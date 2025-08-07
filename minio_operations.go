package main

import (
	"context"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// 初始化 MinIO 客户端
func (a *App) initMinioClient() error {
	// 创建 MinIO 客户端
	client, err := minio.New(a.minioConfig.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(a.minioConfig.AccessKeyID, a.minioConfig.SecretAccessKey, ""),
		Secure: a.minioConfig.UseSSL,
	})
	
	if err != nil {
		return err
	}
	
	a.minioClient = client
	
	// 检查存储桶是否存在
	exists, err := client.BucketExists(context.Background(), a.minioConfig.BucketName)
	if err != nil {
		return err
	}
	
	// 如果存储桶不存在，创建它
	if !exists {
		err = client.MakeBucket(context.Background(), a.minioConfig.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
	}
	
	return nil
}

// GetMinioConfig 获取 MinIO 配置
func (a *App) GetMinioConfig() MinioConfig {
	return a.minioConfig
}

// UpdateMinioConfig 更新 MinIO 配置
func (a *App) UpdateMinioConfig(endpoint, accessKeyID, secretAccessKey, bucketName string, useSSL, enabled bool) error {
	// 检查用户是否已登录
	if !a.isLoggedIn {
		return fmt.Errorf("用户未登录")
	}
	
	// 更新配置
	a.minioConfig = MinioConfig{
		Endpoint:        endpoint,
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		UseSSL:          useSSL,
		BucketName:      bucketName,
		Enabled:         enabled,
	}
	
	// 保存配置
	if err := a.saveConfig(); err != nil {
		return err
	}
	
	// 如果启用了 MinIO，初始化客户端
	if enabled {
		return a.initMinioClient()
	}
	
	// 如果禁用了 MinIO，清除客户端
	a.minioClient = nil
	return nil
}

// TestMinioConnection 测试 MinIO 连接
func (a *App) TestMinioConnection(endpoint, accessKeyID, secretAccessKey string, useSSL bool) error {
	// 创建临时 MinIO 客户端
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	
	if err != nil {
		return err
	}
	
	// 列出存储桶，测试连接
	_, err = client.ListBuckets(context.Background())
	return err
}

// ListMinioBuckets 列出 MinIO 存储桶
func (a *App) ListMinioBuckets() ([]string, error) {
	// 检查用户是否已登录
	if !a.isLoggedIn {
		return nil, fmt.Errorf("用户未登录")
	}
	
	// 检查 MinIO 是否已启用
	if !a.minioConfig.Enabled || a.minioClient == nil {
		return nil, fmt.Errorf("MinIO 未启用")
	}
	
	// 列出存储桶
	buckets, err := a.minioClient.ListBuckets(context.Background())
	if err != nil {
		return nil, err
	}
	
	// 提取存储桶名称
	var bucketNames []string
	for _, bucket := range buckets {
		bucketNames = append(bucketNames, bucket.Name)
	}
	
	return bucketNames, nil
}

// ListMinioFiles 列出 MinIO 中的文件（使用默认配置的 bucket）
func (a *App) ListMinioFiles(prefix string) ([]MinioFileInfo, error) {
	return a.ListMinioFilesByBucket(a.minioConfig.BucketName, prefix)
}

// ListMinioFilesByBucket 列出指定 bucket 中的文件
func (a *App) ListMinioFilesByBucket(bucketName, prefix string) ([]MinioFileInfo, error) {
	// 检查用户是否已登录
	if !a.isLoggedIn {
		return nil, fmt.Errorf("用户未登录")
	}
	
	// 检查 MinIO 是否已启用
	if !a.minioConfig.Enabled || a.minioClient == nil {
		return nil, fmt.Errorf("MinIO 未启用")
	}
	
	// 检查 bucket 是否存在
	exists, err := a.minioClient.BucketExists(context.Background(), bucketName)
	if err != nil {
		return nil, fmt.Errorf("检查存储桶失败: %v", err)
	}
	if !exists {
		return nil, fmt.Errorf("存储桶 '%s' 不存在", bucketName)
	}
	
	// 创建一个通道接收对象信息
	objectCh := a.minioClient.ListObjects(context.Background(), bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: false,
	})
	
	// 收集文件信息
	var files []MinioFileInfo
	
	// 处理目录
	prefixParts := strings.Split(strings.Trim(prefix, "/"), "/")
	
	// 添加返回上级目录的选项（如果不是根目录）
	if prefix != "" {
		parentPath := ""
		if len(prefixParts) > 1 {
			parentPath = strings.Join(prefixParts[:len(prefixParts)-1], "/")
		}
		
		files = append(files, MinioFileInfo{
			Name:  "..",
			Path:  parentPath,
			IsDir: true,
		})
	}
	
	// 处理对象
	seenDirs := make(map[string]bool)
	for object := range objectCh {
		if object.Err != nil {
			fmt.Printf("列出对象时出错: %v\n", object.Err)
			continue
		}
		
		// 获取相对路径
		name := strings.TrimPrefix(object.Key, prefix)
		if name == "" {
			continue
		}
		
		// 检查是否为目录
		isDir := strings.HasSuffix(object.Key, "/")
		
		// 如果是文件，但包含在子目录中
		if !isDir && strings.Contains(name, "/") {
			// 提取目录名
			dirName := name[:strings.Index(name, "/")]
			dirPath := prefix + dirName + "/"
			
			// 如果目录还没有添加到列表中
			if !seenDirs[dirName] {
				files = append(files, MinioFileInfo{
					Name:  dirName,
					Path:  dirPath,
					IsDir: true,
				})
				seenDirs[dirName] = true
			}
			continue
		}
		
		// 如果是目录，去掉末尾的斜杠
		if isDir {
			name = strings.TrimSuffix(name, "/")
		}
		
		// 跳过已经处理过的目录
		if isDir && seenDirs[name] {
			continue
		}
		
		// 添加到文件列表
		files = append(files, MinioFileInfo{
			Name:         name,
			Path:         object.Key,
			Size:         object.Size,
			LastModified: object.LastModified,
			IsDir:        isDir,
		})
		
		if isDir {
			seenDirs[name] = true
		}
	}
	
	fmt.Printf("列出存储桶 '%s' 中的文件，前缀: '%s'，找到 %d 个项目\n", bucketName, prefix, len(files))
	
	return files, nil
}

