package model

// 방화벽 장비 정보를 나타냅니다.
type Firewall struct {
	Index        int           `json:"index"`                  // 고유 ID (Auto Increment)
	DeviceName   string        `json:"deviceName"`             // 장비 IP 주소
	ServerStatus string        `json:"serverStatus"`           // 서버 상태 (running/stop/-)
	DeployStatus string        `json:"deployStatus"`           // 배포 상태 (success/fail/error/-)
	Version      string        `json:"version"`                // 배포된 템플릿 버전
	DeployResult *DeployResult `json:"deployResult,omitempty"` // 마지막 배포 결과
}

// 배포 결과를 나타냅니다.
type DeployResult struct {
	IP     string       `json:"ip"`             // 장비 IP
	Status string       `json:"status"`         // 배포 상태 (success/fail)
	Info   []ResultInfo `json:"info,omitempty"` // 규칙별 상세 결과
}

// 규칙별 배포 결과 상세 정보를 나타냅니다.
type ResultInfo struct {
	Index  int    `json:"index"`  // 규칙 순서
	Rule   string `json:"rule"`   // 실제 적용된 규칙
	Text   string `json:"text"`   // 서버에서 처리된 규칙 텍스트
	Status string `json:"status"` // 결과 (ok/fail)
	Reason string `json:"reason"` // 사유
}

// 서버 상태 상수
const (
	ServerStatusRunning = "running"
	ServerStatusStop    = "stop"
	ServerStatusUnknown = "-"
)

// 배포 상태 상수
const (
	DeployStatusSuccess = "success"
	DeployStatusFail    = "fail"
	DeployStatusError   = "error"
	DeployStatusUnknown = "-"
)

// 새로운 장비를 생성합니다.
func NewFirewall(deviceName string) *Firewall {
	return &Firewall{
		Index:        -1, // 새 장비는 -1로 시작, 저장 시 ID 할당
		DeviceName:   deviceName,
		ServerStatus: ServerStatusUnknown,
		DeployStatus: DeployStatusUnknown,
		Version:      "-",
	}
}

// 장비 정보가 유효한지 검사합니다.
func (f *Firewall) IsValid() bool {
	return f.DeviceName != ""
}

// 장비의 복사본을 반환합니다.
func (f *Firewall) Clone() *Firewall {
	clone := &Firewall{
		Index:        f.Index,
		DeviceName:   f.DeviceName,
		ServerStatus: f.ServerStatus,
		DeployStatus: f.DeployStatus,
		Version:      f.Version,
	}

	// DeployResult 복사
	if f.DeployResult != nil {
		clone.DeployResult = &DeployResult{
			IP:     f.DeployResult.IP,
			Status: f.DeployResult.Status,
		}
		if len(f.DeployResult.Info) > 0 {
			clone.DeployResult.Info = make([]ResultInfo, len(f.DeployResult.Info))
			copy(clone.DeployResult.Info, f.DeployResult.Info)
		}
	}

	return clone
}

// 서버 상태 코드를 표시 텍스트로 변환합니다.
func GetServerStatusText(status string) string {
	switch status {
	case ServerStatusRunning:
		return "정상"
	case ServerStatusStop:
		return "정지"
	default:
		return "-"
	}
}

// 배포 상태 코드를 표시 텍스트로 변환합니다.
func GetDeployStatusText(status string) string {
	switch status {
	case DeployStatusSuccess:
		return "성공"
	case DeployStatusFail:
		return "실패"
	case DeployStatusError:
		return "확인요망"
	default:
		return "-"
	}
}
