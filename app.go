package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/wailsapp/wails/v2/pkg/options"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App 结构体
type App struct {
	ctx          context.Context
	storagePath  string
	users        map[string]User
	currentUser  string
	isLoggedIn   bool
	usersFile    string
	configFile   string
	minioConfig  MinioConfig
	minioClient  *minio.Client
	syncEnabled  bool
	syncInterval time.Duration
	syncRunning  bool
	syncStopCh   chan bool
	// 群晖Drive风格功能
	syncRules                 []SyncRule
	fileVersions              map[string][]FileVersion
	conflictFiles             []ConflictFile
	shareLinks                map[string]ShareLink
	syncMode                  string // "full", "selective", "backup", "incremental"
	lastSyncTime              time.Time
	configDir                 string
	jsonParser                *JSONParser
	defaultConflictResolution string
	config                    Config
	// 客户端特性
	clientFeatures *ClientFeatures // 客户端特性管理器
}

// JSONParser 是一个JSON解析器包装器
type JSONParser struct{}

// Marshal 将对象序列化为JSON
func (j *JSONParser) Marshal(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

// Unmarshal 将JSON反序列化为对象
func (j *JSONParser) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// NewApp 创建一个新的 App 应用结构体
func NewApp() *App {
	// 在用户目录下创建存储目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	storagePath := filepath.Join(homeDir, "acloud-storage")

	// 确保存储目录存在
	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		os.MkdirAll(storagePath, 0755)
	}

	// 用户文件路径
	usersFile := filepath.Join(storagePath, "users.json")

	// 配置文件路径
	configFile := filepath.Join(storagePath, "config.json")

	// 配置目录
	configDir := filepath.Join(storagePath, "config")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		os.MkdirAll(configDir, 0755)
	}

	app := &App{
		storagePath: storagePath,
		configDir:   configDir,
		users:       make(map[string]User),
		isLoggedIn:  false,
		usersFile:   usersFile,
		configFile:  configFile,
		minioConfig: MinioConfig{
			Endpoint:        "play.min.io",
			AccessKeyID:     "minioadmin",
			SecretAccessKey: "minioadmin",
			UseSSL:          true,
			BucketName:      "acloud-storage",
			Enabled:         false,
		},
		syncEnabled:  false,
		syncInterval: 5 * time.Minute,
		syncRunning:  false,
		syncStopCh:   make(chan bool),
		// 群晖Drive风格功能初始化
		syncRules:                 []SyncRule{},
		fileVersions:              make(map[string][]FileVersion),
		conflictFiles:             []ConflictFile{},
		shareLinks:                make(map[string]ShareLink),
		syncMode:                  "full",
		lastSyncTime:              time.Now().Add(-24 * time.Hour),
		defaultConflictResolution: "ask", // 默认冲突解决方式：询问用户
		jsonParser:                &JSONParser{},
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

	// 初始化同步功能
	if err := a.initSyncFeatures(); err != nil {
		fmt.Printf("初始化同步功能失败: %v\n", err)
	}

	fmt.Println("应用启动完成")
}

