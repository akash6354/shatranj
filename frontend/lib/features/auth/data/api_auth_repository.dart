import '../../../core/network/api_client.dart';
import '../../../core/network/token_storage.dart';
import '../domain/auth_models.dart';
import '../domain/auth_repository.dart';

class ApiAuthRepository implements AuthRepository {
  ApiAuthRepository(this._client, this._tokenStorage);

  final ApiClient _client;
  final TokenStorage _tokenStorage;

  @override
  Future<AuthSession> signIn({
    required String email,
    required String password,
  }) async {
    final response = await _client.post('/auth/login', body: {
      'email': email.trim(),
      'password': password,
    });
    return _storeSession(response);
  }

  @override
  Future<AuthSession> signUp({
    required String displayName,
    required String email,
    required String password,
  }) async {
    final response = await _client.post('/auth/register', body: {
      'email': email.trim(),
      'password': password,
      'display_name': displayName.trim(),
      'username': _usernameFromDisplayName(displayName),
    });
    return _storeSession(response);
  }

  @override
  Future<AuthSession> continueWithGoogle() => Future.error(
        const AuthException('Google sign-in is not available in the backend yet.'),
      );

  @override
  Future<AuthSession> continueWithPhone(String phoneNumber) => Future.error(
        const AuthException('Phone sign-in is not available in the backend yet.'),
      );

  @override
  Future<void> sendPasswordReset(String email) => Future.error(
        const AuthException('Password reset is not available in the backend yet.'),
      );

  Future<AuthSession?> restoreSession() async {
    final token = await _tokenStorage.readAccessToken();
    if (token == null || token.isEmpty) return null;
    try {
      final response = await _client.get('/auth/validate');
      return AuthSession(accessToken: token, user: _user(response));
    } on ApiException catch (error) {
      if (error.statusCode == 401) {
        await _tokenStorage.clear();
        return null;
      }
      rethrow;
    }
  }

  Future<void> logout() => _tokenStorage.clear();

  Future<AuthSession> _storeSession(Map<String, dynamic> response) async {
    final token = response['access_token']?.toString();
    if (token == null ||
        token.isEmpty ||
        response['user'] is! Map<String, dynamic>) {
      throw const AuthException('The authentication response was incomplete.');
    }
    await _tokenStorage.writeAccessToken(token);
    return AuthSession(accessToken: token, user: _user(response));
  }

  AuthUser _user(Map<String, dynamic> response) {
    final data = response['user'] is Map<String, dynamic>
        ? response['user'] as Map<String, dynamic>
        : response;
    return AuthUser(
      id: data['id']?.toString() ?? '',
      email: data['email']?.toString() ?? '',
      displayName: data['display_name']?.toString() ?? '',
      username: data['username']?.toString() ?? '',
    );
  }

  String _usernameFromDisplayName(String value) {
    final username = value
        .trim()
        .toLowerCase()
        .replaceAll(RegExp(r'[^a-z0-9_]'), '_')
        .replaceAll(RegExp('_+'), '_');
    if (username.length >= 3) {
      return username.length > 24 ? username.substring(0, 24) : username;
    }
    return 'player_${DateTime.now().millisecondsSinceEpoch % 100000}';
  }
}
