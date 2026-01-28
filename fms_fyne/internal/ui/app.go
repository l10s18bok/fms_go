// Package ui는 FMS 애플리케이션의 사용자 인터페이스를 구현합니다.
package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"fms/internal/model"
	"fms/internal/repository"
	"fms/internal/storage"
	"fms/internal/ui/component"
	"fms/internal/version"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	fynestorage "fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// 메인 애플리케이션 UI를 관리합니다.
type MainUI struct {
	window      fyne.Window
	store       *storage.JSONStore
	fileStore   *storage.FileStore           // 파일 저장소 (data 디렉토리)
	historyRepo repository.HistoryRepository // 배포 이력 저장소 (SQLite)
	tabs        *container.DocTabs           // 닫기 버튼이 있는 탭
	leftMenu    *fyne.Container              // 왼쪽 메뉴

	// 탭
	firewallTab *FirewallTab // 방화벽 관리 탭
	deviceTab   *DeviceTab
	historyTab  *HistoryTab
	programTab  *ProgramTab // 패키지 관리 탭

	// 탭 아이템 참조 (동적 추가/제거용)
	firewallTabItem *container.TabItem // 방화벽 관리 탭 아이템
	deviceTabItem   *container.TabItem
	historyTabItem  *container.TabItem
	programTabItem  *container.TabItem // 패키지 관리 탭 아이템

	// 동적 편집 탭 관리
	editTabs map[string]*container.TabItem // 파일명 -> 탭 아이템 매핑

	// 상태바 라벨
	statusLabel *widget.Label // 현재 탭 정보 표시
}

// 새로운 메인 UI 인스턴스를 생성합니다.
func NewMainUI(window fyne.Window, store *storage.JSONStore, fileStore *storage.FileStore, historyRepo repository.HistoryRepository) *MainUI {
	ui := &MainUI{
		window:      window,
		store:       store,
		fileStore:   fileStore,
		historyRepo: historyRepo,
		editTabs:    make(map[string]*container.TabItem),
	}

	// 각 탭 생성
	ui.firewallTab = NewFirewallTab(window, fileStore, ui)
	ui.deviceTab = NewDeviceTab(window, store, ui.firewallTab)
	ui.historyTab = NewHistoryTab(window, historyRepo)
	ui.programTab = NewProgramTab(window, store)

	// 탭 간 참조 설정
	ui.deviceTab.SetHistoryTab(ui.historyTab)
	ui.historyTab.SetDeviceTab(ui.deviceTab)

	// 탭 아이템 생성 (동적 추가/제거용)
	ui.firewallTabItem = container.NewTabItemWithIcon("방화벽 관리", theme.DocumentIcon(), ui.firewallTab.Content())
	ui.deviceTabItem = container.NewTabItemWithIcon("장비 관리", theme.ComputerIcon(), ui.deviceTab.Content())
	ui.historyTabItem = container.NewTabItemWithIcon("배포 이력", theme.HistoryIcon(), ui.historyTab.Content())
	ui.programTabItem = container.NewTabItemWithIcon("패키지 관리", theme.FolderOpenIcon(), ui.programTab.Content())

	// DocTabs 컨테이너 생성 (닫기 버튼 지원)
	ui.tabs = container.NewDocTabs()
	ui.tabs.SetTabLocation(container.TabLocationTop)

	// 탭 닫기 시 처리
	ui.tabs.OnClosed = func(tab *container.TabItem) {
		// 편집 탭이 닫히면 editTabs에서 제거
		for fileName, tabItem := range ui.editTabs {
			if tabItem == tab {
				delete(ui.editTabs, fileName)
				break
			}
		}
	}

	// 탭 변경 시 이벤트 처리
	ui.tabs.OnSelected = func(tab *container.TabItem) {
		if tab == ui.deviceTabItem {
			ui.deviceTab.RefreshFiles()
		}
		ui.updateStatusBar(tab)
	}

	// 왼쪽 메뉴 생성
	ui.createLeftMenu()

	// 네이티브 메뉴바 설정
	ui.setupMainMenu()

	// 파일 드롭 핸들러 설정
	ui.setupFileDrop()

	// 초기 탭 열기: 방화벽 관리
	ui.openTab(ui.firewallTabItem)

	// 저장된 설정에서 자동 상태 체크 활성화
	ui.initAutoStatusCheck()

	return ui
}

