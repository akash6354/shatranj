import 'package:flutter/material.dart';

import '../theme/app_theme.dart';

class AppAvatar extends StatelessWidget {
  const AppAvatar({
    super.key,
    this.imageUrl,
    this.initials,
    this.size = 44,
    this.backgroundColor,
    this.borderColor,
    this.showOnline = false,
  });

  final String? imageUrl;
  final String? initials;
  final double size;
  final Color? backgroundColor;
  final Color? borderColor;
  final bool showOnline;

  @override
  Widget build(BuildContext context) {
    final avatar = CircleAvatar(
      radius: size / 2,
      backgroundColor: backgroundColor ?? AppColors.surfaceElevated,
      backgroundImage: imageUrl == null ? null : NetworkImage(imageUrl!),
      child: imageUrl == null
          ? Text(
              initials ?? '?',
              style: AppTextStyles.bodyStrong.copyWith(
                color: AppColors.gold,
                fontSize: size * .34,
              ),
            )
          : null,
    );
    return Stack(
      clipBehavior: Clip.none,
      children: [
        Container(
          padding: const EdgeInsets.all(2),
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            border: Border.all(color: borderColor ?? AppColors.borderStrong),
          ),
          child: avatar,
        ),
        if (showOnline)
          Positioned(
            right: -1,
            bottom: 1,
            child: Container(
              width: size * .25,
              height: size * .25,
              decoration: BoxDecoration(
                color: AppColors.success,
                shape: BoxShape.circle,
                border: Border.all(color: AppColors.background, width: 2),
              ),
            ),
          ),
      ],
    );
  }
}
