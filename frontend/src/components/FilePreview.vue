<script setup>
import { ref, onMounted, watch } from 'vue'
import { GetFileType, ReadFile, GetFilePreview } from '../../wailsjs/go/main/App'

const props = defineProps({
  file: Object,
  visible: Boolean
})

const emit = defineEmits(['close'])

const fileContent = ref(null)
const fileType = ref('')
const loading = ref(false)
const error = ref(null)
const isBase64 = ref(false)

// 支持预览的文件类型
const previewableTypes = {
  text: ['text/plain', 'text/html', 'text/css', 'text/javascript', 'application/json', 'application/xml'],
  image: ['image/jpeg', 'image/png', 'image/gif', 'image/svg+xml', 'image/webp'],
  pdf: ['application/pdf'],
  code: [
    'text/x-go', 'text/x-python', 'text/x-java', 'text/x-c', 'text/x-c++', 
    'application/javascript', 'application/typescript'
  ]
}

// 判断文件类型是否可预览
const isPreviewable = (type) => {
  return [...previewableTypes.text, ...previewableTypes.image, ...previewableTypes.pdf, ...previewableTypes.code].some(
    t => type.startsWith(t)
  )
}

// 判断文件类型
const getFileCategory = (type) => {
  if (previewableTypes.text.some(t => type.startsWith(t))) return 'text'
  if (previewableTypes.image.some(t => type.startsWith(t))) return 'image'
  if (previewableTypes.pdf.some(t => type.startsWith(t))) return 'pdf'
  if (previewableTypes.code.some(t => type.startsWith(t))) return 'code'
  return 'unknown'
}

// 加载文件内容
const loadFileContent = async () => {
  if (!props.file || !props.visible) return
  
  loading.value = true
  error.value = null
  
  try {
    // 获取文件预览信息
    const preview = await GetFilePreview(props.file.path)
    
    // 调试输出
    console.log('文件预览信息:', preview)
    console.log('MIME类型:', preview.MimeType)
    console.log('内容类型:', typeof preview.Content)
    console.log('是否Base64编码:', preview.IsBase64)
    
    // 设置文件类型和内容
    fileType.value = preview.MimeType
    fileContent.value = preview.Content
    isBase64.value = preview.IsBase64
    
    // 如果文件类型不可预览，则显示错误信息
    if (!isPreviewable(fileType.value)) {
      error.value = `不支持预览此类型的文件: ${fileType.value}`
      loading.value = false
      return
    }
  } catch (err) {
    console.error('加载文件预览失败:', err)
    error.value = `加载文件失败: ${err.message || '未知错误'}`
  } finally {
    loading.value = false
  }
}

// 关闭预览
const closePreview = () => {
  emit('close')
}

// 当文件或可见性变化时重新加载
watch(() => [props.file, props.visible], () => {
  if (props.visible && props.file) {
    loadFileContent()
  }
}, { immediate: true })

// 获取图片的 Data URL
const getImageDataUrl = () => {
  if (!fileContent.value) return ''
  
  // 如果内容已经是Base64编码，直接使用
  if (isBase64.value) {
    return `data:${fileType.value};base64,${fileContent.value}`
  }
  
  // 否则尝试编码（这种情况不应该发生，因为图片应该总是Base64编码的）
  console.warn('图片内容不是Base64编码，这可能导致显示问题')
  try {
    return `data:${fileType.value};base64,${btoa(fileContent.value)}`
  } catch (error) {
    console.error('创建图片URL时出错:', error)
    return ''
  }
}

// 获取文本内容
const getTextContent = () => {
  if (!fileContent.value) return ''
  
  // 如果不是Base64编码，直接返回内容（文本文件的正常情况）
  if (!isBase64.value) {
    return fileContent.value
  }
  
  // 如果是Base64编码但需要显示为文本，尝试解码
  // 这种情况不应该经常发生，因为文本文件应该直接返回文本内容
  console.warn('文本内容是Base64编码，尝试解码')
  try {
    const decoded = atob(fileContent.value)
    return decoded
  } catch (error) {
    console.error('解码Base64内容时出错:', error)
    return '无法解码文件内容'
  }
}
</script>

