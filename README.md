# 🛡️ Kube-Sentinel

> **Kubernetes 集群自愈监控系统** — 基于 Kubernetes Informer 机制的实时事件监控与告警平台

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-1.20+-326CE5?style=flat&logo=kubernetes)](https://kubernetes.io/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## ✨ 特性

### 核心功能

- **🔄 Informer 事件驱动监控** — 基于 Kubernetes `SharedInformerFactory` 实现的 List-Watch 机制，实时监听 Pod 生命周期事件，替代低效的轮询方案
- **🔍 智能异常检测** — 自动识别容器重启次数超阈值、CrashLoopBackOff、ImagePullBackOff、ErrImagePull 等异常状态
- **📊 Prometheus 集成** — 实时查询 Pod CPU/内存使用率，通过 PromQL 发现资源异常
- **📝 结构化日志** — 基于 `zap` 的高性能日志系统，支持 Debug/Info/Warn/Error 多级别输出
- **🛡️ 优雅关闭** — 监听 SIGINT/SIGTERM 信号，实现所有 goroutine 的安全退出
- **⚙️ 灵活配置** — 支持环境变量和配置文件的多种参数管理方式

### 技术亮点

- **生产级架构** — 采用 Go 标准项目布局，清晰的目录结构和职责分离
- **高性能** — Informer 本地缓存机制，减少对 Kubernetes API Server 的请求压力
- **可扩展** — 预留告警通知接口（微信/钉钉 Webhook），易于集成现有告警系统

## 🏗️ 架构

```
┌─────────────────────────────────────────────────────────────┐
│                      Kube-Sentinel                          │
│                                                             │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐  │
│  │  Pod Watcher  │    │  Prometheus  │    │   Alerter    │  │
│  │  (Informer)   │◄───│   Client     │◄───│  (Webhook)   │  │
│  └──────┬───────┘    └──────┬───────┘    └──────┬───────┘  │
│         │                   │                    │          │
│  ┌──────▼───────────────────▼────────────────────▼───────┐  │
│  │                  Event Processor                       │  │
│  └───────────────────────┬───────────────────────────────┘  │
│                          │                                  │
│  ┌───────────────────────▼───────────────────────────────┐  │
│  │              Structured Logger (zap)                   │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
         │                          │
         ▼                          ▼
  K8s API Server              Prometheus
  (List-Watch)                (PromQL)
```

## 📋 前置要求

- **Go** 1.22 或更高版本
- **Kubernetes** 集群（v1.20+）或 Minikube/Kind 等本地环境
- **Prometheus**（可选，用于资源监控）
- **kubectl** 配置好集群访问权限

## 🚀 快速开始

### 方式一：直接运行

```bash
# 1. 克隆仓库
git clone https://github.com/Kirara707/kube-sentinel.git
cd kube-sentinel

# 2. 编译项目
go build -o bin/sentinel ./cmd/sentinel

# 3. 运行（自动使用 ~/.kube/config）
./bin/sentinel
```

### 方式二：使用 Makefile

```bash
# 编译
make build

# 运行
make run

# 启用 Prometheus 查询
make run-with-prom
```

### 方式三：Docker 部署

```bash
# 构建镜像
make docker

# 部署到 Kubernetes 集群
make deploy-minikube

# 查看日志
make logs
```

## 🧪 测试场景

```bash
# 制造 ImagePullBackOff 事件
kubectl run broken --image=nginx:nonexistent

# 制造容器崩溃重启
kubectl run crasher --image=busybox -- sh -c "exit 1"

# 正常创建和删除
kubectl run nginx --image=nginx
kubectl delete pod nginx
```

## ⚙️ 配置

Kube-Sentinel 支持通过环境变量进行配置：

| 环境变量 | 默认值 | 描述 |
|---------|--------|------|
| `KUBECONFIG` | `~/.kube/config` | kubeconfig 文件路径 |
| `WATCH_NAMESPACE` | `""` (所有命名空间) | 监控的目标命名空间 |
| `RESTART_THRESHOLD` | `3` | 容器重启告警阈值 |
| `RESYNC_INTERVAL` | `30s` | Informer 全量同步间隔 |
| `PROMETHEUS_ENABLED` | `false` | 是否启用 Prometheus 查询 |
| `PROMETHEUS_URL` | `http://localhost:9090` | Prometheus 服务地址 |
| `QUERY_INTERVAL` | `60s` | Prometheus 查询间隔 |
| `LOG_LEVEL` | `info` | 日志级别 (debug/info/warn/error) |

### 配置示例

```bash
# 监控特定命名空间并启用 Prometheus
export WATCH_NAMESPACE=production
export PROMETHEUS_ENABLED=true
export PROMETHEUS_URL=http://prometheus:9090
export RESTART_THRESHOLD=5

./bin/sentinel
```

## 📁 项目结构

```
kube-sentinel/
├── cmd/
│   └── sentinel/
│       └── main.go                 # 应用程序入口
├── internal/
│   ├── config/
│   │   ├── app.go                  # 配置管理与加载
│   │   └── app_test.go             # 配置单元测试
│   ├── k8s/
│   │   ├── client.go               # Kubernetes Clientset 初始化
│   │   └── config.go               # kubeconfig 加载器
│   ├── monitor/
│   │   ├── pod_watcher.go          # Informer Pod 监控核心逻辑
│   │   ├── pod_watcher_test.go     # 监控器单元测试
│   │   └── types.go                # 数据类型定义
│   └── prometheus/
│       ├── client.go               # Prometheus 客户端
│       └── query.go                # PromQL 查询实现
├── pkg/
│   └── logger/
│       └── logger.go               # 结构化日志封装
├── deployments/
│   ├── Dockerfile                  # 多阶段 Docker 构建
│   ├── deployment.yaml             # Kubernetes 部署清单
│   └── rbac.yaml                   # RBAC 权限配置
├── Makefile                        # 构建自动化脚本
├── go.mod                          # Go 模块定义
└── go.sum                          # 依赖版本锁定
```

## 🔧 开发

### 运行测试

```bash
# 运行所有测试
make test

# 或使用 go test
go test -v ./...
```

### 代码检查

```bash
# Go 静态分析
make lint
```

### 清理构建产物

```bash
make clean
```

## 🎯 监控事件

Kube-Sentinel 监控以下 Kubernetes 事件：

| 事件类型 | 触发条件 | 告警级别 |
|---------|---------|---------|
| Pod 新增 | Pod 被创建 | Info |
| Pod 更新 | Pod 状态变更 | Info |
| Pod 删除 | Pod 被删除 | Info |
| 容器重启 | 重启次数 ≥ 阈值 | Warn |
| 异常状态 | CrashLoopBackOff/ImagePullBackOff 等 | Warn |

## 🛣️ Roadmap

- [ ] 集成微信/钉钉 Webhook 告警
- [ ] 支持 Node 事件监控
- [ ] 支持 Deployment/StatefulSet 监控
- [ ] 添加 Dashboard 展示
- [ ] 支持自定义告警规则
- [ ] 添加 Metrics 端点

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 提交 Pull Request

## 📮 联系方式

- 作者: Kirara707
- 项目链接: [https://github.com/Kirara707/kube-sentinel](https://github.com/Kirara707/kube-sentinel)

---

**⭐ 如果这个项目对你有帮助，请给它一个 Star！**