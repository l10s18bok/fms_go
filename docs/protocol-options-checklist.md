# Implementation Plan: 프로토콜 옵션 확장

**Status**: ✅ Completed
**Started**: 2026-01-06
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
- [x] ICMP Type/Code 옵션 지원
- [x] 기존 규칙과의 하위 호환성 유지
- [x] 쿼리 스트링 형식의 직관적인 명령어 표현

### User Impact
- 초보자: 프리셋에서 일반적인 보안 규칙 선택 가능
- 고급자: 체크박스로 직접 플래그 조합 설정 가능

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
- [x] 기존 규칙 빌더 구현 완료
- [ ] 백엔드 Agent 서버 호환성 확인 (추후)

### External Dependencies
- `fyne.io/fyne/v2` - GUI 프레임워크

---

## 🧪 Test Strategy

### Testing Approach
**TDD Principle**: Write tests FIRST, then implement to make them pass

### Test Pyramid for This Feature
| Test Type | Coverage Target | Purpose |
|-----------|-----------------|---------|
| **Unit Tests** | ≥80% | 파서 함수, 헬퍼 함수, 데이터 모델 |
| **Integration Tests** | Critical paths | UI ↔ Parser 연동, 탭 전환 동기화 |
| **Manual Tests** | Key user flows | 규칙 추가/수정/삭제 워크플로우 |

### Test File Organization
```
fms_fyne/
├── test/
│   ├── model/
│   │   └── protocol_options_test.go  # ProtocolOptions 테스트
│   │   └── tcp_flags_preset_test.go  # TCP Flags 프리셋 테스트
│   └── parser/
│       └── protocol_parser_test.go   # 프로토콜 파싱 함수 테스트
```

---

## 🚀 Implementation Phases

### Phase 1: 데이터 모델 확장
**Goal**: ProtocolOptions 구조체와 TCP Flags 프리셋 정의
**Status**: ✅ Completed

#### Tasks

**🔴 RED: Write Failing Tests First**
- [x] **Test 1.1**: ProtocolOptions 구조체 테스트
  - File: `test/model/protocol_options_test.go`
  - Test cases:
    - IsEmpty() 메서드
    - HasTCPOptions() 메서드
    - HasICMPOptions() 메서드

- [x] **Test 1.2**: TCP Flags 프리셋 테스트
  - File: `test/model/tcp_flags_preset_test.go`
  - Test cases:
    - GetTCPFlagsPresets() 반환 값
    - ToFlagsString() 변환
    - FindPresetByFlags() 검색

**🟢 GREEN: Implement to Make Tests Pass**
- [x] **Task 1.3**: `internal/model/rule.go` 수정
  - ProtocolOptions 구조체 추가
    - [x] TCPFlags string 필드
    - [x] ICMPType string 필드
    - [x] ICMPCode string 필드
    - [x] IsEmpty() 메서드
    - [x] HasTCPOptions() 메서드
    - [x] HasICMPOptions() 메서드
  - FirewallRule 구조체에 Options 필드 추가
  - NewFirewallRule() 수정 (Options 초기화)

- [x] **Task 1.4**: TCP Flags 프리셋 구현
  - [x] TCPFlagsPreset 구조체 정의
  - [x] GetTCPFlagsPresets() 함수
    - 없음 (모든 TCP 패킷)
    - 새 연결만 (SYN)
    - 확립된 연결 (ACK)
    - NULL 스캔 차단
    - XMAS 스캔 차단
    - SYN+FIN 차단
    - 커스텀
  - [x] ToFlagsString() 메서드
  - [x] FindPresetByFlags() 함수

- [x] **Task 1.5**: 헬퍼 함수 추가
  - [x] GetTCPFlagsList() - 체크박스용 플래그 목록
  - [x] GetICMPTypeOptions() - UI Select용 ICMP 타입 목록
  - [x] GetICMPCodeOptions() - UI Select용 ICMP Code 목록 (Type 3 전용)
  - [x] ICMPTypeNameToNumber() - 이름 → 숫자 변환
  - [x] ICMPTypeNumberToName() - 숫자 → 이름 변환
  - [x] ICMPCodeNameToNumber() - Code 이름 → 숫자 변환
  - [x] ICMPCodeNumberToName() - Code 숫자 → 이름 변환

