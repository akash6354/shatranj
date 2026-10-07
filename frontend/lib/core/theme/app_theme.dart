import 'package:flutter/material.dart';

import 'app_colors.dart';
import 'app_radius.dart';
import 'app_spacing.dart';
import 'app_text_styles.dart';

export 'app_colors.dart';
export 'app_radius.dart';
export 'app_shadows.dart';
export 'app_spacing.dart';
export 'app_text_styles.dart';

abstract final class AppTheme {
  static ThemeData get dark {
    final scheme = ColorScheme.fromSeed(
      seedColor: AppColors.gold,
      brightness: Brightness.dark,
    ).copyWith(
      primary: AppColors.gold,
      onPrimary: AppColors.background,
      secondary: AppColors.mint,
      onSecondary: AppColors.background,
      surface: AppColors.surface,
      onSurface: AppColors.text,
      error: AppColors.danger,
    );

    return ThemeData(
      brightness: Brightness.dark,
      colorScheme: scheme,
      scaffoldBackgroundColor: AppColors.background,
      canvasColor: AppColors.background,
      fontFamily: 'Roboto',
      useMaterial3: true,
      visualDensity: VisualDensity.standard,
      textTheme: const TextTheme(
        displayLarge: AppTextStyles.display,
        headlineMedium: AppTextStyles.headline,
        titleLarge: AppTextStyles.title,
        titleMedium: AppTextStyles.titleSmall,
        bodyLarge: AppTextStyles.body,
        bodyMedium: AppTextStyles.body,
        labelLarge: AppTextStyles.button,
        labelMedium: AppTextStyles.label,
      ),
      cardTheme: CardThemeData(
        color: AppColors.surface,
        elevation: 0,
        margin: EdgeInsets.zero,
        shape: RoundedRectangleBorder(
          borderRadius: AppRadius.cardRadius,
          side: const BorderSide(color: AppColors.border),
        ),
      ),
      dividerTheme: const DividerThemeData(
        color: AppColors.border,
        thickness: 1,
        space: AppSpacing.md,
      ),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: AppColors.surfaceMuted,
        hintStyle: AppTextStyles.body.copyWith(color: AppColors.muted),
        labelStyle: AppTextStyles.body,
        contentPadding: const EdgeInsets.symmetric(
          horizontal: AppSpacing.md,
          vertical: AppSpacing.sm + 2,
        ),
        border: OutlineInputBorder(
          borderRadius: AppRadius.buttonRadius,
          borderSide: const BorderSide(color: AppColors.border),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: AppRadius.buttonRadius,
          borderSide: const BorderSide(color: AppColors.border),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: AppRadius.buttonRadius,
          borderSide: const BorderSide(color: AppColors.gold, width: 1.2),
        ),
        errorBorder: OutlineInputBorder(
          borderRadius: AppRadius.buttonRadius,
          borderSide: const BorderSide(color: AppColors.danger),
        ),
      ),
      navigationBarTheme: NavigationBarThemeData(
        backgroundColor: AppColors.background,
        elevation: 0,
        indicatorColor: AppColors.gold.withValues(alpha: .18),
        labelTextStyle: WidgetStatePropertyAll(AppTextStyles.label),
        iconTheme: WidgetStateProperty.resolveWith(
          (states) => IconThemeData(
            color: states.contains(WidgetState.selected)
                ? AppColors.gold
                : AppColors.muted,
          ),
        ),
      ),
      progressIndicatorTheme: const ProgressIndicatorThemeData(
        color: AppColors.gold,
        linearTrackColor: AppColors.border,
      ),
      snackBarTheme: SnackBarThemeData(
        backgroundColor: AppColors.surfaceElevated,
        contentTextStyle: AppTextStyles.bodyStrong,
        shape: RoundedRectangleBorder(borderRadius: AppRadius.buttonRadius),
        behavior: SnackBarBehavior.floating,
      ),
      chipTheme: ChipThemeData(
        backgroundColor: AppColors.surfaceMuted,
        selectedColor: AppColors.gold,
        disabledColor: AppColors.surfaceMuted,
        side: const BorderSide(color: AppColors.border),
        shape: RoundedRectangleBorder(borderRadius: AppRadius.pillRadius),
        labelStyle: AppTextStyles.caption,
        secondaryLabelStyle: AppTextStyles.caption.copyWith(
          color: AppColors.background,
          fontWeight: FontWeight.w700,
        ),
        padding: const EdgeInsets.symmetric(horizontal: AppSpacing.sm),
      ),
    );
  }
}
