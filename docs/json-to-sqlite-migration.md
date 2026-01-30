# FMS 데이터 저장소 SQLite 마이그레이션 구현 문서

## 1. 개요

FMS 애플리케이션의 데이터 저장 방식을 JSON 파일 기반에서 SQLite 데이터베이스로 전환한 구현 내용을 정리한 문서입니다.

### 마이그레이션 대상

| 데이터 | 마이그레이션 전 | 마이그레이션 후 | 상태 |
|--------|----------------|----------------|------|
| 장비 (Firewall) | `config/firewallList.json` | SQLite `firewalls` 테이블 | ✅ 구현 완료 |
| 패키지 (Program) | `config/programList.json` | SQLite `programs` 테이블 | ✅ 구현 완료 |
| 배포 이력 (History) | - | SQLite `deploy_history` 테이블 | ✅ 기존 구현 |
| 설정 (Config) | `config/config.json` | JSON 유지 | - 변경 없음 |
| 방화벽 규칙 파일 | `data/*.rules` | 파일 유지 | - 변경 없음 |

### SQLite 드라이버

- `modernc.org/sqlite` (CGo-free, Pure Go 구현)
- Windows 환경에서 C 컴파일러 없이 빌드 가능

---

## 2. 아키텍처

### 저장소 계층 구조

```
main.go
  ├── SQLiteStore (storage/sqlite_store.go)
  │     ├── SQLiteFirewallRepository (repository/sqlite_firewall_repo.go)
  │     ├── SQLiteProgramRepository (repository/sqlite_program_repo.go)
  │     └── SQLiteHistoryRepository (repository/sqlite_history_repo.go)
  ├── JSONStore (storage/json_store.go)     → Config 전용
  └── FileStore (storage/file_store.go)     → 규칙 파일 전용
```

### 초기화 흐름 (main.go)

```go
// SQLite 저장소 초기화
sqliteStore, err := storage.NewSQLiteStore(configDir)
defer sqliteStore.Close()

// Repository 생성
historyRepo := repository.NewSQLiteHistoryRepository(sqliteStore)
firewallRepo := repository.NewSQLiteFirewallRepository(sqliteStore)
programRepo := repository.NewSQLiteProgramRepository(sqliteStore)

// UI에 Repository 주입
mainUI := ui.NewMainUI(w, jsonStore, fileStore, historyRepo, firewallRepo, programRepo)
```

---

## 3. 데이터베이스 스키마

### 3.1 firewalls 테이블

```sql
CREATE TABLE IF NOT EXISTS firewalls (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    device_name TEXT NOT NULL,
    device_ip TEXT,
    server_status TEXT DEFAULT '-',
    deploy_status TEXT DEFAULT '-',
    version TEXT DEFAULT '-',
    deploy_result TEXT,
    device_id TEXT,
    device_pw TEXT,
    device_ppk TEXT,
    last_checked_at TEXT,
    location TEXT,
    os_info TEXT,
    cpu_usage REAL DEFAULT 0,
    memory_usage REAL DEFAULT 0,
    disk_usage REAL DEFAULT 0,
    network_usage REAL DEFAULT 0,
    program_upload_path TEXT,
    program_update_scheme TEXT,
    program_update_path TEXT,
    program_update_port INTEGER DEFAULT 0,
    firewall_deploy_scheme TEXT,
    firewall_deploy_path TEXT,
    firewall_deploy_port INTEGER DEFAULT 0,
    device_info_scheme TEXT,
    device_info_path TEXT,
    device_info_port INTEGER DEFAULT 0
);
```

### 3.2 firewall_program_versions 테이블

```sql
CREATE TABLE IF NOT EXISTS firewall_program_versions (
    firewall_id INTEGER NOT NULL,
    program_name TEXT NOT NULL,
    version TEXT,
    PRIMARY KEY (firewall_id, program_name),
    FOREIGN KEY (firewall_id) REFERENCES firewalls(id) ON DELETE CASCADE
);
```

### 3.3 programs 테이블

