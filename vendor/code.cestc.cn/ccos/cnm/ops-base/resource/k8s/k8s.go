package k8s

import (
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"k8s.io/client-go/dynamic"

	"code.cestc.cn/ccos/cnm/ops-base/logging"
)

type Config struct {
	NeedConnect bool   `yaml:"needConnect"`
	Url         string `yaml:"url"`
	KubeConfig  string `yaml:"kubeConfig"`
}

type K8s struct {
	kubernetesClient *kubernetes.Clientset
	dynamicClient    dynamic.Interface
	config           *rest.Config
}

func NewKubernetsClient(url, configPath string) (k *K8s, err error) {
	k = &K8s{}
	config, err := clientcmd.BuildConfigFromFlags(url, configPath)
	if err != nil {
		logging.Errorw("build config error", zap.Error(err))
		return nil, err
	}
	k.config = config
	k.kubernetesClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		logging.Errorw("new client set error", zap.Error(err))
		return nil, err
	}
	k.dynamicClient, err = dynamic.NewForConfig(config)
	if err != nil {
		logging.Errorw("new dynamic client error", zap.Error(err))
		return nil, err
	}
	return k, nil
}

func (k *K8s) ClientSet() *kubernetes.Clientset {
	return k.kubernetesClient
}

func (k *K8s) DynamicClient() dynamic.Interface {
	return k.dynamicClient
}

func (k *K8s) Config() *rest.Config {
	return k.config
}
