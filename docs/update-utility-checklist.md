# 체크리스트: smartfw/smartway 업데이트 유틸리티

**상태**: 🔄 진행 중
**시작일**: 2026-01-12
**최종 수정**: 2026-01-15
**대상 프로젝트**: fms_fyne
**관련 문서**: [update-utility-prd.md](update-utility-prd.md), [fyne-layout-update.md](fyne-layout-update.md)

---

**⚠️ 중요 진행 규칙**:

1. ✅ 각 Phase 완료 후 Quality Gate 검증 필수
2. 🧪 **Phase 1-2 완료 전까지 패키지 업데이트 기능 구현 금지**
3. ⚠️ **기존 방화벽 룰 배포 기능이 새 UI에서 정상 동작해야 다음 Phase 진행**
4. 📅 각 Phase 완료 시 "최종 수정" 날짜 업데이트
5. 📝 이슈 발생 시 Notes 섹션에 기록

⛔ **Quality Gate 미통과 시 다음 Phase 진행 금지**

---

## 📋 개요

### 목표
1. 기존 FMS UI를 DocTabs + 왼쪽 메뉴 구조로 마이그레이션
2. 패키지 관리 탭 신규 구현
3. 장비관리 탭 확장 (SSH 인증 정보, 패키지 버전)
4. 배포 이력 탭 통합 (방화벽 룰 + 패키지 업데이트)
5. 패키지 업데이트 기능 구현 (SSH/SFTP + HTTP)

### 성공 기준
- [ ] 기존 방화벽 룰 배포 기능이 새 UI에서 정상 동작
- [ ] 패키지 업데이트 기능 정상 동작 (SSH + SFTP + HTTP)
- [ ] 배포/업데이트 이력 통합 관리
- [ ] SSH 키(PPK) 기반 인증으로 파일 업로드

### 아키텍처 (fms_fyne)

```
┌─────────────────────────────────────────────────────────────┐
│                    UI Layer (Fyne)                          │
│  ┌──────────────┬──────────────────────────────────────┐   │
│  │  왼쪽 메뉴    │  DocTabs (동적 탭 관리)               │   │
│  │  - 방화벽 룰  │  TemplateTab, ProgramTab,            │   │
│  │  - 패키지   │  DeviceTab, HistoryTab               │   │
│  │  - 장비 관리  │                                       │   │
│  │  - 배포 이력  │                                       │   │
│  └──────────────┴──────────────────────────────────────┘   │
├─────────────────────────────────────────────────────────────┤
│                    Component Layer                          │
│  (PagedTable, NewCustomButton, Form, Dialog 등)             │
├─────────────────────────────────────────────────────────────┤
│                    Service Layer (신규)                      │
│  (SSHService, UpdateService)                                │
├─────────────────────────────────────────────────────────────┤
│                    HTTP Layer                               │
│  (client.go - HTTP 통신)                                    │
├─────────────────────────────────────────────────────────────┤
│                    Storage Layer                            │
│  (JSONStore - 파일 저장/조회)                                │
├─────────────────────────────────────────────────────────────┤
│                    Model Layer                              │
│  (Firewall, ProcessInfo, DeployHistory 등)                  │
└─────────────────────────────────────────────────────────────┘
```

### 통신 흐름 (패키지 업데이트)

```
┌─────────────┐         SSH/SFTP          ┌──────────────┐
│  FMS 유틸   │ ──────────────────────────>│   Agent 서버  │
│  (Fyne)     │   1. SSH 연결              │  (운영장비)   │
│             │   2. 파일 업로드 (SFTP)     │              │
│             │      → /download/          │              │
│             │                            │              │
│             │         HTTP POST          │              │
│             │ ──────────────────────────>│              │
│             │   3. /program-update       │   Agent가    │
│             │                            │   압축해제,   │
│             │   4. 응답 (성공/실패)       │   설치 실행   │
│             │<──────────────────────────│              │
└─────────────┘                            └──────────────┘
```

---

## 🚀 구현 Phase

---

