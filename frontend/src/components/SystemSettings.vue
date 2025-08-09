<template>
  <div class="system-settings">
    <a-card title="系统设置" :bordered="false">
      <a-tabs v-model:activeKey="activeTab" type="card">
        <!-- 基本设置 -->
        <a-tab-pane key="basic" tab="基本设置">
          <a-space direction="vertical" size="large" style="width: 100%">
            <!-- 开机自启动 -->
            <a-card size="small" title="启动设置">
              <a-row :gutter="16" align="middle">
                <a-col :span="18">
                  <div>
                    <div class="setting-title">开机自启动</div>
                    <div class="setting-desc">系统启动时自动运行应用程序</div>
                  </div>
                </a-col>
                <a-col :span="6">
                  <a-switch 
                    v-model:checked="settings.autoStart" 
                    @change="handleAutoStartChange"
                  />
                </a-col>
              </a-row>
            </a-card>

            <!-- 系统托盘 -->
            <a-card size="small" title="系统托盘">
              <a-row :gutter="16" align="middle">
                <a-col :span="18">
                  <div>
                    <div class="setting-title">最小化到托盘</div>
                    <div class="setting-desc">关闭窗口时最小化到系统托盘而不是退出</div>
                  </div>
                </a-col>
                <a-col :span="6">
                  <a-switch 
                    v-model:checked="settings.minimizeToTray" 
                    @change="handleMinimizeToTrayChange"
                  />
                </a-col>
              </a-row>
            </a-card>

            <!-- 通知设置 -->
            <a-card size="small" title="通知设置">
              <a-row :gutter="16" align="middle">
                <a-col :span="18">
                  <div>
                    <div class="setting-title">系统通知</div>
                    <div class="setting-desc">启用系统通知提醒</div>
                  </div>
                </a-col>
                <a-col :span="6">
                  <a-switch 
                    v-model:checked="settings.notifications" 
                    @change="handleNotificationsChange"
                  />
                </a-col>
              </a-row>
            </a-card>
          </a-space>
        </a-tab-pane>

        <!-- 同步设置 -->
        <a-tab-pane key="sync" tab="同步设置">
          <a-space direction="vertical" size="large" style="width: 100%">
            <!-- 自动同步 -->
            <a-card size="small" title="自动同步">
              <a-row :gutter="16" align="middle">
                <a-col :span="18">
                  <div>
                    <div class="setting-title">启用自动同步</div>
                    <div class="setting-desc">自动同步本地文件到云存储</div>
                  </div>
                </a-col>
                <a-col :span="6">
                  <a-switch 
                    v-model:checked="syncStatus.enabled" 
                    @change="handleSyncToggle"
                  />
                </a-col>
              </a-row>
            </a-card>

            <!-- 同步间隔 -->
            <a-card size="small" title="同步间隔">
              <a-row :gutter="16" align="middle">
                <a-col :span="12">
                  <div class="setting-title">同步检查间隔</div>
                </a-col>
                <a-col :span="12">
                  <a-select 
                    v-model:value="settings.syncInterval" 
                    style="width: 100%"
                    @change="handleSyncIntervalChange"
                  >
                    <a-select-option value="1m">1分钟</a-select-option>
                    <a-select-option value="5m">5分钟</a-select-option>
                    <a-select-option value="15m">15分钟</a-select-option>
                    <a-select-option value="30m">30分钟</a-select-option>
                    <a-select-option value="1h">1小时</a-select-option>
                  </a-select>
                </a-col>
              </a-row>
            </a-card>

            <!-- 同步模式 -->
            <a-card size="small" title="同步模式">
              <a-radio-group 
                v-model:value="settings.syncMode" 
                @change="handleSyncModeChange"
              >
                <a-row :gutter="16">
                  <a-col :span="8">
                    <a-radio value="full">
                      <div>
                        <div class="setting-title">完全同步</div>
                        <div class="setting-desc">同步所有文件和文件夹</div>
                      </div>
                    </a-radio>
                  </a-col>
                  <a-col :span="8">
                    <a-radio value="selective">
                      <div>
                        <div class="setting-title">选择性同步</div>
                        <div class="setting-desc">只同步选定的文件和文件夹</div>
                      </div>
                    </a-radio>
                  </a-col>
                  <a-col :span="8">
                    <a-radio value="backup">
                      <div>
                        <div class="setting-title">备份模式</div>
                        <div class="setting-desc">只上传，不下载</div>
                      </div>
                    </a-radio>
                  </a-col>
                </a-row>
              </a-radio-group>
            </a-card>
          </a-space>
        </a-tab-pane>

        <!-- 更新设置 -->
        <a-tab-pane key="update" tab="更新设置">
          <a-space direction="vertical" size="large" style="width: 100%">
            <!-- 当前版本 -->
            <a-card size="small" title="版本信息">
              <a-descriptions :column="1" bordered size="small">
                <a-descriptions-item label="当前版本">
                  {{ systemInfo.version }}
                </a-descriptions-item>
                <a-descriptions-item label="操作系统">
                  {{ systemInfo.os }} ({{ systemInfo.arch }})
                </a-descriptions-item>
                <a-descriptions-item label="存储路径">
                  {{ systemInfo.storage_path }}
                </a-descriptions-item>
              </a-descriptions>
            </a-card>

            <!-- 更新检查 -->
            <a-card size="small" title="更新检查">
              <a-row :gutter="16" align="middle">
                <a-col :span="18">
                  <div>
                    <div class="setting-title">自动检查更新</div>
                    <div class="setting-desc">定期检查应用程序更新</div>
                  </div>
                </a-col>
                <a-col :span="6">
                  <a-switch 
                    v-model:checked="settings.autoUpdate" 
                    @change="handleAutoUpdateChange"
                  />
                </a-col>
              </a-row>
              
              <a-divider />
              
              <a-row :gutter="16">
                <a-col :span="12">
                  <a-button 
                    type="primary" 
                    :loading="updateChecking"
                    @click="checkForUpdates"
                    block
                  >
                    ↻ 检查更新
                  </a-button>
                </a-col>
                <a-col :span="12">
                  <a-button @click="showUpdateHistory" block>
                    📜 更新历史
                  </a-button>
                </a-col>
              </a-row>
            </a-card>
          </a-space>
        </a-tab-pane>

        <!-- 存储设置 -->
        <a-tab-pane key="storage" tab="存储设置">
          <a-space direction="vertical" size="large" style="width: 100%">
            <!-- MinIO 配置 -->
            <a-card size="small" title="MinIO 对象存储">
              <a-row :gutter="16" align="middle" style="margin-bottom: 16px">
                <a-col :span="18">
                  <div>
                    <div class="setting-title">启用 MinIO 存储</div>
                    <div class="setting-desc">使用 MinIO 对象存储作为云端存储后端</div>
                  </div>
                </a-col>
                <a-col :span="6">
                  <a-switch 
                    v-model:checked="minioConfig.enabled" 
                    @change="saveMinioConfig"
                  />
                </a-col>
              </a-row>

              <div v-if="minioConfig.enabled">
                <a-divider />
                
                <a-row :gutter="16" style="margin-bottom: 16px">
                  <a-col :span="12">
                    <div class="setting-title">服务器地址</div>
                    <a-input 
                      v-model:value="minioConfig.endpoint" 
                      placeholder="例如: play.min.io"
                      style="margin-top: 8px"
                    />
                  </a-col>
                  <a-col :span="12">
                    <div class="setting-title">存储桶名称</div>
                    <a-input 
                      v-model:value="minioConfig.bucketName" 
                      placeholder="例如: hkce-cloud"
                      style="margin-top: 8px"
                    />
                  </a-col>
                </a-row>

                <a-row :gutter="16" style="margin-bottom: 16px">
                  <a-col :span="12">
                    <div class="setting-title">访问密钥 ID</div>
                    <a-input 
                      v-model:value="minioConfig.accessKeyId" 
                      placeholder="访问密钥 ID"
                      style="margin-top: 8px"
                    />
                  </a-col>
                  <a-col :span="12">
                    <div class="setting-title">秘密访问密钥</div>
                    <a-input-password 
                      v-model:value="minioConfig.secretAccessKey" 
                      placeholder="秘密访问密钥"
                      style="margin-top: 8px"
                    />
                  </a-col>
                </a-row>

                <a-row :gutter="16" align="middle" style="margin-bottom: 16px">
                  <a-col :span="18">
                    <div>
                      <div class="setting-title">使用 SSL 连接</div>
                      <div class="setting-desc">启用 HTTPS 安全连接</div>
                    </div>
                  </a-col>
                  <a-col :span="6">
                    <a-switch v-model:checked="minioConfig.useSSL" />
                  </a-col>
                </a-row>

                <a-divider />

                <a-row :gutter="16">
                  <a-col :span="12">
                    <a-button 
                      type="default" 
                      :loading="testLoading"
                      @click="testMinioConnection"
                      block
                    >
                      🔗 测试连接
                    </a-button>
                  </a-col>
                  <a-col :span="12">
                    <a-button 
                      type="primary" 
                      :loading="minioLoading"
                      @click="saveMinioConfig"
                      block
                    >
                      ✓ 保存配置
                    </a-button>
                  </a-col>
                </a-row>
              </div>
            </a-card>

            <!-- 本地存储设置 -->
            <a-card size="small" title="本地存储">
              <a-row :gutter="16" align="middle">
                <a-col :span="18">
                  <div>
                    <div class="setting-title">存储路径</div>
                    <div class="setting-desc">{{ systemInfo.storage_path || '未设置' }}</div>
                  </div>
                </a-col>
                <a-col :span="6">
                  <a-button size="small">
                    📁 更改路径
                  </a-button>
                </a-col>
              </a-row>
            </a-card>
          </a-space>
        </a-tab-pane>

        <!-- 高级设置 -->
        <a-tab-pane key="advanced" tab="高级设置">
          <a-space direction="vertical" size="large" style="width: 100%">
            <!-- 日志设置 -->
            <a-card size="small" title="日志设置">
              <a-row :gutter="16" align="middle">
                <a-col :span="12">
                  <div class="setting-title">日志级别</div>
                </a-col>
                <a-col :span="12">
                  <a-select 
                    v-model:value="settings.logLevel" 
                    style="width: 100%"
                    @change="handleLogLevelChange"
                  >
                    <a-select-option value="debug">调试</a-select-option>
                    <a-select-option value="info">信息</a-select-option>
                    <a-select-option value="warn">警告</a-select-option>
                    <a-select-option value="error">错误</a-select-option>
                  </a-select>
                </a-col>
              </a-row>
            </a-card>

            <!-- 性能设置 -->
            <a-card size="small" title="性能设置">
              <a-row :gutter="16" align="middle">
                <a-col :span="18">
                  <div>
                    <div class="setting-title">硬件加速</div>
                    <div class="setting-desc">启用GPU硬件加速（需要重启）</div>
                  </div>
                </a-col>
                <a-col :span="6">
                  <a-switch 
                    v-model:checked="settings.hardwareAcceleration" 
                    @change="handleHardwareAccelerationChange"
                  />
                </a-col>
              </a-row>
            </a-card>

            <!-- 数据管理 -->
            <a-card size="small" title="数据管理">
              <a-space>
                <a-button @click="exportSettings">
                  📤 导出设置
                </a-button>
                <a-button @click="importSettings">
                  📥 导入设置
                </a-button>
                <a-button danger @click="resetSettings">
                  🔄 重置设置
                </a-button>
              </a-space>
            </a-card>
          </a-space>
        </a-tab-pane>
      </a-tabs>
    </a-card>

    <!-- 更新历史对话框 -->
    <a-modal
      v-model:open="updateHistoryVisible"
      title="更新历史"
      :footer="null"
      width="600px"
    >
      <a-timeline>
        <a-timeline-item
          v-for="update in updateHistory"
          :key="update.version"
          :color="update.current ? 'green' : 'blue'"
        >
          <template #dot>
            <span v-if="update.current" style="color: green">✅</span>
          </template>
          <div>
            <div style="font-weight: bold">{{ update.version }}</div>
            <div style="color: #666; font-size: 12px">{{ update.date }}</div>
            <div style="margin-top: 8px">{{ update.description }}</div>
          </div>
        </a-timeline-item>
      </a-timeline>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { GetSystemInfo, SetAutoStart, CheckForUpdatesManually, ToggleSyncStatus, GetSyncStatus, GetMinioConfig, UpdateMinioConfig, TestMinioConnection } from '../../wailsjs/go/main/App'

