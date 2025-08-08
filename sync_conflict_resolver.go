package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 冲突解决方式常量
const (
	ConflictResolutionLocal  = "local"  // 使用本地文件
	ConflictResolutionRemote = "remote" // 使用远程文件
	ConflictResolutionBoth   = "both"   // 保留两者
	ConflictResolutionSkip   = "skip"   // 跳过
	ConflictResolutionAsk    = "ask"    // 询问用户
)

// detectConflicts 检测同步冲突
func (a *App) detectConflicts(localPath, remotePath string) ([]ConflictFile, error) {
	var conflicts []ConflictFile
	
	// 获取本地文件列表
	localFiles, err := getAllFiles(localPath)
	if err != nil {
		return nil, fmt.Errorf("获取本地文件列表失败: %v", err)
	}
	
	// 获取远程文件列表
	remoteFiles, err := a.ListMinioFiles(remotePath)
	if err != nil {
		return nil, fmt.Errorf("获取远程文件列表失败: %v", err)
	}
	
	// 创建远程文件映射，用于快速查找
	remoteFileMap := make(map[string]MinioFileInfo)
	for _, file := range remoteFiles {
		if !file.IsDir {
			remoteFileMap[file.Path] = file
		}
	}
	
	// 检查每个本地文件
	for _, localFile := range localFiles {
		// 计算相对路径
		relPath, err := filepath.Rel(localPath, localFile)
		if err != nil {
			return nil, fmt.Errorf("计算相对路径失败: %v", err)
		}
		
		// 转换为远程路径
		remotePath := filepath.Join(remotePath, relPath)
		remotePath = strings.ReplaceAll(remotePath, "\\", "/")
		
		// 检查远程文件是否存在
		remoteFile, exists := remoteFileMap[remotePath]
		if !exists {
			continue // 远程文件不存在，不是冲突
		}
		
		// 获取本地文件的修改时间
		localModTime, err := getFileModTime(localFile)
		if err != nil {
			return nil, fmt.Errorf("获取本地文件修改时间失败: %v", err)
		}
		
		// 获取远程文件的修改时间
		remoteModTime := remoteFile.LastModified
		
		// 检查修改时间
		if localModTime.After(a.lastSyncTime) && remoteModTime.After(a.lastSyncTime) {
			// 本地和远程都有修改，这是一个潜在冲突
			
			// 计算本地文件的校验和
			localChecksum, err := calculateMD5(localFile)
			if err != nil {
				return nil, fmt.Errorf("计算本地文件校验和失败: %v", err)
			}
			
			// 下载远程文件并计算校验和
			remoteData, err := a.DownloadFileFromMinio(remotePath)
			if err != nil {
				return nil, fmt.Errorf("下载远程文件失败: %v", err)
			}
			
			remoteChecksum, err := calculateMD5FromBytes(remoteData)
			if err != nil {
				return nil, fmt.Errorf("计算远程文件校验和失败: %v", err)
			}
			
			// 如果校验和相同，不是冲突
			if localChecksum == remoteChecksum {
				continue
			}
			
			conflicts = append(conflicts, ConflictFile{
				Path:          localFile,
				LocalModTime:  localModTime,
				RemoteModTime: remoteModTime,
				Resolution:    "pending",
			})
		}
	}
	
	return conflicts, nil
}

