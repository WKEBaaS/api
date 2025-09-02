package kube_project

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
func (r *KubeProjectService) GenProjectHost(ref string) string {
	return ref + "." + r.config.App.ExternalDomain
}

func (r *KubeProjectService) GetAuthAPIURL(ref string) string {
	u := url.URL{
		Scheme: "https",
		Host:   r.GenProjectHost(ref),
		Path:   "/api/auth",
	}
	return u.String()
}

func (r *KubeProjectService) GetRESTAPIURL(ref string) string {
	u := url.URL{
		Scheme: "https",
		Host:   r.GenProjectHost(ref),
		Path:   "/api/rest",
	}
	return u.String()
}

func (r *KubeProjectService) GetJWKSConfigMapName(ref string) string {
	return generateResourceName(ref, "jwks")
}

func (r *KubeProjectService) GetMigrationJobName(ref string) string {
	return generateResourceName(ref, "migration")
}

func (*KubeProjectService) GetAuthAPIDeploymentName(ref string) string {
	return generateResourceName(ref, AuthAPIComponent)
}

func (*KubeProjectService) GetAuthAPIContainerName(ref string) string {
	return generateResourceName(ref, AuthAPIComponent)
}

func (*KubeProjectService) GetAuthAPIServiceName(ref string) string {
	return generateResourceName(ref, AuthAPIComponent)
}

// ===== REST API (PostgREST) =====
func (*KubeProjectService) GetRESTAPIDeploymentName(ref string) string {
	return generateResourceName(ref, RestAPIComponent)
}

func (*KubeProjectService) GetRESTAPIContainerName(ref string, component string) string {
	return generateResourceName(ref, RestAPIComponent, component)
}

func (*KubeProjectService) GetRESTAPIServiceName(ref string) string {
	return generateResourceName(ref, RestAPIComponent)
}

func (*KubeProjectService) GenerateScalarAPIConfig(source string) string {
	config := fmt.Sprintf(`{
	"sources": [
		{
			"url": %s
		}
	],
	"theme": "purple"
}`, source)

	return config
}

func (*KubeProjectService) GetAPIIngressRouteName(ref string) string {
	return generateResourceName(ref, APIIngressComponent)
}

func (*KubeProjectService) GetDBIngressRouteTCPName(ref string) string {
	return generateResourceName(ref, DBComponent)
}

func (*KubeProjectService) GetDatabaseRoleSecretName(ref string, role string) string {
	return generateResourceName(ref, role)
}

func generateResourceName(parts ...string) string {
	return strings.Join(parts, "-")
}
