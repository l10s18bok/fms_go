# 규칙 빌더 UI 개선 PRD (Product Requirements Document)

## 문서 정보
- **버전**: 1.0
- **작성일**: 2026-01-05
- **적용 대상**: fms_fyne, fms_wails

---

## 1. 개요

### 1.1 배경
현재 FMS 템플릿 편집기는 텍스트 직접 입력 방식으로 방화벽 규칙을 관리합니다. 사용자가 명령어 형식을 정확히 알아야 하며, 오타나 형식 오류가 발생하기 쉽습니다.

### 1.2 목표
텍스트 입력 방식을 유지하면서, **폼 기반 규칙 빌더 UI**를 추가하여 사용자가 드롭다운과 입력 필드로 쉽게 규칙을 추가/편집할 수 있도록 개선합니다.

### 1.3 핵심 원칙
- 기존 텍스트 편집 기능 100% 유지
- 두 모드 간 자동 데이터 동기화
- 기존 JSON 저장 형태 변경 없음

---

## 2. 현재 상태

### 2.1 명령어 형식
```
agent -m=insert -c=INPUT -p=tcp --dport=9010 -a=DROP
```

### 2.2 현재 UI
- MultiLineEntry 위젯으로 텍스트 직접 입력
- 파일: `fms_fyne/internal/ui/template_tab.go`

### 2.3 데이터 저장 형식 (변경 없음)
```json
{
  "version": "v1.1.0",
  "contents": "agent -m=insert -c=INPUT -p=tcp --dport=9010 -a=DROP\n..."
}
```

---

## 3. 지원 필드

### 3.1 firewall_t 구조체 기반 필드 정의

| 필드 | 타입 | 가능한 값 | UI 컴포넌트 | 필수 |
|------|------|-----------|-------------|------|
| chain | int | INPUT(0), OUTPUT(1), FORWARD(2) | Select | O |
| protocol | int | TCP(6), UDP(17), ICMP(1), ANY(255) | Select | O |
| action | int | DROP(0), ACCEPT(1), REJECT(2) | Select | O |
| dport | int | 포트 번호 (0-65535) | Entry | X |
| sip | string | Source IP (콤마리스트 지원) | Entry | X |
| dip | string | Destination IP (콤마리스트 지원) | Entry | X |
| black | bool | 블랙리스트 규칙 여부 | Check | X |
| white | bool | 화이트리스트 규칙 여부 | Check | X |

### 3.2 명령어 매핑

| 필드 | 명령어 옵션 | 예시 |
|------|-------------|------|
| chain | -c= | -c=INPUT |
| protocol | -p= | -p=tcp |
| action | -a= | -a=DROP |
| dport | --dport= | --dport=9010 |
| sip | --sip= | --sip=192.168.1.0/24 |
| dip | --dip= | --dip=10.0.0.1 |
| black | --black | --black |
| white | --white | --white |

---

## 4. UI 설계

### 4.1 전체 레이아웃

```
┌──────────────────────────────────────────────────────────────────────┐
│ 템플릿 탭                                                            │
├────────────────┬─────────────────────────────────────────────────────┤
│                │  [텍스트 편집] [규칙 빌더]  <- 서브 탭              │
│  좌측 패널     ├─────────────────────────────────────────────────────┤
│  (템플릿 목록) │                                                     │
│                │  ┌───────────────────────────────────────────────┐  │
│  ○ v1.1.0      │  │ [X]│Chain │Proto│Action│DPort│ SIP │ B │ W │    │
│  ○ v1.0.0      │  ├────┼──────┼─────┼──────┼─────┼─────┼───┼───┤    │
│                │  │ X  │INPUT │ TCP │ DROP │9010 │ ANY │ □ │ □ │    │
│                │  │ X  │INPUT │ TCP │ DROP │9020 │ ANY │ □ │ □ │    │
│                │  └───────────────────────────────────────────────┘  │
│                │                                                     │
│                │  ┌─ 규칙 추가 ────────────────────────────────────┐ │
│                │  │ Chain:[INPUT v] Proto:[TCP v] Action:[DROP v] │ │
│                │  │ DPort:[     ] SIP:[          ] DIP:[        ] │ │
│                │  │ [ ] Black  [ ] White           [+ 추가]       │ │
│                │  └───────────────────────────────────────────────┘ │
├────────────────┴─────────────────────────────────────────────────────┤
│ v1.0.0                                         [저장] [삭제]        │
└──────────────────────────────────────────────────────────────────────┘
```

### 4.2 서브 탭 구조

