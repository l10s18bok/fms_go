// Package ui는 FMS 애플리케이션의 사용자 인터페이스를 구현합니다.
package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"fms/internal/model"
	"fms/internal/storage"

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
	tabs        *container.AppTabs
	templateTab *TemplateTab
	deviceTab   *DeviceTab
	historyTab  *HistoryTab
}

// 새로운 메인 UI 인스턴스를 생성합니다.
func NewMainUI(window fyne.Window, store *storage.JSONStore) *MainUI {
	ui := &MainUI{
		window: window,
		store:  store,
	}

	// 각 탭 생성
	ui.templateTab = NewTemplateTab(window, store)
	ui.deviceTab = NewDeviceTab(window, store, ui.templateTab)
	ui.historyTab = NewHistoryTab(window, store)

	// 탭 간 참조 설정
	ui.deviceTab.SetHistoryTab(ui.historyTab)
	ui.historyTab.SetDeviceTab(ui.deviceTab)

	// 탭 컨테이너 생성
	ui.tabs = container.NewAppTabs(
		container.NewTabItemWithIcon("템플릿 관리", theme.DocumentIcon(), ui.templateTab.Content()),
		container.NewTabItemWithIcon("장비 관리", theme.ComputerIcon(), ui.deviceTab.Content()),
		container.NewTabItemWithIcon("배포 이력", theme.HistoryIcon(), ui.historyTab.Content()),
	)
	ui.tabs.SetTabLocation(container.TabLocationTop)

	// 탭 변경 시 이벤트 처리
	ui.tabs.OnSelected = func(tab *container.TabItem) {
		if ui.tabs.SelectedIndex() == 1 { // 장비 관리 탭
			ui.deviceTab.RefreshTemplates()
		}
	}

	// 네이티브 메뉴바 설정
	ui.setupMainMenu()

	return ui
}

// 메인 UI 컨텐츠를 반환합니다.
func (m *MainUI) Content() fyne.CanvasObject {
	// 탭만 반환 (메뉴바는 네이티브 메뉴로 이동)
	return m.tabs
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

// 설정 다이얼로그를 표시합니다.
func (m *MainUI) showSettingsDialog() {
	// 현재 설정 로드
	config, err := m.store.GetConfig()
	if err != nil {
		dialog.ShowError(err, m.window)
		return
	}

	// 연결 모드 라디오 그룹 (Agent Server는 임시로 비활성화)
	connectionMode := widget.NewRadioGroup([]string{"Agent Server (준비중)", "Direct"}, nil)
	connectionMode.SetSelected("Direct")
	connectionMode.Disable() // Agent 모드 임시 비활성화

	// Agent Server URL 입력 필드 (Agent 모드 비활성화로 인해 항상 비활성화)
	agentURLEntry := widget.NewEntry()
	agentURLEntry.SetText(config.AgentServerURL)
	agentURLEntry.SetPlaceHolder("http://172.24.10.6:8080")
	agentURLEntry.Disable() // Agent 모드 임시 비활성화

	// 타임아웃 입력 필드
	timeoutEntry := widget.NewEntry()
	timeoutEntry.SetText(strconv.Itoa(config.GetTimeoutSeconds()))
	timeoutEntry.SetPlaceHolder("10")

	// 연결 모드에 따라 URL 입력 필드 활성화/비활성화
	updateURLEntryState := func() {
		if connectionMode.Selected == "Agent Server" {
			agentURLEntry.Enable()
		} else {
			agentURLEntry.Disable()
		}
	}
	updateURLEntryState()

	connectionMode.OnChanged = func(selected string) {
		updateURLEntryState()
	}

	// 설정 경로 표시 (읽기 전용)
	configPathLabel := widget.NewLabel(m.store.GetConfigDir())

	// 폼 생성
	formItems := []*widget.FormItem{
		widget.NewFormItem("Connection", connectionMode),
		widget.NewFormItem("Agent Server URL", agentURLEntry),
		widget.NewFormItem("Timeout (초)", timeoutEntry),
		widget.NewFormItem("", widget.NewLabel("")), // 빈 줄
		widget.NewFormItem("설정 저장 경로", configPathLabel),
	}

	// 다이얼로그 표시
	dialog.ShowForm("설정", "저장", "취소", formItems, func(ok bool) {
		if !ok {
			return
		}

		// 연결 모드 설정
		var newConnectionMode string
		if connectionMode.Selected == "Agent Server" {
			newConnectionMode = model.ConnectionModeAgent
		} else {
			newConnectionMode = model.ConnectionModeDirect
		}

		// Agent Server URL 검증 (Agent 모드일 경우)
		if newConnectionMode == model.ConnectionModeAgent && agentURLEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("Agent Server URL을 입력해주세요"), m.window)
			return
		}

		// 타임아웃 값 파싱
		timeoutSeconds, err := strconv.Atoi(timeoutEntry.Text)
		if err != nil || timeoutSeconds < 5 || timeoutSeconds > 120 {
			dialog.ShowError(fmt.Errorf("타임아웃은 5~120 사이의 숫자를 입력해주세요"), m.window)
			return
		}

		// 설정 저장
		newConfig := &model.Config{
			ConnectionMode: newConnectionMode,
			AgentServerURL: agentURLEntry.Text,
			TimeoutSeconds: timeoutSeconds,
		}

		if err := m.store.SaveConfig(newConfig); err != nil {
			dialog.ShowError(err, m.window)
			return
		}

		dialog.ShowInformation("성공", "설정이 저장되었습니다.", m.window)
	}, m.window)
}

