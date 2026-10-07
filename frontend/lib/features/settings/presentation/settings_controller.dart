import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/storage/local_storage.dart';
import '../../../core/widgets/chess/chess_models.dart';
import '../data/local_settings_repository.dart';
import '../domain/settings_models.dart';
import '../domain/settings_repository.dart';

final localStorageProvider = Provider<LocalStorage>((_) => InMemoryLocalStorage());
final settingsRepositoryProvider = Provider<SettingsRepository>(
  (ref) => LocalSettingsRepository(ref.watch(localStorageProvider)),
);

final settingsControllerProvider =
    NotifierProvider<SettingsController, AsyncValue<SettingsState>>(
  SettingsController.new,
);

class SettingsController extends Notifier<AsyncValue<SettingsState>> {
  late SettingsRepository _repository;

  @override
  AsyncValue<SettingsState> build() {
    _repository = ref.watch(settingsRepositoryProvider);
    _load();
    return const AsyncLoading();
  }

  Future<void> _load() async {
    state = await AsyncValue.guard(_repository.load);
  }

  Future<void> update(SettingsState Function(SettingsState) change) async {
    final current = state.valueOrNull;
    if (current == null) return;
    final next = change(current);
    state = AsyncData(next);
    await _repository.save(next);
  }

  Future<void> setTheme(AppThemeMode value) => update((current) => SettingsState(
        themeMode: value,
        boardTheme: current.boardTheme,
        pieceTheme: current.pieceTheme,
        showLegalMoves: current.showLegalMoves,
        showMoveSuggestions: current.showMoveSuggestions,
        soundEnabled: current.soundEnabled,
        vibrationEnabled: current.vibrationEnabled,
      ));

  Future<void> setBoard(BoardTheme value) => update((current) => SettingsState(
        themeMode: current.themeMode,
        boardTheme: value,
        pieceTheme: current.pieceTheme,
        showLegalMoves: current.showLegalMoves,
        showMoveSuggestions: current.showMoveSuggestions,
        soundEnabled: current.soundEnabled,
        vibrationEnabled: current.vibrationEnabled,
      ));

  Future<void> setPieces(PieceTheme value) => update((current) => SettingsState(
        themeMode: current.themeMode,
        boardTheme: current.boardTheme,
        pieceTheme: value,
        showLegalMoves: current.showLegalMoves,
        showMoveSuggestions: current.showMoveSuggestions,
        soundEnabled: current.soundEnabled,
        vibrationEnabled: current.vibrationEnabled,
      ));

  Future<void> setLegalMoves(bool value) => update((current) => SettingsState(
        themeMode: current.themeMode,
        boardTheme: current.boardTheme,
        pieceTheme: current.pieceTheme,
        showLegalMoves: value,
        showMoveSuggestions: current.showMoveSuggestions,
        soundEnabled: current.soundEnabled,
        vibrationEnabled: current.vibrationEnabled,
      ));

  Future<void> setSuggestions(bool value) => update((current) => SettingsState(
        themeMode: current.themeMode,
        boardTheme: current.boardTheme,
        pieceTheme: current.pieceTheme,
        showLegalMoves: current.showLegalMoves,
        showMoveSuggestions: value,
        soundEnabled: current.soundEnabled,
        vibrationEnabled: current.vibrationEnabled,
      ));

  Future<void> setSound(bool value) => update((current) => SettingsState(
        themeMode: current.themeMode,
        boardTheme: current.boardTheme,
        pieceTheme: current.pieceTheme,
        showLegalMoves: current.showLegalMoves,
        showMoveSuggestions: current.showMoveSuggestions,
        soundEnabled: value,
        vibrationEnabled: current.vibrationEnabled,
      ));

  Future<void> setVibration(bool value) => update((current) => SettingsState(
        themeMode: current.themeMode,
        boardTheme: current.boardTheme,
        pieceTheme: current.pieceTheme,
        showLegalMoves: current.showLegalMoves,
        showMoveSuggestions: current.showMoveSuggestions,
        soundEnabled: current.soundEnabled,
        vibrationEnabled: value,
      ));
}
