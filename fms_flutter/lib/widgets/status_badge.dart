import 'package:flutter/material.dart';
import '../theme/app_theme.dart';

/// 상태 뱃지 위젯 - Wails의 badge 클래스와 동일
class StatusBadge extends StatelessWidget {
  final String text;
  final BadgeType type;

  const StatusBadge({
    super.key,
    required this.text,
    this.type = BadgeType.info,
  });

  /// 서버/배포 상태에 따른 뱃지 생성
  factory StatusBadge.fromStatus(String status) {
    final lowerStatus = status.toLowerCase();

    if (lowerStatus == 'running' || lowerStatus == 'success' || lowerStatus == 'ok') {
      return StatusBadge(
        text: lowerStatus == 'running' ? '정상' : '성공',
        type: BadgeType.success,
      );
    } else if (lowerStatus == 'stop' || lowerStatus == 'fail' || lowerStatus == 'error' || lowerStatus == 'unfind' || lowerStatus == 'validation') {
      String displayText;
      if (lowerStatus == 'stop') {
        displayText = '정지';
      } else if (lowerStatus == 'fail' || lowerStatus == 'error' || lowerStatus == 'unfind' || lowerStatus == 'validation') {
        displayText = '실패';
      } else {
        displayText = status;
      }
      return StatusBadge(
        text: displayText,
        type: BadgeType.danger,
      );
    } else if (lowerStatus == 'write') {
      return const StatusBadge(
        text: '진행중',
        type: BadgeType.warning,
      );
    } else {
      return StatusBadge(
        text: status.isEmpty ? '-' : status,
        type: BadgeType.info,
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    Color bgColor;
    Color textColor;

    switch (type) {
      case BadgeType.success:
        bgColor = AppTheme.successColor.withValues(alpha: 0.2);
        textColor = AppTheme.successColor;
        break;
      case BadgeType.danger:
        bgColor = AppTheme.dangerColor.withValues(alpha: 0.2);
        textColor = AppTheme.dangerColor;
        break;
      case BadgeType.warning:
        bgColor = AppTheme.warningColor.withValues(alpha: 0.2);
        textColor = AppTheme.warningColor;
        break;
      case BadgeType.info:
        bgColor = AppTheme.infoColor.withValues(alpha: 0.2);
        textColor = AppTheme.infoColor;
        break;
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(
        color: bgColor,
        borderRadius: BorderRadius.circular(12),
      ),
      child: Text(
        text,
        style: TextStyle(
          color: textColor,
          fontSize: 12,
          fontWeight: FontWeight.w500,
        ),
      ),
    );
  }
}

enum BadgeType { success, danger, warning, info }