// 왼쪽 메뉴를 생성합니다.
func (m *MainUI) createLeftMenu() {
	// 방화벽 관리 버튼 (텍스트만, 위아래 간격 4, 테마 반응형)
	ruleBtn := component.NewCustomButton("방화벽 관리", nil, nil, nil, func() {
		m.openTab(m.firewallTabItem)
	}, 4, 4, 0, 0)

	// 패키지 관리 버튼 (텍스트만, 위아래 간격 4, 테마 반응형)
	programBtn := component.NewCustomButton("패키지 관리", nil, nil, nil, func() {
		m.openTab(m.programTabItem)
	}, 4, 4, 0, 0)

	// 장비 관리 버튼 (텍스트만, 위아래 간격 4, 테마 반응형)
	deviceBtn := component.NewCustomButton("장비 관리", nil, nil, nil, func() {
		m.openTab(m.deviceTabItem)
	}, 4, 4, 0, 0)

	// 배포 이력 버튼 (텍스트만, 위아래 간격 4, 테마 반응형)
	historyBtn := component.NewCustomButton("배포 이력", nil, nil, nil, func() {
		m.openTab(m.historyTabItem)
	}, 4, 4, 0, 0)

	// 왼쪽 메뉴 레이아웃: 방화벽 관리 → 패키지 관리 → 장비 관리 → 배포 이력
	m.leftMenu = container.NewVBox(
		ruleBtn,
		programBtn,
		deviceBtn,
		historyBtn,
	)
}

// 탭을 열거나 이미 열려있으면 선택합니다.
func (m *MainUI) openTab(tabItem *container.TabItem) {
	// 이미 열려있는지 확인
	for _, item := range m.tabs.Items {
		if item == tabItem {
			m.tabs.Select(tabItem)
			return
		}
	}
	// 새로 추가하고 선택
	m.tabs.Append(tabItem)
	m.tabs.Select(tabItem)
}

// 메인 UI 컨텐츠를 반환합니다.
func (m *MainUI) Content() fyne.CanvasObject {
	// 왼쪽 메뉴 + 오른쪽 탭 영역
	leftPanel := container.NewVBox(
		widget.NewLabel(""), // 상단 여백
		m.leftMenu,
	)

	// 왼쪽 패널에 최소 너비 적용 (150px)
	leftPanelWithSize := container.New(&minWidthLayout{minWidth: 150}, leftPanel)

	// 구분선 추가
	separator := widget.NewSeparator()

	// 왼쪽 메뉴 + 구분선 + 오른쪽 탭 영역
	leftWithSeparator := container.NewBorder(
		nil, nil, nil, separator,
		leftPanelWithSize,
	)

	// 하단 상태바 (왼쪽: 탭 정보, 오른쪽: 버전)
	m.statusLabel = widget.NewLabel("")
	versionLabel := widget.NewLabel(version.GetVersionString())
	bottomBar := container.NewBorder(widget.NewSeparator(), nil, m.statusLabel, versionLabel, nil)

	// 왼쪽 메뉴 너비 고정 (Border 레이아웃 사용)
	return container.NewBorder(
		nil,               // 상단
		bottomBar,         // 하단 (버전 표시)
		leftWithSeparator, // 왼쪽 (고정)
		nil,               // 오른쪽
		m.tabs,            // 중앙 (확장)
	)
}

// minWidthLayout 최소 너비를 보장하는 커스텀 레이아웃
type minWidthLayout struct {
	minWidth float32
}

func (l *minWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(l.minWidth, 0)
	}
	childMin := objects[0].MinSize()
	width := childMin.Width
	if width < l.minWidth {
		width = l.minWidth
	}
	return fyne.NewSize(width, childMin.Height)
}

func (l *minWidthLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, obj := range objects {
		obj.Resize(size)
		obj.Move(fyne.NewPos(0, 0))
	}
}

