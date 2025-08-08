package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// RunSyncCommand 运行同步命令行工具
func (a *App) RunSyncCommand() {
	// 检查命令行参数
	if len(os.Args) < 2 || os.Args[1] != "sync" {
		return
	}

	// 登录
	// 检查是否已登录
	if !a.isLoggedIn {
		fmt.Println("错误: 用户未登录")
		return
	}

	// 检查 MinIO 是否已启用
	if !a.minioConfig.Enabled || a.minioClient == nil {
		fmt.Println("MinIO 未启用，无法使用同步功能")
		os.Exit(1)
	}

	// 解析子命令
	if len(os.Args) >= 3 {
		switch os.Args[2] {
		case "start":
			a.cmdStartSync()
		case "stop":
			a.cmdStopSync()
		case "status":
			a.cmdSyncStatus()
		case "run":
			a.cmdRunSync()
		case "add-rule":
			a.cmdAddSyncRule()
		case "list-rules":
			a.cmdListSyncRules()
		case "remove-rule":
			a.cmdRemoveSyncRule()
		case "enable-rule":
			a.cmdEnableSyncRule()
		case "disable-rule":
			a.cmdDisableSyncRule()
		case "history":
			a.cmdSyncHistory()
		case "conflicts":
			a.cmdSyncConflicts()
		case "resolve":
			a.cmdResolveConflict()
		default:
			a.showSyncHelp()
		}
	} else {
		a.showSyncHelp()
	}

	os.Exit(0)
}

// showSyncHelp 显示同步命令帮助
func (a *App) showSyncHelp() {
	fmt.Println("HKCE Cloud 同步命令行工具")
	fmt.Println("用法: hkce-cloud sync <子命令> [参数...]")
	fmt.Println("\n可用子命令:")
	fmt.Println("  start                         - 启动同步服务")
	fmt.Println("  stop                          - 停止同步服务")
	fmt.Println("  status                        - 显示同步状态")
	fmt.Println("  run [mode]                    - 执行一次同步 (mode: full, selective, backup, incremental)")
	fmt.Println("  add-rule <名称> <本地路径> <远程路径> <方向> - 添加同步规则")
	fmt.Println("  list-rules                    - 列出同步规则")
	fmt.Println("  remove-rule <ID>              - 删除同步规则")
	fmt.Println("  enable-rule <ID>              - 启用同步规则")
	fmt.Println("  disable-rule <ID>             - 禁用同步规则")
	fmt.Println("  history                       - 显示同步历史")
	fmt.Println("  conflicts                     - 显示同步冲突")
	fmt.Println("  resolve <路径> <解决方式>      - 解决同步冲突 (local, remote, both, skip)")
}

