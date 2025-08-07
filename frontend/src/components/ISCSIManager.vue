<template>
  <div class="iscsi-manager">
    <!-- iSCSI服务状态 -->
    <a-card class="service-status-card" title="iSCSI服务状态">
      <template #extra>
        <a-switch 
          :model-value="iscsiEnabled"
          @change="toggleISCSIService"
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
            :value="iscsiEnabled ? '运行中' : '已停止'"
            :value-style="{ color: iscsiEnabled ? '#00b42a' : '#f53f3f' }"
          >
            <template #prefix>
              <icon-check-circle v-if="iscsiEnabled" />
              <icon-close-circle v-else />
            </template>
          </a-statistic>
        </a-col>
        
        <a-col :span="6">
          <a-statistic
            title="监听端口"
            :value="iscsiPort"
          >
            <template #prefix><icon-wifi /></template>
          </a-statistic>
        </a-col>
        
        <a-col :span="6">
          <a-statistic
            title="活动目标器"
            :value="stats.activeTargets"
            :suffix="`/ ${stats.totalTargets}`"
          >
            <template #prefix><icon-storage /></template>
          </a-statistic>
        </a-col>
        
        <a-col :span="6">
          <a-statistic
            title="活动会话"
            :value="stats.activeSessions"
            :suffix="`/ ${stats.totalSessions}`"
          >
            <template #prefix><icon-link /></template>
          </a-statistic>
        </a-col>
      </a-row>
    </a-card>

    <!-- 功能选项卡 -->
    <a-card class="iscsi-tabs-card">
      <a-tabs v-model:active-key="activeTab" type="card">
        <!-- 目标器管理 -->
        <a-tab-pane key="targets" title="目标器管理">
          <template #title>
            <icon-storage />
            目标器管理
          </template>
          
          <div class="targets-section">
            <!-- 创建目标器 -->
            <a-card title="创建新目标器" class="create-target-card" :bordered="false">
              <a-form :model="newTarget" layout="inline">
                <a-form-item label="目标器名称" required>
                  <a-input 
                    v-model="newTarget.name" 
                    placeholder="输入目标器名称"
                    style="width: 200px"
                  />
                </a-form-item>
                <a-form-item label="IQN">
                  <a-input 
                    v-model="newTarget.iqn" 
                    placeholder="自动生成或手动输入"
                    style="width: 300px"
                  />
                </a-form-item>
                <a-form-item label="端口">
                  <a-input-number 
                    v-model="newTarget.port" 
                    :min="1024" 
                    :max="65535"
                    placeholder="3260"
                    style="width: 100px"
                  />
                </a-form-item>
                <a-form-item>
                  <a-button 
                    type="primary" 
                    @click="createTarget"
                    :loading="creating"
                  >
                    创建目标器
                  </a-button>
                </a-form-item>
              </a-form>
            </a-card>
            
            <!-- 目标器列表 -->
            <a-card title="目标器列表" :bordered="false">
              <a-table 
                :columns="targetColumns" 
                :data="targets"
                :pagination="false"
                row-key="id"
              >
                <template #status="{ record }">
                  <a-tag 
                    :color="record.status === 'active' ? 'green' : 'gray'"
                  >
                    {{ record.status === 'active' ? '运行中' : '已停止' }}
                  </a-tag>
                </template>
                
                <template #luns="{ record }">
                  <a-tag v-for="lun in record.luns" :key="lun.id" color="blue">
                    LUN{{ lun.id }}: {{ formatSize(lun.size) }}
                  </a-tag>
                </template>
                
                <template #actions="{ record }">
                  <a-space>
                    <a-button 
                      v-if="record.status !== 'active'"
                      type="primary" 
                      size="small"
                      @click="startTarget(record.id)"
                    >
                      启动
                    </a-button>
                    <a-button 
                      v-else
                      status="danger" 
                      size="small"
                      @click="stopTarget(record.id)"
                    >
                      停止
                    </a-button>
                    <a-button 
                      size="small"
                      @click="showTargetConfig(record)"
                    >
                      配置
                    </a-button>
                    <a-popconfirm
                      content="确定要删除这个目标器吗？"
                      @ok="deleteTarget(record.id)"
                    >
                      <a-button 
                        status="danger" 
                        size="small"
                        :disabled="record.status === 'active'"
                      >
                        删除
                      </a-button>
                    </a-popconfirm>
                  </a-space>
                </template>
              </a-table>
            </a-card>
          </div>
        </a-tab-pane>

        <!-- LUN管理 -->
        <a-tab-pane key="luns" title="LUN管理">
          <template #title>
            <icon-hard-drive />
            LUN管理
          </template>
          
          <div v-if="selectedTarget" class="lun-section">
            <a-card :title="`目标器: ${selectedTarget.name} - LUN管理`" :bordered="false">
              <!-- 添加LUN -->
              <a-card title="添加新LUN" class="add-lun-card" :bordered="false">
                <a-form :model="newLUN" layout="vertical">
                  <a-row :gutter="16">
                    <a-col :span="6">
                      <a-form-item label="LUN ID" required>
                        <a-input-number 
                          v-model="newLUN.id" 
                          :min="0" 
                          :max="255"
                          style="width: 100%"
                        />
                      </a-form-item>
                    </a-col>
                    <a-col :span="6">
                      <a-form-item label="LUN名称" required>
                        <a-input v-model="newLUN.name" placeholder="输入LUN名称" />
                      </a-form-item>
                    </a-col>
                    <a-col :span="6">
                      <a-form-item label="类型" required>
                        <a-select v-model="newLUN.type" style="width: 100%">
                          <a-option value="file">文件</a-option>
                          <a-option value="block">块设备</a-option>
                          <a-option value="memory">内存</a-option>
                        </a-select>
                      </a-form-item>
                    </a-col>
                    <a-col :span="6">
                      <a-form-item label="大小(GB)" required>
                        <a-input-number 
                          v-model="newLUN.sizeGB" 
                          :min="0.1" 
                          :step="0.1"
                          style="width: 100%"
                        />
                      </a-form-item>
                    </a-col>
                  </a-row>
                  
                  <a-row :gutter="16">
                    <a-col :span="12">
                      <a-form-item label="路径" required>
                        <a-input 
                          v-model="newLUN.path" 
                          placeholder="文件路径或块设备路径"
                        />
                      </a-form-item>
                    </a-col>
                    <a-col :span="6">
                      <a-form-item label="只读">
                        <a-switch v-model="newLUN.readOnly" />
                      </a-form-item>
                    </a-col>
                    <a-col :span="6">
                      <a-form-item>
                        <a-button 
                          type="primary" 
                          @click="addLUN"
                          :loading="addingLUN"
                          style="margin-top: 32px"
                        >
                          添加LUN
                        </a-button>
                      </a-form-item>
                    </a-col>
                  </a-row>
                </a-form>
              </a-card>
              
              <!-- LUN列表 -->
              <a-table 
                :columns="lunColumns" 
                :data="selectedTarget.luns"
                :pagination="false"
                row-key="id"
              >
                <template #type="{ record }">
                  <a-tag :color="getLUNTypeColor(record.type)">
                    {{ getLUNTypeName(record.type) }}
                  </a-tag>
                </template>
                
                <template #size="{ record }">
                  {{ formatSize(record.size) }}
                </template>
                
                <template #readOnly="{ record }">
                  <a-tag :color="record.readOnly ? 'orange' : 'green'">
                    {{ record.readOnly ? '只读' : '读写' }}
                  </a-tag>
                </template>
                
                <template #status="{ record }">
                  <a-tag :color="record.status === 'online' ? 'green' : 'red'">
                    {{ record.status === 'online' ? '在线' : '离线' }}
                  </a-tag>
                </template>
                
                <template #actions="{ record }">
                  <a-popconfirm
                    content="确定要移除这个LUN吗？"
                    @ok="removeLUN(record.id)"
                  >
                    <a-button status="danger" size="small">
                      移除
                    </a-button>
                  </a-popconfirm>
                </template>
              </a-table>
            </a-card>
          </div>
          
          <a-empty v-else description="请先选择一个目标器">
            <a-button type="primary" @click="activeTab = 'targets'">
              去选择目标器
            </a-button>
          </a-empty>
        </a-tab-pane>

        <!-- 发起者管理 -->
        <a-tab-pane key="initiators" title="发起者管理">
          <template #title>
            <icon-user />
            发起者管理
          </template>
          
          <div v-if="selectedTarget" class="initiator-section">
            <a-card :title="`目标器: ${selectedTarget.name} - 发起者管理`" :bordered="false">
              <!-- 添加发起者 -->
              <a-card title="添加新发起者" class="add-initiator-card" :bordered="false">
                <a-form :model="newInitiator" layout="vertical">
                  <a-row :gutter="16">
                    <a-col :span="12">
                      <a-form-item label="发起者IQN" required>
                        <a-input 
                          v-model="newInitiator.iqn" 
                          placeholder="iqn.1993-08.org.debian:01:client"
                        />
                      </a-form-item>
                    </a-col>
                    <a-col :span="6">
                      <a-form-item label="访问权限" required>
                        <a-select v-model="newInitiator.access" style="width: 100%">
                          <a-option value="rw">读写</a-option>
                          <a-option value="ro">只读</a-option>
                        </a-select>
                      </a-form-item>
                    </a-col>
                    <a-col :span="6">
                      <a-form-item>
                        <a-button 
                          type="primary" 
                          @click="addInitiator"
                          :loading="addingInitiator"
                          style="margin-top: 32px"
                        >
                          添加发起者
                        </a-button>
                      </a-form-item>
                    </a-col>
                  </a-row>
                  
                  <a-row :gutter="16">
                    <a-col :span="12">
                      <a-form-item label="允许的IP地址">
                        <a-input 
                          v-model="newInitiator.ipAddressesStr" 
                          placeholder="192.168.1.100,192.168.1.101 (用逗号分隔)"
                        />
                      </a-form-item>
                    </a-col>
                    <a-col :span="6">
                      <a-form-item label="CHAP用户名">
                        <a-input v-model="newInitiator.username" placeholder="可选" />
                      </a-form-item>
                    </a-col>
                    <a-col :span="6">
                      <a-form-item label="CHAP密码">
                        <a-input-password v-model="newInitiator.password" placeholder="可选" />
                      </a-form-item>
                    </a-col>
                  </a-row>
                </a-form>
              </a-card>
              
              <!-- 发起者列表 -->
              <a-table 
                :columns="initiatorColumns" 
                :data="selectedTarget.initiators"
                :pagination="false"
                row-key="iqn"
              >
                <template #ipAddresses="{ record }">
                  <a-tag v-for="ip in record.ipAddresses" :key="ip" color="blue">
                    {{ ip }}
                  </a-tag>
                </template>
                
                <template #access="{ record }">
                  <a-tag :color="record.access === 'rw' ? 'green' : 'orange'">
                    {{ record.access === 'rw' ? '读写' : '只读' }}
                  </a-tag>
                </template>
                
                <template #auth="{ record }">
                  <a-tag v-if="record.username" color="purple">
                    CHAP认证
                  </a-tag>
                  <a-tag v-else color="gray">
                    无认证
                  </a-tag>
                </template>
                
                <template #actions="{ record }">
                  <a-popconfirm
                    content="确定要移除这个发起者吗？"
                    @ok="removeInitiator(record.iqn)"
                  >
                    <a-button status="danger" size="small">
                      移除
                    </a-button>
                  </a-popconfirm>
                </template>
              </a-table>
            </a-card>
          </div>
          
          <a-empty v-else description="请先选择一个目标器">
            <a-button type="primary" @click="activeTab = 'targets'">
              去选择目标器
            </a-button>
          </a-empty>
        </a-tab-pane>

        <!-- 会话监控 -->
        <a-tab-pane key="sessions" title="会话监控">
          <template #title>
            <icon-link />
            会话监控
          </template>
          
          <a-card title="活动会话" :bordered="false">
            <template #extra>
              <a-button @click="loadSessions">
                <template #icon><icon-refresh /></template>
                刷新
              </a-button>
            </template>
            
            <a-table 
              :columns="sessionColumns" 
              :data="sessions"
              :pagination="false"
              row-key="id"
            >
              <template #status="{ record }">
                <a-tag :color="record.status === 'connected' ? 'green' : 'red'">
                  {{ record.status === 'connected' ? '已连接' : '已断开' }}
                </a-tag>
              </template>
              
              <template #connectedAt="{ record }">
                {{ formatDateTime(record.connectedAt) }}
              </template>
              
              <template #lastActivity="{ record }">
                {{ formatDateTime(record.lastActivity) }}
              </template>
              
              <template #traffic="{ record }">
                <div>
                  <div>读取: {{ formatSize(record.bytesRead) }}</div>
                  <div>写入: {{ formatSize(record.bytesWritten) }}</div>
                </div>
              </template>
            </a-table>
          </a-card>
        </a-tab-pane>

        <!-- 认证设置 -->
        <a-tab-pane key="auth" title="认证设置">
          <template #title>
            <icon-lock />
            认证设置
          </template>
          
          <div v-if="selectedTarget" class="auth-section">
            <a-card :title="`目标器: ${selectedTarget.name} - 认证设置`" :bordered="false">
              <a-form :model="authConfig" layout="vertical">
                <a-form-item label="认证类型" required>
                  <a-radio-group v-model="authConfig.type">
                    <a-radio value="none">无认证</a-radio>
                    <a-radio value="chap">CHAP认证</a-radio>
                    <a-radio value="mutual_chap">双向CHAP认证</a-radio>
                  </a-radio-group>
                </a-form-item>
                
                <div v-if="authConfig.type !== 'none'">
                  <a-row :gutter="16">
                    <a-col :span="12">
                      <a-form-item label="目标器用户名" required>
                        <a-input v-model="authConfig.username" />
                      </a-form-item>
                    </a-col>
                    <a-col :span="12">
                      <a-form-item label="目标器密码" required>
                        <a-input-password v-model="authConfig.password" />
                      </a-form-item>
                    </a-col>
                  </a-row>
                  
                  <div v-if="authConfig.type === 'mutual_chap'">
                    <a-row :gutter="16">
                      <a-col :span="12">
                        <a-form-item label="双向认证用户名" required>
                          <a-input v-model="authConfig.mutualUsername" />
                        </a-form-item>
                      </a-col>
                      <a-col :span="12">
                        <a-form-item label="双向认证密码" required>
                          <a-input-password v-model="authConfig.mutualPassword" />
                        </a-form-item>
                      </a-col>
                    </a-row>
                  </div>
                </div>
                
                <a-form-item>
                  <a-button 
                    type="primary" 
                    @click="saveAuthConfig"
                    :loading="savingAuth"
                  >
                    保存认证设置
                  </a-button>
                </a-form-item>
              </a-form>
            </a-card>
          </div>
          
          <a-empty v-else description="请先选择一个目标器">
            <a-button type="primary" @click="activeTab = 'targets'">
              去选择目标器
            </a-button>
          </a-empty>
        </a-tab-pane>
      </a-tabs>
    </a-card>

    <!-- 目标器配置对话框 -->
    <a-modal
      v-model:visible="showConfigModal"
      title="目标器配置"
      width="800px"
      @ok="saveTargetConfig"
      @cancel="showConfigModal = false"
    >
      <a-form v-if="editingTarget" :model="editingTarget" layout="vertical">
        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="目标器名称" required>
              <a-input v-model="editingTarget.name" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="IQN" required>
              <a-input v-model="editingTarget.iqn" />
            </a-form-item>
          </a-col>
        </a-row>
        
        <a-form-item label="监听端口">
          <a-input-number 
            v-model="editingTarget.port" 
            :min="1024" 
            :max="65535"
            style="width: 200px"
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { 
  CreateISCSITarget, GetISCSITargets, UpdateISCSITarget, DeleteISCSITarget,
  StartISCSITarget, StopISCSITarget, AddISCSILUN, RemoveISCSILUN,
  AddISCSIInitiator, RemoveISCSIInitiator, SetISCSIAuth,
  GetISCSISessions, GetISCSIStats, EnableISCSIService, DisableISCSIService,
  IsISCSIEnabled
} from '../../wailsjs/go/main/App'

