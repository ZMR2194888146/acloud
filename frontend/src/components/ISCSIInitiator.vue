<template>
  <div class="iscsi-initiator">
    <!-- iSCSI发起器状态 -->
    <a-card class="service-status-card" title="iSCSI发起器状态">
      <template #extra>
        <a-switch 
          :model-value="initiatorEnabled"
          @change="toggleInitiatorService"
          :loading="serviceToggling"
        >
          <template #checked>已启用</template>
          <template #unchecked>已禁用</template>
        </a-switch>
      </template>
      
      <a-row :gutter="16">
        <a-col :span="6">
          <a-statistic
            title="服务状态"
            :value="initiatorEnabled ? '运行中' : '已停止'"
            :value-style="{ color: initiatorEnabled ? '#00b42a' : '#f53f3f' }"
          >
            <template #prefix>
              <icon-check-circle v-if="initiatorEnabled" />
              <icon-close-circle v-else />
            </template>
          </a-statistic>
        </a-col>
        
        <a-col :span="6">
          <a-statistic
            title="已连接目标"
            :value="stats.connectedTargets"
            :suffix="`/ ${stats.totalTargets}`"
          >
            <template #prefix><icon-link /></template>
          </a-statistic>
        </a-col>
        
        <a-col :span="6">
          <a-statistic
            title="挂载磁盘"
            :value="stats.mountedDisks"
          >
            <template #prefix><icon-hard-drive /></template>
          </a-statistic>
        </a-col>
        
        <a-col :span="6">
          <a-statistic
            title="总容量"
            :value="formatSize(stats.totalCapacity)"
          >
            <template #prefix><icon-storage /></template>
          </a-statistic>
        </a-col>
      </a-row>
    </a-card>

    <!-- 功能选项卡 -->
    <a-card class="iscsi-tabs-card">
      <a-tabs v-model:active-key="activeTab" type="card">
        <!-- 目标发现 -->
        <a-tab-pane key="discovery" title="目标发现">
          <template #title>
            <icon-search />
            目标发现
          </template>
          
          <div class="discovery-section">
            <!-- 发现新目标 -->
            <a-card title="发现iSCSI目标" class="discovery-card" :bordered="false">
              <a-form :model="discoveryForm" layout="vertical">
                <a-row :gutter="16">
                  <a-col :span="8">
                    <a-form-item label="目标服务器IP" required>
                      <a-input 
                        v-model="discoveryForm.serverIp" 
                        placeholder="192.168.1.100"
                      />
                    </a-form-item>
                  </a-col>
                  <a-col :span="4">
                    <a-form-item label="端口">
                      <a-input-number 
                        v-model="discoveryForm.port" 
                        :min="1" 
                        :max="65535"
                        placeholder="3260"
                      />
                    </a-form-item>
                  </a-col>
                  <a-col :span="6">
                    <a-form-item label="CHAP用户名">
                      <a-input 
                        v-model="discoveryForm.username" 
                        placeholder="可选"
                      />
                    </a-form-item>
                  </a-col>
                  <a-col :span="6">
                    <a-form-item label="CHAP密码">
                      <a-input-password 
                        v-model="discoveryForm.password" 
                        placeholder="可选"
                      />
                    </a-form-item>
                  </a-col>
                </a-row>
                
                <a-form-item>
                  <a-space>
                    <a-button 
                      type="primary" 
                      @click="discoverTargets"
                      :loading="discovering"
                    >
                      <template #icon><icon-search /></template>
                      发现目标
                    </a-button>
                    <a-button @click="clearDiscovery">
                      <template #icon><icon-refresh /></template>
                      清空结果
                    </a-button>
                  </a-space>
                </a-form-item>
              </a-form>
            </a-card>
            
            <!-- 发现的目标列表 -->
            <a-card title="发现的目标" :bordered="false">
              <a-table 
                :columns="discoveredColumns" 
                :data="discoveredTargets"
                :pagination="false"
                row-key="iqn"
              >
                <template #portal="{ record }">
                  {{ record.serverIp }}:{{ record.port }}
                </template>
                
                <template #status="{ record }">
                  <a-tag 
                    :color="record.available ? 'green' : 'red'"
                  >
                    {{ record.available ? '可用' : '不可用' }}
                  </a-tag>
                </template>
                
                <template #actions="{ record }">
                  <a-space>
                    <a-button 
                      type="primary" 
                      size="small"
                      @click="connectTarget(record)"
                      :disabled="!record.available || isConnected(record.iqn)"
                    >
                      {{ isConnected(record.iqn) ? '已连接' : '连接' }}
                    </a-button>
                    <a-button 
                      size="small"
                      @click="showTargetDetails(record)"
                    >
                      详情
                    </a-button>
                  </a-space>
                </template>
              </a-table>
              
              <a-empty v-if="discoveredTargets.length === 0" description="暂无发现的目标">
                <a-button type="primary" @click="discoverTargets">
                  开始发现
                </a-button>
              </a-empty>
            </a-card>
          </div>
        </a-tab-pane>

        <!-- 连接管理 -->
        <a-tab-pane key="connections" title="连接管理">
          <template #title>
            <icon-link />
            连接管理
          </template>
          
          <div class="connections-section">
            <!-- 活动连接 -->
            <a-card title="活动连接" :bordered="false">
              <template #extra>
                <a-button @click="loadConnections">
                  <template #icon><icon-refresh /></template>
                  刷新
                </a-button>
              </template>
              
              <a-table 
                :columns="connectionColumns" 
                :data="activeConnections"
                :pagination="false"
                row-key="sessionId"
              >
                <template #portal="{ record }">
                  {{ record.serverIp }}:{{ record.port }}
                </template>
                
                <template #status="{ record }">
                  <a-tag 
                    :color="record.status === 'connected' ? 'green' : 'orange'"
                  >
                    {{ getConnectionStatusText(record.status) }}
                  </a-tag>
                </template>
                
                <template #connectedAt="{ record }">
                  {{ formatDateTime(record.connectedAt) }}
                </template>
                
                <template #traffic="{ record }">
                  <div class="traffic-info">
                    <div>↑ {{ formatSize(record.bytesWritten) }}</div>
                    <div>↓ {{ formatSize(record.bytesRead) }}</div>
                  </div>
                </template>
                
                <template #actions="{ record }">
                  <a-space>
                    <a-button 
                      size="small"
                      @click="showConnectionDetails(record)"
                    >
                      详情
                    </a-button>
                    <a-popconfirm
                      content="确定要断开这个连接吗？"
                      @ok="disconnectTarget(record.sessionId)"
                    >
                      <a-button 
                        status="danger" 
                        size="small"
                      >
                        断开
                      </a-button>
                    </a-popconfirm>
                  </a-space>
                </template>
              </a-table>
              
              <a-empty v-if="activeConnections.length === 0" description="暂无活动连接" />
            </a-card>
          </div>
        </a-tab-pane>

        <!-- 磁盘管理 -->
        <a-tab-pane key="disks" title="磁盘管理">
          <template #title>
            <icon-hard-drive />
            磁盘管理
          </template>
          
          <div class="disks-section">
            <!-- iSCSI磁盘列表 -->
            <a-card title="iSCSI磁盘" :bordered="false">
              <template #extra>
                <a-space>
                  <a-button @click="scanDisks">
                    <template #icon><icon-scan /></template>
                    扫描磁盘
                  </a-button>
                  <a-button @click="loadDisks">
                    <template #icon><icon-refresh /></template>
                    刷新
                  </a-button>
                </a-space>
              </template>
              
              <a-table 
                :columns="diskColumns" 
                :data="iscsiDisks"
                :pagination="false"
                row-key="devicePath"
              >
                <template #size="{ record }">
                  {{ formatSize(record.size) }}
                </template>
                
                <template #status="{ record }">
                  <a-tag 
                    :color="record.mounted ? 'green' : 'gray'"
                  >
                    {{ record.mounted ? '已挂载' : '未挂载' }}
                  </a-tag>
                </template>
                
                <template #mountPoint="{ record }">
                  <span v-if="record.mountPoint">{{ record.mountPoint }}</span>
                  <span v-else class="text-gray">-</span>
                </template>
                
                <template #actions="{ record }">
                  <a-space>
                    <a-button 
                      v-if="!record.mounted"
                      type="primary" 
                      size="small"
                      @click="mountDisk(record)"
                    >
                      挂载
                    </a-button>
                    <a-button 
                      v-else
                      status="warning" 
                      size="small"
                      @click="unmountDisk(record)"
                    >
                      卸载
                    </a-button>
                    <a-button 
                      size="small"
                      @click="showDiskDetails(record)"
                    >
                      详情
                    </a-button>
                    <a-button 
                      v-if="record.mounted"
                      size="small"
                      @click="openDiskInExplorer(record)"
                    >
                      <template #icon><icon-folder /></template>
                      打开
                    </a-button>
                  </a-space>
                </template>
              </a-table>
              
              <a-empty v-if="iscsiDisks.length === 0" description="暂无iSCSI磁盘">
                <a-button type="primary" @click="scanDisks">
                  扫描磁盘
                </a-button>
              </a-empty>
            </a-card>
          </div>
        </a-tab-pane>

        <!-- 发起器配置 -->
        <a-tab-pane key="config" title="发起器配置">
          <template #title>
            <icon-settings />
            发起器配置
          </template>
          
          <div class="config-section">
            <a-row :gutter="16">
              <a-col :span="12">
                <a-card title="发起器设置" :bordered="false">
                  <a-form :model="initiatorConfig" layout="vertical">
                    <a-form-item label="发起器名称" required>
                      <a-input v-model="initiatorConfig.name" />
                    </a-form-item>
                    
                    <a-form-item label="发起器IQN" required>
                      <a-input v-model="initiatorConfig.iqn" />
                      <template #help>
                        格式: iqn.yyyy-mm.reverse.domain.name:identifier
                      </template>
                    </a-form-item>
                    
                    <a-form-item label="默认端口">
                      <a-input-number 
                        v-model="initiatorConfig.defaultPort" 
                        :min="1" 
                        :max="65535"
                      />
                    </a-form-item>
                    
                    <a-form-item label="连接超时(秒)">
                      <a-input-number 
                        v-model="initiatorConfig.timeout" 
                        :min="5" 
                        :max="300"
                      />
                    </a-form-item>
                    
                    <a-form-item>
                      <a-button 
                        type="primary" 
                        @click="saveInitiatorConfig"
                        :loading="savingConfig"
                      >
                        保存配置
                      </a-button>
                    </a-form-item>
                  </a-form>
                </a-card>
              </a-col>
              
              <a-col :span="12">
                <a-card title="高级设置" :bordered="false">
                  <a-form :model="advancedConfig" layout="vertical">
                    <a-form-item label="自动重连">
                      <a-switch v-model="advancedConfig.autoReconnect" />
                      <template #help>
                        连接断开时自动尝试重连
                      </template>
                    </a-form-item>
                    
                    <a-form-item label="重连间隔(秒)">
                      <a-input-number 
                        v-model="advancedConfig.reconnectInterval" 
                        :min="5" 
                        :max="300"
                        :disabled="!advancedConfig.autoReconnect"
                      />
                    </a-form-item>
                    
                    <a-form-item label="最大重连次数">
                      <a-input-number 
                        v-model="advancedConfig.maxReconnectAttempts" 
                        :min="1" 
                        :max="100"
                        :disabled="!advancedConfig.autoReconnect"
                      />
                    </a-form-item>
                    
                    <a-form-item label="启用多路径">
                      <a-switch v-model="advancedConfig.enableMultipath" />
                      <template #help>
                        启用多路径IO以提高性能和可靠性
                      </template>
                    </a-form-item>
                    
                    <a-form-item label="队列深度">
                      <a-input-number 
                        v-model="advancedConfig.queueDepth" 
                        :min="1" 
                        :max="256"
                      />
                    </a-form-item>
                    
                    <a-form-item>
                      <a-button 
                        type="primary" 
                        @click="saveAdvancedConfig"
                        :loading="savingAdvanced"
                      >
                        保存高级设置
                      </a-button>
                    </a-form-item>
                  </a-form>
                </a-card>
              </a-col>
            </a-row>
          </div>
        </a-tab-pane>

        <!-- 性能监控 -->
        <a-tab-pane key="monitoring" title="性能监控">
          <template #title>
            <icon-dashboard />
            性能监控
          </template>
          
          <div class="monitoring-section">
            <!-- 性能统计 -->
            <a-row :gutter="16" class="performance-stats">
              <a-col :span="6">
                <a-card class="stat-card" :bordered="false">
                  <a-statistic
                    title="读取IOPS"
                    :value="performanceStats.readIOPS"
                    suffix="ops/s"
                  >
                    <template #prefix><icon-arrow-down /></template>
                  </a-statistic>
                </a-card>
              </a-col>
              
              <a-col :span="6">
                <a-card class="stat-card" :bordered="false">
                  <a-statistic
                    title="写入IOPS"
                    :value="performanceStats.writeIOPS"
                    suffix="ops/s"
                  >
                    <template #prefix><icon-arrow-up /></template>
                  </a-statistic>
                </a-card>
              </a-col>
              
              <a-col :span="6">
                <a-card class="stat-card" :bordered="false">
                  <a-statistic
                    title="读取带宽"
                    :value="formatSize(performanceStats.readBandwidth)"
                    suffix="/s"
                  >
                    <template #prefix><icon-download /></template>
                  </a-statistic>
                </a-card>
              </a-col>
              
              <a-col :span="6">
                <a-card class="stat-card" :bordered="false">
                  <a-statistic
                    title="写入带宽"
                    :value="formatSize(performanceStats.writeBandwidth)"
                    suffix="/s"
                  >
                    <template #prefix><icon-upload /></template>
                  </a-statistic>
                </a-card>
              </a-col>
            </a-row>

            <!-- 延迟统计 -->
            <a-card title="延迟统计" :bordered="false">
              <a-row :gutter="16">
                <a-col :span="8">
                  <a-statistic
                    title="平均延迟"
                    :value="performanceStats.avgLatency"
                    suffix="ms"
                  />
                </a-col>
                <a-col :span="8">
                  <a-statistic
                    title="最小延迟"
                    :value="performanceStats.minLatency"
                    suffix="ms"
                  />
                </a-col>
                <a-col :span="8">
                  <a-statistic
                    title="最大延迟"
                    :value="performanceStats.maxLatency"
                    suffix="ms"
                  />
                </a-col>
              </a-row>
            </a-card>
          </div>
        </a-tab-pane>
      </a-tabs>
    </a-card>

    <!-- 目标详情对话框 -->
    <a-modal
      v-model:visible="showTargetModal"
      title="目标详情"
      width="600px"
      @ok="showTargetModal = false"
      @cancel="showTargetModal = false"
    >
      <div v-if="selectedTarget">
        <a-descriptions :column="2" bordered>
          <a-descriptions-item label="目标IQN">
            {{ selectedTarget.iqn }}
          </a-descriptions-item>
          <a-descriptions-item label="服务器地址">
            {{ selectedTarget.serverIp }}:{{ selectedTarget.port }}
          </a-descriptions-item>
          <a-descriptions-item label="状态">
            <a-tag :color="selectedTarget.available ? 'green' : 'red'">
              {{ selectedTarget.available ? '可用' : '不可用' }}
            </a-tag>
          </a-descriptions-item>
          <a-descriptions-item label="认证类型">
            {{ selectedTarget.authType || '无认证' }}
          </a-descriptions-item>
          <a-descriptions-item label="LUN数量">
            {{ selectedTarget.lunCount || 0 }}
          </a-descriptions-item>
          <a-descriptions-item label="发现时间">
            {{ formatDateTime(selectedTarget.discoveredAt) }}
          </a-descriptions-item>
        </a-descriptions>
      </div>
    </a-modal>

    <!-- 磁盘挂载对话框 -->
    <a-modal
      v-model:visible="showMountModal"
      title="挂载磁盘"
      width="500px"
      @ok="confirmMount"
      @cancel="showMountModal = false"
    >
      <div v-if="selectedDisk">
        <a-form :model="mountForm" layout="vertical">
          <a-form-item label="设备路径">
            <a-input :value="selectedDisk.devicePath" readonly />
          </a-form-item>
          
          <a-form-item label="挂载点" required>
            <a-input 
              v-model="mountForm.mountPoint" 
              placeholder="/mnt/iscsi-disk"
            />
            <template #help>
              指定磁盘挂载的目录路径
            </template>
          </a-form-item>
          
          <a-form-item label="文件系统">
            <a-select v-model="mountForm.filesystem" placeholder="自动检测">
              <a-option value="auto">自动检测</a-option>
              <a-option value="ext4">ext4</a-option>
              <a-option value="ntfs">NTFS</a-option>
              <a-option value="xfs">XFS</a-option>
              <a-option value="btrfs">Btrfs</a-option>
            </a-select>
          </a-form-item>
          
          <a-form-item label="挂载选项">
            <a-checkbox-group v-model="mountForm.options">
              <a-checkbox value="rw">读写模式</a-checkbox>
              <a-checkbox value="noatime">禁用访问时间更新</a-checkbox>
              <a-checkbox value="sync">同步写入</a-checkbox>
            </a-checkbox-group>
          </a-form-item>
        </a-form>
      </div>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

