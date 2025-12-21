package model

// 연결 모드 상수
const (
	ConnectionModeAgent  = "agent"  // 에이전트 서버를 통한 연결
	ConnectionModeDirect = "direct" // 직접 연결
)

// 기본 타임아웃 (초)
const DefaultTimeoutSeconds = 10

// 애플리케이션 설정을 나타냅니다.
type Config struct {
	ConnectionMode string `json:"connectionMode"` // 연결 모드: "agent" 또는 "direct"
	AgentServerURL string `json:"agentServerURL"` // 에이전트 서버 URL (예: http://172.24.10.6:8080)
	TimeoutSeconds int    `json:"timeoutSeconds"` // HTTP 타임아웃 (초)
}

// 기본 설정을 반환합니다.
func DefaultConfig() *Config {
	return &Config{
		ConnectionMode: ConnectionModeDirect,
		AgentServerURL: "http://172.24.10.6:8080",
		TimeoutSeconds: DefaultTimeoutSeconds,
	}
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

// 연결 모드가 에이전트 모드인지 확인합니다.
func (c *Config) IsAgentMode() bool {
	return c.ConnectionMode == ConnectionModeAgent
}

// 연결 모드가 직접 연결 모드인지 확인합니다.
func (c *Config) IsDirectMode() bool {
	return c.ConnectionMode == ConnectionModeDirect
}
