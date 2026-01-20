package component

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ColumnDef 테이블 컬럼 정의
type ColumnDef struct {
	Header string  // 헤더 텍스트
	Width  float32 // 컬럼 너비 (0이면 자동)
}

// EditColumnType 편집 컬럼 타입
type EditColumnType int

const (
	EditTypeEntry  EditColumnType = iota // Entry 위젯 (텍스트 입력)
	EditTypeSelect                       // Select 위젯 (드롭다운)
)

// EditColumnConfig 편집 가능 컬럼 설정
type EditColumnConfig struct {
	Type       EditColumnType                        // 편집 위젯 타입
	Options    []string                              // Select 타입일 때 옵션 목록
	OnEdit     func(row int, oldValue, newValue string) bool // 편집 완료 콜백 (false 반환 시 취소)
	GetValue   func(row int) string                  // 현재 값 조회 콜백
}

// PagedTableConfig 페이지네이션 테이블 설정
type PagedTableConfig struct {
	Columns          []ColumnDef                                    // 컬럼 정의 (첫 번째는 체크박스 컬럼)
	PageSize         int                                            // 페이지당 항목 수
	OnCellUpdate     func(row int, col int, cell fyne.CanvasObject) // 셀 업데이트 콜백
	OnRowSelected    func(row int)                                  // 행 선택 콜백 (단일 클릭/Enter)
	OnRowDoubleClick func(row int)                                  // 행 더블클릭 콜백
	OnCheckChange    func(row int, checked bool)                    // 체크 변경 콜백
	EditableColumns  map[int]EditColumnConfig                       // 인라인 편집 가능 컬럼 (컬럼 인덱스 → 설정)
}

// PagedTable 페이지네이션 지원 테이블
type PagedTable struct {
	widget.BaseWidget

	config       PagedTableConfig
	table        *widget.Table
	totalItems   int          // 전체 데이터 수
	currentPage  int          // 현재 페이지 (0부터 시작)
	checkedRows  map[int]bool // 체크된 행 (실제 데이터 인덱스 기준)
	selectedRow  int          // 현재 선택된 행 (페이지 내 인덱스, -1이면 선택 없음)
	allSelected  bool         // 전체 선택 상태

	// 인라인 편집 상태
	editingCell    *widget.TableCellID // 현재 편집 중인 셀 (nil이면 편집 중 아님)
	editingValue   string              // 편집 전 원본 값
	window         fyne.Window         // ESC 키 처리용 윈도우

	// 페이지네이션 UI
	pageLabel   *widget.Label
	prevBtn     *widget.Button
	nextBtn     *widget.Button
	pageContent *fyne.Container

	// 전체 컨테이너
	content *fyne.Container
}

// NewPagedTable 새 페이지네이션 테이블 생성
func NewPagedTable(config PagedTableConfig) *PagedTable {
	return NewPagedTableWithWindow(config, nil)
}

// NewPagedTableWithWindow 윈도우와 함께 새 페이지네이션 테이블 생성 (인라인 편집 시 ESC 키 처리용)
func NewPagedTableWithWindow(config PagedTableConfig, window fyne.Window) *PagedTable {
	if config.PageSize <= 0 {
		config.PageSize = 10 // 기본값
	}

	t := &PagedTable{
		config:      config,
		currentPage: 0,
		totalItems:  0,
		checkedRows: make(map[int]bool),
		selectedRow: -1,
		allSelected: false,
		window:      window,
	}
	t.ExtendBaseWidget(t)
	t.createUI()
	return t
}

