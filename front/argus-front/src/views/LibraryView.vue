<script setup lang="ts">
import { httpClient } from '@/utils/http.ts'
import { CloseOutline, FolderOutline, CheckmarkCircle, EllipseOutline } from '@vicons/ionicons5'
import { onMounted, ref } from 'vue'
import { NSpace, NInput, NButton, NIcon, NTag,  NCard, NEmpty, useMessage } from 'naive-ui'

// 消息提示
const message = useMessage()

// 当前输入的路径
const inputPath = ref('')

// 已添加的路径列表
interface LibraryPath {
  id: string
  path: string
  enabled: boolean
}

const libraryPaths = ref<LibraryPath[]>([
  { id: '1', path: 'C:\\Users\\Documents\\教师资料', enabled: true },
  { id: '2', path: 'D:\\程序员学习资料', enabled: true }
])

// 选择文件夹
// const selectFolder = async () => {
//   // 这里应该调用 Electron 或其他桌面框架的文件选择 API
//   // 示例代码，实际需要根据你的环境调整
//   try {
//     // const result = await window.electronAPI.selectFolder()
//     // if (result) {
//     //   inputPath.value = result
//     // }
//
//     // 临时模拟
//     inputPath.value = 'C:\\Users\\Documents\\示例路径'
//     message.info('请在实际环境中实现文件夹选择功能')
//   } catch (error) {
//     message.error('选择文件夹失败')
//   }
// }


// 选择文件夹（使用浏览器 File System Access API）
const selectFolder = async () => {
  try {
    // 检查浏览器是否支持
    if (!('showDirectoryPicker' in window)) {
      message.warning('您的浏览器不支持文件夹选择功能，请手动输入路径')
      return
    }

    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    //@ts-expect-error
    const dirHandle = await window.showDirectoryPicker()
    // 获取文件夹名称作为路径（实际项目中可能需要完整路径）
    inputPath.value = dirHandle.name
  } catch (err) {
    // 用户取消选择
    console.log('用户取消选择文件夹')
  }
}


// 添加路径
const addPath = async () => {
  if (!inputPath.value.trim()) {
    message.warning('请输入或选择文件路径')
    return
  }

  // 检查是否已存在
  if (libraryPaths.value.some(item => item.path === inputPath.value)) {
    message.warning('该路径已存在')
    return
  }

  try {
    // 发送到后端
    const response = await httpClient.post('/v1/library', {
      path: inputPath.value
    })

    // 添加到列表
    libraryPaths.value.push({
      id: Date.now().toString(),
      path: inputPath.value,
      enabled: true
    })

    // 清空输入
    inputPath.value = ''
    message.success('添加成功')
  } catch (error) {
    message.error('添加失败，请重试')
  }
}

// 切换启用状态
const toggleEnabled = async (item: LibraryPath) => {
  try {
    item.enabled = !item.enabled

    // 发送到后端
    await httpClient.put(`/api/library/${item.id}`, {
      enabled: item.enabled
    })

    message.success(item.enabled ? '已启用' : '已禁用')
  } catch (error) {
    // 恢复原状态
    item.enabled = !item.enabled
    message.error('操作失败')
  }
}

// 删除路径
const removePath = async (item: LibraryPath) => {
  try {
    // 发送到后端
    await httpClient.delete(`/api/library/${item.id}`)

    // 从列表中移除
    const index = libraryPaths.value.findIndex(p => p.id === item.id)
    if (index > -1) {
      libraryPaths.value.splice(index, 1)
    }

    message.success('删除成功')
  } catch (error) {
    message.error('删除失败')
  }
}

// 开始任务
const startTask = async () => {
  const enabledPaths = libraryPaths.value.filter(item => item.enabled)

  if (enabledPaths.length === 0) {
    message.warning('请至少选择一个启用的路径')
    return
  }

  try {
    const response = await httpClient.post('/api/task/create', {
      paths: enabledPaths.map(item => item.path)
    })

    message.success('任务已创建，正在处理中...')
  } catch (error) {
    message.error('创建任务失败')
  }
}

onMounted(() => {
  httpClient.get("v1/library").then((res) => {
    console.log(res);
  })
})
</script>

<template>
  <div class="library-manager">
    <!-- 顶部：添加资料库区域 -->
    <n-card class="add-section" title="添加资料库" size="small">
      <n-space>
        <n-input
          v-model:value="inputPath"
          size="large"
          placeholder="输入文件夹路径..."
          style="width: 500px"
          @keyup.enter="addPath"
        >
          <template #prefix>
            <n-icon :component="FolderOutline" />
          </template>
          <template #suffix>
            <n-button
              type="tertiary"
              size="small"
              @click="selectFolder"
            >
              选择文件夹
            </n-button>
          </template>
        </n-input>
        <n-button
          type="primary"
          size="large"
          @click="addPath"
        >
          添加
        </n-button>
      </n-space>
    </n-card>

    <!-- 中间：已选择的路径列表 -->
    <n-card class="paths-section" title="资料库列表" size="small">
      <div v-if="libraryPaths.length > 0" class="paths-container">
        <n-tag
          v-for="item in libraryPaths"
          :key="item.id"
          class="library-tag"
          :type="item.enabled ? 'success' : 'default'"
          size="large"
        >
          <n-space align="center" :size="12">
            <n-icon
              size="20"
              :component="item.enabled ? CheckmarkCircle : EllipseOutline"
              class="check-icon"
              @click="toggleEnabled(item)"
            />
            <span class="path-text">{{ item.path }}</span>
            <n-icon
              size="18"
              :component="CloseOutline"
              class="close-icon"
              @click="removePath(item)"
            />
          </n-space>
        </n-tag>
      </div>
      <n-empty v-else description="暂无资料库，请添加文件夹路径" />
    </n-card>

    <!-- 底部：操作按钮 -->
    <n-card class="action-section" size="small">
      <n-space justify="start">
        <n-button
          type="primary"
          size="large"
          @click="startTask"
          :disabled="libraryPaths.length === 0"
        >
          <template #icon>
            <span class="iconfont icon-zhongxinkaishi"></span>
          </template>
          开始检索
        </n-button>
      </n-space>
    </n-card>
  </div>
</template>

<style scoped>
.library-manager {
  padding: 20px;
  margin: 0 auto;
}

.add-section,
.paths-section,
.action-section {
  margin-bottom: 20px;
}

.paths-container {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.library-tag {
  padding: 8px 12px;
  border-radius: 6px;
  transition: all 0.3s ease;
}

.library-tag:hover {
  transform: translateY(-2px);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.path-text {
  font-size: 14px;
  max-width: 400px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.check-icon {
  cursor: pointer;
  transition: all 0.2s ease;
}

.check-icon:hover {
  transform: scale(1.1);
}

.close-icon {
  cursor: pointer;
  color: #666;
  transition: all 0.2s ease;
}

.close-icon:hover {
  color: #d03050;
  transform: scale(1.1);
}

/* 深色主题适配 */
:deep(.n-card) {
  border-radius: 8px;
}

:deep(.n-card-header) {
  padding: 16px 20px;
}

:deep(.n-card__content) {
  padding: 20px;
}
</style>
