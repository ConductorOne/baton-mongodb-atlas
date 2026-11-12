package connector

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	ent "github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.mongodb.org/atlas-sdk/v20250312006/admin"
	"go.uber.org/zap"
)

var (
	// https://www.mongodb.com/docs/api/doc/atlas-admin-api-v2/operation/operation-addprojectuser#operation-addprojectuser-body-application-vnd-atlas-2025-02-19-json-roles
	userRolesProjectEntitlementMap = map[string]string{
		"GROUP_OWNER":                   ownerEntitlement,
		"GROUP_CLUSTER_MANAGER":         clusterManagerEntitlement,
		"GROUP_STREAM_PROCESSING_OWNER": groupStreamProcessingOwnerEntitlement,
		"GROUP_DATA_ACCESS_ADMIN":       dataAccessAdminEntitlement,
		"GROUP_DATA_ACCESS_READ_WRITE":  dataAccessReadAndWriteEntitlement,
		"GROUP_DATA_ACCESS_READ_ONLY":   dataAccessReadOnlyEntitlement,
		"GROUP_READ_ONLY":               readOnlyEntitlement,
		"GROUP_SEARCH_INDEX_EDITOR":     searchIndexEditorEntitlement,
		"GROUP_BACKUP_MANAGER":          groupBackupManagerEntitlement,
		"GROUP_OBSERVABILITY_VIEWER":    groupObservabilityViewerEntitlement,
		"GROUP_DATABASE_ACCESS_ADMIN":   groupDatabaseAccessAdminEntitlement,
	}
	projectEntitlementsUserRolesMap = reverseMap(userRolesProjectEntitlementMap)
)

type projectBuilder struct {
	resourceType *v2.ResourceType
	client       *admin.APIClient
}

func (o *projectBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return projectResourceType
}

func newProjectResource(ctx context.Context, organizationId *v2.ResourceId, project admin.Group) (*v2.Resource, error) {
	projectId := *project.Id

	profile := map[string]interface{}{
		"project_id": projectId,
		"name":       project.Name,
	}

	resource, err := rs.NewGroupResource(
		project.Name,
		projectResourceType,
		projectId,
		nil,
		rs.WithParentResourceID(organizationId),
		rs.WithResourceProfile(profile),
		rs.WithAnnotation(
			&v2.ChildResourceType{ResourceTypeId: databaseUserResourceType.Id},
			&v2.ChildResourceType{ResourceTypeId: mongoClusterResourceType.Id},
		),
	)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

func newProjectBuilder(client *admin.APIClient) *projectBuilder {
	return &projectBuilder{
		resourceType: projectResourceType,
		client:       client,
	}
}

func (p *projectBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, opts rs.SyncOpAttrs) ([]*v2.Resource, *rs.SyncOpResults, error) {
	if parentResourceID == nil {
		return nil, nil, nil
	}

	bag, page, err := parsePageToken(opts.PageToken.Token, &v2.ResourceId{ResourceType: p.resourceType.Id})
	if err != nil {
		return nil, nil, err
	}

	logger := ctxzap.Extract(ctx)
	logger.Debug(
		"fetching a page of projects",
		zap.Int("pageNum", page),
		zap.Int("ItemsPerPage", resourcePageSize),
	)

	projects, resp, err := p.client.
		ProjectsApi.
		ListProjects(ctx).
		PageNum(page).
		ItemsPerPage(resourcePageSize).
		IncludeCount(true).
		Execute() //nolint:bodyclose // The SDK handles closing the response body
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list projects: %w", parseToUHttpError(resp, err))
	}

	if projects == nil {
		return nil, nil, nil
	}

	var resources []*v2.Resource
	for _, project := range *projects.Results {
		resource, err := newProjectResource(ctx, parentResourceID, project)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create project resource: %w", err)
		}

		resources = append(resources, resource)
	}

	if isLastPage(len(*projects.Results), resourcePageSize) {
		return resources, nil, nil
	}

	nextPage, err := getPageTokenFromPage(bag, page+1)
	if err != nil {
		return nil, nil, err
	}

	return resources, &rs.SyncOpResults{NextPageToken: nextPage}, nil
}

