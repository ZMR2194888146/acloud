<script setup>
import { ref, onMounted } from 'vue'
import { GetMinioConfig, UpdateMinioConfig, TestMinioConnection } from '../../wailsjs/go/main/App'

const emit = defineEmits(['config-updated'])

const minioConfig = ref({
  endpoint: '',
  accessKeyId: '',
  secretAccessKey: '',
  useSSL: true,
  bucketName: '',
  enabled: false
})

const loading = ref(false)
const testLoading = ref(false)
const message = ref('')
const messageType = ref('') // 'success' 或 'error'

// 加载 MinIO 配置
const loadConfig = async () => {
  try {
    const config = await GetMinioConfig()
    minioConfig.value = config
  } catch (error) {
    showMessage(`加载配置失败: ${error.message || '未知错误'}`, 'error')
  }
}

// 保存 MinIO 配置
const saveConfig = async () => {
  loading.value = true
  message.value = ''
  
  try {
    await UpdateMinioConfig(
      minioConfig.value.endpoint,
      minioConfig.value.accessKeyId,
      minioConfig.value.secretAccessKey,
      minioConfig.value.bucketName,
      minioConfig.value.useSSL,
      minioConfig.value.enabled
    )
    
    showMessage('配置已保存', 'success')
    emit('config-updated', minioConfig.value.enabled)
  } catch (error) {
    showMessage(`保存配置失败: ${error.message || '未知错误'}`, 'error')
  } finally {
    loading.value = false
  }
}

// 测试连接
const testConnection = async () => {
  testLoading.value = true
  message.value = ''
  
  try {
    await TestMinioConnection(
      minioConfig.value.endpoint,
      minioConfig.value.accessKeyId,
      minioConfig.value.secretAccessKey,
      minioConfig.value.useSSL
    )
    
    showMessage('连接测试成功', 'success')
  } catch (error) {
    showMessage(`连接测试失败: ${error.message || '未知错误'}`, 'error')
  } finally {
    testLoading.value = false
  }
}

// 显示消息
const showMessage = (msg, type) => {
  message.value = msg
  messageType.value = type
  
  // 3秒后自动清除成功消息
  if (type === 'success') {
    setTimeout(() => {
      if (message.value === msg) {
        message.value = ''
      }
    }, 3000)
  }
}

// 组件挂载时加载配置
onMounted(() => {
  loadConfig()
})
</script>

<template>
  <div class="minio-config">
    <h2>MinIO 配置</h2>
    
    <div v-if="message" class="message" :class="messageType">
      {{ message }}
    </div>
    
    <div class="form-group">
      <label>
        <input type="checkbox" v-model="minioConfig.enabled" />
        启用 MinIO 存储
      </label>
    </div>
    
    <div class="form-group">
      <label for="endpoint">服务器地址</label>
      <input 
        type="text" 
        id="endpoint" 
        v-model="minioConfig.endpoint" 
        placeholder="例如: play.min.io"
      />
    </div>
    
    <div class="form-group">
      <label for="accessKeyId">访问密钥 ID</label>
      <input 
        type="text" 
        id="accessKeyId" 
        v-model="minioConfig.accessKeyId" 
        placeholder="访问密钥 ID"
      />
    </div>
    
    <div class="form-group">
      <label for="secretAccessKey">秘密访问密钥</label>
      <input 
        type="password" 
        id="secretAccessKey" 
        v-model="minioConfig.secretAccessKey" 
        placeholder="秘密访问密钥"
      />
    </div>
    
    <div class="form-group">
      <label for="bucketName">存储桶名称</label>
      <input 
        type="text" 
        id="bucketName" 
        v-model="minioConfig.bucketName" 
        placeholder="例如: acloud-storage"
      />
    </div>
    
    <div class="form-group">
      <label>
        <input type="checkbox" v-model="minioConfig.useSSL" />
        使用 SSL 连接
      </label>
    </div>
    
    <div class="form-actions">
      <button 
        @click="testConnection" 
        :disabled="testLoading" 
        class="test-button"
      >
        {{ testLoading ? '测试中...' : '测试连接' }}
      </button>
      
      <button 
        @click="saveConfig" 
        :disabled="loading" 
        class="save-button"
      >
        {{ loading ? '保存中...' : '保存配置' }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.minio-config {
  background-color: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  margin-bottom: 20px;
}

h2 {
  margin-top: 0;
  margin-bottom: 20px;
  color: #1890ff;
  font-size: 20px;
}

.form-group {
  margin-bottom: 15px;
}

label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
}

input[type="text"],
input[type="password"] {
  width: 100%;
  padding: 10px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  font-size: 14px;
}

input[type="checkbox"] {
  margin-right: 8px;
}

.form-actions {
  display: flex;
  justify-content: space-between;
  margin-top: 20px;
}

button {
  padding: 10px 20px;
  border: none;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.3s;
}

.test-button {
  background-color: #52c41a;
  color: white;
}

.test-button:hover {
  background-color: #73d13d;
}

.save-button {
  background-color: #1890ff;
  color: white;
}

.save-button:hover {
  background-color: #40a9ff;
}

button:disabled {
  background-color: #d9d9d9;
  cursor: not-allowed;
}

.message {
  padding: 10px;
  border-radius: 4px;
  margin-bottom: 15px;
}

.success {
  background-color: #f6ffed;
  border: 1px solid #b7eb8f;
  color: #52c41a;
}

.error {
  background-color: #fff2f0;
  border: 1px solid #ffccc7;
  color: #ff4d4f;
}
</style>