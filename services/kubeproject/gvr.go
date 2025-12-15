// Package kubeproject
//
// kubernetes related repository for project management
package kubeproject

import "k8s.io/apimachinery/pkg/runtime/schema"

// CloudNativePG Cluster GVR
var clusterGVR = schema.GroupVersionResource{
	Group:    "postgresql.cnpg.io",
	Version:  "v1",
	Resource: "clusters", // Resource Name: 通常是CRD定義中的 `spec.names.plural`
}

var databaseGVR = schema.GroupVersionResource{
	Group:    "postgresql.cnpg.io",
	Version:  "v1",
	Resource: "databases",
}

var ingressRouteTCPGVR = schema.GroupVersionResource{
	Group:    "traefik.io",
	Version:  "v1alpha1",
	Resource: "ingressroutetcps",
}

var ingressRouteGVR = schema.GroupVersionResource{
	Group:    "traefik.io",
	Version:  "v1alpha1",
	Resource: "ingressroutes",
}
