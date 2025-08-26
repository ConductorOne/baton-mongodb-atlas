package main

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var publicKeyField = field.StringField("public-key",
	field.WithDescription(``),
	field.WithRequired(true),
)
var privateKeyField = field.StringField("private-key",
	field.WithDescription(``),
	field.WithRequired(true),
)
var createInviteKeyField = field.BoolField("create-invite-key",
	field.WithDescription("Create the invitation user email"),
	field.WithRequired(false),
)

var enableSyncDatabases = field.BoolField("enable-sync-databases",
	field.WithDescription("Enable sync of databases as resources"),
	field.WithRequired(false),
)

var enableMongoDriver = field.BoolField("enable-mongo-driver",
	field.WithDescription("Enable MongoDB driver for additional functionality such as collection management"),
	field.WithRequired(false),
)

var deleteDatabaseUserWithReadOnly = field.BoolField(
	"delete-database-user-with-read-only",
	field.WithDescription("Delete database users that only have read@admin when revoke"),
	field.WithRequired(false),
)

var configFields = []field.SchemaField{
	publicKeyField,
	privateKeyField,
	createInviteKeyField,
	enableSyncDatabases,
	enableMongoDriver,
	deleteDatabaseUserWithReadOnly,
}

var configRelations = []field.SchemaFieldRelationship{
	field.FieldsDependentOn(
		[]field.SchemaField{
			enableMongoDriver,
		},
		[]field.SchemaField{
			enableSyncDatabases,
		},
	),
}

var cfg = field.Configuration{
	Fields:      configFields,
	Constraints: configRelations,
}