// 네이티브 메뉴바를 설정합니다.
func (m *MainUI) setupMainMenu() {
	// 파일 메뉴
	fileMenu := fyne.NewMenu("파일",
		fyne.NewMenuItem("Import", func() {
			m.showImportDialog()
		}),
		fyne.NewMenuItem("Export", func() {
			m.showExportDialog()
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Reset", func() {
			m.showResetDialog()
		}),
	)

	// 도구 메뉴
	toolsMenu := fyne.NewMenu("도구",
		fyne.NewMenuItem("설정", func() {
			m.showSettingsDialog()
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("라이트 테마", func() {
			fyne.CurrentApp().Settings().SetTheme(theme.LightTheme())
			m.saveThemeSetting(model.ThemeLight)
		}),
		fyne.NewMenuItem("다크 테마", func() {
			fyne.CurrentApp().Settings().SetTheme(theme.DarkTheme())
			m.saveThemeSetting(model.ThemeDark)
		}),
	)

	// 도움말 메뉴
	helpMenu := fyne.NewMenu("도움말",
		fyne.NewMenuItem("도움말", func() {
			m.showHelpDialog()
		}),
	)

	// 네이티브 메뉴바 생성 및 설정
	mainMenu := fyne.NewMainMenu(fileMenu, toolsMenu, helpMenu)
	m.window.SetMainMenu(mainMenu)
}

// 앱 시작 시 저장된 설정에서 자동 상태 체크를 초기화합니다.
func (m *MainUI) initAutoStatusCheck() {
	config, err := m.store.GetConfig()
	if err != nil {
		return
	}
	if config.AutoStatusCheck {
		m.deviceTab.SetAutoStatusCheck(true, config.GetStatusCheckInterval())
	}
}

// 설정 다이얼로그를 표시합니다.
func (m *MainUI) showSettingsDialog() {
	// 현재 설정 로드
	config, err := m.store.GetConfig()
	if err != nil {
		dialog.ShowError(err, m.window)
		return
	}

	// 타임아웃 입력 필드
	timeoutEntry := widget.NewEntry()
	timeoutEntry.SetText(strconv.Itoa(config.GetTimeoutSeconds()))
	timeoutEntry.SetPlaceHolder("10")

	// 패키지 업데이트 요청 스킴/경로/포트
	programUpdateSchemeSelect := widget.NewSelect([]string{"http", "https"}, nil)
	programUpdateSchemeSelect.SetSelected(config.GetProgramUpdateScheme())
	programUpdatePathEntry := widget.NewEntry()
	programUpdatePathEntry.SetText(config.ProgramUpdatePath)
	programUpdatePathEntry.SetPlaceHolder(model.DefaultProgramUpdatePath)
	programUpdatePortEntry := widget.NewEntry()
	programUpdatePortEntry.SetText(strconv.Itoa(config.ProgramUpdatePort))
	programUpdatePortEntry.SetPlaceHolder(strconv.Itoa(model.DefaultAPIPort))

	// 방화벽 규칙 배포 스킴/경로/포트
	firewallDeploySchemeSelect := widget.NewSelect([]string{"http", "https"}, nil)
	firewallDeploySchemeSelect.SetSelected(config.GetFirewallDeployScheme())
	firewallDeployPathEntry := widget.NewEntry()
	firewallDeployPathEntry.SetText(config.FirewallDeployPath)
	firewallDeployPathEntry.SetPlaceHolder(model.DefaultFirewallDeployPath)
	firewallDeployPortEntry := widget.NewEntry()
	firewallDeployPortEntry.SetText(strconv.Itoa(config.FirewallDeployPort))
	firewallDeployPortEntry.SetPlaceHolder(strconv.Itoa(model.DefaultAPIPort))

	// 장비 상세보기 및 상태 체크 스킴/경로/포트
	deviceInfoSchemeSelect := widget.NewSelect([]string{"http", "https"}, nil)
	deviceInfoSchemeSelect.SetSelected(config.GetDeviceInfoScheme())
	deviceInfoPathEntry := widget.NewEntry()
	deviceInfoPathEntry.SetText(config.DeviceInfoPath)
	deviceInfoPathEntry.SetPlaceHolder(model.DefaultDeviceInfoPath)
	deviceInfoPortEntry := widget.NewEntry()
	deviceInfoPortEntry.SetText(strconv.Itoa(config.DeviceInfoPort))
	deviceInfoPortEntry.SetPlaceHolder(strconv.Itoa(model.DefaultAPIPort))

	// 장비 체크 주기 입력 필드
	statusCheckIntervalEntry := widget.NewEntry()
	statusCheckIntervalEntry.SetText(strconv.Itoa(config.GetStatusCheckInterval()))
	statusCheckIntervalEntry.SetPlaceHolder("60")

	// 설정 경로 표시 (읽기 전용)
	configPathLabel := widget.NewLabel(m.store.GetConfigDir())

	// 경로+포트 컨테이너 생성 함수
	makePathPortRow := func(pathEntry, portEntry *widget.Entry) fyne.CanvasObject {
		return container.NewBorder(nil, nil, nil,
			container.NewHBox(widget.NewLabel("PORT:"), container.NewGridWrap(fyne.NewSize(60, 36), portEntry)),
			pathEntry,
		)
	}

	// 라벨+스킴 선택 + 경로/포트 컨테이너 생성 함수
	makeRouterSettingRow := func(schemeSelect *widget.Select, pathEntry, portEntry *widget.Entry) fyne.CanvasObject {
		pathPortRow := makePathPortRow(pathEntry, portEntry)
		return container.NewVBox(
			container.NewGridWrap(fyne.NewSize(80, 36), schemeSelect),
			pathPortRow,
		)
	}

	// 폼 생성
	formItems := []*widget.FormItem{
		widget.NewFormItem("장비 자동 체크 주기(초)", statusCheckIntervalEntry),
		widget.NewFormItem("", widget.NewLabel("")), // 빈 줄
		widget.NewFormItem("HTTP 응답대기(초)", timeoutEntry),
		widget.NewFormItem("", widget.NewLabel("")), // 빈 줄
		widget.NewFormItem("패키지 업데이트 요청", makeRouterSettingRow(programUpdateSchemeSelect, programUpdatePathEntry, programUpdatePortEntry)),
		widget.NewFormItem("방화벽 규칙 배포", makeRouterSettingRow(firewallDeploySchemeSelect, firewallDeployPathEntry, firewallDeployPortEntry)),
		widget.NewFormItem("장비 보고", makeRouterSettingRow(deviceInfoSchemeSelect, deviceInfoPathEntry, deviceInfoPortEntry)),
		widget.NewFormItem("", widget.NewLabel("")), // 빈 줄
		widget.NewFormItem("설정 저장 경로", configPathLabel),
	}

	// 다이얼로그 표시
	dialog.ShowForm("⚙ 설정", "저장", "취소", formItems, func(ok bool) {
		if !ok {
			return
		}

		// 타임아웃 값 파싱
		timeoutSeconds, err := strconv.Atoi(timeoutEntry.Text)
		if err != nil || timeoutSeconds < 5 || timeoutSeconds > 120 {
			dialog.ShowError(fmt.Errorf("타임아웃은 5~120 사이의 숫자를 입력해주세요"), m.window)
			return
		}

		// 포트 값 파싱
		programUpdatePort, err := strconv.Atoi(programUpdatePortEntry.Text)
		if err != nil || programUpdatePort < 1 || programUpdatePort > 65535 {
			dialog.ShowError(fmt.Errorf("패키지 업데이트 포트는 1~65535 사이의 숫자를 입력해주세요"), m.window)
			return
		}
		firewallDeployPort, err := strconv.Atoi(firewallDeployPortEntry.Text)
		if err != nil || firewallDeployPort < 1 || firewallDeployPort > 65535 {
			dialog.ShowError(fmt.Errorf("방화벽 배포 포트는 1~65535 사이의 숫자를 입력해주세요"), m.window)
			return
		}
		deviceInfoPort, err := strconv.Atoi(deviceInfoPortEntry.Text)
		if err != nil || deviceInfoPort < 1 || deviceInfoPort > 65535 {
			dialog.ShowError(fmt.Errorf("장비 상세보기 포트는 1~65535 사이의 숫자를 입력해주세요"), m.window)
			return
		}

		// 장비 체크 주기 값 파싱
		statusCheckInterval, err := strconv.Atoi(statusCheckIntervalEntry.Text)
		if err != nil || statusCheckInterval < 10 || statusCheckInterval > 300 {
			dialog.ShowError(fmt.Errorf("장비 체크 주기는 10~300 사이의 숫자를 입력해주세요"), m.window)
			return
		}

		// 설정 저장
		newConfig := &model.Config{
			TimeoutSeconds:       timeoutSeconds,
			Theme:                config.Theme,
			AutoStatusCheck:      config.AutoStatusCheck,
			StatusCheckInterval:  statusCheckInterval,
			ProgramUpdateScheme:  programUpdateSchemeSelect.Selected,
			ProgramUpdatePath:    programUpdatePathEntry.Text,
			ProgramUpdatePort:    programUpdatePort,
			FirewallDeployScheme: firewallDeploySchemeSelect.Selected,
			FirewallDeployPath:   firewallDeployPathEntry.Text,
			FirewallDeployPort:   firewallDeployPort,
			DeviceInfoScheme:     deviceInfoSchemeSelect.Selected,
			DeviceInfoPath:       deviceInfoPathEntry.Text,
			DeviceInfoPort:       deviceInfoPort,
		}

		if err := m.store.SaveConfig(newConfig); err != nil {
			dialog.ShowError(err, m.window)
			return
		}

		// DeviceTab에 장비 체크 주기 설정 적용
		m.deviceTab.SetAutoStatusCheck(config.AutoStatusCheck, statusCheckInterval)

		dialog.ShowInformation("성공", "설정이 저장되었습니다.", m.window)
	}, m.window)
}

// 도움말 다이얼로그를 표시합니다.
func (m *MainUI) showHelpDialog() {
	component.ShowHelpPopup("도움말", component.AppHelpText, m.window.Canvas().Content())
}

// 현재 선택된 탭 유형을 반환합니다. (0: 방화벽, 1: 장비, 2: 이력, 3: 패키지, -1: 없음)
func (m *MainUI) getSelectedTabType() int {
	selected := m.tabs.Selected()
	if selected == nil {
		return -1
	}
	switch selected {
	case m.firewallTabItem:
		return 0
	case m.deviceTabItem:
		return 1
	case m.historyTabItem:
		return 2
	case m.programTabItem:
		return 3
	default:
		return -1
	}
}

// Import 다이얼로그를 표시합니다.
func (m *MainUI) showImportDialog() {
	// 현재 탭에 따라 데이터 종류 결정
	tabType := m.getSelectedTabType()
	if tabType < 0 {
		dialog.ShowInformation("알림", "먼저 탭을 선택해주세요.", m.window)
		return
	}

	// 방화벽 관리 탭은 Import 지원 안함 (파일 기반)
	if tabType == 0 {
		dialog.ShowInformation("알림", "방화벽 관리 탭은 Import 기능을 지원하지 않습니다.\n파일추가/수정 버튼을 사용해주세요.", m.window)
		return
	}

	// 배포 이력 탭은 Import 지원 안함 (SQLite DB)
	if tabType == 2 {
		dialog.ShowInformation("알림", "배포 이력 탭은 Import 기능을 지원하지 않습니다.", m.window)
		return
	}

	// 탭별 데이터 타입명 (방화벽 관리, 배포 이력 제외)
	tabNames := []string{"", "장비", "", "패키지"}
	tabName := tabNames[tabType]

	// 파일 선택 다이얼로그
	openDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, m.window)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()

		// JSON 읽기
		data, err := os.ReadFile(reader.URI().Path())
		if err != nil {
			dialog.ShowError(err, m.window)
			return
		}

		// 확인 팝업 표시
		dialog.ShowConfirm("Import 확인",
			fmt.Sprintf("기존 %s 데이터가 모두 삭제되고 새로운 데이터로 교체됩니다.\n계속 진행하시겠습니까?", tabName),
			func(confirmed bool) {
				if !confirmed {
					return
				}

				// 현재 탭에 따라 처리
				switch tabType {
				case 1: // 장비 관리 탭
					var firewalls []*model.Firewall
					if err := json.Unmarshal(data, &firewalls); err != nil {
						dialog.ShowError(fmt.Errorf("JSON 형태의 파일이 아닙니다: %v", err), m.window)
						return
					}

					// 기존 데이터 모두 삭제
					if err := m.store.ClearFirewalls(); err != nil {
						dialog.ShowError(err, m.window)
						return
					}

					// 장비 형식 검증: deviceName이 유효한지 확인
					validCount := 0
					for _, fw := range firewalls {
						if fw.DeviceName == "" || fw.DeviceName == "-" {
							continue // 유효하지 않은 장비는 건너뜀
						}
						if err := m.store.SaveFirewall(fw); err != nil {
							dialog.ShowError(err, m.window)
							return
						}
						validCount++
					}
					if validCount == 0 {
						dialog.ShowError(fmt.Errorf("유효한 장비 데이터가 없습니다. 장비 형식의 JSON 파일을 선택해주세요."), m.window)
						return
					}
					dialog.ShowInformation("성공", fmt.Sprintf("%d개의 장비 정보가 가져오기 되었습니다.", validCount), m.window)
					m.deviceTab.ReloadDevices()
				case 3: // 패키지 탭
					var programs []*model.ProcessInfo
					if err := json.Unmarshal(data, &programs); err != nil {
						dialog.ShowError(fmt.Errorf("JSON 형태의 파일이 아닙니다: %v", err), m.window)
						return
					}

					// 기존 데이터 모두 삭제
					if err := m.store.ClearPrograms(); err != nil {
						dialog.ShowError(err, m.window)
						return
					}

					// 패키지 형식 검증: ProcessName이 유효한지 확인
					validCount := 0
					for _, p := range programs {
						if p.ProcessName == "" || p.ProcessName == "-" {
							continue // 유효하지 않은 패키지은 건너뜀
						}
						if err := m.store.SaveProgram(p); err != nil {
							dialog.ShowError(err, m.window)
							return
						}
						validCount++
					}
					if validCount == 0 {
						dialog.ShowError(fmt.Errorf("유효한 패키지 데이터가 없습니다. 패키지 형식의 JSON 파일을 선택해주세요."), m.window)
						return
					}
					dialog.ShowInformation("성공", fmt.Sprintf("%d개의 패키지이 가져오기 되었습니다.", validCount), m.window)
					m.programTab.RefreshPrograms()
				}
			}, m.window)
	}, m.window)

	openDialog.SetFilter(fynestorage.NewExtensionFileFilter([]string{".json"}))

	// 실행 파일 위치의 config 폴더를 시작 경로로 설정, 없으면 실행 파일 디렉토리
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		configDir := filepath.Join(exeDir, "config")
		if uri, err := fynestorage.ListerForURI(fynestorage.NewFileURI(configDir)); err == nil {
			openDialog.SetLocation(uri)
		} else {
			// config 폴더가 없으면 실행 파일 디렉토리로 설정
			if uri, err := fynestorage.ListerForURI(fynestorage.NewFileURI(exeDir)); err == nil {
				openDialog.SetLocation(uri)
			}
		}
	}

	openDialog.Show()
}

