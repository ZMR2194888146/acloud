<template>
  <div class="sync-rule-manager">
    <h2>文件同步规则管理</h2>
    
    <div v-if="error" class="error-message">
      {{ error }}
      <button class="close-error" @click="error = ''">×</button>
    </div>
    
    <!-- 同步规则列表 -->
    <div class="rule-list">
      <a-table 
        :columns="columns" 
        :data="rules" 
        :loading="loading"
        :pagination="false"
        class="rule-table"
      >
        <template #empty>
          <div class="empty-state">
            <icon-inbox class="empty-icon" />
            <p>暂无同步规则，请添加新规则</p>
          </div>
        </template>
        
        <template #direction="{ record }">
          <a-tag :color="getDirectionColor(record.direction)">
            {{ getDirectionText(record.direction) }}
          </a-tag>
        </template>
        
        <template #enabled="{ record }">
          <a-switch 
            v-model="record.enabled" 
            @change="(value) => toggleRuleStatus(record.id, value)"
          />
        </template>
        
        <template #operation="{ record }">
          <div class="operation-btns">
            <a-button type="text" size="small" @click="editRule(record)">
              <icon-edit />
            </a-button>
            <a-button type="text" size="small" @click="confirmDelete(record)">
              <icon-delete />
            </a-button>
          </div>
        </template>
      </a-table>
    </div>
    
    <!-- 添加/编辑规则表单 -->
    <a-card class="rule-form-card" :title="isEditing ? '编辑同步规则' : '添加同步规则'">
      <a-form :model="form" layout="vertical">
        <a-form-item field="name" label="规则名称" :rules="[{ required: true, message: '请输入规则名称' }]">
          <a-input v-model="form.name" placeholder="请输入规则名称" />
        </a-form-item>
        
        <a-form-item field="localPath" label="本地同步文件夹" :rules="[{ required: true, message: '请选择本地文件夹' }]">
          <div class="path-input">
            <a-input v-model="form.localPath" placeholder="选择要同步的本地文件夹路径" readonly />
            <a-button @click="selectLocalFolder" type="outline">
              <template #icon><icon-folder /></template>
              浏览本地文件夹
            </a-button>
          </div>
          <div class="path-tips">
            选择本地计算机上要同步的文件夹
          </div>
        </a-form-item>
        
        <a-form-item field="remotePath" label="远程同步文件夹" :rules="[{ required: true, message: '请选择远程路径' }]">
          <div class="path-input">
            <a-input v-model="form.remotePath" placeholder="选择云端存储的文件夹路径" />
            <a-button @click="showRemoteFolderSelector = true" type="outline">
              <template #icon><icon-cloud /></template>
              浏览云端文件夹
            </a-button>
          </div>
          <div class="path-tips">
            选择云端存储中的目标文件夹，例如：backup/documents
          </div>
        </a-form-item>
        
        <a-form-item field="direction" label="同步方向">
          <a-radio-group v-model="form.direction">
            <a-radio value="upload">上传 (本地 → 远程)</a-radio>
            <a-radio value="download">下载 (远程 → 本地)</a-radio>
            <a-radio value="bidirectional">双向同步</a-radio>
          </a-radio-group>
        </a-form-item>
        
        <a-form-item field="filters" label="过滤规则 (可选)">
          <a-input-tag
            v-model="form.filters"
            placeholder="添加过滤规则，例如：*.tmp"
            allow-clear
          />
          <div class="filter-tips">
            支持通配符，例如：*.tmp（忽略临时文件）、.git*（忽略Git文件）
          </div>
        </a-form-item>
        
        <a-form-item field="enabled">
          <a-checkbox v-model="form.enabled">启用此规则</a-checkbox>
        </a-form-item>
        
        <div class="form-actions">
          <a-button @click="resetForm">取消</a-button>
          <a-button type="primary" @click="saveRule" :loading="saving">
            {{ isEditing ? '更新' : '添加' }}
          </a-button>
        </div>
      </a-form>
    </a-card>
    
    <!-- 手动同步按钮 -->
    <div class="manual-sync">
      <a-button type="primary" @click="triggerManualSync" :loading="syncing" :disabled="rules.length === 0">
        <template #icon><icon-sync /></template>
        立即同步
      </a-button>
      <a-button @click="refreshRules" :disabled="loading">
        <template #icon><icon-refresh /></template>
        刷新列表
      </a-button>
    </div>
    
    <!-- 同步状态 -->
    <div v-if="syncStatus" class="sync-status">
      <h3>同步状态</h3>
      <div class="status-info">
        <div class="status-item">
          <span class="label">上次同步：</span>
          <span>{{ formatDate(syncStatus.lastSync) }}</span>
        </div>
        <div class="status-item">
          <span class="label">同步模式：</span>
          <span>{{ getSyncModeText(syncStatus.syncMode) }}</span>
        </div>
        <div class="status-item">
          <span class="label">上传文件：</span>
          <span>{{ syncStatus.filesUploaded }}</span>
        </div>
        <div class="status-item">
          <span class="label">下载文件：</span>
          <span>{{ syncStatus.filesDownloaded }}</span>
        </div>
        <div class="status-item">
          <span class="label">冲突数量：</span>
          <span>{{ syncStatus.conflictCount }}</span>
        </div>
      </div>
      
      <div v-if="syncStatus.errors && syncStatus.errors.length > 0" class="sync-errors">
        <h4>同步错误</h4>
        <ul>
          <li v-for="(err, index) in syncStatus.errors" :key="index">{{ err }}</li>
        </ul>
      </div>
    </div>
    
    <!-- 远程文件夹选择器模态框 -->
    <a-modal 
      v-model:visible="showRemoteFolderSelector" 
      title="选择远程同步文件夹"
      width="600px"
      @ok="selectRemoteFolder(form.remotePath)"
      @cancel="showRemoteFolderSelector = false"
    >
      <div class="remote-folder-selector">
        <div class="selector-header">
          <a-input 
            v-model="form.remotePath" 
            placeholder="输入远程文件夹路径，例如：backup/documents"
            style="margin-bottom: 16px"
          />
          <a-space>
            <a-button @click="createRemoteFolderPath" type="outline">
              <template #icon><icon-plus /></template>
              新建路径
            </a-button>
            <a-button @click="loadRemoteBuckets" :loading="loadingRemoteFolders">
              <template #icon><icon-refresh /></template>
              刷新
            </a-button>
          </a-space>
        </div>
        
        <div class="folder-browser">
          <div class="browser-header">
            <h4>浏览云端文件夹</h4>
            <div class="current-path">
              当前路径: {{ currentRemotePath || '根目录' }}
            </div>
          </div>
          
          <div class="folder-list" v-loading="loadingRemoteFolders">
            <div v-if="remoteBuckets.length === 0" class="empty-state">
              <p>请先配置MinIO连接</p>
            </div>
            
            <div v-else>
              <div class="bucket-section">
                <h5>存储桶</h5>
                <div 
                  v-for="bucket in remoteBuckets" 
                  :key="bucket"
                  class="folder-item bucket-item"
                  @click="loadRemoteFolders(bucket)"
                >
                  <icon-archive class="folder-icon" />
                  <span>{{ bucket }}</span>
                </div>
              </div>
              
              <div v-if="remoteFolders.length > 0" class="folders-section">
                <h5>文件夹</h5>
                <div 
                  v-for="folder in remoteFolders" 
                  :key="folder.path"
                  class="folder-item"
                  @click="selectRemoteFolder(folder.path)"
                >
                  <icon-folder class="folder-icon" />
                  <span>{{ folder.name }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
        
        <div class="selector-tips">
          <p><strong>提示：</strong></p>
          <ul>
            <li>可以直接在上方输入框中输入路径</li>
            <li>点击存储桶可以浏览其中的文件夹</li>
            <li>点击"新建路径"可以创建新的文件夹路径</li>
            <li>路径格式示例：backup/documents、sync/photos</li>
          </ul>
        </div>
      </div>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { 
  GetSyncRules, 
  AddSyncRule, 
  UpdateSyncRule, 
  RemoveSyncRule, 
  EnableSyncRule, 
  DisableSyncRule,
  TriggerManualSync,
  GetSyncStatus,
  ListMinioBuckets,
  ListMinioFilesByBucket,
  GetMinioConfig
} from '../../wailsjs/go/main/App'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'

// 表格列定义
const columns = [
  { title: '规则名称', dataIndex: 'name' },
  { title: '本地路径', dataIndex: 'localPath', ellipsis: true },
  { title: '远程路径', dataIndex: 'remotePath', ellipsis: true },
  { title: '同步方向', dataIndex: 'direction', slotName: 'direction' },
  { title: '状态', dataIndex: 'enabled', slotName: 'enabled' },
  { title: '操作', slotName: 'operation', width: 100 }
]

// 状态变量
const rules = ref([])
const loading = ref(false)
const saving = ref(false)
const syncing = ref(false)
const error = ref('')
const isEditing = ref(false)
const syncStatus = ref(null)

// 远程文件夹选择器相关
const showRemoteFolderSelector = ref(false)
const remoteFolders = ref([])
const remoteBuckets = ref([])
const currentRemotePath = ref('')
const loadingRemoteFolders = ref(false)

// 表单数据
const form = reactive({
  id: '',
  name: '',
  localPath: '',
  remotePath: '',
  direction: 'bidirectional',
  filters: [],
  enabled: true
})

// 加载同步规则
const loadRules = async () => {
  loading.value = true
  error.value = ''
  
  try {
    const result = await GetSyncRules()
    rules.value = result || []
  } catch (err) {
    error.value = `加载同步规则失败: ${err.message || err}`
    rules.value = []
  } finally {
    loading.value = false
  }
}

// 加载同步状态
const loadSyncStatus = async () => {
  try {
    const status = await GetSyncStatus()
    syncStatus.value = status
  } catch (err) {
    console.error('获取同步状态失败:', err)
  }
}

// 刷新规则列表
const refreshRules = () => {
  loadRules()
  loadSyncStatus()
}

// 选择本地文件夹
const selectLocalFolder = async () => {
  try {
    // 使用HTML5文件API作为备选方案
    const input = document.createElement('input')
    input.type = 'file'
    input.webkitdirectory = true
    input.directory = true
    
    input.onchange = (event) => {
      const files = event.target.files
      if (files && files.length > 0) {
        // 获取选择的文件夹路径
        const firstFile = files[0]
        const pathParts = firstFile.webkitRelativePath.split('/')
        if (pathParts.length > 0) {
          // 构建完整的本地路径
          const folderName = pathParts[0]
          form.localPath = folderName
        }
      }
    }
    
    input.click()
  } catch (err) {
    error.value = `选择文件夹失败: ${err.message || err}`
    // 如果文件夹选择失败，允许用户手动输入路径
    console.log('文件夹选择失败，请手动输入路径')
  }
}

// 加载远程存储桶列表
const loadRemoteBuckets = async () => {
  try {
    const buckets = await ListMinioBuckets()
    remoteBuckets.value = buckets || []
  } catch (err) {
    console.error('加载存储桶失败:', err)
    remoteBuckets.value = []
  }
}

// 加载远程文件夹列表
const loadRemoteFolders = async (bucketName, path = '') => {
  loadingRemoteFolders.value = true
  try {
    const files = await ListMinioFilesByBucket(bucketName, path)
    remoteFolders.value = files.filter(file => file.isDir) || []
    currentRemotePath.value = path
  } catch (err) {
    console.error('加载远程文件夹失败:', err)
    remoteFolders.value = []
  } finally {
    loadingRemoteFolders.value = false
  }
}

// 选择远程文件夹
const selectRemoteFolder = (folderPath) => {
  form.remotePath = folderPath
  showRemoteFolderSelector.value = false
}

// 创建新的远程文件夹路径
const createRemoteFolderPath = () => {
  const newPath = prompt('请输入新文件夹路径:')
  if (newPath) {
    form.remotePath = newPath.trim()
    showRemoteFolderSelector.value = false
  }
}

// 保存规则
const saveRule = async () => {
  // 表单验证
  if (!form.name) {
    error.value = '请输入规则名称'
    return
  }
  
  if (!form.localPath) {
    error.value = '请选择本地文件夹'
    return
  }
  
  if (!form.remotePath) {
    error.value = '请输入远程路径'
    return
  }
  
  saving.value = true
  error.value = ''
  
  try {
    const ruleData = {
      id: form.id || generateID(),
      name: form.name,
      localPath: form.localPath,
      remotePath: form.remotePath,
      direction: form.direction,
      filters: form.filters,
      enabled: form.enabled
    }
    
    if (isEditing.value) {
      await UpdateSyncRule(ruleData)
    } else {
      await AddSyncRule(ruleData)
    }
    
    // 重置表单
    resetForm()
    
    // 重新加载规则列表
    await loadRules()
  } catch (err) {
    error.value = `保存同步规则失败: ${err.message || err}`
  } finally {
    saving.value = false
  }
}

// 编辑规则
const editRule = (rule) => {
  isEditing.value = true
  form.id = rule.id
  form.name = rule.name
  form.localPath = rule.localPath
  form.remotePath = rule.remotePath
  form.direction = rule.direction
  form.filters = rule.filters || []
  form.enabled = rule.enabled
}

// 确认删除
const confirmDelete = async (rule) => {
  if (confirm(`确定要删除同步规则 "${rule.name}" 吗？`)) {
    try {
      await RemoveSyncRule(rule.id)
      await loadRules()
    } catch (err) {
      error.value = `删除同步规则失败: ${err.message || err}`
    }
  }
}

// 切换规则状态
const toggleRuleStatus = async (id, enabled) => {
  try {
    if (enabled) {
      await EnableSyncRule(id)
    } else {
      await DisableSyncRule(id)
    }
    
    // 重新加载规则列表
    await loadRules()
  } catch (err) {
    error.value = `更新规则状态失败: ${err.message || err}`
    // 回滚UI状态
    await loadRules()
  }
}

// 触发手动同步
const triggerManualSync = async () => {
  syncing.value = true
  error.value = ''
  
  try {
    await TriggerManualSync()
  } catch (err) {
    error.value = `触发同步失败: ${err.message || err}`
  } finally {
    syncing.value = false
  }
}

// 重置表单
const resetForm = () => {
  isEditing.value = false
  form.id = ''
  form.name = ''
  form.localPath = ''
  form.remotePath = ''
  form.direction = 'bidirectional'
  form.filters = []
  form.enabled = true
}

// 获取同步方向文本
const getDirectionText = (direction) => {
  switch (direction) {
    case 'upload': return '上传'
    case 'download': return '下载'
    case 'bidirectional': return '双向'
    default: return '未知'
  }
}

// 获取同步方向颜色
const getDirectionColor = (direction) => {
  switch (direction) {
    case 'upload': return 'blue'
    case 'download': return 'green'
    case 'bidirectional': return 'purple'
    default: return 'gray'
  }
}

// 获取同步模式文本
const getSyncModeText = (mode) => {
  switch (mode) {
    case 'full': return '完整同步'
    case 'selective': return '选择性同步'
    case 'backup': return '备份同步'
    case 'incremental': return '增量同步'
    default: return '未知'
  }
}

// 格式化日期
const formatDate = (date) => {
  if (!date) return '从未同步'
  return new Date(date).toLocaleString()
}

// 生成唯一ID
const generateID = () => {
  return 'rule_' + Date.now() + '_' + Math.floor(Math.random() * 1000)
}

// 监听同步事件
onMounted(() => {
  // 加载规则和状态
  loadRules()
  loadSyncStatus()
  
  // 初始化时加载远程存储桶
  loadRemoteBuckets()
  
  // 监听同步开始事件
  EventsOn('sync-started', () => {
    syncing.value = true
  })
  
  // 监听同步完成事件
  EventsOn('sync-completed', (status) => {
    syncing.value = false
    syncStatus.value = status
  })
  
  // 监听同步状态变化事件
  EventsOn('sync-status-changed', () => {
    loadSyncStatus()
  })
})

// 清理事件监听
onUnmounted(() => {
  EventsOff('sync-started')
  EventsOff('sync-completed')
  EventsOff('sync-status-changed')
})
</script>

<style scoped>
.sync-rule-manager {
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

.rule-list {
  margin-bottom: 24px;
}

.rule-table {
  width: 100%;
  border-radius: 8px;
  overflow: hidden;
}

.empty-state {
  padding: 40px 0;
  text-align: center;
  color: #999;
}

.empty-icon {
  font-size: 48px;
  color: #d9d9d9;
  margin-bottom: 16px;
}

.operation-btns {
  display: flex;
  gap: 8px;
}

.rule-form-card {
  margin-bottom: 24px;
  border-radius: 8px;
}

.path-input {
  display: flex;
  gap: 8px;
}

.filter-tips {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 16px;
}

.manual-sync {
  display: flex;
  gap: 12px;
  margin-bottom: 24px;
}

.sync-status {
  background-color: #f9f9f9;
  border-radius: 8px;
  padding: 16px;
  margin-top: 24px;
}

.sync-status h3 {
  margin-top: 0;
  margin-bottom: 16px;
  font-size: 16px;
  color: #333;
}

.status-info {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 12px;
}

.status-item {
  display: flex;
  align-items: center;
}

.status-item .label {
  font-weight: 500;
  margin-right: 8px;
  color: #666;
}

.sync-errors {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #eee;
}

.sync-errors h4 {
  margin-top: 0;
  margin-bottom: 12px;
  font-size: 14px;
  color: #ff4d4f;
}

.sync-errors ul {
  margin: 0;
  padding-left: 20px;
  color: #ff4d4f;
  font-size: 13px;
}

/* 路径输入样式 */
.path-tips {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}

/* 远程文件夹选择器样式 */
.remote-folder-selector {
  max-height: 500px;
}

.selector-header {
  margin-bottom: 16px;
}

.folder-browser {
  border: 1px solid #e8e8e8;
  border-radius: 6px;
  padding: 16px;
  margin-bottom: 16px;
  max-height: 300px;
  overflow-y: auto;
}

.browser-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid #f0f0f0;
}

.browser-header h4 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
}

.current-path {
  font-size: 12px;
  color: #666;
}

.bucket-section, .folders-section {
  margin-bottom: 16px;
}

.bucket-section h5, .folders-section h5 {
  margin: 0 0 8px 0;
  font-size: 13px;
  font-weight: 600;
  color: #333;
}

.folder-item {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.2s;
  margin-bottom: 4px;
}

.folder-item:hover {
  background-color: #f5f5f5;
}

.folder-item.bucket-item {
  background-color: #e6f7ff;
  border: 1px solid #91d5ff;
}

.folder-item.bucket-item:hover {
  background-color: #bae7ff;
}

.folder-icon {
  margin-right: 8px;
  color: #1890ff;
}

.selector-tips {
  background-color: #f6f8fa;
  padding: 12px;
  border-radius: 4px;
  font-size: 12px;
}

.selector-tips p {
  margin: 0 0 8px 0;
  font-weight: 600;
}

.selector-tips ul {
  margin: 0;
  padding-left: 16px;
}

.selector-tips li {
  margin-bottom: 4px;
  color: #666;
}
</style>
