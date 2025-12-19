# FMS 개발 절차서 (DEV_PROC)

## 개요

Go Fyne 기반 FMS(Firewall Management System) 데스크톱 애플리케이션 개발 절차서입니다.
**탭 기반 분리형 UI**를 채택하여 구현합니다.

---

## UI 레이아웃 설계

### 메인 윈도우 구조

```
┌─────────────────────────────────────────────────────────────────┐
│  FMS - Firewall Management System                    [설정] [?] │
├─────────────────────────────────────────────────────────────────┤
│  [템플릿 관리]  [장비 관리]  [배포 이력]                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│                      (탭 컨텐츠 영역)                            │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 탭 1: 템플릿 관리

```
┌─────────────────────────────────────────────────────────────────┐
│  [템플릿 관리]  [장비 관리]  [배포 이력]                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────┐  ┌──────────────────────────────────────────┐ │
│  │ 템플릿 목록   │  │ 템플릿 내용                               │ │
│  │              │  │                                          │ │
│  │ ● v2.0      │  │ req|INSERT|...|INPUT|FLUSH|...           │ │
│  │ ○ v1.1      │  │ req|INSERT|...|INPUT|ACCEPT|TCP|...      │ │
│  │ ○ v1.0      │  │ req|INSERT|...|INPUT|DROP|UDP|...        │ │
│  │              │  │                                          │ │
│  │              │  │                                          │ │
│  │              │  │                                          │ │
│  │ [+새로만들기] │  │                                          │ │
│  └──────────────┘  └──────────────────────────────────────────┘ │
│                                                                 │
│  버전명: [v2.0________]  [저장] [삭제] [Export] [Import] [Reset]│
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**컴포넌트 목록:**
| 컴포넌트 | Fyne 위젯 | 설명 |
|----------|-----------|------|
| 템플릿 목록 | `widget.RadioGroup` | 단일 선택 라디오 버튼 |
| 템플릿 내용 | `widget.Entry` (MultiLine) | 여러 줄 텍스트 입력 |
| 버전명 입력 | `widget.Entry` | 단일 줄 텍스트 입력 |
| 새로만들기 | `widget.Button` | 새 템플릿 생성 |
| 저장 | `widget.Button` | 템플릿 저장 |
| 삭제 | `widget.Button` | 선택된 템플릿 삭제 |
| Export | `widget.Button` | JSON 파일로 내보내기 |
| Import | `widget.Button` | JSON 파일에서 가져오기 |
| Reset | `widget.Button` | 데이터 초기화 |

### 탭 2: 장비 관리