## Phase 0: 공통 컴포넌트 준비 ✅

**목표**: 공통 테이블 컴포넌트 구현
**상태**: ✅ 완료

### 0.1 PagedTable 컴포넌트

- [x] PagedTable 구조체 정의
  - 페이지네이션 지원
  - 체크박스 선택
  - 단일 클릭/더블클릭 콜백
  - 파일: `fms_fyne/internal/ui/component/table.go`

- [x] PagedTableConfig 정의
  ```go
  type PagedTableConfig struct {
      Columns          []ColumnDef
      PageSize         int
      OnCellUpdate     func(row int, col int, cell fyne.CanvasObject)
      OnRowSelected    func(row int)      // 단일 클릭/Enter
      OnRowDoubleClick func(row int)      // 더블클릭
      OnCheckChange    func(row int, checked bool)
  }
  ```

- [x] 주요 메서드 구현
  - SetData(totalItems int)
  - GetCheckedRows() []int
  - SelectAll() / DeselectAll()
  - NextPage() / PrevPage() / GoToPage()

### 0.2 NewCustomButton 컴포넌트

- [x] NewCustomButton 함수 (기존)
  - 파일: `fms_fyne/internal/ui/component/button.go`
  - 사용법:
  ```go
  // NewCustomButton(label, icon, iconColor, bgColor, onTap, margin...)
  //
  // - label: 버튼 텍스트
  // - icon: 아이콘 리소스 (nil이면 텍스트만 표시)
  // - iconColor: 아이콘/텍스트 색상 (nil이면 자동 결정)
  // - bgColor: 배경색 (nil이면 투명 - 텍스트 버튼)
  // - onTap: 클릭 콜백
  // - margin: 외부 여백 (상, 하, 좌, 우 순서, 생략 시 0)

  // 예시 1: 빨간 텍스트 버튼 (투명 배경)
  deleteBtn := component.NewCustomButton("삭제", nil, themes.Colors["red"], nil, func() {
      // 삭제 로직
  })

  // 예시 2: 아이콘 + 텍스트 버튼 (파란 배경)
  saveBtn := component.NewCustomButton("저장", theme.DocumentSaveIcon(),
      color.White, themes.Colors["blue"], func() {
      // 저장 로직
  })

  // 예시 3: 아이콘만 있는 버튼
  refreshBtn := component.NewCustomButton("", theme.ViewRefreshIcon(),
      themes.Colors["gray"], nil, func() {
      // 새로고침 로직
  })
  ```

- [x] 버튼 사용 가이드라인
  - **모든 버튼은 `component.NewCustomButton` 사용 권장**
  - `widget.NewButton` 대신 `NewCustomButton` 사용으로 일관된 스타일 유지
  - 위험한 동작 (삭제 등): 빨간 텍스트 (`themes.Colors["red"]`)
  - 주요 동작 (저장, 배포 등): 파란 배경 (`themes.Colors["blue"]`)
  - 보조 동작 (취소, 닫기 등): 투명 배경 (bgColor = nil)

---

## Phase 1: UI 레이아웃 마이그레이션

**목표**: AppTabs → DocTabs + 왼쪽 메뉴 구조로 변경
**상태**: ⏳ 대기
**참고 문서**: [fyne-layout-update.md](fyne-layout-update.md)

### 1.1 MainUI 구조 변경

#### 1.1.1 app.go 수정
- [ ] MainUI 구조체 변경
  ```go
  type MainUI struct {
      tabs        *container.DocTabs  // AppTabs → DocTabs
      leftMenu    *fyne.Container     // 신규

      // 탭 아이템 참조 (동적 추가/제거용)
      templateTabItem *container.TabItem
      programTabItem  *container.TabItem  // 신규
      deviceTabItem   *container.TabItem
      historyTabItem  *container.TabItem
  }
  ```
  - 파일: `fms_fyne/internal/ui/app.go`

- [ ] createLeftMenu() 함수 구현
  - 방화벽  관리 버튼
  - 패키지 관리 버튼
  - 장비 관리 버튼
  - 배포 이력 버튼
  - 최소 너비 150px (minWidthLayout 커스텀 레이아웃)

