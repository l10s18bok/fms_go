package component

import (
	"fmt"

	"fms/internal/model"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// DNATForm DNAT (포트 포워딩) 규칙 추가 폼
type DNATForm struct {
	onAdd func(*model.NATRule)

	// UI 요소
	commandSel     *FixedWidthSelect // INSERT/DELETE
	protoSel       *FixedWidthSelect // 프로토콜
	matchPortEntry *widget.Entry     // 매칭 포트 (MatchPort)
	matchIPEntry   *widget.Entry     // 매칭 IP (MatchIP)
	transIPEntry   *widget.Entry     // 변환 IP (TransIP)
	transPortEntry *widget.Entry     // 변환 포트 (TransPort)
	inIfEntry      *widget.Entry     // 입력 인터페이스 (InIF)
	// descEntry      *widget.Entry     // 설명 (선택) - 현재 미사용
	addBtn  fyne.CanvasObject
	content *fyne.Container
}

// NewDNATForm 새 DNAT 폼 생성
func NewDNATForm(onAdd func(*model.NATRule)) *DNATForm {
	form := &DNATForm{
		onAdd: onAdd,
	}
	form.createUI()
	form.Reset()
	return form
}

// createUI UI 생성
func (f *DNATForm) createUI() {
	selectWidth := float32(100)

	// Command 선택 (INSERT/DELETE)
	f.commandSel = NewFixedWidthSelect([]string{"INSERT", "DELETE"}, nil, 120)

	// 프로토콜 선택
	f.protoSel = NewFixedWidthSelect(model.GetProtocolOptions(), nil, selectWidth)

	// 매칭 포트 (MatchPort) - 필수 필드
	f.matchPortEntry = widget.NewEntry()
	f.matchPortEntry.SetPlaceHolder("Port (필수)")
	f.matchPortEntry.Validator = func(s string) error {
		if s == "" {
			return fmt.Errorf("필수 입력")
		}
		return nil
	}

	// 매칭 IP (MatchIP) - 선택
	f.matchIPEntry = widget.NewEntry()
	f.matchIPEntry.SetPlaceHolder("MatchIP")

	// 변환 IP (TransIP) - 필수 필드
	f.transIPEntry = widget.NewEntry()
	f.transIPEntry.SetPlaceHolder("TransIP (필수)")
	f.transIPEntry.Validator = func(s string) error {
		if s == "" {
			return fmt.Errorf("필수 입력")
		}
		return nil
	}

	// 변환 포트 (TransPort)
	f.transPortEntry = widget.NewEntry()
	f.transPortEntry.SetPlaceHolder("Port")

	// 입력 인터페이스 (InIF)
	f.inIfEntry = widget.NewEntry()
	f.inIfEntry.SetPlaceHolder("eth0")

	// // 설명 (선택) - 현재 미사용
	// f.descEntry = widget.NewEntry()
	// f.descEntry.SetPlaceHolder("설명")

	// 행 높이
	rowHeight := float32(36)

	// 도움말 버튼 (아이콘)
	helpBtn := widget.NewButtonWithIcon("", theme.QuestionIcon(), func() {
		f.showDNATHelp()
	})

	// 첫 번째 행: Cmd, Proto, MatchPort, MatchIP, ? (오른쪽 끝)
	row1 := container.NewBorder(
		nil, nil, nil,
		helpBtn, // 오른쪽 끝에 도움말 버튼
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(35, rowHeight), widget.NewLabel("Cmd:")),
			container.NewGridWrap(fyne.NewSize(120, rowHeight), f.commandSel),
			container.NewGridWrap(fyne.NewSize(50, rowHeight), widget.NewLabel("Proto:")),
			container.NewGridWrap(fyne.NewSize(80, rowHeight), f.protoSel),
			container.NewGridWrap(fyne.NewSize(85, rowHeight), widget.NewLabel("MatchPort:")),
			container.NewGridWrap(fyne.NewSize(100, rowHeight), f.matchPortEntry),
			container.NewGridWrap(fyne.NewSize(70, rowHeight), widget.NewLabel("MatchIP:")),
			container.NewGridWrap(fyne.NewSize(150, rowHeight), f.matchIPEntry),
			layout.NewSpacer(),
		),
	)

	// 두 번째 행: TransIP, TransPort, InIF
	row2 := container.NewHBox(
		container.NewGridWrap(fyne.NewSize(70, rowHeight), widget.NewLabel("TransIP:")),
		container.NewGridWrap(fyne.NewSize(170, rowHeight), f.transIPEntry),
		container.NewGridWrap(fyne.NewSize(75, rowHeight), widget.NewLabel("TransPort:")),
		container.NewGridWrap(fyne.NewSize(100, rowHeight), f.transPortEntry),
		container.NewGridWrap(fyne.NewSize(45, rowHeight), widget.NewLabel("InIF:")),
		container.NewGridWrap(fyne.NewSize(100, rowHeight), f.inIfEntry),
	)

	// 전체 레이아웃 (행 간격 추가)
	f.content = container.NewVBox(
		row1,
		widget.NewSeparator(), // 행 사이 간격
		row2,
	)
}

// showDNATHelp DNAT 도움말 팝업 표시
func (f *DNATForm) showDNATHelp() {
	ShowHelpPopup("DNAT 도움말", DNATHelpText, f.content)
}

// SubmitRule 규칙 생성 및 콜백 호출 (다이얼로그에서 호출 가능)
// 성공 시 true, 실패 시 false 반환
func (f *DNATForm) SubmitRule() bool {
	// 필수 필드 validation 체크
	hasError := false

	if f.matchPortEntry.Text == "" {
		f.matchPortEntry.SetValidationError(fmt.Errorf("필수 입력"))
		hasError = true
	}

	if f.transIPEntry.Text == "" {
		f.transIPEntry.SetValidationError(fmt.Errorf("필수 입력"))
		hasError = true
	}

	if hasError {
		return false
	}

	matchIP := f.matchIPEntry.Text
	if matchIP == "" {
		matchIP = "ANY"
	}

	rule := &model.NATRule{
		Command:       model.StringToCommand(f.commandSel.Selected),
		NATType:       model.NATTypeDNAT,
		Protocol:      model.StringToProtocol(f.protoSel.Selected),
		MatchIP:       matchIP,
		MatchPort:     f.matchPortEntry.Text,
		TranslateIP:   f.transIPEntry.Text,
		TranslatePort: f.transPortEntry.Text,
		InInterface:   f.inIfEntry.Text,
		// Description:   f.descEntry.Text, // 현재 미사용
	}

	if f.onAdd != nil {
		f.onAdd(rule)
	}

	f.Reset()
	return true
}

// Reset 폼 초기화
func (f *DNATForm) Reset() {
	f.commandSel.SetSelected("INSERT")
	f.protoSel.SetSelected("tcp")
	f.matchPortEntry.SetText("")
	f.matchIPEntry.SetText("")
	f.transIPEntry.SetText("")
	f.transPortEntry.SetText("")
	f.inIfEntry.SetText("")
	// f.descEntry.SetText("") // 현재 미사용
}

// Content UI 컨테이너 반환
func (f *DNATForm) Content() *fyne.Container {
	return f.content
}
