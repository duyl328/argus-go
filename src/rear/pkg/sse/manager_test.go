package sse

import (
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	manager := NewManager(nil)
	if manager == nil {
		t.Fatal("Manager should not be nil")
	}

	if len(manager.clients) != 0 {
		t.Errorf("Expected 0 clients, got %d", len(manager.clients))
	}

	manager.Close()
}

func TestManagerWithCustomOptions(t *testing.T) {
	opts := &Options{
		PingInterval:    10 * time.Second,
		MaxClients:      50,
		AllowedOrigins:  []string{"http://localhost:3000"},
		EnableCORS:      true,
		EventBufferSize: 10,
		ClientTimeout:   2 * time.Minute,
	}

	manager := NewManager(opts)
	if manager == nil {
		t.Fatal("Manager should not be nil")
	}

	if manager.options.MaxClients != 50 {
		t.Errorf("Expected MaxClients 50, got %d", manager.options.MaxClients)
	}

	if manager.options.PingInterval != 10*time.Second {
		t.Errorf("Expected PingInterval 10s, got %v", manager.options.PingInterval)
	}

	manager.Close()
}

func TestEventChannelBroadcast(t *testing.T) {
	manager := NewManager(nil)
	defer manager.Close()

	eventChannel := manager.GetEventChannel()

	// 创建一个测试客户端
	testClient := &Client{
		ID:        "test-client-1",
		EventChan: make(chan *Event, 10),
		CloseChan: make(chan struct{}),
		LastPing:  time.Now(),
	}

	// 手动添加客户端到管理器
	manager.clientsMutex.Lock()
	manager.clients[testClient.ID] = testClient
	manager.clientsMutex.Unlock()

	// 通过事件通道发送消息
	testMessage := "test broadcast message"

	go func() {
		eventChannel <- testMessage
	}()

	// 等待事件传播
	select {
	case event := <-testClient.EventChan:
		if event.Data != testMessage {
			t.Errorf("Expected message '%s', got '%s'", testMessage, event.Data)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for broadcast event")
	}
}

func TestBroadcastMethod(t *testing.T) {
	manager := NewManager(nil)
	defer manager.Close()

	// 创建测试客户端
	testClient := &Client{
		ID:        "test-client-2",
		EventChan: make(chan *Event, 10),
		CloseChan: make(chan struct{}),
		LastPing:  time.Now(),
	}

	// 手动添加客户端
	manager.clientsMutex.Lock()
	manager.clients[testClient.ID] = testClient
	manager.clientsMutex.Unlock()

	testMessage := "direct broadcast test"
	manager.Broadcast(testMessage)

	select {
	case event := <-testClient.EventChan:
		if event.Data != testMessage {
			t.Errorf("Expected message '%s', got '%s'", testMessage, event.Data)
		}
		if event.ID == "" {
			t.Error("Event ID should not be empty")
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for direct broadcast")
	}
}

func TestBroadcastEventMethod(t *testing.T) {
	manager := NewManager(nil)
	defer manager.Close()

	testClient := &Client{
		ID:        "test-client-3",
		EventChan: make(chan *Event, 10),
		CloseChan: make(chan struct{}),
		LastPing:  time.Now(),
	}

	manager.clientsMutex.Lock()
	manager.clients[testClient.ID] = testClient
	manager.clientsMutex.Unlock()

	testEventType := "test-event"
	testMessage := "test event message"

	manager.BroadcastEvent(testEventType, testMessage)

	select {
	case event := <-testClient.EventChan:
		if event.Data != testMessage {
			t.Errorf("Expected message '%s', got '%s'", testMessage, event.Data)
		}
		if event.Event != testEventType {
			t.Errorf("Expected event type '%s', got '%s'", testEventType, event.Event)
		}
		if event.ID == "" {
			t.Error("Event ID should not be empty")
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for event broadcast")
	}
}

func TestSendToClient(t *testing.T) {
	manager := NewManager(nil)
	defer manager.Close()

	// 创建两个测试客户端
	client1 := &Client{
		ID:        "test-client-1",
		EventChan: make(chan *Event, 10),
		CloseChan: make(chan struct{}),
		LastPing:  time.Now(),
	}

	client2 := &Client{
		ID:        "test-client-2",
		EventChan: make(chan *Event, 10),
		CloseChan: make(chan struct{}),
		LastPing:  time.Now(),
	}

	manager.clientsMutex.Lock()
	manager.clients[client1.ID] = client1
	manager.clients[client2.ID] = client2
	manager.clientsMutex.Unlock()

	testMessage := "private message to client 1"
	success := manager.SendToClient(client1.ID, testMessage)

	if !success {
		t.Error("SendToClient should return true for existing client")
	}

	// 检查client1收到消息
	select {
	case event := <-client1.EventChan:
		if event.Data != testMessage {
			t.Errorf("Client1 expected message '%s', got '%s'", testMessage, event.Data)
		}
	case <-time.After(2 * time.Second):
		t.Error("Client1 timeout waiting for private message")
	}

	// 检查client2没有收到消息
	select {
	case <-client2.EventChan:
		t.Error("Client2 should not receive private message meant for Client1")
	case <-time.After(100 * time.Millisecond):
		// 这是预期行为
	}
}

func TestSendToNonExistentClient(t *testing.T) {
	manager := NewManager(nil)
	defer manager.Close()

	success := manager.SendToClient("non-existent-client", "test message")
	if success {
		t.Error("SendToClient should return false for non-existent client")
	}
}

func TestGetClientCount(t *testing.T) {
	manager := NewManager(nil)
	defer manager.Close()

	if manager.GetClientCount() != 0 {
		t.Errorf("Expected client count 0, got %d", manager.GetClientCount())
	}

	// 添加测试客户端
	testClient := &Client{
		ID:        "test-client",
		EventChan: make(chan *Event, 10),
		CloseChan: make(chan struct{}),
		LastPing:  time.Now(),
	}

	manager.clientsMutex.Lock()
	manager.clients[testClient.ID] = testClient
	manager.clientsMutex.Unlock()

	if manager.GetClientCount() != 1 {
		t.Errorf("Expected client count 1, got %d", manager.GetClientCount())
	}
}

func TestGetClientIDs(t *testing.T) {
	manager := NewManager(nil)
	defer manager.Close()

	ids := manager.GetClientIDs()
	if len(ids) != 0 {
		t.Errorf("Expected 0 client IDs, got %d", len(ids))
	}

	// 添加测试客户端
	testClients := map[string]*Client{
		"client-1": {
			ID:        "client-1",
			EventChan: make(chan *Event, 10),
			CloseChan: make(chan struct{}),
			LastPing:  time.Now(),
		},
		"client-2": {
			ID:        "client-2",
			EventChan: make(chan *Event, 10),
			CloseChan: make(chan struct{}),
			LastPing:  time.Now(),
		},
	}

	manager.clientsMutex.Lock()
	for id, client := range testClients {
		manager.clients[id] = client
	}
	manager.clientsMutex.Unlock()

	ids = manager.GetClientIDs()
	if len(ids) != 2 {
		t.Errorf("Expected 2 client IDs, got %d", len(ids))
	}

	// 检查是否包含正确的ID
	idMap := make(map[string]bool)
	for _, id := range ids {
		idMap[id] = true
	}

	if !idMap["client-1"] || !idMap["client-2"] {
		t.Error("Client IDs not returned correctly")
	}
}

func TestGetClientInfo(t *testing.T) {
	manager := NewManager(nil)
	defer manager.Close()

	// 测试不存在的客户端
	_, exists := manager.GetClientInfo("non-existent")
	if exists {
		t.Error("GetClientInfo should return false for non-existent client")
	}

	// 添加测试客户端
	originalClient := &Client{
		ID:          "test-client",
		EventChan:   make(chan *Event, 10),
		CloseChan:   make(chan struct{}),
		LastPing:    time.Now(),
		RemoteAddr:  "127.0.0.1",
		UserAgent:   "test-agent",
		ConnectedAt: time.Now(),
	}

	manager.clientsMutex.Lock()
	manager.clients[originalClient.ID] = originalClient
	manager.clientsMutex.Unlock()

	clientInfo, exists := manager.GetClientInfo("test-client")
	if !exists {
		t.Error("GetClientInfo should return true for existing client")
	}

	if clientInfo.ID != originalClient.ID {
		t.Errorf("Expected client ID '%s', got '%s'", originalClient.ID, clientInfo.ID)
	}

	if clientInfo.RemoteAddr != originalClient.RemoteAddr {
		t.Errorf("Expected remote addr '%s', got '%s'", originalClient.RemoteAddr, clientInfo.RemoteAddr)
	}
}

func TestManagerClose(t *testing.T) {
	manager := NewManager(nil)

	// 添加一个测试客户端
	testClient := &Client{
		ID:        "test-client",
		EventChan: make(chan *Event, 10),
		CloseChan: make(chan struct{}),
		LastPing:  time.Now(),
	}

	manager.clientsMutex.Lock()
	manager.clients[testClient.ID] = testClient
	manager.clientsMutex.Unlock()

	// 关闭管理器
	manager.Close()

	// 验证客户端被清理
	if len(manager.clients) != 0 {
		t.Errorf("Expected 0 clients after close, got %d", len(manager.clients))
	}

	// 验证context被取消
	select {
	case <-manager.ctx.Done():
		// 这是预期行为
	case <-time.After(100 * time.Millisecond):
		t.Error("Manager context should be cancelled after Close()")
	}
}

func TestMaxClientsLimit(t *testing.T) {
	opts := &Options{
		MaxClients:      2,
		EventBufferSize: 10,
		PingInterval:    30 * time.Second,
	}

	manager := NewManager(opts)
	defer manager.Close()

	// 创建3个测试客户端，但只有2个应该被接受
	clients := []*Client{
		{
			ID:        "client-1",
			EventChan: make(chan *Event, 10),
			CloseChan: make(chan struct{}),
			LastPing:  time.Now(),
		},
		{
			ID:        "client-2",
			EventChan: make(chan *Event, 10),
			CloseChan: make(chan struct{}),
			LastPing:  time.Now(),
		},
		{
			ID:        "client-3",
			EventChan: make(chan *Event, 10),
			CloseChan: make(chan struct{}),
			LastPing:  time.Now(),
		},
	}

	// 添加客户端
	for _, client := range clients {
		manager.addClient(client)
	}

	// 验证只有2个客户端被接受
	if manager.GetClientCount() != 2 {
		t.Errorf("Expected max 2 clients, got %d", manager.GetClientCount())
	}

	// 验证第三个客户端的CloseChan被关闭
	select {
	case <-clients[2].CloseChan:
		// 这是预期行为，第三个客户端应该被拒绝
	case <-time.After(100 * time.Millisecond):
		t.Error("Third client should be rejected and CloseChan should be closed")
	}
}

func BenchmarkBroadcast(b *testing.B) {
	manager := NewManager(&Options{
		EventBufferSize: 1000,
		MaxClients:      100,
		PingInterval:    30 * time.Second,
		AllowedOrigins:  []string{"*"},
		EnableCORS:      true,
		ClientTimeout:   5 * time.Minute,
	})
	defer manager.Close()

	// 添加一些测试客户端
	for i := 0; i < 10; i++ {
		client := &Client{
			ID:        string(rune('A' + i)),
			EventChan: make(chan *Event, 100),
			CloseChan: make(chan struct{}),
			LastPing:  time.Now(),
		}
		manager.clientsMutex.Lock()
		manager.clients[client.ID] = client
		manager.clientsMutex.Unlock()

		// 启动goroutine来消费事件，避免channel阻塞
		go func(c *Client) {
			for {
				select {
				case <-c.EventChan:
					// 消费事件
				case <-c.CloseChan:
					return
				}
			}
		}(client)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager.Broadcast("benchmark test message")
	}
}
