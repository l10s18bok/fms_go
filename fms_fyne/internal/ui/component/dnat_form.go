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
	protoSel       *FixedWidthSelect // 프로토콜
	matchPortEntry *widget.Entry     // 외부 포트 (매칭 포트)
	matchIPEntry   *widget.Entry     // 소스 IP (선택)
	transIPEntry   *widget.Entry     // 내부 IP (변환 대상)
	transPortEntry *widget.Entry     // 내부 포트 (변환 포트)
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

	// 프로토콜 선택
	f.protoSel = NewFixedWidthSelect(model.GetProtocolOptions(), nil, selectWidth)

	// 외부 포트 (매칭 포트) - 필수 필드
	f.matchPortEntry = widget.NewEntry()
	f.matchPortEntry.SetPlaceHolder("Port (필수)")
	f.matchPortEntry.Validator = func(s string) error {
		if s == "" {
			return fmt.Errorf("필수 입력")
		}
		return nil
	}

	// 소스 IP (선택)
	f.matchIPEntry = widget.NewEntry()
	f.matchIPEntry.SetPlaceHolder("Source IP")

	// 내부 IP - 필수 필드
	f.transIPEntry = widget.NewEntry()
	f.transIPEntry.SetPlaceHolder("Dest IP (필수)")
	f.transIPEntry.Validator = func(s string) error {
		if s == "" {
			return fmt.Errorf("필수 입력")
		}
		return nil
	}

	// 내부 포트
	f.transPortEntry = widget.NewEntry()
	f.transPortEntry.SetPlaceHolder("Port")

	// // 설명 (선택) - 현재 미사용
	// f.descEntry = widget.NewEntry()
	// f.descEntry.SetPlaceHolder("설명")

	// 레이블 너비 통일 (규칙 빌더와 동일)
	labelWidth := float32(50)
	rowHeight := float32(36)

	// 도움말 버튼 (아이콘)
	helpBtn := widget.NewButtonWithIcon("", theme.QuestionIcon(), func() {
		f.showDNATHelp()
	})

	// 첫 번째 행: Proto, ExtPort, SIP, ? (오른쪽 끝)
	row1 := container.NewBorder(
		nil, nil, nil,
		helpBtn, // 오른쪽 끝에 도움말 버튼
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("Proto:")),
			container.NewGridWrap(fyne.NewSize(100, rowHeight), f.protoSel),
			container.NewGridWrap(fyne.NewSize(60, rowHeight), widget.NewLabel("ExtPort:")),
			container.NewGridWrap(fyne.NewSize(150, rowHeight), f.matchPortEntry),
			container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("SIP:")),
			container.NewGridWrap(fyne.NewSize(180, rowHeight), f.matchIPEntry),
			layout.NewSpacer(),
		),
	)

	// 두 번째 행: DstIP, DstPort
	row2 := container.NewHBox(
		container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("DIP:")),
		container.NewGridWrap(fyne.NewSize(180, rowHeight), f.transIPEntry),
		container.NewGridWrap(fyne.NewSize(60, rowHeight), widget.NewLabel("DPort:")),
		container.NewGridWrap(fyne.NewSize(150, rowHeight), f.transPortEntry),
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
		NATType:       model.NATTypeDNAT,
		Protocol:      model.StringToProtocol(f.protoSel.Selected),
		MatchIP:       matchIP,
		MatchPort:     f.matchPortEntry.Text,
		TranslateIP:   f.transIPEntry.Text,
		TranslatePort: f.transPortEntry.Text,
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
	f.protoSel.SetSelected("tcp")
	f.matchPortEntry.SetText("")
	f.matchIPEntry.SetText("")
	f.transIPEntry.SetText("")
	f.transPortEntry.SetText("")
	// f.descEntry.SetText("") // 현재 미사용
}

// Content UI 컨테이너 반환
func (f *DNATForm) Content() *fyne.Container {
	return f.content
}
