package camera

// 定义关键字
/*----------------以下是camera------------------------*/
// 命令
const (
	PTZControl          = "PTZControl"          // 云台控制
	Angle               = "Angle"               // 转动角度
	Snapshot            = "Snapshot"            // 快照
	SnapshotURL         = "SnapshotURL"         // 快照路径
	Zoom                = "Zoom"                // 放大缩小
	SetPreset           = "SetPreset"           // 设置预置位置
	GetPresets          = "GetPresets"          // 获取所有预置位置
	GotoPreset          = "GotoPreset"          // 转到预置位置
	RemovePreset        = "RemovePreset"        // 移除预置位置
	SetHomePosition     = "SetHomePosition"     // 设置Home位
	GotoHomePosition    = "GotoHomePosition"    // 转到Home位
	TimeCalibration     = "TimeCalibration"     // 时间校准
	TimeCalibrationData = "TimeCalibrationData" // 时间校准值
	CameraStatus = "CameraStatus" // 设备状态
	/*----------------结束------------------------*/

	// 命令回执
	SetLocalTimeInfo = "SetLocalTimeInfo" // 设置本地时间
)
