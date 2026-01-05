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

- [ ] Chain 상수 정의
  - [ ] ChainINPUT (0)
  - [ ] ChainOUTPUT (1)
  - [ ] ChainFORWARD (2)
  - [ ] ChainPREROUTING (3)
  - [ ] ChainPOSTROUTING (4)
- [ ] Protocol 상수 정의
  - [ ] ProtocolTCP (6)
  - [ ] ProtocolUDP (17)
  - [ ] ProtocolICMP (1)
  - [ ] ProtocolANY (255)
- [ ] Action 상수 정의
  - [ ] ActionDROP (0)
  - [ ] ActionACCEPT (1)
  - [ ] ActionREJECT (2)
- [ ] FirewallRule 구조체 정의
- [ ] 문자열 변환 헬퍼 메서드
  - [ ] ChainToString()
  - [ ] StringToChain()
  - [ ] ProtocolToString()
  - [ ] StringToProtocol()
  - [ ] ActionToString()
  - [ ] StringToAction()
  - [ ] GetChainOptions() - UI Select용
  - [ ] GetProtocolOptions() - UI Select용
  - [ ] GetActionOptions() - UI Select용

### Step 1.2: `internal/parser/rule_parser.go` 생성

- [ ] ParseLine(line string) (*FirewallRule, error)
  - [ ] 빈 줄 처리
  - [ ] 주석 라인(#) 처리
  - [ ] agent 형식 파싱
  - [ ] -c= (chain) 파싱
  - [ ] -p= (protocol) 파싱
  - [ ] -a= (action) 파싱
  - [ ] --dport= 파싱
  - [ ] --sip= 파싱
  - [ ] --dip= 파싱
  - [ ] --black 플래그 파싱
  - [ ] --white 플래그 파싱
  - [ ] --geoip 플래그 파싱
- [ ] RuleToLine(rule *FirewallRule) string
  - [ ] agent 형식으로 변환
  - [ ] 필수 필드 출력
  - [ ] 선택 필드 조건부 출력
- [ ] ParseTextToRules(text string) ([]*FirewallRule, []error)
  - [ ] 줄 단위 분리
  - [ ] 각 줄 파싱
  - [ ] 오류 수집
- [ ] RulesToText(rules []*FirewallRule) string
  - [ ] 각 규칙 변환
  - [ ] 줄바꿈으로 연결

---

## Phase 2: UI 컴포넌트

### Step 2.1: `internal/ui/component/rule_row.go` 생성

- [ ] RuleRow 구조체 정의
  - [ ] rule *model.FirewallRule
  - [ ] onDelete func()
  - [ ] onChange func()
- [ ] UI 요소 생성
  - [ ] 삭제 버튼 (theme.DeleteIcon)
  - [ ] Chain Select 위젯
  - [ ] Protocol Select 위젯
  - [ ] Action Select 위젯
  - [ ] DPort Entry 위젯
  - [ ] SIP Entry 위젯
  - [ ] DIP Entry 위젯
  - [ ] Black Check 위젯
  - [ ] White Check 위젯
  - [ ] GeoIP Check 위젯
- [ ] 컨테이너 레이아웃 (HBox)
- [ ] NewRuleRow() 생성자
- [ ] GetRule() 메서드
- [ ] SetRule() 메서드
- [ ] Content() 메서드

### Step 2.2: `internal/ui/component/rule_list.go` 생성

- [ ] RuleList 구조체 정의
  - [ ] rows []*RuleRow
  - [ ] onChange func()
  - [ ] container *fyne.Container
- [ ] 헤더 행 생성
  - [ ] 컬럼 Label들
- [ ] 스크롤 가능한 VBox
- [ ] NewRuleList() 생성자
- [ ] AddRule(rule *FirewallRule) 메서드
- [ ] RemoveRule(index int) 메서드
- [ ] GetRules() []*FirewallRule 메서드
- [ ] SetRules(rules []*FirewallRule) 메서드
- [ ] Clear() 메서드
- [ ] Content() 메서드
- [ ] Refresh() 메서드

### Step 2.3: `internal/ui/component/rule_form.go` 생성

- [ ] RuleForm 구조체 정의
  - [ ] onAdd func(*FirewallRule)
- [ ] UI 요소 생성
  - [ ] Chain Select (기본값: INPUT)
  - [ ] Protocol Select (기본값: TCP)
  - [ ] Action Select (기본값: DROP)
  - [ ] DPort Entry
  - [ ] SIP Entry
  - [ ] DIP Entry
  - [ ] Black Check
  - [ ] White Check
  - [ ] GeoIP Check
  - [ ] 추가 버튼
- [ ] NewRuleForm() 생성자
- [ ] Reset() 메서드
- [ ] Content() 메서드

---

## Phase 3: 템플릿 탭 통합

### Step 3.1: `internal/ui/rule_builder.go` 생성

- [ ] RuleBuilder 구조체 정의
  - [ ] ruleList *component.RuleList
  - [ ] ruleForm *component.RuleForm
  - [ ] onChange func()
- [ ] NewRuleBuilder() 생성자
- [ ] Content() 메서드
- [ ] GetRules() []*FirewallRule 메서드
- [ ] SetRules(rules []*FirewallRule) 메서드
- [ ] Clear() 메서드

### Step 3.2: `internal/ui/template_tab.go` 수정

- [ ] TemplateTab 구조체 필드 추가
  - [ ] ruleBuilder *RuleBuilder
  - [ ] subTabs *container.AppTabs
- [ ] createTemplateContentPanel() 수정
  - [ ] 텍스트 편집 탭 생성
  - [ ] 규칙 빌더 탭 생성
  - [ ] container.NewAppTabs() 사용
  - [ ] OnSelected 핸들러 설정
- [ ] onSubTabChanged() 핸들러 추가
  - [ ] 텍스트 -> 빌더: ParseTextToRules() 호출
  - [ ] 빌더 -> 텍스트: RulesToText() 호출
- [ ] onSaveTemplate() 수정
  - [ ] 현재 활성 탭 확인
  - [ ] 규칙 빌더 탭이면 텍스트로 변환 후 저장
- [ ] onTemplateSelected() 수정
  - [ ] 두 뷰 모두 동기화

---

## Phase 4: 테스트 및 검증

### 파서 테스트

- [ ] 빈 줄 처리 테스트
- [ ] 주석 라인 처리 테스트
- [ ] 기본 규칙 파싱 테스트
  - [ ] `agent -m=insert -c=INPUT -p=tcp --dport=9010 -a=DROP`
- [ ] 확장 필드 파싱 테스트
  - [ ] --sip, --dip 포함
  - [ ] --black, --white, --geoip 플래그 포함
- [ ] 규칙 -> 텍스트 변환 테스트
- [ ] 왕복 변환 테스트 (파싱 후 다시 텍스트로)

### UI 테스트

- [ ] 빌드 성공 확인
- [ ] 앱 실행 확인
- [ ] 템플릿 탭 표시 확인
- [ ] 서브 탭 전환 확인
- [ ] 규칙 추가 동작 확인
  - [ ] 폼 입력
  - [ ] 추가 버튼 클릭
  - [ ] 테이블에 행 추가됨
- [ ] 규칙 삭제 동작 확인
  - [ ] 삭제 버튼 클릭
  - [ ] 테이블에서 행 제거됨
- [ ] 규칙 수정 동작 확인
  - [ ] 테이블 내 위젯으로 직접 수정
- [ ] 탭 전환 시 데이터 동기화 확인
  - [ ] 텍스트 편집 -> 규칙 빌더
  - [ ] 규칙 빌더 -> 텍스트 편집
- [ ] 저장 동작 확인
  - [ ] 규칙 빌더 탭에서 저장
  - [ ] JSON 파일 확인
- [ ] 로드 동작 확인
  - [ ] 템플릿 선택 시 두 뷰 모두 업데이트

---

## 완료 체크

- [ ] 모든 Phase 완료
- [ ] 빌드 오류 없음
- [ ] 기본 기능 동작 확인
- [ ] fms_wails 적용 준비

---

## 메모

(구현 중 발견한 이슈나 변경사항을 여기에 기록)

