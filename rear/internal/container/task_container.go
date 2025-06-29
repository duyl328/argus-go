package container

import (
	"rear/internal/workflow"
)

type TaskContainer struct {
	// 照片任务处理管理
	ImgTaskManager *workflow.ImgTaskManager
	// 其他服务...

	// 数据库服务
	DbContainer *DbContainer
}

func NewTaskContainer(con *DbContainer) *TaskContainer {
	return &TaskContainer{
		DbContainer:    con,
		ImgTaskManager: workflow.NewImgTaskManager(5),
	}
}
