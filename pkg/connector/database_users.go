package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"go.mongodb.org/atlas-sdk/v20250312006/admin"
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
		rs.WithStatus(v2.UserTrait_Status_STATUS_ENABLED), // The only possible state for this type of user.
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
	users, resp, err := o.client.DatabaseUsersApi.ListDatabaseUsers(
		ctx,
		parentResourceID.GetResource(),
	).IncludeCount(true).PageNum(page).ItemsPerPage(resourcePageSize).Execute() //nolint:bodyclose // The SDK handles closing the response body
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to list database users: %w", parseToUHttpError(resp, err))
	}

	if users.Results == nil {
		return nil, "", nil, nil
	}

	var resources []*v2.Resource
	for _, user := range *users.Results {
		resource, err := newDatabaseUserResource(ctx, parentResourceID, user)
		if err != nil {
			return nil, "", nil, fmt.Errorf("failed to create database user resource: %w", err)
		}

		resources = append(resources, resource)
	}

	if isLastPage(len(*users.Results), resourcePageSize) {
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

func (o *databaseUserBuilder) Delete(ctx context.Context, resourceId *v2.ResourceId, parentResourceID *v2.ResourceId) (annotations.Annotations, error) {
	dbUserId := resourceId.Resource

	if parentResourceID == nil {
		return nil, fmt.Errorf("database user must have a parent resource: parent resource ID is nil")
	}

	groupId := parentResourceID.Resource

	resp, err := o.client.DatabaseUsersApi.DeleteDatabaseUser(ctx, groupId, "admin", dbUserId).Execute() //nolint:bodyclose // The SDK handles closing the response body
	if err != nil {
		return nil, fmt.Errorf("failed to delete database user: %w", parseToUHttpError(resp, err))
	}

	return nil, nil
}