// Export 다이얼로그를 표시합니다.
func (m *MainUI) showExportDialog() {
	// 현재 탭에 따라 데이터 종류 결정
	tabType := m.getSelectedTabType()
	if tabType < 0 {
		dialog.ShowInformation("알림", "먼저 탭을 선택해주세요.", m.window)
		return
	}

	// 방화벽 관리 탭은 Export 지원 안함 (파일 기반)
	if tabType == 0 {
		dialog.ShowInformation("알림", "방화벽 관리 탭은 Export 기능을 지원하지 않습니다.\ndata 디렉토리에서 파일을 직접 복사해주세요.", m.window)
		return
	}

	// 배포 이력 탭은 Export 지원 안함 (SQLite DB)
	if tabType == 2 {
		dialog.ShowInformation("알림", "배포 이력 탭은 Export 기능을 지원하지 않습니다.", m.window)
		return
	}

	// 데이터 확인
	switch tabType {
	case 1: // 장비 관리 탭
		firewalls, err := m.store.GetAllFirewalls()
		if err != nil {
			dialog.ShowError(err, m.window)
			return
		}
		if len(firewalls) == 0 {
			dialog.ShowInformation("알림", "내보낼 장비 정보가 없습니다.", m.window)
			return
		}
	case 3: // 패키지 탭
		programs, err := m.store.GetAllPrograms()
		if err != nil {
			dialog.ShowError(err, m.window)
			return
		}
		if len(programs) == 0 {
			dialog.ShowInformation("알림", "내보낼 패키지이 없습니다.", m.window)
			return
		}
	}

	// 파일 저장 다이얼로그
	saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, m.window)
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()

		var data []byte
		var jsonErr error

		// 현재 탭에 따라 처리
		switch tabType {
		case 1: // 장비 관리 탭
			firewalls, err := m.store.GetAllFirewalls()
			if err != nil {
				dialog.ShowError(err, m.window)
				return
			}
			data, jsonErr = json.MarshalIndent(firewalls, "", "  ")
		case 3: // 패키지 탭
			programs, err := m.store.GetAllPrograms()
			if err != nil {
				dialog.ShowError(err, m.window)
				return
			}
			data, jsonErr = json.MarshalIndent(programs, "", "  ")
		}

		if jsonErr != nil {
			dialog.ShowError(jsonErr, m.window)
			return
		}

		if _, err := writer.Write(data); err != nil {
			dialog.ShowError(err, m.window)
			return
		}

		dialog.ShowInformation("성공", "데이터가 내보내기 되었습니다.", m.window)
	}, m.window)

	// 기본 파일명 설정
	switch tabType {
	case 1:
		saveDialog.SetFileName("firewallList.json")
	case 3:
		saveDialog.SetFileName("programList.json")
	}

	// 실행 파일 위치의 config 폴더를 시작 경로로 설정, 없으면 실행 파일 디렉토리
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		configDir := filepath.Join(exeDir, "config")
		if uri, err := fynestorage.ListerForURI(fynestorage.NewFileURI(configDir)); err == nil {
			saveDialog.SetLocation(uri)
		} else {
			// config 폴더가 없으면 실행 파일 디렉토리로 설정
			if uri, err := fynestorage.ListerForURI(fynestorage.NewFileURI(exeDir)); err == nil {
				saveDialog.SetLocation(uri)
			}
		}
	}

	saveDialog.Show()
}

