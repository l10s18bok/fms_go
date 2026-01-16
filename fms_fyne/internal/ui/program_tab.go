package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"fms/internal/model"
	"fms/internal/storage"
	"fms/internal/themes"
	"fms/internal/ui/component"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	fynestorage "fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

// ProgramTab 프로그램 관리 탭
type ProgramTab struct {
	window  fyne.Window
	store   *storage.JSONStore
	content fyne.CanvasObject

	// UI 컴포넌트
	searchBox    *component.SearchBox // 검색 컴포넌트 (공통)
	programTable *component.PagedTable

	// 데이터
	programs         []*model.ProcessInfo
	filteredPrograms []*model.ProcessInfo
}

// NewProgramTab 새로운 프로그램 탭을 생성합니다.
func NewProgramTab(window fyne.Window, store *storage.JSONStore) *ProgramTab {
	tab := &ProgramTab{
		window:   window,
		store:    store,
		programs: []*model.ProcessInfo{},
	}
	tab.createUI()
	tab.loadPrograms()
	return tab
}

// Content 탭 컨텐츠를 반환합니다.
func (t *ProgramTab) Content() fyne.CanvasObject {
	return t.content
}

// createUI UI를 생성합니다.
func (t *ProgramTab) createUI() {
	// 검색 컴포넌트 (공통)
	t.searchBox = component.NewSearchBox(component.SearchBoxConfig{
		Placeholder: "검색...",
		Width:       200,
		OnSearch: func(text string) {
			t.filterPrograms(text)
		},
	})

	// 삭제 버튼
	deleteBtn := component.NewCustomButton("삭제", nil, themes.Colors["red"], nil, func() {
		t.onDeleteSelected()
	})

	// 추가/수정 버튼
	addEditBtn := component.NewCustomButton("추가/수정", nil, nil, themes.Colors["blue"], func() {
		t.onAddOrEdit()
	})

	// 상단 영역 (버튼 간격 추가)
	topBarLine := container.NewBorder(
		nil, nil,
		t.searchBox.Content(),
		container.NewHBox(
			deleteBtn,
			widget.NewLabel("  "), // 버튼 간격
			addEditBtn,
		),
		nil,
	)

	// 상단 영역에 상하좌우 10px 패딩 적용
	topBar := container.New(layout.NewCustomPaddedLayout(10, 10, 10, 10), topBarLine)

	// 테이블 설정
	t.programTable = component.NewPagedTable(component.PagedTableConfig{
		Columns: []component.ColumnDef{
			{Header: "선택", Width: 50},
			{Header: "이름", Width: 150},
			{Header: "버전", Width: 100},
			{Header: "업로드 경로", Width: 225},
			{Header: "로컬파일 경로", Width: 375},
			{Header: "추가(수정)시간", Width: 150},
		},
		PageSize: 10,
		OnCellUpdate: func(row int, col int, cell fyne.CanvasObject) {
			t.updateCell(row, col, cell)
		},
		OnRowDoubleClick: func(row int) {
			t.onRowDoubleClick(row)
		},
	})

	// 메인 컨텐츠
	t.content = container.NewBorder(
		topBar,
		nil, nil, nil,
		t.programTable.Content(),
	)
}

// loadPrograms 프로그램 목록을 로드합니다.
func (t *ProgramTab) loadPrograms() {
	programs, err := t.store.GetAllPrograms()
	if err != nil {
		dialog.ShowError(err, t.window)
		return
	}

	// ID 내림차순 정렬 (최신순 - ID가 클수록 최신)
	sort.Slice(programs, func(i, j int) bool {
		return programs[i].ID > programs[j].ID
	})

	t.programs = programs
	t.filteredPrograms = programs
	t.programTable.SetData(len(t.filteredPrograms))
}

// RefreshPrograms 프로그램 목록을 새로고침합니다.
func (t *ProgramTab) RefreshPrograms() {
	t.loadPrograms()
}