```
┌─────────────────────────────────────────────────────────────────┐
│  [템플릿 관리]  [장비 관리]  [배포 이력]                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  배포 템플릿: [v2.0 ▼]        [서버상태확인] [선택장비에 배포]   │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ ☑ │ 장비명(IP)      │ SSH계정 │ 서버상태 │ 배포상태 │ 버전  ││
│  ├─────────────────────────────────────────────────────────────┤│
│  │ ☑ │ 192.168.1.1    │ root   │ ● 정상  │ ✓ 성공  │ v2.0  ││
│  │ ☐ │ 192.168.1.2    │ admin  │ ○ 정지  │ ✗ 실패  │ -     ││
│  │ ☐ │ 192.168.1.3    │ root   │ ● 정상  │ - 대기  │ v1.1  ││
│  │    │                │        │         │         │       ││
│  └─────────────────────────────────────────────────────────────┘│
│                                                                 │
│  [전체선택] [전체해제]                    [+추가] [삭제] [저장]  │
│                                                                 │
│  ── 장비 상세 정보 ──────────────────────────────────────────── │
│  장비 IP: [192.168.1.1___]  SSH 계정: [root____]                │
│  인증방식: ○ SSH키  ● 비밀번호   키경로/비밀번호: [**********]  │
│  SSH 포트: [22__]                                               │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**컴포넌트 목록:**
| 컴포넌트 | Fyne 위젯 | 설명 |
|----------|-----------|------|
| 배포 템플릿 선택 | `widget.Select` | 드롭다운 선택 |
| 장비 테이블 | `widget.Table` | 장비 목록 테이블 |
| 체크박스 | `widget.Check` | 장비 선택 |
| 서버상태확인 | `widget.Button` | SSH 연결 테스트 |
| 배포 | `widget.Button` | 선택 장비에 배포 |
| 장비 상세 | `widget.Form` | 장비 정보 입력 폼 |
| 인증방식 | `widget.RadioGroup` | SSH키/비밀번호 선택 |

### 탭 3: 배포 이력

```
┌─────────────────────────────────────────────────────────────────┐
│  [템플릿 관리]  [장비 관리]  [배포 이력]                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │ 시간                │ 장비           │ 템플릿 │ 결과        ││
│  ├─────────────────────────────────────────────────────────────┤│
│  │ 2024-01-15 14:30:22│ 192.168.1.1   │ v2.0  │ ✓ 성공      ││
│  │ 2024-01-15 14:30:20│ 192.168.1.2   │ v2.0  │ ✗ 실패      ││
│  │ 2024-01-15 10:15:33│ 192.168.1.1   │ v1.1  │ ✓ 성공      ││
│  │                    │               │       │             ││
│  └─────────────────────────────────────────────────────────────┘│
│                                                                 │
│  ── 상세 결과 ───────────────────────────────────────────────── │
│  │ 규칙                              │ 상태   │ 사유          │ │
│  │ req|INSERT|...|FLUSH|...         │ 성공   │ -             │ │
│  │ req|INSERT|...|ACCEPT|TCP|...    │ 성공   │ -             │ │
│  │ req|INSERT|...|DROP|UDP|...      │ 실패   │ 권한 없음      │ │
│                                                                 │
│                                                    [이력 삭제]  │
└─────────────────────────────────────────────────────────────────┘
```

**컴포넌트 목록:**
| 컴포넌트 | Fyne 위젯 | 설명 |
|----------|-----------|------|
| 이력 테이블 | `widget.Table` | 배포 이력 목록 |
| 상세 결과 | `widget.Table` | 규칙별 결과 |
| 이력 삭제 | `widget.Button` | 선택된 이력 삭제 |

---

## 반응형 레이아웃 가이드

데스크탑과 모바일에서 UI가 깨지지 않도록 Fyne의 반응형 컨테이너를 활용합니다.

### 반응형 컨테이너 종류

| 컨테이너 | 함수 | 용도 | 동작 |
|----------|------|------|------|
| 적응형 그리드 | `container.NewAdaptiveGrid(n)` | 화면 방향 대응 | 가로→n열, 세로→n행 자동 전환 |
| 그리드 래핑 | `container.NewGridWrap(size)` | 아이템 자동 배치 | 화면 크기에 맞춰 자동 줄바꿈 |
| 수평 분할 | `container.NewHSplit(left, right)` | 좌우 패널 분할 | 드래그로 비율 조절 가능 |
| 수직 분할 | `container.NewVSplit(top, bottom)` | 상하 패널 분할 | 드래그로 비율 조절 가능 |
| 테두리 | `container.NewBorder(t,b,l,r,center)` | 고정+확장 영역 | 중앙 영역 자동 확장 |
| 스크롤 | `container.NewScroll(content)` | 긴 컨텐츠 | 내용이 넘치면 스크롤 |

### FMS 메인 레이아웃 구현

```go
// internal/ui/layout.go
package ui

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
)

// 메인 레이아웃 - 데스크탑/모바일 모두 대응
func CreateMainLayout(tabs *container.AppTabs) fyne.CanvasObject {
    // 상단 툴바 고정, 탭 컨텐츠 확장
    return container.NewBorder(
        createToolbar(),  // 상단 고정
        nil,              // 하단 없음
        nil,              // 좌측 없음
        nil,              // 우측 없음
        tabs,             // 중앙 (자동 확장)
    )
}
```

### 탭 1: 템플릿 관리 (반응형)

```go
// internal/ui/template_tab.go
func CreateTemplateTab() fyne.CanvasObject {
    // 좌측: 템플릿 목록 (30%), 우측: 템플릿 내용 (70%)
    split := container.NewHSplit(
        createTemplateList(),    // 좌측 패널
        createTemplateContent(), // 우측 패널
    )
    split.Offset = 0.25 // 25% : 75%

    // 하단 버튼 그룹
    buttons := container.NewHBox(
        widget.NewButton("저장", nil),
        widget.NewButton("삭제", nil),
        widget.NewButton("Export", nil),
        widget.NewButton("Import", nil),
    )

    return container.NewBorder(
        nil,     // 상단 없음
        buttons, // 하단 버튼 고정
        nil,
        nil,
        split,   // 중앙 분할 패널 (자동 확장)
    )
}
```

### 탭 2: 장비 관리 (반응형)

```go
// internal/ui/device_tab.go
func CreateDeviceTab() fyne.CanvasObject {
    // 상단: 배포 컨트롤
    deployControl := container.NewHBox(
        widget.NewLabel("배포 템플릿:"),
        widget.NewSelect([]string{"v2.0", "v1.1"}, nil),
        widget.NewButton("서버상태확인", nil),
        widget.NewButton("배포", nil),
    )

    // 중앙: 장비 테이블 (스크롤 가능)
    deviceTable := widget.NewTable(/* ... */)
    scrollableTable := container.NewScroll(deviceTable)

    // 하단: 장비 상세 정보
    deviceDetail := createDeviceDetailForm()

    // 전체 레이아웃
    return container.NewBorder(
        deployControl,   // 상단 고정
        deviceDetail,    // 하단 고정
        nil,
        nil,
        scrollableTable, // 중앙 스크롤 (자동 확장)
    )
}
```

### 모바일 대응: AdaptiveGrid 활용

```go
// 버튼 그룹 - 모바일에서 세로 배치
func createResponsiveButtons() fyne.CanvasObject {
    return container.NewAdaptiveGrid(4, // 가로: 4열, 세로: 4행
        widget.NewButton("저장", nil),
        widget.NewButton("삭제", nil),
        widget.NewButton("Export", nil),
        widget.NewButton("Import", nil),
    )
}

