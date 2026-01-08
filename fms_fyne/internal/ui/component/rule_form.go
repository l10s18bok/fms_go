package component

import (
	"strings"

	"fms/internal/model"
	"fms/internal/themes"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
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

// RuleForm 규칙 추가 폼 컴포넌트 (일반 규칙용)
type RuleForm struct {
	onAdd func(*model.FirewallRule)

	// UI 요소
	chainSel   *FixedWidthSelect
	protoSel   *FixedWidthSelect
	actionSel  *FixedWidthSelect
	dportEntry *widget.Entry
	sipEntry   *widget.Entry
	dipEntry   *widget.Entry
	addBtn     fyne.CanvasObject
	content    *fyne.Container

	// TCP Flags 옵션 UI
	tcpFlagsPresetSel *widget.Select
	tcpMaskChecks     map[string]*widget.Check // 검사할 플래그
	tcpSetChecks      map[string]*widget.Check // 설정된 플래그
	tcpOptionsBox     *fyne.Container

	// ICMP 옵션 UI
	icmpTypeSel    *widget.Select
	icmpTypeEntry  *widget.Entry   // 커스텀 숫자용
	icmpCodeSel    *widget.Select  // Code 드롭다운
	icmpCodeEntry  *widget.Entry   // Code 커스텀 숫자용
	icmpCodeRow    *fyne.Container // Code 행 (조건부 표시용)
	icmpOptionsBox *fyne.Container

	// 옵션 컨테이너
	optionsContainer *fyne.Container

	// 플래그: 프리셋 변경 중 체크박스 이벤트 무시
	updatingFromPreset bool
}

// NewRuleForm 새 규칙 추가 폼 생성
func NewRuleForm(onAdd func(*model.FirewallRule)) *RuleForm {
	form := &RuleForm{
		onAdd:         onAdd,
		tcpMaskChecks: make(map[string]*widget.Check),
		tcpSetChecks:  make(map[string]*widget.Check),
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

	// Protocol 선택 (OnChanged 핸들러 추가)
	f.protoSel = NewFixedWidthSelect(model.GetProtocolOptions(), func(s string) {
		f.onProtocolChanged(s)
	}, selectWidth)

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

	// 추가 버튼 (진한 회색 배경)
	f.addBtn = NewCustomButton("+ 추가", nil, nil, themes.Colors["darkgray"], func() {
		f.submitRule()
	})

	// TCP Flags 옵션 UI 생성
	f.createTCPFlagsUI()

	// ICMP 옵션 UI 생성
	f.createICMPOptionsUI()

	// 옵션 컨테이너 (프로토콜에 따라 동적으로 표시)
	f.optionsContainer = container.NewVBox()

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

	// 전체 폼 레이아웃 (Black/White 체크박스 제거됨 - BlackWhiteForm에서 별도 처리)
	formContent := container.NewVBox(row1, row2, f.optionsContainer)

	// 헤더: "규칙 추가" 레이블 + 오른쪽에 추가 버튼
	header := container.NewBorder(
		nil, nil, // top, bottom
		widget.NewLabel("규칙 추가"),                              // left
		container.NewGridWrap(fyne.NewSize(80, 36), f.addBtn), // right
	)

	// 테두리가 있는 카드 형태
	f.content = container.NewVBox(
		widget.NewSeparator(),
		header,
		formContent,
	)
}

// createTCPFlagsUI TCP Flags 옵션 UI 생성
func (f *RuleForm) createTCPFlagsUI() {
	// 프리셋 목록 생성
	presets := model.GetTCPFlagsPresets()
	presetNames := make([]string, len(presets))
	for i, p := range presets {
		presetNames[i] = p.Name
	}

	f.tcpFlagsPresetSel = widget.NewSelect(presetNames, func(s string) {
		f.onTCPPresetChanged(s)
	})

	// 플래그 체크박스 생성
	flags := model.GetTCPFlagsList()
	rowHeight := float32(30)
	labelWidth := float32(50) // 레이블 너비 통일

	// 검사할 플래그 행
	maskRow := container.NewHBox(
		container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("Mask:")),
	)
	for _, flag := range flags {
		check := widget.NewCheck(strings.ToUpper(flag), func(b bool) {
			if !f.updatingFromPreset {
				f.tcpFlagsPresetSel.SetSelected("Custom")
			}
		})
		f.tcpMaskChecks[flag] = check
		maskRow.Add(container.NewGridWrap(fyne.NewSize(60, rowHeight), check))
	}

	// 설정된 플래그 행
	setRow := container.NewHBox(
		container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("Set:")),
	)
	for _, flag := range flags {
		check := widget.NewCheck(strings.ToUpper(flag), func(b bool) {
			if !f.updatingFromPreset {
				f.tcpFlagsPresetSel.SetSelected("Custom")
			}
		})
		f.tcpSetChecks[flag] = check
		setRow.Add(container.NewGridWrap(fyne.NewSize(60, rowHeight), check))
	}

	// 프리셋 행
	presetRow := container.NewHBox(
		container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("Preset:")),
		container.NewGridWrap(fyne.NewSize(180, rowHeight), f.tcpFlagsPresetSel),
	)

	// 헬프 버튼 ("?" 아이콘)
	helpBtn := widget.NewButtonWithIcon("", theme.QuestionIcon(), func() {
		f.showTCPFlagsHelp()
	})

	// 헤더: "TCP Flags" + 헬프 버튼
	headerRow := container.NewHBox(
		widget.NewLabel("TCP Flags"),
		helpBtn,
	)

	f.tcpOptionsBox = container.NewVBox(
		headerRow,
		presetRow,
		maskRow,
		setRow,
	)
}

