# FMS Wails (Firewall Management System)

Go + Wails + React를 이용한 방화벽 관리 시스템 데스크톱 애플리케이션입니다.

---

## 프로젝트 구조

```
fms_wails/
├── CLAUDE.md           # Claude 개발 가이드
├── README.md           # 프로젝트 문서 (현재 파일)
├── app.go              # Wails 백엔드 API (프론트엔드에서 호출)
├── main.go             # 앱 진입점
├── wails.json          # Wails 설정
├── go.mod / go.sum     # Go 모듈 정의
├── internal/           # Go 백엔드 로직
│   ├── model/          # 데이터 모델
│   │   ├── template.go   # Template 모델
│   │   ├── firewall.go   # Firewall 모델
│   │   ├── history.go    # DeployHistory 모델
│   │   └── config.go     # Config 모델
│   ├── storage/        # JSON 파일 기반 저장소
│   │   ├── storage.go    # 인터페이스 정의
│   │   └── json_store.go # JSON 파일 저장소 구현
│   ├── deploy/         # 배포 로직
│   │   └── deployer.go   # HTTP 기반 배포
│   ├── http/           # HTTP 클라이언트
│   │   └── client.go     # Agent/Direct 모드 HTTP 통신
│   └── version/        # 버전 정보
│       └── version.go    # 앱 버전 상수
├── frontend/           # React 프론트엔드
│   ├── src/
│   │   ├── App.tsx     # 메인 앱 컴포넌트
│   │   ├── App.css     # 전역 스타일
│   │   └── components/ # 탭별 컴포넌트
│   │       ├── TemplateTab.tsx   # 템플릿 관리 탭
│   │       ├── DeviceTab.tsx     # 장비 관리 탭
│   │       └── HistoryTab.tsx    # 배포 이력 탭
│   └── wailsjs/        # Wails 자동 생성 바인딩
└── build/              # 빌드 관련 리소스
    ├── darwin/         # macOS 설정
    ├── windows/        # Windows 설정
    └── bin/            # 빌드 결과물
```

---

## 주요 기능

### 1. 템플릿 관리
- 템플릿 목록 표시 (리스트 선택)
- 템플릿 조회/저장/삭제

### 2. 장비(방화벽) 관리
- 장비 목록 테이블 (체크박스, 장비명, 서버상태, 배포상태, 버전)
- 장비 추가/편집/삭제
- 서버 상태 확인 (선택된 장비만 또는 전체)

### 3. 배포
- 선택한 장비에 템플릿 배포
- 배포 진행 상태 표시
- HTTP를 통한 원격 장비 통신

### 4. 배포 이력
- 배포 이력 목록 및 상세 조회
- 규칙별 성공/실패 확인
- 전체 삭제 기능

### 5. 데이터 관리
- JSON 파일로 내보내기/가져오기
- 데이터 초기화
- 파일에서 새로고침 (ReloadData)

---

## 기술 스택

### 백엔드 (Go)
- `github.com/wailsapp/wails/v2` - 데스크톱 앱 프레임워크
- `net/http` - HTTP 클라이언트
- `encoding/json` - 데이터 저장

### 프론트엔드 (React + TypeScript)
- `react` - UI 프레임워크
- `vite` - 빌드 도구
- `typescript` - 타입 안정성
- `recharts` - 차트 라이브러리

### 통신 방식
- **Agent 모드**: 에이전트 서버를 통한 배포 (HTTP)
- **Direct 모드**: 각 장비에 직접 연결 (HTTP)

---

## 데이터 구조

### 설정 (Config)
```go
type Config struct {
    ConnectionMode string `json:"connectionMode"` // "agent" 또는 "direct"
    AgentServerURL string `json:"agentServerURL"` // 에이전트 서버 URL
    TimeoutSeconds int    `json:"timeoutSeconds"` // HTTP 타임아웃 (5~120초)
}
```

### 템플릿 (Template)
```go
type Template struct {
    Version  string `json:"version"`  // 템플릿 버전 (고유 키)
    Contents string `json:"contents"` // 규칙 내용
}
```

### 장비 (Firewall)
```go
type Firewall struct {
    Index        int    `json:"index"`        // 고유 ID (Auto Increment)
    DeviceName   string `json:"deviceName"`   // 장비 IP 주소
    ServerStatus string `json:"serverStatus"` // running/stop/-
    DeployStatus string `json:"deployStatus"` // success/fail/error/-
    Version      string `json:"version"`      // 배포된 템플릿 버전
}
```

### 배포 이력 (DeployHistory)
```go
type DeployHistory struct {
    ID           int          `json:"id"`           // 고유 ID
    DeviceName   string       `json:"deviceName"`   // 장비 IP
    Version      string       `json:"version"`      // 템플릿 버전
    Status       string       `json:"status"`       // success/fail
    DeployedAt   time.Time    `json:"deployedAt"`   // 배포 시간
    TotalRules   int          `json:"totalRules"`   // 총 규칙 수
    SuccessCount int          `json:"successCount"` // 성공 수
    FailCount    int          `json:"failCount"`    // 실패 수
    Results      []RuleResult `json:"results"`      // 규칙별 결과
}

type RuleResult struct {
    Rule   string `json:"rule"`   // 규칙 번호
    Text   string `json:"text"`   // 규칙 내용
    Status string `json:"status"` // ok/error/unfind/validation
    Reason string `json:"reason"` // 결과 메시지
}
```