// createUI UI 생성
func (t *PagedTable) createUI() {
	// 테이블 생성
	t.table = widget.NewTable(
		// Length: 행/열 수 반환
		func() (rows, cols int) {
			pageRows := t.getPageRowCount()
			return pageRows + 1, len(t.config.Columns) // +1 for header
		},
		// CreateCell: 셀 위젯 생성
		func() fyne.CanvasObject {
			// 행 선택 배경 (투명 → 선택 시 하이라이트)
			background := canvas.NewRectangle(color.Transparent)

			// 체크박스용 버튼 (투명, 클릭 감지용)
			checkBtn := widget.NewButton("", nil)
			checkBtn.Importance = widget.LowImportance

			// 체크박스 표시용 텍스트
			checkText := canvas.NewText("", color.Black)
			checkText.TextSize = 16
			checkText.Alignment = fyne.TextAlignCenter

			// 일반 라벨 (Truncation 설정으로 텍스트가 컬럼 너비를 초과하지 않도록 함)
			label := widget.NewLabel("")
			label.Truncation = fyne.TextTruncateEllipsis

			// 커스텀 셀용 컨테이너
			customContainer := container.NewStack()

			// 배경을 가장 아래에, 나머지를 위에 배치
			// checkBtn은 체크박스 컬럼에서만 표시됨
			innerStack := container.NewStack(background, checkBtn, checkText, label, customContainer)

			// 더블탭 감지 래퍼로 감싸서 반환
			return newDoubleTappableCell(innerStack)
		},
		// UpdateCell: 셀 데이터 업데이트
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			t.updateCell(id, cell)
		},
	)

	// 컬럼 너비 설정
	for i, col := range t.config.Columns {
		if col.Width > 0 {
			t.table.SetColumnWidth(i, col.Width)
		}
	}

	// 셀 선택 이벤트 (클릭/Enter 키/방향키)
	t.table.OnSelected = func(id widget.TableCellID) {
		if id.Row == 0 {
			t.table.UnselectAll()
			return // 헤더 클릭 무시
		}

		pageRow := id.Row - 1
		dataIndex := t.getDataIndex(pageRow)
		if dataIndex < 0 || dataIndex >= t.totalItems {
			t.table.UnselectAll()
			return
		}

		// 행 선택 상태 업데이트
		t.selectedRow = pageRow

		// 행 선택 콜백 호출 (체크박스 토글은 checkBtn에서 처리)
		if t.config.OnRowSelected != nil {
			t.config.OnRowSelected(dataIndex)
		}

		// 행 전체 하이라이트를 위해 테이블 새로고침
		t.table.Refresh()
	}

	// 페이지네이션 UI 생성
	t.pageLabel = widget.NewLabel("page 1/1")
	t.prevBtn = widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		t.PrevPage()
	})
	t.nextBtn = widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() {
		t.NextPage()
	})

	t.pageContent = container.NewHBox(
		t.prevBtn,
		t.pageLabel,
		t.nextBtn,
	)

	// 페이지네이션을 중앙에 배치
	paginationBar := container.NewCenter(t.pageContent)

	// 테이블 외곽선 (상, 하, 좌, 우)
	borderColor := color.RGBA{R: 200, G: 200, B: 200, A: 255}
	borderTop := canvas.NewRectangle(borderColor)
	borderTop.SetMinSize(fyne.NewSize(0, 1))
	borderBottom := canvas.NewRectangle(borderColor)
	borderBottom.SetMinSize(fyne.NewSize(0, 1))
	borderLeft := canvas.NewRectangle(borderColor)
	borderLeft.SetMinSize(fyne.NewSize(1, 0))
	borderRight := canvas.NewRectangle(borderColor)
	borderRight.SetMinSize(fyne.NewSize(1, 0))

	// 테이블을 외곽선으로 감싸기
	borderedTable := container.NewBorder(
		borderTop,
		borderBottom,
		borderLeft,
		borderRight,
		t.table,
	)

	// 전체 레이아웃
	t.content = container.NewBorder(
		nil,
		paginationBar,
		nil, nil,
		borderedTable,
	)

	t.updatePaginationUI()
}

