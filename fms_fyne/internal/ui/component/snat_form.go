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

// SNATForm SNAT/MASQUERADE 규칙 추가 폼
type SNATForm struct {
	onAdd func(*model.NATRule)

	// UI 요소
	natTypeSel   *FixedWidthSelect // NAT 타입 (SNAT / MASQUERADE)
	protoSel     *FixedWidthSelect // 프로토콜
	matchIPEntry *widget.Entry     // 소스 네트워크
	inIfEntry    *widget.Entry     // 입력 인터페이스
	outIfEntry   *widget.Entry     // 출력 인터페이스
	transIPEntry *widget.Entry     // 변환 IP (SNAT만, 선택)
	// descEntry    *widget.Entry     // 설명 (선택) - 현재 미사용
	addBtn fyne.CanvasObject
	content      *fyne.Container

	// 변환 IP 행 (조건부 표시용)
	transIPRow *fyne.Container
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
	selectWidth := float32(120)

	// NAT 타입 선택 (SNAT / MASQUERADE)
	f.natTypeSel = NewFixedWidthSelect(model.GetSNATTypeOptions(), func(s string) {
		f.onNATTypeChanged(s)
	}, selectWidth)

	// 프로토콜 선택
	f.protoSel = NewFixedWidthSelect(model.GetProtocolOptions(), nil, float32(100))

	// 소스 네트워크 - 필수 필드
	f.matchIPEntry = widget.NewEntry()
	f.matchIPEntry.SetPlaceHolder("Source IP (필수)")
	f.matchIPEntry.Validator = func(s string) error {
		if s == "" {
			return fmt.Errorf("필수 입력")
		}
		return nil
	}

	// 입력 인터페이스
	f.inIfEntry = widget.NewEntry()
	f.inIfEntry.SetPlaceHolder("eth1")

	// 출력 인터페이스
	f.outIfEntry = widget.NewEntry()
	f.outIfEntry.SetPlaceHolder("eth0")

	// 변환 IP (SNAT만)
	f.transIPEntry = widget.NewEntry()
	f.transIPEntry.SetPlaceHolder("Translate IP")

	// // 설명 (선택) - 현재 미사용
	// f.descEntry = widget.NewEntry()
	// f.descEntry.SetPlaceHolder("설명")

	// 레이블 너비 통일 (규칙 빌더와 동일)
	labelWidth := float32(50)
	rowHeight := float32(36)

	// 도움말 버튼 (아이콘)
	helpBtn := widget.NewButtonWithIcon("", theme.QuestionIcon(), func() {
		f.showSNATHelp()
	})

	// 첫 번째 행: Type, Proto, SIP, ? (오른쪽 끝)
	row1 := container.NewBorder(
		nil, nil, nil,
		helpBtn, // 오른쪽 끝에 도움말 버튼
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("Type:")),
			container.NewGridWrap(fyne.NewSize(130, rowHeight), f.natTypeSel),
			container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("Proto:")),
			container.NewGridWrap(fyne.NewSize(100, rowHeight), f.protoSel),
			container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("SIP:")),
			container.NewGridWrap(fyne.NewSize(180, rowHeight), f.matchIPEntry),
			layout.NewSpacer(),
		),
	)

	// 두 번째 행: InIF, OutIF
	row2 := container.NewHBox(
		container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("InIF:")),
		container.NewGridWrap(fyne.NewSize(150, rowHeight), f.inIfEntry),
		container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("OutIF:")),
		container.NewGridWrap(fyne.NewSize(150, rowHeight), f.outIfEntry),
	)

	// 세 번째 행: TransIP (SNAT만 표시)
	f.transIPRow = container.NewHBox(
		container.NewGridWrap(fyne.NewSize(60, rowHeight), widget.NewLabel("TransIP:")),
		container.NewGridWrap(fyne.NewSize(180, rowHeight), f.transIPEntry),
	)

	// 전체 레이아웃 (행 간격 추가)
	f.content = container.NewVBox(
		row1,
		widget.NewSeparator(), // 행 사이 간격
		row2,
		widget.NewSeparator(), // 행 사이 간격
		f.transIPRow,
	)
}

// showSNATHelp SNAT 도움말 팝업 표시
func (f *SNATForm) showSNATHelp() {
	ShowHelpPopup("SNAT/MASQUERADE 도움말", SNATHelpText, f.content)
}

// onNATTypeChanged NAT 타입 변경 시 변환 IP 행 표시/숨김
func (f *SNATForm) onNATTypeChanged(natType string) {
	if natType == "SNAT" {
		f.transIPRow.Show()
	} else {
		// MASQUERADE는 변환 IP 불필요
		f.transIPRow.Hide()
		f.transIPEntry.SetText("")
	}
	f.content.Refresh()
}

// SubmitRule 규칙 생성 및 콜백 호출 (다이얼로그에서 호출 가능)
// 성공 시 true, 실패 시 false 반환
func (f *SNATForm) SubmitRule() bool {
	// 필수 필드 validation 체크
	if f.matchIPEntry.Text == "" {
		f.matchIPEntry.SetValidationError(fmt.Errorf("필수 입력"))
		return false
	}

	natType := model.StringToNATType(f.natTypeSel.Selected)

	rule := &model.NATRule{
		NATType:      natType,
		Protocol:     model.StringToProtocol(f.protoSel.Selected),
		MatchIP:      f.matchIPEntry.Text,
		InInterface:  f.inIfEntry.Text,
		OutInterface: f.outIfEntry.Text,
		// Description:  f.descEntry.Text, // 현재 미사용
	}

	// SNAT만 변환 IP 설정
	if natType == model.NATTypeSNAT {
		rule.TranslateIP = f.transIPEntry.Text
	}

	if f.onAdd != nil {
		f.onAdd(rule)
	}

	f.Reset()
	return true
}

// Reset 폼 초기화
func (f *SNATForm) Reset() {
	f.natTypeSel.SetSelected("SNAT")
	f.protoSel.SetSelected("tcp")
	f.matchIPEntry.SetText("")
	f.inIfEntry.SetText("")
	f.outIfEntry.SetText("")
	f.transIPEntry.SetText("")
	// f.descEntry.SetText("") // 현재 미사용
	f.transIPRow.Show()
}

// Content UI 컨테이너 반환
func (f *SNATForm) Content() *fyne.Container {
	return f.content
}