const emit = defineEmits(['config-updated'])

// 使用简化的消息提示
const message = {
  success: (msg) => console.log('Success:', msg),
  error: (msg) => console.error('Error:', msg),
  warning: (msg) => console.warn('Warning:', msg)
}

const activeTab = ref('basic')
const updateChecking = ref(false)
const updateHistoryVisible = ref(false)
const testLoading = ref(false)
const minioLoading = ref(false)

// 设置数据
const settings = reactive({
  autoStart: false,
  minimizeToTray: true,
  notifications: true,
  syncInterval: '5m',
  syncMode: 'full',
  autoUpdate: true,
  logLevel: 'info',
  hardwareAcceleration: true
})

// 同步状态
const syncStatus = reactive({
  enabled: false,
  running: false,
  interval: '',
  mode: 'full'
})

// 系统信息
const systemInfo = reactive({
  version: '1.0.0',
  os: '',
  arch: '',
  storage_path: '',
  auto_start: false
})

// MinIO 配置
const minioConfig = reactive({
  enabled: false,
  endpoint: '',
  accessKeyId: '',
  secretAccessKey: '',
  bucketName: '',
  useSSL: true
})

// 更新历史
const updateHistory = ref([
  {
    version: '1.0.0',
    date: '2024-01-15',
    description: '初始版本发布，包含基本的文件管理和云存储功能',
    current: true
  }
])

