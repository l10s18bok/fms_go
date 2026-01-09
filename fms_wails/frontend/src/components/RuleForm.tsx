import { useState, useEffect } from 'react';
import { model } from '../../wailsjs/go/models';
import {
    GetChainOptions,
    GetProtocolOptions,
    GetActionOptions,
    GetTCPFlagsPresets,
    GetTCPFlagsList,
    GetICMPTypeOptions
} from '../../wailsjs/go/main/App';
import { TCP_FLAGS_HELP, ICMP_HELP } from '../constants/helpTexts';
import {
    CHAIN_INPUT, CHAIN_OUTPUT, CHAIN_FORWARD,
    PROTOCOL_TCP, PROTOCOL_UDP, PROTOCOL_ICMP, PROTOCOL_ANY,
    ACTION_DROP, ACTION_ACCEPT,
} from '../constants/ruleConstants';

// 문자열을 Chain 값으로 변환
const stringToChain = (s: string): number => {
    switch (s.toUpperCase()) {
        case 'INPUT': return CHAIN_INPUT;
        case 'OUTPUT': return CHAIN_OUTPUT;
        case 'FORWARD': return CHAIN_FORWARD;
        default: return CHAIN_INPUT;
    }
};

// 문자열을 Protocol 값으로 변환
const stringToProtocol = (s: string): number => {
    switch (s.toLowerCase()) {
        case 'tcp': return PROTOCOL_TCP;
        case 'udp': return PROTOCOL_UDP;
        case 'icmp': return PROTOCOL_ICMP;
        case 'any': return PROTOCOL_ANY;
        default: return PROTOCOL_TCP;
    }
};

// 문자열을 Action 값으로 변환
const stringToAction = (s: string): number => {
    switch (s.toUpperCase()) {
        case 'DROP': return ACTION_DROP;
        case 'ACCEPT': return ACTION_ACCEPT;
        default: return ACTION_DROP;
    }
};

interface RuleFormProps {
    onAdd: (rule: model.FirewallRule) => void;
    editRule?: model.FirewallRule | null;
    editIndex?: number;
    onUpdate?: (index: number, rule: model.FirewallRule) => void;
    onCancel?: () => void;
}