// 폼 필드 - 모바일에서 세로 배치
func createResponsiveForm() fyne.CanvasObject {
    return container.NewAdaptiveGrid(2, // 가로: 2열, 세로: 2행
        widget.NewLabel("IP 주소:"),
        widget.NewEntry(),
        widget.NewLabel("SSH 계정:"),
        widget.NewEntry(),
    )
}
```

### 레이아웃 가이드라인

| 상황 | 권장 컨테이너 | 이유 |
|------|--------------|------|
| 고정 + 확장 영역 | `Border` | 툴바/상태바 고정, 컨텐츠 확장 |
| 좌우/상하 분할 | `HSplit` / `VSplit` | 사용자 조절 가능한 패널 |
| 긴 목록/테이블 | `Scroll` | 내용이 넘칠 때 스크롤 |
| 모바일 대응 버튼 | `AdaptiveGrid` | 화면 회전 시 자동 재배치 |
| 동일 크기 아이템 | `GridWrap` | 자동 줄바꿈 |

---

## 재사용 가능한 커스텀 컴포넌트

자주 사용하는 UI 패턴을 컴포넌트로 분리하여 관리합니다.

### 컴포넌트 목록

| 컴포넌트 | 파일 | 용도 |
|----------|------|------|
| ActionButton | `component/button.go` | 스타일이 적용된 액션 버튼 |
| IconButton | `component/button.go` | 아이콘 버튼 |
| ButtonGroup | `component/button.go` | 버튼 그룹 (가로 배치) |
| LabeledEntry | `component/entry.go` | 라벨이 붙은 입력 필드 |
| LabeledSelect | `component/entry.go` | 라벨이 붙은 드롭다운 |
| LabeledPassword | `component/entry.go` | 라벨이 붙은 비밀번호 입력 |
| StatusBadge | `component/status.go` | 상태 표시 뱃지 (색상) |
| StatusIcon | `component/status.go` | 상태 표시 아이콘 |
| DataTable | `component/table.go` | 데이터 테이블 (정렬, 선택) |
| CheckableList | `component/list.go` | 체크박스가 있는 목록 |
| ConfirmDialog | `component/dialog.go` | 확인 다이얼로그 |
| AlertDialog | `component/dialog.go` | 알림 다이얼로그 |
| ProgressDialog | `component/dialog.go` | 진행률 다이얼로그 |
| Toast | `component/toast.go` | 토스트 알림 |
| Card | `component/card.go` | 카드 컨테이너 |
| Toolbar | `component/toolbar.go` | 툴바 |

### 컴포넌트 구현 예시

#### ActionButton - 스타일이 적용된 버튼

```go
// internal/ui/component/button.go
package component

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/widget"
    "fyne.io/fyne/v2/theme"
    "fyne.io/fyne/v2/canvas"
    "image/color"
)

// 버튼 스타일 타입
type ButtonStyle int

const (
    ButtonPrimary   ButtonStyle = iota // 파란색 (주요 액션)
    ButtonSuccess                       // 초록색 (성공/저장)
    ButtonDanger                        // 빨간색 (삭제/위험)
    ButtonSecondary                     // 회색 (보조)
)

// ActionButton 생성
func NewActionButton(label string, style ButtonStyle, onTap func()) *widget.Button {
    btn := widget.NewButton(label, onTap)

    switch style {
    case ButtonPrimary:
        btn.Importance = widget.HighImportance
    case ButtonDanger:
        btn.Importance = widget.DangerImportance
    case ButtonSuccess:
        btn.Importance = widget.SuccessImportance
    default:
        btn.Importance = widget.MediumImportance
    }

    return btn
}

// ButtonGroup - 버튼들을 가로로 배치
func NewButtonGroup(buttons ...*widget.Button) *fyne.Container {
    objects := make([]fyne.CanvasObject, len(buttons))
    for i, btn := range buttons {
        objects[i] = btn
    }
    return container.NewHBox(objects...)
}
```

#### LabeledEntry - 라벨이 붙은 입력 필드

```go
// internal/ui/component/entry.go
package component

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
)

// LabeledEntry - 라벨과 입력 필드를 함께 표시
type LabeledEntry struct {
    widget.BaseWidget
    Label    *widget.Label
    Entry    *widget.Entry
    OnChange func(string)
}

