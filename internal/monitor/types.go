package monitor

import (
	"time"
)

// EventType 事件类型枚举
type EventType string

const (
	EventAdd    EventType = "ADD"
	EventUpdate EventType = "UPDATE"
	EventDelete EventType = "DELETE"
)

// PodHealthEvent Pod 健康事件
type PodHealthEvent struct {
	EventType     EventType `json:"eventType"`     // 事件类型
	PodName       string    `json:"podName"`       // Pod 名称
	Namespace     string    `json:"namespace"`     // 命名空间
	Phase         string    `json:"phase"`         // Running/Pending/Failed...
	NodeName      string    `json:"nodeName"`      // 所在节点
	Timestamp     time.Time `json:"timestamp"`     // 事件时间
	Message       string    `json:"message"`       // 事件描述
}

// RestartAlert Pod 重启告警
type RestartAlert struct {
	PodName        string    `json:"podName"`        // Pod 名称
	Namespace      string    `json:"namespace"`      // 命名空间
	ContainerName  string    `json:"containerName"`  // 容器名称
	RestartCount   int32     `json:"restartCount"`   // 重启次数
	LastState      string    `json:"lastState"`      // 上次终止原因
	ExitCode       int32     `json:"exitCode"`       // 退出码
	Timestamp      time.Time `json:"timestamp"`      // 告警时间
}

// PodMetric Pod 性能指标（从 Prometheus 获取）
type PodMetric struct {
	PodName      string    `json:"podName"`
	Namespace    string    `json:"namespace"`
	CPUUsage     float64   `json:"cpuUsage"`     // CPU 使用率（核数）
	MemoryUsage  float64   `json:"memoryUsage"`  // 内存使用量（bytes）
	Timestamp    time.Time `json:"timestamp"`
}
