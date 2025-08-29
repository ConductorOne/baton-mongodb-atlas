package connector

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"go.mongodb.org/mongo-driver/mongo/options"

	"go.uber.org/zap"

	"github.com/conductorone/baton-mongodb-atlas/pkg/connector/mongodriver"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.mongodb.org/atlas-sdk/v20250312006/admin"
)

// Const roles for db https://www.mongodb.com/docs/atlas/mongodb-users-roles-and-privileges/#std-label-atlas-user-privileges
// Only that uses DB.
var dbRoles = []string{
	"dbAdmin",   // Only db
	"read",      // DB and collections
	"readWrite", // DB and collections
}

type databaseBuilder struct {
	client                         *admin.APIClient
	enableMongoDriver              bool
	mongodriver                    *mongodriver.MongoDriver
	deleteDatabaseUserWithReadOnly bool
}

func (o *databaseBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return databaseResourceType
}

func newDatabaseBuilder(
	client *admin.APIClient,
	enableMongoDriver bool,
	mongodriver *mongodriver.MongoDriver,
	deleteDatabaseUserWithReadOnly bool,
) *databaseBuilder {
	return &databaseBuilder{
		client:                         client,
		enableMongoDriver:              enableMongoDriver,
		mongodriver:                    mongodriver,
		deleteDatabaseUserWithReadOnly: deleteDatabaseUserWithReadOnly,
	}
}

// List returns all the users from the database as resource objects.
// Users include a UserTrait because they are the 'shape' of a standard user.
func (o *databaseBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	if parentResourceID == nil {
		return nil, "", nil, nil
	}

	if parentResourceID.ResourceType != mongoClusterResourceType.Id {
		return nil, "", nil, fmt.Errorf("invalid parent resource type: %s", parentResourceID.ResourceType)
	}

	bag, page, err := parsePageToken(pToken.Token, &v2.ResourceId{ResourceType: databaseResourceType.Id})
	if err != nil {
		return nil, "", nil, err
	}

	splited := strings.Split(parentResourceID.Resource, "/")
	if len(splited) != 3 {
		return nil, "", nil, fmt.Errorf("invalid parent resource ID: %s", parentResourceID.Resource)
	}

	groupID := splited[0]
	// clusterID := splited[1]
	clusterName := splited[2]

	clusterInfo, _, err := o.client.ClustersApi.GetCluster(ctx, groupID, clusterName).
		Execute() //nolint:bodyclose // The SDK handles closing the response body
	if err != nil {
		return nil, "", nil, err
	}

	connectionsStrings := clusterInfo.GetConnectionStrings()
	if connectionsStrings.Standard == nil {
		l.Warn("Cluster does not have a standard connection string", zap.Any("clusterInfo", clusterInfo))
		return nil, "", nil, fmt.Errorf("Cluster does not have a standard connection string: %s", clusterName)
	}

	connectionString := strings.Split(*connectionsStrings.Standard, ",")
	if len(connectionString) == 0 {
		return nil, "", nil, fmt.Errorf("cluster %s does not have a valid connection string", clusterName)
	}
	process := strings.TrimPrefix(connectionString[0], "mongodb://")

	var databases []string
	if o.enableMongoDriver {
		l.Info("using mongo driver to list databases")

		_, mongoDriver, err := o.mongodriver.Connect(ctx, groupID, clusterName)
		if err != nil {
			l.Error("failed to connect to MongoDB Atlas cluster skipping database sync", zap.String("group_id", groupID), zap.String("cluster_name", clusterName), zap.Error(err))
			// We are skipping databases if we can't connect to the cluster.
			return nil, "", nil, nil
		}

		names, err := mongoDriver.ListDatabaseNames(ctx, bson.M{}, &options.ListDatabasesOptions{
			AuthorizedDatabases: boolPointer(false),
			NameOnly:            boolPointer(true),
		})
		if err != nil {
			return nil, "", nil, err
		}

		databases = names
	} else {
		l.Info("using atlas api to list databases")

		execute, _, err := o.client.MonitoringAndLogsApi.ListDatabases(ctx, groupID, process).
			PageNum(page).
			ItemsPerPage(resourcePageSize).
			Execute() //nolint:bodyclose // The SDK handles closing the response body
		if err != nil {
			return nil, "", nil, err
		}

		if execute.Results == nil || len(execute.GetResults()) == 0 {
			return nil, "", nil, nil
		}

		for _, database := range execute.GetResults() {
			databases = append(databases, database.GetDatabaseName())
		}
	}

	resources := make([]*v2.Resource, 0)

	for _, database := range databases {
		if database == "" {
			l.Warn("Skipping database with empty name")
			continue
		}

		resource, err := newDatabaseResource(groupID, clusterName, database, parentResourceID, o.enableMongoDriver)
		if err != nil {
			return nil, "", nil, wrapError(err, "failed to create resource")
		}

		resources = append(resources, resource)
	}

	nextPage := ""
	if !o.enableMongoDriver {
		nextPage, err = getPageTokenFromPage(bag, page+1)
		if err != nil {
			return nil, "", nil, err
		}
	}

	return resources, nextPage, nil, nil
}

