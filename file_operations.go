package main

import (
	"encoding/base64"
	"fmt"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// ListFiles 列出指定目录下的所有文件和文件夹
func (a *App) ListFiles(dirPath string) ([]FileInfo, error) {
	// 检查用户是否已登录
	if !a.isLoggedIn {
		return nil, fmt.Errorf("用户未登录")
	}
	
	// 如果路径为空，则使用根存储路径
	fullPath := a.storagePath
	if dirPath != "" {
		fullPath = filepath.Join(a.storagePath, dirPath)
	}
	
	// 检查路径是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("路径不存在")
	}
	
	// 读取目录内容
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}
	
	// 转换为 FileInfo 结构体
	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		
		relativePath := filepath.Join(dirPath, entry.Name())
		files = append(files, FileInfo{
			Name:      entry.Name(),
			Path:      relativePath,
			Size:      info.Size(),
			IsDir:     entry.IsDir(),
			UpdatedAt: info.ModTime(),
		})
	}
	
	// 按照文件夹在前，文件在后的顺序排序
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir && !files[j].IsDir {
			return true
		}
		if !files[i].IsDir && files[j].IsDir {
			return false
		}
		return files[i].Name < files[j].Name
	})
	
	return files, nil
}

// CreateFolder 创建新文件夹
func (a *App) CreateFolder(path, name string) error {
	// 检查用户是否已登录
	if !a.isLoggedIn {
		return fmt.Errorf("用户未登录")
	}
	
	fullPath := filepath.Join(a.storagePath, path, name)
	return os.MkdirAll(fullPath, 0755)
}

// DeleteFile 删除文件或文件夹
func (a *App) DeleteFile(path string) error {
	// 检查用户是否已登录
	if !a.isLoggedIn {
		return fmt.Errorf("用户未登录")
	}
	
	fullPath := filepath.Join(a.storagePath, path)
	return os.RemoveAll(fullPath)
}

// RenameFile 重命名文件或文件夹
func (a *App) RenameFile(oldPath, newName string) error {
	// 检查用户是否已登录
	if !a.isLoggedIn {
		return fmt.Errorf("用户未登录")
	}
	
	oldFullPath := filepath.Join(a.storagePath, oldPath)
	dir := filepath.Dir(oldFullPath)
	newFullPath := filepath.Join(dir, newName)
	return os.Rename(oldFullPath, newFullPath)
}

// SaveFile 保存上传的文件
func (a *App) SaveFile(path, name string, content []byte) error {
	// 检查用户是否已登录
	if !a.isLoggedIn {
		return fmt.Errorf("用户未登录")
	}
	
	dirPath := filepath.Join(a.storagePath, path)
	
	// 确保目录存在
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return err
	}
	
	filePath := filepath.Join(dirPath, name)
	return os.WriteFile(filePath, content, 0644)
}

// UploadFile 处理文件上传
func (a *App) UploadFile(path string, fileData []byte, fileName string) error {
	// 检查用户是否已登录
	if !a.isLoggedIn {
		return fmt.Errorf("用户未登录")
	}
	
	dirPath := filepath.Join(a.storagePath, path)
	
	// 确保目录存在
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return err
	}
	
	filePath := filepath.Join(dirPath, fileName)
	return os.WriteFile(filePath, fileData, 0644)
}

// UploadFileString 处理文件上传（字符串版本）
func (a *App) UploadFileString(path string, fileContent string, fileName string) error {
	// 检查用户是否已登录
	if !a.isLoggedIn {
		return fmt.Errorf("用户未登录")
	}
	
	dirPath := filepath.Join(a.storagePath, path)
	
	// 确保目录存在
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return err
	}
	
	filePath := filepath.Join(dirPath, fileName)
	return os.WriteFile(filePath, []byte(fileContent), 0644)
}

// DownloadFile 获取文件内容用于下载
func (a *App) DownloadFile(path string) ([]byte, error) {
	// 检查用户是否已登录
	if !a.isLoggedIn {
		return nil, fmt.Errorf("用户未登录")
	}
	
	fullPath := filepath.Join(a.storagePath, path)
	return os.ReadFile(fullPath)
}

// ReadFile 读取文件内容
func (a *App) ReadFile(path string) ([]byte, error) {
	// 检查用户是否已登录
	if !a.isLoggedIn {
		return nil, fmt.Errorf("用户未登录")
	}
	
	fullPath := filepath.Join(a.storagePath, path)
	return os.ReadFile(fullPath)
}

