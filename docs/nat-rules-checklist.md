# Implementation Plan: NAT 규칙 확장

**Status**: ✅ Complete
**Started**: 2026-01-06
**Last Updated**: 2026-01-08
**Related PRD**: [nat-rules-prd.md](./nat-rules-prd.md)

---

**⚠️ CRITICAL INSTRUCTIONS**: After completing each phase:
1. ✅ Check off completed task checkboxes
2. 🧪 Run all quality gate validation commands
3. ⚠️ Verify ALL quality gate items pass
4. 📅 Update "Last Updated" date above
5. 📝 Document learnings in Notes section
6. ➡️ Only then proceed to next phase

⛔ **DO NOT skip quality gates or proceed with failing checks**

---

## 📋 Overview

### Feature Description
템플릿 편집기에 NAT 규칙(DNAT, SNAT, MASQUERADE) 지원을 추가하여 포트 포워딩 및 소스 NAT 기능을 제공합니다.

### Success Criteria
- [ ] DNAT (Destination NAT) 규칙 지원 - 포트 포워딩
- [ ] SNAT (Source NAT) 규칙 지원 - 소스 주소 변환
- [ ] 인터페이스 기반 규칙 지원 (in_interface, out_interface)
- [ ] 기존 규칙 빌더와 통합
- [ ] smartfw 커널 모듈 형식과 호환

### User Impact
- 포트 포워딩 설정을 직관적인 UI로 관리 가능
- SNAT/MASQUERADE를 통한 내부 네트워크 외부 통신 설정

---

## 🏗️ Architecture Decisions

| Decision | Rationale | Trade-offs |
|----------|-----------|------------|
| 필터 규칙과 NAT 규칙 분리 | 체인(PREROUTING/POSTROUTING)이 다름 | UI 복잡도 증가 |
| 별도 서브 탭 추가 | 명확한 구분, 혼동 방지 | 탭 전환 필요 |
| smartfw 형식 호환 | 기존 커널 모듈과 호환 | 형식 제약 |
| NATRule 별도 구조체 | FirewallRule과 필드가 많이 다름 | 코드 중복 가능성 |

---

## 📦 Dependencies

### Required Before Starting
- [x] 기존 규칙 빌더 구현 완료
- [ ] 백엔드 Agent 서버 NAT 지원 여부 확인 (추후)

### External Dependencies
- `fyne.io/fyne/v2` - GUI 프레임워크

---

## 🧪 Test Strategy

### Testing Approach
**TDD Principle**: Write tests FIRST, then implement to make them pass

### Test Pyramid for This Feature
| Test Type | Coverage Target | Purpose |
|-----------|-----------------|---------|
| **Unit Tests** | ≥80% | 파서 함수, 데이터 모델, 헬퍼 함수 |
| **Integration Tests** | Critical paths | UI ↔ Parser 연동, smartfw 변환 |
| **Manual Tests** | Key user flows | DNAT/SNAT 규칙 추가/삭제 워크플로우 |

### Test File Organization
```
fms_fyne/
├── test/
│   ├── model/
│   │   └── nat_rule_test.go      # NATRule 테스트
│   │   └── nat_type_test.go      # NATType 열거형 테스트
│   └── parser/
│       └── nat_parser_test.go    # NAT 파싱 함수 테스트
```

---

## 🚀 Implementation Phases

### Phase 1: 데이터 모델
**Goal**: NATType 열거형과 NATRule 구조체 정의
**Status**: ✅ Complete

#### Tasks

**🔴 RED: Write Failing Tests First**
- [x] **Test 1.1**: NATType 열거형 테스트
  - File: `test/model/nat_type_test.go`
  - Test cases:
    - NATTypeToString() 변환
    - StringToNATType() 변환
    - GetNATTypeOptions() 반환 값

- [x] **Test 1.2**: NATRule 구조체 테스트
  - File: `test/model/nat_rule_test.go`
  - Test cases:
    - NewNATRule() 생성
    - 필드 초기값 확인

**🟢 GREEN: Implement to Make Tests Pass**
- [x] **Task 1.3**: `internal/model/nat_rule.go` 생성
  - [x] NATType 열거형 정의
    - NATTypeDNAT (0)
    - NATTypeSNAT (1)
    - NATTypeMASQUERADE (2)
  - [x] NATRule 구조체 정의
    - NATType NATType
    - Protocol Protocol
    - MatchIP string
    - MatchPort string
    - TranslateIP string
    - TranslatePort string
    - InInterface string
    - OutInterface string
    - Description string
  - [x] NewNATRule() 생성자
  - [x] 문자열 변환 헬퍼
    - NATTypeToString()
    - StringToNATType()
    - GetNATTypeOptions() - UI Select용

