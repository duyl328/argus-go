<template>
  <div class="timeline-container">
    <RecycleScroller
      :items="blocks"
      :key-field="'id'"
      :min-item-size="100"
      class="scroller"
      v-slot="{ item }"
    >
      <!-- 日期标题 -->
      <div v-if="item.type === 'header'" class="header">
        {{ item.label }}
      </div>

      <!-- 照片行 -->
      <div v-else-if="item.type === 'photoRow'" class="photo-row" :style="{ height: item.rowHeight + 'px' }">
        <div
          v-for="photo in item.photos"
          :key="photo.hash"
          class="photo"
          :style="{
            backgroundColor: photo.color,
            width: photo.displayWidth + 'px',
            height: item.rowHeight + 'px'
          }"
        ></div>
      </div>

      <!-- 未来可能出现的卡片 -->
      <div v-else-if="item.type === 'card'" class="card">
        {{ item.text }}
      </div>
    </RecycleScroller>
  </div>
</template>

<script setup lang="ts">
import { RecycleScroller } from 'vue-virtual-scroller'

// 模拟接口返回的数据结构
const photoApiData = {
  hash: [
    "9df97ccc44529d1d377b36f50f79cd5ab9c0920dd58352630412e4b3f8eceaf5",
    "70886b5e8bf59eeb3cc5fbb1f714c3da5c371e1a0c83ac5ad9c20621c703ff4f"
  ],
  isImage: [true, true],
  takenAt: ["2025-08-02T14:16:36", "2025-08-02T14:16:36"],
  ratio: [1, 0.46153846]
}

const timelineApiData = [
  { date: "2024-10-05", count: 1202 },
  { date: "2024-11-01", count: 1 },
  { date: "2024-11-04", count: 4 },
  { date: "2025-08-02", count: 3 }
]

// 随机颜色生成
function randomColor() {
  const colors = ['#f87171', '#fbbf24', '#34d399', '#60a5fa', '#a78bfa', '#f472b6']
  return colors[Math.floor(Math.random() * colors.length)]
}

// 生成模拟照片
function generatePhotos(count) {
  const photos = []
  for (let i = 0; i < count; i++) {
    const ratio = (Math.random() * 1.5 + 0.5).toFixed(2) // 比例 0.5 ~ 2.0
    photos.push({
      hash: Math.random().toString(36).substring(2, 15),
      color: randomColor(),
      ratio: parseFloat(ratio)
    })
  }
  return photos
}

// 将照片按行布局，每行总宽度接近 containerWidth
function buildPhotoRows(photos, containerWidth, rowHeight) {
  const rows = []
  let currentRow = []
  let currentRatioSum = 0

  photos.forEach(photo => {
    currentRow.push(photo)
    currentRatioSum += photo.ratio

    // 如果行里照片太多，就换行
    if (currentRatioSum >= containerWidth / rowHeight) {
      const scale = containerWidth / (currentRatioSum * rowHeight)
      rows.push({ photos: currentRow, scale })
      currentRow = []
      currentRatioSum = 0
    }
  })

  if (currentRow.length > 0) {
    const scale = containerWidth / (currentRatioSum * rowHeight)
    rows.push({ photos: currentRow, scale })
  }

  return rows
}

// 生成完整 blocks
function generateBlocks() {
  const blocks = []
  let idCounter = 0
  const containerWidth = 800 // 假设容器宽度 800px
  const rowHeight = 200 // 每行高度 200px

  timelineApiData.forEach(entry => {
    // 日期标题
    blocks.push({ id: idCounter++, type: 'header', label: entry.date })

    // 生成该日期对应的照片
    const photos = generatePhotos(entry.count)
    const rows = buildPhotoRows(photos, containerWidth, rowHeight)

    rows.forEach(row => {
      const scaledHeight = rowHeight * row.scale
      const scaledPhotos = row.photos.map(p => ({
        ...p,
        displayWidth: p.ratio * scaledHeight
      }))

      blocks.push({
        id: idCounter++,
        type: 'photoRow',
        rowHeight: scaledHeight,
        photos: scaledPhotos
      })
    })
  })

  return blocks
}

const blocks = generateBlocks()
</script>

<style>
@import "vue-virtual-scroller/dist/vue-virtual-scroller.css";

.timeline-container {
  padding: 16px;
}

.scroller {
  height: 100vh;
  overflow-y: auto;
  display: block;
}

.header {
  font-size: 18px;
  font-weight: bold;
  padding: 12px 0;
  background: #f5f5f5;
  position: sticky;
  top: 0;
  z-index: 10;
}

.photo-row {
  display: flex;
  gap: 4px;
  margin: 8px 0;
}

.photo {
  border-radius: 6px;
  flex-shrink: 0;
}

.card {
  padding: 12px;
  margin: 8px 0;
  border-radius: 12px;
  background: #ddd;
}
</style>
