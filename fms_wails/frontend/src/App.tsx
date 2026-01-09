import { useState, useRef, useEffect, useCallback } from 'react';
import './App.css';
import { APP_HELP } from './constants/helpTexts';
import TemplateTab, { TemplateTabRef } from './components/TemplateTab';
import DeviceTab, { DeviceTabRef } from './components/DeviceTab';
import HistoryTab, { HistoryTabRef } from './components/HistoryTab';
// [DEMO] 차트 데모 - 삭제 시 이 import와 ChartDemo 관련 코드 제거
import ChartDemo from './components/ChartDemo';
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
    DeleteAllTemplates,
    DeleteAllFirewalls,
    DeleteAllHistory,
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
    // [DEMO] 차트 데모 - 삭제 시 이 state와 ChartDemo 관련 코드 제거
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

            // 탭별 데이터 타입명
            const tabNames: Record<TabType, string> = {
                template: '템플릿',
                device: '장비',
                history: '배포 이력'
            };
            const tabName = tabNames[activeTab];

            // 확인 팝업 표시
            const confirmed = await ConfirmDialog(
                'Import 확인',
                `기존 ${tabName} 데이터가 모두 삭제되고 새로운 데이터로 교체됩니다.\n계속 진행하시겠습니까?`
            );

            if (!confirmed) {
                e.target.value = '';
                return;
            }

            // 현재 탭에 따라 데이터 타입 확인 및 저장
            let importedCount = 0;

            if (activeTab === 'template') {
                if (!Array.isArray(data)) {
                    alert('유효한 템플릿 데이터가 아닙니다.');
                    return;
                }
                // 기존 데이터 모두 삭제
                await DeleteAllTemplates();

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
                // 기존 데이터 모두 삭제
                await DeleteAllFirewalls();

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
                // 기존 데이터 모두 삭제
                await DeleteAllHistory();

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

            {/* [DEMO] 차트 데모 모달 - 삭제 시 이 컴포넌트와 관련 state 제거 */}
            <ChartDemo isOpen={showChartModal} onClose={() => setShowChartModal(false)} />

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
                            <h4>{APP_HELP.title}</h4>
                            <p>버전: {appVersion}</p>

                            {APP_HELP.sections.map((section, i) => (
                                <div key={i}>
                                    <h5>[{section.name}]</h5>
                                    {section.items.map((item, j) => (
                                        <p key={j}>• {item}</p>
                                    ))}
                                </div>
                            ))}

                            <h5>[연결 모드] (설정에서 변경)</h5>
                            {APP_HELP.connectionModes.map((mode, i) => (
                                <div key={i}>
                                    <p>• {mode.name}: {mode.desc}</p>
                                    {mode.endpoints.map((ep, j) => (
                                        <p key={j}>  - {ep}</p>
                                    ))}
                                </div>
                            ))}

                            <h5>[규칙 포맷]</h5>
                            <p><code>{APP_HELP.ruleFormat.pattern}</code></p>
                            <p>예시:</p>
                            {APP_HELP.ruleFormat.examples.map((ex, i) => (
                                <p key={i}><code>{ex}</code></p>
                            ))}
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