// initSyncFeatures 初始化同步功能
func (a *App) initSyncFeatures() error {
	// 加载同步规则
	if err := a.LoadSyncRules(); err != nil {
		return fmt.Errorf("加载同步规则失败: %v", err)
	}

	// 加载同步历史
	history, err := a.loadSyncHistory()
	if err != nil {
		fmt.Printf("加载同步历史记录失败: %v，将创建新的历史记录\n", err)
	} else {
		// 如果有历史记录，更新最后同步时间
		if len(history.Entries) > 0 {
			a.lastSyncTime = history.Entries[len(history.Entries)-1].Timestamp
		}
	}

	// 如果同步功能已启用，启动同步服务
	if a.syncEnabled && a.isLoggedIn && a.minioConfig.Enabled {
		go func() {
			// 延迟几秒启动，确保应用完全初始化
			time.Sleep(3 * time.Second)
			if err := a.StartSync(); err != nil {
				fmt.Printf("启动同步服务失败: %v\n", err)
			}
		}()
	}

	// 注册同步状态监听器
	wailsRuntime.EventsOn(a.ctx, "sync-status-request", func(optionalData ...interface{}) {
		wailsRuntime.EventsEmit(a.ctx, "sync-status-update", a.GetSyncStatus())
	})

	return nil
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
		// 等待同步服务停止
		time.Sleep(500 * time.Millisecond)
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

	// 更新配置
	a.config.SyncConfig.Enabled = a.syncEnabled
	a.SaveConfig()

	// 如果启用同步，尝试启动同步服务
	if a.syncEnabled && a.isLoggedIn && a.minioConfig.Enabled && !a.syncRunning {
		go func() {
			if err := a.StartSync(); err != nil {
				fmt.Printf("启动同步服务失败: %v\n", err)
				a.SendNotification("同步服务", fmt.Sprintf("启动同步服务失败: %v", err))
			}
		}()
	} else if !a.syncEnabled && a.syncRunning {
		// 如果禁用同步，停止同步服务
		go func() {
			if err := a.StopSync(); err != nil {
				fmt.Printf("停止同步服务失败: %v\n", err)
				a.SendNotification("同步服务", fmt.Sprintf("停止同步服务失败: %v", err))
			}
		}()
	}

	// 发送状态更新到前端
	wailsRuntime.EventsEmit(a.ctx, "sync-status-changed", a.syncEnabled)

	return a.syncEnabled
}

// GetSyncStatus 获取同步状态
func (a *App) GetSyncStatus() SyncStatus {
	return SyncStatus{
		Running:         a.syncRunning,
		LastSync:        a.lastSyncTime,
		FilesUploaded:   0, // 这里可以从最近一次同步记录中获取
		FilesDownloaded: 0, // 这里可以从最近一次同步记录中获取
		Errors:          []string{},
		SyncMode:        a.syncMode,
		ConflictCount:   len(a.conflictFiles),
	}
}

// MinimizeToTray 最小化到系统托盘
func (a *App) MinimizeToTray() {
	wailsRuntime.WindowHide(a.ctx)
	a.SendNotification("ACloud", "应用已最小化到系统托盘")
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
		"os":           runtime.GOOS,
		"arch":         runtime.GOARCH,
		"version":      a.GetAppVersion(),
		"storage_path": a.storagePath,
		"auto_start":   false, // 这里可以实际检查自启动状态
	}
}

// Config 应用配置结构体
type Config struct {
	Minio      MinioConfig `json:"minio"`
	SyncConfig struct {
		Enabled                   bool   `json:"enabled"`
		Interval                  int    `json:"interval"` // 秒
		Mode                      string `json:"mode"`
		DefaultConflictResolution string `json:"defaultConflictResolution"`
	} `json:"sync"`
	ISCSIConfig struct {
		Enabled         bool                 `json:"enabled"`
		InitiatorName   string               `json:"initiatorName"`
		InitiatorConfig ISCSIInitiatorConfig `json:"initiatorConfig"`
	} `json:"iscsi"`
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
	var config Config

	if err := json.Unmarshal(data, &config); err != nil {
		return
	}

	// 保存配置对象
	a.config = config

	// 更新配置
	a.minioConfig = config.Minio
	a.syncEnabled = config.SyncConfig.Enabled
	a.syncInterval = time.Duration(config.SyncConfig.Interval) * time.Second
	a.syncMode = config.SyncConfig.Mode
	a.defaultConflictResolution = config.SyncConfig.DefaultConflictResolution

	// 如果同步间隔太短，设置为默认值
	if a.syncInterval < time.Minute {
		a.syncInterval = 5 * time.Minute
	}

	// 如果同步模式无效，设置为默认值
	if a.syncMode == "" || (a.syncMode != "full" && a.syncMode != "selective" && a.syncMode != "backup" && a.syncMode != "incremental") {
		a.syncMode = "full"
	}

	// 如果冲突解决方式无效，设置为默认值
	if a.defaultConflictResolution == "" || (a.defaultConflictResolution != "local" && a.defaultConflictResolution != "remote" && a.defaultConflictResolution != "both" && a.defaultConflictResolution != "skip" && a.defaultConflictResolution != "ask") {
		a.defaultConflictResolution = "ask"
	}
}

