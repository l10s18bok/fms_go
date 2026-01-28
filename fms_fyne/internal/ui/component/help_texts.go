package component

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ShowHelpPopup 공통 헬프 팝업 표시 함수
// title: 팝업 제목, content: 도움말 내용, parent: 부모 CanvasObject
func ShowHelpPopup(title string, content string, parent fyne.CanvasObject) {
	// 헬프 내용 생성
	helpLabel := widget.NewLabel(content)

	// 스크롤 가능한 컨테이너
	scroll := container.NewScroll(helpLabel)
	scroll.SetMinSize(fyne.NewSize(400, 400))

	// 닫기 버튼
	var popup *widget.PopUp
	closeBtn := widget.NewButton("닫기", func() {
		if popup != nil {
			popup.Hide()
		}
	})

	// 팝업 내용 (제목 + 내용 + 닫기 버튼)
	titleLabel := widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	popupContent := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		scroll,
		widget.NewSeparator(),
		container.NewCenter(closeBtn),
	)

	// 모달 팝업 (화면 중앙에 표시)
	canvas := fyne.CurrentApp().Driver().CanvasForObject(parent)
	popup = widget.NewModalPopUp(container.NewPadded(popupContent), canvas)
	popup.Show()
}

// 헬프 텍스트 상수 정의
// 각 폼과 다이얼로그에서 사용되는 도움말 텍스트를 중앙에서 관리합니다.

// AppHelpText 애플리케이션 전체 도움말
const AppHelpText = `FMS - Firewall Management System

[규칙 파일 관리]
• 방화벽 규칙 파일을 생성/수정/삭제합니다

[장비 관리]
• 관리할 방화벽 장비(IP)를 추가/수정/삭제합니다
• 장비 상태 확인: 수동(새로고침) 또는 자동(토글)
• 장비 더블클릭: 상세 정보 및 프로세스 목록 확인
• 배포 기능:
  - 방화벽 룰 배포: 규칙 파일을 선택하여 배포
  - 패키지 업데이트: 등록된 패키지를 장비에 배포

[배포 이력]
• 배포 결과를 확인할 수 있습니다
• 규칙별 성공/실패 상태를 확인합니다

[Import/Export]
• 현재 탭의 데이터를 JSON 파일로 내보내거나 가져옵니다`

// TCPFlagsHelpText TCP Flags 옵션 도움말
const TCPFlagsHelpText = `TCP Flags 옵션 설명:

[프리셋]

• None
  - 모든 TCP 패킷을 매칭합니다.

• New Connection (SYN)
  - 새로운 연결 요청만 매칭합니다.
  - SYN 플래그만 설정된 패킷을 탐지합니다.

• Established (ACK)
  - 이미 연결된 세션의 패킷을 매칭합니다.
  - ACK 플래그가 설정된 패킷을 탐지합니다.

• NULL Scan Block
  - 모든 플래그가 해제된 비정상 패킷을 탐지합니다.
  - 포트 스캔 공격 방어에 사용됩니다.

• XMAS Scan Block
  - FIN, PSH, URG가 동시에 설정된 비정상 패킷입니다.
  - 포트 스캔 공격 방어에 사용됩니다.

• SYN+FIN Block
  - SYN과 FIN이 동시에 설정된 비정상 패킷입니다.
  - 정상적인 TCP에서는 발생하지 않습니다.

• Custom
  - 체크박스로 직접 플래그를 설정합니다.

[플래그 설명]

• SYN (Synchronize)
  - 연결 시작을 요청합니다.

• ACK (Acknowledge)
  - 데이터 수신을 확인합니다.

• FIN (Finish)
  - 연결 종료를 요청합니다.

• RST (Reset)
  - 연결을 강제로 종료합니다.

• PSH (Push)
  - 데이터를 즉시 전달하도록 요청합니다.

• URG (Urgent)
  - 긴급 데이터임을 표시합니다.

[Mask / Set 설명]

• Mask: 검사할 플래그를 선택합니다.
• Set: 실제로 설정되어야 할 플래그를 선택합니다.

예) Mask=SYN,ACK / Set=SYN
  → SYN,ACK 중 SYN만 설정된 패킷 (새 연결)`

