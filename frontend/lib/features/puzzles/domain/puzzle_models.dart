import '../../../core/widgets/chess/chess_models.dart';

enum PuzzleStatus { idle, active, solved, failed }

class Puzzle {
  const Puzzle({
    required this.id,
    required this.fen,
    required this.sideToMove,
    required this.themes,
    required this.difficulty,
    required this.rating,
    required this.explanation,
  });

  final String id;
  final String fen;
  final ChessColor sideToMove;
  final List<String> themes;
  final String difficulty;
  final int rating;
  final String explanation;
}

class PuzzleMove {
  const PuzzleMove({
    required this.from,
    required this.to,
    required this.uci,
    this.san,
  });

  final String from;
  final String to;
  final String uci;
  final String? san;
}

class PuzzleResult {
  const PuzzleResult({
    required this.success,
    required this.ratingChange,
    required this.accuracy,
    required this.message,
  });

  final bool success;
  final int? ratingChange;
  final double? accuracy;
  final String message;
}

class PuzzleState {
  const PuzzleState({
    required this.puzzle,
    required this.status,
    required this.moves,
    this.currentFen,
    this.selectedSquare,
    this.hintSquare,
    this.result,
    this.error,
  });

  final Puzzle puzzle;
  final PuzzleStatus status;
  final List<PuzzleMove> moves;
  final String? currentFen;
  final String? selectedSquare;
  final String? hintSquare;
  final PuzzleResult? result;
  final String? error;

  PuzzleState copyWith({
    PuzzleStatus? status,
    List<PuzzleMove>? moves,
    String? currentFen,
    String? selectedSquare,
    String? hintSquare,
    PuzzleResult? result,
    String? error,
  }) =>
      PuzzleState(
        puzzle: puzzle,
        status: status ?? this.status,
        moves: moves ?? this.moves,
        currentFen: currentFen ?? this.currentFen,
        selectedSquare: selectedSquare ?? this.selectedSquare,
        hintSquare: hintSquare ?? this.hintSquare,
        result: result ?? this.result,
        error: error,
      );
}

class PuzzleStats {
  const PuzzleStats({
    required this.rating,
    required this.solved,
    required this.attempted,
    required this.streak,
  });

  final int rating;
  final int solved;
  final int attempted;
  final int streak;
}