// 加载系统信息
const loadSystemInfo = async () => {
  try {
    const info = await GetSystemInfo()
    Object.assign(systemInfo, info)
    settings.autoStart = info.auto_start
  } catch (error) {
    console.error('加载系统信息失败:', error)
    message.error('加载系统信息失败')
  }
}

// 加载同步状态
const loadSyncStatus = async () => {
  try {
    const status = await GetSyncStatus()
    Object.assign(syncStatus, status)
  } catch (error) {
    console.error('加载同步状态失败:', error)
  }
}

// 加载 MinIO 配置
const loadMinioConfig = async () => {
  try {
    const config = await GetMinioConfig()
    Object.assign(minioConfig, config)
  } catch (error) {
    console.error('加载MinIO配置失败:', error)
    message.error('加载MinIO配置失败')
  }
}

// 处理开机自启动变更
const handleAutoStartChange = async (checked) => {
  try {
    await SetAutoStart(checked)
    message.success(checked ? '已启用开机自启动' : '已禁用开机自启动')
  } catch (error) {
    console.error('设置开机自启动失败:', error)
    message.error('设置开机自启动失败')
    settings.autoStart = !checked // 回滚状态
  }
}

// 处理最小化到托盘变更
const handleMinimizeToTrayChange = (checked) => {
  message.success(checked ? '已启用最小化到托盘' : '已禁用最小化到托盘')
}

