import 'package:flutter/material.dart';

import '../../../core/theme/app_theme.dart';

class PlaceholderScreen extends StatelessWidget {
  const PlaceholderScreen({
    required this.title,
    required this.icon,
    this.message,
    super.key,
  });

  final String title;
  final IconData icon;
  final String? message;

  @override
  Widget build(BuildContext context) => LayoutBuilder(
        builder: (context, constraints) => SingleChildScrollView(
          padding: const EdgeInsets.all(AppSpacing.page),
          child: Center(
            child: ConstrainedBox(
              constraints: BoxConstraints(
                maxWidth: constraints.maxWidth > 700 ? 620 : double.infinity,
                minHeight: constraints.maxHeight - AppSpacing.page * 2,
              ),
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(icon, color: AppColors.gold, size: 54),
                  const SizedBox(height: AppSpacing.md),
                  Text(title, style: AppTextStyles.headline),
                  const SizedBox(height: AppSpacing.xs),
                  Text(
                    message ?? 'This section is ready for the next phase.',
                    style: AppTextStyles.body,
                    textAlign: TextAlign.center,
                  ),
                ],
              ),
            ),
          ),
        ),
      );
}
