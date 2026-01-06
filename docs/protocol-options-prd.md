# 프로토콜 옵션 확장 PRD (Product Requirements Document)

## 문서 정보
- **버전**: 1.0
- **작성일**: 2026-01-06
- **적용 대상**: fms_fyne, fms_wails
- **관련 문서**: [rule-builder-prd.md](./rule-builder-prd.md)

---

## 1. 개요

### 1.1 배경
현재 규칙 빌더는 프로토콜(TCP, UDP, ICMP, ANY)만 선택 가능하며, 프로토콜별 세부 옵션을 지정할 수 없습니다. 방화벽 규칙의 정밀한 제어를 위해 TCP Flags, ICMP Type/Code 등의 옵션 지원이 필요합니다.

### 1.2 목표
- TCP Flags 옵션 지원 (SYN, ACK, FIN, RST, PSH, URG)
- ICMP Type/Code 옵션 지원
- 기존 규칙과의 하위 호환성 유지
- 쿼리 스트링 형식의 직관적인 명령어 표현

### 1.3 핵심 원칙
- 기존 `-p=tcp` 형식 100% 호환
- 옵션은 선택 사항 (필수 아님)
- UI에서 프로토콜 선택 시 동적으로 옵션 필드 표시

---

## 2. 명령어 형식

### 2.1 쿼리 스트링 형식 채택

프로토콜 옵션은 HTTP 쿼리 스트링과 유사한 형식을 사용합니다.

```
-p={protocol}?{option1}={value1}&{option2}={value2}
```

### 2.2 형식 예시

| 프로토콜 | 옵션 | 명령어 예시 |
|----------|------|-------------|
| TCP | 기본 (옵션 없음) | `-p=tcp` |
| TCP | flags | `-p=tcp?flags=syn/syn` |
| TCP | flags (복수) | `-p=tcp?flags=syn,ack/syn` |
| ICMP | 기본 (옵션 없음) | `-p=icmp` |
| ICMP | type (이름) | `-p=icmp?type=echo-request` |
| ICMP | type (숫자) | `-p=icmp?type=8` |
| ICMP | type + code | `-p=icmp?type=3&code=0` |
| UDP | 기본 | `-p=udp` |
| ANY | 기본 | `-p=any` |

### 2.3 전체 명령어 예시

```bash
# 기존 형식 (그대로 동작)
agent -m=insert -c=INPUT -p=tcp --dport=9010 -a=DROP

# TCP SYN 패킷만 차단
agent -m=insert -c=INPUT -p=tcp?flags=syn/syn --dport=80 -a=DROP

# ICMP ping 요청 차단
agent -m=insert -c=INPUT -p=icmp?type=echo-request -a=DROP

# ICMP 목적지 도달 불가 (port unreachable) 차단
agent -m=insert -c=INPUT -p=icmp?type=3&code=3 -a=DROP
```

---

## 3. TCP Flags 옵션

### 3.1 지원 플래그

| 플래그 | 설명 | 용도 |
|--------|------|------|
| syn | Synchronize | 연결 시작 요청 |
| ack | Acknowledge | 확인 응답 |
| fin | Finish | 연결 종료 요청 |
| rst | Reset | 연결 강제 종료 |
| psh | Push | 데이터 즉시 전달 |
| urg | Urgent | 긴급 데이터 표시 |

### 3.2 Flags 형식

```
flags={검사할플래그}/{설정된플래그}
```

- **검사할플래그**: 검사 대상 플래그 목록 (쉼표로 구분)
- **설정된플래그**: 실제로 설정되어야 할 플래그 목록 (쉼표로 구분)

### 3.3 Flags 예시

| 형식 | 의미 | iptables 동등 |
|------|------|---------------|
| `syn/syn` | SYN만 설정 | `--tcp-flags SYN SYN` |
| `syn,ack/syn` | SYN,ACK 검사, SYN만 설정 | `--tcp-flags SYN,ACK SYN` |
| `syn,rst,ack,fin/syn` | 새 연결 (--syn) | `--tcp-flags SYN,RST,ACK,FIN SYN` |
| `fin,syn,rst,psh,ack,urg/fin,psh,urg` | XMAS 스캔 | `--tcp-flags ALL FIN,PSH,URG` |

