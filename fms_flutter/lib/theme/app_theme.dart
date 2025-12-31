import 'package:flutter/material.dart';

/// FMS 앱 테마 - Wails 버전 CSS와 동일한 컬러 스킴
class AppTheme {
  // 주요 색상 (Wails CSS와 동일)
  static const Color primaryColor = Color(0xFFE94560); // #e94560
  static const Color primaryDark = Color(0xFFD63D56); // #d63d56

  static const Color backgroundColor = Color(0xFF1A1A2E); // #1a1a2e
  static const Color surfaceColor = Color(0xFF16213E); // #16213e
  static const Color borderColor = Color(0xFF0F3460); // #0f3460

  static const Color textPrimary = Color(0xFFEEEEEE); // #eee
  static const Color textSecondary = Color(0xFFAAAAAA); // #aaa
  static const Color textMuted = Color(0xFF666666); // #666
  static const Color textDisabled = Color(0xFF888888); // #888

  // 상태 색상
  static const Color successColor = Color(0xFF27AE60); // #27ae60
  static const Color dangerColor = Color(0xFFE74C3C); // #e74c3c
  static const Color warningColor = Color(0xFFF1C40F); // #f1c40f
  static const Color infoColor = Color(0xFF3498DB); // #3498db

  // 버튼 색상
  static const Color secondaryButtonColor = Color(0xFF0F3460); // #0f3460
  static const Color dangerButtonColor = Color(0xFFC0392B); // #c0392b

  static ThemeData get darkTheme {
    return ThemeData(
      useMaterial3: true,
      brightness: Brightness.dark,
      colorScheme: const ColorScheme.dark(
        primary: primaryColor,
        secondary: primaryColor,
        surface: surfaceColor,
        error: dangerColor,
      ),
      scaffoldBackgroundColor: backgroundColor,
      appBarTheme: const AppBarTheme(
        backgroundColor: surfaceColor,
        foregroundColor: textPrimary,
        elevation: 0,
      ),
      cardTheme: CardThemeData(
        color: surfaceColor,
        elevation: 0,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(8),
          side: const BorderSide(color: borderColor),
        ),
      ),
      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ElevatedButton.styleFrom(
          backgroundColor: primaryColor,
          foregroundColor: Colors.white,
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(6),
          ),
        ),
      ),
      outlinedButtonTheme: OutlinedButtonThemeData(
        style: OutlinedButton.styleFrom(
          foregroundColor: textPrimary,
          backgroundColor: secondaryButtonColor,
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(6),
          ),
          side: BorderSide.none,
        ),
      ),
      textButtonTheme: TextButtonThemeData(
        style: TextButton.styleFrom(
          foregroundColor: textSecondary,
        ),
      ),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: backgroundColor,
        contentPadding:
            const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(6),
          borderSide: const BorderSide(color: borderColor),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(6),
          borderSide: const BorderSide(color: borderColor),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(6),
          borderSide: const BorderSide(color: primaryColor),
        ),
        hintStyle: const TextStyle(color: textMuted),
        labelStyle: const TextStyle(color: textSecondary),
      ),
      dialogTheme: DialogThemeData(
        backgroundColor: surfaceColor,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(12),
          side: const BorderSide(color: borderColor),
        ),
      ),
      tabBarTheme: TabBarThemeData(
        labelColor: Colors.white,
        unselectedLabelColor: textSecondary,
        indicator: BoxDecoration(
          color: primaryColor,
          borderRadius: BorderRadius.circular(6),
        ),
      ),
      dividerColor: borderColor,
      textTheme: const TextTheme(
        bodyLarge: TextStyle(color: textPrimary),
        bodyMedium: TextStyle(color: textPrimary),
        bodySmall: TextStyle(color: textSecondary),
        titleLarge: TextStyle(color: primaryColor, fontWeight: FontWeight.w600),
        titleMedium:
            TextStyle(color: primaryColor, fontWeight: FontWeight.w600),
        labelLarge: TextStyle(color: textPrimary),
      ),
      checkboxTheme: CheckboxThemeData(
        fillColor: WidgetStateProperty.resolveWith((states) {
          if (states.contains(WidgetState.selected)) {
            return primaryColor;
          }
          return Colors.transparent;
        }),
        side: const BorderSide(color: textSecondary),
      ),
      radioTheme: RadioThemeData(
        fillColor: WidgetStateProperty.resolveWith((states) {
          if (states.contains(WidgetState.selected)) {
            return primaryColor;
          }
          return textSecondary;
        }),
      ),
      popupMenuTheme: PopupMenuThemeData(
        color: surfaceColor,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(6),
          side: const BorderSide(color: borderColor),
        ),
      ),
    );
  }
}
