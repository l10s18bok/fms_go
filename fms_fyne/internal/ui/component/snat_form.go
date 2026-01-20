package component

import (
	"fms/internal/model"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// SNATForm SNAT 규칙 추가 폼
type SNATForm struct {
	onAdd func(*model.NATRule)

	// UI 요소
	protoSel     *FixedWidthSelect // 프로토콜
	matchIPEntry *widget.Entry     // Match_IP
	outIfEntry   *widget.Entry     // Out_Face (출력 인터페이스)
	transIPEntry *widget.Entry     // Trans_IP (변환 IP)
	// descEntry    *widget.Entry     // 설명 (선택) - 현재 미사용
	addBtn  fyne.CanvasObject
	content *fyne.Container
}

// NewSNATForm 새 SNAT 폼 생성
func NewSNATForm(onAdd func(*model.NATRule)) *SNATForm {
	form := &SNATForm{
		onAdd: onAdd,
	}
	form.createUI()
	form.Reset()
	return form
}

// createUI UI 생성
func (f *SNATForm) createUI() {
	// 프로토콜 선택
	f.protoSel = NewFixedWidthSelect(model.GetProtocolOptions(), nil, float32(80))

	// Match_IP
	f.matchIPEntry = widget.NewEntry()
	f.matchIPEntry.SetPlaceHolder("Match IP")

	// Out_Face (출력 인터페이스)
	f.outIfEntry = widget.NewEntry()
	f.outIfEntry.SetPlaceHolder("Out Interface")

	// Trans_IP (변환 IP)
	f.transIPEntry = widget.NewEntry()
	f.transIPEntry.SetPlaceHolder("Trans IP")

	// 행 높이
	rowHeight := float32(36)

	// 도움말 버튼 (아이콘)
	helpBtn := widget.NewButtonWithIcon("", theme.QuestionIcon(), func() {
		f.showSNATHelp()
	})

	// 첫 번째 행: Proto, Match_IP, ? (오른쪽 끝)
	row1 := container.NewBorder(
		nil, nil, nil,
		helpBtn, // 오른쪽 끝에 도움말 버튼
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(50, rowHeight), widget.NewLabel("Proto:")),
			container.NewGridWrap(fyne.NewSize(80, rowHeight), f.protoSel),
			container.NewGridWrap(fyne.NewSize(70, rowHeight), widget.NewLabel("Match_IP:")),
			container.NewGridWrap(fyne.NewSize(180, rowHeight), f.matchIPEntry),
			layout.NewSpacer(),
		),
	)

	// 두 번째 행: Out_Face, Trans_IP
	row2 := container.NewHBox(
		container.NewGridWrap(fyne.NewSize(70, rowHeight), widget.NewLabel("Out_Face:")),
		container.NewGridWrap(fyne.NewSize(150, rowHeight), f.outIfEntry),
		container.NewGridWrap(fyne.NewSize(65, rowHeight), widget.NewLabel("Trans_IP:")),
		container.NewGridWrap(fyne.NewSize(180, rowHeight), f.transIPEntry),
	)

	// 전체 레이아웃 (행 간격 추가)
	f.content = container.NewVBox(
		row1,
		widget.NewSeparator(), // 행 사이 간격
		row2,
	)
}

// showSNATHelp SNAT 도움말 팝업 표시
func (f *SNATForm) showSNATHelp() {
	ShowHelpPopup("SNAT 도움말", SNATHelpText, f.content)
}

// SubmitRule 규칙 생성 및 콜백 호출 (다이얼로그에서 호출 가능)
// 성공 시 true, 실패 시 false 반환
func (f *SNATForm) SubmitRule() bool {
	rule := &model.NATRule{
		NATType:      model.NATTypeSNAT,
		Protocol:     model.StringToProtocol(f.protoSel.Selected),
		MatchIP:      f.matchIPEntry.Text,
		OutInterface: f.outIfEntry.Text,
		TranslateIP:  f.transIPEntry.Text,
	}

	if f.onAdd != nil {
		f.onAdd(rule)
	}

	f.Reset()
	return true
}

// Reset 폼 초기화
func (f *SNATForm) Reset() {
	f.protoSel.SetSelected("tcp")
	f.matchIPEntry.SetText("")
	f.outIfEntry.SetText("")
	f.transIPEntry.SetText("")
}

// Content UI 컨테이너 반환
func (f *SNATForm) Content() *fyne.Container {
	return f.content
}
