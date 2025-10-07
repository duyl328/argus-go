package sse

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Client struct {
	ID          string
	Writer      gin.ResponseWriter
	Request     *http.Request
	Context     *gin.Context
	EventChan   chan *Event
	CloseChan   chan struct{}
	LastPing    time.Time
	UserAgent   string
	RemoteAddr  string
	ConnectedAt time.Time

	// 订阅管理
	subscriptions     map[string]bool // key: 订阅的路径, value: true
	subscriptionMutex sync.RWMutex    // 保护 subscriptions
}

type Event struct {
	ID    string `json:"id,omitempty"`
	Event string `json:"event,omitempty"`
	Data  string `json:"data"`
	Retry int    `json:"retry,omitempty"`
}

type Manager struct {
	clients        map[string]*Client
	clientsMutex   sync.RWMutex
	broadcastChan  chan *Event
	registerChan   chan *Client
	unregisterChan chan string
	eventChan      chan string
	ctx            context.Context
	cancel         context.CancelFunc
	pingInterval   time.Duration
	options        *Options
}

type Options struct {
	PingInterval    time.Duration
	MaxClients      int
	AllowedOrigins  []string
	EnableCORS      bool
	EventBufferSize int
	ClientTimeout   time.Duration
}

func DefaultOptions() *Options {
	return &Options{
		PingInterval:    30 * time.Second,
		MaxClients:      1000,
		AllowedOrigins:  []string{"*"},
		EnableCORS:      true,
		EventBufferSize: 100,
		ClientTimeout:   5 * time.Minute,
	}
}

func NewManager(opts *Options) *Manager {
	if opts == nil {
		opts = DefaultOptions()
	}

	// 确保必要的字段有有效值
	if opts.PingInterval <= 0 {
		opts.PingInterval = 30 * time.Second
	}
	if opts.EventBufferSize <= 0 {
		opts.EventBufferSize = 100
	}
	if opts.MaxClients <= 0 {
		opts.MaxClients = 1000
	}

	ctx, cancel := context.WithCancel(context.Background())

	manager := &Manager{
		clients:        make(map[string]*Client),
		broadcastChan:  make(chan *Event, opts.EventBufferSize),
		registerChan:   make(chan *Client),
		unregisterChan: make(chan string),
		eventChan:      make(chan string, opts.EventBufferSize),
		ctx:            ctx,
		cancel:         cancel,
		pingInterval:   opts.PingInterval,
		options:        opts,
	}

	go manager.run()
	go manager.handleExternalEvents()
	if opts.PingInterval > 0 {
		go manager.periodicPing()
	}

	return manager
}

func (m *Manager) run() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("SSE Manager panic recovered: %v\n", r)
		}
	}()

	for {
		select {
		case <-m.ctx.Done():
			return
		case client := <-m.registerChan:
			m.addClient(client)
		case clientID := <-m.unregisterChan:
			m.removeClient(clientID)
		case event := <-m.broadcastChan:
			m.broadcastToAllClients(event)
		}
	}
}

func (m *Manager) handleExternalEvents() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("SSE Event Handler panic recovered: %v\n", r)
		}
	}()

	for {
		select {
		case <-m.ctx.Done():
			return
		case eventData := <-m.eventChan:
			event := &Event{
				ID:   uuid.New().String(),
				Data: eventData,
			}
			select {
			case m.broadcastChan <- event:
			case <-time.After(1 * time.Second):
			}
		}
	}
}

func (m *Manager) periodicPing() {
	ticker := time.NewTicker(m.pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.pingAllClients()
		}
	}
}

func (m *Manager) addClient(client *Client) {
	if m.options.MaxClients > 0 && len(m.clients) >= m.options.MaxClients {
		close(client.CloseChan)
		return
	}

	m.clientsMutex.Lock()
	m.clients[client.ID] = client
	m.clientsMutex.Unlock()

	fmt.Printf("Client connected: %s (Total: %d)\n", client.ID, len(m.clients))
}

