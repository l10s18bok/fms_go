# NAT 규칙 PRD v2.0 (Smartfw 문서 기반)

## 문서 정보
- **버전**: 2.0
- **작성일**: 2026-01-27
- **기준 문서**: Smartfw-manual.docx (3.4 NAT 룰)
- **적용 대상**: fms_fyne

---

## 1. 개요

### 1.1 배경
Smartfw-manual.docx 문서에 정의된 NAT 규칙 형식에 맞춰 FMS NAT 빌더 UI를 재정의합니다.

### 1.2 변경 사항
기존 nat-rules-prd.md 대비 주요 변경:
- `--nat-type=dnat` → `-p="TCP?DNAT"`
- `--match-port=` → `--dport=`
- `--to-dest=` → `--dest=`
- Action: `-a=NAT` 추가

---

## 2. Smartfw 문서 기반 NAT 규칙 형식

### 2.1 SNAT (Source NAT)

내부에서 외부로 나가는 트래픽의 **소스 주소**를 변환합니다.

**형식:**
```bash
agent -m=insert -s=SOURCE_IP -p="PROTO?SNAT" --dest=TRANSLATE_IP -a=NAT
```

**예시:**
```bash
# 192.168.0.0/24 네트워크 → 203.0.113.10 IP로 SNAT
agent -m=insert -s=192.168.0.0/24 -p="ANY?SNAT" --dest=203.0.113.10 -a=NAT
```

### 2.2 DNAT (Destination NAT / 포트 포워딩)

외부에서 들어오는 트래픽의 **목적지 주소/포트**를 변환합니다.

**형식:**
```bash
agent -m=insert -p="PROTO?DNAT" --dest=DEST_IP:DEST_PORT --dport=MATCH_PORT -a=NAT
```

**예시:**
```bash
# 외부 80 포트 → 내부 192.168.0.10:8080으로 포워딩
agent -m=insert -p="TCP?DNAT" --dest=192.168.0.10:8080 --dport=80 -a=NAT
```

### 2.3 MASQUERADE

SNAT의 특수 형태로, 동적 IP 환경에서 사용합니다. `--dest` 없이 `-o` 인터페이스만 지정합니다.

**형식:**
```bash
agent -m=insert -s=SOURCE_IP -p="PROTO?SNAT" -o=OUT_IF -a=NAT
```

**예시:**
```bash
# eth0 인터페이스의 IP로 자동 MASQUERADE
agent -m=insert -s=192.168.0.0/24 -p="ANY?SNAT" -o=eth0 -a=NAT
```

---

## 3. 데이터 모델

### 3.1 상수 정의

```go
// NATType NAT 타입
type NATType int
const (
    NATTypeDNAT       NATType = 0  // Destination NAT (포트 포워딩)
    NATTypeSNAT       NATType = 1  // Source NAT
    NATTypeMASQUERADE NATType = 2  // Masquerade (동적 IP용 SNAT)
)
```

### 3.2 NATRule 구조체

```go
type NATRule struct {
    // 기본 필드
    NATType      NATType   // DNAT, SNAT, MASQUERADE
    Protocol     Protocol  // TCP, UDP, ANY

    // 매칭 조건
    MatchIP      string    // 소스 IP (-s)
    MatchPort    string    // 매칭 포트 (--dport)

    // 변환 대상
    TranslateIP   string   // 변환할 IP (--dest)
    TranslatePort string   // 변환할 포트 (--dest의 :PORT 부분)

    // 인터페이스
    InInterface  string    // 입력 인터페이스 (-i)
    OutInterface string    // 출력 인터페이스 (-o)

    // 추가 옵션
    Description  string    // 규칙 설명
}
```

---

## 4. UI 설계

### 4.1 NAT 테이블 컬럼 (9개)

| 인덱스 | 컬럼명 | 너비 | 위젯 타입 | 설명 |
|--------|--------|------|-----------|------|
| 0 | 삭제 | 36px | Button | 행 삭제 버튼 |
| 1 | Type | 100px | Label | SNAT/DNAT/MASQUERADE |
| 2 | Proto | 70px | Label | TCP/UDP/ANY |
| 3 | MatchIP | 18% | Label | 소스 IP (-s) |
| 4 | TransIP | 18% | Label | 변환 IP (--dest) |
| 5 | MatchPort | 10% | Label | 매칭 포트 (--dport) |
| 6 | TransPort | 10% | Label | 변환 포트 |
| 7 | InIF | 8% | Label | 입력 인터페이스 |
| 8 | OutIF | 8% | Label | 출력 인터페이스 |

### 4.2 NAT 타입별 필드 사용

