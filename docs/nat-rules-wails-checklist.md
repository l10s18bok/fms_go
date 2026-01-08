# Implementation Plan: NAT 규칙 확장 (fms_wails)

**Status**: ✅ Complete
**Started**: 2026-01-08
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
- [x] DNAT (Destination NAT) 규칙 지원 - 포트 포워딩
- [x] SNAT (Source NAT) 규칙 지원 - 소스 주소 변환
- [x] MASQUERADE 규칙 지원
- [x] 인터페이스 기반 규칙 지원 (in_interface, out_interface)
- [x] 기존 규칙 빌더와 통합
- [x] smartfw 커널 모듈 형식과 호환

### User Impact
- 포트 포워딩 설정을 직관적인 UI로 관리 가능
- SNAT/MASQUERADE를 통한 내부 네트워크 외부 통신 설정

### Dependencies
- [x] **rule-builder-wails 구현 완료** (선행 필수)
- [x] **protocol-options-wails 구현 완료** (선행 필수)

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
- [x] rule-builder-wails 구현 완료
- [x] protocol-options-wails 구현 완료
- [x] fms_fyne nat-rules 구현 완료 (참조용)

### External Dependencies
- `github.com/wailsapp/wails/v2`
- React + TypeScript

---

## 🚀 Implementation Phases

### Phase 1: Go 백엔드 - 데이터 모델
**Goal**: NATType 열거형과 NATRule 구조체 정의
**Status**: ✅ Complete

#### Tasks

**🔴 RED: Write Failing Tests First**
- [x] **Test 1.1**: NATType 열거형 테스트
  - File: `internal/model/nat_rule_test.go`
  - Test cases:
    - NATTypeToString() 변환
    - StringToNATType() 변환
    - GetNATTypeOptions() 반환 값

- [x] **Test 1.2**: NATRule 구조체 테스트
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
    - GetNATTypeOptions()

**🔵 REFACTOR: Clean Up Code**
- [x] **Task 1.4**: 코드 품질 개선
  - [x] 상수 정리
  - [x] 문서화 주석 추가

#### Quality Gate ✋

**Build & Tests**:
- [x] `go build ./...` 성공
- [x] `go test ./internal/model/...` 100% 통과

---

### Phase 2: Go 백엔드 - 파서
**Goal**: NAT 규칙 파싱/변환 함수 구현
**Status**: ✅ Complete

#### Tasks

**🔴 RED: Write Failing Tests First**
- [x] **Test 2.1**: ParseNATLine() 테스트
  - File: `internal/parser/nat_parser_test.go`
  - Test cases:
    - DNAT 규칙 파싱
    - SNAT 규칙 파싱
    - MASQUERADE 규칙 파싱
    - 잘못된 형식 에러 처리

- [x] **Test 2.2**: NATRuleToLine() 테스트
- [x] **Test 2.3**: NATRuleToSmartfw() 테스트
- [x] **Test 2.4**: 왕복 변환 테스트

**🟢 GREEN: Implement to Make Tests Pass**
- [x] **Task 2.5**: `internal/parser/nat_parser.go` 생성
  - [x] ParseNATLine() 함수
  - [x] NATRuleToLine() 함수
  - [x] NATRuleToSmartfw() 함수
  - [x] ParseTextToNATRules() 함수
  - [x] NATRulesToText() 함수

**🔵 REFACTOR: Clean Up Code**
- [x] **Task 2.6**: 코드 품질 개선
  - [x] 에러 처리 개선
  - [x] 파싱 로직 최적화

#### Quality Gate ✋

**Build & Tests**:
- [x] `go build ./...` 성공
- [x] `go test ./internal/parser/...` 100% 통과

---

### Phase 3: Wails API 확장
**Goal**: NAT 규칙 관련 API 추가
**Status**: ✅ Complete

#### Tasks

- [x] **Task 3.1**: `app.go`에 NAT API 추가
  - [x] ParseNATRules(text string) - 텍스트를 NAT 규칙 배열로 변환
  - [x] NATRulesToText(rulesJSON string) - NAT 규칙을 텍스트로 변환
  - [x] GetNATTypeOptions() - NAT 타입 옵션 반환
  - [x] GetSNATTypeOptions() - SNAT 전용 옵션 반환
  - [x] NewNATRule(), NewDNATRule(), NewSNATRule() - 새 규칙 생성

- [x] **Task 3.2**: Wails 바인딩 재생성

#### Quality Gate ✋

- [x] `wails build` 성공

---

### Phase 4: React NAT UI 컴포넌트
**Goal**: NAT 규칙 테이블, DNAT 폼, SNAT 폼 구현
**Status**: ✅ Complete

#### Tasks

- [x] **Task 4.1**: `frontend/src/components/NATTable.tsx` 생성
  - [x] NATRule 인터페이스 정의
  - [x] 테이블 헤더 (삭제, 타입, 프로토콜, 매칭, 변환, 인터페이스, 설명)
  - [x] 테이블 행 렌더링
  - [x] 삭제 버튼 기능

- [x] **Task 4.2**: `frontend/src/components/DNATForm.tsx` 생성
  - [x] Protocol Select
  - [x] 외부 포트 Entry (matchPort)
  - [x] 소스 IP Entry (matchIP, 선택)
  - [x] 내부 IP Entry (translateIP)
  - [x] 내부 포트 Entry (translatePort)
  - [x] 추가 버튼
  - [x] 도움말 버튼