func (m *Manager) removeClient(clientID string) {
	m.clientsMutex.Lock()
	client, exists := m.clients[clientID]
	if exists {
		delete(m.clients, clientID)
		close(client.CloseChan)
	}
	m.clientsMutex.Unlock()

	if exists {
		fmt.Printf("Client disconnected: %s (Total: %d)\n", clientID, len(m.clients))
	}
}

func (m *Manager) broadcastToAllClients(event *Event) {
	m.clientsMutex.RLock()
	clientCount := len(m.clients)

	// 如果没有客户端，直接返回（避免向空列表广播）
	if clientCount == 0 {
		m.clientsMutex.RUnlock()
		fmt.Printf("跳过广播: 当前无客户端连接\n")
		return
	}

	clients := make([]*Client, 0, clientCount)
	for _, client := range m.clients {
		clients = append(clients, client)
	}
	m.clientsMutex.RUnlock()

	// 并发发送，避免单个客户端阻塞
	for _, client := range clients {
		go func(c *Client) {
			// 双重检查：发送前再次确认客户端未关闭
			select {
			case <-c.CloseChan:
				// 客户端已关闭，跳过
				fmt.Printf("跳过向已关闭客户端 %s 发送事件\n", c.ID)
				return
			default:
				// 继续发送
			}

			// 尝试发送事件，带超时和关闭检测
			select {
			case c.EventChan <- event:
				// 发送成功
			case <-time.After(100 * time.Millisecond):
				// 发送超时，断开客户端
				fmt.Printf("向客户端 %s 发送事件超时，断开连接\n", c.ID)
				m.UnregisterClient(c.ID)
			case <-c.CloseChan:
				// 发送过程中客户端关闭
				fmt.Printf("发送过程中客户端 %s 关闭\n", c.ID)
				return
			}
		}(client)
	}
}

func (m *Manager) pingAllClients() {
	pingEvent := &Event{
		Event: "ping",
		Data:  fmt.Sprintf(`{"time":"%s"}`, time.Now().Format(time.RFC3339)),
	}

	// 记录心跳发送
	fmt.Printf("发送心跳事件 (客户端数: %d)\n", len(m.clients))

	select {
	case m.broadcastChan <- pingEvent:
		fmt.Printf("心跳事件已加入广播队列\n")
	case <-time.After(1 * time.Second):
		fmt.Printf("心跳事件发送超时\n")
	}
}

