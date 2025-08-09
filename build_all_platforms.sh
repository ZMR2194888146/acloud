#!/bin/bash

echo "=== ACloud 多平台编译脚本 ==="
echo "开始编译 ACloud 应用程序..."

# 创建构建目录
mkdir -p build/bin

# 定义版本号
VERSION="1.0.0"

echo ""
echo "1. 编译 Windows 版本 (amd64)..."
wails build -platform windows/amd64 -o "acloud-windows-amd64.exe"
if [ $? -eq 0 ]; then
    echo "✅ Windows 版本编译成功: acloud-windows-amd64.exe"
    # 重命名文件以包含架构信息
    if [ -f "build/bin/acloud.exe" ]; then
        mv "build/bin/acloud.exe" "build/bin/acloud-v${VERSION}-windows-amd64.exe"
        echo "📦 已重命名为: acloud-v${VERSION}-windows-amd64.exe"
    fi
else
    echo "❌ Windows 版本编译失败"
fi

echo ""
echo "2. 编译 macOS 版本 (arm64)..."
wails build -platform darwin/arm64
if [ $? -eq 0 ]; then
    echo "✅ macOS ARM64 版本编译成功"
    # 重命名 macOS 应用包
    if [ -d "build/bin/acloud.app" ]; then
        mv "build/bin/acloud.app" "build/bin/acloud-v${VERSION}-darwin-arm64.app"
        echo "📦 已重命名为: acloud-v${VERSION}-darwin-arm64.app"
    fi
else
    echo "❌ macOS ARM64 版本编译失败"
fi

echo ""
echo "3. 编译 macOS 版本 (amd64)..."
wails build -platform darwin/amd64
if [ $? -eq 0 ]; then
    echo "✅ macOS Intel 版本编译成功"
    # 重命名 macOS 应用包
    if [ -d "build/bin/acloud.app" ]; then
        mv "build/bin/acloud.app" "build/bin/acloud-v${VERSION}-darwin-amd64.app"
        echo "📦 已重命名为: acloud-v${VERSION}-darwin-amd64.app"
    fi
else
    echo "❌ macOS Intel 版本编译失败"
fi

echo ""
echo "4. 尝试编译 Linux 版本 (amd64)..."
echo "注意: Linux 交叉编译可能需要在 Linux 环境中进行"
wails build -platform linux/amd64
if [ $? -eq 0 ]; then
    echo "✅ Linux 版本编译成功"
    # 重命名 Linux 可执行文件
    if [ -f "build/bin/acloud" ]; then
        mv "build/bin/acloud" "build/bin/acloud-v${VERSION}-linux-amd64"
        echo "📦 已重命名为: acloud-v${VERSION}-linux-amd64"
    fi
else
    echo "❌ Linux 版本编译失败 (可能需要在 Linux 环境中编译)"
fi

echo ""
echo "5. 尝试编译 Linux 版本 (arm64)..."
echo "注意: Linux ARM64 交叉编译可能需要在对应环境中进行"
wails build -platform linux/arm64
if [ $? -eq 0 ]; then
    echo "✅ Linux ARM64 版本编译成功"
    # 重命名 Linux ARM64 可执行文件
    if [ -f "build/bin/acloud" ]; then
        mv "build/bin/acloud" "build/bin/acloud-v${VERSION}-linux-arm64"
        echo "📦 已重命名为: acloud-v${VERSION}-linux-arm64"
    fi
else
    echo "❌ Linux ARM64 版本编译失败 (可能需要在对应环境中编译)"
fi

echo ""
echo "=== 编译结果 ==="
echo "构建文件位置: build/bin/"
ls -la build/bin/

echo ""
echo "=== 文件信息 ==="
for file in build/bin/*; do
    if [ -f "$file" ]; then
        echo "📄 文件: $(basename "$file")"
        file "$file"
        echo "📏 大小: $(du -h "$file" | cut -f1)"
        echo "🕐 修改时间: $(stat -f "%Sm" -t "%Y-%m-%d %H:%M:%S" "$file" 2>/dev/null || stat -c "%y" "$file" 2>/dev/null)"
        echo ""
    elif [ -d "$file" ] && [[ "$file" == *.app ]]; then
        echo "📱 应用包: $(basename "$file")"
        if [ -f "$file/Contents/MacOS/"* ]; then
            file "$file/Contents/MacOS/"*
            echo "📏 大小: $(du -sh "$file" | cut -f1)"
            echo "🕐 修改时间: $(stat -f "%Sm" -t "%Y-%m-%d %H:%M:%S" "$file" 2>/dev/null || stat -c "%y" "$file" 2>/dev/null)"
        else
            echo "⚠️  状态: 应用包结构不完整"
        fi
        echo ""
    fi
done

echo ""
echo "=== 创建发布包 ==="
RELEASE_DIR="build/release"
mkdir -p "$RELEASE_DIR"

# 创建 ZIP 压缩包 (所有平台统一使用 ZIP 格式)
for file in build/bin/acloud-v${VERSION}-*; do
    if [ -f "$file" ] || [ -d "$file" ]; then
        filename=$(basename "$file")
        echo "📦 正在打包: $filename (ZIP格式)"
        
        if [[ "$file" == *.app ]]; then
            # macOS 应用包打包为 ZIP
            (cd build/bin && zip -r "../release/${filename}.zip" "$(basename "$file")")
        elif [[ "$file" == *linux* ]]; then
            # Linux 可执行文件打包为 ZIP
            (cd build/bin && zip "../release/${filename}.zip" "$(basename "$file")")
        else
            # Windows 和其他可执行文件打包为 ZIP
            (cd build/bin && zip "../release/${filename}.zip" "$(basename "$file")")
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
if [ -d "$RELEASE_DIR" ] && [ "$(ls -A "$RELEASE_DIR")" ]; then
    ls -la "$RELEASE_DIR"
    echo ""
    echo "📊 发布包大小统计:"
    for zip_file in "$RELEASE_DIR"/*.zip; do
        if [ -f "$zip_file" ]; then
            echo "  $(basename "$zip_file"): $(du -h "$zip_file" | cut -f1)"
        fi
    done
else
    echo "⚠️  没有创建发布包"
fi

echo ""
echo "=== 编译摘要 ==="
echo "🎯 项目: ACloud v${VERSION}"
echo "📅 编译时间: $(date '+%Y-%m-%d %H:%M:%S')"
echo "💻 编译环境: $(uname -s) $(uname -m)"
echo "🔧 Wails 版本: $(wails version 2>/dev/null || echo "未知")"

echo ""
echo "🎉 编译完成！"
echo "📁 构建文件: build/bin/"
echo "📦 发布包: build/release/"
