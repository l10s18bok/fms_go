/**
 * ChartDemo - 차트 데모 컴포넌트
 *
 * 이 파일은 데모용 차트 컴포넌트입니다.
 * 나중에 삭제하려면 이 파일을 삭제하고 App.tsx에서 import와 사용 부분을 제거하세요.
 *
 * 삭제 방법:
 * 1. 이 파일 삭제: src/components/ChartDemo.tsx
 * 2. App.tsx에서 다음 줄 삭제:
 *    - import ChartDemo from './components/ChartDemo';
 *    - const [showChartModal, setShowChartModal] = useState(false);
 *    - <ChartDemo ... /> 컴포넌트
 *    - onClick={() => setShowChartModal(true)} (footer 버전 클릭 부분)
 * 3. recharts 패키지가 더 이상 필요 없으면: npm uninstall recharts
 */

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

interface ChartDemoProps {
    isOpen: boolean;
    onClose: () => void;
}

// 데모 데이터 - 월별 배포 통계 (임의의 값)
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

function ChartDemo({ isOpen, onClose }: ChartDemoProps) {
    if (!isOpen) return null;

    return (
        <div className="modal-overlay" onClick={onClose}>
            <div className="modal modal-chart" onClick={(e) => e.stopPropagation()}>
                <div className="modal-header">
                    <h3 className="modal-title">월별 배포 통계 (Demo)</h3>
                    <button className="modal-close" onClick={onClose}>
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
                    <button className="btn btn-primary" onClick={onClose}>
                        닫기
                    </button>
                </div>
            </div>
        </div>
    );
}

export default ChartDemo;
