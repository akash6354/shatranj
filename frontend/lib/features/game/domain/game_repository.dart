import 'game_models.dart';
import 'game_socket.dart';

abstract interface class GameRepository {
  Future<GameState> getGame(String gameId);

  GameSocket createSocket();
}
