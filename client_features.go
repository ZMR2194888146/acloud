package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// ClientFeatures 客户端特性管理器
type ClientFeatures struct {
	app           *App
	fileWatcher   *fsnotify.Watcher
	updateChecker *UpdateChecker
	trayManager   *TrayManager
	notifier      *NotificationManager
}

// UpdateChecker 更新检查器
type UpdateChecker struct {
	currentVersion string
	updateURL      string
	checkInterval  time.Duration
	lastCheck      time.Time
}

// TrayManager 系统托盘管理器
type TrayManager struct {
	isVisible bool
	menuItems []TrayMenuItem
}

// TrayMenuItem 托盘菜单项
type TrayMenuItem struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Type     string `json:"type"` // "normal", "separator", "checkbox"
	Enabled  bool   `json:"enabled"`
	Checked  bool   `json:"checked"`
	Shortcut string `json:"shortcut"`
}

// NotificationManager 通知管理器
type NotificationManager struct {
	enabled bool
}

// UpdateInfo 更新信息
type UpdateInfo struct {
	Version     string `json:"version"`
	Description string `json:"description"`
	DownloadURL string `json:"download_url"`
	ReleaseDate string `json:"release_date"`
	Mandatory   bool   `json:"mandatory"`
}

// NewClientFeatures 创建客户端特性管理器
func NewClientFeatures(app *App) *ClientFeatures {
	return &ClientFeatures{
		app: app,
		updateChecker: &UpdateChecker{
			currentVersion: "1.0.0",
			updateURL:      "https://api.github.com/repos/your-repo/releases/latest",
			checkInterval:  24 * time.Hour,
		},
		trayManager: &TrayManager{
			isVisible: false,
			menuItems: []TrayMenuItem{
				{ID: "show", Label: "显示主窗口", Type: "normal", Enabled: true},
				{ID: "separator1", Type: "separator"},
				{ID: "sync", Label: "开始同步", Type: "checkbox", Enabled: true},
				{ID: "settings", Label: "设置", Type: "normal", Enabled: true},
				{ID: "separator2", Type: "separator"},
				{ID: "quit", Label: "退出", Type: "normal", Enabled: true},
			},
		},
		notifier: &NotificationManager{
			enabled: true,
		},
	}
}

// InitializeClientFeatures 初始化客户端特性
func (cf *ClientFeatures) InitializeClientFeatures(ctx context.Context) error {
	// 初始化文件监控
	if err := cf.initFileWatcher(); err != nil {
		log.Printf("初始化文件监控失败: %v", err)
	}

	// 初始化系统托盘
	if err := cf.initSystemTray(ctx); err != nil {
		log.Printf("初始化系统托盘失败: %v", err)
	}

	// 启动更新检查
	go cf.startUpdateChecker(ctx)

	// 启动文件监控服务
	go cf.startFileWatcherService()

	return nil
}

// initFileWatcher 初始化文件监控器
func (cf *ClientFeatures) initFileWatcher() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("创建文件监控器失败: %v", err)
	}

	cf.fileWatcher = watcher

	// 添加监控目录
	err = watcher.Add(cf.app.storagePath)
	if err != nil {
		return fmt.Errorf("添加监控目录失败: %v", err)
	}

	return nil
}

// initSystemTray 初始化系统托盘
func (cf *ClientFeatures) initSystemTray(ctx context.Context) error {
	// 由于Wails v2对系统托盘的支持有限，这里提供基础实现
	// 在实际项目中可能需要使用第三方库如 systray
	
	cf.trayManager.isVisible = true
	
	// 设置托盘图标和菜单
	// 这里是模拟实现，实际需要根据平台特定的API
	log.Println("系统托盘已初始化")
	
	return nil
}

// startUpdateChecker 启动更新检查器
func (cf *ClientFeatures) startUpdateChecker(ctx context.Context) {
	ticker := time.NewTicker(cf.updateChecker.checkInterval)
	defer ticker.Stop()

	// 立即检查一次更新
	cf.checkForUpdates(ctx)

	for {
		select {
		case <-ticker.C:
			cf.checkForUpdates(ctx)
		case <-ctx.Done():
			return
		}
	}
}

