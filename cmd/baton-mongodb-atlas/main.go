package main

import (
	"context"
	"fmt"
	"os"

	cfg "github.com/conductorone/baton-mongodb-atlas/pkg/config"

	configschema "github.com/conductorone/baton-sdk/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/types"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"

	"github.com/conductorone/baton-mongodb-atlas/pkg/connector"
)

var version = "dev"
var connectorName = "baton-mongodb-atlas"

func main() {
	ctx := context.Background()
	_, cmd, err := configschema.DefineConfiguration(
		ctx,
		connectorName,
		getConnector,
		cfg.Config,
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	cmd.Version = version
	err = cmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func getConnector(ctx context.Context, cc *cfg.Mongodbatlas) (types.ConnectorServer, error) {
	l := ctxzap.Extract(ctx)

	if err := cfg.ValidateConfig(cc); err != nil {
		return nil, err
	}

	cb, err := connector.New(
		ctx,
		cc.Mongodbatlas_public_key,
		cc.Mongodbatlas_private_key,
		cc.Mongodbatlas_create_invite,
		cc.Mongodbatlas_enable_sync_database,
		cc.Mongodbatlas_enable_mongo_driver,
		cc.Mongodbatlas_enable_delete_database_user_with_read_only,
	)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}

	c, err := connectorbuilder.NewConnector(ctx, cb)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}

	return c, nil
}