// 데이터 초기화 다이얼로그를 표시합니다.
func (m *MainUI) showResetDialog() {
	// 경고 다이얼로그 표시
	dialog.ShowConfirm("⚠️ 경고",
		"모든 데이터(방화벽 파일, 장비, 배포이력, 패키지, 앱설정)를 초기화하시겠습니까?",
		func(ok bool) {
			if !ok {
				return
			}

			// 모든 방화벽 파일 삭제
			if err := m.fileStore.ClearAllFiles(); err != nil {
				dialog.ShowError(err, m.window)
				return
			}

			// 모든 장비 삭제
			if err := m.store.ClearFirewalls(); err != nil {
				dialog.ShowError(err, m.window)
				return
			}

			// 모든 배포이력 삭제 (SQLite)
			if err := m.historyRepo.Clear(); err != nil {
				dialog.ShowError(err, m.window)
				return
			}

			// 모든 패키지 삭제
			if err := m.store.ClearPrograms(); err != nil {
				dialog.ShowError(err, m.window)
				return
			}

			// 설정 초기화 (기본값으로)
			if err := m.store.SaveConfig(model.DefaultConfig()); err != nil {
				dialog.ShowError(err, m.window)
				return
			}

			// UI 초기화 (서버 상태 체크 없이, 다이얼로그 없이)
			m.firewallTab.RefreshFiles()
			m.deviceTab.ReloadDevices()
			m.historyTab.ReloadHistory()
			m.programTab.RefreshPrograms()

			dialog.ShowInformation("완료", "모든 데이터가 초기화되었습니다.", m.window)
		}, m.window)
}

