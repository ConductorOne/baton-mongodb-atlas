package connector

import (
	"context"
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"go.mongodb.org/atlas-sdk/v20250312006/admin"
)

// orgApiKeyDetail is the axis-2 detail string for organization API keys, per RFC §2.8
// (<platform>.<object>.<purpose>, lowercase, dot-delimited).
const orgApiKeyDetail = "mongodb.org_api_key" //nolint:gosec // G101 false positive: axis-2 detail label, not a credential.

type orgApiKeyBuilder struct {
	resourceType *v2.ResourceType
	client       *admin.APIClient
}

func (o *orgApiKeyBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return orgApiKeyResourceType
}

func newOrgApiKeyResource(_ context.Context, organizationId *v2.ResourceId, apiKey admin.ApiKeyUserDetails) (*v2.Resource, error) {
	keyId := apiKey.GetId()

	displayName := apiKey.GetDesc()
	if displayName == "" {
		displayName = apiKey.GetPublicKey()
	}
	if displayName == "" {
		displayName = keyId
	}

	secretTraits := []rs.SecretTraitOption{
		rs.WithSecretType(v2.SecretTrait_CREDENTIAL_TYPE_STATIC_SECRET),
		rs.WithSecretDetail(orgApiKeyDetail),
	}

	resourceOpts := []rs.ResourceOption{
		rs.WithParentResourceID(organizationId),
	}
	if publicKey := apiKey.GetPublicKey(); publicKey != "" {
		resourceOpts = append(resourceOpts, rs.WithDescription(fmt.Sprintf("Organization API key (public key %s)", publicKey)))
	}

	resource, err := rs.NewSecretResource(
		displayName,
		orgApiKeyResourceType,
		keyId,
		secretTraits,
		resourceOpts...,
	)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

func newOrgApiKeyBuilder(client *admin.APIClient) *orgApiKeyBuilder {
	return &orgApiKeyBuilder{
		resourceType: orgApiKeyResourceType,
		client:       client,
	}
}

// List returns the organization's programmatic API keys as secret resources.
// API keys are read-only static-secret credentials, so they carry no entitlements or grants.
func (o *orgApiKeyBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, opts rs.SyncOpAttrs) ([]*v2.Resource, *rs.SyncOpResults, error) {
	if parentResourceID == nil {
		return nil, nil, nil
	}

	bag, page, err := parsePageToken(opts.PageToken.Token, &v2.ResourceId{ResourceType: o.resourceType.Id})
	if err != nil {
		return nil, nil, err
	}

	apiKeys, resp, err := o.client.ProgrammaticAPIKeysApi.ListApiKeys(
		ctx,
		parentResourceID.GetResource(),
	).PageNum(page).ItemsPerPage(resourcePageSize).Execute() //nolint:bodyclose // The SDK handles closing the response body
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list organization API keys: %w", parseToUHttpError(resp, err))
	}

	if apiKeys == nil || apiKeys.Results == nil {
		return nil, nil, nil
	}

	var resources []*v2.Resource
	for _, apiKey := range *apiKeys.Results {
		resource, err := newOrgApiKeyResource(ctx, parentResourceID, apiKey)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create organization API key resource: %w", err)
		}

		resources = append(resources, resource)
	}

	if isLastPage(len(*apiKeys.Results), resourcePageSize) {
		return resources, nil, nil
	}

	nextPage, err := getPageTokenFromPage(bag, page+1)
	if err != nil {
		return nil, nil, err
	}

	return resources, &rs.SyncOpResults{NextPageToken: nextPage}, nil
}

// Entitlements always returns an empty slice; API keys are credentials, not grantable resources.
func (o *orgApiKeyBuilder) Entitlements(_ context.Context, _ *v2.Resource, _ rs.SyncOpAttrs) ([]*v2.Entitlement, *rs.SyncOpResults, error) {
	return nil, nil, nil
}

// Grants always returns an empty slice; API keys are credentials, not grantable resources.
func (o *orgApiKeyBuilder) Grants(_ context.Context, _ *v2.Resource, _ rs.SyncOpAttrs) ([]*v2.Grant, *rs.SyncOpResults, error) {
	return nil, nil, nil
}
