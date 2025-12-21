import { useState, useEffect, forwardRef, useImperativeHandle } from 'react';
import {
    GetAllTemplates,
    GetTemplate,
    SaveTemplate,
    DeleteTemplate,
    ConfirmDialog
} from '../../wailsjs/go/main/App';

interface Template {
    version: string;
    contents: string;
}

export interface TemplateTabRef {
    refresh: () => void;
}

const TemplateTab = forwardRef<TemplateTabRef>((_, ref) => {
    const [templates, setTemplates] = useState<Template[]>([]);
    const [selectedVersion, setSelectedVersion] = useState<string>('');
    const [version, setVersion] = useState('');
    const [contents, setContents] = useState('');
    const [isNew, setIsNew] = useState(false);

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
        }
    }));

    const handleSelect = async (ver: string) => {
        setSelectedVersion(ver);
        setIsNew(false);
        const template = await GetTemplate(ver);
        if (template) {
            setVersion(template.version);
            setContents(template.contents);
        }
    };

    const handleNew = () => {
        setSelectedVersion('');
        setIsNew(true);
        setVersion('');
        setContents('');
    };

    const handleSave = async () => {
        if (!version.trim()) {
            alert('버전을 입력하세요.');
            return;
        }
        await SaveTemplate(version, contents);
        await loadTemplates();
        setSelectedVersion(version);
        setIsNew(false);
    };

    const handleDelete = async () => {
        if (!selectedVersion) return;
        const result = await ConfirmDialog('삭제 확인', `"${selectedVersion}" 템플릿을 삭제하시겠습니까?`);
        if (result !== '확인') return;

        await DeleteTemplate(selectedVersion);
        await loadTemplates();
        setSelectedVersion('');
        setVersion('');
        setContents('');
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
                <div className="card-title">
                    {isNew ? '새 템플릿' : selectedVersion ? `템플릿: ${selectedVersion}` : '템플릿 선택'}
                </div>

                {(selectedVersion || isNew) ? (
                    <>
                        <div className="form-group">
                            <label>버전</label>
                            <input
                                type="text"
                                className="input"
                                value={version}
                                onChange={(e) => setVersion(e.target.value)}
                                placeholder="예: v1.0.0"
                                disabled={!isNew}
                            />
                        </div>

                        <div className="form-group">
                            <label>규칙 내용</label>
                            <textarea
                                className="textarea"
                                value={contents}
                                onChange={(e) => setContents(e.target.value)}
                                placeholder="방화벽 규칙을 입력하세요..."
                                style={{ minHeight: '300px' }}
                            />
                        </div>

                        <div className="btn-group">
                            <button className="btn btn-primary" onClick={handleSave}>
                                저장
                            </button>
                            {!isNew && (
                                <button className="btn btn-danger" onClick={handleDelete}>
                                    삭제
                                </button>
                            )}
                        </div>
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