- [ ] openTab() 함수 구현
  - 이미 열린 탭이면 선택
  - 새 탭이면 추가 후 선택

- [ ] 전체 레이아웃 변경
  ```go
  // 왼쪽: 메뉴 + 구분선 (고정 너비)
  // 중앙: DocTabs (자동 확장)
  container.NewBorder(nil, nil, leftMenuWithSeparator, nil, tabs)
  ```

### 1.2 용어 변경

- [ ] template_tab.go 주석/UI 텍스트 변경
  - "템플릿" → "방화벽 룰" 또는 "룰"

- [ ] rule_builder.go 주석 변경
  - "규칙 빌더" → "룰 빌더"

### 1.3 Quality Gate ✋

**빌드 & 테스트**:
- [ ] `go build` 성공 (fms_fyne 디렉토리)
- [ ] 기존 테스트 통과: `go test ./...`

**기능 검증** (수동 테스트):
- [ ] 왼쪽 메뉴 4개 버튼 표시
- [ ] 메뉴 클릭 시 해당 탭 열림
- [ ] 탭 닫기(×) 버튼 동작
- [ ] 동일 탭 재클릭 시 기존 탭 선택 (중복 생성 안함)
- [ ] **방화벽 관리 탭 기존 기능 정상 동작** ⭐
- [ ] **장비 관리 탭 기존 기능 정상 동작** ⭐
- [ ] **배포 이력 탭 기존 기능 정상 동작** ⭐

---

## Phase 2: 패키지 관리 탭 구현

**목표**: 패키지 관리 탭 신규 구현
**상태**: ⏳ 대기
**선행 조건**: Phase 1 Quality Gate 통과

### 2.1 Model 추가

#### 2.1.1 ProcessInfo 모델 생성
- [ ] ProcessInfo 구조체 정의
  ```go
  type ProcessInfo struct {
      ID              int    `json:"id"`
      ProcessName     string `json:"process_name"`
      ProcessFilePath string `json:"process_file_path"`   // 로컬 파일 경로
      ProcessUploadPath string `json:"process_upload_path"` // 서버 업로드 경로
      ProcessVersion  string `json:"process_version"`
      ProcessCreatedAt string `json:"process_created_at"`
  }
  ```
  - 파일: `fms_fyne/internal/model/process_info.go` (신규)

### 2.2 Storage 확장

#### 2.2.1 JSONStore 확장
- [ ] processes.json 저장/로드 메서드 추가
  - GetAllProcesses() ([]*ProcessInfo, error)
  - SaveProcess(*ProcessInfo) error
  - DeleteProcess(id int) error
  - 파일: `fms_fyne/internal/storage/json_store.go`

### 2.3 ProgramTab UI 구현

#### 2.3.1 ProgramTab 구조체 생성
- [ ] ProgramTab 구조체 및 NewProgramTab 함수
  - 파일: `fms_fyne/internal/ui/program_tab.go` (신규)

#### 2.3.2 상단 영역 UI
- [ ] 검색 입력 필드 (widget.Entry)
- [ ] [찾기] 버튼
- [ ] [삭제] 버튼
- [ ] [추가/수정] 버튼

#### 2.3.3 패키지 목록 테이블 (PagedTable 사용)
- [ ] 컬럼 정의
  - 선택 (체크박스)
  - 이름
  - 버전
  - 업로드 경로
  - 로컬파일 경로
  - 추가(수정)시간
- [ ] 페이지네이션 (10개/페이지)
- [ ] 더블클릭 시 수정 다이얼로그 표시

#### 2.3.4 추가/수정 다이얼로그
- [ ] dialog.NewCustom 사용
  - 이름 입력 (widget.Entry)
  - 버전 입력 (widget.Entry)
  - 업로드 경로 입력 (widget.Entry, 기본값: /download/)
  - 로컬 파일 경로 + [찾아보기...] 버튼
