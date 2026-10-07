import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/chess/chess_widgets.dart';
import '../../../core/widgets/app_widgets.dart';
import '../domain/review_models.dart';
import 'review_controller.dart';

class ReviewScreen extends ConsumerWidget {
  const ReviewScreen({required this.gameId, super.key});

  final String gameId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final review = ref.watch(reviewProvider(gameId));
    return Scaffold(
      appBar: AppBar(
        leading: BackButton(onPressed: () => Navigator.maybePop(context)),
        title: const Text('Game Review'),
      ),
      body: SafeArea(
        child: review.when(
          loading: () => const AppLoading(message: 'Preparing your review'),
          error: (error, _) => AppError(
            title: 'Review unavailable',
            message: error.toString(),
            onRetry: () => ref.invalidate(reviewProvider(gameId)),
          ),
          data: (data) => _ReviewContent(review: data),
        ),
      ),
    );
  }
}

class _ReviewContent extends ConsumerWidget {
  const _ReviewContent({required this.review});

  final GameReview review;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final tab = ref.watch(reviewTabProvider);
    return Column(
      children: [
        _ReviewTabs(
          selected: tab,
          onSelected: (value) =>
              ref.read(reviewTabProvider.notifier).state = value,
        ),
        Expanded(
          child: switch (tab) {
            ReviewTab.overview => _Overview(review: review),
            ReviewTab.moveByMove => _MoveByMove(review: review),
            ReviewTab.insights => _Insights(review: review),
          },
        ),
      ],
    );
  }
}

class _ReviewTabs extends StatelessWidget {
  const _ReviewTabs({required this.selected, required this.onSelected});

  final ReviewTab selected;
  final ValueChanged<ReviewTab> onSelected;

  @override
  Widget build(BuildContext context) => Padding(
        padding: const EdgeInsets.fromLTRB(AppSpacing.page, AppSpacing.xs, AppSpacing.page, AppSpacing.sm),
        child: Row(
          children: ReviewTab.values
              .map(
                (tab) => Expanded(
                  child: AppChoicePill(
                    label: switch (tab) {
                      ReviewTab.overview => 'Overview',
                      ReviewTab.moveByMove => 'Move by Move',
                      ReviewTab.insights => 'Insights',
                    },
                    selected: selected == tab,
                    onSelected: (_) => onSelected(tab),
                  ),
                ),
              )
              .toList(),
        ),
      );
}

class _Overview extends StatelessWidget {
  const _Overview({required this.review});

  final GameReview review;

  @override
  Widget build(BuildContext context) => LayoutBuilder(
        builder: (context, constraints) {
          final wide = constraints.maxWidth > 760;
          return SingleChildScrollView(
            padding: const EdgeInsets.fromLTRB(AppSpacing.page, AppSpacing.sm, AppSpacing.page, AppSpacing.xl),
            child: Center(
              child: ConstrainedBox(
                constraints: const BoxConstraints(maxWidth: 980),
                child: Column(
                  children: [
                    _ResultCard(review: review),
                    const SizedBox(height: AppSpacing.md),
                    if (wide)
                      Row(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Expanded(child: _AccuracyCard(review: review)),
                          const SizedBox(width: AppSpacing.md),
                          Expanded(child: _EvaluationGraph(review: review)),
                        ],
                      )
                    else ...[
                      _AccuracyCard(review: review),
                      const SizedBox(height: AppSpacing.md),
                      _EvaluationGraph(review: review),
                    ],
                    const SizedBox(height: AppSpacing.md),
                    _PhaseStatsCard(phases: review.phases),
                  ],
                ),
              ),
            ),
          );
        },
      );
}

class _ResultCard extends StatelessWidget {
  const _ResultCard({required this.review});

  final GameReview review;

  @override
  Widget build(BuildContext context) => AppGradientCard(
        child: Column(
          children: [
            Row(
              children: [
                Expanded(child: _PlayerSummary(player: review.white)),
                Column(
                  children: [
                    const Icon(Icons.emoji_events_rounded, color: AppColors.gold, size: 34),
                    const SizedBox(height: AppSpacing.xs),
                    Text(review.result, style: AppTextStyles.title),
                    Text(review.resultLabel, style: AppTextStyles.caption, textAlign: TextAlign.center),
                  ],
                ),
                Expanded(child: _PlayerSummary(player: review.black, rightAligned: true)),
              ],
            ),
            const Divider(height: AppSpacing.lg),
            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                const Icon(Icons.star_rounded, color: AppColors.gold, size: 18),
                const SizedBox(width: AppSpacing.xs),
                Text('Game rating ${review.rating}', style: AppTextStyles.bodyStrong),
                const SizedBox(width: AppSpacing.sm),
                AppPill(label: '+${review.ratingChange}', color: AppColors.success, icon: Icons.trending_up_rounded),
              ],
            ),
          ],
        ),
      );
}

