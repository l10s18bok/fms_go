// 헬프 텍스트 상수 정의
// 각 폼과 다이얼로그에서 사용되는 도움말 텍스트를 중앙에서 관리합니다.

// 애플리케이션 전체 도움말 (파일 메뉴 > 도움말)
export const APP_HELP = {
    title: 'FMS - Firewall Management System',
    version: '1.1.0',
    sections: [
        {
            name: '템플릿 관리',
            items: ['방화벽 규칙 템플릿을 생성/수정/삭제합니다'],
        },
        {
            name: '장비 관리',
            items: [
                '관리할 방화벽 장비(IP)를 등록합니다',
                '서버 상태를 확인하고 템플릿을 배포합니다',
            ],
        },
        {
            name: '배포 이력',
            items: [
                '배포 결과를 확인할 수 있습니다',
                '규칙별 성공/실패 상태를 상세히 확인합니다',
            ],
        },
        {
            name: 'Import/Export',
            items: ['현재 탭의 데이터를 JSON 파일로 내보내거나 가져옵니다'],
        },
    ],
    connectionModes: [
        {
            name: 'Agent Server',
            desc: 'Agent 서버(예: http://172.24.10.6:8080)를 통해 연결',
            endpoints: [
                '상태확인: POST /agent/req-respCheck',
                '배포: POST /agent/req-deploy',
            ],
        },
        {
            name: 'Direct',
            desc: '각 장비에 직접 HTTP 연결 (포트 80)',
            endpoints: [
                '상태확인: GET http://{장비IP}/respCheck',
                '배포: POST http://{장비IP}/deploy',
            ],
        },
    ],
    ruleFormat: {
        pattern: 'req|INSERT|{ID}|{CHAIN}|{ACTION}|{PROTOCOL}|{SRC}|{DST}|{옵션들}',
        examples: [
            'req|INSERT|3813792919|INPUT|FLUSH|ANY|ANY|ANY|||',
            'req|INSERT|3813792919|INPUT|ACCEPT|TCP|192.168.1.0/24|ANY|80||',
        ],
    },
};

// TCP Flags 도움말 (JSX용)
export const TCP_FLAGS_HELP = {
    title: 'TCP Flags 옵션 설명',
    presets: [
        { name: 'None', desc: '모든 TCP 패킷을 매칭합니다.' },
        { name: 'New Connection (SYN)', desc: '새로운 연결 요청만 매칭합니다. SYN 플래그만 설정된 패킷을 탐지합니다.' },
        { name: 'Established (ACK)', desc: '이미 연결된 세션의 패킷을 매칭합니다. ACK 플래그가 설정된 패킷을 탐지합니다.' },
        { name: 'NULL Scan Block', desc: '모든 플래그가 해제된 비정상 패킷을 탐지합니다. 포트 스캔 공격 방어에 사용됩니다.' },
        { name: 'XMAS Scan Block', desc: 'FIN, PSH, URG가 동시에 설정된 비정상 패킷입니다. 포트 스캔 공격 방어에 사용됩니다.' },
        { name: 'SYN+FIN Block', desc: 'SYN과 FIN이 동시에 설정된 비정상 패킷입니다. 정상적인 TCP에서는 발생하지 않습니다.' },
        { name: 'Custom', desc: '체크박스로 직접 플래그를 설정합니다.' },
    ],
    flags: [
        { name: 'SYN', desc: '연결 시작을 요청합니다.' },
        { name: 'ACK', desc: '데이터 수신을 확인합니다.' },
        { name: 'FIN', desc: '연결 종료를 요청합니다.' },
        { name: 'RST', desc: '연결을 강제로 종료합니다.' },
        { name: 'PSH', desc: '데이터를 즉시 전달하도록 요청합니다.' },
        { name: 'URG', desc: '긴급 데이터임을 표시합니다.' },
    ],
    maskSetDesc: {
        mask: '검사할 플래그를 선택합니다.',
        set: '실제로 설정되어야 할 플래그를 선택합니다.',
        example: 'Mask=SYN,ACK / Set=SYN → SYN,ACK 중 SYN만 설정된 패킷 (새 연결)',
    },
};

// ICMP 도움말 (JSX용)
export const ICMP_HELP = {
    title: 'ICMP Type 옵션 설명',
    types: [
        { name: 'None', desc: '모든 ICMP 패킷을 매칭합니다.' },
        { name: 'echo-reply', desc: 'ping 응답 패킷입니다. echo-request에 대한 응답입니다.' },
        { name: 'destination-unreachable', desc: '목적지에 도달할 수 없음을 알리는 패킷입니다.' },
        { name: 'source-quench', desc: '송신 속도를 줄이라는 요청입니다. (현재는 거의 사용되지 않음)' },
        { name: 'echo-redirect', desc: '더 좋은 라우팅 경로가 있음을 알립니다. 보안상 차단하는 경우가 많습니다.' },
        { name: 'echo-request', desc: 'ping 요청 패킷입니다. 이 타입을 차단하면 외부에서 ping이 안 됩니다.' },
        { name: 'time-exceeded', desc: 'TTL이 0이 되어 패킷이 폐기됨을 알립니다. traceroute에서 사용됩니다.' },
        { name: 'parameter-problem', desc: 'IP 헤더에 문제가 있음을 알립니다.' },
        { name: 'timestamp-request', desc: '타임스탬프 요청 패킷입니다.' },
        { name: 'timestamp-reply', desc: '타임스탬프 응답 패킷입니다.' },
        { name: 'information-request', desc: '네트워크 정보 요청입니다. (거의 사용되지 않음)' },
        { name: 'information-reply', desc: '네트워크 정보 응답입니다. (거의 사용되지 않음)' },
        { name: 'addressmask-request', desc: '서브넷 마스크 요청입니다.' },
        { name: 'addressmask-reply', desc: '서브넷 마스크 응답입니다.' },
    ],
};