// 保存配置
func (a *App) saveConfig() error {
	// 创建配置对象
	config := Config{
		Minio: a.minioConfig,
	}

	// 同步配置
	config.SyncConfig.Enabled = a.syncEnabled
	config.SyncConfig.Interval = int(a.syncInterval.Seconds())
	config.SyncConfig.Mode = a.syncMode
	config.SyncConfig.DefaultConflictResolution = a.defaultConflictResolution

	// 更新配置对象
	a.config = config

	// 序列化为JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// 写入配置文件
	return os.WriteFile(a.configFile, data, 0644)
}

// SaveConfig 保存配置的公共方法
func (a *App) SaveConfig() error {
	return a.saveConfig()
}

// initMinioClient 初始化MinIO客户端
func (a *App) initMinioClient() error {
	// 创建MinIO客户端
	client, err := minio.New(a.minioConfig.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(a.minioConfig.AccessKeyID, a.minioConfig.SecretAccessKey, ""),
		Secure: a.minioConfig.UseSSL,
	})
	if err != nil {
		return fmt.Errorf("创建MinIO客户端失败: %v", err)
	}

	// 检查存储桶是否存在
	exists, err := client.BucketExists(context.Background(), a.minioConfig.BucketName)
	if err != nil {
		return fmt.Errorf("检查存储桶失败: %v", err)
	}

	// 如果存储桶不存在，创建它
	if !exists {
		err = client.MakeBucket(context.Background(), a.minioConfig.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("创建存储桶失败: %v", err)
		}
	}

	// 保存客户端
	a.minioClient = client

	return nil
}

// GetMinioFileInfo 获取MinIO文件信息
func (a *App) GetMinioFileInfo(path string) (MinioFileInfo, error) {
	if a.minioClient == nil {
		return MinioFileInfo{}, fmt.Errorf("MinIO客户端未初始化")
	}

	// 获取对象信息
	info, err := a.minioClient.StatObject(context.Background(), a.minioConfig.BucketName, path, minio.StatObjectOptions{})
	if err != nil {
		return MinioFileInfo{}, err
	}

	// 转换为MinioFileInfo
	return MinioFileInfo{
		Name:         filepath.Base(path),
		Path:         path,
		Size:         info.Size,
		LastModified: info.LastModified,
		IsDir:        false,
	}, nil
}

// UploadFileToMinio 上传文件到MinIO
func (a *App) UploadFileToMinio(localPath, remotePath string) error {
	if a.minioClient == nil {
		return fmt.Errorf("MinIO客户端未初始化")
	}

	// 打开文件
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %v", err)
	}

	// 上传文件
	_, err = a.minioClient.PutObject(context.Background(), a.minioConfig.BucketName, remotePath, file, fileInfo.Size(), minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return fmt.Errorf("上传文件失败: %v", err)
	}

	return nil
}

// DownloadFileFromMinio 从MinIO下载文件
func (a *App) DownloadFileFromMinio(remotePath string) ([]byte, error) {
	if a.minioClient == nil {
		return nil, fmt.Errorf("MinIO客户端未初始化")
	}

	// 获取对象
	obj, err := a.minioClient.GetObject(context.Background(), a.minioConfig.BucketName, remotePath, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取对象失败: %v", err)
	}
	defer obj.Close()

	// 读取对象内容
	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("读取对象内容失败: %v", err)
	}

	return data, nil
}

// CreateMinioFolder 在MinIO中创建文件夹
func (a *App) CreateMinioFolder(path string) error {
	if a.minioClient == nil {
		return fmt.Errorf("MinIO客户端未初始化")
	}

	// 确保路径以斜杠结尾
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	// 创建空对象作为文件夹
	_, err := a.minioClient.PutObject(context.Background(), a.minioConfig.BucketName, path, nil, 0, minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("创建文件夹失败: %v", err)
	}

	return nil
}