// Entitlements always returns an empty slice for users.
func (p *projectBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ rs.SyncOpAttrs) ([]*v2.Entitlement, *rs.SyncOpResults, error) {
	var rv []*v2.Entitlement

	for _, e := range userRolesProjectEntitlementMap {
		assigmentOptions := []ent.EntitlementOption{
			ent.WithGrantableTo(userResourceType),
			ent.WithDescription(fmt.Sprintf("Member of %s team", resource.DisplayName)),
			ent.WithDisplayName(fmt.Sprintf("%s team %s", resource.DisplayName, e)),
		}

		entitlement := ent.NewPermissionEntitlement(resource, e, assigmentOptions...)
		rv = append(rv, entitlement)
	}

	assigmentOptions := []ent.EntitlementOption{
		ent.WithGrantableTo(databaseUserResourceType),
		ent.WithDescription(fmt.Sprintf("Member of %s team", resource.DisplayName)),
		ent.WithDisplayName(fmt.Sprintf("%s team %s", resource.DisplayName, memberEntitlement)),
	}

	entitlement := ent.NewAssignmentEntitlement(resource, memberEntitlement, assigmentOptions...)
	rv = append(rv, entitlement)

	return rv, nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (p *projectBuilder) Grants(ctx context.Context, resource *v2.Resource, opts rs.SyncOpAttrs) ([]*v2.Grant, *rs.SyncOpResults, error) {
	bag, page, err := parsePageToken(opts.PageToken.Token, &v2.ResourceId{ResourceType: databaseUserResourceType.Id}, &v2.ResourceId{ResourceType: userResourceType.Id})
	if err != nil {
		return nil, nil, err
	}

	var rv []*v2.Grant
	var count int
	switch bag.Current().ResourceTypeID {
	case databaseUserResourceType.Id:
		grants, c, err := p.GrantDatabaseUsers(ctx, resource, page)
		if err != nil {
			return nil, nil, err
		}
		count = c
		rv = append(rv, grants...)
	case userResourceType.Id:
		grants, c, err := p.GrantUsers(ctx, resource, page)
		if err != nil {
			return nil, nil, err
		}
		count = c
		rv = append(rv, grants...)
	}

	if isLastPage(count, resourcePageSize) {
		nextPage, err := bag.NextToken("")
		if err != nil {
			return nil, nil, err
		}

		// Process the next resource type.
		return rv, &rs.SyncOpResults{NextPageToken: nextPage}, nil
	}

	nextPage, err := getPageTokenFromPage(bag, page+1)
	if err != nil {
		return nil, nil, err
	}

	return rv, &rs.SyncOpResults{NextPageToken: nextPage}, nil
}

func (p *projectBuilder) GrantUsers(ctx context.Context, resource *v2.Resource, page int) ([]*v2.Grant, int, error) {
	members, resp, err := p.client.MongoDBCloudUsersApi.ListProjectUsers(
		ctx,
		resource.Id.Resource,
	).PageNum(page).ItemsPerPage(resourcePageSize).IncludeCount(true).Execute() //nolint:bodyclose // The SDK handles closing the response body
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list project users: %w", parseToUHttpError(resp, err))
	}

	if members.Results == nil {
		return nil, 0, err
	}

	var rv []*v2.Grant
	for _, member := range *members.Results {
		userResource, err := newUserResource(ctx, resource.ParentResourceId, &member)
		if err != nil {
			return nil, *members.TotalCount, fmt.Errorf("failed to create user resource: %w", err)
		}

		for _, roleName := range member.Roles {
			if entitlement, ok := userRolesProjectEntitlementMap[roleName]; ok {
				rv = append(rv, grant.NewGrant(resource, entitlement, userResource.Id))
			}
		}
	}

	return rv, len(*members.Results), nil
}

func (p *projectBuilder) GrantDatabaseUsers(ctx context.Context, resource *v2.Resource, page int) ([]*v2.Grant, int, error) {
	members, resp, err := p.client.DatabaseUsersApi.ListDatabaseUsers(
		ctx,
		resource.Id.Resource,
	).PageNum(page).ItemsPerPage(resourcePageSize).IncludeCount(true).Execute() //nolint:bodyclose // The SDK handles closing the response body
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list project database users: %w", parseToUHttpError(resp, err))
	}

	if members.Results == nil {
		return nil, 0, err
	}

	var rv []*v2.Grant
	for _, member := range *members.Results {
		userResource, err := newDatabaseUserResource(ctx, resource.ParentResourceId, member)
		if err != nil {
			return nil, *members.TotalCount, fmt.Errorf("failed to create database user resource: %w", err)
		}

		rv = append(rv, grant.NewGrant(resource, memberEntitlement, userResource.Id))
	}

	return rv, len(*members.Results), nil
}

// getProjectUserRoles returns the project roles of a user. Atlas is assumed to answer 404 when the user is not in the project.
func (p *projectBuilder) getProjectUserRoles(ctx context.Context, projectId, userId string) ([]string, bool, error) {
	user, resp, err := p.client.MongoDBCloudUsersApi.GetProjectUser(ctx, projectId, userId).Execute() //nolint:bodyclose // The SDK handles closing the response body
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("baton-mongodb-atlas: failed to get project user: %w", parseToUHttpError(resp, err))
	}
	if user == nil {
		return nil, false, nil
	}

	return user.GetRoles(), true, nil
}

func (p *projectBuilder) addProjectRole(ctx context.Context, projectId, userId, role string) error {
	_, resp, err := p.client.MongoDBCloudUsersApi.AddProjectRole(
		ctx,
		projectId,
		userId,
		admin.NewAddOrRemoveGroupRole(role),
	).Execute() //nolint:bodyclose // The SDK handles closing the response body
	if err != nil {
		return fmt.Errorf("baton-mongodb-atlas: failed to add project role: %w", parseToUHttpError(resp, err))
	}

	return nil
}

