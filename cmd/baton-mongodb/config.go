package main

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-sdk/pkg/cli"
	"github.com/spf13/cobra"
)

// config defines the external configuration required for the connector to run.
type config struct {
	cli.BaseConfig `mapstructure:",squash"` // Puts the base config options in the same place as the connector options

	OrganizationId string `mapstructure:"organization-id"`
	PublicKey      string `mapstructure:"public-key"`
	PrivateKey     string `mapstructure:"private-key"`
}

// validateConfig is run after the configuration is loaded, and should return an error if it isn't valid.
func validateConfig(ctx context.Context, cfg *config) error {
	if cfg.OrganizationId == "" {
		return fmt.Errorf("organization-id is required")
	}

	if cfg.PublicKey == "" {
		return fmt.Errorf("public-key is required")
	}

	if cfg.PrivateKey == "" {
		return fmt.Errorf("private-key is required")
	}

	return nil
}

func cmdFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String("organization-id", "", "Organization ID")
	cmd.PersistentFlags().String("public-key", "", "Public Key")
	cmd.PersistentFlags().String("private-key", "", "Private Key")
}
