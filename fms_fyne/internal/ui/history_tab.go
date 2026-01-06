package ui

import (
	"sort"
	"time"

	"fms/internal/model"
	"fms/internal/storage"
	"fms/internal/themes"
	"fms/internal/ui/component"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// 배포 이력 탭을 구현합니다.
type HistoryTab struct {
	window    fyne.Window
	store     *storage.JSONStore
	deviceTab *DeviceTab
	content   fyne.CanvasObject

	// UI 컴포넌트
	historyTable *widget.Table // 이력 테이블
	detailTable  *widget.Table // 상세 결과 테이블

	// 데이터
	histories            []*model.DeployHistory
	selectedHistoryIndex int
	selectedHistory      *model.DeployHistory
}

// 새로운 배포 이력 탭을 생성합니다.
func NewHistoryTab(window fyne.Window, store *storage.JSONStore) *HistoryTab {
	tab := &HistoryTab{
		window:               window,
		store:                store,
		histories:            []*model.DeployHistory{},
		selectedHistoryIndex: -1,
	}
	tab.createUI()
	tab.loadHistory()
	return tab
}

// 이력 탭의 UI를 생성합니다.
func (h *HistoryTab) createUI() {
	// 상단: 이력 테이블
	historyPanel := h.createHistoryTablePanel()

	// 하단: 상세 결과
	detailPanel := h.createDetailPanel()

	// 상하 분할 (60% : 40%)
	split := container.NewVSplit(historyPanel, detailPanel)
	split.Offset = 0.6

	h.content = split
}

// 이력 테이블 패널을 생성합니다.
func (h *HistoryTab) createHistoryTablePanel() fyne.CanvasObject {
	// 테이블 헤더
	headers := []string{"시간", "장비", "템플릿", "결과"}

	// 테이블 생성
	h.historyTable = widget.NewTable(
		// 크기 함수
		func() (int, int) {
			return len(h.histories) + 1, len(headers)
		},
		// 셀 생성 함수
		func() fyne.CanvasObject {
			return widget.NewLabel("                    ")
		},
		// 셀 업데이트 함수
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)

			if id.Row == 0 {
				label.SetText(headers[id.Col])
				label.TextStyle = fyne.TextStyle{Bold: true}
			} else {
				idx := id.Row - 1
				if idx < len(h.histories) {
					history := h.histories[idx]
					switch id.Col {
					case 0:
						label.SetText(history.GetTimestampString())
					case 1:
						label.SetText(history.DeviceIP)
					case 2:
						label.SetText(history.TemplateVer)
					case 3:
						label.SetText(model.GetDeployStatusText(history.Status))
					}
				}
			}
		},
	)

	// 열 너비 설정
	h.historyTable.SetColumnWidth(0, 180) // 시간
	h.historyTable.SetColumnWidth(1, 150) // 장비
	h.historyTable.SetColumnWidth(2, 100) // 템플릿
	h.historyTable.SetColumnWidth(3, 100) // 결과

	// 이력 선택 시 상세 표시
	h.historyTable.OnSelected = func(id widget.TableCellID) {
		if id.Row > 0 { // 헤더 제외
			h.selectedHistoryIndex = id.Row - 1
			if h.selectedHistoryIndex < len(h.histories) {
				h.selectedHistory = h.histories[h.selectedHistoryIndex]
				h.detailTable.Refresh()
			}
		}
	}

	// 삭제 버튼 (투명 배경 + 빨간 텍스트)
	deleteBtn := component.NewCustomButton("이력 삭제", nil, themes.Colors["red"], nil, func() {
		h.onDeleteHistory()
	})

	// 전체 삭제 버튼 (투명 배경 + 빨간 텍스트)
	clearBtn := component.NewCustomButton("전체 삭제", nil, themes.Colors["red"], nil, func() {
		h.onClearHistory()
	})

	// 스크롤 가능한 테이블
	scrollableTable := container.NewScroll(h.historyTable)

	// 상단 헤더 (배포 이력 라벨)
	header := widget.NewLabel("배포 이력")

	// 하단 버튼 영역 (이력삭제 좌측 + margin, 전체삭제 우측 + margin)
	bottomButtons := container.NewPadded(container.NewBorder(nil, nil, deleteBtn, clearBtn, nil))

	return container.NewBorder(
		header,
		bottomButtons,
		nil, nil,
		scrollableTable,
	)
}