// filterPrograms 검색어로 프로그램을 필터링합니다.
func (t *ProgramTab) filterPrograms(query string) {
	if query == "" {
		t.filteredPrograms = t.programs
	} else {
		query = strings.ToLower(query)
		filtered := []*model.ProcessInfo{}
		for _, p := range t.programs {
			if strings.Contains(strings.ToLower(p.ProcessName), query) ||
				strings.Contains(strings.ToLower(p.ProcessVersion), query) ||
				strings.Contains(strings.ToLower(p.ProcessUploadPath), query) ||
				strings.Contains(strings.ToLower(p.ProcessFilePath), query) {
				filtered = append(filtered, p)
			}
		}
		t.filteredPrograms = filtered
		if len(filtered) == 0 {
			dialog.ShowInformation("검색 결과", "검색 결과가 없습니다.", t.window)
		}
	}
	t.programTable.SetData(len(t.filteredPrograms))
}

// updateCell 테이블 셀을 업데이트합니다.
func (t *ProgramTab) updateCell(row int, col int, cell fyne.CanvasObject) {
	if row >= len(t.filteredPrograms) {
		return
	}
	p := t.filteredPrograms[row]

	switch col {
	case 0: // 선택 (체크박스)
		// PagedTable에서 처리
	case 1: // 이름
		if label, ok := cell.(*widget.Label); ok {
			label.SetText(p.ProcessName)
		}
	case 2: // 버전
		if label, ok := cell.(*widget.Label); ok {
			label.SetText(p.ProcessVersion)
		}
	case 3: // 업로드 경로
		if label, ok := cell.(*widget.Label); ok {
			label.SetText(p.ProcessUploadPath)
		}
	case 4: // 로컬파일 경로
		if label, ok := cell.(*widget.Label); ok {
			label.SetText(p.ProcessFilePath)
		}
	case 5: // 추가(수정)시간
		if label, ok := cell.(*widget.Label); ok {
			label.SetText(p.ProcessCreatedAt)
		}
	}
}

// onRowDoubleClick 행 더블클릭 시 수정 다이얼로그를 표시합니다.
func (t *ProgramTab) onRowDoubleClick(row int) {
	if row >= len(t.filteredPrograms) {
		return
	}
	t.showEditDialog(t.filteredPrograms[row])
}

// onDeleteSelected 선택된 프로그램을 삭제합니다.
func (t *ProgramTab) onDeleteSelected() {
	checkedRows := t.programTable.GetCheckedRows()
	if len(checkedRows) == 0 {
		return
	}

	dialog.ShowConfirm("삭제 확인",
		fmt.Sprintf("선택한 %d개 프로그램을 삭제하시겠습니까?", len(checkedRows)),
		func(ok bool) {
			if !ok {
				return
			}

			// 삭제 실행
			for _, row := range checkedRows {
				if row < len(t.filteredPrograms) {
					p := t.filteredPrograms[row]
					if err := t.store.DeleteProgram(p.ID); err != nil {
						dialog.ShowError(err, t.window)
						return
					}
				}
			}

			t.loadPrograms()
			dialog.ShowInformation("삭제 완료", fmt.Sprintf("%d개 프로그램이 삭제되었습니다.", len(checkedRows)), t.window)
		}, t.window)
}

// onAddOrEdit 추가/수정 버튼 클릭 처리
func (t *ProgramTab) onAddOrEdit() {
	checkedRows := t.programTable.GetCheckedRows()
	if len(checkedRows) == 0 {
		// 추가 모드
		t.showEditDialog(nil)
	} else if len(checkedRows) == 1 {
		// 수정 모드
		if checkedRows[0] < len(t.filteredPrograms) {
			t.showEditDialog(t.filteredPrograms[checkedRows[0]])
		}
	} else {
		// 다중 선택 시 첫 번째만 수정
		if checkedRows[0] < len(t.filteredPrograms) {
			t.showEditDialog(t.filteredPrograms[checkedRows[0]])
		}
	}
}