// 响应式数据
const initiatorEnabled = ref(false)
const serviceToggling = ref(false)
const activeTab = ref('discovery')

// 统计数据
const stats = ref({
  connectedTargets: 0,
  totalTargets: 0,
  mountedDisks: 0,
  totalCapacity: 0
})

// 目标发现
const discovering = ref(false)
const discoveryForm = ref({
  serverIp: '',
  port: 3260,
  username: '',
  password: ''
})

const discoveredTargets = ref([])
const activeConnections = ref([])
const iscsiDisks = ref([])

// 配置
const initiatorConfig = ref({
  name: 'ACloud-Initiator',
  iqn: 'iqn.2024-01.com.acloud:initiator',
  defaultPort: 3260,
  timeout: 30
})

const advancedConfig = ref({
  autoReconnect: true,
  reconnectInterval: 10,
  maxReconnectAttempts: 5,
  enableMultipath: false,
  queueDepth: 32
})

// 性能统计
const performanceStats = ref({
  readIOPS: 0,
  writeIOPS: 0,
  readBandwidth: 0,
  writeBandwidth: 0,
  avgLatency: 0,
  minLatency: 0,
  maxLatency: 0
})

// 对话框状态
const showTargetModal = ref(false)
const showMountModal = ref(false)
const selectedTarget = ref(null)
const selectedDisk = ref(null)
const savingConfig = ref(false)
const savingAdvanced = ref(false)

