package component

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// 상태 타입입니다.
type StatusType int

const (
	StatusSuccess StatusType = iota // 성공/정상 (초록)
	StatusError                     // 에러/실패 (빨강)
	StatusWarning                   // 경고 (노랑)
	StatusInfo                      // 정보 (파랑)
	StatusPending                   // 대기 (회색)
)

// 상태별 색상
var statusColors = map[StatusType]color.Color{
	StatusSuccess: color.RGBA{R: 40, G: 167, B: 69, A: 255},   // 초록
	StatusError:   color.RGBA{R: 220, G: 53, B: 69, A: 255},   // 빨강
	StatusWarning: color.RGBA{R: 255, G: 193, B: 7, A: 255},   // 노랑
	StatusInfo:    color.RGBA{R: 0, G: 123, B: 255, A: 255},   // 파랑
	StatusPending: color.RGBA{R: 108, G: 117, B: 125, A: 255}, // 회색
}

// 상태 텍스트와 색상 뱃지를 생성합니다.
func NewStatusBadge(text string, status StatusType) *fyne.Container {
	circle := canvas.NewCircle(statusColors[status])
	circle.Resize(fyne.NewSize(10, 10))

	label := widget.NewLabel(text)

	return container.NewHBox(circle, label)
}

// 상태 아이콘만 표시합니다.
func NewStatusIcon(status StatusType) *canvas.Circle {
	circle := canvas.NewCircle(statusColors[status])
	circle.Resize(fyne.NewSize(12, 12))
	return circle
}

// 상태 코드를 StatusType으로 변환합니다.
func GetStatusType(code string) StatusType {
	switch code {
	case "running", "success", "ok":
		return StatusSuccess
	case "stop", "fail", "error":
		return StatusError
	case "warning":
		return StatusWarning
	default:
		return StatusPending
	}
}

// 상태 코드를 표시 텍스트로 변환합니다.
func GetStatusText(code string) string {
	textMap := map[string]string{
		"running": "정상",
		"stop":    "정지",
		"success": "성공",
		"error":   "확인요망",
		"fail":    "실패",
	}
	if text, ok := textMap[code]; ok {
		return text
	}
	return "-"
}
