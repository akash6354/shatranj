import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/app_widgets.dart';
import '../domain/play_models.dart';
import 'play_controller.dart';

class MatchmakingScreen extends ConsumerWidget {
  const MatchmakingScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(matchmakingProvider);
    ref.listen(matchmakingProvider, (_, next) {
      if (next.status == SearchStatus.found && context.mounted) {
        context.push('/play/matchmaking/found');
      }
    });
    return Scaffold(
      appBar: AppBar(title: const Text('Quick Match')),
      body: SafeArea(
        child: _MatchmakingContent(
          state: state,
          onTimeControl: (value) => ref.read(matchmakingProvider.notifier).setTimeControl(value),
          onRatingRange: (value) => ref.read(matchmakingProvider.notifier).setRatingRange(value),
          onSearch: () => ref.read(matchmakingProvider.notifier).search(),
          onCancel: () => ref.read(matchmakingProvider.notifier).cancel(),
        ),
      ),
    );
  }
}

class _MatchmakingContent extends StatelessWidget {
  const _MatchmakingContent({
    required this.state,
    required this.onTimeControl,
    required this.onRatingRange,
    required this.onSearch,
    required this.onCancel,
  });

  final MatchmakingState state;
  final ValueChanged<TimeControl> onTimeControl;
  final ValueChanged<int> onRatingRange;
  final VoidCallback onSearch;
  final VoidCallback onCancel;

  @override
  Widget build(BuildContext context) {
    final searching = state.status == SearchStatus.searching;
    return SingleChildScrollView(
      padding: const EdgeInsets.all(AppSpacing.page),
      child: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 560),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text('Find an opponent', style: AppTextStyles.headline),
              const SizedBox(height: AppSpacing.lg),
              const Text('Time control', style: AppTextStyles.titleSmall),
              const SizedBox(height: AppSpacing.sm),
              Wrap(
                spacing: AppSpacing.sm,
                runSpacing: AppSpacing.sm,
                children: _timeControls
                    .map((control) => AppChoicePill(
                          label: control.label,
                          selected: control.label == state.timeControl.label,
                          onSelected: (_) => onTimeControl(control),
                        ))
                    .toList(),
              ),
              const SizedBox(height: AppSpacing.lg),
              AppCard(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text('Rating range: ±${state.ratingRange}', style: AppTextStyles.bodyStrong),
                    Slider(
                      value: state.ratingRange.toDouble(),
                      min: 100,
                      max: 500,
                      divisions: 4,
                      label: '±${state.ratingRange}',
                      onChanged: searching ? null : (value) => onRatingRange(value.round()),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: AppSpacing.xl),
              if (searching) ...[
                const Center(child: CircularProgressIndicator()),
                const SizedBox(height: AppSpacing.md),
                const Center(child: Text('Searching for an opponent…', style: AppTextStyles.body)),
                const SizedBox(height: AppSpacing.lg),
                AppButton(label: 'Cancel search', variant: AppButtonVariant.outline, expand: true, onPressed: onCancel),
              ] else
                AppButton(label: 'Find opponent', icon: Icons.search_rounded, expand: true, onPressed: onSearch),
            ],
          ),
        ),
      ),
    );
  }
}

const _timeControls = [
  TimeControl(1, 0),
  TimeControl(3, 2),
  TimeControl(5, 0),
  TimeControl(10, 5),
];

class MatchFoundScreen extends ConsumerWidget {
  const MatchFoundScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(matchmakingProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Opponent found')),
      body: SafeArea(
        child: Center(
          child: Padding(
            padding: const EdgeInsets.all(AppSpacing.page),
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 480),
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const CircleAvatar(radius: 38, child: Icon(Icons.person_rounded, size: 38)),
                  const SizedBox(height: AppSpacing.md),
                  Text(state.opponentName ?? 'Opponent', style: AppTextStyles.headline),
                  Text('${state.opponentRating ?? 0} rating • ${state.timeControl.label}', style: AppTextStyles.body),
                  const SizedBox(height: AppSpacing.xl),
                  AppButton(label: 'Enter game', icon: Icons.sports_esports_rounded, expand: true, onPressed: () => context.go('/game/quick-match-${state.timeControl.label}')),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
