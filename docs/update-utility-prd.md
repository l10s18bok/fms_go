# PRD: smartfw/smartway 업데이트 유틸리티

## 문서 정보

| 항목 | 내용 |
|------|------|
| 문서 버전 | 1.0 |
| 작성일 | 2026-01-12 |
| 상태 | Draft |
| 관련 프로젝트 | fms_fyne |

---

## 1. 개요

### 1.1 배경

현재 FMS(Firewall Management System)는 Agent 서버를 통해 방화벽 장비에 **방화벽 룰(규칙)**을 배포하는 기능을 제공합니다. 그러나 smartfw/smartway **프로그램 자체의 업데이트**는 별도의 수동 작업이 필요합니다.

본 문서는 smartfw/smartway 프로그램을 원격으로 업데이트할 수 있는 유틸리티의 요구사항을 정의합니다.

### 1.2 목적

- smartfw/smartway 프로그램의 원격 업데이트 자동화
- 다수의 장비에 대한 일괄 업데이트 지원
- 장비 상태 및 버전 정보의 주기적 수집
- 업데이트 이력 관리

### 1.3 기존 FMS와의 차이점

| 구분 | 기존 FMS (방화벽 룰 배포) | 업데이트 유틸리티 |
|------|----------------------|------------------|
| 대상 | 방화벽 규칙 (방화벽 룰) | smartfw/smartway 프로그램 |
| 통신 방식 | HTTP (Agent 경유 또는 Direct) | SSH + HTTP (1차), HTTPS (향후) |
| 인증 | 없음 (Agent 신뢰) | SSH 비밀번호 또는 PPK 키 |
| 파일 전송 | HTTP POST Body | SSH (SCP/SFTP) |

---

## 2. 기능 요구사항

### 2.1 프로그램 관리

#### 2.1.1 프로그램 정보 테이블

| 필드명 | 타입 | 설명 | 필수 |
|--------|------|------|------|
| process_name | string | 업데이트 프로세스 이름 | O |
| process_file_path | string | 업데이트 할 파일 저장 경로 (local) | O |
| process_upload_path | string | 서버에 업로드 할 경로 (서버 쪽 저장 경로) | O |
| process_version | string | 프로세스 버전 | O |
| process_created_at | string | 추가(수정) 일시 | O |

**JSON 예시 (programs.json):**
```json
[
  {
    "process_name": "smartfw",
    "process_file_path": "C:\\Updates\\smartfw-v2.1.0.tar",
    "process_upload_path": "/download/",
    "process_version": "2.1.0",
    "process_created_at": "2026-01-12 14:00:00"
  },
  {
    "process_name": "smartway",
    "process_file_path": "C:\\Updates\\smartway-v1.5.0.tar",
    "process_upload_path": "/download/",
    "process_version": "1.5.0",
    "process_created_at": "2026-01-12 15:30:00"
  }
]
```

#### 2.1.2 프로그램 등록 기능

- 이름 입력
- 버전 정보 입력
- 업로드 경로 입력 (기본값: /download/)
- 파일찾기 (찾아보기...) - 로컬 파일 선택 (파일 브라우저)
  - ⚠️ **구현 시 주의**: 다이얼로그 내에서 파일 선택 다이얼로그 호출 시 fyne-docs 스킬의 "다이얼로그 및 팝업 중첩" 섹션 참고 필요 (`.claude/skills/fyne-docs/SKILL.md`)

#### 2.1.3 프로그램 목록 표시

| 컬럼 | 설명 |
|------|------|
| 선택 | 체크박스 (복수 선택) |
| 이름 | process_name |
| 버전 | process_version |
| 업로드 경로 | process_upload_path |
| 로컬파일 경로 | process_file_path |
| 추가(수정)시간 | 등록/수정 일시 |

---

### 2.2 장비 관리

#### 2.2.1 장비 정보 테이블

장비 정보를 저장하기 위한 데이터 구조:

| 필드명 | 타입 | 설명 | 필수 |
|--------|------|------|------|
| device_name | string | 장비 이름 | O |
| device_ip | string | 장비 IP | O |
| device_id | string | 접속 ID | O |
| device_pw | string | 접속 비밀번호 | △ (PPK 없을 때) |
| device_ppk | string | 접속 PPK 파일 경로 | △ (PW 없을 때) |

**인증 방식:**
- **비밀번호 인증**: device_id + device_pw 사용
- **키 기반 인증**: device_id + device_ppk 사용 (권장)
- 둘 중 하나는 반드시 설정되어야 함

**JSON 예시 (firewalls.json):**
```json
[
  {
    "device_name": "firewall-01",
    "device_ip": "192.168.1.100",
    "device_id": "root",
    "device_pw": "password123",
    "device_ppk": ""
  },
  {
    "device_name": "firewall-02",
    "device_ip": "192.168.1.101",
    "device_id": "root",
    "device_pw": "",
    "device_ppk": "C:\\Keys\\server.ppk"
  }
]
```

#### 2.2.2 장비 등록 기능

- 장비명, IP, SSH 아이디 입력
- 인증 방식 선택 (비밀번호 / PPK 키)
  - 비밀번호 선택 시: 비밀번호 입력 필드 표시
  - PPK 선택 시: PPK 파일 경로 선택 (파일 브라우저)
- 등록 전 SSH 연결 테스트 기능 (선택사항)

#### 2.2.3 장비 목록 표시

| 컬럼 | 설명 |
|------|------|
| 선택 | 체크박스 (다중 선택) |
| 장비명 | device_name |
| IP 주소 | device_ip |
| SSH 아이디 | device_id |
| 인증 방식 | "비밀번호" 또는 "PPK" |
| 연결 상태 | 마지막 연결 성공/실패 |
| 현재 버전 | 설치된 smartfw/smartway 버전 |

#### 2.2.4 장비 상태 수집

**서버 연결 상태 확인 (기존):**

서버가 실행 중인지 확인합니다.

**엔드포인트:**
```
GET http://{device_ip}:{HTTP_PORT}/agent/respCheck
```

