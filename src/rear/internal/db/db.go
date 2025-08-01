package db

import (
	"context"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm/logger"
	"rear/internal/config"
	"sync"
	"time"

	"gorm.io/gorm"
)

// WriteTask 写任务接口
type WriteTask interface {
	Execute(db *gorm.DB) error
	GetPriority() int // 优先级，用于任务排序
	GetID() string    // 任务ID，用于去重或追踪
}

// WriteTaskResult 写任务结果
type WriteTaskResult struct {
	TaskID string
	Error  error
}

// TaskCallback 任务回调
type TaskCallback struct {
	OnSuccess func(taskID string)
	OnError   func(taskID string, err error)
}

// DatabaseManager 数据库管理器
type DatabaseManager struct {
	db     *gorm.DB
	dbType config.DatabaseType

	// 写任务队列
	writeQueue  chan WriteTaskWithCallback
	resultQueue chan WriteTaskResult
	stopCh      chan struct{}
	wg          sync.WaitGroup

	// 读写分离配置
	maxWriteWorkers int
	maxRetries      int
}

// WriteTaskWithCallback 带回调的写任务
type WriteTaskWithCallback struct {
	Task     WriteTask
	Callback TaskCallback
}

// NewDatabaseManager 创建数据库管理器
func NewDatabaseManager(db *gorm.DB, dbType config.DatabaseType) *DatabaseManager {
	maxWorkers := 1 // SQLite默认单线程
	if dbType == config.MySQL {
		maxWorkers = 10 // MySQL可以多线程
	}

	dm := &DatabaseManager{
		db:              db,
		dbType:          dbType,
		writeQueue:      make(chan WriteTaskWithCallback, 1000), // 缓冲队列
		resultQueue:     make(chan WriteTaskResult, 1000),
		stopCh:          make(chan struct{}),
		maxWriteWorkers: maxWorkers,
		maxRetries:      3,
	}

	dm.start()
	return dm
}

// start 启动数据库管理器
func (dm *DatabaseManager) start() {
	// 启动写任务处理器
	for i := 0; i < dm.maxWriteWorkers; i++ {
		dm.wg.Add(1)
		go dm.writeWorker(i)
	}

	// 启动结果处理器
	dm.wg.Add(1)
	go dm.resultProcessor()
}

// writeWorker 写任务工作协程
func (dm *DatabaseManager) writeWorker(workerID int) {
	defer dm.wg.Done()

	for {
		select {
		case taskWithCallback := <-dm.writeQueue:
			dm.executeWriteTask(taskWithCallback)
		case <-dm.stopCh:
			return
		}
	}
}

// executeWriteTask 执行写任务
func (dm *DatabaseManager) executeWriteTask(taskWithCallback WriteTaskWithCallback) {
	task := taskWithCallback.Task
	callback := taskWithCallback.Callback

	var err error
	for retry := 0; retry < dm.maxRetries; retry++ {
		// 对于SQLite，可以使用事务来优化性能
		if dm.dbType == config.SQLite {
			err = dm.db.Transaction(func(tx *gorm.DB) error {
				return task.Execute(tx)
			})
		} else {
			err = task.Execute(dm.db)
		}

		if err == nil {
			break
		}

		// 重试延迟
		if retry < dm.maxRetries-1 {
			time.Sleep(time.Millisecond * time.Duration(100*(retry+1)))
		}
	}

	// 发送结果
	result := WriteTaskResult{
		TaskID: task.GetID(),
		Error:  err,
	}

	select {
	case dm.resultQueue <- result:
	default:
		// 结果队列满了，直接执行回调
		if err != nil && callback.OnError != nil {
			callback.OnError(task.GetID(), err)
		} else if err == nil && callback.OnSuccess != nil {
			callback.OnSuccess(task.GetID())
		}
	}
}

// resultProcessor 结果处理器
func (dm *DatabaseManager) resultProcessor() {
	defer dm.wg.Done()

	for {
		select {
		case result := <-dm.resultQueue:
			// 这里可以进行结果处理，比如记录日志、监控等
			fmt.Printf("Task %s completed with error: %v\n", result.TaskID, result.Error)
		case <-dm.stopCh:
			return
		}
	}
}

// SubmitWriteTask 提交写任务
func (dm *DatabaseManager) SubmitWriteTask(task WriteTask, callback TaskCallback) error {
	taskWithCallback := WriteTaskWithCallback{
		Task:     task,
		Callback: callback,
	}

	select {
	case dm.writeQueue <- taskWithCallback:
		return nil
	default:
		return fmt.Errorf("write queue is full")
	}
}

// SubmitWriteTaskSync 同步提交写任务
func (dm *DatabaseManager) SubmitWriteTaskSync(ctx context.Context, task WriteTask) error {
	done := make(chan error, 1)

	callback := TaskCallback{
		OnSuccess: func(taskID string) {
			done <- nil
		},
		OnError: func(taskID string, err error) {
			done <- err
		},
	}

	if err := dm.SubmitWriteTask(task, callback); err != nil {
		return err
	}

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ReadDirect 直接读取（不经过队列）
func (dm *DatabaseManager) ReadDirect() *gorm.DB {
	return dm.db
}

// GetDB 获取数据库连接（仅用于读操作）
func (dm *DatabaseManager) GetDB() *gorm.DB {
	return dm.db
}

// Stop 停止数据库管理器
func (dm *DatabaseManager) Stop() {
	close(dm.stopCh)
	dm.wg.Wait()
}

func InitDatabase() error {
	databaseConfig := getDatabaseConfig()

	var db *gorm.DB
	var err error

	switch databaseConfig.Type {
	case config.MySQL:
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			databaseConfig.Username, databaseConfig.Password, databaseConfig.Host, databaseConfig.Port, databaseConfig.Database)
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	case config.SQLite:
		// SQLite特殊配置
		db, err = gorm.Open(sqlite.Open(databaseConfig.Database), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err != nil {
			return err
		}

		// SQLite优化设置
		sqlDB, _ := db.DB()
		sqlDB.SetMaxIdleConns(1)    // SQLite单连接
		sqlDB.SetMaxOpenConns(1)    // SQLite单连接
		sqlDB.SetConnMaxLifetime(0) // 永不关闭

		// SQLite WAL模式优化
		db.Exec("PRAGMA journal_mode=WAL;")
		db.Exec("PRAGMA synchronous=NORMAL;")
		db.Exec("PRAGMA cache_size=1000000;")
		db.Exec("PRAGMA foreign_keys=true;")
		db.Exec("PRAGMA temp_store=memory;")

	default:
		return fmt.Errorf("unsupported database type: %s", databaseConfig.Type)
	}

	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}
	DB = db
	Manger = NewDatabaseManager(db, databaseConfig.Type)
	return nil
}

func getDatabaseConfig() config.DatabaseConfig {
	// 初始化数据库
	return config.DatabaseConfig{
		Type:         config.SQLite,
		Database:     "test.db",
		MaxIdleConns: 1,
		MaxOpenConns: 1,
		MaxLifetime:  0,
	}
}

var DB *gorm.DB
var Manger *DatabaseManager

func GetDB() *gorm.DB {
	return DB
}
func GetManger() *DatabaseManager {
	return Manger
}
func IsSQLite() bool {
	cfg := getDatabaseConfig()
	return cfg.Type == config.SQLite
}
