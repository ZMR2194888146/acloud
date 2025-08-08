package main

import (
	"context"
	"fmt"
	"mime"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

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

// ListMinioFilesByBucket 按存储桶列出MinIO中的文件
func (a *App) ListMinioFilesByBucket(bucketName, path string) ([]MinioFileInfo, error) {
	// 检查用户是否已登录
	if !a.isLoggedIn {
		return nil, fmt.Errorf("用户未登录")
	}

	// 检查 MinIO 是否已启用
	if !a.minioConfig.Enabled || a.minioClient == nil {
		return nil, fmt.Errorf("MinIO 未启用")
	}

	// 确保路径以斜杠结尾
	if path != "" && !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

	// 列出对象
	var files []MinioFileInfo

	// 创建通道接收对象信息
	objectCh := a.minioClient.ListObjects(context.Background(), bucketName, minio.ListObjectsOptions{
		Prefix:    path,
		Recursive: false, // 只列出当前目录的内容
	})

	// 遍历对象
	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("列出对象失败: %v", object.Err)
		}

		// 跳过当前目录
		if object.Key == path {
			continue
		}

		// 判断是否是目录
		isDir := strings.HasSuffix(object.Key, "/")

		// 添加到文件列表
		files = append(files, MinioFileInfo{
			Name:         filepath.Base(object.Key),
			Path:         object.Key,
			Size:         object.Size,
			LastModified: object.LastModified,
			IsDir:        isDir,
		})
	}

	return files, nil
}
