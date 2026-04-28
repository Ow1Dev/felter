package httpserver

import "context"

var ErrUserNotFound = context.Canceled

type ProviderUserInfo struct {
	Sub   string
	Email string
}

type Provider interface {
	Type() string
	BuildAuthURL(state, redirectURI string) string
	ExchangeCode(ctx context.Context, code, redirectURI string) (accessToken string, err error)
	GetUserInfo(ctx context.Context, accessToken string) (*ProviderUserInfo, error)
}

var _ Provider = (*KeycloakProvider)(nil)
