import 'play_models.dart';

abstract interface class MatchmakingRepository {
  Future<MatchmakingState> search(TimeControl timeControl, int ratingRange);
  Future<void> cancel();
}

abstract interface class ComputerGameRepository {
  Future<String> createGame({
    required ComputerDifficulty difficulty,
    required PlayerColor color,
    required TimeControl timeControl,
  });
}

abstract interface class FriendGameRepository {
  Future<String> createRoom(TimeControl timeControl);
  Future<String> joinRoom(String code);
}
