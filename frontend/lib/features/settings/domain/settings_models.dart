import '../../../core/widgets/chess/chess_models.dart';

enum AppThemeMode { dark, system }

class SettingsState {
  const SettingsState({
    required this.themeMode,
    required this.boardTheme,
    required this.pieceTheme,
    required this.showLegalMoves,
    required this.showMoveSuggestions,
    required this.soundEnabled,
    required this.vibrationEnabled,
    this.isLoading = false,
  });

  factory SettingsState.defaults() => const SettingsState(
        themeMode: AppThemeMode.dark,
        boardTheme: BoardTheme.classic,
        pieceTheme: PieceTheme.unicode,
        showLegalMoves: true,
        showMoveSuggestions: true,
        soundEnabled: true,
        vibrationEnabled: true,
      );

  final AppThemeMode themeMode;
  final BoardTheme boardTheme;
  final PieceTheme pieceTheme;
  final bool showLegalMoves;
  final bool showMoveSuggestions;
  final bool soundEnabled;
  final bool vibrationEnabled;
  final bool isLoading;
}
