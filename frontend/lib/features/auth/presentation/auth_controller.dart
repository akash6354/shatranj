import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/network_providers.dart';
import '../data/api_auth_repository.dart';
import '../domain/auth_models.dart';

final authTokenStorageProvider = networkTokenStorageProvider;
final apiClientProvider = networkApiClientProvider;
final authRepositoryProvider = Provider<ApiAuthRepository>(
  (ref) => ApiAuthRepository(
    ref.watch(apiClientProvider),
    ref.watch(authTokenStorageProvider),
  ),
);

final authControllerProvider =
    NotifierProvider<AuthController, AsyncValue<AuthSession?>>(
  AuthController.new,
);

class AuthController extends Notifier<AsyncValue<AuthSession?>> {
  late final ApiAuthRepository _repository;

  @override
  AsyncValue<AuthSession?> build() {
    _repository = ref.watch(authRepositoryProvider);
    _restore();
    return const AsyncLoading();
  }

  Future<void> signIn({
    required String email,
    required String password,
  }) async {
    await _run(() => _repository.signIn(email: email, password: password));
  }

  Future<void> signUp({
    required String displayName,
    required String email,
    required String password,
  }) async {
    await _run(() => _repository.signUp(
          displayName: displayName,
          email: email,
          password: password,
        ));
  }

  Future<void> continueWithGoogle() async {
    await _run(_repository.continueWithGoogle);
  }

  Future<void> continueWithPhone(String phoneNumber) async {
    await _run(() => _repository.continueWithPhone(phoneNumber));
  }

  Future<void> sendPasswordReset(String email) async {
    await _runWithoutSession(() => _repository.sendPasswordReset(email));
  }

  void clearError() {
    if (state.hasError) state = const AsyncData(null);
  }

  Future<void> logout() async {
    await _repository.logout();
    state = const AsyncData(null);
  }

  Future<void> _restore() async {
    state = await AsyncValue.guard(_repository.restoreSession);
  }

  Future<void> _run(Future<AuthSession> Function() operation) async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(operation);
  }

  Future<void> _runWithoutSession(Future<void> Function() operation) async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(() async {
      await operation();
      return null;
    });
  }
}
