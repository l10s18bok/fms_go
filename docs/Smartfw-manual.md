# Smartfw 매뉴얼

**2025. 12. 34**

---

## 목차

1. [Smartfw 개요](#1-smartfw-개요)
   - 1.1 [구성 프로그램](#11-구성-프로그램)
   - 1.2 [용어 설명](#12-용어-설명)
2. [L3/L4 IDS/IPS 기능](#2-l3l4-idsips-기능)
   - 2.1 [IP Layer](#21-ip-layer)
   - 2.1.1 [TCP](#211-tcp)
   - 2.1.2 [UDP](#212-udp)
   - 2.1.3 [ICMP](#213-icmp)
3. [L4 기본 방화벽](#3-l4-기본-방화벽)
4. [방화벽 운영 참고 사항](#4-방화벽-운영-참고-사항)
   - 4.1 [메모리 사용량](#41-메모리-사용량)
   - 4.2 [제약 사항](#42-제약-사항)
5. [방화벽 명령어 가이드](#5-방화벽-명령어-가이드)
   - 5.1 [기본 명령어](#51-기본-명령어)
   - 5.2 [로그 레벨 변경](#52-로그-레벨-변경)
   - 5.3 [일반 룰](#53-일반-룰)
   - 5.4 [NAT 룰](#54-nat-룰)
   - 5.5 [IPS 룰](#55-ips-룰)
   - 5.6 [ACL (화이트리스트 & 블록리스트)](#56-acl-화이트리스트--블록리스트)
   - 5.7 [적용된 룰 확인](#57-적용된-룰-확인)
   - 5.8 [룰 적용 히스토리](#58-룰-적용-히스토리)
6. [부록](#6-부록)
   - 6.1 [성능 테스트](#61-성능-테스트)

---

## 수정 이력

| No. | 날짜 | 작성자 | 수정내용 | 참고 |
|-----|------|--------|----------|------|
| 1 | 2025-12-30 | 이희석 | 초안작성 | 1.0.0 |

---

## 1. Smartfw 개요

SmartFW는 Linux 커널 공간에서 동작하는 Hash 베이스 네트워크 방화벽 시스템으로, Netfilter 프레임워크와 ConnTrack 기반의 커널 모듈로 구현되었다.

L4의 구성 요소인 IP, PORT, PROTOCOL 기반의 방어 룰 적용과 더불어 여러 공격 기법을 방어할 수 있는 IDS/IPS 기반의 필터링도 가능하다.

다음 목차에서부터 기능 구성 및 설정 가능한 항목들에 대해 정의한다.

### 1.1 구성 프로그램

| 이름 | 정의 |
|------|------|
| Smartfw | 커널에 설치될 커널 방화벽 |
| Smartfw_agent | 커널 방화벽의 룰 관리 및 방화벽으로부터 노티를 받는 어플리케이션 |

### 1.2 용어 설명

| 용어 | 정의 |
|------|------|
| Action | 룰 필터에서 감지 했을 때 동작할 행위 |
| | - **Accept**: 허용 |
| | - **DROP**: 차단 |
| | - **IDS**: 패킷은 허용, NOTI 전송 |
| | - **IPS**: 패킷차단, NOTI 전송 |

---

## 2. L3/L4 IDS/IPS 기능

Smartfw에서 지원하는 IDS/IPS 관련 기능으로는 다음과 같으며, 정의된 타입 별로 룰을 통해 수치 및 액션, 사용여부에 대한 조정이 가능하다.

현재 버전 기준으로는 기본 값으로 오탐율이 적고 필수적으로 적용해야 할 룰에 대해서는 필터가 활성화되고, 기본 Action이 IPS이며 그 외에는 Action이 IDS 또는 사용 안함으로 설정되어 있다.

또한 시간 주기가 필요한 룰들의 경우 예시로 TCP SYN-FLOOD와 같이 기본 seconds 설정은 1초이며, 설정할 경우 변경이 가능하다.

개별 항목에 대한 설정은 아래 타입별로 설정을 확인할 수 있다.

### 2.1 IP Layer

IP Layer의 필터링은 프로토콜과 연관없이 Prerouting Chain에서 적용되며, 운영 환경에 따라 조정이 필요할 수 있다.

| 필터명 | 정의 |
|--------|------|
| **land-attack** | Source IP = Destination IP인 패킷 탐지 |
| **ip-spoofing** | IP Spoofing 공격 차단 (외부에서 불가능한 Source IP) |
| **Multicast Source** (224.0.0.0/4 ~ 239.255.255.255) | Multicast IP를 Source로 사용하는 패킷 차단. Multicast는 목적지로만 사용 가능 |
| **Broadcast Source** (255.255.255.255) | Broadcast IP를 Source로 사용하는 패킷 차단 |
| **Reserved IP** - 0.0.0.0/8 | 0.0.0.0/8 범위를 Source로 사용하는 패킷 차단. 단, 0.0.0.0은 DHCP에서 사용하므로 별도 처리 |
| **Class E Reserved** (240.0.0.0/4) | Class E 예약 주소 (240~255)를 Source로 사용하는 패킷 차단 |
| **0.0.0.0 Source** (DHCP 제외) | 0.0.0.0을 Source로 사용하는 패킷 차단. 예외: UDP 68→67 (DHCP Client → Server)는 허용 |
| **ip-tunnel** | IP 터널링 프로토콜 차단. GRE (Protocol 47): Generic Routing Encapsulation, IPIP (Protocol 4): IP-in-IP tunneling |
| **ip-fragment** | IP Fragmentation 기반 공격 차단 |
| | - Fragment Offset 조작: Fragment Offset > 65535 - 20 과 같이 IP패킷 최대크기 초과 시도 |
| | - Teardrop Attack: Fragment의 끝 위치가 65535 초과 |
| | - Reserved Flag 설정: IP Header Flags의 Reserved 비트(최상위 비트) 설정 |
| | - DF + MF 동시 설정: Don't Fragment와 More Fragments 플래그 동시 설정 |
| **ttl-attack** | 비정상적으로 낮은 TTL 패킷 차단. TTL ≤ 1. 멀티캐스트는 예외적으로 허용 |
| **port-scan** | 포트 스캔 공격 탐지 및 차단 |
| **ip-protocol** | 설정한 프로토콜 외 비정상 프로토콜로 판단. TCP, UDP, ICMP 프로토콜 외 비정상으로 판단 |
| **ip-options** | IP Options 필드를 가진 패킷 차단. IP Header Length (IHL) > 5 (즉, Options 존재). 멀티캐스트(224.x.x.x \| 239.x.x.x)는 Router Alert 옵션을 정상적으로 사용하므로 제외 |
| **ip-fragment-tiny** | Tiny Fragment Attack 차단. First Fragment (Offset=0, MF=1), 패킷 크기 < 68 bytes |
| **MCAST-DST-PING** | 멀티캐스트 목적지로 ICMP(PING)을 보내는 패킷 차단 |

#### IP Layer 필터 기본 설정

| 필터명 | 기본 활성화 여부 | 기본 Limit | 기본 seconds(초) | 기본 Action |
|--------|------------------|------------|------------------|-------------|
| land-attack | O (1) | X | X | IPS |
| ip-spoofing | O (1) | X | X | IPS |
| ip-tunnel | O (1) | X | X | IPS |
| ip-fragment | O (1) | X | X | DROP |
| ttl-attack | O (1) | X | X | DROP |
| port-scan | O (1) | 32 | 1 | IPS |
| ip-protocol | X (0) | X | X | IPS |
| ip-options | O (1) | X | X | IDS |
| ip-fragment-tiny | O (1) | X | X | IDS |
| MCAST-DST-PING | O (1) | | | IPS |

### 2.1.1 TCP

#### 2.1.1.1 기본 필터

기본 필터의 경우 정책을 바꿀 수 없으며, 운영 환경에서 무조건적으로 방어 및 허용해야 할 필터이다.

| 필터명 | 정의 |
|--------|------|
| **flag-invalid** | 비정상적인 TCP 플래그가 셋팅되어 온 패킷 |
| | - NULL SCAN: 모든 플래그가 0 |
| | - XMAS SCAN: FIN+PSH+URG 플래그 1 |
| | - SYN+FIN: 모순되는 플래그 조합(연결시작과 종료) |
| | - SYN+RST: 모순되는 플래그 조합(연결시작과 리셋) |
| | - FIN+RST: 모순되는 플래그 조합(종료와 리셋) |
| | - ALL FLAG: 모든 플래그가 1 |
| **conntrack-invalid** | 세션을 관리하는 Conntrack에서 Invalid 상태가 나온 패킷 |
| **new-Syn** | TCP에서 신규 세션이 들어왔을 때 Syn플래그만 활성화 된 패킷이 아닐 경우 |
| **ESTABLISHED/RELATED** | 기본 룰 및 IDS/IPS 필터가 통과된 신뢰있는 세션으로, 최초 세션 연결에서 동시 접속 제한에 걸리지 않을 경우 무조건 허용된다. |

#### 2.1.1.2 IDS/IPS 필터

기능 구성은 정상적인 플래그를 통한 공격을 방어하기 위해 FLOOD 공격 방어와, 여러 FLOOD 공격을 혼합하는것을 대비한 TOTAL, 장시간에 거쳐 한 개의 IP가 동시에 세션을 지속적으로 늘리는 공격을 방어한다.

기본 설정은 Syn-Flood와 Concurrent-conn 기능만 활성화되어 있으며, 각 항목별로 IP || Port || IP+PORT로 예외적으로 별도의 정책을 적용할 수 있다.

| 필터명 | 정의 |
|--------|------|
| **syn-flood** | SYN 플래그를 활성화한 패킷을 통한 공격 |
| **concurrent-conn** | 1개 IP당 TCP 세션을 지속적으로 늘리는 공격 |
| **synack-flood** | SYN-ACK 플래그를 활성화한 패킷을 통한 공격 |
| **ack-flood** | ACK 플래그를 활성화한 패킷을 통한 공격 |
| **rst-flood** | RST 플래그를 활성화한 패킷을 통한 공격 |
| **fin-flood** | FIN 플래그를 활성화한 패킷을 통한 공격 |
| **pshack-flood** | PSH-ACK 플래그를 활성화한 패킷을 통한 공격 |
| **tcp-total** | 위 여러가지 유형의 패킷을 가변적으로 사용하여 공격 |

#### TCP IDS/IPS 필터 기본 설정

| 필터명 | 기본 활성화 여부 | 기본 Limit | 기본 seconds(초) | 기본 Action |
|--------|------------------|------------|------------------|-------------|
| syn-flood | O (1) | 50개 | 1 | IPS |
| concurrent-conn | O (1) | 200개 | 1 | IPS |
| synack-flood | X (0) | 1000개 | 1 | IPS |
| ack-flood | X (0) | 2000개 | 1 | IPS |
| rst-flood | X (0) | 500개 | 1 | IPS |
| fin-flood | X (0) | 500개 | 1 | IPS |
| pshack-flood | X (0) | 2000개 | 1 | IPS |
| tcp-total | X (0) | 3000개 | 1 | IPS |

### 2.1.2 UDP

#### IDS/IPS 필터

기능 구성은 UDP 패킷을 지속적으로 전송하는 FLOOD 공격 방어와, 사이즈가 큰 UDP 페이로드를 이용한 공격 방어가 있다.

각 항목별로 IP || Port || IP+PORT로 예외적으로 별도의 정책을 적용할 수 있다.

| 필터명 | 정의 |
|--------|------|
| **udp-flood** | UDP 패킷을 지속적으로 전송하여 공격 |
| **udp-bytes** | UDP 패킷의 전송횟수는 적으나, 사이즈가 큰 페이로드 패킷으로 공격 |

#### UDP IDS/IPS 필터 기본 설정

| 필터명 | 기본 활성화 여부 | 기본 Limit | 기본 seconds(초) | 기본 Action |
|--------|------------------|------------|------------------|-------------|
| udp-flood | O (1) | 500개 | 1 | IPS |
| udp-bytes | O (1) | 50000byte(50kb) | 1 | IPS |

### 2.1.3 ICMP

#### IDS/IPS 필터

기능 구성은 ECHO_REQ 패킷을 지속적으로 전송하는 FLOOD 공격 방어와, 사이즈가 큰 ICMP 패킷을 전송하는 공격 방어가 있다.

각 항목별로 IP || Port || IP+PORT로 예외적으로 별도의 정책을 적용할 수 있다.

| 필터명 | 정의 |
|--------|------|
| **icmp-flood** | ICMP 패킷을 지속적으로 전송하여 공격 |
| **icmp-bytes** | ICMP 패킷의 전송횟수는 적으나, 사이즈가 큰 페이로드 패킷으로 공격 |

#### ICMP IDS/IPS 필터 기본 설정

| 필터명 | 기본 활성화 여부 | 기본 Limit | 기본 seconds(초) | 기본 Action |
|--------|------------------|------------|------------------|-------------|
| icmp-flood | O (1) | 50개 | 1 | IPS |
| icmp-bytes | O (1) | 20000byte(20kb) | 1 | IPS |

---

## 3. L4 기본 방화벽

기존 리눅스의 Iptables(firewalld)에서 지원하는 L4 레이어의 기본적인 방화벽 룰은 모두 지원한다.

| 요소 | 정의 |
|------|------|
| **Chain** | PREROUTING, INPUT, FORWARD, OUTPUT, POSTROUTING (단일 입력만 가능) |
| **action** | ACCEPT, DROP, IDS, IPS (단일 입력만 가능) |
| **Source_ip** | CIDR, 하이픈(-) 범위, 콤마(,) 형식 지원 |
| | 예: `10.10.10.10/24,192.168.31.30-192.168.31.35,7.7.7.7,8.8.8.8`와 같이 복합적인 입력도 지원 |
| **protocol** | TCP, UDP, ICMP, ANY … (단일 입력만 가능) |
| **Dest_ip** | CIDR, 하이픈(-) 범위, 콤마(,) 형식 지원 |
| | 예: `10.10.10.10/24,192.168.31.30-192.168.31.35,7.7.7.7,8.8.8.8`와 같이 복합적인 입력도 지원 |
| **Dest_port** | 하이픈(-) 범위, 콤마(,) 형식 지원 |
| | 예: `8080-8090,10100,20000-40000`와 같이 복합 형식도 지원 |
| **In_Interface** | 단일 입력만 가능 |
| **Out_Interface** | 단일 입력만 가능 |

---

## 4. 방화벽 운영 참고 사항

### 4.1 메모리 사용량

Smartfw는 Hash 기반의 방화벽으로 방화벽이 동작함에 따라 룰 저장 및 접속 IP에 대한 메모리 저장이 들어간다.

각 항목별 메모리 사용량에 대한 정보는 다음과 같다.

| 종류 | 단위 | 사이즈 | 비고 |
|------|------|--------|------|
| 일반 룰 저장 | 룰 1개당 | 88 byte | |
| 인터페이스 기반 룰 | 룰 1개당 | 168 byte | |
| Rate_limit | 접속 IP 1개당 | 416 byte | 스케줄러 60초. 120초 경과될 경우 정리 |
| Conn_limit | 접속 IP 1개당 | 56 byte | 스케줄러 60초. 커넥션 0이면서, 60초 경과 됐을 경우 |
| ACL (white/block LIST) | IP 1개당 | 56 byte | |
| NAT | 룰 1개당 | 148 byte | |
| Interface_card | 인터페이스 1개당 | 72 byte | |
| Detect_LOG | 중복 발생 로그 1개당 | 240 byte | 스케줄러 10분. 10분간 갱신 안됐다면 정리 |

#### 운영 환경 기준 (메모리 2GB 기준)

각 항목별로 단일 저장 했을 때의 최대 수량:

| 항목 | 2GB 기준 | 비고 |
|------|----------|------|
| 일반 룰 저장 | 2400만개 | |
| 인터페이스 기반 룰 | 1270만개 | 많아 질수록 룰 검색 성능 저하 발생 |
| NAT | 1440만개 | 많아 질수록 룰 검색 성능 저하 발생 |
| 동시 접속 IP 트래킹 | 450만개 | |
| ACL (White/block list) | 3800만개 | |

### 4.2 제약 사항

| 항목 | 제약 |
|------|------|
| 입력 룰 커맨드 | 입력 값 총합 255자 (값 기준) |
| 룰 IP 입력 | Source × Dest IP (단일 IP) 65537개 까지. Interface 카드에 정책 입력 시 이상으로 사용가능 |
| Port 입력 | 범위 포트 룰은 최대 10개까지 입력가능 (예: 1010-1015, 2020-2025, …10개까지). 단일 포트 지정은 별도 공간 |
| 인터페이스 카드 | 1개씩만 입력 가능 |

---

## 5. 방화벽 명령어 가이드

### 5.1 기본 명령어

#### 적용된 룰 확인

```bash
./agent -L
```

**용도**: 적용되어 있는 룰을 프린트(출력)해준다.

#### 룰 히스토리

```bash
./agent -history
```

**용도**: 최근 100건까지의 입력된 룰을 보여준다.

### 5.2 로그 레벨 변경

#### PKT_LOG

```bash
./agent -a=PKT_LOG_ON / OFF
```

**용도**: 패킷이 들어오고 나갈 때 로그 (기본값 OFF)

Protocol, Sip, dip, port 부분에 입력값이 있을 경우 해당 값만 필터링 가능

#### CONF_LOG

```bash
./agent -a=CONF_LOG_ON / OFF
```

**용도**: 방화벽 룰의 저장, 수정, 삭제에 대한 로그 (기본값 OFF)

#### DEBUG_LOG

```bash
./agent -a=DEBUG_LOG_ON / OFF
```

**용도**: 상세분석 용 (디버깅 로그) (기본값 OFF)

#### DETECT_LOG

**사용 여부 변경 불가**

**용도**: IDS, IPS, DROP에 대한 로그 기록

- 중복 디텍팅 로그는 한번 찍은 후 설정된 시간(300초)동안은 동일 로그에 대한 카운팅만 함
- 300초 후 또 발생할 경우 5분 동안 발생했던 카운팅과 함께 로그 기록

**예시**:

```
[ 3416.772796] [Smartfw-Detect] type=DROP-acl-rule&proto=TCP&direct=192.168.44.16 -> 192.168.44.13:9090
[ 3446.791160] [Smartfw-Detect] type=DROP-acl-rule&proto=TCP&direct=192.168.44.16 -> 192.168.44.13:9090&sup=1193 in 30s
(30초 이내 1193건의 중복로그 발생)
[ 3476.802584] [Smartfw-Detect] type=DROP-acl-rule&proto=TCP&direct=192.168.44.16 -> 192.168.44.13:9090&sup=1193 in 30s
(30초 이내 1193건의 중복로그 발생)
```

### 5.3 일반 룰

| 요소 | 입력 값 | 비고 |
|------|---------|------|
| Command | `-c` | Insert, delete |
| Chain | `-a` | PREROUTING, INPUT, FORWARD, OUTPUT, POSTROUTING (단일 입력만 가능) |
| action | `-a` | ACCEPT, DROP, IDS, IPS (단일 입력만 가능) |
| Source_ip | `-s` | CIDR, 하이픈(-) 범위, 콤마(,) 형식 지원. IP의 수량이 Source, dest 조합으로 65536가 초과될 경우 등록 불가 (예외: 인터페이스에 설정이 들어갈 경우 허용) |
| protocol | `-p` | TCP, UDP, ICMP, ANY … |
| Dest_ip | `--dest` | CIDR, 하이픈(-) 범위, 콤마(,) 형식 지원. IP의 수량이 Source, dest 조합으로 65536가 초과될 경우 등록 불가 (예외: 인터페이스에 설정이 들어갈 경우 허용) |
| Dest_port | `--dport` | 하이픈(-) 범위, 콤마(,) 형식 지원. 예: `8080-8090,10100,20000-40000` |
| In_Interface | `-i` | 단일 입력만 가능. `+`, `*`, `!` 정규식 적용가능 (eth+, eth*, !eth0, !eth*) |
| Out_Interface | `-o` | 단일 입력만 가능. `+`, `*`, `!` 정규식 적용가능 (eth+, eth*, !eth0, !eth*) |

#### 디폴트 Action 변경

```bash
./agent -a=DEFAULT_ACCEPT / DEFAULT_DROP
```

**용도**: 방화벽의 디폴트 룰을 ACCEPT 또는 DROP으로 변경 (기본값은 DROP)

> **주의**: 디폴트 변경 시 룰들은 초기화 됨.

### 5.4 NAT 룰

Action 필드에 NAT라는 명령문과 source_ip, dest_ip, dest_port, out_interface 입력 값을 사용하여 설정한다.

#### 5.4.1 SNAT

Iptables와 다르게, MASQUERADE 사용은 dest_ip가 입력되지 않을 경우 out_interface의 IP를 사용하기 때문에 명시적으로 입력하지 않아도 된다.

**Smartfw**:
```bash
./agent -f -c=insert -s=192.168.0.0/24 -p="ANY?SNAT" --dest=203.0.113.10 -a=NAT
```

**iptables**:
```bash
iptables -t nat -A POSTROUTING -s 192.168.0.0/24 -o eth0 -j SNAT --to-source 203.0.113.10
```

**Smartfw (MASQUERADE)**:
```bash
./agent -f -c=insert -s=192.168.0.0/24 -p="ANY?SNAT" -o=eth0 -a=NAT
```

**iptables (MASQUERADE)**:
```bash
iptables -t nat -A POSTROUTING -s 192.168.0.0/24 -o eth0 -j MASQUERADE
```

#### 5.4.2 DNAT

**Smartfw**:
```bash
./agent -f -m=insert -p="TCP?DNAT" --dest=192.168.0.10 --dport=80,8080 -a=NAT
```

**iptables**:
```bash
iptables -t nat -A PREROUTING -p tcp --dport 80 -j DNAT --to-destination 192.168.0.10:8080
```

### 5.5 IPS 룰

Action 필드에 원하는 action을 지정하고 protocol 부분에 `IPS?`로 시작하는 룰을 적용할 수 있다.

#### 형식

```
protocol에 IPS?ipstype&limit=카운팅제한&seconds=적용시간&enable=사용여부
```

입력된 값만 업데이트 하며, 기존에 입력된 값이 없을 경우 기본값을 사용한다.

IP, port에 지정이 가능한 IPS 룰들은 `-s` 또는 `--dport`를 사용하여 지정할 수 있다.

#### 명령어 예시 (IP Layer)

```bash
./agent -f -m=insert -p="IPS?land-attack&enable=0" -a=IPS
./agent -f -m=insert -p="IPS?ip-protocol&enable=1" -a=IDS
```

#### 명령어 예시 (IPS 프로토콜 필터)

```bash
./agent -f -m=insert -p="IPS?syn-flood&limit=50&seconds=2&enable=1" -a=IPS
./agent -f -m=insert -p="IPS?udp-flood&limit=50&seconds=2&enable=1" -a=IPS
```

### 5.6 ACL (화이트리스트 & 블록리스트)

#### 특정 IP 조회

```bash
./agent -f -m=insert -s=192.168.30.30 -a=searchacl
```

#### 추가 (Block)

```bash
./agent -f -m=insert -s=192.168.30.30 -a=blocklist
```

#### 추가 (White)

```bash
./agent -f -m=insert -s=192.168.30.30 -a=whitelist
```

### 5.7 적용된 룰 확인

룰은 최대 100개 (최신 입력) 까지만 출력 됨.

```bash
./agent -L
# 또는
./agent --list
```

> Conf_log 옵션이 켜져있을 경우 메모리 사용량도 출력 됨.

### 5.8 룰 적용 히스토리

최근 적용된 룰을 최신-과거 순으로 100개까지의 이력을 볼 수 있다.

```bash
./agent -H
# 또는
./agent --history
```

---

## 6. 부록

### 6.1 성능 테스트

기존 line 기반의 리눅스 기본 방화벽인 iptables(firewalld)와 HashMap 기반의 Smartfw 비교

#### 6.1.1 테스트 장비

| 사양 | 값 |
|------|-----|
| CPU | 2 Core |
| MEMORY | 2 GB |

#### 6.1.2 참고사항

두 방화벽 모두 동일한 룰셋을 적용하지만 Smartfw의 경우 룰 설정이 없더라도 목차에서 설명된 기본으로 작동하는 IPS 필터까지 적용된 환경

#### 6.1.3 컬럼 정의

| 컬럼 | 정의 |
|------|------|
| **bits_per_second** | iperf3가 보고한 전체 전송 속도(bps, bits/second) |
| **packets_per_second** | 초당 패킷 처리량(pps). TCP: bps / (packet_size × 8), UDP: packets / seconds |
| **retransmits** (TCP) | TCP 재전송 횟수 (네트워크 혼잡, 방화벽 처리 비용 등에 영향을 받음) |
| **latency_ms** | 클라이언트 → 방화벽 서버로 `ping -c 10 -q`를 수행한 결과의 평균 RTT(ms) |
| **loss_percent** (UDP) | UDP 전송 시 패킷 손실률(%) - TCP 테스트에서는 항상 0으로 기록됨 |
| **jitter_ms** (UDP) | UDP 지터(ms) - TCP 테스트에서는 0 |
| **cpu_usage_pct** | iperf 실행 전부터 실행 동안의 /proc/stat 비교로 계산한 평균 CPU 사용률(%) - 시간 평균 |
| **firewall_cpu_pct** | cpu_usage_pct에서 iperf3가 사용한 CPU 사용을 제외한 평균 CPU 사용률(%) |

#### 6.1.4 테스트 참고 사항

| 항목 | 값 |
|------|-----|
| Default RULE | ACCEPT |
| 적용 룰 | 각 프로토콜에 대해 단일 IP로 갯수만큼 설정 |
| 비고 | Smartfw의 PROC READ: 10ms 지연, CPU 3프로 사용 |
| iperf3 부하 지속시간 | 30초 |
| iperf3 패킷 사이즈 | 64 byte |
| iperf3 병렬(쓰레드) | 4개 |

#### 6.1.5 대량 룰 테스트 (TCP)

Layer 4에서 룰 수량이 늘어 날수록 Iptables과 같은 line 기반 방화벽은 CPU 부하가 심해지면서 성능저하가 발생함.

Iptables의 경우 ACL(White/Block) 추가도 1개의 룰로 카운팅 됨.

![TCP 테스트 결과](images/image1.png)

#### 6.1.6 대량 룰 테스트 (UDP)

TCP와 동일한 결과 값, UDP의 경우 패킷 발생이 더 많고, CPU 부하가 더 가기 때문에 Iptables의 경우 ksoftirqd/0 CPU 부하(코어가 풀)

![UDP 테스트 결과](images/image2.png)

#### 6.1.7 대량 룰 테스트 결과 요약

**TCP 테스트**:
- BPS/PPS의 경우 일정하면서 수치가 높을수록 좋음
- Latency의 경우 일정하면서 수치가 낮을수록 좋음

**UDP 테스트**:
- BPS/PPS의 경우 일정하면서 수치가 높을수록 좋음
- Latency의 경우 일정하면서 수치가 낮을수록 좋음
- CPU의 경우 일정하면서 수치가 낮을수록 좋음

![성능 비교 그래프](images/image3.png)