**🔵 REFACTOR: Clean Up Code**
- [x] **Task 1.4**: 코드 품질 개선
  - [x] 상수 정리
  - [x] 문서화 주석 추가

#### Quality Gate ✋

**⚠️ STOP: Do NOT proceed to Phase 2 until ALL checks pass**

**Build & Tests**:
- [x] `go build ./...` 성공
- [x] 수동 테스트 완료
- [x] 기능 동작 확인

**Code Quality**:
- [x] `go vet ./...` 오류 없음
- [x] `go fmt ./...` 적용됨

**Validation Commands**:
```bash
cd fms_fyne
go build ./...
go test ./test/model/... -v
go test ./test/model/... -cover
go vet ./...
```

---

### Phase 2: 파서
**Goal**: NAT 규칙 파싱/변환 함수 구현
**Status**: ✅ Complete

#### Tasks

**🔴 RED: Write Failing Tests First**
- [x] **Test 2.1**: ParseNATLine() 테스트
  - File: `test/parser/nat_parser_test.go`
  - Test cases:
    - DNAT 규칙 파싱
    - SNAT 규칙 파싱
    - MASQUERADE 규칙 파싱
    - 잘못된 형식 에러 처리

- [x] **Test 2.2**: NATRuleToLine() 테스트
  - File: `test/parser/nat_parser_test.go`
  - Test cases:
    - DNAT → agent 명령어
    - SNAT → agent 명령어

- [x] **Test 2.3**: NATRuleToSmartfw() 테스트
  - File: `test/parser/nat_parser_test.go`
  - Test cases:
    - DNAT → smartfw 형식
    - SNAT → smartfw 형식

- [x] **Test 2.4**: 왕복 변환 테스트
  - File: `test/parser/nat_parser_test.go`
  - 파싱 → 포맷 → 파싱 일관성 확인

**🟢 GREEN: Implement to Make Tests Pass**
- [x] **Task 2.5**: `internal/parser/nat_parser.go` 생성
  - [x] ParseNATLine() 함수 구현
    - NAT 규칙 라인 파싱
    - DNAT 형식 파싱
    - SNAT 형식 파싱
    - MASQUERADE 형식 파싱
  - [x] NATRuleToLine() 함수 구현
    - agent 명령어 형식으로 변환
  - [x] NATRuleToSmartfw() 함수 구현
    - DNAT smartfw 형식 변환
    - SNAT smartfw 형식 변환
  - [x] ParseTextToNATRules() 함수 구현
    - 전체 텍스트에서 NAT 규칙 추출
  - [x] NATRulesToText() 함수 구현
    - NAT 규칙 목록을 텍스트로 변환

**🔵 REFACTOR: Clean Up Code**
- [x] **Task 2.6**: 코드 품질 개선
  - [x] 에러 처리 개선
  - [x] 파싱 로직 최적화

#### Quality Gate ✋

**Build & Tests**:
- [x] `go build ./...` 성공
- [x] 수동 테스트 완료
- [x] 기능 동작 확인

**Validation Commands**:
```bash
cd fms_fyne
go build ./...
go test ./test/parser/... -v
go test ./test/parser/... -cover
```

---

### Phase 3: UI 컴포넌트
**Goal**: NAT 규칙 테이블, 폼, 빌더 UI 구현 (필터 규칙 빌더 패턴과 동일)
**Status**: ✅ Complete

> **Note**: 필터 규칙 빌더(RuleTable, RuleForm)와 동일한 패턴 적용
> - 테이블: `widget.Table` 기반 (고정 너비 + 비율 컬럼)
> - 폼: 탭 구조로 NAT 타입별 분리 (DNAT / SNAT·MASQ)

#### Tasks

