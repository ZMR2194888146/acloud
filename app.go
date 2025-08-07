package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/wailsapp/wails/v2/pkg/options"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App 结构体
type App struct {
	ctx         context.Context
	storagePath string
	users       map[string]User
	currentUser string
	isLoggedIn  bool
	usersFile   string
	configFile  string
	minioConfig MinioConfig
	minioClient *minio.Client
	syncEnabled bool
	syncInterval time.Duration
	syncRunning bool
	syncStopCh  chan bool
	// 群晖Drive风格功能
	syncRules     []SyncRule
	fileVersions  map[string][]FileVersion
	conflictFiles []ConflictFile
	shareLinks    map[string]ShareLink
	syncMode      string // "full", "selective", "backup"
	// iSCSI客户端功能
	iscsiDiscoveredTargets []ISCSIDiscoveredTarget      // 发现的iSCSI目标器列表
	iscsiConnections       map[string]*ISCSIConnection  // 活动连接列表
	iscsiDisks            []ISCSIDisk                   // iSCSI磁盘列表
	iscsiInitiatorConfig  ISCSIInitiatorConfig          // 发起器配置
	iscsiEnabled          bool                          // iSCSI客户端是否启用
	// 客户端特性
	clientFeatures *ClientFeatures                     // 客户端特性管理器
}

// NewApp 创建一个新的 App 应用结构体
func NewApp() *App {
	// 在用户目录下创建存储目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	storagePath := filepath.Join(homeDir, "hkce-cloud-storage")
	
	// 确保存储目录存在
	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		os.MkdirAll(storagePath, 0755)
	}
	
	// 用户文件路径
	usersFile := filepath.Join(storagePath, "users.json")
	
	// 配置文件路径
	configFile := filepath.Join(storagePath, "config.json")
	
	app := &App{
		storagePath: storagePath,
		users:       make(map[string]User),
		isLoggedIn:  false,
		usersFile:   usersFile,
		configFile:  configFile,
		minioConfig: MinioConfig{
			Endpoint:        "play.min.io",
			AccessKeyID:     "minioadmin",
			SecretAccessKey: "minioadmin",
			UseSSL:          true,
			BucketName:      "hkce-cloud",
			Enabled:         false,
		},
		syncEnabled:  false,
		syncInterval: 5 * time.Minute,
		syncRunning:  false,
		syncStopCh:   make(chan bool),
		// 群晖Drive风格功能初始化
		syncRules:     []SyncRule{},
		fileVersions:  make(map[string][]FileVersion),
		conflictFiles: []ConflictFile{},
		shareLinks:    make(map[string]ShareLink),
		syncMode:      "full",
		// iSCSI客户端功能初始化
		iscsiDiscoveredTargets: []ISCSIDiscoveredTarget{},
		iscsiConnections:       make(map[string]*ISCSIConnection),
		iscsiDisks:            []ISCSIDisk{},
		iscsiInitiatorConfig: ISCSIInitiatorConfig{
			InitiatorName:                "iqn.2024-01.com.hkce-cloud:initiator",
			DefaultPort:                  3260,
			LoginTimeout:                 30,
			LogoutTimeout:                15,
			NOPOutInterval:               10,
			NOPOutTimeout:                15,
			MaxConnections:               8,
			HeaderDigest:                 "None",
			DataDigest:                   "None",
			MaxRecvDataSegmentLength:     262144,
		},
		iscsiEnabled: false,
	}
	
	// 初始化客户端特性
	app.clientFeatures = NewClientFeatures(app)
	
	// 加载用户数据
	app.loadUsers()
	
	// 如果没有用户，创建默认用户
	if len(app.users) == 0 {
		// 创建默认管理员用户
		app.Register("admin", "admin")
	}
	
	// 加载配置
	app.loadConfig()
	
	// 初始化 MinIO 客户端
	if app.minioConfig.Enabled {
		app.initMinioClient()
	}
	
	return app
}

// startup 在应用启动时调用。上下文被保存以便我们可以调用运行时方法
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	
	// 初始化客户端特性
	if err := a.clientFeatures.InitializeClientFeatures(ctx); err != nil {
		fmt.Printf("初始化客户端特性失败: %v\n", err)
	}
	
	// 恢复窗口状态
	if err := a.clientFeatures.RestoreWindowState(ctx); err != nil {
		fmt.Printf("恢复窗口状态失败: %v\n", err)
	}
	
	fmt.Println("应用启动完成")
}

// domReady 在DOM准备就绪时调用
func (a *App) domReady(ctx context.Context) {
	// DOM准备就绪后的初始化工作
	fmt.Println("DOM已准备就绪")
}