func NewLabeledEntry(label, placeholder string, onChange func(string)) *fyne.Container {
    lbl := widget.NewLabel(label)
    entry := widget.NewEntry()
    entry.SetPlaceHolder(placeholder)

    if onChange != nil {
        entry.OnChanged = onChange
    }

    return container.NewBorder(nil, nil, lbl, nil, entry)
}

// LabeledPassword - 라벨과 비밀번호 입력 필드
func NewLabeledPassword(label, placeholder string, onChange func(string)) *fyne.Container {
    lbl := widget.NewLabel(label)
    entry := widget.NewPasswordEntry()
    entry.SetPlaceHolder(placeholder)

    if onChange != nil {
        entry.OnChanged = onChange
    }

    return container.NewBorder(nil, nil, lbl, nil, entry)
}

// LabeledSelect - 라벨과 드롭다운
func NewLabeledSelect(label string, options []string, onChange func(string)) *fyne.Container {
    lbl := widget.NewLabel(label)
    sel := widget.NewSelect(options, onChange)

    return container.NewBorder(nil, nil, lbl, nil, sel)
}

// LabeledMultiLineEntry - 라벨과 여러 줄 입력 필드
func NewLabeledMultiLineEntry(label string, onChange func(string)) *fyne.Container {
    lbl := widget.NewLabel(label)
    entry := widget.NewMultiLineEntry()

    if onChange != nil {
        entry.OnChanged = onChange
    }

    return container.NewBorder(lbl, nil, nil, nil, entry)
}
```

#### StatusBadge - 상태 표시 뱃지

```go
// internal/ui/component/status.go
package component

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/canvas"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
    "image/color"
)

// 상태 타입
type StatusType int

const (
    StatusSuccess StatusType = iota // 성공/정상 (초록)
    StatusError                      // 에러/실패 (빨강)
    StatusWarning                    // 경고 (노랑)
    StatusInfo                       // 정보 (파랑)
    StatusPending                    // 대기 (회색)
)

// 상태별 색상
var statusColors = map[StatusType]color.Color{
    StatusSuccess: color.RGBA{R: 40, G: 167, B: 69, A: 255},  // 초록
    StatusError:   color.RGBA{R: 220, G: 53, B: 69, A: 255},  // 빨강
    StatusWarning: color.RGBA{R: 255, G: 193, B: 7, A: 255},  // 노랑
    StatusInfo:    color.RGBA{R: 0, G: 123, B: 255, A: 255},  // 파랑
    StatusPending: color.RGBA{R: 108, G: 117, B: 125, A: 255}, // 회색
}

// StatusBadge - 상태 텍스트와 색상 뱃지
func NewStatusBadge(text string, status StatusType) *fyne.Container {
    circle := canvas.NewCircle(statusColors[status])
    circle.Resize(fyne.NewSize(10, 10))

    label := widget.NewLabel(text)

    return container.NewHBox(circle, label)
}

// StatusIcon - 상태 아이콘만 표시
func NewStatusIcon(status StatusType) *canvas.Circle {
    circle := canvas.NewCircle(statusColors[status])
    circle.Resize(fyne.NewSize(12, 12))
    return circle
}

// 상태 코드 → StatusType 변환
func GetStatusType(code string) StatusType {
    switch code {
    case "running", "success", "ok":
        return StatusSuccess
    case "stop", "fail", "error":
        return StatusError
    case "warning":
        return StatusWarning
    default:
        return StatusPending
    }
}

// 상태 코드 → 표시 텍스트 변환
func GetStatusText(code string) string {
    textMap := map[string]string{
        "running": "정상",
        "stop":    "정지",
        "success": "성공",
        "error":   "확인요망",
        "fail":    "실패",
    }
    if text, ok := textMap[code]; ok {
        return text
    }
    return "-"
}
```

#### DataTable - 데이터 테이블

```go
// internal/ui/component/table.go
package component

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/widget"
)

// DataTable - 데이터 테이블 래퍼
type DataTable struct {
    widget.Table
    Headers    []string
    Data       [][]string
    OnSelect   func(row int)
    SelectedRow int
}

func NewDataTable(headers []string, data [][]string, onSelect func(row int)) *widget.Table {
    table := widget.NewTable(
        // 크기 함수
        func() (int, int) {
            return len(data) + 1, len(headers) // +1 for header
        },
        // 셀 생성 함수
        func() fyne.CanvasObject {
            return widget.NewLabel("")
        },
        // 셀 업데이트 함수
        func(id widget.TableCellID, cell fyne.CanvasObject) {
            label := cell.(*widget.Label)
            if id.Row == 0 {
                // 헤더
                label.SetText(headers[id.Col])
                label.TextStyle = fyne.TextStyle{Bold: true}
            } else {
                // 데이터
                if id.Row-1 < len(data) && id.Col < len(data[id.Row-1]) {
                    label.SetText(data[id.Row-1][id.Col])
                }
            }
        },
    )

    if onSelect != nil {
        table.OnSelected = func(id widget.TableCellID) {
            if id.Row > 0 { // 헤더 제외
                onSelect(id.Row - 1)
            }
        }
    }

    return table
}
```

#### CheckableList - 체크박스가 있는 목록

```go
// internal/ui/component/list.go
package component

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
)

