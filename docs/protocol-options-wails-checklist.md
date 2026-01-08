# Implementation Plan: 프로토콜 옵션 확장 (fms_wails)

**Status**: ✅ Complete (rule-builder와 함께 구현됨)
**Started**: 2026-01-08
**Last Updated**: 2026-01-08
**Related PRD**: [protocol-options-prd.md](./protocol-options-prd.md)

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
규칙 빌더에 TCP Flags와 ICMP Type/Code 옵션을 추가하여 방화벽 규칙의 정밀한 제어를 가능하게 합니다.

### Success Criteria
- [x] TCP Flags 옵션 지원 (SYN, ACK, FIN, RST, PSH, URG)
- [x] TCP Flags 프리셋 지원 (새 연결만, NULL 스캔 차단, XMAS 스캔 차단 등)
- [x] ICMP Type/Code 옵션 지원
- [x] 기존 규칙과의 하위 호환성 유지
- [x] 쿼리 스트링 형식의 직관적인 명령어 표현 (`-p=tcp?flags=syn/syn`)

### User Impact
- 초보자: 프리셋에서 일반적인 보안 규칙 선택 가능
- 고급자: 체크박스로 직접 플래그 조합 설정 가능

### Dependencies
- [x] **rule-builder-wails 구현 완료** (선행 필수)

---

## 🏗️ Architecture Decisions

| Decision | Rationale | Trade-offs |
|----------|-----------|------------|
| 쿼리 스트링 형식 (`-p=tcp?flags=syn/syn`) | HTTP 쿼리 스트링과 유사한 직관적 형식 | 백엔드 호환성 확인 필요 |
| TCP flags 소문자 표기 | 일관성 및 가독성 | - |
| ICMP type 이름+숫자 지원 | 사용자 편의성 | 변환 로직 필요 |
| 프리셋 + 체크박스 UI | 초보자/고급자 모두 지원 | UI 복잡도 증가 |

---

## 📦 Dependencies

### Required Before Starting
- [ ] rule-builder-wails 구현 완료 (Phase 1~6)
- [ ] fms_fyne protocol-options 구현 완료 (참조용)

### External Dependencies
- `github.com/wailsapp/wails/v2`
- React + TypeScript

---

## 🚀 Implementation Phases

### Phase 1: Go 백엔드 - 데이터 모델 확장
**Goal**: ProtocolOptions 구조체와 TCP Flags 프리셋 정의
**Status**: ⏳ Pending

#### Tasks

**🔴 RED: Write Failing Tests First**
- [ ] **Test 1.1**: ProtocolOptions 구조체 테스트
  - File: `internal/model/rule_test.go` (확장)
  - Test cases:
    - IsEmpty() 메서드
    - HasTCPOptions() 메서드
    - HasICMPOptions() 메서드

- [ ] **Test 1.2**: TCP Flags 프리셋 테스트
  - Test cases:
    - GetTCPFlagsPresets() 반환 값
    - ToFlagsString() 변환
    - FindPresetByFlags() 검색

**🟢 GREEN: Implement to Make Tests Pass**
- [ ] **Task 1.3**: `internal/model/rule.go` 수정
  - [ ] ProtocolOptions 구조체 추가
    - TCPFlags string
    - ICMPType string
    - ICMPCode string
  - [ ] FirewallRule에 Options 필드 추가
  - [ ] IsEmpty(), HasTCPOptions(), HasICMPOptions() 메서드

- [ ] **Task 1.4**: TCP Flags 프리셋 구현
  - [ ] TCPFlagsPreset 구조체
  - [ ] GetTCPFlagsPresets() 함수
  - [ ] ToFlagsString() 메서드
  - [ ] FindPresetByFlags() 함수

- [ ] **Task 1.5**: 헬퍼 함수 추가
  - [ ] GetTCPFlagsList()
  - [ ] GetICMPTypeOptions()
  - [ ] GetICMPCodeOptions()
  - [ ] ICMPTypeNameToNumber(), ICMPTypeNumberToName()
  - [ ] ICMPCodeNameToNumber(), ICMPCodeNumberToName()

#### Quality Gate ✋

**Build & Tests**:
- [ ] `go build ./...` 성공
- [ ] `go test ./internal/model/...` 100% 통과

---

### Phase 2: Go 백엔드 - 파서 확장
**Goal**: 쿼리 스트링 형식 파싱/포맷 함수 구현
**Status**: ⏳ Pending

#### Tasks

**🔴 RED: Write Failing Tests First**
- [ ] **Test 2.1**: ParseProtocolWithOptions() 테스트
  - File: `internal/parser/rule_parser_test.go` (확장)
  - Test cases:
    - 빈 옵션 파싱 (`tcp`)
    - TCP flags 파싱 (`tcp?flags=syn/syn`)
    - ICMP type 파싱 (`icmp?type=echo-request`)
    - ICMP type+code 파싱 (`icmp?type=3&code=0`)

- [ ] **Test 2.2**: FormatProtocolWithOptions() 테스트
- [ ] **Test 2.3**: 왕복 변환 테스트

