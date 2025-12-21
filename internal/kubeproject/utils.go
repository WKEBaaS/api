package kubeproject

import (
	"fmt"
	"net/url"
	"strings"
)

// Constants for better maintainability
const (
	AuthAPIComponent    = "auth-api"
	RestAPIComponent    = "rest-api"
	APIIngressComponent = "api"
	DBComponent         = "db"
	PGRSTComponent      = "pgrst"
	OpenAPIComponent    = "openapi"

	RoleApp           = "app"
	RoleAuthenticator = "authenticator"
	RoleAppAdmin      = "app-admin"
)

// ===== Auth API =====

func (s *service) GetProjectHost(ref string) string {
	return ref + "." + s.config.App.ExternalDomain
}

func (s *service) GetAuthAPIURL(ref string) string {
	u := url.URL{
		Scheme: "https",
		Host:   s.GetProjectHost(ref),
		Path:   "/api/auth",
	}
	return u.String()
}

func (s *service) GetRESTAPIURL(ref string) string {
	u := url.URL{
		Scheme: "https",
		Host:   s.GetProjectHost(ref),
		Path:   "/api/rest",
	}
	return u.String()
}

func (s *service) GetJWKSConfigMapName(ref string) string {
	return generateResourceName(ref, "jwks")
}

func (s *service) GetMigrationJobName(ref string) string {
	return generateResourceName(ref, "migration")
}

func (*service) GetAuthAPIDeploymentName(ref string) string {
	return generateResourceName(ref, AuthAPIComponent)
}

func (*service) GetAuthAPIContainerName(ref string) string {
	return generateResourceName(ref, AuthAPIComponent)
}

func (*service) GetAuthAPIServiceName(ref string) string {
	return generateResourceName(ref, AuthAPIComponent)
}

// ===== REST API (PostgREST) =====

func (*service) GetRESTAPIDeploymentName(ref string) string {
	return generateResourceName(ref, RestAPIComponent)
}

func (*service) GetRESTAPIContainerName(ref string, component string) string {
	return generateResourceName(ref, RestAPIComponent, component)
}

func (*service) GetRESTAPIServiceName(ref string) string {
	return generateResourceName(ref, RestAPIComponent)
}

func (*service) GenerateScalarAPIConfig(source string) string {
	config := fmt.Sprintf(`{
	"sources": [
		{
			"url": "%s"
		}
	],
	"theme": "purple"
}`, source)

	return config
}

func (*service) GetAPIIngressRouteName(ref string) string {
	return generateResourceName(ref, APIIngressComponent)
}

func (*service) GetDBIngressRouteTCPName(ref string) string {
	return generateResourceName(ref, DBComponent)
}

func (*service) GetDatabaseRoleSecretName(ref string, role string) string {
	return generateResourceName(ref, role)
}

func (*service) GetDatabaseRWServiceName(ref string) string {
	return generateResourceName(ref, "rw")
}

func generateResourceName(parts ...string) string {
	return strings.Join(parts, "-")
}
