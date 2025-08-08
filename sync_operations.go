package main

import (
	"fmt"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// StartSync 启动同步服务
func (a *App) StartSync() error {
	// 检查用户是否已登录
	if !a.isLoggedIn {
		return fmt.Errorf("用户未登录")
	}

	// 检查 MinIO 是否已启用
	if !a.minioConfig.Enabled || a.minioClient == nil {
		return fmt.Errorf("MinIO 未启用")
	}

	// 如果同步已经在运行，返回错误
	if a.syncRunning {
		return fmt.Errorf("同步服务已在运行")
	}

	// 设置同步状态
	a.syncEnabled = true
	a.syncRunning = true

	// 发送状态更新到前端
	wailsRuntime.EventsEmit(a.ctx, "sync-status-changed", a.syncEnabled)

	// 启动同步服务
	go a.syncService()

	fmt.Println("同步服务已启动")
	return nil
}

// StopSync 停止同步服务
func (a *App) StopSync() error {
	// 检查用户是否已登录
	if !a.isLoggedIn {
		return fmt.Errorf("用户未登录")
	}

	// 如果同步未在运行，返回错误
	if !a.syncRunning {
		return fmt.Errorf("同步服务未在运行")
	}

	// 停止同步服务
	a.syncStopCh <- true

	// 设置同步状态
	a.syncEnabled = false
	a.syncRunning = false

	// 发送状态更新到前端
	wailsRuntime.EventsEmit(a.ctx, "sync-status-changed", a.syncEnabled)

	fmt.Println("同步服务已停止")
	return nil
}

// syncService 同步服务主循环
func (a *App) syncService() {
	// 创建定时器
	ticker := time.NewTicker(a.syncInterval)
	defer ticker.Stop()

	// 立即执行一次同步
	a.performSync()

	// 主循环
	for {
		select {
		case <-ticker.C:
			// 定时执行同步
			if a.syncEnabled {
				a.performSync()
			}
		case <-a.syncStopCh:
			// 收到停止信号
			fmt.Println("同步服务收到停止信号")
			return
		}
	}
}

// performSync 执行同步操作
func (a *App) performSync() {
	fmt.Println("开始执行同步...")
	startTime := time.Now()

	// 创建同步状态
	status := SyncStatus{
		Running:         true,
		LastSync:        time.Now(),
		FilesUploaded:   0,
		FilesDownloaded: 0,
		Errors:          []string{},
		SyncMode:        a.syncMode,
		ConflictCount:   0,
	}

	// 发送同步开始事件
	wailsRuntime.EventsEmit(a.ctx, "sync-started", status)

	// 根据同步模式执行不同的同步策略
	var err error
	switch a.syncMode {
	case "full":
		err = a.fullSync(&status)
	case "selective":
		err = a.selectiveSync(&status)
	case "backup":
		err = a.backupSync(&status)
	case "incremental":
		err = a.incrementalSync(&status)
	default:
		err = fmt.Errorf("未知的同步模式: %s", a.syncMode)
	}

	// 更新同步状态
	status.Running = false
	if err != nil {
		status.Errors = append(status.Errors, err.Error())
	}

	// 记录同步历史
	duration := time.Since(startTime)
	a.recordSyncHistory(status, duration)

	// 保存同步报告
	if err := a.SaveSyncReport(status); err != nil {
		fmt.Printf("保存同步报告失败: %v\n", err)
	}

	// 发送同步完成事件
	wailsRuntime.EventsEmit(a.ctx, "sync-completed", status)

	fmt.Printf("同步完成: 上传 %d 个文件, 下载 %d 个文件, %d 个冲突, %d 个错误\n",
		status.FilesUploaded, status.FilesDownloaded, status.ConflictCount, len(status.Errors))
}

// TriggerManualSync 触发手动同步
func (a *App) TriggerManualSync() error {
	// 检查用户是否已登录
	if !a.isLoggedIn {
		return fmt.Errorf("用户未登录")
	}

	// 检查 MinIO 是否已启用
	if !a.minioConfig.Enabled || a.minioClient == nil {
		return fmt.Errorf("MinIO 未启用")
	}

	// 如果同步已经在运行，返回错误
	if a.syncRunning {
		return fmt.Errorf("同步服务已在运行")
	}

	// 执行同步
	go a.performSync()

	fmt.Println("已触发手动同步")
	return nil
}
