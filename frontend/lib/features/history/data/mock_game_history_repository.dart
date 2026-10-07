import '../domain/game_history_models.dart';
import '../domain/game_history_repository.dart';

class MockGameHistoryRepository implements GameHistoryRepository {
  static final _games = [
    GameHistoryItem(
      id: 'demo-game',
      opponent: 'Rohan',
      opponentRating: 1223,
      result: HistoryResult.win,
      resultLabel: 'Win',
      color: PlayedColor.white,
      timeControl: '3 + 2',
      playedAt: DateTime(2026, 10, 7, 18, 42),
      ratingChange: 12,
      moveCount: 52,
      duration: const Duration(minutes: 8, seconds: 14),
      accuracy: 87.6,
      endReason: 'Checkmate',
    ),
    GameHistoryItem(
      id: 'game-002',
      opponent: 'Meera',
      opponentRating: 1301,
      result: HistoryResult.draw,
      resultLabel: 'Draw',
      color: PlayedColor.black,
      timeControl: '10 + 0',
      playedAt: DateTime(2026, 10, 6, 21, 10),
      ratingChange: 0,
      moveCount: 44,
      duration: const Duration(minutes: 19, seconds: 2),
      accuracy: 81.4,
      endReason: 'Agreement',
    ),
    GameHistoryItem(
      id: 'game-003',
      opponent: 'Dev',
      opponentRating: 1198,
      result: HistoryResult.loss,
      resultLabel: 'Loss',
      color: PlayedColor.white,
      timeControl: '5 + 0',
      playedAt: DateTime(2026, 10, 5, 17, 25),
      ratingChange: -9,
      moveCount: 36,
      duration: const Duration(minutes: 7, seconds: 48),
      accuracy: 74.2,
      endReason: 'Resignation',
    ),
    GameHistoryItem(
      id: 'game-004',
      opponent: 'Aarav',
      opponentRating: 1275,
      result: HistoryResult.abandoned,
      resultLabel: 'Abandoned',
      color: PlayedColor.black,
      timeControl: '3 + 0',
      playedAt: DateTime(2026, 10, 4, 11, 4),
      ratingChange: -5,
      moveCount: 9,
      duration: const Duration(minutes: 1, seconds: 10),
      accuracy: null,
      endReason: 'Abandoned game',
    ),
  ];

  @override
  Future<List<GameHistoryItem>> listGames() async => _games;

  @override
  Future<GameHistoryItem> getGame(String gameId) async =>
      _games.firstWhere((game) => game.id == gameId);
}