**🔵 REFACTOR: Clean Up Code**
- [x] **Task 1.6**: 코드 품질 개선
  - [x] 중복 제거
  - [x] 명명 개선
  - [x] 인라인 문서화

#### Quality Gate ✋

**⚠️ STOP: Do NOT proceed to Phase 2 until ALL checks pass**

**Build & Tests**:
- [x] `go build ./...` 성공
- [x] `go test ./test/model/...` 100% 통과
- [x] 테스트 커버리지 ≥80%

**Code Quality**:
- [x] `go vet ./...` 오류 없음
- [x] `go fmt ./...` 적용됨

---

### Phase 2: 파서 확장
**Goal**: 쿼리 스트링 형식 파싱/포맷 함수 구현
**Status**: ✅ Completed

#### Tasks

**🔴 RED: Write Failing Tests First**
- [x] **Test 2.1**: ParseProtocolWithOptions() 테스트
  - File: `test/parser/protocol_parser_test.go`
  - Test cases:
    - 빈 옵션 파싱 (`tcp`)
    - TCP flags 파싱 (`tcp?flags=syn/syn`)
    - TCP flags 복수 파싱 (`tcp?flags=syn,ack/syn`)
    - ICMP type 이름 파싱 (`icmp?type=echo-request`)
    - ICMP type 숫자 파싱 (`icmp?type=8`)
    - ICMP type+code 파싱 (`icmp?type=3&code=0`)

- [x] **Test 2.2**: FormatProtocolWithOptions() 테스트
  - File: `test/parser/protocol_parser_test.go`
  - Test cases:
    - Protocol + nil Options
    - Protocol + TCPFlags
    - Protocol + ICMPType
    - Protocol + ICMPType + ICMPCode

- [x] **Test 2.3**: 왕복 변환 테스트
  - File: `test/parser/protocol_parser_test.go`
  - 파싱 → 포맷 → 파싱 일관성 확인

**🟢 GREEN: Implement to Make Tests Pass**
- [x] **Task 2.4**: `internal/parser/rule_parser.go` 수정
  - [x] ParseProtocolWithOptions() 함수 구현
    - "?" 기준 분리
    - 쿼리 스트링 파싱
    - flags 옵션 파싱
    - type 옵션 파싱
    - code 옵션 파싱
  - [x] FormatProtocolWithOptions() 함수 구현
    - Protocol + Options → 문자열 변환
    - TCP flags 포맷
    - ICMP type/code 포맷
  - [x] ParseLine() 수정
    - `-p=` 파싱 시 ParseProtocolWithOptions() 사용
  - [x] RuleToLine() 수정
    - Options가 있으면 FormatProtocolWithOptions() 사용

**🔵 REFACTOR: Clean Up Code**
- [x] **Task 2.5**: 코드 품질 개선
  - [x] 에러 처리 개선
  - [x] 파싱 로직 최적화

#### Quality Gate ✋

**Build & Tests**:
- [x] `go build ./...` 성공
- [x] `go test ./test/parser/...` 100% 통과
- [x] 테스트 커버리지 ≥80%

---

### Phase 3: UI 컴포넌트 수정
**Goal**: 규칙 폼과 테이블에 프로토콜 옵션 UI 추가
**Status**: ✅ Completed

#### Tasks

**🟢 GREEN: Implement UI Components**
- [x] **Task 3.1**: `internal/ui/component/rule_form.go` 수정
  - [x] TCP Flags 옵션 UI 추가
    - 프리셋 Select 위젯
    - 검사할 플래그 체크박스 그룹 (6개)
    - 설정된 플래그 체크박스 그룹 (6개)
    - 프리셋 선택 시 체크박스 자동 설정
    - 체크박스 변경 시 프리셋 → "커스텀" 전환
  - [x] ICMP 옵션 UI 개선
    - [x] Type Select 위젯
    - [x] Code Select 위젯 (드롭다운으로 변경)
    - [x] Type이 destination-unreachable (3)일 때만 Code 드롭다운 표시
    - [x] 다른 Type 선택 시 Code 숨김 및 초기화
    - [x] "커스텀 숫자" 선택 시 Entry 표시
  - [x] 프로토콜 Select OnChanged 수정
    - TCP 선택 시 TCP 옵션 표시
    - ICMP 선택 시 ICMP 옵션 표시
    - UDP/ANY 선택 시 옵션 숨김
  - [x] submitRule() 수정
    - Options 값 추출
    - FirewallRule에 Options 설정
  - [x] Reset() 수정
    - 옵션 필드 초기화