func (m *Manager) RegisterClient(c *gin.Context) {
	if m.options.EnableCORS {
		m.setupCORS(c)
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("X-Accel-Buffering", "no") // 禁用 Nginx 缓冲

	clientID := uuid.New().String()
	client := &Client{
		ID:            clientID,
		Writer:        c.Writer,
		Request:       c.Request,
		Context:       c,
		EventChan:     make(chan *Event, 50),
		CloseChan:     make(chan struct{}),
		LastPing:      time.Now(),
		UserAgent:     c.GetHeader("User-Agent"),
		RemoteAddr:    c.ClientIP(),
		ConnectedAt:   time.Now(),
		subscriptions: make(map[string]bool), // 初始化订阅映射
	}

	m.registerChan <- client

	// 立即发送一个连接成功事件
	connectEvent := &Event{
		Event: "connected",
		Data:  fmt.Sprintf(`{"client_id":"%s","time":"%s"}`, clientID, time.Now().Format(time.RFC3339)),
	}
	select {
	case client.EventChan <- connectEvent:
	case <-time.After(100 * time.Millisecond):
	}

	// 使用 context 来确保资源清理
	streamCtx, streamCancel := context.WithCancel(c.Request.Context())
	defer streamCancel()

	// 使用独立的 goroutine 发送 keepalive
	go func() {
		ticker := time.NewTicker(5 * time.Second) // 5秒发送一次 keepalive
		defer ticker.Stop()
		defer fmt.Printf("Keepalive goroutine exited for client %s\n", clientID)

		for {
			select {
			case <-ticker.C:
				// 双重检查：发送前确认客户端未关闭
				select {
				case <-client.CloseChan:
					return
				case <-streamCtx.Done():
					return
				default:
				}

				keepAliveEvent := &Event{
					Event: "keepalive",
					Data:  fmt.Sprintf(`{"time":"%s"}`, time.Now().Format(time.RFC3339)),
				}

				// 非阻塞发送，避免阻塞 keepalive 协程
				select {
				case client.EventChan <- keepAliveEvent:
					// 成功发送，不打印日志（减少噪音）
				case <-client.CloseChan:
					return
				case <-streamCtx.Done():
					return
				default:
					// EventChan 已满，跳过本次 keepalive
					fmt.Printf("Keepalive skipped for client %s (channel full)\n", clientID)
				}

			case <-client.CloseChan:
				return
			case <-streamCtx.Done():
				return
			case <-m.ctx.Done():
				return
			}
		}
	}()

	c.Stream(func(w io.Writer) bool {
		select {
		case <-client.CloseChan:
			fmt.Printf("Client %s CloseChan triggered in Stream\n", clientID)
			streamCancel()
			return false
		case event := <-client.EventChan:
			// 写入事件，如果失败则断开连接
			err := m.writeEvent(w, event)
			if err != nil {
				fmt.Printf("Failed to write SSE event to client %s: %v\n", clientID, err)
				streamCancel() // 取消 context，通知 keepalive 协程退出
				go m.UnregisterClient(clientID)
				return false
			}
			client.LastPing = time.Now()
			return true
		case <-streamCtx.Done():
			fmt.Printf("Client %s context cancelled in Stream\n", clientID)
			return false
		}
	})

	// Stream 结束后确保客户端被注销
	fmt.Printf("Stream ended for client %s, unregistering\n", clientID)
	go m.UnregisterClient(clientID)
}

func (m *Manager) setupCORS(c *gin.Context) {
	origin := c.GetHeader("Origin")
	if m.isAllowedOrigin(origin) {
		c.Header("Access-Control-Allow-Origin", origin)
	}
	c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Accept, Cache-Control")
	c.Header("Access-Control-Allow-Credentials", "true")
}

func (m *Manager) isAllowedOrigin(origin string) bool {
	for _, allowed := range m.options.AllowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

func (m *Manager) writeEvent(w io.Writer, event *Event) error {
	// 防御性编程：捕获写入错误
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("SSE writeEvent panic recovered: %v, event: %+v\n", r, event)
		}
	}()

	if event == nil {
		return nil
	}

	// 检查 Writer 是否可用
	if w == nil {
		return fmt.Errorf("writer is nil")
	}

	var err error

	// 一次性构建完整的 SSE 消息，减少多次 Fprintf 调用
	var message string

	if event.ID != "" {
		message += fmt.Sprintf("id: %s\n", event.ID)
	}

	if event.Event != "" {
		message += fmt.Sprintf("event: %s\n", event.Event)
	}

	if event.Retry > 0 {
		message += fmt.Sprintf("retry: %d\n", event.Retry)
	}

	message += fmt.Sprintf("data: %s\n\n", event.Data)

	// 一次性写入完整消息
	n, err := fmt.Fprint(w, message)
	if err != nil {
		return fmt.Errorf("failed to write SSE event: %w (bytes written: %d)", err, n)
	}

	// 立即刷新缓冲区
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	} else {
		return fmt.Errorf("writer does not support flushing")
	}

	return nil
}

func (m *Manager) UnregisterClient(clientID string) {
	// 双重检查：确保 manager 未关闭
	select {
	case <-m.ctx.Done():
		// Manager 已关闭，直接返回
		return
	default:
		// Manager 仍在运行，继续注销流程
	}

	select {
	case m.unregisterChan <- clientID:
	case <-time.After(1 * time.Second):
		// 超时也要检查是否因为 manager 关闭导致
	case <-m.ctx.Done():
		// 发送过程中 manager 关闭
		return
	}
}

