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
		ResyncInterval:    getEnvWithParse("RESYNC_INTERVAL", 30*time.Second, time.ParseDuration),
		Namespace:         getEnv("WATCH_NAMESPACE", ""), // 空 = 所有命名空间
		RestartThreshold:  getEnvWithParse("RESTART_THRESHOLD", int32(3), parseToInt32),
		PrometheusURL:     getEnv("PROMETHEUS_URL", "http://localhost:9090"),
		PrometheusEnabled: getEnvWithParse("PROMETHEUS_ENABLED", false, strconv.ParseBool),
		QueryInterval:     getEnvWithParse("QUERY_INTERVAL", 60*time.Second, time.ParseDuration),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
	}

	// Validate configuration
	if errs := cfg.Validate(); len(errs) > 0 {
		for _, err := range errs {
			fmt.Printf("[Config] ⚠️  Validation error: %v\n", err)
		}
		// Non-blocking: log warnings but continue with defaults
	}

	fmt.Printf("[Config] 已加载配置: Namespace=%s, RestartThreshold=%d, PrometheusEnabled=%v\n",
		cfg.NamespaceDisplay(), cfg.RestartThreshold, cfg.PrometheusEnabled)
	return cfg
}

// Validate validates the configuration and returns any errors found
func (c *AppConfig) Validate() []error {
	var errs []error

	// Validate RestartThreshold
	if c.RestartThreshold < 0 {
		errs = append(errs, fmt.Errorf("RestartThreshold must be non-negative, got %d", c.RestartThreshold))
	}

	// Validate ResyncInterval
	if c.ResyncInterval < 0 {
		errs = append(errs, fmt.Errorf("ResyncInterval must be non-negative, got %v", c.ResyncInterval))
	}
	if c.ResyncInterval > 0 && c.ResyncInterval < 10*time.Second {
		errs = append(errs, fmt.Errorf("ResyncInterval recommended minimum is 10s, got %v", c.ResyncInterval))
	}

	// Validate QueryInterval
	if c.QueryInterval < 0 {
		errs = append(errs, fmt.Errorf("QueryInterval must be non-negative, got %v", c.QueryInterval))
	}

	// Validate LogLevel
	validLogLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true,
	}
	if !validLogLevels[c.LogLevel] {
		errs = append(errs, fmt.Errorf("LogLevel must be one of: debug, info, warn, error, got %s", c.LogLevel))
	}

	return errs
}

// NamespaceDisplay 返回可读的命名空间显示
func (c *AppConfig) NamespaceDisplay() string {
	if c.Namespace == "" {
		return "ALL"
	}
	return c.Namespace
}

// --- 环境变量辅助函数 ---

// Parser is a generic parser function type for environment variables
type Parser[T any] func(string) (T, error)

// getEnv reads an environment variable and parses it using the provided parser
func getEnvWithParse[T any](key string, fallback T, parse Parser[T]) T {
	if val := os.Getenv(key); val != "" {
		if parsed, err := parse(val); err == nil {
			return parsed
		}
	}
	return fallback
}

// getEnv reads a string environment variable
func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

// parseToInt32 converts a string to int32
func parseToInt32(s string) (int32, error) {
	n, err := strconv.ParseInt(s, 10, 32)
	return int32(n), err
}
