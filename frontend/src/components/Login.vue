<template>
  <div class="login-container">
    <!-- 背景装饰元素 -->
    <div class="background-decoration">
      <div class="floating-shape shape-1"></div>
      <div class="floating-shape shape-2"></div>
      <div class="floating-shape shape-3"></div>
      <div class="floating-shape shape-4"></div>
      <div class="floating-shape shape-5"></div>
    </div>

    <!-- 主登录卡片 -->
    <div class="login-wrapper">
      <div class="login-header">
        <div class="logo-container">
          <div class="logo-icon">
            <icon-cloud class="logo" />
          </div>
          <h1 class="app-title">HKCE Cloud</h1>
          <p class="app-subtitle">现代化桌面网盘应用</p>
        </div>
      </div>

      <a-card class="login-card" :bordered="false">
        <!-- 状态提示条 -->
        <div v-if="statusMessage" :class="['status-bar', statusType]">
          <icon-check-circle v-if="statusType === 'success'" />
          <icon-exclamation-circle v-else-if="statusType === 'error'" />
          <icon-info-circle v-else />
          {{ statusMessage }}
        </div>

        <div class="card-content">
          <!-- 登录表单 -->
          <div class="form-section" :class="{ 'form-hidden': showRegister }">
            <h2 class="form-title">
              <icon-user class="title-icon" />
              用户登录
            </h2>
            
            <!-- 简化登录表单，避免使用Arco Design的表单验证 -->
            <div class="simple-form">
              <div class="form-item">
                <a-input
                  v-model="form.username"
                  placeholder="请输入用户名"
                  size="large"
                  class="animated-input"
                  @keyup.enter="handleLogin"
                >
                  <template #prefix>
                    <icon-user class="input-icon" />
                  </template>
                </a-input>
              </div>

              <div class="form-item">
                <a-input-password
                  v-model="form.password"
                  placeholder="请输入密码"
                  size="large"
                  class="animated-input"
                  @keyup.enter="handleLogin"
                >
                  <template #prefix>
                    <icon-lock class="input-icon" />
                  </template>
                </a-input-password>
              </div>

              <div class="form-item">
                <a-button
                  type="primary"
                  size="large"
                  :loading="loading"
                  class="submit-btn login-btn"
                  long
                  @click="handleLogin"
                >
                  <span v-if="!loading">登录</span>
                  <span v-else>登录中...</span>
                </a-button>
              </div>
            </div>

            <div class="form-footer">
              <a-button 
                type="text" 
                @click="toggleForm"
                class="toggle-btn"
              >
                还没有账号？立即注册
              </a-button>
            </div>
          </div>

          <!-- 注册表单 -->
          <div class="form-section" :class="{ 'form-hidden': !showRegister }">
            <h2 class="form-title">
              <icon-user-add class="title-icon" />
              用户注册
            </h2>
            
            <!-- 简化注册表单 -->
            <div class="simple-form">
              <div class="form-item">
                <a-input
                  v-model="registerForm.username"
                  placeholder="请输入新用户名"
                  size="large"
                  class="animated-input"
                  @keyup.enter="handleRegister"
                >
                  <template #prefix>
                    <icon-user-add class="input-icon" />
                  </template>
                </a-input>
              </div>

              <div class="form-item">
                <a-input-password
                  v-model="registerForm.password"
                  placeholder="请输入新密码"
                  size="large"
                  class="animated-input"
                  @keyup.enter="handleRegister"
                >
                  <template #prefix>
                    <icon-lock class="input-icon" />
                  </template>
                </a-input-password>
              </div>

              <div class="form-item">
                <a-button
                  type="primary"
                  size="large"
                  :loading="registerLoading"
                  class="submit-btn register-btn"
                  long
                  @click="handleRegister"
                >
                  <span v-if="!registerLoading">注册</span>
                  <span v-else>注册中...</span>
                </a-button>
              </div>
            </div>

            <div class="form-footer">
              <a-button 
                type="text" 
                @click="toggleForm"
                class="toggle-btn"
              >
                已有账号？立即登录
              </a-button>
            </div>
          </div>
        </div>
      </a-card>

      <!-- 底部信息 -->
      <div class="login-footer">
        <p class="copyright">© 2024 HKCE Cloud. All rights reserved.</p>
        <div class="features">
          <span class="feature-item">
            <icon-safe class="feature-icon" />
            安全可靠
          </span>
          <span class="feature-item">
            <icon-cloud class="feature-icon" />
            云端同步
          </span>
          <span class="feature-item">
            <icon-desktop class="feature-icon" />
            跨平台
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { Login, Register } from '../../wailsjs/go/main/App'

