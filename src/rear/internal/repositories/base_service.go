package repositories

import (
	"rear/internal/db"
	"sync"
)

// 写操作类型
type WriteOperation struct {
	Execute  func() error
	Response chan error
}

// BaseService 处理数据库写操作的单线程执行
type BaseService struct {
	writeQueue chan WriteOperation
	once       sync.Once
}

var baseService = &BaseService{
	// 缓冲队列
	writeQueue: make(chan WriteOperation, 1000),
}

// 初始化并启动写操作处理协程
func InitBaseService() {
	baseService.once.Do(func() {
		go baseService.processWriteOperations()
	})
}

// 处理写操作的协程
func (s *BaseService) processWriteOperations() {
	for op := range s.writeQueue {
		err := op.Execute() // 串行执行每个写操作
		op.Response <- err  // 将结果返回给调用者
		close(op.Response)
	}
}

// ExecuteWrite 执行写操作（增删改）
func ExecuteWrite(fn func() error) error {
	// 如果不是 SQLite，直接执行
	if !db.IsSQLite() {
		return fn()
	}

	// SQLite 走单线程队列
	op := WriteOperation{
		Execute:  fn,
		Response: make(chan error, 1),
	}

	baseService.writeQueue <- op // 将操作放入队列
	return <-op.Response         // 等待执行结果
}

// ExecuteRead 执行读操作（可以并发）
func ExecuteRead(fn func() error) error {
	return fn()
}
