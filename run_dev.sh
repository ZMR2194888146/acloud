#!/bin/bash

echo "=== ACloud 开发环境启动 ==="
echo "正在启动 Wails 开发服务器..."
echo ""

# 检查依赖
echo "检查 Go 模块依赖..."
go mod tidy

echo ""
echo "启动开发服务器..."
wails dev

echo ""
echo "开发服务器已停止"