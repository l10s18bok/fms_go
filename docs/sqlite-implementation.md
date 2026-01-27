# SQLite 데이터베이스 도입 설계 문서

## 1. 개요

### 1.1 목적
- 배포 이력(DeployHistory)을 SQLite 데이터베이스로 저장
- 향후 서버 상태 관리 데이터 등 확장 가능한 구조 설계

### 1.2 선택한 라이브러리
**modernc.org/sqlite** (순수 Go)
- CGO 불필요 → Windows 빌드 간편
- 크로스 컴파일 용이
- 널리 사용되며 안정적

---

## 2. 현재 구조 분석

### 2.1 Repository 패턴 (기존)
```
internal/repository/
├── interfaces.go           # Repository 인터페이스 정의
├── json_history_repo.go    # JSON 기반 구현체
├── json_firewall_repo.go
├── json_program_repo.go
└── json_rule_file_repo.go
```

### 2.2 HistoryRepository 인터페이스
```go
type HistoryRepository interface {
    GetAll() ([]*model.DeployHistory, error)
    GetByID(id int) (*model.DeployHistory, error)
    GetByType(historyType string) ([]*model.DeployHistory, error)
    Save(h *model.DeployHistory) error
    Delete(id int) error
    Clear() error
    Count() int
}
```

### 2.3 DeployHistory 모델
```go
type DeployHistory struct {
    ID          int            // 고유 ID (Auto Increment)
    Type        string         // 이력 유형 (firewall/program)
    Timestamp   utils.JSONTime // 배포 시간
    DeviceName  string         // 장비명
    DeviceIP    string         // 장비 IP
    TemplateVer string         // 규칙 파일 버전
    ProgramName string         // 패키지 이름
    ProgramVer  string         // 패키지 버전
    Message     string         // 결과 메시지
    Status      string         // 배포 상태
    Results     []RuleResult   // 규칙별 결과 (1:N)
}

type RuleResult struct {
    Rule   string // 규칙 내용
    Text   string // 규칙 설명
    Status string // 결과 (ok/error/unfind/validation)
    Reason string // 실패 사유
}
```

---

## 3. SQLite 스키마 설계

### 3.1 테이블 구조

#### deploy_history (배포 이력)
| 컬럼 | 타입 | 설명 |
|------|------|------|
| id | INTEGER PRIMARY KEY | 자동 증가 ID |
| type | TEXT | 이력 유형 (firewall/program) |
| timestamp | DATETIME | 배포 시간 |
| device_name | TEXT | 장비명 |
| device_ip | TEXT | 장비 IP |
| template_version | TEXT | 규칙 파일 버전 |
| program_name | TEXT | 패키지 이름 |
| program_version | TEXT | 패키지 버전 |
| message | TEXT | 결과 메시지 |
| status | TEXT | 배포 상태 |

#### rule_results (규칙 결과)
| 컬럼 | 타입 | 설명 |
|------|------|------|
| id | INTEGER PRIMARY KEY | 자동 증가 ID |
| history_id | INTEGER | 외래키 (deploy_history.id) |
| rule | TEXT | 규칙 내용 |
| text | TEXT | 규칙 설명 |
| status | TEXT | 결과 상태 |
| reason | TEXT | 실패 사유 |

### 3.2 DDL
```sql
-- 배포 이력 테이블
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

-- 규칙 결과 테이블
CREATE TABLE IF NOT EXISTS rule_results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    history_id INTEGER NOT NULL,
    rule TEXT,
    text TEXT,
    status TEXT,
    reason TEXT,
    FOREIGN KEY (history_id) REFERENCES deploy_history(id) ON DELETE CASCADE
);

-- 인덱스
CREATE INDEX IF NOT EXISTS idx_history_type ON deploy_history(type);
CREATE INDEX IF NOT EXISTS idx_history_device_ip ON deploy_history(device_ip);
CREATE INDEX IF NOT EXISTS idx_history_timestamp ON deploy_history(timestamp);
CREATE INDEX IF NOT EXISTS idx_rule_results_history_id ON rule_results(history_id);
```

---

## 4. 구현 구조

### 4.1 파일 구조 (신규/수정)
```
internal/
├── storage/
│   ├── json_store.go       # 기존 (유지)
│   ├── sqlite_store.go     # 신규: SQLite 연결 관리
│   └── migration.go        # 신규: JSON → SQLite 마이그레이션
├── repository/
│   ├── interfaces.go       # 기존 (유지)
│   ├── json_history_repo.go    # 기존 (유지)
│   ├── sqlite_history_repo.go  # 신규: SQLite 구현체
│   └── factory.go          # 신규: Repository 팩토리
└── model/
    └── config.go           # 수정: StorageType 추가
```

### 4.2 설정 확장
```go
type Config struct {
    // 기존 필드...
    ConnectionMode string `json:"connectionMode"`
    AgentServerURL string `json:"agentServerURL"`
    TimeoutSeconds int    `json:"timeoutSeconds"`

    // 신규 필드
    StorageType string `json:"storageType"` // "json" 또는 "sqlite"
}
```

### 4.3 팩토리 패턴
```go
// factory.go
func NewHistoryRepository(config *model.Config, jsonStore *storage.JSONStore, sqliteStore *storage.SQLiteStore) HistoryRepository {
    if config.StorageType == "sqlite" {
        return NewSQLiteHistoryRepository(sqliteStore)
    }
    return NewJSONHistoryRepository(jsonStore)
}
```

---

## 5. 마이그레이션 전략

### 5.1 JSON → SQLite 마이그레이션
1. 애플리케이션 시작 시 SQLite DB 존재 여부 확인
2. DB가 없고 history.json이 있으면 자동 마이그레이션 제안
3. 사용자 선택에 따라 마이그레이션 실행

### 5.2 마이그레이션 프로세스
```
1. history.json 읽기
2. SQLite 테이블 생성
3. 각 DeployHistory를 deploy_history 테이블에 INSERT
4. 각 RuleResult를 rule_results 테이블에 INSERT
5. 완료 후 config.json의 StorageType을 "sqlite"로 변경
```

---

## 6. 향후 확장 계획

SQLite 기반 구조가 완성되면 다음 데이터도 테이블로 추가 가능:

| 테이블 | 용도 |
|--------|------|
| server_status_history | 서버 상태 이력 |
| firewalls | 장비 정보 (현재 JSON) |
| rule_files | 규칙 파일 (현재 JSON) |
| programs | 패키지 정보 (현재 JSON) |

---

## 7. 참고 자료

- [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) - 순수 Go SQLite 드라이버
- [database/sql](https://pkg.go.dev/database/sql) - Go 표준 DB 인터페이스
