package component

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// 토스트 타입입니다.
type ToastType int

const (
	ToastSuccess ToastType = iota
	ToastError
	ToastInfo
)

// 토스트 색상
var toastColors = map[ToastType]color.Color{
	ToastSuccess: color.RGBA{R: 40, G: 167, B: 69, A: 230},
	ToastError:   color.RGBA{R: 220, G: 53, B: 69, A: 230},
	ToastInfo:    color.RGBA{R: 0, G: 123, B: 255, A: 230},
}

// 토스트 메시지를 표시합니다 (일정 시간 후 사라짐).
func ShowToast(parent fyne.Window, message string, toastType ToastType, duration time.Duration) {
	// 테두리만 있는 배경 (내부는 투명)
	borderColor := toastColors[toastType]
	bg := canvas.NewRectangle(color.Transparent)
	bg.SetMinSize(fyne.NewSize(200, 50))
	bg.StrokeColor = borderColor
	bg.StrokeWidth = 2

	label := widget.NewLabel(message)
	label.Alignment = fyne.TextAlignCenter

	// 패딩을 추가하여 텍스트가 잘리지 않도록 함
	paddedLabel := container.NewPadded(label)
	toast := container.NewStack(bg, container.NewCenter(paddedLabel))

	// 팝업으로 표시
	popup := widget.NewModalPopUp(toast, parent.Canvas())
	popup.Show()

	// 일정 시간 후 숨김
	go func() {
		time.Sleep(duration)
		fyne.Do(func() {
			popup.Hide()
		})
	}()
}

// 성공 토스트를 표시합니다.
func ShowSuccessToast(parent fyne.Window, message string) {
	ShowToast(parent, message, ToastSuccess, 2*time.Second)
}

// 에러 토스트를 표시합니다.
func ShowErrorToast(parent fyne.Window, message string) {
	ShowToast(parent, message, ToastError, 3*time.Second)
}

// 정보 토스트를 표시합니다.
func ShowInfoToast(parent fyne.Window, message string) {
	ShowToast(parent, message, ToastInfo, 2*time.Second)
}