// 挂载表单
const mountForm = ref({
  mountPoint: '',
  filesystem: 'auto',
  options: ['rw']
})

// 表格列定义
const discoveredColumns = [
  { title: '目标IQN', dataIndex: 'iqn', key: 'iqn' },
  { title: '服务器地址', key: 'portal', slotName: 'portal' },
  { title: '状态', key: 'status', slotName: 'status' },
  { title: 'LUN数量', dataIndex: 'lunCount', key: 'lunCount' },
  { title: '操作', key: 'actions', slotName: 'actions', width: 150 }
]

const connectionColumns = [
  { title: '目标IQN', dataIndex: 'targetIqn', key: 'targetIqn' },
  { title: '服务器地址', key: 'portal', slotName: 'portal' },
  { title: '状态', key: 'status', slotName: 'status' },
  { title: '连接时间', key: 'connectedAt', slotName: 'connectedAt' },
  { title: '流量统计', key: 'traffic', slotName: 'traffic' },
  { title: '操作', key: 'actions', slotName: 'actions', width: 120 }
]

const diskColumns = [
  { title: '设备路径', dataIndex: 'devicePath', key: 'devicePath' },
  { title: '目标IQN', dataIndex: 'targetIqn', key: 'targetIqn' },
  { title: '大小', key: 'size', slotName: 'size' },
  { title: '状态', key: 'status', slotName: 'status' },
  { title: '挂载点', key: 'mountPoint', slotName: 'mountPoint' },
  { title: '操作', key: 'actions', slotName: 'actions', width: 200 }
]

