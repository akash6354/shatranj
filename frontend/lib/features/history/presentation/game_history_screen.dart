import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/routing/app_router.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/app_widgets.dart';
import '../domain/game_history_models.dart';
import 'game_history_controller.dart';

class GameHistoryScreen extends ConsumerWidget {
  const GameHistoryScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final history = ref.watch(gameHistoryProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Game History')),
      body: SafeArea(
        child: history.when(
          loading: () => const AppLoading(message: 'Loading games'),
          error: (error, _) => AppError(
            title: 'Could not load game history',
            message: error.toString(),
            onRetry: () => ref.invalidate(gameHistoryProvider),
          ),
          data: (state) => _HistoryList(games: state.games),
        ),
      ),
    );
  }
}

class _HistoryList extends StatelessWidget {
  const _HistoryList({required this.games});

  final List<GameHistoryItem> games;

  @override
  Widget build(BuildContext context) => ListView.separated(
        padding: const EdgeInsets.fromLTRB(AppSpacing.page, AppSpacing.sm, AppSpacing.page, AppSpacing.xl),
        itemCount: games.length,
        separatorBuilder: (_, _) => const SizedBox(height: AppSpacing.sm),
        itemBuilder: (context, index) => _HistoryTile(game: games[index]),
      );
}

class _HistoryTile extends StatelessWidget {
  const _HistoryTile({required this.game});

  final GameHistoryItem game;

  @override
  Widget build(BuildContext context) => AppCard(
        onTap: () => context.push('${AppRoutes.gameHistory}/${game.id}'),
        child: Row(
          children: [
            _ResultBadge(result: game.result),
            const SizedBox(width: AppSpacing.sm),
            const AppAvatar(initials: 'P', size: 42),
            const SizedBox(width: AppSpacing.sm),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(game.opponent, style: AppTextStyles.bodyStrong),
                  Text('${game.opponentRating} • ${game.color.name} • ${game.timeControl}', style: AppTextStyles.caption),
                  const SizedBox(height: AppSpacing.xs),
                  Text(_dateLabel(game.playedAt), style: AppTextStyles.caption),
                ],
              ),
            ),
            Column(
              crossAxisAlignment: CrossAxisAlignment.end,
              children: [
                Text(_ratingLabel(game.ratingChange), style: AppTextStyles.bodyStrong.copyWith(color: _ratingColor(game.ratingChange))),
                if (game.accuracy != null) Text('${game.accuracy!.toStringAsFixed(1)}%', style: AppTextStyles.caption),
                const Icon(Icons.chevron_right_rounded, color: AppColors.muted),
              ],
            ),
          ],
        ),
      );
}

class _ResultBadge extends StatelessWidget {
  const _ResultBadge({required this.result});

  final HistoryResult result;

  @override
  Widget build(BuildContext context) => Container(
        width: 10,
        height: 46,
        decoration: BoxDecoration(
          color: _resultColor(result),
          borderRadius: AppRadius.pillRadius,
        ),
      );
}

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

String _dateLabel(DateTime date) =>
    '${date.day}/${date.month}/${date.year} • ${date.hour.toString().padLeft(2, '0')}:${date.minute.toString().padLeft(2, '0')}';
