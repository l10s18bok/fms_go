import { useState, useEffect, forwardRef, useImperativeHandle } from 'react';
import { GetAllHistory, DeleteHistory, ConfirmDialog } from '../../wailsjs/go/main/App';

// Go model과 동일한 구조
interface RuleResult {
    rule: string;
    text: string;
    status: string;  // ok/error/unfind/validation
    reason: string;
}

interface DeployHistory {
    id: number;
    timestamp: string;       // Go time.Time은 JSON으로 문자열 변환
    deviceIp: string;
    templateVersion: string;
    status: string;          // success/fail/error
    results: RuleResult[];
}

export interface HistoryTabRef {
    refresh: () => void;
}

const HistoryTab = forwardRef<HistoryTabRef>((_, ref) => {
    const [history, setHistory] = useState<DeployHistory[]>([]);
    const [selectedHistory, setSelectedHistory] = useState<DeployHistory | null>(null);

    useEffect(() => {
        loadHistory();
    }, []);

    const loadHistory = async () => {
        const data = await GetAllHistory();
        // 최신순 정렬
        const sorted = (data || []).sort((a: DeployHistory, b: DeployHistory) =>
            new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()
        );
        setHistory(sorted as DeployHistory[]);
    };

    // 부모 컴포넌트에서 호출할 수 있도록 refresh 메서드 노출
    useImperativeHandle(ref, () => ({
        refresh: () => {
            loadHistory();
            setSelectedHistory(null);
        }
    }));

    const handleDelete = async (id: number) => {
        const result = await ConfirmDialog('삭제 확인', '이 배포 이력을 삭제하시겠습니까?');
        // Windows에서는 "Yes", "예", "확인" 등 다양한 값이 반환될 수 있음
        if (result !== '확인' && result !== 'Yes' && result !== '예') return;
        await DeleteHistory(id);
        await loadHistory();
        if (selectedHistory?.id === id) {
            setSelectedHistory(null);
        }
    };

    const handleDeleteAll = async () => {
        if (history.length === 0) {
            return;
        }
        const result = await ConfirmDialog('전체 삭제', `${history.length}개의 배포 이력을 모두 삭제하시겠습니까?`);
        // Windows에서는 "Yes", "예", "확인" 등 다양한 값이 반환될 수 있음
        if (result !== '확인' && result !== 'Yes' && result !== '예') return;

        for (const h of history) {
            await DeleteHistory(h.id);
        }
        await loadHistory();
        setSelectedHistory(null);
    };

    const formatDate = (dateStr: string) => {
        const date = new Date(dateStr);
        return date.toLocaleString('ko-KR');
    };

    // 배포 상태 배지
    const getStatusBadge = (status: string) => {
        if (status === 'success') {
            return <span className="badge badge-success">성공</span>;
        } else if (status === 'fail') {
            return <span className="badge badge-danger">실패</span>;
        } else if (status === 'error') {
            return <span className="badge badge-warning">오류</span>;
        }
        return <span className="badge badge-info">{status || '-'}</span>;
    };

    // 규칙 결과 배지
    const getRuleStatusBadge = (status: string) => {
        const lowerStatus = status?.toLowerCase() || '';
        if (lowerStatus === 'ok') {
            return <span className="badge badge-success">성공</span>;
        } else if (lowerStatus === 'error' || lowerStatus === 'unfind' || lowerStatus === 'validation') {
            return <span className="badge badge-danger">실패</span>;
        } else if (lowerStatus === 'write') {
            return <span className="badge badge-warning">진행중</span>;
        }
        return <span className="badge badge-info">{status || '-'}</span>;
    };

    // 규칙 결과 통계 계산
    const getResultStats = (results: RuleResult[]) => {
        if (!results) return { total: 0, success: 0, fail: 0 };
        const total = results.length;
        const success = results.filter(r => r.status.toLowerCase() === 'ok').length;
        const fail = total - success;
        return { total, success, fail };
    };

    return (
        <div className="split-layout">
            {/* 왼쪽: 이력 목록 */}
            <div className="card">
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
                    <div className="card-title" style={{ marginBottom: 0 }}>배포 이력</div>
                    <button
                        className="btn btn-danger"
                        onClick={handleDeleteAll}
                        style={{ padding: '4px 12px', fontSize: '0.85rem' }}
                        disabled={history.length === 0}
                    >
                        전체 삭제
                    </button>
                </div>
                <ul className="list">
                    {history.length === 0 ? (
                        <li className="list-item" style={{ color: '#666' }}>
                            배포 이력이 없습니다
                        </li>
                    ) : (
                        history.map((h) => (
                            <li
                                key={h.id}
                                className={`list-item ${selectedHistory?.id === h.id ? 'active' : ''}`}
                                onClick={() => setSelectedHistory(h)}
                            >
                                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                                    <div>
                                        <div style={{ fontWeight: 500 }}>{h.deviceIp}</div>
                                        <div style={{ fontSize: '0.8rem', color: '#888' }}>
                                            {formatDate(h.timestamp)}
                                        </div>
                                    </div>
                                    {getStatusBadge(h.status)}
                                </div>
                            </li>
                        ))
                    )}
                </ul>
            </div>

            {/* 오른쪽: 상세 정보 */}
            <div className="card">
                <div className="card-title">상세 정보</div>

                {selectedHistory ? (
                    <>
                        <div style={{ marginBottom: '20px' }}>
                            <table className="table">
                                <tbody>
                                    <tr>
                                        <th style={{ width: '120px' }}>장비 IP</th>
                                        <td>{selectedHistory.deviceIp}</td>
                                    </tr>
                                    <tr>
                                        <th>템플릿 버전</th>
                                        <td>{selectedHistory.templateVersion}</td>
                                    </tr>
                                    <tr>
                                        <th>배포 시간</th>
                                        <td>{formatDate(selectedHistory.timestamp)}</td>
                                    </tr>
                                    <tr>
                                        <th>상태</th>
                                        <td>{getStatusBadge(selectedHistory.status)}</td>
                                    </tr>
                                    <tr>
                                        <th>결과</th>
                                        <td>
                                            {(() => {
                                                const stats = getResultStats(selectedHistory.results);
                                                return `총 ${stats.total}개 / 성공 ${stats.success}개 / 실패 ${stats.fail}개`;
                                            })()}
                                        </td>
                                    </tr>
                                </tbody>
                            </table>
                        </div>

                        {selectedHistory.results && selectedHistory.results.length > 0 && (
                            <div>
                                <h4 style={{ marginBottom: '12px', color: '#e94560' }}>규칙별 결과</h4>
                                <div style={{ maxHeight: '300px', overflowY: 'auto' }}>
                                    <table className="table">
                                        <thead>
                                            <tr>
                                                <th>규칙</th>
                                                <th style={{ width: '80px' }}>결과</th>
                                                <th>사유</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            {selectedHistory.results.map((r, idx) => (
                                                <tr key={idx}>
                                                    <td style={{ fontFamily: 'monospace', fontSize: '0.85rem', wordBreak: 'break-all' }}>
                                                        {r.rule}
                                                    </td>
                                                    <td>
                                                        {getRuleStatusBadge(r.status)}
                                                    </td>
                                                    <td>{r.reason || '-'}</td>
                                                </tr>
                                            ))}
                                        </tbody>
                                    </table>
                                </div>
                            </div>
                        )}

                        <div style={{ marginTop: '20px' }}>
                            <button
                                className="btn btn-danger"
                                onClick={() => handleDelete(selectedHistory.id)}
                            >
                                이력 삭제
                            </button>
                        </div>
                    </>
                ) : (
                    <div className="empty-state">
                        <div className="empty-state-icon">📜</div>
                        <p>왼쪽에서 배포 이력을 선택하세요</p>
                    </div>
                )}
            </div>
        </div>
    );
});

export default HistoryTab;