// showTCPFlagsHelp TCP Flags 옵션 헬프 팝업 표시
func (f *RuleForm) showTCPFlagsHelp() {
	ShowHelpPopup("TCP Flags 도움말", TCPFlagsHelpText, f.content)
}

// createICMPOptionsUI ICMP 옵션 UI 생성
func (f *RuleForm) createICMPOptionsUI() {
	// ICMP Type 선택
	f.icmpTypeSel = widget.NewSelect(model.GetICMPTypeOptions(), func(s string) {
		f.onICMPTypeChanged(s)
	})

	// Type 커스텀 숫자 입력
	f.icmpTypeEntry = widget.NewEntry()
	f.icmpTypeEntry.SetPlaceHolder("Type #")
	f.icmpTypeEntry.Hide()

	// Code 드롭다운 (destination-unreachable 전용)
	f.icmpCodeSel = widget.NewSelect(model.GetICMPCodeOptions(), func(s string) {
		f.onICMPCodeChanged(s)
	})

	// Code 커스텀 숫자 입력
	f.icmpCodeEntry = widget.NewEntry()
	f.icmpCodeEntry.SetPlaceHolder("Code #")
	f.icmpCodeEntry.Hide()

	rowHeight := float32(36)

	typeRow := container.NewHBox(
		widget.NewLabel("Type:"),
		container.NewGridWrap(fyne.NewSize(200, rowHeight), f.icmpTypeSel),
		container.NewGridWrap(fyne.NewSize(80, rowHeight), f.icmpTypeEntry),
	)

	// Code 행 (조건부 표시)
	f.icmpCodeRow = container.NewHBox(
		widget.NewLabel("Code:"),
		container.NewGridWrap(fyne.NewSize(200, rowHeight), f.icmpCodeSel),
		container.NewGridWrap(fyne.NewSize(80, rowHeight), f.icmpCodeEntry),
	)
	f.icmpCodeRow.Hide() // 기본적으로 숨김

	// 헬프 버튼 ("?" 아이콘)
	helpBtn := widget.NewButtonWithIcon("", theme.QuestionIcon(), func() {
		f.showICMPHelp()
	})

	// 헤더: "ICMP Options" + 헬프 버튼
	headerRow := container.NewHBox(
		widget.NewLabel("ICMP Options"),
		helpBtn,
	)

	f.icmpOptionsBox = container.NewVBox(
		headerRow,
		typeRow,
		f.icmpCodeRow,
	)
}