// DNAT 도움말 (JSX용)
export const DNAT_HELP = {
    title: '포트 포워딩 (DNAT) 도움말',
    description: '외부에서 들어오는 트래픽을 내부 서버로 전달합니다.\n예) 공인 IP:8080 → 내부 서버 192.168.1.10:80',
    fields: [
        { name: 'ExtPort (외부 포트)', desc: '외부에서 접속할 포트 번호입니다. 필수 입력 항목입니다.', example: '8080, 443, 22' },
        { name: 'SIP (소스 IP)', desc: '접속을 허용할 출발지 IP입니다. 비워두면 모든 IP에서 접속을 허용합니다.', example: '192.168.1.0/24, 10.0.0.5' },
        { name: 'DIP (목적지 IP)', desc: '트래픽을 전달할 내부 서버 IP입니다. 필수 입력 항목입니다.', example: '192.168.1.10, 10.0.0.100' },
        { name: 'DPort (목적지 포트)', desc: '내부 서버의 실제 서비스 포트입니다. 비워두면 ExtPort와 동일한 포트를 사용합니다.', example: '80, 443, 3389' },
    ],
    examples: [
        { title: '웹 서버 포트 포워딩', config: 'ExtPort=8080, DIP=192.168.1.10, DPort=80', result: '외부:8080 접속 시 내부 192.168.1.10:80으로 전달' },
        { title: 'SSH 포트 포워딩 (특정 IP만 허용)', config: 'ExtPort=2222, SIP=10.0.0.0/8, DIP=192.168.1.5, DPort=22', result: '10.x.x.x 대역에서만 SSH 접속 허용' },
    ],
};

// SNAT 도움말 (JSX용)
export const SNAT_HELP = {
    title: '소스 NAT (SNAT/MASQUERADE) 도움말',
    description: '내부 네트워크에서 외부로 나가는 트래픽의 소스 IP를 변환합니다.\n예) 내부 192.168.1.x → 공인 IP로 변환하여 인터넷 접속',
    natTypes: [
        { name: 'SNAT', desc: '고정 IP 환경에서 사용합니다. 변환할 IP를 직접 지정합니다.', example: '내부 → 고정 공인 IP' },
        { name: 'MASQUERADE', desc: '유동 IP 환경(PPPoE, DHCP)에서 사용합니다. 출력 인터페이스의 IP를 자동으로 사용합니다.', example: '내부 → 인터넷 공유기의 현재 IP' },
    ],
    fields: [
        { name: 'SIP (소스 IP/네트워크)', desc: 'NAT를 적용할 내부 네트워크입니다. 필수 입력 항목입니다.', example: '192.168.1.0/24, 10.0.0.0/8' },
        { name: 'InIF (입력 인터페이스)', desc: '트래픽이 들어오는 내부 인터페이스입니다. 선택 사항입니다.', example: 'eth1, br0, lan0' },
        { name: 'OutIF (출력 인터페이스)', desc: '트래픽이 나가는 외부 인터페이스입니다. 선택 사항입니다.', example: 'eth0, ppp0, wan0' },
        { name: 'TransIP (변환 IP) - SNAT만', desc: '소스 IP를 변환할 공인 IP입니다. SNAT 선택 시 필수입니다.', example: '203.0.113.1, 1.2.3.4' },
    ],
    examples: [
        { title: '가정용 공유기 (MASQUERADE)', config: 'Type=MASQUERADE, SIP=192.168.1.0/24, OutIF=eth0', result: '내부 192.168.1.x가 eth0의 IP로 변환되어 외부 통신' },
        { title: '기업 네트워크 (SNAT)', config: 'Type=SNAT, SIP=10.0.0.0/8, TransIP=203.0.113.1', result: '내부 10.x.x.x가 고정 IP 203.0.113.1로 변환' },
    ],
};

// Black/White 도움말 (JSX용)
export const BLACK_WHITE_HELP = {
    title: 'Black/White 도움말',
    types: [
        { name: 'Black', desc: '블랙리스트 - 해당 IP에서 오는 모든 트래픽을 차단합니다.' },
        { name: 'White', desc: '화이트리스트 - 해당 IP에서 오는 모든 트래픽을 허용합니다.' },
    ],
    fields: [
        { name: 'IP', desc: '차단/허용할 IP 주소 또는 네트워크', example: '192.168.1.100, 10.0.0.0/8' },
    ],
};
