import 'package:flutter/material.dart';

import 'app_colors.dart';

abstract final class AppTextStyles {
  static const display = TextStyle(
    fontSize: 32,
    height: 1.12,
    fontWeight: FontWeight.w800,
    letterSpacing: -0.6,
    color: AppColors.text,
  );
  static const headline = TextStyle(
    fontSize: 24,
    height: 1.18,
    fontWeight: FontWeight.w800,
    letterSpacing: -0.3,
    color: AppColors.text,
  );
  static const title = TextStyle(
    fontSize: 18,
    height: 1.25,
    fontWeight: FontWeight.w800,
    color: AppColors.text,
  );
  static const titleSmall = TextStyle(
    fontSize: 15,
    height: 1.25,
    fontWeight: FontWeight.w700,
    color: AppColors.text,
  );
  static const body = TextStyle(
    fontSize: 14,
    height: 1.45,
    fontWeight: FontWeight.w400,
    color: AppColors.textSecondary,
  );
  static const bodyStrong = TextStyle(
    fontSize: 14,
    height: 1.4,
    fontWeight: FontWeight.w600,
    color: AppColors.text,
  );
  static const caption = TextStyle(
    fontSize: 12,
    height: 1.3,
    fontWeight: FontWeight.w400,
    color: AppColors.muted,
  );
  static const label = TextStyle(
    fontSize: 11,
    height: 1.2,
    fontWeight: FontWeight.w700,
    letterSpacing: 0.25,
    color: AppColors.muted,
  );
  static const button = TextStyle(
    fontSize: 14,
    height: 1.2,
    fontWeight: FontWeight.w700,
    color: AppColors.background,
  );
  static const numeric = TextStyle(
    fontSize: 28,
    height: 1.1,
    fontWeight: FontWeight.w800,
    color: AppColors.text,
  );
}