func newDatabaseResource(groupID string, clusterName string, dbName string, parentId *v2.ResourceId, enableMongoDriver bool) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"db_name": dbName,
	}

	appTraits := []rs.AppTraitOption{
		rs.WithAppProfile(profile),
	}

	rsOptions := []rs.ResourceOption{
		rs.WithParentResourceID(parentId),
	}

	if enableMongoDriver {
		rsOptions = append(rsOptions, rs.WithAnnotation(&v2.ChildResourceType{
			ResourceTypeId: collectionResourceType.Id,
		}))
	}

	id := fmt.Sprintf("%s/%s/%s", groupID, clusterName, dbName)

	resource, err := rs.NewAppResource(
		fmt.Sprintf("%s - %s", clusterName, dbName),
		databaseResourceType,
		id,
		appTraits,
		rsOptions...,
	)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

// Entitlements always returns an empty slice for users.
func (o *databaseBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	ents := make([]*v2.Entitlement, 0)

	for _, role := range dbRoles {
		ent := entitlement.NewAssignmentEntitlement(
			resource,
			role,
			entitlement.WithGrantableTo(databaseUserResourceType),
			entitlement.WithDisplayName(fmt.Sprintf("%s - %s", resource.DisplayName, role)),
		)
		ents = append(ents, ent)
	}

	return ents, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *databaseBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	splited := strings.Split(resource.Id.Resource, "/")
	if len(splited) != 3 {
		return nil, "", nil, fmt.Errorf("invalid resource ID: %s", resource.Id.Resource)
	}

	groupID := splited[0]

	bag, page, err := parsePageToken(pToken.Token, &v2.ResourceId{ResourceType: databaseResourceType.Id})
	if err != nil {
		return nil, "", nil, err
	}

	dbUsers, _, err := o.client.DatabaseUsersApi.ListDatabaseUsers(ctx, groupID).
		IncludeCount(true).PageNum(page).ItemsPerPage(resourcePageSize).
		Execute() //nolint:bodyclose // The SDK handles closing the response body
	if err != nil {
		return nil, "", nil, err
	}

	if len(dbUsers.GetResults()) == 0 {
		return nil, "", nil, nil
	}

	var grants []*v2.Grant
	for _, user := range dbUsers.GetResults() {
		userId := &v2.ResourceId{
			ResourceType: databaseUserResourceType.Id,
			Resource:     user.Username,
		}

		for _, role := range user.GetRoles() {
			// We only want to return grants for roles that are not collection specific.
			if role.HasCollectionName() {
				continue
			}

			if !slices.Contains(dbRoles, role.GetRoleName()) {
				continue
			}

			grants = append(grants, grant.NewGrant(resource, role.RoleName, userId))
		}
	}

	nextPage, err := getPageTokenFromPage(bag, page+1)
	if err != nil {
		return nil, "", nil, err
	}

	return grants, nextPage, nil, nil
}

