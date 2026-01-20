package ui

import (
	"fmt"
	"strings"

	"fms/internal/model"
	"fms/internal/storage"
	"fms/internal/themes"
	"fms/internal/ui/component"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// FirewallTab 방화벽 관리 탭 (파일 목록)
type FirewallTab struct {
	window    fyne.Window
	fileStore *storage.FileStore
	mainUI    *MainUI // 동적 탭 열기 위한 참조
	content   fyne.CanvasObject

	// UI 컴포넌트
	searchBox *component.SearchBox
	fileTable *component.PagedTable

	// 데이터
	files         []*model.FirewallFile
	filteredFiles []*model.FirewallFile
}

// NewFirewallTab 새로운 방화벽 관리 탭을 생성합니다.
func NewFirewallTab(window fyne.Window, fileStore *storage.FileStore, mainUI *MainUI) *FirewallTab {
	tab := &FirewallTab{
		window:    window,
		fileStore: fileStore,
		mainUI:    mainUI,
		files:     []*model.FirewallFile{},
	}
	tab.createUI()
	tab.loadFiles()
	return tab
}

// Content 탭 컨텐츠를 반환합니다.
func (t *FirewallTab) Content() fyne.CanvasObject {
	return t.content
}

// createUI UI를 생성합니다.
func (t *FirewallTab) createUI() {
	// 검색 컴포넌트
	t.searchBox = component.NewSearchBox(component.SearchBoxConfig{
		Placeholder: "검색...",
		Width:       200,
		OnSearch: func(text string) {
			t.filterFiles(text)
		},
	})

	// 삭제 버튼
	deleteBtn := component.NewCustomButton("선택삭제", nil, themes.Colors["red"], nil, func() {
		t.onDeleteSelected()
	})

	// 파일추가/수정 버튼
	addEditBtn := component.NewCustomButton("파일추가/수정", nil, nil, themes.Colors["blue"], func() {
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

	// 테이블 설정 (5컬럼: 선택, 파일이름, 만든날짜, 수정한날짜, 버전)
	t.fileTable = component.NewPagedTableWithWindow(component.PagedTableConfig{
		Columns: []component.ColumnDef{
			{Header: "선택", Width: 50},
			{Header: "파일이름", Width: 200},
			{Header: "만든 날짜", Width: 180},
			{Header: "수정한 날짜", Width: 180},
			{Header: "버전", Width: 100},
		},
		PageSize: 10,
		OnCellUpdate: func(row int, col int, cell fyne.CanvasObject) {
			t.updateCell(row, col, cell)
		},
		OnRowDoubleClick: func(row int) {
			t.onRowDoubleClick(row)
		},
		// 인라인 편집 설정: 파일이름 컬럼 (컬럼 1)
		EditableColumns: map[int]component.EditColumnConfig{
			1: { // 파일이름 컬럼
				Type: component.EditTypeEntry,
				GetValue: func(row int) string {
					if row < len(t.filteredFiles) {
						return t.filteredFiles[row].FileName
					}
					return ""
				},
				OnEdit: func(row int, oldValue, newValue string) bool {
					return t.onInlineEditFileName(row, oldValue, newValue)
				},
			},
		},
	}, t.window)

	// 메인 컨텐츠
	t.content = container.NewBorder(
		topBar,
		nil, nil, nil,
		t.fileTable.Content(),
	)
}

// loadFiles 파일 목록을 로드합니다.
func (t *FirewallTab) loadFiles() {
	files, err := t.fileStore.GetAllFiles()
	if err != nil {
		dialog.ShowError(err, t.window)
		return
	}

	t.files = files
	t.filteredFiles = files
	t.fileTable.SetData(len(t.filteredFiles))
}

// RefreshFiles 파일 목록을 새로고침합니다.
func (t *FirewallTab) RefreshFiles() {
	t.loadFiles()
}

// filterFiles 검색어로 파일을 필터링합니다.
func (t *FirewallTab) filterFiles(query string) {
	if query == "" {
		t.filteredFiles = t.files
	} else {
		query = strings.ToLower(query)
		filtered := []*model.FirewallFile{}
		for _, f := range t.files {
			if strings.Contains(strings.ToLower(f.FileName), query) ||
				strings.Contains(strings.ToLower(f.Version), query) {
				filtered = append(filtered, f)
			}
		}
		t.filteredFiles = filtered
		if len(filtered) == 0 {
			dialog.ShowInformation("검색 결과", "검색 결과가 없습니다.", t.window)
		}
	}
	t.fileTable.SetData(len(t.filteredFiles))
}

// updateCell 테이블 셀을 업데이트합니다.
func (t *FirewallTab) updateCell(row int, col int, cell fyne.CanvasObject) {
	if row >= len(t.filteredFiles) {
		return
	}
	f := t.filteredFiles[row]

	switch col {
	case 0: // 선택 (체크박스) - PagedTable에서 처리
	case 1: // 파일이름
		if label, ok := cell.(*widget.Label); ok {
			label.SetText(f.FileName)
		}
	case 2: // 만든 날짜
		if label, ok := cell.(*widget.Label); ok {
			label.SetText(f.GetDisplayCreatedAt())
		}
	case 3: // 수정한 날짜
		if label, ok := cell.(*widget.Label); ok {
			label.SetText(f.GetDisplayModifiedAt())
		}
	case 4: // 버전
		if label, ok := cell.(*widget.Label); ok {
			label.SetText(f.Version)
		}
	}
}

// onRowDoubleClick 행 더블클릭 시 편집 탭을 엽니다.
func (t *FirewallTab) onRowDoubleClick(row int) {
	if row >= len(t.filteredFiles) {
		return
	}
	file := t.filteredFiles[row]

	// 파일 내용 로드
	loadedFile, err := t.fileStore.GetFile(file.FileName)
	if err != nil {
		dialog.ShowError(err, t.window)
		return
	}

	// MainUI를 통해 편집 탭 열기
	if t.mainUI != nil {
		t.mainUI.OpenFirewallEditTab(loadedFile)
	}
}

// onDeleteSelected 선택된 파일을 삭제합니다.
func (t *FirewallTab) onDeleteSelected() {
	checkedRows := t.fileTable.GetCheckedRows()
	if len(checkedRows) == 0 {
		return
	}

	dialog.ShowConfirm("삭제 확인",
		fmt.Sprintf("선택한 %d개 파일을 삭제하시겠습니까?", len(checkedRows)),
		func(ok bool) {
			if !ok {
				return
			}

			// 삭제할 파일명 수집
			fileNames := make([]string, 0, len(checkedRows))
			for _, row := range checkedRows {
				if row < len(t.filteredFiles) {
					fileNames = append(fileNames, t.filteredFiles[row].FileName)
				}
			}

			// 삭제 실행
			if err := t.fileStore.DeleteFiles(fileNames); err != nil {
				dialog.ShowError(err, t.window)
				return
			}

			t.loadFiles()
			dialog.ShowInformation("삭제 완료", fmt.Sprintf("%d개 파일이 삭제되었습니다.", len(fileNames)), t.window)
		}, t.window)
}

// onAddOrEdit 파일추가/수정 버튼 클릭 처리
func (t *FirewallTab) onAddOrEdit() {
	checkedRows := t.fileTable.GetCheckedRows()

	// 여러 개 선택 시 안내 다이얼로그 표시
	if len(checkedRows) > 1 {
		dialog.ShowInformation("안내", "수정하려면 하나만 선택해주세요.", t.window)
		return
	}

	if len(checkedRows) == 0 {
		// 추가 모드
		t.showFileDialog(nil)
	} else {
		// 수정 모드 - 해당 파일의 수정 다이얼로그 표시
		if checkedRows[0] < len(t.filteredFiles) {
			file := t.filteredFiles[checkedRows[0]]
			t.showFileDialog(file)
		}
	}
}

// showFileDialog 파일 추가/수정 다이얼로그를 표시합니다.
// file이 nil이면 추가 모드, file이 있으면 수정 모드
func (t *FirewallTab) showFileDialog(file *model.FirewallFile) {
	isEditMode := file != nil

	// 입력 필드 설정
	entryWidth := float32(300)
	rowHeight := float32(36)
	labelWidth := float32(80)
	rowSpacing := float32(20)
	buttonSpacing := float32(60)

	// 선택 드롭다운 (파일생성/파일찾기) - 추가 모드에서만 사용
	modeSelect := widget.NewSelect([]string{"파일생성", "파일찾기"}, nil)
	modeSelect.SetSelected("파일생성")

	// 파일명 입력 필드
	fileNameEntry := widget.NewEntry()
	// "파일찾기" 모드에서 파일명 수정 방지를 위한 변수
	isFileBrowseMode := false
	selectedFileName := ""

	if isEditMode {
		fileNameEntry.SetText(file.FileName)
	} else {
		fileNameEntry.SetPlaceHolder("파일명 (예: rules-v1_0_1.txt)")
	}

	// 파일찾기 모드에서 파일명 수정 방지
	fileNameEntry.OnChanged = func(s string) {
		if isFileBrowseMode && selectedFileName != "" && s != selectedFileName {
			fileNameEntry.SetText(selectedFileName)
		}
	}

	// 파일 경로 라벨 (파일찾기 모드용)
	filePathLabel := widget.NewLabel("파일을 선택해주세요")
	filePathLabel.Truncation = fyne.TextTruncateEllipsis
	filePath := ""

	// 찾아보기 버튼
	browseBtn := widget.NewButton("찾아보기...", nil)
	browseBtn.Hide() // 초기에는 숨김

	// 커스텀 팝업
	var popup *widget.PopUp

	// 모드 변경 시 UI 업데이트 (추가 모드에서만 사용)
	modeSelect.OnChanged = func(selected string) {
		if selected == "파일생성" {
			isFileBrowseMode = false
			selectedFileName = ""
			fileNameEntry.SetPlaceHolder("파일명 (예: rules-v1_0_1.txt)")
			fileNameEntry.SetText("")
			browseBtn.Hide()
			filePathLabel.Hide()
		} else {
			isFileBrowseMode = true
			selectedFileName = ""
			fileNameEntry.SetPlaceHolder("파일을 선택해주세요")
			fileNameEntry.SetText("")
			browseBtn.Show()
			filePathLabel.Show()
		}
	}

	// 헤더
	var headerText *widget.RichText
	if isEditMode {
		headerText = widget.NewRichTextFromMarkdown("## 파일 수정")
	} else {
		headerText = widget.NewRichTextFromMarkdown("## 파일 추가")
	}

	// 선택 행 (추가 모드에서만 표시)
	selectRow := container.NewHBox(
		container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("선택:")),
		container.NewGridWrap(fyne.NewSize(entryWidth, rowHeight), modeSelect),
	)

	// 파일찾기 행 (추가 모드에서만 표시)
	browseRow := container.NewHBox(
		container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("파일찾기:")),
		container.NewGridWrap(fyne.NewSize(entryWidth-80, rowHeight), filePathLabel),
		browseBtn,
	)

	// 폼 컨텐츠 구성
	var formContent *fyne.Container
	if isEditMode {
		// 수정 모드: 파일명만 표시
		formContent = container.NewVBox(
			// 파일명
			container.NewHBox(
				container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("파일명:")),
				container.NewGridWrap(fyne.NewSize(entryWidth, rowHeight), fileNameEntry),
			),
			container.NewGridWrap(fyne.NewSize(1, buttonSpacing), layout.NewSpacer()),
		)
	} else {
		// 추가 모드: 선택, 파일명, 파일찾기 모두 표시
		formContent = container.NewVBox(
			// 선택
			selectRow,
			container.NewGridWrap(fyne.NewSize(1, rowSpacing), layout.NewSpacer()),
			// 파일명
			container.NewHBox(
				container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("파일명:")),
				container.NewGridWrap(fyne.NewSize(entryWidth, rowHeight), fileNameEntry),
			),
			container.NewGridWrap(fyne.NewSize(1, rowSpacing), layout.NewSpacer()),
			// 파일찾기 (파일찾기 모드용)
			browseRow,
			container.NewGridWrap(fyne.NewSize(1, buttonSpacing), layout.NewSpacer()),
		)
		// 초기 상태에서 파일찾기 관련 숨김
		filePathLabel.Hide()
	}

	// 완료 처리 함수
	onComplete := func() {
		var fileName string

		if isEditMode {
			// 수정 모드
			fileName = strings.TrimSpace(fileNameEntry.Text)
			if fileName == "" {
				dialog.ShowError(fmt.Errorf("파일명을 입력해주세요"), t.window)
				return
			}
			// 파일명이 변경된 경우
			if fileName != file.FileName {
				// 중복 확인
				if t.fileStore.FileExists(fileName) {
					dialog.ShowError(fmt.Errorf("이미 존재하는 파일명입니다: %s", fileName), t.window)
					return
				}
				// 파일 이름 변경
				if err := t.fileStore.RenameFile(file.FileName, fileName); err != nil {
					dialog.ShowError(err, t.window)
					return
				}
			}

			popup.Hide()
			t.fileTable.ClearChecked()
			t.loadFiles()

			// 변경된 파일의 편집 탭 열기
			loadedFile, err := t.fileStore.GetFile(fileName)
			if err != nil {
				dialog.ShowError(err, t.window)
				return
			}
			if t.mainUI != nil {
				// 기존 탭이 열려있으면 탭 이름 업데이트
				if fileName != file.FileName {
					t.mainUI.UpdateFirewallEditTabName(file.FileName, loadedFile)
				}
				t.mainUI.OpenFirewallEditTab(loadedFile)
			}
		} else {
			// 추가 모드
			var contents string

			if modeSelect.Selected == "파일생성" {
				// 파일생성 모드
				fileName = strings.TrimSpace(fileNameEntry.Text)
				if fileName == "" {
					dialog.ShowError(fmt.Errorf("파일명을 입력해주세요"), t.window)
					return
				}
				contents = "" // 빈 파일
			} else {
				// 파일찾기 모드
				if filePath == "" {
					dialog.ShowError(fmt.Errorf("파일을 선택해주세요"), t.window)
					return
				}
				// 파일 경로에서 파일명 추출
				parts := strings.Split(filePath, "/")
				if len(parts) == 0 {
					parts = strings.Split(filePath, "\\")
				}
				fileName = parts[len(parts)-1]
			}

			// 중복 확인
			if t.fileStore.FileExists(fileName) {
				dialog.ShowError(fmt.Errorf("이미 존재하는 파일명입니다: %s", fileName), t.window)
				return
			}

			// 파일 저장
			if modeSelect.Selected == "파일찾기" {
				// 파일 복사
				if err := t.fileStore.CopyFileToData(filePath, fileName); err != nil {
					dialog.ShowError(err, t.window)
					return
				}
			} else {
				// 빈 파일 생성
				newFile := model.NewFirewallFile(fileName)
				newFile.Contents = contents
				if err := t.fileStore.SaveFile(newFile); err != nil {
					dialog.ShowError(err, t.window)
					return
				}
			}

			popup.Hide()
			t.fileTable.ClearChecked()
			t.loadFiles()
		}
	}

	// 버튼
	cancelBtn := component.NewCustomButton("취소", nil, themes.Colors["black"], themes.Colors["lightgray"], func() {
		popup.Hide()
	}, 5, 5, 5, 5)
	completeBtn := component.NewCustomButton("완료", nil, nil, themes.Colors["blue"], onComplete, 5, 5, 5, 5)

	// 버튼 컨테이너
	btnContainer := container.NewHBox(
		layout.NewSpacer(),
		cancelBtn,
		container.NewGridWrap(fyne.NewSize(60, 1), layout.NewSpacer()),
		completeBtn,
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

	// 고정 크기 컨테이너 (수정 모드는 더 작게)
	var dialogWidth, dialogHeight float32
	if isEditMode {
		dialogWidth = 500
		dialogHeight = 250
	} else {
		dialogWidth = 500
		dialogHeight = 400
	}

	paddedContent := container.New(layout.NewCustomPaddedLayout(20, 20, 20, 20), content)
	sizedContent := container.NewGridWrap(fyne.NewSize(dialogWidth, dialogHeight), paddedContent)

	// 팝업 생성
	popup = widget.NewModalPopUp(sizedContent, t.window.Canvas())

	// 찾아보기 버튼 동작 설정 (추가 모드에서만 사용)
	if !isEditMode {
		browseBtn.OnTapped = func() {
			popup.Hide()

			fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
				if reader != nil {
					filePath = reader.URI().Path()
					filePathLabel.SetText(filePath)
					// 파일명 자동 설정
					parts := strings.Split(filePath, "/")
					if len(parts) == 0 {
						parts = strings.Split(filePath, "\\")
					}
					// 파일명 설정 및 수정 방지를 위한 값 저장
					selectedFileName = parts[len(parts)-1]
					fileNameEntry.SetText(selectedFileName)
					reader.Close()
				}
				popup.Show()
			}, t.window)
			fileDialog.Show()
		}
	}

	popup.Show()
}