// showICMPHelp ICMP 옵션 헬프 팝업 표시
func (f *RuleForm) showICMPHelp() {
	ShowHelpPopup("ICMP Options 도움말", ICMPOptionsHelpText, f.content)
}

// onProtocolChanged 프로토콜 변경 시 옵션 UI 전환 및 포트 필드 활성/비활성화
func (f *RuleForm) onProtocolChanged(proto string) {
	f.optionsContainer.Objects = nil

	switch strings.ToLower(proto) {
	case "tcp":
		f.optionsContainer.Add(f.tcpOptionsBox)
		f.setTCPOptionsEnabled(true)
		f.dportEntry.Enable()
		f.dportEntry.SetPlaceHolder("포트")
	case "udp":
		// UDP: TCP 옵션 박스 표시하되 비활성화
		f.optionsContainer.Add(f.tcpOptionsBox)
		f.setTCPOptionsEnabled(false)
		f.dportEntry.Enable()
		f.dportEntry.SetPlaceHolder("포트")
	case "icmp":
		f.optionsContainer.Add(f.icmpOptionsBox)
		f.setICMPOptionsEnabled(true)
		// ICMP는 포트 개념이 없음
		f.dportEntry.Disable()
		f.dportEntry.SetText("")
		f.dportEntry.SetPlaceHolder("N/A")
	case "any":
		// ANY: TCP 옵션 박스 표시하되 비활성화
		f.optionsContainer.Add(f.tcpOptionsBox)
		f.setTCPOptionsEnabled(false)
		f.dportEntry.Enable()
		f.dportEntry.SetPlaceHolder("포트")
	default:
		f.optionsContainer.Add(f.tcpOptionsBox)
		f.setTCPOptionsEnabled(false)
		f.dportEntry.Enable()
		f.dportEntry.SetPlaceHolder("포트")
	}

	f.optionsContainer.Refresh()
}

// setTCPOptionsEnabled TCP 옵션 위젯들 활성/비활성화
func (f *RuleForm) setTCPOptionsEnabled(enabled bool) {
	if enabled {
		f.tcpFlagsPresetSel.Enable()
		for _, check := range f.tcpMaskChecks {
			check.Enable()
		}
		for _, check := range f.tcpSetChecks {
			check.Enable()
		}
	} else {
		f.tcpFlagsPresetSel.Disable()
		for _, check := range f.tcpMaskChecks {
			check.Disable()
		}
		for _, check := range f.tcpSetChecks {
			check.Disable()
		}
	}
}

// setICMPOptionsEnabled ICMP 옵션 위젯들 활성/비활성화
func (f *RuleForm) setICMPOptionsEnabled(enabled bool) {
	if enabled {
		f.icmpTypeSel.Enable()
		f.icmpCodeSel.Enable()
	} else {
		f.icmpTypeSel.Disable()
		f.icmpCodeSel.Disable()
	}
}

// onTCPPresetChanged TCP 프리셋 변경 시 체크박스 업데이트
func (f *RuleForm) onTCPPresetChanged(presetName string) {
	if presetName == "Custom" {
		return
	}

	f.updatingFromPreset = true
	defer func() { f.updatingFromPreset = false }()

	// 프리셋 찾기
	presets := model.GetTCPFlagsPresets()
	var preset *model.TCPFlagsPreset
	for i, p := range presets {
		if p.Name == presetName {
			preset = &presets[i]
			break
		}
	}

	if preset == nil {
		return
	}

	// 모든 체크박스 초기화
	for _, check := range f.tcpMaskChecks {
		check.SetChecked(false)
	}
	for _, check := range f.tcpSetChecks {
		check.SetChecked(false)
	}

	// 프리셋에 따라 체크박스 설정
	for _, flag := range preset.MaskFlags {
		if check, ok := f.tcpMaskChecks[flag]; ok {
			check.SetChecked(true)
		}
	}
	for _, flag := range preset.SetFlags {
		if check, ok := f.tcpSetChecks[flag]; ok {
			check.SetChecked(true)
		}
	}
}

