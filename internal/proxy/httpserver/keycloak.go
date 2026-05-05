package httpserver

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type KeycloakProvider struct {
	URL           string
	ClientID      string
	ClientSecret  string
	RedirectURI   string
	Realm         string
	TokenEndpoint string
	UserInfoURL   string
}

func NewKeycloakProvider(url, realm, clientID, clientSecret, redirectURI string) *KeycloakProvider {
	return &KeycloakProvider{
		URL:          url,
		Realm:        realm,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURI:  redirectURI,
	}
}

func (p *KeycloakProvider) Type() string {
	return "keycloak"
}

func (p *KeycloakProvider) BuildAuthURL(state, redirectURI string) string {
	return fmt.Sprintf("%s/realms/%s/protocol/openid-connect/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=openid&state=%s",
		p.URL, p.Realm, p.ClientID, url.QueryEscape(redirectURI), state)
}

type oidcTokenResponse struct {
	AccessToken string `json:"access_token"`
	IdToken     string `json:"id_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func (p *KeycloakProvider) ExchangeCode(ctx context.Context, code, redirectURI string) (string, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("client_id", p.ClientID)
	data.Set("client_secret", p.ClientSecret)

	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", p.URL, p.Realm)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.PostForm(tokenURL, data)
	if err != nil {
		return "", fmt.Errorf("exchange code: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}

	var tokenResp oidcTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("unmarshal token response: %w, body: %s", err, string(body))
	}
	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("empty access token, body: %s", string(body))
	}
	return tokenResp.AccessToken, nil
}

type oidcUserInfo struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
}

func (p *KeycloakProvider) GetUserInfo(ctx context.Context, accessToken string) (*ProviderUserInfo, error) {
	userInfoURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/userinfo", p.URL, p.Realm)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, userInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo returned status %d: %s", resp.StatusCode, string(body))
	}

	var userInfo oidcUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("unmarshal user info: %w, body: %s", err, string(body))
	}
	return &ProviderUserInfo{Sub: userInfo.Sub, Email: userInfo.Email}, nil
}

func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
