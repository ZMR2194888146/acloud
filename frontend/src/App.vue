<script setup>
import { ref, onMounted } from 'vue'
import { ListFiles, CreateFolder, DeleteFile, RenameFile, GetStoragePath, UploadFile, UploadFileString, DownloadFile, IsLoggedIn, GetCurrentUser, Logout, OpenInExplorer, OpenFileInExplorer, GetSyncStatus, ToggleSyncStatus, GetSystemInfo, SetAutoStart, CheckForUpdatesManually, GetSyncRules, AddSyncRule, UpdateSyncRule, RemoveSyncRule, EnableSyncRule, DisableSyncRule } from '../wailsjs/go/main/App'
import Login from './components/Login.vue'
import FilePreview from './components/FilePreview.vue'
import MinioFiles from './components/MinioFiles.vue'
import SystemSettings from './components/SystemSettings.vue'
import SyncRuleManager from './components/SyncRuleManager.vue'

// 登录状态
const isLoggedIn = ref(false)
const currentUser = ref('')

// UI状态
const sidebarCollapsed = ref(false)
const activeView = ref('files')
const viewMode = ref('list')
const windowWidth = ref(window.innerWidth)
const windowHeight = ref(window.innerHeight)

// 当前路径
const currentPath = ref('')
// 文件列表
const files = ref([])
// 新文件夹名称
const newFolderName = ref('')
// 存储路径
const storagePath = ref('')
// 重命名相关
const isRenaming = ref(false)
const renameItem = ref(null)
const newName = ref('')
// 面包屑导航
const breadcrumbs = ref([{ name: '根目录', path: '' }])
// 文件预览相关
const previewVisible = ref(false)
const currentPreviewFile = ref(null)

// MinIO 相关
const minioEnabled = ref(false)
const showMinioConfig = ref(false)


// 同步相关
const syncRunning = ref(false)
const syncInterval = ref(5)
const showSyncConfig = ref(false)
const syncStatus = ref({
  running: false,
  lastSync: null,
  filesUploaded: 0,
  filesDownloaded: 0,
  errors: [],
  syncMode: 'full',
  conflictCount: 0
})

// 群晖Drive风格功能
const showVersionHistory = ref(false)
const showSyncRules = ref(false)
const showConflicts = ref(false)
const showShareLinks = ref(false)
const selectedFile = ref(null)
const fileVersions = ref([])
const syncRules = ref([])
const conflicts = ref([])
const shareLinks = ref([])
const syncMode = ref('full')
const activeTab = ref('control')
const newShareLink = ref({
  password: '',
  expiryHours: 24,
  maxDownloads: 0
})
const newSyncRule = ref({
  name: '',
  localPath: '',
  remotePath: '',
  direction: 'bidirectional',
  filters: []
})

// 菜单点击处理
const handleMenuClick = ({ key }) => {
  activeView.value = key
  if (key === 'files') {
    loadFiles()
  }
}

// 加载文件列表
const loadFiles = async () => {
  try {
    const result = await ListFiles(currentPath.value)
    files.value = result || []
  } catch (error) {
    console.error('加载文件失败:', error)
    files.value = []
  }
}

// 获取存储路径
const loadStoragePath = async () => {
  try {
    storagePath.value = await GetStoragePath()
  } catch (error) {
    console.error('获取存储路径失败:', error)
  }
}

// 打开文件夹
const openFolder = (folder) => {
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
    await CreateFolder(currentPath.value, newFolderName.value)
    newFolderName.value = ''
    loadFiles()
  } catch (error) {
    console.error('创建文件夹失败:', error)
  }
}

// 删除文件或文件夹
const deleteItem = async (item) => {
  try {
    await DeleteFile(item.path)
    loadFiles()
  } catch (error) {
    console.error('删除失败:', error)
  }
}

// 开始重命名
const startRename = (item) => {
  isRenaming.value = true
  renameItem.value = item
  newName.value = item.name
}

// 完成重命名
const finishRename = async () => {
  if (!newName.value || newName.value === renameItem.value.name) {
    cancelRename()
    return
  }
  
  try {
    await RenameFile(renameItem.value.path, newName.value)
    loadFiles()
  } catch (error) {
    console.error('重命名失败:', error)
  } finally {
    cancelRename()
  }
}