### 3.4 일반적인 보안 규칙

| 용도 | flags 값 | 설명 |
|------|----------|------|
| 새 연결만 허용 | `syn,rst,ack,fin/syn` | SYN flood 방지 |
| NULL 스캔 차단 | `fin,syn,rst,psh,ack,urg/` | 모든 플래그 해제된 패킷 |
| XMAS 스캔 차단 | `fin,syn,rst,psh,ack,urg/fin,psh,urg` | 비정상 플래그 조합 |
| SYN+FIN 차단 | `syn,fin/syn,fin` | 비정상 플래그 조합 |

---

## 4. ICMP Type/Code 옵션

### 4.1 지원 Type 값

숫자와 이름 모두 지원합니다.

| Type | 이름 | 설명 |
|------|------|------|
| 0 | echo-reply | 핑 응답 (pong) |
| 3 | destination-unreachable | 목적지 도달 불가 |
| 4 | source-quench | 소스 억제 (deprecated) |
| 5 | redirect | 라우팅 리다이렉트 |
| 8 | echo-request | 핑 요청 (ping) |
| 11 | time-exceeded | TTL 초과 |
| 12 | parameter-problem | 파라미터 문제 |
| 13 | timestamp-request | 타임스탬프 요청 |
| 14 | timestamp-reply | 타임스탬프 응답 |

### 4.2 Destination Unreachable (Type 3) Code 값

| Code | 이름 | 설명 |
|------|------|------|
| 0 | network-unreachable | 네트워크 도달 불가 |
| 1 | host-unreachable | 호스트 도달 불가 |
| 2 | protocol-unreachable | 프로토콜 도달 불가 |
| 3 | port-unreachable | 포트 도달 불가 |
| 4 | fragmentation-needed | 단편화 필요 |
| 5 | source-route-failed | 소스 라우트 실패 |

### 4.3 ICMP 예시

| 용도 | 명령어 | 설명 |
|------|--------|------|
| ping 차단 | `-p=icmp?type=echo-request` | 핑 요청 차단 |
| ping 차단 (숫자) | `-p=icmp?type=8` | 위와 동일 |
| pong 허용 | `-p=icmp?type=echo-reply` | 핑 응답 허용 |
| port unreachable | `-p=icmp?type=3&code=3` | 포트 도달 불가 |
| TTL 초과 | `-p=icmp?type=time-exceeded` | traceroute용 |

---

## 5. 데이터 모델

### 5.1 ProtocolOptions 구조체 (신규)

```go
// ProtocolOptions 프로토콜별 세부 옵션
type ProtocolOptions struct {
    // TCP 옵션
    TCPFlags string // 예: "syn/syn", "syn,ack/syn"

    // ICMP 옵션
    ICMPType string // 예: "echo-request", "8"
    ICMPCode string // 예: "0", "3" (선택)
}

// IsEmpty 옵션이 비어있는지 확인
func (o *ProtocolOptions) IsEmpty() bool

// HasTCPOptions TCP 옵션이 있는지 확인
func (o *ProtocolOptions) HasTCPOptions() bool

// HasICMPOptions ICMP 옵션이 있는지 확인
func (o *ProtocolOptions) HasICMPOptions() bool
```

### 5.2 FirewallRule 구조체 확장

```go
type FirewallRule struct {
    Chain    Chain
    Protocol Protocol
    Options  *ProtocolOptions // 신규 필드
    Action   Action
    DPort    string
    SIP      string
    DIP      string
    Black    bool
    White    bool
}
```

### 5.3 TCP Flags 프리셋 구조체

