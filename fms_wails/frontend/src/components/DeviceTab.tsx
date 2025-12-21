import { useState, useEffect, forwardRef, useImperativeHandle } from 'react';
import {
    GetAllFirewalls,
    SaveFirewall,
    DeleteFirewall,
    CheckAllServerStatus,
    CheckServerStatus,
    GetAllTemplates,
    Deploy,
    ConfirmDialog
} from '../../wailsjs/go/main/App';

// 간소화된 Firewall 인터페이스 (HTTP 방식)
interface Firewall {
    index: number;
    deviceName: string;
    serverStatus: string;
    deployStatus: string;
    version: string;
}

interface Template {
    version: string;
    contents: string;
}

export interface DeviceTabRef {
    refresh: () => void;
}

const DeviceTab = forwardRef<DeviceTabRef>((_, ref) => {
    const [firewalls, setFirewalls] = useState<Firewall[]>([]);
    const [templates, setTemplates] = useState<Template[]>([]);
    const [selectedIndexes, setSelectedIndexes] = useState<number[]>([]);
    const [showModal, setShowModal] = useState(false);
    const [showDeployModal, setShowDeployModal] = useState(false);
    const [selectedTemplate, setSelectedTemplate] = useState('');
    const [editingFirewall, setEditingFirewall] = useState<Firewall | null>(null);
    const [isDeploying, setIsDeploying] = useState(false);
    const [isChecking, setIsChecking] = useState(false);

    const emptyFirewall: Firewall = {
        index: 0,
        deviceName: '',
        serverStatus: '-',
        deployStatus: '-',
        version: '-'
    };

    useEffect(() => {
        loadFirewalls();
        loadTemplates();
    }, []);

    const loadFirewalls = async () => {
        const data = await GetAllFirewalls();
        setFirewalls((data || []) as Firewall[]);
    };

    const loadTemplates = async () => {
        const data = await GetAllTemplates();
        setTemplates((data || []) as Template[]);
    };

    // 부모 컴포넌트에서 호출할 수 있도록 refresh 메서드 노출
    useImperativeHandle(ref, () => ({
        refresh: () => {
            loadFirewalls();
            loadTemplates();
            setSelectedIndexes([]);
            setShowModal(false);
            setShowDeployModal(false);
            setEditingFirewall(null);
        }
    }));

    const handleAdd = () => {
        setEditingFirewall({ ...emptyFirewall });
        setShowModal(true);
    };

    const handleEdit = (fw: Firewall) => {
        setEditingFirewall({ ...fw });
        setShowModal(true);
    };

    const handleSave = async () => {
        if (!editingFirewall) return;
        if (!editingFirewall.deviceName.trim()) {
            alert('장비명(IP)을 입력하세요.');
            return;
        }
        await SaveFirewall(JSON.stringify(editingFirewall));
        await loadFirewalls();
        setShowModal(false);
        setEditingFirewall(null);
    };

    const handleDelete = async () => {
        if (selectedIndexes.length === 0) {
            alert('삭제할 장비를 선택하세요.');
            return;
        }
        const result = await ConfirmDialog('삭제 확인', `${selectedIndexes.length}개 장비를 삭제하시겠습니까?`);
        if (result !== '확인') return;

        for (const idx of selectedIndexes) {
            await DeleteFirewall(idx);
        }
        await loadFirewalls();
        setSelectedIndexes([]);
    };

    const handleCheckStatus = async () => {
        setIsChecking(true);
        try {
            if (selectedIndexes.length === 0) {
                // 선택된 장비가 없으면 전체 확인
                await CheckAllServerStatus();
            } else {
                // 선택된 장비만 확인
                for (const idx of selectedIndexes) {
                    await CheckServerStatus(idx);
                }
            }
            await loadFirewalls();
        } finally {
            setIsChecking(false);
        }
    };

    const handleToggleSelect = (index: number) => {
        setSelectedIndexes((prev) =>
            prev.includes(index)
                ? prev.filter((i) => i !== index)
                : [...prev, index]
        );
    };

    const handleSelectAll = (e: React.ChangeEvent<HTMLInputElement>) => {
        if (e.target.checked) {
            // 체크가 되었을 때만 전체 선택
            setSelectedIndexes(firewalls.map((f) => f.index));
        } else {
            // 체크 해제 시 전체 해제
            setSelectedIndexes([]);
        }
    };

    const handleDeployClick = async () => {
        if (selectedIndexes.length === 0) {
            alert('배포할 장비를 선택하세요.');
            return;
        }
        // 배포 시 최신 템플릿 목록 로드
        const latestTemplates = await GetAllTemplates();
        const templateList = (latestTemplates || []) as Template[];
        setTemplates(templateList);

        if (templateList.length === 0) {
            alert('배포할 템플릿이 없습니다.');
            return;
        }
        setSelectedTemplate(templateList[0]?.version || '');
        setShowDeployModal(true);
    };

    const handleDeploy = async () => {
        if (!selectedTemplate) {
            alert('템플릿을 선택하세요.');
            return;
        }

        setIsDeploying(true);
        setShowDeployModal(false);

        let successCount = 0;
        let failCount = 0;

        for (const idx of selectedIndexes) {
            try {
                await Deploy(idx, selectedTemplate);
                successCount++;
            } catch (e) {
                console.error(`배포 실패 (장비 ${idx}):`, e);
                failCount++;
            }
        }

        await loadFirewalls();
        setIsDeploying(false);

        if (failCount === 0) {
            alert(`${successCount}개 장비에 배포가 완료되었습니다.`);
        } else {
            alert(`배포 완료: 성공 ${successCount}개, 실패 ${failCount}개`);
        }
    };

    const getStatusBadge = (status: string) => {
        if (status === 'running' || status === 'success') {
            return <span className="badge badge-success">{status === 'running' ? '정상' : '성공'}</span>;
        } else if (status === 'stop' || status === 'fail') {
            return <span className="badge badge-danger">{status === 'stop' ? '정지' : '실패'}</span>;
        } else if (status === 'error') {
            return <span className="badge badge-warning">확인요망</span>;
        }
        return <span className="badge badge-info">{status || '-'}</span>;
    };

    return (
        <div>
            <div className="card">
                <div className="card-title">장비 목록</div>

                <div className="btn-group" style={{ marginBottom: '16px' }}>
                    <button className="btn btn-primary" onClick={handleAdd} disabled={isDeploying || isChecking}>
                        + 장비 추가
                    </button>
                    <button className="btn btn-secondary" onClick={handleCheckStatus} disabled={isDeploying || isChecking}>
                        {isChecking ? '확인중...' : '상태 확인'}
                    </button>
                    <button className="btn btn-primary" onClick={handleDeployClick} disabled={isDeploying || isChecking}>
                        {isDeploying ? '배포 중...' : '배포'}
                    </button>
                    <button className="btn btn-danger" onClick={handleDelete} disabled={isDeploying || isChecking}>
                        삭제
                    </button>
                </div>

                <table className="table">
                    <thead>
                        <tr>
                            <th>
                                <input
                                    type="checkbox"
                                    checked={selectedIndexes.length === firewalls.length && firewalls.length > 0}
                                    onChange={(e) => handleSelectAll(e)}
                                />
                            </th>
                            <th>장비명(IP)</th>
                            <th>서버상태</th>
                            <th>배포상태</th>
                            <th>버전</th>
                            <th>작업</th>
                        </tr>
                    </thead>
                    <tbody>
                        {firewalls.length === 0 ? (
                            <tr>
                                <td colSpan={6} style={{ textAlign: 'center', color: '#666' }}>
                                    등록된 장비가 없습니다
                                </td>
                            </tr>
                        ) : (
                            firewalls.map((fw) => (
                                <tr key={fw.index}>
                                    <td>
                                        <input
                                            type="checkbox"
                                            checked={selectedIndexes.includes(fw.index)}
                                            onChange={() => handleToggleSelect(fw.index)}
                                        />
                                    </td>
                                    <td>{fw.deviceName}</td>
                                    <td>{getStatusBadge(fw.serverStatus)}</td>
                                    <td>{getStatusBadge(fw.deployStatus)}</td>
                                    <td>{fw.version || '-'}</td>
                                    <td>
                                        <button
                                            className="btn btn-secondary"
                                            onClick={() => handleEdit(fw)}
                                            style={{ padding: '4px 12px' }}
                                        >
                                            편집
                                        </button>
                                    </td>
                                </tr>
                            ))
                        )}
                    </tbody>
                </table>
            </div>

            {/* 장비 편집 모달 - 간소화 (IP만 입력) */}
            {showModal && editingFirewall && (
                <div className="modal-overlay" onClick={() => setShowModal(false)}>
                    <div className="modal" onClick={(e) => e.stopPropagation()}>
                        <div className="modal-header">
                            <h3 className="modal-title">
                                {editingFirewall.index ? '장비 편집' : '장비 추가'}
                            </h3>
                            <button className="modal-close" onClick={() => setShowModal(false)}>
                                ×
                            </button>
                        </div>

                        <div className="form-group">
                            <label>장비명(IP)</label>
                            <input
                                type="text"
                                className="input"
                                value={editingFirewall.deviceName}
                                onChange={(e) =>
                                    setEditingFirewall({ ...editingFirewall, deviceName: e.target.value })
                                }
                                placeholder="192.168.1.100"
                            />
                        </div>

                        <div className="modal-footer">
                            <button className="btn btn-secondary" onClick={() => setShowModal(false)}>
                                취소
                            </button>
                            <button className="btn btn-primary" onClick={handleSave}>
                                저장
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {/* 배포 모달 */}
            {showDeployModal && (
                <div className="modal-overlay" onClick={() => setShowDeployModal(false)}>
                    <div className="modal" onClick={(e) => e.stopPropagation()}>
                        <div className="modal-header">
                            <h3 className="modal-title">템플릿 배포</h3>
                            <button className="modal-close" onClick={() => setShowDeployModal(false)}>
                                ×
                            </button>
                        </div>

                        <p style={{ marginBottom: '16px' }}>
                            {selectedIndexes.length}개 장비에 배포할 템플릿을 선택하세요.
                        </p>

                        <div className="form-group">
                            <label>템플릿 선택</label>
                            <select
                                className="input"
                                value={selectedTemplate}
                                onChange={(e) => setSelectedTemplate(e.target.value)}
                            >
                                {templates.map((t) => (
                                    <option key={t.version} value={t.version}>
                                        {t.version}
                                    </option>
                                ))}
                            </select>
                        </div>

                        <div className="modal-footer">
                            <button className="btn btn-secondary" onClick={() => setShowDeployModal(false)}>
                                취소
                            </button>
                            <button className="btn btn-primary" onClick={handleDeploy}>
                                배포 실행
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
});

export default DeviceTab;