// 方法定义
const toggleInitiatorService = async (enabled) => {
  serviceToggling.value = true
  try {
    // 这里调用后端API启用/禁用iSCSI发起器服务
    initiatorEnabled.value = enabled
  } catch (error) {
    console.error('切换发起器服务失败:', error)
  } finally {
    serviceToggling.value = false
  }
}

const discoverTargets = async () => {
  if (!discoveryForm.value.serverIp) return
  
  discovering.value = true
  try {
    // 这里调用后端API发现iSCSI目标
    // const targets = await DiscoverISCSITargets(discoveryForm.value)
    // discoveredTargets.value = targets
    
    // 模拟数据
    discoveredTargets.value = [
      {
        iqn: 'iqn.2024-01.com.example:target1',
        serverIp: discoveryForm.value.serverIp,
        port: discoveryForm.value.port,
        available: true,
        lunCount: 2,
        authType: 'CHAP',
        discoveredAt: new Date()
      }
    ]
  } catch (error) {
    console.error('发现目标失败:', error)
  } finally {
    discovering.value = false
  }
}

const clearDiscovery = () => {
  discoveredTargets.value = []
}

const connectTarget = async (target) => {
  try {
    // 这里调用后端API连接到iSCSI目标
    // await ConnectISCSITarget(target.iqn, target.serverIp, target.port)
    
    // 刷新连接列表
    loadConnections()
  } catch (error) {
    console.error('连接目标失败:', error)
  }
}

