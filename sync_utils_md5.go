package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// calculateMD5 计算文件的MD5哈希值
func calculateMD5(filePath string) (string, error) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 创建MD5哈希对象
	hash := md5.New()

	// 将文件内容复制到哈希对象
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("计算哈希值失败: %v", err)
	}

	// 计算哈希值并转换为十六进制字符串
	hashInBytes := hash.Sum(nil)
	hashString := hex.EncodeToString(hashInBytes)

	return hashString, nil
}

// calculateMD5FromBytes 从字节数组计算MD5哈希值
func calculateMD5FromBytes(data []byte) (string, error) {
	// 创建MD5哈希对象
	hash := md5.New()

	// 将数据写入哈希对象
	if _, err := hash.Write(data); err != nil {
		return "", fmt.Errorf("计算哈希值失败: %v", err)
	}

	// 计算哈希值并转换为十六进制字符串
	hashInBytes := hash.Sum(nil)
	hashString := hex.EncodeToString(hashInBytes)

	return hashString, nil
}