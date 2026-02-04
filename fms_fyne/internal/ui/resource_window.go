// Package ui는 사용자 인터페이스를 구현합니다.
package ui

import (
	"fmt"
	"image/color"
	"log"
	"sync"
	"time"

	"fms/internal/http"
	"fms/internal/model"
	"fms/internal/storage"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ResourceWindowManager는 장비별 리소스 모니터링 창을 관리합니다.
type ResourceWindowManager struct {
	app         fyne.App
	configStore *storage.JSONStore
	windows     map[int]*ResourceWindow // 장비 Index -> 창
	mu          sync.RWMutex
}

// NewResourceWindowManager는 새로운 리소스 창 관리자를 생성합니다.
func NewResourceWindowManager(app fyne.App, configStore *storage.JSONStore) *ResourceWindowManager {
	return &ResourceWindowManager{
		app:         app,
		configStore: configStore,
		windows:     make(map[int]*ResourceWindow),
	}
}

// OpenOrFocus는 장비의 리소스 창을 열거나 이미 열려있으면 포커스합니다.
func (m *ResourceWindowManager) OpenOrFocus(fw *model.Firewall, parentWindow fyne.Window) {
	m.mu.Lock()

	// 이미 열려있는 창이 있으면 포커스
	if w, ok := m.windows[fw.Index]; ok {
		m.mu.Unlock()
		w.window.RequestFocus()
		return
	}
	m.mu.Unlock()

	// 로딩 다이얼로그 표시
	ip := fw.DeviceIP
	if ip == "" {
		ip = fw.DeviceName
	}
	progressLabel := widget.NewLabel(fmt.Sprintf("리소스 정보 조회 중... (%s)", ip))
	progressBar := widget.NewProgressBarInfinite()
	progressContent := container.NewVBox(progressLabel, progressBar)
	progressDialog := dialog.NewCustomWithoutButtons("로딩 중", progressContent, parentWindow)
	progressDialog.Show()

	// UI 스레드에서 창 생성
	fyne.Do(func() {
		rw := NewResourceWindow(m.app, m.configStore, fw, func() {
			m.mu.Lock()
			delete(m.windows, fw.Index)
			m.mu.Unlock()
		})

		progressDialog.Hide()
		m.mu.Lock()
		m.windows[fw.Index] = rw
		m.mu.Unlock()
		rw.Show()

		// 백그라운드에서 데이터 로드 및 자동 갱신 시작
		go func() {
			fyne.Do(func() {
				rw.refresh()
			})
			rw.startAutoRefresh()
		}()
	})
}

// CloseAll은 모든 리소스 창을 닫습니다.
func (m *ResourceWindowManager) CloseAll() {
	// 데드락 방지: 먼저 창 목록을 복사한 후 잠금 해제
	m.mu.Lock()
	windowsToClose := make([]*ResourceWindow, 0, len(m.windows))
	for _, w := range m.windows {
		windowsToClose = append(windowsToClose, w)
	}
	m.windows = make(map[int]*ResourceWindow)
	m.mu.Unlock()

	// 잠금 해제 후 창 닫기 (SetOnClosed 콜백에서 mutex 접근 가능)
	for _, w := range windowsToClose {
		w.Close()
	}
}

// ResourceWindow는 단일 장비의 리소스 모니터링 창입니다.
type ResourceWindow struct {
	window      fyne.Window
	configStore *storage.JSONStore
	firewall    *model.Firewall
	onClose     func()

	// UI 컴포넌트
	lastUpdateLabel *widget.Label
	statusLabel     *widget.Label

	// 테이블 데이터
	cpuData     [][]string
	memoryData  [][]string
	diskData    [][]string
	networkData [][]string
	processData [][]string

	// 테이블 위젯
	cpuTable     *widget.Table
	memoryTable  *widget.Table
	diskTable    *widget.Table
	networkTable *widget.Table
	processTable *widget.Table

	// 자동 갱신
	ticker     *time.Ticker
	stopTicker chan struct{}

	// 컬럼 너비 비율 (동적 조절용)
	cpuColRatios     []float32
	memColRatios     []float32
	diskColRatios    []float32
	netColRatios     []float32
	procColRatios    []float32
}

// NewResourceWindow는 새로운 리소스 모니터링 창을 생성합니다.
// 주의: 이 함수는 반드시 fyne.Do() 안에서 호출해야 합니다.
func NewResourceWindow(app fyne.App, configStore *storage.JSONStore, fw *model.Firewall, onClose func()) *ResourceWindow {
	// 창 제목 설정
	ip := fw.DeviceIP
	if ip == "" {
		ip = fw.DeviceName
	}
	title := fmt.Sprintf("리소스 모니터링 - %s (%s)", fw.DeviceName, ip)

	rw := &ResourceWindow{
		window:      app.NewWindow(title),
		configStore: configStore,
		firewall:    fw,
		onClose:     onClose,
		stopTicker:  make(chan struct{}),
	}

	rw.createUI()
	rw.window.Resize(fyne.NewSize(1050, 800))
	rw.window.SetOnClosed(func() {
		rw.stopAutoRefresh()
		if rw.onClose != nil {
			rw.onClose()
		}
	})

	return rw
}

// createUI는 창의 UI를 생성합니다.
func (rw *ResourceWindow) createUI() {
	// 컬럼 비율 초기화 (기존 고정 너비 기반)
	rw.cpuColRatios = []float32{90, 90, 90, 90, 80, 90, 90, 80, 80, 80}     // 총 860
	rw.memColRatios = []float32{1, 1, 1, 1, 1, 1, 1, 1}                     // 균등 8개
	rw.diskColRatios = []float32{200, 130, 110, 110, 110, 110}             // 총 770
	rw.netColRatios = []float32{120, 160, 110, 110, 110, 110, 100}         // 총 820
	rw.procColRatios = []float32{150, 130, 90, 90, 110, 100, 110}          // 총 780

	// 상단 헤더
	rw.lastUpdateLabel = widget.NewLabel("마지막 갱신: -")
	rw.statusLabel = widget.NewLabel("")

	refreshBtn := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		rw.refresh()
	})

	header := container.NewHBox(
		rw.lastUpdateLabel,
		layout.NewSpacer(),
		rw.statusLabel,
		refreshBtn,
	)

	// CPU 테이블
	cpuSection := rw.createCPUSection()

	// 메모리 테이블
	memorySection := rw.createMemorySection()

	// 디스크 테이블
	diskSection := rw.createDiskSection()

	// 네트워크 테이블
	networkSection := rw.createNetworkSection()

	// 프로세스 테이블
	processSection := rw.createProcessSection()

	// 5개 섹션이 가용 공간을 균등하게 나눠 갖도록 GridWithRows 사용
	content := container.NewGridWithRows(5,
		cpuSection,
		memorySection,
		diskSection,
		networkSection,
		processSection,
	)

	// 전체 레이아웃
	mainContent := container.NewBorder(
		container.NewVBox(header, widget.NewSeparator()),
		nil, nil, nil,
		content,
	)

	// 크기 변경 감지 래퍼로 감싸서 컬럼 너비 동적 조절
	resizable := newResourceResizableContainer(container.NewPadded(mainContent), func(size fyne.Size) {
		// 패딩(16*2) + Card 여백(24) 제외한 가용 너비
		availableWidth := size.Width - 56
		if availableWidth > 200 {
			rw.updateColumnWidths(availableWidth)
		}
	})

	rw.window.SetContent(resizable)
}