const disconnectTarget = async (sessionId) => {
  try {
    // 这里调用后端API断开iSCSI连接
    // await DisconnectISCSITarget(sessionId)
    
    // 刷新连接列表
    loadConnections()
  } catch (error) {
    console.error('断开连接失败:', error)
  }
}

const loadConnections = async () => {
  try {
    // 这里调用后端API获取活动连接
    // const connections = await GetISCSIConnections()
    // activeConnections.value = connections
    
    // 模拟数据
    activeConnections.value = []
  } catch (error) {
    console.error('加载连接失败:', error)
  }
}

const scanDisks = async () => {
  try {
    // 这里调用后端API扫描iSCSI磁盘
    // const disks = await ScanISCSIDisks()
    // iscsiDisks.value = disks
    
    // 模拟数据
    iscsiDisks.value = []
  } catch (error) {
    console.error('扫描磁盘失败:', error)
  }
}

const loadDisks = async () => {
  try {
    // 这里调用后端API获取iSCSI磁盘列表
    // const disks = await GetISCSIDisks()
    // iscsiDisks.value = disks
  } catch (error) {
    console.error('加载磁盘失败:', error)
  }
}

const mountDisk = (disk) => {
  selectedDisk.value = disk
  mountForm.value.mountPoint = `/mnt/iscsi-${disk.devicePath.split('/').pop()}`
  showMountModal.value = true
}

