package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SyncConfig 存储同步配置
type SyncConfig struct {
	LocalPath  string `json:"localPath"`
	RemotePath string `json:"remotePath"`
	Direction  string `json:"direction"` // "up", "down", "both"
	Interval   int    `json:"interval"`  // 同步间隔（秒）
}

// SyncRule 同步规则
type SyncRule struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	LocalPath  string   `json:"localPath"`
	RemotePath string   `json:"remotePath"`
	Direction  string   `json:"direction"`
	Filters    []string `json:"filters"`
	Enabled    bool     `json:"enabled"`
}

// SyncStatus 同步状态
type SyncStatus struct {
	Running         bool      `json:"running"`
	LastSync        time.Time `json:"lastSync"`
	FilesUploaded   int       `json:"filesUploaded"`
	FilesDownloaded int       `json:"filesDownloaded"`
	Errors          []string  `json:"errors"`
	SyncMode        string    `json:"syncMode"`
	ConflictCount   int       `json:"conflictCount"`
}

// ConflictFile 冲突文件
type ConflictFile struct {
	Path          string    `json:"path"`
	LocalModTime  time.Time `json:"localModTime"`
	RemoteModTime time.Time `json:"remoteModTime"`
	Resolution    string    `json:"resolution"` // "local", "remote", "none"
}