class _PlayerSummary extends StatelessWidget {
  const _PlayerSummary({required this.player, this.rightAligned = false});

  final ReviewPlayer player;
  final bool rightAligned;

  @override
  Widget build(BuildContext context) => Column(
        crossAxisAlignment: rightAligned ? CrossAxisAlignment.end : CrossAxisAlignment.start,
        children: [
          const AppAvatar(initials: 'P', size: 42),
          const SizedBox(height: AppSpacing.xs),
          Text(player.name, style: AppTextStyles.bodyStrong),
          Text('${player.rating} • ${player.color}', style: AppTextStyles.caption),
          const SizedBox(height: AppSpacing.xs),
          Text('${player.accuracy.toStringAsFixed(1)}%', style: AppTextStyles.title),
          const Text('Accuracy', style: AppTextStyles.caption),
        ],
      );
}

class _AccuracyCard extends StatelessWidget {
  const _AccuracyCard({required this.review});

  final GameReview review;

  @override
  Widget build(BuildContext context) => AppCard(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const AppSectionHeader(title: 'Game statistics'),
            const SizedBox(height: AppSpacing.md),
            Row(
              children: [
                _AccuracyScore(label: 'You', value: review.white.accuracy),
                _AccuracyScore(label: 'Opponent', value: review.black.accuracy),
              ],
            ),
            const SizedBox(height: AppSpacing.md),
            _MoveCounts(stats: review.stats),
          ],
        ),
      );
}

class _AccuracyScore extends StatelessWidget {
  const _AccuracyScore({required this.label, required this.value});

  final String label;
  final double value;

  @override
  Widget build(BuildContext context) => Expanded(
        child: Column(
          children: [
            Text(value.toStringAsFixed(1), style: AppTextStyles.numeric),
            Text(label, style: AppTextStyles.caption),
            const SizedBox(height: AppSpacing.xs),
            LinearProgressIndicator(
              value: value / 100,
              minHeight: 6,
              borderRadius: AppRadius.pillRadius,
              color: AppColors.mint,
              backgroundColor: AppColors.border,
            ),
          ],
        ),
      );
}

class _MoveCounts extends StatelessWidget {
  const _MoveCounts({required this.stats});

  final ReviewStats stats;

  @override
  Widget build(BuildContext context) {
    final counts = [
      ('Brilliant', stats.brilliant, AppColors.info),
      ('Best', stats.best, AppColors.success),
      ('Good', stats.good, AppColors.mint),
      ('Inaccuracy', stats.inaccuracies, AppColors.warning),
      ('Mistake', stats.mistakes, AppColors.dangerSoft),
      ('Blunder', stats.blunders, AppColors.danger),
      ('Missed win', stats.missedWins, AppColors.danger),
    ];
    return Wrap(
      spacing: AppSpacing.sm,
      runSpacing: AppSpacing.sm,
      children: counts
          .map(
            (item) => AppPill(
              label: '${item.$1} ${item.$2}',
              color: item.$3,
              icon: Icons.circle,
            ),
          )
          .toList(),
    );
  }
}

class _EvaluationGraph extends StatelessWidget {
  const _EvaluationGraph({required this.review});

  final GameReview review;

  @override
  Widget build(BuildContext context) => AppCard(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const AppSectionHeader(title: 'Evaluation graph'),
            const SizedBox(height: AppSpacing.md),
            SizedBox(
              height: 150,
              width: double.infinity,
              child: CustomPaint(
                painter: _GraphPainter(review.moves.map((move) => move.evaluation).toList()),
              ),
            ),
            const SizedBox(height: AppSpacing.sm),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: review.phases
                  .map((phase) => Text(phase.name, style: AppTextStyles.caption))
                  .toList(),
            ),
          ],
        ),
      );
}

class _GraphPainter extends CustomPainter {
  _GraphPainter(this.values);

  final List<double> values;

  @override
  void paint(Canvas canvas, Size size) {
    final gridPaint = Paint()..color = AppColors.border;
    final linePaint = Paint()
      ..color = AppColors.mint
      ..strokeWidth = 3
      ..style = PaintingStyle.stroke;
    for (var index = 1; index < 4; index++) {
      final y = size.height * index / 4;
      canvas.drawLine(Offset(0, y), Offset(size.width, y), gridPaint);
    }
    if (values.length < 2) return;
    final path = Path();
    for (var index = 0; index < values.length; index++) {
      final x = size.width * index / (values.length - 1);
      final normalized = ((values[index] + 1) / 2).clamp(0.0, 1.0);
      final y = size.height - normalized * size.height;
      if (index == 0) {
        path.moveTo(x, y);
      } else {
        path.lineTo(x, y);
      }
    }
    canvas.drawPath(path, linePaint);
  }

