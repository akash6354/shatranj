import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/routing/app_router.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/app_widgets.dart';
import '../domain/game_history_models.dart';
import 'game_history_controller.dart';

class GameHistoryDetailScreen extends ConsumerWidget {
  const GameHistoryDetailScreen({required this.gameId, super.key});

  final String gameId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final game = ref.watch(gameHistoryDetailProvider(gameId));
    return Scaffold(
      appBar: AppBar(title: const Text('Game details')),
      body: SafeArea(
        child: game.when(
          loading: () => const AppLoading(),
          error: (error, _) => AppError(title: 'Game unavailable', message: error.toString()),
          data: (data) => _DetailContent(game: data),
        ),
      ),
    );
  }
}

class _DetailContent extends StatelessWidget {
  const _DetailContent({required this.game});

  final GameHistoryItem game;

  @override
  Widget build(BuildContext context) => SingleChildScrollView(
        padding: const EdgeInsets.all(AppSpacing.page),
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 620),
            child: Column(
              children: [
                AppGradientCard(
                  child: Column(
                    children: [
                      Icon(_resultIcon(game.result), color: _resultColor(game.result), size: 56),
                      const SizedBox(height: AppSpacing.sm),
                      Text(game.resultLabel, style: AppTextStyles.headline, textAlign: TextAlign.center),
                      const SizedBox(height: AppSpacing.xs),
                      Text(game.endReason, style: AppTextStyles.body),
                      const SizedBox(height: AppSpacing.md),
                      AppPill(label: _ratingLabel(game.ratingChange), color: _ratingColor(game.ratingChange), icon: Icons.trending_up_rounded),
                    ],
                  ),
                ),
                const SizedBox(height: AppSpacing.md),
                AppCard(
                  child: Column(
                    children: [
                      _DetailRow(label: 'Opponent', value: '${game.opponent} (${game.opponentRating})'),
                      _DetailRow(label: 'Color', value: game.color.name),
                      _DetailRow(label: 'Time control', value: game.timeControl),
                      _DetailRow(label: 'Moves', value: '${game.moveCount}'),
                      _DetailRow(label: 'Duration', value: _durationLabel(game.duration)),
                      _DetailRow(label: 'Accuracy', value: game.accuracy == null ? 'Unavailable' : '${game.accuracy!.toStringAsFixed(1)}%'),
                      _DetailRow(label: 'Played', value: _dateLabel(game.playedAt)),
                    ],
                  ),
                ),
                const SizedBox(height: AppSpacing.md),
                Row(
                  children: [
                    Expanded(child: AppButton(label: 'Game review', icon: Icons.analytics_outlined, onPressed: () => context.push(AppRoutes.gameReview(game.id)))),
                    const SizedBox(width: AppSpacing.sm),
                    Expanded(child: AppButton(label: 'Share result', icon: Icons.share_outlined, variant: AppButtonVariant.outline, onPressed: () => ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Result ready to share'))))),
                  ],
                ),
                const SizedBox(height: AppSpacing.sm),
                AppButton(label: 'Return to home', variant: AppButtonVariant.ghost, onPressed: () => context.go(AppRoutes.home)),
              ],
            ),
          ),
        ),
      );
}

class _DetailRow extends StatelessWidget {
  const _DetailRow({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) => Padding(
        padding: const EdgeInsets.symmetric(vertical: AppSpacing.xs),
        child: Row(
          children: [
            Expanded(child: Text(label, style: AppTextStyles.caption)),
            Text(value, style: AppTextStyles.bodyStrong),
          ],
        ),
      );
}

IconData _resultIcon(HistoryResult result) => switch (result) {
      HistoryResult.win => Icons.emoji_events_rounded,
      HistoryResult.loss => Icons.close_rounded,
      HistoryResult.draw => Icons.handshake_rounded,
      HistoryResult.abandoned => Icons.flag_outlined,
    };

Color _resultColor(HistoryResult result) => switch (result) {
      HistoryResult.win => AppColors.success,
      HistoryResult.loss => AppColors.danger,
      HistoryResult.draw => AppColors.warning,
      HistoryResult.abandoned => AppColors.muted,
    };

Color _ratingColor(int value) => value > 0
    ? AppColors.success
    : value < 0
        ? AppColors.danger
        : AppColors.muted;

String _ratingLabel(int value) => value > 0 ? '+$value' : '$value';

String _durationLabel(Duration duration) =>
    '${duration.inMinutes}:${duration.inSeconds.remainder(60).toString().padLeft(2, '0')}';

String _dateLabel(DateTime date) =>
    '${date.day}/${date.month}/${date.year} • ${date.hour.toString().padLeft(2, '0')}:${date.minute.toString().padLeft(2, '0')}';