```go
// TCPFlagsPreset TCP Flags 프리셋 정의
type TCPFlagsPreset struct {
    Name        string   // 프리셋 이름 (UI 표시용)
    MaskFlags   []string // 검사할 플래그
    SetFlags    []string // 설정된 플래그
    Description string   // 설명
}

// GetTCPFlagsPresets 프리셋 목록 반환
func GetTCPFlagsPresets() []TCPFlagsPreset {
    return []TCPFlagsPreset{
        {
            Name:        "없음",
            MaskFlags:   nil,
            SetFlags:    nil,
            Description: "모든 TCP 패킷 매칭",
        },
        {
            Name:        "새 연결만 (SYN)",
            MaskFlags:   []string{"syn", "rst", "ack", "fin"},
            SetFlags:    []string{"syn"},
            Description: "새 연결 요청만 매칭",
        },
        {
            Name:        "확립된 연결 (ACK)",
            MaskFlags:   []string{"ack"},
            SetFlags:    []string{"ack"},
            Description: "기존 연결 패킷만 매칭",
        },
        {
            Name:        "NULL 스캔 차단",
            MaskFlags:   []string{"syn", "rst", "ack", "fin", "psh", "urg"},
            SetFlags:    nil,
            Description: "플래그 없는 비정상 패킷",
        },
        {
            Name:        "XMAS 스캔 차단",
            MaskFlags:   []string{"syn", "rst", "ack", "fin", "psh", "urg"},
            SetFlags:    []string{"fin", "psh", "urg"},
            Description: "비정상 플래그 조합",
        },
        {
            Name:        "SYN+FIN 차단",
            MaskFlags:   []string{"syn", "fin"},
            SetFlags:    []string{"syn", "fin"},
            Description: "비정상 플래그 조합",
        },
        {
            Name:        "커스텀",
            MaskFlags:   nil,
            SetFlags:    nil,
            Description: "직접 체크박스 설정",
        },
    }
}

// PresetToFlags 프리셋을 flags 문자열로 변환
// 예: "syn,rst,ack,fin/syn"
func (p *TCPFlagsPreset) ToFlagsString() string

// FindPresetByFlags flags 문자열에 매칭되는 프리셋 찾기
// 매칭되는 프리셋 없으면 "커스텀" 반환
func FindPresetByFlags(flags string) *TCPFlagsPreset
```

### 5.4 헬퍼 함수

```go
// TCP flags 옵션 목록 (체크박스용)
func GetTCPFlagsList() []string {
    return []string{"syn", "ack", "fin", "rst", "psh", "urg"}
}

// ICMP type 옵션 목록 (UI Select용)
func GetICMPTypeOptions() []string {
    return []string{
        "없음",                    // 옵션 없음
        "echo-request (8)",       // ping 요청
        "echo-reply (0)",         // ping 응답
        "destination-unreachable (3)",
        "time-exceeded (11)",
        "redirect (5)",
        "커스텀 숫자...",
    }
}

// ICMP Code 옵션 목록 (Type 3 - destination-unreachable 전용)
func GetICMPCodeOptions() []string {
    return []string{
        "없음",                      // 옵션 없음 (모든 Code)
        "network-unreachable (0)",
        "host-unreachable (1)",
        "protocol-unreachable (2)",
        "port-unreachable (3)",
        "fragmentation-needed (4)",
        "source-route-failed (5)",
    }
}

// ICMP type 이름을 숫자로 변환
func ICMPTypeNameToNumber(name string) (int, error)

// ICMP type 숫자를 이름으로 변환
func ICMPTypeNumberToName(num int) string

// ICMP code 이름을 숫자로 변환
func ICMPCodeNameToNumber(name string) (int, error)

// ICMP code 숫자를 이름으로 변환
func ICMPCodeNumberToName(num int) string
```

---

## 6. 파서 확장

### 6.1 프로토콜 파싱 함수

```go
// ParseProtocolWithOptions 프로토콜 문자열을 파싱
// 입력: "tcp?flags=syn/syn" 또는 "tcp"
// 출력: Protocol, *ProtocolOptions, error
func ParseProtocolWithOptions(s string) (Protocol, *ProtocolOptions, error)

// FormatProtocolWithOptions 프로토콜과 옵션을 문자열로 변환
// 입력: Protocol=TCP, Options={TCPFlags: "syn/syn"}
// 출력: "tcp?flags=syn/syn"
func FormatProtocolWithOptions(p Protocol, opts *ProtocolOptions) string
```

