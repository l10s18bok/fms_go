package component

import (
	"fms/internal/model"
	"fms/internal/themes"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
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

	// 외부 포트 (매칭 포트)
	f.matchPortEntry = widget.NewEntry()
	f.matchPortEntry.SetPlaceHolder("Port")

	// 소스 IP (선택)
	f.matchIPEntry = widget.NewEntry()
	f.matchIPEntry.SetPlaceHolder("Source IP")

	// 내부 IP
	f.transIPEntry = widget.NewEntry()
	f.transIPEntry.SetPlaceHolder("Dest IP")

	// 내부 포트
	f.transPortEntry = widget.NewEntry()
	f.transPortEntry.SetPlaceHolder("Port")

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

	// 첫 번째 행: Proto, ExtPort, SIP
	row1 := container.NewHBox(
		container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("Proto:")),
		container.NewGridWrap(fyne.NewSize(100, rowHeight), f.protoSel),
		container.NewGridWrap(fyne.NewSize(60, rowHeight), widget.NewLabel("ExtPort:")),
		container.NewGridWrap(fyne.NewSize(100, rowHeight), f.matchPortEntry),
		container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("SIP:")),
		container.NewGridWrap(fyne.NewSize(180, rowHeight), f.matchIPEntry),
	)

	// 두 번째 행: DstIP, DstPort
	row2 := container.NewHBox(
		container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("DIP:")),
		container.NewGridWrap(fyne.NewSize(180, rowHeight), f.transIPEntry),
		container.NewGridWrap(fyne.NewSize(60, rowHeight), widget.NewLabel("DPort:")),
		container.NewGridWrap(fyne.NewSize(100, rowHeight), f.transPortEntry),
	)

	// 헬프 버튼 ("?" 아이콘)
	helpBtn := widget.NewButtonWithIcon("", theme.QuestionIcon(), func() {
		f.showDNATHelp()
	})

	// 헤더: "포트 포워딩 (DNAT)" 레이블 + 헬프 버튼 + 오른쪽에 추가 버튼
	headerLeft := container.NewHBox(
		widget.NewLabel("포트 포워딩 (DNAT) 추가"),
		helpBtn,
	)
	header := container.NewBorder(
		nil, nil,
		headerLeft,
		container.NewGridWrap(fyne.NewSize(80, 36), f.addBtn),
	)

	// 전체 레이아웃
	f.content = container.NewVBox(
		widget.NewSeparator(),
		header,
		row1,
		row2,
	)
}

// showDNATHelp DNAT 도움말 팝업 표시
func (f *DNATForm) showDNATHelp() {
	ShowHelpPopup("DNAT 도움말", DNATHelpText, f.content)
}

// submitRule 규칙 생성 및 콜백 호출
func (f *DNATForm) submitRule() {
	// 필수 필드 확인
	if f.matchPortEntry.Text == "" || f.transIPEntry.Text == "" {
		return
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