**응답:** HTTP 200 OK (연결 성공)

---

**프로그램 버전 정보 조회 (신규):**

설정된 주기(예: 5분, 10분, 30분)마다 등록된 모든 장비의 프로그램 버전 정보 수집:

**엔드포인트:**
```
GET http://{device_ip}:{HTTP_PORT}/device-report
```

**응답 예시:**
```json
{
  "processes": [
    {"name": "smartway", "version": "1.0.0"},
    {"name": "smartfw", "version": "2.0.0"}
  ]
}
```

**참고:** 프로그램 종류는 고정되지 않으며, 장비에 설치된 프로그램에 따라 동적으로 반환됩니다.

---

**수집 정보 저장:**
- 장비별 최신 상태 정보 유지
- 버전 정보 자동 갱신
- 연결 실패 시 상태를 "연결 실패"로 표시

**수동 상태 확인:**
- "상태 확인" 버튼으로 선택된 장비의 즉시 상태 조회 가능

**서버 상태 체크 메커니즘 (Fyne 구현 참고):**

서버 상태 체크는 다음과 같은 데이터 흐름으로 동작합니다.

**사용 필드:**
- `firewalls.json`의 **`serverStatus`** 필드에 상태 저장 (값: `"running"`, `"stop"`, `"-"`)
- **`lastCheckedAt`** 필드에 마지막 상태 확인 시간 저장

**데이터 흐름:**

```
client.go (CheckHealth)
        ↓
deployer.go (HealthCheck)
        ↓
firewalls.json (serverStatus 필드)
```

**코드 흐름 설명:**

1. **client.go - `CheckHealth()` 함수**: API를 호출하고 상태 문자열 반환
   - Agent 모드: `CheckHealthViaAgent()` → `POST /agent/req-respCheck`
   - Direct 모드: `CheckHealthDirect()` → `GET http://{장비IP}/agent/respCheck`
   - 반환값: `model.ServerStatusRunning` 또는 `model.ServerStatusStop`

2. **deployer.go - `HealthCheck()` 함수**: 결과를 Firewall 객체에 반영
   ```go
   func (d *Deployer) HealthCheck(fw *model.Firewall) error {
       status, err := d.client.CheckHealth(fw)
       fw.ServerStatus = status  // ← 여기서 반영됨
       return err
   }
   ```

3. **firewalls.json 저장**: UI에서 장비 목록을 저장할 때 `serverStatus` 필드가 JSON에 기록됨

**구현 참고사항:**
- `lastCheckedAt` 필드는 상태 체크 시 현재 시간으로 업데이트되어야 함
- 상태 체크 실패 시 `serverStatus`는 `"stop"`으로 설정

#### 2.2.5 업데이트 배포

**업데이트 흐름:**
1. 유틸리티 → SSH로 `process_file_path`의 파일을 서버의 `process_upload_path`로 업로드
2. 유틸리티 → HTTP로 Agent에 업데이트 요청
3. Agent → 압축 해제, 버전 파일 생성, 프로세스 재시작 (Agent 내부 처리)

**배포 프로세스:**

```
1. 장비 선택 (다중 선택 가능)
2. 프로그램 선택 (단일 선택)
3. 배포 실행
   ├─ SSH 연결 (비밀번호 또는 PPK)
   ├─ 파일 업로드 (SCP/SFTP)
   └─ HTTP API 호출 (POST /program-update)
4. 배포 결과 표시
```

**SSH 연결:**

**비밀번호 인증:**
```
SSH 연결 → device_ip:22
아이디: device_id
비밀번호: device_pw
```

**PPK 키 인증:**
```
SSH 연결 → device_ip:22
아이디: device_id
키 파일: device_ppk
```

**파일 업로드:**

SSH 연결 후 SCP 또는 SFTP 프로토콜을 사용하여 파일 전송:

```
Source: process_file_path (로컬)
Destination: {device_ip}:{process_upload_path}/{파일명}
```

**업데이트 API 호출:**

파일 업로드 완료 후 HTTP API 호출:

**엔드포인트:**
```
POST http://{device_ip}:{HTTP_PORT}/program-update
```

**요청 본문:**
```json
{
  "process_name": "smartfw",
  "process_version": "2.1.0",
  "process_file_path": "/tmp/updates/smartfw_v2.1.0.tar.gz"
}
```

**응답:**
```json
{
  "status": "success",
  "message": "Update completed",
  "installed_version": "2.1.0"
}
```

**참고:** HTTPS 포트는 고정값 사용 (설정 또는 상수로 관리)

**배포 결과 표시:**

각 장비별 배포 결과:
- 성공: 녹색 표시, 설치된 버전 표시
- 실패: 빨간색 표시, 오류 메시지 표시
- 진행 중: 진행률 표시

---

### 2.3 배포 이력 관리

#### 2.3.1 이력 정보

| 필드 | 설명 |
|------|------|
| 일시 | 배포/업데이트 실행 시간 |
| 장비명 | 대상 장비 이름 |
| 장비 IP | 대상 장비 IP |
| 유형 | "방화벽 룰" 또는 "프로그램" |
| 버전 | 배포된 방화벽 룰 버전 또는 프로그램 버전 |
| 결과 | 성공/실패 |

#### 2.3.2 이력 조회

- 유형별 필터 (방화벽 룰 / 프로그램)
- 검색 기능 (장비명, IP, 버전)
- 결과별 필터 (성공/실패)

---

## 3. UI 설계

### 3.1 현재 FMS 구조 분석

현재 fms_fyne의 UI 구조:

```
MainUI (app.go)
├── 방화벽 룰 관리 (TemplateTab)
│   ├── 룰 목록 (라디오 버튼 선택)
│   ├── 규칙 빌더 (RuleBuilder)
│   └── NAT 빌더 (NATBuilder)
├── 프로그램 관리 (ProgramTab) ← 신규
│   ├── 프로그램 목록 테이블 (체크박스 다중 선택)
│   └── 버튼: 찾기, 삭제, 추가/수정
├── 장비 관리 (DeviceTab)
│   ├── 장비 목록 테이블 (체크박스 다중 선택)
│   ├── 방화벽 룰 선택 드롭다운
│   └── 버튼: 추가, 삭제, 상태확인, 배포
└── 배포 이력 (HistoryTab)
    ├── 배포 이력 테이블
    └── 상세 결과 테이블
```

