import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/app_widgets.dart';
import '../../../core/widgets/chess/chess_widgets.dart';
import '../domain/puzzle_models.dart';
import 'puzzle_controller.dart';

class PuzzleHomeScreen extends ConsumerWidget {
  const PuzzleHomeScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final puzzle = ref.watch(dailyPuzzleProvider);
    final stats = ref.watch(puzzleStatsProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Puzzles')),
      body: SafeArea(
        child: puzzle.when(
          loading: () => const AppLoading(message: 'Loading puzzles'),
          error: (error, _) => AppError(title: 'Puzzles unavailable', message: error.toString()),
          data: (daily) => stats.when(
            loading: () => const AppLoading(),
            error: (error, _) => AppError(title: 'Stats unavailable', message: error.toString()),
            data: (summary) => _PuzzleHomeContent(puzzle: daily, stats: summary),
          ),
        ),
      ),
    );
  }
}

class _PuzzleHomeContent extends StatelessWidget {
  const _PuzzleHomeContent({required this.puzzle, required this.stats});

  final Puzzle puzzle;
  final PuzzleStats stats;

  @override
  Widget build(BuildContext context) => SingleChildScrollView(
        padding: const EdgeInsets.fromLTRB(AppSpacing.page, AppSpacing.sm, AppSpacing.page, AppSpacing.xl),
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 900),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                AppGradientCard(
                  child: Row(
                    children: [
                      const Icon(Icons.extension_rounded, color: AppColors.gold, size: 52),
                      const SizedBox(width: AppSpacing.md),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            const Text('Sharpen your tactics', style: AppTextStyles.headline),
                            const SizedBox(height: AppSpacing.xs),
                            Text('${stats.rating} puzzle rating • ${stats.streak} day streak', style: AppTextStyles.body),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: AppSpacing.lg),
                const AppSectionHeader(title: 'Daily puzzle', subtitle: 'A new challenge every day'),
                const SizedBox(height: AppSpacing.sm),
                AppCard(
                  onTap: () => context.push('/puzzles/${puzzle.id}/play'),
                  child: Row(
                    children: [
                      SizedBox(width: 150, child: ChessBoardPreview(fen: puzzle.fen)),
                      const SizedBox(width: AppSpacing.md),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(puzzle.difficulty, style: AppPillText.style),
                            const SizedBox(height: AppSpacing.xs),
                            Text(puzzle.themes.join(' • '), style: AppTextStyles.caption),
                            const SizedBox(height: AppSpacing.sm),
                            const Text('Find the best move', style: AppTextStyles.titleSmall),
                            const SizedBox(height: AppSpacing.md),
                            const AppButton(label: 'Solve puzzle', expand: true),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: AppSpacing.lg),
                const AppSectionHeader(title: 'Puzzle themes'),
                const SizedBox(height: AppSpacing.sm),
                const Wrap(
                  spacing: AppSpacing.sm,
                  runSpacing: AppSpacing.sm,
                  children: [
                    AppPill(label: 'Tactics', icon: Icons.flash_on_rounded),
                    AppPill(label: 'Opening', icon: Icons.menu_book_rounded),
                    AppPill(label: 'Endgame', icon: Icons.flag_rounded),
                    AppPill(label: 'Checkmate', icon: Icons.gps_fixed_rounded),
                  ],
                ),
                const SizedBox(height: AppSpacing.lg),
                AppCard(
                  child: Row(
                    children: [
                      _Stat(value: '${stats.solved}', label: 'Solved'),
                      _Stat(value: '${stats.attempted}', label: 'Attempted'),
                      _Stat(value: '${stats.streak}', label: 'Streak'),
                    ],
                  ),
                ),
                const SizedBox(height: AppSpacing.lg),
                const AppSectionHeader(title: 'Recent puzzles', actionLabel: 'See all'),
                const SizedBox(height: AppSpacing.sm),
                AppCard(
                  onTap: () => context.push('/puzzles/${puzzle.id}/play'),
                  child: Row(
                    children: [
                      const Icon(Icons.history_rounded, color: AppColors.mint),
                      const SizedBox(width: AppSpacing.sm),
                      const Expanded(child: Text('Daily puzzle • Medium', style: AppTextStyles.bodyStrong)),
                      Text('+18', style: AppTextStyles.bodyStrong.copyWith(color: AppColors.success)),
                      const Icon(Icons.chevron_right_rounded, color: AppColors.muted),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
      );
}

class ChessBoardPreview extends StatelessWidget {
  const ChessBoardPreview({required this.fen, super.key});
  final String fen;

  @override
  Widget build(BuildContext context) => ChessBoard(fen: fen, showCoordinates: false);
}

class _Stat extends StatelessWidget {
  const _Stat({required this.value, required this.label});
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

abstract final class AppPillText {
  static const style = TextStyle(
    color: AppColors.gold,
    fontSize: 12,
    fontWeight: FontWeight.w700,
  );
}
