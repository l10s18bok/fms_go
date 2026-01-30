package ui

import (
	"fmt"
	"image/color"
	"log"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"fms/internal/deploy"
	// "fms/internal/http" // 임시 주석처리 - 프로세스 리스트 기능 비활성화
	"fms/internal/model"
	"fms/internal/repository"
	"fms/internal/storage"
	"fms/internal/themes"
	"fms/internal/ui/component"
	"fms/internal/usecase"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"

	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	ttwidget "github.com/dweymouth/fyne-tooltip/widget"
)

// 장비 관리 탭을 구현합니다. (PRD 3.3.3 기준)
type DeviceTab struct {
	window       fyne.Window
	firewallRepo repository.FirewallRepository
	programRepo  repository.ProgramRepository
	configStore  *storage.JSONStore // Config 접근용
	firewallTab  *FirewallTab
	historyTab   *HistoryTab
	programTab   *ProgramTab
	content      fyne.CanvasObject

	// UI 컴포넌트
	searchBox   *component.SearchBox  // 검색 컴포넌트 (공통)
	calendarBtn *ttwidget.Button      // 날짜 선택 아이콘 버튼
	deviceTable *component.PagedTable // 장비 테이블 (공통 컴포넌트)

	// 상태 요약 표시
	statusGreenLabel  *widget.Label
	statusYellowLabel *widget.Label
	statusRedLabel    *widget.Label

	// 데이터 (DB 페이지네이션)
	firewalls     []*model.Firewall // 전체 장비 (상태 체크용)
	pageData      []*model.Firewall // 현재 페이지 데이터
	totalCount    int               // 검색 적용 후 전체 건수
	searchKeyword string            // 검색 키워드
	startDate     string            // 시작일
	endDate       string            // 종료일

	// 새로고침 상태
	isRefreshing bool
	refreshBtn   *ttwidget.Button

	// 자동 상태 체크
	autoCheckEnabled  bool
	autoCheckInterval int
	autoCheckTicker   *time.Ticker
	stopAutoCheck     chan struct{}
	statusBorder      *canvas.Rectangle // 상태 박스 테두리
	autoCheckBtn      *ttwidget.Button  // 자동 상태 체크 토글 버튼
}

