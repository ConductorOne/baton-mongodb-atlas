package connector

import (
	"context"
	"fmt"
	"strings"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"

	"github.com/conductorone/baton-mongodb-atlas/pkg/connector/mongodriver"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"go.mongodb.org/mongo-driver/mongo"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"go.mongodb.org/atlas-sdk/v20250312006/admin"
)

type databaseBuilder struct {
	client      *admin.APIClient
	mongodriver *mongodriver.MongoDriver
}

func (o *databaseBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return databaseResourceType
}

func newDatabaseBuilder(client *admin.APIClient, mongodriver *mongodriver.MongoDriver) *databaseBuilder {
	return &databaseBuilder{
		client:      client,
		mongodriver: mongodriver,
	}
}

// List returns all the users from the database as resource objects.
// Users include a UserTrait because they are the 'shape' of a standard user.
func (o *databaseBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	if parentResourceID == nil {
		return nil, "", nil, nil
	}

	if parentResourceID.ResourceType != mongoClusterResourceType.Id {
		return nil, "", nil, fmt.Errorf("invalid parent resource type: %s", parentResourceID.ResourceType)
	}

	splited := strings.Split(parentResourceID.Resource, "/")
	if len(splited) != 3 {
		return nil, "", nil, fmt.Errorf("invalid parent resource ID: %s", parentResourceID.Resource)
	}

	groupID := splited[0]
	// clusterID := splited[1]
	clusterName := splited[2]

	_, client, err := o.mongodriver.Connect(ctx, groupID, clusterName)
	if err != nil {
		l.Error("failed to connect to MongoDB Atlas cluster", zap.String("group_id", groupID), zap.String("cluster_name", clusterName), zap.Error(err))
		return nil, "", nil, err
	}

	databases, err := client.ListDatabases(ctx, bson.D{}, nil)
	if err != nil {
		l.Error("failed to list databases", zap.Error(err))
		return nil, "", nil, err
	}

	resources := make([]*v2.Resource, 0)

	for _, database := range databases.Databases {
		resource, err := newDatabaseResource(groupID, clusterName, database, parentResourceID)
		if err != nil {
			return nil, "", nil, wrapError(err, "failed to create resource")
		}

		resources = append(resources, resource)
	}

	return resources, "", nil, nil
}

func newDatabaseResource(
	groupID string,
	clusterName string,
	db mongo.DatabaseSpecification,
	parentId *v2.ResourceId,
) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"db_name": db.Name,
	}

	appTraits := []rs.AppTraitOption{
		rs.WithAppProfile(profile),
	}

	id := fmt.Sprintf("%s/%s/%s", groupID, clusterName, db.Name)

	resource, err := rs.NewAppResource(
		fmt.Sprintf("%s - %s", clusterName, db.Name),
		databaseResourceType,
		id,
		appTraits,
		rs.WithParentResourceID(parentId),
		rs.WithAnnotation(&v2.ChildResourceType{
			ResourceTypeId: collectionResourceType.Id,
		}),
	)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

// Entitlements always returns an empty slice for users.
func (o *databaseBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *databaseBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}
