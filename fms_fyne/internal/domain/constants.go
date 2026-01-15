// Package domain은 FMS 애플리케이션의 핵심 엔티티와 비즈니스 규칙을 정의합니다.
package domain

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

// 규칙 결과 상태 상수
const (
	RuleStatusOK         = "ok"
	RuleStatusError      = "error"
	RuleStatusUnfind     = "unfind"
	RuleStatusValidation = "validation"
	RuleStatusWrite      = "write"
)

// 이력 유형 상수
const (
	HistoryTypeFirewall = "firewall" // 방화벽 룰 배포
	HistoryTypeProgram  = "program"  // 프로그램 업데이트
)

// 연결 모드 상수
const (
	ConnectionModeAgent  = "agent"  // 에이전트 서버를 통한 연결
	ConnectionModeDirect = "direct" // 직접 연결
)

// 인증 타입 상수
type AuthType string

const (
	AuthTypeNone     AuthType = "none"     // 인증 정보 없음
	AuthTypePassword AuthType = "password" // 비밀번호 인증
	AuthTypePPK      AuthType = "ppk"      // PPK 키 인증
)

// 기본 타임아웃 (초)
const DefaultTimeoutSeconds = 5

// 기본 SSH 포트
const DefaultSSHPort = 22

// 기본 HTTP 포트
const DefaultHTTPPort = 8080