// updateColumnWidths는 가용 너비에 맞게 모든 테이블의 컬럼 너비를 조절합니다.
func (rw *ResourceWindow) updateColumnWidths(availableWidth float32) {
	// CPU 테이블
	if rw.cpuTable != nil {
		total := sumFloat32(rw.cpuColRatios)
		for i, ratio := range rw.cpuColRatios {
			rw.cpuTable.SetColumnWidth(i, availableWidth*ratio/total)
		}
	}

	// 메모리 테이블 (균등 분배)
	if rw.memoryTable != nil {
		colWidth := availableWidth / float32(len(rw.memColRatios))
		for i := range rw.memColRatios {
			rw.memoryTable.SetColumnWidth(i, colWidth)
		}
	}

	// 디스크 테이블
	if rw.diskTable != nil {
		total := sumFloat32(rw.diskColRatios)
		for i, ratio := range rw.diskColRatios {
			rw.diskTable.SetColumnWidth(i, availableWidth*ratio/total)
		}
	}

	// 네트워크 테이블
	if rw.networkTable != nil {
		total := sumFloat32(rw.netColRatios)
		for i, ratio := range rw.netColRatios {
			rw.networkTable.SetColumnWidth(i, availableWidth*ratio/total)
		}
	}

	// 프로세스 테이블
	if rw.processTable != nil {
		total := sumFloat32(rw.procColRatios)
		for i, ratio := range rw.procColRatios {
			rw.processTable.SetColumnWidth(i, availableWidth*ratio/total)
		}
	}
}

