package connector

import (
	"context"
	"fmt"
	"strings"

	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"go.mongodb.org/atlas-sdk/v20231001002/admin"
)

type userBuilder struct {
	resourceType *v2.ResourceType
	client       *admin.APIClient
}

func (o *userBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return userResourceType
}

func newUserResource(ctx context.Context, organizationId *v2.ResourceId, user admin.CloudAppUser) (*v2.Resource, error) {
	userId := *user.Id

	profile := map[string]interface{}{
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"email":      user.EmailAddress,
		"login":      user.Username,
		"user_id":    userId,
		"county":     user.Country,
	}

	userTraits := []rs.UserTraitOption{
		rs.WithUserProfile(profile),
		rs.WithUserLogin(user.Username),
		rs.WithStatus(v2.UserTrait_Status_STATUS_UNSPECIFIED),
	}

	resource, err := rs.NewUserResource(
		user.Username,
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

	users, _, err := o.client.OrganizationsApi.ListOrganizationUsers(ctx, parentResourceID.GetResource()).PageNum(page).ItemsPerPage(resourcePageSize).Execute() //nolint:bodyclose // The SDK handles closing the response body
	if err != nil {
		return nil, "", nil, wrapError(err, "failed to list users")
	}

	var resources []*v2.Resource
	for _, user := range users.Results {
		resource, err := newUserResource(ctx, parentResourceID, user)
		if err != nil {
			return nil, "", nil, wrapError(err, "failed to create user resource")
		}

		resources = append(resources, resource)
	}

	if isLastPage(len(users.Results), resourcePageSize) {
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

func (o *userBuilder) CreateAccount(ctx context.Context, accountInfo *v2.AccountInfo, credentialOptions *v2.CredentialOptions) (connectorbuilder.CreateAccountResponse, []*v2.PlaintextData, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	profile := accountInfo.Profile.AsMap()

	l.Info("aaa", zap.Any("profile", profile))

	orgId, ok := profile["organizationId"].(string)
	if orgId == "" || !ok {
		return nil, nil, annotations.Annotations{}, fmt.Errorf("organizationId is empty")
	}

	email, ok := profile["email"].(string)
	if email == "" || !ok {
		return nil, nil, annotations.Annotations{}, fmt.Errorf("email is empty")
	}

	_, _, err := o.client.OrganizationsApi.CreateOrganizationInvitation(
		ctx,
		orgId,
		&admin.OrganizationInvitationRequest{
			Username:             &email,
			Roles:                parseStrList(profile["roles"], []string{"ORG_MEMBER"}),
			TeamIds:              parseStrList(profile["teamIds"], []string{}),
			GroupRoleAssignments: nil,
		},
	).Execute() //nolint:bodyclose // The SDK handles closing the response body

	if err != nil {
		l.Error(
			"failed to create organization invitation",
			zap.Error(err),
		)
		return nil, nil, nil, err
	}

	response := &v2.CreateAccountResponse_SuccessResult{
		IsCreateAccountResult: true,
	}

	return response, nil, nil, err
}

func parseStrList(strFrom any, defaultValue []string) []string {
	str, ok := strFrom.(string)
	if !ok {
		return defaultValue
	}

	if str == "" {
		return defaultValue
	}

	return strings.Split(strings.TrimSpace(str), ",")
}

func (o *userBuilder) CreateAccountCapabilityDetails(ctx context.Context) (*v2.CredentialDetailsAccountProvisioning, annotations.Annotations, error) {
	return &v2.CredentialDetailsAccountProvisioning{
		SupportedCredentialOptions: []v2.CapabilityDetailCredentialOption{
			v2.CapabilityDetailCredentialOption_CAPABILITY_DETAIL_CREDENTIAL_OPTION_NO_PASSWORD,
		},
		PreferredCredentialOption: v2.CapabilityDetailCredentialOption_CAPABILITY_DETAIL_CREDENTIAL_OPTION_NO_PASSWORD,
	}, nil, nil
}

func newUserBuilder(client *admin.APIClient) *userBuilder {
	return &userBuilder{
		resourceType: userResourceType,
		client:       client,
	}
}