```sql
CREATE TABLE IF NOT EXISTS programs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    process_name TEXT NOT NULL,
    process_file_path TEXT NOT NULL,
    process_version TEXT NOT NULL,
    process_created_at TEXT
);
```

### 3.4 deploy_history / rule_results 테이블

```sql
CREATE TABLE IF NOT EXISTS deploy_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT,
    timestamp DATETIME,
    device_name TEXT,
    device_ip TEXT,
    template_version TEXT,
    program_name TEXT,
    program_version TEXT,
    message TEXT,
    status TEXT
);

CREATE TABLE IF NOT EXISTS rule_results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    history_id INTEGER NOT NULL,
    rule TEXT,
    text TEXT,
    status TEXT,
    reason TEXT,
    FOREIGN KEY (history_id) REFERENCES deploy_history(id) ON DELETE CASCADE
);
```

### 3.5 인덱스

```sql
CREATE INDEX IF NOT EXISTS idx_firewalls_device_ip ON firewalls(device_ip);
CREATE INDEX IF NOT EXISTS idx_firewalls_device_name ON firewalls(device_name);
CREATE INDEX IF NOT EXISTS idx_programs_name ON programs(process_name);
CREATE INDEX IF NOT EXISTS idx_history_type ON deploy_history(type);
CREATE INDEX IF NOT EXISTS idx_history_device_ip ON deploy_history(device_ip);
CREATE INDEX IF NOT EXISTS idx_history_timestamp ON deploy_history(timestamp);
CREATE INDEX IF NOT EXISTS idx_rule_results_history_id ON rule_results(history_id);
```

---

## 4. Repository 인터페이스

### 4.1 FirewallRepository

```go
type FirewallRepository interface {
    GetAll() ([]*model.Firewall, error)
    GetByIndex(index int) (*model.Firewall, error)
    GetByIP(ip string) (*model.Firewall, error)
    GetPage(req model.PageRequest) (*model.PageResult[model.Firewall], error)
    Save(fw *model.Firewall) error
    Delete(index int) error
    Clear() error
    Count() int
}
```

### 4.2 ProgramRepository

```go
type ProgramRepository interface {
    GetAll() ([]*model.ProcessInfo, error)
    GetByID(id int) (*model.ProcessInfo, error)
    GetByName(name string) (*model.ProcessInfo, error)
    GetPage(req model.PageRequest) (*model.PageResult[model.ProcessInfo], error)
    Save(p *model.ProcessInfo) error
    Delete(id int) error
    Clear() error
    Count() int
}
```

### 4.3 HistoryRepository

```go
type HistoryRepository interface {
    GetAll() ([]*model.DeployHistory, error)
    GetByID(id int) (*model.DeployHistory, error)
    GetByType(historyType string) ([]*model.DeployHistory, error)
    GetPage(req model.PageRequest) (*model.PageResult[model.DeployHistory], error)
    Save(h *model.DeployHistory) error
    Delete(id int) error
    Clear() error
    Count() int
}
```

---

## 5. DB 레벨 페이지네이션

### 5.1 개요

장비, 패키지, 배포이력 3개 탭 모두 동일한 DB 레벨 페이지네이션 구조를 사용합니다. `LIMIT/OFFSET` 기반으로 현재 페이지에 해당하는 데이터만 DB에서 조회합니다.

### 5.2 PageRequest / PageResult

```go
type PageRequest struct {
    Page     int    // 현재 페이지 (0-based)
    PageSize int    // 페이지당 건수
    Limit    int    // LIMIT (PageSize와 동일하게 사용)
    Offset   int    // OFFSET (Page * PageSize)
    Keyword  string // 검색 키워드
    Filter   string // 필터 (배포이력: type 필터)
}

type PageResult[T any] struct {
    Items      []*T
    TotalCount int
    Page       int
    PageSize   int
}
```

### 5.3 탭별 구현 구조

3개 탭 모두 동일한 패턴으로 구현되어 있습니다:

