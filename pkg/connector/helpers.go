package connector

import (
	"fmt"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
)

func wrapError(err error, message string) error {
	return fmt.Errorf("mongo-db-connector: %s: %w", message, err)
}

func getSkippEntitlementsAndGrantsAnnotations() annotations.Annotations {
	annotations := annotations.Annotations{}
	annotations.Update(&v2.SkipEntitlementsAndGrants{})

	return annotations
}
