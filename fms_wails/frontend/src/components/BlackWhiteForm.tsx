import { useState, useEffect } from 'react';
import { model } from '../../wailsjs/go/models';
import { BLACK_WHITE_HELP } from '../constants/helpTexts';

// Chain 상수
const CHAIN_INPUT = 0;

// Protocol 상수
const PROTOCOL_ANY = 3;

// Action 상수
const ACTION_DROP = 0;
const ACTION_ACCEPT = 1;

interface BlackWhiteFormProps {
    onAdd: (rule: model.FirewallRule) => void;
    editRule?: model.FirewallRule | null;
    editIndex?: number;
    onUpdate?: (index: number, rule: model.FirewallRule) => void;
    onCancel?: () => void;
}

const BlackWhiteForm = ({ onAdd, editRule, editIndex, onUpdate, onCancel }: BlackWhiteFormProps) => {
    const [showHelp, setShowHelp] = useState(false);

    // 폼 상태
    const [ruleType, setRuleType] = useState<'Black' | 'White'>('Black');
    const [sip, setSip] = useState('');

    // 편집 모드일 때 폼 값 설정
    useEffect(() => {
        if (editRule) {
            setRuleType(editRule.white ? 'White' : 'Black');
            setSip(editRule.sip || '');
        }
    }, [editRule]);

    // 폼 초기화
    const resetForm = () => {
        setRuleType('Black');
        setSip('');
    };

    // 규칙 생성
    const createRule = (): model.FirewallRule => {
        const rule = new model.FirewallRule();
        rule.chain = CHAIN_INPUT;
        rule.protocol = PROTOCOL_ANY;
        rule.action = ruleType === 'Black' ? ACTION_DROP : ACTION_ACCEPT;
        rule.sip = sip || undefined;
        rule.black = ruleType === 'Black';
        rule.white = ruleType === 'White';
        return rule;
    };

    // 추가/수정 처리
    const handleSubmit = () => {
        if (!sip.trim()) {
            alert('IP 주소를 입력하세요.');
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
                    <span className="rule-form-title">{editRule ? 'Black/White 규칙 수정' : 'Black/White 규칙 추가'}</span>
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
                    <div className="help-modal" onClick={(e) => e.stopPropagation()}>
                        <div className="help-modal-header">
                            <span className="help-modal-title">{BLACK_WHITE_HELP.title}</span>
                            <button className="help-modal-close" onClick={() => setShowHelp(false)}>×</button>
                        </div>
                        <div className="help-modal-content">
                            {BLACK_WHITE_HELP.types.map((t, i) => (
                                <p key={i}><strong>{t.name}:</strong> {t.desc}</p>
                            ))}
                            <br />
                            {BLACK_WHITE_HELP.fields.map((f, i) => (
                                <p key={i}><strong>{f.name}:</strong> {f.desc}<br />예: {f.example}</p>
                            ))}
                        </div>
                    </div>
                </div>
            )}

            <div className="rule-form-grid">
                <div className="rule-form-group">
                    <label>Type</label>
                    <select
                        className="select"
                        value={ruleType}
                        onChange={(e) => setRuleType(e.target.value as 'Black' | 'White')}
                    >
                        <option value="Black">Black (차단)</option>
                        <option value="White">White (허용)</option>
                    </select>
                </div>

                <div className="rule-form-group" style={{ flex: 2 }}>
                    <label>IP</label>
                    <input
                        type="text"
                        className="input"
                        value={sip}
                        onChange={(e) => setSip(e.target.value)}
                        placeholder="예: 192.168.1.100"
                    />
                </div>
            </div>

        </div>
    );
};

export default BlackWhiteForm;
