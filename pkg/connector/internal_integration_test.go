package connector

import (
	"context"
	"os"
	"testing"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/atlas-sdk/v20231001002/admin"
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
	pToken := &pagination.Token{
		Size:  0,
		Token: `{"states":null,"current_state":{"token":"1","resource_type_id":"database_user","resource_id":""}}`,
	}
	cli, err := admin.NewClient(admin.UseDigestAuth(publicKey, privateKey))
	assert.Nil(t, err)

	user := &databaseUserBuilder{
		resourceType: databaseUserResourceType,
		client:       cli,
	}
	for pToken.Token != "" {
		_, token, _, err := user.List(ctx, parentResourceID, pToken)
		assert.Nil(t, err)
		pToken.Token = token
	}
}