// 响应式数据
const iscsiEnabled = ref(false)
const iscsiPort = ref(3260)
const serviceToggling = ref(false)
const activeTab = ref('targets')
const targets = ref([])
const sessions = ref([])
const stats = ref({
  totalTargets: 0,
  activeTargets: 0,
  totalSessions: 0,
  activeSessions: 0,
  totalLUNs: 0,
  totalStorage: 0,
  usedStorage: 0
})

// 选中的目标器
const selectedTarget = ref(null)

// 创建目标器相关
const creating = ref(false)
const newTarget = ref({
  name: '',
  iqn: '',
  port: 3260
})

// LUN管理相关
const addingLUN = ref(false)
const newLUN = ref({
  id: 0,
  name: '',
  type: 'file',
  path: '',
  sizeGB: 1,
  readOnly: false
})

// 发起者管理相关
const addingInitiator = ref(false)
const newInitiator = ref({
  iqn: '',
  ipAddressesStr: '',
  username: '',
  password: '',
  access: 'rw'
})

// 认证配置相关
const savingAuth = ref(false)
const authConfig = ref({
  type: 'none',
  username: '',
  password: '',
  mutualUsername: '',
  mutualPassword: ''
})

// 目标器配置对话框
const showConfigModal = ref(false)
const editingTarget = ref(null)

