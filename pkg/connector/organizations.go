package connector

import (
	"context"
	"fmt"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"go.mongodb.org/atlas-sdk/v20250312006/admin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	userRolesOrganizationEntitlementMap = map[string]string{
		"ORG_OWNER":             ownerEntitlement,
		"ORG_GROUP_CREATOR":     projectCreatorEntitlement,
		"ORG_BILLING_ADMIN":     billingAdminEntitlement,
		"ORG_BILLING_READ_ONLY": billingViewerEntitlement,
		"ORG_READ_ONLY":         readOnlyEntitlement,
		"ORG_MEMBER":            memberEntitlement,
	}
	userRolesOrganizationEntitlementMapReversed = reverseMap(userRolesOrganizationEntitlementMap)
	organizationUserEntitlements                = []string{
		ownerEntitlement,
		projectCreatorEntitlement,
		billingAdminEntitlement,
		billingViewerEntitlement,
		readOnlyEntitlement,
		memberEntitlement,
	}
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

	organizations, _, err := o.client.OrganizationsApi.ListOrganizations(ctx).PageNum(page).ItemsPerPage(resourcePageSize).IncludeCount(true).Execute() //nolint:bodyclose // The SDK handles closing the response body
	if err != nil {
		return nil, "", nil, wrapError(err, "failed to list organizations")
	}

	if organizations.Results == nil {
		return nil, "", nil, nil
	}

	var resources []*v2.Resource
	for _, organization := range *organizations.Results {
		resource, err := newOrganizationResource(organization)
		if err != nil {
			return nil, "", nil, wrapError(err, "failed to create organization resource")
		}

		resources = append(resources, resource)
	}

	if isLastPage(len(*organizations.Results), resourcePageSize) {
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

	for _, e := range organizationUserEntitlements {
		assigmentOptions := []entitlement.EntitlementOption{
			entitlement.WithGrantableTo(userResourceType),
			entitlement.WithDescription(fmt.Sprintf("Member of %s organization", resource.DisplayName)),
			entitlement.WithDisplayName(fmt.Sprintf("%s organization %s", resource.DisplayName, memberEntitlement)),
		}
		ent := entitlement.NewPermissionEntitlement(resource, e, assigmentOptions...)
		rv = append(rv, ent)
	}

	assigmentOptions := []entitlement.EntitlementOption{
		entitlement.WithGrantableTo(teamResourceType),
		entitlement.WithDescription(fmt.Sprintf("Member of %s organization", resource.DisplayName)),
		entitlement.WithDisplayName(fmt.Sprintf("%s organization %s", resource.DisplayName, memberEntitlement)),
		entitlement.WithAnnotation(&v2.EntitlementImmutable{}),
	}
	ent := entitlement.NewAssignmentEntitlement(resource, memberEntitlement, assigmentOptions...)
	rv = append(rv, ent)

	assigmentOptions = []entitlement.EntitlementOption{
		entitlement.WithGrantableTo(databaseUserResourceType),
		entitlement.WithDescription(fmt.Sprintf("Member of %s organization", resource.DisplayName)),
		entitlement.WithDisplayName(fmt.Sprintf("%s organization %s", resource.DisplayName, partEntitlement)),
		entitlement.WithAnnotation(&v2.EntitlementImmutable{}),
	}
	ent = entitlement.NewAssignmentEntitlement(resource, partEntitlement, assigmentOptions...)
	rv = append(rv, ent)

	assigmentOptions = []entitlement.EntitlementOption{
		entitlement.WithGrantableTo(projectResourceType),
		entitlement.WithDescription(fmt.Sprintf("Part of %s organization", resource.DisplayName)),
		entitlement.WithDisplayName(fmt.Sprintf("%s organization %s", resource.DisplayName, partEntitlement)),
		entitlement.WithAnnotation(&v2.EntitlementImmutable{}),
	}
	ent = entitlement.NewAssignmentEntitlement(resource, partEntitlement, assigmentOptions...)
	rv = append(rv, ent)

	return rv, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *organizationBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	bag, page, err := parsePageToken(
		pToken.Token,
		&v2.ResourceId{ResourceType: teamResourceType.Id},
		&v2.ResourceId{ResourceType: projectResourceType.Id},
		&v2.ResourceId{ResourceType: userResourceType.Id},
	)
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
	case userResourceType.Id:
		grants, c, err := o.GrantUsers(ctx, resource, page)
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

func (o *organizationBuilder) Grant(ctx context.Context, resource *v2.Resource, entitlement *v2.Entitlement) ([]*v2.Grant, annotations.Annotations, error) {
	if resource.Id.ResourceType != userResourceType.Id {
		return nil, nil, fmt.Errorf("baton-mongodb-atlas: cannot grant to resource type %s", resource.Id.ResourceType)
	}

	orgId := entitlement.Resource.Id.Resource
	userId := resource.Id.Resource
	userTrait, err := rs.GetUserTrait(resource)
	if err != nil {
		return nil, nil, status.Errorf(codes.FailedPrecondition, "baton-mongodb-atlas: resource %s does not have a user trait", userId)
	}

	emails := userTrait.GetEmails()
	var primaryEmail string
	for _, email := range emails {
		if email.IsPrimary {
			primaryEmail = email.GetAddress()
			break
		}
	}

	if primaryEmail == "" {
		return nil, nil, status.Errorf(codes.FailedPrecondition, "baton-mongodb-atlas: no primary email found for user %s", userId)
	}

	role := userRolesOrganizationEntitlementMapReversed[entitlement.Slug]

	response, _, err := o.client.MongoDBCloudUsersApi.GetOrganizationUser(ctx, orgId, userId).Execute() //nolint:bodyclose // The SDK handles closing the response body
	if err != nil {
		return nil, nil, err
	}

	var newRoles []string

	if response.Roles.OrgRoles != nil {
		for _, s := range *response.Roles.OrgRoles {
			if s == role {
				return nil, annotations.New(&v2.GrantAlreadyExists{}), nil
			}
		}

		newRoles = append(newRoles, *response.Roles.OrgRoles...)
	}

	newRoles = append(newRoles, role)

	_, _, err = o.client.MongoDBCloudUsersApi.UpdateOrganizationUser(
		ctx,
		orgId,
		userId,
		&admin.OrgUserUpdateRequest{
			Roles: &admin.OrgUserRolesRequest{
				OrgRoles:             newRoles,
				GroupRoleAssignments: response.Roles.GroupRoleAssignments,
			},
			TeamIds: response.TeamIds,
		},
	).Execute() //nolint:bodyclose // The SDK handles closing the response body

	if err != nil {
		return nil, nil, err
	}

	newGrant := grant.NewGrant(resource, entitlement.Slug, &v2.ResourceId{
		ResourceType: userResourceType.Id,
		Resource:     userId,
	})

	return []*v2.Grant{newGrant}, nil, nil
}

func (o *organizationBuilder) Revoke(ctx context.Context, grant *v2.Grant) (annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	if grant.Principal.Id.ResourceType != userResourceType.Id {
		return nil, fmt.Errorf("baton-mongodb-atlas: cannot revoke to resource type %s", grant.Principal.Id.ResourceType)
	}

	orgId := grant.Entitlement.Resource.Id.Resource
	userId := grant.Principal.Id.Resource
	role := userRolesOrganizationEntitlementMapReversed[grant.Entitlement.Slug]

	response, _, err := o.client.MongoDBCloudUsersApi.GetOrganizationUser(ctx, orgId, userId).Execute() //nolint:bodyclose // The SDK handles closing the response body
	if err != nil {
		return nil, err
	}

	if response.Roles.OrgRoles == nil {
		return annotations.New(&v2.GrantAlreadyRevoked{}), nil
	}

	var newRoles []string

	for _, s := range *response.Roles.OrgRoles {
		if s == role {
			continue
		}

		newRoles = append(newRoles, s)
	}

	if len(newRoles) == len(*response.Roles.OrgRoles) {
		return annotations.New(&v2.GrantAlreadyRevoked{}), nil
	}

	if len(newRoles) == 0 {
		l.Info(
			"baton-mongodb-atlas: no roles to assign to user, removing from org",
			zap.String("orgId", orgId),
			zap.String("userId", userId),
		)

		_, err = o.client.MongoDBCloudUsersApi.RemoveOrganizationUser(ctx, orgId, userId).Execute() //nolint:bodyclose // The SDK handles closing the response body
	} else {
		_, _, err = o.client.MongoDBCloudUsersApi.UpdateOrganizationUser(
			ctx,
			orgId,
			userId,
			&admin.OrgUserUpdateRequest{
				Roles: &admin.OrgUserRolesRequest{
					OrgRoles:             newRoles,
					GroupRoleAssignments: response.Roles.GroupRoleAssignments,
				},
				TeamIds: response.TeamIds,
			},
		).Execute() //nolint:bodyclose // The SDK handles closing the response body
	}

	if err != nil {
		return nil, wrapError(err, "failed to remove user role for organization user")
	}

	return nil, nil
}

func (o *organizationBuilder) GrantTeams(ctx context.Context, orgResource *v2.Resource, page int) ([]*v2.Grant, int, error) {
	teams, _, err := o.client.TeamsApi.ListOrganizationTeams(ctx, orgResource.Id.Resource).PageNum(page).ItemsPerPage(resourcePageSize).IncludeCount(true).Execute() //nolint:bodyclose // The SDK handles closing the response body
	if err != nil {
		return nil, 0, wrapError(err, "failed to list teams")
	}

	if teams.Results == nil {
		return nil, 0, err
	}

	var rv []*v2.Grant
	for _, team := range *teams.Results {
		teamResource, err := newTeamResource(ctx, orgResource.Id, team)
		if err != nil {
			return nil, *teams.TotalCount, wrapError(err, "failed to create team grant")
		}

		g := grant.NewGrant(
			orgResource,
			memberEntitlement,
			teamResource.Id,
			grant.WithAnnotation(&v2.GrantImmutable{}),
		)

		rv = append(rv, g)
	}

	return rv, len(*teams.Results), nil
}

func (o *organizationBuilder) GrantProjects(ctx context.Context, orgResource *v2.Resource, page int) ([]*v2.Grant, int, error) {
	projects, _, err := o.client.ProjectsApi.ListProjects(ctx).PageNum(page).ItemsPerPage(resourcePageSize).IncludeCount(true).Execute() //nolint:bodyclose // The SDK handles closing the response body
	if err != nil {
		return nil, 0, wrapError(err, "failed to list projects")
	}

	if projects.Results == nil {
		return nil, 0, err
	}

	var rv []*v2.Grant
	for _, project := range *projects.Results {
		if project.OrgId != orgResource.Id.Resource {
			continue
		}

		projectResource, err := newProjectResource(ctx, orgResource.Id, project)
		if err != nil {
			return nil, *projects.TotalCount, wrapError(err, "failed to create project grant")
		}

		g := grant.NewGrant(
			orgResource,
			partEntitlement,
			projectResource.Id,
			grant.WithAnnotation(
				&v2.GrantExpandable{
					EntitlementIds:  []string{fmt.Sprintf("project:%s:%s", projectResource.Id.Resource, memberEntitlement)},
					Shallow:         true,
					ResourceTypeIds: []string{databaseUserResourceType.Id},
				},
				&v2.GrantImmutable{},
			),
		)

		rv = append(rv, g)
	}

	return rv, len(*projects.Results), nil
}

func (o *organizationBuilder) GrantUsers(ctx context.Context, orgResource *v2.Resource, page int) ([]*v2.Grant, int, error) {
	users, _, err := o.client.MongoDBCloudUsersApi.ListOrganizationUsers(ctx, orgResource.Id.Resource).PageNum(page).ItemsPerPage(resourcePageSize).IncludeCount(true).Execute() //nolint:bodyclose // The SDK handles closing the response body
	if err != nil {
		return nil, 0, wrapError(err, "failed to list organization users")
	}

	if users.Results == nil {
		return nil, 0, err
	}

	var rv []*v2.Grant
	for _, user := range *users.Results {
		userResource, err := newUserResource(ctx, orgResource.Id, &user)
		if err != nil {
			return nil, *users.TotalCount, wrapError(err, "failed to create user resource")
		}

		for _, roleName := range *user.Roles.OrgRoles {
			if entitlementTarget, ok := userRolesOrganizationEntitlementMap[roleName]; ok {
				rv = append(
					rv,
					grant.NewGrant(
						orgResource,
						entitlementTarget,
						userResource.Id,
					),
				)
			}
		}
	}

	return rv, len(*users.Results), nil
}

func newOrganizationBuilder(client *admin.APIClient) *organizationBuilder {
	return &organizationBuilder{
		resourceType: organizationResourceType,
		client:       client,
	}
}
