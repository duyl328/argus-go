/**
 * SSE (Server-Sent Events) 服务
 * 用于实时接收服务器推送的事件
 */

export interface SSEEvent {
  id?: string
  event?: string
  data: string
  retry?: number
}

export interface FileSystemChangeEvent {
  type: 'create' | 'modify' | 'delete' | 'rename'
  path: string
  name: string
  timestamp: string
  is_dir: boolean
}

type SSEEventCallback = (event: SSEEvent) => void
type FileSystemChangeCallback = (event: FileSystemChangeEvent) => void

class SSEService {
  private eventSource: EventSource | null = null
  private baseURL: string
  private clientID: string | null = null // 存储客户端 ID
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  private reconnectDelay = 1000
  private listeners: Map<string, Set<SSEEventCallback>> = new Map()
  private fileSystemChangeCallbacks: Set<FileSystemChangeCallback> = new Set()
  private isManualClose = false
  private lastHeartbeat: number = Date.now()
  private heartbeatCheckInterval: ReturnType<typeof setInterval> | null = null
  private heartbeatTimeout = 90000 // 90秒没收到心跳就重连（后端每5秒keepalive，每10秒ping）

  constructor() {
    // 从环境变量获取 API 基础 URL
    this.baseURL = import.meta.env.VITE_APP_API_URL || 'http://127.0.0.1:9484/api'
  }