- [x] **Task 4.3**: `frontend/src/components/SNATForm.tsx` 생성
  - [x] NAT Type Select (SNAT / MASQUERADE)
  - [x] Protocol Select
  - [x] 소스 네트워크 Entry (matchIP)
  - [x] 입력 인터페이스 Entry (inInterface)
  - [x] 출력 인터페이스 Entry (outInterface)
  - [x] 변환 IP Entry (translateIP, SNAT만)
  - [x] 추가 버튼
  - [x] 도움말 버튼
  - [x] NAT 타입 변경 시 TransIP 행 표시/숨김

- [x] **Task 4.4**: 도움말 컴포넌트
  - [x] DNAT 도움말 팝업
  - [x] SNAT 도움말 팝업

#### Quality Gate ✋

**Manual Testing**:
- [x] DNAT 폼 레이아웃 확인
- [x] SNAT 폼 레이아웃 확인
- [x] NAT 타입 탭 전환 동작
- [x] 규칙 추가 → 테이블에 표시
- [x] 규칙 삭제 기능

---

### Phase 5: 템플릿 탭 NAT 통합
**Goal**: TemplateTab에 NAT 규칙 서브 탭 추가
**Status**: ✅ Complete

#### Tasks

- [x] **Task 5.1**: `TemplateTab.tsx` 수정
  - [x] "NAT 규칙" 서브 탭 추가
  - [x] NATTable, DNATForm, SNATForm 통합
  - [x] NAT 폼 탭 구조 (DNAT / SNAT)

- [x] **Task 5.2**: 데이터 동기화
  - [x] NAT 규칙 저장 로직
  - [x] NAT 규칙 로드 로직
  - [x] 텍스트 ↔ NAT 규칙 변환

- [x] **Task 5.3**: 저장 기능 수정
  - [x] 필터 규칙 + NAT 규칙 함께 저장

#### Quality Gate ✋

**Manual Testing**:
- [x] NAT 서브 탭 전환 동작
- [x] DNAT 규칙 추가/삭제
- [x] SNAT 규칙 추가/삭제
- [x] 저장 후 재로드 시 NAT 규칙 유지

---

### Phase 6: 통합 테스트
**Goal**: 전체 기능 통합 테스트 및 smartfw 변환 검증
**Status**: ✅ Complete

#### Tasks

**DNAT 테스트**
- [x] **Task 6.1**: DNAT 규칙 추가 테스트
- [x] **Task 6.2**: DNAT 규칙 삭제 테스트
- [x] **Task 6.3**: DNAT → smartfw 변환 확인
  - 예상: `req|INSERT|{ID}|ANY|NAT|ANY|TCP?DNAT|192.168.30.180|6080,8080||`

**SNAT 테스트**
- [x] **Task 6.4**: SNAT 규칙 추가 테스트
- [x] **Task 6.5**: SNAT 규칙 삭제 테스트
- [x] **Task 6.6**: SNAT → smartfw 변환 확인

**통합 테스트**
- [x] **Task 6.7**: 필터 규칙 + NAT 규칙 함께 저장
- [x] **Task 6.8**: 템플릿 로드 시 NAT 규칙 표시
- [x] **Task 6.9**: 탭 전환 시 데이터 동기화

#### Quality Gate ✋

**Final Checklist**:
- [x] 모든 Phase 완료
- [x] 빌드 오류 없음
- [x] DNAT 기능 동작 확인
- [x] SNAT 기능 동작 확인
- [x] 기존 필터 규칙과 호환성 확인

---

## 📊 Progress Tracking

### Completion Status
- **Phase 1**: ✅ 100% (데이터 모델)
- **Phase 2**: ✅ 100% (파서)
- **Phase 3**: ✅ 100% (API)
- **Phase 4**: ✅ 100% (UI 컴포넌트)
- **Phase 5**: ✅ 100% (통합)
- **Phase 6**: ✅ 100% (테스트)

**Overall Progress**: 100% complete

---

## 📝 Notes & Learnings

### 2026-01-08
- 체크리스트 문서 생성
- Phase 1-6 구현 완료
- 생성된 파일:
  - `internal/model/nat_rule.go` - NAT 규칙 모델
  - `internal/model/nat_rule_test.go` - 모델 테스트
  - `internal/parser/nat_parser.go` - NAT 파서
  - `internal/parser/nat_parser_test.go` - 파서 테스트
  - `app.go` - NAT API 바인딩 추가
  - `frontend/src/components/NATTable.tsx` - NAT 규칙 테이블
  - `frontend/src/components/DNATForm.tsx` - DNAT 폼
  - `frontend/src/components/SNATForm.tsx` - SNAT 폼
  - `frontend/src/components/TemplateTab.tsx` - NAT 서브 탭 통합
  - `frontend/src/App.css` - NAT 스타일 추가
- wails build 성공

---

## 📚 References

- [nat-rules-prd.md](./nat-rules-prd.md) - 상세 기능 요구사항
- [nat-rules-checklist.md](./nat-rules-checklist.md) - fms_fyne 구현 체크리스트
- fms_fyne/internal/model/nat_rule.go - 참조 코드
- fms_fyne/internal/parser/nat_parser.go - 참조 코드
