# FMS (Firewall Management System) - Claude 개발 가이드

## 프로젝트 개요

Go Fyne 패키지를 이용하여 구현하는 방화벽 관리 시스템(FMS) 데스크톱 애플리케이션입니다.
기존 웹 애플리케이션(index.html)의 기능을 Go Fyne으로 재구현합니다.

---

## 중요 지침

### 기존 코드 수정 금지

- **`smartfw_hs/` 폴더 내의 모든 파일 수정 금지**
- **`index.html` 파일 수정 금지**
- 기존 코드의 함수명, 변수명 등을 임의로 수정하면 안됨
- 기존 코드는 참조용으로만 사용

### 언어 및 커뮤니케이션

- **모든 응답, 주석, 문서, 커밋 메시지를 한글로 작성**
- 질문에 대답할 때 확실하지 않으면 추론으로 대답하지 말 것
- 모르면 코드를 찾아보고 답변해야 함

### 개발 환경

| 환경 | 플랫폼 | 용도 |
|------|--------|------|
| 개발 환경 | windows (로컬) | 코드 작성 및 빌드 |

---

## 프로젝트 구조

```
fms_go/
├── CLAUDE.md               # Claude 개발 가이드 (현재 파일)
├── flutter_wails_fyne.md   # 프레임워크 비교 문서
├── index.html              # 원본 웹 애플리케이션 (참조용, 수정 금지)
├── fms_fyne/               # Go Fyne 프로젝트
├── fms_flutter/            # Flutter 프로젝트
└── fms_wails/              # Wails 프로젝트
```

---

## 주요 기능

### 1. 템플릿 관리
- 템플릿 목록 표시 (라디오 버튼 선택)
- 템플릿 조회/저장/삭제

### 2. 장비(방화벽) 관리
- 장비 목록 테이블 (체크박스, 장비명, 서버상태, 배포상태, 버전)
- 장비 추가/저장/삭제

### 3. 서버 운영
- 서버 상태 확인 (HTTP 연결 테스트)
- 배포 (Agent 서버를 통해 템플릿을 방화벽 장비에 전달)
- 배포 결과 상세 표시

### 4. 데이터 Import/Export
- JSON 파일로 내보내기/가져오기
- 데이터 초기화

---

## 기술 스택

### 공통 패키지
- `net/http` - HTTP 통신 (Agent/Direct 모드)
- `encoding/json` - 데이터 저장

### fms_fyne 전용
- `fyne.io/fyne/v2` - GUI 프레임워크

### fms_wails 전용
- `github.com/wailsapp/wails/v2` - 데스크톱 앱 프레임워크
- React + TypeScript - 프론트엔드

### 통신 방식
- **Agent 모드**: 프록시 서버(Agent)를 경유하여 방화벽 장비와 통신
- **Direct 모드**: 방화벽 장비에 직접 HTTP 연결

---

## 데이터 구조

### 템플릿
```go
type Template struct {
    Version  string `json:"version"`
    Contents string `json:"contents"`
}
```

### 장비
```go
type Firewall struct {
    Index        int           `json:"index"`
    DeviceName   string        `json:"deviceName"`   // 장비 IP 주소
    ServerStatus string        `json:"serverStatus"` // "running", "stop", "-"
    DeployStatus string        `json:"deployStatus"` // "success", "fail", "-"
    Version      string        `json:"version"`
    DeployResult *DeployResult `json:"deployResult,omitempty"`
}
```

### 설정
```go
type Config struct {
    ConnectionMode string `json:"connectionMode"` // "agent" 또는 "direct"
    AgentServerURL string `json:"agentServerURL"` // 예: http://{agent-server}:8080
    TimeoutSeconds int    `json:"timeoutSeconds"` // HTTP 타임아웃 (5~120초)
}
```

---

## 통신 아키텍처

### Agent 모드 (권장)
```
FMS 클라이언트 → Agent 서버 (프록시) → 방화벽 장비들
```

- 상태 확인: `POST http://{agent-server}:8080/agent/req-respCheck`
- 배포: `POST http://{agent-server}:8080/agent/req-deploy`

### Direct 모드
```
FMS 클라이언트 → 방화벽 장비 (직접 연결)
```

- 상태 확인: `GET http://{장비IP}/respCheck`
- 배포: `POST http://{장비IP}/agent/req-deploy`

---

## 템플릿 규칙 포맷

규칙은 Agent 서버를 통해 방화벽 장비의 `/proc/smartfw` 커널 모듈로 전달됩니다.

```
req|INSERT|{ID}|{CHAIN}|{ACTION}|{PROTOCOL}|{SRC}|{DST}|{옵션들}
```

### 예시
```
req|INSERT|3813792919|INPUT|FLUSH|ANY|ANY|ANY|||
req|INSERT|3813792919|INPUT|ACCEPT|TCP|192.168.1.0/24|ANY|80||
```

---

## 참조 문서

- [FMS_SPEC.md](FMS_SPEC.md) - 상세 기능 명세 및 구현 예시 코드
- [DEV_PROC.md](DEV_PROC.md) - 개발 절차서, UI 레이아웃, 반응형 가이드, 커스텀 컴포넌트
- [index.html](index.html) - 원본 웹 애플리케이션 (참조용)
- [smartfw_hs/Makefile](smartfw_hs/Makefile) - 규칙 포맷 및 SSH 설정 참조

---

## 빌드 및 실행

### 빌드 파일명 규칙

| 플랫폼 | 파일명 | 비고 |
|--------|--------|------|
| Windows | `fms_fyne.exe` | 운영 환경 |

### Windows에서 빌드
```bash
cd fms_fyne
go mod download
go mod tidy
go build -ldflags "-H windowsgui -s -w" -o fms_fyne.exe .
```
