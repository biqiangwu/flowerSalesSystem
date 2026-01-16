.PHONY: test build run docker-build k8s-setup k8s-deploy k8s-clean deploy

# 默认目标
default:
	@echo "可用命令:"
	@echo "  make test         - 运行所有测试"
	@echo "  make build        - 构建二进制"
	@echo "  make run          - 运行服务器"
	@echo "  make docker-build - 构建 Docker 镜像"
	@echo "  make k8s-setup    - 初始化 Kind 集群"
	@echo "  make k8s-deploy   - Helm 部署"
	@echo "  make k8s-clean    - 清理 K8s 资源"
	@echo "  make deploy       - 完整部署流程"

# 运行所有测试
test:
	go test -v -race ./...

# 构建二进制
build:
	go build -o bin/server ./cmd/server

# 运行服务器
run:
	go run ./cmd/server

# 构建 Docker 镜像
docker-build:
	docker build -t flowersales:latest .

# 初始化 Kind 集群和基础设施
k8s-setup:
	kind create cluster --config=k8s/kind-cluster.yaml
	kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/main/config/manifests/metallb-native.yaml
	kubectl wait --namespace metallb-system --for=condition=ready pod --selector=app=metallb --timeout=300s
	kubectl apply -f k8s/metallb-native.yaml
	kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.10.0/deploy/static/provider/kind/deploy.yaml
	kubectl wait --namespace ingress-nginx --for=condition=ready pod --selector=app.kubernetes.io/component=controller --timeout=300s

# 使用 Helm 部署应用
k8s-deploy:
	helm install flowersales ./helm/flowersales

# 清理 K8s 资源
k8s-clean:
	helm uninstall flowersales 2>/dev/null || true
	kind delete cluster 2>/dev/null || true

# 完整部署流程
deploy: docker-build k8s-setup k8s-deploy
	@echo ""
	@echo "部署完成！"
	@echo "请添加 /etc/hosts 条目: 127.0.0.1 flowers.local"
	@echo "然后访问: http://flowers.local/"