// onInlineEditFileName 인라인 편집으로 파일명을 변경합니다.
func (t *FirewallTab) onInlineEditFileName(row int, oldValue, newValue string) bool {
	if row >= len(t.filteredFiles) {
		return false
	}

	newValue = strings.TrimSpace(newValue)
	if newValue == "" {
		dialog.ShowError(fmt.Errorf("파일명을 입력해주세요"), t.window)
		return false
	}

	// 값이 같으면 변경 없음
	if newValue == oldValue {
		return true
	}

	// 중복 확인
	if t.fileStore.FileExists(newValue) {
		dialog.ShowError(fmt.Errorf("이미 존재하는 파일명입니다: %s", newValue), t.window)
		return false
	}

	// 파일 이름 변경
	if err := t.fileStore.RenameFile(oldValue, newValue); err != nil {
		dialog.ShowError(err, t.window)
		return false
	}

	// 편집 탭이 열려있으면 탭 이름 업데이트
	if t.mainUI != nil {
		loadedFile, err := t.fileStore.GetFile(newValue)
		if err == nil {
			t.mainUI.UpdateFirewallEditTabName(oldValue, loadedFile)
		}
	}

	// 목록 새로고침
	t.loadFiles()
	return true
}

// GetFileNames 모든 파일명 목록을 반환합니다. (배포 시 드롭다운용)
func (t *FirewallTab) GetFileNames() []string {
	names := make([]string, len(t.files))
	for i, f := range t.files {
		names[i] = f.FileName
	}
	return names
}

// GetFileContents 특정 파일의 내용을 반환합니다.
func (t *FirewallTab) GetFileContents(fileName string) (string, error) {
	file, err := t.fileStore.GetFile(fileName)
	if err != nil {
		return "", err
	}
	return file.Contents, nil
}
