import '../../../core/network/api_client.dart';
import '../domain/profile_models.dart';
import '../domain/profile_repository.dart';

class ApiProfileRepository implements ProfileRepository {
  ApiProfileRepository(this._client);

  final ApiClient _client;

  @override
  Future<Profile> getProfile() async {
    final profile = await _client.get('/profiles/me');
    final ratings = await Future.wait([
      _client.get('/ratings/blitz'),
      _client.get('/ratings/rapid'),
      _client.get('/ratings/puzzle'),
    ]);
    final blitz = _rating(ratings[0]);
    final rapid = _rating(ratings[1]);
    final puzzle = _rating(ratings[2]);
    final games = blitz['games'] as int;
    final wins = blitz['wins'] as int;
    return Profile(
      username: _string(profile, 'username'),
      displayName: _string(profile, 'display_name'),
      avatarUrl: _string(profile, 'avatar_url'),
      rating: blitz['rating'] as int,
      blitzRating: blitz['rating'] as int,
      rapidRating: rapid['rating'] as int,
      puzzleRating: puzzle['rating'] as int,
      games: games,
      winRate: games == 0 ? 0 : wins * 100 / games,
      streak: 0,
      achievements: const [],
      friends: 0,
    );
  }

  Map<String, Object> _rating(Map<String, dynamic> response) {
    return {
      'rating': _int(response, 'rating'),
      'games': _int(response, 'games'),
      'wins': _int(response, 'wins'),
    };
  }

  String _string(Map<String, dynamic> value, String key) =>
      value[key]?.toString() ?? '';

  int _int(Map<String, dynamic> value, String key) =>
      (value[key] as num?)?.toInt() ?? 0;
}
