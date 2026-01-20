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

버전: 1.1.0

[템플릿 관리]
• 방화벽 규칙 템플릿을 생성/수정/삭제합니다

[장비 관리]
• 관리할 방화벽 장비(IP)를 등록합니다
• 서버 상태를 확인하고 템플릿을 배포합니다

[배포 이력]
• 배포 결과를 확인할 수 있습니다
• 규칙별 성공/실패 상태를 상세히 확인합니다

[Import/Export]
• 현재 탭의 데이터를 JSON 파일로 내보내거나 가져옵니다

[연결 모드] (설정에서 변경)
• Agent Server: Agent 서버(예: http://172.24.10.6:8080)를 통해 연결
  - 상태확인: POST /agent/req-respCheck
  - 배포: POST /agent/req-deploy
• Direct: 각 장비에 직접 HTTP 연결 (포트 80)
  - 상태확인: GET http://{장비IP}/respCheck
  - 배포: POST http://{장비IP}/deploy

[규칙 포맷]
req|INSERT|{ID}|{CHAIN}|{ACTION}|{PROTOCOL}|{SRC}|{DST}|{옵션들}

예시:
req|INSERT|3813792919|INPUT|FLUSH|ANY|ANY|ANY|||
req|INSERT|3813792919|INPUT|ACCEPT|TCP|192.168.1.0/24|ANY|80||`

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

// GeneralRuleHelpText 일반 규칙 도움말
const GeneralRuleHelpText = `일반 규칙 도움말:

방화벽 규칙을 세부적으로 설정합니다.

[입력 필드 설명]

• Chain (체인)
  - INPUT: 서버로 들어오는 트래픽
  - OUTPUT: 서버에서 나가는 트래픽
  - FORWARD: 서버를 경유하는 트래픽

• Action (동작)
  - ACCEPT: 트래픽 허용
  - DROP: 트래픽 차단 (응답 없음)
  - REJECT: 트래픽 차단 (거부 응답 전송)
  - RETURN: 현재 체인 종료
  - FLUSH: 체인의 모든 규칙 삭제
  - LOG: 로그 기록

• Protocol (프로토콜)
  - TCP: 웹, SSH, FTP 등 연결 지향
  - UDP: DNS, NTP 등 비연결 지향
  - ICMP: ping, traceroute 등
  - ANY: 모든 프로토콜

• SIP / DIP (소스/목적지 IP)
  - 특정 IP 또는 네트워크 지정
  - 예: 192.168.1.100, 10.0.0.0/8
  - 비워두면 ANY (모든 IP)

• 포트 옵션 (TCP/UDP)
  - SPort: 소스 포트
  - DPort: 목적지 포트
  - 예: 80, 443, 22, 1024:65535

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

• Match_Port (매칭 포트)
  - 외부에서 접속할 포트 번호입니다.
  - 필수 입력 항목입니다.
  - 예: 8080, 443, 22

• Match_IP (매칭 IP)
  - 접속을 허용할 출발지 IP입니다.
  - 비워두면 모든 IP에서 접속을 허용합니다.
  - 예: 192.168.1.0/24, 10.0.0.5

• Trans_IP (변환 IP)
  - 트래픽을 전달할 내부 서버 IP입니다.
  - 필수 입력 항목입니다.
  - 예: 192.168.1.10, 10.0.0.100

• D_Port (목적지 포트)
  - 내부 서버의 실제 서비스 포트입니다.
  - 비워두면 Match_Port와 동일한 포트를 사용합니다.
  - 예: 80, 443, 3389

• In_IFace (입력 인터페이스)
  - 트래픽이 들어오는 인터페이스입니다.
  - 선택 사항입니다.
  - 예: eth0, wan0

[사용 예시]

웹 서버 포트 포워딩:
  Match_Port=8080, Trans_IP=192.168.1.10, D_Port=80
  → 외부:8080 접속 시 내부 192.168.1.10:80으로 전달

SSH 포트 포워딩 (특정 IP만 허용):
  Match_Port=2222, Match_IP=10.0.0.0/8, Trans_IP=192.168.1.5, D_Port=22
  → 10.x.x.x 대역에서만 SSH 접속 허용`

// SNATHelpText SNAT 도움말
const SNATHelpText = `소스 NAT (SNAT) 도움말:

내부 네트워크에서 외부로 나가는 트래픽의 소스 IP를 변환합니다.
예) 내부 192.168.1.x → 공인 IP로 변환하여 인터넷 접속

[입력 필드 설명]

• SIP (소스 IP/네트워크)
  - NAT를 적용할 내부 네트워크입니다.
  - 필수 입력 항목입니다.
  - 예: 192.168.1.0/24, 10.0.0.0/8

• TransIP (변환 IP)
  - 소스 IP를 변환할 공인 IP입니다.
  - 예: 203.0.113.1, 1.2.3.4

• InIF (입력 인터페이스)
  - 트래픽이 들어오는 내부 인터페이스입니다.
  - 선택 사항입니다.
  - 예: eth1, br0, lan0

• OutIF (출력 인터페이스)
  - 트래픽이 나가는 외부 인터페이스입니다.
  - 선택 사항입니다.
  - 예: eth0, ppp0, wan0

[사용 예시]

기업 네트워크 (SNAT):
  SIP=10.0.0.0/8, TransIP=203.0.113.1
  → 내부 10.x.x.x가 고정 IP 203.0.113.1로 변환

인터페이스 지정:
  SIP=192.168.1.0/24, TransIP=1.2.3.4, InIF=eth1, OutIF=eth0
  → eth1에서 들어온 192.168.1.x가 eth0으로 나갈 때 1.2.3.4로 변환`