// updateCell 셀 업데이트
func (t *PagedTable) updateCell(id widget.TableCellID, cell fyne.CanvasObject) {
	// doubleTappableCell에서 내부 스택 추출
	dtCell := cell.(*doubleTappableCell)
	stack := dtCell.content.(*fyne.Container)
	background := stack.Objects[0].(*canvas.Rectangle)
	checkBtn := stack.Objects[1].(*widget.Button)
	checkText := stack.Objects[2].(*canvas.Text)
	label := stack.Objects[3].(*widget.Label)
	customContainer := stack.Objects[4].(*fyne.Container)

	// 모든 요소 숨기기
	checkBtn.Hide()
	checkText.Hidden = true
	label.Hide()
	customContainer.Hide()

	// 배경색 초기화
	background.FillColor = color.Transparent

	// 더블탭 콜백 초기화
	dtCell.onDoubleTap = nil
	dtCell.onTap = nil

	if id.Row == 0 {
		// 헤더
		if id.Col == 0 {
			// 첫 번째 컬럼 헤더: 전체 선택 체크박스
			checkBtn.Show()
			checkBtn.OnTapped = func() {
				t.toggleSelectAll()
			}

			checkText.Hidden = false
			if t.allSelected {
				checkText.Text = "✔"
				checkText.Color = color.RGBA{R: 220, G: 20, B: 20, A: 255}
				checkText.TextStyle = fyne.TextStyle{Bold: true}
			} else {
				checkText.Text = "☐"
				checkText.Color = color.RGBA{R: 100, G: 100, B: 100, A: 255}
				checkText.TextStyle = fyne.TextStyle{Bold: false}
			}
			checkText.Refresh()
		} else {
			label.SetText(t.config.Columns[id.Col].Header)
			label.TextStyle = fyne.TextStyle{Bold: true}
			label.Show()
		}
		background.Refresh()
		return
	}

	pageRow := id.Row - 1
	dataIndex := t.getDataIndex(pageRow)
	if dataIndex < 0 || dataIndex >= t.totalItems {
		background.Refresh()
		return
	}

	// 더블탭 콜백 설정 (OnRowDoubleClick이 설정된 경우)
	if t.config.OnRowDoubleClick != nil {
		dtCell.onDoubleTap = func() {
			t.config.OnRowDoubleClick(dataIndex)
		}
	}

	// 선택된 행이면 배경 하이라이트
	if pageRow == t.selectedRow {
		background.FillColor = color.RGBA{R: 200, G: 220, B: 255, A: 255} // 연한 파란색
	}
	background.Refresh()

	if id.Col == 0 {
		// 체크박스 컬럼 - 버튼으로 클릭 감지
		checkBtn.Show()
		checkBtn.OnTapped = func() {
			t.toggleCheck(dataIndex, pageRow)
		}

		checkText.Hidden = false
		if t.checkedRows[dataIndex] {
			checkText.Text = "✔"
			checkText.Color = color.RGBA{R: 220, G: 20, B: 20, A: 255}
			checkText.TextStyle = fyne.TextStyle{Bold: true}
		} else {
			checkText.Text = "☐"
			checkText.Color = color.RGBA{R: 100, G: 100, B: 100, A: 255}
			checkText.TextStyle = fyne.TextStyle{Bold: false}
		}
		checkText.Refresh()
	} else {
		// 편집 가능 컬럼 확인
		editConfig, isEditable := t.config.EditableColumns[id.Col]

		// 현재 편집 중인 셀인지 확인
		isEditingThis := t.editingCell != nil && t.editingCell.Row == id.Row && t.editingCell.Col == id.Col

		if isEditingThis && isEditable {
			// 편집 모드: 편집 위젯 표시
			customContainer.Objects = nil

			if editConfig.Type == EditTypeSelect {
				// Select 위젯
				selectWidget := widget.NewSelect(editConfig.Options, func(selected string) {
					// 선택 변경 시 즉시 저장
					t.finishEditing(dataIndex, id.Col, selected, editConfig)
				})
				// 현재 값 설정
				if editConfig.GetValue != nil {
					selectWidget.SetSelected(editConfig.GetValue(dataIndex))
				}
				customContainer.Objects = append(customContainer.Objects, selectWidget)
			} else {
				// Entry 위젯 (ESC 키 및 포커스 잃음 처리 포함)
				var currentValue string
				if editConfig.GetValue != nil {
					currentValue = editConfig.GetValue(dataIndex)
				}
				entry := newEditableEntry(
					currentValue,
					func(text string) {
						// Enter 키 누르면 저장
						t.finishEditing(dataIndex, id.Col, text, editConfig)
					},
					func() {
						// ESC 키 또는 포커스 잃으면 취소
						t.cancelEditing()
					},
				)
				customContainer.Objects = append(customContainer.Objects, entry)
			}
			customContainer.Show()
			label.Hide()
		} else {
			// 일반 모드: 라벨 표시
			if t.config.OnCellUpdate != nil {
				t.config.OnCellUpdate(dataIndex, id.Col, label)
				label.Show()
			}

			// 편집 가능 컬럼이면 클릭 시 편집 모드 진입
			if isEditable {
				dtCell.onTap = func() {
					t.startEditing(id, dataIndex, editConfig)
				}
			}
		}
	}
}

