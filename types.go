package main

import (
	"time"
)

// User 结构体用于存储用户信息
type User struct {
	Username string `json:"username"`
	Password string `json:"password"` // 存储密码哈希
}

// FileInfo 结构体用于存储文件信息
type FileInfo struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Size      int64     `json:"size"`
	IsDir     bool      `json:"isDir"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// FilePreview 文件预览信息结构体
type FilePreview struct {
	MimeType string `json:"mimeType"`
	Content  string `json:"content"`
	IsBase64 bool   `json:"isBase64"`
}

// ISCSIPerformanceStats iSCSI性能统计
type ISCSIPerformanceStats struct {
	TotalConnections  int               `json:"totalConnections"`
	ActiveConnections int               `json:"activeConnections"`
	TotalDisks        int               `json:"totalDisks"`
	MountedDisks      int               `json:"mountedDisks"`
	TotalIOPS         int64             `json:"totalIops"`
	TotalBandwidth    float64           `json:"totalBandwidth"`
	AverageLatency    float64           `json:"averageLatency"`
	ConnectionStats   []ISCSIConnection `json:"connectionStats"`
	DiskStats         []ISCSIDisk       `json:"diskStats"`
}


// SystemInfo 系统信息
type SystemInfo struct {
	OS           string  `json:"os"`
	Arch         string  `json:"arch"`
	CPUCount     int     `json:"cpuCount"`
	CPUPercent   float64 `json:"cpuPercent"`
	MemoryTotal  uint64  `json:"memoryTotal"`
	MemoryUsed   uint64  `json:"memoryUsed"`
	MemoryFree   uint64  `json:"memoryFree"`
	DiskTotal    uint64  `json:"diskTotal"`
	DiskUsed     uint64  `json:"diskUsed"`
	DiskFree     uint64  `json:"diskFree"`
	GoVersion    string  `json:"goVersion"`
	NumGoroutine int     `json:"numGoroutine"`
}

// ClientFeaturesStatus 客户端特性状态
type ClientFeaturesStatus struct {
	AutoStart     bool `json:"autoStart"`
	Notifications bool `json:"notifications"`
	TrayEnabled   bool `json:"trayEnabled"`
	FileWatcher   bool `json:"fileWatcher"`
	UpdateChecker bool `json:"updateChecker"`
}

// ClientFeaturesConfig 客户端特性配置
type ClientFeaturesConfig struct {
	AutoStart      bool          `json:"autoStart"`
	Notifications  bool          `json:"notifications"`
	TrayEnabled    bool          `json:"trayEnabled"`
	FileWatcher    bool          `json:"fileWatcher"`
	UpdateChecker  bool          `json:"updateChecker"`
	UpdateInterval time.Duration `json:"updateInterval"`
}

