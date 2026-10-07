import '../../../core/network/api_client.dart';
import '../domain/tournament_models.dart';
import '../domain/tournament_repository.dart';

class ApiTournamentRepository implements TournamentRepository {
  ApiTournamentRepository(this._client);
  final ApiClient _client;

  @override
  Future<List<Tournament>> listTournaments() async {
    final response = await _client.get('/tournaments');
    return _list(response).map(_parseTournament).toList();
  }

  @override
  Future<Tournament> getTournament(String tournamentId) async =>
      _parseTournament(await _client.get('/tournaments/$tournamentId'));

  @override
  Future<void> register(String tournamentId) async {
    await _client.post('/tournaments/$tournamentId/registration');
  }

  @override
  Future<void> withdraw(String tournamentId) async {
    await _client.delete('/tournaments/$tournamentId/registration');
  }

  @override
  Future<List<TournamentPlayer>> getStandings(String tournamentId) async {
    final response = await _client.get('/tournaments/$tournamentId/standings');
    return _list(response).map(_parsePlayer).toList();
  }

  List<Map<String, dynamic>> _list(Map<String, dynamic> response) {
    final value = response['data'];
    return value is List
        ? value.whereType<Map>().map(Map<String, dynamic>.from).toList()
        : const [];
  }

  Tournament _parseTournament(Map<String, dynamic> data) {
    final participants = data['participants'];
    final rounds = data['rounds'];
    return Tournament(
      id: _string(data, 'id'),
      name: _string(data, 'name'),
      description: _string(data, 'description'),
      format: _string(data, 'format'),
      status: TournamentStatus.values.firstWhere(
        (item) => item.name == _string(data, 'status'),
        orElse: () => TournamentStatus.upcoming,
      ),
      timeControlSeconds: _int(data, 'time_control_seconds'),
      incrementSeconds: _int(data, 'increment_seconds'),
      maxPlayers: _int(data, 'max_players'),
      startsAt: _date(data['starts_at']),
      participants: participants is List
          ? participants
              .whereType<Map>()
              .map((item) => _parsePlayer(Map<String, dynamic>.from(item)))
              .toList()
          : const [],
      rounds: rounds is List
          ? rounds
              .whereType<Map>()
              .map((item) => TournamentRound(
                    number: _int(Map<String, dynamic>.from(item), 'number'),
                    status: _string(Map<String, dynamic>.from(item), 'status'),
                    currentGameId: null,
                  ))
              .toList()
          : const [],
    );
  }

  TournamentPlayer _parsePlayer(Map<String, dynamic> data) => TournamentPlayer(
        userId: _string(data, 'user_id'),
        username: _string(data, 'username'),
        rating: _int(data, 'rating'),
        score: _int(data, 'score'),
        rank: data['rank'] is num ? (data['rank'] as num).toInt() : null,
      );

  String _string(Map<String, dynamic> data, String key) =>
      data[key]?.toString() ?? '';

  int _int(Map<String, dynamic> data, String key) =>
      (data[key] as num?)?.toInt() ?? 0;

  DateTime _date(Object? value) =>
      DateTime.tryParse(value?.toString() ?? '') ?? DateTime.now();
}
