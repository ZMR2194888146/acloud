package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// incrementalSync 执行增量同步
func (a *App) incrementalSync(status *SyncStatus) error {
	fmt.Println("执行增量同步...")

	// 获取所有同步规则
	rules := a.GetSyncRules()
	if len(rules) == 0 {
		return fmt.Errorf("没有同步规则")
	}

	// 获取上次同步时间
	lastSyncTime := a.lastSyncTime

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

		// 根据方向执行同步
		switch rule.Direction {
		case "upload":
			err := a.incrementalSyncUp(config, lastSyncTime, status, rule.Filters)
			if err != nil {
				status.Errors = append(status.Errors, fmt.Sprintf("上传同步失败: %v", err))
			}
		case "download":
			err := a.incrementalSyncDown(config, lastSyncTime, status, rule.Filters)
			if err != nil {
				status.Errors = append(status.Errors, fmt.Sprintf("下载同步失败: %v", err))
			}
		case "bidirectional":
			// 先上传再下载
			err := a.incrementalSyncUp(config, lastSyncTime, status, rule.Filters)
			if err != nil {
				status.Errors = append(status.Errors, fmt.Sprintf("上传同步失败: %v", err))
			}

			err = a.incrementalSyncDown(config, lastSyncTime, status, rule.Filters)
			if err != nil {
				status.Errors = append(status.Errors, fmt.Sprintf("下载同步失败: %v", err))
			}
		default:
			status.Errors = append(status.Errors, fmt.Sprintf("无效的同步方向: %s", rule.Direction))
		}
	}

	return nil
}

// incrementalSyncUp 执行增量上传同步
func (a *App) incrementalSyncUp(config SyncConfig, lastSyncTime time.Time, status *SyncStatus, filters []string) error {
	fmt.Printf("开始增量上传同步: %s -> %s\n", config.LocalPath, config.RemotePath)

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
		// 检查是否匹配过滤规则
		if matchesFilter(localFile, filters) {
			continue
		}

		// 获取文件修改时间
		fileInfo, err := os.Stat(localFile)
		if err != nil {
			status.Errors = append(status.Errors, fmt.Sprintf("获取文件信息失败: %v", err))
			continue
		}

		// 如果文件在上次同步后被修改，则上传
		if fileInfo.ModTime().After(lastSyncTime) {
			// 计算相对路径
			relPath, err := filepath.Rel(config.LocalPath, localFile)
			if err != nil {
				status.Errors = append(status.Errors, fmt.Sprintf("计算相对路径失败: %v", err))
				continue
			}

			// 转换为远程路径
			remotePath := filepath.Join(config.RemotePath, relPath)
			remotePath = strings.ReplaceAll(remotePath, "\\", "/")

			// 检查远程文件是否存在
			remoteFile, exists := remoteFileMap[remotePath]
			if !exists || fileInfo.ModTime().After(remoteFile.LastModified) {
				// 文件不存在或本地文件更新，上传
				fmt.Printf("上传文件: %s -> %s\n", localFile, remotePath)
				err = a.UploadFileToMinio(localFile, remotePath)
				if err != nil {
					status.Errors = append(status.Errors, fmt.Sprintf("上传文件失败: %v", err))
				} else {
					uploadCount++
					status.FilesUploaded++
				}
			}
		}
	}

	fmt.Printf("增量上传同步完成，共上传 %d 个文件\n", uploadCount)
	return nil
}

// incrementalSyncDown 执行增量下载同步
func (a *App) incrementalSyncDown(config SyncConfig, lastSyncTime time.Time, status *SyncStatus, filters []string) error {
	fmt.Printf("开始增量下载同步: %s -> %s\n", config.RemotePath, config.LocalPath)

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

		// 如果文件在上次同步后被修改，则下载
		if remoteFile.LastModified.After(lastSyncTime) {
			// 计算相对路径
			relPath := strings.TrimPrefix(remoteFile.Path, config.RemotePath)
			if strings.HasPrefix(relPath, "/") {
				relPath = relPath[1:]
			}

			// 检查是否匹配过滤规则
			if matchesFilter(relPath, filters) {
				continue
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
					status.Errors = append(status.Errors, fmt.Sprintf("创建本地目录失败: %v", err))
					continue
				}

				// 下载文件
				data, err := a.DownloadFileFromMinio(remoteFile.Path)
				if err != nil {
					status.Errors = append(status.Errors, fmt.Sprintf("下载文件失败: %v", err))
					continue
				}

				// 写入本地文件
				err = os.WriteFile(localPath, data, 0644)
				if err != nil {
					status.Errors = append(status.Errors, fmt.Sprintf("写入本地文件失败: %v", err))
					continue
				}

				downloadCount++
				status.FilesDownloaded++
			} else {
				// 文件存在，检查修改时间
				localModTime, err := getFileModTime(existingLocalPath)
				if err != nil {
					status.Errors = append(status.Errors, fmt.Sprintf("获取本地文件修改时间失败: %v", err))
					continue
				}

				if remoteFile.LastModified.After(localModTime) {
					// 远程文件更新，下载
					fmt.Printf("下载更新的文件: %s -> %s\n", remoteFile.Path, localPath)

					// 下载文件
					data, err := a.DownloadFileFromMinio(remoteFile.Path)
					if err != nil {
						status.Errors = append(status.Errors, fmt.Sprintf("下载文件失败: %v", err))
						continue
					}

					// 写入本地文件
					err = os.WriteFile(localPath, data, 0644)
					if err != nil {
						status.Errors = append(status.Errors, fmt.Sprintf("写入本地文件失败: %v", err))
						continue
					}

					downloadCount++
					status.FilesDownloaded++
				}
			}
		}
	}

	fmt.Printf("增量下载同步完成，共下载 %d 个文件\n", downloadCount)
	return nil
}