// ICMPOptionsHelpText ICMP 옵션 도움말
const ICMPOptionsHelpText = `ICMP Type 옵션 설명:

• None
  - 모든 ICMP 패킷을 매칭합니다.

• echo-reply
  - ping 응답 패킷입니다.
  - echo-request에 대한 응답입니다.

• destination-unreachable
  - 목적지에 도달할 수 없음을 알리는 패킷입니다.
  - 네트워크, 호스트, 포트 등이 도달 불가할 때 발생합니다.

• source-quench
  - 송신 속도를 줄이라는 요청입니다.
  - 네트워크 혼잡 제어용으로 현재는 거의 사용되지 않습니다.

• echo-redirect
  - 더 좋은 라우팅 경로가 있음을 알립니다.
  - 보안상 차단하는 경우가 많습니다.

• echo-request
  - ping 요청 패킷입니다.
  - 상대 호스트가 살아있는지 확인할 때 사용됩니다.
  - 이 타입을 차단하면 외부에서 ping이 안 됩니다.

• time-exceeded
  - TTL이 0이 되어 패킷이 폐기됨을 알립니다.
  - traceroute 명령에서 경로 추적에 사용됩니다.

• parameter-problem
  - IP 헤더에 문제가 있음을 알립니다.

• timestamp-request
  - 타임스탬프 요청 패킷입니다.
  - 네트워크 지연 측정에 사용됩니다.

• timestamp-reply
  - 타임스탬프 응답 패킷입니다.

• information-request
  - 네트워크 정보 요청입니다. (거의 사용되지 않음)

• information-reply
  - 네트워크 정보 응답입니다. (거의 사용되지 않음)

• addressmask-request
  - 서브넷 마스크 요청입니다.

• addressmask-reply
  - 서브넷 마스크 응답입니다.`

// IPSOptionsHelpText IPS 옵션 도움말
const IPSOptionsHelpText = `IPS (Intrusion Prevention System) 옵션 설명:

Smartfw의 IPS 기능은 다양한 네트워크 공격을 탐지하고 차단합니다.

[IP Layer 필터]

• land-attack - Source IP = Dest IP 패킷 차단
• ip-spoofing - IP 스푸핑 공격 차단
• ip-tunnel - GRE/IPIP 터널링 차단
• ip-fragment - IP 조각화 공격 차단 (Teardrop 포함)
• ttl-attack - 비정상 TTL(≤1) 패킷 차단
• port-scan - 포트 스캔 탐지/차단
• ip-protocol - 비정상 프로토콜 차단
• ip-options - IP Options 필드 패킷 차단
• ip-fragment-tiny - Tiny Fragment 공격 차단
• MCAST-DST-PING - 멀티캐스트 ICMP 차단

[TCP 필터]

• syn-flood - SYN 패킷 대량 전송 공격
• concurrent-conn - IP당 동시 세션 제한
• synack-flood - SYN-ACK 플러드 공격
• ack-flood - ACK 플러드 공격
• rst-flood - RST 플러드 공격
• fin-flood - FIN 플러드 공격
• pshack-flood - PSH-ACK 플러드 공격
• tcp-total - TCP 종합 플러드 공격

[UDP 필터]

• udp-flood - UDP 패킷 대량 전송 공격
• udp-bytes - UDP 대용량 페이로드 공격

[ICMP 필터]

• icmp-flood - ICMP 패킷 대량 전송 공격
• icmp-bytes - ICMP 대용량 페이로드 공격

[파라미터 설명]

• limit: 허용할 최대 패킷/바이트 수
• seconds: 측정 시간 간격 (초)
• enable: 활성화 여부 (1=활성, 0=비활성)

[예시]

SYN Flood: limit=50&seconds=1&enable=1
→ 초당 50개 이상 SYN 패킷 차단

Port Scan: limit=32&seconds=1&enable=1
→ 초당 32개 이상 포트 스캔 차단`

// GeneralRuleHelpText 일반 규칙 도움말
const GeneralRuleHelpText = `일반 규칙 도움말:

방화벽 규칙을 세부적으로 설정합니다.
(Smartfw-manual.docx 기준)

[입력 필드 설명]

• Chain (체인) - 옵션: -c
  - INPUT: 서버로 들어오는 트래픽
  - OUTPUT: 서버에서 나가는 트래픽
  - FORWARD: 서버를 경유하는 트래픽
  - PREROUTING: 라우팅 전 (NAT용)
  - POSTROUTING: 라우팅 후 (NAT용)

• Action (동작) - 옵션: -a
  - ACCEPT: 트래픽 허용
  - DROP: 트래픽 차단 (응답 없음)
  - IDS: 허용 + 로그 기록 (Intrusion Detection)
  - IPS: 차단 + 로그 기록 (Intrusion Prevention)

• Protocol (프로토콜) - 옵션: -p
  - TCP: 웹, SSH, FTP 등 연결 지향
  - UDP: DNS, NTP 등 비연결 지향
  - ICMP: ping, traceroute 등
  - ANY: 모든 프로토콜

• SIP (소스 IP) - 옵션: -s
  - 출발지 IP 또는 네트워크 지정
  - 예: 192.168.1.100, 10.0.0.0/8
  - CIDR, 범위(-), 콤마(,) 형식 지원
  - 비워두면 ANY (모든 IP)

• DIP (목적지 IP) - 옵션: --dest
  - 목적지 IP 또는 네트워크 지정
  - 예: 192.168.1.100, 10.0.0.0/8
  - CIDR, 범위(-), 콤마(,) 형식 지원
  - 비워두면 ANY (모든 IP)

• Port (목적지 포트) - 옵션: --dport
  - 예: 80, 443, 80-90, 80,443,8080
  - 범위(-), 콤마(,) 형식 지원

• InIF / OutIF (인터페이스) - 옵션: -i, -o
  - 트래픽이 들어오거나 나가는 네트워크 인터페이스
  - 예: eth0, eth1, wan0
  - 정규식 지원: eth+ (eth로 시작하는 모든 인터페이스)
  - 비워두면 모든 인터페이스에 적용

• TCP Flags (TCP 전용)
  - 특정 TCP 플래그 조합을 필터링
  - SYN, ACK, FIN, RST, PSH, URG

• ICMP Type (ICMP 전용)
  - echo-request (ping), echo-reply 등`

