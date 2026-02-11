package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// 测试默认值
	os.Clearenv()
	cfg := Load()

	if cfg.RestartThreshold != 3 {
		t.Errorf("expected RestartThreshold=3, got %d", cfg.RestartThreshold)
	}

	if cfg.ResyncInterval != 30*time.Second {
		t.Errorf("expected ResyncInterval=30s, got %v", cfg.ResyncInterval)
	}

	if cfg.LogLevel != "info" {
		t.Errorf("expected LogLevel=info, got %s", cfg.LogLevel)
	}
}

func TestLoadCustom(t *testing.T) {
	// 测试自定义配置
	os.Setenv("RESTART_THRESHOLD", "5")
	os.Setenv("WATCH_NAMESPACE", "production")
	os.Setenv("LOG_LEVEL", "debug")

	cfg := Load()

	if cfg.RestartThreshold != 5 {
		t.Errorf("expected RestartThreshold=5, got %d", cfg.RestartThreshold)
	}

	if cfg.Namespace != "production" {
		t.Errorf("expected Namespace=production, got %s", cfg.Namespace)
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("expected LogLevel=debug, got %s", cfg.LogLevel)
	}

	os.Clearenv()
}
