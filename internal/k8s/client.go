package k8s

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// NewClientset 创建 Kubernetes Clientset
// Clientset 是与 K8s API Server 交互的核心客户端
func NewClientset(config *rest.Config) (*kubernetes.Clientset, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("创建 K8s Clientset 失败: %w", err)
	}

	// 验证连接：尝试获取集群版本
	version, err := clientset.Discovery().ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("连接 K8s API Server 失败: %w", err)
	}

	fmt.Printf("[K8s] 已连接到集群，K8s 版本: %s\n", version.GitVersion)
	return clientset, nil
}