// 새로운 장비 관리 탭을 생성합니다.
func NewDeviceTab(window fyne.Window, firewallRepo repository.FirewallRepository, programRepo repository.ProgramRepository, configStore *storage.JSONStore, firewallTab *FirewallTab) *DeviceTab {
	tab := &DeviceTab{
		window:            window,
		firewallRepo:      firewallRepo,
		programRepo:       programRepo,
		configStore:       configStore,
		firewallTab:       firewallTab,
		firewalls: []*model.Firewall{},
		pageData:  []*model.Firewall{},
	}
	tab.createUI()
	tab.loadFirewalls()

	// config에서 자동 상태 체크 설정 동기화
	tab.SyncAutoCheckFromConfig()

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
		Placeholder: "",
		Width:       200,
		OnSearch: func(text string) {
			d.applyFilter()
		},
	})

	// 날짜 선택 아이콘 버튼 (Calendar 팝업)
	d.calendarBtn = ttwidget.NewButtonWithIcon("", theme.Icon(theme.IconNameCalendar), func() {
		calendar := widget.NewCalendar(time.Now(), func(t time.Time) {
			dateStr := t.Format("2006-01-02")
			current := strings.TrimSpace(d.searchBox.GetText())
			if current != "" && !strings.Contains(current, "~") {
				d.searchBox.SetText(current + " ~ " + dateStr)
			} else {
				d.searchBox.SetText(dateStr)
			}
		})
		c := fyne.CurrentApp().Driver().CanvasForObject(d.calendarBtn)
		pop := widget.NewPopUp(calendar, c)
		btnPos := fyne.CurrentApp().Driver().AbsolutePositionForObject(d.calendarBtn)
		pop.ShowAtPosition(fyne.NewPos(btnPos.X, btnPos.Y+d.calendarBtn.Size().Height))
	})
	d.calendarBtn.Importance = widget.LowImportance
	d.calendarBtn.SetToolTip("기간검색")

	// 상태 요약 레이블
	d.statusGreenLabel = widget.NewLabel("0")
	d.statusYellowLabel = widget.NewLabel("0")
	d.statusRedLabel = widget.NewLabel("0")

	// 색상 원 (widget.Label 사용하여 수직 정렬 일치)
	greenDot := widget.NewLabelWithStyle("●", fyne.TextAlignCenter, fyne.TextStyle{})
	greenDot.Importance = widget.SuccessImportance
	yellowDot := widget.NewLabelWithStyle("●", fyne.TextAlignCenter, fyne.TextStyle{})
	yellowDot.Importance = widget.WarningImportance
	redDot := widget.NewLabelWithStyle("●", fyne.TextAlignCenter, fyne.TextStyle{})
	redDot.Importance = widget.DangerImportance

	// 자동 상태 체크 토글 버튼 (초기: Play 아이콘)
	d.autoCheckBtn = ttwidget.NewButtonWithIcon("", theme.MediaPlayIcon(), func() {
		d.onToggleAutoCheck()
	})
	d.autoCheckBtn.SetToolTip("전체장비 자동체크")

	// 새로고침 버튼 (🔄)
	d.refreshBtn = ttwidget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		d.onRefreshAll()
	})
	d.refreshBtn.SetToolTip("선택장비체크")

	statusContent := container.NewHBox(
		greenDot, widget.NewLabel("연결:"), d.statusGreenLabel,
		yellowDot, widget.NewLabel("알수없음:"), d.statusYellowLabel,
		redDot, widget.NewLabel("연결안됨:"), d.statusRedLabel,
	)

	// 라운드 테두리 배경
	borderColor := color.RGBA{R: 200, G: 200, B: 200, A: 255}
	d.statusBorder = canvas.NewRectangle(borderColor)
	d.statusBorder.StrokeWidth = 1
	d.statusBorder.StrokeColor = borderColor
	d.statusBorder.FillColor = color.Transparent
	d.statusBorder.CornerRadius = 8

	// 패딩을 위한 컨테이너 (상하 0, 좌우 10)
	paddedStatus := container.New(layout.NewCustomPaddedLayout(0, 0, 10, 10), statusContent)

	// 테두리와 내용을 스택으로 결합
	statusBox := container.NewStack(d.statusBorder, paddedStatus)

	// 상태 요약 컨테이너 (토글 버튼 + 새로고침 버튼 포함)
	statusSummary := container.NewHBox(
		statusBox,
		d.autoCheckBtn,
		d.refreshBtn,
	)

	// 삭제 버튼 (빨간 배경, 흰색 텍스트)
	deleteBtn := component.NewCustomButton("삭제", theme.DeleteIcon(), nil, themes.Colors["red"], func() {
		d.onDeleteDevices()
	})

	// 배포 버튼 (종이비행기 아이콘)
	deployBtn := component.NewCustomButton("배포", theme.MailSendIcon(), nil, themes.Colors["darkgray"], func() {
		d.onDeploy()
	})

	// 추가/수정 버튼
	addEditBtn := component.NewCustomButton("+추가/수정", nil, nil, themes.Colors["blue"], func() {
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
	headerLine := container.NewBorder(nil, nil, container.NewHBox(d.searchBox.Content(), d.calendarBtn), buttonArea, container.NewCenter(statusSummary))
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
	// PRD 테이블 컬럼: 선택, 장비명, 장비 IP, 위치, 장비상태, OS 정보, 상태 체크시간, CPU, 메모리, 디스크, 네트워크
	d.deviceTable = component.NewPagedTableWithWindow(component.PagedTableConfig{
		Columns: []component.ColumnDef{
			{Header: "선택", Width: 50},
			{Header: "장비명", Width: 150},
			{Header: "장비 IP", Width: 150},
			{Header: "위치", Width: 150},
			{Header: "장비상태", Width: 80},
			{Header: "OS 정보", Width: 150},
			{Header: "상태 체크시간", Width: 150},
			{Header: "CPU", Width: 70},
			{Header: "메모리", Width: 70},
			{Header: "디스크", Width: 70},
			{Header: "네트워크", Width: 100},
		},
		PageSize: 15,
		OnCellUpdate: func(row int, col int, cell fyne.CanvasObject) {
			d.updateDeviceCell(row, col, cell)
		},
		OnCustomCell: func(row int, col int, container *fyne.Container) bool {
			return d.updateCustomCell(row, col, container)
		},
		OnRowSelected: func(row int) {
			// 단일 클릭 - 선택 (현재 미사용)
		},
		OnRowDoubleClick: func(row int) {
			// 더블 클릭 - 상세보기
			if row >= 0 && row < len(d.pageData) {
				d.showDetailDialog(d.pageData[row])
			}
		},
		// 장비명(1), 장비 IP(2), 위치(3), OS정보(5) 컬럼 인라인 편집 설정
		EditableColumns: map[int]component.EditColumnConfig{
			1: { // 장비명 컬럼
				Type: component.EditTypeEntry,
				GetValue: func(row int) string {
					if row >= 0 && row < len(d.pageData) {
						fw := d.pageData[row]
						if fw.DeviceName != "" && fw.DeviceName != fw.DeviceIP {
							return fw.DeviceName
						}
					}
					return ""
				},
				OnEdit: func(row int, oldValue, newValue string) bool {
					if row >= 0 && row < len(d.pageData) {
						fw := d.pageData[row]
						fw.DeviceName = newValue
						// DeviceName이 비어있으면 IP 사용
						if fw.DeviceName == "" {
							fw.DeviceName = fw.DeviceIP
						}
						if err := d.firewallRepo.Save(fw); err != nil {
							dialog.ShowError(err, d.window)
							return false
						}
						return true
					}
					return false
				},
			},
			2: { // 장비 IP 컬럼
				Type: component.EditTypeEntry,
				GetValue: func(row int) string {
					if row >= 0 && row < len(d.pageData) {
						fw := d.pageData[row]
						if fw.DeviceIP != "" {
							return fw.DeviceIP
						}
						return fw.DeviceName
					}
					return ""
				},
				OnEdit: func(row int, oldValue, newValue string) bool {
					if row >= 0 && row < len(d.pageData) {
						// IP 형식 유효성 검사
						if newValue != "" && !isValidIPOrHostPort(newValue) {
							dialog.ShowError(fmt.Errorf("올바른 IP 주소 형식이 아닙니다"), d.window)
							return false
						}
						fw := d.pageData[row]
						fw.DeviceIP = newValue
						if err := d.firewallRepo.Save(fw); err != nil {
							dialog.ShowError(err, d.window)
							return false
						}
						return true
					}
					return false
				},
			},
			3: { // 위치 컬럼
				Type: component.EditTypeEntry,
				GetValue: func(row int) string {
					if row >= 0 && row < len(d.pageData) {
						return d.pageData[row].Location
					}
					return ""
				},
				OnEdit: func(row int, oldValue, newValue string) bool {
					if row >= 0 && row < len(d.pageData) {
						fw := d.pageData[row]
						fw.Location = newValue
						if err := d.firewallRepo.Save(fw); err != nil {
							dialog.ShowError(err, d.window)
							return false
						}
						return true
					}
					return false
				},
			},
			5: { // OS 정보 컬럼
				Type: component.EditTypeEntry,
				GetValue: func(row int) string {
					if row >= 0 && row < len(d.pageData) {
						return d.pageData[row].OSInfo
					}
					return ""
				},
				OnEdit: func(row int, oldValue, newValue string) bool {
					if row >= 0 && row < len(d.pageData) {
						fw := d.pageData[row]
						fw.OSInfo = newValue
						if err := d.firewallRepo.Save(fw); err != nil {
							dialog.ShowError(err, d.window)
							return false
						}
						return true
					}
					return false
				},
			},
		},
		OnPageLoad: func(page, pageSize int) int {
			return d.loadPage(page, pageSize)
		},
	}, d.window)

	return d.deviceTable.Content()
}

// 장비 테이블 셀을 업데이트합니다.
func (d *DeviceTab) updateDeviceCell(row int, col int, cell fyne.CanvasObject) {
	label := cell.(*widget.Label)

	if row >= len(d.pageData) {
		label.SetText("")
		return
	}

	fw := d.pageData[row]

	switch col {
	case 1: // 장비명
		if fw.DeviceName != "" && fw.DeviceName != fw.DeviceIP {
			label.SetText(fw.DeviceName)
		} else {
			label.SetText("-")
		}
	case 2: // 장비 IP
		if fw.DeviceIP != "" {
			label.SetText(fw.DeviceIP)
		} else {
			label.SetText(fw.DeviceName)
		}
	case 3: // 위치
		if fw.Location != "" {
			label.SetText(fw.Location)
		} else {
			label.SetText("-")
		}
	case 4: // 장비상태 - OnCustomCell에서 처리
		// 커스텀 셀에서 처리하므로 여기서는 아무것도 하지 않음
	case 5: // OS 정보
		if fw.OSInfo != "" {
			label.SetText(fw.OSInfo)
		} else {
			label.SetText("-")
		}
	case 6: // 상태 체크시간
		if fw.LastCheckedAt != "" {
			label.SetText(fw.LastCheckedAt)
		} else {
			label.SetText("-")
		}
	case 7: // CPU 사용률
		if fw.ServerStatus == model.ServerStatusRunning && fw.CPUUsage > 0 {
			label.SetText(fmt.Sprintf("%.1f%%", fw.CPUUsage))
		} else {
			label.SetText("-")
		}
	case 8: // 메모리 사용률
		if fw.ServerStatus == model.ServerStatusRunning && fw.MemoryUsage > 0 {
			label.SetText(fmt.Sprintf("%.1f%%", fw.MemoryUsage))
		} else {
			label.SetText("-")
		}
	case 9: // 디스크 사용률
		if fw.ServerStatus == model.ServerStatusRunning && fw.DiskUsage > 0 {
			label.SetText(fmt.Sprintf("%.1f%%", fw.DiskUsage))
		} else {
			label.SetText("-")
		}
	case 10: // 네트워크
		if fw.ServerStatus == model.ServerStatusRunning && fw.NetworkUsage > 0 {
			label.SetText(fmt.Sprintf("%.2f MB/s", fw.NetworkUsage))
		} else {
			label.SetText("-")
		}
	}
}

// 커스텀 셀을 업데이트합니다. (장비상태 컬럼에 색상 원 표시)
func (d *DeviceTab) updateCustomCell(row int, col int, cont *fyne.Container) bool {
	// 장비상태 컬럼(4)만 커스텀 처리
	if col != 4 {
		return false
	}

	if row >= len(d.pageData) {
		return false
	}

	fw := d.pageData[row]

	// 색상 원 생성
	var dotColor = themes.Colors["yellow"] // 기본: 알수없음
	switch fw.ServerStatus {
	case model.ServerStatusRunning:
		dotColor = themes.Colors["green"]
	case model.ServerStatusStop:
		dotColor = themes.Colors["red"]
	}

	dot := canvas.NewText("●", dotColor)
	dot.TextSize = 18
	dot.Alignment = fyne.TextAlignCenter

	cont.Objects = append(cont.Objects, container.NewCenter(dot))
	return true
}

// 검색을 실행합니다.
func (d *DeviceTab) applyFilter() {
	keyword := strings.TrimSpace(d.searchBox.GetText())

	prevKeyword := d.searchKeyword
	prevStartDate := d.startDate
	prevEndDate := d.endDate

	// 날짜 파싱 (단일: YYYY-MM-DD, 범위: YYYY-MM-DD ~ YYYY-MM-DD)
	d.startDate = ""
	d.endDate = ""
	d.searchKeyword = keyword
	if strings.Contains(keyword, "~") {
		parts := strings.SplitN(keyword, "~", 2)
		sd := strings.TrimSpace(parts[0])
		ed := strings.TrimSpace(parts[1])
		if _, err := time.Parse("2006-01-02", sd); err == nil {
			d.startDate = sd
		}
		if _, err := time.Parse("2006-01-02", ed); err == nil {
			d.endDate = ed
		}
		if d.startDate != "" || d.endDate != "" {
			d.searchKeyword = ""
		}
	} else if _, err := time.Parse("2006-01-02", keyword); err == nil {
		d.startDate = keyword
		d.endDate = keyword
		d.searchKeyword = ""
	}
	total := d.loadPage(0, d.deviceTable.GetPageSize())

	// 검색 결과가 없으면 이전 상태 유지
	if total == 0 && keyword != "" {
		dialog.ShowInformation("검색 결과", "검색 결과가 없습니다.", d.window)
		d.searchKeyword = prevKeyword
		d.startDate = prevStartDate
		d.endDate = prevEndDate
		d.loadPage(0, d.deviceTable.GetPageSize())
	}

	if d.deviceTable != nil {
		d.deviceTable.SetData(d.totalCount)
	}
}

// DB에서 해당 페이지 데이터를 조회합니다.
func (d *DeviceTab) loadPage(page, pageSize int) int {
	req := model.PageRequest{
		Offset:    page * pageSize,
		Limit:     pageSize,
		Keyword:   d.searchKeyword,
		StartDate: d.startDate,
		EndDate:   d.endDate,
	}
	result, err := d.firewallRepo.GetPage(req)
	if err != nil {
		d.pageData = []*model.Firewall{}
		d.totalCount = 0
		return 0
	}
	d.pageData = result.Items
	d.totalCount = result.TotalCount
	return result.TotalCount
}

// GetExportData 현재 검색 조건에 맞는 전체 장비를 반환합니다.
func (d *DeviceTab) GetExportData() ([]*model.Firewall, error) {
	if d.searchKeyword == "" {
		return d.firewallRepo.GetAll()
	}
	req := model.PageRequest{
		Offset:  0,
		Limit:   100000,
		Keyword: d.searchKeyword,
	}
	result, err := d.firewallRepo.GetPage(req)
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

// IsFiltered 검색 필터가 적용되어 있는지 반환합니다.
func (d *DeviceTab) IsFiltered() bool {
	return d.searchKeyword != ""
}

// 탭의 컨텐츠를 반환합니다.
func (d *DeviceTab) Content() fyne.CanvasObject {
	return d.content
}

// 저장소에서 장비 목록을 로드합니다.
func (d *DeviceTab) loadFirewalls() {
	// 전체 장비 (상태 체크용)
	firewalls, err := d.firewallRepo.GetAll()
	if err != nil {
		dialog.ShowError(err, d.window)
		return
	}
	d.firewalls = firewalls

	// Index 내림차순 정렬 (최신순 - Index가 클수록 최신)
	sort.Slice(d.firewalls, func(i, j int) bool {
		return d.firewalls[i].Index > d.firewalls[j].Index
	})

	// 페이지 데이터 로드
	total := d.loadPage(0, d.deviceTable.GetPageSize())
	if d.deviceTable != nil {
		d.deviceTable.SetData(total)
	}

	// 상태 요약 업데이트
	d.updateStatusSummary()
}

// 장비 추가/수정 다이얼로그를 표시합니다. (PRD 3.3.3 기준)
func (d *DeviceTab) showAddEditDialog() {
	var editingFw *model.Firewall

	// 체크된 장비가 있으면 수정 모드 (PagedTable 사용)
	checkedRows := d.deviceTable.GetCheckedRows()

	// 여러 개 선택 시 안내 다이얼로그 표시
	if len(checkedRows) > 1 {
		dialog.ShowInformation("안내", "수정하려면 하나만 선택해주세요.", d.window)
		return
	}

	if len(checkedRows) == 1 && checkedRows[0] < len(d.pageData) {
		editingFw = d.pageData[checkedRows[0]]
	}

	// 입력 필드 너비
	entryWidth := float32(250)
	rowHeight := float32(32)
	labelWidth := float32(80)
	rowSpacing := float32(8) // 라인 간격

	deviceNameEntry := widget.NewEntry()
	deviceNameEntry.SetPlaceHolder("장비 이름")

	serverIPEntry := widget.NewEntry()
	serverIPEntry.SetPlaceHolder("192.168.1.1 또는 192.168.1.1:8080")

	locationEntry := widget.NewEntry()
	locationEntry.SetPlaceHolder("예: 서버실 1층")

	osInfoEntry := widget.NewEntry()
	osInfoEntry.SetPlaceHolder("예: Ubuntu 22.04")

	// 배포 경로 필드들
	pathEntryWidth := float32(150)
	portEntryWidth := float32(70)
	pathLabelWidth := float32(120)

	programUploadPathEntry := widget.NewEntry()
	programUploadPathEntry.SetPlaceHolder("/download/")

	// 패키지 업데이트 요청 스킴/경로/포트
	programUpdateSchemeSelect := widget.NewSelect([]string{"http", "https"}, nil)
	programUpdateSchemeSelect.SetSelected("http")
	programUpdatePathEntry := widget.NewEntry()
	programUpdatePathEntry.SetPlaceHolder("/program-update")
	programUpdatePortEntry := widget.NewEntry()
	programUpdatePortEntry.SetPlaceHolder("8080")

	// 방화벽 규칙 배포 스킴/경로/포트
	firewallDeploySchemeSelect := widget.NewSelect([]string{"http", "https"}, nil)
	firewallDeploySchemeSelect.SetSelected("http")
	firewallDeployPathEntry := widget.NewEntry()
	firewallDeployPathEntry.SetPlaceHolder("/agent/req-deploy")
	firewallDeployPortEntry := widget.NewEntry()
	firewallDeployPortEntry.SetPlaceHolder("8080")

	// 장비보고 스킴/경로/포트
	deviceInfoSchemeSelect := widget.NewSelect([]string{"http", "https"}, nil)
	deviceInfoSchemeSelect.SetSelected("http")
	deviceInfoPathEntry := widget.NewEntry()
	deviceInfoPathEntry.SetPlaceHolder("/device-report")
	deviceInfoPortEntry := widget.NewEntry()
	deviceInfoPortEntry.SetPlaceHolder("8080")

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

	// SSH ID 입력 영역 (PW/PPK 공통)
	sshIDEntryWidth := float32(160)
	sshIDContainer := container.NewHBox(
		container.NewGridWrap(fyne.NewSize(70, rowHeight), widget.NewLabel("SSH ID:")),
		container.NewGridWrap(fyne.NewSize(sshIDEntryWidth, rowHeight), sshIDEntry),
	)

	// PW 입력 영역 (비밀번호만)
	sshPWEntryWidth := float32(240)
	pwContainer := container.NewHBox(
		container.NewGridWrap(fyne.NewSize(70, rowHeight), widget.NewLabel("비밀번호:")),
		container.NewGridWrap(fyne.NewSize(sshPWEntryWidth, rowHeight), sshPWEntry),
	)

	// PPK 입력 영역 (초기에 숨김)
	ppkBrowseBtn := widget.NewButton("찾아보기...", nil)
	ppkContainer := container.NewVBox(
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(50, rowHeight), widget.NewLabel("PPK:")),
			ppkBrowseBtn,
		),
		ppkPathLabel,
	)
	ppkContainer.Hide()

	// 접속선택에 따른 동적 필드 전환
	authSelect.OnChanged = func(selected string) {
		if selected == "PW" {
			pwContainer.Show()
			ppkContainer.Hide()
		} else {
			sshPWEntry.SetText("")
			pwContainer.Hide()
			ppkContainer.Show()
		}
	}

	// 수정 모드면 기존 값 채우기, 아니면 전역 설정에서 기본값 가져오기
	if editingFw != nil {
		deviceNameEntry.SetText(editingFw.DeviceName)
		if editingFw.DeviceIP != "" {
			serverIPEntry.SetText(editingFw.DeviceIP)
		} else {
			serverIPEntry.SetText(editingFw.DeviceName) // 기존 데이터 호환
		}
		locationEntry.SetText(editingFw.Location)
		osInfoEntry.SetText(editingFw.OSInfo)
		// 배포 경로 필드
		programUploadPathEntry.SetText(editingFw.ProgramUploadPath)
		if editingFw.ProgramUpdateScheme != "" {
			programUpdateSchemeSelect.SetSelected(editingFw.ProgramUpdateScheme)
		}
		programUpdatePathEntry.SetText(editingFw.ProgramUpdatePath)
		if editingFw.ProgramUpdatePort > 0 {
			programUpdatePortEntry.SetText(fmt.Sprintf("%d", editingFw.ProgramUpdatePort))
		}
		if editingFw.FirewallDeployScheme != "" {
			firewallDeploySchemeSelect.SetSelected(editingFw.FirewallDeployScheme)
		}
		firewallDeployPathEntry.SetText(editingFw.FirewallDeployPath)
		if editingFw.FirewallDeployPort > 0 {
			firewallDeployPortEntry.SetText(fmt.Sprintf("%d", editingFw.FirewallDeployPort))
		}
		if editingFw.DeviceInfoScheme != "" {
			deviceInfoSchemeSelect.SetSelected(editingFw.DeviceInfoScheme)
		}
		deviceInfoPathEntry.SetText(editingFw.DeviceInfoPath)
		if editingFw.DeviceInfoPort > 0 {
			deviceInfoPortEntry.SetText(fmt.Sprintf("%d", editingFw.DeviceInfoPort))
		}
		// SSH 인증 정보
		sshIDEntry.SetText(editingFw.DeviceID)
		sshPWEntry.SetText(editingFw.DevicePW)
		if editingFw.DevicePPK != "" {
			ppkPath = editingFw.DevicePPK
			ppkPathLabel.SetText(editingFw.DevicePPK)
			authSelect.SetSelected("PPK")
		}
	} else {
		// 새 장비 추가 시 전역 설정에서 기본값 가져오기
		programUploadPathEntry.SetText(model.DefaultRemotePath) // 기본 업로드 경로
		config, err := d.configStore.GetConfig()
		if err == nil {
			programUpdateSchemeSelect.SetSelected(config.GetProgramUpdateScheme())
			programUpdatePathEntry.SetText(config.ProgramUpdatePath)
			if config.ProgramUpdatePort > 0 {
				programUpdatePortEntry.SetText(fmt.Sprintf("%d", config.ProgramUpdatePort))
			}
			firewallDeploySchemeSelect.SetSelected(config.GetFirewallDeployScheme())
			firewallDeployPathEntry.SetText(config.FirewallDeployPath)
			if config.FirewallDeployPort > 0 {
				firewallDeployPortEntry.SetText(fmt.Sprintf("%d", config.FirewallDeployPort))
			}
			deviceInfoSchemeSelect.SetSelected(config.GetDeviceInfoScheme())
			deviceInfoPathEntry.SetText(config.DeviceInfoPath)
			if config.DeviceInfoPort > 0 {
				deviceInfoPortEntry.SetText(fmt.Sprintf("%d", config.DeviceInfoPort))
			}
		}
	}

	// 다이얼로그 컨텐츠
	buttonSpacing := float32(30) // 버튼 윗쪽 간격

	// 헤더 (큰 폰트 - RichText 사용, 수정 모드에 따라 동적 변경)
	title := "장비 추가"
	if editingFw != nil {
		title = "장비 수정"
	}
	headerText := widget.NewRichTextFromMarkdown("## " + title)

	// 폼 컨텐츠 (이미지 순서대로 배치)
	formContent := container.NewVBox(
		// 장비명
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("장비명:")),
			container.NewGridWrap(fyne.NewSize(entryWidth, rowHeight), deviceNameEntry),
		),
		container.NewGridWrap(fyne.NewSize(1, rowSpacing), layout.NewSpacer()), // 간격
		// 장비 IP
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("장비 IP:")),
			container.NewGridWrap(fyne.NewSize(entryWidth, rowHeight), serverIPEntry),
		),
		container.NewGridWrap(fyne.NewSize(1, rowSpacing), layout.NewSpacer()), // 간격
		// 위치
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("위치:")),
			container.NewGridWrap(fyne.NewSize(entryWidth, rowHeight), locationEntry),
		),
		container.NewGridWrap(fyne.NewSize(1, rowSpacing), layout.NewSpacer()), // 간격
		// OS 정보
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("OS 정보:")),
			container.NewGridWrap(fyne.NewSize(entryWidth, rowHeight), osInfoEntry),
		),
		container.NewGridWrap(fyne.NewSize(1, rowSpacing), layout.NewSpacer()), // 간격
		// 접속선택
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("접속선택:")),
			container.NewGridWrap(fyne.NewSize(entryWidth, rowHeight), authSelect),
		),
		// SSH ID (PW/PPK 공통)
		sshIDContainer,
		container.NewGridWrap(fyne.NewSize(1, rowSpacing/2), layout.NewSpacer()), // 간격
		// PW 영역 (비밀번호)
		pwContainer,
		container.NewGridWrap(fyne.NewSize(1, rowSpacing/2), layout.NewSpacer()), // 간격
		// PPK 영역 (점선 테두리)
		ppkContainer,
		container.NewGridWrap(fyne.NewSize(1, rowSpacing), layout.NewSpacer()), // 간격
		widget.NewSeparator(),
		container.NewGridWrap(fyne.NewSize(1, rowSpacing), layout.NewSpacer()), // 간격
		// 업로드 경로 (SFTP)
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(pathLabelWidth, rowHeight), widget.NewLabel("업로드경로:")),
			container.NewGridWrap(fyne.NewSize(pathEntryWidth, rowHeight), programUploadPathEntry),
		),
		container.NewGridWrap(fyne.NewSize(1, rowSpacing/2), layout.NewSpacer()), // 간격
		// 패키지 업데이트 요청 (라벨 + 스킴)
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(pathLabelWidth, rowHeight), widget.NewLabel("패키지업데이트:")),
			container.NewGridWrap(fyne.NewSize(80, rowHeight), programUpdateSchemeSelect),
		),
		// 패키지 업데이트 요청 (경로 + 포트)
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(pathLabelWidth, rowHeight), layout.NewSpacer()),
			container.NewGridWrap(fyne.NewSize(pathEntryWidth, rowHeight), programUpdatePathEntry),
			widget.NewLabel("포트:"),
			container.NewGridWrap(fyne.NewSize(portEntryWidth, rowHeight), programUpdatePortEntry),
		),
		container.NewGridWrap(fyne.NewSize(1, rowSpacing/2), layout.NewSpacer()), // 간격
		// 방화벽 규칙 배포 (라벨 + 스킴)
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(pathLabelWidth, rowHeight), widget.NewLabel("방화벽 배포:")),
			container.NewGridWrap(fyne.NewSize(80, rowHeight), firewallDeploySchemeSelect),
		),
		// 방화벽 규칙 배포 (경로 + 포트)
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(pathLabelWidth, rowHeight), layout.NewSpacer()),
			container.NewGridWrap(fyne.NewSize(pathEntryWidth, rowHeight), firewallDeployPathEntry),
			widget.NewLabel("포트:"),
			container.NewGridWrap(fyne.NewSize(portEntryWidth, rowHeight), firewallDeployPortEntry),
		),
		container.NewGridWrap(fyne.NewSize(1, rowSpacing/2), layout.NewSpacer()), // 간격
		// 장비보고 (라벨 + 스킴)
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(pathLabelWidth, rowHeight), widget.NewLabel("장비보고:")),
			container.NewGridWrap(fyne.NewSize(80, rowHeight), deviceInfoSchemeSelect),
		),
		// 장비보고 (경로 + 포트)
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(pathLabelWidth, rowHeight), layout.NewSpacer()),
			container.NewGridWrap(fyne.NewSize(pathEntryWidth, rowHeight), deviceInfoPathEntry),
			widget.NewLabel("포트:"),
			container.NewGridWrap(fyne.NewSize(portEntryWidth, rowHeight), deviceInfoPortEntry),
		),
		container.NewGridWrap(fyne.NewSize(1, buttonSpacing), layout.NewSpacer()), // 버튼 간격
	)

	// 커스텀 팝업 생성
	var popup *widget.PopUp

	// 저장 처리 함수
	onSave := func() {
		// 유효성 검사
		if serverIPEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("장비 IP를 입력해주세요"), d.window)
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
		}

		fw.DeviceName = deviceNameEntry.Text
		fw.DeviceIP = serverIPEntry.Text
		fw.Location = locationEntry.Text
		fw.OSInfo = osInfoEntry.Text

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

		// API 경로 설정 저장
		fw.ProgramUploadPath = programUploadPathEntry.Text
		fw.ProgramUpdateScheme = programUpdateSchemeSelect.Selected
		fw.ProgramUpdatePath = programUpdatePathEntry.Text
		if programUpdatePortEntry.Text != "" {
			if port, err := strconv.Atoi(programUpdatePortEntry.Text); err == nil {
				fw.ProgramUpdatePort = port
			}
		}
		fw.FirewallDeployScheme = firewallDeploySchemeSelect.Selected
		fw.FirewallDeployPath = firewallDeployPathEntry.Text
		if firewallDeployPortEntry.Text != "" {
			if port, err := strconv.Atoi(firewallDeployPortEntry.Text); err == nil {
				fw.FirewallDeployPort = port
			}
		}
		fw.DeviceInfoScheme = deviceInfoSchemeSelect.Selected
		fw.DeviceInfoPath = deviceInfoPathEntry.Text
		if deviceInfoPortEntry.Text != "" {
			if port, err := strconv.Atoi(deviceInfoPortEntry.Text); err == nil {
				fw.DeviceInfoPort = port
			}
		}

		if err := d.firewallRepo.Save(fw); err != nil {
			dialog.ShowError(err, d.window)
			return
		}

		popup.Hide()

		// 체크 해제 및 새로고침
		d.deviceTable.ClearChecked()
		d.loadFirewalls()
		dialog.ShowInformation("성공", "장비 정보가 저장되었습니다.", d.window)
	}

	// 버튼 (간격 3배 = 60)
	cancelBtn := component.NewCustomButton("취소", nil, themes.Colors["black"], themes.Colors["lightgray"], func() {
		popup.Hide()
	}, 5, 5, 5, 5)
	saveBtn := component.NewCustomButton("저장", nil, nil, themes.Colors["blue"], onSave, 5, 5, 5, 5)

	// 버튼 컨테이너 (중앙 정렬, 버튼 간격 3배 = 60)
	btnContainer := container.NewHBox(
		layout.NewSpacer(),
		cancelBtn,
		container.NewGridWrap(fyne.NewSize(60, 1), layout.NewSpacer()), // 버튼 간격 3배
		saveBtn,
		layout.NewSpacer(),
	)

	// 전체 컨텐츠
	content := container.NewVBox(
		headerText,
		container.NewGridWrap(fyne.NewSize(1, rowSpacing), layout.NewSpacer()), // 헤더 아래 간격
		formContent,
		btnContainer,
		container.NewGridWrap(fyne.NewSize(1, 10), layout.NewSpacer()), // 하단 여백
	)

	// 고정 크기 컨테이너 (API 경로 필드 추가로 크기 확장)
	paddedContent := container.New(layout.NewCustomPaddedLayout(-5, 20, 20, 20), content)
	sizedContent := container.NewGridWrap(fyne.NewSize(480, 822), paddedContent)

	// 팝업 생성
	popup = widget.NewModalPopUp(sizedContent, d.window.Canvas())

	// PPK 찾아보기 버튼 - 다이얼로그 중첩 처리 (fyne-docs 스킬 참고)
	ppkBrowseBtn.OnTapped = func() {
		popup.Hide() // 부모 다이얼로그 숨김

		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if reader != nil {
				ppkPath = reader.URI().Path()
				ppkPathLabel.SetText(ppkPath)
				reader.Close()
			}
			popup.Show() // 부모 다이얼로그 다시 표시
		}, d.window)

		fileDialog.Show()
	}

	popup.Show()
}

