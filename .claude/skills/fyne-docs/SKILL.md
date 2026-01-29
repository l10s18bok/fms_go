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

## 다이얼로그 및 팝업 중첩

### 중첩 제약사항

Fyne의 다이얼로그와 모달 팝업은 canvas 전체를 덮는 오버레이를 사용하므로 **동시에 여러 개를 띄우면 문제가 발생**합니다.

| 컴포넌트 | 동시 중첩 | 비고 |
|----------|-----------|------|
| dialog.* | ❌ | Modal 오버레이 충돌 |
| widget.PopUp (비모달) | ⚠️ | 가능하나 권장 안함 |
| widget.ModalPopUp | ❌ | 오버레이 충돌 |
| widget.PopUpMenu | ⚠️ | 메뉴 전용 |

### Dialog 인터페이스 메서드

```go
type Dialog interface {
    Show()              // 다이얼로그 표시
    Hide()              // 다이얼로그 숨김 (상태 유지)
    Dismiss()           // 다이얼로그 닫기 (v2.6+)
    SetOnClosed(func()) // 닫힘 콜백 설정
    Refresh()
    Resize(fyne.Size)
    MinSize() fyne.Size
    SetDismissText(string)
}
```

### 해결 패턴: Hide/Show 방식

다이얼로그 안에서 다른 다이얼로그(예: FileOpen)를 띄워야 할 때는 **부모 다이얼로그를 숨긴 후 자식 다이얼로그를 표시**하고, 자식이 닫히면 부모를 다시 표시합니다.

```go
// 부모 다이얼로그 (예: 장비 추가/수정)
parentDialog := dialog.NewCustom("장비 추가/수정", "취소", content, window)

// [찾아보기...] 버튼 클릭 시
browseBtn.OnTapped = func() {
    parentDialog.Hide()  // 1. 부모 다이얼로그 숨김

    fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
        if reader != nil {
            filePath = reader.URI().Path()
            fileLabel.SetText(filePath)
            reader.Close()
        }
        parentDialog.Show()  // 3. 파일 선택 후 부모 다이얼로그 다시 표시
    }, window)
    fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".pem", ".key"}))
    fileDialog.Show()  // 2. 파일 다이얼로그 표시
}

parentDialog.Show()
```

**핵심 포인트:**
- `Hide()`는 다이얼로그를 닫지 않고 숨기기만 하므로 **입력값이 유지**됨
- FileOpen 콜백에서 부모 다이얼로그를 `Show()`로 다시 표시
- 같은 다이얼로그 인스턴스를 재사용하므로 상태 보존

### 대안: 인라인 편집 방식

다이얼로그 중첩을 완전히 피하려면 다이얼로그 대신 **테이블 인라인 편집** 방식을 사용:

```
[추가] 클릭 → 테이블에 빈 행 추가 → 행 내에서 직접 편집
                                    → [찾아보기...] 버튼은 메인 윈도우에서 FileOpen 호출
```

## HBox/Border 수직 정렬 문제

`HBox`와 `Border`는 자식 위젯을 세로 방향으로 **늘려서(stretch)** 컨테이너 높이에 맞춥니다. 따라서 높이가 다른 위젯(예: Label과 Button)을 나란히 배치하면 수직 중앙이 맞지 않을 수 있습니다.

**해결:** `container.NewCenter()`로 각 위젯을 감싸면 원래 크기를 유지하면서 중앙 정렬됩니다.

```go
// ❌ 수직 정렬 안 맞음
titleRow := container.NewHBox(titleLabel, helpBtn)

// ✅ 수직 중앙 정렬
titleRow := container.NewHBox(container.NewCenter(titleLabel), container.NewCenter(helpBtn))
```
