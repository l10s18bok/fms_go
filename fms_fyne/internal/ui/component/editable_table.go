package component

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// cellEntry Esc 키 처리가 가능한 커스텀 Entry
type cellEntry struct {
	widget.Entry
	onEscape func() // Esc 키 콜백
}

func newCellEntry() *cellEntry {
	e := &cellEntry{}
	e.ExtendBaseWidget(e)
	return e
}

func (e *cellEntry) TypedKey(key *fyne.KeyEvent) {
	if key.Name == fyne.KeyEscape && e.onEscape != nil {
		e.onEscape()
		return
	}
	e.Entry.TypedKey(key)
}

// EditableCellType 편집 가능 셀 타입
type EditableCellType int

const (
	CellTypeLabel     EditableCellType = iota // Label (읽기 전용)
	CellTypeEntry                             // Entry (텍스트 입력)
	CellTypeSelect                            // Select (드롭다운)
	CellTypeCheck                             // Check (체크박스) - 항상 표시
	CellTypeButton                            // Button (버튼) - 항상 표시
	CellTypeHyperlink                         // Hyperlink (클릭 가능 링크)
)

// EditableCellConfig 셀 설정
type EditableCellConfig struct {
	Type       EditableCellType // 셀 타입
	Text       string           // Label/Entry/Hyperlink 텍스트
	Options    []string         // Select 옵션 목록
	Selected   string           // Select 선택된 값
	Checked    bool             // Check 체크 상태
	Icon       fyne.Resource    // Button 아이콘
	OnChanged  func(value any)  // 값 변경 콜백 (Select: string, Entry: string, Check: bool)
	OnTapped   func()           // Button/Hyperlink 클릭 콜백
}

// EditableTableColumn 컬럼 정의
type EditableTableColumn struct {
	Header     string  // 헤더 텍스트
	Width      float32 // 고정 너비 (0이면 비율 사용)
	WidthRatio float32 // 너비 비율 (Width가 0일 때)
}

// EditableTableConfig 테이블 설정
type EditableTableConfig struct {
	Columns       []EditableTableColumn                // 컬럼 정의
	GetRowCount   func() int                           // 행 수 반환
	GetCellConfig func(row, col int) EditableCellConfig // 셀 설정 반환
}

// EditableTable 클릭 시 편집 전환 테이블
type EditableTable struct {
	widget.BaseWidget

	config        EditableTableConfig
	table         *widget.Table
	borderedTable *fyne.Container
	lastWidth     float32

	// 편집 상태
	editingCell  *widget.TableCellID // 현재 편집 중인 셀
	isRefreshing bool                // Refresh 중복 방지
}

// NewEditableTable 새 테이블 생성
func NewEditableTable(config EditableTableConfig) *EditableTable {
	t := &EditableTable{
		config: config,
	}
	t.ExtendBaseWidget(t)
	t.createTable()
	return t
}

