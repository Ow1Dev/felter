// Package httpserver provides the HTTP handlers and OIDC provider integration
// for the authentication proxy.
package httpserver

import "context"

// ErrUserNotFound is returned when a user cannot be found by the provider.
var ErrUserNotFound = context.Canceled

// ProviderUserInfo holds the identity information returned by an OIDC provider.
type ProviderUserInfo struct {
	Sub   string
	Email string
}

// Provider defines the interface for OIDC authentication providers.
type Provider interface {
	Type() string
	BuildAuthURL(state, redirectURI string) string
	ExchangeCode(ctx context.Context, code, redirectURI string) (accessToken string, err error)
	GetUserInfo(ctx context.Context, accessToken string) (*ProviderUserInfo, error)
}

var _ Provider = (*KeycloakProvider)(nil)
