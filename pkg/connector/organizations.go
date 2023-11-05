package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"go.mongodb.org/atlas-sdk/v20231001002/admin"
)

type organizationBuilder struct {
	resourceType *v2.ResourceType
	client       *admin.APIClient
}

func (o *organizationBuilder) ResourceType(context context.Context) *v2.ResourceType {
	return organizationResourceType
}

func newOrganizationResource(organization admin.AtlasOrganization) (*v2.Resource, error) {
	organizationId := *organization.Id

	resource, err := rs.NewResource(
		organization.Name,
		organizationResourceType,
		organizationId,
		rs.WithAnnotation(
			&v2.ChildResourceType{ResourceTypeId: userResourceType.Id},
			&v2.ChildResourceType{ResourceTypeId: teamResourceType.Id},
			&v2.ChildResourceType{ResourceTypeId: projectResourceType.Id},
		),
	)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

func (o *organizationBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	bag, page, err := parsePageToken(pToken.Token, &v2.ResourceId{ResourceType: o.resourceType.Id})
	if err != nil {
		return nil, "", nil, err
	}

	organizations, _, err := o.client.OrganizationsApi.ListOrganizations(ctx).PageNum(page).ItemsPerPage(resourcePageSize).IncludeCount(true).Execute()
	if err != nil {
		return nil, "", nil, wrapError(err, "failed to list organizations")
	}

	var resources []*v2.Resource
	for _, organization := range organizations.Results {
		resource, err := newOrganizationResource(organization)
		if err != nil {
			return nil, "", nil, wrapError(err, "failed to create organization resource")
		}

		resources = append(resources, resource)
	}

	if isLastPage(*organizations.TotalCount, resourcePageSize) {
		return resources, "", nil, nil
	}

	nextPage, err := getPageTokenFromPage(bag, page+1)
	if err != nil {
		return nil, "", nil, err
	}

	return resources, nextPage, nil, nil
}

func (o *organizationBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var rv []*v2.Entitlement

	assigmentOptions := []entitlement.EntitlementOption{
		entitlement.WithGrantableTo(userResourceType),
		entitlement.WithDescription(fmt.Sprintf("Member of %s organization", resource.DisplayName)),
		entitlement.WithDisplayName(fmt.Sprintf("%s organization %s", resource.DisplayName, memberEntitlement)),
	}
	ent := entitlement.NewAssignmentEntitlement(resource, memberEntitlement, assigmentOptions...)
	rv = append(rv, ent)

	assigmentOptions = []entitlement.EntitlementOption{
		entitlement.WithGrantableTo(teamResourceType),
		entitlement.WithDescription(fmt.Sprintf("Member of %s organization", resource.DisplayName)),
		entitlement.WithDisplayName(fmt.Sprintf("%s organization %s", resource.DisplayName, memberEntitlement)),
	}
	ent = entitlement.NewAssignmentEntitlement(resource, memberEntitlement, assigmentOptions...)
	rv = append(rv, ent)

	assigmentOptions = []entitlement.EntitlementOption{
		entitlement.WithGrantableTo(projectResourceType),
		entitlement.WithDescription(fmt.Sprintf("Part of %s organization", resource.DisplayName)),
		entitlement.WithDisplayName(fmt.Sprintf("%s organization %s", resource.DisplayName, partEntitlement)),
	}
	ent = entitlement.NewAssignmentEntitlement(resource, partEntitlement, assigmentOptions...)
	rv = append(rv, ent)

	return rv, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *organizationBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	bag, page, err := parsePageToken(pToken.Token, &v2.ResourceId{ResourceType: teamResourceType.Id}, &v2.ResourceId{ResourceType: projectResourceType.Id})
	if err != nil {
		return nil, "", nil, err
	}

	var rv []*v2.Grant
	var count int
	switch bag.Current().ResourceTypeID {
	case teamResourceType.Id:
		grants, c, err := o.GrantTeams(ctx, resource, page)
		if err != nil {
			return nil, "", nil, err
		}
		count = c
		rv = append(rv, grants...)

	case projectResourceType.Id:
		grants, c, err := o.GrantProjects(ctx, resource, page)
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

func (o *organizationBuilder) GrantTeams(ctx context.Context, resource *v2.Resource, page int) ([]*v2.Grant, int, error) {
	teams, _, err := o.client.TeamsApi.ListOrganizationTeams(ctx, resource.Id.Resource).PageNum(page).ItemsPerPage(resourcePageSize).IncludeCount(true).Execute()
	if err != nil {
		return nil, 0, wrapError(err, "failed to list teams")
	}

	var rv []*v2.Grant
	for _, team := range teams.Results {
		teamResource, err := newTeamResource(ctx, resource.ParentResourceId, team)
		if err != nil {
			return nil, *teams.TotalCount, wrapError(err, "failed to create team grant")
		}

		g := grant.NewGrant(
			resource,
			memberEntitlement,
			teamResource.Id,
			grant.WithAnnotation(
				&v2.GrantExpandable{
					EntitlementIds:  []string{fmt.Sprintf("team:%s:%s", teamResource.Id.Resource, memberEntitlement)},
					Shallow:         true,
					ResourceTypeIds: []string{userResourceType.Id},
				},
			),
		)

		rv = append(rv, g)
	}

	return rv, *teams.TotalCount, nil
}

func (o *organizationBuilder) GrantProjects(ctx context.Context, resource *v2.Resource, page int) ([]*v2.Grant, int, error) {
	projects, _, err := o.client.ProjectsApi.ListProjects(ctx).PageNum(page).ItemsPerPage(resourcePageSize).IncludeCount(true).Execute()
	if err != nil {
		return nil, 0, wrapError(err, "failed to list projects")
	}

	var rv []*v2.Grant
	for _, project := range projects.Results {
		if project.OrgId != resource.Id.Resource {
			continue
		}

		projectResource, err := newProjectResource(ctx, resource.ParentResourceId, project)
		if err != nil {
			return nil, *projects.TotalCount, wrapError(err, "failed to create project grant")
		}

		g := grant.NewGrant(
			resource,
			partEntitlement,
			projectResource.Id,
			// TODO: DB users
			// grant.WithAnnotation(
			// 	&v2.GrantExpandable{
			// 		EntitlementIds:  []string{fmt.Sprintf("project:%s:%s", projectResource.Id.Resource, partEntitlement)},
			// 		Shallow:         true,
			// 		ResourceTypeIds: []string{userResourceType.Id},
			// 	},
			// ),
		)

		rv = append(rv, g)
	}

	return rv, *projects.TotalCount, nil
}

func newOrganizationBuilder(client *admin.APIClient) *organizationBuilder {
	return &organizationBuilder{
		resourceType: organizationResourceType,
		client:       client,
	}
}
