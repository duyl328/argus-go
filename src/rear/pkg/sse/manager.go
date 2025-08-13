package sse

import (
	"context"
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
	clients := make([]*Client, 0, len(m.clients))
	for _, client := range m.clients {
		clients = append(clients, client)
	}
	m.clientsMutex.RUnlock()

	for _, client := range clients {
		select {
		case client.EventChan <- event:
		case <-time.After(100 * time.Millisecond):
			go m.UnregisterClient(client.ID)
		}
	}
}

func (m *Manager) pingAllClients() {
	pingEvent := &Event{
		Event: "ping",
		Data:  fmt.Sprintf(`{"time":"%s"}`, time.Now().Format(time.RFC3339)),
	}

	select {
	case m.broadcastChan <- pingEvent:
	case <-time.After(1 * time.Second):
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

	clientID := uuid.New().String()
	client := &Client{
		ID:          clientID,
		Writer:      c.Writer,
		Request:     c.Request,
		Context:     c,
		EventChan:   make(chan *Event, 50),
		CloseChan:   make(chan struct{}),
		LastPing:    time.Now(),
		UserAgent:   c.GetHeader("User-Agent"),
		RemoteAddr:  c.ClientIP(),
		ConnectedAt: time.Now(),
	}

	m.registerChan <- client

	c.Stream(func(w io.Writer) bool {
		select {
		case <-client.CloseChan:
			return false
		case event := <-client.EventChan:
			m.writeEvent(w, event)
			client.LastPing = time.Now()
			return true
		case <-c.Request.Context().Done():
			go m.UnregisterClient(clientID)
			return false
		}
	})

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

func (m *Manager) writeEvent(w io.Writer, event *Event) {
	if event.ID != "" {
		fmt.Fprintf(w, "id: %s\n", event.ID)
	}
	if event.Event != "" {
		fmt.Fprintf(w, "event: %s\n", event.Event)
	}
	if event.Retry > 0 {
		fmt.Fprintf(w, "retry: %d\n", event.Retry)
	}
	fmt.Fprintf(w, "data: %s\n\n", event.Data)

	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (m *Manager) UnregisterClient(clientID string) {
	select {
	case m.unregisterChan <- clientID:
	case <-time.After(1 * time.Second):
	}
}

func (m *Manager) Broadcast(message string) {
	event := &Event{
		ID:   uuid.New().String(),
		Data: message,
	}

	select {
	case m.broadcastChan <- event:
	case <-time.After(1 * time.Second):
	}
}

func (m *Manager) BroadcastEvent(eventType, message string) {
	event := &Event{
		ID:    uuid.New().String(),
		Event: eventType,
		Data:  message,
	}

	select {
	case m.broadcastChan <- event:
	case <-time.After(1 * time.Second):
	}
}

func (m *Manager) SendToClient(clientID, message string) bool {
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
	m.cancel()

	m.clientsMutex.Lock()
	for _, client := range m.clients {
		close(client.CloseChan)
	}
	m.clients = make(map[string]*Client)
	m.clientsMutex.Unlock()

	close(m.broadcastChan)
	close(m.registerChan)
	close(m.unregisterChan)
	close(m.eventChan)
}