  /**
   * 连接到 SSE 服务器
   */
  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.eventSource && this.eventSource.readyState !== EventSource.CLOSED) {
        console.log('SSE 已经连接')
        resolve()
        return
      }

      this.isManualClose = false
      const sseURL = `${this.baseURL}/v1/sse/connect`

      console.log(`正在连接 SSE: ${sseURL}`)

      this.eventSource = new EventSource(sseURL)

      // 连接打开
      this.eventSource.onopen = () => {
        const isReconnect = this.reconnectAttempts > 0
        if (isReconnect) {
          console.log('✅ SSE 重连成功')
        } else {
          console.log('✅ SSE 连接已建立')
        }

        this.reconnectAttempts = 0
        this.lastHeartbeat = Date.now()
        this.startHeartbeatCheck()

        // 在连接建立后注册自定义事件监听器
        if (!isReconnect) {
          console.log('[SSE] 开始注册自定义事件监听器')
        }

        // 监听 connected 事件（获取客户端 ID）
        this.eventSource!.addEventListener('connected', (event: MessageEvent) => {
          try {
            const data = JSON.parse(event.data)
            this.clientID = data.client_id
            console.log('✅ SSE 客户端 ID:', this.clientID)
          } catch (error) {
            console.error('解析 connected 事件失败:', error)
          }
        })

        // 监听 ping 事件
        this.eventSource!.addEventListener('ping', (event: MessageEvent) => {
          console.log('[SSE] 收到 ping 事件:', event.data)
          this.lastHeartbeat = Date.now()
        })

        // 监听 keepalive 事件
        this.eventSource!.addEventListener('keepalive', (event: MessageEvent) => {
          console.log('[SSE] 收到 keepalive 事件:', event.data)
          this.lastHeartbeat = Date.now()
        })

        // 监听 filesystem-change 事件
        this.eventSource!.addEventListener('filesystem-change', (event: MessageEvent) => {
          console.log('[SSE] 收到 filesystem-change 事件:', event.data)
          this.lastHeartbeat = Date.now()

          try {
            const fsEvent: FileSystemChangeEvent = JSON.parse(event.data)
            console.log('文件系统变化:', fsEvent)

            // 通知所有订阅者
            this.fileSystemChangeCallbacks.forEach(callback => {
              callback(fsEvent)
            })
          } catch (error) {
            console.error('解析文件系统变化事件失败:', error)
          }
        })

        console.log('[SSE] 自定义事件监听器注册完成')
        resolve()
      }

      // 接收消息（默认 message 事件）
      this.eventSource.onmessage = (event) => {
        console.log('[SSE] 收到 message 事件:', event.data)
        this.lastHeartbeat = Date.now()
        this.handleMessage({ data: event.data })
      }

      // 错误处理
      this.eventSource.onerror = (error) => {
        // 只在非重连状态时输出完整错误（避免日志噪音）
        if (this.reconnectAttempts === 0) {
          console.warn('SSE 连接中断:', error)
        } else {
          console.log('SSE 连接断开（正常重连中...）')
        }

        // 不管readyState如何，都尝试重连（因为ERR_INCOMPLETE_CHUNKED_ENCODING时状态可能不是CLOSED）
        const shouldReconnect = !this.isManualClose && this.reconnectAttempts < this.maxReconnectAttempts

        if (this.eventSource?.readyState === EventSource.CLOSED || shouldReconnect) {
          // 清理当前连接
          if (this.eventSource) {
            this.eventSource.close()
            this.eventSource = null
          }

          // 自动重连（如果不是手动关闭）
          if (shouldReconnect) {
            this.reconnectAttempts++
            const delay = this.reconnectDelay * this.reconnectAttempts
            console.log(`🔄 SSE 自动重连中... (${this.reconnectAttempts}/${this.maxReconnectAttempts})`)

            setTimeout(() => {
              this.connect().catch(err => {
                console.error('❌ SSE 重连失败:', err)
              })
            }, delay)
          } else if (this.reconnectAttempts >= this.maxReconnectAttempts) {
            console.error('❌ SSE 重连次数已达上限，请检查网络或刷新页面')
          }

          // 只在第一次错误时reject
          if (this.reconnectAttempts === 0) {
            reject(error)
          }
        }
      }
    })
  }

  /**
   * 启动心跳检查
   */
  private startHeartbeatCheck() {
    // 清除旧的定时器
    if (this.heartbeatCheckInterval) {
      clearInterval(this.heartbeatCheckInterval)
    }

    // 每10秒检查一次心跳
    this.heartbeatCheckInterval = setInterval(() => {
      const timeSinceLastHeartbeat = Date.now() - this.lastHeartbeat

      if (timeSinceLastHeartbeat > this.heartbeatTimeout) {
        console.warn(`心跳超时 (${timeSinceLastHeartbeat}ms)，主动重连...`)

        // 清理当前连接
        if (this.eventSource) {
          this.eventSource.close()
          this.eventSource = null
        }

        // 重连
        this.connect().catch(err => {
          console.error('心跳超时重连失败:', err)
        })
      }
    }, 10000) // 10秒检查一次
  }

  /**
   * 停止心跳检查
   */
  private stopHeartbeatCheck() {
    if (this.heartbeatCheckInterval) {
      clearInterval(this.heartbeatCheckInterval)
      this.heartbeatCheckInterval = null
    }
  }

  /**
   * 断开 SSE 连接
   */
  disconnect() {
    if (this.eventSource) {
      this.isManualClose = true
      this.stopHeartbeatCheck()
      this.eventSource.close()
      this.eventSource = null
      this.listeners.clear()
      this.fileSystemChangeCallbacks.clear()
      console.log('SSE 连接已断开')
    }
  }

  /**
   * 添加事件监听器
   */
  addEventListener(eventType: string, callback: SSEEventCallback) {
    if (!this.listeners.has(eventType)) {
      this.listeners.set(eventType, new Set())

      // 注册到 EventSource
      if (this.eventSource) {
        console.log(`[SSE] 注册事件监听器: ${eventType}`)
        this.eventSource.addEventListener(eventType, (event: MessageEvent) => {
          console.log(`[SSE] 收到事件: ${eventType}, 数据:`, event.data)
          this.handleMessage({
            event: eventType,
            data: event.data,
            id: (event as any).lastEventId
          })
        })
      } else {
        console.warn(`[SSE] 无法注册事件监听器 ${eventType}：eventSource 不存在`)
      }
    }

    this.listeners.get(eventType)!.add(callback)
  }

  /**
   * 移除事件监听器
   */
  removeEventListener(eventType: string, callback: SSEEventCallback) {
    const callbacks = this.listeners.get(eventType)
    if (callbacks) {
      callbacks.delete(callback)
      if (callbacks.size === 0) {
        this.listeners.delete(eventType)
      }
    }
  }

  /**
   * 订阅文件系统变化事件
   */
  onFileSystemChange(callback: FileSystemChangeCallback) {
    this.fileSystemChangeCallbacks.add(callback)

    // 返回取消订阅函数
    return () => {
      this.fileSystemChangeCallbacks.delete(callback)
    }
  }

  /**
   * 处理接收到的消息
   */
  private handleMessage(event: SSEEvent) {
    // 更新心跳时间（所有事件都算作心跳）
    this.lastHeartbeat = Date.now()

    const eventType = event.event || 'message'
    const callbacks = this.listeners.get(eventType)

    if (callbacks) {
      callbacks.forEach(callback => {
        callback(event)
      })
    }
  }

  /**
   * 获取连接状态
   */
  getState(): number {
    return this.eventSource?.readyState ?? EventSource.CLOSED
  }

  /**
   * 是否已连接
   */
  isConnected(): boolean {
    return this.eventSource?.readyState === EventSource.OPEN
  }

  /**
   * 获取客户端 ID
   */
  getClientID(): string | null {
    return this.clientID
  }

  /**
   * 订阅指定路径的文件系统变化
   */
  async subscribe(path: string): Promise<boolean> {
    if (!this.clientID) {
      console.error('无法订阅: 客户端 ID 未初始化')
      return false
    }

    try {
      const response = await fetch(`${this.baseURL}/v1/sse/subscribe`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          client_id: this.clientID,
          path: path
        })
      })

      const result = await response.json()
      if (result.success) {
        console.log(`✅ 订阅成功: ${path}`)
        return true
      } else {
        console.error(`订阅失败: ${result.message}`)
        return false
      }
    } catch (error) {
      console.error('订阅请求失败:', error)
      return false
    }
  }

  /**
   * 取消订阅指定路径
   */
  async unsubscribe(path: string): Promise<boolean> {
    if (!this.clientID) {
      console.error('无法取消订阅: 客户端 ID 未初始化')
      return false
    }

    try {
      const response = await fetch(`${this.baseURL}/v1/sse/unsubscribe`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          client_id: this.clientID,
          path: path
        })
      })

      const result = await response.json()
      if (result.success) {
        console.log(`✅ 取消订阅成功: ${path}`)
        return true
      } else {
        console.error(`取消订阅失败: ${result.message}`)
        return false
      }
    } catch (error) {
      console.error('取消订阅请求失败:', error)
      return false
    }
  }

  /**
   * 获取当前所有订阅
   */
  async getSubscriptions(): Promise<string[]> {
    if (!this.clientID) {
      console.error('无法获取订阅: 客户端 ID 未初始化')
      return []
    }

    try {
      const response = await fetch(`${this.baseURL}/v1/sse/subscriptions/${this.clientID}`)
      const result = await response.json()

      if (result.success) {
        return result.data.subscriptions || []
      } else {
        console.error(`获取订阅失败: ${result.message}`)
        return []
      }
    } catch (error) {
      console.error('获取订阅请求失败:', error)
      return []
    }
  }
}

// 单例模式
export const sseService = new SSEService()