// sumFloat32는 float32 슬라이스의 합을 반환합니다.
func sumFloat32(values []float32) float32 {
	var sum float32
	for _, v := range values {
		sum += v
	}
	return sum
}

// resourceResizableContainer는 크기 변경 감지 컨테이너입니다.
type resourceResizableContainer struct {
	widget.BaseWidget
	wrapped  fyne.CanvasObject
	onResize func(fyne.Size)
}

func newResourceResizableContainer(wrapped fyne.CanvasObject, onResize func(fyne.Size)) *resourceResizableContainer {
	r := &resourceResizableContainer{wrapped: wrapped, onResize: onResize}
	r.ExtendBaseWidget(r)
	return r
}

func (r *resourceResizableContainer) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(r.wrapped)
}

func (r *resourceResizableContainer) Resize(size fyne.Size) {
	r.BaseWidget.Resize(size)
	r.wrapped.Resize(size)
	if r.onResize != nil {
		r.onResize(size)
	}
}

// configureReadOnlyTable은 읽기 전용 테이블 설정을 적용합니다.
// - 헤더 행 고정 (StickyRowCount = 1)
// - 셀 선택 비활성화 (OnSelected에서 UnselectAll 호출)
func configureReadOnlyTable(table *widget.Table) {
	table.StickyRowCount = 1
	table.OnSelected = func(id widget.TableCellID) {
		table.UnselectAll()
	}
}

// createCPUSection은 CPU 섹션을 생성합니다.
func (rw *ResourceWindow) createCPUSection() fyne.CanvasObject {
	// CPU 테이블 (헤더 포함)
	cpuHeaders := []string{"User", "Kernel", "Idle", "IOWait", "Nice", "HW IRQ", "SW IRQ", "Avg1", "Avg5", "Avg15"}
	rw.cpuData = [][]string{make([]string, len(cpuHeaders))}
	headerBgColor := color.RGBA{R: 230, G: 230, B: 230, A: 255}

	rw.cpuTable = widget.NewTable(
		func() (int, int) { return len(rw.cpuData) + 1, len(cpuHeaders) }, // +1 for header
		func() fyne.CanvasObject {
			bg := canvas.NewRectangle(color.Transparent)
			label := widget.NewLabel("")
			label.Alignment = fyne.TextAlignCenter
			return container.NewStack(bg, container.NewCenter(label))
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			stack := obj.(*fyne.Container)
			bg := stack.Objects[0].(*canvas.Rectangle)
			label := stack.Objects[1].(*fyne.Container).Objects[0].(*widget.Label)
			if id.Row == 0 {
				// 헤더 행
				bg.FillColor = headerBgColor
				label.SetText(cpuHeaders[id.Col])
			} else {
				bg.FillColor = color.Transparent
				if id.Row-1 < len(rw.cpuData) && id.Col < len(rw.cpuData[id.Row-1]) {
					label.SetText(rw.cpuData[id.Row-1][id.Col])
				}
			}
			bg.Refresh()
		},
	)

	rw.cpuTable.SetColumnWidth(0, 90)
	rw.cpuTable.SetColumnWidth(1, 90)
	rw.cpuTable.SetColumnWidth(2, 90)
	rw.cpuTable.SetColumnWidth(3, 90)
	rw.cpuTable.SetColumnWidth(4, 80)
	rw.cpuTable.SetColumnWidth(5, 90)
	rw.cpuTable.SetColumnWidth(6, 90)
	rw.cpuTable.SetColumnWidth(7, 80)
	rw.cpuTable.SetColumnWidth(8, 80)
	rw.cpuTable.SetColumnWidth(9, 80)
	configureReadOnlyTable(rw.cpuTable)

	// 최소 높이 설정 (테이블이 가용 너비를 채우도록)
	minSizer := canvas.NewRectangle(color.Transparent)
	minSizer.SetMinSize(fyne.NewSize(0, 80))
	tableWithMinHeight := container.NewStack(minSizer, rw.cpuTable)
	tableWithBorder := container.NewVBox(widget.NewSeparator(), tableWithMinHeight, widget.NewSeparator())

	return widget.NewCard("CPU", "", tableWithBorder)
}

