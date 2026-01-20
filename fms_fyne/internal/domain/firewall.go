package domain

// DeviceProcessInfo 장비에 설치된 패키지 정보 (device-report 응답용)
type DeviceProcessInfo struct {
	Name    string `json:"name"`    // 패키지 이름
	Version string `json:"version"` // 설치된 버전
}

// Firewall 방화벽 장비 정보 엔티티
type Firewall struct {
	// 기존 필드
	Index        int           `json:"index"`                  // 고유 ID (Auto Increment)
	DeviceName   string        `json:"deviceName"`             // 장비 이름
	ServerStatus string        `json:"serverStatus"`           // 서버 상태 (running/stop/-)
	DeployStatus string        `json:"deployStatus"`           // 배포 상태 (success/fail/error/-)
	Version      string        `json:"version"`                // 배포된 룰 버전
	DeployResult *DeployResult `json:"deployResult,omitempty"` // 마지막 배포 결과

	// 신규 필드 (SSH 인증)
	DeviceIP  string `json:"device_ip"`  // 장비 IP 주소
	DeviceID  string `json:"device_id"`  // SSH 접속 ID
	DevicePW  string `json:"device_pw"`  // SSH 비밀번호
	DevicePPK string `json:"device_ppk"` // PPK 파일 경로

	// 신규 필드 (패키지 버전)
	Processes []DeviceProcessInfo `json:"processes"` // 설치된 패키지 목록

	// 신규 필드 (상태 정보)
	LastCheckedAt string `json:"lastCheckedAt"` // 마지막 상태 확인 시간
}

// DeployResult 배포 결과
type DeployResult struct {
	IP     string       `json:"ip"`             // 장비 IP
	Status string       `json:"status"`         // 배포 상태 (success/fail)
	Info   []ResultInfo `json:"info,omitempty"` // 규칙별 상세 결과
}

// ResultInfo 규칙별 배포 결과 상세 정보
type ResultInfo struct {
	Index  int    `json:"index"`  // 규칙 순서
	Rule   string `json:"rule"`   // 실제 적용된 규칙
	Text   string `json:"text"`   // 서버에서 처리된 규칙 텍스트
	Status string `json:"status"` // 결과 (ok/fail)
	Reason string `json:"reason"` // 사유
}

// NewFirewall 새로운 장비를 생성합니다.
func NewFirewall(deviceName string) *Firewall {
	return &Firewall{
		Index:        -1, // 새 장비는 -1로 시작, 저장 시 ID 할당
		DeviceName:   deviceName,
		ServerStatus: ServerStatusUnknown,
		DeployStatus: DeployStatusUnknown,
		Version:      "-",
		Processes:    []DeviceProcessInfo{},
	}
}

// NewFirewallWithIP 장비명과 IP로 새로운 장비를 생성합니다.
func NewFirewallWithIP(deviceName, deviceIP string) *Firewall {
	fw := NewFirewall(deviceName)
	fw.DeviceIP = deviceIP
	return fw
}

// IsValid 장비 정보가 유효한지 검사합니다.
func (f *Firewall) IsValid() bool {
	return f.DeviceName != "" && f.DeviceIP != ""
}

// GetAuthType 인증 방식을 반환합니다.
func (f *Firewall) GetAuthType() AuthType {
	if f.DevicePPK != "" {
		return AuthTypePPK
	}
	if f.DevicePW != "" {
		return AuthTypePassword
	}
	return AuthTypeNone
}

// HasSSHCredentials SSH 인증 정보가 설정되어 있는지 확인합니다.
func (f *Firewall) HasSSHCredentials() bool {
	return f.DeviceID != "" && (f.DevicePW != "" || f.DevicePPK != "")
}

// GetProcessVersion 특정 패키지의 버전을 반환합니다.
func (f *Firewall) GetProcessVersion(processName string) string {
	for _, p := range f.Processes {
		if p.Name == processName {
			return p.Version
		}
	}
	return "-"
}

// Clone 장비의 복사본을 반환합니다.
func (f *Firewall) Clone() *Firewall {
	clone := &Firewall{
		Index:         f.Index,
		DeviceName:    f.DeviceName,
		ServerStatus:  f.ServerStatus,
		DeployStatus:  f.DeployStatus,
		Version:       f.Version,
		DeviceIP:      f.DeviceIP,
		DeviceID:      f.DeviceID,
		DevicePW:      f.DevicePW,
		DevicePPK:     f.DevicePPK,
		LastCheckedAt: f.LastCheckedAt,
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

	// Processes 복사
	if len(f.Processes) > 0 {
		clone.Processes = make([]DeviceProcessInfo, len(f.Processes))
		copy(clone.Processes, f.Processes)
	} else {
		clone.Processes = []DeviceProcessInfo{}
	}

	return clone
}

// GetServerStatusText 서버 상태 코드를 표시 텍스트로 변환합니다.
func GetServerStatusText(status string) string {
	switch status {
	case ServerStatusRunning:
		return "정상"
	case ServerStatusStop:
		return "연결안됨"
	default:
		return "알수없음"
	}
}

// GetDeployStatusText 배포 상태 코드를 표시 텍스트로 변환합니다.
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

// GetAuthTypeText 인증 타입을 표시 텍스트로 변환합니다.
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
