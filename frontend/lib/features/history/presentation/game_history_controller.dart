import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/mock_game_history_repository.dart';
import '../domain/game_history_models.dart';
import '../domain/game_history_repository.dart';

final gameHistoryRepositoryProvider = Provider<GameHistoryRepository>(
  (_) => MockGameHistoryRepository(),
);

final gameHistoryProvider = FutureProvider<GameHistoryState>((ref) async {
  final games = await ref.watch(gameHistoryRepositoryProvider).listGames();
  return GameHistoryState(games: games);
});

final gameHistoryDetailProvider =
    FutureProvider.family<GameHistoryItem, String>((ref, gameId) {
  return ref.watch(gameHistoryRepositoryProvider).getGame(gameId);
});