| 필드 | DNAT | SNAT | MASQUERADE |
|------|------|------|------------|
| MatchIP | 선택 | 필수 | 필수 |
| TransIP | 필수 | 필수 | - |
| MatchPort | 필수 | - | - |
| TransPort | 선택 | - | - |
| InIF | 선택 | - | - |
| OutIF | - | 선택 | 필수 |

### 4.3 DNAT 폼 레이아웃

```
┌─ 포트 포워딩 (DNAT) 추가 ───────────────────────────────────────────────┐
│ Row 1: Proto [TCP v]  MatchPort [____]  MatchIP [________]  [?]         │
│ Row 2: TransIP [__________]  TransPort [____]  InIF [____]              │
│                                                         [+ 추가]        │
└─────────────────────────────────────────────────────────────────────────┘
```

### 4.4 SNAT/MASQUERADE 폼 레이아웃

```
┌─ 소스 NAT (SNAT/MASQUERADE) 추가 ───────────────────────────────────────┐
│ Row 1: Proto [ANY v]  MatchIP [__________]  [?]                         │
│ Row 2: OutIF [____]  TransIP [__________]                               │
│                                                         [+ 추가]        │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 5. 파서 인터페이스

### 5.1 명령어 매핑

| 필드 | 옵션 | 파싱 예시 |
|------|------|----------|
| Protocol+NATType | -p= | -p="TCP?DNAT" |
| MatchIP | -s= | -s=192.168.0.0/24 |
| MatchPort | --dport= | --dport=80 |
| TranslateIP:Port | --dest= | --dest=192.168.0.10:8080 |
| InInterface | -i= | -i=eth0 |
| OutInterface | -o= | -o=eth0 |
| NAT Action | -a=NAT | (필수) |

### 5.2 프로토콜+NAT타입 형식

```
-p="PROTO?NAT_TYPE"

예시:
- -p="TCP?DNAT"       → Protocol=TCP, NATType=DNAT
- -p="UDP?DNAT"       → Protocol=UDP, NATType=DNAT
- -p="ANY?SNAT"       → Protocol=ANY, NATType=SNAT
- -p="ANY?MASQUERADE" → Protocol=ANY, NATType=MASQUERADE
```

### 5.3 함수 시그니처

```go
// NAT 라인 파싱
func ParseNATLine(line string) (*NATRule, error)

// NATRule을 agent 명령어로 변환
func NATRuleToLine(rule *NATRule) string

// NATRule을 smartfw 형식으로 변환
func NATRuleToSmartfw(rule *NATRule, id string) string

// 전체 텍스트에서 NAT 규칙 추출
func ParseTextToNATRules(text string) ([]*NATRule, []string, []error)

// NAT 규칙 라인인지 확인
func IsNATLine(line string) bool
```

---

## 6. smartfw 커널 모듈 형식

### 6.1 DNAT 형식

```
req|INSERT|{ID}|ANY|NAT|{MatchIP}|{Proto}?DNAT|{TransIP}|{MatchPort},{TransPort}|{InIF}|{OutIF}
```

예시:
```
req|INSERT|3813792919|ANY|NAT|ANY|TCP?DNAT|192.168.30.180|6080,8080||
```

### 6.2 SNAT 형식

```
req|INSERT|{ID}|ANY|NAT|{MatchIP}|{Proto}?SNAT|{TransIP}|{Ports}|{InIF}|{OutIF}
```

예시:
```
req|INSERT|3813792919|ANY|NAT|192.168.45.0/24|TCP?SNAT|ANY|ANY|eth1|eth0
```

### 6.3 MASQUERADE 형식

```
req|INSERT|{ID}|ANY|NAT|{MatchIP}|{Proto}?MASQUERADE|ANY|ANY|{InIF}|{OutIF}
```

예시:
```
req|INSERT|3813792919|ANY|NAT|192.168.0.0/24|ANY?MASQUERADE|ANY|ANY||eth0
```

---

## 7. 파일 구조

### 7.1 관련 파일

| 경로 | 용도 |
|------|------|
| `internal/model/nat_rule.go` | NATRule 구조체, NATType 상수 |
| `internal/parser/nat_parser.go` | NAT 규칙 파싱/변환 |
| `internal/ui/component/nat_table.go` | NAT 테이블 (9개 컬럼) |
| `internal/ui/component/dnat_form.go` | DNAT 규칙 추가 폼 |
| `internal/ui/component/snat_form.go` | SNAT/MASQ 규칙 추가 폼 |
| `internal/ui/component/help_texts.go` | NAT 도움말 텍스트 |

---

## 8. 변경 이력

| 버전 | 날짜 | 변경 내용 |
|------|------|----------|
| 1.0 | 2026-01-06 | 초기 버전 (nat-rules-prd.md) |
| 2.0 | 2026-01-27 | Smartfw 문서 기반 재정의 - 명령어 형식 변경, 테이블 컬럼 분리 |