### 6.2 파싱 로직

```go
func ParseProtocolWithOptions(s string) (Protocol, *ProtocolOptions, error) {
    // 1. "?" 기준으로 분리
    parts := strings.SplitN(s, "?", 2)
    protocol := StringToProtocol(parts[0])

    if len(parts) == 1 {
        // 옵션 없음
        return protocol, nil, nil
    }

    // 2. 쿼리 스트링 파싱
    opts := &ProtocolOptions{}
    params := strings.Split(parts[1], "&")

    for _, param := range params {
        kv := strings.SplitN(param, "=", 2)
        if len(kv) != 2 {
            continue
        }

        switch kv[0] {
        case "flags":
            opts.TCPFlags = kv[1]
        case "type":
            opts.ICMPType = kv[1]
        case "code":
            opts.ICMPCode = kv[1]
        }
    }

    return protocol, opts, nil
}
```

---

## 7. UI 설계

### 7.1 규칙 추가 폼 - TCP 선택 시 (프리셋 + 체크박스)

TCP Flags는 **프리셋 드롭다운**과 **체크박스 그룹**을 조합하여 제공합니다.
- 초보자: 프리셋에서 일반적인 보안 규칙 선택
- 고급자: 체크박스로 직접 플래그 조합 설정

```
┌─ 규칙 추가 ─────────────────────────────────────────────────────┐
│ Chain: [INPUT    v]  Proto: [TCP v]  Action: [DROP   v]         │
│                                                                  │
│ ┌─ TCP Flags 옵션 ─────────────────────────────────────────────┐│
│ │ 프리셋: [새 연결만 (SYN)              v]                      ││
│ │         ├─ 없음 (모든 TCP 패킷)                               ││
│ │         ├─ 새 연결만 (SYN)                                    ││
│ │         ├─ 확립된 연결 (ACK)                                  ││
│ │         ├─ NULL 스캔 차단                                     ││
│ │         ├─ XMAS 스캔 차단                                     ││
│ │         ├─ SYN+FIN 차단                                       ││
│ │         └─ 커스텀...                                          ││
│ │                                                               ││
│ │ 검사할 플래그: [✓]SYN [✓]ACK [ ]FIN [✓]RST [ ]PSH [ ]URG     ││
│ │ 설정된 플래그: [✓]SYN [ ]ACK [ ]FIN [ ]RST [ ]PSH [ ]URG     ││
│ └──────────────────────────────────────────────────────────────┘│
│                                                                  │
│ DPort: [      ]  SIP: [              ]  DIP: [              ]   │
│ [ ] Black   [ ] White                              [+ 추가]     │
└─────────────────────────────────────────────────────────────────┘
```

#### TCP Flags 프리셋 목록

| 프리셋 | 검사할 플래그 | 설정된 플래그 | 용도 |
|--------|---------------|---------------|------|
| 없음 | - | - | 모든 TCP 패킷 매칭 |
| 새 연결만 (SYN) | syn,rst,ack,fin | syn | 새 연결 요청만 매칭 |
| 확립된 연결 (ACK) | ack | ack | 기존 연결 패킷만 매칭 |
| NULL 스캔 차단 | syn,rst,ack,fin,psh,urg | (없음) | 플래그 없는 비정상 패킷 |
| XMAS 스캔 차단 | syn,rst,ack,fin,psh,urg | fin,psh,urg | 비정상 플래그 조합 |
| SYN+FIN 차단 | syn,fin | syn,fin | 비정상 플래그 조합 |
| 커스텀 | 사용자 지정 | 사용자 지정 | 직접 체크박스 설정 |

#### 프리셋 선택 시 동작
1. 프리셋 선택 → 해당하는 체크박스 자동 설정
2. 체크박스 직접 수정 → 프리셋이 "커스텀"으로 변경
3. "없음" 선택 → 모든 체크박스 해제, flags 옵션 미적용

### 7.2 규칙 추가 폼 - ICMP 선택 시

**Type 선택에 따른 Code 드롭다운 조건부 표시:**
- `destination-unreachable (3)` 선택 시에만 Code 드롭다운 표시
- 다른 Type 선택 시 Code는 숨김

