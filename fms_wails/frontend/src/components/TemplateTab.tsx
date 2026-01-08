import { useState, useEffect, forwardRef, useImperativeHandle } from 'react';
import {
    GetAllTemplates,
    GetTemplate,
    SaveTemplate,
    DeleteTemplate,
    ConfirmDialog,
    ParseRules,
    RulesToText,
    ParseNATRules,
    NATRulesToText
} from '../../wailsjs/go/main/App';
import { model } from '../../wailsjs/go/models';
import RuleTable from './RuleTable';
import RuleForm from './RuleForm';
import BlackWhiteForm from './BlackWhiteForm';
import NATTable from './NATTable';
import DNATForm from './DNATForm';
import SNATForm from './SNATForm';

interface Template {
    version: string;
    contents: string;
}

type SubTabType = 'text' | 'builder' | 'nat';
type NATFormType = 'dnat' | 'snat';
type RuleFormType = 'general' | 'blackwhite';

export interface TemplateTabRef {
    refresh: () => void;
}

const TemplateTab = forwardRef<TemplateTabRef>((_, ref) => {
    const [templates, setTemplates] = useState<Template[]>([]);
    const [selectedVersion, setSelectedVersion] = useState<string>('');
    const [version, setVersion] = useState('');
    const [contents, setContents] = useState('');
    const [isNew, setIsNew] = useState(false);
    const [subTab, setSubTab] = useState<SubTabType>('text');

    // 규칙 빌더 상태
    const [rules, setRules] = useState<model.FirewallRule[]>([]);
    const [comments, setComments] = useState<string[]>([]);
    const [parseErrors, setParseErrors] = useState<string[]>([]);
    const [editRule, setEditRule] = useState<model.FirewallRule | null>(null);
    const [editIndex, setEditIndex] = useState<number | undefined>(undefined);

    // NAT 규칙 상태
    const [natRules, setNatRules] = useState<model.NATRule[]>([]);
    const [natComments, setNatComments] = useState<string[]>([]);
    const [natParseErrors, setNatParseErrors] = useState<string[]>([]);
    const [editNatRule, setEditNatRule] = useState<model.NATRule | null>(null);
    const [editNatIndex, setEditNatIndex] = useState<number | undefined>(undefined);
    const [natFormType, setNatFormType] = useState<NATFormType>('dnat');
    const [ruleFormType, setRuleFormType] = useState<RuleFormType>('general');

    useEffect(() => {
        loadTemplates();
    }, []);

    const loadTemplates = async () => {
        const data = await GetAllTemplates();
        setTemplates(data || []);
    };

    // 부모 컴포넌트에서 호출할 수 있도록 refresh 메서드 노출
    useImperativeHandle(ref, () => ({
        refresh: () => {
            loadTemplates();
            setSelectedVersion('');
            setVersion('');
            setContents('');
            setIsNew(false);
            setRules([]);
            setComments([]);
            setParseErrors([]);
            setNatRules([]);
            setNatComments([]);
            setNatParseErrors([]);
            setSubTab('text');
        }
    }));

    // 텍스트 → 규칙 파싱
    const parseContentsToRules = async (text: string) => {
        if (!text.trim()) {
            setRules([]);
            setComments([]);
            setParseErrors([]);
            return;
        }

        const result = await ParseRules(text);
        setRules(result.rules || []);
        setComments(result.comments || []);
        setParseErrors(result.errors || []);
    };

    // 규칙 → 텍스트 변환
    const rulesToContents = async (ruleList: model.FirewallRule[], commentList: string[]) => {
        const text = await RulesToText(
            JSON.stringify(ruleList),
            JSON.stringify(commentList)
        );
        setContents(text);
    };

    // 텍스트 → NAT 규칙 파싱
    const parseContentsToNATRules = async (text: string) => {
        if (!text.trim()) {
            setNatRules([]);
            setNatComments([]);
            setNatParseErrors([]);
            return;
        }

        const result = await ParseNATRules(text);
        setNatRules(result.rules || []);
        setNatComments(result.comments || []);
        setNatParseErrors(result.errors || []);
    };

    // NAT 규칙 → 텍스트 변환
    const natRulesToContents = async (ruleList: model.NATRule[], commentList: string[]) => {
        const text = await NATRulesToText(
            JSON.stringify(ruleList),
            JSON.stringify(commentList)
        );
        setContents(text);
    };

    // 빌더 내용을 텍스트로 통합 (fyne의 syncBuildersToText와 동일)
    const syncBuildersToText = async (): Promise<string> => {
        // 필터 규칙 (주석 포함)
        const filterText = await RulesToText(
            JSON.stringify(rules),
            JSON.stringify(comments)
        );

        // NAT 규칙 (주석 제외 - 필터 규칙에서 이미 포함됨)
        const natText = await NATRulesToText(
            JSON.stringify(natRules),
            JSON.stringify([])  // 빈 배열 - 주석 중복 방지
        );

        // 통합
        let finalText = '';
        if (filterText && natText) {
            finalText = filterText + '\n' + natText;
        } else if (filterText) {
            finalText = filterText;
        } else {
            finalText = natText;
        }

        setContents(finalText);
        return finalText;
    };

    // 서브 탭 전환 시 데이터 동기화
    const handleSubTabChange = async (tab: SubTabType) => {
        if (tab === 'builder' && subTab === 'text') {
            // 텍스트 → 규칙 빌더로 전환: 파싱
            await parseContentsToRules(contents);
            await parseContentsToNATRules(contents);
        } else if (tab === 'text' && subTab === 'builder') {
            // 규칙 빌더 → 텍스트로 전환: 양쪽 빌더 내용 통합
            await syncBuildersToText();
        } else if (tab === 'nat' && subTab === 'text') {
            // 텍스트 → NAT 규칙 빌더로 전환: 파싱
            await parseContentsToRules(contents);
            await parseContentsToNATRules(contents);
        } else if (tab === 'text' && subTab === 'nat') {
            // NAT 규칙 빌더 → 텍스트로 전환: 양쪽 빌더 내용 통합
            await syncBuildersToText();
        } else if (tab === 'nat' && subTab === 'builder') {
            // 규칙 빌더 → NAT로 전환: 현재 규칙 유지, NAT 규칙은 이미 있음
            // 별도 처리 불필요 (각 빌더가 독립적으로 규칙 유지)
        } else if (tab === 'builder' && subTab === 'nat') {
            // NAT → 규칙 빌더로 전환: 현재 NAT 규칙 유지, 일반 규칙은 이미 있음
            // 별도 처리 불필요 (각 빌더가 독립적으로 규칙 유지)
        }
        setSubTab(tab);
        setEditRule(null);
        setEditIndex(undefined);
        setEditNatRule(null);
        setEditNatIndex(undefined);
    };

    const handleSelect = async (ver: string) => {
        setSelectedVersion(ver);
        setIsNew(false);
        const template = await GetTemplate(ver);
        if (template) {
            setVersion(template.version);
            setContents(template.contents);
            // 양쪽 빌더에 모두 파싱 (어느 탭이든 규칙 유지)
            await parseContentsToRules(template.contents);
            await parseContentsToNATRules(template.contents);
        }
        setEditRule(null);
        setEditIndex(undefined);
        setEditNatRule(null);
        setEditNatIndex(undefined);
    };

    const handleNew = () => {
        setSelectedVersion('');
        setIsNew(true);
        setVersion('');
        setContents('');
        setRules([]);
        setComments([]);
        setParseErrors([]);
        setNatRules([]);
        setNatComments([]);
        setNatParseErrors([]);
        setEditRule(null);
        setEditIndex(undefined);
        setEditNatRule(null);
        setEditNatIndex(undefined);
    };

    const handleSave = async () => {
        if (!version.trim()) {
            alert('버전을 입력하세요.');
            return;
        }

        // 빌더 탭이면 양쪽 빌더 내용을 통합하여 저장
        let contentsToSave = contents;
        if (subTab === 'builder' || subTab === 'nat') {
            contentsToSave = await syncBuildersToText();
        }

        if (!contentsToSave.trim()) {
            alert('규칙 내용을 입력해주세요.');
            return;
        }

        await SaveTemplate(version, contentsToSave);
        await loadTemplates();
        setSelectedVersion(version);
        setIsNew(false);
        alert('템플릿이 저장되었습니다.');
    };

    const handleDelete = async () => {
        if (!selectedVersion) {
            alert('삭제할 템플릿이 선택되지 않았습니다.');
            return;
        }
        const result = await ConfirmDialog('삭제 확인', `"${selectedVersion}" 템플릿을 삭제하시겠습니까?`);
        if (result !== '확인' && result !== 'Yes' && result !== '예') {
            return;
        }

        try {
            await DeleteTemplate(selectedVersion);
            await loadTemplates();
            setSelectedVersion('');
            setVersion('');
            setContents('');
            setIsNew(false);
            setRules([]);
            setComments([]);
            setParseErrors([]);
            setNatRules([]);
            setNatComments([]);
            setNatParseErrors([]);
            alert('템플릿이 삭제되었습니다.');
        } catch (err) {
            console.error('템플릿 삭제 실패:', err);
            alert(`템플릿 삭제 실패: ${err}`);
        }
    };

    // 규칙 추가
    const handleAddRule = (rule: model.FirewallRule) => {
        setRules([...rules, rule]);
    };

    // 규칙 삭제
    const handleDeleteRule = (index: number) => {
        setRules(rules.filter((_, i) => i !== index));
    };

    // 규칙 편집 시작
    const handleEditRule = (index: number, rule: model.FirewallRule) => {
        setEditRule(rule);
        setEditIndex(index);
        // Black/White 규칙이면 해당 탭으로 전환
        if (rule.black || rule.white) {
            setRuleFormType('blackwhite');
        } else {
            setRuleFormType('general');
        }
    };

    // 규칙 수정
    const handleUpdateRule = (index: number, rule: model.FirewallRule) => {
        const newRules = [...rules];
        newRules[index] = rule;
        setRules(newRules);
        setEditRule(null);
        setEditIndex(undefined);
    };

    // 편집 취소
    const handleCancelEdit = () => {
        setEditRule(null);
        setEditIndex(undefined);
    };

    // NAT 규칙 추가
    const handleAddNatRule = (rule: model.NATRule) => {
        setNatRules([...natRules, rule]);
    };

    // NAT 규칙 삭제
    const handleDeleteNatRule = (index: number) => {
        setNatRules(natRules.filter((_, i) => i !== index));
    };

    // NAT 규칙 편집 시작
    const handleEditNatRule = (index: number, rule: model.NATRule) => {
        setEditNatRule(rule);
        setEditNatIndex(index);
        // DNAT인지 SNAT/MASQUERADE인지에 따라 폼 전환
        if (rule.natType === 0) {
            setNatFormType('dnat');
        } else {
            setNatFormType('snat');
        }
    };

    // NAT 규칙 수정
    const handleUpdateNatRule = (index: number, rule: model.NATRule) => {
        const newRules = [...natRules];
        newRules[index] = rule;
        setNatRules(newRules);
        setEditNatRule(null);
        setEditNatIndex(undefined);
    };

    // NAT 편집 취소
    const handleCancelNatEdit = () => {
        setEditNatRule(null);
        setEditNatIndex(undefined);
    };

    return (
        <div className="split-layout">
            {/* 왼쪽: 템플릿 목록 */}
            <div className="card">
                <div className="card-title">템플릿 목록</div>
                <button className="btn btn-primary" onClick={handleNew} style={{ width: '100%', marginBottom: '16px' }}>
                    + 새 템플릿
                </button>
                <ul className="list">
                    {templates.length === 0 ? (
                        <li className="list-item" style={{ color: '#666' }}>
                            템플릿이 없습니다
                        </li>
                    ) : (
                        templates.map((t) => (
                            <li
                                key={t.version}
                                className={`list-item ${selectedVersion === t.version ? 'active' : ''}`}
                                onClick={() => handleSelect(t.version)}
                            >
                                {t.version}
                            </li>
                        ))
                    )}
                </ul>
            </div>

            {/* 오른쪽: 템플릿 편집 */}
            <div className="card">
                {(selectedVersion || isNew) ? (
                    <>
                        <div className="template-header">
                            <label className="template-label">템플릿</label>
                            <input
                                type="text"
                                className="input template-version-input"
                                value={version}
                                onChange={(e) => setVersion(e.target.value)}
                                placeholder="예: v1.0.0"
                            />
                            <div className="template-header-buttons">
                                <button className="btn btn-primary btn-sm" onClick={handleSave}>
                                    저장
                                </button>
                                {!isNew && (
                                    <button className="btn btn-danger btn-sm" onClick={handleDelete}>
                                        삭제
                                    </button>
                                )}
                            </div>
                        </div>

                        {/* 서브 탭 */}
                        <div className="sub-tabs">
                            <button
                                className={`sub-tab-btn ${subTab === 'text' ? 'active' : ''}`}
                                onClick={() => handleSubTabChange('text')}
                            >
                                텍스트 편집
                            </button>
                            <button
                                className={`sub-tab-btn ${subTab === 'builder' ? 'active' : ''}`}
                                onClick={() => handleSubTabChange('builder')}
                            >
                                규칙 빌더
                            </button>
                            <button
                                className={`sub-tab-btn ${subTab === 'nat' ? 'active' : ''}`}
                                onClick={() => handleSubTabChange('nat')}
                            >
                                NAT 규칙
                            </button>
                        </div>

                        {/* 텍스트 편집 탭 */}
                        {subTab === 'text' && (
                            <div className="form-group flex-grow">
                                <label>규칙 내용</label>
                                <textarea
                                    className="textarea"
                                    value={contents}
                                    onChange={(e) => setContents(e.target.value)}
                                    placeholder="방화벽 규칙을 입력하세요..."
                                />
                            </div>
                        )}

                        {/* 규칙 빌더 탭 */}
                        {subTab === 'builder' && (
                            <div className="rule-builder-container">
                                {/* 파싱 에러 표시 */}
                                {parseErrors.length > 0 && (
                                    <div className="protocol-options" style={{ marginBottom: '16px', borderColor: '#e74c3c' }}>
                                        <div className="protocol-options-title" style={{ color: '#e74c3c' }}>
                                            파싱 오류 ({parseErrors.length}개)
                                        </div>
                                        <ul style={{ fontSize: '0.8rem', color: '#e74c3c', paddingLeft: '20px' }}>
                                            {parseErrors.slice(0, 5).map((err, i) => (
                                                <li key={i}>{err}</li>
                                            ))}
                                            {parseErrors.length > 5 && (
                                                <li>... 외 {parseErrors.length - 5}개</li>
                                            )}
                                        </ul>
                                    </div>
                                )}

                                {/* 규칙 테이블 */}
                                <RuleTable
                                    rules={rules}
                                    onDelete={handleDeleteRule}
                                    onEdit={handleEditRule}
                                />

                                {/* 규칙 폼 타입 선택 탭 (fyne 스타일) */}
                                <div className="sub-tabs" style={{ marginTop: '16px' }}>
                                    <button
                                        className={`sub-tab-btn ${ruleFormType === 'general' ? 'active' : ''}`}
                                        onClick={() => {
                                            setRuleFormType('general');
                                            setEditRule(null);
                                            setEditIndex(undefined);
                                        }}
                                    >
                                        일반 규칙
                                    </button>
                                    <button
                                        className={`sub-tab-btn ${ruleFormType === 'blackwhite' ? 'active' : ''}`}
                                        onClick={() => {
                                            setRuleFormType('blackwhite');
                                            setEditRule(null);
                                            setEditIndex(undefined);
                                        }}
                                    >
                                        Black/White
                                    </button>
                                </div>

                                {/* 일반 규칙 폼 */}
                                {ruleFormType === 'general' && (
                                    <RuleForm
                                        onAdd={handleAddRule}
                                        editRule={editRule && !editRule.black && !editRule.white ? editRule : null}
                                        editIndex={editRule && !editRule.black && !editRule.white ? editIndex : undefined}
                                        onUpdate={handleUpdateRule}
                                        onCancel={handleCancelEdit}
                                    />
                                )}

                                {/* Black/White 폼 */}
                                {ruleFormType === 'blackwhite' && (
                                    <BlackWhiteForm
                                        onAdd={handleAddRule}
                                        editRule={editRule && (editRule.black || editRule.white) ? editRule : null}
                                        editIndex={editRule && (editRule.black || editRule.white) ? editIndex : undefined}
                                        onUpdate={handleUpdateRule}
                                        onCancel={handleCancelEdit}
                                    />
                                )}
                            </div>
                        )}

                        {/* NAT 규칙 탭 */}
                        {subTab === 'nat' && (
                            <div className="rule-builder-container">
                                {/* 파싱 에러 표시 */}
                                {natParseErrors.length > 0 && (
                                    <div className="protocol-options" style={{ marginBottom: '16px', borderColor: '#e74c3c' }}>
                                        <div className="protocol-options-title" style={{ color: '#e74c3c' }}>
                                            파싱 오류 ({natParseErrors.length}개)
                                        </div>
                                        <ul style={{ fontSize: '0.8rem', color: '#e74c3c', paddingLeft: '20px' }}>
                                            {natParseErrors.slice(0, 5).map((err, i) => (
                                                <li key={i}>{err}</li>
                                            ))}
                                            {natParseErrors.length > 5 && (
                                                <li>... 외 {natParseErrors.length - 5}개</li>
                                            )}
                                        </ul>
                                    </div>
                                )}

                                {/* NAT 규칙 테이블 */}
                                <NATTable
                                    rules={natRules}
                                    onDelete={handleDeleteNatRule}
                                    onEdit={handleEditNatRule}
                                />

                                {/* NAT 폼 타입 선택 */}
                                <div className="sub-tabs" style={{ marginTop: '16px' }}>
                                    <button
                                        className={`sub-tab-btn ${natFormType === 'dnat' ? 'active' : ''}`}
                                        onClick={() => {
                                            setNatFormType('dnat');
                                            setEditNatRule(null);
                                            setEditNatIndex(undefined);
                                        }}
                                    >
                                        DNAT (포트포워딩)
                                    </button>
                                    <button
                                        className={`sub-tab-btn ${natFormType === 'snat' ? 'active' : ''}`}
                                        onClick={() => {
                                            setNatFormType('snat');
                                            setEditNatRule(null);
                                            setEditNatIndex(undefined);
                                        }}
                                    >
                                        SNAT / MASQUERADE
                                    </button>
                                </div>

                                {/* DNAT 폼 */}
                                {natFormType === 'dnat' && (
                                    <DNATForm
                                        onAdd={handleAddNatRule}
                                        editRule={editNatRule?.natType === 0 ? editNatRule : null}
                                        editIndex={editNatRule?.natType === 0 ? editNatIndex : undefined}
                                        onUpdate={handleUpdateNatRule}
                                        onCancel={handleCancelNatEdit}
                                    />
                                )}

                                {/* SNAT 폼 */}
                                {natFormType === 'snat' && (
                                    <SNATForm
                                        onAdd={handleAddNatRule}
                                        editRule={editNatRule?.natType !== 0 ? editNatRule : null}
                                        editIndex={editNatRule?.natType !== 0 ? editNatIndex : undefined}
                                        onUpdate={handleUpdateNatRule}
                                        onCancel={handleCancelNatEdit}
                                    />
                                )}
                            </div>
                        )}
                    </>
                ) : (
                    <div className="empty-state">
                        <div className="empty-state-icon">📋</div>
                        <p>왼쪽에서 템플릿을 선택하거나</p>
                        <p>새 템플릿을 생성하세요</p>
                    </div>
                )}
            </div>
        </div>
    );
});

export default TemplateTab;