// 取消重命名
const cancelRename = () => {
  isRenaming.value = false
  renameItem.value = null
  newName.value = ''
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

// 触发文件上传对话框
const triggerFileUpload = () => {
  document.getElementById('file-upload').click()
}

// 处理文件上传
const handleFileUpload = async (event) => {
  const files = event.target.files
  if (!files || files.length === 0) return
  
  for (let i = 0; i < files.length; i++) {
    const file = files[i]
    try {
      // 读取文件内容
      const fileContent = await readFileAsText(file)
      // 上传文件
      await UploadFileString(currentPath.value, fileContent, file.name)
    } catch (error) {
      console.error(`上传文件 ${file.name} 失败:`, error)
      
      // 如果文本读取失败（可能是二进制文件），尝试使用 ArrayBuffer
      try {
        const fileData = await readFileAsArrayBuffer(file)
        await UploadFile(currentPath.value, new Uint8Array(fileData), file.name)
      } catch (err) {
        console.error(`二次尝试上传文件 ${file.name} 失败:`, err)
      }
    }
  }
  
  // 重置文件输入框
  event.target.value = null
  // 重新加载文件列表
  loadFiles()
}

// 将文件读取为文本
const readFileAsText = (file) => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(reader.result)
    reader.onerror = () => reject(reader.error)
    reader.readAsText(file)
  })
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

// 下载文件
const downloadFile = async (file) => {
  if (file.isDir) return
  
  try {
    // 获取文件内容
    const fileData = await DownloadFile(file.path)
    
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
  } catch (error) {
    console.error(`下载文件 ${file.name} 失败:`, error)
  }
}

// 预览文件
const previewFile = (file) => {
  if (file.isDir) return
  
  previewVisible.value = true
  currentPreviewFile.value = file
}

// 关闭预览
const closePreview = () => {
  previewVisible.value = false
  currentPreviewFile.value = null
}

// 检查登录状态
const checkLoginStatus = async () => {
  try {
    isLoggedIn.value = await IsLoggedIn()
    if (isLoggedIn.value) {
      currentUser.value = await GetCurrentUser()
      loadFiles()
      loadStoragePath()
    }
  } catch (error) {
    console.error('检查登录状态失败:', error)
  }
}

// 处理登录成功
const handleLoginSuccess = async () => {
  isLoggedIn.value = true
  currentUser.value = await GetCurrentUser()
  loadFiles()
  loadStoragePath()
}

// 处理登出
const handleLogout = async () => {
  try {
    await Logout()
    isLoggedIn.value = false
    currentUser.value = ''
    files.value = []
  } catch (error) {
    console.error('登出失败:', error)
  }
}


// 在资源管理器中打开当前目录
const openCurrentFolderInExplorer = async () => {
  try {
    await OpenInExplorer(currentPath.value)
  } catch (error) {
    console.error('打开资源管理器失败:', error)
  }
}

// 在资源管理器中打开文件所在目录并选中文件
const openFileInExplorer = async (file) => {
  try {
    await OpenFileInExplorer(file.path)
  } catch (error) {
    console.error('在资源管理器中打开文件失败:', error)
  }
}

// 启动同步
const startSync = async () => {
  try {
    const newStatus = await ToggleSyncStatus()
    syncRunning.value = newStatus
    updateSyncStatus()
  } catch (error) {
    console.error('启动同步失败:', error)
  }
}

// 停止同步
const stopSync = async () => {
  try {
    const newStatus = await ToggleSyncStatus()
    syncRunning.value = newStatus
    updateSyncStatus()
  } catch (error) {
    console.error('停止同步失败:', error)
  }
}

// 更新同步状态
const updateSyncStatus = async () => {
  try {
    const status = await GetSyncStatus()
    syncStatus.value = status
    syncRunning.value = status.running
  } catch (error) {
    console.error('获取同步状态失败:', error)
  }
}

// 设置同步间隔
const setSyncInterval = async () => {
  try {
    // 这里可以添加设置同步间隔的逻辑
    console.log('设置同步间隔:', syncInterval.value)
  } catch (error) {
    console.error('设置同步间隔失败:', error)
  }
}