const RuleForm = ({ onAdd, editRule, editIndex, onUpdate, onCancel }: RuleFormProps) => {
    // 옵션 상태
    const [chainOptions, setChainOptions] = useState<string[]>([]);
    const [protocolOptions, setProtocolOptions] = useState<string[]>([]);
    const [actionOptions, setActionOptions] = useState<string[]>([]);
    const [tcpFlagsPresets, setTcpFlagsPresets] = useState<model.TCPFlagsPreset[]>([]);
    const [tcpFlagsList, setTcpFlagsList] = useState<string[]>([]);
    const [icmpTypeOptions, setIcmpTypeOptions] = useState<string[]>([]);

    // 폼 상태
    const [chain, setChain] = useState('INPUT');
    const [protocol, setProtocol] = useState('tcp');
    const [action, setAction] = useState('DROP');
    const [dport, setDport] = useState('');
    const [sip, setSip] = useState('');
    const [dip, setDip] = useState('');

    // TCP Flags 상태
    const [tcpFlagsPreset, setTcpFlagsPreset] = useState('None');
    const [maskFlags, setMaskFlags] = useState<Record<string, boolean>>({});
    const [setFlags, setSetFlags] = useState<Record<string, boolean>>({});

    // ICMP 상태
    const [icmpType, setIcmpType] = useState('None');

    // 도움말 표시 상태
    const [showTcpHelp, setShowTcpHelp] = useState(false);
    const [showIcmpHelp, setShowIcmpHelp] = useState(false);

    // 옵션 로드
    useEffect(() => {
        const loadOptions = async () => {
            const [chains, protocols, actions, presets, flags, types] = await Promise.all([
                GetChainOptions(),
                GetProtocolOptions(),
                GetActionOptions(),
                GetTCPFlagsPresets(),
                GetTCPFlagsList(),
                GetICMPTypeOptions()
            ]);
            setChainOptions(chains);
            setProtocolOptions(protocols);
            setActionOptions(actions);
            setTcpFlagsPresets(presets);
            setTcpFlagsList(flags);
            setIcmpTypeOptions(types);

            // 플래그 초기화
            const initMask: Record<string, boolean> = {};
            const initSet: Record<string, boolean> = {};
            flags.forEach((f: string) => {
                initMask[f] = false;
                initSet[f] = false;
            });
            setMaskFlags(initMask);
            setSetFlags(initSet);
        };
        loadOptions();
    }, []);

    // 편집 모드일 때 폼 값 설정
    useEffect(() => {
        if (editRule) {
            setChain(chainOptions[editRule.chain] || 'INPUT');
            setProtocol(protocolOptions[editRule.protocol] || 'tcp');
            setAction(actionOptions[editRule.action] || 'DROP');
            setDport(editRule.dport || '');
            setSip(editRule.sip || '');
            setDip(editRule.dip || '');

            // 프로토콜 옵션
            if (editRule.options?.tcpFlags) {
                const [maskStr, setStr] = editRule.options.tcpFlags.split('/');
                const maskArr = maskStr ? maskStr.split(',') : [];
                const setArr = setStr ? setStr.split(',') : [];

                const newMask: Record<string, boolean> = {};
                const newSet: Record<string, boolean> = {};
                tcpFlagsList.forEach(f => {
                    newMask[f] = maskArr.includes(f);
                    newSet[f] = setArr.includes(f);
                });
                setMaskFlags(newMask);
                setSetFlags(newSet);

                // 프리셋 찾기
                const preset = tcpFlagsPresets.find(p => {
                    if (!p.maskFlags || !p.setFlags) return false;
                    return JSON.stringify(p.maskFlags.sort()) === JSON.stringify(maskArr.sort()) &&
                           JSON.stringify(p.setFlags.sort()) === JSON.stringify(setArr.sort());
                });
                setTcpFlagsPreset(preset?.name || 'Custom');
            } else {
                setTcpFlagsPreset('None');
                resetFlags();
            }

            if (editRule.options?.icmpType) {
                setIcmpType(editRule.options.icmpType);
            } else {
                setIcmpType('None');
            }
        }
    }, [editRule, chainOptions, protocolOptions, actionOptions, tcpFlagsPresets, tcpFlagsList]);

    // 플래그 초기화
    const resetFlags = () => {
        const initMask: Record<string, boolean> = {};
        const initSet: Record<string, boolean> = {};
        tcpFlagsList.forEach(f => {
            initMask[f] = false;
            initSet[f] = false;
        });
        setMaskFlags(initMask);
        setSetFlags(initSet);
    };

    // 폼 초기화
    const resetForm = () => {
        setChain('INPUT');
        setProtocol('tcp');
        setAction('DROP');
        setDport('');
        setSip('');
        setDip('');
        setTcpFlagsPreset('None');
        resetFlags();
        setIcmpType('None');
    };

    // 프리셋 변경 시 체크박스 업데이트
    const handlePresetChange = (presetName: string) => {
        setTcpFlagsPreset(presetName);

        if (presetName === 'None' || presetName === 'Custom') {
            if (presetName === 'None') {
                resetFlags();
            }
            return;
        }

        const preset = tcpFlagsPresets.find(p => p.name === presetName);
        if (!preset) return;

        const newMask: Record<string, boolean> = {};
        const newSet: Record<string, boolean> = {};
        tcpFlagsList.forEach(f => {
            newMask[f] = preset.maskFlags?.includes(f) || false;
            newSet[f] = preset.setFlags?.includes(f) || false;
        });
        setMaskFlags(newMask);
        setSetFlags(newSet);
    };

    // Mask 체크박스 변경
    const handleMaskChange = (flag: string, checked: boolean) => {
        setMaskFlags(prev => ({ ...prev, [flag]: checked }));
        setTcpFlagsPreset('Custom');
    };

    // Set 체크박스 변경
    const handleSetChange = (flag: string, checked: boolean) => {
        setSetFlags(prev => ({ ...prev, [flag]: checked }));
        // SET에 선택된 플래그는 자동으로 MASK에도 선택
        if (checked) {
            setMaskFlags(prev => ({ ...prev, [flag]: true }));
        }
        setTcpFlagsPreset('Custom');
    };

    // TCP flags 문자열 생성
    const getTCPFlags = (): string => {
        const maskArr = tcpFlagsList.filter(f => maskFlags[f]);
        const setArr = tcpFlagsList.filter(f => setFlags[f]);

        if (maskArr.length === 0) return '';
        return `${maskArr.join(',')}/${setArr.join(',')}`;
    };

    // 규칙 생성
    const createRule = (): model.FirewallRule => {
        const rule = new model.FirewallRule();
        rule.chain = stringToChain(chain);
        rule.protocol = stringToProtocol(protocol);
        rule.action = stringToAction(action);
        rule.dport = dport || undefined;
        rule.sip = sip || undefined;
        rule.dip = dip || undefined;
        rule.black = false;
        rule.white = false;

        // 프로토콜 옵션
        const options = new model.ProtocolOptions();
        let hasOptions = false;

        if (protocol === 'tcp') {
            const tcpFlags = getTCPFlags();
            if (tcpFlags) {
                options.tcpFlags = tcpFlags;
                hasOptions = true;
            }
        }

        if (protocol === 'icmp' && icmpType !== 'None') {
            options.icmpType = icmpType;
            hasOptions = true;
        }

        if (hasOptions) {
            rule.options = options;
        }

        return rule;
    };

    // 추가/수정 처리
    const handleSubmit = () => {
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

    // 프로토콜에 따라 옵션 표시 및 필드 활성화
    const showTcpOptions = protocol === 'tcp' || protocol === 'udp' || protocol === 'any';
    const showIcmpOptions = protocol === 'icmp';
    const tcpOptionsEnabled = protocol === 'tcp';
    const portEnabled = protocol !== 'icmp';

    return (
        <div className="rule-form">
            {/* 헤더: 규칙 추가 + 추가 버튼 */}
            <div className="rule-form-header">
                <span className="rule-form-title">{editRule ? '규칙 수정' : '규칙 추가'}</span>
                <div className="rule-form-header-buttons">
                    {editRule && (
                        <button className="btn btn-secondary btn-sm" onClick={handleCancel}>
                            취소
                        </button>
                    )}
                    <button className="btn btn-primary btn-sm" onClick={handleSubmit}>
                        {editRule ? '수정' : '+ 추가'}
                    </button>
                </div>
            </div>

            {/* 첫 번째 행: Chain, Proto, Action, Port */}
            <div className="rule-form-row">
                <div className="rule-form-field">
                    <label>Chain:</label>
                    <select className="select select-sm" value={chain} onChange={(e) => setChain(e.target.value)}>
                        {chainOptions.map((opt) => (
                            <option key={opt} value={opt}>{opt}</option>
                        ))}
                    </select>
                </div>
                <div className="rule-form-field">
                    <label>Proto:</label>
                    <select
                        className="select select-sm"
                        value={protocol}
                        onChange={(e) => {
                            setProtocol(e.target.value);
                            setTcpFlagsPreset('None');
                            resetFlags();
                            setIcmpType('None');
                        }}
                    >
                        {protocolOptions.map((opt) => (
                            <option key={opt} value={opt}>{opt}</option>
                        ))}
                    </select>
                </div>
                <div className="rule-form-field">
                    <label>Action:</label>
                    <select className="select select-sm" value={action} onChange={(e) => setAction(e.target.value)}>
                        {actionOptions.map((opt) => (
                            <option key={opt} value={opt}>{opt}</option>
                        ))}
                    </select>
                </div>
                <div className="rule-form-field">
                    <label>Port:</label>
                    <input
                        type="text"
                        className="input input-sm"
                        value={dport}
                        onChange={(e) => setDport(e.target.value)}
                        placeholder="포트"
                        disabled={!portEnabled}
                    />
                </div>
            </div>

            {/* 두 번째 행: SIP, DIP */}
            <div className="rule-form-row">
                <div className="rule-form-field" style={{ flex: 1 }}>
                    <label>SIP:</label>
                    <input
                        type="text"
                        className="input input-sm"
                        value={sip}
                        onChange={(e) => setSip(e.target.value)}
                        placeholder="Source IP"
                        style={{ width: '200px' }}
                    />
                </div>
                <div className="rule-form-field" style={{ flex: 1 }}>
                    <label>DIP:</label>
                    <input
                        type="text"
                        className="input input-sm"
                        value={dip}
                        onChange={(e) => setDip(e.target.value)}
                        placeholder="Dest IP"
                        style={{ width: '200px' }}
                    />
                </div>
            </div>

            {/* TCP Flags 옵션 (tcp, udp, any 모두 표시, tcp만 활성화) */}
            {showTcpOptions && (
                <div className={`tcp-flags-section ${!tcpOptionsEnabled ? 'disabled' : ''}`}>
                    <div className="tcp-flags-header">
                        <span>TCP Flags</span>
                        <button
                            className="help-btn-inline"
                            onClick={() => setShowTcpHelp(!showTcpHelp)}
                            title="도움말"
                            disabled={!tcpOptionsEnabled}
                        >
                            ?
                        </button>
                    </div>

                    {showTcpHelp && tcpOptionsEnabled && (
                        <div className="help-modal-overlay" onClick={() => setShowTcpHelp(false)}>
                            <div className="help-modal help-modal-large" onClick={(e) => e.stopPropagation()}>
                                <div className="help-modal-header">
                                    <span className="help-modal-title">{TCP_FLAGS_HELP.title}</span>
                                    <button className="help-modal-close" onClick={() => setShowTcpHelp(false)}>×</button>
                                </div>
                                <div className="help-modal-content help-modal-scroll">
                                    <p><strong>[프리셋]</strong></p>
                                    {TCP_FLAGS_HELP.presets.map((p, i) => (
                                        <p key={i}>• <strong>{p.name}</strong><br />  - {p.desc}</p>
                                    ))}
                                    <br />
                                    <p><strong>[플래그 설명]</strong></p>
                                    {TCP_FLAGS_HELP.flags.map((f, i) => (
                                        <p key={i}>• <strong>{f.name}</strong> - {f.desc}</p>
                                    ))}
                                    <br />
                                    <p><strong>[Mask / Set 설명]</strong></p>
                                    <p>• <strong>Mask:</strong> {TCP_FLAGS_HELP.maskSetDesc.mask}</p>
                                    <p>• <strong>Set:</strong> {TCP_FLAGS_HELP.maskSetDesc.set}</p>
                                    <p>예) {TCP_FLAGS_HELP.maskSetDesc.example}</p>
                                </div>
                            </div>
                        </div>
                    )}

                    {/* Preset 행 */}
                    <div className="rule-form-row">
                        <div className="rule-form-field">
                            <label>Preset:</label>
                            <select
                                className="select select-sm"
                                value={tcpFlagsPreset}
                                onChange={(e) => handlePresetChange(e.target.value)}
                                disabled={!tcpOptionsEnabled}
                                style={{ width: '180px' }}
                            >
                                {tcpFlagsPresets.map((preset) => (
                                    <option key={preset.name} value={preset.name}>{preset.name}</option>
                                ))}
                            </select>
                        </div>
                    </div>

                    {/* Mask 체크박스 행 */}
                    <div className="rule-form-row">
                        <div className="rule-form-field">
                            <label>Mask:</label>
                        </div>
                        {tcpFlagsList.map(flag => (
                            <label key={`mask-${flag}`} className="checkbox-inline">
                                <input
                                    type="checkbox"
                                    checked={maskFlags[flag] || false}
                                    onChange={(e) => handleMaskChange(flag, e.target.checked)}
                                    disabled={!tcpOptionsEnabled}
                                />
                                {flag.toUpperCase()}
                            </label>
                        ))}
                    </div>

                    {/* Set 체크박스 행 */}
                    <div className="rule-form-row">
                        <div className="rule-form-field">
                            <label>Set:</label>
                        </div>
                        {tcpFlagsList.map(flag => (
                            <label key={`set-${flag}`} className="checkbox-inline">
                                <input
                                    type="checkbox"
                                    checked={setFlags[flag] || false}
                                    onChange={(e) => handleSetChange(flag, e.target.checked)}
                                    disabled={!tcpOptionsEnabled}
                                />
                                {flag.toUpperCase()}
                            </label>
                        ))}
                    </div>
                </div>
            )}

            {/* ICMP 옵션 */}
            {showIcmpOptions && (
                <div className="icmp-section">
                    <div className="tcp-flags-header">
                        <span>ICMP Options</span>
                        <button
                            className="help-btn-inline"
                            onClick={() => setShowIcmpHelp(!showIcmpHelp)}
                            title="도움말"
                        >
                            ?
                        </button>
                    </div>

                    {showIcmpHelp && (
                        <div className="help-modal-overlay" onClick={() => setShowIcmpHelp(false)}>
                            <div className="help-modal help-modal-large" onClick={(e) => e.stopPropagation()}>
                                <div className="help-modal-header">
                                    <span className="help-modal-title">{ICMP_HELP.title}</span>
                                    <button className="help-modal-close" onClick={() => setShowIcmpHelp(false)}>×</button>
                                </div>
                                <div className="help-modal-content help-modal-scroll">
                                    {ICMP_HELP.types.map((t, i) => (
                                        <p key={i}>• <strong>{t.name}</strong> - {t.desc}</p>
                                    ))}
                                </div>
                            </div>
                        </div>
                    )}

                    <div className="rule-form-row">
                        <div className="rule-form-field">
                            <label>Type:</label>
                            <select
                                className="select select-sm"
                                value={icmpType}
                                onChange={(e) => setIcmpType(e.target.value)}
                                style={{ width: '200px' }}
                            >
                                {icmpTypeOptions.map((opt) => (
                                    <option key={opt} value={opt}>{opt}</option>
                                ))}
                            </select>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default RuleForm;
