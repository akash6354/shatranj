import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/network_providers.dart';
import '../data/api_game_repository.dart';
import '../domain/game_models.dart';
import '../domain/game_repository.dart';
import '../domain/game_socket.dart';

final gameRepositoryProvider = Provider<GameRepository>(
  (ref) => ApiGameRepository(
    ref.watch(networkApiClientProvider),
    ref.watch(networkTokenStorageProvider),
  ),
);

final gameControllerProvider =
    AutoDisposeNotifierProviderFamily<
      GameController,
      AsyncValue<GameState>,
      String
    >(GameController.new);

class GameController
    extends AutoDisposeFamilyNotifier<AsyncValue<GameState>, String> {
  late final GameRepository _repository;
  late final GameSocket _socket;
  StreamSubscription<GameState>? _subscription;

  @override
  AsyncValue<GameState> build(String gameId) {
    _repository = ref.watch(gameRepositoryProvider);
    _socket = _repository.createSocket();
    ref.onDispose(() async {
      await _subscription?.cancel();
      await _socket.dispose();
    });
    _load(gameId);
    return const AsyncLoading();
  }

  Future<void> _load(String gameId) async {
    try {
      final initial = await _repository.getGame(gameId);
      state = AsyncData(initial);
      _subscription = _socket.events.listen(
        (event) => state = AsyncData(event),
        onError: (Object error, StackTrace stackTrace) {
          state = AsyncError(error, stackTrace);
        },
      );
      await _socket.connect(gameId);
    } catch (error, stackTrace) {
      state = AsyncError(error, stackTrace);
    }
  }

  Future<void> sendMove(String from, String to) async {
    await _runAction(() => _socket.sendMove('$from$to'));
  }

  Future<void> offerDraw() async {
    await _runAction(_socket.offerDraw);
  }

  Future<void> respondToDraw({required bool accept}) async {
    await _runAction(() => _socket.respondToDraw(accept: accept));
  }

  Future<void> resign() async {
    await _runAction(_socket.resign);
  }

  Future<void> reconnect() async {
    final current = state.valueOrNull;
    if (current != null) {
      state = AsyncData(
        current.copyWith(connection: GameConnectionStatus.reconnecting),
      );
    }
    await _runAction(_socket.reconnect);
  }

  Future<void> _runAction(Future<void> Function() action) async {
    try {
      await action();
    } catch (error, stackTrace) {
      state = AsyncError<GameState>(error, stackTrace);
    }
  }
}
