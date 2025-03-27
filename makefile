FRONTEND_DIR = ./web
BACKEND_DIR = .

.PHONY: all build-frontend start-backend generate-swagger swagger-docs

all: build-frontend start-backend

build-frontend:
	@echo "Building frontend..."
	@cd $(FRONTEND_DIR) && npm install && DISABLE_ESLINT_PLUGIN='true' VITE_REACT_APP_VERSION=$(cat VERSION) npm run build

start-backend:
	@echo "Starting backend dev server..."
	@cd $(BACKEND_DIR) && go run main.go &

# 编译生成OpenAPI文档工具
build-swagger-tool:
	@echo "Building swagger document generator..."
	@cd $(BACKEND_DIR) && go build -o bin/swagger-generator docs/generate_swagger.go

# 运行文档生成工具
generate-swagger: build-swagger-tool
	@echo "Generating OpenAPI 3.0 documents..."
	@mkdir -p web/public/swagger
	@mkdir -p docs/swagger
	@bin/swagger-generator
	@echo "OpenAPI 3.0 documentation generated successfully"

# 一键生成OpenAPI文档并启动服务
swagger-docs: generate-swagger start-backend
	@echo "OpenAPI documentation is available at http://localhost:3000/swagger/index.html"

# 生成文档并重新构建前端（完整构建）
build-with-docs: generate-swagger build-frontend start-backend
	@echo "Build completed with OpenAPI documentation"
