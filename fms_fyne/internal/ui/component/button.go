// Package component는 재사용 가능한 UI 컴포넌트를 제공합니다.
package component

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
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

// 아이콘, 텍스트, 스타일이 적용된 버튼을 패딩과 함께 생성합니다.
// top, bottom, left, right 순서로 패딩을 지정합니다. (기본값: 0)
func NewIconTextButtonWithPadding(label string, icon fyne.Resource, style ButtonStyle, onTap func(), top, bottom, left, right float32) fyne.CanvasObject {
	btn := NewIconTextButton(label, icon, style, onTap)
	return container.New(layout.NewCustomPaddedLayout(top, bottom, left, right), btn)
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

// 전천후 커스텀 버튼을 생성합니다.
// - icon: nil이면 텍스트만 표시
// - iconColor: 아이콘/텍스트 색상 (nil이면 자동 결정: 밝은 배경→검정, 어두운 배경→흰색)
// - bgColor: nil이면 투명 배경 (텍스트 버튼)
// - margin: 외부 여백 (상, 하, 좌, 우 순서, 생략 시 0)
func NewCustomButton(label string, icon fyne.Resource, iconColor, bgColor color.Color, onTap func(), margin ...float32) fyne.CanvasObject {
	// 외부 여백 기본값
	var mTop, mBottom, mLeft, mRight float32 = 0, 0, 0, 0
	if len(margin) >= 4 {
		mTop, mBottom, mLeft, mRight = margin[0], margin[1], margin[2], margin[3]
	}

	// 텍스트/아이콘 색상 결정 (기본값: 흰색)
	var textColor color.Color = color.White
	if iconColor != nil {
		textColor = iconColor
	}

	// 컨텐츠 생성
	var content fyne.CanvasObject

	if icon != nil && label != "" {
		// 아이콘 + 텍스트
		iconImg := canvas.NewImageFromResource(theme.NewInvertedThemedResource(icon))
		iconImg.SetMinSize(fyne.NewSize(16, 16))
		iconImg.FillMode = canvas.ImageFillContain
		text := canvas.NewText(label, textColor)
		text.TextStyle = fyne.TextStyle{Bold: true}
		content = container.NewHBox(iconImg, text)
	} else if icon != nil {
		// 아이콘만
		iconImg := canvas.NewImageFromResource(theme.NewInvertedThemedResource(icon))
		iconImg.SetMinSize(fyne.NewSize(16, 16))
		iconImg.FillMode = canvas.ImageFillContain
		content = iconImg
	} else {
		// 텍스트만
		text := canvas.NewText(label, textColor)
		text.TextStyle = fyne.TextStyle{Bold: true}
		content = text
	}

	// 내부 패딩 적용 (기본값 고정: 8, 8, 16, 16)
	paddedContent := container.New(layout.NewCustomPaddedLayout(8, 8, 16, 16), container.NewCenter(content))

	// 버튼 생성
	var btn fyne.CanvasObject
	if bgColor != nil {
		bg := canvas.NewRectangle(bgColor)
		bg.CornerRadius = 4
		btn = NewTappableContainer(container.NewStack(bg, paddedContent), onTap)
	} else {
		btn = NewTappableContainer(paddedContent, onTap)
	}

	// 외부 여백 적용
	if mTop > 0 || mBottom > 0 || mLeft > 0 || mRight > 0 {
		return container.New(layout.NewCustomPaddedLayout(mTop, mBottom, mLeft, mRight), btn)
	}

	return btn
}

// 밝은 색상인지 판별합니다.
func isLightColor(c color.Color) bool {
	r, g, b, _ := c.RGBA()
	// 밝기 계산 (0-65535 범위)
	brightness := (r*299 + g*587 + b*114) / 1000
	return brightness > 32767
}