**🟢 GREEN: Implement UI Components**
- [x] **Task 3.1**: `internal/ui/component/nat_table.go` 생성
  - [x] NATTable 구조체 정의 (RuleTable 패턴 참조)
    - widget.BaseWidget 상속
    - rules []*NATRule
    - table *widget.Table
    - onChange func()
    - lastWidth float32
  - [x] 컬럼 인덱스 상수 정의
    - colDelete, colType, colProto, colMatch, colTranslate, colInterface, colDesc
  - [x] 고정 너비 컬럼 상수
    - fixedWidthDelete = 36
    - fixedWidthType = 80
  - [x] 가변 컬럼 비율 정의
  - [x] createTable() - widget.NewTable 사용
    - ShowHeaderRow = true
    - ShowHeaderColumn = false (행 번호 제거)
  - [x] updateCell() - 셀 업데이트
  - [x] updateColumnWidths() - 고정 + 비율 기반
  - [x] Resize() - 크기 변경 시 컬럼 너비 재계산
  - [x] CRUD 메서드: AddRule(), RemoveRule(), GetRules(), SetRules(), Clear()

- [x] **Task 3.2**: `internal/ui/component/dnat_form.go` 생성
  - [x] DNATForm 구조체 정의
    - protocol Select
    - matchPort Entry (외부 포트)
    - matchIP Entry (소스 IP, 선택)
    - translateIP Entry (내부 IP)
    - translatePort Entry (내부 포트)
    - onAdd func(*NATRule)
  - [x] DNAT 전용 폼 레이아웃
  - [x] NewDNATForm() 생성자
  - [x] submitRule() - NATType=DNAT 고정
  - [x] Reset(), Content() 메서드
  - [x] 도움말 버튼 추가 (showDNATHelp)

- [x] **Task 3.3**: `internal/ui/component/snat_form.go` 생성
  - [x] SNATForm 구조체 정의
    - natTypeSel Select (SNAT / MASQUERADE)
    - protocol Select
    - matchIP Entry (소스 네트워크)
    - inInterface Entry
    - outInterface Entry
    - translateIP Entry (선택, SNAT만)
    - onAdd func(*NATRule)
  - [x] SNAT/MASQ 폼 레이아웃
  - [x] NewSNATForm() 생성자
  - [x] submitRule()
  - [x] Reset(), Content() 메서드
  - [x] 도움말 버튼 추가 (showSNATHelp)
  - [x] NAT 타입 변경 시 TransIP 행 표시/숨김 (onNATTypeChanged)

- [x] **Task 3.4**: `internal/ui/nat_builder.go` 생성
  - [x] NATBuilder 구조체 정의 (RuleBuilder 패턴 참조)
    - natTable *component.NATTable
    - dnatForm *component.DNATForm
    - snatForm *component.SNATForm
    - formTabs *container.AppTabs
    - onChange func()
  - [x] createUI() - 테이블 위, 폼 탭 아래 레이아웃
  - [x] NewNATBuilder() 생성자
  - [x] GetRules(), SetRules(), Clear(), Refresh() 메서드

- [x] **Task 3.5**: `internal/ui/template_tab.go` 수정
  - [x] TemplateTab 구조체 필드 추가
    - natBuilder *NATBuilder
  - [x] 서브 탭에 "NAT 규칙" 탭 추가
  - [x] NAT 탭 전환 핸들러 추가
  - [x] onSaveTemplate() 수정
    - NAT 규칙도 함께 저장
  - [x] onTemplateSelected() 수정
    - NAT 규칙도 로드

- [x] **Task 3.6**: `internal/ui/component/help_texts.go` 도움말 텍스트 추가
  - [x] DNATHelpText 상수 정의
  - [x] SNATHelpText 상수 정의
  - [x] ShowHelpPopup() 공통 함수 활용

**🔵 REFACTOR: Clean Up Code**
- [x] **Task 3.7**: UI 코드 품질 개선
  - [x] 중복 UI 로직 추출
  - [x] 이벤트 핸들러 정리

#### Quality Gate ✋

**Build & Tests**:
- [x] `go build ./...` 성공
- [x] 앱 실행 확인

**Manual Testing**:
- [x] DNAT 폼 레이아웃 확인
- [x] SNAT 폼 레이아웃 확인
- [x] NAT 타입 탭 전환 동작
- [x] 규칙 추가 → 테이블에 표시
- [x] 규칙 삭제 기능
- [x] 컬럼 너비 자동 조정 확인

**Validation Commands**:
```bash
cd fms_fyne
go build -ldflags "-H windowsgui -s -w" -o fms_fyne.exe .
./fms_fyne.exe
```

---

### Phase 4: 테스트 및 검증
**Goal**: 전체 기능 통합 테스트 및 smartfw 변환 검증
**Status**: ✅ Complete

#### Tasks