// toggleSelectAll 전체 선택/해제 토글
func (t *PagedTable) toggleSelectAll() {
	t.allSelected = !t.allSelected

	if t.allSelected {
		// 전체 선택
		for i := 0; i < t.totalItems; i++ {
			t.checkedRows[i] = true
		}
	} else {
		// 전체 해제
		t.checkedRows = make(map[int]bool)
	}

	// 체크 변경 콜백 호출 (각 항목에 대해)
	if t.config.OnCheckChange != nil {
		for i := 0; i < t.totalItems; i++ {
			t.config.OnCheckChange(i, t.allSelected)
		}
	}

	t.table.Refresh()
}

// toggleCheck 체크박스 토글 (버튼 클릭용)
func (t *PagedTable) toggleCheck(dataIndex int, pageRow int) {
	t.checkedRows[dataIndex] = !t.checkedRows[dataIndex]

	// 체크 해제 시 선택 행 하이라이트도 해제
	if t.checkedRows[dataIndex] {
		t.selectedRow = pageRow
	} else {
		t.selectedRow = -1
	}

	// 전체 선택 상태 업데이트
	t.updateAllSelectedState()

	if t.config.OnCheckChange != nil {
		t.config.OnCheckChange(dataIndex, t.checkedRows[dataIndex])
	}
	if t.config.OnRowSelected != nil {
		t.config.OnRowSelected(dataIndex)
	}

	t.table.Refresh()
}

// updateAllSelectedState 전체 선택 상태 업데이트
func (t *PagedTable) updateAllSelectedState() {
	if t.totalItems == 0 {
		t.allSelected = false
		return
	}

	// 모든 항목이 체크되었는지 확인
	allChecked := true
	for i := 0; i < t.totalItems; i++ {
		if !t.checkedRows[i] {
			allChecked = false
			break
		}
	}
	t.allSelected = allChecked
}

// getPageRowCount 현재 페이지의 행 수 반환
func (t *PagedTable) getPageRowCount() int {
	if t.totalItems == 0 {
		return 0
	}

	startIdx := t.currentPage * t.config.PageSize
	remaining := t.totalItems - startIdx
	if remaining > t.config.PageSize {
		return t.config.PageSize
	}
	return remaining
}

// getDataIndex 페이지 내 행 인덱스를 실제 데이터 인덱스로 변환
func (t *PagedTable) getDataIndex(pageRow int) int {
	return t.currentPage*t.config.PageSize + pageRow
}

// getTotalPages 전체 페이지 수 반환
func (t *PagedTable) getTotalPages() int {
	if t.totalItems == 0 {
		return 1
	}
	pages := t.totalItems / t.config.PageSize
	if t.totalItems%t.config.PageSize > 0 {
		pages++
	}
	return pages
}