const emit = defineEmits(['login-success'])

const loading = ref(false)
const registerLoading = ref(false)
const showRegister = ref(false)
const statusMessage = ref('')
const statusType = ref('info') // 'info', 'success', 'error'

const form = reactive({
  username: '',
  password: ''
})

const registerForm = reactive({
  username: '',
  password: ''
})

const rules = {
  username: [
    { required: true, message: '请输入用户名' }
  ],
  password: [
    { required: true, message: '请输入密码' }
  ]
}

const registerRules = {
  username: [
    { required: true, message: '请输入用户名' },
    { min: 3, message: '用户名至少3个字符' }
  ],
  password: [
    { required: true, message: '请输入密码' },
    { min: 6, message: '密码至少6个字符' }
  ]
}

const toggleForm = () => {
  showRegister.value = !showRegister.value
  // 切换表单时清除状态消息
  statusMessage.value = ''
}

// 设置状态消息
const setStatus = (message, type = 'info') => {
  statusMessage.value = message
  statusType.value = type
  
  // 5秒后自动清除成功消息
  if (type === 'success') {
    setTimeout(() => {
      if (statusMessage.value === message) {
        statusMessage.value = ''
      }
    }, 5000)
  }
}

// 使用与紧急测试登录相同的逻辑，确保一致性
const handleLogin = async (e) => {
  // 阻止表单默认提交行为
  if (e) e.preventDefault()
  
  // 表单验证
  if (!form.username) {
    setStatus('请输入用户名', 'error')
    return
  }
  
  if (!form.password) {
    setStatus('请输入密码', 'error')
    return
  }
  
  // 设置登录中状态
  loading.value = true
  setStatus('正在登录...', 'info')
  
  try {
    // 使用与测试函数相同的调用方式
    const result = await window.go.main.App.Login(form.username, form.password)
    
    if (result && result.success) {
      setStatus('登录成功！正在进入系统...', 'success')
      
      // 延迟一下再跳转，让用户看到成功消息
      setTimeout(() => {
        emit('login-success')
      }, 1000)
    } else {
      setStatus(result?.message || '登录失败：用户名或密码错误', 'error')
    }
  } catch (error) {
    setStatus(`登录错误: ${error.toString()}`, 'error')
  } finally {
    loading.value = false
  }
}

const handleRegister = async () => {
  // 表单验证
  if (!registerForm.username) {
    setStatus('请输入用户名', 'error')
    return
  }
  
  if (registerForm.username.length < 3) {
    setStatus('用户名至少3个字符', 'error')
    return
  }
  
  if (!registerForm.password) {
    setStatus('请输入密码', 'error')
    return
  }
  
  if (registerForm.password.length < 6) {
    setStatus('密码至少6个字符', 'error')
    return
  }
  
  // 设置注册中状态
  registerLoading.value = true
  setStatus('正在注册...', 'info')
  
  try {
    console.log('尝试注册:', registerForm.username, registerForm.password)
    const result = await Register(registerForm.username, registerForm.password)
    console.log('注册结果:', result)
    
    if (result && result.success) {
      setStatus('注册成功！', 'success')
      
      // 注册成功后切换到登录表单
      form.username = registerForm.username
      form.password = registerForm.password
      showRegister.value = false
      
      // 清空注册表单
      registerForm.username = ''
      registerForm.password = ''
      
      // 延迟一下再自动登录
      setTimeout(() => {
        handleLogin()
      }, 1500)
    } else {
      console.error('注册失败:', result?.message || '未知错误')
      setStatus(result?.message || '注册失败：用户名可能已存在', 'error')
    }
  } catch (error) {
    console.error('注册错误:', error)
    setStatus('注册失败：' + (error.message || '网络错误'), 'error')
  } finally {
    registerLoading.value = false
  }
}