// CheckableItem - 체크 가능한 아이템
type CheckableItem struct {
    ID      int
    Label   string
    Checked bool
}

// CheckableList - 체크박스가 있는 목록
type CheckableList struct {
    Items      []CheckableItem
    OnChange   func(id int, checked bool)
    container  *fyne.Container
}

func NewCheckableList(items []CheckableItem, onChange func(id int, checked bool)) *fyne.Container {
    vbox := container.NewVBox()

    for _, item := range items {
        itemCopy := item // 클로저용 복사
        check := widget.NewCheck(item.Label, func(checked bool) {
            if onChange != nil {
                onChange(itemCopy.ID, checked)
            }
        })
        check.Checked = item.Checked
        vbox.Add(check)
    }

    return vbox
}

// SelectAll - 전체 선택
func SelectAll(list *fyne.Container, checked bool) {
    for _, obj := range list.Objects {
        if check, ok := obj.(*widget.Check); ok {
            check.SetChecked(checked)
        }
    }
}
```

#### ConfirmDialog - 확인 다이얼로그

```go
// internal/ui/component/dialog.go
package component

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/dialog"
    "fyne.io/fyne/v2/widget"
)

// ConfirmDialog - 확인/취소 다이얼로그
func ShowConfirmDialog(parent fyne.Window, title, message string, onConfirm func()) {
    dialog.ShowConfirm(title, message, func(confirmed bool) {
        if confirmed && onConfirm != nil {
            onConfirm()
        }
    }, parent)
}

// AlertDialog - 알림 다이얼로그
func ShowAlertDialog(parent fyne.Window, title, message string) {
    dialog.ShowInformation(title, message, parent)
}

// ErrorDialog - 에러 다이얼로그
func ShowErrorDialog(parent fyne.Window, title string, err error) {
    dialog.ShowError(err, parent)
}

// ProgressDialog - 진행률 다이얼로그
func ShowProgressDialog(parent fyne.Window, title string) *dialog.ProgressDialog {
    progress := dialog.NewProgress(title, "처리 중...", parent)
    progress.Show()
    return progress
}

// InputDialog - 입력 다이얼로그
func ShowInputDialog(parent fyne.Window, title, placeholder string, onSubmit func(string)) {
    entry := widget.NewEntry()
    entry.SetPlaceHolder(placeholder)

    dialog.ShowForm(title, "확인", "취소", []*widget.FormItem{
        widget.NewFormItem("", entry),
    }, func(confirmed bool) {
        if confirmed && onSubmit != nil {
            onSubmit(entry.Text)
        }
    }, parent)
}
```

#### Toast - 토스트 알림

```go
// internal/ui/component/toast.go
package component

import (
    "time"
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/canvas"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
    "image/color"
)

// Toast 타입
type ToastType int

const (
    ToastSuccess ToastType = iota
    ToastError
    ToastInfo
)

// Toast 색상
var toastColors = map[ToastType]color.Color{
    ToastSuccess: color.RGBA{R: 40, G: 167, B: 69, A: 230},
    ToastError:   color.RGBA{R: 220, G: 53, B: 69, A: 230},
    ToastInfo:    color.RGBA{R: 0, G: 123, B: 255, A: 230},
}

// ShowToast - 토스트 메시지 표시 (일정 시간 후 사라짐)
func ShowToast(parent fyne.Window, message string, toastType ToastType, duration time.Duration) {
    bg := canvas.NewRectangle(toastColors[toastType])
    label := widget.NewLabel(message)
    label.Alignment = fyne.TextAlignCenter

    toast := container.NewStack(bg, container.NewCenter(label))
    toast.Resize(fyne.NewSize(300, 50))

    // 팝업으로 표시
    popup := widget.NewModalPopUp(toast, parent.Canvas())
    popup.Show()

    // 일정 시간 후 숨김
    go func() {
        time.Sleep(duration)
        popup.Hide()
    }()
}

// 편의 함수들
func ShowSuccessToast(parent fyne.Window, message string) {
    ShowToast(parent, message, ToastSuccess, 2*time.Second)
}

func ShowErrorToast(parent fyne.Window, message string) {
    ShowToast(parent, message, ToastError, 3*time.Second)
}

