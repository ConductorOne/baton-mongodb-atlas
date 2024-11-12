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

var configFields = []field.SchemaField{
	publicKeyField,
	privateKeyField,
}

var configRelations = []field.SchemaFieldRelationship{}

var cfg = field.Configuration{
	Fields:      configFields,
	Constraints: configRelations,
}
