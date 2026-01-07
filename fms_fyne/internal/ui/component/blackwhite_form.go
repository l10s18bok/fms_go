package component

import (
	"fms/internal/model"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// BlackWhiteForm Black/White 규칙 추가 폼 컴포넌트
// 일반 규칙보다 간소화된 폼: 타입(Black/White)과 IP만 입력
type BlackWhiteForm struct {
	onAdd func(*model.FirewallRule)

	// UI 요소
	typeSel  *widget.Select // "Black" / "White"
	sipEntry *widget.Entry  // Source IP
	addBtn   fyne.CanvasObject
	content  *fyne.Container
}

// NewBlackWhiteForm 새 Black/White 폼 생성
func NewBlackWhiteForm(onAdd func(*model.FirewallRule)) *BlackWhiteForm {
	form := &BlackWhiteForm{
		onAdd: onAdd,
	}
	form.createUI()
	form.Reset()
	return form
}

// createUI UI 생성
func (f *BlackWhiteForm) createUI() {
	// 타입 선택 (Black/White)
	f.typeSel = widget.NewSelect([]string{"Black", "White"}, nil)

	// SIP 입력
	f.sipEntry = widget.NewEntry()
	f.sipEntry.SetPlaceHolder("차단/허용할 IP (예: 192.168.1.100)")

	// 추가 버튼
	f.addBtn = NewColoredButton("+ 추가", ButtonDark, func() {
		f.submitRule()
	})

	// 레이블 너비 통일
	labelWidth := float32(50)
	rowHeight := float32(36)

	// 입력 행: 타입, IP
	row := container.NewHBox(
		container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("타입:")),
		container.NewGridWrap(fyne.NewSize(100, rowHeight), f.typeSel),
		container.NewGridWrap(fyne.NewSize(30, rowHeight), widget.NewLabel("IP:")),
		container.NewGridWrap(fyne.NewSize(300, rowHeight), f.sipEntry),
	)

	// 헤더: "Black/White 규칙 추가" 레이블 + 오른쪽에 추가 버튼
	header := container.NewBorder(
		nil, nil, // top, bottom
		widget.NewLabel("Black/White 규칙 추가"), // left
		container.NewGridWrap(fyne.NewSize(80, 36), f.addBtn), // right
	)

	// 전체 레이아웃
	f.content = container.NewVBox(
		widget.NewSeparator(),
		header,
		row,
	)
}

// submitRule 규칙 생성 및 콜백 호출
func (f *BlackWhiteForm) submitRule() {
	// IP가 비어있으면 무시
	if f.sipEntry.Text == "" {
		return
	}

	isBlack := f.typeSel.Selected == "Black"
	isWhite := f.typeSel.Selected == "White"

	// 기본 Action 설정: Black=DROP, White=ACCEPT
	action := model.ActionDROP
	if isWhite {
		action = model.ActionACCEPT
	}

	rule := &model.FirewallRule{
		Chain:    model.ChainINPUT,  // 고정
		Protocol: model.ProtocolANY, // 고정 (모든 프로토콜)
		Action:   action,
		SIP:      f.sipEntry.Text,
		Black:    isBlack,
		White:    isWhite,
	}

	if f.onAdd != nil {
		f.onAdd(rule)
	}

	f.Reset()
}

// Reset 폼 초기화
func (f *BlackWhiteForm) Reset() {
	f.typeSel.SetSelected("Black")
	f.sipEntry.SetText("")
}

// Content UI 컨테이너 반환
func (f *BlackWhiteForm) Content() *fyne.Container {
	return f.content
}
