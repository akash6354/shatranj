import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/mock_play_repositories.dart';
import '../domain/play_models.dart';
import '../domain/play_repositories.dart';

final matchmakingRepositoryProvider = Provider<MatchmakingRepository>(
  (_) => MockMatchmakingRepository(),
);
final computerGameRepositoryProvider = Provider<ComputerGameRepository>(
  (_) => MockComputerGameRepository(),
);
final friendGameRepositoryProvider = Provider<FriendGameRepository>(
  (_) => MockFriendGameRepository(),
);

final matchmakingProvider = AutoDisposeNotifierProvider<MatchmakingController,
    MatchmakingState>(MatchmakingController.new);

class MatchmakingController extends AutoDisposeNotifier<MatchmakingState> {
  @override
  MatchmakingState build() => const MatchmakingState(
        timeControl: TimeControl(3, 2),
      );

  void setTimeControl(TimeControl value) => state = state.copyWith(timeControl: value);
  void setRatingRange(int value) => state = state.copyWith(ratingRange: value);

  Future<void> search() async {
    state = state.copyWith(status: SearchStatus.searching);
    final result = await ref.read(matchmakingRepositoryProvider).search(
          state.timeControl,
          state.ratingRange,
        );
    state = result;
  }

  Future<void> cancel() async {
    await ref.read(matchmakingRepositoryProvider).cancel();
    state = state.copyWith(status: SearchStatus.cancelled);
  }
}