// ListMinioFiles 列出MinIO中的文件
func (a *App) ListMinioFiles(path string) ([]MinioFileInfo, error) {
	if a.minioClient == nil {
		return nil, fmt.Errorf("MinIO客户端未初始化")
	}

	// 确保路径以斜杠结尾
	if path != "" && !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	// 列出对象
	var files []MinioFileInfo

	// 创建通道接收对象信息
	objectCh := a.minioClient.ListObjects(context.Background(), a.minioConfig.BucketName, minio.ListObjectsOptions{
		Prefix:    path,
		Recursive: true,
	})

	// 遍历对象
	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("列出对象失败: %v", object.Err)
		}

		// 跳过当前目录
		if object.Key == path {
			continue
		}

		// 判断是否是目录
		isDir := strings.HasSuffix(object.Key, "/")

		// 添加到文件列表
		files = append(files, MinioFileInfo{
			Name:         filepath.Base(object.Key),
			Path:         object.Key,
			Size:         object.Size,
			LastModified: object.LastModified,
			IsDir:        isDir,
		})
	}

	return files, nil
}



// 同步进度监控功能
type SyncProgress struct {
	TotalFiles      int     `json:"totalFiles"`
	ProcessedFiles  int     `json:"processedFiles"`
	UploadedFiles   int     `json:"uploadedFiles"`
	DownloadedFiles int     `json:"downloadedFiles"`
	CurrentFile     string  `json:"currentFile"`
	Progress        float64 `json:"progress"`
	Status          string  `json:"status"` // "running", "paused", "completed", "error"
	Error           string  `json:"error"`
}

// 同步进度监控
var syncProgress SyncProgress

// UpdateSyncProgress 更新同步进度
func (a *App) UpdateSyncProgress(progress SyncProgress) {
	syncProgress = progress

	// 发送进度更新到前端
	wailsRuntime.EventsEmit(a.ctx, "sync-progress-update", progress)
}

// GetSyncProgress 获取同步进度
func (a *App) GetSyncProgress() SyncProgress {
	return syncProgress
}

// ResetSyncProgress 重置同步进度
func (a *App) ResetSyncProgress() {
	syncProgress = SyncProgress{
		TotalFiles:      0,
		ProcessedFiles:  0,
		UploadedFiles:   0,
		DownloadedFiles: 0,
		CurrentFile:     "",
		Progress:        0,
		Status:          "idle",
		Error:           "",
	}

	// 发送进度更新到前端
	wailsRuntime.EventsEmit(a.ctx, "sync-progress-update", syncProgress)
}

// 同步日志记录功能
type SyncLogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"` // "info", "warning", "error"
	Message   string    `json:"message"`
	File      string    `json:"file"`
}

// 同步日志
var syncLogs []SyncLogEntry

// LogSyncEvent 记录同步事件
func (a *App) LogSyncEvent(level, message, file string) {
	// 创建日志条目
	entry := SyncLogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		File:      file,
	}

	// 添加到日志
	syncLogs = append(syncLogs, entry)

	// 限制日志数量
	if len(syncLogs) > 1000 {
		syncLogs = syncLogs[len(syncLogs)-1000:]
	}

	// 发送日志更新到前端
	wailsRuntime.EventsEmit(a.ctx, "sync-log-update", entry)

	// 如果是错误，发送通知
	if level == "error" {
		a.SendNotification("同步错误", message)
	}
}

// GetSyncLogs 获取同步日志
func (a *App) GetSyncLogs() []SyncLogEntry {
	return syncLogs
}

// ClearSyncLogs 清除同步日志
func (a *App) ClearSyncLogs() {
	syncLogs = []SyncLogEntry{}

	// 发送日志更新到前端
	wailsRuntime.EventsEmit(a.ctx, "sync-logs-cleared", nil)
}

// FileVersion 文件版本
type FileVersion struct {
	Path      string    `json:"path"`
	Version   int       `json:"version"`
	Size      int64     `json:"size"`
	ModTime   time.Time `json:"modTime"`
	Checksum  string    `json:"checksum"`
	CreatedBy string    `json:"createdBy"`
}

// ShareLink 分享链接
type ShareLink struct {
	ID        string    `json:"id"`
	Path      string    `json:"path"`
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expiresAt"`
	Password  string    `json:"password"`
	Views     int       `json:"views"`
	MaxViews  int       `json:"maxViews"`
	CreatedAt time.Time `json:"createdAt"`
	CreatedBy string    `json:"createdBy"`
}

