<template>
  <div v-if="visible" class="debug-panel">
    <div class="debug-header">
      <h3>🔬 虚拟滚动调试面板</h3>
      <button @click="$emit('close')" class="close-btn">✕</button>
    </div>

    <div class="debug-content">
      <!-- 数据生成器 -->
      <section class="debug-section">
        <h4>📊 测试数据生成器</h4>
        <div class="input-group">
          <label>文件数量:</label>
          <input v-model.number="testCount" type="number" min="10" max="100000" step="100" />
          <button @click="generateTestData" class="btn-primary">生成测试数据</button>
        </div>
        <div class="presets">
          <button @click="() => generateTestData(100)">100</button>
          <button @click="() => generateTestData(500)">500</button>
          <button @click="() => generateTestData(1000)">1K</button>
          <button @click="() => generateTestData(5000)">5K</button>
          <button @click="() => generateTestData(10000)">1W</button>
          <button @click="() => generateTestData(50000)">5W</button>
        </div>
      </section>

      <!-- 性能指标 -->
      <section class="debug-section">
        <h4>📈 性能指标</h4>
        <table class="metrics-table">
          <tr>
            <td>总文件数:</td>
            <td><strong>{{ metrics.totalItems }}</strong></td>
          </tr>
          <tr>
            <td>渲染DOM数:</td>
            <td><strong>{{ metrics.renderedItems }}</strong></td>
          </tr>
          <tr>
            <td>DOM节约率:</td>
            <td><strong>{{ metrics.savings }}%</strong></td>
          </tr>
          <tr>
            <td>虚拟滚动:</td>
            <td><strong>{{ metrics.virtualScrollEnabled ? '✅ 已启用' : '❌ 未启用' }}</strong></td>
          </tr>
          <tr>
            <td>视图模式:</td>
            <td><strong>{{ metrics.viewMode === 'grid' ? '网格' : '列表' }}</strong></td>
          </tr>
          <tr>
            <td>缩放级别:</td>
            <td><strong>{{ metrics.zoomLevel }}%</strong></td>
          </tr>
        </table>
      </section>

      <!-- 虚拟滚动状态 -->
      <section class="debug-section">
        <h4>🔄 虚拟滚动状态</h4>
        <table class="metrics-table">
          <tr>
            <td>滚动位置:</td>
            <td>{{ metrics.scrollTop }}px</td>
          </tr>
          <tr>
            <td>开始索引:</td>
            <td>{{ metrics.startIndex }}</td>
          </tr>
          <tr>
            <td>结束索引:</td>
            <td>{{ metrics.endIndex }}</td>
          </tr>
          <tr>
            <td>总高度:</td>
            <td>{{ metrics.totalHeight }}px</td>
          </tr>
          <tr v-if="metrics.viewMode === 'grid'">
            <td>列数:</td>
            <td>{{ metrics.columns }}</td>
          </tr>
        </table>
      </section>

      <!-- API 集成状态 -->
      <section class="debug-section">
        <h4>🔌 API 集成</h4>
        <table class="metrics-table">
          <tr>
            <td>数据源:</td>
            <td><strong>{{ useRealApi ? '真实 API' : 'Mock 数据' }}</strong></td>
          </tr>
          <tr v-if="useRealApi">
            <td>加载状态:</td>
            <td><strong>{{ apiLoading ? '⏳ 加载中' : '✅ 就绪' }}</strong></td>
          </tr>
          <tr v-if="useRealApi && apiError">
            <td>错误信息:</td>
            <td class="error-text"><strong>{{ apiError }}</strong></td>
          </tr>
        </table>
        <button @click="$emit('toggle-api')" class="btn-primary" style="width: 100%; margin-top: 8px;">
          {{ useRealApi ? '切换到 Mock 数据' : '切换到真实 API' }}
        </button>
      </section>

      <!-- 操作按钮 -->
      <section class="debug-section">
        <h4>⚡ 快速操作</h4>
        <div class="action-buttons">
          <button @click="scrollToTop">滚动到顶部</button>
          <button @click="scrollToBottom">滚动到底部</button>
          <button @click="scrollToMiddle">滚动到中间</button>
          <button @click="clearData">清空数据</button>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'

interface DebugMetrics {
  totalItems: number
  renderedItems: number
  savings: number
  virtualScrollEnabled: boolean
  viewMode: string
  zoomLevel: number
  scrollTop: number
  startIndex: number
  endIndex: number
  totalHeight: number
  columns?: number
}