// 表格列定义
const targetColumns = [
  { title: '名称', dataIndex: 'name', key: 'name' },
  { title: 'IQN', dataIndex: 'iqn', key: 'iqn' },
  { title: '端口', dataIndex: 'port', key: 'port' },
  { title: '状态', dataIndex: 'status', key: 'status', slotName: 'status' },
  { title: 'LUNs', dataIndex: 'luns', key: 'luns', slotName: 'luns' },
  { title: '操作', key: 'actions', slotName: 'actions', width: 200 }
]

const lunColumns = [
  { title: 'LUN ID', dataIndex: 'id', key: 'id' },
  { title: '名称', dataIndex: 'name', key: 'name' },
  { title: '类型', dataIndex: 'type', key: 'type', slotName: 'type' },
  { title: '路径', dataIndex: 'path', key: 'path' },
  { title: '大小', dataIndex: 'size', key: 'size', slotName: 'size' },
  { title: '权限', dataIndex: 'readOnly', key: 'readOnly', slotName: 'readOnly' },
  { title: '状态', dataIndex: 'status', key: 'status', slotName: 'status' },
  { title: '操作', key: 'actions', slotName: 'actions' }
]

const initiatorColumns = [
  { title: '发起者IQN', dataIndex: 'iqn', key: 'iqn' },
  { title: 'IP地址', dataIndex: 'ipAddresses', key: 'ipAddresses', slotName: 'ipAddresses' },
  { title: '访问权限', dataIndex: 'access', key: 'access', slotName: 'access' },
  { title: '认证', dataIndex: 'username', key: 'auth', slotName: 'auth' },
  { title: '操作', key: 'actions', slotName: 'actions' }
]

