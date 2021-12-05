// Package encdns defines the Resolver interface that encrypted DNS resolvers must satisfy.
package encdns

import (
	"context"
)

type Resolver interface {
	LookupTXT(ctx context.Context, host string) ([]string, error)
}
