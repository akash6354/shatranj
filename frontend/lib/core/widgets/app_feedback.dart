import 'package:flutter/material.dart';

import '../theme/app_theme.dart';
import 'app_buttons.dart';

class AppLoading extends StatelessWidget {
  const AppLoading({super.key, this.message});

  final String? message;

  @override
  Widget build(BuildContext context) => Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const SizedBox(
              width: 28,
              height: 28,
              child: CircularProgressIndicator(strokeWidth: 2.5),
            ),
            if (message != null) ...[
              const SizedBox(height: AppSpacing.sm),
              Text(message!, style: AppTextStyles.caption),
            ],
          ],
        ),
      );
}

class AppError extends StatelessWidget {
  const AppError({
    super.key,
    this.title = 'Something went wrong',
    this.message,
    this.onRetry,
  });

  final String title;
  final String? message;
  final VoidCallback? onRetry;

  @override
  Widget build(BuildContext context) => Center(
        child: Padding(
          padding: const EdgeInsets.all(AppSpacing.xl),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Icon(Icons.error_outline_rounded, color: AppColors.danger, size: 36),
              const SizedBox(height: AppSpacing.sm),
              Text(title, style: AppTextStyles.titleSmall, textAlign: TextAlign.center),
              if (message != null) ...[
                const SizedBox(height: AppSpacing.xs),
                Text(message!, style: AppTextStyles.caption, textAlign: TextAlign.center),
              ],
              if (onRetry != null) ...[
                const SizedBox(height: AppSpacing.md),
                AppButton(
                  label: 'Try again',
                  icon: Icons.refresh_rounded,
                  variant: AppButtonVariant.outline,
                  onPressed: onRetry,
                ),
              ],
            ],
          ),
        ),
      );
}

class AppEmptyState extends StatelessWidget {
  const AppEmptyState({
    super.key,
    required this.title,
    this.message,
    this.icon = Icons.inbox_outlined,
  });

  final String title;
  final String? message;
  final IconData icon;

  @override
  Widget build(BuildContext context) => Center(
        child: Padding(
          padding: const EdgeInsets.all(AppSpacing.xl),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(icon, color: AppColors.muted, size: 38),
              const SizedBox(height: AppSpacing.sm),
              Text(title, style: AppTextStyles.titleSmall, textAlign: TextAlign.center),
              if (message != null) ...[
                const SizedBox(height: AppSpacing.xs),
                Text(message!, style: AppTextStyles.caption, textAlign: TextAlign.center),
              ],
            ],
          ),
        ),
      );
}
