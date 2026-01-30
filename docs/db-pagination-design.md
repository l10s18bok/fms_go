# DB 레벨 페이지네이션 설계 문서

## 1. 배경 및 목적

### 현재 문제점
- 배포이력 조회 시 `GetAllHistory()`로 **전체 데이터를 메모리에 로드** 후 UI에서 페이지 분할
- 각 이력마다 `getRuleResults()` 별도 호출 → **N+1 쿼리 문제** (이력 100개 = DB 쿼리 101번)
- 데이터가 누적될수록 메모리 사용량과 초기 로딩 시간 증가

### 목표
- 페이지 변경 시 **해당 페이지 데이터만** DB에서 조회 (LIMIT/OFFSET)
- 필터/검색도 DB WHERE 절로 처리하여 메모리 필터링 제거
- N+1 쿼리를 배치 조회로 해결
- 향후 다른 데이터(장비, 패키지 등)도 DB 저장 시 동일 패턴 적용 가능하도록 범용 설계

---

## 2. 현재 구조 분석

### 데이터 흐름 (Before)
```
DB (SQLite)
  ↓ GetAllHistory() — 전체 로드, N+1 쿼리
[]*DeployHistory (메모리 전체)
  ↓ sort.Slice() — ID 내림차순
h.histories (전체)
  ↓ applyFilter() — 메모리 내 유형 필터 + 키워드 검색
h.filteredHistories (필터링 결과 전체)
  ↓ SetData(len(filteredHistories))
PagedTable — UI에서 페이지 분할 표시 (기본 15, 윈도우 크기에 따라 5~32 동적 조절)
```

### 관련 파일
| 파일 | 역할 |
|------|------|
| `model/deploy_history.go` | DeployHistory 모델 |
| `repository/interfaces.go` | HistoryRepository 인터페이스 |
| `repository/sqlite_history_repo.go` | SQLite 어댑터 |
| `storage/sqlite_store.go` | 실제 DB 쿼리 (GetAllHistory 등) |
| `ui/component/table.go` | PagedTable 컴포넌트 |
| `ui/history_tab.go` | 배포 이력 탭 UI |

---

## 3. 변경 설계

### 설계 방침
- **캐시 없이 단순하게**: 페이지 이동마다 DB 조회 (로컬 SQLite라 수 ms 수준)
- **UI PageSize = DB LIMIT**: 화면에 표시되는 행 수만큼만 DB에서 조회

### 데이터 흐름 (After)
```
PagedTable 페이지 변경 이벤트
  ↓ OnPageLoad(page, pageSize) 콜백
HistoryTab.loadPage(page, pageSize)
  ↓ PageRequest{Offset=page*pageSize, Limit=pageSize, Filter, Keyword}
HistoryRepository.GetPage(req)
  ↓
SQLiteStore.GetHistoryPage(req)
  ↓ ① SELECT COUNT(*) ... (TotalCount 조회)
  ↓ ② SELECT ... LIMIT pageSize OFFSET page*pageSize
  ↓ ③ SELECT ... WHERE history_id IN (...) (배치 조회)
PageResult[DeployHistory]{Items, TotalCount}
  ↓
h.pageData = Items (현재 페이지 데이터만 보유)
PagedTable에 TotalCount 반영 → "page 1/34" 표시
```

### 3.1 공통 모델 (신규)

**파일: `model/pagination.go`**

```go
// PageRequest 페이지네이션 요청
type PageRequest struct {
    Offset   int    // 시작 위치 (0-based)
    Limit    int    // 가져올 건수 (= UI PageSize)
    Filter   string // 유형 필터 ("firewall", "program", "" = 전체)
    Keyword  string // 검색 키워드 ("" = 검색 없음)
}

// PageResult 페이지네이션 응답 (제네릭)
type PageResult[T any] struct {
    Items      []*T // 조회된 데이터
    TotalCount int  // 필터/검색 적용 후 전체 건수
}
```

### 3.2 Repository 인터페이스 확장

**파일: `repository/interfaces.go`**

HistoryRepository에 메서드 추가:

```go
type HistoryRepository interface {
    // ... 기존 메서드 유지 ...

    // 신규: DB 레벨 페이지네이션
    GetPage(req model.PageRequest) (*model.PageResult[model.DeployHistory], error)
}
```

> `CountFiltered`는 별도 메서드로 분리하지 않고 `GetPage` 내부에서 카운트를 함께 조회하여 `PageResult.TotalCount`로 반환합니다.

### 3.3 SQLiteStore DB 쿼리 구현

**파일: `storage/sqlite_store.go`**

#### GetHistoryPage — 핵심 쿼리 (3단계)