const sessionColumns = [
  { title: '目标器IQN', dataIndex: 'targetIqn', key: 'targetIqn' },
  { title: '发起者IQN', dataIndex: 'initiatorIqn', key: 'initiatorIqn' },
  { title: '发起者IP', dataIndex: 'initiatorIp', key: 'initiatorIp' },
  { title: '状态', dataIndex: 'status', key: 'status', slotName: 'status' },
  { title: '连接时间', dataIndex: 'connectedAt', key: 'connectedAt', slotName: 'connectedAt' },
  { title: '最后活动', dataIndex: 'lastActivity', key: 'lastActivity', slotName: 'lastActivity' },
  { title: '流量统计', key: 'traffic', slotName: 'traffic' }
]

// 方法定义
const loadData = async () => {
  try {
    iscsiEnabled.value = await IsISCSIEnabled()
    const targetsData = await GetISCSITargets()
    targets.value = targetsData || []
    const statsData = await GetISCSIStats()
    stats.value = statsData || stats.value
  } catch (error) {
    console.error('加载iSCSI数据失败:', error)
  }
}

const loadSessions = async () => {
  try {
    const sessionsData = await GetISCSISessions()
    sessions.value = sessionsData || []
  } catch (error) {
    console.error('加载会话数据失败:', error)
  }
}