- [x] **Task 3.2**: `internal/ui/component/rule_row.go` 수정
  - [x] 옵션 컬럼 추가 (읽기 전용 Label)
  - [x] syncFromRule() 수정 (rule.Options 값 표시)
  - [x] 프로토콜 Select OnChanged 수정 (옵션 초기화 - syncing 플래그로 제어)
  - [x] triggerChange() 수정 (옵션 변경 시에도 호출)
  - [x] **옵션 컬럼을 읽기 전용 Label로 구현**
    - [x] 옵션 문자열만 표시 (예: `flags=syn,ack,fin,rst/syn`)
    - [x] `parser.FormatOptionsOnly()` 함수로 포맷팅
    - [x] 옵션 없을 시 "-" 표시
  - [x] `syncing` 플래그 추가 (syncFromRule 중 콜백 무시)
  - [x] updateOptionsLabel() 함수로 옵션 레이블 업데이트

- [x] **Task 3.3**: `internal/ui/component/rule_list.go` 수정
  - [x] 헤더에 "옵션" 컬럼 추가
  - [x] 컬럼 너비 조정

- [x] **Task 3.4**: 프로토콜별 필드 활성화/비활성화
  - [x] ICMP 선택 시 포트 필드 비활성화 ("N/A" placeholder)
  - [x] UDP/ANY 선택 시 TCP Flags 옵션 영역 비활성화 (회색 표시)
  - [x] setTCPOptionsEnabled() 헬퍼 함수
  - [x] setICMPOptionsEnabled() 헬퍼 함수

- [x] **Task 3.5**: 도움말 버튼 추가
  - [x] TCP Flags 영역에 "?" 버튼 추가
  - [x] ICMP Options 영역에 "?" 버튼 추가
  - [x] 모달 팝업으로 도움말 표시 (widget.NewModalPopUp)
  - [x] 스크롤 가능한 컨텐츠

- [x] **Task 3.6**: `internal/ui/component/help_texts.go` 생성 (신규)
  - [x] ShowHelpPopup() 공통 함수
  - [x] TCPFlagsHelpText 상수
  - [x] ICMPOptionsHelpText 상수
  - [x] AppHelpText 상수
  - [x] DNATHelpText 상수
  - [x] SNATHelpText 상수

**🔵 REFACTOR: Clean Up Code**
- [x] **Task 3.7**: UI 코드 품질 개선
  - [x] 중복 UI 로직 추출
  - [x] 이벤트 핸들러 정리

#### Quality Gate ✋

**Build & Tests**:
- [x] `go build ./...` 성공
- [x] 앱 실행 확인

**Manual Testing**:
- [x] TCP 프리셋 선택 → 체크박스 자동 설정
- [x] 체크박스 수정 → "커스텀" 전환
- [x] TCP flags 규칙 추가 → 테이블에 표시
- [x] ICMP type 규칙 추가 → 테이블에 표시
- [x] 프로토콜 변경 시 옵션 초기화

---

### Phase 4: 통합 테스트 및 검증
**Goal**: 전체 기능 통합 테스트 및 하위 호환성 검증
**Status**: ✅ Completed

#### Tasks

- [x] **Task 4.1**: 탭 전환 동기화 테스트
  - [x] 규칙 빌더 → 텍스트 편집: 옵션 포함 변환
  - [x] 텍스트 편집 → 규칙 빌더: 옵션 파싱 및 표시

- [x] **Task 4.2**: 저장/로드 테스트
  - [x] 옵션 포함 규칙 저장
  - [x] 재로드 후 옵션 정상 표시

- [x] **Task 4.3**: 하위 호환성 테스트
  - [x] 기존 규칙 (`-p=tcp`) 정상 동작
  - [x] 기존 템플릿 파일 로드 정상

#### Quality Gate ✋

