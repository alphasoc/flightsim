package encdns

type ProviderType int

// Providers of encrypted DNS (in some form or another - DoH/DoT/etc).
const (
	GoogleProvider ProviderType = iota
	CloudFlareProvider
	Quad9Provider
	OpenDNSProvider
)
