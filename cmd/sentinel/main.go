package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Kirara707/kube-sentinel/internal/config"
	k8sClient "github.com/Kirara707/kube-sentinel/internal/k8s"
	"github.com/Kirara707/kube-sentinel/internal/monitor"
	promClient "github.com/Kirara707/kube-sentinel/internal/prometheus"
	"github.com/Kirara707/kube-sentinel/pkg/logger"
)

const banner = `
╔══════════════════════════════════════════════════════╗
║           🛡️  Kube-Sentinel v0.1.0                  ║
║     Kubernetes 集群自愈监控系统                        ║
║     Informer + Prometheus + 告警通知                  ║
╚══════════════════════════════════════════════════════╝
`

func main() {
	fmt.Print(banner)

	// 1. 加载配置
	cfg := config.Load()

	// 2. 初始化日志
	logger.Init(cfg.LogLevel)
	defer logger.Sync()

	logger.Log.Info("Kube-Sentinel 正在启动...")

	// 3. 加载 kubeconfig，创建 K8s 客户端
	kubeConfig, err := k8sClient.LoadKubeConfig(cfg.KubeConfigPath)
	if err != nil {
		logger.Log.Fatalf("加载 kubeconfig 失败: %v", err)
	}

	clientset, err := k8sClient.NewClientset(kubeConfig)
	if err != nil {
		logger.Log.Fatalf("创建 K8s 客户端失败: %v", err)
	}

	// 4. 全局 stop 通道 —— 用于优雅关闭所有 goroutine
	stopCh := make(chan struct{})

	// 5. 启动 Pod Informer（核心功能）
	watcher := monitor.NewPodWatcher(
		clientset,
		cfg.ResyncInterval,
		cfg.RestartThreshold,
		cfg.Namespace,
	)

	if err := watcher.Start(stopCh); err != nil {
		logger.Log.Fatalf("启动 Pod Informer 失败: %v", err)
	}

	// 6. 启动 Prometheus 查询（可选）
	if cfg.PrometheusEnabled {
		prom, err := promClient.NewClient(cfg.PrometheusURL)
		if err != nil {
			logger.Log.Warnf("Prometheus 客户端初始化失败: %v（将跳过指标查询）", err)
		} else {
			if err := prom.CheckHealth(); err != nil {
				logger.Log.Warnf("Prometheus 连接异常: %v（将跳过指标查询）", err)
			} else {
				go prom.StartPeriodicQuery(cfg.QueryInterval, stopCh)
			}
		}
	} else {
		logger.Log.Info("Prometheus 查询未启用（设置 PROMETHEUS_ENABLED=true 启用）")
	}

	logger.Log.Info("✅ Kube-Sentinel 已完全启动！正在监听集群事件...")

	// 7. 监听 OS 信号 —— 优雅关闭
	//    面试金句："收到 SIGTERM 后，通过关闭 stopCh 通知所有 goroutine 退出，
	//    实现了 graceful shutdown"
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh
	logger.Log.Infof("收到信号 %v，正在优雅关闭...", sig)
	close(stopCh)

	logger.Log.Info("Kube-Sentinel 已关闭。")
}
