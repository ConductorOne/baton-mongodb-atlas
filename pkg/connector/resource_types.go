package connector

import (
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

// The user resource type is for all user objects from the database.
var (
	userResourceType = &v2.ResourceType{
		Id:          "user",
		DisplayName: "User",
		Description: "A MongoDB Atlas Organization User",
		Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_USER},
		Annotations: getSkippEntitlementsAndGrantsAnnotations(),
	}
	teamResourceType = &v2.ResourceType{
		Id:          "team",
		DisplayName: "Team",
		Description: "A MongoDB Atlas Team",
		Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_GROUP},
	}
)
