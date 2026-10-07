import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/routing/app_router.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/app_widgets.dart';
import '../domain/puzzle_models.dart';
import 'puzzle_controller.dart';

class PuzzleResultScreen extends ConsumerWidget {
  const PuzzleResultScreen({required this.puzzleId, super.key});
  final String puzzleId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(puzzleControllerProvider(puzzleId));
    final result = state.valueOrNull?.result;
    return Scaffold(
      body: SafeArea(
        child: Center(
          child: result == null
              ? const AppLoading()
              : _ResultContent(puzzleId: puzzleId, result: result),
        ),
      ),
    );
  }
}

class _ResultContent extends ConsumerWidget {
  const _ResultContent({required this.puzzleId, required this.result});
  final String puzzleId;
  final PuzzleResult result;

  @override
  Widget build(BuildContext context, WidgetRef ref) => SingleChildScrollView(
    padding: const EdgeInsets.all(AppSpacing.page),
    child: ConstrainedBox(
      constraints: const BoxConstraints(maxWidth: 480),
      child: Column(
        children: [
          Icon(
            result.success ? Icons.celebration_rounded : Icons.refresh_rounded,
            color: result.success ? AppColors.success : AppColors.warning,
            size: 72,
          ),
          const SizedBox(height: AppSpacing.md),
          Text(
            result.success ? 'Puzzle solved!' : 'Not quite',
            style: AppTextStyles.headline,
          ),
          const SizedBox(height: AppSpacing.xs),
          Text(
            result.message,
            style: AppTextStyles.body,
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: AppSpacing.xl),
          AppCard(
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceAround,
              children: [
                _Metric(
                  value: result.ratingChange == null
                      ? '—'
                      : result.ratingChange! > 0
                      ? '+${result.ratingChange}'
                      : '${result.ratingChange}',
                  label: 'Rating change',
                ),
                _Metric(
                  value: result.accuracy == null
                      ? '—'
                      : '${result.accuracy!.toStringAsFixed(0)}%',
                  label: 'Accuracy',
                ),
              ],
            ),
          ),
          const SizedBox(height: AppSpacing.lg),
          AppButton(
            label: result.success ? 'Next puzzle' : 'Retry',
            icon: Icons.arrow_forward_rounded,
            expand: true,
            onPressed: () => context.go(
              '/puzzles/${result.success ? 'daily-001' : puzzleId}/play',
            ),
          ),
          const SizedBox(height: AppSpacing.sm),
          AppButton(
            label: 'Return to puzzles',
            variant: AppButtonVariant.outline,
            expand: true,
            onPressed: () => context.go(AppRoutes.puzzles),
          ),
        ],
      ),
    ),
  );
}

class _Metric extends StatelessWidget {
  const _Metric({required this.value, required this.label});
  final String value;
  final String label;

  @override
  Widget build(BuildContext context) => Column(
    children: [
      Text(value, style: AppTextStyles.numeric),
      Text(label, style: AppTextStyles.caption),
    ],
  );
}