func (o *databaseBuilder) Grant(ctx context.Context, resource *v2.Resource, entitlement *v2.Entitlement) ([]*v2.Grant, annotations.Annotations, error) {
	if resource.Id.ResourceType != databaseUserResourceType.Id {
		return nil, nil, fmt.Errorf("invalid resource type: %s", resource.Id.ResourceType)
	}

	// We want database Id
	splited := strings.Split(entitlement.Resource.Id.Resource, "/")
	if len(splited) != 3 {
		return nil, nil, fmt.Errorf("invalid resource ID: %s", resource.Id.Resource)
	}

	groupID := splited[0]
	dbName := splited[2]
	role := entitlement.Slug

	dbUsername := resource.Id.Resource

	dbUser, _, err := o.client.DatabaseUsersApi.GetDatabaseUser(ctx, groupID, "admin", dbUsername).
		Execute() //nolint:bodyclose // The SDK handles closing the response body
	if err != nil {
		return nil, nil, err
	}

	newRoles := append(dbUser.GetRoles(), admin.DatabaseUserRole{
		DatabaseName: dbName,
		RoleName:     role,
	})

	dbUser.Roles = &newRoles

	_, _, err = o.client.DatabaseUsersApi.UpdateDatabaseUser(ctx, groupID, "admin", dbUsername, dbUser).
		Execute() //nolint:bodyclose // The SDK handles closing the response body
	if err != nil {
		return nil, nil, err
	}

	userId := &v2.ResourceId{
		ResourceType: databaseUserResourceType.Id,
		Resource:     dbUser.Username,
	}

	return []*v2.Grant{
		grant.NewGrant(resource, role, userId),
	}, nil, nil
}

func (o *databaseBuilder) Revoke(ctx context.Context, grant *v2.Grant) (annotations.Annotations, error) {
	if grant.Principal.Id.ResourceType != databaseUserResourceType.Id {
		return nil, fmt.Errorf("invalid resource type: %s", grant.Principal.Id.ResourceType)
	}

	splited := strings.Split(grant.Entitlement.Resource.Id.Resource, "/")
	if len(splited) != 3 {
		return nil, fmt.Errorf("invalid resource ID: %s", grant.Entitlement.Resource.Id.Resource)
	}

	groupID := splited[0]
	dbName := splited[2]
	role := grant.Entitlement.Slug

	dbUsername := grant.Principal.Id.Resource

	dbUser, _, err := o.client.DatabaseUsersApi.GetDatabaseUser(ctx, groupID, "admin", dbUsername).
		Execute() //nolint:bodyclose // The SDK handles closing the response body
	if err != nil {
		return nil, err
	}

	// Remove the role from the user
	var newRoles []admin.DatabaseUserRole
	for _, r := range dbUser.GetRoles() {
		if r.DatabaseName == dbName && r.RoleName == role {
			continue // Skip the role we want to remove
		}
		newRoles = append(newRoles, r)
	}

	if o.shouldDeleteUser(newRoles) {
		_, err := o.client.DatabaseUsersApi.DeleteDatabaseUser(ctx, groupID, "admin", dbUsername).
			Execute() //nolint:bodyclose // The SDK handles closing the response body
		if err != nil {
			return nil, err
		}
	} else {
		dbUser.Roles = &newRoles
		_, _, err = o.client.DatabaseUsersApi.UpdateDatabaseUser(ctx, groupID, "admin", dbUsername, dbUser).
			Execute() //nolint:bodyclose // The SDK handles closing the response body
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (o *databaseBuilder) shouldDeleteUser(roles []admin.DatabaseUserRole) bool {
	if len(roles) == 0 {
		return true
	}

	if len(roles) == 1 {
		if !o.deleteDatabaseUserWithReadOnly {
			return false
		}

		if roles[0].RoleName == "read" && roles[0].DatabaseName == "admin" {
			return true
		}
	}

	return false
}

func boolPointer(v bool) *bool {
	return &v
}