```go
// 공통 데이터 필드
pageData      []*model.T    // 현재 페이지 데이터
totalCount    int           // 검색 적용 후 전체 건수
searchKeyword string        // 검색 키워드

// 공통 메서드
loadPage(page, pageSize int) int   // DB에서 페이지 데이터 조회
GetExportData() ([]*model.T, error) // Export용 데이터 조회
IsFiltered() bool                   // 검색/필터 적용 여부
```

### 5.4 PagedTable OnPageLoad 콜백

각 탭에서 `PagedTable` 컴포넌트에 `OnPageLoad` 콜백을 설정하여 페이지 변경 시 DB에서 데이터를 조회합니다:

```go
tab.deviceTable.OnPageLoad = func(page, pageSize int) int {
    return tab.loadPage(page, pageSize)
}
```

### 5.5 이전 방식과의 차이

| 항목 | 이전 (메모리 필터링) | 현재 (DB 페이지네이션) |
|------|---------------------|----------------------|
| 데이터 로드 | `GetAll()` → 메모리에 전체 보관 | `GetPage()` → 현재 페이지만 |
| 검색 | 메모리에서 `strings.Contains` 필터링 | SQL `LIKE` 검색 |
| 정렬 | `sort.Slice()` | SQL `ORDER BY` |
| 페이지 이동 | 메모리 슬라이싱 | `LIMIT/OFFSET` 재조회 |
| 메모리 사용 | 전체 데이터 상주 | 현재 페이지만 |

---

## 6. Import 기능

### 6.1 지원 범위

| 탭 | Import 지원 | 비고 |
|----|------------|------|
| 방화벽 관리 | ❌ | 파일 기반 (data 디렉토리) |
| 장비 관리 | ✅ | JSON → SQLite |
| 배포 이력 | ❌ | DB 전용 데이터 |
| 패키지 관리 | ✅ | JSON → SQLite |

### 6.2 Import 모드 선택

DB에 기존 데이터가 있는 경우 3버튼 다이얼로그로 Import 방식을 선택합니다:

```
┌─ Import 방식 선택 ────────────────────────┐
│                                           │
│ 장비 데이터를 어떤 방식으로 가져오시겠습니까? │
│ 교체: 기존 데이터를 삭제하고 새로운 데이터로 교체 │
│ 병합: 기존 데이터를 유지하고 중복 항목은 덮어쓰기 │
│                                           │
│ [교체]  [병합]  [취소]                      │
└───────────────────────────────────────────┘
```

DB에 기존 데이터가 없으면 모드 선택 없이 바로 Import를 실행합니다.

### 6.3 교체 모드

1. `Clear()` → 기존 데이터 전부 삭제
2. JSON 파일의 모든 항목을 `Save()` (Index/ID = -1, INSERT)

### 6.4 병합 모드

#### 장비 (device_ip 기준)

```go
for _, fw := range firewalls {
    existing, err := m.firewallRepo.GetByIP(fw.DeviceIP)
    if err == nil && existing != nil {
        fw.Index = existing.Index  // 기존 ID → UPDATE
        updatedCount++
    } else {
        fw.Index = -1              // 신규 → INSERT
    }
    m.firewallRepo.Save(fw)
}
```

#### 패키지 (process_name 기준)

```go
for _, p := range programs {
    existing, err := m.programRepo.GetByName(p.ProcessName)
    if err == nil && existing != nil {
        p.ID = existing.ID  // 기존 ID → UPDATE
        updatedCount++
    } else {
        p.ID = -1           // 신규 → INSERT
    }
    m.programRepo.Save(p)
}
```

### 6.5 Import 결과 메시지

- 교체 모드: `"N개의 장비 정보가 가져오기 되었습니다."`
- 병합 모드: `"N개의 장비 정보가 가져오기 되었습니다. (신규: X, 업데이트: Y)"`

### 6.6 JSON 파일 호환성

기존 `firewallList.json`, `programList.json` 파일을 그대로 Import할 수 있습니다. Export 시에도 동일한 JSON 형식을 사용합니다.

---

## 7. Export 기능

### 7.1 지원 범위

