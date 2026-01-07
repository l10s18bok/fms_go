# 규칙 빌더 구현 체크리스트

## 문서 정보
- **버전**: 1.0
- **작성일**: 2026-01-05
- **관련 PRD**: [rule-builder-prd.md](./rule-builder-prd.md)

---

## Phase 0: 문서 준비

- [x] `docs/` 디렉토리 생성
- [x] `docs/rule-builder-prd.md` 작성
- [x] `docs/rule-builder-checklist.md` 작성

---

## Phase 1: 데이터 모델 및 파서

### Step 1.1: `internal/model/rule.go` 생성

- [x] Chain 상수 정의
  - [x] ChainINPUT (0)
  - [x] ChainOUTPUT (1)
  - [x] ChainFORWARD (2)
  - [x] ChainPREROUTING (3)
  - [x] ChainPOSTROUTING (4)
- [x] Protocol 상수 정의
  - [x] ProtocolTCP (6)
  - [x] ProtocolUDP (17)
  - [x] ProtocolICMP (1)
  - [x] ProtocolANY (255)
- [x] Action 상수 정의
  - [x] ActionDROP (0)
  - [x] ActionACCEPT (1)
  - [x] ActionREJECT (2)
- [x] FirewallRule 구조체 정의
- [x] 문자열 변환 헬퍼 메서드
  - [x] ChainToString()
  - [x] StringToChain()
  - [x] ProtocolToString()
  - [x] StringToProtocol()
  - [x] ActionToString()
  - [x] StringToAction()
  - [x] GetChainOptions() - UI Select용
  - [x] GetProtocolOptions() - UI Select용
  - [x] GetActionOptions() - UI Select용

### Step 1.2: `internal/parser/rule_parser.go` 생성

