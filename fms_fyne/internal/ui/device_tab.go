package ui

import (
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"fms/internal/deploy"
	"fms/internal/model"
	"fms/internal/storage"
	"fms/internal/themes"
	"fms/internal/ui/component"
	"fms/internal/usecase"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	fynestorage "fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// 장비 관리 탭을 구현합니다. (PRD 3.3.3 기준)
type DeviceTab struct {
	window      fyne.Window
	store       *storage.JSONStore
	templateTab *TemplateTab
	historyTab  *HistoryTab
	programTab  *ProgramTab
	content     fyne.CanvasObject

	// UI 컴포넌트
	searchBox   *component.SearchBox  // 검색 컴포넌트 (공통)
	deviceTable *component.PagedTable // 장비 테이블 (공통 컴포넌트)

	// 상태 요약 표시
	statusGreenLabel  *widget.Label
	statusYellowLabel *widget.Label
	statusRedLabel    *widget.Label

	// 데이터
	firewalls           []*model.Firewall
	filteredFirewalls   []*model.Firewall // 검색 필터링된 목록
	selectedDeviceIndex int

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
		filteredFirewalls:   []*model.Firewall{},
		selectedDeviceIndex: -1,
	}
	tab.createUI()
	tab.loadFirewalls()
	return tab
}

// 장비 탭의 UI를 생성합니다. (PRD 3.3.3 레이아웃)
func (d *DeviceTab) createUI() {
	// 상단: 검색 + 상태요약 + 버튼
	topPanel := d.createTopPanel()

	// 중앙: 장비 테이블
	tablePanel := d.createDeviceTablePanel()

	// 전체 레이아웃
	d.content = container.NewBorder(
		topPanel, // 상단 고정
		nil,      // 하단 없음
		nil, nil,
		tablePanel, // 중앙 (자동 확장)
	)
}

// 상단 패널을 생성합니다. (PRD 3.3.3 기준)
func (d *DeviceTab) createTopPanel() fyne.CanvasObject {
	// 검색 컴포넌트 (공통)
	d.searchBox = component.NewSearchBox(component.SearchBoxConfig{
		Placeholder: "장비명/IP 검색...",
		Width:       200,
		OnSearch: func(text string) {
			d.applyFilter()
		},
	})

	// 상태 요약 레이블
	d.statusGreenLabel = widget.NewLabel("0")
	d.statusYellowLabel = widget.NewLabel("0")
	d.statusRedLabel = widget.NewLabel("0")

	// 색상이 있는 상태 표시
	greenDot := canvas.NewText("●", themes.Colors["green"])
	greenDot.TextSize = 21
	yellowDot := canvas.NewText("●", themes.Colors["yellow"])
	yellowDot.TextSize = 21
	redDot := canvas.NewText("●", themes.Colors["red"])
	redDot.TextSize = 21

	// 새로고침 버튼 (🔄)
	d.refreshBtn = widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		d.onRefreshAll()
	})

	// 상태 요약 컨테이너 (새로고침 버튼 포함)
	statusSummary := container.NewHBox(
		greenDot, widget.NewLabel("연결:"), d.statusGreenLabel,
		widget.NewLabel(" "),
		yellowDot, widget.NewLabel("알수없음:"), d.statusYellowLabel,
		widget.NewLabel(" "),
		redDot, widget.NewLabel("연결안됨:"), d.statusRedLabel,
		widget.NewLabel("  "), // 새로고침 버튼 앞 간격
		d.refreshBtn,
	)

	// 삭제 버튼 (투명 배경, 빨간 텍스트)
	deleteBtn := component.NewCustomButton("삭제", nil, themes.Colors["red"], nil, func() {
		d.onDeleteDevices()
	})

	// 배포 버튼
	deployBtn := component.NewCustomButton("배포", nil, nil, themes.Colors["darkgray"], func() {
		d.onDeploy()
	})

	// 추가/수정 버튼
	addEditBtn := component.NewCustomButton("추가/수정", nil, nil, themes.Colors["blue"], func() {
		d.showAddEditDialog()
	})

	// 버튼 영역 (버튼 간격 추가) - 순서: 삭제, 배포, 추가/수정
	buttonArea := container.NewHBox(
		deleteBtn,
		widget.NewLabel("  "), // 버튼 간격
		deployBtn,
		widget.NewLabel("  "), // 버튼 간격
		addEditBtn,
	)

	// 헤더 라인 (위아래 패딩 10px, 오른쪽 마진 10px) - 상태 요약 중앙 배치
	headerLine := container.NewBorder(nil, nil, d.searchBox.Content(), buttonArea, container.NewCenter(statusSummary))
	paddedHeader := container.New(layout.NewCustomPaddedLayout(10, 10, 0, 10), headerLine)

	return paddedHeader
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

