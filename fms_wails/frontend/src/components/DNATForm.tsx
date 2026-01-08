import { useState, useEffect } from 'react';
import { model } from '../../wailsjs/go/models';
import { GetProtocolOptions } from '../../wailsjs/go/main/App';
import { DNAT_HELP } from '../constants/helpTexts';

// Protocol 상수
const PROTOCOL_TCP = 0;
const PROTOCOL_UDP = 1;
const PROTOCOL_ANY = 3;

// 문자열을 Protocol 값으로 변환
const stringToProtocol = (s: string): number => {
    switch (s.toLowerCase()) {
        case 'tcp': return PROTOCOL_TCP;
        case 'udp': return PROTOCOL_UDP;
        case 'any': return PROTOCOL_ANY;
        default: return PROTOCOL_TCP;
    }
};

interface DNATFormProps {
    onAdd: (rule: model.NATRule) => void;
    editRule?: model.NATRule | null;
    editIndex?: number;
    onUpdate?: (index: number, rule: model.NATRule) => void;
    onCancel?: () => void;
}

const DNATForm = ({ onAdd, editRule, editIndex, onUpdate, onCancel }: DNATFormProps) => {
    const [protocolOptions, setProtocolOptions] = useState<string[]>([]);
    const [showHelp, setShowHelp] = useState(false);

    // 폼 상태
    const [protocol, setProtocol] = useState('tcp');
    const [matchPort, setMatchPort] = useState('');
    const [matchIP, setMatchIP] = useState('');
    const [translateIP, setTranslateIP] = useState('');
    const [translatePort, setTranslatePort] = useState('');

    // 옵션 로드
    useEffect(() => {
        GetProtocolOptions().then(setProtocolOptions);
    }, []);

    // 편집 모드일 때 폼 값 설정
    useEffect(() => {
        if (editRule) {
            setProtocol(protocolOptions[editRule.protocol] || 'tcp');
            setMatchPort(editRule.matchPort || '');
            setMatchIP(editRule.matchIP || '');
            setTranslateIP(editRule.translateIP || '');
            setTranslatePort(editRule.translatePort || '');
        }
    }, [editRule, protocolOptions]);

    // 폼 초기화
    const resetForm = () => {
        setProtocol('tcp');
        setMatchPort('');
        setMatchIP('');
        setTranslateIP('');
        setTranslatePort('');
    };

    // 규칙 생성
    const createRule = (): model.NATRule => {
        const rule = new model.NATRule();
        rule.natType = 0; // DNAT
        rule.protocol = stringToProtocol(protocol);
        rule.matchPort = matchPort || undefined;
        rule.matchIP = matchIP || 'ANY';
        rule.translateIP = translateIP || undefined;
        rule.translatePort = translatePort || undefined;
        return rule;
    };

    // 추가/수정 처리
    const handleSubmit = () => {
        if (!translateIP) {
            alert('내부 IP를 입력하세요.');
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

    return (
        <div className="rule-form">
            <div className="rule-form-header">
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                    <span className="rule-form-title">{editRule ? 'DNAT 규칙 수정' : 'DNAT 규칙 추가 (포트 포워딩)'}</span>
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
                            <span className="help-modal-title">{DNAT_HELP.title}</span>
                            <button className="help-modal-close" onClick={() => setShowHelp(false)}>×</button>
                        </div>
                        <div className="help-modal-content help-modal-scroll">
                            <p>{DNAT_HELP.description.split('\n').map((line, i) => <span key={i}>{line}<br /></span>)}</p>
                            <br />
                            <p><strong>[입력 필드 설명]</strong></p>
                            {DNAT_HELP.fields.map((f, i) => (
                                <p key={i}>• <strong>{f.name}</strong><br />  - {f.desc}<br />  - 예: {f.example}</p>
                            ))}
                            <br />
                            <p><strong>[사용 예시]</strong></p>
                            {DNAT_HELP.examples.map((ex, i) => (
                                <p key={i}>{ex.title}:<br />  {ex.config}<br />  → {ex.result}</p>
                            ))}
                        </div>
                    </div>
                </div>
            )}

            <div className="rule-form-grid">
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
                    <label>외부 포트</label>
                    <input
                        type="text"
                        className="input"
                        value={matchPort}
                        onChange={(e) => setMatchPort(e.target.value)}
                        placeholder="예: 80"
                    />
                </div>

                <div className="rule-form-group">
                    <label>소스 IP (선택)</label>
                    <input
                        type="text"
                        className="input"
                        value={matchIP}
                        onChange={(e) => setMatchIP(e.target.value)}
                        placeholder="예: 10.0.0.0/8"
                    />
                </div>

                <div className="rule-form-group">
                    <label>내부 IP *</label>
                    <input
                        type="text"
                        className="input"
                        value={translateIP}
                        onChange={(e) => setTranslateIP(e.target.value)}
                        placeholder="예: 192.168.1.100"
                    />
                </div>

                <div className="rule-form-group">
                    <label>내부 포트</label>
                    <input
                        type="text"
                        className="input"
                        value={translatePort}
                        onChange={(e) => setTranslatePort(e.target.value)}
                        placeholder="예: 8080"
                    />
                </div>
            </div>

        </div>
    );
};

export default DNATForm;
