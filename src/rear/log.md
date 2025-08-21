2025-08-16 17:50:46.628	DEBUG	api/img_api.go:505	生成缩略图完成	{"original": "E:\\整合\\kkimkkimmy\\Pic\\116055250_The light is nice. ️_01_8cc4a8a3-0f83-4c22-bf8c-c5b9d1de3419.jpg", "thumbnail": "C:\\Users\\charlatans\\AppData\\Local\\JetBrains\\GoLand2025.1\\tmp\\GoLand\\cache\\thumbnail\\fa\\35\\f2\\fa35f27766e9cbeb84a557e0b74e9680f252c5000f33b248d38f6c799dbf79c7\\720X720.jpg", "size": 720}
2025-08-16 17:50:46.628	DEBUG	workflow/img_task.go:233	任务步骤更新	{"task_id": "b2a5347d-fabb-49de-ad2c-e0b625a2486e", "step_name": "生成缩略图", "overall_progress": 0.85}
2025-08-16 17:50:46.628	DEBUG	workflow/img_task.go:233	任务步骤更新	{"task_id": "b2a5347d-fabb-49de-ad2c-e0b625a2486e", "step_name": "保存数据", "overall_progress": 0.9}
2025-08-16 17:50:46.628	INFO	workflow/img_task.go:367	保存图像信息到数据库	{"hash": "fa35f27766e9cbeb84a557e0b74e9680f252c5000f33b248d38f6c799dbf79c7", "path": "E:\\整合\\kkimkkimmy\\Pic\\116055250_The light is nice. ️_01_8cc4a8a3-0f83-4c22-bf8c-c5b9d1de3419.jpg"}
2025-08-16 17:50:46.628	WARN	workflow/img_task.go:393	没有EXIF数据可保存	{"hash": "fa35f27766e9cbeb84a557e0b74e9680f252c5000f33b248d38f6c799dbf79c7"}
panic: runtime error: invalid memory address or nil pointer dereference
[signal 0xc0000005 code=0x0 addr=0x18 pc=0x7ff6ab5ed44a]

goroutine 269 [running]:
rear/internal/repositories.(*PhotoRepository).CreatePhotoFromImageAPI(0x7ff6aba65a4a?, {0xc0001297c0, 0x40}, {0xc0003ff2d0, 0x65}, {0x7ff6ab9591a0, 0x0})
D:/go-argus/src/rear/internal/repositories/photo_repository.go:200 +0x26a
rear/internal/workflow.(*PictureTask).saveToDatabase(0xc0004a4600, 0x7ff6aba5eddf?)
D:/go-argus/src/rear/internal/workflow/img_task.go:398 +0xa85
rear/internal/workflow.(*PictureTask).Run(0xc0004a4600)
D:/go-argus/src/rear/internal/workflow/img_task.go:318 +0x26e
rear/internal/workflow.(*ImgTaskManager).run.func1(0xc0004a4600)
D:/go-argus/src/rear/internal/workflow/img_task.go:619 +0x51
created by rear/internal/workflow.(*ImgTaskManager).run in goroutine 14
D:/go-argus/src/rear/internal/workflow/img_task.go:617 +0x38
panic: runtime error: invalid memory address or nil pointer dereference
[signal 0xc0000005 code=0x0 addr=0x18 pc=0x7ff6ab5ed44a]

goroutine 265 [running]:
rear/internal/repositories.(*PhotoRepository).CreatePhotoFromImageAPI(0x7ff6aba65a4a?, {0xc003d95340, 0x40}, {0xc0003ff110, 0x65}, {0x7ff6ab9591a0, 0x0})
D:/go-argus/src/rear/internal/repositories/photo_repository.go:200 +0x26a
rear/internal/workflow.(*PictureTask).saveToDatabase(0xc0004a4480, 0x7ff6aba5eddf?)
D:/go-argus/src/rear/internal/workflow/img_task.go:398 +0xa85
rear/internal/workflow.(*PictureTask).Run(0xc0004a4480)
D:/go-argus/src/rear/internal/workflow/img_task.go:318 +0x26e
rear/internal/workflow.(*ImgTaskManager).run.func1(0xc0004a4480)
D:/go-argus/src/rear/internal/workflow/img_task.go:619 +0x51
created by rear/internal/workflow.(*ImgTaskManager).run in goroutine 14
D:/go-argus/src/rear/internal/workflow/img_task.go:617 +0x38