// 상세보기 다이얼로그를 표시합니다. (PRD 3.3.3 기준)
func (d *DeviceTab) showDetailDialog(fw *model.Firewall) {
	// 레이아웃 설정
	rowHeight := float32(28)
	rowSpacing := float32(8)
	labelWidth := float32(100)
	valueWidth := float32(200)

	// 장비상태 텍스트
	statusText := "알수없음"
	switch fw.ServerStatus {
	case model.ServerStatusRunning:
		statusText = "정상"
	case model.ServerStatusStop:
		statusText = "연결안됨"
	}

	// 접속정보 타입 (PW or KEY)
	authType := "-"
	if fw.DevicePPK != "" {
		authType = "KEY"
	} else if fw.DevicePW != "" {
		authType = "PW"
	}

	// ID 값
	idValue := fw.DeviceID
	if idValue == "" {
		idValue = "-"
	}

	// PW 값 (마스킹)
	pwValue := "-"
	if fw.DevicePW != "" {
		pwValue = "********"
	}

	// PPK 경로
	ppkValue := fw.DevicePPK
	if ppkValue == "" {
		ppkValue = "-"
	}

	// IP 값
	ipValue := fw.DeviceIP
	if ipValue == "" {
		ipValue = fw.DeviceName
	}

	// 위치 값
	locationValue := fw.Location
	if locationValue == "" {
		locationValue = "-"
	}

	// OS 정보 값
	osInfoValue := fw.OSInfo
	if osInfoValue == "" {
		osInfoValue = "-"
	}

	// config 가져오기
	config, _ := d.configStore.GetConfig()

	// API 경로 값 (비어있으면 전역 설정 -> 기본값 순으로 표시)
	uploadPath := fw.ProgramUploadPath
	if uploadPath == "" {
		uploadPath = model.DefaultRemotePath
	}

	programUpdatePath := fw.ProgramUpdatePath
	if programUpdatePath == "" && config != nil {
		programUpdatePath = config.ProgramUpdatePath
	}
	if programUpdatePath == "" {
		programUpdatePath = model.DefaultProgramUpdatePath
	}
	programUpdatePort := fw.ProgramUpdatePort
	if programUpdatePort == 0 && config != nil {
		programUpdatePort = config.ProgramUpdatePort
	}
	if programUpdatePort == 0 {
		programUpdatePort = model.DefaultAPIPort
	}
	programUpdateScheme := fw.ProgramUpdateScheme
	if programUpdateScheme == "" && config != nil {
		programUpdateScheme = config.GetProgramUpdateScheme()
	}
	if programUpdateScheme == "" {
		programUpdateScheme = model.SchemeHTTP
	}

	firewallDeployPath := fw.FirewallDeployPath
	if firewallDeployPath == "" && config != nil {
		firewallDeployPath = config.FirewallDeployPath
	}
	if firewallDeployPath == "" {
		firewallDeployPath = model.DefaultFirewallDeployPath
	}
	firewallDeployPort := fw.FirewallDeployPort
	if firewallDeployPort == 0 && config != nil {
		firewallDeployPort = config.FirewallDeployPort
	}
	if firewallDeployPort == 0 {
		firewallDeployPort = model.DefaultAPIPort
	}
	firewallDeployScheme := fw.FirewallDeployScheme
	if firewallDeployScheme == "" && config != nil {
		firewallDeployScheme = config.GetFirewallDeployScheme()
	}
	if firewallDeployScheme == "" {
		firewallDeployScheme = model.SchemeHTTP
	}

	deviceInfoPath := fw.DeviceInfoPath
	if deviceInfoPath == "" && config != nil {
		deviceInfoPath = config.DeviceInfoPath
	}
	if deviceInfoPath == "" {
		deviceInfoPath = model.DefaultDeviceInfoPath
	}
	deviceInfoPort := fw.DeviceInfoPort
	if deviceInfoPort == 0 && config != nil {
		deviceInfoPort = config.DeviceInfoPort
	}
	if deviceInfoPort == 0 {
		deviceInfoPort = model.DefaultAPIPort
	}
	deviceInfoScheme := fw.DeviceInfoScheme
	if deviceInfoScheme == "" && config != nil {
		deviceInfoScheme = config.GetDeviceInfoScheme()
	}
	if deviceInfoScheme == "" {
		deviceInfoScheme = model.SchemeHTTP
	}

	// 배포정보 - 방화벽 룰 버전
	ruleVersion := fw.Version
	if ruleVersion == "" || ruleVersion == "-" {
		ruleVersion = "-"
	}

	// 커스텀 팝업 생성
	var popup *widget.PopUp

	// 헤더 (큰 폰트 - RichText 사용)
	headerText := widget.NewRichTextFromMarkdown("## 상세보기")

	// 폼 컨텐츠
	formContent := container.NewVBox(
		// 장비명
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("장비명:")),
			container.NewGridWrap(fyne.NewSize(valueWidth, rowHeight), widget.NewLabel(fw.DeviceName)),
		),
		// IP
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("IP:")),
			container.NewGridWrap(fyne.NewSize(valueWidth, rowHeight), widget.NewLabel(ipValue)),
		),
		// 위치
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("위치:")),
			container.NewGridWrap(fyne.NewSize(valueWidth, rowHeight), widget.NewLabel(locationValue)),
		),
		// OS 정보
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("OS 정보:")),
			container.NewGridWrap(fyne.NewSize(valueWidth, rowHeight), widget.NewLabel(osInfoValue)),
		),
		// 연결상태
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("연결상태:")),
			container.NewGridWrap(fyne.NewSize(valueWidth, rowHeight), widget.NewLabel(statusText)),
		),
		container.NewGridWrap(fyne.NewSize(1, rowSpacing), layout.NewSpacer()),
		// 접속정보
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("접속정보:")),
			container.NewGridWrap(fyne.NewSize(valueWidth, rowHeight), widget.NewLabel(authType)),
		),
		// ID
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("ID:")),
			container.NewGridWrap(fyne.NewSize(valueWidth, rowHeight), widget.NewLabel(idValue)),
		),
		// PW
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("PW:")),
			container.NewGridWrap(fyne.NewSize(valueWidth, rowHeight), widget.NewLabel(pwValue)),
		),
		// PPK 경로 (접속정보가 KEY인 경우)
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("PPK 경로:")),
			container.NewGridWrap(fyne.NewSize(valueWidth, rowHeight), widget.NewLabel(ppkValue)),
		),
		container.NewGridWrap(fyne.NewSize(1, rowSpacing), layout.NewSpacer()),
		// 업로드 경로
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("업로드 경로:")),
			container.NewGridWrap(fyne.NewSize(valueWidth, rowHeight), widget.NewLabel(uploadPath)),
		),
		// 패키지업데이트
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("패키지업데이트:")),
			container.NewGridWrap(fyne.NewSize(valueWidth, rowHeight), widget.NewLabel(fmt.Sprintf("%s %s, 포트:%d", programUpdateScheme, programUpdatePath, programUpdatePort))),
		),
		// 방화벽 배포
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("방화벽 배포:")),
			container.NewGridWrap(fyne.NewSize(valueWidth, rowHeight), widget.NewLabel(fmt.Sprintf("%s %s, 포트:%d", firewallDeployScheme, firewallDeployPath, firewallDeployPort))),
		),
		// 장비보고 (상태체크 포함)
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("장비보고:")),
			container.NewGridWrap(fyne.NewSize(valueWidth, rowHeight), widget.NewLabel(fmt.Sprintf("%s %s, 포트:%d", deviceInfoScheme, deviceInfoPath, deviceInfoPort))),
		),
		container.NewGridWrap(fyne.NewSize(1, rowSpacing), layout.NewSpacer()),
		// 배포정보
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("배포정보:")),
		),
		// 방화벽 룰 버전
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("  룰 버전:")),
			container.NewGridWrap(fyne.NewSize(valueWidth, rowHeight), widget.NewLabel(ruleVersion)),
		),
		container.NewGridWrap(fyne.NewSize(1, rowSpacing), layout.NewSpacer()),
	)

	// 확인 버튼
	confirmBtn := component.NewCustomButton("확인", nil, nil, themes.Colors["blue"], func() {
		popup.Hide()
	}, 5, 5, 5, 5)

	// 버튼 컨테이너 (중앙 정렬)
	btnContainer := container.NewHBox(
		layout.NewSpacer(),
		confirmBtn,
		layout.NewSpacer(),
	)

	// 전체 컨텐츠
	contentItems := []fyne.CanvasObject{
		headerText,
		container.NewGridWrap(fyne.NewSize(1, rowSpacing), layout.NewSpacer()),
		formContent,
	}
	contentItems = append(contentItems,
		btnContainer,
		container.NewGridWrap(fyne.NewSize(1, 15), layout.NewSpacer()),
	)
	content := container.NewVBox(contentItems...)

	// 다이얼로그 높이 계산 (프로세스 리스트 유무에 따라) - 임시 주석처리
	dialogHeight := float32(700)

	// 고정 크기 컨테이너
	paddedContent := container.New(layout.NewCustomPaddedLayout(15, 15, 15, 15), content)
	sizedContent := container.NewGridWrap(fyne.NewSize(400, dialogHeight), paddedContent)

	// 팝업 생성
	popup = widget.NewModalPopUp(sizedContent, d.window.Canvas())
	popup.Show()
}

