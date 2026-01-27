package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
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

	projectTraits := []rs.GroupTraitOption{
		rs.WithGroupProfile(profile),
	}

	resource, err := rs.NewGroupResource(
		project.Name,
		projectResourceType,
		projectId,
		projectTraits,
		rs.WithParentResourceID(organizationId),
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

func (p *projectBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	if parentResourceID == nil {
		return nil, "", nil, nil
	}

	bag, page, err := parsePageToken(pToken.Token, &v2.ResourceId{ResourceType: p.resourceType.Id})
	if err != nil {
		return nil, "", nil, err
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
		return nil, "", nil, fmt.Errorf("failed to list projects: %w", parseToUHttpError(resp, err))
	}

	if projects == nil {
		return nil, "", nil, nil
	}

	var resources []*v2.Resource
	for _, project := range *projects.Results {
		resource, err := newProjectResource(ctx, parentResourceID, project)
		if err != nil {
			return nil, "", nil, fmt.Errorf("failed to create project resource: %w", err)
		}

		resources = append(resources, resource)
	}

	if isLastPage(len(*projects.Results), resourcePageSize) {
		return resources, "", nil, nil
	}

	nextPage, err := getPageTokenFromPage(bag, page+1)
	if err != nil {
		return nil, "", nil, err
	}

	return resources, nextPage, nil, nil
}

// Entitlements always returns an empty slice for users.
func (p *projectBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
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

	return rv, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (p *projectBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	bag, page, err := parsePageToken(pToken.Token, &v2.ResourceId{ResourceType: databaseUserResourceType.Id}, &v2.ResourceId{ResourceType: userResourceType.Id})
	if err != nil {
		return nil, "", nil, err
	}

	var rv []*v2.Grant
	var count int
	switch bag.Current().ResourceTypeID {
	case databaseUserResourceType.Id:
		grants, c, err := p.GrantDatabaseUsers(ctx, resource, page)
		if err != nil {
			return nil, "", nil, err
		}
		count = c
		rv = append(rv, grants...)
	case userResourceType.Id:
		grants, c, err := p.GrantUsers(ctx, resource, page)
		if err != nil {
			return nil, "", nil, err
		}
		count = c
		rv = append(rv, grants...)
	}

	if isLastPage(count, resourcePageSize) {
		nextPage, err := bag.NextToken("")
		if err != nil {
			return nil, "", nil, err
		}

		// Process the next resource type.
		return rv, nextPage, nil, nil
	}

	nextPage, err := getPageTokenFromPage(bag, page+1)
	if err != nil {
		return nil, "", nil, err
	}

	return rv, nextPage, nil, nil
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

	username := trait.GetLogin()
	_, resp, err := p.client.MongoDBCloudUsersApi.AddProjectUser(
		ctx,
		entitlement.Resource.Id.Resource,
		&admin.GroupUserRequest{
			Username: username,
			Roles:    []string{entitlementSlug},
		},
	).Execute() //nolint:bodyclose // The SDK handles closing the response body
	if err != nil {
		err = fmt.Errorf("failed to add user to project: %w", parseToUHttpError(resp, err))

		l.Error(
			"failed to add user to project",
			zap.Error(err),
			zap.String("user_id", principal.Id.Resource),
			zap.String("project_id", entitlement.Resource.Id.Resource),
		)

		return nil, err
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

	resp, err := p.client.MongoDBCloudUsersApi.RemoveProjectUser(
		ctx,
		grant.Entitlement.Resource.Id.Resource,
		grant.Principal.Id.Resource,
	).Execute() //nolint:bodyclose // The SDK handles closing the response body
	if err != nil {
		err = fmt.Errorf("failed to remove user from project: %w", parseToUHttpError(resp, err))

		l.Error(
			"failed to remove user from project",
			zap.Error(err),
			zap.String("user_id", grant.Principal.Id.Resource),
			zap.String("project_id", grant.Entitlement.Resource.Id.Resource),
		)

		return nil, err
	}

	return nil, nil
}
