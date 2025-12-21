package storage

import "fms_wails/internal/model"

// ExportData는 내보내기/가져오기용 데이터 구조입니다.
type ExportData struct {
	Templates []*model.Template      `json:"templates"`
	Firewalls []*model.Firewall      `json:"firewalls"`
	History   []*model.DeployHistory `json:"history"`
}
