import 'game_models.dart';

abstract interface class GameSocket {
  Stream<GameState> get events;

  Future<void> connect(String gameId);

  Future<void> sendMove(String uci);

  Future<void> offerDraw();

  Future<void> respondToDraw({required bool accept});

  Future<void> resign();

  Future<void> reconnect();

  Future<void> disconnect();

  Future<void> dispose();
}