// updatePaginationUI 페이지네이션 UI 업데이트
func (t *PagedTable) updatePaginationUI() {
	totalPages := t.getTotalPages()
	t.pageLabel.SetText(fmt.Sprintf("page %d/%d", t.currentPage+1, totalPages))

	// 버튼 활성화/비활성화
	if t.currentPage <= 0 {
		t.prevBtn.Disable()
	} else {
		t.prevBtn.Enable()
	}

	if t.currentPage >= totalPages-1 {
		t.nextBtn.Disable()
	} else {
		t.nextBtn.Enable()
	}
}

// SetData 데이터 설정 (전체 항목 수)
func (t *PagedTable) SetData(totalItems int) {
	t.totalItems = totalItems
	t.currentPage = 0
	t.selectedRow = -1
	t.allSelected = false
	t.table.Refresh()
	t.updatePaginationUI()
}

// Refresh 테이블 새로고침
func (t *PagedTable) Refresh() {
	t.table.Refresh()
	t.updatePaginationUI()
}

// NextPage 다음 페이지로 이동
func (t *PagedTable) NextPage() {
	if t.currentPage < t.getTotalPages()-1 {
		t.currentPage++
		t.selectedRow = -1
		t.table.Refresh()
		t.updatePaginationUI()
	}
}

// PrevPage 이전 페이지로 이동
func (t *PagedTable) PrevPage() {
	if t.currentPage > 0 {
		t.selectedRow = -1
		t.currentPage--
		t.table.Refresh()
		t.updatePaginationUI()
	}
}

// GoToPage 특정 페이지로 이동
func (t *PagedTable) GoToPage(page int) {
	totalPages := t.getTotalPages()
	if page < 0 {
		page = 0
	} else if page >= totalPages {
		page = totalPages - 1
	}
	t.currentPage = page
	t.selectedRow = -1
	t.table.Refresh()
	t.updatePaginationUI()
}

// GetCurrentPage 현재 페이지 반환
func (t *PagedTable) GetCurrentPage() int {
	return t.currentPage
}

// GetCheckedRows 체크된 행 인덱스 목록 반환
func (t *PagedTable) GetCheckedRows() []int {
	result := []int{}
	for idx, checked := range t.checkedRows {
		if checked {
			result = append(result, idx)
		}
	}
	return result
}

// SetChecked 특정 행의 체크 상태 설정
func (t *PagedTable) SetChecked(dataIndex int, checked bool) {
	t.checkedRows[dataIndex] = checked
	t.table.Refresh()
}

// ClearChecked 모든 체크 해제
func (t *PagedTable) ClearChecked() {
	t.checkedRows = make(map[int]bool)
	t.allSelected = false
	t.table.Refresh()
}

// SelectAll 현재 페이지 전체 선택
func (t *PagedTable) SelectAll() {
	for i := 0; i < t.getPageRowCount(); i++ {
		dataIndex := t.getDataIndex(i)
		if dataIndex < t.totalItems {
			t.checkedRows[dataIndex] = true
		}
	}
	t.table.Refresh()
}

// SelectAllData 전체 데이터 선택
func (t *PagedTable) SelectAllData() {
	for i := 0; i < t.totalItems; i++ {
		t.checkedRows[i] = true
	}
	t.allSelected = true
	t.table.Refresh()
}

// DeselectAll 전체 선택 해제
func (t *PagedTable) DeselectAll() {
	t.checkedRows = make(map[int]bool)
	t.allSelected = false
	t.table.Refresh()
}

// IsChecked 특정 행의 체크 상태 반환
func (t *PagedTable) IsChecked(dataIndex int) bool {
	return t.checkedRows[dataIndex]
}

// GetCheckedCount 체크된 항목 수 반환
func (t *PagedTable) GetCheckedCount() int {
	count := 0
	for _, checked := range t.checkedRows {
		if checked {
			count++
		}
	}
	return count
}

// SetColumnWidth 컬럼 너비 설정
func (t *PagedTable) SetColumnWidth(col int, width float32) {
	t.table.SetColumnWidth(col, width)
}

// Content 컨텐츠 반환
func (t *PagedTable) Content() fyne.CanvasObject {
	return t.content
}

