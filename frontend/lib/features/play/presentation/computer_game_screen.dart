import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/app_widgets.dart';
import '../domain/play_models.dart';
import 'play_controller.dart';

class ComputerGameScreen extends StatefulWidget {
  const ComputerGameScreen({super.key});

  @override
  State<ComputerGameScreen> createState() => _ComputerGameScreenState();
}

class _ComputerGameScreenState extends State<ComputerGameScreen> {
  ComputerDifficulty _difficulty = ComputerDifficulty.medium;
  PlayerColor _color = PlayerColor.random;
  TimeControl _timeControl = const TimeControl(10, 5);

  @override
  Widget build(BuildContext context) => Scaffold(
        appBar: AppBar(title: const Text('Play Computer')),
        body: SafeArea(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(AppSpacing.page),
            child: Center(
              child: ConstrainedBox(
                constraints: const BoxConstraints(maxWidth: 560),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text('Set up your game', style: AppTextStyles.headline),
                    const SizedBox(height: AppSpacing.lg),
                    const Text('Difficulty', style: AppTextStyles.titleSmall),
                    const SizedBox(height: AppSpacing.sm),
                    Wrap(
                      spacing: AppSpacing.sm,
                      children: ComputerDifficulty.values
                          .map((value) => AppChoicePill(
                                label: _difficultyLabel(value),
                                selected: _difficulty == value,
                                onSelected: (_) => setState(() => _difficulty = value),
                              ))
                          .toList(),
                    ),
                    const SizedBox(height: AppSpacing.lg),
                    const Text('Your color', style: AppTextStyles.titleSmall),
                    const SizedBox(height: AppSpacing.sm),
                    Wrap(
                      spacing: AppSpacing.sm,
                      children: PlayerColor.values
                          .map((value) => AppChoicePill(
                                label: _colorLabel(value),
                                selected: _color == value,
                                onSelected: (_) => setState(() => _color = value),
                              ))
                          .toList(),
                    ),
                    const SizedBox(height: AppSpacing.lg),
                    const Text('Time control', style: AppTextStyles.titleSmall),
                    const SizedBox(height: AppSpacing.sm),
                    Wrap(
                      spacing: AppSpacing.sm,
                      runSpacing: AppSpacing.sm,
                      children: _computerControls
                          .map((value) => AppChoicePill(
                                label: value.label,
                                selected: _timeControl.label == value.label,
                                onSelected: (_) => setState(() => _timeControl = value),
                              ))
                          .toList(),
                    ),
                    const SizedBox(height: AppSpacing.xl),
                    Consumer(
                      builder: (context, ref, _) => AppButton(
                        label: 'Start game',
                        icon: Icons.play_arrow_rounded,
                        expand: true,
                        onPressed: () async {
                          final id = await ref.read(computerGameRepositoryProvider).createGame(
                                difficulty: _difficulty,
                                color: _color,
                                timeControl: _timeControl,
                              );
                          if (context.mounted) context.go('/game/$id');
                        },
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
        ),
      );
}

String _difficultyLabel(ComputerDifficulty value) => switch (value) {
      ComputerDifficulty.easy => 'Easy',
      ComputerDifficulty.medium => 'Medium',
      ComputerDifficulty.hard => 'Hard',
      ComputerDifficulty.grandmaster => 'Grandmaster',
    };

String _colorLabel(PlayerColor value) => switch (value) {
      PlayerColor.white => 'White',
      PlayerColor.black => 'Black',
      PlayerColor.random => 'Random',
    };

const _computerControls = [TimeControl(1, 0), TimeControl(5, 0), TimeControl(10, 5)];
