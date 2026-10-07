import 'tournament_models.dart';

abstract interface class TournamentRepository {
  Future<List<Tournament>> listTournaments();
  Future<Tournament> getTournament(String tournamentId);
  Future<void> register(String tournamentId);
  Future<void> withdraw(String tournamentId);
  Future<List<TournamentPlayer>> getStandings(String tournamentId);
}
