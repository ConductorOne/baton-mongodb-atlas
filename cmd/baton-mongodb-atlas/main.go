package main

import (
	"context"

	"github.com/conductorone/baton-sdk/pkg/connectorrunner"

	cfg "github.com/conductorone/baton-mongodb-atlas/pkg/config"

	"github.com/conductorone/baton-mongodb-atlas/pkg/connector"
	configschema "github.com/conductorone/baton-sdk/pkg/config"
)

var version = "dev"
var connectorName = "baton-mongodb-atlas"

func main() {
	ctx := context.Background()
	configschema.RunConnector(ctx,
		"connectorName",
		version,
		cfg.Config,
		connector.New,
		connectorrunner.WithDefaultCapabilitiesConnectorBuilderV2(&connector.MongoDB{}),
	)
}