// UploadFileToMinio 上传文件到 MinIO
func (a *App) UploadFileToMinio(localPath, remotePath string) error {
	// 检查用户是否已登录
	if !a.isLoggedIn {
		return fmt.Errorf("用户未登录")
	}
	
	// 检查 MinIO 是否已启用
	if !a.minioConfig.Enabled || a.minioClient == nil {
		return fmt.Errorf("MinIO 未启用")
	}
	
	// 打开本地文件
	file, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}
	
	// 获取文件类型
	contentType := mime.TypeByExtension(filepath.Ext(localPath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	
	// 上传文件
	_, err = a.minioClient.PutObject(context.Background(), a.minioConfig.BucketName, remotePath, file, fileInfo.Size(), minio.PutObjectOptions{
		ContentType: contentType,
	})
	
	return err
}

// UploadDataToMinio 上传数据到 MinIO
func (a *App) UploadDataToMinio(data []byte, remotePath string) error {
	// 检查用户是否已登录
	if !a.isLoggedIn {
		return fmt.Errorf("用户未登录")
	}
	
	// 检查 MinIO 是否已启用
	if !a.minioConfig.Enabled || a.minioClient == nil {
		return fmt.Errorf("MinIO 未启用")
	}
	
	// 获取文件类型
	contentType := mime.TypeByExtension(filepath.Ext(remotePath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	
	// 上传数据
	_, err := a.minioClient.PutObject(
		context.Background(),
		a.minioConfig.BucketName,
		remotePath,
		strings.NewReader(string(data)),
		int64(len(data)),
		minio.PutObjectOptions{ContentType: contentType},
	)
	
	return err
}

// DownloadFileFromMinio 从 MinIO 下载文件
func (a *App) DownloadFileFromMinio(remotePath string) ([]byte, error) {
	// 检查用户是否已登录
	if !a.isLoggedIn {
		return nil, fmt.Errorf("用户未登录")
	}
	
	// 检查 MinIO 是否已启用
	if !a.minioConfig.Enabled || a.minioClient == nil {
		return nil, fmt.Errorf("MinIO 未启用")
	}
	
	// 获取对象
	object, err := a.minioClient.GetObject(context.Background(), a.minioConfig.BucketName, remotePath, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer object.Close()
	
	// 读取对象内容
	return io.ReadAll(object)
}

// DeleteFileFromMinio 从 MinIO 删除文件
func (a *App) DeleteFileFromMinio(remotePath string) error {
	// 检查用户是否已登录
	if !a.isLoggedIn {
		return fmt.Errorf("用户未登录")
	}
	
	// 检查 MinIO 是否已启用
	if !a.minioConfig.Enabled || a.minioClient == nil {
		return fmt.Errorf("MinIO 未启用")
	}
	
	// 删除对象
	return a.minioClient.RemoveObject(context.Background(), a.minioConfig.BucketName, remotePath, minio.RemoveObjectOptions{})
}

// CreateMinioFolder 在 MinIO 中创建文件夹
func (a *App) CreateMinioFolder(folderPath string) error {
	// 检查用户是否已登录
	if !a.isLoggedIn {
		return fmt.Errorf("用户未登录")
	}
	
	// 检查 MinIO 是否已启用
	if !a.minioConfig.Enabled || a.minioClient == nil {
		return fmt.Errorf("MinIO 未启用")
	}
	
	// 确保路径以斜杠结尾
	if !strings.HasSuffix(folderPath, "/") {
		folderPath += "/"
	}
	
	// 创建一个空对象作为文件夹标记
	_, err := a.minioClient.PutObject(
		context.Background(),
		a.minioConfig.BucketName,
		folderPath,
		strings.NewReader(""),
		0,
		minio.PutObjectOptions{ContentType: "application/directory"},
	)
	
	return err
}