// cmdStartSync 启动同步服务
func (a *App) cmdStartSync() {
	fmt.Println("正在启动同步服务...")
	err := a.StartSync()
	if err != nil {
		fmt.Printf("启动同步服务失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("同步服务已启动")
}

// cmdStopSync 停止同步服务
func (a *App) cmdStopSync() {
	fmt.Println("正在停止同步服务...")
	err := a.StopSync()
	if err != nil {
		fmt.Printf("停止同步服务失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("同步服务已停止")
}

// cmdSyncStatus 显示同步状态
func (a *App) cmdSyncStatus() {
	status := a.GetSyncStatus()
	fmt.Println("同步状态:")
	fmt.Printf("  启用: %v\n", a.syncEnabled)
	fmt.Printf("  运行中: %v\n", status.Running)
	fmt.Printf("  同步间隔: %s\n", a.syncInterval)
	fmt.Printf("  同步模式: %s\n", a.syncMode)
	fmt.Printf("  冲突数量: %d\n", status.ConflictCount)
	fmt.Printf("  规则数量: %d\n", len(a.syncRules))
}

// cmdRunSync 执行一次同步
func (a *App) cmdRunSync() {
	mode := "full"
	if len(os.Args) >= 4 {
		mode = os.Args[3]
	}

	// 设置同步模式
	a.syncMode = mode

	fmt.Printf("正在执行%s同步...\n", mode)
	err := a.TriggerManualSync()
	if err != nil {
		fmt.Printf("执行同步失败: %v\n", err)
		os.Exit(1)
	}

	// 等待同步完成
	fmt.Println("同步已启动，请等待完成...")
	time.Sleep(2 * time.Second)
}

// cmdAddSyncRule 添加同步规则
func (a *App) cmdAddSyncRule() {
	if len(os.Args) < 7 {
		fmt.Println("错误: 参数不足")
		fmt.Println("用法: hkce-cloud sync add-rule <名称> <本地路径> <远程路径> <方向>")
		fmt.Println("方向: upload, download, bidirectional")
		os.Exit(1)
	}

	name := os.Args[3]
	localPath := os.Args[4]
	remotePath := os.Args[5]
	direction := os.Args[6]

	// 验证方向
	if direction != "upload" && direction != "download" && direction != "bidirectional" {
		fmt.Println("错误: 无效的同步方向")
		fmt.Println("有效的方向: upload, download, bidirectional")
		os.Exit(1)
	}

	// 验证本地路径
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		fmt.Printf("警告: 本地路径不存在: %s\n", localPath)
		// 创建本地路径
		if err := os.MkdirAll(localPath, 0755); err != nil {
			fmt.Printf("创建本地路径失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("已创建本地路径: %s\n", localPath)
	}

	// 创建同步规则
	rule := SyncRule{
		ID:         fmt.Sprintf("rule_%d", time.Now().Unix()),
		Name:       name,
		LocalPath:  localPath,
		RemotePath: remotePath,
		Direction:  direction,
		Filters:    []string{},
		Enabled:    true,
	}

	// 添加过滤器
	if len(os.Args) > 7 {
		rule.Filters = strings.Split(os.Args[7], ",")
	}

	// 添加规则
	a.AddSyncRule(rule)

	fmt.Printf("已添加同步规则: %s\n", name)
}

// cmdListSyncRules 列出同步规则
func (a *App) cmdListSyncRules() {
	rules := a.GetSyncRules()

	if len(rules) == 0 {
		fmt.Println("没有同步规则")
		return
	}

	fmt.Println("同步规则列表:")
	for i, rule := range rules {
		status := "启用"
		if !rule.Enabled {
			status = "禁用"
		}

		fmt.Printf("%d. %s (%s)\n", i+1, rule.Name, status)
		fmt.Printf("   ID: %s\n", rule.ID)
		fmt.Printf("   本地路径: %s\n", rule.LocalPath)
		fmt.Printf("   远程路径: %s\n", rule.RemotePath)
		fmt.Printf("   方向: %s\n", rule.Direction)

		if len(rule.Filters) > 0 {
			fmt.Printf("   过滤器: %s\n", strings.Join(rule.Filters, ", "))
		}

		fmt.Println()
	}
}

// cmdRemoveSyncRule 删除同步规则
func (a *App) cmdRemoveSyncRule() {
	if len(os.Args) < 4 {
		fmt.Println("错误: 缺少规则ID")
		fmt.Println("用法: hkce-cloud sync remove-rule <ID>")
		os.Exit(1)
	}

	ruleID := os.Args[3]

	// 删除规则
	err := a.RemoveSyncRule(ruleID)
	if err != nil {
		fmt.Printf("删除同步规则失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("已删除同步规则: %s\n", ruleID)
}

// cmdEnableSyncRule 启用同步规则
func (a *App) cmdEnableSyncRule() {
	if len(os.Args) < 4 {
		fmt.Println("错误: 缺少规则ID")
		fmt.Println("用法: hkce-cloud sync enable-rule <ID>")
		os.Exit(1)
	}

	ruleID := os.Args[3]

	// 启用规则
	err := a.EnableSyncRule(ruleID)
	if err != nil {
		fmt.Printf("启用同步规则失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("已启用同步规则: %s\n", ruleID)
}

// cmdDisableSyncRule 禁用同步规则
func (a *App) cmdDisableSyncRule() {
	if len(os.Args) < 4 {
		fmt.Println("错误: 缺少规则ID")
		fmt.Println("用法: hkce-cloud sync disable-rule <ID>")
		os.Exit(1)
	}

	ruleID := os.Args[3]

	// 禁用规则
	err := a.DisableSyncRule(ruleID)
	if err != nil {
		fmt.Printf("禁用同步规则失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("已禁用同步规则: %s\n", ruleID)
}

// cmdSyncHistory 显示同步历史
func (a *App) cmdSyncHistory() {
	entries, err := a.GetSyncHistory()
	if err != nil {
		fmt.Printf("获取同步历史失败: %v\n", err)
		os.Exit(1)
	}

	if len(entries) == 0 {
		fmt.Println("没有同步历史记录")
		return
	}

	fmt.Println("同步历史记录:")
	for i, entry := range entries {
		fmt.Printf("%d. %s (%s)\n", i+1, entry.Timestamp.Format("2006-01-02 15:04:05"), entry.Status)
		fmt.Printf("   ID: %s\n", entry.ID)
		fmt.Printf("   模式: %s\n", entry.SyncMode)
		fmt.Printf("   上传文件: %d\n", entry.FilesUploaded)
		fmt.Printf("   下载文件: %d\n", entry.FilesDownloaded)
		fmt.Printf("   冲突数量: %d\n", entry.ConflictCount)
		fmt.Printf("   错误数量: %d\n", entry.ErrorCount)
		fmt.Printf("   持续时间: %.2f秒\n", float64(entry.Duration)/1000.0)
		fmt.Println()
	}

	// 显示统计信息
	stats := a.GetSyncHistoryStats()
	fmt.Println("同步统计信息:")
	fmt.Printf("  总记录数: %d\n", stats["totalEntries"])
	fmt.Printf("  总上传文件: %d\n", stats["totalUploaded"])
	fmt.Printf("  总下载文件: %d\n", stats["totalDownloaded"])
	fmt.Printf("  总冲突数量: %d\n", stats["totalConflicts"])
	fmt.Printf("  总错误数量: %d\n", stats["totalErrors"])
	fmt.Printf("  平均持续时间: %.2f秒\n", float64(stats["averageDuration"].(float64))/1000.0)
	fmt.Printf("  成功率: %.2f%%\n", stats["successRate"].(float64)*100)
}

// cmdSyncConflicts 显示同步冲突
func (a *App) cmdSyncConflicts() {
	conflicts := a.GetConflictFiles()

	if len(conflicts) == 0 {
		fmt.Println("没有同步冲突")
		return
	}

	fmt.Println("同步冲突列表:")
	for i, conflict := range conflicts {
		fmt.Printf("%d. %s\n", i+1, conflict.Path)
		fmt.Printf("   本地修改时间: %s\n", conflict.LocalModTime.Format("2006-01-02 15:04:05"))
		fmt.Printf("   远程修改时间: %s\n", conflict.RemoteModTime.Format("2006-01-02 15:04:05"))
		fmt.Printf("   解决状态: %s\n", conflict.Resolution)
		fmt.Println()
	}
}

// cmdResolveConflict 解决同步冲突
func (a *App) cmdResolveConflict() {
	if len(os.Args) < 5 {
		fmt.Println("错误: 参数不足")
		fmt.Println("用法: hkce-cloud sync resolve <路径> <解决方式>")
		fmt.Println("解决方式: local, remote, both, skip")
		os.Exit(1)
	}

	path := os.Args[3]
	resolution := os.Args[4]

	// 验证解决方式
	if resolution != "local" && resolution != "remote" && resolution != "both" && resolution != "skip" {
		fmt.Println("错误: 无效的解决方式")
		fmt.Println("有效的解决方式: local, remote, both, skip")
		os.Exit(1)
	}

	// 解决冲突
	err := a.ResolveConflict(path, resolution)
	if err != nil {
		fmt.Printf("解决冲突失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("已解决冲突: %s (使用%s)\n", path, resolution)
}

// AddSyncRule 添加同步规则
func (a *App) AddSyncRule(rule SyncRule) {
	a.syncRules = append(a.syncRules, rule)
	a.SaveSyncRules()
}

// RemoveSyncRule 删除同步规则
func (a *App) RemoveSyncRule(ruleID string) error {
	for i, rule := range a.syncRules {
		if rule.ID == ruleID {
			a.syncRules = append(a.syncRules[:i], a.syncRules[i+1:]...)
			a.SaveSyncRules()
			return nil
		}
	}

	return fmt.Errorf("未找到同步规则: %s", ruleID)
}

// EnableSyncRule 启用同步规则
func (a *App) EnableSyncRule(ruleID string) error {
	for i, rule := range a.syncRules {
		if rule.ID == ruleID {
			a.syncRules[i].Enabled = true
			a.SaveSyncRules()
			return nil
		}
	}

	return fmt.Errorf("未找到同步规则: %s", ruleID)
}

// DisableSyncRule 禁用同步规则
func (a *App) DisableSyncRule(ruleID string) error {
	for i, rule := range a.syncRules {
		if rule.ID == ruleID {
			a.syncRules[i].Enabled = false
			a.SaveSyncRules()
			return nil
		}
	}

	return fmt.Errorf("未找到同步规则: %s", ruleID)
}

// GetSyncRules 获取同步规则
func (a *App) GetSyncRules() []SyncRule {
	return a.syncRules
}

// SaveSyncRules 保存同步规则
func (a *App) SaveSyncRules() error {
	// 规则文件路径
	rulesPath := filepath.Join(a.configDir, "sync_rules.json")

	// 序列化为JSON
	data, err := a.jsonParser.Marshal(a.syncRules)
	if err != nil {
		return fmt.Errorf("序列化同步规则失败: %v", err)
	}

	// 写入文件
	err = os.WriteFile(rulesPath, data, 0644)
	if err != nil {
		return fmt.Errorf("写入同步规则文件失败: %v", err)
	}

	return nil
}

// LoadSyncRules 加载同步规则
func (a *App) LoadSyncRules() error {
	// 规则文件路径
	rulesPath := filepath.Join(a.configDir, "sync_rules.json")

	// 检查文件是否存在
	if _, err := os.Stat(rulesPath); os.IsNotExist(err) {
		a.syncRules = []SyncRule{}
		return nil
	}

	// 读取文件
	data, err := os.ReadFile(rulesPath)
	if err != nil {
		return fmt.Errorf("读取同步规则文件失败: %v", err)
	}

	// 解析JSON
	err = a.jsonParser.Unmarshal(data, &a.syncRules)
	if err != nil {
		return fmt.Errorf("解析同步规则失败: %v", err)
	}

	return nil
}
