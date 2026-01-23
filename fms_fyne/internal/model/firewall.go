package model

// 방화벽 장비 정보를 나타냅니다. (PRD 4.2 기준)
type Firewall struct {
	Index        int           `json:"index"`                  // 고유 ID (Auto Increment)
	DeviceName   string        `json:"deviceName"`             // 장비명 (PRD: device_name)
	DeviceIP     string        `json:"device_ip,omitempty"`    // 장비 IP (PRD: device_ip)
	ServerStatus string        `json:"serverStatus"`           // 서버 상태 (running/stop/-)
	DeployStatus string        `json:"deployStatus"`           // 배포 상태 (success/fail/error/-)
	Version      string        `json:"version"`                // 배포된 템플릿 버전
	DeployResult *DeployResult `json:"deployResult,omitempty"` // 마지막 배포 결과

	// SSH 인증 정보 (패키지 업데이트용)
	DeviceID  string `json:"device_id,omitempty"`  // SSH 사용자 ID (PRD: device_id)
	DevicePW  string `json:"device_pw,omitempty"`  // SSH 비밀번호 (PRD: device_pw)
	DevicePPK string `json:"device_ppk,omitempty"` // PPK 키 파일 경로 (PRD: device_ppk)

	// 패키지 버전 정보
	ProgramVersions map[string]string `json:"program_versions,omitempty"` // 패키지명 -> 버전

	// 상태 정보 (PRD: lastCheckedAt)
	LastCheckedAt string `json:"lastCheckedAt,omitempty"` // 마지막 상태 확인 시간

	// 장비별 API 경로 설정 (비어있으면 전역 설정 사용)
	ProgramUploadPath  string `json:"programUploadPath,omitempty"`  // 패키지 업로드 경로 (SFTP)
	ProgramUpdatePath  string `json:"programUpdatePath,omitempty"`  // 패키지 업데이트 요청 경로
	ProgramUpdatePort  int    `json:"programUpdatePort,omitempty"`  // 패키지 업데이트 요청 포트
	FirewallDeployPath string `json:"firewallDeployPath,omitempty"` // 방화벽 규칙 배포 경로
	FirewallDeployPort int    `json:"firewallDeployPort,omitempty"` // 방화벽 규칙 배포 포트
	DeviceReportPath   string `json:"deviceReportPath,omitempty"`   // 서버 상태 체크 경로
	DeviceReportPort   int    `json:"deviceReportPort,omitempty"`   // 서버 상태 체크 포트
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

// 장비명과 IP로 새로운 장비를 생성합니다.
func NewFirewallWithIP(deviceName, deviceIP string) *Firewall {
	fw := NewFirewall(deviceName)
	fw.DeviceIP = deviceIP
	return fw
}

// 인증 방식을 반환합니다.
func (f *Firewall) GetAuthType() AuthType {
	if f.DevicePPK != "" {
		return AuthTypePPK
	}
	if f.DevicePW != "" {
		return AuthTypePassword
	}
	return AuthTypeNone
}

// SSH 인증 정보가 설정되어 있는지 확인합니다.
func (f *Firewall) HasSSHCredentials() bool {
	return f.DeviceID != "" && (f.DevicePW != "" || f.DevicePPK != "")
}

// 특정 패키지의 버전을 반환합니다.
func (f *Firewall) GetProcessVersion(processName string) string {
	if f.ProgramVersions == nil {
		return "-"
	}
	if version, ok := f.ProgramVersions[processName]; ok {
		return version
	}
	return "-"
}

// 인증 타입을 표시 텍스트로 변환합니다.
func GetAuthTypeText(authType AuthType) string {
	switch authType {
	case AuthTypePassword:
		return "비밀번호"
	case AuthTypePPK:
		return "PPK"
	default:
		return "-"
	}
}

// 장비 정보가 유효한지 검사합니다.
func (f *Firewall) IsValid() bool {
	return f.DeviceName != ""
}

// 장비의 복사본을 반환합니다.
func (f *Firewall) Clone() *Firewall {
	clone := &Firewall{
		Index:              f.Index,
		DeviceName:         f.DeviceName,
		DeviceIP:           f.DeviceIP,
		ServerStatus:       f.ServerStatus,
		DeployStatus:       f.DeployStatus,
		Version:            f.Version,
		DeviceID:           f.DeviceID,
		DevicePW:           f.DevicePW,
		DevicePPK:          f.DevicePPK,
		LastCheckedAt:      f.LastCheckedAt,
		ProgramUploadPath:  f.ProgramUploadPath,
		ProgramUpdatePath:  f.ProgramUpdatePath,
		ProgramUpdatePort:  f.ProgramUpdatePort,
		FirewallDeployPath: f.FirewallDeployPath,
		FirewallDeployPort: f.FirewallDeployPort,
		DeviceReportPath:   f.DeviceReportPath,
		DeviceReportPort:   f.DeviceReportPort,
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

	// ProgramVersions 복사
	if f.ProgramVersions != nil {
		clone.ProgramVersions = make(map[string]string)
		for k, v := range f.ProgramVersions {
			clone.ProgramVersions[k] = v
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