// 장비 테이블 패널을 생성합니다. (PRD 3.3.3 기준)
func (d *DeviceTab) createDeviceTablePanel() fyne.CanvasObject {
	// PRD 테이블 컬럼: 선택, 장비명, 서버 IP, 서버상태, 보고시간, 접속방식
	d.deviceTable = component.NewPagedTable(component.PagedTableConfig{
		Columns: []component.ColumnDef{
			{Header: "선택", Width: 50},
			{Header: "장비명", Width: 240},
			{Header: "서버 IP", Width: 300},
			{Header: "서버상태", Width: 80},
			{Header: "보고시간", Width: 225},
			{Header: "접속방식", Width: 80},
		},
		PageSize: 15,
		OnCellUpdate: func(row int, col int, cell fyne.CanvasObject) {
			d.updateDeviceCell(row, col, cell)
		},
		OnRowSelected: func(row int) {
			// 단일 클릭 - 선택
			if row >= 0 && row < len(d.filteredFirewalls) {
				d.selectedDeviceIndex = d.getOriginalIndex(d.filteredFirewalls[row])
			}
		},
		OnRowDoubleClick: func(row int) {
			// 더블 클릭 - 상세보기
			if row >= 0 && row < len(d.filteredFirewalls) {
				d.showDetailDialog(d.filteredFirewalls[row])
			}
		},
	})

	return d.deviceTable.Content()
}

// 장비 테이블 셀을 업데이트합니다.
func (d *DeviceTab) updateDeviceCell(row int, col int, cell fyne.CanvasObject) {
	label := cell.(*widget.Label)

	if row >= len(d.filteredFirewalls) {
		label.SetText("")
		return
	}

	fw := d.filteredFirewalls[row]

	switch col {
	case 1: // 장비명
		if fw.DeviceName != "" && fw.DeviceName != fw.DeviceIP {
			label.SetText(fw.DeviceName)
		} else {
			label.SetText("-")
		}
	case 2: // 서버 IP
		if fw.DeviceIP != "" {
			label.SetText(fw.DeviceIP)
		} else {
			label.SetText(fw.DeviceName)
		}
	case 3: // 서버상태
		switch fw.ServerStatus {
		case model.ServerStatusRunning:
			label.SetText("정상")
		case model.ServerStatusStop:
			label.SetText("연결안됨")
		default:
			label.SetText("-")
		}
	case 4: // 보고시간
		if fw.LastCheckedAt != "" {
			label.SetText(fw.LastCheckedAt)
		} else {
			label.SetText("-")
		}
	case 5: // 접속방식
		if fw.DevicePPK != "" {
			label.SetText("PPK")
		} else if fw.DevicePW != "" {
			label.SetText("PW")
		} else {
			label.SetText("-")
		}
	}
}

// 필터링된 목록에서 원본 인덱스를 찾습니다.
func (d *DeviceTab) getOriginalIndex(fw *model.Firewall) int {
	for i, f := range d.firewalls {
		if f.Index == fw.Index {
			return i
		}
	}
	return -1
}

// 검색 필터를 적용합니다.
func (d *DeviceTab) applyFilter() {
	searchText := strings.ToLower(strings.TrimSpace(d.searchBox.GetText()))

	if searchText == "" {
		d.filteredFirewalls = d.firewalls
	} else {
		filtered := []*model.Firewall{}
		for _, fw := range d.firewalls {
			// 장비명 또는 IP로 검색
			if strings.Contains(strings.ToLower(fw.DeviceName), searchText) ||
				strings.Contains(strings.ToLower(fw.DeviceIP), searchText) {
				filtered = append(filtered, fw)
			}
		}
		d.filteredFirewalls = filtered
	}

	d.deviceTable.SetData(len(d.filteredFirewalls))
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

	// Index 내림차순 정렬 (최신순 - Index가 클수록 최신)
	sort.Slice(d.firewalls, func(i, j int) bool {
		return d.firewalls[i].Index > d.firewalls[j].Index
	})

	// 필터 적용
	d.applyFilter()

	// 상태 요약 업데이트
	d.updateStatusSummary()
}