const toggleISCSIService = async (enabled) => {
  serviceToggling.value = true
  try {
    if (enabled) {
      await EnableISCSIService(iscsiPort.value)
    } else {
      await DisableISCSIService()
    }
    iscsiEnabled.value = enabled
  } catch (error) {
    console.error('切换iSCSI服务失败:', error)
  } finally {
    serviceToggling.value = false
  }
}

const createTarget = async () => {
  if (!newTarget.value.name) {
    return
  }
  
  creating.value = true
  try {
    await CreateISCSITarget(
      newTarget.value.name,
      newTarget.value.iqn,
      newTarget.value.port || 3260
    )
    
    // 重置表单
    newTarget.value = {
      name: '',
      iqn: '',
      port: 3260
    }
    
    // 重新加载数据
    await loadData()
  } catch (error) {
    console.error('创建目标器失败:', error)
  } finally {
    creating.value = false
  }
}

const startTarget = async (targetId) => {
  try {
    await StartISCSITarget(targetId)
    await loadData()
  } catch (error) {
    console.error('启动目标器失败:', error)
  }
}

const stopTarget = async (targetId) => {
  try {
    await StopISCSITarget(targetId)
    await loadData()
  } catch (error) {
    console.error('停止目标器失败:', error)
  }
}

const deleteTarget = async (targetId) => {
  try {
    await DeleteISCSITarget(targetId)
    await loadData()
  } catch (error) {
    console.error('删除目标器失败:', error)
  }
}