// createMemorySection은 메모리 섹션을 생성합니다.
func (rw *ResourceWindow) createMemorySection() fyne.CanvasObject {
	memHeaders := []string{"Total", "Used", "Free", "Available", "Buff/Cache", "Swap Total", "Swap Used", "Swap Free"}
	rw.memoryData = [][]string{make([]string, len(memHeaders))}
	headerBgColor := color.RGBA{R: 230, G: 230, B: 230, A: 255}

	rw.memoryTable = widget.NewTable(
		func() (int, int) { return len(rw.memoryData) + 1, len(memHeaders) }, // +1 for header
		func() fyne.CanvasObject {
			bg := canvas.NewRectangle(color.Transparent)
			label := widget.NewLabel("")
			label.Alignment = fyne.TextAlignCenter
			return container.NewStack(bg, container.NewCenter(label))
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			stack := obj.(*fyne.Container)
			bg := stack.Objects[0].(*canvas.Rectangle)
			label := stack.Objects[1].(*fyne.Container).Objects[0].(*widget.Label)
			if id.Row == 0 {
				bg.FillColor = headerBgColor
				label.SetText(memHeaders[id.Col])
			} else {
				bg.FillColor = color.Transparent
				if id.Row-1 < len(rw.memoryData) && id.Col < len(rw.memoryData[id.Row-1]) {
					label.SetText(rw.memoryData[id.Row-1][id.Col])
				}
			}
			bg.Refresh()
		},
	)

	for i := 0; i < len(memHeaders); i++ {
		rw.memoryTable.SetColumnWidth(i, 115)
	}
	configureReadOnlyTable(rw.memoryTable)

	// 최소 높이 설정 (테이블이 가용 너비를 채우도록)
	minSizer := canvas.NewRectangle(color.Transparent)
	minSizer.SetMinSize(fyne.NewSize(0, 80))
	tableWithMinHeight := container.NewStack(minSizer, rw.memoryTable)
	tableWithBorder := container.NewVBox(widget.NewSeparator(), tableWithMinHeight, widget.NewSeparator())

	return widget.NewCard("메모리", "", tableWithBorder)
}

