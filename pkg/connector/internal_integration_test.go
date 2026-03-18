package connector

import (
	"context"
	"os"
	"testing"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/atlas-sdk/v20250312006/admin"
)

var (
	publicKey  = os.Getenv("BATON_PUBLIC_KEY")
	privateKey = os.Getenv("BATON_PRIVATE_KEY")
)

func TestDatabaseUserBuilderList(t *testing.T) {
	if publicKey == "" && privateKey == "" {
		t.Skip()
	}

	ctx := context.Background()
	parentResourceID := &v2.ResourceId{
		ResourceType: "project",
		Resource:     "654be90ed21dc34308aba9bb",
	}
	pToken := pagination.Token{ //nolint:gosec // Not a credential, this is a pagination token.
		Size:  0,
		Token: `{"states":null,"current_state":{"token":"1","resource_type_id":"database_user","resource_id":""}}`,
	}
	cli, err := admin.NewClient(admin.UseDigestAuth(publicKey, privateKey))
	assert.Nil(t, err)

	user := &databaseUserBuilder{
		resourceType: databaseUserResourceType,
		client:       cli,
	}
	_, syncResp, err := user.List(ctx, parentResourceID, resource.SyncOpAttrs{PageToken: pToken})
	assert.Nil(t, err)
	pToken.Token = syncResp.NextPageToken
}
