# 규칙 추가 다이얼로그 리팩토링 체크리스트

> **관련 문서:** [rule-dialog-prd.md](rule-dialog-prd.md)

---

## Phase 1: 기존 폼 컴포넌트 수정

기존 폼들의 내부 헤더와 추가 버튼을 외부에서 제어할 수 있도록 수정합니다.

### 1.1 RuleForm 수정
- [x] `addBtn` 필드를 외부에서 접근 가능하도록 변경 (또는 콜백 방식)
- [x] `submitRule()` 메서드를 public으로 변경 (`SubmitRule()`)
- [ ] 헤더 영역 제거 옵션 추가 (다이얼로그에서 사용 시) - 기존 폼 재사용으로 불필요

### 1.2 BlackWhiteForm 수정
- [x] `submitRule()` 메서드를 public으로 변경 (`SubmitRule()`)
- [ ] 헤더 영역 제거 옵션 추가 - 기존 폼 재사용으로 불필요

### 1.3 DNATForm 수정
- [x] `submitRule()` 메서드를 public으로 변경 (`SubmitRule()`)
- [ ] 헤더 영역 제거 옵션 추가 - 기존 폼 재사용으로 불필요

### 1.4 SNATForm 수정
- [x] `submitRule()` 메서드를 public으로 변경 (`SubmitRule()`)
- [ ] 헤더 영역 제거 옵션 추가 - 기존 폼 재사용으로 불필요

---

## Phase 2: 다이얼로그 컴포넌트 생성

### 2.1 RuleAddDialog 생성
**파일:** `internal/ui/component/rule_add_dialog.go`

- [x] 구조체 정의
  ```go
  type RuleAddDialog struct {
      window         fyne.Window
      onAdd          func(*model.FirewallRule)
      generalForm    *RuleForm
      blackWhiteForm *BlackWhiteForm
      tabs           *container.AppTabs
      popup          *widget.PopUp
  }
  ```
- [x] `NewRuleAddDialog()` 생성자 구현
- [x] 내부 탭 UI 구성 ("일반 규칙", "Black/White")
- [x] `Show()` 메서드 구현
- [x] `Hide()` 메서드 구현
- [x] 다이얼로그 버튼 ("추가", "닫기") 연동
- [x] 규칙 추가 후 폼 초기화

### 2.2 NATAddDialog 생성
**파일:** `internal/ui/component/nat_add_dialog.go`

- [x] 구조체 정의
  ```go
  type NATAddDialog struct {
      window   fyne.Window
      onAdd    func(*model.NATRule)
      dnatForm *DNATForm
      snatForm *SNATForm
      tabs     *container.AppTabs
      popup    *widget.PopUp
  }
  ```
- [x] `NewNATAddDialog()` 생성자 구현
- [x] 내부 탭 UI 구성 ("DNAT (포트 포워딩)", "SNAT/MASQUERADE")
- [x] `Show()` 메서드 구현
- [x] `Hide()` 메서드 구현
- [x] 다이얼로그 버튼 ("추가", "닫기") 연동
- [x] 규칙 추가 후 폼 초기화

---

## Phase 3: Builder 수정

### 3.1 RuleBuilder 수정
**파일:** `internal/ui/rule_builder.go`

- [x] 기존 필드 제거
  - [x] `generalForm *component.RuleForm` 제거
  - [x] `blackWhiteForm *component.BlackWhiteForm` 제거
  - [x] `formTabs *container.AppTabs` 제거
- [x] 새 필드 추가
  - [x] `window fyne.Window` 추가 (다이얼로그용)
  - [x] `addDialog *component.RuleAddDialog` 추가
- [x] `NewRuleBuilder()` 시그니처 변경: `window` 파라미터 추가
- [x] `createUI()` 수정
  - [x] 테이블 상단에 "추가" 버튼 배치
  - [x] 하단 `formTabs` 제거
  - [x] 레이아웃 변경: `container.NewBorder(addButton, nil, nil, nil, table)`
- [x] `ResetTabs()` 메서드 수정 (다이얼로그 탭 초기화로 변경)

### 3.2 NATBuilder 수정
**파일:** `internal/ui/nat_builder.go`

- [x] 기존 필드 제거
  - [x] `dnatForm *component.DNATForm` 제거
  - [x] `snatForm *component.SNATForm` 제거
  - [x] `formTabs *container.AppTabs` 제거
