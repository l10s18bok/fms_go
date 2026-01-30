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
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// 배포 이력 탭을 구현합니다. (PRD 3.3.4 기준)
type HistoryTab struct {
	window      fyne.Window
	historyRepo repository.HistoryRepository // SQLite 기반 배포 이력 저장소
	deviceTab   *DeviceTab
	content     fyne.CanvasObject

	// UI 컴포넌트
	historyTable *component.PagedTable // 이력 테이블 (공통 컴포넌트)
	typeFilter   *widget.Select        // 유형 필터
	searchBox    *component.SearchBox  // 검색 컴포넌트 (공통)

	// 데이터 (DB 페이지네이션)
	pageData      []*model.DeployHistory // 현재 페이지 데이터만
	totalCount    int                    // 필터/검색 적용 후 전체 건수
	currentFilter string                 // 현재 유형 필터값
	searchKeyword string                 // 검색 키워드
}

// 새로운 배포 이력 탭을 생성합니다.
func NewHistoryTab(window fyne.Window, historyRepo repository.HistoryRepository) *HistoryTab {
	tab := &HistoryTab{
		window:      window,
		historyRepo: historyRepo,
		pageData:    []*model.DeployHistory{},
	}
	tab.createUI()
	tab.loadHistory()
	return tab
}

// 이력 탭의 UI를 생성합니다. (PRD 3.3.4 기준 - 상세 패널 없음)
func (h *HistoryTab) createUI() {
	// 이력 테이블 패널만 사용 (상세 패널 제거)
	h.content = h.createHistoryTablePanel()
}

// 이력 테이블 패널을 생성합니다. (PRD 3.3.4 기준)
// 컬럼: 선택, 시간, 장비명, 장비 IP, 유형, 버전, 결과, 메시지
func (h *HistoryTab) createHistoryTablePanel() fyne.CanvasObject {
	// 메시지 컬럼 너비: 긴 텍스트를 표시할 수 있도록 충분히 넓게 설정
	// 다른 컬럼 합계: 50+150+120+150+80+80+60 = 690
	messageColumnWidth := float32(600) // 충분히 넓은 기본값

	h.historyTable = component.NewPagedTable(component.PagedTableConfig{
		Columns: []component.ColumnDef{
			{Header: "선택", Width: 50},
			{Header: "시간", Width: 150},
			{Header: "장비명", Width: 120},
			{Header: "장비 IP", Width: 150},
			{Header: "유형", Width: 80},
			{Header: "버전", Width: 80},
			{Header: "결과", Width: 60},
			{Header: "메시지", Width: messageColumnWidth},
		},
		PageSize: 15,
		OnCellUpdate: func(row int, col int, cell fyne.CanvasObject) {
			h.updateHistoryCell(row, col, cell)
		},
		OnRowSelected:    func(row int) {},
		OnRowDoubleClick: func(row int) {},
		OnPageLoad: func(page, pageSize int) int {
			return h.loadPage(page, pageSize)
		},
	})

	// 유형 필터 드롭다운 (전체, 패키지, 방화벽 룰)
	h.typeFilter = widget.NewSelect([]string{"전체", "패키지", "방화벽 룰"}, func(selected string) {
		h.applyFilter()
	})
	h.typeFilter.SetSelected("전체") // 기본값: 전체

	// 검색 컴포넌트 (공통)
	h.searchBox = component.NewSearchBox(component.SearchBoxConfig{
		Placeholder: "시간, 장비명, IP, 유형, 버전, 메시지 검색",
		Width:       300,
		OnSearch: func(text string) {
			h.onSearch()
		},
	})

	// 삭제 버튼 (휴지통 아이콘 + 삭제, 빨간 배경, 흰색 텍스트)
	deleteBtn := component.NewCustomButton("삭제", theme.DeleteIcon(), nil, themes.Colors["red"], func() {
		h.onDeleteHistory()
	})

	// 상단 헤더 (PRD 3.3.4: [유형선택 ▼] 검색: [____] [찾기] [선택삭제])
	headerLine := container.NewBorder(
		nil, nil,
		container.NewHBox(
			h.typeFilter,
			h.searchBox.Content(),
		),
		container.NewHBox(deleteBtn),
		nil,
	)

	// 헤더에 상하좌우 10px 패딩 적용
	paddedHeader := container.New(layout.NewCustomPaddedLayout(10, 10, 10, 10), headerLine)

	return container.NewBorder(
		paddedHeader,
		nil,
		nil, nil,
		h.historyTable.Content(),
	)
}