// 直接测试登录函数
const testDirectLogin = async () => {
  setStatus('正在直接测试登录...', 'info')
  
  try {
    // 直接调用全局函数
    const result = await window.go.main.App.Login('admin', 'admin')
    setStatus(`直接登录结果: ${JSON.stringify(result)}`, 'info')
    
    if (result && result.success) {
      setStatus('登录成功！正在进入系统...', 'success')
      
      // 延迟一下再跳转，让用户看到成功消息
      setTimeout(() => {
        emit('login-success')
      }, 1000)
    } else {
      setStatus(`登录失败: ${result?.message || '未知错误'}`, 'error')
    }
  } catch (error) {
    setStatus(`直接登录错误: ${error.toString()}`, 'error')
  }
}

onMounted(() => {
  // 添加页面加载动画
  const loginWrapper = document.querySelector('.login-wrapper')
  if (loginWrapper) {
    loginWrapper.classList.add('fade-in')
  }
  
  // 设置默认用户名提示
  setStatus('请使用 admin/admin 登录系统', 'info')
})
</script>

<style scoped>
/* 主容器 */
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  position: relative;
  overflow: hidden;
}

/* 状态提示条 */
.status-bar {
  padding: 12px 16px;
  margin: -20px -20px 20px -20px;
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  border-radius: 0;
  animation: fadeIn 0.3s ease-out;
  transition: all 0.3s ease;
}

.status-bar.info {
  background-color: #e6f7ff;
  color: #1890ff;
  border-bottom: 1px solid #91d5ff;
}

.status-bar.success {
  background-color: #f6ffed;
  color: #52c41a;
  border-bottom: 1px solid #b7eb8f;
}

.status-bar.error {
  background-color: #fff2f0;
  color: #ff4d4f;
  border-bottom: 1px solid #ffccc7;
}

/* 背景装饰 */
.background-decoration {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
  z-index: 1;
}

.floating-shape {
  position: absolute;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  animation: float 6s ease-in-out infinite;
}

.shape-1 {
  width: 80px;
  height: 80px;
  top: 10%;
  left: 10%;
  animation-delay: 0s;
}

.shape-2 {
  width: 120px;
  height: 120px;
  top: 20%;
  right: 15%;
  animation-delay: 1s;
}

.shape-3 {
  width: 60px;
  height: 60px;
  bottom: 30%;
  left: 20%;
  animation-delay: 2s;
}

.shape-4 {
  width: 100px;
  height: 100px;
  bottom: 20%;
  right: 10%;
  animation-delay: 3s;
}

.shape-5 {
  width: 40px;
  height: 40px;
  top: 50%;
  left: 5%;
  animation-delay: 4s;
}

@keyframes float {
  0%, 100% {
    transform: translateY(0px) rotate(0deg);
    opacity: 0.7;
  }
  50% {
    transform: translateY(-20px) rotate(180deg);
    opacity: 1;
  }
}

/* 登录包装器 */
.login-wrapper {
  position: relative;
  z-index: 2;
  width: 100%;
  max-width: 420px;
  opacity: 0;
  transform: translateY(30px);
  transition: all 0.8s cubic-bezier(0.4, 0, 0.2, 1);
}

.login-wrapper.fade-in {
  opacity: 1;
  transform: translateY(0);
}

/* 登录头部 */
.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.logo-container {
  animation: slideDown 0.8s ease-out 0.2s both;
}

.logo-icon {
  width: 80px;
  height: 80px;
  margin: 0 auto 16px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.3);
  transition: all 0.3s ease;
}

.logo-icon:hover {
  transform: scale(1.05);
  background: rgba(255, 255, 255, 0.3);
}

.logo {
  font-size: 36px;
  color: white;
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.1);
  }
}

.app-title {
  font-size: 32px;
  font-weight: 700;
  color: white;
  margin: 0 0 8px 0;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
  letter-spacing: 1px;
}

