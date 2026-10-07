import 'package:flutter/material.dart';

import '../theme/app_theme.dart';

class AppPill extends StatelessWidget {
  const AppPill({
    super.key,
    required this.label,
    this.icon,
    this.color,
    this.backgroundColor,
    this.onTap,
  });

  final String label;
  final IconData? icon;
  final Color? color;
  final Color? backgroundColor;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final content = Padding(
      padding: const EdgeInsets.symmetric(
        horizontal: AppSpacing.sm,
        vertical: AppSpacing.xs,
      ),
      child: Row(
        children: [
          if (icon != null) ...[
            Icon(icon, size: 14, color: color ?? AppColors.gold),
            const SizedBox(width: AppSpacing.xs),
          ],
          Flexible(
            child: Text(
              label,
              overflow: TextOverflow.ellipsis,
              style: AppTextStyles.caption.copyWith(
                color: color ?? AppColors.textSecondary,
                fontWeight: FontWeight.w600,
              ),
            ),
          ),
        ],
      ),
    );
    return Material(
      color: backgroundColor ?? AppColors.surfaceMuted,
      shape: RoundedRectangleBorder(
        borderRadius: AppRadius.pillRadius,
        side: const BorderSide(color: AppColors.border),
      ),
      child: InkWell(
        onTap: onTap,
        borderRadius: AppRadius.pillRadius,
        child: content,
      ),
    );
  }
}

class AppChoicePill extends StatelessWidget {
  const AppChoicePill({
    super.key,
    required this.label,
    required this.selected,
    required this.onSelected,
  });

  final String label;
  final bool selected;
  final ValueChanged<bool> onSelected;

  @override
  Widget build(BuildContext context) => GestureDetector(
        onTap: () => onSelected(!selected),
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 160),
          padding: const EdgeInsets.symmetric(
            horizontal: AppSpacing.md,
            vertical: AppSpacing.sm,
          ),
          decoration: BoxDecoration(
            color: selected ? AppColors.gold : AppColors.surfaceMuted,
            borderRadius: AppRadius.pillRadius,
            border: Border.all(
              color: selected ? AppColors.gold : AppColors.border,
            ),
          ),
          child: Text(
            label,
            style: AppTextStyles.caption.copyWith(
              color: selected ? AppColors.background : AppColors.textSecondary,
              fontWeight: FontWeight.w700,
            ),
          ),
        ),
      );
}
