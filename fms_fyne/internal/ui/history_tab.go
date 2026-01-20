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
	"fyne.io/fyne/v2/widget"
)

// 배포 이력 탭을 구현합니다. (PRD 3.3.4 기준)
type HistoryTab struct {
	window    fyne.Window
	store     *storage.JSONStore
	deviceTab *DeviceTab
	content   fyne.CanvasObject

	// UI 컴포넌트
	historyTable *component.PagedTable // 이력 테이블 (공통 컴포넌트)
	typeFilter   *widget.Select        // 유형 필터
	searchBox    *component.SearchBox  // 검색 컴포넌트 (공통)

	// 데이터
	histories         []*model.DeployHistory
	filteredHistories []*model.DeployHistory // 필터링된 이력
	searchKeyword     string                 // 검색 키워드
}

// 새로운 배포 이력 탭을 생성합니다.
func NewHistoryTab(window fyne.Window, store *storage.JSONStore) *HistoryTab {
	tab := &HistoryTab{
		window:    window,
		store:     store,
		histories: []*model.DeployHistory{},
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
// 컬럼: 선택, 시간, 장비명, 장비 IP, 유형, 버전, 결과
func (h *HistoryTab) createHistoryTablePanel() fyne.CanvasObject {
	h.historyTable = component.NewPagedTable(component.PagedTableConfig{
		Columns: []component.ColumnDef{
			{Header: "선택", Width: 50},
			{Header: "시간", Width: 150},
			{Header: "장비명", Width: 150},
			{Header: "장비 IP", Width: 225},
			{Header: "유형", Width: 100},
			{Header: "버전", Width: 80},
			{Header: "결과", Width: 60},
		},
		PageSize: 15,
		OnCellUpdate: func(row int, col int, cell fyne.CanvasObject) {
			h.updateHistoryCell(row, col, cell)
		},
		OnRowSelected:    func(row int) {},
		OnRowDoubleClick: func(row int) {},
	})

	// 유형 필터 드롭다운 (전체, 패키지, 방화벽 룰)
	h.typeFilter = widget.NewSelect([]string{"전체", "패키지", "방화벽 룰"}, func(selected string) {
		h.applyFilter()
	})
	h.typeFilter.SetSelected("전체") // 기본값: 전체

	// 검색 컴포넌트 (공통)
	h.searchBox = component.NewSearchBox(component.SearchBoxConfig{
		Placeholder: "시간, 장비명, IP, 버전 검색",
		Width:       200,
		OnSearch: func(text string) {
			h.onSearch()
		},
	})

	// 선택 삭제 버튼
	deleteBtn := component.NewCustomButton("선택삭제", nil, themes.Colors["red"], nil, func() {
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

	if row >= len(h.filteredHistories) {
		label.SetText("")
		return
	}

	history := h.filteredHistories[row]

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
	}
}

// 검색을 실행합니다. (PRD 3.3.4 검색 기능)
func (h *HistoryTab) onSearch() {
	h.searchKeyword = strings.TrimSpace(h.searchBox.GetText())
	h.applyFilter()

	// 검색 결과가 없으면 다이얼로그 표시
	if len(h.filteredHistories) == 0 && h.searchKeyword != "" {
		dialog.ShowInformation("검색 결과", "검색 결과가 없습니다.", h.window)
	}
}

// 필터와 검색을 적용합니다.
func (h *HistoryTab) applyFilter() {
	selected := h.typeFilter.Selected

	// 1단계: 유형 필터 적용
	var typeFiltered []*model.DeployHistory
	if selected == "전체" || selected == "" {
		// 전체: 모든 이력 표시
		typeFiltered = h.histories
	} else if selected == "방화벽 룰" {
		for _, history := range h.histories {
			// 방화벽 룰: "firewall", "template"(레거시), 빈 문자열(레거시)
			if history.Type == model.HistoryTypeFirewall || history.Type == "template" || history.Type == "" {
				typeFiltered = append(typeFiltered, history)
			}
		}
	} else { // 패키지
		for _, history := range h.histories {
			if history.Type == model.HistoryTypeProgram {
				typeFiltered = append(typeFiltered, history)
			}
		}
	}

	// 2단계: 검색 키워드 적용
	if h.searchKeyword == "" {
		h.filteredHistories = typeFiltered
	} else {
		keyword := strings.ToLower(h.searchKeyword)
		h.filteredHistories = []*model.DeployHistory{}
		for _, history := range typeFiltered {
			// 검색 대상: 시간, 장비명, 장비 IP, 버전 (부분 일치)
			if strings.Contains(strings.ToLower(history.GetTimestampString()), keyword) ||
				strings.Contains(strings.ToLower(history.DeviceName), keyword) ||
				strings.Contains(strings.ToLower(history.DeviceIP), keyword) ||
				strings.Contains(strings.ToLower(history.TemplateVer), keyword) ||
				strings.Contains(strings.ToLower(history.ProgramVer), keyword) {
				h.filteredHistories = append(h.filteredHistories, history)
			}
		}
	}

	if h.historyTable != nil {
		h.historyTable.SetData(len(h.filteredHistories))
	}
}

// 탭의 컨텐츠를 반환합니다.
func (h *HistoryTab) Content() fyne.CanvasObject {
	return h.content
}

// 저장소에서 배포 이력을 로드합니다.
func (h *HistoryTab) loadHistory() {
	histories, err := h.store.GetAllHistory()
	if err != nil {
		fyne.Do(func() {
			dialog.ShowError(err, h.window)
		})
		return
	}

	h.histories = histories

	// ID 내림차순 정렬 (최신순 - ID가 클수록 최신)
	sort.Slice(h.histories, func(i, j int) bool {
		return h.histories[i].ID > h.histories[j].ID
	})

	// UI 업데이트는 메인 스레드에서 실행
	fyne.Do(func() {
		// 필터 적용
		h.applyFilter()
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
		for _, row := range checkedRows {
			if row < len(h.filteredHistories) {
				history := h.filteredHistories[row]
				deviceIPs[history.DeviceIP] = true
				if err := h.store.DeleteHistory(history.ID); err != nil {
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
	if err := h.store.SaveHistory(history); err != nil {
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
	histories, err := h.store.GetAllHistory()
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
