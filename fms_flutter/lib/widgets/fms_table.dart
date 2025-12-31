import 'package:flutter/material.dart';
import '../theme/app_theme.dart';

/// FMS 테이블 위젯 - Wails의 .table 클래스와 동일
class FmsTable extends StatelessWidget {
  final List<String> headers;
  final List<TableRow> rows;
  final List<double>? columnWidths;

  const FmsTable({
    super.key,
    required this.headers,
    required this.rows,
    this.columnWidths,
  });

  @override
  Widget build(BuildContext context) {
    return Table(
      columnWidths: columnWidths != null
          ? Map.fromIterables(
              List.generate(columnWidths!.length, (i) => i),
              columnWidths!.map((w) => w == -1
                  ? const FlexColumnWidth()
                  : FixedColumnWidth(w)),
            )
          : null,
      border: TableBorder(
        horizontalInside: BorderSide(color: AppTheme.borderColor),
      ),
      children: [
        // 헤더 행
        TableRow(
          decoration: BoxDecoration(
            color: AppTheme.borderColor,
          ),
          children: headers
              .map((h) => TableCell(
                    child: Padding(
                      padding: const EdgeInsets.all(12),
                      child: Text(
                        h,
                        style: const TextStyle(
                          color: Colors.white,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ),
                  ))
              .toList(),
        ),
        // 데이터 행
        ...rows,
      ],
    );
  }
}

/// FMS 테이블 셀 위젯
class FmsTableCell extends StatelessWidget {
  final Widget child;
  final EdgeInsetsGeometry? padding;
  final Alignment alignment;

  const FmsTableCell({
    super.key,
    required this.child,
    this.padding,
    this.alignment = Alignment.centerLeft,
  });

  @override
  Widget build(BuildContext context) {
    return TableCell(
      verticalAlignment: TableCellVerticalAlignment.middle,
      child: Padding(
        padding: padding ?? const EdgeInsets.all(12),
        child: Align(
          alignment: alignment,
          child: child,
        ),
      ),
    );
  }
}

/// 호버 가능한 테이블 행
class FmsHoverTableRow extends StatefulWidget {
  final List<Widget> cells;
  final VoidCallback? onTap;
  final bool isSelected;

  const FmsHoverTableRow({
    super.key,
    required this.cells,
    this.onTap,
    this.isSelected = false,
  });

  @override
  State<FmsHoverTableRow> createState() => _FmsHoverTableRowState();
}

class _FmsHoverTableRowState extends State<FmsHoverTableRow> {
  bool _isHovered = false;

  @override
  Widget build(BuildContext context) {
    return MouseRegion(
      onEnter: (_) => setState(() => _isHovered = true),
      onExit: (_) => setState(() => _isHovered = false),
      child: GestureDetector(
        onTap: widget.onTap,
        child: Container(
          decoration: BoxDecoration(
            color: widget.isSelected
                ? AppTheme.primaryColor.withValues(alpha: 0.2)
                : _isHovered
                    ? AppTheme.primaryColor.withValues(alpha: 0.1)
                    : Colors.transparent,
            border: widget.isSelected
                ? const Border(
                    left: BorderSide(color: AppTheme.primaryColor, width: 3),
                  )
                : null,
          ),
          child: Row(
            children: widget.cells
                .map((cell) => Expanded(
                      child: Padding(
                        padding: const EdgeInsets.all(12),
                        child: cell,
                      ),
                    ))
                .toList(),
          ),
        ),
      ),
    );
  }
}
