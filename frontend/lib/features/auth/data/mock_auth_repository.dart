import 'dart:async';

import '../domain/auth_models.dart';
import '../domain/auth_repository.dart';

/// Development-only repository used until the REST authentication contract is wired.
class MockAuthRepository implements AuthRepository {
  Future<void> _delay() => Future<void>.delayed(const Duration(milliseconds: 450));

  @override
  Future<AuthSession> signIn({
    required String email,
    required String password,
  }) async {
    await _delay();
    if (email.trim().isEmpty || password.isEmpty) {
      throw const AuthException('Enter your email and password to continue.');
    }
    return _session(email: email, displayName: 'Akash');
  }

  @override
  Future<AuthSession> signUp({
    required String displayName,
    required String email,
    required String password,
  }) async {
    await _delay();
    if (displayName.trim().isEmpty || email.trim().isEmpty || password.length < 8) {
      throw const AuthException('Use a name, email, and password with at least 8 characters.');
    }
    return _session(email: email, displayName: displayName);
  }

  @override
  Future<AuthSession> continueWithGoogle() async {
    await _delay();
    return _session(email: 'google-user@example.com', displayName: 'Google Player');
  }

  @override
  Future<AuthSession> continueWithPhone(String phoneNumber) async {
    await _delay();
    if (phoneNumber.trim().isEmpty) {
      throw const AuthException('Enter a phone number to continue.');
    }
    return _session(email: 'phone-user@example.com', displayName: 'Phone Player');
  }

  @override
  Future<void> sendPasswordReset(String email) async {
    await _delay();
    if (email.trim().isEmpty) {
      throw const AuthException('Enter your email to request a reset link.');
    }
  }

  AuthSession _session({
    required String email,
    required String displayName,
  }) {
    return AuthSession(
      accessToken: 'mock-token',
      user: AuthUser(
        id: 'mock-user',
        email: email.trim(),
        displayName: displayName.trim(),
        username: displayName.trim().toLowerCase().replaceAll(' ', '_'),
      ),
    );
  }
}
