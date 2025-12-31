import 'package:flutter/material.dart';
import '../models/models.dart';
import '../services/services.dart';
import '../theme/app_theme.dart';
import '../widgets/widgets.dart';

/// 템플릿 관리 탭 - Wails의 TemplateTab.tsx와 동일
class TemplateTab extends StatefulWidget {
  final StorageService storage;
  final VoidCallback? onRefresh;

  const TemplateTab({
    super.key,
    required this.storage,
    this.onRefresh,
  });

  @override
  State<TemplateTab> createState() => TemplateTabState();
}

class TemplateTabState extends State<TemplateTab> {
  List<Template> _templates = [];
  String? _selectedVersion;
  bool _isNew = false;

  final _versionController = TextEditingController();
  final _contentsController = TextEditingController();

  @override
  void initState() {
    super.initState();
    _loadTemplates();
  }

  @override
  void dispose() {
    _versionController.dispose();
    _contentsController.dispose();
    super.dispose();
  }

  /// 외부에서 호출 가능한 새로고침
  void refresh() {
    _loadTemplates();
    setState(() {
      _selectedVersion = null;
      _isNew = false;
      _versionController.clear();
      _contentsController.clear();
    });
  }

  Future<void> _loadTemplates() async {
    final templates = await widget.storage.getAllTemplates();
    setState(() {
      _templates = templates;
    });
  }

  Future<void> _handleSelect(String version) async {
    final template = await widget.storage.getTemplate(version);
    if (template != null) {
      setState(() {
        _selectedVersion = version;
        _isNew = false;
        _versionController.text = template.version;
        _contentsController.text = template.contents;
      });
    }
  }

  void _handleNew() {
    setState(() {
      _selectedVersion = null;
      _isNew = true;
      _versionController.clear();
      _contentsController.clear();
    });
  }

  Future<void> _handleSave() async {
    final version = _versionController.text.trim();
    final contents = _contentsController.text.trim();

    if (version.isEmpty) {
      _showMessage('버전을 입력하세요.');
      return;
    }
    if (contents.isEmpty) {
      _showMessage('규칙 내용을 입력해주세요.');
      return;
    }

    await widget.storage.saveTemplate(version, contents);
    await _loadTemplates();
    setState(() {
      _selectedVersion = version;
      _isNew = false;
    });
    _showMessage('템플릿이 저장되었습니다.');
  }

  Future<void> _handleDelete() async {
    if (_selectedVersion == null) {
      _showMessage('삭제할 템플릿이 선택되지 않았습니다.');
      return;
    }

    final confirmed = await showConfirmDialog(
      context,
      title: '삭제 확인',
      message: '"$_selectedVersion" 템플릿을 삭제하시겠습니까?',
      isDanger: true,
    );

    if (confirmed) {
      await widget.storage.deleteTemplate(_selectedVersion!);
      await _loadTemplates();
      setState(() {
        _selectedVersion = null;
        _isNew = false;
        _versionController.clear();
        _contentsController.clear();
      });
      _showMessage('템플릿이 삭제되었습니다.');
    }
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
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // 왼쪽: 템플릿 목록
        SizedBox(
          width: 300,
          child: FmsCard(
            title: '템플릿 목록',
            child: Column(
              children: [
                // 새 템플릿 버튼
                SizedBox(
                  width: double.infinity,
                  child: FmsButton(
                    text: '+ 새 템플릿',
                    onPressed: _handleNew,
                  ),
                ),
                const SizedBox(height: 16),
                // 템플릿 리스트
                Expanded(
                  child: _templates.isEmpty
                      ? const Center(
                          child: Text(
                            '템플릿이 없습니다',
                            style: TextStyle(color: AppTheme.textMuted),
                          ),
                        )
                      : ListView.builder(
                          itemCount: _templates.length,
                          itemBuilder: (context, index) {
                            final template = _templates[index];
                            final isSelected =
                                _selectedVersion == template.version;
                            return _TemplateListItem(
                              version: template.version,
                              isSelected: isSelected,
                              onTap: () => _handleSelect(template.version),
                            );
                          },
                        ),
                ),
              ],
            ),
          ),
        ),
        const SizedBox(width: 24),
        // 오른쪽: 템플릿 편집
        Expanded(
          child: FmsCard(
            title: _isNew
                ? '새 템플릿'
                : _selectedVersion != null
                    ? '템플릿: $_selectedVersion'
                    : '템플릿 선택',
            child: _selectedVersion != null || _isNew
                ? _buildEditForm()
                : const EmptyState(
                    icon: '📋',
                    message: '왼쪽에서 템플릿을 선택하거나',
                    subMessage: '새 템플릿을 생성하세요',
                  ),
          ),
        ),
      ],
    );
  }

  Widget _buildEditForm() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // 버전 입력
        const Text(
          '버전',
          style: TextStyle(
            color: AppTheme.textSecondary,
            fontSize: 13,
          ),
        ),
        const SizedBox(height: 6),
        TextField(
          controller: _versionController,
          decoration: const InputDecoration(
            hintText: '예: v1.0.0',
          ),
        ),
        const SizedBox(height: 16),
        // 규칙 내용 입력
        const Text(
          '규칙 내용',
          style: TextStyle(
            color: AppTheme.textSecondary,
            fontSize: 13,
          ),
        ),
        const SizedBox(height: 6),
        Expanded(
          child: TextField(
            controller: _contentsController,
            maxLines: null,
            expands: true,
            textAlignVertical: TextAlignVertical.top,
            style: const TextStyle(
              fontFamily: 'monospace',
              fontSize: 13,
            ),
            decoration: const InputDecoration(
              hintText: '방화벽 규칙을 입력하세요...',
              alignLabelWithHint: true,
            ),
          ),
        ),
        const SizedBox(height: 16),
        // 버튼 그룹
        Row(
          children: [
            FmsButton(
              text: '저장',
              onPressed: _handleSave,
            ),
            if (!_isNew) ...[
              const SizedBox(width: 8),
              FmsButton(
                text: '삭제',
                type: FmsButtonType.danger,
                onPressed: _handleDelete,
              ),
            ],
          ],
        ),
      ],
    );
  }
}

/// 템플릿 목록 아이템
class _TemplateListItem extends StatefulWidget {
  final String version;
  final bool isSelected;
  final VoidCallback onTap;

  const _TemplateListItem({
    required this.version,
    required this.isSelected,
    required this.onTap,
  });

  @override
  State<_TemplateListItem> createState() => _TemplateListItemState();
}

class _TemplateListItemState extends State<_TemplateListItem> {
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
          child: Text(
            widget.version,
            style: const TextStyle(
              color: AppTheme.textPrimary,
            ),
          ),
        ),
      ),
    );
  }
}
