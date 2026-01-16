package mongodriver

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/conductorone/baton-mongodb-atlas/pkg/connector/mongoconfig"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/crypto"
	"go.mongodb.org/atlas-sdk/v20250312006/admin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type userSessionTuple struct {
	pass string
	user *admin.CloudDatabaseUser
}

type MongoDriver struct {
	adminClient *admin.APIClient
	accountTTL  time.Duration
	mutex       sync.Mutex

	// groupID -> userSessionTuple
	accountsPerGroupId map[string]*userSessionTuple
	clients            map[string]*mongo.Client
	proxy              *mongoconfig.MongoProxy
}

func NewMongoDriver(
	adminClient *admin.APIClient,
	accountTTL time.Duration,
	proxy *mongoconfig.MongoProxy,
) *MongoDriver {
	return &MongoDriver{
		adminClient:        adminClient,
		accountTTL:         accountTTL,
		accountsPerGroupId: make(map[string]*userSessionTuple),
		clients:            make(map[string]*mongo.Client),
		proxy:              proxy,
	}
}

func (m *MongoDriver) createUser(ctx context.Context, groupId string) (*admin.CloudDatabaseUser, string, error) {
	password, err := crypto.GeneratePassword(ctx, &v2.LocalCredentialOptions{
		Options: &v2.LocalCredentialOptions_RandomPassword_{
			RandomPassword: &v2.LocalCredentialOptions_RandomPassword{
				Length: 30,
			},
		},
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate password: %w", err)
	}

	id, err := randomString(10)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate random string: %w", err)
	}

	username := "baton_mongodb_atlas_" + id

	deletedTime := time.Now().UTC().Add(m.accountTTL)
	userDescription := "Created by Baton, Automatically deleted after " + deletedTime.Format(time.RFC3339)

	dbUser, _, err := m.adminClient.DatabaseUsersApi.CreateDatabaseUser(
		ctx,
		groupId,
		&admin.CloudDatabaseUser{
			GroupId:         groupId,
			Password:        &password,
			Username:        username,
			DatabaseName:    "admin",
			DeleteAfterDate: &deletedTime,
			Description:     &userDescription,
			Roles: &[]admin.DatabaseUserRole{
				{
					DatabaseName: "admin",
					RoleName:     "readAnyDatabase",
				},
			},
		},
	).Execute()
	if err != nil {
		return nil, "", fmt.Errorf("failed to create database user: %w", err)
	}

	return dbUser, password, nil
}

func (m *MongoDriver) Close(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, client := range m.clients {
		err := client.Disconnect(ctx)
		if err != nil {
			return fmt.Errorf("failed to disconnect mongo client: %w", err)
		}
	}
	return nil
}

