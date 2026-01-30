package ui

import (
	"fmt"
	"strings"
	"time"

	"fms/internal/model"
	"fms/internal/repository"
	"fms/internal/themes"
	"fms/internal/ui/component"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	fynestorage "fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ProgramTab 패키지 관리 탭
type ProgramTab struct {
	window      fyne.Window
	programRepo repository.ProgramRepository
	content     fyne.CanvasObject

	// UI 컴포넌트
	searchBox    *component.SearchBox // 검색 컴포넌트 (공통)
	programTable *component.PagedTable

	// 데이터 (DB 페이지네이션)
	pageData      []*model.ProcessInfo // 현재 페이지 데이터
	totalCount    int                  // 검색 적용 후 전체 건수
	searchKeyword string               // 검색 키워드
}

// NewProgramTab 새로운 패키지 탭을 생성합니다.
func NewProgramTab(window fyne.Window, programRepo repository.ProgramRepository) *ProgramTab {
	tab := &ProgramTab{
		window:      window,
		programRepo: programRepo,
		pageData:    []*model.ProcessInfo{},
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

	// 삭제 버튼 (빨간 배경, 흰색 텍스트)
	deleteBtn := component.NewCustomButton("삭제", theme.DeleteIcon(), nil, themes.Colors["red"], func() {
		t.onDeleteSelected()
	})

	// 추가/수정 버튼
	addEditBtn := component.NewCustomButton("+추가/수정", nil, nil, themes.Colors["blue"], func() {
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
	t.programTable = component.NewPagedTableWithWindow(component.PagedTableConfig{
		Columns: []component.ColumnDef{
			{Header: "선택", Width: 50},
			{Header: "패키지명", Width: 150},
			{Header: "버전", Width: 100},
			{Header: "로컬파일 경로", Width: 500},
			{Header: "추가(수정)시간", Width: 150},
		},
		PageSize: 15,
		OnCellUpdate: func(row int, col int, cell fyne.CanvasObject) {
			t.updateCell(row, col, cell)
		},
		OnRowDoubleClick: func(row int) {
			t.onRowDoubleClick(row)
		},
		// 패키지명(1), 버전(2) 컬럼 인라인 편집 설정
		EditableColumns: map[int]component.EditColumnConfig{
			1: { // 패키지명 컬럼
				Type: component.EditTypeEntry,
				GetValue: func(row int) string {
					if row >= 0 && row < len(t.pageData) {
						return t.pageData[row].ProcessName
					}
					return ""
				},
				OnEdit: func(row int, oldValue, newValue string) bool {
					if row >= 0 && row < len(t.pageData) {
						if newValue == "" {
							dialog.ShowError(fmt.Errorf("패키지명을 입력해주세요"), t.window)
							return false
						}
						p := t.pageData[row]
						p.ProcessName = newValue
						p.ProcessCreatedAt = time.Now().Format("2006-01-02 15:04:05")
						if err := t.programRepo.Save(p); err != nil {
							dialog.ShowError(err, t.window)
							return false
						}
						return true
					}
					return false
				},
			},
			2: { // 버전 컬럼
				Type: component.EditTypeEntry,
				GetValue: func(row int) string {
					if row >= 0 && row < len(t.pageData) {
						return t.pageData[row].ProcessVersion
					}
					return ""
				},
				OnEdit: func(row int, oldValue, newValue string) bool {
					if row >= 0 && row < len(t.pageData) {
						if newValue == "" {
							dialog.ShowError(fmt.Errorf("버전을 입력해주세요"), t.window)
							return false
						}
						p := t.pageData[row]
						p.ProcessVersion = newValue
						p.ProcessCreatedAt = time.Now().Format("2006-01-02 15:04:05")
						if err := t.programRepo.Save(p); err != nil {
							dialog.ShowError(err, t.window)
							return false
						}
						return true
					}
					return false
				},
			},
		},
		OnPageLoad: func(page, pageSize int) int {
			return t.loadPage(page, pageSize)
		},
	}, t.window)

	// 메인 컨텐츠
	t.content = container.NewBorder(
		topBar,
		nil, nil, nil,
		t.programTable.Content(),
	)
}

// loadPrograms 패키지 목록을 로드합니다.
func (t *ProgramTab) loadPrograms() {
	total := t.loadPage(0, t.programTable.GetPageSize())
	if t.programTable != nil {
		t.programTable.SetData(total)
	}
}

// RefreshPrograms 패키지 목록을 새로고침합니다.
func (t *ProgramTab) RefreshPrograms() {
	t.loadPrograms()
}

// DB에서 해당 페이지 데이터를 조회합니다.
func (t *ProgramTab) loadPage(page, pageSize int) int {
	req := model.PageRequest{
		Offset:  page * pageSize,
		Limit:   pageSize,
		Keyword: t.searchKeyword,
	}
	result, err := t.programRepo.GetPage(req)
	if err != nil {
		t.pageData = []*model.ProcessInfo{}
		t.totalCount = 0
		return 0
	}
	t.pageData = result.Items
	t.totalCount = result.TotalCount
	return result.TotalCount
}

// filterPrograms 검색어로 패키지를 필터링합니다.
func (t *ProgramTab) filterPrograms(query string) {
	keyword := strings.TrimSpace(query)

	prevKeyword := t.searchKeyword
	t.searchKeyword = keyword
	total := t.loadPage(0, t.programTable.GetPageSize())

	// 검색 결과가 없으면 이전 상태 유지
	if total == 0 && keyword != "" {
		dialog.ShowInformation("검색 결과", "검색 결과가 없습니다.", t.window)
		t.searchKeyword = prevKeyword
		t.loadPage(0, t.programTable.GetPageSize())
	}

	if t.programTable != nil {
		t.programTable.SetData(t.totalCount)
	}
}

// GetExportData 현재 검색 조건에 맞는 전체 패키지를 반환합니다.
func (t *ProgramTab) GetExportData() ([]*model.ProcessInfo, error) {
	if t.searchKeyword == "" {
		return t.programRepo.GetAll()
	}
	req := model.PageRequest{
		Offset:  0,
		Limit:   100000,
		Keyword: t.searchKeyword,
	}
	result, err := t.programRepo.GetPage(req)
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

// IsFiltered 검색 필터가 적용되어 있는지 반환합니다.
func (t *ProgramTab) IsFiltered() bool {
	return t.searchKeyword != ""
}

// updateCell 테이블 셀을 업데이트합니다.
func (t *ProgramTab) updateCell(row int, col int, cell fyne.CanvasObject) {
	if row >= len(t.pageData) {
		return
	}
	p := t.pageData[row]

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
	case 3: // 로컬파일 경로
		if label, ok := cell.(*widget.Label); ok {
			label.SetText(p.ProcessFilePath)
		}
	case 4: // 추가(수정)시간
		if label, ok := cell.(*widget.Label); ok {
			label.SetText(p.ProcessCreatedAt)
		}
	}
}

// onRowDoubleClick 행 더블클릭 시 수정 다이얼로그를 표시합니다.
func (t *ProgramTab) onRowDoubleClick(row int) {
	if row >= len(t.pageData) {
		return
	}
	t.showEditDialog(t.pageData[row])
}

// onDeleteSelected 선택된 패키지을 삭제합니다.
func (t *ProgramTab) onDeleteSelected() {
	checkedRows := t.programTable.GetCheckedRows()
	if len(checkedRows) == 0 {
		return
	}

	dialog.ShowConfirm("삭제 확인",
		fmt.Sprintf("선택한 %d개 패키지을 삭제하시겠습니까?", len(checkedRows)),
		func(ok bool) {
			if !ok {
				return
			}

			// 삭제 실행
			for _, row := range checkedRows {
				if row < len(t.pageData) {
					p := t.pageData[row]
					if err := t.programRepo.Delete(p.ID); err != nil {
						dialog.ShowError(err, t.window)
						return
					}
				}
			}

			t.loadPrograms()
			dialog.ShowInformation("삭제 완료", fmt.Sprintf("%d개 패키지이 삭제되었습니다.", len(checkedRows)), t.window)
		}, t.window)
}

// onAddOrEdit 추가/수정 버튼 클릭 처리
func (t *ProgramTab) onAddOrEdit() {
	checkedRows := t.programTable.GetCheckedRows()

	// 여러 개 선택 시 안내 다이얼로그 표시
	if len(checkedRows) > 1 {
		dialog.ShowInformation("안내", "수정하려면 하나만 선택해주세요.", t.window)
		return
	}

	if len(checkedRows) == 0 {
		// 추가 모드
		t.showEditDialog(nil)
	} else {
		// 수정 모드
		if checkedRows[0] < len(t.pageData) {
			t.showEditDialog(t.pageData[checkedRows[0]])
		}
	}
}

// showEditDialog 패키지 추가/수정 다이얼로그를 표시합니다.
func (t *ProgramTab) showEditDialog(program *model.ProcessInfo) {
	isEdit := program != nil

	// 입력 필드 너비 (장비 관리와 동일)
	entryWidth := float32(250)
	rowHeight := float32(36)
	labelWidth := float32(100)
	rowSpacing := float32(20)    // 라인 간격 2배
	buttonSpacing := float32(60) // 버튼 간격 3배

	// 입력 필드 생성
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("패키지명")

	versionEntry := widget.NewEntry()
	versionEntry.SetPlaceHolder("버전 (예: v1.0.0)")

	filePathLabel := widget.NewLabel("파일을 선택해주세요")
	filePathLabel.Truncation = fyne.TextTruncateEllipsis
	filePath := ""

	// 수정 모드일 경우 기존 값 설정
	if isEdit {
		nameEntry.SetText(program.ProcessName)
		versionEntry.SetText(program.ProcessVersion)
		filePath = program.ProcessFilePath
		filePathLabel.SetText(filePath)
	}

	// 커스텀 팝업 생성
	var popup *widget.PopUp

	// 찾아보기 버튼
	browseBtn := widget.NewButton("찾아보기...", nil)

	// 헤더 (큰 폰트 - RichText 사용)
	title := "패키지 추가"
	if isEdit {
		title = "패키지 수정"
	}
	headerText := widget.NewRichTextFromMarkdown("## " + title)

	// 폼 컨텐츠 (라인 간격 2배)
	formContent := container.NewVBox(
		// 패키지명
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("패키지명:")),
			container.NewGridWrap(fyne.NewSize(entryWidth, rowHeight), nameEntry),
		),
		container.NewGridWrap(fyne.NewSize(1, rowSpacing), layout.NewSpacer()), // 간격 2배
		// 버전
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("버전:")),
			container.NewGridWrap(fyne.NewSize(entryWidth, rowHeight), versionEntry),
		),
		container.NewGridWrap(fyne.NewSize(1, rowSpacing), layout.NewSpacer()), // 간격 2배
		widget.NewSeparator(),
		container.NewGridWrap(fyne.NewSize(1, rowSpacing), layout.NewSpacer()), // 간격 2배
		// 로컬 파일
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("로컬 파일:")),
			container.NewGridWrap(fyne.NewSize(entryWidth-80, rowHeight), filePathLabel),
			browseBtn,
		),
		container.NewGridWrap(fyne.NewSize(1, buttonSpacing), layout.NewSpacer()), // 버튼 간격 3배
	)

	// 저장 처리 함수
	onSave := func() {
		// 유효성 검사
		if nameEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("패키지명을 입력해주세요"), t.window)
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
			p = program.Clone() // 기존 객체의 복사본 사용 (ID 유지)
		} else {
			p = model.NewProcessInfo(nameEntry.Text, versionEntry.Text)
		}
		p.ProcessName = nameEntry.Text
		p.ProcessVersion = versionEntry.Text
		p.ProcessFilePath = filePath
		p.ProcessCreatedAt = time.Now().Format("2006-01-02 15:04:05")

		if err := t.programRepo.Save(p); err != nil {
			dialog.ShowError(err, t.window)
			return
		}

		popup.Hide()
		t.programTable.ClearChecked()
		t.loadPrograms()
		if isEdit {
			dialog.ShowInformation("수정 완료", "패키지이 수정되었습니다.", t.window)
		} else {
			dialog.ShowInformation("추가 완료", "패키지이 추가되었습니다.", t.window)
		}
	}

	// 버튼 (간격 3배 = 60)
	cancelBtn := component.NewCustomButton("취소", nil, themes.Colors["black"], themes.Colors["lightgray"], func() {
		popup.Hide()
	}, 5, 5, 5, 5)
	saveBtn := component.NewCustomButton("저장", nil, nil, themes.Colors["blue"], onSave, 5, 5, 5, 5)

	// 버튼 컨테이너 (중앙 정렬, 버튼 간격 3배 = 60)
	btnContainer := container.NewHBox(
		layout.NewSpacer(),
		cancelBtn,
		container.NewGridWrap(fyne.NewSize(60, 1), layout.NewSpacer()), // 버튼 간격 3배
		saveBtn,
		layout.NewSpacer(),
	)

	// 전체 컨텐츠
	content := container.NewVBox(
		headerText,
		container.NewGridWrap(fyne.NewSize(1, rowSpacing), layout.NewSpacer()), // 헤더 아래 간격
		formContent,
		btnContainer,
		container.NewGridWrap(fyne.NewSize(1, 20), layout.NewSpacer()), // 하단 여백
	)

	// 고정 크기 컨테이너
	paddedContent := container.New(layout.NewCustomPaddedLayout(20, 20, 20, 20), content)
	sizedContent := container.NewGridWrap(fyne.NewSize(450, 400), paddedContent)

	// 팝업 생성
	popup = widget.NewModalPopUp(sizedContent, t.window.Canvas())

	// 찾아보기 버튼 동작 설정 (Hide/Show 패턴)
	browseBtn.OnTapped = func() {
		popup.Hide() // 부모 다이얼로그 숨김

		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if reader != nil {
				filePath = reader.URI().Path()
				filePathLabel.SetText(filePath)
				reader.Close()
			}
			popup.Show() // 부모 다이얼로그 다시 표시
		}, t.window)
		fileDialog.SetFilter(fynestorage.NewExtensionFileFilter([]string{".tar", ".tar.gz", ".tgz", ".zip"}))
		fileDialog.Show()
	}

	popup.Show()
}

// GetAllPrograms 모든 패키지 목록을 반환합니다.
func (t *ProgramTab) GetAllPrograms() []*model.ProcessInfo {
	programs, err := t.programRepo.GetAll()
	if err != nil {
		return []*model.ProcessInfo{}
	}
	return programs
}
