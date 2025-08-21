package connector

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/crypto"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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
	GetOrgMembershipStatus() string
}

type userBuilder struct {
	resourceType    *v2.ResourceType
	client          *admin.APIClient
	createInviteKey bool
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
		rs.WithEmail(user.GetUsername(), true),
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

func (o *userBuilder) CreateAccount(ctx context.Context, accountInfo *v2.AccountInfo, credentialOptions *v2.CredentialOptions) (connectorbuilder.CreateAccountResponse, []*v2.PlaintextData, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	var err error

	profile := accountInfo.Profile.AsMap()

	orgId, ok := profile["organizationId"].(string)
	if orgId == "" || !ok {
		return nil, nil, annotations.Annotations{}, fmt.Errorf("organizationId is empty")
	}

	email, ok := profile["email"].(string)
	if email == "" || !ok {
		return nil, nil, annotations.Annotations{}, fmt.Errorf("email is empty")
	}

	groupId, ok := profile["groupId"].(string)
	if groupId == "" || !ok {
		return nil, nil, annotations.Annotations{}, fmt.Errorf("groupId is empty")
	}

	username, ok := profile["username"].(string)
	if username == "" || !ok {
		return nil, nil, annotations.Annotations{}, fmt.Errorf("username is empty")
	}

	var userId string

	if o.createInviteKey {
		l.Info("creating organization user")
		userId, err = o.createUserIfNotExists(ctx, orgId, email, profile)
		if err != nil {
			l.Error(
				"failed to create organization invitation",
				zap.Error(err),
			)
			return nil, nil, nil, err
		}
	}

	l.Info("creating database user", zap.String("userId", userId))

	var user atlasUserResponse
	if userId != "" {
		user, _, err = o.client.MongoDBCloudUsersApi.GetOrganizationUser(ctx, orgId, userId).Execute() //nolint:bodyclose // The SDK handles closing the response body
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to get user by id: %w", err)
		}
	} else {
		response, _, err := o.client.MongoDBCloudUsersApi.ListOrganizationUsers(ctx, orgId).Username(email).Execute() //nolint:bodyclose // The SDK handles closing the response body
		if err != nil {
			if atlasErr, ok := admin.AsError(err); ok {
				switch atlasErr.ErrorCode {
				case "CANNOT_ADD_PENDING_USER":
					return nil, nil, nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("the user '%s' has a pending invite in the organization", email))
				case "NOT_USER_ADMIN":
					return nil, nil, nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("the user '%s' is not in the organization either received an invite, enable createInviteKey to create the invite", email))
				}
			}

			return nil, nil, nil, fmt.Errorf("failed to get user by username: %w", err)
		}

		if response.Results == nil {
			return nil, nil, nil, fmt.Errorf("user '%s' not found, results is nil", email)
		}

		for _, userResponse := range *response.Results {
			if userResponse.GetUsername() == email {
				user = &userResponse
				break
			}
		}

		if user == nil {
			return nil, nil, nil, fmt.Errorf("user '%s' not found", email)
		}
	}

	password, err := crypto.GeneratePassword(credentialOptions)
	if err != nil {
		return nil, nil, nil, err
	}

	defaultDatabase := "admin"

	// TODO(golds): Needs to support more usernames
	// https://www.mongodb.com/docs/api/doc/atlas-admin-api-v2/operation/operation-createdatabaseuser#operation-createdatabaseuser-body-application-vnd-atlas-2023-01-01-json-username
	_, _, err = o.client.DatabaseUsersApi.CreateDatabaseUser(
		ctx,
		groupId,
		&admin.CloudDatabaseUser{
			GroupId:      groupId,
			Password:     &password,
			Username:     username,
			DatabaseName: defaultDatabase,
			Roles: &[]admin.DatabaseUserRole{
				{
					DatabaseName: defaultDatabase,
					RoleName:     "read",
				},
			},
		},
	).Execute() //nolint:bodyclose // The SDK handles closing the response body
	if err != nil {
		l.Error(
			"failed to create database user",
			zap.Error(err),
		)
		return nil, nil, nil, err
	}

	resource, err := newUserResource(
		ctx,
		&v2.ResourceId{
			ResourceType: organizationResourceType.Id,
			Resource:     orgId,
		},
		user,
	)
	if err != nil {
		return nil, nil, nil, err
	}

	response := &v2.CreateAccountResponse_SuccessResult{
		IsCreateAccountResult: true,
		Resource:              resource,
	}

	plaintextData := []*v2.PlaintextData{
		{
			Name:        "password",
			Description: "The password for the database user",
			Schema:      "text/plain",
			Bytes:       []byte(password),
		},
	}

	return response, plaintextData, nil, err
}

func parseStrList(strFrom any, defaultValue []string) *[]string {
	str, ok := strFrom.(string)
	if !ok {
		return &defaultValue
	}

	if str == "" {
		return &defaultValue
	}

	temp := strings.Split(strings.TrimSpace(str), ",")
	return &temp
}

func (o *userBuilder) CreateAccountCapabilityDetails(ctx context.Context) (*v2.CredentialDetailsAccountProvisioning, annotations.Annotations, error) {
	return &v2.CredentialDetailsAccountProvisioning{
		SupportedCredentialOptions: []v2.CapabilityDetailCredentialOption{
			v2.CapabilityDetailCredentialOption_CAPABILITY_DETAIL_CREDENTIAL_OPTION_RANDOM_PASSWORD,
		},
		PreferredCredentialOption: v2.CapabilityDetailCredentialOption_CAPABILITY_DETAIL_CREDENTIAL_OPTION_RANDOM_PASSWORD,
	}, nil, nil
}

func (o *userBuilder) createUserIfNotExists(ctx context.Context, orgId, email string, profile map[string]any) (string, error) {
	l := ctxzap.Extract(ctx)

	orgUser, httpResponse, err := o.client.MongoDBCloudUsersApi.CreateOrganizationUser(
		ctx,
		orgId,
		&admin.OrgUserRequest{
			Username: email,
			Roles: admin.OrgUserRolesRequest{
				OrgRoles:             *parseStrList(profile["roles"], []string{"ORG_MEMBER"}),
				GroupRoleAssignments: &[]admin.GroupRoleAssignment{},
			},
			TeamIds: parseStrList(profile["teamIds"], []string{}),
		},
	).Execute() //nolint:bodyclose // The SDK handles closing the response body

	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == http.StatusConflict {
			l.Info(
				"user already exists, skipping creation",
				zap.String("email", email),
				zap.String("orgId", orgId),
			)
			return "", nil
		}
		l.Error(
			"failed to create organization invitation",
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to create organization invitation: %w", err)
	}

	return orgUser.Id, nil
}

func newUserBuilder(client *admin.APIClient, createInviteKey bool) *userBuilder {
	return &userBuilder{
		resourceType:    userResourceType,
		client:          client,
		createInviteKey: createInviteKey,
	}
}
