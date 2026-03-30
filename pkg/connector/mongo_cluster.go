package connector

import (
	"context"
	"fmt"
	"strconv"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"go.mongodb.org/atlas-sdk/v20250312006/admin"
)

const (
	clusterStateDeleting = "DELETING"
	clusterStateDeleted  = "DELETED"
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
func (o *mongoClusterBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, opts rs.SyncOpAttrs) ([]*v2.Resource, *rs.SyncOpResults, error) {
	if parentResourceID == nil {
		return nil, nil, nil
	}

	if parentResourceID.ResourceType != projectResourceType.Id {
		return nil, nil, fmt.Errorf("invalid parent resource type: expected %s, got %s", projectResourceType.Id, parentResourceID.ResourceType)
	}

	currentPage := 1

	if opts.PageToken.Token != "" {
		tempPage, err := strconv.Atoi(opts.PageToken.Token)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid pagination token: %w", err)
		}

		currentPage = tempPage
	}

	response, resp, err := o.client.ClustersApi.ListClusters(ctx, parentResourceID.GetResource()).
		PageNum(currentPage).
		IncludeDeletedWithRetainedBackups(true).
		Execute() //nolint:bodyclose // The SDK handles closing the response body
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list clusters: %w", parseToUHttpError(resp, err))
	}

	resources := make([]*v2.Resource, 0, len(response.GetResults()))

	for _, cluster := range response.GetResults() {
		resource, err := newMongoClusterResource(cluster, parentResourceID, o.enableSyncDatabases)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create resource: %w", err)
		}

		resources = append(resources, resource)
	}

	nextPage := ""
	if response.Results == nil || len(*response.Results) != 0 {
		nextPage = strconv.Itoa(currentPage + 1)
	}

	return resources, &rs.SyncOpResults{NextPageToken: nextPage}, nil
}

func newMongoClusterResource(
	cluster admin.ClusterDescription20240805,
	parentId *v2.ResourceId,
	enableSyncDatabases bool,
) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"cluster_id":    cluster.GetId(),
		"cluster_name":  cluster.GetName(),
		"cluster_type":  cluster.GetClusterType(),
		"group_id":      cluster.GetGroupId(),
		"cluster_state": cluster.GetStateName(),
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

	state := cluster.GetStateName()
	// Cluster is available if it is not deleting or deleted. If the cluster is in any of these states
	// we omit the database sync as the cluster cannot be reached.
	clusterAvailable := state != clusterStateDeleting && state != clusterStateDeleted
	if enableSyncDatabases && clusterAvailable {
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
func (o *mongoClusterBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ rs.SyncOpAttrs) ([]*v2.Entitlement, *rs.SyncOpResults, error) {
	return nil, nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *mongoClusterBuilder) Grants(ctx context.Context, resource *v2.Resource, opts rs.SyncOpAttrs) ([]*v2.Grant, *rs.SyncOpResults, error) {
	return nil, nil, nil
}
