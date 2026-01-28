// Package model은 FMS 애플리케이션의 데이터 모델을 정의합니다.
package model

// 기본 타임아웃 (초)
const DefaultTimeoutSeconds = 30

// 기본 SSH 포트
const DefaultSSHPort = 22

// 기본 HTTP 포트
const DefaultHTTPPort = 8080

// 기본 원격 업로드 경로
const DefaultRemotePath = "/download/"

// 인증 타입 상수
type AuthType string

const (
	AuthTypeNone     AuthType = "none"     // 인증 정보 없음
	AuthTypePassword AuthType = "password" // 비밀번호 인증
	AuthTypePPK      AuthType = "ppk"      // PPK 키 인증
)

// 테마 상수
const (
	ThemeLight = "light"
	ThemeDark  = "dark"
)

// 기본 장비 체크 주기 (초)
const DefaultStatusCheckInterval = 60

// 기본 API 경로
const (
	DefaultProgramUpdatePath  = "/program-update"
	DefaultFirewallDeployPath = "/agent/req-deploy"
	DefaultDeviceInfoPath     = "/device-report" // 장비 상세보기 및 상태 체크 경로
)

// 기본 API 포트
const DefaultAPIPort = 8080

// HTTP 스킴 상수
const (
	SchemeHTTP  = "http"
	SchemeHTTPS = "https"
)

// 애플리케이션 설정을 나타냅니다.
type Config struct {
	TimeoutSeconds      int    `json:"timeoutSeconds"`      // HTTP 타임아웃 (초)
	Theme               string `json:"theme"`               // 테마: "light" 또는 "dark"
	AutoStatusCheck     bool   `json:"autoStatusCheck"`     // 자동 상태 체크 활성화 여부
	StatusCheckInterval int    `json:"statusCheckInterval"` // 장비 체크 주기 (초)

	// API 경로 설정
	ProgramUpdateScheme string `json:"programUpdateScheme"` // 패키지 업데이트 요청 스킴 (http/https)
	ProgramUpdatePath   string `json:"programUpdatePath"`   // 패키지 업데이트 요청 경로
	ProgramUpdatePort   int    `json:"programUpdatePort"`   // 패키지 업데이트 요청 포트
	FirewallDeployScheme string `json:"firewallDeployScheme"` // 방화벽 규칙 배포 스킴 (http/https)
	FirewallDeployPath  string `json:"firewallDeployPath"`  // 방화벽 규칙 배포 경로
	FirewallDeployPort  int    `json:"firewallDeployPort"`  // 방화벽 규칙 배포 포트
	DeviceInfoScheme    string `json:"deviceInfoScheme"`    // 장비 상세보기 스킴 (http/https)
	DeviceInfoPath      string `json:"deviceInfoPath"`      // 장비 상세보기 및 상태 체크 경로
	DeviceInfoPort      int    `json:"deviceInfoPort"`      // 장비 상세보기 및 상태 체크 포트
}

// 기본 설정을 반환합니다.
func DefaultConfig() *Config {
	return &Config{
		TimeoutSeconds:       DefaultTimeoutSeconds,
		Theme:                ThemeLight,
		AutoStatusCheck:      false,
		StatusCheckInterval:  DefaultStatusCheckInterval,
		ProgramUpdateScheme:  SchemeHTTP,
		ProgramUpdatePath:    DefaultProgramUpdatePath,
		ProgramUpdatePort:    DefaultAPIPort,
		FirewallDeployScheme: SchemeHTTP,
		FirewallDeployPath:   DefaultFirewallDeployPath,
		FirewallDeployPort:   DefaultAPIPort,
		DeviceInfoScheme:     SchemeHTTP,
		DeviceInfoPath:       DefaultDeviceInfoPath,
		DeviceInfoPort:       DefaultAPIPort,
	}
}

// 스킴 값을 반환합니다 (기본값: http)
func (c *Config) GetProgramUpdateScheme() string {
	if c.ProgramUpdateScheme == SchemeHTTPS {
		return SchemeHTTPS
	}
	return SchemeHTTP
}

func (c *Config) GetFirewallDeployScheme() string {
	if c.FirewallDeployScheme == SchemeHTTPS {
		return SchemeHTTPS
	}
	return SchemeHTTP
}

func (c *Config) GetDeviceInfoScheme() string {
	if c.DeviceInfoScheme == SchemeHTTPS {
		return SchemeHTTPS
	}
	return SchemeHTTP
}

// 장비 체크 주기를 반환합니다 (최소 10초, 최대 300초)
func (c *Config) GetStatusCheckInterval() int {
	if c.StatusCheckInterval < 10 {
		return 10
	}
	if c.StatusCheckInterval > 300 {
		return 300
	}
	return c.StatusCheckInterval
}

// 타임아웃 값을 반환합니다 (최소 5초, 최대 120초)
func (c *Config) GetTimeoutSeconds() int {
	if c.TimeoutSeconds < 5 {
		return 5
	}
	if c.TimeoutSeconds > 120 {
		return 120
	}
	return c.TimeoutSeconds
}