  @override
  bool shouldRepaint(covariant _GraphPainter oldDelegate) => oldDelegate.values != values;
}

class _PhaseStatsCard extends StatelessWidget {
  const _PhaseStatsCard({required this.phases});

  final List<PhaseStats> phases;

  @override
  Widget build(BuildContext context) => AppCard(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const AppSectionHeader(title: 'Game phases'),
            const SizedBox(height: AppSpacing.sm),
            ...phases.map(
              (phase) => Padding(
                padding: const EdgeInsets.symmetric(vertical: AppSpacing.xs),
                child: Row(
                  children: [
                    Expanded(child: Text(phase.name, style: AppTextStyles.bodyStrong)),
                    Text('${phase.accuracy.toStringAsFixed(0)}%', style: AppTextStyles.bodyStrong),
                    const SizedBox(width: AppSpacing.sm),
                    Text('${phase.moves} moves • ${phase.highlight}', style: AppTextStyles.caption),
                  ],
                ),
              ),
            ),
          ],
        ),
      );
}

class _MoveByMove extends ConsumerWidget {
  const _MoveByMove({required this.review});

  final GameReview review;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final selected = ref.watch(selectedReviewMoveProvider);
    final move = review.moves[selected.clamp(0, review.moves.length - 1)];
    return LayoutBuilder(
      builder: (context, constraints) {
        final wide = constraints.maxWidth > 760;
        final detail = _MoveDetail(
          review: review,
          move: move,
          index: selected,
          onPrevious: selected == 0 ? null : () => ref.read(selectedReviewMoveProvider.notifier).state--,
          onNext: selected == review.moves.length - 1 ? null : () => ref.read(selectedReviewMoveProvider.notifier).state++,
        );
        return SingleChildScrollView(
          padding: const EdgeInsets.fromLTRB(AppSpacing.page, AppSpacing.sm, AppSpacing.page, AppSpacing.xl),
          child: wide
              ? Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Expanded(flex: 3, child: detail),
                    const SizedBox(width: AppSpacing.md),
                    Expanded(flex: 2, child: _MoveList(review: review, selected: selected, onSelected: (value) => ref.read(selectedReviewMoveProvider.notifier).state = value)),
                  ],
                )
              : Column(
                  children: [
                    detail,
                    const SizedBox(height: AppSpacing.md),
                    _MoveList(review: review, selected: selected, onSelected: (value) => ref.read(selectedReviewMoveProvider.notifier).state = value),
                  ],
                ),
        );
      },
    );
  }
}

class _MoveDetail extends StatelessWidget {
  const _MoveDetail({required this.review, required this.move, required this.index, required this.onPrevious, required this.onNext});

  final GameReview review;
  final ReviewMove move;
  final int index;
  final VoidCallback? onPrevious;
  final VoidCallback? onNext;

  @override
  Widget build(BuildContext context) => Column(
        children: [
          AppCard(
            padding: EdgeInsets.zero,
            child: ChessBoard(
              fen: move.fenAfter,
              lastMove: (from: move.white.length == 2 ? move.white.substring(0, 2) : 'e2', to: 'e4'),
            ),
          ),
          const SizedBox(height: AppSpacing.md),
          AppCard(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    _ClassificationPill(classification: move.classification),
                    const Spacer(),
                    Text('Move ${index + 1}', style: AppTextStyles.caption),
                  ],
                ),
                const SizedBox(height: AppSpacing.sm),
                Text(move.explanation, style: AppTextStyles.body),
                if (move.bestMove != null) ...[
                  const SizedBox(height: AppSpacing.sm),
                  Text('Best move: ${move.bestMove}', style: AppTextStyles.bodyStrong),
                ],
                const SizedBox(height: AppSpacing.sm),
                Row(
                  children: [
                    Text('Evaluation ${move.evaluation >= 0 ? '+' : ''}${move.evaluation.toStringAsFixed(1)}', style: AppTextStyles.caption),
                    const Spacer(),
                    TextButton.icon(onPressed: onPrevious, icon: const Icon(Icons.chevron_left), label: const Text('Previous')),
                    FilledButton.icon(onPressed: onNext, icon: const Icon(Icons.chevron_right), label: const Text('Next')),
                  ],
                ),
              ],
            ),
          ),
        ],
      );
}

