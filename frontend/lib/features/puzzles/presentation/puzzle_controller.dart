import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/network_providers.dart';
import '../data/api_puzzle_repository.dart';
import '../domain/puzzle_models.dart';
import '../domain/puzzle_repository.dart';

final puzzleRepositoryProvider = Provider<PuzzleRepository>(
  (ref) => ApiPuzzleRepository(ref.watch(networkApiClientProvider)),
);

final puzzleStatsProvider = FutureProvider<PuzzleStats>(
  (ref) => ref.watch(puzzleRepositoryProvider).getStats(),
);

final dailyPuzzleProvider = FutureProvider<Puzzle>(
  (ref) => ref.watch(puzzleRepositoryProvider).getDailyPuzzle(),
);

final puzzleControllerProvider = AutoDisposeNotifierProviderFamily<
    PuzzleController, AsyncValue<PuzzleState>, String>(PuzzleController.new);

class PuzzleController
    extends AutoDisposeFamilyNotifier<AsyncValue<PuzzleState>, String> {
  late PuzzleRepository _repository;

  @override
  AsyncValue<PuzzleState> build(String puzzleId) {
    ref.keepAlive();
    _repository = ref.watch(puzzleRepositoryProvider);
    _load(puzzleId);
    return const AsyncLoading();
  }

  Future<void> _load(String puzzleId) async {
    try {
      final puzzle = await _repository.getPuzzle(puzzleId);
      state = AsyncData(PuzzleState(
        puzzle: puzzle,
        status: PuzzleStatus.active,
        moves: const [],
        currentFen: puzzle.fen,
      ));
    } catch (error, stackTrace) {
      state = AsyncError(error, stackTrace);
    }
  }

  Future<void> submitMove(String from, String to) async {
    final current = state.valueOrNull;
    if (current == null || current.status != PuzzleStatus.active) return;
    final move = PuzzleMove(from: from, to: to, uci: '$from$to');
    final response = await _repository.submitMove(current.puzzle.id, move);
    final moves = [...current.moves, move];
    if (response == null) {
      state = AsyncData(current.copyWith(
        status: PuzzleStatus.failed,
        moves: moves,
        result: const PuzzleResult(
          success: false,
          ratingChange: null,
          accuracy: null,
          message: 'The backend rejected that move.',
        ),
      ));
      return;
    }
    state = AsyncData(current.copyWith(
      status: PuzzleStatus.solved,
      moves: moves,
      result: const PuzzleResult(
        success: true,
        ratingChange: null,
        accuracy: null,
        message: 'The backend accepted your solution.',
      ),
    ));
  }

  Future<void> showHint() async {
    final current = state.valueOrNull;
    if (current == null) return;
    final hint = await _repository.getHint(current.puzzle.id);
    state = AsyncData(current.copyWith(hintSquare: hint));
  }

  Future<void> retry() async => _load(arg);
}
