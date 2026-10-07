import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/app_widgets.dart';
import '../../../core/widgets/chess/chess_widgets.dart';
import '../domain/puzzle_models.dart';
import 'puzzle_controller.dart';

class PuzzlePlayScreen extends ConsumerStatefulWidget {
  const PuzzlePlayScreen({required this.puzzleId, super.key});
  final String puzzleId;

  @override
  ConsumerState<PuzzlePlayScreen> createState() => _PuzzlePlayScreenState();
}

class _PuzzlePlayScreenState extends ConsumerState<PuzzlePlayScreen> {
  String? _selected;

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(puzzleControllerProvider(widget.puzzleId));
    ref.listen(puzzleControllerProvider(widget.puzzleId), (_, next) {
      final result = next.valueOrNull?.result;
      if (result != null && context.mounted) {
        context.go('/puzzles/${widget.puzzleId}/result');
      }
    });
    return Scaffold(
      appBar: AppBar(
        leading: BackButton(onPressed: () => context.pop()),
        title: const Text('Puzzle'),
        actions: [
          IconButton(
            tooltip: 'Hint',
            onPressed: () => ref.read(puzzleControllerProvider(widget.puzzleId).notifier).showHint(),
            icon: const Icon(Icons.lightbulb_outline_rounded),
          ),
        ],
      ),
      body: SafeArea(
        child: state.when(
          loading: () => const AppLoading(message: 'Preparing puzzle'),
          error: (error, _) => AppError(title: 'Puzzle unavailable', message: error.toString()),
          data: (puzzleState) => _PuzzleBoardContent(
            state: puzzleState,
            selected: _selected,
            onTap: _onSquareTap,
            onHint: () => ref.read(puzzleControllerProvider(widget.puzzleId).notifier).showHint(),
            onSolution: () => _showSolution(puzzleState),
          ),
        ),
      ),
    );
  }

  void _onSquareTap(String square) {
    final state = ref.read(puzzleControllerProvider(widget.puzzleId)).valueOrNull;
    if (state == null || state.status != PuzzleStatus.active) return;
    if (_selected == null) {
      setState(() => _selected = square);
    } else {
      final from = _selected!;
      setState(() => _selected = null);
      ref.read(puzzleControllerProvider(widget.puzzleId).notifier).submitMove(from, square);
    }
  }

  Future<void> _showSolution(PuzzleState state) async {
    final solution = await ref.read(puzzleRepositoryProvider).getSolution(state.puzzle.id);
    if (!mounted) return;
    showDialog<void>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Solution'),
        content: Text(solution.map((move) => move.san ?? move.uci).join('  ')),
        actions: [TextButton(onPressed: () => context.pop(), child: const Text('Close'))],
      ),
    );
  }
}

class _PuzzleBoardContent extends StatelessWidget {
  const _PuzzleBoardContent({
    required this.state,
    required this.selected,
    required this.onTap,
    required this.onHint,
    required this.onSolution,
  });

  final PuzzleState state;
  final String? selected;
  final ValueChanged<String> onTap;
  final VoidCallback onHint;
  final VoidCallback onSolution;

  @override
  Widget build(BuildContext context) => SingleChildScrollView(
        padding: const EdgeInsets.fromLTRB(AppSpacing.page, AppSpacing.sm, AppSpacing.page, AppSpacing.xl),
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 640),
            child: Column(
              children: [
                Row(
                  children: [
                    AppPill(label: state.puzzle.difficulty, color: AppColors.gold),
                    const SizedBox(width: AppSpacing.sm),
                    AppPill(label: '${state.puzzle.rating} rating', icon: Icons.star_rounded),
                    const Spacer(),
                    Text('${state.moves.length} moves', style: AppTextStyles.caption),
                  ],
                ),
                const SizedBox(height: AppSpacing.md),
                AppCard(
                  padding: EdgeInsets.zero,
                  child: ChessBoard(
                    fen: state.currentFen ?? state.puzzle.fen,
                    orientation: state.puzzle.sideToMove == ChessColor.white
                        ? BoardOrientation.white
                        : BoardOrientation.black,
                    selectedSquare: selected,
                    onSquareTap: onTap,
                  ),
                ),
                const SizedBox(height: AppSpacing.md),
                AppCard(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          const Icon(Icons.gps_fixed_rounded, color: AppColors.gold),
                          const SizedBox(width: AppSpacing.sm),
                          const Expanded(child: Text('Find the best move', style: AppTextStyles.titleSmall)),
                          Text(state.puzzle.sideToMove == ChessColor.white ? 'White to move' : 'Black to move', style: AppTextStyles.caption),
                        ],
                      ),
                      const SizedBox(height: AppSpacing.sm),
                      Text(state.puzzle.explanation, style: AppTextStyles.body),
                      if (state.hintSquare != null) ...[
                        const SizedBox(height: AppSpacing.sm),
                        AppPill(label: 'Hint: look at ${state.hintSquare}', icon: Icons.lightbulb_rounded, color: AppColors.warning),
                      ],
                    ],
                  ),
                ),
                const SizedBox(height: AppSpacing.md),
                Row(
                  children: [
                    Expanded(child: AppButton(label: 'Hint', icon: Icons.lightbulb_outline, variant: AppButtonVariant.outline, onPressed: onHint)),
                    const SizedBox(width: AppSpacing.sm),
                    Expanded(child: AppButton(label: 'Show solution', icon: Icons.visibility_outlined, variant: AppButtonVariant.ghost, onPressed: onSolution)),
                  ],
                ),
              ],
            ),
          ),
        ),
      );
}