// BlackWhiteHelpText Black/White 규칙 도움말
const BlackWhiteHelpText = `Black/White 리스트 도움말:

간단하게 IP를 차단하거나 허용합니다.

[타입 선택]

• Black (차단)
  - 지정한 IP에서 오는 모든 트래픽을 차단합니다.
  - Action: DROP
  - Chain: INPUT

• White (허용)
  - 지정한 IP에서 오는 모든 트래픽을 허용합니다.
  - Action: ACCEPT
  - Chain: INPUT

[입력 필드]

• IP
  - 차단/허용할 IP 주소입니다.
  - 필수 입력 항목입니다.
  - 예: 192.168.1.100, 10.0.0.0/8

[사용 예시]

악성 IP 차단:
  Type=Black, IP=203.0.113.50
  → 해당 IP에서 오는 모든 접속 차단

신뢰 IP 허용:
  Type=White, IP=192.168.1.0/24
  → 내부 네트워크에서 오는 접속 허용`

// DNATHelpText DNAT (포트 포워딩) 도움말
const DNATHelpText = `포트 포워딩 (DNAT) 도움말:

외부에서 들어오는 트래픽을 내부 서버로 전달합니다.
예) 공인 IP:8080 → 내부 서버 192.168.1.10:80

[입력 필드 설명]

• Proto (프로토콜)
  - 포트 포워딩을 적용할 프로토콜입니다.
  - tcp, udp 중 선택합니다.

• MatchPort (매칭 포트)
  - 외부에서 접속할 포트 번호입니다.
  - 필수 입력 항목입니다.
  - 예: 8080, 443, 22

• MatchIP (매칭 IP)
  - 접속을 허용할 출발지 IP입니다.
  - 비워두면 모든 IP에서 접속을 허용합니다.
  - 예: 192.168.1.0/24, 10.0.0.5

• TransIP (변환 IP)
  - 트래픽을 전달할 내부 서버 IP입니다.
  - 필수 입력 항목입니다.
  - 예: 192.168.1.10, 10.0.0.100

• TransPort (변환 포트)
  - 내부 서버의 실제 서비스 포트입니다.
  - 비워두면 MatchPort와 동일한 포트를 사용합니다.
  - 예: 80, 443, 3389

• InIF (입력 인터페이스)
  - 트래픽이 들어오는 인터페이스입니다.
  - 선택 사항입니다.
  - 예: eth0, wan0

[사용 예시]

웹 서버 포트 포워딩:
  MatchPort=8080, TransIP=192.168.1.10, TransPort=80
  → 외부:8080 접속 시 내부 192.168.1.10:80으로 전달

SSH 포트 포워딩 (특정 IP만 허용):
  MatchPort=2222, MatchIP=10.0.0.0/8, TransIP=192.168.1.5, TransPort=22
  → 10.x.x.x 대역에서만 SSH 접속 허용`

// SNATHelpText SNAT 도움말
const SNATHelpText = `소스 NAT (SNAT) 도움말:

내부 네트워크에서 외부로 나가는 트래픽의 소스 IP를 변환합니다.
예) 내부 192.168.1.x → 공인 IP로 변환하여 인터넷 접속

[입력 필드 설명]

• Proto (프로토콜)
  - NAT를 적용할 프로토콜입니다.
  - tcp, udp 중 선택합니다.

• MatchIP (매칭 IP/네트워크)
  - NAT를 적용할 내부 네트워크입니다.
  - 예: 192.168.1.0/24, 10.0.0.0/8

• OutIF (출력 인터페이스)
  - 트래픽이 나가는 외부 인터페이스입니다.
  - 선택 사항입니다.
  - 예: eth0, ppp0, wan0

• TransIP (변환 IP)
  - 소스 IP를 변환할 공인 IP입니다.
  - 예: 203.0.113.1, 1.2.3.4

[사용 예시]

기업 네트워크 (SNAT):
  MatchIP=10.0.0.0/8, TransIP=203.0.113.1
  → 내부 10.x.x.x가 고정 IP 203.0.113.1로 변환

인터페이스 지정:
  MatchIP=192.168.1.0/24, TransIP=1.2.3.4, OutIF=eth0
  → 192.168.1.x가 eth0으로 나갈 때 1.2.3.4로 변환`

