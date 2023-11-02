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

	return rv, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *organizationBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	bag, page, err := parsePageToken(pToken.Token, &v2.ResourceId{ResourceType: teamResourceType.Id})
	if err != nil {
		return nil, "", nil, err
	}

	teams, _, err := o.client.TeamsApi.ListOrganizationTeams(ctx, resource.Id.Resource).PageNum(page).ItemsPerPage(resourcePageSize).IncludeCount(true).Execute()
	if err != nil {
		return nil, "", nil, wrapError(err, "failed to list teams")
	}

	var rv []*v2.Grant
	for _, team := range teams.Results {
		teamResource, err := newTeamResource(ctx, team)
		if err != nil {
			return nil, "", nil, wrapError(err, "failed to create team grant")
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

	if isLastPage(*teams.TotalCount, resourcePageSize) {
		return rv, "", nil, nil
	}

	nextPage, err := getPageTokenFromPage(bag, page+1)
	if err != nil {
		return nil, "", nil, err
	}

	return rv, nextPage, nil, nil
}

func newOrganizationBuilder(client *admin.APIClient) *organizationBuilder {
	return &organizationBuilder{
		resourceType: organizationResourceType,
		client:       client,
	}
}
