import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/routing/app_router.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/app_widgets.dart';
import '../domain/game_history_models.dart';
import 'game_history_controller.dart';

class GameResultScreen extends ConsumerWidget {
  const GameResultScreen({required this.gameId, super.key});

  final String gameId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final game = ref.watch(gameHistoryDetailProvider(gameId));
    return Scaffold(
      body: SafeArea(
        child: game.when(
          loading: () => const AppLoading(message: 'Preparing result'),
          error: (error, _) => AppError(title: 'Result unavailable', message: error.toString()),
          data: (data) => _ResultContent(game: data),
        ),
      ),
    );
  }
}

class _ResultContent extends StatelessWidget {
  const _ResultContent({required this.game});

  final GameHistoryItem game;

  @override
  Widget build(BuildContext context) => SingleChildScrollView(
        padding: const EdgeInsets.all(AppSpacing.page),
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 560),
            child: Column(
              children: [
                const SizedBox(height: AppSpacing.xl),
                Icon(_resultIcon(game.result), color: _resultColor(game.result), size: 72),
                const SizedBox(height: AppSpacing.md),
                Text(game.resultLabel, style: AppTextStyles.headline, textAlign: TextAlign.center),
                const SizedBox(height: AppSpacing.xs),
                Text(game.endReason, style: AppTextStyles.body),
                const SizedBox(height: AppSpacing.xl),
                AppCard(
                  child: Column(
                    children: [
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceAround,
                        children: [
                          _Metric(value: _ratingLabel(game.ratingChange), label: 'Rating change', color: _ratingColor(game.ratingChange)),
                          _Metric(value: game.accuracy == null ? '—' : '${game.accuracy!.toStringAsFixed(1)}%', label: 'Accuracy'),
                          _Metric(value: '${game.moveCount}', label: 'Moves'),
                          _Metric(value: _durationLabel(game.duration), label: 'Duration'),
                        ],
                      ),
                      const SizedBox(height: AppSpacing.lg),
                      AppPill(label: '${game.color.name} • ${game.timeControl}', icon: Icons.timer_outlined, color: AppColors.goldSoft),
                    ],
                  ),
                ),
                const SizedBox(height: AppSpacing.lg),
                AppButton(label: 'Rematch', icon: Icons.refresh_rounded, expand: true, onPressed: () => context.go('${AppRoutes.game}/${game.id}')),
                const SizedBox(height: AppSpacing.sm),
                AppButton(label: 'Game review', icon: Icons.analytics_outlined, variant: AppButtonVariant.secondary, expand: true, onPressed: () => context.push(AppRoutes.gameReview(game.id))),
                const SizedBox(height: AppSpacing.sm),
                AppButton(label: 'Share result', icon: Icons.share_outlined, variant: AppButtonVariant.outline, expand: true, onPressed: () => ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Result ready to share')))),
                const SizedBox(height: AppSpacing.md),
                TextButton.icon(onPressed: () => context.go(AppRoutes.home), icon: const Icon(Icons.home_outlined), label: const Text('Return to home')),
              ],
            ),
          ),
        ),
      );
}

class _Metric extends StatelessWidget {
  const _Metric({required this.value, required this.label, this.color});

  final String value;
  final String label;
  final Color? color;

  @override
  Widget build(BuildContext context) => Column(
        children: [
          Text(value, style: AppTextStyles.title.copyWith(color: color)),
          const SizedBox(height: AppSpacing.xxs),
          Text(label, style: AppTextStyles.caption),
        ],
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

Color _ratingColor(int value) => value > 0 ? AppColors.success : value < 0 ? AppColors.danger : AppColors.muted;
String _ratingLabel(int value) => value > 0 ? '+$value' : '$value';
String _durationLabel(Duration duration) => '${duration.inMinutes}:${duration.inSeconds.remainder(60).toString().padLeft(2, '0')}';