// 도움말 다이얼로그를 표시합니다.
func (m *MainUI) showHelpDialog() {
	helpText := `FMS - Firewall Management System

버전: 1.1.0

[템플릿 관리]
• 방화벽 규칙 템플릿을 생성/수정/삭제합니다

[장비 관리]
• 관리할 방화벽 장비(IP)를 등록합니다
• 서버 상태를 확인하고 템플릿을 배포합니다

[배포 이력]
• 배포 결과를 확인할 수 있습니다
• 규칙별 성공/실패 상태를 상세히 확인합니다

[Import/Export]
• 현재 탭의 데이터를 JSON 파일로 내보내거나 가져옵니다

[연결 모드] (설정에서 변경)
• Agent Server: Agent 서버(예: http://172.24.10.6:8080)를 통해 연결
  - 상태확인: POST /agent/req-respCheck
  - 배포: POST /agent/req-deploy
• Direct: 각 장비에 직접 HTTP 연결 (포트 80)
  - 상태확인: GET http://{장비IP}/respCheck
  - 배포: POST http://{장비IP}/deploy

[규칙 포맷]
req|INSERT|{ID}|{CHAIN}|{ACTION}|{PROTOCOL}|{SRC}|{DST}|{옵션들}

예시:
req|INSERT|3813792919|INPUT|FLUSH|ANY|ANY|ANY|||
req|INSERT|3813792919|INPUT|ACCEPT|TCP|192.168.1.0/24|ANY|80||`

	// 도움말 내용 표시
	helpLabel := widget.NewLabel(helpText)
	helpLabel.Wrapping = fyne.TextWrapWord

	scrollContent := container.NewScroll(helpLabel)
	scrollContent.SetMinSize(fyne.NewSize(500, 400))

	dialog.ShowCustom("도움말", "닫기", scrollContent, m.window)
}

// Import 다이얼로그를 표시합니다.
func (m *MainUI) showImportDialog() {
	// 현재 탭에 따라 데이터 종류 결정
	tabIndex := m.tabs.SelectedIndex()

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

		// 현재 탭에 따라 처리
		switch tabIndex {
		case 0: // 템플릿 탭
			var templates []*model.Template
			if err := json.Unmarshal(data, &templates); err != nil {
				dialog.ShowError(fmt.Errorf("JSON 형태의 파일이 아닙니다: %v", err), m.window)
				return
			}
			// 템플릿 형식 검증: version과 contents가 유효한지 확인
			validCount := 0
			for _, tmpl := range templates {
				if tmpl.Version == "" || tmpl.Version == "-" || tmpl.Contents == "" || tmpl.Contents == "-" {
					continue // 유효하지 않은 템플릿은 건너뜀
				}
				if err := m.store.SaveTemplate(tmpl); err != nil {
					dialog.ShowError(err, m.window)
					return
				}
				validCount++
			}
			if validCount == 0 {
				dialog.ShowError(fmt.Errorf("유효한 템플릿 데이터가 없습니다. 템플릿 형식의 JSON 파일을 선택해주세요."), m.window)
				return
			}
			dialog.ShowInformation("성공", fmt.Sprintf("%d개의 템플릿이 가져오기 되었습니다.", validCount), m.window)
			m.templateTab.RefreshTemplates()
		case 1: // 장비 관리 탭
			var firewalls []*model.Firewall
			if err := json.Unmarshal(data, &firewalls); err != nil {
				dialog.ShowError(fmt.Errorf("JSON 형태의 파일이 아닙니다: %v", err), m.window)
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
		case 2: // 배포 이력 탭
			var histories []*model.DeployHistory
			if err := json.Unmarshal(data, &histories); err != nil {
				dialog.ShowError(fmt.Errorf("JSON 형태의 파일이 아닙니다: %v", err), m.window)
				return
			}
			// 배포 이력 형식 검증: deviceIp와 templateVersion이 유효한지 확인
			validCount := 0
			for _, h := range histories {
				if h.DeviceIP == "" || h.DeviceIP == "-" || h.TemplateVer == "" || h.TemplateVer == "-" {
					continue // 유효하지 않은 이력은 건너뜀
				}
				if err := m.store.SaveHistory(h); err != nil {
					dialog.ShowError(err, m.window)
					return
				}
				validCount++
			}
			if validCount == 0 {
				dialog.ShowError(fmt.Errorf("유효한 배포 이력 데이터가 없습니다. 배포 이력 형식의 JSON 파일을 선택해주세요."), m.window)
				return
			}
			dialog.ShowInformation("성공", fmt.Sprintf("%d개의 배포 이력이 가져오기 되었습니다.", validCount), m.window)
			m.historyTab.RefreshHistory()
		}
	}, m.window)

	openDialog.SetFilter(fynestorage.NewExtensionFileFilter([]string{".json"}))

	// 실행 파일 위치의 config 폴더를 시작 경로로 설정, 없으면 실행 파일 디렉토리
	exePath, _ := os.Executable()
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

	openDialog.Show()
}