// ResolveConflict 解决单个冲突
func (a *App) ResolveConflict(path string, resolution string) error {
	// 验证解决方式
	validResolutions := []string{
		ConflictResolutionLocal,
		ConflictResolutionRemote,
		ConflictResolutionBoth,
		ConflictResolutionSkip,
	}
	
	valid := false
	for _, r := range validResolutions {
		if r == resolution {
			valid = true
			break
		}
	}
	
	if !valid {
		return fmt.Errorf("无效的冲突解决方式: %s", resolution)
	}
	
	// 查找冲突文件
	var conflict *ConflictFile
	for i, c := range a.conflictFiles {
		if c.Path == path {
			conflict = &a.conflictFiles[i]
			break
		}
	}
	
	if conflict == nil {
		return fmt.Errorf("未找到冲突文件: %s", path)
	}
	
	// 根据解决方式处理冲突
	switch resolution {
	case ConflictResolutionLocal:
		// 使用本地文件，上传到远程
		remotePath := a.getRemotePathForLocalFile(conflict.Path)
		err := a.UploadFileToMinio(conflict.Path, remotePath)
		if err != nil {
			return fmt.Errorf("上传文件失败: %v", err)
		}
		conflict.Resolution = ConflictResolutionLocal
		
	case ConflictResolutionRemote:
		// 使用远程文件，下载到本地
		remotePath := a.getRemotePathForLocalFile(conflict.Path)
		data, err := a.DownloadFileFromMinio(remotePath)
		if err != nil {
			return fmt.Errorf("下载文件失败: %v", err)
		}
		err = os.WriteFile(conflict.Path, data, 0644)
		if err != nil {
			return fmt.Errorf("写入本地文件失败: %v", err)
		}
		conflict.Resolution = ConflictResolutionRemote
		
	case ConflictResolutionBoth:
		// 保留两者，重命名本地文件
		localDir := filepath.Dir(conflict.Path)
		localBase := filepath.Base(conflict.Path)
		ext := filepath.Ext(localBase)
		name := localBase[:len(localBase)-len(ext)]
		newLocalPath := filepath.Join(localDir, fmt.Sprintf("%s_local_%s%s", name, time.Now().Format("20060102_150405"), ext))
		
		// 重命名本地文件
		err := os.Rename(conflict.Path, newLocalPath)
		if err != nil {
			return fmt.Errorf("重命名本地文件失败: %v", err)
		}
		
		// 下载远程文件到原路径
		remotePath := a.getRemotePathForLocalFile(conflict.Path)
		data, err := a.DownloadFileFromMinio(remotePath)
		if err != nil {
			return fmt.Errorf("下载文件失败: %v", err)
		}
		err = os.WriteFile(conflict.Path, data, 0644)
		if err != nil {
			return fmt.Errorf("写入本地文件失败: %v", err)
		}
		conflict.Resolution = ConflictResolutionBoth
		
	case ConflictResolutionSkip:
		// 跳过，不做任何操作
		conflict.Resolution = ConflictResolutionSkip
	}
	
	return nil
}

// ResolveAllConflicts 解决所有冲突
func (a *App) ResolveAllConflicts(resolution string) error {
	if len(a.conflictFiles) == 0 {
		return nil
	}
	
	// 验证解决方式
	validResolutions := []string{
		ConflictResolutionLocal,
		ConflictResolutionRemote,
		ConflictResolutionBoth,
		ConflictResolutionSkip,
	}
	
	valid := false
	for _, r := range validResolutions {
		if r == resolution {
			valid = true
			break
		}
	}
	
	if !valid {
		return fmt.Errorf("无效的冲突解决方式: %s", resolution)
	}
	
	// 解决所有冲突
	for _, conflict := range a.conflictFiles {
		if conflict.Resolution == "pending" {
			err := a.ResolveConflict(conflict.Path, resolution)
			if err != nil {
				return fmt.Errorf("解决冲突失败: %v", err)
			}
		}
	}
	
	return nil
}

// getRemotePathForLocalFile 获取本地文件对应的远程路径
func (a *App) getRemotePathForLocalFile(localPath string) string {
	// 查找匹配的同步规则
	for _, rule := range a.syncRules {
		if strings.HasPrefix(localPath, rule.LocalPath) {
			// 计算相对路径
			relPath, err := filepath.Rel(rule.LocalPath, localPath)
			if err != nil {
				continue
			}
			
			// 转换为远程路径
			remotePath := filepath.Join(rule.RemotePath, relPath)
			remotePath = strings.ReplaceAll(remotePath, "\\", "/")
			return remotePath
		}
	}
	
	// 如果没有找到匹配的规则，使用默认路径
	return filepath.Base(localPath)
}

// GetConflictCount 获取冲突文件数量
func (a *App) GetConflictCount() int {
	return len(a.conflictFiles)
}

// HasPendingConflicts 检查是否有待解决的冲突
func (a *App) HasPendingConflicts() bool {
	for _, conflict := range a.conflictFiles {
		if conflict.Resolution == "pending" {
			return true
		}
	}
	return false
}