<template>
  <div v-if="visible" class="file-preview-overlay" @click.self="closePreview">
    <div class="file-preview-container">
      <div class="file-preview-header">
        <h3>{{ file ? file.name : '文件预览' }}</h3>
        <button class="close-button" @click="closePreview">&times;</button>
      </div>
      
      <div class="file-preview-content">
        <div v-if="loading" class="loading">
          <div class="spinner"></div>
          <p>加载中...</p>
        </div>
        
        <div v-else-if="error" class="error-message">
          {{ error }}
        </div>
        
        <div v-else-if="fileContent">
          <!-- 文本文件预览 -->
          <pre v-if="getFileCategory(fileType) === 'text'" class="text-preview">{{ getTextContent() }}</pre>
          
          <!-- 代码文件预览 -->
          <pre v-else-if="getFileCategory(fileType) === 'code'" class="code-preview">{{ getTextContent() }}</pre>
          
          <!-- 图片预览 -->
          <div v-else-if="getFileCategory(fileType) === 'image'" class="image-preview">
            <img :src="getImageDataUrl()" :alt="file.name" />
          </div>
          
          <!-- PDF预览 -->
          <div v-else-if="getFileCategory(fileType) === 'pdf'" class="pdf-preview">
            <p>PDF预览暂不支持，请下载后查看。</p>
          </div>
          
          <!-- 不支持的文件类型 -->
          <div v-else class="unsupported">
            <p>不支持预览此类型的文件: {{ fileType }}</p>
            <p>文件路径: {{ file.path }}</p>
            <p>是否Base64编码: {{ isBase64 ? '是' : '否' }}</p>
          </div>
        </div>
        
        <div v-else class="no-content">
          <p>无法加载文件内容</p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.file-preview-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.7);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
  animation: fadeIn 0.3s ease-out;
}

.file-preview-container {
  width: 80%;
  height: 80%;
  background-color: white;
  border-radius: 12px;
  box-shadow: 0 5px 25px rgba(0, 0, 0, 0.2);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.file-preview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px 20px;
  background-color: #f0f2f5;
  border-bottom: 1px solid #e8e8e8;
}

.file-preview-header h3 {
  margin: 0;
  font-size: 18px;
  color: #333;
  font-weight: 600;
}

.close-button {
  background: none;
  border: none;
  font-size: 24px;
  color: #999;
  cursor: pointer;
  transition: color 0.2s ease;
}

.close-button:hover {
  color: #333;
}

.file-preview-content {
  flex: 1;
  padding: 20px;
  overflow: auto;
  display: flex;
  justify-content: center;
  align-items: flex-start;
}

.loading {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  height: 100%;
  width: 100%;
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

.error-message {
  color: #ff4d4f;
  text-align: center;
  padding: 20px;
  background-color: #fff2f0;
  border: 1px solid #ffccc7;
  border-radius: 8px;
  width: 100%;
}

.text-preview, .code-preview {
  width: 100%;
  height: 100%;
  overflow: auto;
  background-color: #f5f5f5;
  padding: 15px;
  border-radius: 8px;
  white-space: pre-wrap;
  font-family: 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.5;
  color: #333;
}

.code-preview {
  background-color: #282c34;
  color: #abb2bf;
}

.image-preview {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100%;
  height: 100%;
  overflow: auto;
}

.image-preview img {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

.pdf-preview, .unsupported, .no-content {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 100%;
  height: 100%;
  color: #666;
  text-align: center;
  background-color: #f5f5f5;
  border-radius: 8px;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

@media (max-width: 768px) {
  .file-preview-container {
    width: 95%;
    height: 90%;
  }
}
</style>