const showTargetConfig = (target) => {
  editingTarget.value = { ...target }
  selectedTarget.value = target
  showConfigModal.value = true
}

const saveTargetConfig = async () => {
  if (!editingTarget.value) return
  
  try {
    await UpdateISCSITarget(
      editingTarget.value.id,
      editingTarget.value.name,
      editingTarget.value.iqn,
      editingTarget.value.port
    )
    
    showConfigModal.value = false
    await loadData()
  } catch (error) {
    console.error('保存目标器配置失败:', error)
  }
}

const addLUN = async () => {
  if (!selectedTarget.value || !newLUN.value.name) {
    return
  }
  
  addingLUN.value = true
  try {
    const sizeBytes = newLUN.value.sizeGB * 1024 * 1024 * 1024 // 转换为字节
    
    await AddISCSILUN(
      selectedTarget.value.id,
      newLUN.value.id,
      newLUN.value.name,
      newLUN.value.type,
      newLUN.value.path,
      sizeBytes,
      newLUN.value.readOnly
    )
    
    // 重置表单
    newLUN.value = {
      id: 0,
      name: '',
      type: 'file',
      path: '',
      sizeGB: 1,
      readOnly: false
    }
    
    // 重新加载数据
    await loadData()
  } catch (error) {
    console.error('添加LUN失败:', error)
  } finally {
    addingLUN.value = false
  }
}

const removeLUN = async (lunId) => {
  if (!selectedTarget.value) return
  
  try {
    await RemoveISCSILUN(selectedTarget.value.id, lunId)
    await loadData()
  } catch (error) {
    console.error('移除LUN失败:', error)
  }
}

