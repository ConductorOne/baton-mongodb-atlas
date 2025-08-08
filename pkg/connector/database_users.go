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

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"go.mongodb.org/atlas-sdk/v20231001002/admin"
)

type databaseUserBuilder struct {
	resourceType    *v2.ResourceType
	client          *admin.APIClient
	createInviteKey bool
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

func newDatabaseUserBuilder(client *admin.APIClient, createInviteKey bool) *databaseUserBuilder {
	return &databaseUserBuilder{
		resourceType:    databaseUserResourceType,
		client:          client,
		createInviteKey: createInviteKey,
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
	users, _, err := o.client.DatabaseUsersApi.ListDatabaseUsers(ctx, parentResourceID.GetResource()).IncludeCount(true).PageNum(page).ItemsPerPage(resourcePageSize).Execute() //nolint:bodyclose // The SDK handles closing the response body
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
func (o *databaseUserBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *databaseUserBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func (o *databaseUserBuilder) CreateAccount(ctx context.Context, accountInfo *v2.AccountInfo, credentialOptions *v2.CredentialOptions) (connectorbuilder.CreateAccountResponse, []*v2.PlaintextData, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

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

	databaseName, ok := profile["databaseName"].(string)
	if databaseName == "" || !ok {
		return nil, nil, annotations.Annotations{}, fmt.Errorf("databaseName is empty")
	}

	username, ok := profile["username"].(string)
	if username == "" || !ok {
		return nil, nil, annotations.Annotations{}, fmt.Errorf("username is empty")
	}

	if o.createInviteKey {
		err := o.createUserIfNotExists(ctx, orgId, email, profile)
		if err != nil {
			l.Error(
				"failed to create organization invitation",
				zap.Error(err),
			)
			return nil, nil, nil, err
		}
	}

	password, err := crypto.GeneratePassword(credentialOptions)
	if err != nil {
		return nil, nil, nil, err
	}

	// TODO(golds): Needs to support more usernames
	// https://www.mongodb.com/docs/api/doc/atlas-admin-api-v2/operation/operation-createdatabaseuser#operation-createdatabaseuser-body-application-vnd-atlas-2023-01-01-json-username
	_, _, err = o.client.DatabaseUsersApi.CreateDatabaseUser(
		ctx,
		groupId,
		&admin.CloudDatabaseUser{
			GroupId:      groupId,
			Password:     &password,
			Username:     username,
			DatabaseName: databaseName,
			Roles: []admin.DatabaseUserRole{
				{
					DatabaseName: databaseName,
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

	userFromApi, _, err := o.client.DatabaseUsersApi.GetDatabaseUser(
		ctx,
		groupId,
		databaseName,
		username,
	).Execute() //nolint:bodyclose // The SDK handles closing the response body
	if err != nil {
		return nil, nil, nil, err
	}

	resource, err := newDatabaseUserResource(ctx, &v2.ResourceId{
		ResourceType: projectResourceType.Id,
		Resource:     groupId,
	}, *userFromApi)

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

func (o *databaseUserBuilder) CreateAccountCapabilityDetails(ctx context.Context) (*v2.CredentialDetailsAccountProvisioning, annotations.Annotations, error) {
	return &v2.CredentialDetailsAccountProvisioning{
		SupportedCredentialOptions: []v2.CapabilityDetailCredentialOption{
			v2.CapabilityDetailCredentialOption_CAPABILITY_DETAIL_CREDENTIAL_OPTION_RANDOM_PASSWORD,
		},
		PreferredCredentialOption: v2.CapabilityDetailCredentialOption_CAPABILITY_DETAIL_CREDENTIAL_OPTION_RANDOM_PASSWORD,
	}, nil, nil
}

func (o *databaseUserBuilder) createUserIfNotExists(ctx context.Context, orgId, email string, profile map[string]any) error {
	l := ctxzap.Extract(ctx)

	_, httpResponse, err := o.client.OrganizationsApi.CreateOrganizationInvitation(
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
		if httpResponse != nil && httpResponse.StatusCode == http.StatusConflict {
			l.Info("user already exists, skipping creation", zap.String("email", email))
			return nil
		}
		l.Error(
			"failed to create organization invitation",
			zap.Error(err),
		)
		return err
	}

	return nil
}
