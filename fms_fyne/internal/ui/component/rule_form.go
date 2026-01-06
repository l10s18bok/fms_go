package component

import (
	"fms/internal/model"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// FixedWidthSelect MinSize를 고정한 Select 위젯
type FixedWidthSelect struct {
	widget.Select
	fixedWidth float32
}

// NewFixedWidthSelect 고정 너비 Select 생성
func NewFixedWidthSelect(options []string, changed func(string), width float32) *FixedWidthSelect {
	s := &FixedWidthSelect{fixedWidth: width}
	s.Options = options
	s.OnChanged = changed
	s.ExtendBaseWidget(s)
	return s
}

// MinSize 고정 너비 반환
func (s *FixedWidthSelect) MinSize() fyne.Size {
	min := s.Select.MinSize()
	return fyne.NewSize(s.fixedWidth, min.Height)
}

// RuleForm 규칙 추가 폼 컴포넌트
type RuleForm struct {
	onAdd func(*model.FirewallRule)

	// UI 요소
	chainSel   *FixedWidthSelect
	protoSel   *FixedWidthSelect
	actionSel  *FixedWidthSelect
	dportEntry *widget.Entry
	sipEntry   *widget.Entry
	dipEntry   *widget.Entry
	blackCheck *widget.Check
	whiteCheck *widget.Check
	addBtn     fyne.CanvasObject
	content    *fyne.Container
}

// NewRuleForm 새 규칙 추가 폼 생성
func NewRuleForm(onAdd func(*model.FirewallRule)) *RuleForm {
	form := &RuleForm{
		onAdd: onAdd,
	}
	form.createUI()
	form.Reset()
	return form
}

// createUI UI 생성
func (f *RuleForm) createUI() {
	// 드롭다운 고정 너비
	selectWidth := float32(100)

	// Chain 선택
	f.chainSel = NewFixedWidthSelect(model.GetChainOptions(), nil, selectWidth)

	// Protocol 선택
	f.protoSel = NewFixedWidthSelect(model.GetProtocolOptions(), nil, selectWidth)

	// Action 선택
	f.actionSel = NewFixedWidthSelect(model.GetActionOptions(), nil, selectWidth)

	// DPort 입력
	f.dportEntry = widget.NewEntry()
	f.dportEntry.SetPlaceHolder("포트")

	// SIP 입력
	f.sipEntry = widget.NewEntry()
	f.sipEntry.SetPlaceHolder("Source IP")

	// DIP 입력
	f.dipEntry = widget.NewEntry()
	f.dipEntry.SetPlaceHolder("Dest IP")

	// 체크박스들
	f.blackCheck = widget.NewCheck("Black", nil)
	f.whiteCheck = widget.NewCheck("White", nil)

	// 추가 버튼 (진한 회색 배경)
	f.addBtn = NewColoredButton("+ 추가", ButtonDark, func() {
		f.submitRule()
	})

	// 레이블 너비 통일
	labelWidth := float32(50)
	rowHeight := float32(36)

	// 첫 번째 행: Chain, Protocol, Action, DPort
	row1 := container.NewHBox(
		container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("Chain:")),
		container.NewGridWrap(fyne.NewSize(100, rowHeight), f.chainSel),
		container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("Proto:")),
		container.NewGridWrap(fyne.NewSize(100, rowHeight), f.protoSel),
		container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("Action:")),
		container.NewGridWrap(fyne.NewSize(100, rowHeight), f.actionSel),
		container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("Port:")),
		container.NewGridWrap(fyne.NewSize(140, rowHeight), f.dportEntry),
	)

	// 두 번째 행: SIP, DIP
	row2 := container.NewHBox(
		container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("SIP:")),
		container.NewGridWrap(fyne.NewSize(230, rowHeight), f.sipEntry),
		container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("DIP:")),
		container.NewGridWrap(fyne.NewSize(230, rowHeight), f.dipEntry),
	)

	// 세 번째 행: 체크박스
	row3 := container.NewHBox(
		container.NewGridWrap(fyne.NewSize(80, rowHeight), f.blackCheck),
		container.NewGridWrap(fyne.NewSize(80, rowHeight), f.whiteCheck),
	)

	// 전체 폼 레이아웃
	formContent := container.NewVBox(row1, row2, row3)

	// 헤더: "규칙 추가" 레이블 + 오른쪽에 추가 버튼
	header := container.NewBorder(
		nil, nil, // top, bottom
		widget.NewLabel("규칙 추가"), // left
		container.NewGridWrap(fyne.NewSize(80, 36), f.addBtn), // right
	)

	// 테두리가 있는 카드 형태
	f.content = container.NewVBox(
		widget.NewSeparator(),
		header,
		formContent,
	)
}

// submitRule 규칙 생성 및 콜백 호출
func (f *RuleForm) submitRule() {
	rule := &model.FirewallRule{
		Chain:    model.StringToChain(f.chainSel.Selected),
		Protocol: model.StringToProtocol(f.protoSel.Selected),
		Action:   model.StringToAction(f.actionSel.Selected),
		DPort:    f.dportEntry.Text,
		SIP:      f.sipEntry.Text,
		DIP:      f.dipEntry.Text,
		Black:    f.blackCheck.Checked,
		White:    f.whiteCheck.Checked,
	}

	if f.onAdd != nil {
		f.onAdd(rule)
	}

	f.Reset()
}

// Reset 폼 초기화
func (f *RuleForm) Reset() {
	f.chainSel.SetSelected("INPUT")
	f.protoSel.SetSelected("tcp")
	f.actionSel.SetSelected("DROP")
	f.dportEntry.SetText("")
	f.sipEntry.SetText("")
	f.dipEntry.SetText("")
	f.blackCheck.SetChecked(false)
	f.whiteCheck.SetChecked(false)
}

// Content UI 컨테이너 반환
func (f *RuleForm) Content() *fyne.Container {
	return f.content
}
