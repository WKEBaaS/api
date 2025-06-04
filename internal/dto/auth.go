package dto

import "net/http"

type AuthLoginInput struct {
	RedirectURL string `query:"redirect_url" example:"http://example.com" doc:"Redirect URL"`
}

type AuthCallbackInput struct {
	StateCookie string `cookie:"state" example:"d0sP4Bmr98VQc5WV4799W" doc:"State (NanoID format"`
	NonceCookie string `cookie:"nonce" example:"d0sP4Bmr98VQc5WV4799W" doc:"Nonce (NanoID format)"`
	RedirectURL string `cookie:"redirect_url" example:"http://example.com" doc:"Redirect URL"`
	State       string `query:"state" format:"uuid" example:"d0sP4Bmr98VQc5WV4799W" doc:"State (NanoID format)"`
	Code        string `query:"code" example:"code" doc:"Code received from OAuth2 provider"`
}

type AuthLoginOutput struct {
	Status         int
	Url            string       `header:"Location"`
	NonceCookie    *http.Cookie `header:"Set-Cookie"`
	StateCookie    *http.Cookie `header:"Set-Cookie"`
	RedirectCookie *http.Cookie `header:"Set-Cookie"`
}

type AuthCallbackOutput struct {
	Status      int
	Url         string       `header:"Location"`
	TokenCookie *http.Cookie `header:"Set-Cookie"`
	Body        struct {
		Message string `json:"message"`
		Ok      bool   `json:"ok" doc:"true if the request was successful"`
	}
}