func (p *projectBuilder) Grant(ctx context.Context, principal *v2.Resource, entitlement *v2.Entitlement) (annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	if principal.Id.ResourceType != userResourceType.Id &&
		principal.Id.ResourceType != databaseUserResourceType.Id {
		err := fmt.Errorf("only users can be granted to projects: expected %s or %s, got %s", userResourceType.Id, databaseUserResourceType.Id, principal.Id.ResourceType)

		l.Warn(
			"mongodb connector: only users can be granted to projects",
			zap.Error(err),
			zap.String("principal_id", principal.Id.Resource),
			zap.String("principal_type", principal.Id.ResourceType),
		)

		return nil, err
	}

	trait, err := rs.GetUserTrait(principal)
	if err != nil {
		return nil, fmt.Errorf("failed to get user trait: %w", err)
	}

	var entitlementSlug string
	if slug, ok := projectEntitlementsUserRolesMap[entitlement.Slug]; !ok {
		err := fmt.Errorf("unknown entitlement: entitlement %s is not recognized", entitlement.Slug)

		l.Warn(
			"mongodb connector: unknown entitlement",
			zap.Error(err),
			zap.String("entitlement_slug", entitlement.Slug),
		)

		return nil, err
	} else {
		entitlementSlug = slug
	}

	projectId := entitlement.Resource.Id.Resource
	userId := principal.Id.Resource

	roles, found, err := p.getProjectUserRoles(ctx, projectId, userId)
	if err != nil {
		return nil, err
	}

	if found {
		if slices.Contains(roles, entitlementSlug) {
			return annotations.New(&v2.GrantAlreadyExists{}), nil
		}
		if err := p.addProjectRole(ctx, projectId, userId, entitlementSlug); err != nil {
			return nil, err
		}
		return nil, nil
	}

	_, resp, err := p.client.MongoDBCloudUsersApi.AddProjectUser(
		ctx,
		projectId,
		&admin.GroupUserRequest{
			Username: trait.GetLogin(),
			Roles:    []string{entitlementSlug},
		},
	).Execute() //nolint:bodyclose // The SDK handles closing the response body
	if err != nil {
		// The user may have joined the project after the lookup; Atlas reports this as USER_ALREADY_IN_GROUP.
		if resp != nil &&
			(resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusConflict) &&
			admin.IsErrorCode(err, "USER_ALREADY_IN_GROUP") {
			l.Debug(
				"user already in project, adding role instead",
				zap.String("user_id", userId),
				zap.String("project_id", projectId),
			)
			if err := p.addProjectRole(ctx, projectId, userId, entitlementSlug); err != nil {
				return nil, err
			}
			return nil, nil
		}

		return nil, fmt.Errorf("baton-mongodb-atlas: failed to add user to project: %w", parseToUHttpError(resp, err))
	}

	return nil, nil
}

func (p *projectBuilder) Revoke(ctx context.Context, grant *v2.Grant) (annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	if grant.Principal.Id.ResourceType != userResourceType.Id {
		err := fmt.Errorf("only users can be revoked from projects: expected %s, got %s", userResourceType.Id, grant.Principal.Id.ResourceType)

		l.Warn(
			"mongodb connector: only users can be revoked from projects",
			zap.Error(err),
			zap.String("principal_id", grant.Principal.Id.Resource),
			zap.String("principal_type", grant.Principal.Id.ResourceType),
		)

		return nil, err
	}

	role, ok := projectEntitlementsUserRolesMap[grant.Entitlement.Slug]
	if !ok {
		return nil, fmt.Errorf("baton-mongodb-atlas: unknown entitlement: entitlement %s is not recognized", grant.Entitlement.Slug)
	}

	projectId := grant.Entitlement.Resource.Id.Resource
	userId := grant.Principal.Id.Resource

	roles, found, err := p.getProjectUserRoles(ctx, projectId, userId)
	if err != nil {
		return nil, err
	}

	if !found || !slices.Contains(roles, role) {
		return annotations.New(&v2.GrantAlreadyRevoked{}), nil
	}

	// Atlas requires at least one project role, so removing the last one means removing the user from the project.
	if len(roles) == 1 {
		resp, err := p.client.MongoDBCloudUsersApi.RemoveProjectUser(ctx, projectId, userId).Execute() //nolint:bodyclose // The SDK handles closing the response body
		if err != nil {
			if resp != nil && resp.StatusCode == http.StatusNotFound {
				return annotations.New(&v2.GrantAlreadyRevoked{}), nil
			}
			return nil, fmt.Errorf("baton-mongodb-atlas: failed to remove user from project: %w", parseToUHttpError(resp, err))
		}
		return nil, nil
	}

	_, resp, err := p.client.MongoDBCloudUsersApi.RemoveProjectRole(
		ctx,
		projectId,
		userId,
		admin.NewAddOrRemoveGroupRole(role),
	).Execute() //nolint:bodyclose // The SDK handles closing the response body
	if err != nil {
		return nil, fmt.Errorf("baton-mongodb-atlas: failed to remove project role: %w", parseToUHttpError(resp, err))
	}

	return nil, nil
}