// createDiskSection은 디스크 섹션을 생성합니다.
func (rw *ResourceWindow) createDiskSection() fyne.CanvasObject {
	diskHeaders := []string{"파티션", "마운트", "전체", "사용", "여유", "사용률"}
	rw.diskData = [][]string{}
	headerBgColor := color.RGBA{R: 230, G: 230, B: 230, A: 255}

	rw.diskTable = widget.NewTable(
		func() (int, int) {
			rows := len(rw.diskData)
			if rows == 0 {
				rows = 1
			}
			return rows + 1, len(diskHeaders) // +1 for header
		},
		func() fyne.CanvasObject {
			bg := canvas.NewRectangle(color.Transparent)
			label := widget.NewLabel("")
			label.Alignment = fyne.TextAlignCenter
			return container.NewStack(bg, container.NewCenter(label))
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			stack := obj.(*fyne.Container)
			bg := stack.Objects[0].(*canvas.Rectangle)
			label := stack.Objects[1].(*fyne.Container).Objects[0].(*widget.Label)
			if id.Row == 0 {
				bg.FillColor = headerBgColor
				label.SetText(diskHeaders[id.Col])
			} else {
				bg.FillColor = color.Transparent
				dataRow := id.Row - 1
				if dataRow < len(rw.diskData) && id.Col < len(rw.diskData[dataRow]) {
					label.SetText(rw.diskData[dataRow][id.Col])
				} else {
					label.SetText("-")
				}
			}
			bg.Refresh()
		},
	)

	rw.diskTable.SetColumnWidth(0, 200)
	rw.diskTable.SetColumnWidth(1, 130)
	rw.diskTable.SetColumnWidth(2, 110)
	rw.diskTable.SetColumnWidth(3, 110)
	rw.diskTable.SetColumnWidth(4, 110)
	rw.diskTable.SetColumnWidth(5, 110)
	configureReadOnlyTable(rw.diskTable)

	// 최소 높이 설정 (테이블이 가용 너비를 채우도록)
	minSizer := canvas.NewRectangle(color.Transparent)
	minSizer.SetMinSize(fyne.NewSize(0, 100))
	tableWithMinHeight := container.NewStack(minSizer, rw.diskTable)
	tableWithBorder := container.NewVBox(widget.NewSeparator(), tableWithMinHeight, widget.NewSeparator())

	return widget.NewCard("디스크", "", tableWithBorder)
}

// createNetworkSection은 네트워크 섹션을 생성합니다.
func (rw *ResourceWindow) createNetworkSection() fyne.CanvasObject {
	netHeaders := []string{"인터페이스", "IP", "RX 속도", "TX 속도", "RX 누적", "TX 누적", "에러"}
	rw.networkData = [][]string{}
	headerBgColor := color.RGBA{R: 230, G: 230, B: 230, A: 255}

	rw.networkTable = widget.NewTable(
		func() (int, int) {
			rows := len(rw.networkData)
			if rows == 0 {
				rows = 1
			}
			return rows + 1, len(netHeaders) // +1 for header
		},
		func() fyne.CanvasObject {
			bg := canvas.NewRectangle(color.Transparent)
			label := widget.NewLabel("")
			label.Alignment = fyne.TextAlignCenter
			return container.NewStack(bg, container.NewCenter(label))
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			stack := obj.(*fyne.Container)
			bg := stack.Objects[0].(*canvas.Rectangle)
			label := stack.Objects[1].(*fyne.Container).Objects[0].(*widget.Label)
			if id.Row == 0 {
				bg.FillColor = headerBgColor
				label.SetText(netHeaders[id.Col])
			} else {
				bg.FillColor = color.Transparent
				dataRow := id.Row - 1
				if dataRow < len(rw.networkData) && id.Col < len(rw.networkData[dataRow]) {
					label.SetText(rw.networkData[dataRow][id.Col])
				} else {
					label.SetText("-")
				}
			}
			bg.Refresh()
		},
	)

	rw.networkTable.SetColumnWidth(0, 120)
	rw.networkTable.SetColumnWidth(1, 160)
	rw.networkTable.SetColumnWidth(2, 110)
	rw.networkTable.SetColumnWidth(3, 110)
	rw.networkTable.SetColumnWidth(4, 110)
	rw.networkTable.SetColumnWidth(5, 110)
	rw.networkTable.SetColumnWidth(6, 100)
	configureReadOnlyTable(rw.networkTable)

	// 최소 높이 설정 (테이블이 가용 너비를 채우도록)
	minSizer := canvas.NewRectangle(color.Transparent)
	minSizer.SetMinSize(fyne.NewSize(0, 100))
	tableWithMinHeight := container.NewStack(minSizer, rw.networkTable)
	tableWithBorder := container.NewVBox(widget.NewSeparator(), tableWithMinHeight, widget.NewSeparator())

	return widget.NewCard("네트워크", "", tableWithBorder)
}