- [ ] 다이얼로그 중첩 처리 (Hide/Show 패턴)
  - fyne-docs 스킬 "다이얼로그 및 팝업 중첩" 섹션 참고

#### 2.3.5 검색 기능
- [ ] 검색 대상: 이름, 버전, 업로드 경로, 로컬파일 경로 (부분 일치)
- [ ] 검색 결과 없음 시 다이얼로그 표시

### 2.4 MainUI 수정

- [ ] ProgramTab import 및 인스턴스 생성
- [ ] programTabItem 생성 및 openTab 연동
- [ ] 탭 간 참조 설정 (DeviceTab → ProgramTab)

### 2.5 Quality Gate ✋

**빌드 & 테스트**:
- [ ] `go build` 성공
- [ ] 기존 테스트 통과: `go test ./...`

**기능 검증** (수동 테스트):
- [ ] 패키지 관리 탭 표시
- [ ] 패키지 추가/수정/삭제 정상 동작
- [ ] 검색 기능 정상 동작
- [ ] 페이지네이션 정상 동작
- [ ] 파일 찾아보기 다이얼로그 정상 동작

**데이터 검증**:
- [ ] config/processes.json 파일 정상 생성/저장

**기존 기능 확인** ⭐:
- [ ] 방화벽 관리 탭 정상 동작
- [ ] 장비 관리 탭 정상 동작
- [ ] 배포 이력 탭 정상 동작

---

## Phase 3: 장비관리 탭 확장

**목표**: 장비관리 탭에 SSH 인증 정보, 패키지 버전 추가
**상태**: ⏳ 대기
**선행 조건**: Phase 2 Quality Gate 통과

### 3.1 Model 변경

#### 3.1.1 Firewall 모델 확장
- [ ] 신규 필드 추가
  ```go
  type Firewall struct {
      // 기존 필드
      Index        int           `json:"index"`
      DeviceName   string        `json:"device_name"`
      ServerStatus string        `json:"serverStatus"`
      DeployStatus string        `json:"deployStatus"`
      Version      string        `json:"version"`       // 방화벽 룰 버전
      DeployResult *DeployResult `json:"deployResult,omitempty"`

      // 신규 필드 (SSH 인증)
      DeviceIP     string        `json:"device_ip"`
      DeviceID     string        `json:"device_id"`     // SSH 아이디
      DevicePW     string        `json:"device_pw"`     // 비밀번호
      DevicePPK    string        `json:"device_ppk"`    // PPK 파일 경로

      // 신규 필드 (패키지 버전)
      Processes    []ProcessInfo `json:"processes"`     // 설치된 패키지 목록

      // 신규 필드 (상태 정보)
      LastCheckedAt string       `json:"lastCheckedAt"` // 마지막 상태 확인 시간
  }
  ```
  - 파일: `fms_fyne/internal/model/firewall.go`

- [ ] getAuthType() 헬퍼 함수 추가 (접속방식 표시용)
  ```go
  func (f *Firewall) GetAuthType() string {
      if f.DevicePPK != "" { return "PPK" }
      if f.DevicePW != "" { return "비밀번호" }
      return "-"
  }
  ```

### 3.2 DeviceTab UI 변경

#### 3.2.1 상단 영역 변경
- [ ] 기존 UI 변경
  - 기존: templateSelect, 버튼들
  - 변경: 검색 입력, 상태 카운터, 새로고침, [배포] [삭제] [추가/수정]

- [ ] 상태 카운터 표시
  - "연결: N  알수없음: N  연결안됨: N"

#### 3.2.2 테이블 컬럼 변경 (PagedTable 사용)
- [ ] 컬럼 정의
  - 선택 (체크박스)
  - 장비명
  - 서버 IP
  - 서버상태
  - 보고시간
  - 접속방식

#### 3.2.3 장비 추가/수정 다이얼로그
- [ ] 다이얼로그 필드
  - 장비명 (widget.Entry)
  - 서버 IP (widget.Entry)
  - SSH 아이디 (widget.Entry)
  - 접속방식 라디오 (비밀번호 / PPK)
  - 비밀번호 입력 (비밀번호 선택 시)
  - PPK 파일 경로 + [찾아보기...] (PPK 선택 시)
