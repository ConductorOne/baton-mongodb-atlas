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

	assigmentOptions := []ent.EntitlementOption{
		ent.WithGrantableTo(userResourceType),
		ent.WithDescription(fmt.Sprintf("Member of %s team", resource.DisplayName)),
		ent.WithDisplayName(fmt.Sprintf("%s team %s", resource.DisplayName, memberEntitlement)),
	}

	entitlement := ent.NewAssignmentEntitlement(resource, memberEntitlement, assigmentOptions...)
	rv = append(rv, entitlement)

	return rv, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (p *projectBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	bag, page, err := parsePageToken(pToken.Token, &v2.ResourceId{ResourceType: p.resourceType.Id})
	if err != nil {
		return nil, "", nil, err
	}

	members, _, err := p.client.ProjectsApi.ListProjectUsers(ctx, resource.Id.Resource).PageNum(page).ItemsPerPage(resourcePageSize).IncludeCount(true).Execute()
	if err != nil {
		return nil, "", nil, wrapError(err, "failed to list team members")
	}

	var rv []*v2.Grant
	for _, member := range members.Results {
		userResource, err := newUserResource(ctx, resource.ParentResourceId, member)
		if err != nil {
			return nil, "", nil, wrapError(err, "failed to create user resource")
		}

		rv = append(rv, grant.NewGrant(resource, memberEntitlement, userResource.Id))
	}

	if isLastPage(*members.TotalCount, resourcePageSize) {
		return rv, "", nil, nil
	}

	nextPage, err := getPageTokenFromPage(bag, page+1)
	if err != nil {
		return nil, "", nil, err
	}

	return nil, nextPage, nil, nil
}