// 同步状态监控功能
func (a *App) MonitorSyncStatus() {
	// 定期检查同步状态
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 如果同步服务正在运行，检查状态
			if a.syncRunning {
				// 发送状态更新到前端
				wailsRuntime.EventsEmit(a.ctx, "sync-status-update", a.GetSyncStatus())
			}
		case <-a.ctx.Done():
			// 上下文取消，退出监控
			return
		}
	}
}

// 同步通知功能
func (a *App) SendSyncNotification(title, message string) {
	// 发送系统通知
	a.SendNotification(title, message)

	// 发送通知事件到前端
	wailsRuntime.EventsEmit(a.ctx, "sync-notification", map[string]string{
		"title":   title,
		"message": message,
		"time":    time.Now().Format("2006-01-02 15:04:05"),
	})
}

// 同步错误处理功能
func (a *App) HandleSyncError(err error, operation string, file string) {
	// 记录错误
	a.LogSyncEvent("error", fmt.Sprintf("%s: %v", operation, err), file)

	// 发送错误通知
	a.SendSyncNotification("同步错误", fmt.Sprintf("%s失败: %v", operation, err))

	// 更新同步进度
	progress := a.GetSyncProgress()
	progress.Error = err.Error()
	progress.Status = "error"
	a.UpdateSyncProgress(progress)
}

// 同步性能优化功能
func (a *App) OptimizeSyncPerformance() {
	// 根据系统资源调整同步参数
	cpuCount := runtime.NumCPU()

	// 设置并发数
	concurrency := cpuCount
	if concurrency > 4 {
		concurrency = 4 // 最大并发数限制为4
	}

	// 设置缓冲区大小
	bufferSize := 1024 * 1024 // 1MB

	// 记录优化信息
	a.LogSyncEvent("info", fmt.Sprintf("同步性能优化: 并发数=%d, 缓冲区大小=%dKB", concurrency, bufferSize/1024), "")
}

// 同步状态重置功能
func (a *App) ResetSyncState() {
	// 重置同步状态
	a.syncRunning = false
	a.lastSyncTime = time.Now()
	a.conflictFiles = []ConflictFile{}

	// 重置同步进度
	a.ResetSyncProgress()

	// 发送状态更新到前端
	wailsRuntime.EventsEmit(a.ctx, "sync-status-update", a.GetSyncStatus())

	// 记录日志
	a.LogSyncEvent("info", "同步状态已重置", "")
}

// 同步配置管理功能
func (a *App) UpdateSyncConfig(enabled bool, interval int, mode string, defaultConflictResolution string) error {
	// 验证参数
	if interval < 10 {
		return fmt.Errorf("同步间隔不能小于10秒")
	}

	validModes := []string{"full", "selective", "backup", "incremental"}
	modeValid := false
	for _, m := range validModes {
		if m == mode {
			modeValid = true
			break
		}
	}
	if !modeValid {
		return fmt.Errorf("无效的同步模式: %s", mode)
	}

	validResolutions := []string{"local", "remote", "both", "skip", "ask"}
	resolutionValid := false
	for _, r := range validResolutions {
		if r == defaultConflictResolution {
			resolutionValid = true
			break
		}
	}
	if !resolutionValid {
		return fmt.Errorf("无效的冲突解决方式: %s", defaultConflictResolution)
	}

	// 更新配置
	a.syncEnabled = enabled
	a.syncInterval = time.Duration(interval) * time.Second
	a.syncMode = mode
	a.defaultConflictResolution = defaultConflictResolution

	// 更新配置对象
	a.config.SyncConfig.Enabled = enabled
	a.config.SyncConfig.Interval = interval
	a.config.SyncConfig.Mode = mode
	a.config.SyncConfig.DefaultConflictResolution = defaultConflictResolution

	// 保存配置
	if err := a.SaveConfig(); err != nil {
		return fmt.Errorf("保存配置失败: %v", err)
	}

	// 如果同步服务正在运行，重启它以应用新配置
	if a.syncRunning {
		if err := a.StopSync(); err != nil {
			return fmt.Errorf("停止同步服务失败: %v", err)
		}

		if a.syncEnabled {
			if err := a.StartSync(); err != nil {
				return fmt.Errorf("启动同步服务失败: %v", err)
			}
		}
	} else if a.syncEnabled {
		// 如果同步服务未运行但已启用，启动它
		if err := a.StartSync(); err != nil {
			return fmt.Errorf("启动同步服务失败: %v", err)
		}
	}

	// 发送状态更新到前端
	wailsRuntime.EventsEmit(a.ctx, "sync-config-updated", a.GetSyncStatus())

	return nil
}

