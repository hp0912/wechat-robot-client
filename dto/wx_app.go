package dto

type WxappQrcodeAuthLoginRequest struct {
	URL string `form:"url" json:"url" binding:"required"`
}