const props = defineProps<{
  visible: boolean
  metrics: DebugMetrics
  useRealApi?: boolean
  apiLoading?: boolean
  apiError?: string | null
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'generate', count: number): void
  (e: 'scrollTo', position: 'top' | 'bottom' | 'middle'): void
  (e: 'clear'): void
  (e: 'toggle-api'): void
}>()

const testCount = ref(1000)

function generateTestData(countOrEvent?: number | Event) {
  const count = typeof countOrEvent === 'number' ? countOrEvent : testCount.value
  emit('generate', count)
}

function scrollToTop() {
  emit('scrollTo', 'top')
}

function scrollToBottom() {
  emit('scrollTo', 'bottom')
}

function scrollToMiddle() {
  emit('scrollTo', 'middle')
}

function clearData() {
  emit('clear')
}
</script>

<style scoped>
.debug-panel {
  position: fixed;
  top: 80px;
  right: 20px;
  width: 350px;
  max-height: calc(100vh - 100px);
  background: rgba(0, 0, 0, 0.95);
  color: #00ff00;
  border: 2px solid #00ff00;
  border-radius: 8px;
  padding: 16px;
  font-family: 'Courier New', monospace;
  font-size: 12px;
  z-index: 10000;
  overflow-y: auto;
  box-shadow: 0 8px 32px rgba(0, 255, 0, 0.3);
}

.debug-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid #00ff00;
}

.debug-header h3 {
  margin: 0;
  font-size: 16px;
  color: #00ff00;
}

.close-btn {
  background: none;
  border: 1px solid #00ff00;
  color: #00ff00;
  width: 24px;
  height: 24px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 16px;
  line-height: 1;
  transition: all 0.2s;
}

.close-btn:hover {
  background: #00ff00;
  color: #000;
}

.debug-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.debug-section {
  background: rgba(0, 255, 0, 0.05);
  padding: 12px;
  border-radius: 4px;
  border: 1px solid rgba(0, 255, 0, 0.2);
}

.debug-section h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
  color: #00ffaa;
}

.input-group {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 8px;
}

.input-group label {
  font-size: 11px;
}

.input-group input {
  flex: 1;
  background: #000;
  border: 1px solid #00ff00;
  color: #00ff00;
  padding: 4px 8px;
  border-radius: 4px;
  font-family: inherit;
}

.presets {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.presets button {
  flex: 1;
  min-width: 50px;
  padding: 4px 8px;
  background: #000;
  border: 1px solid #00ff00;
  color: #00ff00;
  border-radius: 4px;
  cursor: pointer;
  font-family: inherit;
  font-size: 11px;
  transition: all 0.2s;
}

.presets button:hover {
  background: #00ff00;
  color: #000;
}

.btn-primary {
  padding: 4px 12px;
  background: #00ff00;
  border: none;
  color: #000;
  border-radius: 4px;
  cursor: pointer;
  font-family: inherit;
  font-size: 11px;
  font-weight: bold;
  transition: all 0.2s;
}

.btn-primary:hover {
  background: #00ffaa;
}

.metrics-table {
  width: 100%;
  border-collapse: collapse;
}

.metrics-table tr {
  border-bottom: 1px solid rgba(0, 255, 0, 0.1);
}

.metrics-table td {
  padding: 6px 4px;
  font-size: 11px;
}

.metrics-table td:first-child {
  color: #00ffaa;
  width: 50%;
}

.metrics-table td:last-child {
  text-align: right;
}

.metrics-table strong {
  color: #00ff00;
  font-weight: bold;
}

.metrics-table .error-text {
  color: #ff0000;
  font-size: 10px;
}

.metrics-table .error-text strong {
  color: #ff4444;
}

.action-buttons {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 6px;
}

.action-buttons button {
  padding: 6px 12px;
  background: #000;
  border: 1px solid #00ff00;
  color: #00ff00;
  border-radius: 4px;
  cursor: pointer;
  font-family: inherit;
  font-size: 11px;
  transition: all 0.2s;
}

.action-buttons button:hover {
  background: #00ff00;
  color: #000;
}

/* 滚动条样式 */
.debug-panel::-webkit-scrollbar {
  width: 8px;
}

.debug-panel::-webkit-scrollbar-track {
  background: #000;
}

.debug-panel::-webkit-scrollbar-thumb {
  background: #00ff00;
  border-radius: 4px;
}

.debug-panel::-webkit-scrollbar-thumb:hover {
  background: #00ffaa;
}
</style>
