package connector

import (
	"context"
	"io"
	"time"

	"github.com/conductorone/baton-mongodb-atlas/pkg/connector/mongodriver"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"go.mongodb.org/atlas-sdk/v20250312006/admin"
)

type MongoDB struct {
	client          *admin.APIClient
	createInviteKey bool
	mongodriver     *mongodriver.MongoDriver
}

// ResourceSyncers returns a ResourceSyncer for each resource type that should be synced from the upstream service.
func (d *MongoDB) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncer {
	return []connectorbuilder.ResourceSyncer{
		newOrganizationBuilder(d.client),
		newUserBuilder(d.client, d.createInviteKey),
		newTeamBuilder(d.client),
		newProjectBuilder(d.client),
		newDatabaseUserBuilder(d.client),
		newMongoClusterBuilder(d.client),
		newDatabaseBuilder(d.client, d.mongodriver),
		newCollectionBuilder(d.client, d.mongodriver),
	}
}

// Asset takes an input AssetRef and attempts to fetch it using the connector's authenticated http client
// It streams a response, always starting with a metadata object, following by chunked payloads for the asset.
func (d *MongoDB) Asset(ctx context.Context, asset *v2.AssetRef) (string, io.ReadCloser, error) {
	return "", nil, nil
}

// Metadata returns metadata about the connector.
func (d *MongoDB) Metadata(ctx context.Context) (*v2.ConnectorMetadata, error) {
	return &v2.ConnectorMetadata{
		DisplayName: "MongoDB Atlas Connector",
		Description: "Provides access to MongoDB Atlas resources.",
		AccountCreationSchema: &v2.ConnectorAccountCreationSchema{
			FieldMap: map[string]*v2.ConnectorAccountCreationSchema_Field{
				"email": {
					DisplayName: "Email",
					Required:    true,
					Description: "The email address of the MongoDB Atlas account.",
					Field: &v2.ConnectorAccountCreationSchema_Field_StringField{
						StringField: &v2.ConnectorAccountCreationSchema_StringField{},
					},
					Order: 1,
				},
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
			},
		},
	}, nil
}

// Validate is called to ensure that the connector is properly configured. It should exercise any API credentials
// to be sure that they are valid.
func (d *MongoDB) Validate(ctx context.Context) (annotations.Annotations, error) {
	return nil, nil
}

// New returns a new instance of the connector.
func New(ctx context.Context, publicKey, privateKey string, createInviteKey bool) (*MongoDB, error) {
	client, err := admin.NewClient(admin.UseDigestAuth(publicKey, privateKey))
	if err != nil {
		return nil, err
	}

	return &MongoDB{
		client:          client,
		createInviteKey: createInviteKey,
		mongodriver:     mongodriver.NewMongoDriver(client, time.Minute*30),
	}, nil
}
