import 'package:flutter/material.dart';

import '../theme/app_theme.dart';

enum AppButtonVariant { primary, secondary, outline, ghost, danger }

class AppButton extends StatelessWidget {
  const AppButton({
    super.key,
    required this.label,
    this.onPressed,
    this.icon,
    this.variant = AppButtonVariant.primary,
    this.expand = false,
    this.loading = false,
  });

  final String label;
  final VoidCallback? onPressed;
  final IconData? icon;
  final AppButtonVariant variant;
  final bool expand;
  final bool loading;

  @override
  Widget build(BuildContext context) {
    final button = _buildButton();
    return expand ? SizedBox(width: double.infinity, child: button) : button;
  }

  Widget _buildButton() {
    final child = loading
        ? const SizedBox(
            width: 18,
            height: 18,
            child: CircularProgressIndicator(strokeWidth: 2),
          )
        : Row(
            mainAxisSize: expand ? MainAxisSize.max : MainAxisSize.min,
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              if (icon != null) ...[
                Icon(icon, size: 18),
                const SizedBox(width: AppSpacing.xs),
              ],
              Flexible(
                child: Text(label, overflow: TextOverflow.ellipsis),
              ),
            ],
          );

    switch (variant) {
      case AppButtonVariant.primary:
        return FilledButton(
          onPressed: loading ? null : onPressed,
          style: FilledButton.styleFrom(
            minimumSize: const Size(0, 46),
            padding: EdgeInsets.symmetric(horizontal: expand ? AppSpacing.md : AppSpacing.lg),
            shape: RoundedRectangleBorder(borderRadius: AppRadius.buttonRadius),
            textStyle: AppTextStyles.button,
          ),
          child: child,
        );
      case AppButtonVariant.secondary:
        return FilledButton(
          onPressed: loading ? null : onPressed,
          style: FilledButton.styleFrom(
            backgroundColor: AppColors.mint,
            foregroundColor: AppColors.background,
            minimumSize: const Size(0, 46),
            padding: EdgeInsets.symmetric(horizontal: expand ? AppSpacing.md : AppSpacing.lg),
            shape: RoundedRectangleBorder(borderRadius: AppRadius.buttonRadius),
            textStyle: AppTextStyles.button,
          ),
          child: child,
        );
      case AppButtonVariant.outline:
        return OutlinedButton(
          onPressed: loading ? null : onPressed,
          style: OutlinedButton.styleFrom(
            foregroundColor: AppColors.text,
            minimumSize: const Size(0, 46),
            padding: EdgeInsets.symmetric(horizontal: expand ? AppSpacing.md : AppSpacing.lg),
            side: const BorderSide(color: AppColors.borderStrong),
            shape: RoundedRectangleBorder(borderRadius: AppRadius.buttonRadius),
            textStyle: AppTextStyles.button.copyWith(color: AppColors.text),
          ),
          child: child,
        );
      case AppButtonVariant.ghost:
        return TextButton(
          onPressed: loading ? null : onPressed,
          style: TextButton.styleFrom(
            foregroundColor: AppColors.gold,
            minimumSize: const Size(0, 44),
            padding: const EdgeInsets.symmetric(horizontal: AppSpacing.sm),
            textStyle: AppTextStyles.button.copyWith(color: AppColors.gold),
          ),
          child: child,
        );
      case AppButtonVariant.danger:
        return FilledButton(
          onPressed: loading ? null : onPressed,
          style: FilledButton.styleFrom(
            backgroundColor: AppColors.danger,
            foregroundColor: AppColors.background,
            minimumSize: const Size(0, 46),
            padding: EdgeInsets.symmetric(horizontal: expand ? AppSpacing.md : AppSpacing.lg),
            shape: RoundedRectangleBorder(borderRadius: AppRadius.buttonRadius),
            textStyle: AppTextStyles.button,
          ),
          child: child,
        );
    }
  }
}

class AppIconButton extends StatelessWidget {
  const AppIconButton({
    super.key,
    required this.icon,
    this.onPressed,
    this.tooltip,
    this.selected = false,
  });

  final IconData icon;
  final VoidCallback? onPressed;
  final String? tooltip;
  final bool selected;

  @override
  Widget build(BuildContext context) {
    final button = IconButton(
      onPressed: onPressed,
      tooltip: tooltip,
      icon: Icon(icon, size: 21),
      style: IconButton.styleFrom(
        foregroundColor: selected ? AppColors.gold : AppColors.textSecondary,
        backgroundColor:
            selected ? AppColors.gold.withValues(alpha: .14) : AppColors.surfaceMuted,
        shape: RoundedRectangleBorder(borderRadius: AppRadius.small),
      ),
    );
    return button;
  }
}