.app-subtitle {
  font-size: 16px;
  color: rgba(255, 255, 255, 0.8);
  margin: 0;
  font-weight: 300;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 登录卡片 */
.login-card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(20px);
  border-radius: 24px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
  overflow: hidden;
  animation: slideUp 0.8s ease-out 0.4s both;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.card-content {
  padding: 40px;
  position: relative;
}

/* 表单部分 */
.form-section {
  transition: all 0.5s cubic-bezier(0.4, 0, 0.2, 1);
  transform: translateX(0);
  opacity: 1;
}

.form-section.form-hidden {
  transform: translateX(-100%);
  opacity: 0;
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
}

.form-title {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  font-size: 24px;
  font-weight: 600;
  color: #2c3e50;
  margin: 0 0 32px 0;
  text-align: center;
}

.title-icon {
  font-size: 28px;
  color: #667eea;
}

/* 表单项 */
.form-item {
  margin-bottom: 24px;
  animation: fadeInUp 0.6s ease-out both;
}

.form-item:nth-child(1) { animation-delay: 0.1s; }
.form-item:nth-child(2) { animation-delay: 0.2s; }
.form-item:nth-child(3) { animation-delay: 0.3s; }

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 输入框样式 */
.animated-input {
  border-radius: 12px;
  border: 2px solid #f0f0f0;
  transition: all 0.3s ease;
  height: 52px;
}

.animated-input:hover {
  border-color: #667eea;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.15);
}

.animated-input:focus-within {
  border-color: #667eea;
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.25);
}

.input-icon {
  color: #6c757d;
  font-size: 18px;
  transition: color 0.3s ease;
}

.animated-input:focus-within .input-icon {
  color: #667eea;
}

/* 按钮样式 */
.submit-btn {
  height: 52px;
  border-radius: 12px;
  font-size: 16px;
  font-weight: 600;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.submit-btn::before {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.2), transparent);
  transition: left 0.5s ease;
}

.submit-btn:hover::before {
  left: 100%;
}

.login-btn {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  box-shadow: 0 4px 15px rgba(102, 126, 234, 0.4);
}

.login-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(102, 126, 234, 0.5);
}

.register-btn {
  background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);
  border: none;
  box-shadow: 0 4px 15px rgba(17, 153, 142, 0.4);
}

.register-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(17, 153, 142, 0.5);
}

/* 表单底部 */
.form-footer {
  text-align: center;
  margin-top: 24px;
}

.toggle-btn {
  color: #495057;
  font-weight: 500;
  transition: all 0.3s ease;
}

.toggle-btn:hover {
  color: #343a40;
  transform: scale(1.05);
}

/* 登录底部 */
.login-footer {
  text-align: center;
  margin-top: 32px;
  animation: fadeIn 1s ease-out 0.8s both;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.copyright {
  color: rgba(255, 255, 255, 0.7);
  font-size: 14px;
  margin: 0 0 16px 0;
}

.features {
  display: flex;
  justify-content: center;
  gap: 24px;
  flex-wrap: wrap;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 6px;
  color: rgba(255, 255, 255, 0.8);
  font-size: 14px;
  font-weight: 500;
  transition: all 0.3s ease;
}

.feature-item:hover {
  color: white;
  transform: translateY(-2px);
}

.feature-icon {
  font-size: 16px;
}

/* 响应式设计 */
@media (max-width: 480px) {
  .login-wrapper {
    max-width: 90%;
    margin: 20px;
  }
  
  .card-content {
    padding: 24px;
  }
  
  .app-title {
    font-size: 28px;
  }
  
  .form-title {
    font-size: 20px;
  }
  
  .features {
    flex-direction: column;
    gap: 12px;
  }
}

/* 加载状态动画 */
.submit-btn:loading {
  pointer-events: none;
}

/* 深色模式适配 */
@media (prefers-color-scheme: dark) {
  .login-card {
    background: rgba(30, 30, 30, 0.95);
  }
  
  .form-title {
    color: #e5e5e5;
  }
  
  .animated-input {
    background: rgba(255, 255, 255, 0.05);
    border-color: rgba(255, 255, 255, 0.1);
    color: #e5e5e5;
  }
  
  .input-icon {
    color: #adb5bd;
  }
  
  .toggle-btn {
    color: #adb5bd;
  }
  
  .toggle-btn:hover {
    color: #e9ecef;
  }
  
  .status-bar.info {
    background-color: rgba(24, 144, 255, 0.1);
    color: #1890ff;
    border-bottom: 1px solid rgba(24, 144, 255, 0.3);
  }
  
  .status-bar.success {
    background-color: rgba(82, 196, 26, 0.1);
    color: #52c41a;
    border-bottom: 1px solid rgba(82, 196, 26, 0.3);
  }
  
  .status-bar.error {
    background-color: rgba(255, 77, 79, 0.1);
    color: #ff4d4f;
    border-bottom: 1px solid rgba(255, 77, 79, 0.3);
  }
}
</style>