// MinioFileInfo MinIO文件信息
type MinioFileInfo struct {
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"lastModified"`
	IsDir        bool      `json:"isDir"`
}

// syncUp 将本地文件同步到远程
func (a *App) syncUp(config SyncConfig) error {
	fmt.Printf("开始上传同步: %s -> %s\n", config.LocalPath, config.RemotePath)

	// 检查本地路径是否存在
	if _, err := os.Stat(config.LocalPath); os.IsNotExist(err) {
		return fmt.Errorf("本地路径不存在: %s", config.LocalPath)
	}

	// 获取本地文件列表
	localFiles, err := getAllFiles(config.LocalPath)
	if err != nil {
		return fmt.Errorf("获取本地文件列表失败: %v", err)
	}

	// 获取远程文件列表
	remoteFiles, err := a.ListMinioFiles(config.RemotePath)
	if err != nil {
		// 如果远程路径不存在，创建它
		if strings.Contains(err.Error(), "not found") {
			err = a.CreateMinioFolder(config.RemotePath)
			if err != nil {
				return fmt.Errorf("创建远程文件夹失败: %v", err)
			}
			remoteFiles = []MinioFileInfo{}
		} else {
			return fmt.Errorf("获取远程文件列表失败: %v", err)
		}
	}

	// 创建远程文件映射，用于快速查找
	remoteFileMap := make(map[string]MinioFileInfo)
	for _, file := range remoteFiles {
		if !file.IsDir {
			remoteFileMap[file.Path] = file
		}
	}

	// 上传新文件或更新的文件
	uploadCount := 0
	for _, localFile := range localFiles {
		// 计算相对路径
		relPath, err := filepath.Rel(config.LocalPath, localFile)
		if err != nil {
			return fmt.Errorf("计算相对路径失败: %v", err)
		}

		// 转换为远程路径
		remotePath := filepath.Join(config.RemotePath, relPath)
		remotePath = strings.ReplaceAll(remotePath, "\\", "/")

		// 获取本地文件的修改时间
		localModTime, err := getFileModTime(localFile)
		if err != nil {
			return fmt.Errorf("获取本地文件修改时间失败: %v", err)
		}

		// 检查远程文件是否存在
		remoteFile, exists := remoteFileMap[remotePath]
		if !exists {
			// 文件不存在，上传
			fmt.Printf("上传新文件: %s -> %s\n", localFile, remotePath)
			err = a.UploadFileToMinio(localFile, remotePath)
			if err != nil {
				return fmt.Errorf("上传文件失败: %v", err)
			}
			uploadCount++
		} else {
			// 文件存在，检查修改时间
			if localModTime.After(remoteFile.LastModified) {
				// 本地文件更新，上传
				fmt.Printf("上传更新的文件: %s -> %s\n", localFile, remotePath)
				err = a.UploadFileToMinio(localFile, remotePath)
				if err != nil {
					return fmt.Errorf("上传文件失败: %v", err)
				}
				uploadCount++
			}
		}
	}

	fmt.Printf("上传同步完成，共上传 %d 个文件\n", uploadCount)
	return nil
}

// syncDown 将远程文件同步到本地
func (a *App) syncDown(config SyncConfig) error {
	fmt.Printf("开始下载同步: %s -> %s\n", config.RemotePath, config.LocalPath)

	// 确保本地路径存在
	if err := os.MkdirAll(config.LocalPath, 0755); err != nil {
		return fmt.Errorf("创建本地目录失败: %v", err)
	}

	// 获取远程文件列表
	remoteFiles, err := a.ListMinioFiles(config.RemotePath)
	if err != nil {
		return fmt.Errorf("获取远程文件列表失败: %v", err)
	}

	// 获取本地文件列表
	localFiles, err := getAllFiles(config.LocalPath)
	if err != nil {
		return fmt.Errorf("获取本地文件列表失败: %v", err)
	}

	// 创建本地文件映射，用于快速查找
	localFileMap := make(map[string]string)
	for _, file := range localFiles {
		// 计算相对路径
		relPath, err := filepath.Rel(config.LocalPath, file)
		if err != nil {
			return fmt.Errorf("计算相对路径失败: %v", err)
		}
		localFileMap[relPath] = file
	}

	// 下载新文件或更新的文件
	downloadCount := 0
	for _, remoteFile := range remoteFiles {
		// 跳过目录
		if remoteFile.IsDir {
			continue
		}

		// 计算相对路径
		relPath := strings.TrimPrefix(remoteFile.Path, config.RemotePath)
		if strings.HasPrefix(relPath, "/") {
			relPath = relPath[1:]
		}

		// 转换为本地路径
		localPath := filepath.Join(config.LocalPath, relPath)

		// 检查本地文件是否存在
		existingLocalPath, exists := localFileMap[relPath]
		if !exists {
			// 文件不存在，下载
			fmt.Printf("下载新文件: %s -> %s\n", remoteFile.Path, localPath)

			// 确保本地目录存在
			localDir := filepath.Dir(localPath)
			if err := os.MkdirAll(localDir, 0755); err != nil {
				return fmt.Errorf("创建本地目录失败: %v", err)
			}

			// 下载文件
			data, err := a.DownloadFileFromMinio(remoteFile.Path)
			if err != nil {
				return fmt.Errorf("下载文件失败: %v", err)
			}

			// 写入本地文件
			err = os.WriteFile(localPath, data, 0644)
			if err != nil {
				return fmt.Errorf("写入本地文件失败: %v", err)
			}

			downloadCount++
		} else {
			// 文件存在，检查修改时间
			localModTime, err := getFileModTime(existingLocalPath)
			if err != nil {
				return fmt.Errorf("获取本地文件修改时间失败: %v", err)
			}

			if remoteFile.LastModified.After(localModTime) {
				// 远程文件更新，下载
				fmt.Printf("下载更新的文件: %s -> %s\n", remoteFile.Path, localPath)

				// 下载文件
				data, err := a.DownloadFileFromMinio(remoteFile.Path)
				if err != nil {
					return fmt.Errorf("下载文件失败: %v", err)
				}

				// 写入本地文件
				err = os.WriteFile(localPath, data, 0644)
				if err != nil {
					return fmt.Errorf("写入本地文件失败: %v", err)
				}

				downloadCount++
			}
		}
	}

	fmt.Printf("下载同步完成，共下载 %d 个文件\n", downloadCount)
	return nil
}

// fullSync 执行完整同步
func (a *App) fullSync(status *SyncStatus) error {
	fmt.Println("执行完整同步...")

	// 获取所有同步规则
	rules := a.GetSyncRules()
	if len(rules) == 0 {
		return fmt.Errorf("没有同步规则")
	}

	// 遍历所有规则
	for _, rule := range rules {
		// 跳过禁用的规则
		if !rule.Enabled {
			continue
		}

		// 创建同步配置
		config := SyncConfig{
			LocalPath:  rule.LocalPath,
			RemotePath: rule.RemotePath,
			Direction:  rule.Direction,
			Interval:   60, // 默认60秒
		}

		// 检测冲突
		conflicts, err := a.detectConflicts(rule.LocalPath, rule.RemotePath)
		if err != nil {
			status.Errors = append(status.Errors, fmt.Sprintf("检测冲突失败: %v", err))
		} else if len(conflicts) > 0 {
			// 添加到冲突列表
			a.conflictFiles = append(a.conflictFiles, conflicts...)
			status.ConflictCount += len(conflicts)
			
			// 如果有冲突且默认解决方式不是询问，自动解决冲突
			if a.defaultConflictResolution != "ask" {
				for _, conflict := range conflicts {
					err := a.ResolveConflict(conflict.Path, a.defaultConflictResolution)
					if err != nil {
						status.Errors = append(status.Errors, fmt.Sprintf("解决冲突失败: %v", err))
					}
				}
			}
		}

		// 根据方向执行同步
		switch rule.Direction {
		case "upload":
			err = a.syncUp(config)
			if err == nil {
				status.FilesUploaded++
			}
		case "download":
			err = a.syncDown(config)
			if err == nil {
				status.FilesDownloaded++
			}
		case "bidirectional":
			// 先上传再下载
			err = a.syncUp(config)
			if err == nil {
				status.FilesUploaded++
			} else {
				status.Errors = append(status.Errors, fmt.Sprintf("上传同步失败: %v", err))
			}

			err = a.syncDown(config)
			if err == nil {
				status.FilesDownloaded++
			}
		default:
			err = fmt.Errorf("无效的同步方向: %s", rule.Direction)
		}

		if err != nil {
			status.Errors = append(status.Errors, fmt.Sprintf("同步规则 '%s' 失败: %v", rule.Name, err))
		}
	}

	// 更新最后同步时间
	a.lastSyncTime = time.Now()

	return nil
}

// selectiveSync 执行选择性同步
func (a *App) selectiveSync(status *SyncStatus) error {
	fmt.Println("执行选择性同步...")

	// 获取所有同步规则
	rules := a.GetSyncRules()
	if len(rules) == 0 {
		return fmt.Errorf("没有同步规则")
	}

	// 遍历所有规则
	for _, rule := range rules {
		// 跳过禁用的规则
		if !rule.Enabled {
			continue
		}

		// 创建同步配置
		config := SyncConfig{
			LocalPath:  rule.LocalPath,
			RemotePath: rule.RemotePath,
			Direction:  rule.Direction,
			Interval:   60, // 默认60秒
		}

		// 获取本地文件列表
		localFiles, err := getAllFiles(config.LocalPath)
		if err != nil {
			status.Errors = append(status.Errors, fmt.Sprintf("获取本地文件列表失败: %v", err))
			continue
		}

		// 应用过滤规则
		var filteredFiles []string
		for _, file := range localFiles {
			// 检查是否匹配过滤规则
			if !matchesFilter(file, rule.Filters) {
				filteredFiles = append(filteredFiles, file)
			}
		}

		// 根据方向执行同步
		switch rule.Direction {
		case "upload":
			// 只上传过滤后的文件
			for _, file := range filteredFiles {
				// 计算相对路径
				relPath, err := filepath.Rel(config.LocalPath, file)
				if err != nil {
					status.Errors = append(status.Errors, fmt.Sprintf("计算相对路径失败: %v", err))
					continue
				}

				// 转换为远程路径
				remotePath := filepath.Join(config.RemotePath, relPath)
				remotePath = strings.ReplaceAll(remotePath, "\\", "/")

				// 上传文件
				err = a.UploadFileToMinio(file, remotePath)
				if err != nil {
					status.Errors = append(status.Errors, fmt.Sprintf("上传文件失败: %v", err))
				} else {
					status.FilesUploaded++
				}
			}

		case "download":
			// 执行下载同步
			err = a.syncDown(config)
			if err != nil {
				status.Errors = append(status.Errors, fmt.Sprintf("下载同步失败: %v", err))
			} else {
				status.FilesDownloaded++
			}

		case "bidirectional":
			// 只上传过滤后的文件
			for _, file := range filteredFiles {
				// 计算相对路径
				relPath, err := filepath.Rel(config.LocalPath, file)
				if err != nil {
					status.Errors = append(status.Errors, fmt.Sprintf("计算相对路径失败: %v", err))
					continue
				}

				// 转换为远程路径
				remotePath := filepath.Join(config.RemotePath, relPath)
				remotePath = strings.ReplaceAll(remotePath, "\\", "/")

				// 上传文件
				err = a.UploadFileToMinio(file, remotePath)
				if err != nil {
					status.Errors = append(status.Errors, fmt.Sprintf("上传文件失败: %v", err))
				} else {
					status.FilesUploaded++
				}
			}

			// 执行下载同步
			err = a.syncDown(config)
			if err != nil {
				status.Errors = append(status.Errors, fmt.Sprintf("下载同步失败: %v", err))
			} else {
				status.FilesDownloaded++
			}

		default:
			status.Errors = append(status.Errors, fmt.Sprintf("无效的同步方向: %s", rule.Direction))
		}
	}

	return nil
}

// backupSync 执行备份同步
func (a *App) backupSync(status *SyncStatus) error {
	fmt.Println("执行备份同步...")

	// 获取所有同步规则
	rules := a.GetSyncRules()
	if len(rules) == 0 {
		return fmt.Errorf("没有同步规则")
	}

	// 遍历所有规则
	for _, rule := range rules {
		// 跳过禁用的规则
		if !rule.Enabled {
			continue
		}

		// 创建同步配置
		config := SyncConfig{
			LocalPath:  rule.LocalPath,
			RemotePath: rule.RemotePath,
			Direction:  rule.Direction,
			Interval:   60, // 默认60秒
		}

		// 创建备份文件夹
		backupPath := filepath.Join(config.RemotePath, fmt.Sprintf("backup_%s", time.Now().Format("20060102_150405")))
		err := a.CreateMinioFolder(backupPath)
		if err != nil {
			status.Errors = append(status.Errors, fmt.Sprintf("创建备份文件夹失败: %v", err))
			continue
		}

		// 获取本地文件列表
		localFiles, err := getAllFiles(config.LocalPath)
		if err != nil {
			status.Errors = append(status.Errors, fmt.Sprintf("获取本地文件列表失败: %v", err))
			continue
		}

		// 上传文件到备份文件夹
		for _, file := range localFiles {
			// 计算相对路径
			relPath, err := filepath.Rel(config.LocalPath, file)
			if err != nil {
				status.Errors = append(status.Errors, fmt.Sprintf("计算相对路径失败: %v", err))
				continue
			}

			// 转换为远程路径
			remotePath := filepath.Join(backupPath, relPath)
			remotePath = strings.ReplaceAll(remotePath, "\\", "/")

			// 上传文件
			err = a.UploadFileToMinio(file, remotePath)
			if err != nil {
				status.Errors = append(status.Errors, fmt.Sprintf("上传文件失败: %v", err))
			} else {
				status.FilesUploaded++
			}
		}
	}

	return nil
}

// GetConflictFiles 获取冲突文件
func (a *App) GetConflictFiles() []ConflictFile {
	return a.conflictFiles
}

// 获取文件的最后修改时间
func getFileModTime(path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

// 递归获取目录下所有文件
func getAllFiles(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// 检查文件是否匹配过滤规则
func matchesFilter(file string, filters []string) bool {
	if len(filters) == 0 {
		return false
	}

	filename := filepath.Base(file)
	for _, filter := range filters {
		// 简单的通配符匹配
		if strings.HasPrefix(filter, "*") {
			suffix := filter[1:]
			if strings.HasSuffix(filename, suffix) {
				return true
			}
		} else if strings.HasSuffix(filter, "*") {
			prefix := filter[:len(filter)-1]
			if strings.HasPrefix(filename, prefix) {
				return true
			}
		} else if filter == filename {
			return true
		}
	}

	return false
}