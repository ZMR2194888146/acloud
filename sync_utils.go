package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SyncConfig 结构体已在 sync_implementation.go 中定义

// SetSyncInterval 设置同步间隔
func (a *App) SetSyncInterval(seconds int) error {
	if seconds < 10 {
		return fmt.Errorf("同步间隔不能小于10秒")
	}
	
	a.syncInterval = time.Duration(seconds) * time.Second
	
	// 保存到配置
	a.config.SyncConfig.Interval = seconds
	a.SaveConfig()
	
	return nil
}

// SetSyncMode 设置同步模式
func (a *App) SetSyncMode(mode string) error {
	validModes := []string{"full", "selective", "backup", "incremental"}
	valid := false
	for _, m := range validModes {
		if m == mode {
			valid = true
			break
		}
	}
	
	if !valid {
		return fmt.Errorf("无效的同步模式: %s", mode)
	}
	
	a.syncMode = mode
	
	// 保存到配置
	a.config.SyncConfig.Mode = mode
	a.SaveConfig()
	
	return nil
}

// GetSyncMode 获取当前同步模式
func (a *App) GetSyncMode() string {
	return a.syncMode
}

// GetSyncInterval 获取当前同步间隔
func (a *App) GetSyncInterval() time.Duration {
	return a.syncInterval
}

// IsLocalFileNewer 检查本地文件是否比远程文件新
func (a *App) IsLocalFileNewer(localPath string, remotePath string) (bool, error) {
	// 获取本地文件的修改时间
	localModTime, err := getFileModTime(localPath)
	if err != nil {
		return false, fmt.Errorf("获取本地文件修改时间失败: %v", err)
	}
	
	// 获取远程文件信息
	remoteInfo, err := a.GetMinioFileInfo(remotePath)
	if err != nil {
		return true, nil // 远程文件不存在，认为本地文件更新
	}
	
	return localModTime.After(remoteInfo.LastModified), nil
}

// IsRemoteFileNewer 检查远程文件是否比本地文件新
func (a *App) IsRemoteFileNewer(remotePath string, localPath string) (bool, error) {
	// 获取远程文件信息
	remoteInfo, err := a.GetMinioFileInfo(remotePath)
	if err != nil {
		return false, fmt.Errorf("获取远程文件信息失败: %v", err)
	}
	
	// 检查本地文件是否存在
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		return true, nil // 本地文件不存在，认为远程文件更新
	}
	
	// 获取本地文件的修改时间
	localModTime, err := getFileModTime(localPath)
	if err != nil {
		return false, fmt.Errorf("获取本地文件修改时间失败: %v", err)
	}
	
	return remoteInfo.LastModified.After(localModTime), nil
}

// CreateSyncReport 创建同步报告
func (a *App) CreateSyncReport(status SyncStatus) string {
	var report strings.Builder
	
	report.WriteString("同步报告\n")
	report.WriteString("========\n\n")
	
	report.WriteString(fmt.Sprintf("同步时间: %s\n", status.LastSync.Format("2006-01-02 15:04:05")))
	report.WriteString(fmt.Sprintf("同步模式: %s\n", status.SyncMode))
	report.WriteString(fmt.Sprintf("上传文件数: %d\n", status.FilesUploaded))
	report.WriteString(fmt.Sprintf("下载文件数: %d\n", status.FilesDownloaded))
	report.WriteString(fmt.Sprintf("冲突数: %d\n", status.ConflictCount))
	
	if len(status.Errors) > 0 {
		report.WriteString("\n错误:\n")
		for i, err := range status.Errors {
			report.WriteString(fmt.Sprintf("%d. %s\n", i+1, err))
		}
	}
	
	return report.String()
}

// SaveSyncReport 保存同步报告到文件
func (a *App) SaveSyncReport(status SyncStatus) error {
	// 创建报告目录
	reportsDir := filepath.Join(a.configDir, "sync_reports")
	if err := os.MkdirAll(reportsDir, 0755); err != nil {
		return fmt.Errorf("创建报告目录失败: %v", err)
	}
	
	// 创建报告文件名
	reportFile := filepath.Join(reportsDir, fmt.Sprintf("sync_report_%s.txt", status.LastSync.Format("20060102_150405")))
	
	// 创建报告内容
	report := a.CreateSyncReport(status)
	
	// 写入文件
	err := os.WriteFile(reportFile, []byte(report), 0644)
	if err != nil {
		return fmt.Errorf("写入报告文件失败: %v", err)
	}
	
	return nil
}

// GetSyncReports 获取同步报告列表
func (a *App) GetSyncReports() ([]string, error) {
	// 报告目录
	reportsDir := filepath.Join(a.configDir, "sync_reports")
	
	// 检查目录是否存在
	if _, err := os.Stat(reportsDir); os.IsNotExist(err) {
		return []string{}, nil
	}
	
	// 读取目录
	files, err := os.ReadDir(reportsDir)
	if err != nil {
		return nil, fmt.Errorf("读取报告目录失败: %v", err)
	}
	
	// 过滤报告文件
	var reports []string
	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), "sync_report_") && strings.HasSuffix(file.Name(), ".txt") {
			reports = append(reports, filepath.Join(reportsDir, file.Name()))
		}
	}
	
	return reports, nil
}

// ReadSyncReport 读取同步报告
func (a *App) ReadSyncReport(reportPath string) (string, error) {
	// 读取文件
	data, err := os.ReadFile(reportPath)
	if err != nil {
		return "", fmt.Errorf("读取报告文件失败: %v", err)
	}
	
	return string(data), nil
}