**Final Checklist**:
- [x] 모든 Phase 완료
- [x] 빌드 오류 없음
- [x] 기본 기능 동작 확인
- [x] 하위 호환성 확인

---

## ⚠️ Risk Assessment

| Risk | Probability | Impact | Mitigation Strategy |
|------|-------------|--------|---------------------|
| 백엔드 Agent 서버 미지원 | Medium | High | 사전 백엔드 팀 협의, 형식 문서화 |
| UI 복잡도 증가 | Low | Medium | 프리셋 기본값 제공, 직관적 레이블 |
| 기존 템플릿 호환성 문제 | Low | High | 하위 호환성 테스트 철저히 수행 |

---

## 🔄 Rollback Strategy

### If Phase 1 Fails
- `internal/model/rule.go`에서 Options 필드 제거
- 기존 FirewallRule 구조체 복원

### If Phase 2 Fails
- 파서에서 새 함수 제거
- 기존 ParseLine(), RuleToLine() 복원

### If Phase 3 Fails
- UI 컴포넌트에서 옵션 관련 위젯 제거
- 기존 폼/테이블 레이아웃 복원

---

## 📊 Progress Tracking

### Completion Status
- **Phase 1**: ✅ 100%
- **Phase 2**: ✅ 100%
- **Phase 3**: ✅ 100%
- **Phase 4**: ✅ 100%

**Overall Progress**: 100% complete

---

## 📝 Notes & Learnings

### 2026-01-06
- PRD 문서 작성 완료
- 체크리스트를 feature-planner 형식으로 업데이트
- Phase 1~4 구현 완료
- 테스트 파일 생성: `test/model/`, `test/parser/`
- 모든 유닛 테스트 통과
- exe 빌드 성공
- ICMP Code UI 개선 완료
  - Entry → Select 드롭다운으로 변경
  - Type이 destination-unreachable (3)일 때만 Code 드롭다운 표시
  - GetICMPCodeOptions(), ICMPCodeNameToNumber(), ICMPCodeNumberToName() 함수 추가
  - 관련 테스트 추가 및 통과
- 테이블 행(rule_row.go) 옵션 UI 최종 구현
  - 옵션 컬럼을 읽기 전용 Label로 단순화
  - `parser.FormatOptionsOnly()` 함수 추가 (프로토콜 제외한 옵션 문자열만 반환)
  - 표시 형식: `flags=syn,ack,fin,rst/syn` 또는 `type=echo-request`
  - 옵션 없을 시 "-" 표시
- **버그 수정**: syncFromRule() 시 옵션 초기화 문제
  - 문제: protoSel.SetSelected() 호출 시 OnChanged 콜백이 실행되어 Options를 nil로 초기화
  - 해결: `syncing` 플래그 추가하여 syncFromRule() 실행 중에는 옵션 초기화 건너뜀
  - defer로 syncing 플래그 복원 보장

### 2026-01-08
- **프로토콜별 필드 활성화/비활성화 구현**
  - ICMP 선택 시 포트 필드 비활성화 및 "N/A" placeholder 표시
  - UDP/ANY 선택 시 TCP Flags 옵션 영역 비활성화 (회색 표시)
  - `setTCPOptionsEnabled()`, `setICMPOptionsEnabled()` 헬퍼 함수 추가
- **도움말 버튼 추가**
  - TCP Flags, ICMP Options 영역에 "?" 버튼 추가
  - `widget.NewModalPopUp`으로 중앙 모달 팝업 구현
  - 스크롤 가능한 컨텐츠 영역
- **help_texts.go 파일 생성**
  - 모든 도움말 텍스트 중앙 관리
  - `ShowHelpPopup()` 공통 함수로 통일된 UI 제공
  - TCPFlagsHelpText, ICMPOptionsHelpText 상수 정의
  - DNAT, SNAT 도움말도 함께 관리

---

## 📚 References

- [protocol-options-prd.md](./protocol-options-prd.md) - 상세 기능 요구사항
- [TCP Flags Complete Guide](https://www.actualtests.com/blog/tcp-flags-explained-complete-guide-to-syn-ack-fin-rst-psh-urg-with-examples-and-tcp-header-format/)
- [IANA ICMP Parameters](https://www.iana.org/assignments/icmp-parameters/icmp-parameters.xhtml)
