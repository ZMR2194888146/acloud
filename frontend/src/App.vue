<script setup>
import { ref, onMounted } from 'vue'
import { ListFiles, CreateFolder, DeleteFile, RenameFile, GetStoragePath, UploadFile, UploadFileString, DownloadFile, IsLoggedIn, GetCurrentUser, Logout, OpenInExplorer, OpenFileInExplorer, GetSyncStatus, ToggleSyncStatus, GetSystemInfo, SetAutoStart, CheckForUpdatesManually, GetSyncRules, AddSyncRule, UpdateSyncRule, RemoveSyncRule, EnableSyncRule, DisableSyncRule } from '../wailsjs/go/main/App'
import Login from './components/Login.vue'
import FilePreview from './components/FilePreview.vue'
import MinioConfig from './components/MinioConfig.vue'
import MinioFiles from './components/MinioFiles.vue'
import ISCSIInitiator from './components/ISCSIInitiator.vue'
import SystemSettings from './components/SystemSettings.vue'
import SyncRuleManager from './components/SyncRuleManager.vue'

// 登录状态
const isLoggedIn = ref(false)
const currentUser = ref('')

// UI状态
const sidebarCollapsed = ref(false)
const activeView = ref('files')
const viewMode = ref('list')

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

// iSCSI 相关
const showISCSIManager = ref(false)

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
const handleMenuClick = (key) => {
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

// 处理 MinIO 配置更新
const handleMinioConfigUpdated = (enabled) => {
  minioEnabled.value = enabled
  
  // 如果禁用了 MinIO，切换回本地存储并加载本地文件
  if (!enabled) {
    loadFiles()
    loadStoragePath()
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

// 组件挂载时检查登录状态
onMounted(() => {
  checkLoginStatus()
  
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
        :width="280" 
        :collapsed="sidebarCollapsed"
        collapsible
        @collapse="sidebarCollapsed = $event"
      >
        <!-- Logo区域 -->
        <div class="sidebar-logo">
          <div class="logo-wrapper">
            <div class="logo-icon-wrapper">
              <icon-cloud class="logo-icon" :size="28" />
            </div>
            <span v-if="!sidebarCollapsed" class="logo-text">HKCE Drive</span>
          </div>
        </div>
        
        <!-- 导航菜单 -->
        <a-menu 
          class="sidebar-menu"
          :selected-keys="[activeView]"
          @menu-item-click="handleMenuClick"
        >
          <a-menu-item key="files">
            <template #icon><icon-folder /></template>
            文件管理
          </a-menu-item>
          
          <a-menu-item key="sync">
            <template #icon><icon-sync /></template>
            同步中心
          </a-menu-item>
          
          <a-menu-item key="minio" v-if="minioEnabled">
            <template #icon><icon-cloud /></template>
            云端文件
          </a-menu-item>
          
          <a-menu-item key="iscsi">
            <template #icon><icon-storage /></template>
            iSCSI存储
          </a-menu-item>
          
          <a-menu-item key="settings">
            <template #icon><icon-settings /></template>
            系统设置
          </a-menu-item>
        </a-menu>
        
        <!-- 用户信息 -->
        <div class="sidebar-user" v-if="!sidebarCollapsed">
          <a-card class="user-card" :bordered="false">
            <div class="user-info">
              <a-avatar class="user-avatar" :size="40">
                <icon-user />
              </a-avatar>
              <div class="user-details">
                <div class="user-name">{{ currentUser }}</div>
                <div class="user-status">
                  <a-tag :color="minioEnabled ? 'arcoblue' : 'green'" size="small">
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
              <template #icon><icon-export /></template>
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
                  <icon-home v-if="index === 0" />
                  {{ breadcrumb.name }}
                </a-breadcrumb-item>
              </a-breadcrumb>
            </div>
            
            <div class="header-right">
              <a-space>
                <!-- 搜索框 -->
                <a-input-search
                  class="header-search"
                  placeholder="搜索文件..."
                  style="width: 240px"
                  allow-clear
                />
                
                <!-- 视图切换 -->
                <a-radio-group v-model="viewMode" type="button" size="small">
                  <a-radio value="list">
                    <template #radio="{ checked }">
                      <a-button :type="checked ? 'primary' : 'secondary'" size="small">
                        <template #icon><icon-list /></template>
                      </a-button>
                    </template>
                  </a-radio>
                  <a-radio value="grid">
                    <template #radio="{ checked }">
                      <a-button :type="checked ? 'primary' : 'secondary'" size="small">
                        <template #icon><icon-apps /></template>
                      </a-button>
                    </template>
                  </a-radio>
                </a-radio-group>
                
                <!-- 通知 -->
                <a-badge :count="syncStatus.errors?.length || 0" dot>
                  <a-button type="text" class="notification-btn">
                    <template #icon><icon-notification /></template>
                  </a-button>
                </a-badge>
              </a-space>
            </div>
          </div>
        </a-layout-header>

        <!-- 主内容区域 -->
        <a-layout-content class="modern-content">
          <!-- 文件管理视图 -->
          <div v-if="activeView === 'files'" class="content-section">
            <div class="section-header">
              <div class="section-title">
                <icon-folder class="section-icon" />
                <h2>本地文件管理</h2>
              </div>
              <div class="section-actions">
                <a-space>
                  <a-button type="primary" @click="triggerFileUpload">
                    <template #icon><icon-upload /></template>
                    上传文件
                  </a-button>
                  <a-button @click="openCurrentFolderInExplorer">
                    <template #icon><icon-folder /></template>
                    打开文件夹
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
                      v-model="newFolderName" 
                      placeholder="新建文件夹名称" 
                      @keyup.enter="createFolder"
                      style="width: calc(100% - 100px)"
                    />
                    <a-button type="primary" @click="createFolder">
                      <template #icon><icon-folder-add /></template>
                      创建
                    </a-button>
                  </a-input-group>
                </a-col>
                <a-col :span="8">
                  <div class="view-controls">
                    <a-space>
                      <a-button @click="loadFiles()">
                        <template #icon><icon-refresh /></template>
                        刷新
                      </a-button>
                      <a-button @click="goBack" :disabled="currentPath === ''">
                        <template #icon><icon-left /></template>
                        返回
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
                :data="files" 
                :pagination="false"
                :scroll="{ x: 800 }"
                class="modern-table"
              >
                <template #columns>
                  <a-table-column title="名称" data-index="name" :width="300">
                    <template #cell="{ record }">
                      <div class="file-name-cell">
                        <div class="file-icon-wrapper">
                          <icon-folder v-if="record.isDir" class="file-icon folder-icon" />
                          <icon-file v-else class="file-icon file-icon-style" />
                        </div>
                        
                        <a-input 
                          v-if="isRenaming && renameItem && renameItem.path === record.path"
                          v-model="newName"
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
                    <template #cell="{ record }">
                      <span class="file-size">
                        {{ record.isDir ? '-' : formatSize(record.size) }}
                      </span>
                    </template>
                  </a-table-column>
                  
                  <a-table-column title="修改时间" data-index="updatedAt" :width="180">
                    <template #cell="{ record }">
                      <span class="file-date">
                        {{ formatDate(record.updatedAt) }}
                      </span>
                    </template>
                  </a-table-column>
                  
                  <a-table-column title="操作" :width="280">
                    <template #cell="{ record }">
                      <a-space class="file-actions">
                        <a-button size="small" type="text" @click="startRename(record)">
                          <template #icon><icon-edit /></template>
                        </a-button>
                        
                        <a-popconfirm
                          content="确定要删除这个文件吗？"
                          @ok="deleteItem(record)"
                        >
                          <a-button size="small" type="text" status="danger">
                            <template #icon><icon-delete /></template>
                          </a-button>
                        </a-popconfirm>
                        
                        <a-button 
                          v-if="!record.isDir" 
                          size="small" 
                          type="text"
                          @click="downloadFile(record)"
                        >
                          <template #icon><icon-download /></template>
                        </a-button>
                        
                        <a-button 
                          v-if="!record.isDir" 
                          size="small" 
                          type="text"
                          @click="previewFile(record)"
                        >
                          <template #icon><icon-eye /></template>
                        </a-button>
                        
                        <a-button 
                          size="small"
                          type="text"
                          @click="openFileInExplorer(record)"
                        >
                          <template #icon><icon-folder /></template>
                        </a-button>
                      </a-space>
                    </template>
                  </a-table-column>
                </template>
                
                <template #empty>
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
                          <template #icon><icon-more /></template>
                        </a-button>
                        <template #content>
                          <a-doption @click.stop="startRename(file)">
                            <template #icon><icon-edit /></template>
                            重命名
                          </a-doption>
                          <a-doption @click.stop="deleteItem(file)">
                            <template #icon><icon-delete /></template>
                            删除
                          </a-doption>
                          <a-doption v-if="!file.isDir" @click.stop="downloadFile(file)">
                            <template #icon><icon-download /></template>
                            下载
                          </a-doption>
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
                <icon-sync class="section-icon" />
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
                    :value-style="{ color: syncStatus.running ? '#00b42a' : '#f53f3f' }"
                  >
                    <template #prefix>
                      <icon-sync v-if="syncStatus.running" spin />
                      <icon-pause v-else />
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
                    <template #prefix><icon-clock-circle /></template>
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
                    <template #prefix><icon-upload /></template>
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
                    <template #prefix><icon-download /></template>
                  </a-statistic>
                </a-card>
              </a-col>
            </a-row>

            <!-- 同步控制 -->
            <a-card class="sync-control-card" :bordered="false">
              <template #title>
                <div class="card-title">
                  <icon-settings />
                  同步控制
                </div>
              </template>
              
              <a-space direction="vertical" size="large" fill>
                <a-row :gutter="16">
                  <a-col :span="12">
                    <a-form-item label="同步间隔">
                      <a-input-number 
                        v-model="syncInterval" 
                        :min="1" 
                        :max="1440"
                        suffix="分钟"
                        style="width: 200px"
                      />
                      <a-button type="primary" @click="setSyncInterval" style="margin-left: 8px">
                        设置
                      </a-button>
                    </a-form-item>
                  </a-col>
                  <a-col :span="12">
                    <a-form-item label="同步模式">
                      <a-radio-group v-model="syncMode" type="button">
                        <a-radio value="full">完全同步</a-radio>
                        <a-radio value="selective">选择性同步</a-radio>
                        <a-radio value="backup">备份模式</a-radio>
                      </a-radio-group>
                    </a-form-item>
                  </a-col>
                </a-row>
                
                <a-space>
                  <a-button 
                    v-if="!syncRunning" 
                    type="primary" 
                    size="large"
                    @click="startSync"
                  >
                    <template #icon><icon-play-arrow /></template>
                    启动同步
                  </a-button>
                  <a-button 
                    v-else 
                    status="danger" 
                    size="large"
                    @click="stopSync"
                  >
                    <template #icon><icon-pause /></template>
                    停止同步
                  </a-button>
                  <a-button @click="updateSyncStatus">
                    <template #icon><icon-refresh /></template>
                    刷新状态
                  </a-button>
                </a-space>
              </a-space>
            </a-card>
            
            <!-- 同步规则管理组件 -->
            <SyncRuleManager />
          </div>

          <!-- iSCSI 存储视图 -->
          <div v-if="activeView === 'iscsi'" class="content-section">
            <div class="section-header">
              <div class="section-title">
                <icon-storage class="section-icon" />
                <h2>iSCSI 网络存储</h2>
              </div>
            </div>
            <ISCSIInitiator />
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
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
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

.sidebar-menu .arco-menu-item {
  margin-bottom: 8px;
  border-radius: 12px;
  font-weight: 500;
  transition: all 0.3s ease;
}

.sidebar-menu .arco-menu-item:hover {
  background: rgba(102, 126, 234, 0.1);
  transform: translateX(4px);
}

.sidebar-menu .arco-menu-item-selected {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.sidebar-user {
  position: absolute;
  bottom: 20px;
  left: 20px;
  right: 20px;
}

.user-card {
  background: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  border: 1px solid rgba(255, 255, 255, 0.3);
}

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.user-avatar {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.user-details {
  flex: 1;
}

.user-name {
  font-weight: 600;
  color: #1d2129;
  margin-bottom: 4px;
}

.user-status {
  font-size: 12px;
}

.logout-btn {
  width: 100%;
  border-radius: 8px;
}

/* 主布局样式 */
.main-layout {
  background: transparent;
}

/* 顶部状态栏样式 */
.modern-header {
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(20px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.2);
  box-shadow: 0 2px 20px rgba(0, 0, 0, 0.05);
  height: 64px;
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

.header-search {
  border-radius: 20px;
}

.notification-btn {
  border-radius: 50%;
  width: 36px;
  height: 36px;
}

/* 主内容区域样式 */
.modern-content {
  padding: 24px;
  background: transparent;
  overflow-y: auto;
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

.modern-table .arco-table-th {
  background: #f7f8fa;
  font-weight: 600;
  color: #1d2129;
}

.modern-table .arco-table-td {
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

.modern-table .arco-table-tr:hover .file-actions {
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
@media (max-width: 1200px) {
  .modern-content {
    padding: 16px;
  }
  
  .file-grid {
    grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
    gap: 12px;
  }
}

@media (max-width: 768px) {
  .section-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }
  
  .header-content {
    flex-direction: column;
    gap: 12px;
    padding: 12px 16px;
  }
  
  .sync-stats .arco-col {
    margin-bottom: 16px;
  }
  
  .file-grid {
    grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
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
