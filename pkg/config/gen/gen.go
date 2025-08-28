package main

import (
	cfg "github.com/conductorone/baton-mongodb-atlas/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/config"
)

func main() {
	config.Generate("mongodbAtlas", cfg.Config)
}
