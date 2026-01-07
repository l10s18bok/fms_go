# NAT 규칙 확장 PRD (Product Requirements Document)

## 문서 정보
- **버전**: 1.0
- **작성일**: 2026-01-06
- **적용 대상**: fms_fyne, fms_wails
- **관련 문서**: [rule-builder-prd.md](./rule-builder-prd.md), [protocol-options-prd.md](./protocol-options-prd.md)

---

## 1. 개요

### 1.1 배경
현재 FMS는 기본적인 방화벽 필터 규칙(INPUT, OUTPUT, FORWARD)만 지원합니다. 네트워크 주소 변환(NAT) 규칙을 추가하여 포트 포워딩(DNAT), 소스 NAT(SNAT) 등의 기능을 지원해야 합니다.

### 1.2 목표
- DNAT (Destination NAT) 규칙 지원 - 포트 포워딩
- SNAT (Source NAT) 규칙 지원 - 소스 주소 변환
- 인터페이스 기반 규칙 지원 (in_interface, out_interface)
- 기존 규칙 빌더와 통합

### 1.3 핵심 원칙
- 기존 필터 규칙과 분리된 NAT 전용 UI
- smartfw 커널 모듈 형식과 호환
- 직관적인 포트 포워딩 설정 UI

---

## 2. NAT 개념

### 2.1 DNAT (Destination NAT)

외부에서 들어오는 트래픽의 **목적지 주소/포트**를 변환합니다.

**용도**: 포트 포워딩, 로드 밸런싱

**예시**: 외부 6080 포트 → 내부 192.168.30.180:8080 으로 전달

```
iptables -t nat -A PREROUTING -p tcp --dport 6080 -j DNAT --to-destination 192.168.30.180:8080
```

### 2.2 SNAT (Source NAT)

내부에서 나가는 트래픽의 **소스 주소**를 변환합니다.

**용도**: 여러 내부 호스트가 하나의 공인 IP 공유

**예시**: 192.168.45.0/24 네트워크 → 공인 IP로 변환하여 외부 통신

```
iptables -t nat -A POSTROUTING -s 192.168.45.0/24 -o eth0 -j SNAT --to-source 203.0.113.1
```

### 2.3 MASQUERADE

SNAT의 특수 형태로, 동적 IP 환경에서 사용합니다.

```
iptables -t nat -A POSTROUTING -s 192.168.1.0/24 -o eth0 -j MASQUERADE
```

---

## 3. smartfw 명령어 형식

### 3.1 DNAT 형식

```
req|INSERT|{ID}|{CHAIN}|NAT|{SRC}|{PROTOCOL}?DNAT|{DEST_IP}|{MATCH_PORT},{TRANSLATE_PORT}|{IN_IF}|{OUT_IF}
```

**예시**:
```bash
# 외부 6080 → 내부 192.168.30.180:8080
echo "req|INSERT|3813792919|ANY|NAT|ANY|TCP?DNAT|192.168.30.180|6080,8080||" > /proc/smartfw
```

### 3.2 SNAT 형식

```
req|INSERT|{ID}|{CHAIN}|NAT|{SRC}|{PROTOCOL}?SNAT|{DEST}|{PORTS}|{IN_IF}|{OUT_IF}
```

**예시**:
```bash
# 192.168.45.0/24 → SNAT 적용
echo "req|INSERT|3813792919|ANY|NAT|192.168.45.0/24|TCP?SNAT|ANY|ANY|in_interface|out_interface" > /proc/smartfw
```

---

## 4. 데이터 모델

### 4.1 NATType 열거형

```go
type NATType int

const (
    NATTypeDNAT       NATType = 0 // Destination NAT (포트 포워딩)
    NATTypeSNAT       NATType = 1 // Source NAT
    NATTypeMASQUERADE NATType = 2 // Masquerade
)
```

### 4.2 NATRule 구조체

```go
// NATRule NAT 규칙 구조체
type NATRule struct {
    // 기본 필드
    NATType      NATType  // DNAT, SNAT, MASQUERADE
    Protocol     Protocol // TCP, UDP, ANY

    // 매칭 조건
    MatchIP      string   // 매칭할 IP (소스 또는 목적지)
    MatchPort    string   // 매칭할 포트

    // 변환 대상
    TranslateIP   string  // 변환할 IP
    TranslatePort string  // 변환할 포트

    // 인터페이스
    InInterface  string   // 입력 인터페이스 (예: eth0)
    OutInterface string   // 출력 인터페이스 (예: eth1)

    // 추가 옵션
    Description  string   // 규칙 설명 (선택)
}
```