// DeleteSyncReport 删除同步报告
func (a *App) DeleteSyncReport(reportPath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(reportPath); os.IsNotExist(err) {
		return fmt.Errorf("报告文件不存在: %s", reportPath)
	}
	
	// 删除文件
	err := os.Remove(reportPath)
	if err != nil {
		return fmt.Errorf("删除报告文件失败: %v", err)
	}
	
	return nil
}

// GetSyncStats 获取同步统计信息
func (a *App) GetSyncStats() map[string]interface{} {
	// 获取历史统计
	historyStats := a.GetSyncHistoryStats()
	
	// 获取当前状态
	status := a.GetSyncStatus()
	
	// 合并统计信息
	stats := map[string]interface{}{
		"enabled":          a.syncEnabled,
		"running":          status.Running,
		"interval":         a.syncInterval,
		"mode":             a.syncMode,
		"conflictCount":    status.ConflictCount,
		"ruleCount":        len(a.syncRules),
		"totalUploaded":    historyStats["totalUploaded"],
		"totalDownloaded":  historyStats["totalDownloaded"],
		"totalConflicts":   historyStats["totalConflicts"],
		"totalErrors":      historyStats["totalErrors"],
		"successRate":      historyStats["successRate"],
		"averageDuration":  historyStats["averageDuration"],
	}
	
	return stats
}

// ExportSyncConfig 导出同步配置
func (a *App) ExportSyncConfig(filePath string) error {
	// 创建配置对象
	config := map[string]interface{}{
		"enabled":                 a.syncEnabled,
		"interval":                a.syncInterval.Seconds(),
		"mode":                    a.syncMode,
		"rules":                   a.syncRules,
		"defaultConflictResolution": a.defaultConflictResolution,
	}
	
	// 序列化为JSON
	data, err := a.jsonParser.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}
	
	// 写入文件
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}
	
	return nil
}

// ImportSyncConfig 导入同步配置
func (a *App) ImportSyncConfig(filePath string) error {
	// 读取文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}
	
	// 解析JSON
	var config map[string]interface{}
	err = a.jsonParser.Unmarshal(data, &config)
	if err != nil {
		return fmt.Errorf("解析配置失败: %v", err)
	}
	
	// 应用配置
	if enabled, ok := config["enabled"].(bool); ok {
		a.syncEnabled = enabled
	}
	
	if interval, ok := config["interval"].(float64); ok {
		a.syncInterval = time.Duration(interval) * time.Second
	}
	
	if mode, ok := config["mode"].(string); ok {
		a.syncMode = mode
	}
	
	if resolution, ok := config["defaultConflictResolution"].(string); ok {
		a.defaultConflictResolution = resolution
	}
	
	// 导入规则
	if rulesData, ok := config["rules"]; ok {
		// 重新序列化和反序列化规则
		rulesJSON, err := a.jsonParser.Marshal(rulesData)
		if err != nil {
			return fmt.Errorf("序列化规则失败: %v", err)
		}
		
		err = a.jsonParser.Unmarshal(rulesJSON, &a.syncRules)
		if err != nil {
			return fmt.Errorf("解析规则失败: %v", err)
		}
	}
	
	// 保存配置
	a.SaveConfig()
	a.SaveSyncRules()
	
	return nil
}

// ValidateSyncRule 验证同步规则
func (a *App) ValidateSyncRule(rule SyncRule) error {
	// 检查名称
	if rule.Name == "" {
		return fmt.Errorf("规则名称不能为空")
	}
	
	// 检查本地路径
	if rule.LocalPath == "" {
		return fmt.Errorf("本地路径不能为空")
	}
	
	// 检查远程路径
	if rule.RemotePath == "" {
		return fmt.Errorf("远程路径不能为空")
	}
	
	// 检查同步方向
	if rule.Direction != "upload" && rule.Direction != "download" && rule.Direction != "bidirectional" {
		return fmt.Errorf("无效的同步方向: %s", rule.Direction)
	}
	
	return nil
}

// UpdateSyncRule 更新同步规则
func (a *App) UpdateSyncRule(rule SyncRule) error {
	// 验证规则
	err := a.ValidateSyncRule(rule)
	if err != nil {
		return err
	}
	
	// 查找规则
	for i, r := range a.syncRules {
		if r.ID == rule.ID {
			a.syncRules[i] = rule
			a.SaveSyncRules()
			return nil
		}
	}
	
	return fmt.Errorf("未找到同步规则: %s", rule.ID)
}

// GetSyncRuleByID 根据ID获取同步规则
func (a *App) GetSyncRuleByID(ruleID string) (SyncRule, error) {
	for _, rule := range a.syncRules {
		if rule.ID == ruleID {
			return rule, nil
		}
	}
	
	return SyncRule{}, fmt.Errorf("未找到同步规则: %s", ruleID)
}

// GetSyncRuleByName 根据名称获取同步规则
func (a *App) GetSyncRuleByName(name string) (SyncRule, error) {
	for _, rule := range a.syncRules {
		if rule.Name == name {
			return rule, nil
		}
	}
	
	return SyncRule{}, fmt.Errorf("未找到同步规则: %s", name)
}

// GetEnabledSyncRules 获取已启用的同步规则
func (a *App) GetEnabledSyncRules() []SyncRule {
	var rules []SyncRule
	for _, rule := range a.syncRules {
		if rule.Enabled {
			rules = append(rules, rule)
		}
	}
	
	return rules
}