// createProcessSection은 프로세스 섹션을 생성합니다.
func (rw *ResourceWindow) createProcessSection() fyne.CanvasObject {
	procHeaders := []string{"이름", "버전", "CPU%", "MEM%", "메모리", "상태", "실행시간"}
	rw.processData = [][]string{}
	headerBgColor := color.RGBA{R: 230, G: 230, B: 230, A: 255}

	rw.processTable = widget.NewTable(
		func() (int, int) {
			rows := len(rw.processData)
			if rows == 0 {
				rows = 1
			}
			return rows + 1, len(procHeaders) // +1 for header
		},
		func() fyne.CanvasObject {
			bg := canvas.NewRectangle(color.Transparent)
			label := widget.NewLabel("")
			label.Alignment = fyne.TextAlignCenter
			return container.NewStack(bg, container.NewCenter(label))
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			stack := obj.(*fyne.Container)
			bg := stack.Objects[0].(*canvas.Rectangle)
			label := stack.Objects[1].(*fyne.Container).Objects[0].(*widget.Label)
			if id.Row == 0 {
				bg.FillColor = headerBgColor
				label.SetText(procHeaders[id.Col])
			} else {
				bg.FillColor = color.Transparent
				dataRow := id.Row - 1
				if dataRow < len(rw.processData) && id.Col < len(rw.processData[dataRow]) {
					label.SetText(rw.processData[dataRow][id.Col])
				} else {
					label.SetText("-")
				}
			}
			bg.Refresh()
		},
	)

	rw.processTable.SetColumnWidth(0, 150)
	rw.processTable.SetColumnWidth(1, 130)
	rw.processTable.SetColumnWidth(2, 90)
	rw.processTable.SetColumnWidth(3, 90)
	rw.processTable.SetColumnWidth(4, 110)
	rw.processTable.SetColumnWidth(5, 100)
	rw.processTable.SetColumnWidth(6, 110)
	configureReadOnlyTable(rw.processTable)

	// 최소 높이 설정 (테이블이 가용 너비를 채우도록)
	minSizer := canvas.NewRectangle(color.Transparent)
	minSizer.SetMinSize(fyne.NewSize(0, 150))
	tableWithMinHeight := container.NewStack(minSizer, rw.processTable)
	tableWithBorder := container.NewVBox(widget.NewSeparator(), tableWithMinHeight, widget.NewSeparator())

	return widget.NewCard("프로세스", "", tableWithBorder)
}

// refresh는 데이터를 새로고침합니다.
func (rw *ResourceWindow) refresh() {
	config, err := rw.configStore.GetConfig()
	if err != nil {
		log.Printf("[ERROR] ResourceWindow.refresh: config 로드 실패 - %v", err)
		rw.statusLabel.SetText("설정 로드 실패")
		return
	}

	client := http.NewClient(config)
	report, err := client.GetDeviceReport(rw.firewall)
	if err != nil {
		log.Printf("[ERROR] ResourceWindow.refresh: 데이터 조회 실패 - %v", err)
		rw.statusLabel.SetText("연결 안됨")
		return
	}

	rw.statusLabel.SetText("")
	rw.updateData(report)
	rw.lastUpdateLabel.SetText(fmt.Sprintf("마지막 갱신: %s", time.Now().Format("15:04:05")))
}