- [ ] 다이얼로그 중첩 처리 (Hide/Show 패턴)

#### 3.2.4 배포 다이얼로그
- [ ] 작업 유형 선택 라디오
  - ◉ 방화벽 룰 배포 / ○ 패키지 업데이트
- [ ] 방화벽 룰 배포 모드
  - 룰 버전 드롭다운 (TemplateTab에서 목록 조회)
- [ ] 패키지 업데이트 모드
  - 패키지 드롭다운 (ProgramTab에서 목록 조회)
- [ ] [배포/업데이트] [취소] 버튼

### 3.3 Quality Gate ✋

**빌드 & 테스트**:
- [ ] `go build` 성공
- [ ] 기존 테스트 통과: `go test ./...`

**기능 검증** (수동 테스트):
- [ ] 장비 목록 정상 표시 (신규 컬럼 포함)
- [ ] 장비 추가/수정/삭제 정상 동작
- [ ] SSH 인증 정보 저장/로드 정상
- [ ] 배포 다이얼로그 모드 전환 동작
- [ ] **방화벽 룰 배포 정상 동작** ⭐

**하위 호환성**:
- [ ] 기존 firewalls.json 로드 정상 (신규 필드 없어도 동작)
- [ ] 저장 시 신규 필드 포함하여 저장

---

## Phase 4: 배포이력 탭 확장

**목표**: 배포이력 탭을 방화벽 룰 + 패키지 업데이트 통합 이력으로 변경
**상태**: ⏳ 대기
**선행 조건**: Phase 3 Quality Gate 통과

### 4.1 Model 변경

#### 4.1.1 DeployHistory 모델 확장
- [ ] 모델 변경
  ```go
  type DeployHistory struct {
      ID          int           `json:"id"`
      Timestamp   time.Time     `json:"timestamp"`
      DeviceName  string        `json:"deviceName"`
      DeviceIP    string        `json:"deviceIp"`
      Type        string        `json:"type"`      // "firewall" 또는 "program"
      Version     string        `json:"version"`   // 공통 버전 필드
      Status      string        `json:"status"`    // "success", "fail"

      // Type="firewall"일 때만 사용
      Results     []ResultInfo  `json:"results,omitempty"`

      // Type="program"일 때만 사용
      ProgramName string        `json:"programName,omitempty"`
      Message     string        `json:"message,omitempty"`
  }
  ```
  - 파일: `fms_fyne/internal/model/history.go`

- [ ] 이력 유형 상수 추가
  ```go
  const (
      HistoryTypeFirewall = "firewall"
      HistoryTypeProgram  = "program"
  )
  ```

### 4.2 HistoryTab UI 변경

#### 4.2.1 상단 영역 변경 (PagedTable 사용)
- [ ] 검색 입력 필드
- [ ] [찾기] 버튼
- [ ] [선택삭제] 버튼

#### 4.2.2 이력 테이블 컬럼 변경
- [ ] 컬럼 정의
  - 선택 (체크박스)
  - 일시
  - 장비명
  - 장비 IP
  - 유형 ("방화벽 룰" / "패키지")
  - 버전
  - 결과

#### 4.2.3 검색/필터 기능
- [ ] 검색 대상: 장비명, IP, 버전 (부분 일치)
- [ ] 페이지네이션

#### 4.2.4 상세 결과 패널
- [ ] 유형별 상세 표시 분기
  - 방화벽 룰: 규칙별 결과 테이블 (기존)
  - 패키지: 패키지명, 버전, 메시지

### 4.3 Quality Gate ✋

**빌드 & 테스트**:
- [ ] `go build` 성공
- [ ] 기존 테스트 통과: `go test ./...`

**기능 검증** (수동 테스트):
- [ ] 이력 목록 정상 표시 (신규 컬럼 포함)
- [ ] 검색 기능 정상 동작
- [ ] 선택 삭제 정상 동작
- [ ] 이력 선택 시 상세 결과 표시