// checkForUpdates 检查更新
func (cf *ClientFeatures) checkForUpdates(ctx context.Context) {
	cf.updateChecker.lastCheck = time.Now()

	// 发送HTTP请求检查更新
	resp, err := http.Get(cf.updateChecker.updateURL)
	if err != nil {
		log.Printf("检查更新失败: %v", err)
		return
	}
	defer resp.Body.Close()

	var updateInfo UpdateInfo
	if err := json.NewDecoder(resp.Body).Decode(&updateInfo); err != nil {
		log.Printf("解析更新信息失败: %v", err)
		return
	}

	// 比较版本号
	if cf.isNewerVersion(updateInfo.Version, cf.updateChecker.currentVersion) {
		// 发现新版本，通知用户
		cf.notifyUpdateAvailable(ctx, updateInfo)
	}
}

// isNewerVersion 比较版本号
func (cf *ClientFeatures) isNewerVersion(newVersion, currentVersion string) bool {
	// 简单的版本比较实现
	// 实际项目中应该使用更完善的版本比较库
	return strings.Compare(newVersion, currentVersion) > 0
}

// notifyUpdateAvailable 通知有可用更新
func (cf *ClientFeatures) notifyUpdateAvailable(ctx context.Context, updateInfo UpdateInfo) {
	// 发送系统通知
	cf.sendNotification(ctx, "更新可用", fmt.Sprintf("发现新版本 %s", updateInfo.Version))

	// 在应用中显示更新对话框
	wailsRuntime.EventsEmit(ctx, "update-available", updateInfo)
}

// startFileWatcherService 启动文件监控服务
func (cf *ClientFeatures) startFileWatcherService() {
	if cf.fileWatcher == nil {
		return
	}

	for {
		select {
		case event, ok := <-cf.fileWatcher.Events:
			if !ok {
				return
			}
			cf.handleFileEvent(event)

		case err, ok := <-cf.fileWatcher.Errors:
			if !ok {
				return
			}
			log.Printf("文件监控错误: %v", err)
		}
	}
}

// handleFileEvent 处理文件事件
func (cf *ClientFeatures) handleFileEvent(event fsnotify.Event) {
	log.Printf("文件事件: %s %s", event.Op.String(), event.Name)

	// 根据事件类型处理
	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		cf.handleFileCreated(event.Name)
	case event.Op&fsnotify.Write == fsnotify.Write:
		cf.handleFileModified(event.Name)
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		cf.handleFileDeleted(event.Name)
	case event.Op&fsnotify.Rename == fsnotify.Rename:
		cf.handleFileRenamed(event.Name)
	}

	// 触发同步（如果启用）
	if cf.app.minioConfig.Enabled && cf.app.syncEnabled {
		go cf.triggerSync(event.Name)
	}

	// 发送事件到前端
	wailsRuntime.EventsEmit(cf.app.ctx, "file-changed", map[string]interface{}{
		"type": event.Op.String(),
		"path": event.Name,
		"time": time.Now(),
	})
}

// handleFileCreated 处理文件创建事件
func (cf *ClientFeatures) handleFileCreated(path string) {
	log.Printf("文件已创建: %s", path)
}

// handleFileModified 处理文件修改事件
func (cf *ClientFeatures) handleFileModified(path string) {
	log.Printf("文件已修改: %s", path)
}

// handleFileDeleted 处理文件删除事件
func (cf *ClientFeatures) handleFileDeleted(path string) {
	log.Printf("文件已删除: %s", path)
}

// handleFileRenamed 处理文件重命名事件
func (cf *ClientFeatures) handleFileRenamed(path string) {
	log.Printf("文件已重命名: %s", path)
}

// triggerSync 触发同步
func (cf *ClientFeatures) triggerSync(path string) {
	// 实现文件同步逻辑
	log.Printf("触发同步: %s", path)
}

// sendNotification 发送系统通知
func (cf *ClientFeatures) sendNotification(ctx context.Context, title, message string) {
	if !cf.notifier.enabled {
		return
	}

	// 根据操作系统发送通知
	switch runtime.GOOS {
	case "windows":
		cf.sendWindowsNotification(title, message)
	case "darwin":
		cf.sendMacOSNotification(title, message)
	case "linux":
		cf.sendLinuxNotification(title, message)
	}

	// 同时发送到前端
	wailsRuntime.EventsEmit(ctx, "notification", map[string]string{
		"title":   title,
		"message": message,
		"time":    time.Now().Format("15:04:05"),
	})
}

// sendWindowsNotification 发送Windows通知
func (cf *ClientFeatures) sendWindowsNotification(title, message string) {
	// Windows通知实现
	log.Printf("Windows通知: %s - %s", title, message)
}

// sendMacOSNotification 发送macOS通知
func (cf *ClientFeatures) sendMacOSNotification(title, message string) {
	// macOS通知实现
	log.Printf("macOS通知: %s - %s", title, message)
}