| 탭 | 설명 |
|-----|------|
| 텍스트 편집 | 기존 MultiLineEntry (변경 없음) |
| 규칙 빌더 | 테이블 + 폼 기반 UI (신규) |

### 4.3 규칙 테이블 컬럼

| 컬럼 | 너비 | 위젯 | 설명 |
|------|------|------|------|
| 삭제 | 40px | Button | 행 삭제 버튼 |
| Chain | 100px | Select | 체인 선택 |
| Proto | 80px | Select | 프로토콜 선택 |
| Action | 80px | Select | 액션 선택 |
| DPort | 60px | Entry | 목적지 포트 |
| SIP | 120px | Entry | 소스 IP |
| DIP | 120px | Entry | 목적지 IP |
| B | 30px | Check | 블랙리스트 |
| W | 30px | Check | 화이트리스트 |

### 4.4 규칙 추가 폼

- 모든 필드 입력 위젯 제공
- 기본값: Chain=INPUT, Protocol=TCP, Action=DROP
- [+ 추가] 버튼 클릭 시 테이블에 행 추가
- 추가 후 폼 초기화

---

## 5. 동작 흐름

### 5.1 탭 전환 동기화

```
[텍스트 편집] -> [규칙 빌더] 전환 시:
1. 텍스트 내용을 파싱
2. 파싱된 규칙을 테이블에 표시
3. 파싱 실패한 라인은 오류 표시

[규칙 빌더] -> [텍스트 편집] 전환 시:
1. 테이블의 규칙들을 텍스트로 변환
2. 변환된 텍스트를 에디터에 표시
```

### 5.2 저장 흐름

```
[저장] 버튼 클릭
    ↓
현재 활성 탭 확인
    ↓
(규칙 빌더 탭이면) 규칙 -> 텍스트 변환
    ↓
기존과 동일하게 JSON 저장
```

### 5.3 규칙 추가 흐름

```
폼에 값 입력
    ↓
[+ 추가] 클릭
    ↓
유효성 검사
    ↓
테이블에 새 행 추가
    ↓
폼 초기화
```

---

## 6. 파일 구조

### 6.1 신규 파일

| 경로 | 용도 |
|------|------|
| `internal/model/rule.go` | 규칙 데이터 모델, 상수 정의 |
| `internal/parser/rule_parser.go` | 텍스트 <-> 규칙 변환 |
| `internal/ui/component/rule_row.go` | 규칙 행 위젯 |
| `internal/ui/component/rule_list.go` | 규칙 목록 위젯 |
| `internal/ui/component/rule_form.go` | 규칙 추가 폼 |
| `internal/ui/rule_builder.go` | 규칙 빌더 패널 |

### 6.2 수정 파일

| 경로 | 수정 내용 |
|------|-----------|
| `internal/ui/template_tab.go` | 서브 탭 구조, 탭 전환 핸들러 |

---

## 7. 기술 사양

### 7.1 데이터 모델

```go
type Chain int
const (
    ChainINPUT      Chain = 0
    ChainOUTPUT     Chain = 1
    ChainFORWARD    Chain = 2
    ChainPREROUTING Chain = 3
    ChainPOSTROUTING Chain = 4
)

type Protocol int
const (
    ProtocolTCP  Protocol = 6
    ProtocolUDP  Protocol = 17
    ProtocolICMP Protocol = 1
    ProtocolANY  Protocol = 255
)

type Action int
const (
    ActionDROP   Action = 0
    ActionACCEPT Action = 1
    ActionREJECT Action = 2
)

type FirewallRule struct {
    Chain    Chain
    Protocol Protocol
    Action   Action
    DPort    string
    SIP      string
    DIP      string
    Black    bool
    White    bool
}
```

### 7.2 파서 인터페이스

```go
// 단일 라인 파싱
func ParseLine(line string) (*FirewallRule, error)

// 규칙을 텍스트로 변환
func RuleToLine(rule *FirewallRule) string

// 전체 텍스트 파싱
func ParseTextToRules(text string) ([]*FirewallRule, []error)

// 규칙 목록을 텍스트로 변환
func RulesToText(rules []*FirewallRule) string
```

---

## 8. 참조

- [pfSense Firewall Rules](https://docs.netgate.com/pfsense/en/latest/firewall/configure.html)
- [OPNsense Firewall Rules](https://docs.opnsense.org/manual/firewall.html)
- [Firewall Builder](https://fwbuilder.sourceforge.net/4.0/features.shtml)
- [Vercel WAF UX](https://vercel.com/blog/security-through-design-improved-firewall-experience)