// 배포 다이얼로그를 표시합니다. (PRD 3.3.3 기준 - 통합 다이얼로그)
func (d *DeviceTab) onDeploy() {
	// 체크된 장비 수집 (PagedTable 사용)
	checkedRows := d.deviceTable.GetCheckedRows()
	checkedFirewalls := []*model.Firewall{}
	checkedIPs := []string{}
	for _, row := range checkedRows {
		if row < len(d.pageData) {
			fw := d.pageData[row]
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

	// 레이아웃 설정 (다른 다이얼로그와 동일한 스타일)
	rowHeight := float32(36)
	rowSpacing := float32(20)
	labelWidth := float32(100)
	entryWidth := float32(300)
	buttonSpacing := float32(60)

	// 커스텀 팝업 생성
	var popup *widget.PopUp

	// 헤더 (큰 폰트 - RichText 사용)
	headerText := widget.NewRichTextFromMarkdown("## 배포")

	// 선택한 IP 리스트 (한 줄씩 표시, 3줄 높이 스크롤, 간격 축소)
	// canvas.Text 사용으로 패딩 없이 컴팩트하게 표시
	ipLabels := make([]fyne.CanvasObject, len(checkedIPs))
	for i, ip := range checkedIPs {
		text := canvas.NewText(ip, nil)
		text.TextSize = 14
		ipLabels[i] = text
	}
	ipListContent := container.NewVBox(ipLabels...)
	ipListScroll := container.NewScroll(ipListContent)
	ipListScroll.SetMinSize(fyne.NewSize(entryWidth, 54)) // 약 3줄 높이 (18px * 3)

	// 배포선택 드롭다운
	deployTypeSelect := widget.NewSelect([]string{"방화벽 룰 배포", "패키지 배포"}, nil)
	deployTypeSelect.SetSelected("방화벽 룰 배포")

	// 배포 리스트 (라디오 버튼 그룹)
	var selectedItem string

	// 방화벽 룰 파일 목록
	fileNames := d.firewallTab.GetFileNames()
	// 패키지 목록
	programs, _ := d.programRepo.GetAll()
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

	// 초기 방화벽 룰 리스트 (스크롤 가능)
	radioGroup := createRadioList(fileNames)
	deployListScroll := container.NewScroll(radioGroup)
	deployListScroll.SetMinSize(fyne.NewSize(entryWidth, 150))

	// 배포선택 변경 시 리스트 갱신
	deployTypeSelect.OnChanged = func(selected string) {
		selectedItem = ""
		var newRadio *widget.RadioGroup
		if selected == "방화벽 룰 배포" {
			newRadio = createRadioList(fileNames)
		} else {
			newRadio = createRadioList(programItems)
		}
		deployListScroll.Content = newRadio
		deployListScroll.Refresh()
	}

	// 폼 컨텐츠 (라인 간격 2배)
	formContent := container.NewVBox(
		// 선택한 IP 목록
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("IP 목록:")),
			ipListScroll,
		),
		container.NewGridWrap(fyne.NewSize(1, rowSpacing), layout.NewSpacer()),
		// 배포선택
		container.NewHBox(
			container.NewGridWrap(fyne.NewSize(labelWidth, rowHeight), widget.NewLabel("배포선택:")),
			container.NewGridWrap(fyne.NewSize(entryWidth, rowHeight), deployTypeSelect),
		),
		container.NewGridWrap(fyne.NewSize(1, rowSpacing), layout.NewSpacer()),
		widget.NewSeparator(),
		container.NewGridWrap(fyne.NewSize(1, rowSpacing), layout.NewSpacer()),
		// 배포 목록
		widget.NewLabelWithStyle("배포 목록:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		deployListScroll,
		container.NewGridWrap(fyne.NewSize(1, buttonSpacing), layout.NewSpacer()),
	)

	// 배포 처리 함수
	onDeploy := func() {
		if selectedItem == "" {
			dialog.ShowError(fmt.Errorf("배포할 항목을 선택해주세요"), d.window)
			return
		}

		popup.Hide()

		if deployTypeSelect.Selected == "방화벽 룰 배포" {
			d.executeFirewallDeploy(checkedFirewalls, selectedItem)
		} else {
			// 패키지 배포
			for _, p := range programs {
				displayName := fmt.Sprintf("%s %s", p.ProcessName, p.ProcessVersion)
				if displayName == selectedItem {
					d.executeProgramUpdate(checkedFirewalls, p)
					break
				}
			}
		}
	}

	// 버튼 (간격 3배 = 60)
	cancelBtn := component.NewCustomButton("취소", nil, themes.Colors["black"], themes.Colors["lightgray"], func() {
		popup.Hide()
	}, 5, 5, 5, 5)
	deployBtn := component.NewCustomButton("배포", nil, nil, themes.Colors["blue"], onDeploy, 5, 5, 5, 5)

	// 버튼 컨테이너 (중앙 정렬, 버튼 간격 3배 = 60)
	btnContainer := container.NewHBox(
		layout.NewSpacer(),
		cancelBtn,
		container.NewGridWrap(fyne.NewSize(60, 1), layout.NewSpacer()),
		deployBtn,
		layout.NewSpacer(),
	)

	// 전체 컨텐츠
	content := container.NewVBox(
		headerText,
		container.NewGridWrap(fyne.NewSize(1, rowSpacing), layout.NewSpacer()),
		formContent,
		btnContainer,
	)

	// 고정 크기 컨테이너 (상하 여백 동일하게)
	paddedContent := container.New(layout.NewCustomPaddedLayout(20, 20, 20, 20), content)
	sizedContent := container.NewGridWrap(fyne.NewSize(500, 600), paddedContent)

	// 팝업 생성
	popup = widget.NewModalPopUp(sizedContent, d.window.Canvas())
	popup.Show()
}

// 방화벽 룰 배포를 실행합니다.
func (d *DeviceTab) executeFirewallDeploy(firewalls []*model.Firewall, fileName string) {
	contents, err := d.firewallTab.GetFileContents(fileName)
	if err != nil {
		dialog.ShowError(fmt.Errorf("파일을 찾을 수 없습니다: %s", fileName), d.window)
		return
	}

	if contents == "" {
		dialog.ShowError(fmt.Errorf("선택한 파일에 내용이 없습니다"), d.window)
		return
	}

	// 배포용 RuleFile 구조체 생성
	ruleFile := &model.RuleFile{
		Version:  fileName,
		Contents: contents,
	}

	// 진행률 다이얼로그
	progressLabel := widget.NewLabel("배포 준비 중...")
	progressBar := widget.NewProgressBar()
	progressContent := container.NewVBox(progressLabel, progressBar)
	progressDialog := dialog.NewCustomWithoutButtons("배포 중", progressContent, d.window)
	progressDialog.Show()

	go func() {
		config, err := d.configStore.GetConfig()
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
		timeoutSeconds := config.GetTimeoutSeconds()

		for i, fw := range firewalls {
			ip := fw.DeviceIP
			if ip == "" {
				ip = fw.DeviceName
			}

			// 레이블 설정: 단일 장비면 카운트 생략
			var labelText string
			if total == 1 {
				labelText = fmt.Sprintf("배포 중: %s", ip)
			} else {
				labelText = fmt.Sprintf("배포 중: %s (%d/%d)", ip, i+1, total)
			}

			fyne.Do(func() {
				progressLabel.SetText(labelText)
				progressBar.SetValue(0)
			})

			// 타임아웃 기준 진행률 표시를 위한 타이머
			done := make(chan bool)
			go func() {
				ticker := time.NewTicker(500 * time.Millisecond)
				defer ticker.Stop()
				elapsed := 0
				for {
					select {
					case <-done:
						return
					case <-ticker.C:
						elapsed++
						progress := float64(elapsed) / float64(timeoutSeconds*2) // 0.5초 단위
						if progress > 0.95 {
							progress = 0.95 // 최대 95%까지만 표시
						}
						fyne.Do(func() {
							progressBar.SetValue(progress)
						})
					}
				}
			}()

			result := deployer.Deploy(fw, ruleFile)
			done <- true // 타이머 종료

			// 100% 표시 후 잠시 대기
			fyne.Do(func() {
				progressBar.SetValue(1.0)
			})
			time.Sleep(1 * time.Second)

			if result.Success {
				successCount++
			} else {
				failCount++
			}

			d.firewallRepo.Save(fw)

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

			// 결과 메시지 생성
			var resultTitle string
			var resultMsg string
			if failCount == 0 {
				resultTitle = "배포 완료"
				resultMsg = fmt.Sprintf("규칙 파일: %s\n성공: %d개", ruleFile.Version, successCount)
			} else if successCount == 0 {
				resultTitle = "배포 실패"
				resultMsg = fmt.Sprintf("규칙 파일: %s\n실패: %d개", ruleFile.Version, failCount)
			} else {
				resultTitle = "배포 일부 실패"
				resultMsg = fmt.Sprintf("규칙 파일: %s\n성공: %d개\n실패: %d개", ruleFile.Version, successCount, failCount)
			}
			dialog.ShowInformation(resultTitle, resultMsg, d.window)
		})
	}()
}

