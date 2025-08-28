package config

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var PublicKeyField = field.StringField(
	"mongodbatlas_public_key",
	field.WithDisplayName("Public key"),
	field.WithDescription("Your MongoDB Atlas public key"),
	field.WithRequired(true),
)
var PrivateKeyField = field.StringField(
	"mongodbatlas_private_key",
	field.WithDisplayName("Private key"),
	field.WithDescription("Your MongoDB Atlas private key"),
	field.WithRequired(true),
	field.WithIsSecret(true),
)
var CreateInviteKeyField = field.BoolField(
	"mongodbatlas_create_invite",
	field.WithDisplayName("Create Invite"),
	field.WithDescription("If enabled, Baton will create invites for users that do not have an account in MongoDB Atlas when provisioning."),
	field.WithRequired(false),
)

var EnableSyncDatabases = field.BoolField(
	"mongodbatlas_enable_sync_database",
	field.WithDisplayName("Sync Databases"),
	field.WithDescription("If enabled, Baton will sync database users and roles."),
	field.WithRequired(false),
	field.WithDefaultValue(true),
)

var EnableMongoDriver = field.BoolField(
	"mongodbatlas_enable_mongo_driver",
	field.WithDisplayName("Enable Mongo Driver"),
	field.WithDescription("If enabled, Baton will use the MongoDB Go Driver to fetch database collections."),
	field.WithRequired(false),
	field.WithDefaultValue(false),
)

var DeleteDatabaseUserWithReadOnly = field.BoolField(
	"mongodbatlas_enable_delete_database_user_with_read_only",
	field.WithDisplayName("Enable Delete Database User when only having read@admin"),
	field.WithDescription("If enabled, Baton will delete database users that only have read@admin role when revoking access."),
	field.WithRequired(false),
	field.WithDefaultValue(false),
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
	field.WithConstraints(
		field.FieldsDependentOn(
			[]field.SchemaField{
				EnableMongoDriver,
			},
			[]field.SchemaField{
				EnableSyncDatabases,
			},
		),
	),
)

// ValidateConfig is run after the configuration is loaded, and should return an
// error if it isn't valid. Implementing this function is optional, it only
// needs to perform extra validations that cannot be encoded with configuration
// parameters.
func ValidateConfig(cfg *Mongodbatlas) error {
	return nil
}
