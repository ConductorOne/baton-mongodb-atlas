package config

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var PublicKeyField = field.StringField("public-key",
	field.WithDescription(``),
	field.WithRequired(true),
)
var PrivateKeyField = field.StringField("private-key",
	field.WithDescription(``),
	field.WithRequired(true),
)
var CreateInviteKeyField = field.BoolField("create-invite-key",
	field.WithDescription("Create the invitation user email"),
	field.WithRequired(false),
)

var EnableSyncDatabases = field.BoolField("enable-sync-databases",
	field.WithDescription("Enable sync of databases as resources"),
	field.WithRequired(false),
)

var EnableMongoDriver = field.BoolField("enable-mongo-driver",
	field.WithDescription("Enable MongoDB driver for additional functionality such as collection management"),
	field.WithRequired(false),
)

var DeleteDatabaseUserWithReadOnly = field.BoolField(
	"delete-database-user-with-read-only",
	field.WithDescription("Delete database users that only have read@admin when revoke"),
	field.WithRequired(false),
)

//go:generate go run ./gen
var Config = field.NewConfiguration(
	[]field.SchemaField{
		PublicKeyField,
		PrivateKeyField,
		CreateInviteKeyField,
		EnableSyncDatabases,
		EnableMongoDriver,
		DeleteDatabaseUserWithReadOnly,
	},
	field.WithConnectorDisplayName("MongodbAtlas"),
	field.WithHelpUrl("/docs/baton/mongodb-atlas"),
	field.WithIconUrl("/static/app-icons/mongodb.svg"),
)

// ValidateConfig is run after the configuration is loaded, and should return an
// error if it isn't valid. Implementing this function is optional, it only
// needs to perform extra validations that cannot be encoded with configuration
// parameters.
func ValidateConfig(cfg *Mongodbatlas) error {
	return nil
}
