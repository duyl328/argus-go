package container

import (
	"rear/internal/workflow"
)

type TaskContainer struct {
	// 照片任务处理管理
	ImgTaskManager *workflow.ImgTaskManager
	// 其他服务...
}

func NewTaskContainer() *TaskContainer {
	return &TaskContainer{
		ImgTaskManager: workflow.NewImgTaskManager(10),
	}
}
