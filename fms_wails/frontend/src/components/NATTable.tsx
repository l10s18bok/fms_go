import { model } from '../../wailsjs/go/models';

// NATType 상수
const NAT_TYPE_DNAT = 0;
const NAT_TYPE_SNAT = 1;
const NAT_TYPE_MASQUERADE = 2;

// Protocol 상수
const PROTOCOL_TCP = 0;
const PROTOCOL_UDP = 1;
const PROTOCOL_ANY = 3;

// NATType 값을 문자열로 변환
const natTypeToString = (natType: number): string => {
    switch (natType) {
        case NAT_TYPE_DNAT: return 'DNAT';
        case NAT_TYPE_SNAT: return 'SNAT';
        case NAT_TYPE_MASQUERADE: return 'MASQ';
        default: return 'DNAT';
    }
};

// Protocol 값을 문자열로 변환
const protocolToString = (protocol: number): string => {
    switch (protocol) {
        case PROTOCOL_TCP: return 'tcp';
        case PROTOCOL_UDP: return 'udp';
        case PROTOCOL_ANY: return 'any';
        default: return 'tcp';
    }
};

interface NATTableProps {
    rules: model.NATRule[];
    onDelete: (index: number) => void;
    onEdit?: (index: number, rule: model.NATRule) => void;
}

// Match 정보 포맷 (fyne 스타일: IP:Port)
const formatMatch = (rule: model.NATRule): string => {
    let matchStr = '';
    if (rule.matchIP && rule.matchIP !== 'ANY') {
        matchStr = rule.matchIP;
    }
    if (rule.matchPort) {
        if (matchStr) matchStr += ':';
        matchStr += rule.matchPort;
    }
    return matchStr || 'ANY';
};

// Translate 정보 포맷 (fyne 스타일: IP:Port)
const formatTranslate = (rule: model.NATRule): string => {
    let transStr = '';
    if (rule.translateIP) {
        transStr = rule.translateIP;
    }
    if (rule.translatePort) {
        if (transStr) transStr += ':';
        transStr += rule.translatePort;
    }
    return transStr || '-';
};

// Interface 정보 포맷 (fyne 스타일: IN:xxx OUT:xxx)
const formatInterface = (rule: model.NATRule): string => {
    const parts: string[] = [];
    if (rule.inInterface) parts.push(`IN:${rule.inInterface}`);
    if (rule.outInterface) parts.push(`OUT:${rule.outInterface}`);
    return parts.length > 0 ? parts.join(' ') : '-';
};

const NATTable = ({ rules, onDelete, onEdit }: NATTableProps) => {
    return (
        <div className="rule-table-container">
            <table className="table rule-table nat-table">
                <thead>
                    <tr>
                        <th style={{ width: '40px' }}></th>
                        <th style={{ width: '80px' }}>Type</th>
                        <th style={{ width: '60px' }}>Proto</th>
                        <th style={{ width: '20%' }}>Match</th>
                        <th style={{ width: '25%' }}>Translate</th>
                        <th style={{ width: '25%' }}>Interface</th>
                    </tr>
                </thead>
                <tbody>
                    {rules.length === 0 ? (
                        <tr>
                            <td colSpan={6} style={{ textAlign: 'center', color: '#666', padding: '20px' }}>
                                NAT 규칙이 없습니다. 아래에서 규칙을 추가하세요.
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
                                        title="Delete"
                                    >
                                        X
                                    </button>
                                </td>
                                <td>
                                    <span className={`nat-type-badge ${natTypeToString(rule.natType).toLowerCase()}`}>
                                        {natTypeToString(rule.natType)}
                                    </span>
                                </td>
                                <td>{protocolToString(rule.protocol)}</td>
                                <td>{formatMatch(rule)}</td>
                                <td>{formatTranslate(rule)}</td>
                                <td>{formatInterface(rule)}</td>
                            </tr>
                        ))
                    )}
                </tbody>
            </table>
        </div>
    );
};

export default NATTable;