**현재 Firewall 모델 (firewall.go):**
```go
type Firewall struct {
    Index        int           `json:"index"`
    DeviceName   string        `json:"deviceName"`   // 장비 IP
    ServerStatus string        `json:"serverStatus"` // running, stop, -
    DeployStatus string        `json:"deployStatus"` // success, fail, -
    Version      string        `json:"version"`
    DeployResult *DeployResult `json:"deployResult,omitempty"`
}
```

**탭 간 참조 관계:**
- DeviceTab ↔ HistoryTab: 배포 완료 시 이력 자동 추가, 이력 삭제 시 장비 상태 초기화
- DeviceTab → TemplateTab: 장비에 배포할 방화벽 룰 목록 조회
- DeviceTab → ProgramTab: 장비에 업데이트할 프로그램 목록 조회 (신규)

---

### 3.2 최종 권장안: 통합 UI (작업 모드 전환)

기존 장비관리 탭을 확장하여 "방화벽 룰 배포"와 "프로그램 업데이트" 모드를 라디오 버튼으로 전환하는 방식입니다.

**선정 이유:**

1. **기존 UI 구조 유지**: 사용자 학습 곡선 최소화
2. **장비 정보 일원화**: 장비 테이블에 방화벽 룰 버전 + 프로그램 버전 통합 표시
3. **작업 흐름 단순화**: 모드 전환만으로 다른 작업 수행

**핵심 설계 원칙:**

1. **버전 컬럼으로 상태 표시**
   - 성공 시: 버전 표시 (예: "v1.0", "2.1.0")
   - 실패 또는 미배포 시: "-" 표시
   - 별도의 "배포상태" 컬럼 불필요

2. **SSH 설정은 공통 관리**
   - 모든 장비에 동일한 SSH ID 사용 (예: root)
   - 프로그램 관리 탭에서 SSH 접속 설정 관리
   - 장비별 SSH 정보 저장 불필요

3. **기존 버튼 구조 유지**
   - 하단: [전체선택] [전체해제] + [저장] [삭제]
   - 최하단: 장비 추가/수정 영역

---

### 3.3 권장 UI 레이아웃 (통합 UI)

#### 3.3.1 전체 탭 구조

