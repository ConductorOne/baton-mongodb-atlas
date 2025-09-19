package mongodriver

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/crypto"
	"go.mongodb.org/atlas-sdk/v20250312006/admin"
	"go.mongodb.org/mongo-driver/mongo"
)

type userSessionTuple struct {
	pass string
	user *admin.CloudDatabaseUser
}

type MongoDriver struct {
	adminClient *admin.APIClient
	accountTTL  time.Duration
	mutex       sync.Mutex

	// groupID -> clusterName -> userSessionTuple
	accountsPerGroupId map[string]*userSessionTuple
	clients            map[string]*mongo.Client
}

func NewMongoDriver(adminClient *admin.APIClient, accountTTL time.Duration) *MongoDriver {
	return &MongoDriver{
		adminClient:        adminClient,
		accountTTL:         accountTTL,
		accountsPerGroupId: make(map[string]*userSessionTuple),
		clients:            make(map[string]*mongo.Client),
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
		return nil, "", err
	}

	id, err := randomString(10)
	if err != nil {
		return nil, "", err
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
		return nil, "", err
	}

	return dbUser, password, nil
}

func (m *MongoDriver) Close(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, client := range m.clients {
		err := client.Disconnect(ctx)
		if err != nil {
			return err
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

	uri, err := m.connectionString(ctx, groupID, clusterName)
	if err != nil {
		return nil, nil, err
	}

	escapedPwd := url.QueryEscape(accountTuple.pass)

	uri = fmt.Sprintf("mongodb+srv://%s:%s@%s", accountTuple.user.Username, escapedPwd, uri)

	for i := 0; i < 10; i++ {
		l.Info("Trying to connect to MongoDB", zap.Int("retry", i))

		opts := options.Client().ApplyURI(uri).
			SetMaxConnIdleTime(60 * time.Second).
			SetMaxPoolSize(10)

		client, err := mongo.Connect(ctx, opts)
		if err != nil {
			l.Error("Failed to connect to MongoDB", zap.Error(err))

			time.Sleep(time.Second * 2) // Retry after a delay
			continue
		}

		if err = client.Ping(ctx, nil); err != nil {
			_ = client.Disconnect(ctx)
			l.Error(
				"Failed to ping MongoDB",
				zap.Error(err),
				zap.String("username", accountTuple.user.Username),
				zap.String("cluster_name", clusterName),
			)
			time.Sleep(time.Second * 2) // Retry after a delay
			continue
		}

		m.clients[clientKey] = client

		return accountTuple.user, client, nil
	}

	return nil, nil, fmt.Errorf("failed to connect to MongoDB after multiple attempts")
}

func (m *MongoDriver) connectionString(ctx context.Context, groupID, clusterName string) (string, error) {
	l := ctxzap.Extract(ctx)

	clusterInfo, _, err := m.adminClient.ClustersApi.GetCluster(ctx, groupID, clusterName).
		Execute()
	if err != nil {
		return "", err
	}

	if clusterInfo.ConnectionStrings.StandardSrv == nil {
		l.Error("Cluster does not have a standard connection string", zap.Any("clusterInfo", clusterInfo))
		return "", fmt.Errorf("cluster %s does not have a standard connection string", clusterName)
	}

	return strings.TrimPrefix(*clusterInfo.ConnectionStrings.StandardSrv, "mongodb+srv://"), nil
}
