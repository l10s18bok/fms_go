import { model } from '../../wailsjs/go/models';

// Chain 상수
const CHAIN_INPUT = 0;
const CHAIN_OUTPUT = 1;
const CHAIN_FORWARD = 2;

// Protocol 상수
const PROTOCOL_TCP = 0;
const PROTOCOL_UDP = 1;
const PROTOCOL_ICMP = 2;
const PROTOCOL_ANY = 3;

// Action 상수
const ACTION_DROP = 0;
const ACTION_ACCEPT = 1;

// Chain 값을 문자열로 변환
const chainToString = (chain: number): string => {
    switch (chain) {
        case CHAIN_INPUT: return 'INPUT';
        case CHAIN_OUTPUT: return 'OUTPUT';
        case CHAIN_FORWARD: return 'FORWARD';
        default: return 'INPUT';
    }
};

// Protocol 값을 문자열로 변환
const protocolToString = (protocol: number): string => {
    switch (protocol) {
        case PROTOCOL_TCP: return 'tcp';
        case PROTOCOL_UDP: return 'udp';
        case PROTOCOL_ICMP: return 'icmp';
        case PROTOCOL_ANY: return 'any';
        default: return 'tcp';
    }
};

// Action 값을 문자열로 변환
const actionToString = (action: number): string => {
    switch (action) {
        case ACTION_DROP: return 'DROP';
        case ACTION_ACCEPT: return 'ACCEPT';
        default: return 'DROP';
    }
};

// 프로토콜 옵션 문자열 생성
const formatOptions = (rule: model.FirewallRule): string => {
    if (!rule.options) return '';

    const parts: string[] = [];

    if (rule.options.tcpFlags) {
        parts.push(`flags=${rule.options.tcpFlags}`);
    }
    if (rule.options.icmpType) {
        parts.push(`type=${rule.options.icmpType}`);
    }
    if (rule.options.icmpCode) {
        parts.push(`code=${rule.options.icmpCode}`);
    }

    return parts.join(', ');
};

interface RuleTableProps {
    rules: model.FirewallRule[];
    onDelete: (index: number) => void;
    onEdit?: (index: number, rule: model.FirewallRule) => void;
}

const RuleTable = ({ rules, onDelete, onEdit }: RuleTableProps) => {
    return (
        <div className="rule-table-container">
            <table className="table rule-table">
                <thead>
                    <tr>
                        <th style={{ width: '50px' }}>삭제</th>
                        <th style={{ width: '80px' }}>Chain</th>
                        <th style={{ width: '80px' }}>Protocol</th>
                        <th style={{ width: '150px' }}>옵션</th>
                        <th style={{ width: '80px' }}>Action</th>
                        <th style={{ width: '80px' }}>DPort</th>
                        <th style={{ width: '120px' }}>SIP</th>
                        <th style={{ width: '120px' }}>DIP</th>
                        <th style={{ width: '60px' }}>Black</th>
                        <th style={{ width: '60px' }}>White</th>
                    </tr>
                </thead>
                <tbody>
                    {rules.length === 0 ? (
                        <tr>
                            <td colSpan={10} style={{ textAlign: 'center', color: '#666', padding: '20px' }}>
                                규칙이 없습니다. 아래에서 규칙을 추가하세요.
                            </td>
                        </tr>
                    ) : (
                        rules.map((rule, index) => (
                            <tr
                                key={index}
                                onClick={() => onEdit?.(index, rule)}
                                style={{ cursor: onEdit ? 'pointer' : 'default' }}
                            >
                                <td>
                                    <button
                                        className="btn btn-danger btn-sm"
                                        onClick={(e) => {
                                            e.stopPropagation();
                                            onDelete(index);
                                        }}
                                        title="삭제"
                                    >
                                        X
                                    </button>
                                </td>
                                <td>{chainToString(rule.chain)}</td>
                                <td>{protocolToString(rule.protocol)}</td>
                                <td className="options-cell" title={formatOptions(rule)}>
                                    {formatOptions(rule) || '-'}
                                </td>
                                <td>
                                    <span className={`action-badge ${actionToString(rule.action).toLowerCase()}`}>
                                        {actionToString(rule.action)}
                                    </span>
                                </td>
                                <td>{rule.dport || '-'}</td>
                                <td className="ip-cell" title={rule.sip || ''}>{rule.sip || '-'}</td>
                                <td className="ip-cell" title={rule.dip || ''}>{rule.dip || '-'}</td>
                                <td>{rule.black ? 'Y' : '-'}</td>
                                <td>{rule.white ? 'Y' : '-'}</td>
                            </tr>
                        ))
                    )}
                </tbody>
            </table>
        </div>
    );
};

export default RuleTable;
