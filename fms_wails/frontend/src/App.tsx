import { useState, useRef, useEffect, useCallback } from 'react';
import './App.css';
import TemplateTab, { TemplateTabRef } from './components/TemplateTab';
import DeviceTab, { DeviceTabRef } from './components/DeviceTab';
import HistoryTab, { HistoryTabRef } from './components/HistoryTab';
import {
    AreaChart,
    Area,
    XAxis,
    YAxis,
    CartesianGrid,
    Tooltip,
    ResponsiveContainer,
    Legend
} from 'recharts';
import {
    GetConfig,
    SaveConfig,
    GetConfigDir,
    ResetAll,
    GetAllTemplates,
    GetAllFirewalls,
    GetAllHistory,
    SaveTemplate,
    SaveFirewall,
    SaveHistory,
    SaveFileDialog,
    WriteFileContent,
    ConfirmDialog,
    AlertDialog,
    GetAppVersion
} from '../wailsjs/go/main/App';

type TabType = 'template' | 'device' | 'history';
type MenuType = 'file' | 'tools' | 'help' | null;

// Config 인터페이스
interface Config {
    connectionMode: string;
    agentServerURL: string;
    timeoutSeconds: number;
}

function App() {
    const [activeTab, setActiveTab] = useState<TabType>('template');
    const [activeMenu, setActiveMenu] = useState<MenuType>(null);
    const [showSettingsModal, setShowSettingsModal] = useState(false);
    const [showHelpModal, setShowHelpModal] = useState(false);
    const [config, setConfig] = useState<Config>({
        connectionMode: 'agent',
        agentServerURL: 'http://172.24.10.6:8080',
        timeoutSeconds: 10
    });
    const [configDir, setConfigDir] = useState('');
    const [appVersion, setAppVersion] = useState('');
    const [showChartModal, setShowChartModal] = useState(false);

    // 메뉴 외부 클릭 시 닫기
    const closeMenu = useCallback(() => {
        setActiveMenu(null);
    }, []);

    useEffect(() => {
        if (activeMenu) {
            document.addEventListener('click', closeMenu);
            return () => document.removeEventListener('click', closeMenu);
        }
    }, [activeMenu, closeMenu]);

    // 차트 데모 데이터 - 월별 배포 통계 (임의의 값)
    const chartData = [
        { month: '1월', 성공: 45, 실패: 12, 대기: 8 },
        { month: '2월', 성공: 52, 실패: 8, 대기: 15 },
        { month: '3월', 성공: 78, 실패: 5, 대기: 10 },
        { month: '4월', 성공: 65, 실패: 15, 대기: 12 },
        { month: '5월', 성공: 89, 실패: 3, 대기: 5 },
        { month: '6월', 성공: 95, 실패: 7, 대기: 8 },
        { month: '7월', 성공: 110, 실패: 4, 대기: 6 },
        { month: '8월', 성공: 85, 실패: 10, 대기: 12 },
        { month: '9월', 성공: 102, 실패: 6, 대기: 9 },
        { month: '10월', 성공: 120, 실패: 8, 대기: 7 },
        { month: '11월', 성공: 98, 실패: 5, 대기: 11 },
        { month: '12월', 성공: 130, 실패: 3, 대기: 4 },
    ];

    // 앱 버전 로드
    useEffect(() => {
        GetAppVersion().then(setAppVersion);
    }, []);

    // 각 탭의 ref
    const templateTabRef = useRef<TemplateTabRef>(null);
    const deviceTabRef = useRef<DeviceTabRef>(null);
    const historyTabRef = useRef<HistoryTabRef>(null);

    // Import 파일 입력 ref
    const importInputRef = useRef<HTMLInputElement>(null);

    // Import 처리
    const handleImport = () => {
        importInputRef.current?.click();
    };

    const handleImportFile = async (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (!file) return;

        try {
            const text = await file.text();
            const data = JSON.parse(text);

            // 현재 탭에 따라 데이터 타입 확인 및 저장
            let importedCount = 0;

            if (activeTab === 'template') {
                if (!Array.isArray(data)) {
                    alert('유효한 템플릿 데이터가 아닙니다.');
                    return;
                }
                let skippedCount = 0;
                for (const item of data) {
                    if (item.version && item.contents && item.contents.trim()) {
                        try {
                            await SaveTemplate(item.version, item.contents);
                            importedCount++;
                        } catch (err) {
                            console.error(`템플릿 저장 실패: ${item.version}`, err);
                            skippedCount++;
                        }
                    } else {
                        skippedCount++;
                    }
                }
                templateTabRef.current?.refresh();
                if (skippedCount > 0) {
                    console.log(`${skippedCount}개 템플릿이 유효하지 않아 건너뛰었습니다.`);
                }
            } else if (activeTab === 'device') {
                if (!Array.isArray(data)) {
                    alert('유효한 장비 데이터가 아닙니다.');
                    return;
                }
                for (const item of data) {
                    if (item.deviceName) {
                        await SaveFirewall(JSON.stringify(item));
                        importedCount++;
                    }
                }
                deviceTabRef.current?.refresh();
            } else if (activeTab === 'history') {
                if (!Array.isArray(data)) {
                    alert('유효한 배포 이력 데이터가 아닙니다.');
                    return;
                }
                for (const item of data) {
                    if (item.deviceIp && item.templateVersion) {
                        await SaveHistory(JSON.stringify(item));
                        importedCount++;
                    }
                }
                historyTabRef.current?.refresh();
            }

            if (importedCount > 0) {
                alert(`${importedCount}개 항목을 가져왔습니다.`);
            } else {
                alert('가져올 수 있는 유효한 데이터가 없습니다.');
            }
        } catch (err) {
            alert('파일을 읽는 중 오류가 발생했습니다.');
            console.error(err);
        }

        // 입력 초기화
        e.target.value = '';
    };

    // Export 처리 (네이티브 다이얼로그 사용)
    const handleExport = async () => {
        let data: unknown[] = [];
        let filename = '';

        if (activeTab === 'template') {
            data = await GetAllTemplates();
            filename = 'templates.json';
        } else if (activeTab === 'device') {
            data = await GetAllFirewalls();
            filename = 'firewalls.json';
        } else if (activeTab === 'history') {
            data = await GetAllHistory();
            filename = 'history.json';
        }

        if (!data || data.length === 0) {
            alert('내보낼 데이터가 없습니다.');
            return;
        }

        try {
            const filePath = await SaveFileDialog('파일 내보내기', filename);
            if (!filePath) return;

            await WriteFileContent(filePath, JSON.stringify(data, null, 2));
            alert('파일이 저장되었습니다.');
        } catch (err) {
            alert('파일 저장 중 오류가 발생했습니다.');
            console.error(err);
        }
    };

    // Reset 처리
    const handleReset = async () => {
        const result = await ConfirmDialog('초기화', '모든 데이터(템플릿, 장비, 배포이력)를 초기화하시겠습니까?');
        // Windows에서는 "Yes", "예", "확인" 등 다양한 값이 반환될 수 있음
        if (result !== '확인' && result !== 'Yes' && result !== '예') {
            return;
        }

        try {
            await ResetAll();
            templateTabRef.current?.refresh();
            deviceTabRef.current?.refresh();
            historyTabRef.current?.refresh();
            await AlertDialog('완료', '모든 데이터가 초기화되었습니다.');
        } catch (err) {
            await AlertDialog('오류', '초기화 중 오류가 발생했습니다.');
            console.error(err);
        }
    };

    // 설정 다이얼로그 열기
    const handleOpenSettings = async () => {
        try {
            const cfg = await GetConfig();
            const dir = await GetConfigDir();
            setConfig(cfg as Config);
            setConfigDir(dir);
            setShowSettingsModal(true);
        } catch (err) {
            console.error(err);
        }
    };

    // 설정 저장
    const handleSaveConfig = async () => {
        // 유효성 검사
        if (config.connectionMode === 'agent' && !config.agentServerURL) {
            alert('Agent Server URL을 입력해주세요.');
            return;
        }
        if (config.timeoutSeconds < 5 || config.timeoutSeconds > 120) {
            alert('타임아웃은 5~120 사이의 숫자를 입력해주세요.');
            return;
        }

        try {
            await SaveConfig(JSON.stringify(config));
            setShowSettingsModal(false);
            alert('설정이 저장되었습니다.');
        } catch (err) {
            alert('설정 저장 중 오류가 발생했습니다.');
            console.error(err);
        }
    };

    // 메뉴 토글
    const toggleMenu = (menu: MenuType, e: React.MouseEvent) => {
        e.stopPropagation();
        setActiveMenu(activeMenu === menu ? null : menu);
    };

    return (
        <div id="App">
            {/* 상단 메뉴바 */}
            <header className="app-menubar">
                <div className="menubar-left">
                    {/* 파일 메뉴 */}
                    <div className="menu-item">
                        <button className="menu-btn" onClick={(e) => toggleMenu('file', e)}>
                            파일
                        </button>
                        {activeMenu === 'file' && (
                            <div className="menu-dropdown">
                                <button className="menu-dropdown-item" onClick={() => { handleImport(); closeMenu(); }}>
                                    Import
                                </button>
                                <button className="menu-dropdown-item" onClick={() => { handleExport(); closeMenu(); }}>
                                    Export
                                </button>
                                <div className="menu-divider" />
                                <button className="menu-dropdown-item danger" onClick={() => { handleReset(); closeMenu(); }}>
                                    Reset
                                </button>
                            </div>
                        )}
                    </div>

                    {/* 도구 메뉴 */}
                    <div className="menu-item">
                        <button className="menu-btn" onClick={(e) => toggleMenu('tools', e)}>
                            도구
                        </button>
                        {activeMenu === 'tools' && (
                            <div className="menu-dropdown">
                                <button className="menu-dropdown-item" onClick={() => { handleOpenSettings(); closeMenu(); }}>
                                    설정
                                </button>
                            </div>
                        )}
                    </div>

                    {/* 도움말 메뉴 */}
                    <div className="menu-item">
                        <button className="menu-btn" onClick={(e) => toggleMenu('help', e)}>
                            도움말
                        </button>
                        {activeMenu === 'help' && (
                            <div className="menu-dropdown">
                                <button className="menu-dropdown-item" onClick={() => { setShowHelpModal(true); closeMenu(); }}>
                                    도움말
                                </button>
                            </div>
                        )}
                    </div>
                </div>
            </header>

            {/* 숨겨진 파일 입력 */}
            <input
                type="file"
                ref={importInputRef}
                style={{ display: 'none' }}
                accept=".json"
                onChange={handleImportFile}
            />

            {/* 탭 네비게이션 */}
            <nav className="tab-nav">
                <button
                    className={`tab-btn ${activeTab === 'template' ? 'active' : ''}`}
                    onClick={() => setActiveTab('template')}
                >
                    템플릿 관리
                </button>
                <button
                    className={`tab-btn ${activeTab === 'device' ? 'active' : ''}`}
                    onClick={() => setActiveTab('device')}
                >
                    장비 관리
                </button>
                <button
                    className={`tab-btn ${activeTab === 'history' ? 'active' : ''}`}
                    onClick={() => setActiveTab('history')}
                >
                    배포 이력
                </button>
            </nav>

            {/* 탭 컨텐츠 */}
            <main className="tab-content">
                <div style={{ display: activeTab === 'template' ? 'block' : 'none', height: '100%' }}>
                    <TemplateTab ref={templateTabRef} />
                </div>
                <div style={{ display: activeTab === 'device' ? 'block' : 'none', height: '100%' }}>
                    <DeviceTab ref={deviceTabRef} onDeployComplete={() => historyTabRef.current?.refresh()} />
                </div>
                <div style={{ display: activeTab === 'history' ? 'block' : 'none', height: '100%' }}>
                    <HistoryTab ref={historyTabRef} />
                </div>
            </main>

            {/* 하단 상태바 */}
            <footer className="app-footer">
                <span
                    className="app-version app-version-clickable"
                    onClick={() => setShowChartModal(true)}
                    title="차트 데모 보기"
                >
                    FMS v{appVersion}
                </span>
            </footer>

            {/* 차트 데모 모달 */}
            {showChartModal && (
                <div className="modal-overlay" onClick={() => setShowChartModal(false)}>
                    <div className="modal modal-chart" onClick={(e) => e.stopPropagation()}>
                        <div className="modal-header">
                            <h3 className="modal-title">📊 월별 배포 통계 (Demo)</h3>
                            <button className="modal-close" onClick={() => setShowChartModal(false)}>
                                ×
                            </button>
                        </div>

                        <div className="chart-container">
                            <ResponsiveContainer width="100%" height={350}>
                                <AreaChart
                                    data={chartData}
                                    margin={{ top: 20, right: 30, left: 0, bottom: 0 }}
                                >
                                    <defs>
                                        <linearGradient id="colorSuccess" x1="0" y1="0" x2="0" y2="1">
                                            <stop offset="5%" stopColor="#27ae60" stopOpacity={0.8}/>
                                            <stop offset="95%" stopColor="#27ae60" stopOpacity={0.1}/>
                                        </linearGradient>
                                        <linearGradient id="colorFail" x1="0" y1="0" x2="0" y2="1">
                                            <stop offset="5%" stopColor="#e94560" stopOpacity={0.8}/>
                                            <stop offset="95%" stopColor="#e94560" stopOpacity={0.1}/>
                                        </linearGradient>
                                        <linearGradient id="colorPending" x1="0" y1="0" x2="0" y2="1">
                                            <stop offset="5%" stopColor="#f39c12" stopOpacity={0.8}/>
                                            <stop offset="95%" stopColor="#f39c12" stopOpacity={0.1}/>
                                        </linearGradient>
                                    </defs>
                                    <CartesianGrid strokeDasharray="3 3" stroke="#0f3460" />
                                    <XAxis
                                        dataKey="month"
                                        stroke="#888"
                                        tick={{ fill: '#aaa', fontSize: 12 }}
                                    />
                                    <YAxis
                                        stroke="#888"
                                        tick={{ fill: '#aaa', fontSize: 12 }}
                                    />
                                    <Tooltip
                                        contentStyle={{
                                            backgroundColor: '#16213e',
                                            border: '1px solid #0f3460',
                                            borderRadius: '8px',
                                            color: '#eee'
                                        }}
                                    />
                                    <Legend
                                        wrapperStyle={{ paddingTop: '20px' }}
                                    />
                                    <Area
                                        type="monotone"
                                        dataKey="성공"
                                        stroke="#27ae60"
                                        strokeWidth={2}
                                        fillOpacity={1}
                                        fill="url(#colorSuccess)"
                                    />
                                    <Area
                                        type="monotone"
                                        dataKey="실패"
                                        stroke="#e94560"
                                        strokeWidth={2}
                                        fillOpacity={1}
                                        fill="url(#colorFail)"
                                    />
                                    <Area
                                        type="monotone"
                                        dataKey="대기"
                                        stroke="#f39c12"
                                        strokeWidth={2}
                                        fillOpacity={1}
                                        fill="url(#colorPending)"
                                    />
                                </AreaChart>
                            </ResponsiveContainer>
                        </div>

                        <div className="chart-summary">
                            <div className="chart-stat">
                                <span className="chart-stat-label">총 성공</span>
                                <span className="chart-stat-value success">1,069</span>
                            </div>
                            <div className="chart-stat">
                                <span className="chart-stat-label">총 실패</span>
                                <span className="chart-stat-value fail">86</span>
                            </div>
                            <div className="chart-stat">
                                <span className="chart-stat-label">총 대기</span>
                                <span className="chart-stat-value pending">107</span>
                            </div>
                            <div className="chart-stat">
                                <span className="chart-stat-label">성공률</span>
                                <span className="chart-stat-value success">84.7%</span>
                            </div>
                        </div>

                        <div className="modal-footer">
                            <button className="btn btn-primary" onClick={() => setShowChartModal(false)}>
                                닫기
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {/* 설정 모달 */}
            {showSettingsModal && (
                <div className="modal-overlay" onClick={() => setShowSettingsModal(false)}>
                    <div className="modal" onClick={(e) => e.stopPropagation()}>
                        <div className="modal-header">
                            <h3 className="modal-title">설정</h3>
                            <button className="modal-close" onClick={() => setShowSettingsModal(false)}>
                                ×
                            </button>
                        </div>

                        <div className="form-group">
                            <label>Connection</label>
                            <div className="radio-group">
                                <label className="radio-label" style={{ opacity: 0.5 }}>
                                    <input
                                        type="radio"
                                        name="connectionMode"
                                        value="agent"
                                        checked={config.connectionMode === 'agent'}
                                        disabled
                                    />
                                    Agent Server (준비중)
                                </label>
                                <label className="radio-label">
                                    <input
                                        type="radio"
                                        name="connectionMode"
                                        value="direct"
                                        checked={config.connectionMode === 'direct'}
                                        onChange={(e) => setConfig({ ...config, connectionMode: e.target.value })}
                                    />
                                    Direct
                                </label>
                            </div>
                        </div>

                        <div className="form-group">
                            <label>Agent Server URL</label>
                            <input
                                type="text"
                                className="input"
                                value={config.agentServerURL}
                                onChange={(e) => setConfig({ ...config, agentServerURL: e.target.value })}
                                placeholder="http://172.24.10.6:8080"
                                disabled
                            />
                        </div>

                        <div className="form-group">
                            <label>Timeout (초)</label>
                            <input
                                type="number"
                                className="input"
                                value={config.timeoutSeconds}
                                onChange={(e) => setConfig({ ...config, timeoutSeconds: parseInt(e.target.value) || 10 })}
                                min={5}
                                max={120}
                            />
                        </div>

                        <div className="form-group">
                            <label>설정 저장 경로</label>
                            <input
                                type="text"
                                className="input"
                                value={configDir}
                                disabled
                            />
                        </div>

                        <div className="modal-footer">
                            <button className="btn btn-secondary" onClick={() => setShowSettingsModal(false)}>
                                취소
                            </button>
                            <button className="btn btn-primary" onClick={handleSaveConfig}>
                                저장
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {/* 도움말 모달 */}
            {showHelpModal && (
                <div className="modal-overlay" onClick={() => setShowHelpModal(false)}>
                    <div className="modal modal-wide" onClick={(e) => e.stopPropagation()}>
                        <div className="modal-header">
                            <h3 className="modal-title">도움말</h3>
                            <button className="modal-close" onClick={() => setShowHelpModal(false)}>
                                ×
                            </button>
                        </div>

                        <div className="help-content">
                            <h4>FMS - Firewall Management System</h4>
                            <p>버전: {appVersion}</p>

                            <h5>[템플릿 관리]</h5>
                            <p>• 방화벽 규칙 템플릿을 생성/수정/삭제합니다</p>

                            <h5>[장비 관리]</h5>
                            <p>• 관리할 방화벽 장비(IP)를 등록합니다</p>
                            <p>• 서버 상태를 확인하고 템플릿을 배포합니다</p>

                            <h5>[배포 이력]</h5>
                            <p>• 배포 결과를 확인할 수 있습니다</p>
                            <p>• 규칙별 성공/실패 상태를 상세히 확인합니다</p>

                            <h5>[Import/Export]</h5>
                            <p>• 현재 탭의 데이터를 JSON 파일로 내보내거나 가져옵니다</p>

                            <h5>[연결 모드] (설정에서 변경)</h5>
                            <p>• Agent Server: Agent 서버(예: http://172.24.10.6:8080)를 통해 연결</p>
                            <p>  - 상태확인: POST /agent/req-respCheck</p>
                            <p>  - 배포: POST /agent/req-deploy</p>
                            <p>• Direct: 각 장비에 직접 HTTP 연결 (포트 80)</p>
                            <p>  - 상태확인: GET http://&#123;장비IP&#125;/respCheck</p>
                            <p>  - 배포: POST http://&#123;장비IP&#125;/deploy</p>
                        </div>

                        <div className="modal-footer">
                            <button className="btn btn-primary" onClick={() => setShowHelpModal(false)}>
                                닫기
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}

export default App;
