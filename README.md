# HKCE Cloud iSCSI 客户端

HKCE Cloud iSCSI 客户端是一个用于连接和管理 iSCSI 存储设备的工具。它支持在 Windows、Linux 和 macOS 上运行，提供了统一的接口来发现、连接、挂载和管理 iSCSI 存储。

## 功能特性

- 跨平台支持：Windows、Linux 和 macOS
- 发现 iSCSI 目标
- 连接和断开 iSCSI 目标
- 扫描 iSCSI 磁盘
- 格式化、挂载和卸载 iSCSI 磁盘
- 获取 iSCSI 性能统计

## 系统要求

### Windows
- Windows 10 或更高版本
- Microsoft iSCSI Initiator 服务已启用
- 管理员权限（用于挂载和格式化磁盘）

### Linux
- 现代 Linux 发行版（如 Ubuntu 20.04+, CentOS 7+）
- open-iscsi 包已安装
- sudo 权限（用于挂载和格式化磁盘）

### macOS
- macOS 10.15 (Catalina) 或更高版本
- 第三方 iSCSI 发起器（如 globalSAN iSCSI Initiator）已安装
- 管理员权限（用于挂载和格式化磁盘）

## 安装

### 从源代码构建

1. 确保已安装 Go 1.16 或更高版本
2. 克隆仓库：
   ```
   git clone https://github.com/hkce-cloud/iscsi-client.git
   cd iscsi-client
   ```
3. 构建项目：
   ```
   go build -o iscsi-client
   ```

## 使用方法

### 运行示例程序

```
go run iscsi_example.go
```

示例程序提供了一个交互式界面，引导您完成以下步骤：

1. 发现 iSCSI 目标
2. 连接到选定的目标
3. 扫描 iSCSI 磁盘
4. 格式化、挂载和卸载磁盘
5. 查看性能统计

### 在自己的项目中使用

```go
package main

import (
	"fmt"
	"log"

	"./iscsi"
)

func main() {
	// 创建 iSCSI 管理器
	manager, err := iscsi.NewISCSIManager()
	if err != nil {
		log.Fatalf("创建 iSCSI 管理器失败: %v", err)
	}

	// 初始化 iSCSI 管理器
	if err := manager.Initialize(); err != nil {
		log.Fatalf("初始化 iSCSI 管理器失败: %v", err)
	}

	// 检查 iSCSI 服务是否可用
	if !manager.IsAvailable() {
		log.Fatalf("iSCSI 服务不可用")
	}

	// 发现目标
	targets, err := manager.DiscoverTargets("192.168.1.100", 3260)
	if err != nil {
		log.Fatalf("发现目标失败: %v", err)
	}

	// 连接到目标
	if len(targets) > 0 {
		connectionID, err := manager.ConnectTarget(targets[0], "", "")
		if err != nil {
			log.Fatalf("连接目标失败: %v", err)
		}
		fmt.Printf("成功连接到目标: %s\n", targets[0].IQN)

		// 断开连接
		if err := manager.DisconnectTarget(connectionID); err != nil {
			log.Fatalf("断开连接失败: %v", err)
		}
	}

	// 清理资源
	if err := manager.Cleanup(); err != nil {
		log.Fatalf("清理资源失败: %v", err)
	}
}
```

## 注意事项

- 在 macOS 上，需要安装第三方 iSCSI 发起器，如 globalSAN iSCSI Initiator
- 在 Linux 上，需要安装 open-iscsi 包
- 在 Windows 上，需要启用 Microsoft iSCSI Initiator 服务
- 挂载和格式化磁盘操作需要管理员/root 权限

## 许可证

MIT