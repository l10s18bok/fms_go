# 규칙 빌더 UI PRD v2.0 (Smartfw 문서 기반)

## 문서 정보
- **버전**: 2.0
- **작성일**: 2026-01-27
- **기준 문서**: Smartfw-manual.docx (3.3 일반 룰, 3.6 ACL)
- **적용 대상**: fms_fyne

---

## 1. 개요

### 1.1 배경
Smartfw-manual.docx 문서에 정의된 방화벽 규칙 형식에 맞춰 FMS 규칙 빌더 UI를 재정의합니다.

### 1.2 변경 사항
기존 rule-builder-prd.md 대비 주요 변경:
- `--sip=` → `-s=` (Source IP)
- `--dip=` → `--dest=` (Destination IP)
- Action: `REJECT` → `IDS`, `IPS`
- Interface 필드 추가: `-i=`, `-o=`
- ACL 지원: `-a=blocklist`, `-a=whitelist`

---

## 2. Smartfw 문서 기반 필드 정의

### 2.1 일반 룰 필드 (3.3 섹션)

| 요소 | 옵션 | 타입 | 설명 |
|------|------|------|------|
| Command | -m | string | insert, delete |
| Chain | -c | enum | PREROUTING, INPUT, FORWARD, OUTPUT, POSTROUTING |
| Action | -a | enum | ACCEPT, DROP, IDS, IPS |
| Source_ip | -s | string | CIDR, 범위(-), 콤마(,) 형식 지원 |
| Protocol | -p | enum | TCP, UDP, ICMP, ANY, IPS |
| Dest_ip | --dest | string | CIDR, 범위(-), 콤마(,) 형식 지원 |
| Dest_port | --dport | string | 범위(-), 콤마(,) 형식 지원 |
| In_Interface | -i | string | 단일 입력, 정규식 지원 (eth+, eth*, !eth0) |
| Out_Interface | -o | string | 단일 입력, 정규식 지원 (eth+, eth*, !eth0) |

### 2.2 ACL 필드 (5.6 섹션)

| 타입 | Action 값 | 설명 |
|------|-----------|------|
| blocklist | -a=blocklist | IP 차단 (블랙리스트) |
| whitelist | -a=whitelist | IP 허용 (화이트리스트) |
| searchacl | -a=searchacl | ACL 조회 |

### 2.3 IPS 필드 (5.5 섹션)

IPS 규칙은 프로토콜 필드에 `IPS?타입&파라미터` 형식으로 지정합니다.

**형식:**
```bash
agent -m=insert -p="IPS?타입&limit=N&seconds=N&enable=1" -a=IPS
```

**IPS 타입 목록 (24개):**

| 카테고리 | 타입 | 설명 |
|----------|------|------|
| IP Layer | land-attack | Source IP = Dest IP 패킷 차단 |
| | ip-spoofing | IP 스푸핑 공격 차단 |
| | ip-tunnel | GRE/IPIP 터널링 차단 |
| | ip-fragment | IP 조각화 공격 차단 |
| | ttl-attack | 비정상 TTL(≤1) 패킷 차단 |
| | port-scan | 포트 스캔 탐지/차단 |
| | ip-protocol | 비정상 프로토콜 차단 |
| | ip-options | IP Options 필드 패킷 차단 |
| | ip-fragment-tiny | Tiny Fragment 공격 차단 |
| | MCAST-DST-PING | 멀티캐스트 ICMP 차단 |
| TCP | syn-flood | SYN 패킷 대량 전송 공격 |
| | concurrent-conn | IP당 동시 세션 제한 |
| | synack-flood | SYN-ACK 플러드 공격 |
| | ack-flood | ACK 플러드 공격 |
| | rst-flood | RST 플러드 공격 |
| | fin-flood | FIN 플러드 공격 |
| | pshack-flood | PSH-ACK 플러드 공격 |
| | tcp-total | TCP 종합 플러드 공격 |
| UDP | udp-flood | UDP 패킷 대량 전송 공격 |
| | udp-bytes | UDP 대용량 페이로드 공격 |
| ICMP | icmp-flood | ICMP 패킷 대량 전송 공격 |
| | icmp-bytes | ICMP 대용량 페이로드 공격 |

### 2.4 명령어 예시

```bash
# 기본 방화벽 규칙
agent -m=insert -c=INPUT -p=tcp -a=DROP --dport=22 -s=203.248.252.2

# 인터페이스 지정
agent -m=insert -c=INPUT -p=tcp -a=ACCEPT --dport=80,443 -i=eth0

# 포워딩 규칙
agent -m=insert -c=FORWARD -p=any -a=ACCEPT -i=eth0 -o=eth1

# ACL (블랙리스트)
agent -m=insert -s=192.168.30.30 -a=blocklist

# ACL (화이트리스트)
agent -m=insert -s=10.0.0.0/8 -a=whitelist

# IPS 규칙
agent -m=insert -p="IPS?syn-flood&limit=50&seconds=1&enable=1" -a=IPS
agent -m=insert -p="IPS?land-attack&enable=1" -a=IPS
```

---

## 3. 데이터 모델

### 3.1 상수 정의

