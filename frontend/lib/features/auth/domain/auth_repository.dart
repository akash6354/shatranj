import 'auth_models.dart';

abstract interface class AuthRepository {
  Future<AuthSession> signIn({
    required String email,
    required String password,
  });

  Future<AuthSession> signUp({
    required String displayName,
    required String email,
    required String password,
  });

  Future<AuthSession> continueWithGoogle();

  Future<AuthSession> continueWithPhone(String phoneNumber);

  Future<void> sendPasswordReset(String email);
}
