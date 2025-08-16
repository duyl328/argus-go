package tables

// PhotoAiMetadataTable 表结构 - 保存图片AI分析的元信息
type PhotoAiMetadataTable struct {
	BaseModel
	Hash      string  `gorm:"uniqueIndex;not null" json:"hash"`
	NsfwScore float32 `json:"nsfwScore"`
	// 人物
	//Faces JSON,        -- [ {x,y,w,h,confidence}, ... ]
	// 物体
	//Objects JSON,      -- [ {label: "dog", confidence:0.98}, ... ]
	// 场景
	//Scene JSON,        -- {label:"beach", confidence:0.87}
}