// 处理通知设置变更
const handleNotificationsChange = (checked) => {
  message.success(checked ? '已启用系统通知' : '已禁用系统通知')
}

// 处理同步切换
const handleSyncToggle = async (checked) => {
  try {
    const newStatus = await ToggleSyncStatus()
    syncStatus.enabled = newStatus
    message.success(newStatus ? '已启用自动同步' : '已禁用自动同步')
  } catch (error) {
    console.error('切换同步状态失败:', error)
    message.error('切换同步状态失败')
    syncStatus.enabled = !checked // 回滚状态
  }
}

// 处理同步间隔变更
const handleSyncIntervalChange = (value) => {
  message.success(`同步间隔已设置为 ${value}`)
}

// 处理同步模式变更
const handleSyncModeChange = (e) => {
  const mode = e.target.value
  const modeNames = {
    full: '完全同步',
    selective: '选择性同步',
    backup: '备份模式'
  }
  message.success(`同步模式已设置为 ${modeNames[mode]}`)
}

// 处理自动更新变更
const handleAutoUpdateChange = (checked) => {
  message.success(checked ? '已启用自动更新检查' : '已禁用自动更新检查')
}

// 处理日志级别变更
const handleLogLevelChange = (value) => {
  message.success(`日志级别已设置为 ${value}`)
}

