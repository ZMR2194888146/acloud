<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { ListMinioFiles, ListMinioFilesByBucket, ListMinioBuckets, DownloadFileFromMinio, DeleteFileFromMinio, CreateMinioFolder, UploadDataToMinio } from '../../wailsjs/go/main/App'
import FilePreview from './FilePreview.vue'

const props = defineProps({
  enabled: Boolean
})

const files = ref([])
const currentPath = ref('')
const loading = ref(false)
const error = ref('')
const newFolderName = ref('')
const breadcrumbs = ref([{ name: '根目录', path: '' }])

// Bucket 相关
const buckets = ref([])
const selectedBucket = ref('')
const bucketsLoading = ref(false)

// 文件预览相关
const previewVisible = ref(false)
const currentPreviewFile = ref(null)

// 加载 buckets 列表
const loadBuckets = async () => {
  if (!props.enabled) return
  
  bucketsLoading.value = true
  error.value = ''
  
  try {
    const result = await ListMinioBuckets()
    buckets.value = result || []
    
    // 如果没有选择的 bucket，选择第一个
    if (!selectedBucket.value && buckets.value.length > 0) {
      selectedBucket.value = buckets.value[0]
    }
  } catch (err) {
    error.value = `加载存储桶失败: ${err.message || '未知错误'}`
    buckets.value = []
  } finally {
    bucketsLoading.value = false
  }
}

// 加载文件列表
const loadFiles = async () => {
  if (!props.enabled || !selectedBucket.value) return
  
  loading.value = true
  error.value = ''
  
  try {
    const result = await ListMinioFilesByBucket(selectedBucket.value, currentPath.value)
    files.value = result || []
  } catch (err) {
    error.value = `加载文件失败: ${err.message || '未知错误'}`
    files.value = []
  } finally {
    loading.value = false
  }
}

// 打开文件夹
const openFolder = (folder) => {
  if (folder.name === '..') {
    // 返回上级目录
    goBack()
    return
  }
  
  currentPath.value = folder.path
  
  // 更新面包屑
  const parts = folder.path.split('/')
  breadcrumbs.value = [{ name: '根目录', path: '' }]
  let path = ''
  for (let i = 0; i < parts.length; i++) {
    if (parts[i]) {
      path += (path ? '/' : '') + parts[i]
      breadcrumbs.value.push({ name: parts[i], path })
    }
  }
  
  loadFiles()
}

// 返回上一级
const goBack = () => {
  if (currentPath.value === '') return
  
  const parts = currentPath.value.split('/')
  parts.pop()
  currentPath.value = parts.join('/')
  
  // 更新面包屑
  breadcrumbs.value.pop()
  
  loadFiles()
}

// 导航到指定路径
const navigateTo = (breadcrumb) => {
  currentPath.value = breadcrumb.path
  
  // 更新面包屑
  const index = breadcrumbs.value.findIndex(b => b.path === breadcrumb.path)
  breadcrumbs.value = breadcrumbs.value.slice(0, index + 1)
  
  loadFiles()
}

// 创建文件夹
const createFolder = async () => {
  if (!newFolderName.value) return
  
  try {
    const folderPath = currentPath.value 
      ? `${currentPath.value}/${newFolderName.value}`
      : newFolderName.value
      
    await CreateMinioFolder(folderPath)
    newFolderName.value = ''
    loadFiles()
  } catch (err) {
    error.value = `创建文件夹失败: ${err.message || '未知错误'}`
  }
}

// 删除文件或文件夹
const deleteItem = async (item) => {
  if (confirm(`确定要删除 ${item.name} 吗？`)) {
    try {
      await DeleteFileFromMinio(item.path)
      loadFiles()
    } catch (err) {
      error.value = `删除失败: ${err.message || '未知错误'}`
    }
  }
}

// 下载文件
const downloadFile = async (file) => {
  if (file.isDir) return
  
  try {
    // 获取文件内容
    const fileData = await DownloadFileFromMinio(file.path)
    
    // 创建 Blob 对象
    const blob = new Blob([fileData])
    
    // 创建下载链接
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = file.name
    
    // 触发下载
    document.body.appendChild(a)
    a.click()
    
    // 清理
    setTimeout(() => {
      document.body.removeChild(a)
      URL.revokeObjectURL(url)
    }, 0)
  } catch (err) {
    error.value = `下载文件失败: ${err.message || '未知错误'}`
  }
}

