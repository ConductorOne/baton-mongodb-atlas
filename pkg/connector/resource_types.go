package connector

import (
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

// The user resource type is for all user objects from the database.
var (
	organizationResourceType = &v2.ResourceType{
		Id:          "organization",
		DisplayName: "Organization",
		Description: "A MongoDB Atlas Organization",
	}
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
	projectResourceType = &v2.ResourceType{
		Id:          "project",
		DisplayName: "Project",
		Description: "A MongoDB Atlas Project",
		Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_GROUP},
	}
	databaseUserResourceType = &v2.ResourceType{
		Id:          "database_user",
		DisplayName: "Database User",
		Description: "A MongoDB Atlas Database User",
		Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_USER},
		Annotations: getSkippEntitlementsAndGrantsAnnotations(),
	}

	mongoClusterUserResourceType = &v2.ResourceType{
		Id:          "mongo_cluster",
		DisplayName: "Mongo Cluster",
		Description: "A MongoDB Atlas Cluster",
		Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_APP},
		Annotations: getSkippEntitlementsAndGrantsAnnotations(),
	}
)
