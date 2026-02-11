package prometheus

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/common/model"

	"github.com/Kirara707/kube-sentinel/internal/monitor"
	"github.com/Kirara707/kube-sentinel/pkg/logger"
)

// GetPodCPUUsage 查询 Pod 的 CPU 实时使用率
// PromQL: sum(rate(container_cpu_usage_seconds_total{pod='xxx'}[5m]))
func (c *Client) GetPodCPUUsage(podName, namespace string) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := fmt.Sprintf(
		`sum(rate(container_cpu_usage_seconds_total{pod="%s", namespace="%s"}[5m]))`,
		podName, namespace,
	)

	result, warnings, err := c.api.Query(ctx, query, time.Now())
	if err != nil {
		return 0, fmt.Errorf("查询 Pod CPU 失败: %w", err)
	}

	if len(warnings) > 0 {
		logger.Log.Warnf("Prometheus 查询警告: %v", warnings)
	}

	// 解析结果
	vector, ok := result.(model.Vector)
	if !ok || len(vector) == 0 {
		return 0, nil // 没有数据
	}

	cpuUsage := float64(vector[0].Value)
	return cpuUsage, nil
}

// GetPodMemoryUsage 查询 Pod 的内存使用量（bytes）
// PromQL: sum(container_memory_usage_bytes{pod='xxx'})
func (c *Client) GetPodMemoryUsage(podName, namespace string) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := fmt.Sprintf(
		`sum(container_memory_usage_bytes{pod="%s", namespace="%s"})`,
		podName, namespace,
	)

	result, warnings, err := c.api.Query(ctx, query, time.Now())
	if err != nil {
		return 0, fmt.Errorf("查询 Pod 内存失败: %w", err)
	}

	if len(warnings) > 0 {
		logger.Log.Warnf("Prometheus 查询警告: %v", warnings)
	}

	vector, ok := result.(model.Vector)
	if !ok || len(vector) == 0 {
		return 0, nil
	}

	memUsage := float64(vector[0].Value)
	return memUsage, nil
}

// GetPodMetrics 获取 Pod 的完整指标
func (c *Client) GetPodMetrics(podName, namespace string) (*monitor.PodMetric, error) {
	cpu, err := c.GetPodCPUUsage(podName, namespace)
	if err != nil {
		return nil, err
	}

	mem, err := c.GetPodMemoryUsage(podName, namespace)
	if err != nil {
		return nil, err
	}

	return &monitor.PodMetric{
		PodName:     podName,
		Namespace:   namespace,
		CPUUsage:    cpu,
		MemoryUsage: mem,
		Timestamp:   time.Now(),
	}, nil
}

// StartPeriodicQuery 定时查询所有 Pod 指标
// 在后台协程中运行，每隔 interval 查询一次
func (c *Client) StartPeriodicQuery(interval time.Duration, stopCh <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	logger.Log.Infof("📊 Prometheus 定时查询已启动，间隔: %v", interval)

	for {
		select {
		case <-ticker.C:
			c.queryAllPods()
		case <-stopCh:
			logger.Log.Info("Prometheus 定时查询已停止")
			return
		}
	}
}

// queryAllPods 查询异常 Pod 的指标
func (c *Client) queryAllPods() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 查询所有非 Running 状态的 Pod
	query := `kube_pod_status_phase{phase=~"Failed|Pending|Unknown"}`
	result, _, err := c.api.Query(ctx, query, time.Now())
	if err != nil {
		logger.Log.Errorf("Prometheus 定时查询失败: %v", err)
		return
	}

	vector, ok := result.(model.Vector)
	if !ok {
		return
	}

	for _, sample := range vector {
		podName := string(sample.Metric["pod"])
		namespace := string(sample.Metric["namespace"])
		phase := string(sample.Metric["phase"])

		logger.Log.Warnw("📊 异常 Pod 发现",
			"pod", podName,
			"namespace", namespace,
			"phase", phase,
		)
	}
}
