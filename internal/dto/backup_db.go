package dto

type BackupDbRequest struct {
	Name   string `json:"name" binding:"required,max=63,id"`
	Reason string `json:"reason" binding:"required,max=255"`
}
