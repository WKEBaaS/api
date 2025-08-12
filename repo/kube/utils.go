package kube

func (r *kubeProjectRepository) GenProjectHost(ref string) string {
	return ref + "." + r.config.App.ExternalDomain
}

func (*kubeProjectRepository) GetAPIServiceName(ref string) string {
	return ref + "-api"
}

func (*kubeProjectRepository) GetAPIDeploymentName(ref string) string {
	return ref + "-api"
}

func (*kubeProjectRepository) GetAPIAuthContainerName(ref string) string {
	return ref + "-auth"
}

func (*kubeProjectRepository) GetAPIIngressRouteName(ref string) string {
	return ref + "-api"
}

func (*kubeProjectRepository) GetDBIngressRouteTCPName(ref string) string {
	return ref + "-db-tcp"
}