// showEditDialog 프로그램 추가/수정 다이얼로그를 표시합니다.
func (t *ProgramTab) showEditDialog(program *model.ProcessInfo) {
	isEdit := program != nil

	// 입력 필드 생성
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("프로그램 이름")

	versionEntry := widget.NewEntry()
	versionEntry.SetPlaceHolder("버전 (예: v1.0.0)")

	uploadPathEntry := widget.NewEntry()
	uploadPathEntry.SetPlaceHolder("/download/")
	uploadPathEntry.SetText("/download/")

	filePathLabel := widget.NewLabel("파일을 선택해주세요")
	filePath := ""

	// 수정 모드일 경우 기존 값 설정
	if isEdit {
		nameEntry.SetText(program.ProcessName)
		versionEntry.SetText(program.ProcessVersion)
		uploadPathEntry.SetText(program.ProcessUploadPath)
		filePath = program.ProcessFilePath
		filePathLabel.SetText(filePath)
	}

	// 다이얼로그 참조 (Hide/Show 패턴용)
	var parentDialog dialog.Dialog

	// 찾아보기 버튼
	browseBtn := widget.NewButton("찾아보기...", nil)

	// 파일 선택 행
	fileRow := container.NewBorder(nil, nil, nil, browseBtn, filePathLabel)

	// 폼 컨텐츠
	formContent := container.NewVBox(
		widget.NewLabel("이름:"),
		nameEntry,
		widget.NewLabel("버전:"),
		versionEntry,
		widget.NewLabel("업로드 경로:"),
		uploadPathEntry,
		widget.NewLabel("로컬 파일:"),
		fileRow,
	)

	// 저장 버튼
	saveBtn := widget.NewButton("저장", func() {
		// 유효성 검사
		if nameEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("프로그램 이름을 입력해주세요"), t.window)
			return
		}
		if versionEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("버전을 입력해주세요"), t.window)
			return
		}
		if filePath == "" {
			dialog.ShowError(fmt.Errorf("파일을 선택해주세요"), t.window)
			return
		}

		// 저장
		var p *model.ProcessInfo
		if isEdit {
			p = program
		} else {
			p = model.NewProcessInfo(nameEntry.Text, versionEntry.Text)
		}
		p.ProcessName = nameEntry.Text
		p.ProcessVersion = versionEntry.Text
		p.ProcessUploadPath = uploadPathEntry.Text
		p.ProcessFilePath = filePath
		p.ProcessCreatedAt = time.Now().Format("2006-01-02 15:04:05")

		if err := t.store.SaveProgram(p); err != nil {
			dialog.ShowError(err, t.window)
			return
		}

		parentDialog.Hide()
		t.loadPrograms()
		if isEdit {
			dialog.ShowInformation("수정 완료", "프로그램이 수정되었습니다.", t.window)
		} else {
			dialog.ShowInformation("추가 완료", "프로그램이 추가되었습니다.", t.window)
		}
	})

	// 취소 버튼
	cancelBtn := widget.NewButton("취소", func() {
		parentDialog.Hide()
	})

	// 버튼 영역
	buttons := container.NewHBox(
		cancelBtn,
		saveBtn,
	)

	// 다이얼로그 컨텐츠
	dialogContent := container.NewBorder(
		nil,
		container.NewCenter(buttons),
		nil, nil,
		formContent,
	)

	// 다이얼로그 제목
	title := "프로그램 추가"
	if isEdit {
		title = "프로그램 수정"
	}

	// 다이얼로그 생성
	parentDialog = dialog.NewCustomWithoutButtons(title, dialogContent, t.window)
	parentDialog.Resize(fyne.NewSize(500, 350))

	// 찾아보기 버튼 동작 설정 (Hide/Show 패턴)
	browseBtn.OnTapped = func() {
		parentDialog.Hide() // 부모 다이얼로그 숨김

		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if reader != nil {
				filePath = reader.URI().Path()
				filePathLabel.SetText(filePath)
				reader.Close()
			}
			parentDialog.Show() // 부모 다이얼로그 다시 표시
		}, t.window)
		fileDialog.SetFilter(fynestorage.NewExtensionFileFilter([]string{".tar", ".tar.gz", ".tgz", ".zip"}))
		fileDialog.Show()
	}

	parentDialog.Show()
}

// GetAllPrograms 모든 프로그램 목록을 반환합니다.
func (t *ProgramTab) GetAllPrograms() []*model.ProcessInfo {
	return t.programs
}