func (m *Manager) Broadcast(message string) {
	// 检查 manager 是否已关闭
	select {
	case <-m.ctx.Done():
		return
	default:
	}

	event := &Event{
		ID:   uuid.New().String(),
		Data: message,
	}

	select {
	case m.broadcastChan <- event:
	case <-time.After(1 * time.Second):
	case <-m.ctx.Done():
		return
	}
}

func (m *Manager) BroadcastEvent(eventType, message string) {
	// 检查 manager 是否已关闭
	select {
	case <-m.ctx.Done():
		return
	default:
	}

	// 验证数据有效性
	if message == "" {
		fmt.Printf("Warning: 尝试广播空消息，事件类型: %s\n", eventType)
		return
	}

	// 验证JSON格式
	if eventType == "filesystem-change" {
		var testData interface{}
		if err := json.Unmarshal([]byte(message), &testData); err != nil {
			fmt.Printf("Error: 广播事件数据JSON格式无效: %v, 数据: %s\n", err, message)
			return
		}
	}

	event := &Event{
		ID:    uuid.New().String(),
		Event: eventType,
		Data:  message,
	}

	// 安全打印日志
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("BroadcastEvent日志打印panic: %v\n", r)
		}
	}()

	select {
	case m.broadcastChan <- event:
		// 使用安全的字符串格式化
		fmt.Printf("SSE事件已发送: type=%q, data_length=%d\n", eventType, len(message))
	case <-time.After(1 * time.Second):
		fmt.Printf("Warning: SSE事件发送超时: type=%q\n", eventType)
	case <-m.ctx.Done():
		// Manager 关闭中，放弃发送
		return
	}
}

func (m *Manager) SendToClient(clientID, message string) bool {
	// 检查 manager 是否已关闭
	select {
	case <-m.ctx.Done():
		return false
	default:
	}

	m.clientsMutex.RLock()
	client, exists := m.clients[clientID]
	m.clientsMutex.RUnlock()

	if !exists {
		return false
	}

	event := &Event{
		ID:   uuid.New().String(),
		Data: message,
	}

	select {
	case client.EventChan <- event:
		return true
	case <-time.After(100 * time.Millisecond):
		go m.UnregisterClient(clientID)
		return false
	case <-m.ctx.Done():
		return false
	}
}

func (m *Manager) GetEventChannel() chan<- string {
	return m.eventChan
}

func (m *Manager) GetClientCount() int {
	m.clientsMutex.RLock()
	defer m.clientsMutex.RUnlock()
	return len(m.clients)
}

func (m *Manager) GetClientIDs() []string {
	m.clientsMutex.RLock()
	defer m.clientsMutex.RUnlock()

	ids := make([]string, 0, len(m.clients))
	for id := range m.clients {
		ids = append(ids, id)
	}
	return ids
}

func (m *Manager) GetClientInfo(clientID string) (*Client, bool) {
	m.clientsMutex.RLock()
	defer m.clientsMutex.RUnlock()

	client, exists := m.clients[clientID]
	if !exists {
		return nil, false
	}

	clientCopy := *client
	return &clientCopy, true
}

func (m *Manager) Close() {
	// 1. 取消 context，通知所有 goroutines 退出
	m.cancel()

	// 2. 给 goroutines 一些时间优雅退出
	time.Sleep(100 * time.Millisecond)

	// 3. 关闭所有客户端连接
	m.clientsMutex.Lock()
	for _, client := range m.clients {
		// 安全关闭 CloseChan（检查是否已关闭）
		select {
		case <-client.CloseChan:
			// 已经关闭，跳过
		default:
			close(client.CloseChan)
		}
	}
	m.clients = make(map[string]*Client)
	m.clientsMutex.Unlock()

	// 4. 再等待一小段时间确保所有发送操作完成
	time.Sleep(50 * time.Millisecond)

	// 5. 最后关闭所有 channels
	close(m.broadcastChan)
	close(m.registerChan)
	close(m.unregisterChan)
	close(m.eventChan)

	fmt.Println("SSE Manager 已安全关闭")
}

