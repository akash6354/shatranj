import '../domain/play_models.dart';
import '../domain/play_repositories.dart';

class MockMatchmakingRepository implements MatchmakingRepository {
  @override
  Future<MatchmakingState> search(TimeControl timeControl, int ratingRange) async {
    await Future<void>.delayed(const Duration(seconds: 2));
    return MatchmakingState(
      timeControl: timeControl,
      status: SearchStatus.found,
      ratingRange: ratingRange,
      opponentName: 'Rohan',
      opponentRating: 1223,
    );
  }

  @override
  Future<void> cancel() async {}
}

class MockComputerGameRepository implements ComputerGameRepository {
  @override
  Future<String> createGame({
    required ComputerDifficulty difficulty,
    required PlayerColor color,
    required TimeControl timeControl,
  }) async =>
      'computer-${difficulty.name}-${timeControl.label}';
}

class MockFriendGameRepository implements FriendGameRepository {
  @override
  Future<String> createRoom(TimeControl timeControl) async => 'SHAT-${timeControl.label.replaceAll('+', '')}';

  @override
  Future<String> joinRoom(String code) async => code.trim().toUpperCase();
}
