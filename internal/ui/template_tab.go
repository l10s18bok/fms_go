package ui

import (
	"sort"

	"fms/internal/model"
	"fms/internal/storage"
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

	// 우측: 템플릿 내용 패널
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

	// 패널 레이아웃
	return container.NewBorder(
		widget.NewLabel("템플릿 목록"), // 상단 제목
		nil,                       // 하단 버튼 삭제
		nil, nil,
		listContainer, // 중앙 목록
	)
}

// 템플릿 내용 편집 패널을 생성합니다.
func (t *TemplateTab) createTemplateContentPanel() fyne.CanvasObject {
	// 템플릿 내용 편집기 (여러 줄)
	t.templateContent = widget.NewMultiLineEntry()
	t.templateContent.SetPlaceHolder("템플릿 내용을 입력하세요...\n\n예시:\nreq|INSERT|3813792919|INPUT|FLUSH|ANY|ANY|ANY|||\nreq|INSERT|3813792919|INPUT|ACCEPT|TCP|192.168.1.0/24|ANY|80||")
	t.templateContent.Wrapping = fyne.TextWrapOff

	// Clear 버튼
	clearBtn := widget.NewButton("Clear", func() {
		t.templateContent.SetText("")
		t.selectedVersion = ""
		t.templateList.SetSelected("")
	})

	// 제목과 Clear 버튼을 포함한 헤더
	header := container.NewBorder(nil, nil, widget.NewLabel("템플릿 내용"), clearBtn, nil)

	// 제목과 함께 반환
	return container.NewBorder(
		header,
		nil, nil, nil,
		t.templateContent,
	)
}

// 하단 버튼/입력 영역을 생성합니다.
func (t *TemplateTab) createBottomPanel() fyne.CanvasObject {
	// 앱 버전 표시
	appVersionLabel := widget.NewLabel(version.GetVersionString())

	// 버튼들 (컴포넌트 사용)
	saveBtn := component.NewIconTextButton("저장", theme.ConfirmIcon(), component.ButtonPrimary, func() {
		t.onSaveTemplate()
	})

	deleteBtn := component.NewIconTextButton("삭제", theme.DeleteIcon(), component.ButtonDanger, func() {
		t.onDeleteTemplate()
	})

	// 버튼 그룹 (간격 추가)
	spacer := widget.NewLabel("  ")
	buttons := container.NewHBox(saveBtn, spacer, deleteBtn)

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

	// 선택된 템플릿 찾기
	for _, tmpl := range t.templates {
		if tmpl.Version == version {
			t.templateContent.SetText(tmpl.Contents)
			return
		}
	}
}

// 템플릿 저장 시 호출됩니다.
func (t *TemplateTab) onSaveTemplate() {
	contents := t.templateContent.Text

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
		t.loadTemplates()
	}, t.window)
}

