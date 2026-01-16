package connector

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/conductorone/baton-mongodb-atlas/pkg/connector/mongoconfig"

	"github.com/conductorone/baton-mongodb-atlas/pkg/connector/mongodriver"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/mongodb-forks/digest"
	"go.uber.org/zap"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"go.mongodb.org/atlas-sdk/v20250312006/admin"
)

type MongoDB struct {
	client                         *admin.APIClient
	createInviteKey                bool
	mongodriver                    *mongodriver.MongoDriver
	enableMongoDriver              bool
	enableSyncDatabases            bool
	deleteDatabaseUserWithReadOnly bool
	mProxy                         *mongoconfig.MongoProxy
}

// ResourceSyncers returns a ResourceSyncer for each resource type that should be synced from the upstream service.
func (d *MongoDB) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncer {
	builders := []connectorbuilder.ResourceSyncer{
		newOrganizationBuilder(d.client),
		newUserBuilder(d.client, d.createInviteKey),
		newTeamBuilder(d.client),
		newProjectBuilder(d.client),
		newDatabaseUserBuilder(d.client),
		newMongoClusterBuilder(d.client, d.enableSyncDatabases),
	}

	if d.enableSyncDatabases {
		builders = append(builders, newDatabaseBuilder(d.client, d.enableMongoDriver, d.mongodriver, d.deleteDatabaseUserWithReadOnly))

		if d.enableMongoDriver {
			builders = append(builders, newCollectionBuilder(d.client, d.mongodriver))
		}
	}

	return builders
}

// Asset takes an input AssetRef and attempts to fetch it using the connector's authenticated http client
// It streams a response, always starting with a metadata object, following by chunked payloads for the asset.
func (d *MongoDB) Asset(ctx context.Context, asset *v2.AssetRef) (string, io.ReadCloser, error) {
	return "", nil, nil
}

// Metadata returns metadata about the connector.
func (d *MongoDB) Metadata(ctx context.Context) (*v2.ConnectorMetadata, error) {
	fields := map[string]*v2.ConnectorAccountCreationSchema_Field{
		"username": {
			DisplayName: "Username",
			Required:    true,
			Description: "The username for the database user.",
			Field: &v2.ConnectorAccountCreationSchema_Field_StringField{
				StringField: &v2.ConnectorAccountCreationSchema_StringField{},
			},
			Order: 2,
		},
		"organizationId": {
			DisplayName: "Organization ID",
			Required:    true,
			Description: "The ID of the MongoDB Atlas organization to which the account belongs.",
			Field: &v2.ConnectorAccountCreationSchema_Field_StringField{
				StringField: &v2.ConnectorAccountCreationSchema_StringField{},
			},
			Placeholder: "Enter Organization ID",
			Order:       3,
		},
		"groupId": {
			DisplayName: "Group ID",
			Required:    true,
			Description: "Unique 24-hexadecimal digit string that identifies the project.",
			Field: &v2.ConnectorAccountCreationSchema_Field_StringField{
				StringField: &v2.ConnectorAccountCreationSchema_StringField{},
			},
			Order: 4,
		},
		"roles": {
			DisplayName: "Roles",
			Required:    false,
			Description: "The roles to assign to the account.",
			Field: &v2.ConnectorAccountCreationSchema_Field_StringListField{
				StringListField: &v2.ConnectorAccountCreationSchema_StringListField{
					DefaultValue: make([]string, 0),
				},
			},
			Order: 5,
		},
		"teamIds": {
			DisplayName: "Team IDs",
			Required:    false,
			Description: "The IDs of the teams to which the account belongs.",
			Field: &v2.ConnectorAccountCreationSchema_Field_StringListField{
				StringListField: &v2.ConnectorAccountCreationSchema_StringListField{
					DefaultValue: make([]string, 0),
				},
			},
			Order: 6,
		},
		"authType": {
			DisplayName: "Authentication Type",
			Required:    false,
			Description: "The authentication method for the database user. Defaults to SCRAM-SHA (password-based). Options: SCRAM-SHA, AWS_IAM_USER, X509_CUSTOMER, X509_MANAGED, LDAP_USER, OIDC_WORKLOAD.",
			Field: &v2.ConnectorAccountCreationSchema_Field_StringField{
				StringField: &v2.ConnectorAccountCreationSchema_StringField{},
			},
			Placeholder: "SCRAM-SHA",
			Order:       7,
		},
	}

	if d.createInviteKey {
		fields["email"] = &v2.ConnectorAccountCreationSchema_Field{
			DisplayName: "Email",
			Required:    true,
			Description: "The email address of the MongoDB Atlas account.",
			Field: &v2.ConnectorAccountCreationSchema_Field_StringField{
				StringField: &v2.ConnectorAccountCreationSchema_StringField{},
			},
			Order: 1,
		}
	}

	return &v2.ConnectorMetadata{
		DisplayName: "MongoDB Atlas Connector",
		Description: "Provides access to MongoDB Atlas resources.",
		AccountCreationSchema: &v2.ConnectorAccountCreationSchema{
			FieldMap: fields,
		},
	}, nil
}

// Validate is called to ensure that the connector is properly configured. It should exercise any API credentials
// to be sure that they are valid.
func (d *MongoDB) Validate(ctx context.Context) (annotations.Annotations, error) {
	return nil, nil
}

// New returns a new instance of the connector.
func New(ctx context.Context, publicKey, privateKey string, createInviteKey, enableSyncDatabases, enableMongoDriver, deleteDatabaseUserWithReadOnly bool, mProxy *mongoconfig.MongoProxy) (*MongoDB, error) {
	l := ctxzap.Extract(ctx)
	clientModifiers := []admin.ClientModifier{}

	// If proxy is enabled, create an HTTP client that routes through SOCKS5
	if mProxy != nil && mProxy.Enabled() {
		l.Info(
			"Configuring SOCKS5 proxy for Atlas API",
			zap.String("proxy_address", mProxy.Address()),
		)

		httpTransport, err := mProxy.HTTPTransport()
		if err != nil {
			l.Error("Failed to create SOCKS5 HTTP transport", zap.Error(err))
			return nil, fmt.Errorf("failed to create SOCKS5 HTTP transport: %w", err)
		}

		// Wrap the SOCKS5 transport with digest auth
		digestTransport := digest.NewTransportWithHTTPRoundTripper(publicKey, privateKey, httpTransport)
		httpClient, err := digestTransport.Client()
		if err != nil {
			l.Error("Failed to create digest HTTP client", zap.Error(err))
			return nil, fmt.Errorf("failed to create digest HTTP client: %w", err)
		}

		clientModifiers = append(clientModifiers, admin.UseHTTPClient(httpClient))
		l.Info("Atlas API client configured to use SOCKS5 proxy")
	} else {
		l.Debug("No proxy configured, using direct connection for Atlas API")
		// No proxy, use standard digest auth
		clientModifiers = append(clientModifiers, admin.UseDigestAuth(publicKey, privateKey))
	}

	client, err := admin.NewClient(clientModifiers...)
	if err != nil {
		return nil, err
	}

	return &MongoDB{
		client:                         client,
		createInviteKey:                createInviteKey,
		mongodriver:                    mongodriver.NewMongoDriver(client, time.Minute*30, mProxy),
		enableSyncDatabases:            enableSyncDatabases,
		enableMongoDriver:              enableMongoDriver,
		deleteDatabaseUserWithReadOnly: deleteDatabaseUserWithReadOnly,
		mProxy:                         mProxy,
	}, nil
}
