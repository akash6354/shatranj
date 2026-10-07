import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/app_widgets.dart';
import '../domain/home_models.dart';
import 'home_controller.dart';
import '../../../core/widgets/chess/chess_board.dart';

class HomeScreen extends ConsumerWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final snapshot = ref.watch(homeSnapshotProvider);
    return snapshot.when(
      loading: () => const AppLoading(message: 'Loading your board'),
      error: (error, stackTrace) => AppError(
        title: 'Could not load your home',
        message: error.toString(),
        onRetry: () => ref.invalidate(homeSnapshotProvider),
      ),
      data: (home) => HomeContent(home: home),
    );
  }
}

class HomeContent extends StatelessWidget {
  const HomeContent({required this.home, super.key});

  final HomeSnapshot home;

  @override
  Widget build(BuildContext context) => LayoutBuilder(
        builder: (context, constraints) {
          final wide = constraints.maxWidth >= 760;
          return SingleChildScrollView(
            padding: const EdgeInsets.fromLTRB(
              AppSpacing.page,
              AppSpacing.md,
              AppSpacing.page,
              AppSpacing.xl,
            ),
            child: Center(
              child: ConstrainedBox(
                constraints: const BoxConstraints(maxWidth: 980),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    _HomeHeader(home: home),
                    const SizedBox(height: AppSpacing.lg),
                    _HeroCard(home: home),
                    const SizedBox(height: AppSpacing.lg),
                    _QuickActions(actions: home.quickActions),
                    const SizedBox(height: AppSpacing.lg),
                    _StatsCard(home: home),
                    const SizedBox(height: AppSpacing.lg),
                    if (wide)
                      Row(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Expanded(child: _DailyPuzzleCard(puzzle: home.dailyPuzzle)),
                          const SizedBox(width: AppSpacing.md),
                          Expanded(child: _LearningCard(courses: home.courses)),
                        ],
                      )
                    else ...[
                      _DailyPuzzleCard(puzzle: home.dailyPuzzle),
                      const SizedBox(height: AppSpacing.md),
                      _LearningCard(courses: home.courses),
                    ],
                    const SizedBox(height: AppSpacing.lg),
                    _TournamentCard(),
                  ],
                ),
              ),
            ),
          );
        },
      );
}

class _HomeHeader extends StatelessWidget {
  const _HomeHeader({required this.home});

  final HomeSnapshot home;

  @override
  Widget build(BuildContext context) => AppTopBar(
        title: 'Good morning, ${home.displayName}',
        subtitle: '@${home.username}',
        leading: const AppAvatar(initials: 'A', size: 48, showOnline: true),
        actions: [
          AppIconButton(
            icon: Icons.notifications_none_rounded,
            tooltip: 'Notifications',
            onPressed: () {},
          ),
        ],
      );
}

class _HeroCard extends StatelessWidget {
  const _HeroCard({required this.home});

  final HomeSnapshot home;

  @override
  Widget build(BuildContext context) => AppGradientCard(
        child: Row(
          children: [
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  AppPill(
                    label: '${home.streak} day streak',
                    icon: Icons.local_fire_department_rounded,
                    color: AppColors.goldSoft,
                    backgroundColor: AppColors.gold.withValues(alpha: .12),
                  ),
                  const SizedBox(height: AppSpacing.md),
                  const Text('Keep your\nmomentum going', style: AppTextStyles.headline),
                  const SizedBox(height: AppSpacing.sm),
                  Text(
                    'You are ${home.rating} rated. One game today keeps your streak alive.',
                    style: AppTextStyles.body,
                  ),
                  const SizedBox(height: AppSpacing.md),
                  const AppButton(
                    label: 'Play a game',
                    icon: Icons.play_arrow_rounded,
                  ),
                ],
              ),
            ),
            const SizedBox(width: AppSpacing.sm),
            const Icon(Icons.auto_awesome_rounded, color: AppColors.gold, size: 72),
          ],
        ),
      );
}

class _QuickActions extends StatelessWidget {
  const _QuickActions({required this.actions});

  final List<HomeAction> actions;

  @override
  Widget build(BuildContext context) => LayoutBuilder(
        builder: (context, constraints) {
          final itemWidth = (constraints.maxWidth - AppSpacing.md * 3) / 4;
          return Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const AppSectionHeader(title: 'Explore Shatranj'),
              const SizedBox(height: AppSpacing.sm),
              Wrap(
                spacing: AppSpacing.md,
                runSpacing: AppSpacing.md,
                children: actions
                    .map(
                      (action) => SizedBox(
                        width: itemWidth.clamp(130, 220),
                        child: _ActionCard(action: action),
                      ),
                    )
                    .toList(),
              ),
            ],
          );
        },
      );
}

class _ActionCard extends StatelessWidget {
  const _ActionCard({required this.action});

  final HomeAction action;

