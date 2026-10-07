enum TournamentStatus { upcoming, live, completed, cancelled }

class TournamentPlayer {
  const TournamentPlayer({
    required this.userId,
    required this.username,
    required this.rating,
    required this.score,
    this.rank,
    this.games = 0,
    this.status = 'active',
  });

  final String userId;
  final String username;
  final int rating;
  final int score;
  final int? rank;
  final int games;
  final String status;
}

class TournamentRound {
  const TournamentRound({
    required this.number,
    required this.status,
    this.currentGameId,
  });

  final int number;
  final String status;
  final String? currentGameId;
}

class Tournament {
  const Tournament({
    required this.id,
    required this.name,
    required this.description,
    required this.format,
    required this.status,
    required this.timeControlSeconds,
    required this.incrementSeconds,
    required this.maxPlayers,
    required this.startsAt,
    required this.participants,
    required this.rounds,
    this.isRegistered = false,
  });

  final String id;
  final String name;
  final String description;
  final String format;
  final TournamentStatus status;
  final int timeControlSeconds;
  final int incrementSeconds;
  final int maxPlayers;
  final DateTime startsAt;
  final List<TournamentPlayer> participants;
  final List<TournamentRound> rounds;
  final bool isRegistered;

  String get timeControl =>
      '${timeControlSeconds ~/ 60}+$incrementSeconds';
}

class TournamentState {
  const TournamentState({
    this.tournaments = const [],
    this.selectedTab = 0,
    this.isLoading = false,
    this.error,
  });

  final List<Tournament> tournaments;
  final int selectedTab;
  final bool isLoading;
  final Object? error;
}
