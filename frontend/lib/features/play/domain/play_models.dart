enum PlayMode { quickMatch, friend, computer }

class TimeControl {
  const TimeControl(this.minutes, this.increment);

  final int minutes;
  final int increment;

  String get label => '$minutes+$increment';
  Duration get initialTime => Duration(minutes: minutes);
}

enum SearchStatus { idle, searching, found, cancelled }

class MatchmakingState {
  const MatchmakingState({
    required this.timeControl,
    this.status = SearchStatus.idle,
    this.ratingRange = 200,
    this.opponentName,
    this.opponentRating,
  });

  final TimeControl timeControl;
  final SearchStatus status;
  final int ratingRange;
  final String? opponentName;
  final int? opponentRating;

  MatchmakingState copyWith({
    TimeControl? timeControl,
    SearchStatus? status,
    int? ratingRange,
    String? opponentName,
    int? opponentRating,
  }) =>
      MatchmakingState(
        timeControl: timeControl ?? this.timeControl,
        status: status ?? this.status,
        ratingRange: ratingRange ?? this.ratingRange,
        opponentName: opponentName ?? this.opponentName,
        opponentRating: opponentRating ?? this.opponentRating,
      );
}

enum ComputerDifficulty { easy, medium, hard, grandmaster }

enum PlayerColor { white, black, random }
