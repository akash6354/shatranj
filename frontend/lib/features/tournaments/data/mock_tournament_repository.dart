import '../domain/tournament_models.dart';
import '../domain/tournament_repository.dart';

class MockTournamentRepository implements TournamentRepository {
  static final _players = [
    const TournamentPlayer(userId: 'you', username: 'Akash', rating: 1248, score: 4, rank: 3, games: 3),
    const TournamentPlayer(userId: 'p1', username: 'KnightFox', rating: 1512, score: 6, rank: 1, games: 3),
    const TournamentPlayer(userId: 'p2', username: 'Rohan', rating: 1380, score: 5, rank: 2, games: 3),
    const TournamentPlayer(userId: 'p3', username: 'Mira', rating: 1198, score: 2, rank: 4, games: 3),
  ];

  static final _tournaments = [
    Tournament(
      id: 'arena-blitz',
      name: 'Evening Blitz Arena',
      description: 'Fast games, nonstop action, and a chance to climb the leaderboard.',
      format: 'Arena',
      status: TournamentStatus.live,
      timeControlSeconds: 180,
      incrementSeconds: 2,
      maxPlayers: 64,
      startsAt: DateTime(2026, 10, 7, 20),
      participants: _players,
      rounds: [
        const TournamentRound(number: 3, status: 'in_progress', currentGameId: 'arena-game'),
        const TournamentRound(number: 2, status: 'completed'),
      ],
      isRegistered: true,
    ),
    Tournament(
      id: 'rapid-royale',
      name: 'Rapid Royale',
      description: 'Test your strategic depth in a longer time control.',
      format: 'Swiss',
      status: TournamentStatus.upcoming,
      timeControlSeconds: 600,
      incrementSeconds: 5,
      maxPlayers: 32,
      startsAt: DateTime(2026, 10, 8, 19),
      participants: _players.sublist(0, 2),
      rounds: const [],
    ),
    Tournament(
      id: 'weekend-open',
      name: 'Weekend Open',
      description: 'A community tournament for players of every level.',
      format: 'Swiss',
      status: TournamentStatus.upcoming,
      timeControlSeconds: 300,
      incrementSeconds: 0,
      maxPlayers: 128,
      startsAt: DateTime(2026, 10, 10, 16),
      participants: const [],
      rounds: const [],
    ),
  ];

  @override
  Future<List<Tournament>> listTournaments() async => _tournaments;

  @override
  Future<Tournament> getTournament(String tournamentId) async =>
      _tournaments.firstWhere((item) => item.id == tournamentId);

  @override
  Future<void> register(String tournamentId) async {}

  @override
  Future<void> withdraw(String tournamentId) async {}

  @override
  Future<List<TournamentPlayer>> getStandings(String tournamentId) async => _players;
}
