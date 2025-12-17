package connector

import (
	"context"
	"fmt"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/crypto"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.mongodb.org/atlas-sdk/v20250312006/admin"
	"go.uber.org/zap"
)

// Database user authentication types.
// These constants define the supported authentication methods for MongoDB Atlas database users.
const (
	// AuthTypeScramSHA is the default password-based authentication.
	// Username can be any string. DatabaseName should be "admin".
	AuthTypeScramSHA = "SCRAM-SHA"

	// AuthTypeAWSIAMUser authenticates using AWS IAM user credentials.
	// Username must be an AWS ARN. DatabaseName should be "$external".
	AuthTypeAWSIAMUser = "AWS_IAM_USER"

	// AuthTypeX509Customer authenticates using customer-managed X.509 certificates.
	// Username must be an RFC 2253 Distinguished Name. DatabaseName should be "$external".
	AuthTypeX509Customer = "X509_CUSTOMER"

	// AuthTypeX509Managed authenticates using MongoDB Atlas-managed X.509 certificates.
	// Username must be an RFC 2253 Distinguished Name. DatabaseName should be "$external".
	AuthTypeX509Managed = "X509_MANAGED"

	// AuthTypeLDAPUser authenticates using LDAP user credentials.
	// Username must be an RFC 2253 Distinguished Name. DatabaseName should be "$external".
	AuthTypeLDAPUser = "LDAP_USER"

	// AuthTypeOIDCWorkload authenticates using OIDC workload identity.
	// Username format: Atlas OIDC IdP ID followed by '/' and the IdP user identifier.
	// DatabaseName should be "$external".
	AuthTypeOIDCWorkload = "OIDC_WORKLOAD"
)

// databaseNameAdmin is used for SCRAM-SHA authentication.
const databaseNameAdmin = "admin"

// databaseNameExternal is used for external authentication methods (AWS IAM, x.509, LDAP, OIDC Workload).
const databaseNameExternal = "$external"

// dbTypeUser is the value used for user-based authentication in AWS IAM, LDAP, and OIDC.
const dbTypeUser = "USER"