// ============ 订阅管理方法 ============

// Subscribe 客户端订阅指定路径的文件系统变化
func (m *Manager) Subscribe(clientID, path string) error {
	m.clientsMutex.RLock()
	client, exists := m.clients[clientID]
	m.clientsMutex.RUnlock()

	if !exists {
		return fmt.Errorf("客户端不存在: %s", clientID)
	}

	client.subscriptionMutex.Lock()
	client.subscriptions[path] = true
	client.subscriptionMutex.Unlock()

	fmt.Printf("客户端 %s 订阅了路径: %s\n", clientID, path)
	return nil
}

// Unsubscribe 客户端取消订阅指定路径
func (m *Manager) Unsubscribe(clientID, path string) error {
	m.clientsMutex.RLock()
	client, exists := m.clients[clientID]
	m.clientsMutex.RUnlock()

	if !exists {
		return fmt.Errorf("客户端不存在: %s", clientID)
	}

	client.subscriptionMutex.Lock()
	delete(client.subscriptions, path)
	client.subscriptionMutex.Unlock()

	fmt.Printf("客户端 %s 取消订阅路径: %s\n", clientID, path)
	return nil
}

// GetSubscriptions 获取客户端的所有订阅
func (m *Manager) GetSubscriptions(clientID string) ([]string, error) {
	m.clientsMutex.RLock()
	client, exists := m.clients[clientID]
	m.clientsMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("客户端不存在: %s", clientID)
	}

	client.subscriptionMutex.RLock()
	defer client.subscriptionMutex.RUnlock()

	subs := make([]string, 0, len(client.subscriptions))
	for path := range client.subscriptions {
		subs = append(subs, path)
	}

	return subs, nil
}

// IsSubscribed 检查客户端是否订阅了指定路径
func (m *Manager) IsSubscribed(clientID, path string) bool {
	m.clientsMutex.RLock()
	client, exists := m.clients[clientID]
	m.clientsMutex.RUnlock()

	if !exists {
		return false
	}

	client.subscriptionMutex.RLock()
	defer client.subscriptionMutex.RUnlock()

	return client.subscriptions[path]
}

// BroadcastToSubscribers 向订阅了指定路径的客户端广播事件
func (m *Manager) BroadcastToSubscribers(path, eventType, message string) {
	// 验证数据有效性
	if message == "" {
		fmt.Printf("Warning: 尝试广播空消息，事件类型: %s, 路径: %s\n", eventType, path)
		return
	}

	// 验证JSON格式
	if eventType == "filesystem-change" {
		var testData interface{}
		if err := json.Unmarshal([]byte(message), &testData); err != nil {
			fmt.Printf("Error: 广播事件数据JSON格式无效: %v, 数据: %s\n", err, message)
			return
		}
	}

	event := &Event{
		ID:    uuid.New().String(),
		Event: eventType,
		Data:  message,
	}

	m.clientsMutex.RLock()
	subscribedClients := make([]*Client, 0)

	for _, client := range m.clients {
		client.subscriptionMutex.RLock()
		isSubscribed := client.subscriptions[path]
		client.subscriptionMutex.RUnlock()

		if isSubscribed {
			subscribedClients = append(subscribedClients, client)
		}
	}
	m.clientsMutex.RUnlock()

	if len(subscribedClients) == 0 {
		fmt.Printf("跳过广播: 没有客户端订阅路径 %s\n", path)
		return
	}

	fmt.Printf("向 %d 个订阅了路径 %s 的客户端广播事件\n", len(subscribedClients), path)

	// 并发发送
	for _, client := range subscribedClients {
		go func(c *Client) {
			select {
			case <-c.CloseChan:
				return
			default:
			}

			select {
			case c.EventChan <- event:
				// 发送成功
			case <-time.After(100 * time.Millisecond):
				fmt.Printf("向客户端 %s 发送事件超时\n", c.ID)
			case <-c.CloseChan:
				return
			}
		}(client)
	}
}