**통합 검증** ⭐:
- [ ] **장비관리 탭에서 방화벽 룰 배포 → 이력 탭에 기록 확인**
- [ ] 이력 유형이 "방화벽 룰"로 표시
- [ ] 상세 결과에서 규칙별 결과 정상 표시

**하위 호환성**:
- [ ] 기존 history.json 로드 정상 (type 필드 없으면 "firewall"으로 처리)
- [ ] 기존 templateVersion → version 매핑

---

## Phase 5: 패키지 업데이트 기능 구현

**목표**: SSH/SFTP 서비스 구현 및 패키지 업데이트 연동
**상태**: ⏳ 대기
**선행 조건**: Phase 4 Quality Gate 통과

### 5.1 Service 계층 구현

#### 5.1.1 SSHService 생성
- [ ] SSH 연결 기능
  - PPK 키 기반 인증
  - 비밀번호 인증
  - `golang.org/x/crypto/ssh` 패키지 사용

- [ ] SFTP 파일 업로드 기능
  - `github.com/pkg/sftp` 패키지 사용
  - 진행률 콜백 지원

- 파일: `fms_fyne/internal/service/ssh_service.go` (신규)

#### 5.1.2 UpdateService 생성
- [ ] 패키지 업데이트 실행 흐름
  1. SSHService로 SSH 연결
  2. SSHService로 파일 업로드 (SFTP)
  3. HTTP POST /program-update 호출
  4. 결과 반환

- [ ] HTTP Client 확장
  - POST /program-update 메서드 추가
  - GET /device-report 메서드 추가

- 파일: `fms_fyne/internal/service/update_service.go` (신규)

### 5.2 DeviceTab 패키지 업데이트 연동

#### 5.2.1 업데이트 실행 구현
- [ ] onProgramUpdate() 함수 구현
  - 선택된 패키지 확인
  - 체크된 장비 수집
  - 진행률 다이얼로그 표시
  - UpdateService 호출
  - 결과 처리 및 이력 저장

#### 5.2.2 업데이트 진행 다이얼로그
- [ ] 장비별 진행 상태 표시
  - SSH 연결 중...
  - 파일 업로드 중... (진행률 %)
  - 업데이트 요청 중...
  - 완료/실패

#### 5.2.3 상태 확인 확장
- [ ] GET /device-report 호출
- [ ] Firewall.Processes 갱신

### 5.3 이력 연동

- [ ] 패키지 업데이트 완료 시 이력 저장
  - Type: "program"
  - ProgramName, Version, Message 저장

### 5.4 Quality Gate ✋

**빌드 & 테스트**:
- [ ] `go build` 성공
- [ ] 모든 테스트 통과: `go test ./...`

**기능 검증** (수동 테스트):
- [ ] 장비관리 탭에서 패키지 업데이트 선택
- [ ] 패키지 드롭다운에 등록된 패키지 표시
- [ ] 장비 선택 후 업데이트 실행
- [ ] 업데이트 진행 상태 표시
- [ ] 업데이트 완료 결과 표시

**통합 검증**:
- [ ] 업데이트 완료 후 배포이력 탭에서 이력 확인
- [ ] 이력 유형이 "패키지"으로 표시

**기존 기능 확인** ⭐:
- [ ] 방화벽 룰 배포 정상 동작
- [ ] 방화벽 룰 배포 이력 정상 기록

---

## Phase 6: 통합 테스트 및 마무리

**목표**: 전체 기능 통합 테스트 및 문서화
**상태**: ⏳ 대기
**선행 조건**: Phase 5 Quality Gate 통과

### 6.1 통합 테스트

#### 6.1.1 방화벽 룰 배포 시나리오
- [ ] 장비 등록 → 상태 확인 → 룰 배포 → 이력 확인
- [ ] 다중 장비 일괄 배포
- [ ] 배포 실패 케이스 (연결 불가 장비)