func ShowInfoToast(parent fyne.Window, message string) {
    ShowToast(parent, message, ToastInfo, 2*time.Second)
}
```

#### Card - 카드 컨테이너

```go
// internal/ui/component/card.go
package component

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
)

// Card - 제목이 있는 카드 컨테이너
func NewCard(title string, content fyne.CanvasObject) *widget.Card {
    return widget.NewCard(title, "", content)
}

// CardWithActions - 제목과 액션 버튼이 있는 카드
func NewCardWithActions(title string, content fyne.CanvasObject, actions ...fyne.CanvasObject) *fyne.Container {
    header := container.NewBorder(
        nil, nil, widget.NewLabel(title), container.NewHBox(actions...),
    )

    return container.NewBorder(header, nil, nil, nil, content)
}
```

#### Toolbar - 툴바

```go
// internal/ui/component/toolbar.go
package component

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
)

// ToolbarItem - 툴바 아이템
type ToolbarItem struct {
    Label   string
    Icon    fyne.Resource
    OnTap   func()
}

// NewToolbar - 툴바 생성
func NewToolbar(items ...ToolbarItem) *widget.Toolbar {
    toolbarItems := make([]widget.ToolbarItem, len(items))

    for i, item := range items {
        itemCopy := item
        if item.Icon != nil {
            toolbarItems[i] = widget.NewToolbarAction(item.Icon, itemCopy.OnTap)
        } else {
            toolbarItems[i] = widget.NewToolbarAction(nil, itemCopy.OnTap)
        }
    }

    return widget.NewToolbar(toolbarItems...)
}

// NewToolbarWithSpacer - 스페이서가 포함된 툴바
func NewToolbarWithSpacer(left []ToolbarItem, right []ToolbarItem) *widget.Toolbar {
    items := []widget.ToolbarItem{}

    for _, item := range left {
        itemCopy := item
        items = append(items, widget.NewToolbarAction(item.Icon, itemCopy.OnTap))
    }

    items = append(items, widget.NewToolbarSpacer())

    for _, item := range right {
        itemCopy := item
        items = append(items, widget.NewToolbarAction(item.Icon, itemCopy.OnTap))
    }

    return widget.NewToolbar(items...)
}
```

---

## 프로젝트 구조

```
smartfw_add_rules/
├── CLAUDE.md              # Claude 개발 가이드
├── FMS_SPEC.md            # 기능 명세서
├── DEV_PROC.md            # 개발 절차서 (현재 파일)
├── index.html             # 원본 웹앱 (참조용, 수정금지)
├── smartfw_hs/            # 커널 모듈 (참조용, 수정금지)
│
├── go.mod                 # Go 모듈 정의
├── go.sum                 # 의존성 체크섬
├── main.go                # 앱 진입점
│
├── internal/
│   ├── model/             # 데이터 모델
│   │   ├── template.go    # Template 구조체
│   │   ├── firewall.go    # Firewall 구조체
│   │   └── history.go     # DeployHistory 구조체
│   │
│   ├── storage/           # 데이터 저장소
│   │   ├── storage.go     # 저장소 인터페이스
│   │   └── json_store.go  # JSON 파일 기반 구현
│   │
│   ├── ssh/               # SSH 연결 관리
│   │   ├── client.go      # SSH 클라이언트
│   │   ├── key_auth.go    # SSH 키 인증
│   │   └── password_auth.go # 비밀번호 인증 (AES 암호화)
│   │
│   ├── deploy/            # 배포 로직
│   │   ├── deployer.go    # 배포 실행기
│   │   └── health.go      # 서버 상태 확인
│   │
│   └── ui/                # UI 레이어
│       ├── app.go         # 메인 앱 윈도우
│       ├── tabs.go        # 탭 컨테이너
│       ├── template_tab.go # 템플릿 관리 탭
│       ├── device_tab.go  # 장비 관리 탭
│       ├── history_tab.go # 배포 이력 탭
│       │
│       └── component/     # 재사용 가능한 UI 컴포넌트
│           ├── button.go  # ActionButton, ButtonGroup
│           ├── entry.go   # LabeledEntry, LabeledSelect
│           ├── status.go  # StatusBadge, StatusIcon
│           ├── table.go   # DataTable
│           ├── list.go    # CheckableList
│           ├── dialog.go  # ConfirmDialog, AlertDialog
│           ├── toast.go   # Toast 알림
│           ├── card.go    # Card 컨테이너
│           └── toolbar.go # Toolbar
│
└── data/                  # 데이터 저장 디렉토리
    ├── templates.json     # 템플릿 데이터
    ├── firewalls.json     # 장비 데이터
    └── history.json       # 배포 이력 데이터
