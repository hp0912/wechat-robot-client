package dto

type TFARequest struct {
	Uuid   string `form:"uuid" json:"uuid" binding:"required"`
	Code   string `form:"code" json:"code" binding:"required"`
	Ticket string `form:"ticket" json:"ticket" binding:"required"`
	Data62 string `form:"data62" json:"data62" binding:"required"`
}
