# 규칙 추가 다이얼로그 리팩토링 PRD

## 1. 개요

### 1.1 목적
현재 테이블 하단에 위치한 규칙 추가 탭들을 다이얼로그로 이동하여 UI를 간소화하고, 테이블 영역을 최대화합니다.

### 1.2 배경
- 현재 "룰 빌더" 탭에는 테이블 아래에 "일반 규칙", "Black/White" 탭이 있음
- 현재 "NAT 규칙" 탭에는 테이블 아래에 "DNAT (포트 포워딩)", "SNAT/MASQUERADE" 탭이 있음
- 하단 탭이 화면 공간을 차지하여 테이블 영역이 제한됨

### 1.3 목표
1. 하단 폼 탭들을 다이얼로그로 이동
2. 테이블 상단의 "추가" 버튼으로 다이얼로그 호출
3. 공통 컴포넌트로 구현하여 재사용성 확보
4. 테이블 영역 최대화

---

## 2. 현재 구조 분석

### 2.1 룰 빌더 (RuleBuilder)
```
┌─────────────────────────────────────────────┐
│  [테이블: Chain, Proto, Options, Action...] │
│                                             │
├─────────────────────────────────────────────┤
│  ┌─────────────┬───────────────┐            │
│  │ 일반 규칙   │ Black/White   │  ← 탭     │
│  └─────────────┴───────────────┘            │
│  [규칙 추가 폼 내용...]                      │
└─────────────────────────────────────────────┘
```

**관련 파일:**
- `internal/ui/rule_builder.go` - RuleBuilder 구조체
- `internal/ui/component/rule_form.go` - RuleForm (일반 규칙)
- `internal/ui/component/blackwhite_form.go` - BlackWhiteForm

### 2.2 NAT 빌더 (NATBuilder)
```
┌─────────────────────────────────────────────┐
│  [테이블: Type, Proto, MatchIP, MatchPort...│
│                                             │
├─────────────────────────────────────────────┤
│  ┌──────────────────┬───────────────────┐   │
│  │ DNAT (포트포워딩)│ SNAT/MASQUERADE   │← 탭│
│  └──────────────────┴───────────────────┘   │
│  [NAT 규칙 추가 폼 내용...]                  │
└─────────────────────────────────────────────┘
```

**관련 파일:**
- `internal/ui/nat_builder.go` - NATBuilder 구조체
- `internal/ui/component/dnat_form.go` - DNATForm
- `internal/ui/component/snat_form.go` - SNATForm

---

## 3. 변경 후 구조

### 3.1 룰 빌더 (변경 후)
```
┌─────────────────────────────────────────────┐
│  [+ 추가]                        ← 버튼     │
├─────────────────────────────────────────────┤
│  [테이블: Chain, Proto, Options, Action...] │
│                                             │
│                                             │
│          (테이블 영역 확대)                  │
│                                             │
└─────────────────────────────────────────────┘

         ↓ [+ 추가] 클릭 시

┌─────────────────────────────────────────────┐
│              규칙 추가                  [X] │
├─────────────────────────────────────────────┤
│  ┌─────────────┬───────────────┐            │
│  │ 일반 규칙   │ Black/White   │            │
│  └─────────────┴───────────────┘            │
│  [선택된 탭의 폼 내용...]                    │
│                                             │
│                        [취소]  [추가]       │
└─────────────────────────────────────────────┘
```

### 3.2 NAT 빌더 (변경 후)
```
┌─────────────────────────────────────────────┐
│  [+ 추가]                        ← 버튼     │
├─────────────────────────────────────────────┤
│  [테이블: Type, Proto, MatchIP, MatchPort...│
│                                             │
│          (테이블 영역 확대)                  │
│                                             │
└─────────────────────────────────────────────┘

         ↓ [+ 추가] 클릭 시

┌─────────────────────────────────────────────┐
│            NAT 규칙 추가                [X] │
├─────────────────────────────────────────────┤
│  ┌──────────────────┬───────────────────┐   │
│  │ DNAT (포트포워딩)│ SNAT/MASQUERADE   │   │
│  └──────────────────┴───────────────────┘   │
│  [선택된 탭의 폼 내용...]                    │
│                                             │
│                        [취소]  [추가]       │
└─────────────────────────────────────────────┘
```

---

## 4. 구현 상세

### 4.1 새로운 컴포넌트

#### 4.1.1 RuleAddDialog
**파일:** `internal/ui/component/rule_add_dialog.go`

```go
// RuleAddDialog 규칙 추가 다이얼로그
type RuleAddDialog struct {
    window       fyne.Window
    onAdd        func(*model.FirewallRule)

    // 내부 탭
    generalForm    *RuleForm
    blackWhiteForm *BlackWhiteForm
    tabs           *container.AppTabs

    dialog dialog.Dialog
}
```

**주요 메서드:**
- `NewRuleAddDialog(window, onAdd)` - 생성자
- `Show()` - 다이얼로그 표시
- `Hide()` - 다이얼로그 숨김