### 4.3 DNAT 규칙 예시

```go
// 외부 6080 → 내부 192.168.30.180:8080
rule := &NATRule{
    NATType:       NATTypeDNAT,
    Protocol:      ProtocolTCP,
    MatchIP:       "ANY",           // 모든 소스에서
    MatchPort:     "6080",          // 6080 포트로 들어오면
    TranslateIP:   "192.168.30.180", // 이 IP의
    TranslatePort: "8080",          // 8080 포트로 전달
}
```

### 4.4 SNAT 규칙 예시

```go
// 내부 네트워크 → 외부 통신 시 소스 변환
rule := &NATRule{
    NATType:      NATTypeSNAT,
    Protocol:     ProtocolTCP,
    MatchIP:      "192.168.45.0/24", // 이 네트워크에서 나가는 트래픽
    TranslateIP:  "ANY",             // 자동 (또는 특정 공인 IP)
    InInterface:  "eth1",            // 내부 인터페이스
    OutInterface: "eth0",            // 외부 인터페이스
}
```

---

## 5. 명령어 생성

### 5.1 agent 명령어 형식

기존 필터 규칙과 구분하기 위해 NAT 전용 형식 사용:

```bash
# DNAT
agent -m=insert -t=nat --nat-type=dnat -p=tcp --match-port=6080 --to-dest=192.168.30.180:8080

# SNAT
agent -m=insert -t=nat --nat-type=snat -p=tcp -s=192.168.45.0/24 -i=eth1 -o=eth0
```

### 5.2 smartfw 형식 변환

```go
// NATRuleToSmartfw NAT 규칙을 smartfw 형식으로 변환
func NATRuleToSmartfw(rule *NATRule, id string) string {
    switch rule.NATType {
    case NATTypeDNAT:
        return fmt.Sprintf("req|INSERT|%s|ANY|NAT|%s|%s?DNAT|%s|%s,%s||",
            id,
            rule.MatchIP,
            ProtocolToString(rule.Protocol),
            rule.TranslateIP,
            rule.MatchPort,
            rule.TranslatePort,
        )
    case NATTypeSNAT:
        return fmt.Sprintf("req|INSERT|%s|ANY|NAT|%s|%s?SNAT|%s|%s|%s|%s",
            id,
            rule.MatchIP,
            ProtocolToString(rule.Protocol),
            rule.TranslateIP,
            rule.MatchPort,
            rule.InInterface,
            rule.OutInterface,
        )
    }
    return ""
}
```

---

## 6. UI 설계

### 6.1 NAT 탭 추가

템플릿 편집기에 "NAT 규칙" 서브 탭 추가:

```
┌──────────────────────────────────────────────────────────────────────┐
│ 템플릿 탭                                                            │
├────────────────┬─────────────────────────────────────────────────────┤
│                │  [텍스트 편집] [규칙 빌더] [NAT 규칙]  <- 서브 탭   │
│  좌측 패널     ├─────────────────────────────────────────────────────┤
│  (템플릿 목록) │  ┌───────────────────────────────────────────────┐  │
│                │  │ NAT 규칙 테이블 (widget.Table 기반)           │  │
│  ○ v1.1.0      │  │ - 삭제, 타입, 프로토콜, 매칭, 변환, 인터페이스│  │
│  ○ v1.0.0      │  └───────────────────────────────────────────────┘  │
│                │                                                     │
│                │  [DNAT (포트 포워딩)] [SNAT/MASQ]  <- 폼 전환 탭    │
│                │  ┌───────────────────────────────────────────────┐  │
│                │  │ (선택된 NAT 타입에 따라 다른 폼 표시)          │  │
│                │  └───────────────────────────────────────────────┘  │
└────────────────┴─────────────────────────────────────────────────────┘
```

> **Note**: 필터 규칙 빌더와 동일한 패턴 적용:
> - 테이블은 `widget.Table` 기반 (고정 너비 + 비율 컬럼)
> - 폼은 탭 구조로 NAT 타입별 분리

### 6.2 DNAT 규칙 추가 폼 (포트 포워딩)

