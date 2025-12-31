import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import '../models/models.dart';
import '../services/services.dart';
import '../theme/app_theme.dart';
import '../widgets/widgets.dart';

/// 배포 이력 탭 - Wails의 HistoryTab.tsx와 동일
class HistoryTab extends StatefulWidget {
  final StorageService storage;

  const HistoryTab({
    super.key,
    required this.storage,
  });

  @override
  State<HistoryTab> createState() => HistoryTabState();
}

class HistoryTabState extends State<HistoryTab> {
  List<DeployHistory> _history = [];
  DeployHistory? _selectedHistory;

  @override
  void initState() {
    super.initState();
    _loadHistory();
  }

  /// 외부에서 호출 가능한 새로고침
  void refresh() {
    _loadHistory();
    setState(() {
      _selectedHistory = null;
    });
  }

  Future<void> _loadHistory() async {
    final history = await widget.storage.getAllHistory();
    setState(() {
      _history = history;
    });
  }

  Future<void> _handleDelete(int id) async {
    final confirmed = await showConfirmDialog(
      context,
      title: '삭제 확인',
      message: '이 배포 이력을 삭제하시겠습니까?',
      isDanger: true,
    );

    if (confirmed && mounted) {
      await widget.storage.deleteHistory(id);
      await _loadHistory();
      if (_selectedHistory?.id == id) {
        setState(() {
          _selectedHistory = null;
        });
      }
    }
  }

  Future<void> _handleDeleteAll() async {
    if (_history.isEmpty) return;

    final confirmed = await showConfirmDialog(
      context,
      title: '전체 삭제',
      message: '${_history.length}개의 배포 이력을 모두 삭제하시겠습니까?',
      isDanger: true,
    );

    if (confirmed && mounted) {
      await widget.storage.deleteAllHistory();
      await _loadHistory();
      setState(() {
        _selectedHistory = null;
      });
    }
  }

  String _formatDate(DateTime date) {
    return DateFormat('yyyy. M. d. a h:mm:ss', 'ko_KR').format(date);
  }

  Map<String, int> _getResultStats(List<RuleResult> results) {
    final total = results.length;
    final success =
        results.where((r) => r.status.toLowerCase() == 'ok').length;
    final fail = total - success;
    return {'total': total, 'success': success, 'fail': fail};
  }