**DNAT 테스트**
- [x] **Task 4.1**: DNAT 규칙 추가 테스트
  - [x] 외부 포트 입력
  - [x] 내부 IP/포트 입력
  - [x] 추가 버튼 클릭
  - [x] 테이블에 표시 확인
- [x] **Task 4.2**: DNAT 규칙 삭제 테스트
- [x] **Task 4.3**: DNAT → smartfw 변환 확인
  - 예상 출력: `req|INSERT|{ID}|ANY|NAT|ANY|TCP?DNAT|192.168.30.180|6080,8080||`

**SNAT 테스트**
- [x] **Task 4.4**: SNAT 규칙 추가 테스트
  - [x] 소스 네트워크 입력
  - [x] 인터페이스 입력
  - [x] 추가 버튼 클릭
  - [x] 테이블에 표시 확인
- [x] **Task 4.5**: SNAT 규칙 삭제 테스트
- [x] **Task 4.6**: SNAT → smartfw 변환 확인

**통합 테스트**
- [x] **Task 4.7**: 필터 규칙 + NAT 규칙 함께 저장
- [x] **Task 4.8**: 템플릿 로드 시 NAT 규칙 표시
- [x] **Task 4.9**: 탭 전환 시 데이터 동기화

#### Quality Gate ✋

**Final Checklist**:
- [x] 모든 Phase 완료
- [x] 빌드 오류 없음
- [x] DNAT 기능 동작 확인
- [x] SNAT 기능 동작 확인
- [x] 기존 필터 규칙과 호환성 확인

---

## ⚠️ Risk Assessment

| Risk | Probability | Impact | Mitigation Strategy |
|------|-------------|--------|---------------------|
| 백엔드 Agent 서버 NAT 미지원 | Medium | High | 사전 백엔드 팀 협의, smartfw 형식 문서화 |
| smartfw 형식 불일치 | Medium | High | PRD의 예시 형식 정확히 따름, 테스트 철저히 |
| UI 복잡도 증가 | Low | Medium | DNAT/SNAT 폼 분리, 직관적 레이블 |
| 기존 템플릿 호환성 문제 | Low | Medium | NAT 규칙 별도 저장, 기존 필터 규칙 영향 없음 |

---

## 🔄 Rollback Strategy

### If Phase 1 Fails
- `internal/model/nat_rule.go` 파일 삭제
- 기존 model 코드 영향 없음

### If Phase 2 Fails
- `internal/parser/nat_parser.go` 파일 삭제
- 기존 parser 코드 영향 없음

### If Phase 3 Fails
- UI 컴포넌트 파일들 삭제
  - `nat_form.go`, `nat_row.go`, `nat_list.go`, `nat_builder.go`
- `template_tab.go` 변경 사항 롤백

---

## 📊 Progress Tracking

### Completion Status
- **Phase 1**: ✅ 100% (데이터 모델)
- **Phase 2**: ✅ 100% (파서)
- **Phase 3**: ✅ 100% (UI 컴포넌트)
- **Phase 4**: ✅ 100% (테스트)

**Overall Progress**: 100% complete

---

## 📝 Notes & Learnings

### 2026-01-06
- PRD 문서 작성 완료
- 체크리스트를 feature-planner 형식으로 업데이트
- 백엔드 Agent 서버의 NAT 지원 여부 확인 필요

### 2026-01-08
- **DNAT 폼 구현 완료** (dnat_form.go)
  - 프로토콜 선택, 외부 포트, 소스 IP, 내부 IP/포트 필드
  - 규칙 빌더와 동일한 레이아웃 패턴 적용
  - 도움말 버튼 추가
- **SNAT 폼 구현 완료** (snat_form.go)
  - NAT 타입 선택 (SNAT/MASQUERADE)
  - MASQUERADE 선택 시 TransIP 행 자동 숨김
  - 도움말 버튼 추가
- **도움말 텍스트 중앙 관리**
  - help_texts.go에 DNATHelpText, SNATHelpText 상수 추가
  - ShowHelpPopup() 공통 함수로 통일된 UI 제공
  - 모달 팝업 + 스크롤 가능한 컨텐츠 영역

---

## 📚 References

- [nat-rules-prd.md](./nat-rules-prd.md) - 상세 기능 요구사항
- [iptables NAT 설정 가이드](https://masterdaweb.com/en/blog/examples-of-snat-dnat-with-iptables)
- [pfSense Port Forwarding](https://docs.netgate.com/pfsense/en/latest/nat/port-forwards.html)