// 处理硬件加速变更
const handleHardwareAccelerationChange = (checked) => {
  message.success(checked ? '已启用硬件加速（重启后生效）' : '已禁用硬件加速（重启后生效）')
}

// 检查更新
const checkForUpdates = async () => {
  updateChecking.value = true
  try {
    await CheckForUpdatesManually()
    message.success('更新检查完成')
  } catch (error) {
    console.error('检查更新失败:', error)
    message.error('检查更新失败')
  } finally {
    updateChecking.value = false
  }
}

// 显示更新历史
const showUpdateHistory = () => {
  updateHistoryVisible.value = true
}

// 保存 MinIO 配置
const saveMinioConfig = async () => {
  minioLoading.value = true
  
  try {
    await UpdateMinioConfig(
      minioConfig.endpoint,
      minioConfig.accessKeyId,
      minioConfig.secretAccessKey,
      minioConfig.bucketName,
      minioConfig.useSSL,
      minioConfig.enabled
    )
    
    message.success('MinIO配置已保存')
    emit('config-updated', minioConfig.enabled)
  } catch (error) {
    console.error('保存MinIO配置失败:', error)
    message.error('保存MinIO配置失败')
  } finally {
    minioLoading.value = false
  }
}

// 测试 MinIO 连接
const testMinioConnection = async () => {
  testLoading.value = true
  
  try {
    await TestMinioConnection(
      minioConfig.endpoint,
      minioConfig.accessKeyId,
      minioConfig.secretAccessKey,
      minioConfig.useSSL
    )
    
    message.success('MinIO连接测试成功')
  } catch (error) {
    console.error('MinIO连接测试失败:', error)
    message.error('MinIO连接测试失败')
  } finally {
    testLoading.value = false
  }
}

// 导出设置
const exportSettings = () => {
  const settingsData = JSON.stringify(settings, null, 2)
  const blob = new Blob([settingsData], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'hkce-cloud-settings.json'
  a.click()
  URL.revokeObjectURL(url)
  message.success('设置已导出')
}

// 导入设置
const importSettings = () => {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = '.json'
  input.onchange = (e) => {
    const file = e.target.files[0]
    if (file) {
      const reader = new FileReader()
      reader.onload = (e) => {
        try {
          const importedSettings = JSON.parse(e.target.result)
          Object.assign(settings, importedSettings)
          message.success('设置已导入')
        } catch (error) {
          message.error('导入设置失败：文件格式错误')
        }
      }
      reader.readAsText(file)
    }
  }
  input.click()
}

// 重置设置
const resetSettings = () => {
  Object.assign(settings, {
    autoStart: false,
    minimizeToTray: true,
    notifications: true,
    syncInterval: '5m',
    syncMode: 'full',
    autoUpdate: true,
    logLevel: 'info',
    hardwareAcceleration: true
  })
  message.success('设置已重置为默认值')
}

onMounted(() => {
  loadSystemInfo()
  loadSyncStatus()
  loadMinioConfig()
})
</script>

<style scoped>
.system-settings {
  padding: 20px;
}

.setting-title {
  font-weight: 500;
  margin-bottom: 4px;
}

.setting-desc {
  font-size: 12px;
  color: #666;
  line-height: 1.4;
}

:deep(.ant-card-head) {
  border-bottom: 1px solid #f0f0f0;
}

:deep(.ant-card-body) {
  padding: 16px;
}

:deep(.ant-descriptions-item-label) {
  font-weight: 500;
}
</style>