---

## 상태 값

### 서버 상태 (ServerStatus)
| 값 | 표시 | 설명 |
|---|---|---|
| `running` | 정상 | 서버 정상 동작 |
| `stop` | 정지 | 서버 응답 없음 |
| `-` | - | 확인되지 않음 |

### 배포 상태 (DeployStatus)
| 값 | 표시 | 설명 |
|---|---|---|
| `success` | 성공 | 모든 규칙 배포 성공 |
| `fail` | 실패 | 일부/전체 규칙 실패 |
| `error` | 확인요망 | 통신 오류 등 |
| `-` | - | 배포되지 않음 |

---

## 버전 관리

앱 버전은 `internal/version/version.go`에서 관리됩니다.

```go
const (
    AppVersion  = "1.0.0"
    AppName     = "FMS"
    AppFullName = "Firewall Management System"
)
```

하단 푸터에 버전이 표시되며, 클릭 시 차트 데모를 볼 수 있습니다.

---

## 템플릿 규칙 포맷

규칙은 `/proc/smartfw` 커널 모듈 인터페이스로 전달됩니다.

```
req|INSERT|{ID}|{CHAIN}|{ACTION}|{PROTOCOL}|{SRC}|{DST}|{옵션들}
```

### 예시
```
req|INSERT|3813792919|INPUT|FLUSH|ANY|ANY|ANY|||
req|INSERT|3813792919|INPUT|ACCEPT|TCP|192.168.1.0/24|ANY|80||
```

---

## 빌드 및 실행

### 사전 요구사항
- Go 1.21+
- Node.js 18+
- Wails CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

### 빌드 결과물

| 플랫폼 | 결과물 | 비고 |
|--------|--------|------|
| macOS | `build/bin/fms_wails.app` | 개발 환경 |
| Linux | `build/bin/fms_wails` | 운영 환경 |
| Windows | `build/bin/fms_wails.exe` | Windows 환경 |

### 개발 모드 실행
```bash
cd fms_wails
wails dev
```
- 핫 리로드 지원
- 개발자 도구 사용 가능

### 프로덕션 빌드
```bash
cd fms_wails

# macOS (현재 환경)
wails build

# Linux용 크로스 컴파일
wails build -platform linux/amd64

# Windows용 크로스 컴파일 (MinGW 필요)
wails build -platform windows/amd64
```

### macOS 앱 실행
```bash
./build/bin/fms_wails.app/Contents/MacOS/fms_wails
# 또는
open build/bin/fms_wails.app
```

### Linux 서버에서 실행
```bash
scp build/bin/fms_wails user@192.168.x.x:/path/to/
ssh user@192.168.x.x
./fms_wails
```

---

## Wails 바인딩

Go 함수를 프론트엔드에서 호출하려면 `app.go`에 public 메서드로 정의합니다.

```go
// app.go
func (a *App) GetAllTemplates() []*model.Template {
    // ...
}
```

프론트엔드에서 호출:
```typescript
import { GetAllTemplates } from '../wailsjs/go/main/App';

const templates = await GetAllTemplates();
```

Wails 빌드 시 자동으로 `frontend/wailsjs/` 디렉토리에 TypeScript 바인딩이 생성됩니다.

---

## 설정 파일 위치

앱 실행 시 실행 파일과 같은 위치에 `config/` 디렉토리가 생성됩니다.

```
config/
├── config.json      # 앱 설정 (연결 모드, 타임아웃 등)
├── templates.json   # 템플릿 데이터
├── firewalls.json   # 장비 데이터
└── history.json     # 배포 이력
```

---

## 주요 API 목록

### 설정
- `GetConfig()` - 현재 설정 조회
- `SaveConfig(json)` - 설정 저장

### 템플릿
- `GetAllTemplates()` - 모든 템플릿 조회
- `GetTemplate(version)` - 특정 템플릿 조회
- `SaveTemplate(version, contents)` - 템플릿 저장
- `DeleteTemplate(version)` - 템플릿 삭제

### 장비
- `GetAllFirewalls()` - 모든 장비 조회
- `GetFirewall(index)` - 특정 장비 조회
- `SaveFirewall(json)` - 장비 저장
- `DeleteFirewall(index)` - 장비 삭제
- `CheckServerStatus(index)` - 장비 상태 확인
- `CheckAllServerStatus()` - 전체 장비 상태 확인

### 배포
- `Deploy(firewallIndex, templateVersion)` - 배포 실행

### 이력
- `GetAllHistory()` - 모든 이력 조회
- `DeleteHistory(id)` - 이력 삭제
- `SaveHistory(json)` - 이력 저장 (Import용)

### 데이터 관리
- `ExportData()` - 전체 데이터 내보내기
- `ImportData(json)` - 데이터 가져오기
- `ResetAll()` - 모든 데이터 초기화
- `ReloadData()` - 파일에서 데이터 다시 로드

### 유틸리티
- `GetAppVersion()` - 앱 버전 조회
- `GetAppName()` - 앱 이름 조회
- `GetAppFullName()` - 앱 전체 이름 조회
- `GetConfigDir()` - 설정 디렉토리 경로 조회
- `ConfirmDialog(title, message)` - 확인 대화상자
- `AlertDialog(title, message)` - 알림 대화상자
- `OpenFileDialog(title)` - 파일 열기 대화상자
- `SaveFileDialog(title, filename)` - 파일 저장 대화상자