```
┌─ 규칙 추가 ─────────────────────────────────────────────────────┐
│ Chain: [INPUT    v]  Proto: [ICMP v]  Action: [DROP   v]        │
│                                                                  │
│ ┌─ ICMP 옵션 ──────────────────────────────────────────────────┐│
│ │ Type: [destination-unreachable (3) v]                         ││
│ │       ├─ 없음 (모든 ICMP)                                     ││
│ │       ├─ echo-request (8) - ping 요청                         ││
│ │       ├─ echo-reply (0) - ping 응답                           ││
│ │       ├─ destination-unreachable (3)                          ││
│ │       ├─ time-exceeded (11)                                   ││
│ │       ├─ redirect (5)                                         ││
│ │       └─ 커스텀 숫자...                                       ││
│ │                                                               ││
│ │ Code: [port-unreachable (3)    v]  ← Type 3일 때만 표시       ││
│ │       ├─ 없음 (모든 Code)                                     ││
│ │       ├─ network-unreachable (0)                              ││
│ │       ├─ host-unreachable (1)                                 ││
│ │       ├─ protocol-unreachable (2)                             ││
│ │       ├─ port-unreachable (3)                                 ││
│ │       ├─ fragmentation-needed (4)                             ││
│ │       └─ source-route-failed (5)                              ││
│ └──────────────────────────────────────────────────────────────┘│
│                                                                  │
│ DPort: [      ]  SIP: [              ]  DIP: [              ]   │
│ [ ] Black   [ ] White                              [+ 추가]     │
└─────────────────────────────────────────────────────────────────┘
```

#### ICMP Type/Code 선택 시 동작
1. Type 선택 → 해당 Type 값 설정
2. Type이 `destination-unreachable (3)`이면 → Code 드롭다운 표시
3. 다른 Type 선택 → Code 드롭다운 숨김, Code 값 초기화
4. Code 선택 → 해당 Code 값 설정
5. "커스텀 숫자..." 선택 → 숫자 입력 Entry 표시

### 7.3 규칙 추가 폼 - UDP/ANY 선택 시

```
┌─ 규칙 추가 ─────────────────────────────────────────────────────┐
│ Chain: [INPUT    v]  Proto: [UDP v]  Action: [DROP   v]         │
│                                                                  │
│ (프로토콜 옵션 없음)                                             │
│                                                                  │
│ DPort: [      ]  SIP: [              ]  DIP: [              ]   │
│ [ ] Black   [ ] White                              [+ 추가]     │
└─────────────────────────────────────────────────────────────────┘
```

### 7.4 규칙 테이블 컬럼 확장

| 컬럼 | 너비 | 위젯 | 설명 |
|------|------|------|------|
| 삭제 | 36px | Button | 행 삭제 버튼 |
| Chain | 100px | Select | 체인 선택 |
| Proto | 80px | Select | 프로토콜 선택 |
| 옵션 | 150px | Select/동적 | 프로토콜에 따른 옵션 (아래 참조) |
| Action | 90px | Select | 액션 선택 |
| DPort | 80px | Entry | 목적지 포트 |
| SIP | 140px | Entry | 소스 IP |
| DIP | 140px | Entry | 목적지 IP |
| B | 30px | Check | 블랙리스트 |
| W | 30px | Check | 화이트리스트 |

### 7.5 테이블 행에서 프로토콜별 옵션 UI

테이블의 각 행에서도 프로토콜에 따라 동적으로 옵션 UI가 변경됩니다.

#### TCP 선택 시
```
┌──────────────────────────────────────────────────────────────────────────────┐
│ [🗑] [INPUT v] [tcp v] [새 연결만 (SYN)     v] [DROP v] [포트] [SIP] [DIP] □ □│
└──────────────────────────────────────────────────────────────────────────────┘
```
- 옵션 컬럼: TCP Flags 프리셋 Select (없음, 새 연결만, 확립된 연결, NULL 스캔 차단, XMAS 스캔 차단, SYN+FIN 차단, 커스텀)

