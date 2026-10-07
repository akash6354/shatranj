import 'game_history_models.dart';

abstract interface class GameHistoryRepository {
  Future<List<GameHistoryItem>> listGames();

  Future<GameHistoryItem> getGame(String gameId);
}
