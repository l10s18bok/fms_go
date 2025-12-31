import 'package:flutter/material.dart';
import '../models/models.dart';
import '../services/services.dart';
import '../theme/app_theme.dart';
import '../widgets/widgets.dart';

/// 장비 관리 탭 - Wails의 DeviceTab.tsx와 동일
class DeviceTab extends StatefulWidget {
  final StorageService storage;
  final DeployService deployService;
  final VoidCallback? onDeployComplete;

  const DeviceTab({
    super.key,
    required this.storage,
    required this.deployService,
    this.onDeployComplete,
  });

  @override
  State<DeviceTab> createState() => DeviceTabState();
}

class DeviceTabState extends State<DeviceTab> {
  List<Firewall> _firewalls = [];
  List<Template> _templates = [];
  Set<int> _selectedIndexes = {};
  bool _isDeploying = false;
  bool _isChecking = false;

  @override
  void initState() {
    super.initState();
    _loadData();
  }

  /// 외부에서 호출 가능한 새로고침
  void refresh() {
    _loadData();
    setState(() {
      _selectedIndexes = {};
    });
  }

  Future<void> _loadData() async {
    final firewalls = await widget.storage.getAllFirewalls();
    final templates = await widget.storage.getAllTemplates();
    setState(() {
      _firewalls = firewalls;
      _templates = templates;
    });
  }

  Future<void> _handleAdd() async {
    final result = await _showDeviceDialog(null);
    if (result != null && result.isNotEmpty) {
      await widget.storage.saveFirewall(Firewall(
        index: -1,
        deviceName: result,
      ));
      await _loadData();
    }
  }

  Future<void> _handleEdit(Firewall fw) async {
    final result = await _showDeviceDialog(fw.deviceName);
    if (result != null && result.isNotEmpty) {
      await widget.storage.saveFirewall(fw.copyWith(deviceName: result));
      await _loadData();
    }
  }