// CreateRenderer 렌더러 생성
func (t *PagedTable) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.content)
}

// GetPageSize 페이지 크기 반환
func (t *PagedTable) GetPageSize() int {
	return t.config.PageSize
}

// SetPageSize 페이지 크기 설정
func (t *PagedTable) SetPageSize(size int) {
	if size > 0 {
		t.config.PageSize = size
		t.currentPage = 0
		t.table.Refresh()
		t.updatePaginationUI()
	}
}

// GetTotalItems 전체 항목 수 반환
func (t *PagedTable) GetTotalItems() int {
	return t.totalItems
}

// doubleTappableCell 더블탭 감지 셀 래퍼
type doubleTappableCell struct {
	widget.BaseWidget
	content       fyne.CanvasObject
	onDoubleTap   func()
	onTap         func()
}

func newDoubleTappableCell(content fyne.CanvasObject) *doubleTappableCell {
	c := &doubleTappableCell{content: content}
	c.ExtendBaseWidget(c)
	return c
}

func (c *doubleTappableCell) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.content)
}

func (c *doubleTappableCell) Tapped(_ *fyne.PointEvent) {
	if c.onTap != nil {
		c.onTap()
	}
}

func (c *doubleTappableCell) DoubleTapped(_ *fyne.PointEvent) {
	if c.onDoubleTap != nil {
		c.onDoubleTap()
	}
}

// startEditing 편집 모드 시작
func (t *PagedTable) startEditing(id widget.TableCellID, dataIndex int, config EditColumnConfig) {
	// 이미 편집 중이면 먼저 취소
	if t.editingCell != nil {
		t.cancelEditing()
	}

	// 원본 값 저장
	if config.GetValue != nil {
		t.editingValue = config.GetValue(dataIndex)
	}

	// 편집 셀 설정
	t.editingCell = &id

	// 테이블 새로고침으로 편집 위젯 표시
	t.table.Refresh()
}

// finishEditing 편집 완료 (저장)
func (t *PagedTable) finishEditing(dataIndex int, col int, newValue string, config EditColumnConfig) {
	if t.editingCell == nil {
		return
	}

	oldValue := t.editingValue

	// 값이 변경되었으면 콜백 호출
	if newValue != oldValue && config.OnEdit != nil {
		if !config.OnEdit(dataIndex, oldValue, newValue) {
			// 저장 실패 시 취소
			t.cancelEditing()
			return
		}
	}

	// 편집 상태 초기화
	t.editingCell = nil
	t.editingValue = ""

	// 테이블 새로고침
	t.table.Refresh()
}

// cancelEditing 편집 취소
func (t *PagedTable) cancelEditing() {
	if t.editingCell == nil {
		return
	}

	// 편집 상태 초기화
	t.editingCell = nil
	t.editingValue = ""

	// 테이블 새로고침
	t.table.Refresh()
}

// editableEntry ESC 키 및 포커스 잃음 처리가 가능한 Entry 확장
type editableEntry struct {
	widget.Entry
	onCancel  func()
	submitted bool // 제출 완료 여부 (FocusLost에서 중복 취소 방지)
}

func newEditableEntry(text string, onSubmit func(string), onCancel func()) *editableEntry {
	e := &editableEntry{
		onCancel:  onCancel,
		submitted: false,
	}
	e.ExtendBaseWidget(e)
	e.SetText(text)
	e.OnSubmitted = func(text string) {
		e.submitted = true
		onSubmit(text)
	}
	return e
}

func (e *editableEntry) TypedKey(key *fyne.KeyEvent) {
	if key.Name == fyne.KeyEscape {
		e.submitted = true // ESC도 의도적 종료로 처리
		if e.onCancel != nil {
			e.onCancel()
		}
		return
	}
	e.Entry.TypedKey(key)
}

func (e *editableEntry) FocusLost() {
	e.Entry.FocusLost()
	// 제출되지 않은 상태에서 포커스를 잃으면 편집 취소
	if !e.submitted && e.onCancel != nil {
		e.onCancel()
	}
}
