package dto

import (
	"net/http"
	"net/url"
)

type AuthLoginInput struct {
	RedirectURL string `query:"redirect_url" example:"http://example.com" doc:"Redirect URL" required:"false"`
}

type AuthLoginOutput struct {
	Status         int
	Url            string       `header:"Location"`
	NonceCookie    *http.Cookie `header:"Set-Cookie"`
	StateCookie    *http.Cookie `header:"Set-Cookie"`
	RedirectCookie *http.Cookie `header:"Set-Cookie"`
}

type AuthCallbackInput struct {
	StateCookie string `cookie:"state" example:"d0sP4Bmr98VQc5WV4799W" doc:"State (NanoID format"`
	NonceCookie string `cookie:"nonce" example:"d0sP4Bmr98VQc5WV4799W" doc:"Nonce (NanoID format)"`
	RedirectURL string `cookie:"redirect_url" example:"http://example.com" doc:"Redirect URL"`
	State       string `query:"state" format:"uuid" example:"d0sP4Bmr98VQc5WV4799W" doc:"State (NanoID format)"`
	Code        string `query:"code" example:"code" doc:"Code received from OAuth2 provider"`
}

type AuthCallbackOutput struct {
	Status      int
	Url         string       `header:"Location"`
	TokenCookie *http.Cookie `header:"Set-Cookie"`
	Body        struct {
		Ok bool `json:"ok" doc:"true if the request was successful"`
	}
}

type AuthLogoutInput struct {
	PostLogoutRedirectURI string `query:"post_logout_redirect_uri" example:"http://example.com" doc:"Post logout redirect URI" required:"false"`
}

type AuthLogoutOutput struct {
	Status      int
	Url         url.URL      `header:"Location"`
	TokenCookie *http.Cookie `header:"Set-Cookie" doc:"Token cookie to be cleared"`
}
