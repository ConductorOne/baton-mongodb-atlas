package connector

import (
	"context"
	"fmt"
	"strings"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"

	"github.com/conductorone/baton-mongodb-atlas/pkg/connector/mongodriver"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"go.mongodb.org/atlas-sdk/v20250312006/admin"
)

type collectionBuilder struct {
	client      *admin.APIClient
	mongodriver *mongodriver.MongoDriver
}

func (o *collectionBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return collectionResourceType
}

func newCollectionBuilder(client *admin.APIClient, mongodriver *mongodriver.MongoDriver) *collectionBuilder {
	return &collectionBuilder{
		client:      client,
		mongodriver: mongodriver,
	}
}

// List returns all the users from the database as resource objects.
// Users include a UserTrait because they are the 'shape' of a standard user.
func (o *collectionBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	if parentResourceID == nil {
		return nil, "", nil, nil
	}

	if parentResourceID.ResourceType != databaseResourceType.Id {
		return nil, "", nil, fmt.Errorf("invalid parent resource type: %s", parentResourceID.ResourceType)
	}

	splited := strings.Split(parentResourceID.Resource, "/")
	if len(splited) != 3 {
		return nil, "", nil, fmt.Errorf("invalid parent resource ID: %s", parentResourceID.Resource)
	}

	groupID := splited[0]
	clusterName := splited[1]
	dbName := splited[2]

	_, client, err := o.mongodriver.Connect(ctx, groupID, clusterName)
	if err != nil {
		l.Error("failed to connect to MongoDB Atlas cluster", zap.String("group_id", groupID), zap.String("cluster_name", clusterName), zap.Error(err))
		return nil, "", nil, err
	}

	db := client.Database(dbName, nil)

	collections, err := db.ListCollectionNames(ctx, bson.M{}, nil)
	if err != nil {
		if strings.Contains(err.Error(), "(Unauthorized) not authorized") {
			l.Info("unauthorized to list collections skipping", zap.String("group_id", groupID), zap.String("cluster_name", clusterName), zap.String("db_name", dbName))

			return nil, "", nil, nil
		}

		return nil, "", nil, err
	}

	resources := make([]*v2.Resource, 0)

	for _, collectionName := range collections {
		resource, err := newCollectionResource(groupID, clusterName, dbName, collectionName, parentResourceID)
		if err != nil {
			return nil, "", nil, wrapError(err, "failed to create resource")
		}

		resources = append(resources, resource)
	}

	return resources, "", nil, nil
}

func newCollectionResource(
	groupID string,
	clusterName string,
	dbName string,
	collectionName string,
	parentId *v2.ResourceId,
) (*v2.Resource, error) {
	id := fmt.Sprintf("%s/%s/%s/%s", groupID, clusterName, dbName, collectionName)

	resource, err := rs.NewAppResource(
		fmt.Sprintf("%s - %s", dbName, collectionName),
		collectionResourceType,
		id,
		nil,
		rs.WithParentResourceID(parentId),
	)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

// Entitlements always returns an empty slice for users.
func (o *collectionBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *collectionBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}