// getDatabaseNameForAuthType returns the appropriate database name for the given authentication type.
func getDatabaseNameForAuthType(authType string) string {
	switch authType {
	case AuthTypeScramSHA:
		return databaseNameAdmin
	case AuthTypeAWSIAMUser, AuthTypeX509Customer, AuthTypeX509Managed, AuthTypeLDAPUser, AuthTypeOIDCWorkload:
		return databaseNameExternal
	default:
		// Default to admin for backwards compatibility (SCRAM-SHA)
		return databaseNameAdmin
	}
}

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

	users, resp, err := o.client.MongoDBCloudUsersApi.ListOrganizationUsers(ctx, parentResourceID.GetResource()).PageNum(page).ItemsPerPage(resourcePageSize).Execute() //nolint:bodyclose // The SDK handles closing the response body
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to list users: %w", parseToUHttpError(resp, err))
	}

	if users.Results == nil {
		return nil, "", nil, nil
	}

	var resources []*v2.Resource
	for _, user := range *users.Results {
		resource, err := newUserResource(ctx, parentResourceID, &user)
		if err != nil {
			return nil, "", nil, fmt.Errorf("failed to create user resource: %w", err)
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

func (o *userBuilder) CreateAccount(ctx context.Context, accountInfo *v2.AccountInfo, credentialOptions *v2.LocalCredentialOptions) (connectorbuilder.CreateAccountResponse, []*v2.PlaintextData, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	var err error

	profile := accountInfo.Profile.AsMap()

	orgId, ok := profile["organizationId"].(string)
	if orgId == "" || !ok {
		return nil, nil, annotations.Annotations{}, uhttp.WrapErrors(codes.InvalidArgument, "mongo-db-connector: organizationId is required", fmt.Errorf("organizationId field is missing or empty"))
	}

	groupId, ok := profile["groupId"].(string)
	if groupId == "" || !ok {
		return nil, nil, annotations.Annotations{}, uhttp.WrapErrors(codes.InvalidArgument, "mongo-db-connector: groupId is required", fmt.Errorf("groupId field is missing or empty"))
	}

	username, ok := profile["username"].(string)
	if username == "" || !ok {
		return nil, nil, annotations.Annotations{}, uhttp.WrapErrors(codes.InvalidArgument, "mongo-db-connector: username is required", fmt.Errorf("username field is missing or empty"))
	}

	var userId string

	var user atlasUserResponse
	if o.createInviteKey {
		email, ok := profile["email"].(string)
		if email == "" || !ok {
			return nil, nil, annotations.Annotations{}, uhttp.WrapErrors(codes.InvalidArgument, "mongo-db-connector: email is required", fmt.Errorf("email field is missing or empty"))
		}

		l.Info("creating organization user")
		userId, err = o.createUserIfNotExists(ctx, orgId, email, profile)
		if err != nil {
			l.Error(
				"failed to create organization invitation",
				zap.Error(err),
			)
			return nil, nil, nil, err
		}

		var resp *http.Response
		if userId != "" {
			user, resp, err = o.client.MongoDBCloudUsersApi.GetOrganizationUser(ctx, orgId, userId).Execute() //nolint:bodyclose // The SDK handles closing the response body
			if err != nil {
				return nil, nil, nil, fmt.Errorf("failed to get user by id: %w", parseToUHttpError(resp, err))
			}
		} else {
			var result *admin.PaginatedOrgUser
			result, resp, err = o.client.MongoDBCloudUsersApi.ListOrganizationUsers(ctx, orgId).Username(email).Execute() //nolint:bodyclose // The SDK handles closing the response body
			if err != nil {
				if atlasErr, ok := admin.AsError(err); ok {
					switch atlasErr.ErrorCode {
					case "CANNOT_ADD_PENDING_USER":
						return nil, nil, nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("the user '%s' has a pending invite in the organization", email))
					case "NOT_USER_ADMIN":
						return nil, nil, nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("the user '%s' is not in the organization either received an invite, enable createInviteKey to create the invite", email))
					}
				}

				return nil, nil, nil, fmt.Errorf("failed to get user by username: %w", parseToUHttpError(resp, err))
			}

			if result.Results == nil {
				return nil, nil, nil, uhttp.WrapErrors(codes.NotFound, "mongo-db-connector: user not found", fmt.Errorf("user '%s' not found, results is nil", email))
			}

			for _, userResponse := range *result.Results {
				if userResponse.GetUsername() == email {
					user = &userResponse
					break
				}
			}

			if user == nil {
				l.Info("user was not found by username, creating database user instead", zap.String("email", email))
			}
		}
	}

	l.Info("creating database user", zap.String("userId", userId))

	// Get the authentication type from the profile, default to SCRAM-SHA
	authType, ok := profile["authType"].(string)
	if authType == "" || !ok {
		authType = AuthTypeScramSHA
	}

	// Determine the database name based on authentication type
	databaseName := getDatabaseNameForAuthType(authType)

	// Build the CloudDatabaseUser with the appropriate auth type fields
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
	var plaintextData []*v2.PlaintextData

	// Set the appropriate authentication type fields
	// See: https://www.mongodb.com/docs/api/doc/atlas-admin-api-v2/operation/operation-createdatabaseuser
	switch authType {
	case AuthTypeScramSHA:
		// SCRAM-SHA requires a password
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
		// AWS IAM User authentication - username must be an AWS ARN
		dbUserRequest.AwsIAMType = strPtr(dbTypeUser)

	case AuthTypeX509Customer:
		// Customer-managed X.509 certificate - username must be RFC 2253 Distinguished Name
		dbUserRequest.X509Type = strPtr("CUSTOMER")

	case AuthTypeX509Managed:
		// MongoDB Atlas-managed X.509 certificate - username must be RFC 2253 Distinguished Name
		dbUserRequest.X509Type = strPtr("MANAGED")

	case AuthTypeLDAPUser:
		// LDAP User authentication - username must be RFC 2253 Distinguished Name
		dbUserRequest.LdapAuthType = strPtr(dbTypeUser)

	case AuthTypeOIDCWorkload:
		// OIDC Workload authentication - username format: <Atlas OIDC IdP ID>/<IdP user identifier>
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
		l.Error(
			"failed to create database user",
			zap.Error(err),
		)
		return nil, nil, nil, fmt.Errorf("failed to create database user: %w", parseToUHttpError(resp, err))
	}

	var resource *v2.Resource
	if user != nil {
		resource, err = newUserResource(
			ctx,
			&v2.ResourceId{
				ResourceType: organizationResourceType.Id,
				Resource:     orgId,
			},
			user,
		)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to create user resource: %w", err)
		}
	} else {
		resource, err = newDatabaseUserResource(
			ctx,
			&v2.ResourceId{
				ResourceType: projectResourceType.Id,
				Resource:     groupId,
			},
			*dbUser,
		)
	}

	response := &v2.CreateAccountResponse_SuccessResult{
		IsCreateAccountResult: true,
		Resource:              resource,
	}

	return response, plaintextData, nil, err
}

func (o *userBuilder) Delete(ctx context.Context, resourceId *v2.ResourceId, parentResourceID *v2.ResourceId) (annotations.Annotations, error) {
	userId := resourceId.Resource

	if parentResourceID == nil {
		return nil, fmt.Errorf("parent resource id is required: parent resource id is empty")
	}

	orgId := parentResourceID.Resource

	resp, err := o.client.MongoDBCloudUsersApi.RemoveOrganizationUser(ctx, orgId, userId).Execute() //nolint:bodyclose // The SDK handles closing the response body
	if err != nil {
		return nil, fmt.Errorf("failed to remove organization user: %w", parseToUHttpError(resp, err))
	}

	return nil, nil
}

func parseStrList(strFrom any, defaultValue []string) *[]string {
	strList, ok := strFrom.([]interface{})
	if !ok {
		return &defaultValue
	}
	finalStrList := make([]string, 0, len(strList))
	for _, v := range strList {
		strItem, ok := v.(string)
		if !ok {
			return &defaultValue
		}
		finalStrList = append(finalStrList, strItem)
	}
	if len(finalStrList) == 0 {
		return &defaultValue
	}
	return &finalStrList
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
		return "", fmt.Errorf("failed to create organization invitation: %w", parseToUHttpError(httpResponse, err))
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
