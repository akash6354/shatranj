import 'package:flutter/material.dart';

import '../theme/app_theme.dart';

class AppSectionHeader extends StatelessWidget {
  const AppSectionHeader({
    super.key,
    required this.title,
    this.subtitle,
    this.actionLabel,
    this.onAction,
    this.leading,
  });

  final String title;
  final String? subtitle;
  final String? actionLabel;
  final VoidCallback? onAction;
  final Widget? leading;

  @override
  Widget build(BuildContext context) => Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (leading != null) ...[
            leading!,
            const SizedBox(width: AppSpacing.sm),
          ],
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(title, style: AppTextStyles.title),
                if (subtitle != null) ...[
                  const SizedBox(height: AppSpacing.xxs),
                  Text(subtitle!, style: AppTextStyles.caption),
                ],
              ],
            ),
          ),
          if (actionLabel != null)
            TextButton(
              onPressed: onAction,
              style: TextButton.styleFrom(
                foregroundColor: AppColors.gold,
                padding: const EdgeInsets.symmetric(horizontal: AppSpacing.xs),
                minimumSize: Size.zero,
                tapTargetSize: MaterialTapTargetSize.shrinkWrap,
              ),
              child: Text(actionLabel!, style: AppTextStyles.caption.copyWith(color: AppColors.gold, fontWeight: FontWeight.w700)),
            ),
        ],
      );
}

class AppTopBar extends StatelessWidget {
  const AppTopBar({
    super.key,
    this.title,
    this.subtitle,
    this.leading,
    this.actions = const [],
  });

  final String? title;
  final String? subtitle;
  final Widget? leading;
  final List<Widget> actions;

  @override
  Widget build(BuildContext context) => Row(
        children: [
          if (leading != null) ...[
            leading!,
            const SizedBox(width: AppSpacing.sm),
          ],
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                if (title != null) Text(title!, style: AppTextStyles.title),
                if (subtitle != null) Text(subtitle!, style: AppTextStyles.caption),
              ],
            ),
          ),
          ...actions,
        ],
      );
}
