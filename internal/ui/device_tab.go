package ui

import (
	"fmt"
	"image/color"
	"net"
	"sort"
	"time"

	"fms/internal/deploy"
	"fms/internal/model"
	"fms/internal/storage"
	"fms/internal/ui/component"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// 장비 관리 탭을 구현합니다.
type DeviceTab struct {
	window      fyne.Window
	store       *storage.JSONStore
	templateTab *TemplateTab
	historyTab  *HistoryTab
	content     fyne.CanvasObject

	// UI 컴포넌트
	templateSelect *widget.Select // 배포 템플릿 선택
	deviceTable    *widget.Table  // 장비 테이블
	tableContainer *fyne.Container

	// 상태 요약 표시
	statusGreenLabel  *widget.Label
	statusYellowLabel *widget.Label
	statusRedLabel    *widget.Label

	// 상세 정보 입력 필드
	ipEntry *widget.Entry

	// 에러 표시용 레이블
	ipErrorLabel *canvas.Text

	// 데이터
	firewalls           []*model.Firewall
	selectedDeviceIndex int
	checkedDevices      map[int]bool

	// 새로고침 상태
	isRefreshing bool
	refreshBtn   *widget.Button
}

// 새로운 장비 관리 탭을 생성합니다.
func NewDeviceTab(window fyne.Window, store *storage.JSONStore, templateTab *TemplateTab) *DeviceTab {
	tab := &DeviceTab{
		window:              window,
		store:               store,
		templateTab:         templateTab,
		firewalls:           []*model.Firewall{},
		selectedDeviceIndex: -1,
		checkedDevices:      make(map[int]bool),
	}
	tab.createUI()
	tab.loadFirewalls()
	return tab
}

// 장비 탭의 UI를 생성합니다.
func (d *DeviceTab) createUI() {
	// 상단: 배포 컨트롤
	topPanel := d.createDeployControlPanel()

	// 중앙: 장비 테이블
	tablePanel := d.createDeviceTablePanel()

	// 하단: 장비 상세 정보
	bottomPanel := d.createDetailPanel()

	// 전체 레이아웃
	d.content = container.NewBorder(
		topPanel,    // 상단 고정
		bottomPanel, // 하단 고정
		nil, nil,
		tablePanel, // 중앙 (자동 확장)
	)
}

// 배포 컨트롤 패널을 생성합니다.
func (d *DeviceTab) createDeployControlPanel() fyne.CanvasObject {
	// 배포 템플릿 선택
	d.templateSelect = widget.NewSelect([]string{}, func(selected string) {
		// 템플릿 선택 처리
	})
	d.templateSelect.PlaceHolder = "템플릿 선택..."

	// 상태 요약 레이블 생성
	d.statusGreenLabel = widget.NewLabel("0")
	d.statusYellowLabel = widget.NewLabel("0")
	d.statusRedLabel = widget.NewLabel("0")

	// 색상이 있는 상태 표시 (canvas.Text 사용)
	greenDot := canvas.NewText("●", color.RGBA{R: 0, G: 200, B: 0, A: 255})
	greenDot.TextSize = 21
	yellowDot := canvas.NewText("●", color.RGBA{R: 255, G: 200, B: 0, A: 255})
	yellowDot.TextSize = 21
	redDot := canvas.NewText("●", color.RGBA{R: 220, G: 20, B: 20, A: 255})
	redDot.TextSize = 21

	// 상태 요약 컨테이너
	statusSummary := container.NewHBox(
		greenDot, widget.NewLabel("연결:"), d.statusGreenLabel,
		widget.NewLabel(" "),
		yellowDot, widget.NewLabel("알수없음:"), d.statusYellowLabel,
		widget.NewLabel(" "),
		redDot, widget.NewLabel("연결안됨:"), d.statusRedLabel,
	)

	templateSelector := container.NewHBox(d.templateSelect, widget.NewLabel("  "), statusSummary)

	// 배포 버튼 (텍스트만, 진한 회색 커스텀 버튼)
	deployBtn := component.NewColoredButton("선택장비에 배포", component.ButtonDark, func() {
		d.onDeploy()
	})

	return container.NewVBox(
		container.NewBorder(nil, nil, templateSelector, deployBtn, nil),
		widget.NewSeparator(),
	)
}

// 장비 상태 요약을 업데이트합니다.
func (d *DeviceTab) updateStatusSummary() {
	greenCount := 0
	yellowCount := 0
	redCount := 0

	for _, fw := range d.firewalls {
		switch fw.ServerStatus {
		case model.ServerStatusRunning:
			greenCount++
		case model.ServerStatusStop:
			redCount++
		default:
			yellowCount++
		}
	}

	d.statusGreenLabel.SetText(fmt.Sprintf("%d", greenCount))
	d.statusYellowLabel.SetText(fmt.Sprintf("%d", yellowCount))
	d.statusRedLabel.SetText(fmt.Sprintf("%d", redCount))
}

// 장비 테이블 패널을 생성합니다.
func (d *DeviceTab) createDeviceTablePanel() fyne.CanvasObject {
	// 테이블 생성
	d.deviceTable = widget.NewTable(
		// 크기 함수: 행 수, 열 수 반환
		func() (int, int) {
			return len(d.firewalls) + 1, 5 // +1 for header, 5 columns
		},
		// 셀 생성 함수
		func() fyne.CanvasObject {
			// 체크박스 열을 위한 canvas.Text
			checkText := canvas.NewText("", color.Black)
			checkText.TextSize = 16
			checkText.Alignment = fyne.TextAlignCenter

			// LED 표시용 canvas.Text (● 문자 사용) - 기본 노란색 (알 수 없음)
			ledText := canvas.NewText("●", color.RGBA{R: 255, G: 200, B: 0, A: 255})
			ledText.TextSize = 28
			ledText.Alignment = fyne.TextAlignCenter

			// 일반 Label
			label := widget.NewLabel("                ")

			return container.NewStack(checkText, label, ledText)
		},
		// 셀 업데이트 함수
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			cont := cell.(*fyne.Container)
			checkText := cont.Objects[0].(*canvas.Text)
			label := cont.Objects[1].(*widget.Label)
			ledText := cont.Objects[2].(*canvas.Text)
			headers := []string{"선택", "장비명(IP)", "서버상태", "배포상태", "버전"}

			// 기본적으로 LED 숨김
			ledText.Text = ""
			ledText.Hidden = true

			if id.Row == 0 {
				// 헤더
				checkText.Text = ""
				checkText.Hidden = true
				label.SetText(headers[id.Col])
				label.TextStyle = fyne.TextStyle{Bold: true}
				label.Show()
			} else {
				// 데이터
				idx := id.Row - 1
				if idx < len(d.firewalls) {
					fw := d.firewalls[idx]
					switch id.Col {
					case 0:
						// 체크박스 열: canvas.Text 사용
						label.Hide()
						checkText.Hidden = false
						if d.checkedDevices[fw.Index] {
							checkText.Text = "✔"
							checkText.Color = color.RGBA{R: 220, G: 20, B: 20, A: 255} // 빨간색
							checkText.TextStyle = fyne.TextStyle{Bold: true}
						} else {
							checkText.Text = "☐"
							checkText.Color = color.RGBA{R: 100, G: 100, B: 100, A: 255} // 회색
							checkText.TextStyle = fyne.TextStyle{Bold: false}
						}
						checkText.Refresh()
					case 2:
						// 서버상태 열: LED만 표시 (● 문자)
						checkText.Text = ""
						checkText.Hidden = true
						label.Hide()
						ledText.Text = "●"
						ledText.Hidden = false

						// 서버 상태에 따른 LED 색상 설정
						switch fw.ServerStatus {
						case model.ServerStatusRunning:
							// 정상: 녹색 LED
							ledText.Color = color.RGBA{R: 0, G: 200, B: 0, A: 255}
						case model.ServerStatusStop:
							// 정지: 빨간색 LED
							ledText.Color = color.RGBA{R: 220, G: 20, B: 20, A: 255}
						default:
							// 알 수 없음: 노란색 LED
							ledText.Color = color.RGBA{R: 255, G: 200, B: 0, A: 255}
						}
						ledText.Refresh()
					default:
						// 일반 열: Label 사용
						checkText.Text = ""
						checkText.Hidden = true
						label.Show()
						switch id.Col {
						case 1:
							label.SetText(fw.DeviceName)
						case 3:
							label.SetText(model.GetDeployStatusText(fw.DeployStatus))
						case 4:
							label.SetText(fw.Version)
						}
					}
				}
			}
		},
	)

	// 열 너비 설정
	d.deviceTable.SetColumnWidth(0, 50)  // 선택
	d.deviceTable.SetColumnWidth(1, 150) // 장비명
	d.deviceTable.SetColumnWidth(2, 80)  // 서버상태
	d.deviceTable.SetColumnWidth(3, 80)  // 배포상태
	d.deviceTable.SetColumnWidth(4, 80)  // 버전

	// 셀 선택 이벤트
	d.deviceTable.OnSelected = func(id widget.TableCellID) {
		// 선택 즉시 해제 (같은 셀 재클릭 가능하도록)
		defer d.deviceTable.UnselectAll()

		if id.Row == 0 {
			return // 헤더 클릭 무시
		}

		idx := id.Row - 1
		if idx >= 0 && idx < len(d.firewalls) {
			fw := d.firewalls[idx]

			if id.Col == 0 {
				// 체크박스 토글
				d.checkedDevices[fw.Index] = !d.checkedDevices[fw.Index]
				d.deviceTable.Refresh()
			} else {
				// 장비 선택 - 상세 정보 표시
				d.selectedDeviceIndex = idx
				d.showDeviceDetail(fw)
			}
		}
	}

	// 버튼들
	selectAllBtn := widget.NewButton("전체선택", func() {
		d.onSelectAll(true)
	})
	deselectAllBtn := widget.NewButton("전체해제", func() {
		d.onSelectAll(false)
	})
	deleteBtn := component.NewIconTextButton("삭제", theme.DeleteIcon(), component.ButtonDanger, func() {
		d.onDeleteDevices()
	})
	saveBtn := component.NewIconTextButton("저장", theme.ConfirmIcon(), component.ButtonPrimary, func() {
		d.onSaveDevices()
	})

	leftButtons := container.NewHBox(selectAllBtn, deselectAllBtn)
	rightButtons := container.NewHBox(saveBtn, deleteBtn)

	buttonBar := container.NewBorder(nil, nil, leftButtons, rightButtons, nil)

	// 스크롤 가능한 테이블
	scrollableTable := container.NewScroll(d.deviceTable)
	d.tableContainer = container.NewBorder(
		nil, buttonBar, nil, nil,
		scrollableTable,
	)

	return d.tableContainer
}