// 장비 추가/수정 다이얼로그를 표시합니다. (PRD 3.3.3 기준)
func (d *DeviceTab) showAddEditDialog() {
	var editingFw *model.Firewall

	// 체크된 장비가 있으면 수정 모드 (PagedTable 사용)
	checkedRows := d.deviceTable.GetCheckedRows()
	if len(checkedRows) > 0 && checkedRows[0] < len(d.filteredFirewalls) {
		editingFw = d.filteredFirewalls[checkedRows[0]]
	}

	// 입력 필드
	deviceNameEntry := widget.NewEntry()
	deviceNameEntry.SetPlaceHolder("장비 이름")

	serverIPEntry := widget.NewEntry()
	serverIPEntry.SetPlaceHolder("192.168.1.1 또는 192.168.1.1:8080")

	sshIDEntry := widget.NewEntry()
	sshIDEntry.SetPlaceHolder("root")

	sshPWEntry := widget.NewPasswordEntry()
	sshPWEntry.SetPlaceHolder("비밀번호")

	ppkPathLabel := widget.NewLabel("PPK 파일을 선택해주세요")
	ppkPathLabel.Truncation = fyne.TextTruncateEllipsis
	var ppkPath string

	// 접속선택 드롭다운
	authSelect := widget.NewSelect([]string{"PW", "PPK"}, nil)
	authSelect.SetSelected("PW")

	// PW 입력 영역
	pwContainer := container.NewGridWithColumns(2,
		widget.NewLabel("SSH ID:"), sshIDEntry,
		widget.NewLabel("비밀번호:"), sshPWEntry,
	)

	// PPK 입력 영역 (초기에 숨김)
	ppkBrowseBtn := widget.NewButton("찾아보기...", nil)
	ppkContainer := container.NewGridWithColumns(2,
		widget.NewLabel("PPK:"), container.NewBorder(nil, nil, nil, ppkBrowseBtn, ppkPathLabel),
	)
	ppkContainer.Hide()

	// 접속선택에 따른 동적 필드 전환
	authSelect.OnChanged = func(selected string) {
		if selected == "PW" {
			pwContainer.Show()
			ppkContainer.Hide()
		} else {
			pwContainer.Hide()
			ppkContainer.Show()
		}
	}

	// 수정 모드면 기존 값 채우기
	if editingFw != nil {
		deviceNameEntry.SetText(editingFw.DeviceName)
		if editingFw.DeviceIP != "" {
			serverIPEntry.SetText(editingFw.DeviceIP)
		} else {
			serverIPEntry.SetText(editingFw.DeviceName) // 기존 데이터 호환
		}
		sshIDEntry.SetText(editingFw.DeviceID)
		sshPWEntry.SetText(editingFw.DevicePW)
		if editingFw.DevicePPK != "" {
			ppkPath = editingFw.DevicePPK
			ppkPathLabel.SetText(editingFw.DevicePPK)
			authSelect.SetSelected("PPK")
		}
	}

	// 다이얼로그 컨텐츠
	content := container.NewVBox(
		container.NewGridWithColumns(2,
			widget.NewLabel("장비명:"), deviceNameEntry,
			widget.NewLabel("서버 IP:"), serverIPEntry,
		),
		widget.NewSeparator(),
		container.NewGridWithColumns(2,
			widget.NewLabel("접속선택:"), authSelect,
		),
		pwContainer,
		ppkContainer,
	)

	// 다이얼로그 생성
	var parentDialog dialog.Dialog
	parentDialog = dialog.NewCustomConfirm("장비 추가/수정", "저장", "취소", content, func(ok bool) {
		if !ok {
			return
		}

		// 유효성 검사
		if serverIPEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("서버 IP를 입력해주세요"), d.window)
			return
		}
		if !isValidIPOrHostPort(serverIPEntry.Text) {
			dialog.ShowError(fmt.Errorf("올바른 IP 주소 형식이 아닙니다"), d.window)
			return
		}

		// 장비 저장
		var fw *model.Firewall
		if editingFw != nil {
			fw = editingFw
		} else {
			fw = model.NewFirewall("")
			d.firewalls = append(d.firewalls, fw)
		}

		fw.DeviceName = deviceNameEntry.Text
		fw.DeviceIP = serverIPEntry.Text

		// DeviceName이 비어있으면 IP 사용
		if fw.DeviceName == "" {
			fw.DeviceName = fw.DeviceIP
		}

		if authSelect.Selected == "PW" {
			fw.DeviceID = sshIDEntry.Text
			fw.DevicePW = sshPWEntry.Text
			fw.DevicePPK = ""
		} else {
			fw.DeviceID = sshIDEntry.Text
			fw.DevicePW = ""
			fw.DevicePPK = ppkPath
		}

		if err := d.store.SaveFirewall(fw); err != nil {
			dialog.ShowError(err, d.window)
			return
		}

		// 체크 해제 및 새로고침
		d.deviceTable.ClearChecked()
		d.loadFirewalls()
		dialog.ShowInformation("성공", "장비 정보가 저장되었습니다.", d.window)
	}, d.window)

	// PPK 찾아보기 버튼 - 다이얼로그 중첩 처리 (fyne-docs 스킬 참고)
	ppkBrowseBtn.OnTapped = func() {
		parentDialog.Hide() // 부모 다이얼로그 숨김

		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if reader != nil {
				ppkPath = reader.URI().Path()
				ppkPathLabel.SetText(ppkPath)
				reader.Close()
			}
			parentDialog.Show() // 부모 다이얼로그 다시 표시
		}, d.window)
		fileDialog.SetFilter(fynestorage.NewExtensionFileFilter([]string{".ppk", ".pem", ".key"}))
		fileDialog.Show()
	}

	parentDialog.Show()
}