// 预览文件
const previewFile = (file) => {
  if (file.isDir) return
  
  currentPreviewFile.value = file
  previewVisible.value = true
}

// 关闭预览
const closePreview = () => {
  previewVisible.value = false
  currentPreviewFile.value = null
}

// 触发文件上传对话框
const triggerFileUpload = () => {
  document.getElementById('minio-file-upload').click()
}

// 处理文件上传
const handleFileUpload = async (event) => {
  const files = event.target.files
  if (!files || files.length === 0) return
  
  for (let i = 0; i < files.length; i++) {
    const file = files[i]
    try {
      // 读取文件内容
      const fileContent = await readFileAsArrayBuffer(file)
      
      // 构建远程路径
      const remotePath = currentPath.value 
        ? `${currentPath.value}/${file.name}`
        : file.name
      
      // 上传文件
      await UploadDataToMinio(new Uint8Array(fileContent), remotePath)
    } catch (err) {
      error.value = `上传文件 ${file.name} 失败: ${err.message || '未知错误'}`
    }
  }
  
  // 重置文件输入框
  event.target.value = null
  
  // 重新加载文件列表
  loadFiles()
}

// 将文件读取为 ArrayBuffer
const readFileAsArrayBuffer = (file) => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(reader.result)
    reader.onerror = () => reject(reader.error)
    reader.readAsArrayBuffer(file)
  })
}

// 格式化文件大小
const formatSize = (size) => {
  if (size < 1024) return size + ' B'
  if (size < 1024 * 1024) return (size / 1024).toFixed(2) + ' KB'
  if (size < 1024 * 1024 * 1024) return (size / (1024 * 1024)).toFixed(2) + ' MB'
  return (size / (1024 * 1024 * 1024)).toFixed(2) + ' GB'
}

// 格式化日期
const formatDate = (date) => {
  return new Date(date).toLocaleString()
}

// 当选择的 bucket 改变时重新加载文件
const onBucketChange = () => {
  currentPath.value = ''
  breadcrumbs.value = [{ name: '根目录', path: '' }]
  loadFiles()
}

// 监听 enabled 属性变化
watch(() => props.enabled, (newValue) => {
  if (newValue) {
    loadBuckets()
  }
}, { immediate: true })

// 监听选择的 bucket 变化
watch(() => selectedBucket.value, () => {
  if (selectedBucket.value) {
    onBucketChange()
  }
})

// 处理键盘事件
const handleKeydown = (event) => {
  // F5 键刷新
  if (event.key === 'F5') {
    event.preventDefault()
    if (selectedBucket.value && !loading.value) {
      loadFiles()
    }
  }
}

// 组件挂载时加载 buckets 和文件列表
onMounted(() => {
  if (props.enabled) {
    loadBuckets()
  }
  
  // 添加键盘事件监听
  document.addEventListener('keydown', handleKeydown)
})

// 组件卸载时移除事件监听
onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
})
</script>