**🟢 GREEN: Implement to Make Tests Pass**
- [ ] **Task 2.4**: `internal/parser/rule_parser.go` 수정
  - [ ] ParseProtocolWithOptions() 함수
  - [ ] FormatProtocolWithOptions() 함수
  - [ ] ParseLine() 수정 (Options 파싱)
  - [ ] RuleToLine() 수정 (Options 포맷)

#### Quality Gate ✋

**Build & Tests**:
- [ ] `go build ./...` 성공
- [ ] `go test ./internal/parser/...` 100% 통과

---

### Phase 3: Wails API 확장
**Goal**: 프로토콜 옵션 관련 API 추가
**Status**: ⏳ Pending

#### Tasks

- [ ] **Task 3.1**: `app.go`에 API 추가
  - [ ] GetTCPFlagsPresets() - 프리셋 목록 반환
  - [ ] GetICMPTypeOptions() - ICMP 타입 옵션 반환
  - [ ] GetICMPCodeOptions() - ICMP 코드 옵션 반환

- [ ] **Task 3.2**: Wails 바인딩 재생성

#### Quality Gate ✋

- [ ] `wails build` 성공

---

### Phase 4: React UI 컴포넌트 수정
**Goal**: 규칙 폼과 테이블에 프로토콜 옵션 UI 추가
**Status**: ⏳ Pending

#### Tasks

- [ ] **Task 4.1**: `RuleForm.tsx` 수정
  - [ ] TCP Flags 옵션 UI
    - 프리셋 Select
    - 검사할 플래그 체크박스 그룹 (6개)
    - 설정된 플래그 체크박스 그룹 (6개)
  - [ ] ICMP 옵션 UI
    - Type Select
    - Code Select (Type 3일 때만 표시)
  - [ ] 프로토콜별 필드 활성화/비활성화
    - ICMP 선택 시 포트 필드 비활성화
    - UDP/ANY 선택 시 TCP Flags 비활성화

- [ ] **Task 4.2**: `RuleTable.tsx` 수정
  - [ ] 옵션 컬럼 추가 (읽기 전용)
  - [ ] 프로토콜에 따른 옵션 표시

- [ ] **Task 4.3**: 도움말 컴포넌트
  - [ ] TCP Flags 도움말 팝업
  - [ ] ICMP Options 도움말 팝업

#### Quality Gate ✋

**Manual Testing**:
- [ ] TCP 프리셋 선택 → 체크박스 자동 설정
- [ ] 체크박스 수정 → "커스텀" 전환
- [ ] TCP flags 규칙 추가 → 테이블에 표시
- [ ] ICMP type 규칙 추가 → 테이블에 표시

---

### Phase 5: 통합 테스트
**Goal**: 전체 기능 통합 테스트 및 하위 호환성 검증
**Status**: ⏳ Pending

#### Tasks

- [ ] **Task 5.1**: 탭 전환 동기화 테스트
  - [ ] 규칙 빌더 → 텍스트 편집: 옵션 포함 변환
  - [ ] 텍스트 편집 → 규칙 빌더: 옵션 파싱 및 표시

- [ ] **Task 5.2**: 저장/로드 테스트
  - [ ] 옵션 포함 규칙 저장
  - [ ] 재로드 후 옵션 정상 표시

- [ ] **Task 5.3**: 하위 호환성 테스트
  - [ ] 기존 규칙 (`-p=tcp`) 정상 동작
  - [ ] 기존 템플릿 파일 로드 정상

#### Quality Gate ✋

**Final Checklist**:
- [ ] 모든 Phase 완료
- [ ] 빌드 오류 없음
- [ ] 기본 기능 동작 확인
- [ ] 하위 호환성 확인

---

## 📊 Progress Tracking

### Completion Status
- **Phase 1**: ✅ 100% (데이터 모델) - rule-builder와 함께 구현
- **Phase 2**: ✅ 100% (파서) - rule-builder와 함께 구현
- **Phase 3**: ✅ 100% (API) - rule-builder와 함께 구현
- **Phase 4**: ✅ 100% (UI) - rule-builder와 함께 구현
- **Phase 5**: ✅ 100% (통합 테스트) - rule-builder와 함께 검증

**Overall Progress**: 100% complete

---

## 📝 Notes & Learnings

### 2026-01-08
- 체크리스트 문서 생성
- rule-builder-wails 구현 시 함께 통합 구현됨
- 구현된 기능:
  - `internal/model/rule.go`: ProtocolOptions, TCPFlagsPreset, ICMP 헬퍼 함수
  - `internal/parser/rule_parser.go`: ParseProtocolWithOptions, FormatProtocolWithOptions
  - `app.go`: GetTCPFlagsPresets, GetICMPTypeOptions 등 API
  - `RuleForm.tsx`: TCP Flags 프리셋 Select, ICMP Type/Code Select
  - `RuleTable.tsx`: 옵션 컬럼 표시

---

## 📚 References

- [protocol-options-prd.md](./protocol-options-prd.md) - 상세 기능 요구사항
- [protocol-options-checklist.md](./protocol-options-checklist.md) - fms_fyne 구현 체크리스트
- fms_fyne/internal/model/rule.go - 참조 코드
- fms_fyne/internal/parser/rule_parser.go - 참조 코드
