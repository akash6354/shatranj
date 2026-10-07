import '../../../core/widgets/chess/chess_models.dart';
import '../domain/puzzle_models.dart';
import '../domain/puzzle_repository.dart';

class MockPuzzleRepository implements PuzzleRepository {
  static const _puzzle = Puzzle(
    id: 'daily-001',
    fen: 'r1bqk2r/pppp1ppp/2n2n2/4p3/4P3/2N2N2/PPPP1PPP/R1BQKB1R w KQkq - 2 4',
    sideToMove: ChessColor.white,
    themes: ['Opening', 'Tactics'],
    difficulty: 'Medium',
    rating: 1382,
    explanation: 'Develop your pieces and look for the forcing continuation.',
  );

  static const _solution = [
    PuzzleMove(from: 'e2', to: 'e4', uci: 'e2e4', san: 'e4'),
    PuzzleMove(from: 'g1', to: 'f3', uci: 'g1f3', san: 'Nf3'),
  ];

  @override
  Future<Puzzle> getDailyPuzzle() async => _puzzle;

  @override
  Future<Puzzle> getPuzzle(String puzzleId) async => _puzzle;

  @override
  Future<PuzzleStats> getStats() async =>
      const PuzzleStats(rating: 1382, solved: 126, attempted: 148, streak: 7);

  @override
  Future<List<Puzzle>> getHistory() async => const [_puzzle];

  @override
  Future<PuzzleMove?> submitMove(String puzzleId, PuzzleMove move) async {
    final index = _solution.indexWhere((item) => item.uci == move.uci);
    return index >= 0 ? _solution[index] : null;
  }

  @override
  Future<String?> getHint(String puzzleId) async => _solution.first.from;

  @override
  Future<List<PuzzleMove>> getSolution(String puzzleId) async => _solution;
}