<template>
  <div v-if="enabled" class="minio-files">
    <h2>MinIO 文件管理</h2>
    
    <div v-if="error" class="error-message">
      {{ error }}
      <button class="close-error" @click="error = ''">×</button>
    </div>
    
    <!-- Bucket 选择器 -->
    <div class="bucket-selector">
      <label for="bucket-select">选择存储桶:</label>
      <select 
        id="bucket-select" 
        v-model="selectedBucket" 
        :disabled="bucketsLoading || buckets.length === 0"
        @change="onBucketChange"
      >
        <option value="" disabled>{{ bucketsLoading ? '加载中...' : '请选择存储桶' }}</option>
        <option v-for="bucket in buckets" :key="bucket" :value="bucket">
          {{ bucket }}
        </option>
      </select>
      <button @click="loadBuckets" :disabled="bucketsLoading" class="refresh-buckets">
        {{ bucketsLoading ? '刷新中...' : '刷新' }}
      </button>
    </div>
    
    <div class="toolbar">
      <button @click="goBack" :disabled="currentPath === ''">返回上级</button>
      <button @click="loadFiles" :disabled="loading || !selectedBucket" class="refresh-btn">
        {{ loading ? '刷新中...' : '刷新' }}
      </button>
      <div class="breadcrumbs">
        <span 
          v-for="(breadcrumb, index) in breadcrumbs" 
          :key="index" 
          @click="navigateTo(breadcrumb)"
          class="breadcrumb"
        >
          {{ breadcrumb.name }}
          <span v-if="index < breadcrumbs.length - 1" class="separator">/</span>
        </span>
      </div>
    </div>
    
    <div class="action-bar">
      <div class="new-folder">
        <input 
          type="text" 
          v-model="newFolderName" 
          placeholder="新文件夹名称" 
          @keyup.enter="createFolder"
        />
        <button @click="createFolder">创建文件夹</button>
      </div>
      
      <div class="file-upload">
        <input 
          type="file" 
          id="minio-file-upload" 
          @change="handleFileUpload" 
          multiple
          style="display: none"
        />
        <button @click="triggerFileUpload">上传文件</button>
      </div>
    </div>
    
    <div class="file-list">
      <div v-if="loading" class="loading">
        <div class="spinner"></div>
        <p>加载中...</p>
      </div>
      
      <table v-else>
        <thead>
          <tr>
            <th>名称</th>
            <th>大小</th>
            <th>修改日期</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(item, index) in files" :key="index">
            <td>
              <div class="file-name">
                <span 
                  class="icon" 
                  :class="{ 'folder-icon': item.isDir, 'file-icon': !item.isDir }"
                ></span>
                <span 
                  @click="item.isDir ? openFolder(item) : null"
                  :class="{ 'folder-name': item.isDir, 'file-name-text': !item.isDir }"
                >
                  {{ item.name }}
                </span>
              </div>
            </td>
            <td>{{ item.isDir ? '-' : formatSize(item.size) }}</td>
            <td>{{ item.lastModified ? formatDate(item.lastModified) : '-' }}</td>
            <td class="actions">
              <button v-if="!item.isDir" @click="downloadFile(item)" class="download">下载</button>
              <button v-if="!item.isDir" @click="previewFile(item)" class="preview">预览</button>
              <button v-if="item.name !== '..'" @click="deleteItem(item)" class="delete">删除</button>
            </td>
          </tr>
          <tr v-if="files.length === 0">
            <td colspan="4" class="empty-message">此文件夹为空</td>
          </tr>
        </tbody>
      </table>
    </div>
    
    <!-- 文件预览组件 -->
    <FilePreview 
      :file="currentPreviewFile" 
      :visible="previewVisible" 
      @close="closePreview" 
    />
  </div>
</template>

<style scoped>
.minio-files {
  background-color: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
}

h2 {
  margin-top: 0;
  margin-bottom: 20px;
  color: #1890ff;
  font-size: 20px;
}

.error-message {
  background-color: #fff2f0;
  color: #ff4d4f;
  padding: 12px;
  border-radius: 8px;
  margin-bottom: 20px;
  text-align: center;
  border: 1px solid #ffccc7;
  font-size: 14px;
  position: relative;
}

.close-error {
  position: absolute;
  right: 10px;
  top: 10px;
  background: none;
  border: none;
  font-size: 18px;
  cursor: pointer;
  color: #ff4d4f;
}

