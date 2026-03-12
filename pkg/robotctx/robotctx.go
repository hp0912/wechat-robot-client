package robotctx

import "fmt"

// RobotContext 在工具调用入参中透传的机器人上下文
type RobotContext struct {
	WeChatClientPort string
	RobotID          int64
	RobotCode        string
	RobotRedisDB     uint
	RobotWxID        string
	FromWxID         string
	SenderWxID       string
	MessageID        int64
	RefMessageID     int64
}

// ToEnvVars 将 RobotContext 转换为环境变量键值对
func (rc RobotContext) ToEnvVars() []string {
	return []string{
		"ROBOT_WECHAT_CLIENT_PORT=" + rc.WeChatClientPort,
		fmt.Sprintf("ROBOT_ID=%d", rc.RobotID),
		"ROBOT_CODE=" + rc.RobotCode,
		fmt.Sprintf("ROBOT_REDIS_DB=%d", rc.RobotRedisDB),
		"ROBOT_WX_ID=" + rc.RobotWxID,
		"ROBOT_FROM_WX_ID=" + rc.FromWxID,
		"ROBOT_SENDER_WX_ID=" + rc.SenderWxID,
		fmt.Sprintf("ROBOT_MESSAGE_ID=%d", rc.MessageID),
		fmt.Sprintf("ROBOT_REF_MESSAGE_ID=%d", rc.RefMessageID),
	}
}