// 이력 테이블 셀을 업데이트합니다. (PRD 3.3.4 컬럼 기준)
func (h *HistoryTab) updateHistoryCell(row int, col int, cell fyne.CanvasObject) {
	label := cell.(*widget.Label)

	// row는 전체 데이터 인덱스 (dataIndex = currentPage * pageSize + pageRow)
	// pageData는 현재 페이지 데이터만 보유하므로 페이지 내 인덱스로 변환
	pageRow := row - h.historyTable.GetCurrentPage()*h.historyTable.GetPageSize()
	if pageRow < 0 || pageRow >= len(h.pageData) {
		label.SetText("")
		return
	}

	history := h.pageData[pageRow]

	switch col {
	case 1: // 시간
		label.SetText(history.GetTimestampString())
	case 2: // 장비명
		label.SetText(history.DeviceName)
	case 3: // 장비 IP
		label.SetText(history.DeviceIP)
	case 4: // 유형
		label.SetText(model.GetHistoryTypeText(history.Type))
	case 5: // 버전
		if history.Type == model.HistoryTypeProgram {
			label.SetText(history.ProgramVer)
		} else {
			label.SetText(history.TemplateVer)
		}
	case 6: // 결과
		label.SetText(model.GetDeployStatusText(history.Status))
	case 7: // 메시지
		if history.Message != "" {
			// 첫 번째 줄만 표시 (줄바꿈 이전까지)
			msg := history.Message
			if idx := strings.Index(msg, "\n"); idx != -1 {
				msg = msg[:idx]
			}
			label.SetText(msg)
		} else {
			label.SetText("-")
		}
	}
}

// 검색을 실행합니다. (PRD 3.3.4 검색 기능)
func (h *HistoryTab) onSearch() {
	keyword := strings.TrimSpace(h.searchBox.GetText())

	prevKeyword := h.searchKeyword
	h.searchKeyword = keyword
	total := h.loadPage(0, h.historyTable.GetPageSize())

	// 검색 결과가 없으면 이전 상태 유지
	if total == 0 && keyword != "" {
		dialog.ShowInformation("검색 결과", "검색 결과가 없습니다.", h.window)
		h.searchKeyword = prevKeyword
		h.loadPage(0, h.historyTable.GetPageSize())
	}

	if h.historyTable != nil {
		h.historyTable.SetData(h.totalCount)
	}
}

// 필터와 검색을 적용합니다. (DB WHERE 절로 처리)
func (h *HistoryTab) applyFilter() {
	selected := h.typeFilter.Selected

	switch selected {
	case "방화벽 룰":
		h.currentFilter = model.HistoryTypeFirewall
	case "패키지":
		h.currentFilter = model.HistoryTypeProgram
	default:
		h.currentFilter = ""
	}

	total := h.loadPage(0, h.historyTable.GetPageSize())
	if h.historyTable != nil {
		h.historyTable.SetData(total)
	}
}