- [x] ParseLine(line string) (*FirewallRule, error)
  - [x] 빈 줄 처리
  - [x] 주석 라인(#) 처리
  - [x] agent 형식 파싱
  - [x] -c= (chain) 파싱
  - [x] -p= (protocol) 파싱
  - [x] -a= (action) 파싱
  - [x] --dport= 파싱
  - [x] --sip= 파싱
  - [x] --dip= 파싱
  - [x] --black 플래그 파싱
  - [x] --white 플래그 파싱
- [x] RuleToLine(rule *FirewallRule) string
  - [x] agent 형식으로 변환
  - [x] 필수 필드 출력
  - [x] 선택 필드 조건부 출력
- [x] ParseTextToRules(text string) ([]*FirewallRule, []error)
  - [x] 줄 단위 분리
  - [x] 각 줄 파싱
  - [x] 오류 수집
- [x] RulesToText(rules []*FirewallRule) string
  - [x] 각 규칙 변환
  - [x] 줄바꿈으로 연결

---

## Phase 2: UI 컴포넌트

### Step 2.1: `internal/ui/component/rule_table.go` 생성 (widget.Table 사용)

- [x] RuleTable 구조체 정의
  - [x] rules []*model.FirewallRule
  - [x] table *widget.Table
  - [x] onChange func()
  - [x] lastWidth float32 (중복 업데이트 방지)
- [x] widget.Table 생성 (NewTableWithHeaders 사용)
  - [x] Length 콜백: (rows, cols) 반환
  - [x] CreateCell 콜백: Stack에 모든 위젯 타입 포함
    - [x] canvas.Rectangle 배경 (Select hover 투명 문제 해결)
    - [x] 컬럼 0: 삭제 버튼 (Button)
    - [x] 컬럼 1: Chain (Select)
    - [x] 컬럼 2: Protocol (Select)
    - [x] 컬럼 3: 옵션 (Hyperlink - 클릭 시 팝업)
    - [x] 컬럼 4: Action (Select)
    - [x] 컬럼 5: DPort (Entry)
    - [x] 컬럼 6: SIP (Entry)
    - [x] 컬럼 7: DIP (Entry)
    - [x] 컬럼 8: Black (Check)
    - [x] 컬럼 9: White (Check)
  - [x] UpdateCell 콜백: 셀 데이터 업데이트
  - [x] CreateHeader/UpdateHeader: 헤더 설정
- [x] 컬럼 너비 자동 조절 (비율 기반)
  - [x] columnRatios 배열 정의 (합계 1.0)
  - [x] table.SetColumnWidth() 사용
  - [x] Resize() 메서드로 창 크기에 따라 비율 조절
- [x] NewRuleTable() 생성자
- [x] AddRule(rule *FirewallRule) 메서드
- [x] RemoveRule(index int) 메서드
- [x] GetRules() []*FirewallRule 메서드
- [x] SetRules(rules []*FirewallRule) 메서드
- [x] Clear() 메서드
- [x] Content() 메서드
- [x] Refresh() 메서드
- [x] CreateRenderer() 메서드 (커스텀 위젯)

### Step 2.2: 기존 파일 정리

- [x] `rule_row.go` 삭제
- [x] `rule_list.go` 삭제
- [x] `rule_builder.go`에서 RuleTable 사용하도록 수정

### Step 2.3: `internal/ui/component/rule_form.go` 생성

- [x] RuleForm 구조체 정의
  - [x] onAdd func(*FirewallRule)
- [x] UI 요소 생성
  - [x] Chain Select (기본값: INPUT)
  - [x] Protocol Select (기본값: TCP)
  - [x] Action Select (기본값: DROP)
  - [x] DPort Entry
  - [x] SIP Entry
  - [x] DIP Entry
  - [x] Black Check
  - [x] White Check
  - [x] 추가 버튼
- [x] NewRuleForm() 생성자
- [x] Reset() 메서드
- [x] Content() 메서드

---

## Phase 3: 템플릿 탭 통합

### Step 3.1: `internal/ui/rule_builder.go` 생성

- [x] RuleBuilder 구조체 정의
  - [x] ruleList *component.RuleList
  - [x] ruleForm *component.RuleForm
  - [x] onChange func()
- [x] NewRuleBuilder() 생성자
- [x] Content() 메서드
- [x] GetRules() []*FirewallRule 메서드
- [x] SetRules(rules []*FirewallRule) 메서드
- [x] Clear() 메서드

### Step 3.2: `internal/ui/template_tab.go` 수정

- [x] TemplateTab 구조체 필드 추가
  - [x] ruleBuilder *RuleBuilder
  - [x] subTabs *container.AppTabs
- [x] createTemplateContentPanel() 수정
  - [x] 텍스트 편집 탭 생성
  - [x] 규칙 빌더 탭 생성
  - [x] container.NewAppTabs() 사용
  - [x] OnSelected 핸들러 설정
- [x] onSubTabChanged() 핸들러 추가
  - [x] 텍스트 -> 빌더: ParseTextToRules() 호출
  - [x] 빌더 -> 텍스트: RulesToText() 호출
- [x] onSaveTemplate() 수정
  - [x] 현재 활성 탭 확인
  - [x] 규칙 빌더 탭이면 텍스트로 변환 후 저장
- [x] onTemplateSelected() 수정
  - [x] 두 뷰 모두 동기화

---

## Phase 4: 테스트 및 검증

### 파서 테스트

- [x] 빈 줄 처리 테스트
- [x] 주석 라인 처리 테스트
- [x] 기본 규칙 파싱 테스트
  - [x] `agent -m=insert -c=INPUT -p=tcp --dport=9010 -a=DROP`
- [x] 확장 필드 파싱 테스트
  - [x] --sip, --dip 포함
  - [x] --black, --white 플래그 포함
- [x] 규칙 -> 텍스트 변환 테스트
- [x] 왕복 변환 테스트 (파싱 후 다시 텍스트로)

### UI 테스트

- [x] 빌드 성공 확인
- [x] 앱 실행 확인
- [x] 템플릿 탭 표시 확인
- [x] 서브 탭 전환 확인
- [x] 규칙 추가 동작 확인
  - [x] 폼 입력
  - [x] 추가 버튼 클릭
  - [x] 테이블에 행 추가됨
- [x] 규칙 삭제 동작 확인
  - [x] 삭제 버튼 클릭
  - [x] 테이블에서 행 제거됨
- [x] 규칙 수정 동작 확인
  - [x] 테이블 내 위젯으로 직접 수정
- [x] 탭 전환 시 데이터 동기화 확인
  - [x] 텍스트 편집 -> 규칙 빌더
  - [x] 규칙 빌더 -> 텍스트 편집
- [x] 저장 동작 확인
  - [x] 규칙 빌더 탭에서 저장
  - [x] JSON 파일 확인
- [x] 로드 동작 확인
  - [x] 템플릿 선택 시 두 뷰 모두 업데이트

---

## 완료 체크

- [x] 모든 Phase 완료
- [x] 빌드 오류 없음
- [x] 기본 기능 동작 확인
- [x] 불필요 파일 정리 완료 (rule_row.go, rule_list.go 삭제)
- [ ] fms_wails 적용 준비

---

## 메모

### 2026-01-05
- fms_fyne 규칙 빌더 구현 완료
- 빌드 성공 확인
- UI 테스트 필요 (사용자 확인 필요)

### 2026-01-07
- **widget.Table 기반 RuleTable 구현 완료**
  - 기존 VBox+HBox+GridWrap 조합에서 widget.Table로 전환
  - 창 크기에 따라 자동으로 컬럼 너비 비율 조절
  - widget.BaseWidget 상속하여 커스텀 위젯으로 구현
- **Select hover 투명 문제 해결**
  - 각 셀에 canvas.Rectangle 불투명 배경 추가
  - Stack의 첫 번째 요소로 배경 배치
- **옵션 컬럼 클릭 시 팝업 기능 추가**
  - Label에서 Hyperlink로 변경
  - 클릭 시 오른쪽에 전체 옵션 내용 팝업 표시
  - widget.PopUp 사용
- rule_builder.go에서 RuleList → RuleTable 교체 완료
- 빌드 성공 확인

