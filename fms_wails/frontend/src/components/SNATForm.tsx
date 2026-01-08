import { useState, useEffect } from 'react';
import { model } from '../../wailsjs/go/models';
import { GetProtocolOptions, GetSNATTypeOptions } from '../../wailsjs/go/main/App';
import { SNAT_HELP } from '../constants/helpTexts';

// NATType 상수
const NAT_TYPE_SNAT = 1;
const NAT_TYPE_MASQUERADE = 2;

// Protocol 상수
const PROTOCOL_TCP = 0;
const PROTOCOL_UDP = 1;
const PROTOCOL_ANY = 3;

// 문자열을 NATType 값으로 변환
const stringToNATType = (s: string): number => {
    switch (s.toUpperCase()) {
        case 'SNAT': return NAT_TYPE_SNAT;
        case 'MASQUERADE': return NAT_TYPE_MASQUERADE;
        default: return NAT_TYPE_SNAT;
    }
};

// 문자열을 Protocol 값으로 변환
const stringToProtocol = (s: string): number => {
    switch (s.toLowerCase()) {
        case 'tcp': return PROTOCOL_TCP;
        case 'udp': return PROTOCOL_UDP;
        case 'any': return PROTOCOL_ANY;
        default: return PROTOCOL_TCP;
    }
};

interface SNATFormProps {
    onAdd: (rule: model.NATRule) => void;
    editRule?: model.NATRule | null;
    editIndex?: number;
    onUpdate?: (index: number, rule: model.NATRule) => void;
    onCancel?: () => void;
}