.bucket-selector {
  padding: 15px 20px;
  background-color: #f8f9fa;
  border-bottom: 1px solid #e9ecef;
  border-radius: 8px;
  margin-bottom: 20px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.bucket-selector label {
  font-weight: 500;
  color: #333;
  white-space: nowrap;
}

.bucket-selector select {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid #d9d9d9;
  border-radius: 6px;
  font-size: 14px;
  background-color: white;
  transition: border-color 0.3s ease;
}

.bucket-selector select:focus {
  border-color: #722ed1;
  outline: none;
  box-shadow: 0 0 0 2px rgba(114, 46, 209, 0.2);
}

.bucket-selector select:disabled {
  background-color: #f5f5f5;
  cursor: not-allowed;
}

.refresh-buckets {
  background-color: #722ed1;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s ease;
  white-space: nowrap;
}

.refresh-buckets:hover:not(:disabled) {
  background-color: #9254de;
  transform: translateY(-1px);
}

.refresh-buckets:disabled {
  background-color: #d9d9d9;
  cursor: not-allowed;
  transform: none;
}

.toolbar {
  display: flex;
  align-items: center;
  margin-bottom: 20px;
  background-color: #f9f9f9;
  padding: 12px 15px;
  border-radius: 8px;
}

.toolbar button {
  background-color: #52c41a;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 20px;
  cursor: pointer;
  margin-right: 15px;
  font-weight: 500;
  transition: all 0.3s ease;
}

.toolbar button:hover {
  background-color: #73d13d;
  transform: translateY(-2px);
}

.toolbar button:disabled {
  background-color: #d9d9d9;
  cursor: not-allowed;
  transform: none;
}

.refresh-btn {
  background-color: #1890ff !important;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 20px;
  cursor: pointer;
  margin-right: 15px;
  font-weight: 500;
  transition: all 0.3s ease;
}

.refresh-btn:hover:not(:disabled) {
  background-color: #40a9ff !important;
  transform: translateY(-2px);
}

.refresh-btn:disabled {
  background-color: #d9d9d9 !important;
  cursor: not-allowed;
  transform: none;
}

.breadcrumbs {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
}

.breadcrumb {
  cursor: pointer;
  color: #1890ff;
  margin-right: 5px;
  transition: all 0.2s ease;
}

.breadcrumb:hover {
  color: #40a9ff;
  text-decoration: underline;
}

.separator {
  margin: 0 5px;
  color: #999;
}

.action-bar {
  display: flex;
  justify-content: space-between;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.new-folder {
  display: flex;
  flex: 1;
  margin-right: 20px;
  margin-bottom: 10px;
}

.new-folder input {
  flex: 1;
  padding: 10px 15px;
  border: 1px solid #d9d9d9;
  border-radius: 4px 0 0 4px;
  font-size: 14px;
}

.new-folder button {
  background-color: #1890ff;
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: 0 4px 4px 0;
  cursor: pointer;
  font-weight: 500;
}

.file-upload button {
  background-color: #52c41a;
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 500;
}

.file-list {
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  overflow: hidden;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th, td {
  padding: 12px 15px;
  text-align: left;
  border-bottom: 1px solid #f0f0f0;
}

th {
  background-color: #fafafa;
  font-weight: 600;
  color: #333;
  font-size: 14px;
}

tr:hover {
  background-color: #f5f5f5;
}

.file-name {
  display: flex;
  align-items: center;
}

.icon {
  width: 24px;
  height: 24px;
  margin-right: 12px;
  background-size: contain;
  background-repeat: no-repeat;
}

.folder-icon {
  background-image: url('data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="%23FFC107"><path d="M10 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V8c0-1.1-.9-2-2-2h-8l-2-2z"/></svg>');
}

.file-icon {
  background-image: url('data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="%231890ff"><path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z"/></svg>');
}

.folder-name {
  cursor: pointer;
  color: #1890ff;
  font-weight: 500;
}

.folder-name:hover {
  text-decoration: underline;
}

.file-name-text {
  color: #333;
}

.actions {
  white-space: nowrap;
}

.actions button {
  background-color: #f0f0f0;
  color: #333;
  border: none;
  padding: 6px 12px;
  border-radius: 4px;
  cursor: pointer;
  margin-right: 8px;
  font-size: 0.9em;
  transition: all 0.3s;
}

.actions button:hover {
  background-color: #e0e0e0;
}

.actions button.delete {
  background-color: #ff4d4f;
  color: white;
}

.actions button.delete:hover {
  background-color: #ff7875;
}

.actions button.download {
  background-color: #1890ff;
  color: white;
}

.actions button.download:hover {
  background-color: #40a9ff;
}

.actions button.preview {
  background-color: #722ed1;
  color: white;
}

.actions button.preview:hover {
  background-color: #9254de;
}

.empty-message {
  text-align: center;
  color: #999;
  padding: 40px 0;
  font-size: 16px;
}

.loading {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  padding: 40px 0;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 4px solid rgba(0, 0, 0, 0.1);
  border-radius: 50%;
  border-top-color: #1890ff;
  animation: spin 1s linear infinite;
  margin-bottom: 15px;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>