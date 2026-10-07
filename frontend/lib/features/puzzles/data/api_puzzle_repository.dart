import '../../../core/network/api_client.dart';
import '../../../core/widgets/chess/chess_models.dart';
import '../domain/puzzle_models.dart';
import '../domain/puzzle_repository.dart';

class ApiPuzzleRepository implements PuzzleRepository {
  ApiPuzzleRepository(this._client);

  final ApiClient _client;

  @override
  Future<Puzzle> getDailyPuzzle() async =>
      _parsePuzzle(await _client.get('/puzzles/daily'));

  @override
  Future<Puzzle> getPuzzle(String puzzleId) async =>
      _parsePuzzle(await _client.get('/puzzles/$puzzleId'));

  @override
  Future<PuzzleStats> getStats() async {
    final response = await _client.get('/puzzles/rating');
    return PuzzleStats(
      rating: _int(response, 'rating'),
      solved: 0,
      attempted: 0,
      streak: 0,
    );
  }

  @override
  Future<List<Puzzle>> getHistory() {
    return Future.error(const ApiException(
      501,
      'Puzzle attempt history is not exposed by the backend yet.',
    ));
  }

  @override
  Future<PuzzleMove?> submitMove(String puzzleId, PuzzleMove move) async {
    final response = await _client.post(
      '/puzzles/$puzzleId/attempts',
      body: {
        'moves': [move.uci],
      },
    );
    final correct = response['correct'] == true;
    return correct ? move : null;
  }

  @override
  Future<String?> getHint(String puzzleId) {
    return Future.error(const ApiException(
      501,
      'Puzzle hints are not exposed by the backend yet.',
    ));
  }

  @override
  Future<List<PuzzleMove>> getSolution(String puzzleId) {
    return Future.error(const ApiException(
      501,
      'Puzzle solutions are intentionally not exposed by the backend.',
    ));
  }

  Puzzle _parsePuzzle(Map<String, dynamic> data) {
    final sideToMove = _sideToMove(data['fen']?.toString());
    return Puzzle(
      id: data['id']?.toString() ?? '',
      fen: data['fen']?.toString() ?? '',
      sideToMove: sideToMove,
      themes: _strings(data['themes']),
      difficulty: data['difficulty']?.toString() ?? '',
      rating: _int(data, 'rating'),
      explanation: data['explanation']?.toString() ?? '',
    );
  }

  ChessColor _sideToMove(String? fen) {
    final fields = fen?.split(' ') ?? const [];
    return fields.length > 1 && fields[1] == 'b'
        ? ChessColor.black
        : ChessColor.white;
  }

  List<String> _strings(Object? value) =>
      value is List ? value.map((item) => item.toString()).toList() : const [];

  int _int(Map<String, dynamic> value, String key) =>
      (value[key] as num?)?.toInt() ?? 0;
}
