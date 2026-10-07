import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/app_widgets.dart';

class PlayHomeScreen extends StatelessWidget {
  const PlayHomeScreen({super.key});

  @override
  Widget build(BuildContext context) => Scaffold(
        appBar: AppBar(title: const Text('Play')),
        body: SafeArea(
          child: SingleChildScrollView(
            padding: const EdgeInsets.fromLTRB(AppSpacing.page, AppSpacing.sm, AppSpacing.page, AppSpacing.xl),
            child: Center(
              child: ConstrainedBox(
                constraints: const BoxConstraints(maxWidth: 760),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text('Choose your game', style: AppTextStyles.headline),
                    const SizedBox(height: AppSpacing.xs),
                    const Text('Play a quick match, challenge a friend, or practice against the computer.', style: AppTextStyles.body),
                    const SizedBox(height: AppSpacing.lg),
                    _PlayOption(
                      icon: Icons.flash_on_rounded,
                      title: 'Quick Match',
                      description: 'Find an opponent at your level',
                      onTap: () => context.push('/play/matchmaking'),
                    ),
                    _PlayOption(
                      icon: Icons.people_alt_outlined,
                      title: 'Play Friend',
                      description: 'Create or join a private room',
                      onTap: () => context.push('/play/friend'),
                    ),
                    _PlayOption(
                      icon: Icons.smart_toy_outlined,
                      title: 'Play Computer',
                      description: 'Practice with adjustable difficulty',
                      onTap: () => context.push('/play/computer'),
                    ),
                    const SizedBox(height: AppSpacing.lg),
                    const AppSectionHeader(title: 'Popular time controls'),
                    const SizedBox(height: AppSpacing.sm),
                    const Wrap(
                      spacing: AppSpacing.sm,
                      runSpacing: AppSpacing.sm,
                      children: [
                        AppPill(label: '1+0 Bullet', icon: Icons.bolt_rounded),
                        AppPill(label: '3+2 Blitz', icon: Icons.timer_outlined),
                        AppPill(label: '5+0 Blitz', icon: Icons.timer_outlined),
                        AppPill(label: '10+5 Rapid', icon: Icons.hourglass_bottom_rounded),
                      ],
                    ),
                  ],
                ),
              ),
            ),
          ),
        ),
      );
}

class _PlayOption extends StatelessWidget {
  const _PlayOption({
    required this.icon,
    required this.title,
    required this.description,
    required this.onTap,
  });

  final IconData icon;
  final String title;
  final String description;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) => AppCard(
        margin: const EdgeInsets.only(bottom: AppSpacing.sm),
        onTap: onTap,
        child: Row(
          children: [
            Icon(icon, color: AppColors.gold, size: 32),
            const SizedBox(width: AppSpacing.md),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(title, style: AppTextStyles.titleSmall),
                  const SizedBox(height: AppSpacing.xs),
                  Text(description, style: AppTextStyles.caption),
                ],
              ),
            ),
            const Icon(Icons.chevron_right_rounded, color: AppColors.muted),
          ],
        ),
      );
}
