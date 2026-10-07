import '../../../core/storage/local_storage.dart';
import '../../../core/widgets/chess/chess_models.dart';
import '../domain/settings_models.dart';
import '../domain/settings_repository.dart';

class LocalSettingsRepository implements SettingsRepository {
  LocalSettingsRepository(this._storage);

  final LocalStorage _storage;

  @override
  Future<SettingsState> load() async {
    final defaults = SettingsState.defaults();
    return SettingsState(
      themeMode: _theme(_storage.readString('settings.theme')) ?? defaults.themeMode,
      boardTheme: _board(_storage.readString('settings.board')) ?? defaults.boardTheme,
      pieceTheme: _pieces(_storage.readString('settings.pieces')) ?? defaults.pieceTheme,
      showLegalMoves: _storage.readBool('settings.legal_moves') ?? defaults.showLegalMoves,
      showMoveSuggestions:
          _storage.readBool('settings.suggestions') ?? defaults.showMoveSuggestions,
      soundEnabled: _storage.readBool('settings.sound') ?? defaults.soundEnabled,
      vibrationEnabled:
          _storage.readBool('settings.vibration') ?? defaults.vibrationEnabled,
    );
  }

  @override
  Future<void> save(SettingsState settings) async {
    await _storage.writeString('settings.theme', settings.themeMode.name);
    await _storage.writeString('settings.board', settings.boardTheme.name);
    await _storage.writeString('settings.pieces', settings.pieceTheme.name);
    await _storage.writeBool('settings.legal_moves', settings.showLegalMoves);
    await _storage.writeBool('settings.suggestions', settings.showMoveSuggestions);
    await _storage.writeBool('settings.sound', settings.soundEnabled);
    await _storage.writeBool('settings.vibration', settings.vibrationEnabled);
  }

  AppThemeMode? _theme(String? value) => AppThemeMode.values.where((item) => item.name == value).firstOrNull;
  BoardTheme? _board(String? value) => BoardTheme.values.where((item) => item.name == value).firstOrNull;
  PieceTheme? _pieces(String? value) => PieceTheme.values.where((item) => item.name == value).firstOrNull;
}
