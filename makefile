FRONTEND_DIR = ./web
BACKEND_DIR = .

.PHONY: all start-frontend start-backend

all: start-frontend start-backend

# 启动前端开发服务器
start-frontend:
	@echo "Starting frontend dev server..."
	@cd $(FRONTEND_DIR) && npm start &

# 启动后端开发服务器
start-backend:
	@echo "Starting backend dev server..."
	@cd $(BACKEND_DIR) && go run main.go &