// FirewallEditExampleText 방화벽 규칙 편집 예시 텍스트 (호환용)
const FirewallEditExampleText = `./agent -f -I -c INPUT -a ACCEPT -p TCP -n 80
./agent -f -I -c INPUT -a ACCEPT -p TCP -n 22 -s 192.168.1.0/24
./agent -f -I -c OUTPUT -a ACCEPT -p TCP -e 10.0.0.1 -n 443
./agent -f -I -c INPUT -a DROP -p TCP -n 23
./agent -f -D -c INPUT -a ACCEPT -p TCP -n 80
./agent -f -I -a NAT -p TCP?DNAT -n 80 -e 192.168.1.10:80 -i eth0
./agent -f -I -a NAT -p TCP?SNAT -s 192.168.1.0/24 -e 1.2.3.4 -o eth0
./agent -f -I -a NAT -p TCP?MASQUERADE -s 192.168.1.0/24 -o eth0`

// FirewallExampleRule 방화벽 예시 규칙 구조체
type FirewallExampleRule struct {
	Description string // 규칙 설명 (주석)
	Command     string // 규칙 명령어
}

// FirewallExampleRules 방화벽 예시 규칙 목록
var FirewallExampleRules = []FirewallExampleRule{
	{
		Description: "HTTP 포트 허용",
		Command:     "./agent -f -I -c INPUT -a ACCEPT -p TCP -n 80",
	},
	{
		Description: "특정 소스 IP에서 SSH 허용",
		Command:     "./agent -f -I -c INPUT -a ACCEPT -p TCP -n 22 -s 192.168.1.0/24",
	},
	{
		Description: "특정 목적지로 나가는 트래픽 허용",
		Command:     "./agent -f -I -c OUTPUT -a ACCEPT -p TCP -e 10.0.0.1 -n 443",
	},
	{
		Description: "Telnet 포트 차단",
		Command:     "./agent -f -I -c INPUT -a DROP -p TCP -n 23",
	},
	{
		Description: "룰 삭제 (-D 사용)",
		Command:     "./agent -f -D -c INPUT -a ACCEPT -p TCP -n 80",
	},
	{
		Description: "DNAT (포트 포워딩: 외부 80 → 내부 서버)",
		Command:     "./agent -f -I -a NAT -p TCP?DNAT -n 80 -e 192.168.1.10:80 -i eth0",
	},
	{
		Description: "SNAT (내부 → 공인 IP 변환)",
		Command:     "./agent -f -I -a NAT -p TCP?SNAT -s 192.168.1.0/24 -e 1.2.3.4 -o eth0",
	},
	{
		Description: "MASQUERADE (동적 IP용 SNAT)",
		Command:     "./agent -f -I -a NAT -p TCP?MASQUERADE -s 192.168.1.0/24 -o eth0",
	},
}

// FirewallFileHelpText 방화벽 룰 관리 도움말
const FirewallFileHelpText = `
[파일 추가]
• 드래그 앤 드롭으로 파일을 테이블에 끌어다 놓으면
  자동으로 추가됩니다.
• 또한, "추가/수정" 버튼을 클릭하여 추가할 수도 
  있습니다.

[파일 편집]
• 테이블 행(Row)을 더블클릭하면 방화벽 상세보기에서
  규칙을 작성하고 수정할 수 있습니다.
• 상세보기 창일때 앱 왼쪽 하단에 파일 경로가 
  표시됩니다.  

[파일명 변경]
• 파일이름 셀을 클릭하면 직접 파일명을 수정할
  수 있습니다.(Enter 저장, ESC 취소)
• 변경된 파일명은 즉시 저장됩니다.

[날짜 정보]
• 만든 날짜: 파일이 처음 추가된 날짜입니다.
• 수정한 날짜: 파일 내용이 변경되면 자동으로
  업데이트됩니다.

[버전 정보]
• 파일명에서 첫번째 "-" 을 기준으로 버전 정보를 
  자동으로 추출합니다.  
• 예: rules-v1_0_1.txt → v1.0.1`
