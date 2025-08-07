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

// MinioConfig 存储 MinIO 配置信息
type MinioConfig struct {
	Endpoint        string `json:"endpoint"`
	AccessKeyID     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	UseSSL          bool   `json:"useSSL"`
	BucketName      string `json:"bucketName"`
	Enabled         bool   `json:"enabled"`
}

// SyncConfig 同步配置
type SyncConfig struct {
	Enabled      bool          `json:"enabled"`
	Interval     time.Duration `json:"interval"`
	AutoDownload bool          `json:"autoDownload"`
	AutoUpload   bool          `json:"autoUpload"`
}

// SyncStatus 同步状态
type SyncStatus struct {
	Running         bool      `json:"running"`
	LastSync        time.Time `json:"lastSync"`
	FilesUploaded   int       `json:"filesUploaded"`
	FilesDownloaded int       `json:"filesDownloaded"`
	Errors          []string  `json:"errors"`
	SyncMode        string    `json:"syncMode"`      // "full", "selective", "backup"
	ConflictCount   int       `json:"conflictCount"`
}

// FileVersion 文件版本信息
type FileVersion struct {
	Version   int       `json:"version"`
	Path      string    `json:"path"`
	Size      int64     `json:"size"`
	ModTime   time.Time `json:"modTime"`
	Hash      string    `json:"hash"`
	Comment   string    `json:"comment"`
	CreatedBy string    `json:"createdBy"`
}

// SyncRule 同步规则
type SyncRule struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	LocalPath  string   `json:"localPath"`
	RemotePath string   `json:"remotePath"`
	Direction  string   `json:"direction"` // "bidirectional", "upload", "download"
	Enabled    bool     `json:"enabled"`
	Filters    []string `json:"filters"`  // 文件过滤规则
	Schedule   string   `json:"schedule"` // 同步计划
}

// ConflictFile 冲突文件信息
type ConflictFile struct {
	Path          string    `json:"path"`
	LocalSize     int64     `json:"localSize"`
	RemoteSize    int64     `json:"remoteSize"`
	LocalModTime  time.Time `json:"localModTime"`
	RemoteModTime time.Time `json:"remoteModTime"`
	Resolution    string    `json:"resolution"` // "pending", "local", "remote", "both"
}

// ShareLink 分享链接
type ShareLink struct {
	ID           string    `json:"id"`
	Path         string    `json:"path"`
	Token        string    `json:"token"`
	Password     string    `json:"password"`
	ExpiryTime   time.Time `json:"expiryTime"`
	Downloads    int       `json:"downloads"`
	MaxDownloads int       `json:"maxDownloads"`
	CreatedBy    string    `json:"createdBy"`
}

// MinioFileInfo MinIO 文件信息
type MinioFileInfo struct {
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"lastModified"`
	IsDir        bool      `json:"isDir"`
}

// FilePreview 文件预览信息结构体
type FilePreview struct {
	MimeType string `json:"mimeType"`
	Content  string `json:"content"`
	IsBase64 bool   `json:"isBase64"`
}

// iSCSI 客户端相关结构体

// ISCSIDiscoveredTarget 发现的iSCSI目标器
type ISCSIDiscoveredTarget struct {
	IQN        string `json:"iqn"`        // 目标器IQN
	Portal     string `json:"portal"`     // 目标器地址:端口
	TargetName string `json:"targetName"` // 目标器名称
	Status     string `json:"status"`     // "available", "connected", "error"
}

// ISCSIConnection iSCSI连接信息
type ISCSIConnection struct {
	ID           string    `json:"id"`
	TargetIQN    string    `json:"targetIqn"`
	Portal       string    `json:"portal"`
	Status       string    `json:"status"`       // "connected", "connecting", "disconnected", "error"
	ConnectedAt  time.Time `json:"connectedAt"`
	LastActivity time.Time `json:"lastActivity"`
	BytesRead    int64     `json:"bytesRead"`
	BytesWritten int64     `json:"bytesWritten"`
	IOPS         int64     `json:"iops"`
	Latency      float64   `json:"latency"`   // 延迟(毫秒)
	Bandwidth    float64   `json:"bandwidth"` // 带宽(MB/s)
}

// ISCSIDisk iSCSI磁盘信息
type ISCSIDisk struct {
	DevicePath string `json:"devicePath"` // 设备路径 (/dev/sdX)
	TargetIQN  string `json:"targetIqn"`  // 关联的目标器IQN
	LUN        int    `json:"lun"`        // LUN编号
	Size       int64  `json:"size"`       // 磁盘大小(字节)
	Model      string `json:"model"`      // 磁盘型号
	Serial     string `json:"serial"`     // 序列号
	Status     string `json:"status"`     // "online", "offline", "mounted", "error"
	MountPoint string `json:"mountPoint"` // 挂载点
	FileSystem string `json:"fileSystem"` // 文件系统类型
	UsedSpace  int64  `json:"usedSpace"`  // 已使用空间
	FreeSpace  int64  `json:"freeSpace"`  // 可用空间
}

// ISCSIInitiatorConfig iSCSI发起器配置
type ISCSIInitiatorConfig struct {
	InitiatorName                string `json:"initiatorName"`                // 发起器IQN
	DefaultPort                  int    `json:"defaultPort"`                  // 默认端口
	LoginTimeout                 int    `json:"loginTimeout"`                 // 登录超时(秒)
	LogoutTimeout                int    `json:"logoutTimeout"`                // 登出超时(秒)
	NOPOutInterval               int    `json:"nopOutInterval"`               // NOP-Out间隔(秒)
	NOPOutTimeout                int    `json:"nopOutTimeout"`                // NOP-Out超时(秒)
	MaxConnections               int    `json:"maxConnections"`               // 最大连接数
	HeaderDigest                 string `json:"headerDigest"`                 // 头部摘要算法
	DataDigest                   string `json:"dataDigest"`                   // 数据摘要算法
	MaxRecvDataSegmentLength     int    `json:"maxRecvDataSegmentLength"`     // 最大接收数据段长度
}

// ISCSIPerformanceStats iSCSI性能统计
type ISCSIPerformanceStats struct {
	TotalConnections  int                 `json:"totalConnections"`
	ActiveConnections int                 `json:"activeConnections"`
	TotalDisks        int                 `json:"totalDisks"`
	MountedDisks      int                 `json:"mountedDisks"`
	TotalIOPS         int64               `json:"totalIops"`
	TotalBandwidth    float64             `json:"totalBandwidth"`
	AverageLatency    float64             `json:"averageLatency"`
	ConnectionStats   []ISCSIConnection   `json:"connectionStats"`
	DiskStats         []ISCSIDisk         `json:"diskStats"`
}