// sendLinuxNotification 发送Linux通知
func (cf *ClientFeatures) sendLinuxNotification(title, message string) {
	// Linux通知实现
	log.Printf("Linux通知: %s - %s", title, message)
}

// HandleTrayMenuClick 处理托盘菜单点击
func (cf *ClientFeatures) HandleTrayMenuClick(menuID string) {
	switch menuID {
	case "show":
		wailsRuntime.WindowShow(cf.app.ctx)
		wailsRuntime.WindowSetAlwaysOnTop(cf.app.ctx, true)
		wailsRuntime.WindowSetAlwaysOnTop(cf.app.ctx, false)
	case "sync":
		cf.toggleSync()
	case "settings":
		cf.showSettings()
	case "quit":
		wailsRuntime.Quit(cf.app.ctx)
	}
}

// toggleSync 切换同步状态
func (cf *ClientFeatures) toggleSync() {
	cf.app.syncEnabled = !cf.app.syncEnabled
	
	// 更新托盘菜单状态
	for i := range cf.trayManager.menuItems {
		if cf.trayManager.menuItems[i].ID == "sync" {
			cf.trayManager.menuItems[i].Checked = cf.app.syncEnabled
			if cf.app.syncEnabled {
				cf.trayManager.menuItems[i].Label = "停止同步"
			} else {
				cf.trayManager.menuItems[i].Label = "开始同步"
			}
			break
		}
	}

	// 发送状态更新到前端
	wailsRuntime.EventsEmit(cf.app.ctx, "sync-status-changed", cf.app.syncEnabled)
}

// showSettings 显示设置窗口
func (cf *ClientFeatures) showSettings() {
	wailsRuntime.EventsEmit(cf.app.ctx, "show-settings")
}

// GetTrayMenuItems 获取托盘菜单项
func (cf *ClientFeatures) GetTrayMenuItems() []TrayMenuItem {
	return cf.trayManager.menuItems
}

// SetAutoStart 设置开机自启动
func (cf *ClientFeatures) SetAutoStart(enabled bool) error {
	switch runtime.GOOS {
	case "windows":
		return cf.setWindowsAutoStart(enabled)
	case "darwin":
		return cf.setMacOSAutoStart(enabled)
	case "linux":
		return cf.setLinuxAutoStart(enabled)
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
}

// setWindowsAutoStart 设置Windows开机自启动
func (cf *ClientFeatures) setWindowsAutoStart(enabled bool) error {
	// Windows注册表操作实现
	log.Printf("设置Windows开机自启动: %v", enabled)
	return nil
}

// setMacOSAutoStart 设置macOS开机自启动
func (cf *ClientFeatures) setMacOSAutoStart(enabled bool) error {
	// macOS LaunchAgent实现
	log.Printf("设置macOS开机自启动: %v", enabled)
	return nil
}

// setLinuxAutoStart 设置Linux开机自启动
func (cf *ClientFeatures) setLinuxAutoStart(enabled bool) error {
	// Linux .desktop文件实现
	log.Printf("设置Linux开机自启动: %v", enabled)
	return nil
}

// SaveWindowState 保存窗口状态
func (cf *ClientFeatures) SaveWindowState(ctx context.Context) error {
	// 由于Wails v2的窗口API限制，这里使用简化的实现
	// 在实际项目中可能需要使用其他方式获取窗口状态
	
	// 保存默认窗口状态
	windowState := map[string]interface{}{
		"x":      100,
		"y":      100,
		"width":  1400,
		"height": 900,
		"time":   time.Now(),
	}

	stateFile := filepath.Join(cf.app.storagePath, "window_state.json")
	data, err := json.MarshalIndent(windowState, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(stateFile, data, 0644)
}

// RestoreWindowState 恢复窗口状态
func (cf *ClientFeatures) RestoreWindowState(ctx context.Context) error {
	stateFile := filepath.Join(cf.app.storagePath, "window_state.json")
	
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		return nil // 文件不存在，使用默认状态
	}

	data, err := os.ReadFile(stateFile)
	if err != nil {
		return err
	}

	var windowState map[string]interface{}
	if err := json.Unmarshal(data, &windowState); err != nil {
		return err
	}

	// 由于Wails v2的窗口API限制，这里使用简化的实现
	// 在实际项目中可能需要使用其他方式设置窗口状态
	log.Printf("窗口状态已加载: %+v", windowState)

	return nil
}

// Cleanup 清理客户端特性资源
func (cf *ClientFeatures) Cleanup() {
	if cf.fileWatcher != nil {
		cf.fileWatcher.Close()
	}
	
	log.Println("客户端特性资源已清理")
}