// OpenFirewallEditTab 방화벽 파일 편집 탭을 엽니다.
func (m *MainUI) OpenFirewallEditTab(file *model.FirewallFile) {
	// 이미 열려있는지 확인
	if existingTab, ok := m.editTabs[file.FileName]; ok {
		m.tabs.Select(existingTab)
		// 상태바 업데이트
		m.updateStatusBar(existingTab)
		return
	}

	// 새 편집 탭 생성
	editTab := NewFirewallEditTab(m.window, m.fileStore, m, file)

	// 탭 이름 설정 (truncation 적용)
	tabName := editTab.GetTabName()
	tabItem := container.NewTabItem(tabName, editTab.Content())

	// 탭 추가 및 선택
	m.tabs.Append(tabItem)
	m.tabs.Select(tabItem)

	// editTabs 맵에 등록
	m.editTabs[file.FileName] = tabItem

	// 상태바 업데이트
	m.updateStatusBar(tabItem)
}

// CloseFirewallEditTab 방화벽 파일 편집 탭을 닫습니다.
func (m *MainUI) CloseFirewallEditTab(fileName string) {
	if tabItem, ok := m.editTabs[fileName]; ok {
		m.tabs.Remove(tabItem)
		delete(m.editTabs, fileName)
	}
}

// UpdateFirewallEditTabName 편집 탭의 파일명이 변경되었을 때 탭 이름을 업데이트합니다.
func (m *MainUI) UpdateFirewallEditTabName(oldFileName string, file *model.FirewallFile) {
	if tabItem, ok := m.editTabs[oldFileName]; ok {
		// 맵에서 기존 항목 제거하고 새 파일명으로 등록
		delete(m.editTabs, oldFileName)
		m.editTabs[file.FileName] = tabItem

		// 탭 이름 업데이트
		maxLen := 25
		name := "방화벽관리/" + file.FileName
		if len(name) > maxLen {
			name = name[:maxLen-3] + "..."
		}
		tabItem.Text = name
		m.tabs.Refresh()

		// 현재 선택된 탭이면 상태바도 업데이트
		if m.tabs.Selected() == tabItem {
			m.updateStatusBar(tabItem)
		}
	}
}

