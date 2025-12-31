import 'dart:convert';
import 'dart:io';
import 'package:flutter/material.dart';
import 'package:file_picker/file_picker.dart';
import 'package:fl_chart/fl_chart.dart';
import 'package:intl/date_symbol_data_local.dart';
import 'models/models.dart';
import 'services/services.dart';
import 'screens/screens.dart';
import 'theme/app_theme.dart';
import 'widgets/widgets.dart';

/// FMS - Firewall Management System
/// Flutter Windows 앱
void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await initializeDateFormatting('ko_KR', null);
  runApp(const FmsApp());
}

class FmsApp extends StatelessWidget {
  const FmsApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'FMS - Firewall Management System',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.darkTheme,
      home: const MainScreen(),
    );
  }
}

/// 메인 화면 - Wails의 App.tsx와 동일한 구조
class MainScreen extends StatefulWidget {
  const MainScreen({super.key});

  @override
  State<MainScreen> createState() => _MainScreenState();
}

class _MainScreenState extends State<MainScreen>
    with SingleTickerProviderStateMixin {
  late TabController _tabController;
  late StorageService _storage;
  late DeployService _deployService;

  // 탭 상태 관리용 GlobalKey
  final _templateTabKey = GlobalKey<TemplateTabState>();
  final _deviceTabKey = GlobalKey<DeviceTabState>();
  final _historyTabKey = GlobalKey<HistoryTabState>();

  AppConfig _config = AppConfig();
  String _configDir = '';
  final String _appVersion = '1.0.0';

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 3, vsync: this);
    _storage = StorageService();
    _deployService = DeployService(_storage);
    _loadConfig();
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  Future<void> _loadConfig() async {
    final config = await _storage.getConfig();
    final configDir = await _storage.getConfigDir();
    setState(() {
      _config = config;
      _configDir = configDir;
    });
  }

  // ==================== Import/Export ====================

  Future<void> _handleImport() async {
    final result = await FilePicker.platform.pickFiles(
      type: FileType.custom,
      allowedExtensions: ['json'],
    );

    if (result == null || result.files.isEmpty) return;

    try {
      final file = File(result.files.single.path!);
      final text = await file.readAsString();
      final data = jsonDecode(text) as List<dynamic>;

      int importedCount = 0;
      final currentTab = _tabController.index;

      if (currentTab == 0) {
        // 템플릿
        for (final item in data) {
          if (item['version'] != null &&
              item['contents'] != null &&
              (item['contents'] as String).trim().isNotEmpty) {
            await _storage.saveTemplate(item['version'], item['contents']);
            importedCount++;
          }
        }
        _templateTabKey.currentState?.refresh();
      } else if (currentTab == 1) {
        // 장비
        for (final item in data) {
          if (item['deviceName'] != null) {
            await _storage.saveFirewall(
                Firewall.fromJson(item as Map<String, dynamic>));
            importedCount++;
          }
        }
        _deviceTabKey.currentState?.refresh();
      } else if (currentTab == 2) {
        // 이력
        for (final item in data) {
          if (item['deviceIp'] != null && item['templateVersion'] != null) {
            await _storage.saveHistory(
                DeployHistory.fromJson(item as Map<String, dynamic>));
            importedCount++;
          }
        }
        _historyTabKey.currentState?.refresh();
      }

      if (importedCount > 0) {
        _showMessage('$importedCount개 항목을 가져왔습니다.');
      } else {
        _showMessage('가져올 수 있는 유효한 데이터가 없습니다.');
      }
    } catch (e) {
      _showMessage('파일을 읽는 중 오류가 발생했습니다.');
    }
  }

  Future<void> _handleExport() async {
    List<dynamic> data = [];
    String filename = '';
    final currentTab = _tabController.index;

    if (currentTab == 0) {
      final templates = await _storage.getAllTemplates();
      data = templates.map((t) => t.toJson()).toList();
      filename = 'templates.json';
    } else if (currentTab == 1) {
      final firewalls = await _storage.getAllFirewalls();
      data = firewalls.map((f) => f.toJson()).toList();
      filename = 'firewalls.json';
    } else if (currentTab == 2) {
      final history = await _storage.getAllHistory();
      data = history.map((h) => h.toJson()).toList();
      filename = 'history.json';
    }

    if (data.isEmpty) {
      _showMessage('내보낼 데이터가 없습니다.');
      return;
    }

    final result = await FilePicker.platform.saveFile(
      dialogTitle: '파일 내보내기',
      fileName: filename,
      type: FileType.custom,
      allowedExtensions: ['json'],
    );

    if (result == null) return;

    try {
      final file = File(result);
      await file.writeAsString(jsonEncode(data));
      _showMessage('파일이 저장되었습니다.');
    } catch (e) {
      _showMessage('파일 저장 중 오류가 발생했습니다.');
    }
  }

  Future<void> _handleReset() async {
    final confirmed = await showConfirmDialog(
      context,
      title: '초기화',
      message: '모든 데이터(템플릿, 장비, 배포이력)를 초기화하시겠습니까?',
      isDanger: true,
    );

    if (confirmed && mounted) {
      await _storage.resetAll();
      _templateTabKey.currentState?.refresh();
      _deviceTabKey.currentState?.refresh();
      _historyTabKey.currentState?.refresh();
      _showMessage('모든 데이터가 초기화되었습니다.');
    }
  }

  // ==================== 설정/도움말 ====================

  Future<void> _showSettingsDialog() async {
    await _loadConfig();

    if (!mounted) return;

    AppConfig tempConfig = _config;

    await showDialog(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, setDialogState) => FmsDialog(
          title: '설정',
          width: 500,
          content: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Connection Mode
              const Text(
                'Connection',
                style: TextStyle(
                  color: AppTheme.textSecondary,
                  fontSize: 13,
                ),
              ),
              const SizedBox(height: 6),
              Row(
                children: [
                  Opacity(
                    opacity: 0.5,
                    child: Row(
                      children: [
                        Radio<String>(
                          value: 'agent',
                          groupValue: tempConfig.connectionMode,
                          onChanged: null,
                        ),
                        const Text(
                          'Agent Server (준비중)',
                          style: TextStyle(color: AppTheme.textSecondary),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(width: 16),
                  Row(
                    children: [
                      Radio<String>(
                        value: 'direct',
                        groupValue: tempConfig.connectionMode,
                        onChanged: (value) {
                          setDialogState(() {
                            tempConfig =
                                tempConfig.copyWith(connectionMode: value);
                          });
                        },
                      ),
                      const Text(
                        'Direct',
                        style: TextStyle(color: AppTheme.textPrimary),
                      ),
                    ],
                  ),
                ],
              ),
              const SizedBox(height: 16),
              // Agent Server URL
              const Text(
                'Agent Server URL',
                style: TextStyle(
                  color: AppTheme.textSecondary,
                  fontSize: 13,
                ),
              ),
              const SizedBox(height: 6),
              TextField(
                controller:
                    TextEditingController(text: tempConfig.agentServerURL),
                enabled: false,
                decoration: const InputDecoration(
                  hintText: 'http://172.24.10.6:8080',
                ),
                onChanged: (value) {
                  tempConfig = tempConfig.copyWith(agentServerURL: value);
                },
              ),
              const SizedBox(height: 16),
              // Timeout
              const Text(
                'Timeout (초)',
                style: TextStyle(
                  color: AppTheme.textSecondary,
                  fontSize: 13,
                ),
              ),
              const SizedBox(height: 6),
              TextField(
                controller: TextEditingController(
                    text: tempConfig.timeoutSeconds.toString()),
                keyboardType: TextInputType.number,
                decoration: const InputDecoration(
                  hintText: '10',
                ),
                onChanged: (value) {
                  final parsed = int.tryParse(value);
                  if (parsed != null) {
                    tempConfig = tempConfig.copyWith(timeoutSeconds: parsed);
                  }
                },
              ),
              const SizedBox(height: 16),
              // Config Dir
              const Text(
                '설정 저장 경로',
                style: TextStyle(
                  color: AppTheme.textSecondary,
                  fontSize: 13,
                ),
              ),
              const SizedBox(height: 6),
              TextField(
                controller: TextEditingController(text: _configDir),
                enabled: false,
              ),
            ],
          ),
          actions: [
            FmsButton(
              text: '취소',
              type: FmsButtonType.secondary,
              onPressed: () => Navigator.of(context).pop(),
            ),
            FmsButton(
              text: '저장',
              onPressed: () async {
                // 유효성 검사
                if (tempConfig.connectionMode == 'agent' &&
                    tempConfig.agentServerURL.isEmpty) {
                  _showMessage('Agent Server URL을 입력해주세요.');
                  return;
                }
                if (tempConfig.timeoutSeconds < 5 ||
                    tempConfig.timeoutSeconds > 120) {
                  _showMessage('타임아웃은 5~120 사이의 숫자를 입력해주세요.');
                  return;
                }

                await _storage.saveConfig(tempConfig);
                setState(() {
                  _config = tempConfig;
                });
                if (context.mounted) {
                  Navigator.of(context).pop();
                }
                _showMessage('설정이 저장되었습니다.');
              },
            ),
          ],
        ),
      ),
    );
  }

  void _showHelpDialog() {
    showDialog(
      context: context,
      builder: (context) => FmsDialog(
        title: '도움말',
        width: 600,
        content: SingleChildScrollView(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text(
                'FMS - Firewall Management System',
                style: TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.w600,
                  color: AppTheme.primaryColor,
                ),
              ),
              const SizedBox(height: 8),
              Text(
                '버전: $_appVersion',
                style: const TextStyle(color: AppTheme.textSecondary),
              ),
              const SizedBox(height: 16),
              _buildHelpSection('[템플릿 관리]', [
                '방화벽 규칙 템플릿을 생성/수정/삭제합니다',
              ]),
              _buildHelpSection('[장비 관리]', [
                '관리할 방화벽 장비(IP)를 등록합니다',
                '서버 상태를 확인하고 템플릿을 배포합니다',
              ]),
              _buildHelpSection('[배포 이력]', [
                '배포 결과를 확인할 수 있습니다',
                '규칙별 성공/실패 상태를 상세히 확인합니다',
              ]),
              _buildHelpSection('[Import/Export]', [
                '현재 탭의 데이터를 JSON 파일로 내보내거나 가져옵니다',
              ]),
              _buildHelpSection('[연결 모드] (설정에서 변경)', [
                'Agent Server: Agent 서버를 통해 연결',
                '  - 상태확인: POST /agent/req-respCheck',
                '  - 배포: POST /agent/req-deploy',
                'Direct: 각 장비에 직접 HTTP 연결 (포트 80)',
                '  - 상태확인: GET http://{장비IP}/respCheck',
                '  - 배포: POST http://{장비IP}/deploy',
              ]),
            ],
          ),
        ),
        actions: [
          FmsButton(
            text: '닫기',
            onPressed: () => Navigator.of(context).pop(),
          ),
        ],
      ),
    );
  }

  Widget _buildHelpSection(String title, List<String> items) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const SizedBox(height: 16),
        Text(
          title,
          style: const TextStyle(
            color: AppTheme.infoColor,
            fontWeight: FontWeight.w600,
            fontSize: 14,
          ),
        ),
        const SizedBox(height: 8),
        ...items.map((item) => Padding(
              padding: const EdgeInsets.only(bottom: 4),
              child: Text(
                '• $item',
                style: const TextStyle(
                  color: AppTheme.textSecondary,
                  fontSize: 13,
                ),
              ),
            )),
      ],
    );
  }

  // Wails와 동일한 차트 데이터
  static const _chartData = [
    {'month': '1월', 'success': 45, 'fail': 12, 'pending': 8},
    {'month': '2월', 'success': 52, 'fail': 8, 'pending': 15},
    {'month': '3월', 'success': 78, 'fail': 5, 'pending': 10},
    {'month': '4월', 'success': 65, 'fail': 15, 'pending': 12},
    {'month': '5월', 'success': 89, 'fail': 3, 'pending': 5},
    {'month': '6월', 'success': 95, 'fail': 7, 'pending': 8},
    {'month': '7월', 'success': 110, 'fail': 4, 'pending': 6},
    {'month': '8월', 'success': 85, 'fail': 10, 'pending': 12},
    {'month': '9월', 'success': 102, 'fail': 6, 'pending': 9},
    {'month': '10월', 'success': 120, 'fail': 8, 'pending': 7},
    {'month': '11월', 'success': 98, 'fail': 5, 'pending': 11},
    {'month': '12월', 'success': 130, 'fail': 3, 'pending': 4},
  ];

  void _showChartDemo() {
    showDialog(
      context: context,
      builder: (context) => FmsDialog(
        title: '📊 월별 배포 통계 (Demo)',
        width: 800,
        maxHeight: 600,
        content: SingleChildScrollView(
          child: Column(
            children: [
              // 차트 - Wails의 AreaChart와 동일
              Container(
                height: 350,
                padding: const EdgeInsets.all(20),
                decoration: BoxDecoration(
                  color: AppTheme.backgroundColor,
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: AppTheme.borderColor),
                ),
                child: LineChart(
                  LineChartData(
                    gridData: FlGridData(
                      show: true,
                      drawVerticalLine: true,
                      horizontalInterval: 20,
                      getDrawingHorizontalLine: (value) => FlLine(
                        color: AppTheme.borderColor,
                        strokeWidth: 1,
                        dashArray: [3, 3],
                      ),
                      getDrawingVerticalLine: (value) => FlLine(
                        color: AppTheme.borderColor,
                        strokeWidth: 1,
                        dashArray: [3, 3],
                      ),
                    ),
                    titlesData: FlTitlesData(
                      leftTitles: AxisTitles(
                        sideTitles: SideTitles(
                          showTitles: true,
                          reservedSize: 40,
                          interval: 40,
                          getTitlesWidget: (value, meta) => Padding(
                            padding: const EdgeInsets.only(right: 8),
                            child: Text(
                              value.toInt().toString(),
                              style: const TextStyle(
                                color: Color(0xFFAAAAAA),
                                fontSize: 12,
                              ),
                            ),
                          ),
                        ),
                      ),
                      bottomTitles: AxisTitles(
                        sideTitles: SideTitles(
                          showTitles: true,
                          reservedSize: 30,
                          getTitlesWidget: (value, meta) {
                            final index = value.toInt();
                            if (index >= 0 && index < _chartData.length) {
                              return Padding(
                                padding: const EdgeInsets.only(top: 8),
                                child: Text(
                                  _chartData[index]['month'] as String,
                                  style: const TextStyle(
                                    color: Color(0xFFAAAAAA),
                                    fontSize: 12,
                                  ),
                                ),
                              );
                            }
                            return const Text('');
                          },
                        ),
                      ),
                      topTitles: const AxisTitles(
                          sideTitles: SideTitles(showTitles: false)),
                      rightTitles: const AxisTitles(
                          sideTitles: SideTitles(showTitles: false)),
                    ),
                    borderData: FlBorderData(show: false),
                    minY: 0,
                    maxY: 140,
                    lineTouchData: LineTouchData(
                      touchTooltipData: LineTouchTooltipData(
                        getTooltipColor: (touchedSpot) => AppTheme.surfaceColor,
                        tooltipBorder: BorderSide(color: AppTheme.borderColor),
                        tooltipRoundedRadius: 8,
                        getTooltipItems: (touchedSpots) {
                          return touchedSpots.map((spot) {
                            String label;
                            Color color;
                            if (spot.barIndex == 0) {
                              label = '성공';
                              color = AppTheme.successColor;
                            } else if (spot.barIndex == 1) {
                              label = '실패';
                              color = AppTheme.dangerColor;
                            } else {
                              label = '대기';
                              color = AppTheme.warningColor;
                            }
                            return LineTooltipItem(
                              '$label: ${spot.y.toInt()}',
                              TextStyle(color: color, fontSize: 12),
                            );
                          }).toList();
                        },
                      ),
                    ),
                    lineBarsData: [
                      // 성공 - 녹색 (#27ae60)
                      LineChartBarData(
                        spots: List.generate(
                          _chartData.length,
                          (i) => FlSpot(i.toDouble(),
                              (_chartData[i]['success'] as int).toDouble()),
                        ),
                        isCurved: true,
                        curveSmoothness: 0.3,
                        color: const Color(0xFF27AE60),
                        barWidth: 2,
                        isStrokeCapRound: true,
                        dotData: const FlDotData(show: false),
                        belowBarData: BarAreaData(
                          show: true,
                          gradient: LinearGradient(
                            begin: Alignment.topCenter,
                            end: Alignment.bottomCenter,
                            colors: [
                              const Color(0xFF27AE60).withValues(alpha: 0.8),
                              const Color(0xFF27AE60).withValues(alpha: 0.1),
                            ],
                          ),
                        ),
                      ),
                      // 실패 - 빨강 (#e94560)
                      LineChartBarData(
                        spots: List.generate(
                          _chartData.length,
                          (i) => FlSpot(i.toDouble(),
                              (_chartData[i]['fail'] as int).toDouble()),
                        ),
                        isCurved: true,
                        curveSmoothness: 0.3,
                        color: const Color(0xFFE94560),
                        barWidth: 2,
                        isStrokeCapRound: true,
                        dotData: const FlDotData(show: false),
                        belowBarData: BarAreaData(
                          show: true,
                          gradient: LinearGradient(
                            begin: Alignment.topCenter,
                            end: Alignment.bottomCenter,
                            colors: [
                              const Color(0xFFE94560).withValues(alpha: 0.8),
                              const Color(0xFFE94560).withValues(alpha: 0.1),
                            ],
                          ),
                        ),
                      ),
                      // 대기 - 노랑 (#f39c12)
                      LineChartBarData(
                        spots: List.generate(
                          _chartData.length,
                          (i) => FlSpot(i.toDouble(),
                              (_chartData[i]['pending'] as int).toDouble()),
                        ),
                        isCurved: true,
                        curveSmoothness: 0.3,
                        color: const Color(0xFFF39C12),
                        barWidth: 2,
                        isStrokeCapRound: true,
                        dotData: const FlDotData(show: false),
                        belowBarData: BarAreaData(
                          show: true,
                          gradient: LinearGradient(
                            begin: Alignment.topCenter,
                            end: Alignment.bottomCenter,
                            colors: [
                              const Color(0xFFF39C12).withValues(alpha: 0.8),
                              const Color(0xFFF39C12).withValues(alpha: 0.1),
                            ],
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 16),
              // 범례 - Wails Legend와 동일
              Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  _buildLegendItem('성공', const Color(0xFF27AE60)),
                  const SizedBox(width: 24),
                  _buildLegendItem('실패', const Color(0xFFE94560)),
                  const SizedBox(width: 24),
                  _buildLegendItem('대기', const Color(0xFFF39C12)),
                ],
              ),
              const SizedBox(height: 20),
              // 통계 요약 - Wails chart-summary와 동일
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: AppTheme.surfaceColor,
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: AppTheme.borderColor),
                ),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceAround,
                  children: [
                    _buildChartStat('총 성공', '1,069', const Color(0xFF27AE60)),
                    _buildChartStat('총 실패', '86', const Color(0xFFE94560)),
                    _buildChartStat('총 대기', '107', const Color(0xFFF39C12)),
                    _buildChartStat('성공률', '84.7%', const Color(0xFF27AE60)),
                  ],
                ),
              ),
            ],
          ),
        ),
        actions: [
          FmsButton(
            text: '닫기',
            onPressed: () => Navigator.of(context).pop(),
          ),
        ],
      ),
    );
  }

  Widget _buildLegendItem(String label, Color color) {
    return Row(
      children: [
        Container(
          width: 12,
          height: 12,
          decoration: BoxDecoration(
            color: color,
            borderRadius: BorderRadius.circular(2),
          ),
        ),
        const SizedBox(width: 6),
        Text(
          label,
          style: const TextStyle(
            color: AppTheme.textSecondary,
            fontSize: 13,
          ),
        ),
      ],
    );
  }

  Widget _buildChartStat(String label, String value, Color color) {
    return Column(
      children: [
        Text(
          label,
          style: const TextStyle(
            color: AppTheme.textMuted,
            fontSize: 13,
          ),
        ),
        const SizedBox(height: 6),
        Text(
          value,
          style: TextStyle(
            color: color,
            fontSize: 20,
            fontWeight: FontWeight.w700,
          ),
        ),
      ],
    );
  }

  void _showMessage(String message) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(message),
        backgroundColor: AppTheme.surfaceColor,
        behavior: SnackBarBehavior.floating,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Column(
        children: [
          // 상단 메뉴바
          _buildMenuBar(),
          // 탭 네비게이션
          _buildTabNav(),
          // 탭 컨텐츠
          Expanded(
            child: Padding(
              padding: const EdgeInsets.all(24),
              child: TabBarView(
                controller: _tabController,
                physics: const NeverScrollableScrollPhysics(),
                children: [
                  TemplateTab(
                    key: _templateTabKey,
                    storage: _storage,
                  ),
                  DeviceTab(
                    key: _deviceTabKey,
                    storage: _storage,
                    deployService: _deployService,
                    onDeployComplete: () =>
                        _historyTabKey.currentState?.refresh(),
                  ),
                  HistoryTab(
                    key: _historyTabKey,
                    storage: _storage,
                  ),
                ],
              ),
            ),
          ),
          // 하단 상태바
          _buildFooter(),
        ],
      ),
    );
  }

  Widget _buildMenuBar() {
    return Container(
      height: 36,
      padding: const EdgeInsets.symmetric(horizontal: 8),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: [AppTheme.surfaceColor, AppTheme.backgroundColor],
        ),
        border: const Border(
          bottom: BorderSide(color: AppTheme.borderColor),
        ),
      ),
      child: Row(
        children: [
          _buildMenuButton('파일', [
            _MenuAction('Import', _handleImport),
            _MenuAction('Export', _handleExport),
            null, // divider
            _MenuAction('Reset', _handleReset, isDanger: true),
          ]),
          _buildMenuButton('도구', [
            _MenuAction('설정', _showSettingsDialog),
          ]),
          _buildMenuButton('도움말', [
            _MenuAction('도움말', _showHelpDialog),
          ]),
        ],
      ),
    );
  }

  Widget _buildMenuButton(String label, List<_MenuAction?> actions) {
    return PopupMenuButton<int>(
      offset: const Offset(0, 36),
      tooltip: '',
      child: Container(
        height: 36,
        padding: const EdgeInsets.symmetric(horizontal: 16),
        alignment: Alignment.center,
        child: Text(
          label,
          style: const TextStyle(
            color: AppTheme.textSecondary,
            fontSize: 13,
          ),
        ),
      ),
      itemBuilder: (context) {
        final items = <PopupMenuEntry<int>>[];
        for (int i = 0; i < actions.length; i++) {
          final action = actions[i];
          if (action == null) {
            items.add(const PopupMenuDivider());
          } else {
            items.add(PopupMenuItem<int>(
              value: i,
              child: Text(
                action.label,
                style: TextStyle(
                  color:
                      action.isDanger ? AppTheme.dangerColor : AppTheme.textPrimary,
                ),
              ),
            ));
          }
        }
        return items;
      },
      onSelected: (index) {
        final action = actions[index];
        if (action != null) {
          action.onTap();
        }
      },
    );
  }

  Widget _buildTabNav() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
      decoration: const BoxDecoration(
        color: AppTheme.surfaceColor,
        border: Border(
          bottom: BorderSide(color: AppTheme.borderColor),
        ),
      ),
      child: Row(
        children: [
          _buildTabButton(0, '템플릿 관리'),
          const SizedBox(width: 4),
          _buildTabButton(1, '장비 관리'),
          const SizedBox(width: 4),
          _buildTabButton(2, '배포 이력'),
        ],
      ),
    );
  }

  Widget _buildTabButton(int index, String label) {
    final isActive = _tabController.index == index;
    return GestureDetector(
      onTap: () {
        setState(() {
          _tabController.animateTo(index);
        });
      },
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 10),
        decoration: BoxDecoration(
          color: isActive ? AppTheme.primaryColor : Colors.transparent,
          borderRadius: const BorderRadius.only(
            topLeft: Radius.circular(6),
            topRight: Radius.circular(6),
          ),
        ),
        child: Text(
          label,
          style: TextStyle(
            color: isActive ? Colors.white : AppTheme.textSecondary,
            fontSize: 14,
          ),
        ),
      ),
    );
  }

  Widget _buildFooter() {
    return Container(
      height: 30,
      padding: const EdgeInsets.symmetric(horizontal: 16),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: [AppTheme.surfaceColor, AppTheme.backgroundColor],
        ),
        border: const Border(
          top: BorderSide(color: AppTheme.borderColor),
        ),
      ),
      child: Row(
        children: [
          MouseRegion(
            cursor: SystemMouseCursors.click,
            child: GestureDetector(
              onTap: _showChartDemo,
              child: Text(
                'FMS v$_appVersion',
                style: const TextStyle(
                  color: AppTheme.textMuted,
                  fontSize: 11,
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _MenuAction {
  final String label;
  final VoidCallback onTap;
  final bool isDanger;

  _MenuAction(this.label, this.onTap, {this.isDanger = false});
}
