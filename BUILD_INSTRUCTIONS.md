# ACloud 构建说明

## 📋 构建脚本说明

项目提供了两个构建脚本，用于编译带有系统架构名称的可执行文件：

### 1. `build_with_arch.sh` (推荐)
- **用途**: 简洁高效的多平台编译脚本
- **特点**: 生成带有版本号、系统名称和架构的文件名
- **输出格式**: `acloud-v1.0.0-{系统}-{架构}.{扩展名}`

### 2. `build_all_platforms.sh` (完整版)
- **用途**: 完整的构建脚本，包含打包和发布功能
- **特点**: 编译 + 打包 + 创建发布包
- **输出**: 可执行文件 + ZIP 压缩包

## 🚀 快速开始

### 使用推荐脚本
```bash
./build_with_arch.sh
```

### 使用完整脚本
```bash
./build_all_platforms.sh
```

## 📦 输出文件命名规范

编译完成后，文件将按以下格式命名：

### 可执行文件
```
acloud-v1.0.0-windows-amd64.exe     # Windows 64位
acloud-v1.0.0-darwin-arm64.app      # macOS Apple Silicon
acloud-v1.0.0-darwin-amd64.app      # macOS Intel
acloud-v1.0.0-linux-amd64           # Linux 64位
acloud-v1.0.0-linux-arm64           # Linux ARM64
```

### 发布包 (统一ZIP格式)
```
acloud-v1.0.0-windows-amd64.exe.zip # Windows 64位 ZIP包
acloud-v1.0.0-darwin-arm64.app.zip  # macOS Apple Silicon ZIP包
acloud-v1.0.0-darwin-amd64.app.zip  # macOS Intel ZIP包
acloud-v1.0.0-linux-amd64.zip       # Linux 64位 ZIP包
acloud-v1.0.0-linux-arm64.zip       # Linux ARM64 ZIP包
```

> 📝 **重要说明**: 所有平台（Windows、macOS、Linux）的发布包都统一使用 ZIP 格式，便于跨平台分发和管理。

## 🎯 支持的平台和架构

| 平台 | 架构 | 状态 | 说明 |
|------|------|------|------|
| Windows | amd64 | ✅ 支持 | 完全支持交叉编译 |
| macOS | arm64 | ✅ 支持 | Apple Silicon (M1/M2) |
| macOS | amd64 | ✅ 支持 | Intel 处理器 |
| Linux | amd64 | ⚠️ 受限 | 需要在 Linux 环境编译 |
| Linux | arm64 | ⚠️ 受限 | 需要在 Linux 环境编译 |

## 📁 输出目录结构

```
build/
├── bin/                                    # 可执行文件
│   ├── acloud-v1.0.0-windows-amd64.exe
│   ├── acloud-v1.0.0-darwin-arm64.app/
│   ├── acloud-v1.0.0-darwin-amd64.app/
│   ├── acloud-v1.0.0-linux-amd64
│   └── acloud-v1.0.0-linux-arm64
└── release/                                # 发布包 (所有平台统一ZIP格式)
    ├── acloud-v1.0.0-windows-amd64.exe.zip
    ├── acloud-v1.0.0-darwin-arm64.app.zip
    ├── acloud-v1.0.0-darwin-amd64.app.zip
    ├── acloud-v1.0.0-linux-amd64.zip
    └── acloud-v1.0.0-linux-arm64.zip
```

## 🔧 手动编译单个平台

如果只需要编译特定平台，可以使用以下命令：

```bash
# Windows
wails build -platform windows/amd64

# macOS ARM64
wails build -platform darwin/arm64

# macOS Intel
wails build -platform darwin/amd64

# Linux (需要在 Linux 环境中执行)
wails build -platform linux/amd64
wails build -platform linux/arm64
```

## ⚠️ 注意事项

1. **Linux 编译限制**: 
   - Wails 在 macOS 上不支持 Linux 交叉编译
   - 需要在 Linux 环境中编译 Linux 版本

2. **macOS 应用签名**:
   - 编译的 macOS 应用未签名
   - 分发前可能需要进行代码签名

3. **依赖检查**:
   - 确保已安装 Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
   - 确保前端依赖已安装: `cd frontend && npm install`

## 🐛 故障排除

### 编译失败
```bash
# 清理缓存
go clean -cache
go clean -modcache

# 重新安装前端依赖
cd frontend
rm -rf node_modules package-lock.json
npm install
cd ..

# 重新编译
./build_with_arch.sh
```

### macOS 编译卡住
```bash
# 终止现有编译进程
pkill -f "wails build"

# 清理构建目录
rm -rf build/bin/*

# 重新编译
./build_with_arch.sh
```

## 📊 版本管理

要更改版本号，编辑构建脚本中的 `VERSION` 变量：

```bash
# 在 build_with_arch.sh 中
VERSION="1.0.0"  # 修改为你的版本号
```

## 🎉 完成

编译完成后，你将获得：
- 带有明确系统和架构标识的可执行文件
- 便于分发和管理的文件命名
- 支持多平台的应用程序包

---
*更新时间: 2025/8/9*