```
┌─────────────────────────────────────────────────────────────────────┐
│ [방화벽 룰 관리] [프로그램 관리] [장비 관리] [배포 이력]                 │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│                    (선택된 탭의 내용 표시)                          │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

**변경사항:**
- 기존: [방화벽 룰 관리] [장비 관리] [배포 이력]
- 변경: [방화벽 룰 관리] [프로그램 관리] [장비 관리] [배포 이력]
- "프로그램 관리" 탭 신규 추가 (프로그램 업데이트용 + SSH 설정)

#### 3.3.2 프로그램 관리 탭 (신규)

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│ [방화벽 룰 관리] [프로그램 관리 X] [장비 관리] [배포 이력]                             │
├─────────────────────────────────────────────────────────────────────────────────────┤
│  프로그램 관리                                                                       │
├─────────────────────────────────────────────────────────────────────────────────────┤
│  검색: [                    ] [찾기]                 [삭제]  [추가/수정]            │
├─────────────────────────────────────────────────────────────────────────────────────┤
│ 선택 │ 이름      │ 버전   │ 업로드 경로    │ 로컬파일 경로              │ 추가(수정)시간 │
├──────┼───────────┼────────┼────────────────┼────────────────────────────┼────────────────┤
│  ☑  │ smartfw_2 │ 3.5.1  │ /download      │ C:\Updates\smartfw.tar     │                │
│      │           │        │                │                            │                │
├─────────────────────────────────────────────────────────────────────────────────────┤
│                                << page 1/10 >>                                       │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

**프로그램 추가 및 수정 다이얼로그:**

`[추가/수정]` 버튼 클릭 시 다이얼로그 표시:

```
┌─────────────────────────────────────┐
│         프로그램 추가/수정           │
├─────────────────────────────────────┤
│                                     │
│  이름: [smartfw_2          ]        │
│                                     │
│  버전: [v.3.5.1            ]        │
│                                     │
│  업로드경로: [/download      ]      │
│                                     │
│  파일찾기: [찾아보기...]            │
│                                     │
│        [취소]      [저장]           │
└─────────────────────────────────────┘
```

**버튼 동작 조건:**
- `[삭제]` 버튼: 체크박스 선택이 없으면 아무 동작 안함, 선택 있으면 삭제 확인 다이얼로그 표시 후 삭제
- `[추가/수정]` 버튼: 체크박스 선택이 없으면 추가 모드로 다이얼로그 표시, 선택이 있으면 수정 모드

**삭제 동작:**
1. 체크박스로 선택된 프로그램이 없으면 아무 동작 안함
2. 선택된 프로그램이 있으면 삭제 확인 다이얼로그 표시 ("선택한 N개 프로그램을 삭제하시겠습니까?")
3. 확인 클릭 시 선택된 프로그램 삭제 후 테이블 갱신
4. 취소 클릭 시 다이얼로그 닫기

**다이얼로그 동작:**
1. `[추가/수정]` 버튼 클릭 → 다이얼로그 표시
2. 테이블에서 항목 선택 후 `[추가/수정]` → 선택된 항목 정보로 필드 채움 (수정 모드)
3. 테이블에서 선택 없이 `[추가/수정]` → 빈 필드로 다이얼로그 표시 (추가 모드)
4. `[찾아보기...]` 클릭 → `dialog.FileOpen` (tar 파일 필터)
5. `[저장]` 클릭 → 저장 후 다이얼로그 닫기
6. `[취소]` 클릭 → 다이얼로그 닫기

**구현 시 주의사항 (Fyne 다이얼로그 중첩):**

> ⚠️ `[찾아보기...]` 버튼 구현 시 **fyne-docs 스킬의 "다이얼로그 및 팝업 중첩" 섹션**을 반드시 참고할 것.

추가/수정 다이얼로그(Form Dialog)가 표시된 상태에서 `[찾아보기...]` 버튼을 클릭하면 파일 선택 다이얼로그(FileOpen)가 중첩되어 열립니다. Fyne에서 다이얼로그 중첩 시 발생할 수 있는 문제를 방지하기 위해:

1. **커스텀 다이얼로그 사용**: `dialog.ShowForm` 대신 `dialog.NewCustomWithoutButtons`로 직접 다이얼로그 구성
2. **부모 다이얼로그 참조 유지**: 파일 선택 완료 후 부모 다이얼로그가 정상 동작하도록 참조 관리
3. **콜백 체인 처리**: 파일 선택 콜백에서 선택된 파일 경로를 Entry에 반영

```go
// 참고: fyne-docs 스킬의 다이얼로그 중첩 패턴 적용 필요
// .claude/skills/fyne-docs/SKILL.md 참조
```

**검색 기능 동작:**
1. 검색어 입력 후 `[찾기]` 버튼 클릭 → 검색 실행
2. 검색 결과에 해당하는 프로그램 목록으로 테이블 갱신
3. 검색 대상 필드: 이름, 버전, 업로드 경로, 로컬파일 경로 (부분 일치)
4. 검색어가 빈 상태에서 `[찾기]` 클릭 → 전체 목록 표시
5. 검색 결과가 없을 경우 → "검색 결과가 없습니다." 다이얼로그 표시
6. 페이지네이션은 검색 결과에 맞게 자동 갱신

#### 3.3.3 장비 관리 탭

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│ [방화벽 룰 관리] [프로그램 관리] [장비 관리 X] [배포 이력]                             │
├─────────────────────────────────────────────────────────────────────────────────────┤
│  검색: [          ]  연결: 3  알수없음: 1  연결안됨: 0  🔄  [배포] [삭제] [추가/수정] │
├─────────────────────────────────────────────────────────────────────────────────────┤
│ 선택 │ 장비명          │ 서버 IP         │ 서버상태 │ 보고시간     │ 접속방식      │
├──────┼─────────────────┼─────────────────┼──────────┼──────────────┼───────────────┤
│  ☑   │ smart           │ 192.168.3.1     │ 정상     │ -            │ -             │
│  ☐   │ firewall-02     │ 192.168.3.2     │ 정상     │ -            │ PW            │
│  ☑   │ firewall-03     │ 192.168.3.3     │ 연결안됨 │ -            │ PPK           │
├─────────────────────────────────────────────────────────────────────────────────────┤
│                                << page 1/10 >>                                       │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

**상단 영역 구성:**
- 검색: 텍스트 입력 필드 (장비명/IP로 필터링)
- 연결 상태 표시: `연결: N` `알수없음: N` `연결안됨: N`
- 🔄 새로고침 버튼: 장비 상태 재조회
- `[배포]` 버튼: 선택된 장비에 방화벽 룰 배포 (배포 다이얼로그 표시)
- `[삭제]` 버튼: 선택된 장비 삭제
- `[추가/수정]` 버튼: 장비 추가/수정 다이얼로그 표시

**테이블 컬럼 설명:**

| 컬럼 | 설명 |
|------|------|
| 선택 | 체크박스 (다중 선택 가능) |
| 장비명 | 장비 이름 |
| 서버 IP | 장비 IP 주소 |
| 서버상태 | 정상/알수없음/연결안됨 |
| 보고시간 | 마지막 상태 확인 시간 |
| 접속방식 | SSH 접속 방식 (비밀번호/PPK) |

**버튼 동작 조건:**
- `[배포]` 버튼: 선택이 없으면 아무 동작 안함, 선택 있으면 배포 다이얼로그 표시
- `[삭제]` 버튼: 선택이 없으면 아무 동작 안함, 선택 있으면 삭제 확인 다이얼로그 표시 후 삭제
- `[추가/수정]` 버튼: 선택이 없으면 추가 모드, 선택 있으면 수정 모드

**삭제 동작:**
1. 체크박스로 선택된 장비가 없으면 아무 동작 안함
2. 선택된 장비가 있으면 삭제 확인 다이얼로그 표시 ("선택한 N개 장비를 삭제하시겠습니까?")
3. 확인 클릭 시 선택된 장비 삭제 후 테이블 갱신
4. 취소 클릭 시 다이얼로그 닫기

**테이블 row 더블클릭:**
- 상세보기 다이얼로그 표시

**배포 다이얼로그:**

`[배포]` 버튼 클릭 시 표시:

```
┌─────────────────────────────────────┐
│              배포                    │
├─────────────────────────────────────┤
│                                     │
│  선택한 IP 리스트:                   │
│  192.168.3.1, 192.168.3.3           │ ← 체크된 장비들
│                                     │
│  배포선택: [방화벽 룰 배포     ▼]   │
│                                     │
│  배포 리스트:                        │
│  ┌─────────────────────────────┐    │
│  │ ◉ v1.0.1                    │    │ ← 선택됨 (하이라이트)
│  │ ○ v1.0.0                    │    │
│  │ ○ v0.9.5                    │    │
│  └─────────────────────────────┘    │
│                                     │
│        [취소]      [배포]           │
└─────────────────────────────────────┘
```

| 항목 | 설명 |
|------|------|
| 선택한 IP 리스트 | 테이블에서 선택된 장비 IP |
| 배포선택 | 드롭다운 (방화벽 룰 배포 / 프로그램 배포) |
| 배포 리스트 | 배포선택에 따라 룰 목록 또는 프로그램 목록 표시 (단일 선택) |

**배포선택 옵션:**

| 옵션 | 설명 |
|------|------|
| 방화벽 룰 배포 | 방화벽 룰 관리 탭에 등록된 룰 목록 표시 |
| 프로그램 배포 | 프로그램 관리 탭에 등록된 프로그램 목록 표시 |

**배포 리스트 동작:**

- 라디오 버튼 또는 클릭 선택 방식으로 **단일 선택** 필수
- 선택된 항목은 **하이라이트 표시**되어 시각적으로 구분
- 배포선택 변경 시 배포 리스트 내용이 자동 갱신
- 배포 리스트에서 항목을 선택하지 않으면 `[배포]` 버튼 비활성화 또는 경고 표시
- **리스트 영역은 스크롤 가능**해야 함 (항목이 많을 경우 세로 스크롤)

**배포선택별 리스트 내용:**

| 배포선택 | 배포 리스트 내용 | 데이터 소스 |
|----------|------------------|-------------|
| 방화벽 룰 배포 | 룰 버전 목록 (예: v1.0.1, v1.0.0) | templates.json |
| 프로그램 배포 | 프로그램명 + 버전 (예: smartfw v2.1.0) | programs.json |

**상세보기 다이얼로그:**

테이블 row 더블클릭 시 표시:

```
┌─────────────────────────────────────┐
│            상세보기                  │
├─────────────────────────────────────┤
│                                     │
│  장비명:                             │
│  IP:                                │
│  연결상태:                           │
│                                     │
│  접속정보: PW or KEY                 │
│  ID:                                │
│  PW:                                │
│  (접속정보가 키면 PPK 경로)          │
│                                     │
│  배포정보:                           │
│  방화벽 룰셋 버전(프로그램 버전)      │
│                                     │
│            [확인]                    │
└─────────────────────────────────────┘
```

**장비 추가/수정 다이얼로그:**

`[추가/수정]` 버튼 클릭 시 표시:

```
┌─────────────────────────────────────┐
│          장비 추가/수정              │
├─────────────────────────────────────┤
│                                     │
│  장비명: [                    ]      │
│                                     │
│  서버 IP: [                  ]      │
│                                     │
│  접속선택: [PW              ▼]      │
│  ┌─────────────────────────────┐    │
│  │ SSH ID: [root             ] │    │ ← 접속선택이 PW일 때
│  │ 비밀번호: [               ] │    │
│  └─────────────────────────────┘    │
│  ┌─────────────────────────────┐    │
│  │ PPK: [찾아보기...]         │    │ ← 접속선택이 PPK일 때
│  └─────────────────────────────┘    │
│                                     │
│        [취소]      [저장]           │
└─────────────────────────────────────┘
```

| 필드 | 설명 |
|------|------|
| 장비명 | 장비 이름 입력 |
| 서버 IP | 장비 IP 주소 입력 |
| 접속선택 | 드롭다운 (PW / PPK) |
| SSH ID | 접속선택이 PW일 때 표시, SSH 접속 ID |
| 비밀번호 | 접속선택이 PW일 때 표시, SSH 비밀번호 |
| PPK | 접속선택이 PPK일 때 표시, PPK 파일 경로 선택 |

**접속선택에 따른 동적 필드:**
- `PW` 선택 시: SSH ID + 비밀번호 필드 표시
- `PPK` 선택 시: PPK 파일 찾기 버튼 표시

**구현 시 주의사항 (Fyne 다이얼로그 중첩):**

> ⚠️ `[찾아보기...]` 버튼 구현 시 **fyne-docs 스킬의 "다이얼로그 및 팝업 중첩" 섹션**을 반드시 참고할 것.

장비 추가/수정 다이얼로그가 표시된 상태에서 PPK `[찾아보기...]` 버튼을 클릭하면 파일 선택 다이얼로그가 중첩되어 열립니다. 프로그램 관리 탭의 다이얼로그 중첩과 동일한 패턴 적용 필요.

#### 3.3.4 배포 이력 탭

기존 배포 이력 탭을 확장하여 방화벽 룰 배포와 프로그램 업데이트 이력을 통합 관리합니다.

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│ [방화벽 룰 관리] [프로그램 관리] [장비 관리] [배포 이력]                                 │
├─────────────────────────────────────────────────────────────────────────────────────┤
│  배포 이력 X                                                                         │
├─────────────────────────────────────────────────────────────────────────────────────┤
│  [유형선택 ▼]  검색: [                    ] [찾기]                    [선택삭제]    │
├─────────────────────────────────────────────────────────────────────────────────────┤
│ 선택 │ 시간                │ 장비명      │ 장비 IP              │ 유형         │ 버전   │ 결과 │
├──────┼─────────────────────┼─────────────┼──────────────────────┼──────────────┼────────┼──────┤
│  ☑  │ 2026-01-13 15:28:32 │ smartfw_1   │ 192.168.31.20:9100   │ 방화벽 룰 셋 │ v2.0.0 │ 성공 │
│  ☐  │ 2026-01-13 14:30:00 │ smartfw_2   │ 192.168.31.20:9200   │ 프로그램     │ v3.5.1 │ 성공 │
│  ☐  │ 2026-01-12 10:00:00 │ firewall-03 │ 192.168.1.101        │ 방화벽 룰 셋 │ v1.0   │ 실패 │
│  ☐  │ 2026-01-11 09:00:00 │ firewall-04 │ 192.168.1.102        │ 프로그램     │ v2.1.0 │ 실패 │
├─────────────────────────────────────────────────────────────────────────────────────┤
│                                << page 1/10 >>                                       │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

**유형선택 드롭다운:**
- **기본값**: 프로그램
- **선택 옵션**:
  - `프로그램` - 프로그램 업데이트 이력만 표시
  - `방화벽 룰` - 방화벽 룰 배포 이력만 표시
- 선택 변경 시 테이블 내용이 해당 유형의 이력만 표시되도록 자동 갱신

**테이블 컬럼 설명:**

| 컬럼 | 설명 |
|------|------|
| 선택 | 체크박스 (다중 선택 가능) |
| 시간 | 배포/업데이트 실행 시간 (YYYY-MM-DD HH:MM:SS) |
| 장비명 | 대상 장비 이름 |
| 장비 IP | 대상 장비 IP 주소 (포트 포함) |
| 유형 | "방화벽 룰 셋" 또는 "프로그램" |
| 버전 | 배포된 방화벽 룰 버전 또는 프로그램 버전 |
| 결과 | 성공/실패 |

**검색 기능 동작:**
1. 검색어 입력 후 `[찾기]` 버튼 클릭 → 검색 실행
2. 검색 결과에 해당하는 이력 목록으로 테이블 갱신
3. 검색 대상 필드: 장비명, 장비 IP, 버전 (부분 일치)
4. 검색어가 빈 상태에서 `[찾기]` 클릭 → 전체 목록 표시 (유형선택 필터 적용)
5. 검색 결과가 없을 경우 → "검색 결과가 없습니다." 다이얼로그 표시
6. 페이지네이션은 검색 결과에 맞게 자동 갱신

**선택삭제 동작:**
1. 체크박스로 선택된 이력이 없으면 아무 동작 안함
2. 선택된 이력이 있으면 삭제 확인 다이얼로그 표시 ("선택한 N개 이력을 삭제하시겠습니까?")
3. 확인 클릭 시 선택된 이력 삭제 후 테이블 갱신
4. 취소 클릭 시 다이얼로그 닫기

#### 3.3.5 업데이트/배포 진행 다이얼로그

**업데이트 진행 상태 표시:**
```
┌─────────────────────────────────────────┐
│         업데이트 진행 중                │
├─────────────────────────────────────────┤
│                                         │
│  192.168.1.100                          │
│  [████████████████████████░░] 80%      │
│  상태: 파일 업로드 중...               │
│                                         │
│  192.168.1.101                          │
│  [██████░░░░░░░░░░░░░░░░░░░░] 25%      │
│  상태: SSH 연결 중...                  │
│                                         │
│                [취소]                   │
└─────────────────────────────────────────┘
```

**업데이트 완료 결과:**
```
┌─────────────────────────────────────────┐
│         업데이트 완료                   │
├─────────────────────────────────────────┤
│                                         │
│  ✓ 192.168.1.100: 성공                 │
│    smartfw 2.0.5 → 2.1.0               │
│                                         │
│  ✓ 192.168.1.101: 성공                 │
│    smartfw 2.0.3 → 2.1.0               │
│                                         │
│  ✗ 192.168.1.102: 실패                 │
│    오류: SSH 연결 시간 초과            │
│                                         │
│                [확인]                   │
└─────────────────────────────────────────┘
```

---

## 4. 데이터 모델

### 4.1 UpdateProgram (업데이트 프로그램)

```go
type UpdateProgram struct {
    ProcessName       string `json:"process_name"`        // 업데이트 프로세스 이름
    ProcessFilePath   string `json:"process_file_path"`   // 업데이트 할 파일 저장 경로 (local)
    ProcessUploadPath string `json:"process_upload_path"` // 서버에 업로드 할 경로 (서버 쪽 저장 경로)
    ProcessVersion    string `json:"process_version"`     // 프로세스 버전
    ProcessCreatedAt  string `json:"process_created_at"`  // 추가(수정) 일시
}
```

**저장 위치:** `programs.json`

**JSON 예시:**
```json
[
  {
    "process_name": "smartfw",
    "process_file_path": "C:\\Updates\\smartfw-v2.1.0.tar",
    "process_upload_path": "/download/",
    "process_version": "2.1.0",
    "process_created_at": "2026-01-12 14:00:00"
  },
  {
    "process_name": "smartway",
    "process_file_path": "C:\\Updates\\smartway-v1.5.0.tar",
    "process_upload_path": "/download/",
    "process_version": "1.5.0",
    "process_created_at": "2026-01-12 15:30:00"
  }
]
```

### 4.2 Firewall (장비 - 기존 모델 확장)

기존 Firewall 모델을 확장하여 SSH 접속 정보와 프로그램 버전 정보를 추가합니다.

```go
// ProcessInfo 설치된 프로그램 정보
type ProcessInfo struct {
    Name    string `json:"name"`    // 프로그램 이름
    Version string `json:"version"` // 설치된 버전
}

