<template>
  <!-- 这是一个无UI组件，只处理系统托盘功能 -->
</template>

<script setup>
import { onMounted, onUnmounted } from 'vue'
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
import { HandleTrayMenuClick } from '../../wailsjs/go/main/App'

// 监听系统托盘初始化事件
onMounted(() => {
  // 监听系统托盘菜单项初始化
  EventsOn('init-tray', (menuItems) => {
    console.log('系统托盘菜单项已初始化:', menuItems)
    
    // 这里可以添加额外的前端逻辑，如果需要的话
  })
  
  // 监听系统托盘状态变化
  EventsOn('tray-status-changed', (status) => {
    console.log('系统托盘状态已更改:', status)
  })
  
  // 监听窗口关闭请求
  EventsOn('window-close-requested', () => {
    console.log('窗口关闭请求，应用已最小化到系统托盘')
  })
})

// 组件卸载时移除事件监听
onUnmounted(() => {
  EventsOff('init-tray')
  EventsOff('tray-status-changed')
  EventsOff('window-close-requested')
})

// 处理托盘菜单点击
const handleTrayMenuClick = async (menuID) => {
  try {
    await HandleTrayMenuClick(menuID)
  } catch (error) {
    console.error('处理托盘菜单点击失败:', error)
  }
}

// 导出方法供其他组件使用
defineExpose({
  handleTrayMenuClick
})
</script>