// 格式化同步时间
const formatSyncTime = (time) => {
  if (!time) return '从未同步'
  return new Date(time).toLocaleString()
}

// 窗口大小变化处理
const handleResize = () => {
  windowWidth.value = window.innerWidth
  windowHeight.value = window.innerHeight
  
  // 小屏幕自动折叠侧边栏
  if (windowWidth.value < 1024) {
    sidebarCollapsed.value = true
  } else if (windowWidth.value > 1200) {
    sidebarCollapsed.value = false
  }
}

// 组件挂载时检查登录状态
onMounted(() => {
  checkLoginStatus()
  
  // 初始化窗口大小
  handleResize()
  
  // 监听窗口大小变化
  window.addEventListener('resize', handleResize)
  
  // 定期更新同步状态
  setInterval(() => {
    if (minioEnabled.value && isLoggedIn.value) {
      updateSyncStatus()
    }
  }, 10000) // 每10秒更新一次状态
})
</script>

<template>
  <Login v-if="!isLoggedIn" @login-success="handleLoginSuccess" />
  
  <div v-else class="modern-drive-container">
    <!-- 现代化布局 -->
    <a-layout class="modern-layout">
      <!-- 侧边栏 -->
      <a-layout-sider 
        class="modern-sidebar" 
        :width="200" 
        :collapsed="sidebarCollapsed"
        collapsible
        @collapse="sidebarCollapsed = $event"
      >
        <!-- Logo区域 -->
        <div class="sidebar-logo">
          <div class="logo-wrapper">
            <div class="logo-icon-wrapper">
              <span class="logo-icon">☁️</span>
            </div>
            <span v-if="!sidebarCollapsed" class="logo-text">HKCE Drive</span>
          </div>
        </div>
        
        <!-- 导航菜单 -->
        <a-menu 
          class="sidebar-menu"
          :selected-keys="[activeView]"
          @click="handleMenuClick"
        >
          <a-menu-item key="files">
            <template #icon>⊞</template>
            文件管理
          </a-menu-item>
          
          <a-menu-item key="sync">
            <template #icon>↻</template>
            同步中心
          </a-menu-item>
          
          <a-menu-item key="minio" v-if="minioEnabled">
            <template #icon>⟐</template>
            云端文件
          </a-menu-item>
          
          <a-menu-item key="settings">
            <template #icon>⚙</template>
            系统设置
          </a-menu-item>
        </a-menu>
        
        <!-- 用户信息 -->
        <div class="sidebar-user" v-if="!sidebarCollapsed">
          <a-card class="user-card" :bordered="false">
            <div class="user-info">
              <a-avatar class="user-avatar" :size="40">
                👤
              </a-avatar>
              <div class="user-details">
                <div class="user-name">{{ currentUser }}</div>
                <div class="user-status">
                  <a-tag :color="minioEnabled ? 'blue' : 'green'" size="small">
                    {{ minioEnabled ? '云端模式' : '本地模式' }}
                  </a-tag>
                </div>
              </div>
            </div>
            <a-button 
              class="logout-btn" 
              type="text" 
              size="small"
              @click="handleLogout"
            >
              退出
            </a-button>
          </a-card>
        </div>
      </a-layout-sider>

      <!-- 主内容区域 -->
      <a-layout class="main-layout">
        <!-- 顶部状态栏 -->
        <a-layout-header class="modern-header">
          <div class="header-content">
            <div class="header-left">
              <a-breadcrumb class="modern-breadcrumb">
                <a-breadcrumb-item 
                  v-for="(breadcrumb, index) in breadcrumbs" 
                  :key="index"
                  @click="navigateTo(breadcrumb)"
                  :class="{ clickable: index < breadcrumbs.length - 1 }"
                >
                  <span v-if="index === 0">🏠</span>
                  {{ breadcrumb.name }}
                </a-breadcrumb-item>
              </a-breadcrumb>
            </div>
            
            <div class="header-right">
              <div class="header-controls">
                <!-- 搜索框 -->
                <a-input-search
                  class="header-search"
                  placeholder="搜索文件..."
                  style="width: 240px"
                  allow-clear
                />
                
                <!-- 视图切换 -->
                <a-radio-group v-model:value="viewMode" button-style="solid" size="small" class="view-toggle">
                  <a-radio-button value="list">≡</a-radio-button>
                  <a-radio-button value="grid">⊞</a-radio-button>
                </a-radio-group>
                
                <!-- 通知 -->
                <a-badge :count="syncStatus.errors?.length || 0" dot>
                  <a-button type="text" class="notification-btn">
                    🔔
                  </a-button>
                </a-badge>
              </div>
            </div>
          </div>
        </a-layout-header>

        <!-- 主内容区域 -->
        <a-layout-content class="modern-content">
          <!-- 文件管理视图 -->
          <div v-if="activeView === 'files'" class="content-section">
            <div class="section-header">
              <div class="section-title">
                <span class="section-icon">📁</span>
                <h2>本地文件管理</h2>
              </div>
              <div class="section-actions">
                <a-space>
                  <a-button type="primary" @click="triggerFileUpload">
                    📤 上传文件
                  </a-button>
                  <a-button @click="openCurrentFolderInExplorer">
                    📁 打开文件夹
                  </a-button>
                </a-space>
              </div>
            </div>

            <!-- 快速操作栏 -->
            <a-card class="quick-actions-card" :bordered="false">
              <a-row :gutter="16">
                <a-col :span="16">
                  <a-input-group compact>
                    <a-input 
                      v-model:value="newFolderName" 
                      placeholder="新建文件夹名称" 
                      @keyup.enter="createFolder"
                      style="width: calc(100% - 100px)"
                    />
                    <a-button type="primary" @click="createFolder">
                      ⊞ 创建
                    </a-button>
                  </a-input-group>
                </a-col>
                <a-col :span="8">
                  <div class="view-controls">
                    <a-space>
                      <a-button @click="loadFiles()">
                        🔄 刷新
                      </a-button>
                      <a-button @click="goBack" :disabled="currentPath === ''">
                        ← 返回
                      </a-button>
                    </a-space>
                  </div>
                </a-col>
              </a-row>
            </a-card>

            <!-- 文件列表 -->
            <a-card class="files-card" :bordered="false">
              <!-- 列表视图 -->
              <a-table 
                v-if="viewMode === 'list'"
                :dataSource="files" 
                :pagination="false"
                :scroll="{ x: 800 }"
                class="modern-table"
              >
                <a-table-column title="名称" data-index="name" :width="300">
                  <template #default="{ record }">
                    <div class="file-name-cell">
                      <div class="file-icon-wrapper">
                        <span v-if="record.isDir" class="file-icon folder-icon">📁</span>
                        <span v-else class="file-icon file-icon-style">📄</span>
                      </div>
                      
                      <a-input 
                        v-if="isRenaming && renameItem && renameItem.path === record.path"
                        v-model:value="newName"
                        @keyup.enter="finishRename"
                        @keyup.esc="cancelRename"
                        size="small"
                        class="rename-input"
                      />
                      <span 
                        v-else
                        @click="record.isDir ? openFolder(record) : null"
                        :class="{ 'folder-link': record.isDir, 'file-name': true }"
                      >
                        {{ record.name }}
                      </span>
                    </div>
                  </template>
                </a-table-column>
                
                <a-table-column title="大小" data-index="size" :width="120">
                  <template #default="{ record }">
                    <span class="file-size">
                      {{ record.isDir ? '-' : formatSize(record.size) }}
                    </span>
                  </template>
                </a-table-column>
                
                <a-table-column title="修改时间" data-index="updatedAt" :width="180">
                  <template #default="{ record }">
                    <span class="file-date">
                      {{ formatDate(record.updatedAt) }}
                    </span>
                  </template>
                </a-table-column>
                
                <a-table-column title="操作" :width="280">
                  <template #default="{ record }">
                    <a-space class="file-actions">
                      <a-button size="small" type="text" @click="startRename(record)">
                        ✎
                      </a-button>
                      
                      <a-popconfirm
                        title="确定要删除这个文件吗？"
                        @confirm="deleteItem(record)"
                      >
                        <a-button size="small" type="text" danger>
                          ✕
                        </a-button>
                      </a-popconfirm>
                      
                      <a-button 
                        v-if="!record.isDir" 
                        size="small" 
                        type="text"
                        @click="downloadFile(record)"
                      >
                        ↓
                      </a-button>
                      
                      <a-button 
                        v-if="!record.isDir" 
                        size="small" 
                        type="text"
                        @click="previewFile(record)"
                      >
                        ◉
                      </a-button>
                      
                      <a-button 
                        size="small"
                        type="text"
                        @click="openFileInExplorer(record)"
                      >
                        ⊞
                      </a-button>
                    </a-space>
                  </template>
                </a-table-column>
                
                <template #emptyText>
                  <a-empty description="此文件夹为空">
                    <a-button type="primary" @click="triggerFileUpload">
                      上传第一个文件
                    </a-button>
                  </a-empty>
                </template>
              </a-table>

              <!-- 网格视图 -->
              <div v-else class="grid-view">
                <div class="file-grid">
                  <div 
                    v-for="file in files" 
                    :key="file.path"
                    class="file-card"
                    @click="file.isDir ? openFolder(file) : previewFile(file)"
                  >
                    <div class="file-card-icon">
                      <icon-folder v-if="file.isDir" class="folder-icon-large" />
                      <icon-file v-else class="file-icon-large" />
                    </div>
                    <div class="file-card-info">
                      <div class="file-card-name">{{ file.name }}</div>
                      <div class="file-card-meta">
                        <span v-if="!file.isDir">{{ formatSize(file.size) }}</span>
                        <span>{{ formatDate(file.updatedAt) }}</span>
                      </div>
                    </div>
                    <div class="file-card-actions">
                      <a-dropdown>
                        <a-button type="text" size="small">
                          ⋯
                        </a-button>
                        <template #overlay>
                          <a-menu>
                            <a-menu-item @click.stop="startRename(file)">
                              ✎ 重命名
                            </a-menu-item>
                            <a-menu-item @click.stop="deleteItem(file)">
                              ✕ 删除
                            </a-menu-item>
                            <a-menu-item v-if="!file.isDir" @click.stop="downloadFile(file)">
                              ↓ 下载
                            </a-menu-item>
                          </a-menu>
                        </template>
                      </a-dropdown>
                    </div>
                  </div>
                </div>
              </div>
            </a-card>
          </div>

          <!-- MinIO 云端文件视图 -->
          <div v-if="activeView === 'minio'" class="content-section">
            <MinioFiles :enabled="minioEnabled" />
          </div>

          <!-- 同步中心视图 -->
          <div v-if="activeView === 'sync'" class="content-section">
            <div class="section-header">
              <div class="section-title">
                <span class="section-icon">🔄</span>
                <h2>同步中心</h2>
              </div>
            </div>

            <!-- 同步状态卡片 -->
            <a-row :gutter="16" class="sync-stats">
              <a-col :span="6">
                <a-card class="stat-card" :bordered="false">
                  <a-statistic
                    title="同步状态"
                    :value="syncStatus.running ? '运行中' : '已停止'"
                    :value-style="{ color: syncStatus.running ? '#52c41a' : '#ff4d4f' }"
                  >
                    <template #prefix>
                      <span v-if="syncStatus.running">🔄</span>
                      <span v-else>⏸️</span>
                    </template>
                  </a-statistic>
                </a-card>
              </a-col>
              
              <a-col :span="6">
                <a-card class="stat-card" :bordered="false">
                  <a-statistic
                    title="最后同步"
                    :value="formatSyncTime(syncStatus.lastSync)"
                  >
                    <template #prefix>🕐</template>
                  </a-statistic>
                </a-card>
              </a-col>
              
              <a-col :span="6">
                <a-card class="stat-card" :bordered="false">
                  <a-statistic
                    title="已上传"
                    :value="syncStatus.filesUploaded"
                    suffix="个文件"
                  >
                    <template #prefix>📤</template>
                  </a-statistic>
                </a-card>
              </a-col>
              
              <a-col :span="6">
                <a-card class="stat-card" :bordered="false">
                  <a-statistic
                    title="已下载"
                    :value="syncStatus.filesDownloaded"
                    suffix="个文件"
                  >
                    <template #prefix>↙</template>
                  </a-statistic>
                </a-card>
              </a-col>
            </a-row>

            <!-- 同步控制 -->
            <a-card class="sync-control-card" :bordered="false">
              <template #title>
                <div class="card-title">
                  ⚙️ 同步控制
                </div>
              </template>
              
              <a-space direction="vertical" size="large" style="width: 100%">
                <a-row :gutter="16">
                  <a-col :span="12">
                    <div>
                      <label>同步间隔</label>
                      <div style="display: flex; gap: 8px; margin-top: 8px;">
                        <a-input-number 
                          v-model:value="syncInterval" 
                          :min="1" 
                          :max="1440"
                          addon-after="分钟"
                          style="width: 200px"
                        />
                        <a-button type="primary" @click="setSyncInterval">
                          设置
                        </a-button>
                      </div>
                    </div>
                  </a-col>
                  <a-col :span="12">
                    <div>
                      <label>同步模式</label>
                      <a-radio-group v-model:value="syncMode" button-style="solid" style="margin-top: 8px;">
                        <a-radio-button value="full">完全同步</a-radio-button>
                        <a-radio-button value="selective">选择性同步</a-radio-button>
                        <a-radio-button value="backup">备份模式</a-radio-button>
                      </a-radio-group>
                    </div>
                  </a-col>
                </a-row>
                
                <a-space>
                  <a-button 
                    v-if="!syncRunning" 
                    type="primary" 
                    size="large"
                    @click="startSync"
                  >
                    ▶️ 启动同步
                  </a-button>
                  <a-button 
                    v-else 
                    danger 
                    size="large"
                    @click="stopSync"
                  >
                    ⏸ 停止同步
                  </a-button>
                  <a-button @click="updateSyncStatus">
                    🔄 刷新状态
                  </a-button>
                </a-space>
              </a-space>
            </a-card>
            
            <!-- 同步规则管理组件 -->
            <SyncRuleManager />
          </div>


          <!-- 系统设置视图 -->
          <div v-if="activeView === 'settings'" class="content-section">
            <SystemSettings />
          </div>
        </a-layout-content>
      </a-layout>
    </a-layout>

    <!-- 隐藏的文件上传输入框 -->
    <input 
      type="file" 
      id="file-upload" 
      @change="handleFileUpload" 
      multiple
      style="display: none"
    />

    <!-- 文件预览组件 -->
    <FilePreview 
      :file="currentPreviewFile" 
      :visible="previewVisible" 
      @close="closePreview" 
    />
  </div>