// MinioConfig MinIO配置结构体
type MinioConfig struct {
	Endpoint        string `json:"endpoint"`
	AccessKeyID     string `json:"accessKeyID"`
	SecretAccessKey string `json:"secretAccessKey"`
	UseSSL          bool   `json:"useSSL"`
	BucketName      string `json:"bucketName"`
	Enabled         bool   `json:"enabled"`
}

// ISCSIDiscoveredTarget iSCSI发现的目标器
type ISCSIDiscoveredTarget struct {
	TargetName string `json:"targetName"`
	PortalIP   string `json:"portalIP"`
	PortalPort int    `json:"portalPort"`
}

// ISCSIConnection iSCSI连接
type ISCSIConnection struct {
	TargetName string `json:"targetName"`
	PortalIP   string `json:"portalIP"`
	PortalPort int    `json:"portalPort"`
	Connected  bool   `json:"connected"`
}

// ISCSIDisk iSCSI磁盘
type ISCSIDisk struct {
	TargetName string `json:"targetName"`
	LUN        int    `json:"lun"`
	Size       int64  `json:"size"`
	MountPoint string `json:"mountPoint"`
	Mounted    bool   `json:"mounted"`
}

// ISCSIInitiatorConfig iSCSI发起器配置
type ISCSIInitiatorConfig struct {
	InitiatorName            string `json:"initiatorName"`
	DefaultPort              int    `json:"defaultPort"`
	LoginTimeout             int    `json:"loginTimeout"`
	LogoutTimeout            int    `json:"logoutTimeout"`
	NOPOutInterval           int    `json:"nopOutInterval"`
	NOPOutTimeout            int    `json:"nopOutTimeout"`
	MaxConnections           int    `json:"maxConnections"`
	HeaderDigest             string `json:"headerDigest"`
	DataDigest               string `json:"dataDigest"`
	MaxRecvDataSegmentLength int    `json:"maxRecvDataSegmentLength"`
}

// TrayMenuItem 系统托盘菜单项
type TrayMenuItem struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Type     string `json:"type"` // "normal", "separator", "checkbox"
	Checked  bool   `json:"checked"`
	Disabled bool   `json:"disabled"`
}

// Register 注册新用户
func (a *App) Register(username, password string) error {
	// 检查用户名是否已存在
	if _, exists := a.users[username]; exists {
		return fmt.Errorf("用户名已存在")
	}

	// 创建新用户，密码使用哈希加密
	a.users[username] = User{
		Username: username,
		Password: hashPassword(password),
	}

	// 保存用户数据
	return a.saveUsers()
}

// Login 用户登录
func (a *App) Login(username, password string) AuthResponse {
	// 检查用户名是否存在
	user, exists := a.users[username]
	if !exists {
		return AuthResponse{
			Success: false,
			Message: "用户名不存在",
		}
	}

	// 检查密码是否正确（比较哈希值）
	if user.Password != hashPassword(password) {
		return AuthResponse{
			Success: false,
			Message: "密码错误",
		}
	}

	// 设置当前用户
	a.currentUser = username
	a.isLoggedIn = true

	// 如果同步功能已启用，启动同步服务
	if a.syncEnabled && a.minioConfig.Enabled && !a.syncRunning {
		go func() {
			if err := a.StartSync(); err != nil {
				fmt.Printf("启动同步服务失败: %v\n", err)
			}
		}()
	}

	return AuthResponse{
		Success: true,
		Message: "登录成功",
	}
}

// Logout 用户登出
func (a *App) Logout() {
	// 停止同步服务
	if a.syncRunning {
		if err := a.StopSync(); err != nil {
			fmt.Printf("停止同步服务失败: %v\n", err)
		}
	}

	// 清除当前用户
	a.currentUser = ""
	a.isLoggedIn = false
}

