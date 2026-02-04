// Package model은 애플리케이션의 데이터 모델을 정의합니다.
package model

// DeviceReport는 장비 리포트 API 응답을 나타냅니다.
type DeviceReport struct {
	TotalUsed ReportTotalUsed   `json:"totalUsed"`
	CPUAll    ReportCPUAll      `json:"cpuall"`
	Memory    ReportMemory      `json:"memory"`
	Disk      []ReportDisk      `json:"disk"`
	Interface []ReportInterface `json:"interface"`
	Process   []ReportProcess   `json:"process"`
}

// ReportTotalUsed는 리소스 사용 요약 정보입니다.
type ReportTotalUsed struct {
	CPU          float64 `json:"cpu"`
	Memory       float64 `json:"memory"`
	Disk         float64 `json:"disk"`
	RxBps        float64 `json:"rxBps"`
	TxBps        float64 `json:"txBps"`
	RxBpsDisplay string  `json:"rxBpsDisplay"`
	TxBpsDisplay string  `json:"txBpsDisplay"`
	RxPps        int64   `json:"rxPps"`
	TxPps        int64   `json:"txPps"`
}

// ReportCPUAll은 CPU 상세 정보입니다.
type ReportCPUAll struct {
	Avg1   float64 `json:"avg1"`
	Avg5   float64 `json:"avg5"`
	Avg15  float64 `json:"avg15"`
	User   float64 `json:"user"`
	Kernel float64 `json:"kernel"`
	Idle   float64 `json:"idle"`
	Ni     float64 `json:"ni"`
	IOWait float64 `json:"iowait"`
	HwIrq  float64 `json:"hwIrq"`
	SwIrq  float64 `json:"swIrq"`
	Steal  float64 `json:"steal"`
	Usage  float64 `json:"usage"`
	Total  float64 `json:"total"`
}

// ReportMemory는 메모리 상세 정보입니다.
type ReportMemory struct {
	Total            int64   `json:"total"`
	Free             int64   `json:"free"`
	Used             int64   `json:"used"`
	Available        int64   `json:"available"`
	TotalDisplay     string  `json:"totalDisplay"`
	FreeDisplay      string  `json:"freeDisplay"`
	UsedDisplay      string  `json:"usedDisplay"`
	AvailableDisplay string  `json:"availableDisplay"`
	UsedPercent      float64 `json:"usedPercent"`
	BuffCache        int64   `json:"buffCache"`
	SwapTotal        int64   `json:"swapTotal"`
	SwapFree         int64   `json:"swapFree"`
	SwapUsed         int64   `json:"swapUsed"`
	BuffCacheDisplay string  `json:"buffCacheDisplay"`
	SwapTotalDisplay string  `json:"swapTotalDisplay"`
	SwapFreeDisplay  string  `json:"swapFreeDisplay"`
	SwapUsedDisplay  string  `json:"swapUsedDisplay"`
}

// ReportDisk는 디스크 파티션 정보입니다.
type ReportDisk struct {
	Filesystem  string  `json:"filesystem"`
	Size        int64   `json:"size"`
	Used        int64   `json:"used"`
	Avail       int64   `json:"avail"`
	Free        int64   `json:"free"`
	SizeDisplay string  `json:"sizeDisplay"`
	UsedDisplay string  `json:"usedDisplay"`
	FreeDisplay string  `json:"freeDisplay"`
	UsePercent  float64 `json:"usePercent"`
	MountedOn   string  `json:"mountedOn"`
}

// ReportInterface는 네트워크 인터페이스 정보입니다.
type ReportInterface struct {
	IfName         string  `json:"ifname"`
	IP4            string  `json:"ip4"`
	IP6            string  `json:"ip6"`
	RxPackets      int64   `json:"rxPackets"`
	RxBytes        int64   `json:"rxBytes"`
	RxBytesDisplay string  `json:"rxBytesDisplay"`
	RxErrors       int64   `json:"rxErrors"`
	RxDropped      int64   `json:"rxDropped"`
	TxPackets      int64   `json:"txPackets"`
	TxBytes        int64   `json:"txBytes"`
	TxBytesDisplay string  `json:"txBytesDisplay"`
	TxErrors       int64   `json:"txErrors"`
	TxDropped      int64   `json:"txDropped"`
	RxBps          float64 `json:"rxBps"`
	TxBps          float64 `json:"txBps"`
	RxBpsDisplay   string  `json:"rxBpsDisplay"`
	TxBpsDisplay   string  `json:"txBpsDisplay"`
	RxPps          int64   `json:"rxPps"`
	TxPps          int64   `json:"txPps"`
}

// ReportProcess는 프로세스 정보입니다.
type ReportProcess struct {
	Name       string   `json:"name,omitempty"`
	Version    string   `json:"version,omitempty"`
	User       string   `json:"user"`
	PID        int      `json:"pid"`
	CPUPercent float64  `json:"cpuPercent"`
	MemPercent float64  `json:"memPercent"`
	VSZ        int64    `json:"vsz"`
	RSS        int64    `json:"rss"`
	VSZDisplay string   `json:"vszDisplay"`
	RSSDisplay string   `json:"rssDisplay"`
	Start      string   `json:"start"`
	RunTime    string   `json:"runTime"`
	Status     []string `json:"status"`
	Cmd        string   `json:"cmd"`
	TTY        string   `json:"tty"`
}

// ProcessStatusMap은 프로세스 상태 코드를 한글로 매핑합니다.
var ProcessStatusMap = map[string]string{
	"R": "실행 중",
	"S": "대기 중",
	"D": "I/O 대기",
	"Z": "좀비",
	"T": "중지됨",
	"t": "디버깅 중",
	"X": "종료됨",
	"I": "유휴",
}

// GetProcessStatusText는 프로세스 상태 코드를 텍스트로 변환합니다.
func GetProcessStatusText(status []string) string {
	if len(status) == 0 {
		return "-"
	}
	if text, ok := ProcessStatusMap[status[0]]; ok {
		return text
	}
	return status[0]
}