  @override
  Widget build(BuildContext context) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // 왼쪽: 이력 목록
        SizedBox(
          width: 300,
          child: FmsCard(
            title: '배포 이력',
            trailing: FmsButton(
              text: '전체 삭제',
              type: FmsButtonType.danger,
              onPressed: _history.isEmpty ? null : _handleDeleteAll,
            ),
            child: _history.isEmpty
                ? const Center(
                    child: Text(
                      '배포 이력이 없습니다',
                      style: TextStyle(color: AppTheme.textMuted),
                    ),
                  )
                : ListView.builder(
                    itemCount: _history.length,
                    itemBuilder: (context, index) {
                      final h = _history[index];
                      final isSelected = _selectedHistory?.id == h.id;
                      return _HistoryListItem(
                        history: h,
                        isSelected: isSelected,
                        onTap: () => setState(() => _selectedHistory = h),
                        formatDate: _formatDate,
                      );
                    },
                  ),
          ),
        ),
        const SizedBox(width: 24),
        // 오른쪽: 상세 정보
        Expanded(
          child: FmsCard(
            title: '상세 정보',
            child: _selectedHistory != null
                ? _buildDetailView()
                : const EmptyState(
                    icon: '📜',
                    message: '왼쪽에서 배포 이력을 선택하세요',
                  ),
          ),
        ),
      ],
    );
  }

  Widget _buildDetailView() {
    final h = _selectedHistory!;
    final stats = _getResultStats(h.results);

    return SingleChildScrollView(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // 기본 정보 테이블
          Container(
            decoration: BoxDecoration(
              border: Border.all(color: AppTheme.borderColor),
              borderRadius: BorderRadius.circular(6),
            ),
            child: Table(
              columnWidths: const {
                0: FixedColumnWidth(120),
                1: FlexColumnWidth(),
              },
              border: TableBorder(
                horizontalInside: BorderSide(color: AppTheme.borderColor),
              ),
              children: [
                _buildInfoRow('장비 IP', h.deviceIp),
                _buildInfoRow('템플릿 버전', h.templateVersion),
                _buildInfoRow('배포 시간', _formatDate(h.timestamp)),
                TableRow(
                  children: [
                    const TableCell(
                      child: Padding(
                        padding: EdgeInsets.all(12),
                        child: Text(
                          '상태',
                          style: TextStyle(
                            fontWeight: FontWeight.w600,
                            color: AppTheme.textPrimary,
                          ),
                        ),
                      ),
                    ),
                    TableCell(
                      child: Padding(
                        padding: const EdgeInsets.all(12),
                        child: StatusBadge.fromStatus(h.status),
                      ),
                    ),
                  ],
                ),
                _buildInfoRow('결과',
                    '총 ${stats['total']}개 / 성공 ${stats['success']}개 / 실패 ${stats['fail']}개'),
              ],
            ),
          ),
          const SizedBox(height: 20),
          // 규칙별 결과
          if (h.results.isNotEmpty) ...[
            const Text(
              '규칙별 결과',
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.w600,
                color: AppTheme.primaryColor,
              ),
            ),
            const SizedBox(height: 12),
            Container(
              constraints: const BoxConstraints(maxHeight: 300),
              decoration: BoxDecoration(
                border: Border.all(color: AppTheme.borderColor),
                borderRadius: BorderRadius.circular(6),
              ),
              child: SingleChildScrollView(
                child: Table(
                  columnWidths: const {
                    0: FlexColumnWidth(2),
                    1: FixedColumnWidth(80),
                    2: FlexColumnWidth(1),
                  },
                  border: TableBorder(
                    horizontalInside: BorderSide(color: AppTheme.borderColor),
                  ),
                  children: [
                    // 헤더
                    TableRow(
                      decoration: BoxDecoration(color: AppTheme.borderColor),
                      children: [
                        _buildHeaderCell('규칙'),
                        _buildHeaderCell('결과'),
                        _buildHeaderCell('사유'),
                      ],
                    ),
                    // 데이터
                    ...h.results.map((r) => TableRow(
                          children: [
                            TableCell(
                              child: Padding(
                                padding: const EdgeInsets.all(12),
                                child: Text(
                                  r.rule,
                                  style: const TextStyle(
                                    fontFamily: 'monospace',
                                    fontSize: 12,
                                    color: AppTheme.textPrimary,
                                  ),
                                ),
                              ),
                            ),
                            TableCell(
                              child: Padding(
                                padding: const EdgeInsets.all(12),
                                child: StatusBadge.fromStatus(r.status),
                              ),
                            ),
                            TableCell(
                              child: Padding(
                                padding: const EdgeInsets.all(12),
                                child: Text(
                                  r.reason.isEmpty ? '-' : r.reason,
                                  style: const TextStyle(
                                    color: AppTheme.textPrimary,
                                  ),
                                ),
                              ),
                            ),
                          ],
                        )),
                  ],
                ),
              ),
            ),
          ],
          const SizedBox(height: 20),
          // 삭제 버튼
          FmsButton(
            text: '이력 삭제',
            type: FmsButtonType.danger,
            onPressed: () => _handleDelete(h.id),
          ),
        ],
      ),
    );
  }

  TableRow _buildInfoRow(String label, String value) {
    return TableRow(
      children: [
        TableCell(
          child: Padding(
            padding: const EdgeInsets.all(12),
            child: Text(
              label,
              style: const TextStyle(
                fontWeight: FontWeight.w600,
                color: AppTheme.textPrimary,
              ),
            ),
          ),
        ),
        TableCell(
          child: Padding(
            padding: const EdgeInsets.all(12),
            child: Text(
              value,
              style: const TextStyle(color: AppTheme.textPrimary),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildHeaderCell(String text) {
    return TableCell(
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Text(
          text,
          style: const TextStyle(
            fontWeight: FontWeight.w600,
            color: Colors.white,
          ),
        ),
      ),
    );
  }
}

/// 이력 목록 아이템
class _HistoryListItem extends StatefulWidget {
  final DeployHistory history;
  final bool isSelected;
  final VoidCallback onTap;
  final String Function(DateTime) formatDate;

  const _HistoryListItem({
    required this.history,
    required this.isSelected,
    required this.onTap,
    required this.formatDate,
  });

  @override
  State<_HistoryListItem> createState() => _HistoryListItemState();
}

class _HistoryListItemState extends State<_HistoryListItem> {
  bool _isHovered = false;

  @override
  Widget build(BuildContext context) {
    return MouseRegion(
      onEnter: (_) => setState(() => _isHovered = true),
      onExit: (_) => setState(() => _isHovered = false),
      child: GestureDetector(
        onTap: widget.onTap,
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
          decoration: BoxDecoration(
            color: widget.isSelected
                ? AppTheme.primaryColor.withValues(alpha: 0.2)
                : _isHovered
                    ? AppTheme.primaryColor.withValues(alpha: 0.1)
                    : Colors.transparent,
            border: Border(
              left: widget.isSelected
                  ? const BorderSide(color: AppTheme.primaryColor, width: 3)
                  : BorderSide.none,
              bottom: const BorderSide(color: AppTheme.borderColor),
            ),
          ),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      widget.history.deviceIp,
                      style: const TextStyle(
                        fontWeight: FontWeight.w500,
                        color: AppTheme.textPrimary,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      widget.formatDate(widget.history.timestamp),
                      style: const TextStyle(
                        fontSize: 12,
                        color: AppTheme.textDisabled,
                      ),
                    ),
                  ],
                ),
              ),
              StatusBadge.fromStatus(widget.history.status),
            ],
          ),
        ),
      ),
    );
  }
}