// createTable 테이블 생성
func (t *EditableTable) createTable() {
	t.table = widget.NewTable(
		// Length: 행/열 수 반환
		func() (rows, cols int) {
			if t.config.GetRowCount != nil {
				return t.config.GetRowCount(), len(t.config.Columns)
			}
			return 0, len(t.config.Columns)
		},
		// CreateCell: 셀 위젯 생성
		func() fyne.CanvasObject {
			// 불투명 배경
			bg := canvas.NewRectangle(theme.Color(theme.ColorNameBackground))

			// 모든 위젯 타입 포함
			// 인덱스: 0=bg, 1=Label, 2=Entry(cellEntry), 3=Select, 4=Check, 5=Button, 6=Hyperlink
			label := widget.NewLabel("")
			label.Truncation = fyne.TextTruncateEllipsis

			return container.NewStack(
				bg,
				label,
				newCellEntry(),
				widget.NewSelect([]string{}, nil),
				widget.NewCheck("", nil),
				widget.NewButtonWithIcon("", theme.DeleteIcon(), nil),
				widget.NewHyperlink("", nil),
			)
		},
		// UpdateCell: 셀 데이터 업데이트
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			t.updateCell(id, obj)
		},
	)

	// 헤더 설정
	t.table.ShowHeaderRow = true
	t.table.ShowHeaderColumn = false
	t.table.CreateHeader = func() fyne.CanvasObject {
		return widget.NewLabel("")
	}
	t.table.UpdateHeader = func(id widget.TableCellID, obj fyne.CanvasObject) {
		label := obj.(*widget.Label)
		if id.Col >= 0 && id.Col < len(t.config.Columns) {
			label.SetText(t.config.Columns[id.Col].Header)
		}
	}

	// 셀 선택 이벤트 (클릭 시 편집 모드 진입)
	t.table.OnSelected = func(id widget.TableCellID) {
		// Refresh 중이면 무시
		if t.isRefreshing {
			return
		}

		if t.config.GetRowCount == nil || id.Row < 0 || id.Row >= t.config.GetRowCount() {
			t.table.UnselectAll()
			return
		}
		if t.config.GetCellConfig == nil {
			t.table.UnselectAll()
			return
		}

		cellConfig := t.config.GetCellConfig(id.Row, id.Col)

		// Entry와 Select 타입만 편집 모드 진입
		if cellConfig.Type == CellTypeEntry || cellConfig.Type == CellTypeSelect {
			// 이미 같은 셀을 편집 중이면 무시
			if t.editingCell != nil && t.editingCell.Row == id.Row && t.editingCell.Col == id.Col {
				t.table.UnselectAll()
				return
			}

			// 편집 모드 설정
			t.editingCell = &widget.TableCellID{Row: id.Row, Col: id.Col}
			t.isRefreshing = true
			t.table.UnselectAll()
			t.table.Refresh()
			t.isRefreshing = false
		} else {
			// 다른 타입은 선택만 해제
			t.table.UnselectAll()
		}
	}

	// 초기 컬럼 너비 설정
	t.updateColumnWidths(900)

	// 테이블 외곽선
	borderColor := color.RGBA{R: 200, G: 200, B: 200, A: 255}
	borderTop := canvas.NewRectangle(borderColor)
	borderTop.SetMinSize(fyne.NewSize(0, 1))
	borderBottom := canvas.NewRectangle(borderColor)
	borderBottom.SetMinSize(fyne.NewSize(0, 1))
	borderLeft := canvas.NewRectangle(borderColor)
	borderLeft.SetMinSize(fyne.NewSize(1, 0))
	borderRight := canvas.NewRectangle(borderColor)
	borderRight.SetMinSize(fyne.NewSize(1, 0))

	t.borderedTable = container.NewBorder(
		borderTop, borderBottom, borderLeft, borderRight,
		t.table,
	)
}