```
┌─ 포트 포워딩 (DNAT) 추가 ───────────────────────────────────────────┐
│                                                                      │
│ 타입: [DNAT (포트 포워딩) v]                                         │
│       ├─ DNAT (포트 포워딩)                                          │
│       ├─ SNAT (소스 NAT)                                             │
│       └─ MASQUERADE                                                  │
│                                                                      │
│ ┌─ 매칭 조건 ──────────────────────────────────────────────────────┐│
│ │ 프로토콜: [TCP v]                                                 ││
│ │ 외부 포트: [6080    ] (들어오는 포트)                             ││
│ │ 소스 IP:   [ANY     ] (허용할 소스, 비우면 모두 허용)             ││
│ └──────────────────────────────────────────────────────────────────┘│
│                                                                      │
│ ┌─ 전달 대상 ──────────────────────────────────────────────────────┐│
│ │ 내부 IP:   [192.168.30.180]                                       ││
│ │ 내부 포트: [8080          ]                                       ││
│ └──────────────────────────────────────────────────────────────────┘│
│                                                                      │
│ 설명: [웹 서버 포트 포워딩                    ] (선택)               │
│                                                                      │
│                                                    [+ 추가]          │
└──────────────────────────────────────────────────────────────────────┘
```

### 6.3 SNAT 규칙 추가 폼

```
┌─ 소스 NAT (SNAT) 추가 ──────────────────────────────────────────────┐
│                                                                      │
│ 타입: [SNAT (소스 NAT) v]                                            │
│                                                                      │
│ ┌─ 매칭 조건 ──────────────────────────────────────────────────────┐│
│ │ 프로토콜:    [TCP v]                                              ││
│ │ 소스 네트워크: [192.168.45.0/24] (SNAT 적용할 내부 네트워크)      ││
│ └──────────────────────────────────────────────────────────────────┘│
│                                                                      │
│ ┌─ 인터페이스 ─────────────────────────────────────────────────────┐│
│ │ 입력 인터페이스: [eth1    ] (내부)                                ││
│ │ 출력 인터페이스: [eth0    ] (외부)                                ││
│ └──────────────────────────────────────────────────────────────────┘│
│                                                                      │
│ ┌─ 변환 대상 (선택) ───────────────────────────────────────────────┐│
│ │ 변환 IP: [            ] (비우면 자동)                             ││
│ └──────────────────────────────────────────────────────────────────┘│
│                                                                      │
│ 설명: [내부 네트워크 외부 통신용              ] (선택)               │
│                                                                      │
│                                                    [+ 추가]          │
└──────────────────────────────────────────────────────────────────────┘
```

### 6.4 NAT 규칙 테이블

| 컬럼 | 너비 | 설명 |
|------|------|------|
| 삭제 | 36px | 삭제 버튼 |
| 타입 | 80px | DNAT/SNAT/MASQ |
| 프로토콜 | 70px | TCP/UDP/ANY |
| 매칭 | 200px | 소스 IP/포트 |
| 변환 | 200px | 대상 IP/포트 |
| 인터페이스 | 120px | IN/OUT |
| 설명 | 150px | 규칙 설명 |

---

## 7. 파일 구조

### 7.1 신규 파일

| 경로 | 용도 |
|------|------|
| `internal/model/nat_rule.go` | NAT 규칙 데이터 모델 |
| `internal/parser/nat_parser.go` | NAT 규칙 파싱/변환 |
| `internal/ui/component/nat_table.go` | NAT 규칙 테이블 (widget.Table 기반) |
| `internal/ui/component/dnat_form.go` | DNAT 규칙 추가 폼 |
| `internal/ui/component/snat_form.go` | SNAT/MASQ 규칙 추가 폼 |
| `internal/ui/nat_builder.go` | NAT 빌더 패널 (테이블 + 폼 탭) |

> **Note**: 필터 규칙 빌더(RuleTable, RuleForm, BlackWhiteForm)와 동일한 패턴 적용

### 7.2 수정 파일

| 경로 | 수정 내용 |
|------|-----------|
| `internal/ui/template_tab.go` | NAT 규칙 서브 탭 추가 |

---

## 8. 구현 체크리스트

구현 체크리스트는 별도 문서로 분리되었습니다.

- **체크리스트**: [nat-rules-checklist.md](./nat-rules-checklist.md)

---

## 9. 참조

- [iptables NAT 설정 가이드](https://masterdaweb.com/en/blog/examples-of-snat-dnat-with-iptables)
- [pfSense Port Forwarding](https://docs.netgate.com/pfsense/en/latest/nat/port-forwards.html)
- [OPNsense NAT 문서](https://docs.opnsense.org/manual/nat.html)
- [UniFi NAT 설정](https://help.ui.com/hc/en-us/articles/16437942532759-DNAT-SNAT-and-Masquerading-in-UniFi)
- [Linux NAT Masquerade](https://www.geeksforgeeks.org/linux-unix/using-masquerading-with-iptables-for-network-address-translation-nat/)