// GetExportData 현재 필터/검색 조건에 맞는 전체 배포이력을 반환합니다.
func (h *HistoryTab) GetExportData() ([]*model.DeployHistory, error) {
	if h.currentFilter == "" && h.searchKeyword == "" {
		return h.historyRepo.GetAll()
	}
	// 필터/검색 조건이 있으면 Limit을 크게 설정하여 전체 조회
	req := model.PageRequest{
		Offset:  0,
		Limit:   100000,
		Filter:  h.currentFilter,
		Keyword: h.searchKeyword,
	}
	result, err := h.historyRepo.GetPage(req)
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

// IsFiltered 필터 또는 검색이 적용되어 있는지 반환합니다.
func (h *HistoryTab) IsFiltered() bool {
	return h.currentFilter != "" || h.searchKeyword != ""
}

// 탭의 컨텐츠를 반환합니다.
func (h *HistoryTab) Content() fyne.CanvasObject {
	return h.content
}

// DB에서 해당 페이지 데이터를 조회합니다.
func (h *HistoryTab) loadPage(page, pageSize int) int {
	req := model.PageRequest{
		Offset:  page * pageSize,
		Limit:   pageSize,
		Filter:  h.currentFilter,
		Keyword: h.searchKeyword,
	}
	result, err := h.historyRepo.GetPage(req)
	if err != nil {
		h.pageData = []*model.DeployHistory{}
		h.totalCount = 0
		return 0
	}
	h.pageData = result.Items
	h.totalCount = result.TotalCount
	return result.TotalCount
}

// 저장소에서 배포 이력을 로드합니다. (첫 페이지부터)
func (h *HistoryTab) loadHistory() {
	total := h.loadPage(0, h.historyTable.GetPageSize())
	fyne.Do(func() {
		h.historyTable.SetData(total)
	})
}

// 배포 이력을 새로고침합니다.
func (h *HistoryTab) RefreshHistory() {
	h.loadHistory()

	// 2초 후 자동으로 사라지는 다이얼로그 표시
	infoDialog := dialog.NewInformation("완료", "새로고침 완료", h.window)
	infoDialog.Show()
	go func() {
		time.Sleep(2 * time.Second)
		fyne.Do(func() {
			infoDialog.Hide()
		})
	}()
}

// 배포 이력만 다시 로드합니다. (다이얼로그 없이)
func (h *HistoryTab) ReloadHistory() {
	h.loadHistory()
}

// 선택된 이력을 삭제합니다. (PagedTable 체크된 행 사용)
func (h *HistoryTab) onDeleteHistory() {
	checkedRows := h.historyTable.GetCheckedRows()
	if len(checkedRows) == 0 {
		dialog.ShowInformation("알림", "삭제할 이력을 선택해주세요.", h.window)
		return
	}

	dialog.ShowConfirm("확인", fmt.Sprintf("선택한 %d개 이력을 삭제하시겠습니까?", len(checkedRows)), func(ok bool) {
		if !ok {
			return
		}

		// 삭제할 이력의 장비 IP 수집
		deviceIPs := make(map[string]bool)
		pageOffset := h.historyTable.GetCurrentPage() * h.historyTable.GetPageSize()
		for _, row := range checkedRows {
			pageRow := row - pageOffset
			if pageRow >= 0 && pageRow < len(h.pageData) {
				history := h.pageData[pageRow]
				deviceIPs[history.DeviceIP] = true
				if err := h.historyRepo.Delete(history.ID); err != nil {
					dialog.ShowError(err, h.window)
					return
				}
			}
		}

		// 해당 장비의 남은 이력이 있는지 확인
		for deviceIP := range deviceIPs {
			h.resetDeviceDeployStatusIfNoHistory(deviceIP)
		}

		h.historyTable.ClearChecked()
		dialog.ShowInformation("성공", "선택한 배포 이력이 삭제되었습니다.", h.window)
		h.loadHistory()
	}, h.window)
}

// 새로운 배포 이력을 추가합니다.
func (h *HistoryTab) AddHistory(history *model.DeployHistory) error {
	if err := h.historyRepo.Save(history); err != nil {
		return err
	}
	h.loadHistory()
	return nil
}

// 장비 탭 참조를 설정합니다.
func (h *HistoryTab) SetDeviceTab(deviceTab *DeviceTab) {
	h.deviceTab = deviceTab
}

// 해당 장비의 이력이 없으면 배포 상태를 초기화합니다.
func (h *HistoryTab) resetDeviceDeployStatusIfNoHistory(deviceIP string) {
	if h.deviceTab == nil {
		return
	}

	// 해당 장비의 남은 이력이 있는지 확인
	histories, err := h.historyRepo.GetAll()
	if err != nil {
		return
	}

	hasHistory := false
	for _, history := range histories {
		if history.DeviceIP == deviceIP {
			hasHistory = true
			break
		}
	}

	// 이력이 없으면 장비의 배포 상태 초기화
	if !hasHistory {
		h.deviceTab.ResetDeviceDeployStatus(deviceIP)
	}
}
