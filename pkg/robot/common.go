package robot

type BuiltinString struct {
	String string `json:"string"`
}

type BuiltinBuffer struct {
	Buffer []byte `json:"buffer,omitempty"`
	ILen   int64  `json:"iLen,omitempty"`
}

// RobotStatus 表示机器人状态的枚举类型
type RobotStatus string

const (
	RobotStatusOnline  RobotStatus = "online"
	RobotStatusOffline RobotStatus = "offline"
	RobotStatusError   RobotStatus = "error"
)