// 상세 결과 패널을 생성합니다.
func (h *HistoryTab) createDetailPanel() fyne.CanvasObject {
	// 상세 테이블 헤더
	headers := []string{"규칙", "상태", "사유"}

	// 테이블 생성
	h.detailTable = widget.NewTable(
		// 크기 함수
		func() (int, int) {
			if h.selectedHistory == nil {
				return 1, len(headers) // 헤더만
			}
			return len(h.selectedHistory.Results) + 1, len(headers)
		},
		// 셀 생성 함수
		func() fyne.CanvasObject {
			return widget.NewLabel("                              ")
		},
		// 셀 업데이트 함수
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)

			if id.Row == 0 {
				label.SetText(headers[id.Col])
				label.TextStyle = fyne.TextStyle{Bold: true}
			} else if h.selectedHistory != nil {
				idx := id.Row - 1
				if idx < len(h.selectedHistory.Results) {
					result := h.selectedHistory.Results[idx]
					switch id.Col {
					case 0:
						// 규칙이 너무 길면 축약 (Text 필드 사용)
						text := result.Text
						if text == "" {
							text = result.Rule // Text가 비어있으면 Rule 사용
						}
						if len(text) > 60 {
							text = text[:60] + "..."
						}
						label.SetText(text)
					case 1:
						label.SetText(model.GetRuleStatusText(result.Status))
					case 2:
						label.SetText(model.GetReasonText(result.Reason))
					}
				}
			}
		},
	)

	// 열 너비 설정
	h.detailTable.SetColumnWidth(0, 550) // 규칙
	h.detailTable.SetColumnWidth(1, 60)  // 상태
	h.detailTable.SetColumnWidth(2, 350) // 사유

	// 스크롤 가능한 테이블
	scrollableTable := container.NewScroll(h.detailTable)

	return container.NewBorder(
		widget.NewSeparator(),
		nil,
		nil, nil,
		container.NewBorder(
			widget.NewLabel("상세 결과 (이력을 선택하면 표시됩니다)"),
			nil, nil, nil,
			scrollableTable,
		),
	)
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

	h.selectedHistoryIndex = -1
	h.selectedHistory = nil

	// UI 업데이트는 메인 스레드에서 실행
	fyne.Do(func() {
		h.historyTable.Refresh()
		h.detailTable.Refresh()
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

// 선택된 이력을 삭제합니다.
func (h *HistoryTab) onDeleteHistory() {
	if h.selectedHistoryIndex < 0 || h.selectedHistory == nil {
		dialog.ShowInformation("알림", "삭제할 이력을 선택해주세요.", h.window)
		return
	}

	dialog.ShowConfirm("확인", "선택한 배포 이력을 삭제하시겠습니까?", func(ok bool) {
		if !ok {
			return
		}

		// 삭제할 이력의 장비 IP 저장
		deviceIP := h.selectedHistory.DeviceIP

		if err := h.store.DeleteHistory(h.selectedHistory.ID); err != nil {
			dialog.ShowError(err, h.window)
			return
		}

		// 해당 장비의 남은 이력이 있는지 확인
		h.resetDeviceDeployStatusIfNoHistory(deviceIP)

		dialog.ShowInformation("성공", "배포 이력이 삭제되었습니다.", h.window)
		h.loadHistory()
	}, h.window)
}

// 모든 이력을 삭제합니다.
func (h *HistoryTab) onClearHistory() {
	if len(h.histories) == 0 {
		dialog.ShowInformation("알림", "삭제할 이력이 없습니다.", h.window)
		return
	}

	dialog.ShowConfirm("경고", "모든 배포 이력을 삭제하시겠습니까?", func(ok bool) {
		if !ok {
			return
		}

		// 이력에 있는 모든 장비 IP 수집
		deviceIPs := make(map[string]bool)
		for _, history := range h.histories {
			deviceIPs[history.DeviceIP] = true
		}

		if err := h.store.ClearHistory(); err != nil {
			dialog.ShowError(err, h.window)
			return
		}

		// 모든 장비의 배포 상태 초기화
		if h.deviceTab != nil {
			for deviceIP := range deviceIPs {
				h.deviceTab.ResetDeviceDeployStatus(deviceIP)
			}
		}

		dialog.ShowInformation("성공", "모든 배포 이력이 삭제되었습니다.", h.window)
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