- [x] 새 필드 추가
  - [x] `window fyne.Window` 추가
  - [x] `addDialog *component.NATAddDialog` 추가
- [x] `NewNATBuilder()` 시그니처 변경: `window` 파라미터 추가
- [x] `createUI()` 수정
  - [x] 테이블 상단에 "추가" 버튼 배치
  - [x] 하단 `formTabs` 제거
- [x] `ResetTabs()` 메서드 수정 (다이얼로그 탭 초기화로 변경)

### 3.3 TemplateTab 수정
**파일:** `internal/ui/template_tab.go`

- [x] `NewRuleBuilder()` 호출 시 `window` 파라미터 전달
- [x] `NewNATBuilder()` 호출 시 `window` 파라미터 전달

---

## Phase 4: 빌드 및 테스트

### 4.1 빌드
- [x] `go build` 성공 확인
- [x] 빌드 오류 수정

### 4.2 기능 테스트

#### 룰 빌더 테스트
- [ ] "추가" 버튼 클릭 → 다이얼로그 표시
- [ ] "일반 규칙" 탭 선택 → 폼 표시
- [ ] 일반 규칙 입력 후 "추가" → 테이블에 반영
- [ ] "Black/White" 탭 선택 → 폼 표시
- [ ] Black/White 규칙 입력 후 "추가" → 테이블에 반영
- [ ] "닫기" 버튼 → 다이얼로그 닫힘
- [ ] 연속 추가 테스트 (다이얼로그 유지)

#### NAT 빌더 테스트
- [ ] "추가" 버튼 클릭 → 다이얼로그 표시
- [ ] "DNAT" 탭 선택 → 폼 표시
- [ ] DNAT 규칙 입력 후 "추가" → 테이블에 반영
- [ ] "SNAT/MASQUERADE" 탭 선택 → 폼 표시
- [ ] SNAT 규칙 입력 후 "추가" → 테이블에 반영
- [ ] "닫기" 버튼 → 다이얼로그 닫힘

#### 기존 기능 테스트
- [ ] 테이블에서 규칙 직접 편집
- [ ] 테이블에서 규칙 삭제 (휴지통 버튼)
- [ ] 템플릿 저장 후 로드 → 규칙 유지
- [ ] 텍스트 편집 ↔ 룰 빌더 동기화

---

## Phase 5: 정리 및 커밋

### 5.1 코드 정리
- [ ] 사용하지 않는 import 제거
- [ ] 사용하지 않는 함수/메서드 제거
- [ ] 주석 정리

### 5.2 커밋
- [ ] `git add` - 변경 파일 스테이징
- [ ] `git commit` - 커밋 메시지 작성
  ```
  [Fyne] 규칙 추가 폼을 다이얼로그로 이동

  - RuleAddDialog 컴포넌트 추가 (일반 규칙 + Black/White)
  - NATAddDialog 컴포넌트 추가 (DNAT + SNAT)
  - RuleBuilder에서 하단 탭 제거, 상단에 추가 버튼 배치
  - NATBuilder에서 하단 탭 제거, 상단에 추가 버튼 배치
  - 테이블 영역 최대화
  ```

---

## 진행 상황

| Phase | 항목 | 상태 |
|-------|------|------|
| 1 | 기존 폼 컴포넌트 수정 | ✅ 완료 |
| 2 | 다이얼로그 컴포넌트 생성 | ✅ 완료 |
| 3 | Builder 수정 | ✅ 완료 |
| 4 | 빌드 및 테스트 | 🔄 진행 중 (빌드 완료, 테스트 대기) |
| 5 | 정리 및 커밋 | ⬜ 대기 |

---

## 참고 사항

### Fyne 다이얼로그 중첩 제약
- Fyne의 다이얼로그는 canvas 전체를 덮는 오버레이를 사용
- 다이얼로그 안에서 다른 다이얼로그를 띄우면 충돌 발생
- 해결: `Hide()`/`Show()` 방식으로 부모 다이얼로그 숨김 후 자식 다이얼로그 표시

### 폼 재사용
- 기존 폼 컴포넌트(RuleForm, BlackWhiteForm, DNATForm, SNATForm)는 최대한 재사용
- 폼 내부의 "추가" 버튼은 다이얼로그에서 별도 관리 가능