// 장비 상세 정보 패널을 생성합니다.
func (d *DeviceTab) createDetailPanel() fyne.CanvasObject {
	// 입력 필드 생성
	d.ipEntry = widget.NewEntry()
	d.ipEntry.SetPlaceHolder("192.168.1.1")
	d.ipEntry.OnSubmitted = func(s string) {
		d.onApplyDetail()
	}

	// 에러 레이블 생성 (빨간색 텍스트, 초기에 숨김)
	d.ipErrorLabel = canvas.NewText("", color.RGBA{R: 255, G: 0, B: 0, A: 255})
	d.ipErrorLabel.TextSize = 12
	d.ipErrorLabel.Hidden = true

	// 적용 버튼
	applyBtn := component.NewColoredButton("추가/수정", component.ButtonBlack, func() {
		d.onApplyDetail()
	})

	// IP 입력 필드와 에러 레이블을 VBox로 묶음
	ipContainer := container.NewVBox(d.ipEntry, d.ipErrorLabel)

	// 폼 레이아웃
	form := container.NewVBox(
		widget.NewSeparator(),
		container.NewHBox(
			widget.NewLabelWithStyle("장비 추가/수정", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel("(IP 주소를 입력하거나 테이블에서 선택 후 수정)"),
		),
		container.NewGridWithColumns(4,
			widget.NewLabel("장비 IP:"), ipContainer,
			widget.NewLabel(""), applyBtn,
		),
	)

	return form
}

// 탭의 컨텐츠를 반환합니다.
func (d *DeviceTab) Content() fyne.CanvasObject {
	return d.content
}

// 저장소에서 장비 목록을 로드합니다.
func (d *DeviceTab) loadFirewalls() {
	firewalls, err := d.store.GetAllFirewalls()
	if err != nil {
		dialog.ShowError(err, d.window)
		return
	}

	d.firewalls = firewalls

	// Index로 정렬
	sort.Slice(d.firewalls, func(i, j int) bool {
		return d.firewalls[i].Index < d.firewalls[j].Index
	})

	d.deviceTable.Refresh()

	// 상태 요약 업데이트
	d.updateStatusSummary()
}

// 템플릿 목록을 새로고침합니다.
func (d *DeviceTab) refreshTemplateList() {
	d.templateTab.RefreshTemplates()
	versions := d.templateTab.GetTemplateVersions()
	d.templateSelect.Options = versions
	d.templateSelect.Refresh()
}

// 선택된 장비의 상세 정보를 표시합니다.
func (d *DeviceTab) showDeviceDetail(fw *model.Firewall) {
	d.ipEntry.SetText(fw.DeviceName)
}

// 상세 정보를 선택된 장비에 적용하거나 새 장비를 추가합니다.
func (d *DeviceTab) onApplyDetail() {
	// 에러 레이블 초기화 (숨김)
	d.ipErrorLabel.Hidden = true
	d.ipErrorLabel.Text = ""

	// IP 검증 (IP 또는 IP:PORT 형식 허용)
	if d.ipEntry.Text == "" {
		d.ipErrorLabel.Text = "IP 주소를 입력해주세요"
		d.ipErrorLabel.Hidden = false
		d.ipErrorLabel.Refresh()
		return
	}
	if !isValidIPOrHostPort(d.ipEntry.Text) {
		d.ipErrorLabel.Text = "올바른 IP 주소 형식이 아닙니다 (예: 192.168.1.1 또는 192.168.1.1:8080)"
		d.ipErrorLabel.Hidden = false
		d.ipErrorLabel.Refresh()
		return
	}

	var fw *model.Firewall
	newIP := d.ipEntry.Text

	// 선택된 장비가 있으면 해당 장비 수정, 없으면 새 장비 생성
	if d.selectedDeviceIndex >= 0 && d.selectedDeviceIndex < len(d.firewalls) {
		// 기존 장비 수정 (테이블에서 선택한 장비)
		fw = d.firewalls[d.selectedDeviceIndex]
	} else {
		// 동일 IP 장비가 이미 있는지 확인
		existingIndex := -1
		for i, existing := range d.firewalls {
			if existing.DeviceName == newIP {
				existingIndex = i
				break
			}
		}

		if existingIndex >= 0 {
			// 동일 IP 장비가 있으면 해당 장비 수정
			fw = d.firewalls[existingIndex]
			d.selectedDeviceIndex = existingIndex
		} else {
			// 새 장비 생성
			fw = model.NewFirewall("")
			d.firewalls = append(d.firewalls, fw)
			d.selectedDeviceIndex = len(d.firewalls) - 1
		}
	}

	// 값 적용
	fw.DeviceName = d.ipEntry.Text

	// 장비 저장
	if err := d.store.SaveFirewall(fw); err != nil {
		dialog.ShowError(err, d.window)
		return
	}

	d.deviceTable.Refresh()

	// 상태 요약 업데이트
	d.updateStatusSummary()

	// 입력 필드 클리어
	d.ipEntry.SetText("")
	d.selectedDeviceIndex = -1
}

// 배포 시 호출됩니다.
func (d *DeviceTab) onDeploy() {
	// 템플릿 선택 확인
	if d.templateSelect.Selected == "" {
		dialog.ShowError(fmt.Errorf("배포할 템플릿을 선택해주세요"), d.window)
		return
	}

	// 템플릿 가져오기
	template := d.templateTab.GetTemplate(d.templateSelect.Selected)
	if template == nil {
		dialog.ShowError(fmt.Errorf("템플릿을 찾을 수 없습니다: %s", d.templateSelect.Selected), d.window)
		return
	}

	// 체크된 장비 수집
	checkedFirewalls := []*model.Firewall{}
	for _, fw := range d.firewalls {
		if d.checkedDevices[fw.Index] {
			checkedFirewalls = append(checkedFirewalls, fw)
		}
	}

	if len(checkedFirewalls) == 0 {
		dialog.ShowError(fmt.Errorf("배포할 장비를 선택해주세요"), d.window)
		return
	}

	// 진행률 다이얼로그 표시
	progressLabel := widget.NewLabel("배포 준비 중...")
	progressBar := widget.NewProgressBar()
	progressContent := container.NewVBox(progressLabel, progressBar)
	progressDialog := dialog.NewCustom("배포 진행 중", "취소", progressContent, d.window)
	progressDialog.Show()

	// 백그라운드에서 배포 실행
	go func() {
		config, err := d.store.GetConfig()
		if err != nil {
			fyne.Do(func() {
				progressDialog.Hide()
				dialog.ShowError(err, d.window)
			})
			return
		}

		deployer := deploy.NewDeployer(config)
		total := len(checkedFirewalls)
		successCount := 0
		failCount := 0

		for i, fw := range checkedFirewalls {
			// 진행률 업데이트 (UI 스레드에서 실행)
			idx := i
			fyne.Do(func() {
				progressLabel.SetText(fmt.Sprintf("배포 중: %s (%d/%d)", fw.DeviceName, idx+1, total))
				progressBar.SetValue(float64(idx+1) / float64(total))
			})

			// 배포 실행
			result := deployer.Deploy(fw, template)

			// 결과 처리
			if result.Success {
				successCount++
			} else {
				failCount++
			}

			// 장비 상태 저장
			d.store.SaveFirewall(fw)

			// 이력 저장
			if d.historyTab != nil && result.History != nil {
				d.historyTab.AddHistory(result.History)
			}
		}

		// 결과 처리 및 UI 업데이트 (UI 스레드에서 실행)
		fyne.Do(func() {
			progressLabel.SetText("결과 처리 중... 잠시만 기다려주세요.")
			progressBar.SetValue(1.0)

			// 배포 완료 후 체크 상태 초기화
			d.checkedDevices = make(map[int]bool)

			// UI 업데이트
			d.deviceTable.Refresh()

			// 상태 요약 업데이트
			d.updateStatusSummary()

			// 이력 탭 새로고침
			if d.historyTab != nil {
				d.historyTab.loadHistory()
			}

			// 다이얼로그 닫기 및 결과 표시
			progressDialog.Hide()

			resultMsg := fmt.Sprintf("배포 완료\n\n템플릿: %s\n성공: %d개\n실패: %d개", template.Version, successCount, failCount)
			dialog.ShowInformation("배포 결과", resultMsg, d.window)
		})
	}()
}

// 전체 선택/해제 시 호출됩니다.
func (d *DeviceTab) onSelectAll(selected bool) {
	for _, fw := range d.firewalls {
		d.checkedDevices[fw.Index] = selected
	}
	d.deviceTable.Refresh()
}

// 장비 삭제 시 호출됩니다.
func (d *DeviceTab) onDeleteDevices() {
	// 체크된 장비 확인
	checkedCount := 0
	for _, checked := range d.checkedDevices {
		if checked {
			checkedCount++
		}
	}

	if checkedCount == 0 {
		dialog.ShowInformation("알림", "삭제할 장비를 선택해주세요.", d.window)
		return
	}

	dialog.ShowConfirm("확인", fmt.Sprintf("선택한 %d개의 장비를 삭제하시겠습니까?", checkedCount), func(ok bool) {
		if !ok {
			return
		}

		// 체크된 장비 삭제
		for _, fw := range d.firewalls {
			if d.checkedDevices[fw.Index] {
				if fw.Index > 0 {
					// 저장된 장비만 저장소에서 삭제
					d.store.DeleteFirewall(fw.Index)
				}
			}
		}

		// 목록에서도 삭제
		newFirewalls := []*model.Firewall{}
		for _, fw := range d.firewalls {
			if !d.checkedDevices[fw.Index] {
				newFirewalls = append(newFirewalls, fw)
			}
		}
		d.firewalls = newFirewalls
		d.checkedDevices = make(map[int]bool)
		d.selectedDeviceIndex = -1

		d.deviceTable.Refresh()

		// 상태 요약 업데이트
		d.updateStatusSummary()

		dialog.ShowInformation("성공", "선택한 장비가 삭제되었습니다.", d.window)
	}, d.window)
}

// 장비 저장 시 호출됩니다.
func (d *DeviceTab) onSaveDevices() {
	// 모든 장비 저장
	for _, fw := range d.firewalls {
		if fw.DeviceName == "" {
			continue // IP가 없는 장비는 저장하지 않음
		}

		if err := d.store.SaveFirewall(fw); err != nil {
			dialog.ShowError(err, d.window)
			return
		}
	}

	dialog.ShowInformation("성공", "장비 정보가 저장되었습니다.", d.window)
	d.loadFirewalls()
}

// 장비 목록을 새로고침합니다.
func (d *DeviceTab) RefreshDevices() {
	d.onRefreshAll()
}

// 장비 목록만 새로고침합니다. (서버 상태 체크 없이)
func (d *DeviceTab) ReloadDevices() {
	d.loadFirewalls()
}

// 템플릿 목록만 새로고침합니다.
func (d *DeviceTab) RefreshTemplates() {
	d.refreshTemplateList()
}

// 이력 탭 참조를 설정합니다.
func (d *DeviceTab) SetHistoryTab(historyTab *HistoryTab) {
	d.historyTab = historyTab
}

// 새로고침 버튼 참조를 설정합니다.
func (d *DeviceTab) SetRefreshButton(btn *widget.Button) {
	d.refreshBtn = btn
}

// 해당 IP의 장비 배포 상태를 초기화합니다.
func (d *DeviceTab) ResetDeviceDeployStatus(deviceIP string) {
	for _, fw := range d.firewalls {
		if fw.DeviceName == deviceIP {
			fw.DeployStatus = model.DeployStatusUnknown
			fw.Version = "-"
			d.store.SaveFirewall(fw)
			break
		}
	}
	d.deviceTable.Refresh()
}

// 선택한 장비의 서버 상태를 새로고침합니다.
func (d *DeviceTab) onRefreshAll() {
	// 이미 새로고침 중이면 무시
	if d.isRefreshing {
		return
	}

	// 템플릿 목록 새로고침
	d.refreshTemplateList()

	// 선택된 장비 목록 수집
	selectedFirewalls := make([]*model.Firewall, 0)
	for _, fw := range d.firewalls {
		if d.checkedDevices[fw.Index] {
			selectedFirewalls = append(selectedFirewalls, fw)
		}
	}

	// 선택된 장비가 없으면 종료
	if len(selectedFirewalls) == 0 {
		dialog.ShowInformation("알림", "상태를 확인할 장비를 선택해주세요.", d.window)
		return
	}

	// 새로고침 시작
	d.isRefreshing = true
	if d.refreshBtn != nil {
		d.refreshBtn.Disable()
	}

	// 진행 중 다이얼로그 표시
	progressLabel := widget.NewLabel(fmt.Sprintf("장비 상태 확인 중... (총 %d개)", len(selectedFirewalls)))
	progressBar := widget.NewProgressBarInfinite()
	progressContent := container.NewVBox(progressLabel, progressBar)
	progressDialog := dialog.NewCustomWithoutButtons("새로고침 중", progressContent, d.window)
	progressDialog.Show()

	// 백그라운드에서 선택된 장비 상태 확인 실행
	go func() {
		config, err := d.store.GetConfig()
		if err != nil {
			fyne.Do(func() {
				progressDialog.Hide()
				d.isRefreshing = false
				if d.refreshBtn != nil {
					d.refreshBtn.Enable()
				}
			})
			return
		}

		deployer := deploy.NewDeployer(config)

		// Agent 모드: 배치 호출 (한번에 선택된 장비 상태 확인)
		// Direct 모드: 개별 호출 (장비별로 순차 확인)
		deployer.HealthCheckBatch(selectedFirewalls)

		// 장비 상태 저장
		for _, fw := range selectedFirewalls {
			d.store.SaveFirewall(fw)
		}

		// UI 업데이트 (메인 스레드에서 실행)
		fyne.Do(func() {
			// 테이블 새로고침
			d.deviceTable.Refresh()

			// 상태 요약 업데이트
			d.updateStatusSummary()

			// 진행 다이얼로그 닫기
			progressDialog.Hide()

			// 새로고침 완료
			d.isRefreshing = false
			if d.refreshBtn != nil {
				d.refreshBtn.Enable()
			}

			// 2초 후 자동으로 사라지는 완료 다이얼로그 표시
			infoDialog := dialog.NewInformation("완료", fmt.Sprintf("%d개 장비 상태 확인 완료", len(selectedFirewalls)), d.window)
			infoDialog.Show()
			go func() {
				time.Sleep(2 * time.Second)
				fyne.Do(func() {
					infoDialog.Hide()
				})
			}()
		})
	}()
}

// IP 주소 또는 IP:PORT 형식이 유효한지 검사합니다.
func isValidIPOrHostPort(address string) bool {
	// IP:PORT 형식인 경우
	if host, _, err := net.SplitHostPort(address); err == nil {
		// 호스트 부분이 유효한 IP인지 확인
		return net.ParseIP(host) != nil
	}
	// 순수 IP 주소인 경우
	return net.ParseIP(address) != nil
}