#### 4.1.2 NATAddDialog
**파일:** `internal/ui/component/nat_add_dialog.go`

```go
// NATAddDialog NAT 규칙 추가 다이얼로그
type NATAddDialog struct {
    window fyne.Window
    onAdd  func(*model.NATRule)

    // 내부 탭
    dnatForm *DNATForm
    snatForm *SNATForm
    tabs     *container.AppTabs

    dialog dialog.Dialog
}
```

**주요 메서드:**
- `NewNATAddDialog(window, onAdd)` - 생성자
- `Show()` - 다이얼로그 표시
- `Hide()` - 다이얼로그 숨김

### 4.2 수정할 컴포넌트

#### 4.2.1 RuleBuilder 수정
**파일:** `internal/ui/rule_builder.go`

**변경 사항:**
1. `formTabs` 필드 제거
2. `generalForm`, `blackWhiteForm` 필드 제거
3. `ruleAddDialog` 필드 추가
4. `createUI()` 수정: 테이블 상단에 "추가" 버튼 배치
5. `ResetTabs()` 메서드 제거 또는 수정

#### 4.2.2 NATBuilder 수정
**파일:** `internal/ui/nat_builder.go`

**변경 사항:**
1. `formTabs` 필드 제거
2. `dnatForm`, `snatForm` 필드 제거
3. `natAddDialog` 필드 추가
4. `createUI()` 수정: 테이블 상단에 "추가" 버튼 배치
5. `ResetTabs()` 메서드 제거 또는 수정

### 4.3 기존 폼 컴포넌트 수정

기존 폼 컴포넌트들은 **수정 없이 재사용**합니다:
- `RuleForm` - 다이얼로그 내부에서 사용
- `BlackWhiteForm` - 다이얼로그 내부에서 사용
- `DNATForm` - 다이얼로그 내부에서 사용
- `SNATForm` - 다이얼로그 내부에서 사용

단, 각 폼의 "헤더"와 "추가 버튼"은 다이얼로그에서 관리하므로 폼 내부 헤더/버튼을 숨기거나 제거할 수 있습니다.

---

## 5. UI/UX 고려사항

### 5.1 다이얼로그 크기
- 최소 너비: 700px
- 최소 높이: 400px (내용에 따라 자동 조절)

### 5.2 버튼 스타일
- "추가" 버튼: 테이블 상단 왼쪽, `ButtonPrimary` 스타일
- 다이얼로그 "추가" 버튼: `ButtonSuccess` 스타일
- 다이얼로그 "취소" 버튼: `ButtonSecondary` 스타일

### 5.3 다이얼로그 동작
- 규칙 추가 후 다이얼로그 유지 (연속 추가 가능)
- "취소" 또는 X 버튼 클릭 시 다이얼로그 닫기
- 폼 초기화는 규칙 추가 후 자동 수행

---

## 6. 영향 범위

### 6.1 수정 파일
| 파일 | 변경 유형 | 설명 |
|------|----------|------|
| `component/rule_add_dialog.go` | 신규 | 규칙 추가 다이얼로그 |
| `component/nat_add_dialog.go` | 신규 | NAT 규칙 추가 다이얼로그 |
| `rule_builder.go` | 수정 | 하단 탭 제거, 다이얼로그 연동 |
| `nat_builder.go` | 수정 | 하단 탭 제거, 다이얼로그 연동 |
| `component/rule_form.go` | 수정 (선택) | 헤더/버튼 분리 |
| `component/blackwhite_form.go` | 수정 (선택) | 헤더/버튼 분리 |
| `component/dnat_form.go` | 수정 (선택) | 헤더/버튼 분리 |
| `component/snat_form.go` | 수정 (선택) | 헤더/버튼 분리 |

### 6.2 영향 없는 파일
- `rule_table.go` - 변경 없음
- `nat_table.go` - 변경 없음
- `template_tab.go` - 변경 없음

---

## 7. 테스트 항목

1. **룰 빌더 다이얼로그**
   - [ ] "추가" 버튼 클릭 시 다이얼로그 표시
   - [ ] 일반 규칙 탭에서 규칙 추가
   - [ ] Black/White 탭에서 규칙 추가
   - [ ] 추가된 규칙이 테이블에 반영
   - [ ] 다이얼로그 닫기 동작

2. **NAT 빌더 다이얼로그**
   - [ ] "추가" 버튼 클릭 시 다이얼로그 표시
   - [ ] DNAT 탭에서 규칙 추가
   - [ ] SNAT/MASQUERADE 탭에서 규칙 추가
   - [ ] 추가된 규칙이 테이블에 반영
   - [ ] 다이얼로그 닫기 동작

3. **기존 기능 유지**
   - [ ] 테이블에서 규칙 편집
   - [ ] 테이블에서 규칙 삭제
   - [ ] 템플릿 저장/로드 시 규칙 유지
