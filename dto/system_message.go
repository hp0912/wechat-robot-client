package dto

type MarkAsReadBatchRequest struct {
	IDs []int64 `form:"ids" json:"ids" binding:"required"`
}
