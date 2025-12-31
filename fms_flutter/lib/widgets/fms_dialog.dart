import 'package:flutter/material.dart';
import '../theme/app_theme.dart';
import 'fms_button.dart';

/// FMS 다이얼로그 위젯 - Wails의 .modal 클래스와 동일
class FmsDialog extends StatelessWidget {
  final String title;
  final Widget content;
  final List<Widget>? actions;
  final double? width;
  final double? maxHeight;

  const FmsDialog({
    super.key,
    required this.title,
    required this.content,
    this.actions,
    this.width,
    this.maxHeight,
  });

  @override
  Widget build(BuildContext context) {
    return Dialog(
      backgroundColor: AppTheme.surfaceColor,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: const BorderSide(color: AppTheme.borderColor),
      ),
      child: Container(
        width: width ?? 400,
        constraints: BoxConstraints(
          maxHeight: maxHeight ?? MediaQuery.of(context).size.height * 0.8,
        ),
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // 헤더
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  title,
                  style: const TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.w600,
                    color: AppTheme.primaryColor,
                  ),
                ),
                IconButton(
                  onPressed: () => Navigator.of(context).pop(),
                  icon: const Icon(Icons.close, color: AppTheme.textSecondary),
                  padding: EdgeInsets.zero,
                  constraints: const BoxConstraints(),
                ),
              ],
            ),
            const SizedBox(height: 20),
            // 컨텐츠
            Flexible(child: content),
            // 액션 버튼들
            if (actions != null) ...[
              const SizedBox(height: 20),
              Row(
                mainAxisAlignment: MainAxisAlignment.end,
                children: actions!
                    .map((a) => Padding(
                          padding: const EdgeInsets.only(left: 8),
                          child: a,
                        ))
                    .toList(),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

/// 확인 다이얼로그 표시
Future<bool> showConfirmDialog(
  BuildContext context, {
  required String title,
  required String message,
  String confirmText = '확인',
  String cancelText = '취소',
  bool isDanger = false,
}) async {
  final result = await showDialog<bool>(
    context: context,
    builder: (context) => FmsDialog(
      title: title,
      content: Text(
        message,
        style: const TextStyle(color: AppTheme.textPrimary),
      ),
      actions: [
        FmsButton(
          text: cancelText,
          type: FmsButtonType.secondary,
          onPressed: () => Navigator.of(context).pop(false),
        ),
        FmsButton(
          text: confirmText,
          type: isDanger ? FmsButtonType.danger : FmsButtonType.primary,
          onPressed: () => Navigator.of(context).pop(true),
        ),
      ],
    ),
  );
  return result ?? false;
}

/// 알림 다이얼로그 표시
Future<void> showAlertDialog(
  BuildContext context, {
  required String title,
  required String message,
}) async {
  await showDialog(
    context: context,
    builder: (context) => FmsDialog(
      title: title,
      content: Text(
        message,
        style: const TextStyle(color: AppTheme.textPrimary),
      ),
      actions: [
        FmsButton(
          text: '확인',
          onPressed: () => Navigator.of(context).pop(),
        ),
      ],
    ),
  );
}
