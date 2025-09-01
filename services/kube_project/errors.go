// Package kube
//
// kubernetes related repository for project management
package kube_project

import "errors"

// Errors for kubeProjectRepository
var (
	// cluster errors
	ErrFailedToOpenPostgresClusterYAML        = errors.New("failed to open Postgres cluster YAML file")
	ErrFailedToDecodePostgresClusterYAML      = errors.New("failed to decode Postgres cluster YAML")
	ErrFailedToGetSpecFromPostgresClusterYAML = errors.New("failed to get spec from Postgres cluster YAML")
	ErrSpecNotFoundInPostgresClusterYAML      = errors.New("spec not found in Postgres cluster YAML")
	ErrFailedToSetSpecStorageSize             = errors.New("failed to set storage size in Postgres cluster spec")
	ErrFailedToCreatePostgresCluster          = errors.New("failed to create Postgres cluster")
	ErrFailedToDeletePostgresCluster          = errors.New("failed to delete Postgres cluster")
	// database errors
	ErrFailedToOpenPostgresDatabaseYAML   = errors.New("failed to open Postgres database YAML file")
	ErrFailedToDecodePostgresDatabaseYAML = errors.New("failed to decode Postgres database YAML")
	ErrFailedToSetSpecClusterName         = errors.New("failed to set cluster name in Postgres database spec")
	ErrFeiledToCreatePostgresDatabase     = errors.New("failed to create Postgres database")
	ErrFailedToDeletePostgresDatabase     = errors.New("failed to delete Postgres database")
	ErrFailedToReadDatabaseSecret         = errors.New("failed to read Postgres database secret")
	ErrFailedToResetDatabasePassword      = errors.New("failed to reset Postgres database password")
	// ingress route TCP errors
	ErrFailedToOpenIngressRouteTCPYAML    = errors.New("failed to open IngressRouteTCP YAML file")
	ErrFailedToDecodeIngressRouteTCPYAML  = errors.New("failed to decode IngressRouteTCP YAML")
	ErrFailedToSetSpecRouteTCPName        = errors.New("failed to set route TCP name in IngressRouteTCP spec")
	ErrFailedToSetSpecRouteTCPServiceName = errors.New("failed to set service name in IngressRouteTCP spec")
	ErrFailedToSetSpecTLSSecretName       = errors.New("failed to set TLS secret name in IngressRouteTCP spec")
	ErrFailedToCreateIngressRouteTCP      = errors.New("failed to create IngressRouteTCP")
	ErrFailedToDeleteIngressRouteTCP      = errors.New("failed to delete IngressRouteTCP")
)