```go
func (s *SQLiteStore) GetHistoryPage(req model.PageRequest) (*model.PageResult[model.DeployHistory], error) {
    // 1단계: 필터/검색 조건으로 카운트 조회
    totalCount := s.countHistoryFiltered(req.Filter, req.Keyword)

    // 2단계: 해당 페이지 history 조회 (LIMIT/OFFSET)
    query := `SELECT id, type, timestamp, device_name, device_ip,
              template_version, program_name, program_version, message, status
              FROM deploy_history`
    query += buildWhereClause(req.Filter, req.Keyword)
    query += ` ORDER BY id DESC LIMIT ? OFFSET ?`
    // LIMIT = req.Limit (= UI PageSize), OFFSET = req.Offset

    // 3단계: 조회된 history ID 목록으로 rule_results 배치 조회
    // SELECT * FROM rule_results WHERE history_id IN (?, ?, ...) ORDER BY id
    // → map[int][]RuleResult 매핑 후 각 history에 할당
}
```

#### WHERE 절 빌드

```go
func buildWhereClause(filter, keyword string) (string, []interface{}) {
    var conditions []string
    var args []interface{}

    // 유형 필터
    if filter != "" {
        if filter == "firewall" {
            // 레거시 타입 호환: "firewall", "template", ""
            conditions = append(conditions,
                "(type = ? OR type = ? OR type = ?)")
            args = append(args, "firewall", "template", "")
        } else {
            conditions = append(conditions, "type = ?")
            args = append(args, filter)
        }
    }

    // 키워드 검색 (LIKE)
    if keyword != "" {
        like := "%" + keyword + "%"
        conditions = append(conditions,
            `(device_name LIKE ? OR device_ip LIKE ? OR
              template_version LIKE ? OR program_version LIKE ? OR
              message LIKE ? OR type LIKE ?)`)
        args = append(args, like, like, like, like, like, like)
    }

    if len(conditions) == 0 {
        return "", nil
    }
    return " WHERE " + strings.Join(conditions, " AND "), args
}
```

#### N+1 해결: 배치 조회

```go
func (s *SQLiteStore) getRuleResultsBatch(historyIDs []int) (map[int][]model.RuleResult, error) {
    if len(historyIDs) == 0 {
        return nil, nil
    }
    // placeholders: ?, ?, ? ...
    query := `SELECT history_id, rule, text, status, reason
              FROM rule_results
              WHERE history_id IN (` + placeholders(len(historyIDs)) + `)
              ORDER BY id`
    // 결과를 map[historyID][]RuleResult 로 그룹핑
}
```

### 3.4 SQLiteHistoryRepository 어댑터

**파일: `repository/sqlite_history_repo.go`**

```go
func (r *SQLiteHistoryRepository) GetPage(req model.PageRequest) (*model.PageResult[model.DeployHistory], error) {
    return r.store.GetHistoryPage(req)
}
```

### 3.5 PagedTable 변경

**파일: `ui/component/table.go`**

#### PagedTableConfig에 콜백 추가

```go
type PagedTableConfig struct {
    // ... 기존 필드 ...

    // 페이지 변경 시 데이터 로드 콜백 (설정 시 DB 페이지네이션 모드)
    // 반환값: 전체 항목 수 (totalItems)
    OnPageLoad func(page, pageSize int) int
}
```

#### NextPage/PrevPage/GoToPage/SetPageSize 수정

```go
func (t *PagedTable) NextPage() {
    if t.currentPage < t.getTotalPages()-1 {
        t.currentPage++
        t.selectedRow = -1
        if t.config.OnPageLoad != nil {
            t.totalItems = t.config.OnPageLoad(t.currentPage, t.config.PageSize)
        }
        t.table.Refresh()
        t.updatePaginationUI()
    }
}
// PrevPage, GoToPage 동일 패턴

func (t *PagedTable) SetPageSize(size int) {
    if size > 0 {
        t.config.PageSize = size
        // 페이지 보정
        totalPages := t.getTotalPages()
        if t.currentPage >= totalPages {
            t.currentPage = totalPages - 1
        }
        if t.currentPage < 0 {
            t.currentPage = 0
        }
        // DB 모드: 변경된 PageSize로 현재 페이지 재조회
        if t.config.OnPageLoad != nil {
            t.totalItems = t.config.OnPageLoad(t.currentPage, t.config.PageSize)
        }
        t.table.Refresh()
        t.updatePaginationUI()
    }
}
```

#### SetTotalItems 메서드 추가

```go
// SetTotalItems 전체 항목 수만 갱신 (페이지 리셋 없음, DB 페이지네이션 모드용)
func (t *PagedTable) SetTotalItems(totalItems int) {
    t.totalItems = totalItems
    t.table.Refresh()
    t.updatePaginationUI()
}
```

### 3.6 HistoryTab 변경

**파일: `ui/history_tab.go`**

#### 구조체 필드 변경

