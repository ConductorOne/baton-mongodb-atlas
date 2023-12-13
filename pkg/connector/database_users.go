package connector

import (
	"context"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"go.mongodb.org/atlas-sdk/v20231001002/admin"
)

type databaseUserBuilder struct {
	resourceType *v2.ResourceType
	client       *admin.APIClient
}

func (o *databaseUserBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return databaseUserResourceType
}

func newDatabaseUserResource(ctx context.Context, projectId *v2.ResourceId, user admin.CloudDatabaseUser) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"username":      user.Username,
		"login":         user.Username,
		"database_name": user.DatabaseName,
	}

	userTraits := []rs.UserTraitOption{
		rs.WithUserProfile(profile),
		rs.WithUserLogin(user.Username),
		rs.WithStatus(v2.UserTrait_Status_STATUS_UNSPECIFIED),
	}

	resource, err := rs.NewUserResource(
		user.Username,
		databaseUserResourceType,
		user.Username,
		userTraits,
		rs.WithParentResourceID(projectId),
	)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

func newDatabaseUserBuilder(client *admin.APIClient) *databaseUserBuilder {
	return &databaseUserBuilder{
		resourceType: databaseUserResourceType,
		client:       client,
	}
}

// List returns all the users from the database as resource objects.
// Users include a UserTrait because they are the 'shape' of a standard user.
func (o *databaseUserBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	if parentResourceID == nil {
		return nil, "", nil, nil
	}

	bag, page, err := parsePageToken(pToken.Token, &v2.ResourceId{ResourceType: o.resourceType.Id})
	if err != nil {
		return nil, "", nil, err
	}

	users, _, err := o.client.DatabaseUsersApi.ListDatabaseUsers(ctx, parentResourceID.Resource).IncludeCount(true).PageNum(page).ItemsPerPage(resourcePageSize).Execute()
	if err != nil {
		return nil, "", nil, wrapError(err, "failed to list database users")
	}

	var resources []*v2.Resource
	for _, user := range users.Results {
		resource, err := newDatabaseUserResource(ctx, parentResourceID, user)
		if err != nil {
			return nil, "", nil, wrapError(err, "failed to create database user resource")
		}

		resources = append(resources, resource)
	}

	if isLastPage(*users.TotalCount, resourcePageSize) {
		return resources, "", nil, nil
	}

	nextPage, err := getPageTokenFromPage(bag, page+1)
	if err != nil {
		return nil, "", nil, err
	}

	return resources, nextPage, nil, nil
}

// Entitlements always returns an empty slice for users.
func (o *databaseUserBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *databaseUserBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}
