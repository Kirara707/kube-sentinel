package k8s

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// LoadKubeConfig 加载 Kubernetes 配置
// 支持两种模式：
//   - 集群内运行（InCluster）：自动读取 ServiceAccount Token
//   - 本地开发：从 kubeconfig 文件加载
func LoadKubeConfig(kubeConfigPath string) (*rest.Config, error) {
	// 1. 优先尝试集群内配置（生产环境）
	config, err := rest.InClusterConfig()
	if err == nil {
		fmt.Println("[K8s] 使用集群内配置（InCluster Mode）")
		return config, nil
	}

	// 2. 回退到 kubeconfig 文件（本地开发）
	if kubeConfigPath == "" {
		kubeConfigPath = defaultKubeConfigPath()
	}

	config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("无法加载 kubeconfig（路径: %s）: %w", kubeConfigPath, err)
	}

	fmt.Printf("[K8s] 使用 kubeconfig: %s\n", kubeConfigPath)
	return config, nil
}

// defaultKubeConfigPath 返回默认的 kubeconfig 路径
func defaultKubeConfigPath() string {
	// 优先读取环境变量
	if kubePath := os.Getenv("KUBECONFIG"); kubePath != "" {
		return kubePath
	}

	// 默认路径: ~/.kube/config
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".kube", "config")
}