// updateData는 테이블 데이터를 업데이트합니다.
func (rw *ResourceWindow) updateData(report *model.DeviceReport) {
	// CPU 데이터
	cpu := report.CPUAll
	rw.cpuData = [][]string{{
		fmt.Sprintf("%.1f%%", cpu.User),
		fmt.Sprintf("%.1f%%", cpu.Kernel),
		fmt.Sprintf("%.1f%%", cpu.Idle),
		fmt.Sprintf("%.1f%%", cpu.IOWait),
		fmt.Sprintf("%.1f%%", cpu.Ni),
		fmt.Sprintf("%.1f%%", cpu.HwIrq),
		fmt.Sprintf("%.1f%%", cpu.SwIrq),
		fmt.Sprintf("%.2f", cpu.Avg1),
		fmt.Sprintf("%.2f", cpu.Avg5),
		fmt.Sprintf("%.2f", cpu.Avg15),
	}}
	rw.cpuTable.Refresh()

	// 메모리 데이터
	mem := report.Memory
	rw.memoryData = [][]string{{
		mem.TotalDisplay,
		mem.UsedDisplay,
		mem.FreeDisplay,
		mem.AvailableDisplay,
		mem.BuffCacheDisplay,
		mem.SwapTotalDisplay,
		mem.SwapUsedDisplay,
		mem.SwapFreeDisplay,
	}}
	rw.memoryTable.Refresh()

	// 디스크 데이터
	rw.diskData = make([][]string, len(report.Disk))
	for i, d := range report.Disk {
		rw.diskData[i] = []string{
			d.Filesystem,
			d.MountedOn,
			d.SizeDisplay,
			d.UsedDisplay,
			d.FreeDisplay,
			fmt.Sprintf("%.1f%%", d.UsePercent),
		}
	}
	if len(rw.diskData) == 0 {
		rw.diskData = [][]string{{"-", "-", "-", "-", "-", "-"}}
	}
	rw.diskTable.Refresh()

	// 네트워크 데이터
	rw.networkData = make([][]string, len(report.Interface))
	for i, n := range report.Interface {
		errors := n.RxErrors + n.TxErrors + n.RxDropped + n.TxDropped
		rw.networkData[i] = []string{
			n.IfName,
			n.IP4,
			n.RxBpsDisplay,
			n.TxBpsDisplay,
			n.RxBytesDisplay,
			n.TxBytesDisplay,
			fmt.Sprintf("%d", errors),
		}
	}
	if len(rw.networkData) == 0 {
		rw.networkData = [][]string{{"-", "-", "-", "-", "-", "-", "-"}}
	}
	rw.networkTable.Refresh()

	// 프로세스 데이터 (name + version 있는 것만)
	rw.processData = [][]string{}
	for _, p := range report.Process {
		if p.Name != "" && p.Version != "" {
			rw.processData = append(rw.processData, []string{
				p.Name,
				p.Version,
				fmt.Sprintf("%.1f%%", p.CPUPercent),
				fmt.Sprintf("%.1f%%", p.MemPercent),
				p.RSSDisplay,
				model.GetProcessStatusText(p.Status),
				p.RunTime,
			})
		}
	}
	if len(rw.processData) == 0 {
		rw.processData = [][]string{{"-", "-", "-", "-", "-", "-", "-"}}
	}
	rw.processTable.Refresh()
}

// startAutoRefresh는 자동 갱신을 시작합니다.
func (rw *ResourceWindow) startAutoRefresh() {
	config, err := rw.configStore.GetConfig()
	if err != nil {
		log.Printf("[ERROR] ResourceWindow.startAutoRefresh: config 로드 실패 - %v", err)
		return
	}

	interval := time.Duration(config.GetStatusCheckInterval()) * time.Second
	rw.ticker = time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-rw.stopTicker:
				return
			case <-rw.ticker.C:
				fyne.Do(func() {
					rw.refresh()
				})
			}
		}
	}()

	log.Printf("[DEBUG] ResourceWindow: 자동 갱신 시작 (주기: %v)", interval)
}

// stopAutoRefresh는 자동 갱신을 중지합니다.
func (rw *ResourceWindow) stopAutoRefresh() {
	if rw.ticker != nil {
		rw.ticker.Stop()
	}
	close(rw.stopTicker)
	log.Printf("[DEBUG] ResourceWindow: 자동 갱신 중지")
}

// Show는 창을 표시합니다.
func (rw *ResourceWindow) Show() {
	rw.window.Show()
}

// Close는 창을 닫습니다.
func (rw *ResourceWindow) Close() {
	rw.window.Close()
}