type Firewall struct {
    // 기존 필드
    Index        int           `json:"index"`
    DeviceName   string        `json:"device_name"`   // 장비명
    ServerStatus string        `json:"serverStatus"`  // running, stop, -
    DeployStatus string        `json:"deployStatus"`  // success, fail, -
    Version      string        `json:"version"`       // 방화벽 룰 버전
    DeployResult *DeployResult `json:"deployResult,omitempty"`

    // 신규 필드 (SSH 인증) - 2.2.1 장비 정보 테이블과 일치
    DeviceIP     string        `json:"device_ip"`     // 장비 IP
    DeviceID     string        `json:"device_id"`     // 접속 ID
    DevicePW     string        `json:"device_pw"`     // 접속 비밀번호
    DevicePPK    string        `json:"device_ppk"`    // 접속 PPK 파일 경로

    // 신규 필드 (프로그램 버전) - 동적 목록
    Processes     []ProcessInfo `json:"processes"`    // 설치된 프로그램 목록 (GET /device-report 응답과 동일 구조)

    // 신규 필드 (상태 정보)
    LastCheckedAt string        `json:"lastCheckedAt"` // 마지막 상태 확인 시간
}
```

**저장 위치:** `firewalls.json`

**JSON 예시:**
```json
[
  {
    "index": 1,
    "device_name": "firewall-01",
    "device_ip": "192.168.1.100",
    "device_id": "root",
    "device_pw": "",
    "device_ppk": "C:\\Keys\\server.ppk",
    "serverStatus": "running",
    "deployStatus": "success",
    "version": "v1.0.0",
    "processes": [
      {"name": "smartfw", "version": "2.1.0"},
      {"name": "smartway", "version": "1.5.0"}
    ],
    "lastCheckedAt": "2026-01-13 15:30:00"
  },
  {
    "index": 2,
    "device_name": "firewall-02",
    "device_ip": "192.168.1.101",
    "device_id": "root",
    "device_pw": "password123",
    "device_ppk": "",
    "serverStatus": "stop",
    "deployStatus": "-",
    "version": "-",
    "processes": [],
    "lastCheckedAt": "-"
  }
]
```

**하위 호환성:**
- 기존 JSON에 신규 필드가 없는 경우 빈 값 또는 "-"로 처리
- `processes` 필드가 없으면 빈 배열 `[]`로 처리

---

### 4.3 DeployHistory (통합 이력 - 기존 모델 확장)

기존 DeployHistory 모델을 확장하여 방화벽 룰 배포와 프로그램 업데이트 이력을 통합 관리합니다.

```go
type DeployHistory struct {
    ID          int           `json:"id"`
    Timestamp   time.Time     `json:"timestamp"`
    DeviceName  string        `json:"deviceName"`   // 장비명
    DeviceIP    string        `json:"deviceIp"`     // 장비 IP
    Type        string        `json:"type"`         // "firewall" 또는 "program"
    Version     string        `json:"version"`      // 배포된 버전 (룰/프로그램 공통)
    Status      string        `json:"status"`       // "success", "fail"

    // 방화벽 룰 배포 상세 (Type="firewall"일 때만 사용)
    Results     []ResultInfo  `json:"results,omitempty"`

    // 프로그램 업데이트 상세 (Type="program"일 때만 사용)
    ProgramName string        `json:"programName,omitempty"`
    Message     string        `json:"message,omitempty"`
}
```

**이력 유형 상수:**
```go
const (
    HistoryTypeFirewall = "firewall"  // 방화벽 룰 배포
    HistoryTypeProgram  = "program"   // 프로그램 업데이트
)
```

**설계 원칙:**
- 기존 DeployHistory 구조체를 확장하여 하위 호환성 유지
- Type 필드로 이력 유형 구분
- 각 유형에 맞는 필드만 사용 (`omitempty`로 미사용 필드 생략)
- 기존 이력 데이터에 `type` 필드가 없으면 `"firewall"`으로 간주

**JSON 예시 (history.json):**

```json
[
  {
    "id": 1,
    "timestamp": "2026-01-13 15:28:32",
    "deviceName": "smartfw_1",
    "deviceIp": "192.168.31.20:9100",
    "type": "firewall",
    "version": "v2.0.0",
    "status": "success",
    "results": [
      {
        "rule": "agent -m=insert -c=INPUT -p=tcp --dport=8080 -a=accept",
        "text": "insert|c_123456|input|accept|ANY|tcp|ANY|8080|ANY|ANY",
        "status": "OK",
        "reason": "OK"
      }
    ]
  },
  {
    "id": 2,
    "timestamp": "2026-01-13 14:30:00",
    "deviceName": "smartfw_2",
    "deviceIp": "192.168.31.20:9200",
    "type": "program",
    "version": "3.5.1",
    "status": "success",
    "programName": "smartfw",
    "message": "Update completed successfully"
  },
  {
    "id": 3,
    "timestamp": "2026-01-12 10:00:00",
    "deviceName": "firewall-03",
    "deviceIp": "192.168.1.101",
    "type": "firewall",
    "version": "v1.0",
    "status": "fail",
    "results": []
  },
  {
    "id": 4,
    "timestamp": "2026-01-11 09:00:00",
    "deviceName": "firewall-04",
    "deviceIp": "192.168.1.102",
    "type": "program",
    "version": "2.1.0",
    "status": "fail",
    "programName": "smartway",
    "message": "SSH connection timeout"
  }
]
```

**하위 호환성:**
- 기존 이력 JSON에 `type` 필드가 없는 경우 → `"firewall"`으로 처리
- 기존 이력 JSON에 `deviceName` 필드가 없는 경우 → `deviceIp`를 `deviceName`으로 사용
- 기존 `templateVersion` 필드 → `version`으로 매핑
- 기존 JSON 파일 마이그레이션 불필요

---

## 5. API 명세

### 5.1 장비 측 API (smartfw/smartway 서버)

#### GET /agent/respCheck

서버 연결 상태 확인 (기존)

**응답:** HTTP 200 OK (연결 성공)

---

#### GET /device-report

프로그램 버전 정보 조회 (신규)

**응답:**
```json
{
  "processes": [
    {"name": "smartway", "version": "1.0.0"},
    {"name": "smartfw", "version": "2.0.0"}
  ]
}
```

| 필드 | 타입 | 설명 |
|------|------|------|
| processes | array | 프로세스 목록 |
| processes[].name | string | 프로세스 이름 |
| processes[].version | string | 설치된 버전 |

---

#### POST /agent/firewall-deploy

방화벽 룰 배포 요청

**요청:**
```json
{
  "configInfo": "방화벽 룰 내용"
}
```

| 필드 | 타입 | 설명 |
|------|------|------|
| configInfo | string | 배포할 방화벽 룰 내용 |

**응답:**
```json
{
  "status": "success",
  "message": "배포 완료"
}
```

#### POST /program-update

프로그램 업데이트 실행 요청

**요청:**
```json
{
  "file_path": "/download/smartfw.tar",
  "execute_command": "tar -xvf /download/smartfw.tar -C /opt/smartfw && systemctl restart smartfw",
  "process_name": "smartfw",
  "process_version": "v1.0.0"
}
```

| 필드 | 타입 | 설명 |
|------|------|------|
| file_path | string | SSH로 업로드된 tar 파일의 서버 내 경로 |
| execute_command | string | 실행할 쉘 명령어 (압축 해제, 프로세스 재시작 등) |
| process_name | string | 프로세스 이름 |
| process_version | string | 프로세스 버전 |

**응답:**
```json
{
  "status": "success",
  "message": "완료 메시지",
  "installed_version": "1.0.0"
}
```

| 필드 | 타입 | 설명 |
|------|------|------|
| status | string | 결과 상태 ("success" / "failed") |
| message | string | 상세 메시지 |
| installed_version | string | 설치된 버전 |

---

### 5.2 통신 포트

- 1차 개발: HTTP 사용 (예: 8080)
- 향후: HTTPS 전환 검토
- 포트 번호는 설정 파일 또는 상수로 관리

---

## 6. 보안 고려사항

### 6.1 SSH 인증 정보 보호

- 비밀번호는 암호화하여 저장 (AES-256 등)
- PPK 파일 경로만 저장 (파일 자체는 사용자 관리)
- 메모리에서 인증 정보 사용 후 즉시 삭제

### 6.2 HTTP/HTTPS 통신

- 1차 개발: HTTP로 구현
- 향후 HTTPS 전환 시 고려사항:
  - 자체 서명 인증서 허용 옵션 (내부망 환경)
  - 프로덕션 환경에서는 정식 인증서 권장

### 6.3 권한 관리

- SSH 접속은 최소 권한 계정 사용 권장
- root 계정 사용 시 주의 필요

---

## 7. 시퀀스 다이어그램

### 7.1 전체 프로세스 흐름

```
┌─────────┐                              ┌──────────────┐
│ 유틸리티 │                              │    Agent     │
│ (Fyne)  │                              │  (운영장비)   │
└────┬────┘                              └──────┬───────┘
     │                                          │
     │ 1. SSH 접속 (ID + PPK 또는 Password)     │
     │─────────────────────────────────────────>│
     │                                          │
     │ 2. 파일 업로드 (SCP/SFTP)                │
     │    smartway-v1.0.0.tar                   │
     │    → /download/ (process_upload_path)   │
     │─────────────────────────────────────────>│
     │                                          │
     │ 3. HTTP POST /program-update          │
     │    {                                     │
     │      "file_path": "/download/...",       │
     │      "execute_command": "실행 쉘 명령어" │
     │    }                                     │
     │─────────────────────────────────────────>│
     │                                          │
     │                                          │  Agent 내부 처리
     │                                          │
     │ 4. 응답 (성공/실패)                       │
     │<─────────────────────────────────────────│
     │                                          │
     │ 5. 주기적 상태 수집                       │
     │    HTTP GET /agent/respCheck            │
     │─────────────────────────────────────────>│
     │                                          │
     │    {                                     │
     │      "processes": [                      │
     │        {"name":"smartway","version":"1.0.0"},
     │        {"name":"smartfw","version":"2.0.0"}
     │      ]                                   │
     │    }                                     │
     │<─────────────────────────────────────────│
     │                                          │
