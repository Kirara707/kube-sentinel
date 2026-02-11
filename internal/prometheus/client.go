package prometheus

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"

	"github.com/Kirara707/kube-sentinel/pkg/logger"
)

// Client 封装 Prometheus API 客户端
type Client struct {
	api     v1.API
	address string
}

// NewClient 创建 Prometheus 客户端
func NewClient(address string) (*Client, error) {
	client, err := api.NewClient(api.Config{
		Address: address,
	})
	if err != nil {
		return nil, fmt.Errorf("创建 Prometheus 客户端失败: %w", err)
	}

	return &Client{
		api:     v1.NewAPI(client),
		address: address,
	}, nil
}

// CheckHealth 检测 Prometheus 连通性
func (c *Client) CheckHealth() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 执行一个简单查询来验证连通性
	_, _, err := c.api.Query(ctx, "up", time.Now())
	if err != nil {
		return fmt.Errorf("Prometheus 连接失败 (%s): %w", c.address, err)
	}

	logger.Log.Infof("✅ Prometheus 已连接: %s", c.address)
	return nil
}
