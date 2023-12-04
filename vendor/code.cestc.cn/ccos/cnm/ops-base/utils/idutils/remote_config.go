package idutils

import (
	"context"
	"os"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"

	"code.cestc.cn/ccos/cnm/ops-base/logging"
	"code.cestc.cn/ccos/cnm/ops-base/resource/k8s"
)

const (
	envRemoteIdGenerateUrl = "ID_GENERATE_URL"

	remoteMachinedPath = "/api/ccos-ops-cell-agent/v1/id/machined"

	remoteIdPath  = "/api/id-generator/v1/id"
	IdSegmentPath = "/api/id-generator/v1/id-segment"

	envK8sUrl     = "DEV_K8S_URL"
	envKubeConfig = "DEV_KUBE_CONFIG"
)

type InstallConfig struct {
	GlobalBaseDomain string `yaml:"globalBaseDomain"`
	Region           string `yaml:"region"`
}

var (
	defaultIdGeneratorUrl = "https://cell-agent-svc.ccos-ops-app:8443"
)

var installConfig = &InstallConfig{}

var configOnce = sync.Once{}

func getInstallConfig() *InstallConfig {
	configOnce.Do(func() {
		k8sClient, err := k8s.NewKubernetsClient(os.Getenv(envK8sUrl), os.Getenv(envKubeConfig))
		if err != nil {
			logging.Fatal(err)
		}
		cm, err := k8sClient.ClientSet().CoreV1().ConfigMaps("kube-system").Get(context.Background(), "cluster-config-v1", metav1.GetOptions{})
		if err != nil {
			logging.Fatal(err)
		}
		installConfigStr := cm.Data["install-config"]
		_ = yaml.Unmarshal([]byte(installConfigStr), installConfig)
	})
	return installConfig
}

func getUrl() string {
	url := os.Getenv(envRemoteIdGenerateUrl)
	if len(url) <= 0 {
		url = defaultIdGeneratorUrl
	}
	return url
}
