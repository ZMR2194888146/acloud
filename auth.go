package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
)

// 加载用户数据
func (a *App) loadUsers() {
	// 检查用户文件是否存在
	if _, err := os.Stat(a.usersFile); os.IsNotExist(err) {
		return
	}
	
	// 读取用户文件
	data, err := os.ReadFile(a.usersFile)
	if err != nil {
		return
	}
	
	// 解析用户数据
	var users map[string]User
	if err := json.Unmarshal(data, &users); err != nil {
		return
	}
	
	a.users = users
}

// 保存用户数据
func (a *App) saveUsers() error {
	// 序列化用户数据
	data, err := json.Marshal(a.users)
	if err != nil {
		return err
	}
	
	// 写入用户文件
	return os.WriteFile(a.usersFile, data, 0644)
}

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

// Register 注册新用户
func (a *App) Register(username, password string) AuthResponse {
	// 检查输入参数
	if username == "" || password == "" {
		return AuthResponse{Success: false, Message: "用户名和密码不能为空"}
	}
	
	if len(username) < 3 {
		return AuthResponse{Success: false, Message: "用户名至少3个字符"}
	}
	
	if len(password) < 6 {
		return AuthResponse{Success: false, Message: "密码至少6个字符"}
	}
	
	// 检查用户是否已存在
	if _, exists := a.users[username]; exists {
		return AuthResponse{Success: false, Message: "用户名已存在"}
	}
	
	// 创建新用户
	a.users[username] = User{
		Username: username,
		Password: hashPassword(password),
	}
	
	// 保存用户数据
	if err := a.saveUsers(); err != nil {
		return AuthResponse{Success: false, Message: "保存用户数据失败"}
	}
	
	return AuthResponse{Success: true, Message: "注册成功"}
}

// Login 用户登录
func (a *App) Login(username, password string) AuthResponse {
	// 检查输入参数
	if username == "" || password == "" {
		return AuthResponse{Success: false, Message: "用户名和密码不能为空"}
	}
	
	// 检查用户是否存在
	user, exists := a.users[username]
	if !exists {
		return AuthResponse{Success: false, Message: "用户不存在"}
	}
	
	// 验证密码
	if user.Password != hashPassword(password) {
		return AuthResponse{Success: false, Message: "密码错误"}
	}
	
	// 登录成功
	a.currentUser = username
	a.isLoggedIn = true
	return AuthResponse{Success: true, Message: "登录成功"}
}

// Logout 用户登出
func (a *App) Logout() {
	a.currentUser = ""
	a.isLoggedIn = false
}

// IsLoggedIn 检查是否已登录
func (a *App) IsLoggedIn() bool {
	return a.isLoggedIn
}

// GetCurrentUser 获取当前登录用户
func (a *App) GetCurrentUser() string {
	return a.currentUser
}