// GetFileType 获取文件的MIME类型
func (a *App) GetFileType(path string) (string, error) {
	// 检查用户是否已登录
	if !a.isLoggedIn {
		return "", fmt.Errorf("用户未登录")
	}
	
	fullPath := filepath.Join(a.storagePath, path)
	
	// 通过扩展名判断文件类型
	ext := strings.ToLower(filepath.Ext(fullPath))
	mimeType := mime.TypeByExtension(ext)
	
	// 如果无法通过扩展名确定，则读取文件头部进行判断
	if mimeType == "" {
		// 读取文件头部
		file, err := os.Open(fullPath)
		if err != nil {
			return "", err
		}
		defer file.Close()
		
		// 读取前512字节用于判断文件类型
		buffer := make([]byte, 512)
		_, err = file.Read(buffer)
		if err != nil {
			return "", err
		}
		
		// 通过文件头部判断MIME类型
		mimeType = http.DetectContentType(buffer)
	}
	
	return mimeType, nil
}

// GetFilePreview 获取文件预览信息
func (a *App) GetFilePreview(path string) (*FilePreview, error) {
	// 检查用户是否已登录
	if !a.isLoggedIn {
		return nil, fmt.Errorf("用户未登录")
	}
	
	fullPath := filepath.Join(a.storagePath, path)
	
	// 获取文件类型
	mimeType, err := a.GetFileType(path)
	if err != nil {
		return nil, err
	}
	
	// 读取文件内容
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}
	
	// 判断是否为文本文件
	isText := strings.HasPrefix(mimeType, "text/") || 
		mimeType == "application/json" || 
		mimeType == "application/xml" ||
		mimeType == "application/javascript" ||
		strings.HasSuffix(path, ".md") ||
		strings.HasSuffix(path, ".csv") ||
		strings.HasSuffix(path, ".txt")
	
	// 判断是否为代码文件
	isCode := strings.HasSuffix(path, ".go") ||
		strings.HasSuffix(path, ".py") ||
		strings.HasSuffix(path, ".java") ||
		strings.HasSuffix(path, ".c") ||
		strings.HasSuffix(path, ".cpp") ||
		strings.HasSuffix(path, ".h") ||
		strings.HasSuffix(path, ".js") ||
		strings.HasSuffix(path, ".ts") ||
		strings.HasSuffix(path, ".php") ||
		strings.HasSuffix(path, ".rb")
	
	preview := &FilePreview{
		MimeType: mimeType,
		IsBase64: !(isText || isCode),
	}
	
	// 如果是文本文件或代码文件，直接返回文本内容
	if isText || isCode {
		preview.Content = string(content)
	} else {
		// 如果是二进制文件，转换为Base64编码
		preview.Content = base64.StdEncoding.EncodeToString(content)
	}
	
	fmt.Printf("文件预览: 路径=%s, 类型=%s, 是否Base64=%v\n", path, mimeType, preview.IsBase64)
	
	return preview, nil
}

// OpenInExplorer 在系统资源管理器中打开指定路径
func (a *App) OpenInExplorer(path string) error {
	// 检查用户是否已登录
	if !a.isLoggedIn {
		return fmt.Errorf("用户未登录")
	}
	
	var fullPath string
	if path == "" {
		// 如果路径为空，打开根存储目录
		fullPath = a.storagePath
	} else {
		// 构建完整路径
		fullPath = filepath.Join(a.storagePath, path)
	}
	
	// 检查路径是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("路径不存在: %s", fullPath)
	}
	
	// 根据操作系统打开资源管理器
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", fullPath)
	case "darwin": // macOS
		cmd = exec.Command("open", fullPath)
	case "linux":
		cmd = exec.Command("xdg-open", fullPath)
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
	
	return cmd.Start()
}

// OpenFileInExplorer 在系统资源管理器中打开文件所在的目录并选中文件
func (a *App) OpenFileInExplorer(filePath string) error {
	// 检查用户是否已登录
	if !a.isLoggedIn {
		return fmt.Errorf("用户未登录")
	}
	
	// 构建完整路径
	fullPath := filepath.Join(a.storagePath, filePath)
	
	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在: %s", fullPath)
	}
	
	// 根据操作系统打开资源管理器并选中文件
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", "/select,", fullPath)
	case "darwin": // macOS
		cmd = exec.Command("open", "-R", fullPath)
	case "linux":
		// Linux 下打开包含文件的目录
		dir := filepath.Dir(fullPath)
		cmd = exec.Command("xdg-open", dir)
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}
	
	return cmd.Start()
}