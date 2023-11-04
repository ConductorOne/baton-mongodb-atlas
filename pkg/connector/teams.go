package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	ent "github.com/conductorone/baton-sdk/pkg/types/entitlement"
	grant "github.com/conductorone/baton-sdk/pkg/types/grant"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"go.mongodb.org/atlas-sdk/v20231001002/admin"
)

type teamBuilder struct {
	resourceType *v2.ResourceType
	client       *admin.APIClient
}

func (o *teamBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return teamResourceType
}

func newTeamResource(ctx context.Context, organizationId *v2.ResourceId, team admin.TeamResponse) (*v2.Resource, error) {
	teamId := *team.Id
	teamName := *team.Name

	profile := map[string]interface{}{
		"team_id": teamId,
		"name":    teamName,
	}

	teamTraits := []rs.GroupTraitOption{
		rs.WithGroupProfile(profile),
	}

	resource, err := rs.NewGroupResource(
		teamName,
		teamResourceType,
		teamId,
		teamTraits,
		rs.WithParentResourceID(organizationId),
	)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

func newTeamBuilder(client *admin.APIClient) *teamBuilder {
	return &teamBuilder{
		resourceType: teamResourceType,
		client:       client,
	}
}

func (o *teamBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	if parentResourceID == nil {
		return nil, "", nil, nil
	}

	bag, page, err := parsePageToken(pToken.Token, &v2.ResourceId{ResourceType: o.resourceType.Id})
	if err != nil {
		return nil, "", nil, err
	}

	teams, _, err := o.client.TeamsApi.ListOrganizationTeams(ctx, parentResourceID.Resource).PageNum(page).ItemsPerPage(resourcePageSize).Execute()
	if err != nil {
		return nil, "", nil, wrapError(err, "failed to list teams")
	}

	var resources []*v2.Resource
	for _, team := range teams.Results {
		resource, err := newTeamResource(ctx, parentResourceID, team)
		if err != nil {
			return nil, "", nil, wrapError(err, "failed to create team resource")
		}

		resources = append(resources, resource)
	}

	if isLastPage(*teams.TotalCount, resourcePageSize) {
		return resources, "", nil, nil
	}

	nextPage, err := getPageTokenFromPage(bag, page+1)
	if err != nil {
		return nil, "", nil, err
	}

	return resources, nextPage, nil, nil
}

// Entitlements always returns an empty slice for users.
func (o *teamBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
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
func (o *teamBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	bag, page, err := parsePageToken(pToken.Token, &v2.ResourceId{ResourceType: o.resourceType.Id})
	if err != nil {
		return nil, "", nil, err
	}

	members, _, err := o.client.TeamsApi.ListTeamUsers(ctx, resource.ParentResourceId.Resource, resource.Id.Resource).PageNum(page).ItemsPerPage(resourcePageSize).Execute()
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
