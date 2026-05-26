package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/crypto"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.mongodb.org/atlas-sdk/v20250312006/admin"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
)

type databaseUserBuilder struct {
	resourceType *v2.ResourceType
	client       *admin.APIClient
}

var _ connectorbuilder.AccountManagerV2 = (*databaseUserBuilder)(nil)

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
func (o *databaseUserBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, opts rs.SyncOpAttrs) ([]*v2.Resource, *rs.SyncOpResults, error) {
	if parentResourceID == nil {
		return nil, nil, nil
	}
	bag, page, err := parsePageToken(opts.PageToken.Token, &v2.ResourceId{ResourceType: o.resourceType.Id})
	if err != nil {
		return nil, nil, err
	}
	users, resp, err := o.client.DatabaseUsersApi.ListDatabaseUsers(
		ctx,
		parentResourceID.GetResource(),
	).IncludeCount(true).PageNum(page).ItemsPerPage(resourcePageSize).Execute() //nolint:bodyclose // The SDK handles closing the response body
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list database users: %w", parseToUHttpError(resp, err))
	}

	if users.Results == nil {
		return nil, nil, nil
	}

	var resources []*v2.Resource
	for _, user := range *users.Results {
		resource, err := newDatabaseUserResource(ctx, parentResourceID, user)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create database user resource: %w", err)
		}

		resources = append(resources, resource)
	}

	if isLastPage(len(*users.Results), resourcePageSize) {
		return resources, nil, nil
	}

	nextPage, err := getPageTokenFromPage(bag, page+1)
	if err != nil {
		return nil, nil, err
	}

	return resources, &rs.SyncOpResults{NextPageToken: nextPage}, nil
}

// Entitlements always returns an empty slice for users.
func (o *databaseUserBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ rs.SyncOpAttrs) ([]*v2.Entitlement, *rs.SyncOpResults, error) {
	return nil, nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *databaseUserBuilder) Grants(ctx context.Context, resource *v2.Resource, opts rs.SyncOpAttrs) ([]*v2.Grant, *rs.SyncOpResults, error) {
	return nil, nil, nil
}

func (o *databaseUserBuilder) CreateAccount(ctx context.Context, accountInfo *v2.AccountInfo, credentialOptions *v2.LocalCredentialOptions,
) (connectorbuilder.CreateAccountResponse, []*v2.PlaintextData, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	profile := accountInfo.Profile.AsMap()

	groupId, ok := profile["groupId"].(string)
	if groupId == "" || !ok {
		return nil, nil, nil, uhttp.WrapErrors(codes.InvalidArgument, "mongo-db-connector: groupId is required", fmt.Errorf("groupId field is missing or empty"))
	}

	username, ok := profile["username"].(string)
	if username == "" || !ok {
		return nil, nil, nil, uhttp.WrapErrors(codes.InvalidArgument, "mongo-db-connector: username is required", fmt.Errorf("username field is missing or empty"))
	}

	authType, ok := profile["authType"].(string)
	if authType == "" || !ok {
		authType = AuthTypeScramSHA
	}

	databaseName := getDatabaseNameForAuthType(authType)

	dbUserRequest := &admin.CloudDatabaseUser{
		GroupId:      groupId,
		Username:     username,
		DatabaseName: databaseName,
		Roles: &[]admin.DatabaseUserRole{
			{
				DatabaseName: databaseNameAdmin,
				RoleName:     "read",
			},
		},
	}

	var password string
	var err error
	var plaintextData []*v2.PlaintextData

	switch authType {
	case AuthTypeScramSHA:
		password, err = crypto.GeneratePassword(ctx, credentialOptions)
		if err != nil {
			return nil, nil, nil, uhttp.WrapErrors(codes.Internal, "mongo-db-connector: failed to generate password", err)
		}
		dbUserRequest.Password = &password
		plaintextData = []*v2.PlaintextData{
			{
				Name:        "password",
				Description: "The password for the database user",
				Schema:      "text/plain",
				Bytes:       []byte(password),
			},
		}

	case AuthTypeAWSIAMUser:
		dbUserRequest.AwsIAMType = strPtr(dbTypeUser)

	case AuthTypeX509Customer:
		dbUserRequest.X509Type = strPtr("CUSTOMER")

	case AuthTypeX509Managed:
		dbUserRequest.X509Type = strPtr("MANAGED")

	case AuthTypeLDAPUser:
		dbUserRequest.LdapAuthType = strPtr(dbTypeUser)

	case AuthTypeOIDCWorkload:
		dbUserRequest.OidcAuthType = strPtr(dbTypeUser)

	default:
		return nil, nil, nil, uhttp.WrapErrors(codes.InvalidArgument, fmt.Sprintf("mongo-db-connector: unsupported authentication type: %s", authType))
	}

	l.Info("creating database user",
		zap.String("authType", authType),
		zap.String("databaseName", databaseName),
		zap.String("username", username),
	)

	dbUser, resp, err := o.client.DatabaseUsersApi.CreateDatabaseUser(
		ctx,
		groupId,
		dbUserRequest,
	).Execute() //nolint:bodyclose // The SDK handles closing the response body
	if err != nil {
		l.Error("failed to create database user", zap.Error(err))
		return nil, nil, nil, fmt.Errorf("failed to create database user: %w", parseToUHttpError(resp, err))
	}

	resource, err := newDatabaseUserResource(
		ctx,
		&v2.ResourceId{
			ResourceType: projectResourceType.Id,
			Resource:     groupId,
		},
		*dbUser,
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create database user resource: %w", err)
	}

	response := &v2.CreateAccountResponse_SuccessResult{
		IsCreateAccountResult: true,
		Resource:              resource,
	}

	return response, plaintextData, nil, nil
}

func (o *databaseUserBuilder) CreateAccountCapabilityDetails(ctx context.Context) (*v2.CredentialDetailsAccountProvisioning, annotations.Annotations, error) {
	return &v2.CredentialDetailsAccountProvisioning{
		SupportedCredentialOptions: []v2.CapabilityDetailCredentialOption{
			v2.CapabilityDetailCredentialOption_CAPABILITY_DETAIL_CREDENTIAL_OPTION_RANDOM_PASSWORD,
		},
		PreferredCredentialOption: v2.CapabilityDetailCredentialOption_CAPABILITY_DETAIL_CREDENTIAL_OPTION_RANDOM_PASSWORD,
	}, nil, nil
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