```

---

## 개발 단계

### Phase 1: 프로젝트 초기화 및 기본 구조

**목표**: 프로젝트 셋업 및 기본 윈도우 생성

**작업 항목**:
1. Go 모듈 초기화 (`go mod init fms`)
2. Fyne 패키지 설치 (`go get fyne.io/fyne/v2`)
3. SSH 패키지 설치 (`go get golang.org/x/crypto/ssh`)
4. 디렉토리 구조 생성
5. main.go 작성 (기본 윈도우)
6. 탭 컨테이너 구현

**완료 기준**:
- 앱 실행 시 3개 탭이 있는 빈 윈도우 표시

---

### Phase 2: 데이터 모델 및 저장소

**목표**: 데이터 구조 정의 및 JSON 파일 저장소 구현

**작업 항목**:
1. Template 구조체 정의
2. Firewall 구조체 정의
3. DeployHistory 구조체 정의
4. Storage 인터페이스 정의
5. JSON 파일 기반 저장소 구현
   - Load/Save 함수
   - CRUD 함수

**데이터 모델**:
```go
// internal/model/template.go
type Template struct {
    Version  string `json:"version"`
    Contents string `json:"contents"`
}

// internal/model/firewall.go
type Firewall struct {
    Index        int    `json:"index"`
    DeviceName   string `json:"deviceName"`
    ServerStatus string `json:"serverStatus"`
    DeployStatus string `json:"deployStatus"`
    Version      string `json:"version"`
    AuthType     string `json:"authType"`
    SSHUser      string `json:"sshUser"`
    SSHKeyPath   string `json:"sshKeyPath"`
    SSHPassword  string `json:"sshPassword"`
    SSHPort      int    `json:"sshPort"`
}

// internal/model/history.go
type DeployHistory struct {
    ID         int          `json:"id"`
    Timestamp  time.Time    `json:"timestamp"`
    DeviceIP   string       `json:"deviceIp"`
    TemplateVer string      `json:"templateVersion"`
    Status     string       `json:"status"`
    Results    []RuleResult `json:"results"`
}