| 탭 | Export 지원 | 비고 |
|----|------------|------|
| 방화벽 관리 | ❌ | 파일 기반 (data 디렉토리에서 직접 복사) |
| 장비 관리 | ✅ | 검색 결과 또는 전체 Export |
| 배포 이력 | ✅ | 필터/검색 결과 또는 전체 Export |
| 패키지 관리 | ✅ | 검색 결과 또는 전체 Export |

### 7.2 Export 동작 방식

1. 각 탭의 `IsFiltered()` 메서드로 검색/필터 적용 여부 확인
2. 각 탭의 `GetExportData()` 메서드로 데이터 조회:
   - 검색 없음 → `GetAll()` (전체 데이터)
   - 검색 있음 → `GetPage(Limit: 100000)` (검색 조건의 모든 결과)
3. 건수 확인 다이얼로그 표시:
   - 일반: `"N건의 데이터를 내보내시겠습니까?"`
   - 검색 적용: `"검색 결과 N건의 데이터를 내보내시겠습니까?"`
4. JSON 파일로 저장 (`json.MarshalIndent`)

### 7.3 GetExportData 구현

```go
// 장비 탭 예시
func (d *DeviceTab) GetExportData() ([]*model.Firewall, error) {
    if d.searchKeyword == "" {
        return d.firewallRepo.GetAll()
    }
    result, err := d.firewallRepo.GetPage(model.PageRequest{
        Keyword: d.searchKeyword,
        Limit:   100000,
    })
    return result.Items, err
}
```

### 7.4 IsFiltered 구현

```go
// 장비/패키지: 검색 키워드 여부
func (d *DeviceTab) IsFiltered() bool { return d.searchKeyword != "" }

// 배포이력: 검색 키워드 + 타입 필터
func (h *HistoryTab) IsFiltered() bool {
    return h.currentFilter != "" || h.searchKeyword != ""
}
```

### 7.5 기본 파일명

| 탭 | 기본 파일명 |
|----|------------|
| 장비 관리 | `firewallList.json` |
| 배포 이력 | `deployHistory.json` |
| 패키지 관리 | `programList.json` |

---

## 8. 파일 구조

### 변경된 파일

| 파일 | 설명 |
|------|------|
| `main.go` | SQLiteStore 초기화, Repository 생성, MainUI에 주입 |
| `internal/storage/sqlite_store.go` | SQLite 연결, 스키마 초기화, CRUD 메서드 |
| `internal/repository/interfaces.go` | Repository 인터페이스 정의 |
| `internal/repository/sqlite_firewall_repo.go` | SQLite 장비 저장소 어댑터 |
| `internal/repository/sqlite_program_repo.go` | SQLite 패키지 저장소 어댑터 |
| `internal/repository/sqlite_history_repo.go` | SQLite 배포이력 저장소 어댑터 |
| `internal/ui/app.go` | MainUI에 firewallRepo/programRepo 추가, Import/Export 로직 |
| `internal/ui/device_tab.go` | DB 페이지네이션 전환, GetExportData/IsFiltered 추가 |
| `internal/ui/program_tab.go` | DB 페이지네이션 전환, GetExportData/IsFiltered 추가 |
| `internal/ui/history_tab.go` | GetExportData/IsFiltered 추가 |

### 신규 파일

| 파일 | 설명 |
|------|------|
| `internal/repository/sqlite_firewall_repo.go` | 장비 SQLite Repository |
| `internal/repository/sqlite_program_repo.go` | 패키지 SQLite Repository |

---

## 9. 데이터 파일 위치

```
{실행파일 경로}/
├── fms_fyne.exe
├── config/
│   ├── fms.db              ← SQLite 데이터베이스 (장비, 패키지, 배포이력)
│   └── config.json         ← 설정 (JSON 유지)
└── data/
    └── *.rules             ← 방화벽 규칙 파일
```

- `fms.db`는 앱 최초 실행 시 자동 생성됩니다.
- 기존 JSON 파일(`firewallList.json`, `programList.json`)은 Import 기능을 통해 마이그레이션할 수 있습니다.
