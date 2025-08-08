package main

import (
	"crypto/sha256"
	"encoding/hex"
)

// 哈希密码
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// AuthResponse 认证响应结构
type AuthResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// IsLoggedIn 检查是否已登录
func (a *App) IsLoggedIn() bool {
	return a.isLoggedIn
}

// GetCurrentUser 获取当前登录用户
func (a *App) GetCurrentUser() string {
	return a.currentUser
}