  Future<String?> _showDeviceDialog(String? currentName) async {
    final controller = TextEditingController(text: currentName ?? '');
    final result = await showDialog<String>(
      context: context,
      builder: (context) => FmsDialog(
        title: currentName != null ? '장비 편집' : '장비 추가',
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(
              '장비명(IP)',
              style: TextStyle(
                color: AppTheme.textSecondary,
                fontSize: 13,
              ),
            ),
            const SizedBox(height: 6),
            TextField(
              controller: controller,
              autofocus: true,
              decoration: const InputDecoration(
                hintText: '192.168.1.100',
              ),
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
            onPressed: () => Navigator.of(context).pop(controller.text),
          ),
        ],
      ),
    );
    controller.dispose();
    return result;
  }

  Future<void> _handleDelete() async {
    if (_selectedIndexes.isEmpty) {
      _showMessage('삭제할 장비를 선택하세요.');
      return;
    }

    final confirmed = await showConfirmDialog(
      context,
      title: '삭제 확인',
      message: '${_selectedIndexes.length}개 장비를 삭제하시겠습니까?',
      isDanger: true,
    );

    if (confirmed) {
      for (final index in _selectedIndexes) {
        await widget.storage.deleteFirewall(index);
      }
      await _loadData();
      setState(() {
        _selectedIndexes = {};
      });
    }
  }

  Future<void> _handleCheckStatus() async {
    if (_selectedIndexes.isEmpty) {
      _showMessage('상태를 확인할 장비를 선택해주세요.');
      return;
    }

    setState(() {
      _isChecking = true;
    });

    try {
      final selectedFirewalls =
          _firewalls.where((f) => _selectedIndexes.contains(f.index)).toList();
      await widget.deployService.checkSelectedServerStatus(selectedFirewalls);
      await _loadData();
    } finally {
      setState(() {
        _isChecking = false;
      });
    }
  }

  Future<void> _handleDeployClick() async {
    if (_selectedIndexes.isEmpty) {
      _showMessage('배포할 장비를 선택하세요.');
      return;
    }

    // 최신 템플릿 목록 로드
    final latestTemplates = await widget.storage.getAllTemplates();
    if (latestTemplates.isEmpty) {
      _showMessage('배포할 템플릿이 없습니다.');
      return;
    }

    setState(() {
      _templates = latestTemplates;
    });

    // 템플릿 선택 다이얼로그
    String? selectedTemplate = latestTemplates.first.version;
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, setDialogState) => FmsDialog(
          title: '템플릿 배포',
          content: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                '${_selectedIndexes.length}개 장비에 배포할 템플릿을 선택하세요.',
                style: const TextStyle(color: AppTheme.textPrimary),
              ),
              const SizedBox(height: 16),
              const Text(
                '템플릿 선택',
                style: TextStyle(
                  color: AppTheme.textSecondary,
                  fontSize: 13,
                ),
              ),
              const SizedBox(height: 6),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 12),
                decoration: BoxDecoration(
                  color: AppTheme.backgroundColor,
                  borderRadius: BorderRadius.circular(6),
                  border: Border.all(color: AppTheme.borderColor),
                ),
                child: DropdownButton<String>(
                  value: selectedTemplate,
                  isExpanded: true,
                  underline: const SizedBox(),
                  dropdownColor: AppTheme.surfaceColor,
                  items: _templates
                      .map((t) => DropdownMenuItem(
                            value: t.version,
                            child: Text(t.version),
                          ))
                      .toList(),
                  onChanged: (value) {
                    setDialogState(() {
                      selectedTemplate = value;
                    });
                  },
                ),
              ),
            ],
          ),
          actions: [
            FmsButton(
              text: '취소',
              type: FmsButtonType.secondary,
              onPressed: () => Navigator.of(context).pop(false),
            ),
            FmsButton(
              text: '배포 실행',
              onPressed: () => Navigator.of(context).pop(true),
            ),
          ],
        ),
      ),
    );

    if (confirmed == true && selectedTemplate != null && mounted) {
      await _executeDeploy(selectedTemplate!);
    }
  }

  Future<void> _executeDeploy(String templateVersion) async {
    setState(() {
      _isDeploying = true;
    });

    int successCount = 0;
    int failCount = 0;

    try {
      final template = await widget.storage.getTemplate(templateVersion);
      if (template == null) {
        _showMessage('템플릿을 찾을 수 없습니다.');
        return;
      }

      for (final index in _selectedIndexes) {
        final firewall = _firewalls.firstWhere((f) => f.index == index);
        try {
          final result =
              await widget.deployService.deploy(firewall, template);
          if (result.status == 'success') {
            successCount++;
          } else {
            failCount++;
          }
        } catch (e) {
          failCount++;
        }
      }

      await _loadData();

      // 배포 완료 콜백 호출 (이력 탭 새로고침용)
      widget.onDeployComplete?.call();

      if (failCount == 0) {
        _showMessage('$successCount개 장비에 배포가 완료되었습니다.');
      } else {
        _showMessage('배포 완료: 성공 $successCount개, 실패 $failCount개');
      }
    } finally {
      setState(() {
        _isDeploying = false;
      });
    }
  }

  void _handleToggleSelect(int index) {
    setState(() {
      if (_selectedIndexes.contains(index)) {
        _selectedIndexes.remove(index);
      } else {
        _selectedIndexes.add(index);
      }
    });
  }

  void _handleSelectAll(bool? checked) {
    setState(() {
      if (checked == true) {
        _selectedIndexes = _firewalls.map((f) => f.index).toSet();
      } else {
        _selectedIndexes = {};
      }
    });
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
    return FmsCard(
      title: '장비 목록',
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // 버튼 그룹
          Row(
            children: [
              FmsButton(
                text: '+ 장비 추가',
                onPressed: _isDeploying || _isChecking ? null : _handleAdd,
              ),
              const SizedBox(width: 8),
              FmsButton(
                text: _isChecking ? '확인중...' : '상태 확인',
                type: FmsButtonType.secondary,
                isLoading: _isChecking,
                onPressed:
                    _isDeploying || _isChecking ? null : _handleCheckStatus,
              ),
              const SizedBox(width: 8),
              FmsButton(
                text: _isDeploying ? '배포 중...' : '배포',
                isLoading: _isDeploying,
                onPressed:
                    _isDeploying || _isChecking ? null : _handleDeployClick,
              ),
              const SizedBox(width: 8),
              FmsButton(
                text: '삭제',
                type: FmsButtonType.danger,
                onPressed: _isDeploying || _isChecking ? null : _handleDelete,
              ),
            ],
          ),
          const SizedBox(height: 16),
          // 테이블
          Expanded(
            child: _firewalls.isEmpty
                ? const Center(
                    child: Text(
                      '등록된 장비가 없습니다',
                      style: TextStyle(color: AppTheme.textMuted),
                    ),
                  )
                : SingleChildScrollView(
                    child: _buildTable(),
                  ),
          ),
        ],
      ),
    );
  }

  Widget _buildTable() {
    return Table(
      columnWidths: const {
        0: FixedColumnWidth(50),
        1: FlexColumnWidth(),
        2: FixedColumnWidth(150),
        3: FixedColumnWidth(150),
        4: FixedColumnWidth(150),
        5: FixedColumnWidth(150),
      },
      border: TableBorder(
        horizontalInside: BorderSide(color: AppTheme.borderColor),
      ),
      children: [
        // 헤더
        TableRow(
          decoration: BoxDecoration(color: AppTheme.borderColor),
          children: [
            TableCell(
              child: Padding(
                padding: const EdgeInsets.all(12),
                child: Checkbox(
                  value: _selectedIndexes.length == _firewalls.length &&
                      _firewalls.isNotEmpty,
                  onChanged: _handleSelectAll,
                ),
              ),
            ),
            const TableCell(
              child: Padding(
                padding: EdgeInsets.all(12),
                child: Text(
                  '장비명(IP)',
                  style: TextStyle(
                    color: Colors.white,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ),
            ),
            const TableCell(
              child: Padding(
                padding: EdgeInsets.all(12),
                child: Text(
                  '서버상태',
                  style: TextStyle(
                    color: Colors.white,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ),
            ),
            const TableCell(
              child: Padding(
                padding: EdgeInsets.all(12),
                child: Text(
                  '배포상태',
                  style: TextStyle(
                    color: Colors.white,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ),
            ),
            const TableCell(
              child: Padding(
                padding: EdgeInsets.all(12),
                child: Text(
                  '버전',
                  style: TextStyle(
                    color: Colors.white,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ),
            ),
            const TableCell(
              child: Padding(
                padding: EdgeInsets.all(12),
                child: Text(
                  '작업',
                  style: TextStyle(
                    color: Colors.white,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ),
            ),
          ],
        ),
        // 데이터 행
        ..._firewalls.map((fw) => _buildTableRow(fw)),
      ],
    );
  }

  TableRow _buildTableRow(Firewall fw) {
    return TableRow(
      children: [
        TableCell(
          verticalAlignment: TableCellVerticalAlignment.middle,
          child: Padding(
            padding: const EdgeInsets.all(12),
            child: Checkbox(
              value: _selectedIndexes.contains(fw.index),
              onChanged: (_) => _handleToggleSelect(fw.index),
            ),
          ),
        ),
        TableCell(
          verticalAlignment: TableCellVerticalAlignment.middle,
          child: Padding(
            padding: const EdgeInsets.all(12),
            child: Text(
              fw.deviceName,
              style: const TextStyle(color: AppTheme.textPrimary),
            ),
          ),
        ),
        TableCell(
          verticalAlignment: TableCellVerticalAlignment.middle,
          child: Padding(
            padding: const EdgeInsets.all(12),
            child: StatusBadge.fromStatus(fw.serverStatus),
          ),
        ),
        TableCell(
          verticalAlignment: TableCellVerticalAlignment.middle,
          child: Padding(
            padding: const EdgeInsets.all(12),
            child: StatusBadge.fromStatus(fw.deployStatus),
          ),
        ),
        TableCell(
          verticalAlignment: TableCellVerticalAlignment.middle,
          child: Padding(
            padding: const EdgeInsets.all(12),
            child: Text(
              fw.version.isEmpty ? '-' : fw.version,
              style: const TextStyle(color: AppTheme.textPrimary),
            ),
          ),
        ),
        TableCell(
          verticalAlignment: TableCellVerticalAlignment.middle,
          child: Padding(
            padding: const EdgeInsets.all(12),
            child: FmsButton(
              text: '편집',
              type: FmsButtonType.secondary,
              onPressed: () => _handleEdit(fw),
            ),
          ),
        ),
      ],
    );
  }
}
