import 'puzzle_models.dart';

abstract interface class PuzzleRepository {
  Future<Puzzle> getDailyPuzzle();
  Future<Puzzle> getPuzzle(String puzzleId);
  Future<PuzzleStats> getStats();
  Future<List<Puzzle>> getHistory();
  Future<PuzzleMove?> submitMove(String puzzleId, PuzzleMove move);
  Future<String?> getHint(String puzzleId);
  Future<List<PuzzleMove>> getSolution(String puzzleId);
}
