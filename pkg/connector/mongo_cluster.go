package connector

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"go.mongodb.org/atlas-sdk/v20250312006/admin"
)

type mongoClusterBuilder struct {
	client              *admin.APIClient
	enableSyncDatabases bool
}

func (o *mongoClusterBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return mongoClusterResourceType
}

func newMongoClusterBuilder(client *admin.APIClient, enableSyncDatabases bool) *mongoClusterBuilder {
	return &mongoClusterBuilder{
		client:              client,
		enableSyncDatabases: enableSyncDatabases,
	}
}

// List returns all the users from the database as resource objects.
// Users include a UserTrait because they are the 'shape' of a standard user.
func (o *mongoClusterBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	if parentResourceID == nil {
		return nil, "", nil, nil
	}

	if parentResourceID.ResourceType != projectResourceType.Id {
		return nil, "", nil, fmt.Errorf("invalid parent resource type: %s", parentResourceID.ResourceType)
	}

	currentPage := 1

	if pToken != nil && pToken.Token != "" {
		tempPage, err := strconv.Atoi(pToken.Token)
		if err != nil {
			return nil, "", nil, errors.Join(errors.New("invalid pagination token"), err)
		}

		currentPage = tempPage
	}

	response, resp, err := o.client.ClustersApi.ListClusters(ctx, parentResourceID.GetResource()).
		PageNum(currentPage).
		IncludeDeletedWithRetainedBackups(true).
		Execute() //nolint:bodyclose // The SDK handles closing the response body
	if err != nil {
		return nil, "", nil, wrapErrorWithStatus(resp, err, "failed to list clusters")
	}

	resources := make([]*v2.Resource, 0, len(response.GetResults()))

	for _, cluster := range response.GetResults() {
		resource, err := newMongoClusterResource(cluster, parentResourceID, o.enableSyncDatabases)
		if err != nil {
			return nil, "", nil, wrapError(err, "failed to create resource")
		}

		resources = append(resources, resource)
	}

	nextPage := ""
	if response.Results == nil || len(*response.Results) != 0 {
		nextPage = strconv.Itoa(currentPage + 1)
	}

	return resources, nextPage, nil, nil
}

func newMongoClusterResource(
	cluster admin.ClusterDescription20240805,
	parentId *v2.ResourceId,
	enableSyncDatabases bool,
) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"cluster_id":   cluster.GetId(),
		"cluster_name": cluster.GetName(),
		"cluster_type": cluster.GetClusterType(),
		"group_id":     cluster.GetGroupId(),
	}

	appTraits := []rs.AppTraitOption{
		rs.WithAppProfile(profile),
	}

	name := cluster.GetName()
	if name == "" {
		name = cluster.GetId()
	}

	id := fmt.Sprintf("%s/%s/%s", parentId.GetResource(), cluster.GetId(), cluster.GetName())

	opts := []rs.ResourceOption{
		rs.WithParentResourceID(parentId),
	}

	if enableSyncDatabases {
		opts = append(opts, rs.WithAnnotation(&v2.ChildResourceType{
			ResourceTypeId: databaseResourceType.Id,
		}))
	}

	resource, err := rs.NewAppResource(
		name,
		mongoClusterResourceType,
		id,
		appTraits,
		opts...,
	)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

// Entitlements always returns an empty slice for users.
func (o *mongoClusterBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *mongoClusterBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}