// updateCell 셀 업데이트
func (t *EditableTable) updateCell(id widget.TableCellID, obj fyne.CanvasObject) {
	stack := obj.(*fyne.Container)

	// 모든 위젯 숨기기
	for _, child := range stack.Objects {
		child.Hide()
	}

	// 배경색을 현재 테마에 맞게 업데이트
	bg := stack.Objects[0].(*canvas.Rectangle)
	bg.FillColor = theme.Color(theme.ColorNameBackground)
	bg.Refresh()
	bg.Show()

	if t.config.GetRowCount == nil || id.Row < 0 || id.Row >= t.config.GetRowCount() {
		return
	}
	if t.config.GetCellConfig == nil {
		return
	}

	cellConfig := t.config.GetCellConfig(id.Row, id.Col)

	// 현재 편집 중인 셀인지 확인
	isEditing := t.editingCell != nil && t.editingCell.Row == id.Row && t.editingCell.Col == id.Col

	// 인덱스: 0=bg, 1=Label, 2=Entry, 3=Select, 4=Check, 5=Button, 6=Hyperlink
	switch cellConfig.Type {
	case CellTypeLabel:
		label := stack.Objects[1].(*widget.Label)
		label.SetText(cellConfig.Text)
		label.Show()

	case CellTypeEntry:
		if isEditing {
			// 편집 모드: Entry 표시
			entry := stack.Objects[2].(*cellEntry)
			entry.SetText(cellConfig.Text)
			entry.OnChanged = func(s string) {
				if cellConfig.OnChanged != nil {
					cellConfig.OnChanged(s)
				}
			}
			// Enter 키 입력 시 편집 종료
			entry.OnSubmitted = func(s string) {
				t.finishEditing()
			}
			// Esc 키 입력 시 편집 종료
			entry.onEscape = func() {
				t.finishEditing()
			}
			entry.Show()
		} else {
			// 일반 모드: Label 표시, 클릭 시 편집 모드
			label := stack.Objects[1].(*widget.Label)
			text := cellConfig.Text
			if text == "" {
				text = "-"
			}
			label.SetText(text)
			label.Show()
		}

	case CellTypeSelect:
		if isEditing {
			// 편집 모드: Select 표시
			sel := stack.Objects[3].(*widget.Select)
			// OnChanged를 먼저 nil로 설정하여 SetSelected 시 트리거 방지
			sel.OnChanged = nil
			sel.Options = cellConfig.Options
			sel.SetSelected(cellConfig.Selected)
			// 현재 선택값 저장 (변경 감지용)
			currentSelected := cellConfig.Selected
			sel.OnChanged = func(s string) {
				// 실제로 값이 변경된 경우에만 처리
				if s != currentSelected {
					if cellConfig.OnChanged != nil {
						cellConfig.OnChanged(s)
					}
					// Select 선택 후 편집 종료
					t.finishEditing()
				}
			}
			sel.Show()
		} else {
			// 일반 모드: Label 표시
			label := stack.Objects[1].(*widget.Label)
			text := cellConfig.Selected
			if text == "" {
				text = "-"
			}
			label.SetText(text)
			label.Show()
		}

	case CellTypeCheck:
		// Check는 항상 표시
		check := stack.Objects[4].(*widget.Check)
		check.Checked = cellConfig.Checked
		check.OnChanged = func(b bool) {
			if cellConfig.OnChanged != nil {
				cellConfig.OnChanged(b)
			}
		}
		check.Show()

	case CellTypeButton:
		// Button은 항상 표시
		btn := stack.Objects[5].(*widget.Button)
		btn.SetText("")
		if cellConfig.Icon != nil {
			btn.SetIcon(cellConfig.Icon)
		}
		btn.OnTapped = cellConfig.OnTapped
		btn.Show()

	case CellTypeHyperlink:
		// Hyperlink는 항상 표시
		link := stack.Objects[6].(*widget.Hyperlink)
		text := cellConfig.Text
		if text == "" {
			text = "-"
		}
		link.SetText(text)
		link.OnTapped = cellConfig.OnTapped
		link.Show()
	}
}

// updateColumnWidths 컬럼 너비 업데이트
func (t *EditableTable) updateColumnWidths(totalWidth float32) {
	if totalWidth <= 0 {
		return
	}

	// 고정 너비와 비율 계산
	fixedWidth := float32(0)
	totalRatio := float32(0)

	for _, col := range t.config.Columns {
		if col.Width > 0 {
			fixedWidth += col.Width
		} else {
			totalRatio += col.WidthRatio
		}
	}

	// 스크롤바 여백
	scrollbarWidth := float32(32)
	flexibleWidth := totalWidth - fixedWidth - scrollbarWidth

	// 컬럼 너비 설정
	for i, col := range t.config.Columns {
		if col.Width > 0 {
			t.table.SetColumnWidth(i, col.Width)
		} else if totalRatio > 0 {
			t.table.SetColumnWidth(i, flexibleWidth*(col.WidthRatio/totalRatio))
		}
	}
}

// StartEditing 편집 모드 시작
func (t *EditableTable) StartEditing(row, col int) {
	t.editingCell = &widget.TableCellID{Row: row, Col: col}
	t.table.Refresh()
}

// finishEditing 편집 종료
func (t *EditableTable) finishEditing() {
	t.editingCell = nil
	t.isRefreshing = true
	t.table.Refresh()
	t.isRefreshing = false
}

// Resize 크기 변경 시 컬럼 너비 재계산
func (t *EditableTable) Resize(size fyne.Size) {
	if size.Width != t.lastWidth && size.Width > 0 {
		t.lastWidth = size.Width
		t.updateColumnWidths(size.Width)
	}
	t.BaseWidget.Resize(size)
}

// CreateRenderer 렌더러 생성
func (t *EditableTable) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.borderedTable)
}

// Refresh 테이블 새로고침
func (t *EditableTable) Refresh() {
	t.table.Refresh()
}

// Content 테이블 위젯 반환
func (t *EditableTable) Content() fyne.CanvasObject {
	return t
}

// SetColumnWidth 컬럼 너비 설정
func (t *EditableTable) SetColumnWidth(col int, width float32) {
	t.table.SetColumnWidth(col, width)
}
