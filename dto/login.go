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

type WechatWorkAccessTokenResponse struct {
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type WechatWorkSendMessageRequest struct {
	ToUser  string                `json:"touser"`
	MsgType string                `json:"msgtype"`
	AgentID int64                 `json:"agentid"`
	Text    WechatWorkTextMessage `json:"text"`
	Safe    int                   `json:"safe"`
}

type WechatWorkTextMessage struct {
	Content string `json:"content"`
}

type WechatWorkSendMessageResponse struct {
	ErrCode      int    `json:"errcode"`
	ErrMsg       string `json:"errmsg"`
	InvalidUser  string `json:"invaliduser"`
	InvalidParty string `json:"invalidparty"`
	InvalidTag   string `json:"invalidtag"`
	ResponseCode string `json:"response_code"`
}

type SliderVerifyRequest struct {
	Data62 string `form:"data62" json:"data62" binding:"required"`
	Ticket string `form:"ticket" json:"ticket" binding:"required"`
}

type LoginRequest struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}