#### ICMP 선택 시
```
┌──────────────────────────────────────────────────────────────────────────────┐
│ [🗑] [INPUT v] [icmp v] [echo-request (8)  v] [DROP v] [포트] [SIP] [DIP] □ □│
└──────────────────────────────────────────────────────────────────────────────┘
```
- 옵션 컬럼: ICMP Type Select (없음, echo-request, echo-reply, destination-unreachable, time-exceeded, redirect, 커스텀 숫자...)
- Type이 `destination-unreachable (3)`인 경우 Code Select 추가 표시 필요 (공간 제약으로 팝업 또는 확장 방식 고려)

#### UDP/ANY 선택 시
```
┌──────────────────────────────────────────────────────────────────────────────┐
│ [🗑] [INPUT v] [udp v] [        -         ] [DROP v] [포트] [SIP] [DIP] □ □│
└──────────────────────────────────────────────────────────────────────────────┘
```
- 옵션 컬럼: "-" 텍스트 표시 (Label) 또는 비활성화된 Select

#### 프로토콜 변경 시 동작
1. 프로토콜 Select 변경 → 옵션 UI 동적 전환
2. 기존 옵션 값 초기화
3. 새 프로토콜에 맞는 옵션 UI 표시

---

## 8. 파일 구조

### 8.1 수정 파일

| 경로 | 수정 내용 |
|------|-----------|
| `internal/model/rule.go` | ProtocolOptions 구조체 추가, 헬퍼 함수 추가 |
| `internal/parser/rule_parser.go` | 쿼리 스트링 파싱/포맷 함수 추가 |
| `internal/ui/component/rule_form.go` | 동적 옵션 필드 추가 |
| `internal/ui/component/rule_row.go` | 옵션 컬럼 추가 |
| `internal/ui/component/rule_list.go` | 헤더 컬럼 추가 |

---

## 9. 하위 호환성

### 9.1 기존 규칙 처리

| 기존 형식 | 처리 방식 |
|-----------|-----------|
| `-p=tcp` | Options = nil, 정상 동작 |
| `-p=icmp` | Options = nil, 정상 동작 |

### 9.2 JSON 저장 형식

기존 contents 문자열에 새 형식이 그대로 저장됩니다.

```json
{
  "version": "v1.2.0",
  "contents": "agent -m=insert -c=INPUT -p=tcp?flags=syn/syn --dport=80 -a=DROP\nagent -m=insert -c=INPUT -p=icmp?type=echo-request -a=DROP"
}
```

### 9.3 백엔드 호환성

> **주의**: 백엔드 Agent 서버가 새 형식(`tcp?flags=`)을 지원하는지 확인 필요.
> 미지원 시 Agent 서버 업데이트가 선행되어야 함.

---

## 10. 검증 규칙

### 10.1 TCP Flags 검증

- 허용 플래그: syn, ack, fin, rst, psh, urg
- 형식: `{플래그목록}/{플래그목록}` (슬래시 필수)
- 플래그 구분: 쉼표(,)
- 대소문자: 소문자만 허용

### 10.2 ICMP Type 검증

- 숫자: 0~255 범위
- 이름: 정의된 이름만 허용
- Code: 0~255 범위 (선택)

---

## 11. 구현 체크리스트

구현 체크리스트는 별도 문서로 분리되었습니다.

- **체크리스트**: [protocol-options-checklist.md](./protocol-options-checklist.md)

---

## 12. 참조

- [TCP Flags Complete Guide](https://www.actualtests.com/blog/tcp-flags-explained-complete-guide-to-syn-ack-fin-rst-psh-urg-with-examples-and-tcp-header-format/)
- [iptables TCP flags](https://explainshell.com/explain?cmd=iptables+-A+INPUT+-p+tcp+--tcp-flags+SYN%2CRST%2CACK%2CFIN+SYN+-j+ACCEPT)
- [IANA ICMP Parameters](https://www.iana.org/assignments/icmp-parameters/icmp-parameters.xhtml)
- [Linux iptables ICMP](https://www.cyberciti.biz/tips/linux-iptables-9-allow-icmp-ping.html)