// Export 다이얼로그를 표시합니다.
func (m *MainUI) showExportDialog() {
	// 현재 탭에 따라 데이터 종류 결정
	tabIndex := m.tabs.SelectedIndex()

	// 데이터 확인
	switch tabIndex {
	case 0: // 템플릿 탭
		templates, err := m.store.GetAllTemplates()
		if err != nil {
			dialog.ShowError(err, m.window)
			return
		}
		if len(templates) == 0 {
			dialog.ShowInformation("알림", "내보낼 템플릿이 없습니다.", m.window)
			return
		}
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
	case 2: // 배포 이력 탭
		histories, err := m.store.GetAllHistory()
		if err != nil {
			dialog.ShowError(err, m.window)
			return
		}
		if len(histories) == 0 {
			dialog.ShowInformation("알림", "내보낼 배포 이력이 없습니다.", m.window)
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
		switch tabIndex {
		case 0: // 템플릿 탭
			templates, _ := m.store.GetAllTemplates()
			data, jsonErr = json.MarshalIndent(templates, "", "  ")
		case 1: // 장비 관리 탭
			firewalls, _ := m.store.GetAllFirewalls()
			data, jsonErr = json.MarshalIndent(firewalls, "", "  ")
		case 2: // 배포 이력 탭
			histories, _ := m.store.GetAllHistory()
			data, jsonErr = json.MarshalIndent(histories, "", "  ")
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
	switch tabIndex {
	case 0:
		saveDialog.SetFileName("templateList.json")
	case 1:
		saveDialog.SetFileName("firewallList.json")
	case 2:
		saveDialog.SetFileName("historyList.json")
	}

	// 실행 파일 위치의 config 폴더를 시작 경로로 설정, 없으면 실행 파일 디렉토리
	exePath, _ := os.Executable()
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

	saveDialog.Show()
}

// 데이터 초기화 다이얼로그를 표시합니다.
func (m *MainUI) showResetDialog() {
	// 경고 다이얼로그 표시
	dialog.ShowConfirm("⚠️ 경고",
		"모든 데이터(템플릿, 장비, 배포이력)를 초기화하시겠습니까?",
		func(ok bool) {
			if !ok {
				return
			}

			// 모든 템플릿 삭제
			if err := m.store.ClearTemplates(); err != nil {
				dialog.ShowError(err, m.window)
				return
			}

			// 모든 장비 삭제
			if err := m.store.ClearFirewalls(); err != nil {
				dialog.ShowError(err, m.window)
				return
			}

			// 모든 배포이력 삭제
			if err := m.store.ClearHistory(); err != nil {
				dialog.ShowError(err, m.window)
				return
			}

			// UI 초기화 (서버 상태 체크 없이, 다이얼로그 없이)
			m.templateTab.ClearSelection()
			m.templateTab.RefreshTemplates()
			m.deviceTab.ReloadDevices()
			m.historyTab.ReloadHistory()

			dialog.ShowInformation("완료", "모든 데이터가 초기화되었습니다.", m.window)
		}, m.window)
}