</template>

<style scoped>
/* 现代化容器样式 */
.modern-drive-container {
  min-height: 100vh;
  background: #f5f5f5;
}

.modern-layout {
  min-height: 100vh;
}

/* 侧边栏样式 */
.modern-sidebar {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(20px);
  border-right: 1px solid rgba(255, 255, 255, 0.2);
  box-shadow: 2px 0 20px rgba(0, 0, 0, 0.1);
  height: 100vh;
  overflow: hidden;
  position: fixed;
  left: 0;
  top: 0;
}

.sidebar-logo {
  padding: 24px 20px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
}

.logo-wrapper {
  display: flex;
  align-items: center;
  gap: 12px;
}

.logo-icon-wrapper {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
}

.logo-icon {
  color: white;
}

.logo-text {
  font-size: 20px;
  font-weight: 700;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.sidebar-menu {
  border: none;
  background: transparent;
  padding: 16px 12px;
}

.sidebar-menu .ant-menu-item {
  margin-bottom: 8px;
  border-radius: 12px;
  font-weight: 500;
  transition: all 0.3s ease;
}

.sidebar-menu .ant-menu-item:hover {
  background: rgba(102, 126, 234, 0.1);
  transform: translateX(4px);
}

.sidebar-menu .ant-menu-item-selected {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.sidebar-user {
  position: absolute;
  bottom: 20px;
  left: 20px;
  right: 20px;
  z-index: 10;
}

.user-card {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  border: 1px solid rgba(255, 255, 255, 0.3);
  color: white;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.user-avatar {
  background: rgba(255, 255, 255, 0.2);
  border: 2px solid rgba(255, 255, 255, 0.3);
  flex-shrink: 0;
  width: 40px !important;
  height: 40px !important;
  display: flex;
  align-items: center;
  justify-content: center;
}

.user-details {
  flex: 1;
}

.user-name {
  font-weight: 600;
  color: white;
  margin-bottom: 4px;
}

.user-status {
  font-size: 12px;
}

.logout-btn {
  width: 100%;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.3);
  color: white;
}

.logout-btn:hover {
  background: rgba(255, 255, 255, 0.3);
  color: white;
}

/* 主布局样式 */
.main-layout {
  background: transparent;
  margin-left: 200px;
  transition: margin-left 0.2s;
}

.modern-layout .ant-layout-sider-collapsed + .main-layout {
  margin-left: 48px;
}

/* 顶部状态栏样式 */
.modern-header {
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(20px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.2);
  box-shadow: 0 2px 20px rgba(0, 0, 0, 0.05);
  height: 64px;
  position: fixed;
  top: 0;
  right: 0;
  left: 200px;
  z-index: 5;
  transition: left 0.2s;
}

.modern-layout .ant-layout-sider-collapsed ~ .main-layout .modern-header {
  left: 48px;
}

/* 优化内容区域布局 */
.modern-content {
  padding: 20px;
  background: transparent;
  overflow-y: auto;
  height: calc(100vh - 64px);
  padding-top: 20px;
  margin-top: 64px;
}

/* 优化卡片间距 */
.quick-actions-card,
.files-card,
.sync-control-card,
.settings-card,
.stat-card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(20px);
  border-radius: 12px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  margin-bottom: 16px;
}

/* 优化侧边栏菜单项间距 */
.sidebar-menu {
  border: none;
  background: transparent;
  padding: 12px 8px;
}

.sidebar-menu .ant-menu-item {
  margin-bottom: 6px;
  border-radius: 10px;
  font-weight: 500;
  transition: all 0.3s ease;
  height: 42px;
  line-height: 42px;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 100%;
  padding: 0 24px;
}

.modern-breadcrumb {
  font-weight: 500;
}

.modern-breadcrumb .arco-breadcrumb-item {
  color: #4e5969;
}

.modern-breadcrumb .clickable {
  cursor: pointer;
  transition: color 0.2s;
}

.modern-breadcrumb .clickable:hover {
  color: #165dff;
}

.header-controls {
  display: flex;
  align-items: center;
  gap: 12px;
  height: 32px;
}

.header-search {
  border-radius: 20px;
  height: 32px !important;
}

.header-search .ant-input {
  height: 32px !important;
  line-height: 32px;
}

.header-search .ant-input-search-button {
  height: 32px !important;
}

.view-toggle {
  height: 32px !important;
  display: flex;
  align-items: center;
}

.view-toggle .ant-radio-button-wrapper {
  height: 32px !important;
  line-height: 30px !important;
  display: flex;
  align-items: center;
  justify-content: center;
}

.notification-btn {
  border-radius: 50%;
  width: 32px !important;
  height: 32px !important;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* 主内容区域样式 */
.modern-content {
  padding: 20px;
  background: transparent;
  overflow-y: auto;
  height: calc(100vh - 64px);
  padding-top: 20px;
  margin-top: 64px;
}

.content-section {
  max-width: 1400px;
  margin: 0 auto;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.section-title h2 {
  margin: 0;
  color: white;
  font-size: 24px;
  font-weight: 700;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.section-icon {
  color: white;
  font-size: 28px;
}

/* 卡片样式 */
.quick-actions-card,
.files-card,
.sync-control-card,
.settings-card,
.stat-card {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(20px);
  border-radius: 16px;
  border: 1px solid rgba(255, 255, 255, 0.2);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  margin-bottom: 16px;
}

.card-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
}

/* 同步统计卡片 */
.sync-stats {
  margin-bottom: 24px;
}

.stat-card {
  background: linear-gradient(135deg, #f0f9ff 0%, #e0f2fe 100%);
  border: 1px solid rgba(14, 165, 233, 0.1);
}

/* 表格样式 */
.modern-table {
  border-radius: 12px;
  overflow: hidden;
}

.modern-table .ant-table-thead > tr > th {
  background: #f7f8fa;
  font-weight: 600;
  color: #1d2129;
}

.modern-table .ant-table-tbody > tr > td {
  border-bottom: 1px solid #f2f3f5;
}

.file-name-cell {
  display: flex;
  align-items: center;
  gap: 12px;
}

.file-icon-wrapper {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f2f3f5;
}

.folder-icon {
  color: #ff7d00;
  font-size: 18px;
}

.file-icon-style {
  color: #165dff;
  font-size: 18px;
}

.file-name {
  font-weight: 500;
  color: #1d2129;
}

.folder-link {
  color: #165dff;
  cursor: pointer;
  transition: all 0.2s;
  font-weight: 500;
}

.folder-link:hover {
  color: #0e42d2;
  text-decoration: underline;
}

.file-size,
.file-date {
  color: #86909c;
  font-size: 13px;
}

.file-actions {
  opacity: 0.6;
  transition: opacity 0.2s;
}

.modern-table .ant-table-tbody > tr:hover .file-actions {
  opacity: 1;
}

.rename-input {
  max-width: 200px;
}

/* 网格视图样式 */
.grid-view {
  padding: 16px;
}

.file-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 16px;
}

.file-card {
  background: white;
  border-radius: 12px;
  padding: 16px;
  border: 1px solid #f2f3f5;
  cursor: pointer;
  transition: all 0.3s ease;
  position: relative;
}

.file-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.1);
  border-color: #165dff;
}

.file-card-icon {
  text-align: center;
  margin-bottom: 12px;
}

.folder-icon-large,
.file-icon-large {
  font-size: 48px;
}

.folder-icon-large {
  color: #ff7d00;
}

.file-icon-large {
  color: #165dff;
}

.file-card-info {
  text-align: center;
}

.file-card-name {
  font-weight: 500;
  color: #1d2129;
  margin-bottom: 8px;
  word-break: break-all;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.file-card-meta {
  font-size: 12px;
  color: #86909c;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.file-card-actions {
  position: absolute;
  top: 8px;
  right: 8px;
  opacity: 0;
  transition: opacity 0.2s;
}

.file-card:hover .file-card-actions {
  opacity: 1;
}

.view-controls {
  display: flex;
  justify-content: flex-end;
}

/* 响应式设计 */
@media (max-width: 1400px) {
  .content-section {
    max-width: 100%;
  }
  
  .header-search {
    width: 200px !important;
  }
}

@media (max-width: 1200px) {
  .modern-content {
    padding: 16px;
  }
  
  .file-grid {
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
    gap: 12px;
  }
  
  .section-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
  
  .section-actions {
    width: 100%;
  }
  
  .quick-actions-card .ant-row {
    flex-direction: column;
  }
  
  .quick-actions-card .ant-col {
    width: 100% !important;
    margin-bottom: 12px;
  }
}

@media (max-width: 1024px) {
  .modern-sidebar {
    width: 200px;
  }
  
  .main-layout {
    margin-left: 80px;
  }
  
  .modern-header {
    left: 80px;
  }
  
  .sync-stats .ant-col {
    span: 12 !important;
    margin-bottom: 16px;
  }
}

@media (max-width: 768px) {
  .header-content {
    flex-direction: column;
    gap: 12px;
    padding: 12px 16px;
    height: auto;
    min-height: 64px;
  }
  
  .header-search {
    width: 100% !important;
    max-width: 300px;
  }
  
  .modern-header {
    height: auto;
    min-height: 64px;
  }
  
  .modern-content {
    margin-top: 80px;
    padding: 12px;
  }
  
  .sync-stats .ant-col {
    span: 24 !important;
    margin-bottom: 12px;
  }
  
  .file-grid {
    grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
    gap: 8px;
  }
  
  .modern-table {
    font-size: 14px;
  }
  
  .file-actions {
    flex-wrap: wrap;
  }
}

@media (max-width: 480px) {
  .modern-content {
    padding: 8px;
  }
  
  .section-title h2 {
    font-size: 20px;
  }
  
  .file-grid {
    grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
    gap: 6px;
  }
  
  .file-card {
    padding: 12px;
  }
  
  .quick-actions-card,
  .files-card,
  .sync-control-card {
    border-radius: 8px;
    margin-bottom: 12px;
  }
}

/* 动画效果 */
@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.content-section {
  animation: slideIn 0.3s ease-out;
}

/* 滚动条样式 */
.modern-content::-webkit-scrollbar {
  width: 6px;
}

.modern-content::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 3px;
}

.modern-content::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.3);
  border-radius: 3px;
}

.modern-content::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.5);
}
</style>