  @override
  Widget build(BuildContext context) => AppCard(
        padding: const EdgeInsets.all(AppSpacing.md),
        onTap: () {},
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Container(
              width: 40,
              height: 40,
              decoration: BoxDecoration(
                color: action.color.withValues(alpha: .14),
                borderRadius: AppRadius.small,
              ),
              child: Icon(action.icon, color: action.color),
            ),
            const SizedBox(height: AppSpacing.sm),
            Text(action.title, style: AppTextStyles.titleSmall),
            const SizedBox(height: AppSpacing.xxs),
            Text(action.subtitle, style: AppTextStyles.caption),
          ],
        ),
      );
}

class _StatsCard extends StatelessWidget {
  const _StatsCard({required this.home});

  final HomeSnapshot home;

  @override
  Widget build(BuildContext context) => AppCard(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const AppSectionHeader(title: 'Your progress', subtitle: 'Small steps become strong habits'),
            const SizedBox(height: AppSpacing.md),
            Row(
              children: [
                _Stat(label: 'Rating', value: '${home.rating}', icon: Icons.trending_up_rounded),
                _Stat(label: 'Games', value: '${home.gamesPlayed}', icon: Icons.sports_esports_rounded),
                _Stat(label: 'Puzzles', value: '${home.puzzlesSolved}', icon: Icons.extension_rounded),
                _Stat(label: 'Hours', value: '${home.learningHours}', icon: Icons.schedule_rounded),
              ],
            ),
          ],
        ),
      );
}

class _Stat extends StatelessWidget {
  const _Stat({required this.label, required this.value, required this.icon});

  final String label;
  final String value;
  final IconData icon;

  @override
  Widget build(BuildContext context) => Expanded(
        child: Column(
          children: [
            Icon(icon, color: AppColors.gold, size: 20),
            const SizedBox(height: AppSpacing.xs),
            Text(value, style: AppTextStyles.numeric),
            Text(label, style: AppTextStyles.caption),
          ],
        ),
      );
}

class _DailyPuzzleCard extends StatelessWidget {
  const _DailyPuzzleCard({required this.puzzle});

  final DailyPuzzle puzzle;

  @override
  Widget build(BuildContext context) => AppCard(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const AppSectionHeader(title: 'Daily puzzle', actionLabel: 'View all'),
            const SizedBox(height: AppSpacing.sm),
            Row(
              children: [
                SizedBox(width: 142, child: ChessBoard(fen: puzzle.fen)),
                const SizedBox(width: AppSpacing.md),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(puzzle.title, style: AppTextStyles.titleSmall),
                      const SizedBox(height: AppSpacing.xs),
                      Text(
                        '${puzzle.rating} rating • ${puzzle.difficulty}',
                        style: AppTextStyles.caption,
                      ),
                      const SizedBox(height: AppSpacing.md),
                      const AppButton(
                        label: 'Solve puzzle',
                        variant: AppButtonVariant.secondary,
                        expand: true,
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ],
        ),
      );
}

class _LearningCard extends StatelessWidget {
  const _LearningCard({required this.courses});

  final List<CourseProgress> courses;

  @override
  Widget build(BuildContext context) => AppCard(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const AppSectionHeader(title: 'Continue learning', actionLabel: 'See all'),
            const SizedBox(height: AppSpacing.sm),
            ...courses.map((course) => _CourseRow(course: course)),
          ],
        ),
      );
}

class _CourseRow extends StatelessWidget {
  const _CourseRow({required this.course});

  final CourseProgress course;

  @override
  Widget build(BuildContext context) => Padding(
        padding: const EdgeInsets.symmetric(vertical: AppSpacing.xs),
        child: Row(
          children: [
            Container(
              width: 40,
              height: 40,
              decoration: BoxDecoration(
                color: AppColors.gold.withValues(alpha: .14),
                borderRadius: AppRadius.small,
              ),
              child: Icon(course.icon, color: AppColors.gold),
            ),
            const SizedBox(width: AppSpacing.sm),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(course.title, style: AppTextStyles.bodyStrong),
                  Text(course.category, style: AppTextStyles.caption),
                  const SizedBox(height: AppSpacing.xs),
                  LinearProgressIndicator(
                    value: course.progress / 100,
                    minHeight: 5,
                    borderRadius: AppRadius.pillRadius,
                    backgroundColor: AppColors.border,
                    color: AppColors.mint,
                  ),
                ],
              ),
            ),
            const SizedBox(width: AppSpacing.sm),
            Text('${course.progress}%', style: AppTextStyles.label.copyWith(color: AppColors.mint)),
          ],
        ),
      );
}

class _TournamentCard extends StatelessWidget {
  @override
  Widget build(BuildContext context) => AppGradientCard(
        padding: const EdgeInsets.all(AppSpacing.md),
        child: Row(
          children: [
            const Icon(Icons.emoji_events_rounded, color: AppColors.gold, size: 34),
            const SizedBox(width: AppSpacing.md),
            const Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text('Weekend Blitz Arena', style: AppTextStyles.titleSmall),
                  SizedBox(height: AppSpacing.xxs),
                  Text('Join the next live tournament and test your progress.', style: AppTextStyles.caption),
                ],
              ),
            ),
            const AppButton(label: 'Explore', variant: AppButtonVariant.outline),
          ],
        ),
      );
}