const confirmMount = async () => {
  try {
    // 这里调用后端API挂载磁盘
    // await MountISCSIDisk(selectedDisk.value.devicePath, mountForm.value)
    
    showMountModal.value = false
    loadDisks()
  } catch (error) {
    console.error('挂载磁盘失败:', error)
  }
}

const unmountDisk = async (disk) => {
  try {
    // 这里调用后端API卸载磁盘
    // await UnmountISCSIDisk(disk.devicePath)
    
    loadDisks()
  } catch (error) {
    console.error('卸载磁盘失败:', error)
  }
}

const openDiskInExplorer = async (disk) => {
  try {
    // 这里调用后端API在文件管理器中打开磁盘
    // await OpenInExplorer(disk.mountPoint)
  } catch (error) {
    console.error('打开磁盘失败:', error)
  }
}

const showTargetDetails = (target) => {
  selectedTarget.value = target
  showTargetModal.value = true
}

const showConnectionDetails = (connection) => {
  // 显示连接详情
  console.log('连接详情:', connection)
}

const showDiskDetails = (disk) => {
  // 显示磁盘详情
  console.log('磁盘详情:', disk)
}

const saveInitiatorConfig = async () => {
  savingConfig.value = true
  try {
    // 这里调用后端API保存发起器配置
    // await SaveInitiatorConfig(initiatorConfig.value)
  } catch (error) {
    console.error('保存配置失败:', error)
  } finally {
    savingConfig.value = false
  }
}

