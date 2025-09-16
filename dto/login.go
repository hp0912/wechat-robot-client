package dto

type TFARequest struct {
	Uuid   string `form:"uuid" json:"uuid" binding:"required"`
	Code   string `form:"code" json:"code" binding:"required"`
	Ticket string `form:"ticket" json:"ticket" binding:"required"`
	Data62 string `form:"data62" json:"data62" binding:"required"`
}

type LogoutNotificationRequest struct {
	WxID       string `form:"wxid" json:"wxid"`
	Type       string `form:"type" json:"type"`
	Status     string `form:"status" json:"status"`
	RetryCount int    `form:"retry_count" json:"retry_count"`
}

type PushPlusNotificationRequest struct {
	Token       string `json:"token"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	Template    string `json:"template"`
	Channel     string `json:"channel"`
	Webhook     string `json:"webhook"`
	CallbackUrl string `json:"callbackUrl"`
	Timestamp   string `json:"timestamp"`
	Pre         string `json:"pre"`
}

type PushPlusNotificationResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"`
}

type SliderVerifyRequest struct {
	Data62 string `form:"data62" json:"data62" binding:"required"`
	Ticket string `form:"ticket" json:"ticket" binding:"required"`
}

type LoginRequest struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}