const SNATForm = ({ onAdd, editRule, editIndex, onUpdate, onCancel }: SNATFormProps) => {
    const [protocolOptions, setProtocolOptions] = useState<string[]>([]);
    const [natTypeOptions, setNatTypeOptions] = useState<string[]>([]);
    const [showHelp, setShowHelp] = useState(false);

    // 폼 상태
    const [natType, setNatType] = useState('SNAT');
    const [protocol, setProtocol] = useState('tcp');
    const [matchIP, setMatchIP] = useState('');
    const [inInterface, setInInterface] = useState('');
    const [outInterface, setOutInterface] = useState('');
    const [translateIP, setTranslateIP] = useState('');

    // 옵션 로드
    useEffect(() => {
        Promise.all([
            GetProtocolOptions(),
            GetSNATTypeOptions()
        ]).then(([protocols, natTypes]) => {
            setProtocolOptions(protocols);
            setNatTypeOptions(natTypes);
        });
    }, []);

    // 편집 모드일 때 폼 값 설정
    useEffect(() => {
        if (editRule) {
            setNatType(editRule.natType === NAT_TYPE_MASQUERADE ? 'MASQUERADE' : 'SNAT');
            setProtocol(protocolOptions[editRule.protocol] || 'tcp');
            setMatchIP(editRule.matchIP || '');
            setInInterface(editRule.inInterface || '');
            setOutInterface(editRule.outInterface || '');
            setTranslateIP(editRule.translateIP || '');
        }
    }, [editRule, protocolOptions]);

    // 폼 초기화
    const resetForm = () => {
        setNatType('SNAT');
        setProtocol('tcp');
        setMatchIP('');
        setInInterface('');
        setOutInterface('');
        setTranslateIP('');
    };

    // 규칙 생성
    const createRule = (): model.NATRule => {
        const rule = new model.NATRule();
        rule.natType = stringToNATType(natType);
        rule.protocol = stringToProtocol(protocol);
        rule.matchIP = matchIP || undefined;
        rule.inInterface = inInterface || undefined;
        rule.outInterface = outInterface || undefined;
        rule.translateIP = natType === 'SNAT' ? translateIP : undefined;
        return rule;
    };

    // 추가/수정 처리
    const handleSubmit = () => {
        if (natType === 'SNAT' && !translateIP) {
            alert('변환 IP를 입력하세요.');
            return;
        }

        const rule = createRule();

        if (editRule && editIndex !== undefined && onUpdate) {
            onUpdate(editIndex, rule);
        } else {
            onAdd(rule);
        }
        resetForm();
    };

    // 취소 처리
    const handleCancel = () => {
        resetForm();
        onCancel?.();
    };

    const isMasquerade = natType === 'MASQUERADE';

    return (
        <div className="rule-form">
            <div className="rule-form-header">
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                    <span className="rule-form-title">{editRule ? 'SNAT/MASQUERADE 규칙 수정' : 'SNAT/MASQUERADE 규칙 추가'}</span>
                    <button className="help-btn" onClick={() => setShowHelp(!showHelp)} title="도움말">
                        ?
                    </button>
                </div>
                <div className="rule-form-header-buttons">
                    {editRule && (
                        <button className="btn btn-secondary btn-sm" onClick={handleCancel}>
                            취소
                        </button>
                    )}
                    <button className="btn btn-primary btn-sm" onClick={handleSubmit}>
                        {editRule ? '수정' : '추가'}
                    </button>
                </div>
            </div>

            {showHelp && (
                <div className="help-modal-overlay" onClick={() => setShowHelp(false)}>
                    <div className="help-modal help-modal-large" onClick={(e) => e.stopPropagation()}>
                        <div className="help-modal-header">
                            <span className="help-modal-title">{SNAT_HELP.title}</span>
                            <button className="help-modal-close" onClick={() => setShowHelp(false)}>×</button>
                        </div>
                        <div className="help-modal-content help-modal-scroll">
                            <p>{SNAT_HELP.description.split('\n').map((line, i) => <span key={i}>{line}<br /></span>)}</p>
                            <br />
                            <p><strong>[NAT 타입 선택]</strong></p>
                            {SNAT_HELP.natTypes.map((nt, i) => (
                                <p key={i}>• <strong>{nt.name}</strong><br />  - {nt.desc}<br />  - 예: {nt.example}</p>
                            ))}
                            <br />
                            <p><strong>[입력 필드 설명]</strong></p>
                            {SNAT_HELP.fields.map((f, i) => (
                                <p key={i}>• <strong>{f.name}</strong><br />  - {f.desc}<br />  - 예: {f.example}</p>
                            ))}
                            <br />
                            <p><strong>[사용 예시]</strong></p>
                            {SNAT_HELP.examples.map((ex, i) => (
                                <p key={i}>{ex.title}:<br />  {ex.config}<br />  → {ex.result}</p>
                            ))}
                        </div>
                    </div>
                </div>
            )}

            <div className="rule-form-grid">
                <div className="rule-form-group">
                    <label>NAT 타입</label>
                    <select
                        className="select"
                        value={natType}
                        onChange={(e) => setNatType(e.target.value)}
                    >
                        {natTypeOptions.map((opt) => (
                            <option key={opt} value={opt}>{opt}</option>
                        ))}
                    </select>
                </div>

                <div className="rule-form-group">
                    <label>Protocol</label>
                    <select
                        className="select"
                        value={protocol}
                        onChange={(e) => setProtocol(e.target.value)}
                    >
                        {protocolOptions.filter(opt => opt !== 'icmp').map((opt) => (
                            <option key={opt} value={opt}>{opt}</option>
                        ))}
                    </select>
                </div>

                <div className="rule-form-group">
                    <label>소스 네트워크</label>
                    <input
                        type="text"
                        className="input"
                        value={matchIP}
                        onChange={(e) => setMatchIP(e.target.value)}
                        placeholder="예: 192.168.1.0/24"
                    />
                </div>

                <div className="rule-form-group">
                    <label>입력 인터페이스 (선택)</label>
                    <input
                        type="text"
                        className="input"
                        value={inInterface}
                        onChange={(e) => setInInterface(e.target.value)}
                        placeholder="예: eth1"
                    />
                </div>

                <div className="rule-form-group">
                    <label>출력 인터페이스</label>
                    <input
                        type="text"
                        className="input"
                        value={outInterface}
                        onChange={(e) => setOutInterface(e.target.value)}
                        placeholder="예: eth0"
                    />
                </div>

                {!isMasquerade && (
                    <div className="rule-form-group">
                        <label>변환 IP *</label>
                        <input
                            type="text"
                            className="input"
                            value={translateIP}
                            onChange={(e) => setTranslateIP(e.target.value)}
                            placeholder="예: 203.0.113.1"
                        />
                    </div>
                )}
            </div>

        </div>
    );
};

export default SNATForm;