```go
type HistoryTab struct {
    // ... 기존 유지 ...

    // 데이터 (변경)
    pageData      []*model.DeployHistory // 현재 페이지 데이터만
    totalCount    int                    // 필터/검색 적용 후 전체 건수
    currentFilter string                 // 현재 유형 필터값
    searchKeyword string                 // 검색 키워드 (기존 유지)

    // 삭제: histories, filteredHistories
}
```

#### PagedTable 생성 시 OnPageLoad 콜백 설정

```go
historyTable = component.NewPagedTable(component.PagedTableConfig{
    // ... 기존 설정 ...
    OnPageLoad: func(page, pageSize int) int {
        return h.loadPage(page, pageSize)
    },
})
```

#### loadPage (신규 핵심 메서드)

```go
// loadPage DB에서 해당 페이지 데이터 조회
func (h *HistoryTab) loadPage(page, pageSize int) int {
    req := model.PageRequest{
        Offset:  page * pageSize,
        Limit:   pageSize,
        Filter:  h.currentFilter,
        Keyword: h.searchKeyword,
    }
    result, err := h.historyRepo.GetPage(req)
    if err != nil {
        // 에러 처리
        return 0
    }
    h.pageData = result.Items
    h.totalCount = result.TotalCount
    return result.TotalCount
}
```

#### updateHistoryCell 변경

```go
func (h *HistoryTab) updateHistoryCell(row int, col int, cell fyne.CanvasObject) {
    // row는 페이지 내 인덱스 → pageData에서 직접 접근
    if row >= len(h.pageData) {
        label.SetText("")
        return
    }
    history := h.pageData[row]
    // ... switch col 동일 ...
}
```

#### applyFilter 변경

```go
func (h *HistoryTab) applyFilter() {
    switch h.typeFilter.Selected {
    case "방화벽 룰":
        h.currentFilter = model.HistoryTypeFirewall
    case "패키지":
        h.currentFilter = model.HistoryTypeProgram
    default:
        h.currentFilter = ""
    }
    total := h.loadPage(0, h.historyTable.GetPageSize())
    h.historyTable.SetData(total)
}
```

#### loadHistory 변경

```go
func (h *HistoryTab) loadHistory() {
    total := h.loadPage(0, h.historyTable.GetPageSize())
    fyne.Do(func() {
        h.historyTable.SetData(total)
    })
}
```

#### onDeleteHistory 변경

```go
func (h *HistoryTab) onDeleteHistory() {
    checkedRows := h.historyTable.GetCheckedRows()
    // ...
    for _, row := range checkedRows {
        if row < len(h.pageData) {
            history := h.pageData[row]
            // 삭제 처리
        }
    }
    // 삭제 후 재로드
    h.historyTable.ClearChecked()
    h.loadHistory()
}
```

#### resetDeviceDeployStatusIfNoHistory 변경

```go
func (h *HistoryTab) resetDeviceDeployStatusIfNoHistory(deviceIP string) {
    // deviceIP 기반 조회는 PageRequest에 없으므로 기존 방식 유지
    // (장비별 이력 존재 여부 확인은 빈도가 낮아 성능 영향 없음)
}
```

---

## 4. 영향 범위

| 항목 | 변경 여부 | 설명 |
|------|-----------|------|
| DeviceTab | 변경 없음 | JSONStore 기반, 데이터 적음 |
| ProgramTab | 변경 없음 | JSONStore 기반, 데이터 적음 |
| FirewallTab | 변경 없음 | FileStore 기반, 데이터 적음 |
| HistoryTab | **변경** | DB 페이지네이션 적용 |
| PagedTable | **변경** | OnPageLoad 콜백 추가 |
| RuleBuilder/NATBuilder | 변경 없음 | EditableTable 사용 (별도 컴포넌트) |

---

## 5. 구현 순서

1. `model/pagination.go` — PageRequest, PageResult 모델
2. `repository/interfaces.go` — GetPage 추가
3. `storage/sqlite_store.go` — GetHistoryPage, countHistoryFiltered, getRuleResultsBatch
4. `repository/sqlite_history_repo.go` — 어댑터 메서드
5. `ui/component/table.go` — OnPageLoad, SetTotalItems
6. `ui/history_tab.go` — DB 페이지네이션 연동
7. 빌드 및 동작 검증

---

## 6. 향후 확장

다른 데이터를 DB로 이관할 때 동일 패턴 적용:

1. `PageRequest`/`PageResult[T]`를 그대로 사용
2. 해당 Repository에 `GetPage(req)` 메서드 추가
3. 해당 Tab에서 `OnPageLoad` 콜백 설정

예시: 장비 데이터를 SQLite로 이관 시

```go
type FirewallRepository interface {
    // 기존 ...
    GetPage(req model.PageRequest) (*model.PageResult[model.Firewall], error)
}
```
