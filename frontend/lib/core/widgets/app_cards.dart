import 'package:flutter/material.dart';

import '../theme/app_theme.dart';

class AppCard extends StatelessWidget {
  const AppCard({
    super.key,
    required this.child,
    this.padding = const EdgeInsets.all(AppSpacing.md),
    this.margin = EdgeInsets.zero,
    this.color,
    this.borderColor,
    this.onTap,
    this.elevated = false,
  });

  final Widget child;
  final EdgeInsetsGeometry padding;
  final EdgeInsetsGeometry margin;
  final Color? color;
  final Color? borderColor;
  final VoidCallback? onTap;
  final bool elevated;

  @override
  Widget build(BuildContext context) {
    final content = Padding(padding: padding, child: child);
    final decorated = Container(
      margin: margin,
      decoration: BoxDecoration(
        color: color ?? AppColors.surface,
        borderRadius: AppRadius.cardRadius,
        border: Border.all(color: borderColor ?? AppColors.border),
        boxShadow: elevated ? AppShadows.card : null,
      ),
      child: onTap == null
          ? content
          : Material(
              color: Colors.transparent,
              child: InkWell(
                onTap: onTap,
                borderRadius: AppRadius.cardRadius,
                child: content,
              ),
            ),
    );
    return decorated;
  }
}

class AppGradientCard extends StatelessWidget {
  const AppGradientCard({
    super.key,
    required this.child,
    this.padding = const EdgeInsets.all(AppSpacing.lg),
    this.onTap,
  });

  final Widget child;
  final EdgeInsetsGeometry padding;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) => Container(
        decoration: BoxDecoration(
          gradient: const LinearGradient(
            colors: [Color(0xFF16453E), AppColors.surface],
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
          ),
          borderRadius: AppRadius.cardRadius,
          border: Border.all(color: AppColors.border),
          boxShadow: AppShadows.card,
        ),
        child: Material(
          color: Colors.transparent,
          child: InkWell(
            onTap: onTap,
            borderRadius: AppRadius.cardRadius,
            child: Padding(padding: padding, child: child),
          ),
        ),
      );
}
