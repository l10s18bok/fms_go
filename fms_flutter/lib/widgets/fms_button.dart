import 'package:flutter/material.dart';
import '../theme/app_theme.dart';

/// FMS 버튼 위젯
enum FmsButtonType { primary, secondary, danger }

class FmsButton extends StatelessWidget {
  final String text;
  final VoidCallback? onPressed;
  final FmsButtonType type;
  final bool isLoading;
  final IconData? icon;
  final double? width;

  const FmsButton({
    super.key,
    required this.text,
    this.onPressed,
    this.type = FmsButtonType.primary,
    this.isLoading = false,
    this.icon,
    this.width,
  });

  @override
  Widget build(BuildContext context) {
    Color bgColor;
    Color textColor = Colors.white;

    switch (type) {
      case FmsButtonType.primary:
        bgColor = AppTheme.primaryColor;
        break;
      case FmsButtonType.secondary:
        bgColor = AppTheme.secondaryButtonColor;
        break;
      case FmsButtonType.danger:
        bgColor = AppTheme.dangerButtonColor;
        break;
    }

    return SizedBox(
      width: width,
      child: ElevatedButton(
        onPressed: isLoading ? null : onPressed,
        style: ElevatedButton.styleFrom(
          backgroundColor: bgColor,
          foregroundColor: textColor,
          disabledBackgroundColor: bgColor.withValues(alpha: 0.5),
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(6),
          ),
        ),
        child: isLoading
            ? const SizedBox(
                width: 16,
                height: 16,
                child: CircularProgressIndicator(
                  strokeWidth: 2,
                  valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                ),
              )
            : Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  if (icon != null) ...[
                    Icon(icon, size: 16),
                    const SizedBox(width: 6),
                  ],
                  Text(text),
                ],
              ),
      ),
    );
  }
}
