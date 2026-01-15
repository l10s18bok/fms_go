# Fyne 레이아웃 업데이트

## 개요

FMS Fyne 애플리케이션의 메인 UI 레이아웃을 개선하여 왼쪽 메뉴 + 닫기 가능한 탭 구조로 변경했습니다.

## 변경 일자

2026-01-14

## 변경 내용

### 1. 레이아웃 구조 변경

**변경 전:**
```
┌─────────────────────────────────────────────┐
│  [템플릿 관리] [장비 관리] [배포 이력]       │  ← AppTabs (상단 탭)
├─────────────────────────────────────────────┤
│                                             │
│              탭 콘텐츠                       │
│                                             │
└─────────────────────────────────────────────┘
```

**변경 후:**
```
┌──────────────┬──────────────────────────────┐
│              │ [방화벽 룰 관리 ×] [장비 관리 ×] │  ← DocTabs (닫기 버튼 지원)
│  왼쪽 메뉴    │──────────────────────────────│
│              │                              │
│ ○ 방화벽 룰  │         탭 콘텐츠             │
│ ○ 프로그램   │                              │
│ ○ 장비 관리  │                              │
│ ○ 배포 이력  │                              │
│              │                              │
└──────────────┴──────────────────────────────┘
       ↑                    ↑
   150px 고정          자동 확장
```

### 2. 주요 변경사항

| 항목 | 변경 전 | 변경 후 |
|------|---------|---------|
| 탭 컴포넌트 | `container.AppTabs` | `container.DocTabs` |
| 탭 닫기 버튼 | 없음 | 지원 (×) |
| 왼쪽 메뉴 | 없음 | 버튼 4개 (방화벽 룰 관리, 프로그램 관리, 장비 관리, 배포 이력) |
| 메뉴 너비 | - | 최소 150px (`minWidthLayout`) |
| 구분선 | 없음 | 왼쪽 메뉴와 탭 영역 사이 세로 구분선 |
| 탭 동적 관리 | 불가 | 탭 열기/닫기 동적 지원 |

### 3. 용어 변경

| 변경 전 | 변경 후 |
|---------|---------|
| 템플릿 관리 | 방화벽 룰 관리 |
| 새 템플릿 | 새 룰 추가 |
| 템플릿 목록 | 룰 목록 |
| 템플릿 내용 | 규칙 내용 |
| 규칙 빌더 | 룰 빌더 |

## 수정된 파일

### `fms_fyne/internal/ui/app.go`

주요 변경:

1. **구조체 변경**
   ```go
   type MainUI struct {
       tabs        *container.DocTabs  // AppTabs → DocTabs
       leftMenu    *fyne.Container     // 새로 추가

       // 탭 아이템 참조 (동적 추가/제거용)
       templateTabItem *container.TabItem
       programTabItem  *container.TabItem  // 신규
       deviceTabItem   *container.TabItem
       historyTabItem  *container.TabItem
   }
   ```

2. **왼쪽 메뉴 생성** (`createLeftMenu`)
   - 방화벽 룰 관리 버튼
   - 프로그램 관리 버튼 (신규)
   - 장비 관리 버튼
   - 배포 이력 버튼

3. **탭 동적 관리** (`openTab`)
   - 이미 열린 탭이면 선택
   - 새 탭이면 추가 후 선택

4. **커스텀 레이아웃** (`minWidthLayout`)
   - 왼쪽 메뉴 최소 너비 150px 보장

5. **Content() 함수 변경**
   - `container.NewBorder` 사용
   - 왼쪽: 메뉴 + 구분선 (고정 너비)
   - 중앙: DocTabs (자동 확장)

### `fms_fyne/internal/ui/template_tab.go`

- 주석 및 UI 텍스트 용어 변경 (템플릿 → 룰)

### `fms_fyne/internal/ui/program_tab.go` (신규)

프로그램 관리 탭 신규 구현:

1. **테이블 컬럼**
   - 선택 (체크박스)
   - 이름 (프로그램명)
   - 버전
   - 업로드 경로 (서버 측 저장 경로)
   - 로컬파일 경로
   - 추가(수정)시간

2. **버튼**
   - `[찾기]` - 검색 실행
   - `[삭제]` - 선택된 프로그램 삭제
   - `[추가/수정]` - 프로그램 추가/수정 다이얼로그 표시

3. **검색 기능**
   - 검색 대상: 이름, 버전, 업로드 경로, 로컬파일 경로 (부분 일치)

4. **추가/수정 다이얼로그**
   - 이름, 버전, 업로드경로 입력
   - `[찾아보기...]` 버튼으로 로컬 파일 선택 (tar 파일 필터)

### `fms_fyne/internal/ui/rule_builder.go`

- 주석 용어 변경 (규칙 빌더 → 룰 빌더)

## 기술 참고

### DocTabs vs AppTabs

| 기능 | AppTabs | DocTabs |
|------|---------|---------|
| 닫기 버튼 | ❌ | ✅ |
| 동적 탭 추가/제거 | ⚠️ | ✅ |
| `OnClosed` 콜백 | ❌ | ✅ |
| `CloseIntercept` | ❌ | ✅ |

### minWidthLayout 구현

```go
type minWidthLayout struct {
    minWidth float32
}

func (l *minWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
    if len(objects) == 0 {
        return fyne.NewSize(l.minWidth, 0)
    }
    childMin := objects[0].MinSize()
    width := childMin.Width
    if width < l.minWidth {
        width = l.minWidth
    }
    return fyne.NewSize(width, childMin.Height)
}

func (l *minWidthLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
    for _, obj := range objects {
        obj.Resize(size)
        obj.Move(fyne.NewPos(0, 0))
    }
}
```

## 향후 고려사항

- 탭 헤더 영역에 테두리 추가 (DocTabs 내부 접근 제한으로 현재 미구현)
- 왼쪽 메뉴 아이콘 추가
- 메뉴 선택 상태 하이라이트
- 프로그램 관리 탭 다이얼로그 중첩 처리 (파일 선택 다이얼로그)