func (m *MongoDriver) Connect(ctx context.Context, groupID, clusterName string) (*admin.CloudDatabaseUser, *mongo.Client, error) {
	l := ctxzap.Extract(ctx)

	m.mutex.Lock()
	defer m.mutex.Unlock()

	accountTuple, ok := m.accountsPerGroupId[groupID]
	if !ok {
		user, password, err := m.createUser(ctx, groupID)
		if err != nil {
			return nil, nil, err
		}
		accountTuple = &userSessionTuple{
			user: user,
			pass: password,
		}
		m.accountsPerGroupId[groupID] = accountTuple
	}

	clientKey := fmt.Sprintf("%s-%s", groupID, clusterName)
	client, ok := m.clients[clientKey]
	if ok {
		return accountTuple.user, client, nil
	}

	time.Sleep(time.Second * 5) // Wait for the user to be fully created and available

	connStr, err := m.connectionString(ctx, groupID, clusterName)
	if err != nil {
		return nil, nil, err
	}

	escapedPwd := url.QueryEscape(accountTuple.pass)
	uri := fmt.Sprintf("%s://%s:%s@%s", connStr.scheme, accountTuple.user.Username, escapedPwd, connStr.hosts)

	// Get SOCKS5 dialer if proxy is configured - this routes all connections
	// including DNS resolution through the proxy
	dialer, err := m.proxy.Dialer()
	if err != nil {
		l.Error("Failed to create SOCKS5 dialer", zap.Error(err))
		return nil, nil, fmt.Errorf("failed to create SOCKS5 dialer: %w", err)
	}

	if dialer != nil {
		l.Info(
			"MongoDB connection will use SOCKS5 proxy",
			zap.String("proxy_address", m.proxy.Address()),
			zap.String("cluster_name", clusterName),
		)
	}

	for i := 0; i < 10; i++ {
		l.Info(
			"Trying to connect to MongoDB",
			zap.Int("attempt", i+1),
			zap.String("cluster_name", clusterName),
			zap.String("scheme", connStr.scheme),
		)

		opts := options.Client().
			ApplyURI(uri).
			SetMaxConnIdleTime(60 * time.Second).
			SetMaxPoolSize(10)

		if dialer != nil {
			opts.SetDialer(dialer)
			// Increase timeouts when using SOCKS5 proxy since connections take longer:
			// proxy connect -> SOCKS5 handshake -> proxy connects to MongoDB -> TLS handshake
			opts.SetConnectTimeout(60 * time.Second)
			opts.SetServerSelectionTimeout(60 * time.Second)
		}

		client, err := mongo.Connect(ctx, opts)
		if err != nil {
			l.Error(
				"Failed to connect to MongoDB",
				zap.Error(err),
				zap.Int("attempt", i+1),
				zap.String("cluster_name", clusterName),
			)
			time.Sleep(time.Second * 2) // Retry after a delay
			continue
		}

		if err = client.Ping(ctx, nil); err != nil {
			_ = client.Disconnect(ctx)
			l.Error(
				"Failed to ping MongoDB",
				zap.Error(err),
				zap.Int("attempt", i+1),
				zap.String("username", accountTuple.user.Username),
				zap.String("cluster_name", clusterName),
			)
			time.Sleep(time.Second * 2) // Retry after a delay
			continue
		}

		l.Info(
			"Successfully connected to MongoDB",
			zap.String("cluster_name", clusterName),
			zap.String("username", accountTuple.user.Username),
			zap.Int("attempts", i+1),
		)
		m.clients[clientKey] = client
		return accountTuple.user, client, nil
	}

	l.Error(
		"Failed to connect to MongoDB after multiple attempts",
		zap.String("cluster_name", clusterName),
		zap.Int("max_attempts", 10),
	)
	return nil, nil, fmt.Errorf("failed to connect to MongoDB after multiple attempts")
}

type connectionInfo struct {
	scheme string
	hosts  string
}

func (m *MongoDriver) connectionString(ctx context.Context, groupID, clusterName string) (*connectionInfo, error) {
	l := ctxzap.Extract(ctx)

	clusterInfo, _, err := m.adminClient.ClustersApi.GetCluster(ctx, groupID, clusterName).
		Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster: %w", err)
	}

	// Prefer Standard (mongodb://) connection string as it doesn't require DNS SRV lookups.
	// This is important when using a SOCKS5 proxy since SRV lookups may not go through the proxy.
	if clusterInfo.ConnectionStrings.Standard != nil && *clusterInfo.ConnectionStrings.Standard != "" {
		l.Info(
			"Using standard MongoDB connection string (no SRV lookup required)",
			zap.String("cluster_name", clusterName),
			zap.String("group_id", groupID),
		)
		return &connectionInfo{
			scheme: "mongodb",
			hosts:  strings.TrimPrefix(*clusterInfo.ConnectionStrings.Standard, "mongodb://"),
		}, nil
	}

	// Fall back to SRV connection string if Standard is not available
	if clusterInfo.ConnectionStrings.StandardSrv != nil && *clusterInfo.ConnectionStrings.StandardSrv != "" {
		l.Warn(
			"Standard connection string not available, falling back to SRV (DNS lookups may not go through proxy)",
			zap.String("cluster_name", clusterName),
			zap.String("group_id", groupID),
		)
		return &connectionInfo{
			scheme: "mongodb+srv",
			hosts:  strings.TrimPrefix(*clusterInfo.ConnectionStrings.StandardSrv, "mongodb+srv://"),
		}, nil
	}

	l.Error(
		"Cluster does not have any connection string",
		zap.String("cluster_name", clusterName),
		zap.String("group_id", groupID),
	)
	return nil, fmt.Errorf("cluster %s does not have a connection string", clusterName)
}