class _MoveList extends StatelessWidget {
  const _MoveList({required this.review, required this.selected, required this.onSelected});

  final GameReview review;
  final int selected;
  final ValueChanged<int> onSelected;

  @override
  Widget build(BuildContext context) => AppCard(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const AppSectionHeader(title: 'Moves', actionLabel: 'All moves'),
            const SizedBox(height: AppSpacing.sm),
            ...review.moves.asMap().entries.map(
              (entry) {
                final move = entry.value;
                return ListTile(
                  dense: true,
                  selected: selected == entry.key,
                  selectedTileColor: AppColors.gold.withValues(alpha: .12),
                  onTap: () => onSelected(entry.key),
                  leading: Text('${move.number}.', style: AppTextStyles.caption),
                  title: Text('${move.white}   ${move.black}', style: AppTextStyles.bodyStrong),
                  trailing: _ClassificationIcon(classification: move.classification),
                );
              },
            ),
          ],
        ),
      );
}

class _ClassificationPill extends StatelessWidget {
  const _ClassificationPill({required this.classification});

  final MoveClassification classification;

  @override
  Widget build(BuildContext context) => AppPill(
        label: classification.name,
        icon: Icons.circle,
        color: _classificationColor(classification),
      );
}

class _ClassificationIcon extends StatelessWidget {
  const _ClassificationIcon({required this.classification});

  final MoveClassification classification;

  @override
  Widget build(BuildContext context) => Icon(Icons.circle, size: 12, color: _classificationColor(classification));
}

Color _classificationColor(MoveClassification classification) => switch (classification) {
      MoveClassification.brilliant => AppColors.info,
      MoveClassification.best => AppColors.success,
      MoveClassification.good => AppColors.mint,
      MoveClassification.inaccuracy => AppColors.warning,
      MoveClassification.mistake => AppColors.dangerSoft,
      MoveClassification.blunder => AppColors.danger,
      MoveClassification.missedWin => AppColors.danger,
    };

class _Insights extends StatelessWidget {
  const _Insights({required this.review});

  final GameReview review;

  @override
  Widget build(BuildContext context) => SingleChildScrollView(
        padding: const EdgeInsets.fromLTRB(AppSpacing.page, AppSpacing.sm, AppSpacing.page, AppSpacing.xl),
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 760),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const AppSectionHeader(title: 'Insights', subtitle: 'Patterns from your game'),
                const SizedBox(height: AppSpacing.md),
                Row(
                  children: ReviewTab.values
                      .map(
                        (tab) => Expanded(
                          child: AppChoicePill(
                            label: switch (tab) {
                              ReviewTab.overview => 'Summary',
                              ReviewTab.moveByMove => 'Opening',
                              ReviewTab.insights => 'Middlegame',
                            },
                            selected: tab == ReviewTab.overview,
                            onSelected: (_) {},
                          ),
                        ),
                      )
                      .toList(),
                ),
                const SizedBox(height: AppSpacing.md),
                ...review.insights.map(
                  (insight) => AppCard(
                    margin: const EdgeInsets.only(bottom: AppSpacing.sm),
                    child: Row(
                      children: [
                        Icon(Icons.auto_awesome_rounded, color: Color(insight.color), size: 26),
                        const SizedBox(width: AppSpacing.sm),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(insight.title, style: AppTextStyles.bodyStrong),
                              const SizedBox(height: AppSpacing.xxs),
                              Text(insight.description, style: AppTextStyles.caption),
                            ],
                          ),
                        ),
                        const Icon(Icons.chevron_right_rounded, color: AppColors.muted),
                      ],
                    ),
                  ),
                ),
                const SizedBox(height: AppSpacing.md),
                const AppSectionHeader(title: 'Key moments', actionLabel: 'See all'),
                const SizedBox(height: AppSpacing.sm),
                ...review.moves.where((move) => move.classification != MoveClassification.good).map(
                  (move) => AppCard(
                    margin: const EdgeInsets.only(bottom: AppSpacing.sm),
                    padding: const EdgeInsets.all(AppSpacing.sm),
                    child: Row(
                      children: [
                        SizedBox(width: 72, child: ChessBoard(fen: move.fenAfter, showCoordinates: false)),
                        const SizedBox(width: AppSpacing.sm),
                        Expanded(
                          child: Text('Move ${move.number} • ${move.classification.name}', style: AppTextStyles.bodyStrong),
                        ),
                        _ClassificationIcon(classification: move.classification),
                      ],
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      );
}
