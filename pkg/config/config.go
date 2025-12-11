package config

import (
	"fmt"

	"github.com/conductorone/baton-sdk/pkg/field"
)

var PublicKeyField = field.StringField(
	"public-key",
	field.WithDisplayName("Public key"),
	field.WithDescription("Your MongoDB Atlas public key"),
	field.WithRequired(true),
)
var PrivateKeyField = field.StringField(
	"private-key",
	field.WithDisplayName("Private key"),
	field.WithDescription("Your MongoDB Atlas private key"),
	field.WithRequired(true),
	field.WithIsSecret(true),
)
var CreateInviteKeyField = field.BoolField(
	"create-invite-key",
	field.WithDisplayName("Create Invite"),
	field.WithDescription("If enabled, Baton will create invites for users that do not have an account in MongoDB Atlas when provisioning."),
	field.WithRequired(false),
)

var EnableSyncDatabases = field.BoolField(
	"enable-sync-databases",
	field.WithDisplayName("Sync Databases"),
	field.WithDescription("If enabled, Baton will sync database users and roles."),
	field.WithRequired(false),
	field.WithDefaultValue(true),
)

var EnableMongoDriver = field.BoolField(
	"enable-mongo-driver",
	field.WithDisplayName("Enable Mongo Driver"),
	field.WithDescription("If enabled, Baton will use the MongoDB Go Driver to fetch database collections."),
	field.WithRequired(false),
	field.WithDefaultValue(false),
)

var DeleteDatabaseUserWithReadOnly = field.BoolField(
	"delete-database-user-with-read-only",
	field.WithDisplayName("Enable Delete Database User when only having read@admin"),
	field.WithDescription("If enabled, Baton will delete database users that only have read@admin role when revoking access."),
	field.WithRequired(false),
	field.WithDefaultValue(false),
)

var MongoProxyHost = field.StringField(
	"mongo-proxy-host",
	field.WithDisplayName("Mongo Proxy Host"),
	field.WithDescription("The host of the MongoDB proxy server."),
	field.WithExportTarget(field.ExportTargetOps),
)

var MongoProxyPort = field.IntField(
	"mongo-proxy-port",
	field.WithDisplayName("Mongo Proxy Port"),
	field.WithDescription("The port of the MongoDB proxy server."),
	field.WithExportTarget(field.ExportTargetOps),
)

var MongoProxyUser = field.StringField(
	"mongo-proxy-user",
	field.WithDisplayName("Mongo Proxy User"),
	field.WithDescription("The username for the MongoDB proxy server."),
	field.WithExportTarget(field.ExportTargetOps),
)

var MongoProxyPass = field.StringField(
	"mongo-proxy-pass",
	field.WithDisplayName("Mongo Proxy Password"),
	field.WithDescription("The password for the MongoDB proxy server."),
	field.WithIsSecret(true),
	field.WithExportTarget(field.ExportTargetOps),
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
		// Proxy fields
		MongoProxyHost,
		MongoProxyPort,
		MongoProxyUser,
		MongoProxyPass,
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
	if cfg.PublicKey == "" {
		return fmt.Errorf("config: missing public key")
	}

	if cfg.PrivateKey == "" {
		return fmt.Errorf("config: missing private key")
	}

	return nil
}
