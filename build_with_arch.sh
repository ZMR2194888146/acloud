#!/bin/bash

# ACloud 多平台编译脚本 - 带架构名称版本
# 生成的文件将包含系统名称和架构信息

set -e

VERSION="1.0.0"
APP_NAME="acloud"

echo "🚀 开始编译 ACloud v${VERSION}"
echo "📦 所有文件将包含系统和架构信息"
echo ""

# 创建构建目录
mkdir -p build/bin
cd build/bin

# 清理旧文件
echo "🧹 清理旧的构建文件..."
rm -f acloud* 2>/dev/null || true
rm -rf acloud*.app 2>/dev/null || true

cd ../..

echo ""
echo "=== 开始多平台编译 ==="

# 1. Windows AMD64
echo ""
echo "🪟 编译 Windows AMD64..."
if wails build -platform windows/amd64 -o "${APP_NAME}-v${VERSION}-windows-amd64.exe"; then
    echo "✅ Windows AMD64 编译成功"
    # 移动并重命名文件
    if [ -f "build/bin/${APP_NAME}.exe" ]; then
        mv "build/bin/${APP_NAME}.exe" "build/bin/${APP_NAME}-v${VERSION}-windows-amd64.exe"
    fi
else
    echo "❌ Windows AMD64 编译失败"
fi

# 2. macOS ARM64 (Apple Silicon)
echo ""
echo "🍎 编译 macOS ARM64 (Apple Silicon)..."
if wails build -platform darwin/arm64; then
    echo "✅ macOS ARM64 编译成功"
    # 重命名应用包
    if [ -d "build/bin/${APP_NAME}.app" ]; then
        mv "build/bin/${APP_NAME}.app" "build/bin/${APP_NAME}-v${VERSION}-darwin-arm64.app"
    fi
else
    echo "❌ macOS ARM64 编译失败"
fi

# 3. macOS AMD64 (Intel)
echo ""
echo "🍎 编译 macOS AMD64 (Intel)..."
if wails build -platform darwin/amd64; then
    echo "✅ macOS AMD64 编译成功"
    # 重命名应用包
    if [ -d "build/bin/${APP_NAME}.app" ]; then
        mv "build/bin/${APP_NAME}.app" "build/bin/${APP_NAME}-v${VERSION}-darwin-amd64.app"
    fi
else
    echo "❌ macOS AMD64 编译失败"
fi

# 4. Linux AMD64
echo ""
echo "🐧 尝试编译 Linux AMD64..."
if wails build -platform linux/amd64; then
    echo "✅ Linux AMD64 编译成功"
    # 重命名可执行文件
    if [ -f "build/bin/${APP_NAME}" ]; then
        mv "build/bin/${APP_NAME}" "build/bin/${APP_NAME}-v${VERSION}-linux-amd64"
        chmod +x "build/bin/${APP_NAME}-v${VERSION}-linux-amd64"
    fi
else
    echo "⚠️  Linux AMD64 编译失败 (可能需要在 Linux 环境中编译)"
fi

# 5. Linux ARM64
echo ""
echo "🐧 尝试编译 Linux ARM64..."
if wails build -platform linux/arm64; then
    echo "✅ Linux ARM64 编译成功"
    # 重命名可执行文件
    if [ -f "build/bin/${APP_NAME}" ]; then
        mv "build/bin/${APP_NAME}" "build/bin/${APP_NAME}-v${VERSION}-linux-arm64"
        chmod +x "build/bin/${APP_NAME}-v${VERSION}-linux-arm64"
    fi
else
    echo "⚠️  Linux ARM64 编译失败 (可能需要在对应环境中编译)"
fi

echo ""
echo "=== 编译结果 ==="
echo "📁 构建目录: build/bin/"
echo ""

# 显示编译结果
for file in build/bin/${APP_NAME}-v${VERSION}-*; do
    if [ -e "$file" ]; then
        filename=$(basename "$file")
        if [ -f "$file" ]; then
            size=$(du -h "$file" | cut -f1)
            echo "📄 $filename ($size)"
        elif [ -d "$file" ]; then
            size=$(du -sh "$file" | cut -f1)
            echo "📱 $filename ($size)"
        fi
    fi
done

echo ""
echo "=== 文件命名规范 ==="
echo "格式: ${APP_NAME}-v${VERSION}-{系统}-{架构}.{扩展名}"
echo "示例:"
echo "  • ${APP_NAME}-v${VERSION}-windows-amd64.exe    (Windows 64位)"
echo "  • ${APP_NAME}-v${VERSION}-darwin-arm64.app     (macOS Apple Silicon)"
echo "  • ${APP_NAME}-v${VERSION}-darwin-amd64.app     (macOS Intel)"
echo "  • ${APP_NAME}-v${VERSION}-linux-amd64          (Linux 64位)"
echo "  • ${APP_NAME}-v${VERSION}-linux-arm64          (Linux ARM64)"

echo ""
echo "=== 创建 ZIP 发布包 ==="
RELEASE_DIR="build/release"
mkdir -p "$RELEASE_DIR"

# 为所有平台创建 ZIP 压缩包
for file in build/bin/${APP_NAME}-v${VERSION}-*; do
    if [ -e "$file" ]; then
        filename=$(basename "$file")
        echo "📦 正在打包: $filename (ZIP格式)"
        
        if [ -f "$file" ]; then
            # 可执行文件打包为 ZIP
            (cd build/bin && zip "../release/${filename}.zip" "$(basename "$file")")
        elif [ -d "$file" ]; then
            # macOS 应用包打包为 ZIP
            (cd build/bin && zip -r "../release/${filename}.zip" "$(basename "$file")")
        fi
        
        if [ $? -eq 0 ]; then
            echo "✅ 已创建: build/release/${filename}.zip"
        else
            echo "❌ 打包失败: $filename"
        fi
    fi
done

echo ""
echo "=== 发布包信息 ==="
if [ -d "$RELEASE_DIR" ] && [ "$(ls -A "$RELEASE_DIR" 2>/dev/null)" ]; then
    echo "📁 发布包目录: build/release/"
    for zip_file in "$RELEASE_DIR"/*.zip; do
        if [ -f "$zip_file" ]; then
            size=$(du -h "$zip_file" | cut -f1)
            echo "📦 $(basename "$zip_file") ($size)"
        fi
    done
else
    echo "⚠️  没有创建发布包"
fi

echo ""
echo "🎉 编译和打包完成！"
echo "📁 构建文件: build/bin/"
echo "📦 发布包: build/release/ (所有平台统一使用ZIP格式)"
echo "💡 提示: Linux 版本可能需要在对应的 Linux 环境中编译"
