// 방화벽 규칙 관련 상수 정의
// 각 컴포넌트에서 공통으로 사용되는 상수를 중앙에서 관리합니다.

// Chain 상수
export const CHAIN_INPUT = 0;
export const CHAIN_OUTPUT = 1;
export const CHAIN_FORWARD = 2;

// Protocol 상수
export const PROTOCOL_TCP = 0;
export const PROTOCOL_UDP = 1;
export const PROTOCOL_ICMP = 2;
export const PROTOCOL_ANY = 3;

// Action 상수
export const ACTION_DROP = 0;
export const ACTION_ACCEPT = 1;

// NAT Type 상수
export const NAT_TYPE_DNAT = 0;
export const NAT_TYPE_SNAT = 1;
export const NAT_TYPE_MASQUERADE = 2;

// Chain 이름 매핑
export const CHAIN_NAMES: Record<number, string> = {
    [CHAIN_INPUT]: 'INPUT',
    [CHAIN_OUTPUT]: 'OUTPUT',
    [CHAIN_FORWARD]: 'FORWARD',
};

// Protocol 이름 매핑
export const PROTOCOL_NAMES: Record<number, string> = {
    [PROTOCOL_TCP]: 'TCP',
    [PROTOCOL_UDP]: 'UDP',
    [PROTOCOL_ICMP]: 'ICMP',
    [PROTOCOL_ANY]: 'ANY',
};

// Action 이름 매핑
export const ACTION_NAMES: Record<number, string> = {
    [ACTION_DROP]: 'DROP',
    [ACTION_ACCEPT]: 'ACCEPT',
};

// NAT Type 이름 매핑
export const NAT_TYPE_NAMES: Record<number, string> = {
    [NAT_TYPE_DNAT]: 'DNAT',
    [NAT_TYPE_SNAT]: 'SNAT',
    [NAT_TYPE_MASQUERADE]: 'MASQUERADE',
};