```

---

## 8. 구현 우선순위

### Phase 1: 기본 구조 (필수)
1. 데이터 모델 정의 (UpdateDevice, UpdateProgram, UpdateHistory)
2. 장비 등록/관리 UI
3. SSH 연결 테스트 기능

### Phase 2: 핵심 기능 (필수)
4. 프로그램 등록/관리 UI
5. SSH 파일 업로드 (SCP/SFTP)
6. 업데이트 API 호출
7. 업데이트 실행 UI

### Phase 3: 부가 기능
8. 주기적 상태 수집
9. 업데이트 이력 관리
10. 다중 장비 일괄 업데이트

### Phase 4: 고도화
11. 업데이트 롤백 기능
12. 스케줄 업데이트
13. 업데이트 전 백업 기능

---

## 9. 참고 사항

### 9.1 Go SSH 라이브러리

```go
import "golang.org/x/crypto/ssh"
```

- 비밀번호 인증: `ssh.Password(password)`
- 키 인증: `ssh.PublicKeys(signer)`
- PPK 파일 파싱 필요 (OpenSSH 형식 변환 또는 별도 라이브러리)

### 9.2 Go SCP/SFTP 라이브러리

```go
import "github.com/pkg/sftp"
```

### 9.3 Windows 환경 고려사항

- PPK 파일은 PuTTY 형식 → OpenSSH 형식 변환 필요할 수 있음
- 또는 PPK 직접 파싱 라이브러리 사용

---

## 10. 용어 정의

| 용어 | 설명 |
|------|------|
| smartfw | Smart Firewall - 방화벽 프로그램 |
| smartway | Smart Gateway - 게이트웨이 프로그램 |
| PPK | PuTTY Private Key - PuTTY에서 사용하는 키 형식 |
| SCP | Secure Copy Protocol - SSH 기반 파일 복사 |
| SFTP | SSH File Transfer Protocol - SSH 기반 파일 전송 |

---

## 부록 A: 원본 프로토콜 문서 요약

### 장비 정보 테이블 (device_info)
- device_name: 장비 식별 이름
- device_ip: 장비 IP 주소
- device_id: SSH 접속 아이디
- device_pw: SSH 접속 비밀번호
- device_ppk: PPK 키 파일 경로

### 프로세스 정보 테이블 (process_info)
- process_name: 프로세스/프로그램 이름
- process_file_path: 로컬 업데이트 파일 경로
- process_upload_path: 원격 서버 업로드 경로
- process_version: 프로그램 버전

### API 엔드포인트
- GET /agent/respCheck: 장비 상태 조회
- POST /program-update: 프로그램 업데이트 요청

### 주요 흐름
1. 장비 등록 (SSH 인증 정보 포함)
2. 업데이트 파일 등록
3. SSH 연결 → 파일 업로드 → HTTP API 호출
4. 주기적 장비 상태 수집

---

## 부록 B: Agent 구현 참고사항

> **Note:** 이 섹션은 Agent 측 구현 시 참고 사항입니다. Fyne 유틸리티 구현과는 별개입니다.

**Agent 처리 흐름:**
1. POST /program-update 요청 수신
2. 스크립트 실행 (압축 해제)
   - 압축 해제 위치는 Agent에서 설정 가능해야 함
   - `/root` 고정이 아닌 유연한 구조 필요
3. 요청받은 버전으로 프로세스 버전 파일 생성
4. 프로세스 재시작
5. 결과 응답 반환