```go
// Chain 체인 타입
type Chain int
const (
    ChainINPUT       Chain = 0
    ChainOUTPUT      Chain = 1
    ChainFORWARD     Chain = 2
    ChainPREROUTING  Chain = 3  // NAT용
    ChainPOSTROUTING Chain = 4  // NAT용
)

// Protocol 프로토콜 타입
type Protocol int
const (
    ProtocolTCP  Protocol = 6
    ProtocolUDP  Protocol = 17
    ProtocolICMP Protocol = 1
    ProtocolANY  Protocol = 255
    ProtocolIPS  Protocol = 256  // IPS 규칙용
)

// Action 액션 타입
type Action int
const (
    ActionDROP   Action = 0  // 패킷 차단
    ActionACCEPT Action = 1  // 패킷 허용
    ActionIDS    Action = 2  // 허용 + 로그 (Intrusion Detection)
    ActionIPS    Action = 3  // 차단 + 로그 (Intrusion Prevention)
)
```

### 3.2 FirewallRule 구조체

```go
type FirewallRule struct {
    Chain        Chain            // 체인
    Protocol     Protocol         // 프로토콜
    Options      *ProtocolOptions // 프로토콜 옵션 (TCP Flags, ICMP Type 등)
    Action       Action           // 액션
    DPort        string           // 목적지 포트 (--dport)
    SIP          string           // 소스 IP (-s)
    DIP          string           // 목적지 IP (--dest)
    InInterface  string           // 입력 인터페이스 (-i)
    OutInterface string           // 출력 인터페이스 (-o)
    Black        bool             // 블랙리스트 여부
    White        bool             // 화이트리스트 여부
}
```

---

## 4. UI 설계

### 4.1 테이블 컬럼 (12개)

| 인덱스 | 컬럼명 | 너비 | 위젯 타입 | 편집 | 설명 |
|--------|--------|------|-----------|------|------|
| 0 | 삭제 | 36px | Button | - | 행 삭제 버튼 |
| 1 | Chain | 10% | Select | O | 체인 선택 |
| 2 | Proto | 8% | Select | O | 프로토콜 선택 |
| 3 | Options | 12% | Entry | O | flags=syn/syn, type=8 등 |
| 4 | Action | 10% | Select | O | 액션 선택 |
| 5 | Port | 8% | Entry | O | 목적지 포트 |
| 6 | SIP | 16% | Entry | O | 소스 IP |
| 7 | DIP | 16% | Entry | O | 목적지 IP |
| 8 | InIF | 6% | Entry | O | 입력 인터페이스 |
| 9 | OutIF | 6% | Entry | O | 출력 인터페이스 |
| 10 | Black | 55px | Check | O | 블랙리스트 |
| 11 | White | 55px | Check | O | 화이트리스트 |

### 4.2 규칙 추가 폼 레이아웃

```
┌─ 규칙 추가 ─────────────────────────────────────────────────────────────┐
│ Row 1: Chain [v] Proto [v] Action [v] Port [____]                       │
│ Row 2: SIP [____________] DIP [____________]                            │
│ Row 3: In IF [____] Out IF [____]                                       │
│ Row 4: (프로토콜별 옵션 - TCP Flags, ICMP Type 등)                      │
│                                                         [+ 추가]        │
└─────────────────────────────────────────────────────────────────────────┘
```

### 4.3 Select 옵션 목록

**Chain**:
- INPUT, OUTPUT, FORWARD
- (NAT용: PREROUTING, POSTROUTING)

**Protocol**:
- tcp, udp, icmp, any, IPS

**Action**:
- DROP, ACCEPT, IDS, IPS

---

## 5. 파서 인터페이스

### 5.1 명령어 매핑

| 필드 | 옵션 | 파싱 예시 |
|------|------|----------|
| Chain | -c= | -c=INPUT |
| Protocol | -p= | -p=tcp |
| Action | -a= | -a=DROP |
| DPort | --dport= | --dport=80,443 |
| SIP | -s= | -s=192.168.1.0/24 |
| DIP | --dest= | --dest=10.0.0.1 |
| InInterface | -i= | -i=eth0 |
| OutInterface | -o= | -o=eth1 |
| Black | -a=blocklist | (Action으로 처리) |
| White | -a=whitelist | (Action으로 처리) |

### 5.2 함수 시그니처

```go
// 단일 라인 파싱
func ParseLine(line string) (*FirewallRule, error)

// 규칙을 텍스트로 변환
func RuleToLine(rule *FirewallRule) string

// 전체 텍스트 파싱
func ParseTextToRules(text string) ([]*FirewallRule, []string, []error)

// 규칙 목록을 텍스트로 변환
func RulesToText(rules []*FirewallRule, comments []string) string
```

---

## 6. 파일 구조

### 6.1 관련 파일

| 경로 | 용도 |
|------|------|
| `internal/model/rule.go` | FirewallRule 구조체, 상수 정의 |
| `internal/parser/rule_parser.go` | 텍스트 ↔ 규칙 변환 |
| `internal/ui/component/rule_table.go` | 규칙 테이블 (12개 컬럼) |
| `internal/ui/component/rule_form.go` | 규칙 추가 폼 |
| `internal/ui/component/help_texts.go` | 도움말 텍스트 |

---

## 7. 변경 이력

| 버전 | 날짜 | 변경 내용 |
|------|------|----------|
| 1.0 | 2026-01-05 | 초기 버전 (rule-builder-prd.md) |
| 2.0 | 2026-01-27 | Smartfw 문서 기반 재정의 - Interface 컬럼 추가, Action 변경, 옵션 매핑 수정 |