// loadUsers 加载用户数据
func (a *App) loadUsers() {
	// 检查用户文件是否存在
	if _, err := os.Stat(a.usersFile); os.IsNotExist(err) {
		return
	}

	// 读取用户文件
	data, err := os.ReadFile(a.usersFile)
	if err != nil {
		return
	}

	// 解析用户数据
	var users map[string]User
	if err := json.Unmarshal(data, &users); err != nil {
		return
	}

	// 更新用户数据
	a.users = users
}

// saveUsers 保存用户数据
func (a *App) saveUsers() error {
	// 序列化用户数据
	data, err := json.MarshalIndent(a.users, "", "  ")
	if err != nil {
		return err
	}

	// 写入用户文件
	return os.WriteFile(a.usersFile, data, 0644)
}

// NewClientFeatures 创建客户端特性管理器
func NewClientFeatures(app *App) *ClientFeatures {
	return &ClientFeatures{
		app: app,
		updateChecker: &UpdateChecker{
			currentVersion: "1.0.0",
			updateURL:      "https://example.com/updates",
		},
	}
}

// ClientFeatures 客户端特性管理器
type ClientFeatures struct {
	app           *App
	updateChecker *UpdateChecker
}

// UpdateChecker 更新检查器
type UpdateChecker struct {
	currentVersion string
	updateURL      string
}

// InitializeClientFeatures 初始化客户端特性
func (c *ClientFeatures) InitializeClientFeatures(ctx context.Context) error {
	// 初始化客户端特性
	return nil
}

// RestoreWindowState 恢复窗口状态
func (c *ClientFeatures) RestoreWindowState(ctx context.Context) error {
	// 恢复窗口状态
	return nil
}

// SaveWindowState 保存窗口状态
func (c *ClientFeatures) SaveWindowState(ctx context.Context) error {
	// 保存窗口状态
	return nil
}

// Cleanup 清理资源
func (c *ClientFeatures) Cleanup() {
	// 清理资源
}

// GetTrayMenuItems 获取系统托盘菜单项
func (c *ClientFeatures) GetTrayMenuItems() []TrayMenuItem {
	// 返回系统托盘菜单项
	return []TrayMenuItem{
		{
			ID:       "show",
			Label:    "显示主窗口",
			Type:     "normal",
			Checked:  false,
			Disabled: false,
		},
		{
			ID:       "separator1",
			Type:     "separator",
			Disabled: false,
		},
		{
			ID:       "sync",
			Label:    "同步",
			Type:     "checkbox",
			Checked:  c.app.syncEnabled,
			Disabled: false,
		},
		{
			ID:       "separator2",
			Type:     "separator",
			Disabled: false,
		},
		{
			ID:       "exit",
			Label:    "退出",
			Type:     "normal",
			Checked:  false,
			Disabled: false,
		},
	}
}

// HandleTrayMenuClick 处理托盘菜单点击
func (c *ClientFeatures) HandleTrayMenuClick(menuID string) {
	// 处理托盘菜单点击
	switch menuID {
	case "show":
		c.app.ShowFromTray()
	case "sync":
		c.app.ToggleSyncStatus()
	case "exit":
		wailsRuntime.Quit(c.app.ctx)
	}
}

// SetAutoStart 设置开机自启动
func (c *ClientFeatures) SetAutoStart(enabled bool) error {
	// 设置开机自启动
	return nil
}

// sendNotification 发送系统通知
func (c *ClientFeatures) sendNotification(ctx context.Context, title, message string) {
	// 发送系统通知
	wailsRuntime.EventsEmit(ctx, "notification", map[string]string{
		"title":   title,
		"message": message,
	})
}

// checkForUpdates 检查更新
func (c *ClientFeatures) checkForUpdates(ctx context.Context) {
	// 检查更新
	wailsRuntime.EventsEmit(ctx, "update-check-result", map[string]interface{}{
		"hasUpdate":      false,
		"currentVersion": c.updateChecker.currentVersion,
		"latestVersion":  c.updateChecker.currentVersion,
		"updateURL":      c.updateChecker.updateURL,
	})
}
