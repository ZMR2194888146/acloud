package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SyncHistoryEntry 同步历史条目
type SyncHistoryEntry struct {
	ID              string `json:"id"`
	Timestamp       time.Time `json:"timestamp"`
	SyncMode        string `json:"syncMode"`
	FilesUploaded   int    `json:"filesUploaded"`
	FilesDownloaded int    `json:"filesDownloaded"`
	ConflictCount   int    `json:"conflictCount"`
	ErrorCount      int    `json:"errorCount"`
	Duration        int64  `json:"duration"` // 毫秒
	Status          string `json:"status"`   // "success", "partial", "failed"
}

// SyncHistory 同步历史记录
type SyncHistory struct {
	Entries []SyncHistoryEntry `json:"entries"`
}

// recordSyncHistory 记录同步历史
func (a *App) recordSyncHistory(status SyncStatus, duration time.Duration) error {
	// 加载历史记录
	history, err := a.loadSyncHistory()
	if err != nil {
		fmt.Printf("加载同步历史记录失败: %v，将创建新的历史记录\n", err)
		history = &SyncHistory{
			Entries: []SyncHistoryEntry{},
		}
	}
	
	// 确定同步状态
	syncStatus := "success"
	if len(status.Errors) > 0 {
		if status.FilesUploaded > 0 || status.FilesDownloaded > 0 {
			syncStatus = "partial"
		} else {
			syncStatus = "failed"
		}
	}
	
	// 创建新的历史记录条目
	entry := SyncHistoryEntry{
		ID:              fmt.Sprintf("sync_%d", time.Now().Unix()),
		Timestamp:       time.Now(),
		SyncMode:        status.SyncMode,
		FilesUploaded:   status.FilesUploaded,
		FilesDownloaded: status.FilesDownloaded,
		ConflictCount:   status.ConflictCount,
		ErrorCount:      len(status.Errors),
		Duration:        duration.Milliseconds(),
		Status:          syncStatus,
	}
	
	// 添加到历史记录
	history.Entries = append(history.Entries, entry)
	
	// 限制历史记录数量
	if len(history.Entries) > 100 {
		history.Entries = history.Entries[len(history.Entries)-100:]
	}
	
	// 保存历史记录
	return a.saveSyncHistory(history)
}

// loadSyncHistory 加载同步历史记录
func (a *App) loadSyncHistory() (*SyncHistory, error) {
	// 历史记录文件路径
	historyPath := filepath.Join(a.configDir, "sync_history.json")
	
	// 检查文件是否存在
	if _, err := os.Stat(historyPath); os.IsNotExist(err) {
		return &SyncHistory{
			Entries: []SyncHistoryEntry{},
		}, nil
	}
	
	// 读取文件
	data, err := os.ReadFile(historyPath)
	if err != nil {
		return nil, fmt.Errorf("读取历史记录文件失败: %v", err)
	}
	
	// 解析JSON
	var history SyncHistory
	err = a.jsonParser.Unmarshal(data, &history)
	if err != nil {
		return nil, fmt.Errorf("解析历史记录失败: %v", err)
	}
	
	return &history, nil
}

// saveSyncHistory 保存同步历史记录
func (a *App) saveSyncHistory(history *SyncHistory) error {
	// 历史记录文件路径
	historyPath := filepath.Join(a.configDir, "sync_history.json")
	
	// 序列化为JSON
	data, err := a.jsonParser.Marshal(history)
	if err != nil {
		return fmt.Errorf("序列化历史记录失败: %v", err)
	}
	
	// 写入文件
	err = os.WriteFile(historyPath, data, 0644)
	if err != nil {
		return fmt.Errorf("写入历史记录文件失败: %v", err)
	}
	
	return nil
}

// GetSyncHistory 获取同步历史记录
func (a *App) GetSyncHistory() ([]SyncHistoryEntry, error) {
	history, err := a.loadSyncHistory()
	if err != nil {
		return nil, err
	}
	
	return history.Entries, nil
}

// ClearSyncHistory 清除同步历史记录
func (a *App) ClearSyncHistory() error {
	// 创建空的历史记录
	history := &SyncHistory{
		Entries: []SyncHistoryEntry{},
	}
	
	// 保存空历史记录
	return a.saveSyncHistory(history)
}

// GetSyncHistoryStats 获取同步历史统计信息
func (a *App) GetSyncHistoryStats() map[string]interface{} {
	history, err := a.loadSyncHistory()
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}
	
	// 统计信息
	totalUploaded := 0
	totalDownloaded := 0
	totalConflicts := 0
	totalErrors := 0
	totalDuration := int64(0)
	successCount := 0
	partialCount := 0
	failedCount := 0
	
	for _, entry := range history.Entries {
		totalUploaded += entry.FilesUploaded
		totalDownloaded += entry.FilesDownloaded
		totalConflicts += entry.ConflictCount
		totalErrors += entry.ErrorCount
		totalDuration += entry.Duration
		
		switch entry.Status {
		case "success":
			successCount++
		case "partial":
			partialCount++
		case "failed":
			failedCount++
		}
	}
	
	return map[string]interface{}{
		"totalEntries":     len(history.Entries),
		"totalUploaded":    totalUploaded,
		"totalDownloaded":  totalDownloaded,
		"totalConflicts":   totalConflicts,
		"totalErrors":      totalErrors,
		"totalDuration":    totalDuration,
		"successCount":     successCount,
		"partialCount":     partialCount,
		"failedCount":      failedCount,
		"averageDuration":  float64(totalDuration) / float64(len(history.Entries)),
		"successRate":      float64(successCount) / float64(len(history.Entries)),
	}
}
