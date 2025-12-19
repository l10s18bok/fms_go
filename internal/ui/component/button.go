// Package component는 재사용 가능한 UI 컴포넌트를 제공합니다.
package component

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// 버튼 스타일 타입입니다.
type ButtonStyle int

const (
	ButtonPrimary   ButtonStyle = iota // 파란색 (주요 액션)
	ButtonSuccess                      // 초록색 (성공/저장)
	ButtonDanger                       // 빨간색 (삭제/위험)
	ButtonSecondary                    // 회색 (보조)
	ButtonDark                         // 진한 회색
	ButtonBlack                        // 검정
)

// 스타일이 적용된 액션 버튼을 생성합니다.
func NewActionButton(label string, style ButtonStyle, onTap func()) *widget.Button {
	btn := widget.NewButton(label, onTap)

	switch style {
	case ButtonPrimary:
		btn.Importance = widget.HighImportance
	case ButtonDanger:
		btn.Importance = widget.DangerImportance
	case ButtonSuccess:
		btn.Importance = widget.SuccessImportance
	case ButtonDark, ButtonBlack:
		btn.Importance = widget.LowImportance
	default:
		btn.Importance = widget.MediumImportance
	}

	return btn
}

// 버튼들을 가로로 배치합니다.
func NewButtonGroup(buttons ...*widget.Button) *fyne.Container {
	objects := make([]fyne.CanvasObject, len(buttons))
	for i, btn := range buttons {
		objects[i] = btn
	}
	return container.NewHBox(objects...)
}

// 아이콘 버튼을 생성합니다.
func NewIconButton(icon fyne.Resource, onTap func()) *widget.Button {
	return widget.NewButtonWithIcon("", icon, onTap)
}

// 아이콘, 텍스트, 스타일이 적용된 버튼을 생성합니다.
func NewIconTextButton(label string, icon fyne.Resource, style ButtonStyle, onTap func()) *widget.Button {
	btn := widget.NewButtonWithIcon(label, icon, onTap)

	switch style {
	case ButtonPrimary:
		btn.Importance = widget.HighImportance
	case ButtonDanger:
		btn.Importance = widget.DangerImportance
	case ButtonSuccess:
		btn.Importance = widget.SuccessImportance
	case ButtonDark, ButtonBlack:
		btn.Importance = widget.LowImportance
	default:
		btn.Importance = widget.MediumImportance
	}

	return btn
}

// 탭 이벤트를 처리하는 간단한 컨테이너입니다.
type TappableContainer struct {
	widget.BaseWidget
	content fyne.CanvasObject
	onTap   func()
}

// 탭 가능한 컨테이너를 생성합니다.
func NewTappableContainer(content fyne.CanvasObject, onTap func()) *TappableContainer {
	t := &TappableContainer{
		content: content,
		onTap:   onTap,
	}
	t.ExtendBaseWidget(t)
	return t
}

// 위젯의 렌더러를 생성합니다.
func (t *TappableContainer) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.content)
}

// 탭 이벤트를 처리합니다.
func (t *TappableContainer) Tapped(*fyne.PointEvent) {
	if t.onTap != nil {
		t.onTap()
	}
}

// 보조 탭 이벤트를 처리합니다.
func (t *TappableContainer) TappedSecondary(*fyne.PointEvent) {}

// 투명 배경에 컬러 텍스트 버튼을 생성합니다.
func NewTextButton(text string, textColor color.Color, onTap func()) fyne.CanvasObject {
	label := canvas.NewText(text, textColor)
	label.TextStyle = fyne.TextStyle{Bold: true}

	// 패딩 적용
	paddedLabel := container.New(layout.NewCustomPaddedLayout(6, 6, 12, 12), container.NewCenter(label))

	return NewTappableContainer(paddedLabel, onTap)
}

// 투명 배경에 빨간 텍스트 버튼을 생성합니다.
func NewRedTextButton(text string, onTap func()) fyne.CanvasObject {
	return NewTextButton(text, color.RGBA{R: 220, G: 53, B: 69, A: 255}, onTap)
}

// 회색 배경의 버튼을 생성합니다.
func NewColoredButton(text string, style ButtonStyle, onTap func()) fyne.CanvasObject {
	// 배경색 결정
	var bgColor color.Color
	switch style {
	case ButtonDark:
		bgColor = color.RGBA{R: 80, G: 80, B: 80, A: 255}
	case ButtonBlack:
		bgColor = color.RGBA{R: 40, G: 40, B: 40, A: 255}
	default:
		bgColor = color.RGBA{R: 100, G: 100, B: 100, A: 255}
	}

	// 배경
	bg := canvas.NewRectangle(bgColor)
	bg.CornerRadius = 4

	// 텍스트
	label := canvas.NewText(text, color.White)
	label.TextStyle = fyne.TextStyle{Bold: true}

	// 패딩 적용 (상, 하, 좌, 우)
	paddedLabel := container.New(layout.NewCustomPaddedLayout(8, 8, 16, 16), container.NewCenter(label))

	// 배경 + 텍스트를 TappableContainer로 감싸기
	content := container.NewStack(bg, paddedLabel)
	return NewTappableContainer(content, onTap)
}