// 패키지 업데이트를 실행합니다.
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
	progressDialog := dialog.NewCustomWithoutButtons("업데이트 중", progressContent, d.window)
	progressDialog.Show()

	go func() {
		config, _ := d.configStore.GetConfig()
		timeoutSeconds := 30 // 기본값
		if config != nil {
			timeoutSeconds = config.GetTimeoutSeconds()
		}

		updateUC := usecase.NewUpdateUseCase(config)
		total := len(devices)
		successCount := 0
		failCount := 0

		for i, device := range devices {
			ip := device.DeviceIP
			if ip == "" {
				ip = device.DeviceName
			}

			// 레이블 설정: 단일 장비면 카운트 생략
			var labelText string
			if total == 1 {
				labelText = fmt.Sprintf("업데이트 중: %s", ip)
			} else {
				labelText = fmt.Sprintf("업데이트 중: %s (%d/%d)", ip, i+1, total)
			}

			fyne.Do(func() {
				progressLabel.SetText(labelText)
				progressBar.SetValue(0)
			})

			// 타임아웃 기준 진행률 표시를 위한 타이머
			done := make(chan bool)
			go func() {
				ticker := time.NewTicker(500 * time.Millisecond)
				defer ticker.Stop()
				elapsed := 0
				for {
					select {
					case <-done:
						return
					case <-ticker.C:
						elapsed++
						progress := float64(elapsed) / float64(timeoutSeconds*2)
						if progress > 0.95 {
							progress = 0.95
						}
						fyne.Do(func() {
							progressBar.SetValue(progress)
						})
					}
				}
			}()

			result := updateUC.UpdateProgram(device, program)
			done <- true

			// 100% 표시 후 잠시 대기
			fyne.Do(func() {
				progressBar.SetValue(1.0)
			})
			time.Sleep(1 * time.Second)

			if result.Success {
				successCount++
				d.firewallRepo.Save(device)
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

			// 결과 메시지 생성
			var resultTitle string
			var resultMsg string
			if failCount == 0 {
				resultTitle = "업데이트 완료"
				resultMsg = fmt.Sprintf("패키지: %s\n버전: %s\n성공: %d개", program.ProcessName, program.ProcessVersion, successCount)
			} else if successCount == 0 {
				resultTitle = "업데이트 실패"
				resultMsg = fmt.Sprintf("패키지: %s\n버전: %s\n실패: %d개", program.ProcessName, program.ProcessVersion, failCount)
			} else {
				resultTitle = "업데이트 일부 실패"
				resultMsg = fmt.Sprintf("패키지: %s\n버전: %s\n성공: %d개\n실패: %d개", program.ProcessName, program.ProcessVersion, successCount, failCount)
			}
			dialog.ShowInformation(resultTitle, resultMsg, d.window)
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
			if row < len(d.pageData) {
				fw := d.pageData[row]
				if fw.Index > 0 {
					d.firewallRepo.Delete(fw.Index)
				}
			}
		}

		// 삭제 후 새로고침
		d.deviceTable.ClearChecked()
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
		if row < len(d.pageData) {
			selectedFirewalls = append(selectedFirewalls, d.pageData[row])
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
		config, err := d.configStore.GetConfig()
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
			d.firewallRepo.Save(fw)
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

// 패키지 탭 참조를 설정합니다.
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

// 파일 목록만 새로고침합니다.
func (d *DeviceTab) RefreshFiles() {
	// 방화벽 관리 탭 새로고침
	d.firewallTab.RefreshFiles()
}

// 해당 IP의 장비 배포 상태를 초기화합니다.
func (d *DeviceTab) ResetDeviceDeployStatus(deviceIP string) {
	for _, fw := range d.firewalls {
		if fw.DeviceName == deviceIP || fw.DeviceIP == deviceIP {
			fw.DeployStatus = model.DeployStatusUnknown
			fw.Version = "-"
			d.firewallRepo.Save(fw)
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

// SetAutoStatusCheck 자동 상태 체크 기능을 설정합니다.
func (d *DeviceTab) SetAutoStatusCheck(enabled bool, intervalSeconds int) {
	// 기존 타이머 중지
	d.stopAutoStatusCheck()

	d.autoCheckEnabled = enabled
	d.autoCheckInterval = intervalSeconds

	// 상태 박스 테두리 색상 변경
	if d.statusBorder != nil {
		if enabled {
			// ON: orange 색상
			orangeColor := themes.Colors["orange"]
			d.statusBorder.StrokeColor = orangeColor
		} else {
			// OFF: 기본 회색
			grayColor := color.RGBA{R: 200, G: 200, B: 200, A: 255}
			d.statusBorder.StrokeColor = grayColor
		}
		d.statusBorder.Refresh()
	}

	if enabled && intervalSeconds > 0 {
		d.startAutoStatusCheck()
	}
}

// 자동 상태 체크를 시작합니다.
func (d *DeviceTab) startAutoStatusCheck() {
	log.Printf("[DEBUG] startAutoStatusCheck: 자동 상태 체크 시작 (주기: %d초)", d.autoCheckInterval)
	d.stopAutoCheck = make(chan struct{})
	d.autoCheckTicker = time.NewTicker(time.Duration(d.autoCheckInterval) * time.Second)

	go func() {
		// 시작 직후 바로 상태 체크 수행
		log.Printf("[DEBUG] startAutoStatusCheck: 초기 상태 체크 수행")
		d.performAutoStatusCheck()

		for {
			select {
			case <-d.stopAutoCheck:
				log.Printf("[DEBUG] startAutoStatusCheck: 자동 상태 체크 중지됨")
				return
			case <-d.autoCheckTicker.C:
				log.Printf("[DEBUG] startAutoStatusCheck: 타이머 틱 - performAutoStatusCheck 호출")
				d.performAutoStatusCheck()
			}
		}
	}()
}

// 자동 상태 체크를 중지합니다.
func (d *DeviceTab) stopAutoStatusCheck() {
	if d.autoCheckTicker != nil {
		d.autoCheckTicker.Stop()
		d.autoCheckTicker = nil
	}
	if d.stopAutoCheck != nil {
		close(d.stopAutoCheck)
		d.stopAutoCheck = nil
	}
}

// 모든 장비의 상태를 자동으로 체크합니다.
func (d *DeviceTab) performAutoStatusCheck() {
	log.Printf("[DEBUG] performAutoStatusCheck: 호출됨")

	if d.isRefreshing {
		log.Printf("[DEBUG] performAutoStatusCheck: isRefreshing=true, 스킵")
		return
	}

	// 모든 장비를 대상으로 상태 체크
	if len(d.firewalls) == 0 {
		log.Printf("[DEBUG] performAutoStatusCheck: 장비 목록 비어있음, 스킵")
		return
	}

	log.Printf("[DEBUG] performAutoStatusCheck: %d개 장비 상태 체크 시작", len(d.firewalls))
	d.isRefreshing = true

	go func() {
		config, err := d.configStore.GetConfig()
		if err != nil {
			log.Printf("[ERROR] performAutoStatusCheck: config 로드 실패 - %v", err)
			d.isRefreshing = false
			return
		}

		deployer := deploy.NewDeployer(config)
		deployer.HealthCheckBatch(d.firewalls)

		// 상태 확인 시간 업데이트
		now := time.Now().Format("2006-01-02 15:04:05")
		for _, fw := range d.firewalls {
			fw.LastCheckedAt = now
			d.firewallRepo.Save(fw)
		}

		fyne.Do(func() {
			d.applyFilter()
			d.updateStatusSummary()
			d.isRefreshing = false
		})
	}()
}

// 자동 상태 체크 토글 버튼 클릭 핸들러
func (d *DeviceTab) onToggleAutoCheck() {
	config, err := d.configStore.GetConfig()
	if err != nil {
		dialog.ShowError(err, d.window)
		return
	}

	// 현재 상태 토글
	newEnabled := !config.AutoStatusCheck
	config.AutoStatusCheck = newEnabled

	// config 저장
	if err := d.configStore.SaveConfig(config); err != nil {
		dialog.ShowError(err, d.window)
		return
	}

	// 버튼 아이콘 및 자동 체크 상태 업데이트
	d.updateAutoCheckButton(newEnabled)
	d.SetAutoStatusCheck(newEnabled, config.GetStatusCheckInterval())

	// 참고: Toast 대신 버튼 아이콘(Play/Pause)과 상태 박스 테두리 색상(회색/주황)으로
	// 상태 변경을 표시합니다. Toast 사용 시 fyne-tooltip과 ModalPopUp 간 충돌로
	// "no tool tip layer created for current overlay" 오류가 발생합니다.
}

// 자동 상태 체크 버튼 아이콘과 툴팁을 업데이트합니다.
func (d *DeviceTab) updateAutoCheckButton(enabled bool) {
	if d.autoCheckBtn == nil {
		return
	}

	if enabled {
		// ON 상태: Pause 아이콘 (||)
		d.autoCheckBtn.SetIcon(theme.MediaPauseIcon())
		d.autoCheckBtn.SetToolTip("자동체크중지")
	} else {
		// OFF 상태: Play 아이콘
		d.autoCheckBtn.SetIcon(theme.MediaPlayIcon())
		d.autoCheckBtn.SetToolTip("전체장비 자동체크")
	}
}

// 설정에서 자동 상태 체크 상태를 동기화합니다.
func (d *DeviceTab) SyncAutoCheckFromConfig() {
	config, err := d.configStore.GetConfig()
	if err != nil {
		return
	}

	d.updateAutoCheckButton(config.AutoStatusCheck)
	d.SetAutoStatusCheck(config.AutoStatusCheck, config.GetStatusCheckInterval())
}