const addInitiator = async () => {
  if (!selectedTarget.value || !newInitiator.value.iqn) {
    return
  }
  
  addingInitiator.value = true
  try {
    const ipAddresses = newInitiator.value.ipAddressesStr
      ? newInitiator.value.ipAddressesStr.split(',').map(ip => ip.trim())
      : []
    
    await AddISCSIInitiator(
      selectedTarget.value.id,
      newInitiator.value.iqn,
      ipAddresses,
      newInitiator.value.username,
      newInitiator.value.password,
      newInitiator.value.access
    )
    
    // 重置表单
    newInitiator.value = {
      iqn: '',
      ipAddressesStr: '',
      username: '',
      password: '',
      access: 'rw'
    }
    
    // 重新加载数据
    await loadData()
  } catch (error) {
    console.error('添加发起者失败:', error)
  } finally {
    addingInitiator.value = false
  }
}

const removeInitiator = async (initiatorIqn) => {
  if (!selectedTarget.value) return
  
  try {
    await RemoveISCSIInitiator(selectedTarget.value.id, initiatorIqn)
    await loadData()
  } catch (error) {
    console.error('移除发起者失败:', error)
  }
}

const saveAuthConfig = async () => {
  if (!selectedTarget.value) return
  
  savingAuth.value = true
  try {
    await SetISCSIAuth(
      selectedTarget.value.id,
      authConfig.value.type,
      authConfig.value.username,
      authConfig.value.password,
      authConfig.value.mutualUsername,
      authConfig.value.mutualPassword
    )
    
    await loadData()
  } catch (error) {
    console.error('保存认证配置失败:', error)
  } finally {
    savingAuth.value = false
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

const formatDateTime = (dateTime) => {
  if (!dateTime) return '-'
  return new Date(dateTime).toLocaleString()
}

const getLUNTypeColor = (type) => {
  const colors = {
    file: 'blue',
    block: 'green',
    memory: 'orange'
  }
  return colors[type] || 'gray'
}

const getLUNTypeName = (type) => {
  const names = {
    file: '文件',
    block: '块设备',
    memory: '内存'
  }
  return names[type] || type
}

// 监听选中目标器变化，更新认证配置
const updateAuthConfig = () => {
  if (selectedTarget.value && selectedTarget.value.auth) {
    authConfig.value = { ...selectedTarget.value.auth }
  }
}

// 组件挂载时加载数据
onMounted(() => {
  loadData()
  loadSessions()
  
  // 定期刷新会话数据
  setInterval(() => {
    if (activeTab.value === 'sessions') {
      loadSessions()
    }
  }, 10000) // 每10秒刷新一次
})
</script>

<style scoped>
.iscsi-manager {
  padding: 16px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  min-height: 100vh;
}

.service-status-card {
  margin-bottom: 16px;
  backdrop-filter: blur(10px);
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.iscsi-tabs-card {
  backdrop-filter: blur(10px);
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.create-target-card,
.add-lun-card,
.add-initiator-card {
  margin-bottom: 16px;
  background: rgba(255, 255, 255, 0.05);
}

.targets-section,
.lun-section,
.initiator-section,
.auth-section {
  padding: 16px 0;
}

:deep(.arco-card) {
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(10px);
}

:deep(.arco-card-header) {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(5px);
}

:deep(.arco-table) {
  background: transparent;
}

:deep(.arco-table-thead) {
  background: rgba(255, 255, 255, 0.1);
}

:deep(.arco-table-tbody .arco-table-tr) {
  background: rgba(255, 255, 255, 0.05);
}

:deep(.arco-table-tbody .arco-table-tr:hover) {
  background: rgba(255, 255, 255, 0.1);
}

:deep(.arco-tabs-nav) {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(5px);
}

:deep(.arco-tabs-tab) {
  color: rgba(255, 255, 255, 0.8);
}

:deep(.arco-tabs-tab-active) {
  color: #fff;
  background: rgba(255, 255, 255, 0.2);
}

:deep(.arco-statistic-title) {
  color: rgba(255, 255, 255, 0.8);
}

:deep(.arco-statistic-content) {
  color: #fff;
}
</style>
