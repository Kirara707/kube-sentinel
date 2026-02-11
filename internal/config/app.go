package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// AppConfig 应用全局配置
type AppConfig struct {
	// K8s 配置
	KubeConfigPath string `json:"kubeConfigPath"` // kubeconfig 路径，空则自动检测

	// Informer 配置
	ResyncInterval time.Duration `json:"resyncInterval"` // Informer 全量同步间隔，默认 30s
	Namespace      string        `json:"namespace"`      // 监听的命名空间，空则监听所有

	// 告警阈值
	RestartThreshold int32 `json:"restartThreshold"` // 触发告警的重启次数阈值

	// Prometheus 配置
	PrometheusURL     string        `json:"prometheusUrl"`     // Prometheus 地址
	PrometheusEnabled bool          `json:"prometheusEnabled"` // 是否启用 Prometheus 查询
	QueryInterval     time.Duration `json:"queryInterval"`     // Prometheus 查询间隔

	// 日志配置
	LogLevel string `json:"logLevel"` // debug, info, warn, error
}

// Load 从环境变量加载配置，缺省值采用生产级默认值
func Load() *AppConfig {
	cfg := &AppConfig{
		KubeConfigPath:    getEnv("KUBECONFIG", ""),
		ResyncInterval:    getDurationEnv("RESYNC_INTERVAL", 30*time.Second),
		Namespace:         getEnv("WATCH_NAMESPACE", ""), // 空 = 所有命名空间
		RestartThreshold:  getInt32Env("RESTART_THRESHOLD", 3),
		PrometheusURL:     getEnv("PROMETHEUS_URL", "http://localhost:9090"),
		PrometheusEnabled: getBoolEnv("PROMETHEUS_ENABLED", false),
		QueryInterval:     getDurationEnv("QUERY_INTERVAL", 60*time.Second),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
	}

	fmt.Printf("[Config] 已加载配置: Namespace=%s, RestartThreshold=%d, PrometheusEnabled=%v\n",
		cfg.NamespaceDisplay(), cfg.RestartThreshold, cfg.PrometheusEnabled)
	return cfg
}

// NamespaceDisplay 返回可读的命名空间显示
func (c *AppConfig) NamespaceDisplay() string {
	if c.Namespace == "" {
		return "ALL"
	}
	return c.Namespace
}

// --- 环境变量辅助函数 ---

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		d, err := time.ParseDuration(val)
		if err == nil {
			return d
		}
	}
	return fallback
}

func getInt32Env(key string, fallback int32) int32 {
	if val := os.Getenv(key); val != "" {
		n, err := strconv.ParseInt(val, 10, 32)
		if err == nil {
			return int32(n)
		}
	}
	return fallback
}

func getBoolEnv(key string, fallback bool) bool {
	if val := os.Getenv(key); val != "" {
		b, err := strconv.ParseBool(val)
		if err == nil {
			return b
		}
	}
	return fallback
}
