.PHONY: build run clean test docker

# 项目变量
BINARY_NAME=sentinel
BUILD_DIR=bin
MAIN_PATH=./cmd/sentinel/
DOCKER_IMAGE=kube-sentinel:latest

# 编译（本地开发）
build:
	@echo "🔨 编译 Kube-Sentinel..."
	go build -o $(BUILD_DIR)/$(BINARY_NAME).exe $(MAIN_PATH)
	@echo "✅ 编译完成: $(BUILD_DIR)/$(BINARY_NAME).exe"

# 本地运行
run: build
	@echo "🚀 启动 Kube-Sentinel..."
	./$(BUILD_DIR)/$(BINARY_NAME).exe

# 运行并启用 Prometheus
run-with-prom: build
	PROMETHEUS_ENABLED=true ./$(BUILD_DIR)/$(BINARY_NAME).exe

# 清理构建产物
clean:
	rm -rf $(BUILD_DIR)
	@echo "🧹 已清理"

# 运行测试
test:
	go test -v ./...

# Docker 构建
docker:
	docker build -f deployments/Dockerfile -t $(DOCKER_IMAGE) .
	@echo "🐳 Docker 镜像构建完成: $(DOCKER_IMAGE)"

# 部署到 Minikube
deploy-minikube:
	kubectl apply -f deployments/rbac.yaml
	kubectl apply -f deployments/deployment.yaml
	@echo "🚀 已部署到 Minikube"

# 查看日志
logs:
	kubectl logs -f -n kube-sentinel deployment/kube-sentinel

# Go 代码检查
lint:
	go vet ./...
	@echo "✅ 代码检查通过"
