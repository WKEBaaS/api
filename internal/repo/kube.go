package repo

import (
	"baas-api/internal/configs"
	"log"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubeRepository interface{}

type kubeRepository struct {
	kubeConfig    *rest.Config
	clientset     *kubernetes.Clientset
	dynamicClient *dynamic.DynamicClient
	appConfig     *configs.Config
}

func NewKubeService(config *configs.Config) KubeRepository {
	svc := &kubeRepository{}

	kc, err := clientcmd.BuildConfigFromFlags("", config.Kube.ConfigPath)
	if err != nil {
		log.Panicf("Failed to build kube config %+v", err)
	}
	clientset, err := kubernetes.NewForConfig(kc)
	if err != nil {
		log.Panicf("Failed to create kube client %+v", err)
	}
	dynamicClient, err := dynamic.NewForConfig(kc)
	if err != nil {
		log.Panicf("Failed to create dynamic client %+v", err)
	}
	svc.kubeConfig = kc
	svc.clientset = clientset
	svc.dynamicClient = dynamicClient
	svc.appConfig = config

	return svc
}
