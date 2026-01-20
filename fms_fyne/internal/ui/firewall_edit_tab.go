package ui

import (
	"fmt"
	"strings"

	"fms/internal/model"
	"fms/internal/parser"
	"fms/internal/storage"
	"fms/internal/themes"
	"fms/internal/ui/component"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// FirewallEditTab 방화벽 파일 편집 탭
type FirewallEditTab struct {
	window    fyne.Window
	fileStore *storage.FileStore
	mainUI    *MainUI
	file      *model.FirewallFile
	content   fyne.CanvasObject

	// UI 컴포넌트
	subTabs     *container.AppTabs // 목록보기/Firewall/NAT 서브탭
	textEditor  *widget.Entry      // 텍스트 편집기
	ruleBuilder *RuleBuilder       // 룰 빌더
	natBuilder  *NATBuilder        // NAT 규칙 빌더
	addBtn      fyne.CanvasObject  // 추가 버튼 (탭에 따라 표시/숨김)

	// 상태
	isDirty bool // 변경 여부
}

// NewFirewallEditTab 새로운 편집 탭을 생성합니다.
func NewFirewallEditTab(window fyne.Window, fileStore *storage.FileStore, mainUI *MainUI, file *model.FirewallFile) *FirewallEditTab {
	tab := &FirewallEditTab{
		window:    window,
		fileStore: fileStore,
		mainUI:    mainUI,
		file:      file,
		isDirty:   false,
	}
	tab.createUI()
	tab.loadContent()
	return tab
}

// Content 탭 컨텐츠를 반환합니다.
func (t *FirewallEditTab) Content() fyne.CanvasObject {
	return t.content
}

// GetTabName 탭 이름을 반환합니다. (truncation 처리)
func (t *FirewallEditTab) GetTabName() string {
	maxLen := 25
	name := "방화벽관리/" + t.file.FileName
	if len(name) > maxLen {
		return name[:maxLen-3] + "..."
	}
	return name
}

// GetFileName 파일명을 반환합니다.
func (t *FirewallEditTab) GetFileName() string {
	return t.file.FileName
}

// createUI UI를 생성합니다.
func (t *FirewallEditTab) createUI() {
	// 텍스트 편집기
	t.textEditor = widget.NewMultiLineEntry()
	t.textEditor.SetPlaceHolder("규칙 내용을 입력하세요...\n\n예시:\nagent -m=insert -c=INPUT -p=tcp --dport=9010 -a=DROP")
	t.textEditor.Wrapping = fyne.TextWrapOff
	t.textEditor.OnChanged = func(s string) {
		t.isDirty = true
	}

	// 룰 빌더
	t.ruleBuilder = NewRuleBuilder(t.window, nil)

	// NAT 규칙 빌더
	t.natBuilder = NewNATBuilder(t.window, nil)

	// 텍스트 편집 탭
	textEditTab := container.NewTabItem("목록보기", t.textEditor)

	// 룰 빌더 탭
	ruleBuilderTab := container.NewTabItem("Firewall", t.ruleBuilder.Content())

	// NAT 규칙 탭
	natBuilderTab := container.NewTabItem("NAT", t.natBuilder.Content())

	// 서브 탭 생성
	t.subTabs = container.NewAppTabs(textEditTab, ruleBuilderTab, natBuilderTab)
	t.subTabs.OnSelected = t.onSubTabChanged

	// 추가 버튼 (탭에 따라 표시/숨김)
	t.addBtn = component.NewCustomButton("+ 추가", nil, nil, themes.Colors["darkgray"], func() {
		t.onAddRule()
	}, 5, 5, 5, 5)
	t.addBtn.Hide() // 초기에는 목록보기 탭이므로 숨김

	// 저장 버튼
	saveBtn := component.NewCustomButton("저장", nil, nil, themes.Colors["blue"], func() {
		t.onSave()
	}, 5, 5, 5, 5)

	buttons := container.NewHBox(t.addBtn, saveBtn)

	// 헤더: 저장 버튼 (오른쪽 정렬)
	header := container.NewBorder(nil, nil, nil, buttons, nil)

	// 전체 레이아웃
	t.content = container.NewBorder(
		header,
		nil, nil, nil,
		t.subTabs,
	)
}

// loadContent 파일 내용을 로드합니다.
func (t *FirewallEditTab) loadContent() {
	t.textEditor.SetText(t.file.Contents)

	// 룰 빌더도 동기화 (필터 규칙)
	rules, comments, _ := parser.ParseTextToRules(t.file.Contents)
	t.ruleBuilder.SetRules(rules)
	t.ruleBuilder.SetComments(comments)

	// NAT 빌더도 동기화 (NAT 규칙)
	natRules, natComments, _ := parser.ParseTextToNATRules(t.file.Contents)
	t.natBuilder.SetRules(natRules)
	t.natBuilder.SetComments(natComments)

	t.isDirty = false
}

// onSubTabChanged 서브 탭 전환 시 호출
func (t *FirewallEditTab) onSubTabChanged(tab *container.TabItem) {
	switch tab.Text {
	case "Firewall":
		// 텍스트 -> 룰 빌더로 변환 (필터 규칙만)
		rules, comments, _ := parser.ParseTextToRules(t.textEditor.Text)
		t.ruleBuilder.SetRules(rules)
		t.ruleBuilder.SetComments(comments)
		t.addBtn.Show() // 추가 버튼 표시
	case "NAT":
		// 텍스트 -> NAT 빌더로 변환 (NAT 규칙만)
		natRules, comments, _ := parser.ParseTextToNATRules(t.textEditor.Text)
		t.natBuilder.SetRules(natRules)
		t.natBuilder.SetComments(comments)
		t.addBtn.Show() // 추가 버튼 표시
	case "목록보기":
		// 모든 빌더의 내용을 텍스트로 통합
		t.syncBuildersToText()
		t.addBtn.Hide() // 추가 버튼 숨김
	}
}

// syncBuildersToText 빌더 내용을 텍스트로 동기화
func (t *FirewallEditTab) syncBuildersToText() {
	// 필터 규칙 (주석 포함)
	filterRules := t.ruleBuilder.GetRules()
	filterComments := t.ruleBuilder.GetComments()
	filterText := parser.RulesToText(filterRules, filterComments)

	// NAT 규칙 (주석 제외 - 필터 규칙에서 이미 포함됨)
	natRules := t.natBuilder.GetRules()
	natText := parser.NATRulesToText(natRules, nil) // nil - 주석 중복 방지

	// 통합
	var finalText string
	if filterText != "" && natText != "" {
		finalText = filterText + "\n" + natText
	} else if filterText != "" {
		finalText = filterText
	} else {
		finalText = natText
	}

	t.textEditor.SetText(finalText)
}

// getCurrentContents 현재 활성 탭에서 내용 가져오기
func (t *FirewallEditTab) getCurrentContents() string {
	selectedTab := t.subTabs.Selected().Text

	switch selectedTab {
	case "Firewall":
		// 룰 빌더에서 텍스트로 변환 + NAT 규칙 포함
		t.syncBuildersToText()
		return t.textEditor.Text
	case "NAT":
		// NAT 빌더에서 텍스트로 변환 + 필터 규칙 포함
		t.syncBuildersToText()
		return t.textEditor.Text
	default:
		// 텍스트 편집기에서 직접 반환
		return t.textEditor.Text
	}
}

// onAddRule 규칙 추가 (현재 탭에 따라 다이얼로그 표시)
func (t *FirewallEditTab) onAddRule() {
	selectedTab := t.subTabs.Selected().Text
	switch selectedTab {
	case "Firewall":
		t.ruleBuilder.ShowAddDialog()
	case "NAT":
		t.natBuilder.ShowAddDialog()
	default:
		// 목록보기 탭에서는 아무 동작 안함
	}
}

// onSave 저장 버튼 클릭
func (t *FirewallEditTab) onSave() {
	contents := t.getCurrentContents()

	// 파일명 입력 다이얼로그
	fileNameEntry := widget.NewEntry()
	fileNameEntry.SetText(t.file.FileName)

	entryWidth := float32(300)
	rowHeight := float32(36)
	labelWidth := float32(80)
	rowSpacing := float32(20)
	buttonSpacing := float32(60)

	var popup *widget.PopUp

	// 헤더
	headerText := widget.NewRichTextFromMarkdown("## 파일 저장")

	// 폼 컨텐츠
	formContent := container.NewVBox(
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("파일명:")),
			container.NewGridWrap(fyne.NewSize(entryWidth, rowHeight), fileNameEntry),
		),
		container.NewGridWrap(fyne.NewSize(1, buttonSpacing), layout.NewSpacer()),
	)

	// 저장 처리 함수
	onSaveConfirm := func() {
		newFileName := strings.TrimSpace(fileNameEntry.Text)

		// 파일명 변경 확인
		if newFileName != t.file.FileName {
			// 새 파일명으로 저장
			if t.fileStore.FileExists(newFileName) {
				dialog.ShowError(fmt.Errorf("이미 존재하는 파일명입니다: %s", newFileName), t.window)
				return
			}

			// 기존 파일 삭제 후 새 파일 생성
			if err := t.fileStore.DeleteFile(t.file.FileName); err != nil {
				dialog.ShowError(err, t.window)
				return
			}

			// 파일명 업데이트
			oldFileName := t.file.FileName
			t.file.FileName = newFileName
			t.file.Version = model.ExtractVersion(newFileName)

			// MainUI에 탭 이름 변경 알림
			if t.mainUI != nil {
				t.mainUI.UpdateFirewallEditTabName(oldFileName, t.file)
			}
		}

		// 내용 저장
		t.file.Contents = contents
		if err := t.fileStore.SaveFile(t.file); err != nil {
			dialog.ShowError(err, t.window)
			return
		}

		t.isDirty = false
		popup.Hide()

		// 방화벽 관리 탭 새로고침
		if t.mainUI != nil && t.mainUI.firewallTab != nil {
			t.mainUI.firewallTab.RefreshFiles()
		}

		dialog.ShowInformation("저장 완료", "파일이 저장되었습니다.", t.window)
	}

	// 버튼
	cancelBtn := component.NewCustomButton("취소", nil, themes.Colors["black"], themes.Colors["lightgray"], func() {
		popup.Hide()
	}, 5, 5, 5, 5)
	saveBtn := component.NewCustomButton("저장", nil, nil, themes.Colors["blue"], onSaveConfirm, 5, 5, 5, 5)

	// 버튼 컨테이너
	btnContainer := container.NewHBox(
		layout.NewSpacer(),
		cancelBtn,
		container.NewGridWrap(fyne.NewSize(60, 1), layout.NewSpacer()),
		saveBtn,
		layout.NewSpacer(),
	)

	// 전체 컨텐츠
	content := container.NewVBox(
		headerText,
		container.NewGridWrap(fyne.NewSize(1, rowSpacing), layout.NewSpacer()),
		formContent,
		btnContainer,
		container.NewGridWrap(fyne.NewSize(1, 20), layout.NewSpacer()),
	)

	// 고정 크기 컨테이너
	paddedContent := container.New(layout.NewCustomPaddedLayout(20, 20, 20, 20), content)
	sizedContent := container.NewGridWrap(fyne.NewSize(450, 220), paddedContent)

	// 팝업 생성
	popup = widget.NewModalPopUp(sizedContent, t.window.Canvas())
	popup.Show()
}

// IsDirty 변경 여부를 반환합니다.
func (t *FirewallEditTab) IsDirty() bool {
	return t.isDirty
}

// ResetTabs 모든 탭 위치를 첫 번째 탭으로 초기화
func (t *FirewallEditTab) ResetTabs() {
	if len(t.subTabs.Items) > 0 {
		t.subTabs.SelectIndex(0)
	}
	t.ruleBuilder.ResetTabs()
	t.natBuilder.ResetTabs()
}