#### 6.1.2 패키지 업데이트 시나리오
- [ ] 패키지 등록 → 장비 선택 → 업데이트 → 이력 확인
- [ ] 다중 장비 일괄 업데이트
- [ ] 업데이트 실패 케이스 (SSH 연결 실패, 파일 업로드 실패)

#### 6.1.3 데이터 호환성
- [ ] 기존 firewalls.json 마이그레이션 테스트
- [ ] 기존 history.json 마이그레이션 테스트
- [ ] Export/Import 기능 정상 동작

### 6.2 코드 품질

- [ ] 미사용 코드 정리
- [ ] 주석 정리
- [ ] 에러 핸들링 검토

### 6.3 문서 업데이트

- [ ] PRD 문서 최종 검토
- [ ] CLAUDE.md 업데이트 (신규 구조 반영)
- [ ] 체크리스트 완료 처리

### 6.4 Final Quality Gate ✋

**빌드 & 배포**:
- [ ] `go build -ldflags "-H windowsgui -s -w" -o fms_fyne.exe .` 성공
- [ ] 모든 테스트 통과
- [ ] 실행 파일 정상 동작

**최종 검증**:
- [ ] 모든 Phase Quality Gate 통과 확인
- [ ] 통합 테스트 시나리오 모두 통과
- [ ] 하위 호환성 검증 완료

---

## ⚠️ 위험 요소

| 위험 | 확률 | 영향 | 완화 방안 |
|------|------|------|-----------|
| PPK 키 형식 호환성 | 중 | 중 | PPK 형식 파싱 라이브러리 사용 또는 OpenSSH 변환 안내 |
| 기존 JSON 마이그레이션 실패 | 낮 | 높 | 신규 필드 기본값 처리, 백업 권장 |
| SFTP 라이브러리 호환성 | 낮 | 중 | github.com/pkg/sftp 검증된 라이브러리 사용 |
| Agent API 응답 형식 변경 | 중 | 중 | PRD에 정의된 형식 준수, 에러 핸들링 강화 |
| Fyne 다이얼로그 중첩 이슈 | 중 | 중 | Hide/Show 패턴 적용 |

---

## 🔄 롤백 전략

### Phase 1 롤백
- Git에서 이전 커밋으로 복원
- AppTabs 구조로 복원

### Phase 2-4 롤백
- 신규 파일 삭제: process_info.go, program_tab.go, processes.json
- 이전 Phase 완료 상태로 복원

### Phase 5 롤백
- 신규 파일 삭제: ssh_service.go, update_service.go
- Phase 4 완료 상태로 복원

---

## 📝 Notes & Learnings

### 구현 중 발견 사항
- PagedTable 컴포넌트 구현 완료 (Phase 0)
- 더블클릭 감지 로직 추가 (300ms 이내 재클릭)

### 이슈 및 해결
- (이슈 발생 시 기록)

---

## 📊 진행 상황

| Phase | 상태 | 완료율 | 비고 |
|-------|------|--------|------|
| Phase 0: 공통 컴포넌트 | ✅ 완료 | 100% | PagedTable 구현 |
| Phase 1: UI 레이아웃 마이그레이션 | ⏳ 대기 | 0% | DocTabs + 왼쪽 메뉴 |
| Phase 2: 패키지 관리 탭 | ⏳ 대기 | 0% | 신규 탭 구현 |
| Phase 3: 장비관리 탭 확장 | ⏳ 대기 | 0% | SSH 인증 + 패키지 버전 |
| Phase 4: 배포이력 탭 확장 | ⏳ 대기 | 0% | 통합 이력 |
| Phase 5: 패키지 업데이트 | ⏳ 대기 | 0% | SSH/SFTP + HTTP |
| Phase 6: 통합 테스트 | ⏳ 대기 | 0% | 전체 검증 |

**전체 진행률**: 14% (Phase 0 완료)

---

**다음 작업**: Phase 1 시작 - UI 레이아웃 마이그레이션 (DocTabs + 왼쪽 메뉴)
