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
	"go.mongodb.org/atlas-sdk/v20231001002/admin"
)

var (
	userRolesProjectEntitlementMap = map[string]string{
		"GROUP_OWNER":                  ownerEntitlement,
		"GROUP_CLUSTER_MANAGER":        clusterManagerEntitlement,
		"GROUP_DATA_ACCESS_ADMIN":      dataAccessAdminEntitlement,
		"GROUP_DATA_ACCESS_READ_WRITE": dataAccessReadAndWriteEntitlement,
		"GROUP_DATA_ACCESS_READ_ONLY":  dataAccessReadOnlyEntitlement,
		"GROUP_READ_ONLY":              readOnlyEntitlement,
		"GROUP_SEARCH_INDEX_EDITOR":    searchIndexEditorEntitlement,
	}
	projectEntitlementsUserRolesMap = reverseMap(userRolesProjectEntitlementMap)
	projectUserEntitlements         = []string{
		ownerEntitlement,
		clusterManagerEntitlement,
		dataAccessAdminEntitlement,
		dataAccessReadAndWriteEntitlement,
		dataAccessReadOnlyEntitlement,
		readOnlyEntitlement,
		searchIndexEditorEntitlement,
		memberEntitlement,
	}
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

	projects, _, err := p.client.ProjectsApi.ListProjects(ctx).PageNum(page).ItemsPerPage(resourcePageSize).IncludeCount(true).Execute()
	if err != nil {
		return nil, "", nil, err
	}

	var resources []*v2.Resource
	for _, project := range projects.Results {
		resource, err := newProjectResource(ctx, parentResourceID, project)
		if err != nil {
			return nil, "", nil, err
		}

		resources = append(resources, resource)
	}

	if isLastPage(*projects.TotalCount, resourcePageSize) {
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

	for _, e := range projectUserEntitlements {
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
	members, _, err := p.client.ProjectsApi.ListProjectUsers(ctx, resource.Id.Resource).PageNum(page).ItemsPerPage(resourcePageSize).IncludeCount(true).Execute()
	if err != nil {
		return nil, 0, wrapError(err, "failed to list project users")
	}

	var rv []*v2.Grant
	for _, member := range members.Results {
		userResource, err := newUserResource(ctx, resource.ParentResourceId, member)
		if err != nil {
			return nil, *members.TotalCount, wrapError(err, "failed to create user resource")
		}

		for _, role := range member.Roles {
			if role.GroupId == nil {
				continue
			}

			roleProjectId := *role.GroupId
			if roleProjectId != resource.Id.Resource {
				continue
			}

			roleName := role.RoleName

			if entitlement, ok := userRolesProjectEntitlementMap[*roleName]; ok {
				rv = append(rv, grant.NewGrant(resource, entitlement, userResource.Id))
			}
		}
	}

	return rv, *members.TotalCount, nil
}

func (p *projectBuilder) GrantDatabaseUsers(ctx context.Context, resource *v2.Resource, page int) ([]*v2.Grant, int, error) {
	members, _, err := p.client.DatabaseUsersApi.ListDatabaseUsers(ctx, resource.Id.Resource).PageNum(page).ItemsPerPage(resourcePageSize).IncludeCount(true).Execute()
	if err != nil {
		return nil, 0, wrapError(err, "failed to list project database users")
	}

	var rv []*v2.Grant
	for _, member := range members.Results {
		userResource, err := newDatabaseUserResource(ctx, resource.ParentResourceId, member)
		if err != nil {
			return nil, *members.TotalCount, wrapError(err, "failed to create database user resource")
		}

		rv = append(rv, grant.NewGrant(resource, memberEntitlement, userResource.Id))
	}

	return rv, *members.TotalCount, nil
}