type RuleResult struct {
    Rule   string `json:"rule"`
    Status string `json:"status"`
    Reason string `json:"reason"`
}
```

**완료 기준**:
- 템플릿/장비/이력 데이터 저장 및 로드 정상 동작
- 앱 재시작 후 데이터 유지

---

### Phase 3: 템플릿 관리 탭 UI

**목표**: 템플릿 관리 탭 UI 및 기능 구현

**작업 항목**:
1. 템플릿 목록 (RadioGroup) 구현
2. 템플릿 내용 편집기 (MultiLine Entry) 구현
3. 버전명 입력 필드 구현
4. 버튼 구현
   - 새로만들기: 빈 템플릿 생성
   - 저장: 템플릿 저장 (신규/수정)
   - 삭제: 선택된 템플릿 삭제
5. Export/Import 기능 구현
   - Export: 파일 다이얼로그 → JSON 저장
   - Import: 파일 다이얼로그 → JSON 로드
6. Reset 기능 구현 (확인 다이얼로그 포함)

**완료 기준**:
- 템플릿 CRUD 정상 동작
- Export/Import 정상 동작
- 데이터 유효성 검증 (빈 버전명 방지 등)

---

### Phase 4: SSH 연결 모듈

**목표**: SSH 키 인증 및 비밀번호 인증 구현

**작업 항목**:
1. SSH 클라이언트 인터페이스 정의
2. SSH 키 인증 구현
   - 키 파일 읽기
   - ParsePrivateKey
   - PublicKeys 인증
3. 비밀번호 인증 구현 (선택적)
   - AES 암호화/복호화
   - Password 인증
4. SSH 연결 테스트 함수 구현
5. 원격 명령 실행 함수 구현

**완료 기준**:
- SSH 키로 원격 서버 연결 성공
- 원격 명령 실행 및 결과 수신 성공

---

### Phase 5: 장비 관리 탭 UI

**목표**: 장비 관리 탭 UI 및 기능 구현

**작업 항목**:
1. 배포 템플릿 선택 드롭다운 구현
2. 장비 테이블 구현
   - 체크박스 컬럼
   - 장비 정보 컬럼
   - 상태 표시 (색상/아이콘)
3. 장비 상세 정보 폼 구현
   - IP, SSH계정, 인증방식, 키경로/비밀번호, 포트
4. 버튼 구현
   - 전체선택/전체해제
   - 추가/삭제/저장
5. 서버상태확인 기능 구현
   - 선택된 장비에 SSH 연결 테스트
   - 결과 표시 (정상/정지)

**완료 기준**:
- 장비 CRUD 정상 동작
- 서버 상태 확인 정상 동작
- 장비 선택 시 상세 정보 표시

---

### Phase 6: 배포 기능

**목표**: 템플릿을 선택된 장비에 배포하는 기능 구현

**작업 항목**:
1. 배포 실행 함수 구현
   - 템플릿 내용을 줄 단위로 분리
   - 각 규칙을 `echo 'rule' > /proc/smartfw` 형태로 실행
2. 배포 진행 다이얼로그 구현
   - 진행률 표시
   - 취소 버튼
3. 배포 결과 처리
   - 성공/실패 상태 업데이트
   - 배포 이력 저장
4. 에러 처리
   - SSH 연결 실패
   - 명령 실행 실패
   - 타임아웃

**배포 로직**:
```go
func Deploy(fw Firewall, template Template) (*DeployResult, error) {
    // 1. SSH 연결
    client, err := ConnectSSH(fw)
    if err != nil {
        return nil, err
    }
    defer client.Close()

    // 2. 규칙 분리
    rules := strings.Split(template.Contents, "\n")

    // 3. 각 규칙 실행
    results := []RuleResult{}
    for _, rule := range rules {
        if strings.TrimSpace(rule) == "" {
            continue
        }

        cmd := fmt.Sprintf("echo '%s' > /proc/smartfw", rule)
        err := ExecuteCommand(client, cmd)

        result := RuleResult{Rule: rule}
        if err != nil {
            result.Status = "error"
            result.Reason = err.Error()
        } else {
            result.Status = "ok"
        }
        results = append(results, result)
    }

    return &DeployResult{Results: results}, nil
}
```

**완료 기준**:
- 선택된 장비에 템플릿 배포 성공
- 배포 결과 화면에 표시
- 배포 이력 저장

---

### Phase 7: 배포 이력 탭 UI

**목표**: 배포 이력 조회 및 상세 결과 확인 기능 구현

**작업 항목**:
1. 배포 이력 테이블 구현
   - 시간, 장비, 템플릿, 결과 컬럼
2. 이력 선택 시 상세 결과 표시
   - 규칙별 성공/실패 상태
   - 실패 사유
3. 이력 삭제 기능 구현

**완료 기준**:
- 배포 이력 목록 표시
- 상세 결과 확인 가능
- 이력 삭제 가능

---

### Phase 8: 알림 및 다이얼로그

**목표**: 사용자 피드백 UI 구현

**작업 항목**:
1. 확인 다이얼로그 구현
   - 삭제 확인
   - Reset 확인
   - 배포 확인
2. 알림 메시지 구현
   - 성공 알림 (초록색)
   - 에러 알림 (빨간색)
   - 정보 알림 (파란색)
3. 진행률 다이얼로그 구현
   - 배포 진행 중 표시
   - 서버 상태 확인 중 표시

**완료 기준**:
- 모든 주요 작업에 확인/결과 알림 표시
- 진행 중인 작업 시각적 표시

---

### Phase 9: 설정 및 환경설정

**목표**: 앱 설정 기능 구현

**작업 항목**:
1. 설정 다이얼로그 구현
   - 기본 SSH 포트
   - 기본 SSH 키 경로
   - 데이터 저장 경로
   - SSH 연결 타임아웃
2. 설정 저장/로드 구현
3. 도움말 다이얼로그 구현

**완료 기준**:
- 설정 변경 및 저장 가능
- 앱 재시작 후 설정 유지

---

### Phase 10: 테스트 및 마무리

**목표**: 전체 기능 테스트 및 버그 수정

**작업 항목**:
1. macOS에서 기능 테스트
2. Linux용 크로스 컴파일
3. Linux 서버에서 실제 테스트
   - SSH 연결 테스트
   - 배포 테스트 (실제 /proc/smartfw)
4. 버그 수정
5. 코드 정리 및 주석 추가

**테스트 체크리스트**:
- [ ] 템플릿 생성/수정/삭제
- [ ] 템플릿 Export/Import
- [ ] 장비 추가/수정/삭제
- [ ] SSH 키 인증 연결
- [ ] 서버 상태 확인
- [ ] 템플릿 배포
- [ ] 배포 이력 확인
- [ ] 설정 변경

**완료 기준**:
- 모든 기능 정상 동작
- Linux 서버에서 실제 배포 성공

---

## 빌드 명령어

### 개발 (macOS)
```bash
# 의존성 설치
go mod tidy

# 실행
go run main.go

# 빌드
go build -o fms .
```

### 배포 (Linux x86_64)
```bash
# Linux용 크로스 컴파일
GOOS=linux GOARCH=amd64 go build -o fms-linux .

# 서버로 전송
scp fms-linux user@192.168.x.x:/path/to/

# 서버에서 실행
ssh user@192.168.x.x
chmod +x fms-linux
./fms-linux
```

---

## 참조 문서

- [CLAUDE.md](CLAUDE.md) - Claude 개발 가이드 (중요 지침)
- [FMS_SPEC.md](FMS_SPEC.md) - 기능 명세서 (상세 구현 코드)
- [index.html](index.html) - 원본 웹앱 (참조용)
- [smartfw_hs/Makefile](smartfw_hs/Makefile) - 규칙 포맷 참조