// 테마 설정을 저장합니다.
func (m *MainUI) saveThemeSetting(themeName string) {
	config, err := m.store.GetConfig()
	if err != nil {
		config = model.DefaultConfig()
	}
	config.Theme = themeName
	m.store.SaveConfig(config)
}

// setupFileDrop 파일 드래그 앤 드롭 핸들러를 설정합니다.
func (m *MainUI) setupFileDrop() {
	m.window.SetOnDropped(func(pos fyne.Position, uris []fyne.URI) {
		if len(uris) == 0 {
			return
		}

		// 추가할 파일 목록 생성 (중복 제외)
		var filesToAdd []fyne.URI
		var duplicateFiles []string

		for _, uri := range uris {
			fileName := filepath.Base(uri.Path())
			if m.fileStore.FileExists(fileName) {
				duplicateFiles = append(duplicateFiles, fileName)
			} else {
				filesToAdd = append(filesToAdd, uri)
			}
		}

		// 추가할 파일이 없으면 알림
		if len(filesToAdd) == 0 {
			dialog.ShowInformation("알림", "모든 파일이 이미 존재합니다.", m.window)
			return
		}

		// 확인 메시지 생성
		msg := fmt.Sprintf("%d개 파일을 추가하시겠습니까?", len(filesToAdd))
		if len(duplicateFiles) > 0 {
			msg += fmt.Sprintf("\n(%d개 파일은 중복으로 제외됩니다)", len(duplicateFiles))
		}

		// 확인 다이얼로그 표시
		dialog.ShowConfirm("파일 추가", msg, func(confirmed bool) {
			if !confirmed {
				return
			}

			// 파일 복사
			for _, uri := range filesToAdd {
				fileName := filepath.Base(uri.Path())
				m.fileStore.CopyFileToData(uri.Path(), fileName)
			}

			// 테이블 새로고침
			m.firewallTab.RefreshFiles()
		}, m.window)
	})
}

// updateStatusBar 현재 선택된 탭에 따라 상태바를 업데이트합니다.
func (m *MainUI) updateStatusBar(tab *container.TabItem) {
	if tab == nil || m.statusLabel == nil {
		return
	}

	// 편집 탭인 경우 파일 경로 표시
	for fileName, tabItem := range m.editTabs {
		if tabItem == tab {
			// 파일 경로 조합
			filePath := filepath.Join(m.fileStore.GetDataDir(), fileName)
			m.statusLabel.SetText(filePath)
			return
		}
	}

	// 기본 탭인 경우 빈 문자열
	m.statusLabel.SetText("")
}