// onICMPTypeChanged ICMP Type 변경 시 처리
func (f *RuleForm) onICMPTypeChanged(typeName string) {
	// Custom 옵션 제거됨 - Type Entry 항상 숨김
	f.icmpTypeEntry.Hide()
	f.icmpTypeEntry.SetText("")

	// Code 드롭다운은 항상 숨김 (smartfw에 ICMP_CODE 정의 없음)
	f.icmpCodeRow.Hide()
	f.icmpCodeSel.SetSelected("None")
	f.icmpCodeEntry.SetText("")
	f.icmpCodeEntry.Hide()

	f.icmpOptionsBox.Refresh()
}

// onICMPCodeChanged ICMP Code 변경 시 처리
func (f *RuleForm) onICMPCodeChanged(codeName string) {
	if codeName == "Custom..." {
		f.icmpCodeEntry.Show()
	} else {
		f.icmpCodeEntry.Hide()
		f.icmpCodeEntry.SetText("")
	}
}

// getTCPFlags 체크박스에서 TCP flags 문자열 생성
func (f *RuleForm) getTCPFlags() string {
	var maskFlags, setFlags []string

	flags := model.GetTCPFlagsList()
	for _, flag := range flags {
		if check, ok := f.tcpMaskChecks[flag]; ok && check.Checked {
			maskFlags = append(maskFlags, flag)
		}
		if check, ok := f.tcpSetChecks[flag]; ok && check.Checked {
			setFlags = append(setFlags, flag)
		}
	}

	if len(maskFlags) == 0 {
		return ""
	}

	return strings.Join(maskFlags, ",") + "/" + strings.Join(setFlags, ",")
}

// getICMPType ICMP type 값 가져오기
func (f *RuleForm) getICMPType() string {
	selected := f.icmpTypeSel.Selected
	if selected == "None" || selected == "" {
		return ""
	}
	return selected
}

// getICMPCode ICMP code 값 가져오기
func (f *RuleForm) getICMPCode() string {
	// Code 드롭다운이 숨겨져 있으면 빈 값 반환
	if !f.icmpCodeRow.Visible() {
		return ""
	}

	selected := f.icmpCodeSel.Selected
	if selected == "None" || selected == "" {
		return ""
	}

	if selected == "Custom..." {
		return strings.TrimSpace(f.icmpCodeEntry.Text)
	}

	// "net-unreachable (0)" 형식에서 이름만 추출
	parts := strings.Split(selected, " ")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
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
		Black:    false, // 일반 규칙은 Black/White 아님
		White:    false,
	}

	// 프로토콜 옵션 설정
	proto := strings.ToLower(f.protoSel.Selected)

	switch proto {
	case "tcp":
		tcpFlags := f.getTCPFlags()
		if tcpFlags != "" {
			rule.Options = &model.ProtocolOptions{TCPFlags: tcpFlags}
		}
	case "icmp":
		icmpType := f.getICMPType()
		icmpCode := f.getICMPCode()
		if icmpType != "" || icmpCode != "" {
			rule.Options = &model.ProtocolOptions{
				ICMPType: icmpType,
				ICMPCode: icmpCode,
			}
		}
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

	// TCP Flags 초기화
	f.tcpFlagsPresetSel.SetSelected("None")
	for _, check := range f.tcpMaskChecks {
		check.SetChecked(false)
	}
	for _, check := range f.tcpSetChecks {
		check.SetChecked(false)
	}

	// ICMP 옵션 초기화
	f.icmpTypeSel.SetSelected("None")
	f.icmpTypeEntry.SetText("")
	f.icmpTypeEntry.Hide()
	f.icmpCodeSel.SetSelected("None")
	f.icmpCodeEntry.SetText("")
	f.icmpCodeEntry.Hide()
	f.icmpCodeRow.Hide()

	// 프로토콜에 따른 옵션 UI 표시
	f.onProtocolChanged(f.protoSel.Selected)
}

// Content UI 컨테이너 반환
func (f *RuleForm) Content() *fyne.Container {
	return f.content
}
