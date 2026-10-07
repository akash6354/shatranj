enum HistoryResult { win, loss, draw, abandoned }

enum PlayedColor { white, black }

class GameHistoryItem {
  const GameHistoryItem({
    required this.id,
    required this.opponent,
    required this.opponentRating,
    required this.result,
    required this.resultLabel,
    required this.color,
    required this.timeControl,
    required this.playedAt,
    required this.ratingChange,
    required this.moveCount,
    required this.duration,
    required this.accuracy,
    required this.endReason,
  });

  final String id;
  final String opponent;
  final int opponentRating;
  final HistoryResult result;
  final String resultLabel;
  final PlayedColor color;
  final String timeControl;
  final DateTime playedAt;
  final int ratingChange;
  final int moveCount;
  final Duration duration;
  final double? accuracy;
  final String endReason;
}

class GameHistoryState {
  const GameHistoryState({
    this.games = const [],
    this.isLoading = false,
    this.error,
  });

  final List<GameHistoryItem> games;
  final bool isLoading;
  final Object? error;

  GameHistoryState copyWith({
    List<GameHistoryItem>? games,
    bool? isLoading,
    Object? error,
  }) =>
      GameHistoryState(
        games: games ?? this.games,
        isLoading: isLoading ?? this.isLoading,
        error: error,
      );
}
