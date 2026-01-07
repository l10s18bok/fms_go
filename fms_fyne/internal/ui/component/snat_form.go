package component

import (
	"fms/internal/model"
	"fms/internal/themes"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
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

	// 소스 네트워크
	f.matchIPEntry = widget.NewEntry()
	f.matchIPEntry.SetPlaceHolder("Source IP")

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

	// 추가 버튼
	f.addBtn = NewCustomButton("+ 추가", nil, nil, themes.Colors["darkgray"], func() {
		f.submitRule()
	})

	// 레이블 너비 통일 (규칙 빌더와 동일)
	labelWidth := float32(50)
	rowHeight := float32(36)

	// 첫 번째 행: Type, Proto, SIP
	row1 := container.NewHBox(
		container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("Type:")),
		container.NewGridWrap(fyne.NewSize(130, rowHeight), f.natTypeSel),
		container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("Proto:")),
		container.NewGridWrap(fyne.NewSize(100, rowHeight), f.protoSel),
		container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("SIP:")),
		container.NewGridWrap(fyne.NewSize(180, rowHeight), f.matchIPEntry),
	)

	// 두 번째 행: InIF, OutIF
	row2 := container.NewHBox(
		container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("InIF:")),
		container.NewGridWrap(fyne.NewSize(100, rowHeight), f.inIfEntry),
		container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("OutIF:")),
		container.NewGridWrap(fyne.NewSize(100, rowHeight), f.outIfEntry),
	)

	// 세 번째 행: TransIP (SNAT만 표시)
	f.transIPRow = container.NewHBox(
		container.NewGridWrap(fyne.NewSize(60, rowHeight), widget.NewLabel("TransIP:")),
		container.NewGridWrap(fyne.NewSize(180, rowHeight), f.transIPEntry),
	)

	// 헤더: "소스 NAT (SNAT/MASQ)" 레이블 + 오른쪽에 추가 버튼
	header := container.NewBorder(
		nil, nil,
		widget.NewLabel("소스 NAT (SNAT/MASQUERADE) 추가"),
		container.NewGridWrap(fyne.NewSize(80, 36), f.addBtn),
	)

	// 전체 레이아웃
	f.content = container.NewVBox(
		widget.NewSeparator(),
		header,
		row1,
		row2,
		f.transIPRow,
	)
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

// submitRule 규칙 생성 및 콜백 호출
func (f *SNATForm) submitRule() {
	// 필수 필드 확인: 소스 네트워크
	if f.matchIPEntry.Text == "" {
		return
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
