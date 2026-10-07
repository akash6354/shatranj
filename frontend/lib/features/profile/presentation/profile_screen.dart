import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/app_widgets.dart';
import '../domain/profile_models.dart';
import 'profile_controller.dart';

class ProfileScreen extends ConsumerWidget {
  const ProfileScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final profile = ref.watch(profileProvider);
    return Scaffold(
      appBar: AppBar(
        title: const Text('Profile'),
        actions: [
          AppIconButton(
            icon: Icons.settings_outlined,
            tooltip: 'Settings',
            onPressed: () => context.push('/profile/settings'),
          ),
        ],
      ),
      body: SafeArea(
        child: profile.when(
          loading: () => const AppLoading(message: 'Loading profile'),
          error: (error, _) => AppError(
            message: error.toString(),
            onRetry: () => ref.invalidate(profileProvider),
          ),
          data: (item) => _ProfileContent(profile: item),
        ),
      ),
    );
  }
}

class _ProfileContent extends StatelessWidget {
  const _ProfileContent({required this.profile});
  final Profile profile;

  @override
  Widget build(BuildContext context) => SingleChildScrollView(
        padding: const EdgeInsets.fromLTRB(
          AppSpacing.page,
          AppSpacing.sm,
          AppSpacing.page,
          AppSpacing.xl,
        ),
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 820),
            child: Column(
              children: [
                AppGradientCard(
                  child: Column(
                    children: [
                      const CircleAvatar(
                        radius: 42,
                        child: Icon(Icons.person_rounded, size: 44),
                      ),
                      const SizedBox(height: AppSpacing.md),
                      Text(profile.displayName, style: AppTextStyles.headline),
                      Text('@${profile.username}', style: AppTextStyles.caption),
                      const SizedBox(height: AppSpacing.md),
                      AppPill(
                        label: '${profile.rating} rating',
                        icon: Icons.star_rounded,
                        color: AppColors.gold,
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: AppSpacing.lg),
                _Ratings(profile: profile),
                const SizedBox(height: AppSpacing.lg),
                AppCard(
                  child: Row(
                    children: [
                      _Metric(value: '${profile.games}', label: 'Games'),
                      _Metric(value: '${profile.winRate.toStringAsFixed(0)}%', label: 'Win rate'),
                      _Metric(value: '${profile.streak}', label: 'Streak'),
                      _Metric(value: '${profile.friends}', label: 'Friends'),
                    ],
                  ),
                ),
                const SizedBox(height: AppSpacing.lg),
                const AppSectionHeader(title: 'Achievements'),
                const SizedBox(height: AppSpacing.sm),
                Wrap(
                  spacing: AppSpacing.sm,
                  runSpacing: AppSpacing.sm,
                  children: profile.achievements
                      .map((item) => AppPill(
                            label: item,
                            icon: Icons.emoji_events_outlined,
                            color: AppColors.gold,
                          ))
                      .toList(),
                ),
                const SizedBox(height: AppSpacing.lg),
                AppCard(
                  onTap: () => context.push('/games/history'),
                  child: const Row(
                    children: [
                      Icon(Icons.history_rounded, color: AppColors.mint),
                      SizedBox(width: AppSpacing.sm),
                      Expanded(child: Text('Game history', style: AppTextStyles.bodyStrong)),
                      Icon(Icons.chevron_right_rounded, color: AppColors.muted),
                    ],
                  ),
                ),
                const SizedBox(height: AppSpacing.sm),
                AppCard(
                  onTap: () => context.push('/community/friends'),
                  child: const Row(
                    children: [
                      Icon(Icons.people_alt_outlined, color: AppColors.mint),
                      SizedBox(width: AppSpacing.sm),
                      Expanded(child: Text('Friends', style: AppTextStyles.bodyStrong)),
                      Icon(Icons.chevron_right_rounded, color: AppColors.muted),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
      );
}

class _Ratings extends StatelessWidget {
  const _Ratings({required this.profile});
  final Profile profile;

  @override
  Widget build(BuildContext context) => Row(
        children: [
          Expanded(child: _Rating(value: profile.blitzRating, label: 'Blitz')),
          Expanded(child: _Rating(value: profile.rapidRating, label: 'Rapid')),
          Expanded(child: _Rating(value: profile.puzzleRating, label: 'Puzzles')),
        ],
      );
}

class _Rating extends StatelessWidget {
  const _Rating({required this.value, required this.label});
  final int value;
  final String label;

  @override
  Widget build(BuildContext context) => AppCard(
        margin: const EdgeInsets.symmetric(horizontal: AppSpacing.xxs),
        child: Column(
          children: [
            Text('$value', style: AppTextStyles.numeric),
            Text(label, style: AppTextStyles.caption),
          ],
        ),
      );
}

class _Metric extends StatelessWidget {
  const _Metric({required this.value, required this.label});
  final String value;
  final String label;

  @override
  Widget build(BuildContext context) => Expanded(
        child: Column(
          children: [
            Text(value, style: AppTextStyles.numeric),
            Text(label, style: AppTextStyles.caption),
          ],
        ),
      );
}