// 상세보기 다이얼로그를 표시합니다. (PRD 3.3.3 기준)
func (d *DeviceTab) showDetailDialog(fw *model.Firewall) {
	// 접속정보 텍스트
	authInfo := "-"
	authDetail := ""
	if fw.DevicePPK != "" {
		authInfo = "PPK"
		authDetail = fmt.Sprintf("PPK 경로: %s", fw.DevicePPK)
	} else if fw.DevicePW != "" {
		authInfo = "PW"
		authDetail = fmt.Sprintf("ID: %s", fw.DeviceID)
	}

	// 서버상태 텍스트
	statusText := "알수없음"
	switch fw.ServerStatus {
	case model.ServerStatusRunning:
		statusText = "정상"
	case model.ServerStatusStop:
		statusText = "연결안됨"
	}

	// 배포정보
	deployInfo := fmt.Sprintf("방화벽 룰셋 버전: %s", fw.Version)
	if fw.Version == "" || fw.Version == "-" {
		deployInfo = "방화벽 룰셋 버전: -"
	}

	// 프로그램 버전 정보
	if fw.ProgramVersions != nil && len(fw.ProgramVersions) > 0 {
		for name, ver := range fw.ProgramVersions {
			deployInfo += fmt.Sprintf("\n%s 버전: %s", name, ver)
		}
	}

	content := container.NewVBox(
		container.NewGridWithColumns(2,
			widget.NewLabelWithStyle("장비명:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel(fw.DeviceName),
		),
		container.NewGridWithColumns(2,
			widget.NewLabelWithStyle("IP:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel(func() string {
				if fw.DeviceIP != "" {
					return fw.DeviceIP
				}
				return fw.DeviceName
			}()),
		),
		container.NewGridWithColumns(2,
			widget.NewLabelWithStyle("연결상태:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel(statusText),
		),
		widget.NewSeparator(),
		container.NewGridWithColumns(2,
			widget.NewLabelWithStyle("접속정보:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel(authInfo),
		),
		widget.NewLabel(authDetail),
		widget.NewSeparator(),
		widget.NewLabelWithStyle("배포정보:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(deployInfo),
	)

	dialog.ShowCustom("상세보기", "확인", content, d.window)
}

// 배포 다이얼로그를 표시합니다. (PRD 3.3.3 기준 - 통합 다이얼로그)
func (d *DeviceTab) onDeploy() {
	// 체크된 장비 수집 (PagedTable 사용)
	checkedRows := d.deviceTable.GetCheckedRows()
	checkedFirewalls := []*model.Firewall{}
	checkedIPs := []string{}
	for _, row := range checkedRows {
		if row < len(d.filteredFirewalls) {
			fw := d.filteredFirewalls[row]
			checkedFirewalls = append(checkedFirewalls, fw)
			if fw.DeviceIP != "" {
				checkedIPs = append(checkedIPs, fw.DeviceIP)
			} else {
				checkedIPs = append(checkedIPs, fw.DeviceName)
			}
		}
	}

	if len(checkedFirewalls) == 0 {
		dialog.ShowError(fmt.Errorf("배포할 장비를 선택해주세요"), d.window)
		return
	}

	// 선택한 IP 리스트
	ipListLabel := widget.NewLabel(strings.Join(checkedIPs, ", "))
	ipListLabel.Wrapping = fyne.TextWrapWord

	// 배포선택 드롭다운
	deployTypeSelect := widget.NewSelect([]string{"방화벽 룰 배포", "프로그램 배포"}, nil)
	deployTypeSelect.SetSelected("방화벽 룰 배포")

	// 배포 리스트 (라디오 버튼 그룹)
	var deployListContainer *fyne.Container
	var selectedItem string

	// 방화벽 룰 목록
	templates := d.templateTab.GetTemplateVersions()
	// 프로그램 목록
	programs, _ := d.store.GetAllPrograms()
	programItems := make([]string, len(programs))
	for i, p := range programs {
		programItems[i] = fmt.Sprintf("%s %s", p.ProcessName, p.ProcessVersion)
	}

	// 라디오 그룹 생성 함수
	createRadioList := func(items []string) *widget.RadioGroup {
		radio := widget.NewRadioGroup(items, func(selected string) {
			selectedItem = selected
		})
		return radio
	}

	// 초기 방화벽 룰 리스트
	radioGroup := createRadioList(templates)
	deployListScroll := container.NewScroll(radioGroup)
	deployListScroll.SetMinSize(fyne.NewSize(300, 150))
	deployListContainer = container.NewBorder(nil, nil, nil, nil, deployListScroll)

	// 배포선택 변경 시 리스트 갱신
	deployTypeSelect.OnChanged = func(selected string) {
		selectedItem = ""
		var newRadio *widget.RadioGroup
		if selected == "방화벽 룰 배포" {
			newRadio = createRadioList(templates)
		} else {
			newRadio = createRadioList(programItems)
		}
		deployListScroll.Content = newRadio
		deployListScroll.Refresh()
	}

	// 다이얼로그 컨텐츠
	content := container.NewVBox(
		widget.NewLabelWithStyle("선택한 IP 리스트:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		ipListLabel,
		widget.NewSeparator(),
		container.NewGridWithColumns(2,
			widget.NewLabel("배포선택:"), deployTypeSelect,
		),
		widget.NewSeparator(),
		widget.NewLabelWithStyle("배포 리스트:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		deployListContainer,
	)

	dialog.ShowCustomConfirm("배포", "배포", "취소", content, func(ok bool) {
		if !ok {
			return
		}

		if selectedItem == "" {
			dialog.ShowError(fmt.Errorf("배포할 항목을 선택해주세요"), d.window)
			return
		}

		if deployTypeSelect.Selected == "방화벽 룰 배포" {
			d.executeFirewallDeploy(checkedFirewalls, selectedItem)
		} else {
			// 프로그램 배포
			for _, p := range programs {
				displayName := fmt.Sprintf("%s %s", p.ProcessName, p.ProcessVersion)
				if displayName == selectedItem {
					d.executeProgramUpdate(checkedFirewalls, p)
					break
				}
			}
		}
	}, d.window)
}

// 방화벽 룰 배포를 실행합니다.
func (d *DeviceTab) executeFirewallDeploy(firewalls []*model.Firewall, templateVersion string) {
	template := d.templateTab.GetTemplate(templateVersion)
	if template == nil {
		dialog.ShowError(fmt.Errorf("템플릿을 찾을 수 없습니다: %s", templateVersion), d.window)
		return
	}

	if !template.IsValid() {
		dialog.ShowError(fmt.Errorf("선택한 템플릿에 내용이 없습니다"), d.window)
		return
	}

	// 진행률 다이얼로그
	progressLabel := widget.NewLabel("배포 준비 중...")
	progressBar := widget.NewProgressBar()
	progressContent := container.NewVBox(progressLabel, progressBar)
	progressDialog := dialog.NewCustomWithoutButtons("배포 진행 중", progressContent, d.window)
	progressDialog.Show()

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
		total := len(firewalls)
		successCount := 0
		failCount := 0

		for i, fw := range firewalls {
			idx := i
			fyne.Do(func() {
				ip := fw.DeviceIP
				if ip == "" {
					ip = fw.DeviceName
				}
				progressLabel.SetText(fmt.Sprintf("배포 중: %s (%d/%d)", ip, idx+1, total))
				progressBar.SetValue(float64(idx+1) / float64(total))
			})

			result := deployer.Deploy(fw, template)

			if result.Success {
				successCount++
			} else {
				failCount++
			}

			d.store.SaveFirewall(fw)

			if d.historyTab != nil && result.History != nil {
				d.historyTab.AddHistory(result.History)
			}
		}

		fyne.Do(func() {
			progressDialog.Hide()
			d.deviceTable.ClearChecked()
			d.loadFirewalls()

			if d.historyTab != nil {
				d.historyTab.loadHistory()
			}

			resultMsg := fmt.Sprintf("배포 완료\n\n템플릿: %s\n성공: %d개\n실패: %d개", template.Version, successCount, failCount)
			dialog.ShowInformation("배포 결과", resultMsg, d.window)
		})
	}()
}

// 프로그램 업데이트를 실행합니다.
func (d *DeviceTab) executeProgramUpdate(devices []*model.Firewall, program *model.ProcessInfo) {
	// SSH 인증 정보 확인
	for _, fw := range devices {
		if fw.DeviceID == "" {
			ip := fw.DeviceIP
			if ip == "" {
				ip = fw.DeviceName
			}
			dialog.ShowError(fmt.Errorf("장비 %s의 SSH ID가 설정되지 않았습니다", ip), d.window)
			return
		}
		if fw.DevicePW == "" && fw.DevicePPK == "" {
			ip := fw.DeviceIP
			if ip == "" {
				ip = fw.DeviceName
			}
			dialog.ShowError(fmt.Errorf("장비 %s의 SSH 인증 정보가 설정되지 않았습니다", ip), d.window)
			return
		}
	}

	// 진행률 다이얼로그
	progressLabel := widget.NewLabel("업데이트 준비 중...")
	progressBar := widget.NewProgressBar()
	progressContent := container.NewVBox(progressLabel, progressBar)
	progressDialog := dialog.NewCustomWithoutButtons("프로그램 업데이트 중", progressContent, d.window)
	progressDialog.Show()

	go func() {
		updateUC := usecase.NewUpdateUseCase()
		total := len(devices)
		successCount := 0
		failCount := 0

		for i, device := range devices {
			idx := i
			fyne.Do(func() {
				ip := device.DeviceIP
				if ip == "" {
					ip = device.DeviceName
				}
				progressLabel.SetText(fmt.Sprintf("업데이트 중: %s (%d/%d)", ip, idx+1, total))
				progressBar.SetValue(float64(idx+1) / float64(total))
			})

			result := updateUC.UpdateProgram(device, program)

			if result.Success {
				successCount++
				d.store.SaveFirewall(device)
			} else {
				failCount++
			}

			if d.historyTab != nil && result.History != nil {
				d.historyTab.AddHistory(result.History)
			}
		}

		fyne.Do(func() {
			progressDialog.Hide()
			d.deviceTable.ClearChecked()
			d.loadFirewalls()

			if d.historyTab != nil {
				d.historyTab.loadHistory()
			}

			resultMsg := fmt.Sprintf("프로그램 업데이트 완료\n\n프로그램: %s %s\n성공: %d개\n실패: %d개",
				program.ProcessName, program.ProcessVersion, successCount, failCount)
			dialog.ShowInformation("업데이트 결과", resultMsg, d.window)
		})
	}()
}

// 장비 삭제 시 호출됩니다.
func (d *DeviceTab) onDeleteDevices() {
	checkedRows := d.deviceTable.GetCheckedRows()
	checkedCount := len(checkedRows)

	if checkedCount == 0 {
		return // 선택 없으면 아무 동작 안함 (PRD 기준)
	}

	dialog.ShowConfirm("확인", fmt.Sprintf("선택한 %d개 장비를 삭제하시겠습니까?", checkedCount), func(ok bool) {
		if !ok {
			return
		}

		// 체크된 장비 삭제
		for _, row := range checkedRows {
			if row < len(d.filteredFirewalls) {
				fw := d.filteredFirewalls[row]
				if fw.Index > 0 {
					d.store.DeleteFirewall(fw.Index)
				}
			}
		}

		// 삭제 후 새로고침
		d.deviceTable.ClearChecked()
		d.selectedDeviceIndex = -1
		d.loadFirewalls()

		dialog.ShowInformation("성공", "선택한 장비가 삭제되었습니다.", d.window)
	}, d.window)
}

// 선택한 장비의 서버 상태를 새로고침합니다.
func (d *DeviceTab) onRefreshAll() {
	if d.isRefreshing {
		return
	}

	// 체크된 장비 수집 (PagedTable 사용)
	checkedRows := d.deviceTable.GetCheckedRows()
	selectedFirewalls := make([]*model.Firewall, 0)
	for _, row := range checkedRows {
		if row < len(d.filteredFirewalls) {
			selectedFirewalls = append(selectedFirewalls, d.filteredFirewalls[row])
		}
	}

	if len(selectedFirewalls) == 0 {
		dialog.ShowInformation("알림", "상태를 확인할 장비를 선택해주세요.", d.window)
		return
	}

	d.isRefreshing = true
	if d.refreshBtn != nil {
		d.refreshBtn.Disable()
	}

	progressLabel := widget.NewLabel(fmt.Sprintf("장비 상태 확인 중... (총 %d개)", len(selectedFirewalls)))
	progressBar := widget.NewProgressBarInfinite()
	progressContent := container.NewVBox(progressLabel, progressBar)
	progressDialog := dialog.NewCustomWithoutButtons("새로고침 중", progressContent, d.window)
	progressDialog.Show()

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
		deployer.HealthCheckBatch(selectedFirewalls)

		// 상태 확인 시간 업데이트
		now := time.Now().Format("2006-01-02 15:04:05")
		for _, fw := range selectedFirewalls {
			fw.LastCheckedAt = now
			d.store.SaveFirewall(fw)
		}

		fyne.Do(func() {
			d.applyFilter()
			d.updateStatusSummary()
			progressDialog.Hide()

			d.isRefreshing = false
			if d.refreshBtn != nil {
				d.refreshBtn.Enable()
			}

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

// 이력 탭 참조를 설정합니다.
func (d *DeviceTab) SetHistoryTab(historyTab *HistoryTab) {
	d.historyTab = historyTab
}

// 프로그램 탭 참조를 설정합니다.
func (d *DeviceTab) SetProgramTab(programTab *ProgramTab) {
	d.programTab = programTab
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
	// 템플릿 탭 새로고침
	d.templateTab.RefreshTemplates()
}

// 해당 IP의 장비 배포 상태를 초기화합니다.
func (d *DeviceTab) ResetDeviceDeployStatus(deviceIP string) {
	for _, fw := range d.firewalls {
		if fw.DeviceName == deviceIP || fw.DeviceIP == deviceIP {
			fw.DeployStatus = model.DeployStatusUnknown
			fw.Version = "-"
			d.store.SaveFirewall(fw)
			break
		}
	}
	d.deviceTable.Refresh()
}

// IP 주소 또는 IP:PORT 형식이 유효한지 검사합니다.
func isValidIPOrHostPort(address string) bool {
	if host, _, err := net.SplitHostPort(address); err == nil {
		return net.ParseIP(host) != nil
	}
	return net.ParseIP(address) != nil
}