// beforeClose 在应用关闭前调用
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	// 保存窗口状态
	if err := a.clientFeatures.SaveWindowState(ctx); err != nil {
		fmt.Printf("保存窗口状态失败: %v\n", err)
	}
	
	// 保存应用配置
	if err := a.saveConfig(); err != nil {
		fmt.Printf("保存配置失败: %v\n", err)
	}
	
	// 清理客户端特性资源
	a.clientFeatures.Cleanup()
	
	// 停止同步服务
	if a.syncRunning {
		a.syncStopCh <- true
	}
	
	fmt.Println("应用状态已保存")
	return false // 允许关闭
}

// shutdown 在应用关闭时调用
func (a *App) shutdown(ctx context.Context) {
	fmt.Println("应用正在关闭...")
}

// onSecondInstanceLaunch 当尝试启动第二个实例时调用
func (a *App) onSecondInstanceLaunch(secondInstanceData options.SecondInstanceData) {
	// 显示主窗口并置顶
	wailsRuntime.WindowShow(a.ctx)
	wailsRuntime.WindowSetAlwaysOnTop(a.ctx, true)
	wailsRuntime.WindowSetAlwaysOnTop(a.ctx, false)
	
	// 发送通知
	wailsRuntime.EventsEmit(a.ctx, "second-instance-launched", map[string]interface{}{
		"message": "应用已在运行中",
		"time":    time.Now(),
	})
}

// Greet 返回给定名称的问候语
func (a *App) Greet(name string) string {
	return fmt.Sprintf("你好 %s，欢迎使用网盘！", name)
}

// GetStoragePath 获取存储路径
func (a *App) GetStoragePath() string {
	return a.storagePath
}

// 客户端特性相关API方法

// GetTrayMenuItems 获取系统托盘菜单项
func (a *App) GetTrayMenuItems() []TrayMenuItem {
	return a.clientFeatures.GetTrayMenuItems()
}

// HandleTrayMenuClick 处理托盘菜单点击
func (a *App) HandleTrayMenuClick(menuID string) {
	a.clientFeatures.HandleTrayMenuClick(menuID)
}

// SetAutoStart 设置开机自启动
func (a *App) SetAutoStart(enabled bool) error {
	return a.clientFeatures.SetAutoStart(enabled)
}

// SendNotification 发送系统通知
func (a *App) SendNotification(title, message string) {
	a.clientFeatures.sendNotification(a.ctx, title, message)
}

// GetAppVersion 获取应用版本
func (a *App) GetAppVersion() string {
	return a.clientFeatures.updateChecker.currentVersion
}

// CheckForUpdatesManually 手动检查更新
func (a *App) CheckForUpdatesManually() {
	go a.clientFeatures.checkForUpdates(a.ctx)
}

// ToggleSyncStatus 切换同步状态
func (a *App) ToggleSyncStatus() bool {
	a.syncEnabled = !a.syncEnabled
	
	// 发送状态更新到前端
	wailsRuntime.EventsEmit(a.ctx, "sync-status-changed", a.syncEnabled)
	
	return a.syncEnabled
}

// GetSyncStatus 获取同步状态
func (a *App) GetSyncStatus() map[string]interface{} {
	return map[string]interface{}{
		"enabled": a.syncEnabled,
		"running": a.syncRunning,
		"interval": a.syncInterval.String(),
		"mode": a.syncMode,
	}
}

// MinimizeToTray 最小化到系统托盘
func (a *App) MinimizeToTray() {
	wailsRuntime.WindowHide(a.ctx)
	a.SendNotification("HKCE Cloud", "应用已最小化到系统托盘")
}

// ShowFromTray 从系统托盘显示窗口
func (a *App) ShowFromTray() {
	wailsRuntime.WindowShow(a.ctx)
	wailsRuntime.WindowSetAlwaysOnTop(a.ctx, true)
	wailsRuntime.WindowSetAlwaysOnTop(a.ctx, false)
}

// GetSystemInfo 获取系统信息
func (a *App) GetSystemInfo() map[string]interface{} {
	return map[string]interface{}{
		"os": runtime.GOOS,
		"arch": runtime.GOARCH,
		"version": a.GetAppVersion(),
		"storage_path": a.storagePath,
		"auto_start": false, // 这里可以实际检查自启动状态
	}
}

// 加载配置
func (a *App) loadConfig() {
	// 检查配置文件是否存在
	if _, err := os.Stat(a.configFile); os.IsNotExist(err) {
		// 如果不存在，保存默认配置
		a.saveConfig()
		return
	}
	
	// 读取配置文件
	data, err := os.ReadFile(a.configFile)
	if err != nil {
		return
	}
	
	// 解析配置数据
	var config struct {
		Minio MinioConfig `json:"minio"`
	}
	
	if err := json.Unmarshal(data, &config); err != nil {
		return
	}
	
	// 更新配置
	a.minioConfig = config.Minio
}

// 保存配置
func (a *App) saveConfig() error {
	// 序列化配置数据
	config := struct {
		Minio MinioConfig `json:"minio"`
	}{
		Minio: a.minioConfig,
	}
	
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	
	// 写入配置文件
	return os.WriteFile(a.configFile, data, 0644)
}