const saveAdvancedConfig = async () => {
  savingAdvanced.value = true
  try {
    // 这里调用后端API保存高级配置
    // await SaveAdvancedConfig(advancedConfig.value)
  } catch (error) {
    console.error('保存高级配置失败:', error)
  } finally {
    savingAdvanced.value = false
  }
}

// 工具函数
const formatSize = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatDateTime = (date) => {
  if (!date) return '-'
  return new Date(date).toLocaleString()
}

const isConnected = (iqn) => {
  return activeConnections.value.some(conn => conn.targetIqn === iqn)
}

const getConnectionStatusText = (status) => {
  const statusMap = {
    'connected': '已连接',
    'connecting': '连接中',
    'disconnected': '已断开',
    'error': '错误'
  }
  return statusMap[status] || status
}

// 组件挂载时加载数据
onMounted(() => {
  loadConnections()
  loadDisks()
  
  // 定期更新性能统计
  setInterval(() => {
    // 这里可以调用后端API获取实时性能数据
    // updatePerformanceStats()
  }, 5000)
})
</script>

<style scoped>
.iscsi-initiator {
  padding: 16px;
}

.service-status-card,
.iscsi-tabs-card {
  margin-bottom: 16px;
  border-radius: 12px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
}

.discovery-card {
  margin-bottom: 16px;
  background: linear-gradient(135deg, #f0f9ff 0%, #e0f2fe 100%);
  border-radius: 8px;
}

.discovery-section,
.connections-section,
.disks-section,
.config-section,
.monitoring-section {
  padding: 16px 0;
}

.performance-stats {
  margin-bottom: 24px;
}

.stat-card {
  background: linear-gradient(135deg, #f0f9ff 0%, #e0f2fe 100%);
  border: 1px solid rgba(14, 165, 233, 0.1);
  border-radius: 8px;
}

.traffic-info {
  font-size: 12px;
  line-height: 1.4;
}

.text-gray {
  color: #86909c;
}

/* 表格样式优化 */
.arco-table-th {
  background: #f7f8fa;
  font-weight: 600;
}

.arco-table-td {
  border-bottom: 1px solid #f2f3f5;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .iscsi-initiator {
    padding: 8px;
  }
  
  .performance-stats .arco-col {
    margin-bottom: 16px;
  }
}
</style>
