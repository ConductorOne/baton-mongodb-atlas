package connector

import (
	"context"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"go.mongodb.org/atlas-sdk/v20250312006/admin"
)

type atlasUserResponse interface {
	GetId() string
	GetFirstName() string
	GetLastName() string
	GetUsername() string
	GetCountry() string
}

type userBuilder struct {
	resourceType *v2.ResourceType
	client       *admin.APIClient
}

func (o *userBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return userResourceType
}

func newUserResource(ctx context.Context, organizationId *v2.ResourceId, user atlasUserResponse) (*v2.Resource, error) {
	userId := user.GetId()

	profile := map[string]interface{}{
		"first_name": user.GetFirstName(),
		"last_name":  user.GetLastName(),
		"email":      user.GetUsername(),
		"login":      user.GetUsername(),
		"user_id":    userId,
		"county":     user.GetCountry(),
	}

	userTraits := []rs.UserTraitOption{
		rs.WithUserProfile(profile),
		rs.WithUserLogin(user.GetUsername()),
		rs.WithStatus(v2.UserTrait_Status_STATUS_UNSPECIFIED),
	}

	resource, err := rs.NewUserResource(
		user.GetUsername(),
		userResourceType,
		userId,
		userTraits,
		rs.WithParentResourceID(organizationId),
	)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

// List returns all the users from the database as resource objects.
// Users include a UserTrait because they are the 'shape' of a standard user.
func (o *userBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	if parentResourceID == nil {
		return nil, "", nil, nil
	}

	bag, page, err := parsePageToken(pToken.Token, &v2.ResourceId{ResourceType: o.resourceType.Id})
	if err != nil {
		return nil, "", nil, err
	}

	users, _, err := o.client.MongoDBCloudUsersApi.ListOrganizationUsers(ctx, parentResourceID.GetResource()).PageNum(page).ItemsPerPage(resourcePageSize).Execute() //nolint:bodyclose // The SDK handles closing the response body
	if err != nil {
		return nil, "", nil, wrapError(err, "failed to list users")
	}

	if users.Results == nil {
		return nil, "", nil, nil
	}

	var resources []*v2.Resource
	for _, user := range *users.Results {
		resource, err := newUserResource(ctx, parentResourceID, &user)
		if err != nil {
			return nil, "", nil, wrapError(err, "failed to create user resource")
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
func (o *userBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *userBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func newUserBuilder(client *admin.APIClient) *userBuilder {
	return &userBuilder{
		resourceType: userResourceType,
		client:       client,
	}
}
