class AuthUser {
  const AuthUser({
    required this.id,
    required this.email,
    required this.displayName,
    required this.username,
  });

  final String id;
  final String email;
  final String displayName;
  final String username;
}

class AuthSession {
  const AuthSession({
    required this.accessToken,
    required this.user,
  });

  final String accessToken;
  final AuthUser user;
}

class AuthException implements Exception {
  const AuthException(this.message);

  final String message;

  @override
  String toString() => message;
}
