package kube

func (r *kubeProjectRepository) GenProjectHost(ref string) string {
	return ref + "." + r.config.App.ExternalDomain
}

func (*kubeProjectRepository) GenAPIServiceName(ref string) string {
	return ref + "-api-service"
}

func (*kubeProjectRepository) GenAPIDeploymentName(ref string) string {
	return ref + "-api-deployment"
}

func (*kubeProjectRepository) GenAPIIngressRouteName(ref string) string {
	return ref + "-api-ingress"
}

func (*kubeProjectRepository) GenDBIngressRouteTCPName(ref string) string {
	return ref + "-db-ingress-tcp"
}
