---
name: fyne-docs
description: |
  Fyne GUI 라이브러리 코드 작성 시 자동 활성화.
  fyne.io/fyne/v2 import, widget, container, dialog, canvas 관련 작업 감지 시 사용.
---

# Fyne Documentation Skill

## 자동 활성화 조건

다음 상황에서 이 Skill이 자동으로 활성화됩니다:

- `fyne.io/fyne/v2` 패키지 관련 코드 작성
- Fyne widget, container, dialog, canvas 구현
- Go GUI 개발 관련 질문
- DEV_PROC.md의 커스텀 컴포넌트 구현

## 문서 조회 방법

### WebFetch를 이용한 공식 문서 조회

Fyne 공식 문서 사이트에서 직접 조회합니다.

**기본 URL**: `https://docs.fyne.io/`

| 주제 | URL | 설명 |
|------|-----|------|
| 시작하기 | `https://docs.fyne.io/started/` | 설치 및 첫 앱 |
| Widget | `https://docs.fyne.io/widget/` | 위젯 목록 및 사용법 |
| Container | `https://docs.fyne.io/container/` | 레이아웃 컨테이너 |
| Canvas | `https://docs.fyne.io/canvas/` | 그래픽 요소 |
| Dialog | `https://docs.fyne.io/dialog/` | 다이얼로그 |
| Binding | `https://docs.fyne.io/binding/` | 데이터 바인딩 |
| Theme | `https://docs.fyne.io/theme/` | 테마 커스터마이징 |
| Extend | `https://docs.fyne.io/extend/` | 커스텀 위젯 만들기 |

### 개별 위젯 문서

| 위젯 | URL |
|------|-----|
| Button | `https://docs.fyne.io/widget/button` |
| Entry | `https://docs.fyne.io/widget/entry` |
| Label | `https://docs.fyne.io/widget/label` |
| List | `https://docs.fyne.io/widget/list` |
| Table | `https://docs.fyne.io/widget/table` |
| Tree | `https://docs.fyne.io/widget/tree` |
| Select | `https://docs.fyne.io/widget/select` |
| Check | `https://docs.fyne.io/widget/check` |
| Form | `https://docs.fyne.io/widget/form` |
| Toolbar | `https://docs.fyne.io/widget/toolbar` |
| Menu | `https://docs.fyne.io/widget/menu` |
| Tabs | `https://docs.fyne.io/container/tabs` |

### API 레퍼런스

| 패키지 | URL |
|--------|-----|
| fyne | `https://pkg.go.dev/fyne.io/fyne/v2` |
| widget | `https://pkg.go.dev/fyne.io/fyne/v2/widget` |
| container | `https://pkg.go.dev/fyne.io/fyne/v2/container` |
| dialog | `https://pkg.go.dev/fyne.io/fyne/v2/dialog` |
| canvas | `https://pkg.go.dev/fyne.io/fyne/v2/canvas` |
| theme | `https://pkg.go.dev/fyne.io/fyne/v2/theme` |
| data/binding | `https://pkg.go.dev/fyne.io/fyne/v2/data/binding` |

## 사용 예시

Fyne 코드 작성 시:

```
1. WebFetch 도구로 해당 위젯/기능 문서 조회
   예: WebFetch("https://docs.fyne.io/widget/table", "Table 위젯 사용법 요약")

2. 필요시 pkg.go.dev에서 API 상세 확인
   예: WebFetch("https://pkg.go.dev/fyne.io/fyne/v2/widget#Table", "Table 구조체와 메서드")

3. 프로젝트의 DEV_PROC.md 패턴과 일관성 유지
```

## 주요 Fyne 패키지

| 패키지 | 용도 |
|--------|------|
| `fyne.io/fyne/v2` | 코어 타입 (App, Window, Canvas, Size, Position) |
| `fyne.io/fyne/v2/widget` | UI 위젯 (Button, Entry, Label, Table, List) |
| `fyne.io/fyne/v2/container` | 레이아웃 (VBox, HBox, Border, Grid, Split, Tabs) |
| `fyne.io/fyne/v2/dialog` | 다이얼로그 (Confirm, Information, Error, Progress) |
| `fyne.io/fyne/v2/canvas` | 그래픽 (Rectangle, Circle, Text, Image) |
| `fyne.io/fyne/v2/theme` | 테마 및 아이콘 |
| `fyne.io/fyne/v2/data/binding` | 데이터 바인딩 |

## 프로젝트 참조

이 프로젝트의 Fyne 관련 문서:
- `fms_fyne/DEV_PROC.md` - 커스텀 컴포넌트 구현 예시
- `fms_fyne/FMS_SPEC.md` - UI 기능 명세

## 자주 사용하는 패턴

### 커스텀 위젯 기본 구조

```go
type MyWidget struct {
    widget.BaseWidget
    // 필드들
}

func NewMyWidget() *MyWidget {
    w := &MyWidget{}
    w.ExtendBaseWidget(w)
    return w
}

func (w *MyWidget) CreateRenderer() fyne.WidgetRenderer {
    // 렌더러 구현
}
```

### 테이블 생성

```go
table := widget.NewTable(
    func() (int, int) { return rows, cols },           // 크기
    func() fyne.CanvasObject { return widget.NewLabel("") }, // 셀 생성
    func(id widget.TableCellID, obj fyne.CanvasObject) {     // 셀 업데이트
        obj.(*widget.Label).SetText(data[id.Row][id.Col])
    },
)
```

### 컨테이너 레이아웃

```go
// VBox - 세로 정렬
container.NewVBox(widget1, widget2, widget3)

// HBox - 가로 정렬
container.NewHBox(widget1, widget2, widget3)

// Border - 테두리 레이아웃
container.NewBorder(top, bottom, left, right, center)

// Grid - 그리드 레이아웃
container.NewGridWithColumns(3, widgets...)
```
