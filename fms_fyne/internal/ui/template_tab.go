package ui

import (
	"sort"

	"fms/internal/model"
	"fms/internal/parser"
	"fms/internal/storage"
	"fms/internal/themes"
	"fms/internal/ui/component"
	"fms/internal/version"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// 템플릿 관리 탭을 구현합니다.
type TemplateTab struct {
	window  fyne.Window
	store   *storage.JSONStore
	content fyne.CanvasObject

	// UI 컴포넌트
	templateList    *widget.RadioGroup // 템플릿 목록 (라디오 버튼)
	templateContent *widget.Entry      // 템플릿 내용 편집기
	ruleBuilder     *RuleBuilder       // 규칙 빌더
	natBuilder      *NATBuilder        // NAT 규칙 빌더
	subTabs         *container.AppTabs // 서브 탭 (텍스트 편집 / 규칙 빌더 / NAT 규칙)

	// 데이터
	templates       []*model.Template
	selectedVersion string
}

// 새로운 템플릿 관리 탭을 생성합니다.
func NewTemplateTab(window fyne.Window, store *storage.JSONStore) *TemplateTab {
	tab := &TemplateTab{
		window:    window,
		store:     store,
		templates: []*model.Template{},
	}
	tab.createUI()
	tab.loadTemplates()
	return tab
}

// 템플릿 탭의 UI를 생성합니다.
func (t *TemplateTab) createUI() {
	// 좌측: 템플릿 목록 패널
	leftPanel := t.createTemplateListPanel()

	// 우측: 템플릿 내용 패널 (서브 탭 포함)
	rightPanel := t.createTemplateContentPanel()

	// 좌우 분할 (25% : 75%)
	split := container.NewHSplit(leftPanel, rightPanel)
	split.Offset = 0.25

	// 하단 버튼 및 입력 영역
	bottomPanel := t.createBottomPanel()

	// 전체 레이아웃
	t.content = container.NewBorder(
		nil,         // 상단
		bottomPanel, // 하단 고정
		nil,         // 좌측
		nil,         // 우측
		split,       // 중앙 (자동 확장)
	)
}

// 템플릿 목록 패널을 생성합니다.
func (t *TemplateTab) createTemplateListPanel() fyne.CanvasObject {
	// 템플릿 목록 라디오 그룹 (빈 상태로 시작)
	t.templateList = widget.NewRadioGroup([]string{}, func(selected string) {
		t.selectedVersion = selected
		t.onTemplateSelected(selected)
	})

	// 목록 영역 (스크롤 가능)
	listContainer := container.NewVScroll(t.templateList)

	// 새 템플릿 버튼
	newBtn := component.NewCustomButton("+ 새 템플릿", nil, themes.Colors["black"], themes.Colors["lightgray"], func() {
		t.onNewTemplate()
	}, 5, 0, 0, 5)

	// 헤더: "템플릿 목록" + 오른쪽에 새 템플릿 버튼
	header := container.NewBorder(nil, nil,
		widget.NewLabel("템플릿 목록"),
		newBtn,
	)

	// 패널 레이아웃
	return container.NewBorder(
		header, // 상단 헤더
		nil,    // 하단
		nil, nil,
		listContainer, // 중앙 목록
	)
}

// 템플릿 내용 편집 패널을 생성합니다.
func (t *TemplateTab) createTemplateContentPanel() fyne.CanvasObject {
	// 텍스트 편집기
	t.templateContent = widget.NewMultiLineEntry()
	t.templateContent.SetPlaceHolder("템플릿 내용을 입력하세요...\n\n예시:\nagent -m=insert -c=INPUT -p=tcp --dport=9010 -a=DROP")
	t.templateContent.Wrapping = fyne.TextWrapOff

	// 규칙 빌더
	t.ruleBuilder = NewRuleBuilder(nil)

	// NAT 규칙 빌더
	t.natBuilder = NewNATBuilder(nil)

	// 텍스트 편집 탭
	textEditTab := container.NewTabItem("텍스트 편집", t.templateContent)

	// 규칙 빌더 탭
	ruleBuilderTab := container.NewTabItem("규칙 빌더", t.ruleBuilder.Content())

	// NAT 규칙 탭
	natBuilderTab := container.NewTabItem("NAT 규칙", t.natBuilder.Content())

	// 서브 탭 생성
	t.subTabs = container.NewAppTabs(textEditTab, ruleBuilderTab, natBuilderTab)
	t.subTabs.OnSelected = t.onSubTabChanged

	// 저장, 삭제 버튼
	saveBtn := component.NewCustomButton("저장", theme.ConfirmIcon(), nil, themes.Colors["blue"], func() {
		t.onSaveTemplate()
	}, 5, 5, 5, 5)
	deleteBtn := component.NewCustomButton("삭제", theme.DeleteIcon(), nil, themes.Colors["red"], func() {
		t.onDeleteTemplate()
	}, 5, 5, 5, 5)
	buttons := container.NewHBox(saveBtn, deleteBtn)

	// 헤더: "템플릿 내용" + 저장/삭제 버튼
	header := container.NewBorder(nil, nil, widget.NewLabel("템플릿 내용"), buttons, nil)

	// 제목과 함께 반환
	return container.NewBorder(
		header,
		nil, nil, nil,
		t.subTabs,
	)
}

// onSubTabChanged 서브 탭 전환 시 호출
func (t *TemplateTab) onSubTabChanged(tab *container.TabItem) {
	switch tab.Text {
	case "규칙 빌더":
		// 텍스트 -> 규칙 빌더로 변환 (필터 규칙만)
		rules, comments, _ := parser.ParseTextToRules(t.templateContent.Text)
		t.ruleBuilder.SetRules(rules)
		t.ruleBuilder.SetComments(comments)
	case "NAT 규칙":
		// 텍스트 -> NAT 빌더로 변환 (NAT 규칙만)
		natRules, comments, _ := parser.ParseTextToNATRules(t.templateContent.Text)
		t.natBuilder.SetRules(natRules)
		t.natBuilder.SetComments(comments)
	case "텍스트 편집":
		// 모든 빌더의 내용을 텍스트로 통합
		t.syncBuildersToText()
	}
}

// syncBuildersToText 빌더 내용을 텍스트로 동기화
func (t *TemplateTab) syncBuildersToText() {
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

	t.templateContent.SetText(finalText)
}

// 하단 버튼/입력 영역을 생성합니다.
func (t *TemplateTab) createBottomPanel() fyne.CanvasObject {
	// 앱 버전 표시
	appVersionLabel := widget.NewLabel(version.GetVersionString())

	// 버튼 그룹 없음 (저장/삭제는 규칙 빌더로 이동)
	buttons := container.NewHBox()

	// 하단 레이아웃
	return container.NewVBox(
		widget.NewSeparator(),
		container.NewBorder(nil, nil, appVersionLabel, buttons, nil),
	)
}

// 탭의 컨텐츠를 반환합니다.
func (t *TemplateTab) Content() fyne.CanvasObject {
	return t.content
}

// 저장소에서 템플릿 목록을 로드합니다.
func (t *TemplateTab) loadTemplates() {
	templates, err := t.store.GetAllTemplates()
	if err != nil {
		dialog.ShowError(err, t.window)
		return
	}

	t.templates = templates

	// 버전명으로 정렬
	sort.Slice(t.templates, func(i, j int) bool {
		return t.templates[i].Version > t.templates[j].Version // 내림차순
	})

	// 라디오 그룹 업데이트
	options := make([]string, len(t.templates))
	for i, tmpl := range t.templates {
		options[i] = tmpl.Version
	}
	t.templateList.Options = options

	// 선택 상태 유지 (기존 선택이 없으면 선택 해제)
	if t.selectedVersion == "" {
		t.templateList.SetSelected("")
		t.templateContent.SetText("")
		t.ruleBuilder.Clear()
		t.natBuilder.Clear()
	}
	t.templateList.Refresh()
}

// 템플릿 목록을 새로고침합니다. (외부에서 호출 가능)
func (t *TemplateTab) RefreshTemplates() {
	t.loadTemplates()
}

// 템플릿 선택 상태와 내용을 초기화합니다. (Reset 시 호출)
func (t *TemplateTab) ClearSelection() {
	t.selectedVersion = ""
	t.templateList.SetSelected("")
	t.templateContent.SetText("")
	t.ruleBuilder.Clear()
	t.natBuilder.Clear()
}

// 모든 템플릿 버전 목록을 반환합니다.
func (t *TemplateTab) GetTemplateVersions() []string {
	versions := make([]string, len(t.templates))
	for i, tmpl := range t.templates {
		versions[i] = tmpl.Version
	}
	return versions
}

// 특정 버전의 템플릿을 반환합니다.
func (t *TemplateTab) GetTemplate(version string) *model.Template {
	for _, tmpl := range t.templates {
		if tmpl.Version == version {
			return tmpl
		}
	}
	return nil
}

// 템플릿 선택 시 호출됩니다.
func (t *TemplateTab) onTemplateSelected(version string) {
	if version == "" {
		return
	}

	// 모든 탭 위치 초기화
	t.resetAllTabs()

	// 선택된 템플릿 찾기
	for _, tmpl := range t.templates {
		if tmpl.Version == version {
			t.templateContent.SetText(tmpl.Contents)

			// 규칙 빌더도 동기화 (필터 규칙)
			rules, comments, _ := parser.ParseTextToRules(tmpl.Contents)
			t.ruleBuilder.SetRules(rules)
			t.ruleBuilder.SetComments(comments)

			// NAT 빌더도 동기화 (NAT 규칙)
			natRules, natComments, _ := parser.ParseTextToNATRules(tmpl.Contents)
			t.natBuilder.SetRules(natRules)
			t.natBuilder.SetComments(natComments)
			return
		}
	}
}

// onNewTemplate 새 템플릿 생성
func (t *TemplateTab) onNewTemplate() {
	// 선택 해제
	t.templateList.SetSelected("")
	t.selectedVersion = ""

	// 내용 초기화
	t.templateContent.SetText("")
	t.ruleBuilder.Clear()
	t.natBuilder.Clear()

	// 탭 위치 초기화
	t.resetAllTabs()
}

// resetAllTabs 모든 탭 위치를 첫 번째 탭으로 초기화
func (t *TemplateTab) resetAllTabs() {
	// 서브 탭 (텍스트 편집 / 규칙 빌더 / NAT 규칙) 초기화
	if len(t.subTabs.Items) > 0 {
		t.subTabs.SelectIndex(0)
	}

	// 규칙 빌더 내부 폼 탭 초기화
	t.ruleBuilder.ResetTabs()

	// NAT 빌더 내부 폼 탭 초기화
	t.natBuilder.ResetTabs()
}

// getCurrentContents 현재 활성 탭에서 내용 가져오기
func (t *TemplateTab) getCurrentContents() string {
	selectedTab := t.subTabs.Selected().Text

	switch selectedTab {
	case "규칙 빌더":
		// 규칙 빌더에서 텍스트로 변환 + NAT 규칙 포함
		t.syncBuildersToText()
		return t.templateContent.Text
	case "NAT 규칙":
		// NAT 빌더에서 텍스트로 변환 + 필터 규칙 포함
		t.syncBuildersToText()
		return t.templateContent.Text
	default:
		// 텍스트 편집기에서 직접 반환
		return t.templateContent.Text
	}
}

// 템플릿 저장 시 호출됩니다.
func (t *TemplateTab) onSaveTemplate() {
	contents := t.getCurrentContents()

	if contents == "" {
		dialog.ShowInformation("알림", "템플릿 내용을 입력해주세요.", t.window)
		return
	}

	// 버전명 입력 다이얼로그
	versionEntry := widget.NewEntry()
	versionEntry.SetPlaceHolder("예: v1.0")

	// 기존 선택된 버전이 있으면 기본값으로 설정
	if t.selectedVersion != "" {
		versionEntry.SetText(t.selectedVersion)
	}

	formItems := []*widget.FormItem{
		widget.NewFormItem("버전명", versionEntry),
	}

	dialog.ShowForm("템플릿 저장", "저장", "취소", formItems, func(ok bool) {
		if !ok {
			return
		}

		version := versionEntry.Text
		if version == "" {
			dialog.ShowInformation("알림", "버전명을 입력해주세요.", t.window)
			return
		}

		template := &model.Template{
			Version:  version,
			Contents: contents,
		}

		if err := t.store.SaveTemplate(template); err != nil {
			dialog.ShowError(err, t.window)
			return
		}

		t.loadTemplates()
		t.templateList.SetSelected(version)
		dialog.ShowInformation("알림", "템플릿이 저장되었습니다.", t.window)
	}, t.window)
}

// 템플릿 삭제 시 호출됩니다.
func (t *TemplateTab) onDeleteTemplate() {
	if t.selectedVersion == "" {
		dialog.ShowInformation("알림", "삭제할 템플릿을 선택해주세요.", t.window)
		return
	}

	dialog.ShowConfirm("확인", "선택한 템플릿을 삭제하시겠습니까?", func(ok bool) {
		if !ok {
			return
		}

		if err := t.store.DeleteTemplate(t.selectedVersion); err != nil {
			dialog.ShowError(err, t.window)
			return
		}

		// 선택 초기화
		t.templateList.SetSelected("")
		t.selectedVersion = ""
		t.templateContent.SetText("")
		t.ruleBuilder.Clear()
		t.natBuilder.Clear()
		t.loadTemplates()
		dialog.ShowInformation("알림", "템플릿이 삭제되었습니다.", t.window)
	}, t.window)
}
