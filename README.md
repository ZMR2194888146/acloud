# ACloud

一个基于 Wails 框架开发的现代化桌面云存储应用，提供本地文件管理和云端同步功能。

![ACloud](https://img.shields.io/badge/Version-1.0.0-blue.svg)
![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey.svg)
![License](https://img.shields.io/badge/License-MIT-green.svg)

## ✨ 特性

### 🗂️ 文件管理
- **本地文件管理**：完整的文件和文件夹操作（创建、删除、重命名、移动）
- **多视图模式**：支持列表视图和网格视图
- **文件预览**：内置文件预览功能
- **拖拽上传**：支持拖拽文件上传
- **批量操作**：支持批量文件操作

### ☁️ 云端存储
- **MinIO 集成**：支持 MinIO 对象存储作为云端后端
- **自动同步**：智能文件同步，支持多种同步模式
- **冲突处理**：自动处理文件冲突
- **断点续传**：支持大文件断点续传

### 🔄 同步功能
- **实时同步**：文件变更实时同步到云端
- **选择性同步**：可选择特定文件夹进行同步
- **同步规则**：自定义同步规则和过滤器
- **版本控制**：文件版本历史管理

### ⚙️ 系统设置
- **统一配置**：所有设置集中管理
- **开机自启**：支持开机自动启动
- **系统托盘**：最小化到系统托盘
- **主题切换**：支持明暗主题切换

## 🚀 快速开始

### 环境要求

- **Go**: 1.18 或更高版本
- **Node.js**: 16.0 或更高版本
- **Wails**: v2.0 或更高版本

### 安装依赖

```bash
# 安装 Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 克隆项目
git clone https://github.com/your-username/acloud.git
cd acloud

# 安装前端依赖
cd frontend
npm install
cd ..
```

### 开发模式

```bash
# 启动开发服务器
wails dev
```

### 构建应用

```bash
# 构建生产版本
wails build

# 构建特定平台
wails build -platform windows/amd64
wails build -platform darwin/amd64
wails build -platform linux/amd64
```

## 📁 项目结构

```
acloud/
├── app.go                 # 应用主逻辑
├── main.go               # 程序入口
├── wails.json            # Wails 配置
├── build/                # 构建输出目录
├── frontend/             # 前端代码
│   ├── src/
│   │   ├── components/   # Vue 组件
│   │   ├── assets/       # 静态资源
│   │   ├── App.vue       # 主应用组件
│   │   └── main.js       # 前端入口
│   ├── package.json      # 前端依赖
│   └── vite.config.js    # Vite 配置
└── README.md             # 项目文档
```

## 🔧 配置说明

### MinIO 配置

在系统设置 > 存储设置中配置 MinIO 连接：

```json
{
  "endpoint": "play.min.io",
  "accessKeyId": "your-access-key",
  "secretAccessKey": "your-secret-key",
  "bucketName": "acloud-storage",
  "useSSL": true,
  "enabled": true
}
```

### 同步设置

支持三种同步模式：

- **完全同步**：双向同步所有文件
- **选择性同步**：只同步选定的文件和文件夹
- **备份模式**：只上传到云端，不下载

## 🎨 技术栈

### 后端
- **Go**: 主要编程语言
- **Wails v2**: 桌面应用框架
- **MinIO Go SDK**: 对象存储客户端

### 前端
- **Vue 3**: 前端框架
- **Ant Design Vue**: UI 组件库
- **Vite**: 构建工具
- **JavaScript**: 编程语言

## 📸 界面预览

### 文件管理界面
现代化的文件管理界面，支持列表和网格两种视图模式。

### 同步中心
实时显示同步状态，支持同步规则配置和冲突处理。

### 系统设置
统一的设置管理界面，包含基本设置、同步设置、存储设置等。

## 🔒 安全性

- **加密传输**：支持 HTTPS/TLS 加密传输
- **访问控制**：基于访问密钥的身份验证
- **本地加密**：敏感配置信息本地加密存储

## 🤝 贡献指南

欢迎贡献代码！请遵循以下步骤：

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

### 开发规范

- 遵循 Go 代码规范
- 使用 Vue 3 Composition API
- 保持代码简洁和注释完整
- 添加适当的错误处理

## 📝 更新日志

### v1.0.0 (2024-01-15)
- 🎉 初始版本发布
- ✨ 基本文件管理功能
- ☁️ MinIO 云存储集成
- 🔄 自动同步功能
- ⚙️ 系统设置管理
- 🎨 现代化 UI 设计

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

- [Wails](https://wails.io/) - 优秀的 Go 桌面应用框架
- [Vue.js](https://vuejs.org/) - 渐进式 JavaScript 框架
- [Ant Design Vue](https://antdv.com/) - 企业级 UI 设计语言
- [MinIO](https://min.io/) - 高性能对象存储

## 📞 联系方式

- 项目主页: [GitHub Repository](https://github.com/ZMR2194888146/acloud)
- 问题反馈: [Issues](https://github.com/ZMR2194888146/acloud/issues)
- 邮箱: your-email@example.com

---

<div align="center">
  <p>如果这个项目对你有帮助，请给它一个 ⭐️</p>
  <p>Made with ❤️ by ACloud Team</p>
</div>