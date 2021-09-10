// Package providers ...
package providers

import (
	"context"
	"math/rand"
	"net/http"

	"github.com/alphasoc/flightsim/simulator/encdns"
)

// Provider represents a DoH provider.  addr is used to dial, queryURL is the base for
// DoH queries, and client is the HTTP client used in the actual queries.
type Provider struct {
	addr     string
	queryURL string
	client   *http.Client
}

// Providers supporting DoH.
var providers = []encdns.ProviderType{
	encdns.GoogleProvider,
	encdns.CloudFlareProvider,
	encdns.Quad9Provider,
	encdns.OpenDNSProvider,
}

// NewRandom returns a 'random' Queryable provider.
func NewRandom(ctx context.Context) encdns.Queryable {
	pIdx := encdns.ProviderType(rand.Intn(len(providers)))
	var p encdns.Queryable
	switch providers[pIdx] {
	case encdns.GoogleProvider:
		p = NewGoogle(ctx)
	case encdns.CloudFlareProvider:
		p = NewCloudFlare(ctx)
	case encdns.Quad9Provider:
		p = NewQuad9(ctx)
	case encdns.OpenDNSProvider:
		p = NewOpenDNS(ctx)
	}
	return p
}
