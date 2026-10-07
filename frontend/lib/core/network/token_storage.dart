import 'dart:async';

import '../storage/local_storage.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

abstract interface class TokenStorage {
  Future<String?> readAccessToken();
  Future<void> writeAccessToken(String token);
  Future<void> clear();
}

class LocalTokenStorage implements TokenStorage {
  LocalTokenStorage(this._storage);

  final LocalStorage _storage;
  static const _accessTokenKey = 'auth.access_token';

  @override
  Future<String?> readAccessToken() async => _storage.readString(_accessTokenKey);

  @override
  Future<void> writeAccessToken(String token) =>
      _storage.writeString(_accessTokenKey, token);

  @override
  Future<void> clear() => _storage.remove(_accessTokenKey);
}

class SecureTokenStorage implements TokenStorage {
  SecureTokenStorage([FlutterSecureStorage? storage])
      : _storage = storage ?? const FlutterSecureStorage();

  final FlutterSecureStorage _storage;
  static const _accessTokenKey = 'shatranj.access_token';

  @override
  Future<String?> readAccessToken() async {
    try {
      return await _storage
          .read(key: _accessTokenKey)
          .timeout(const Duration(milliseconds: 250));
    } on TimeoutException {
      return null;
    }
  }

  @override
  Future<void> writeAccessToken(String token) =>
      _storage.write(key: _accessTokenKey, value: token);

  @override
  Future<void> clear() => _storage.delete(key: _accessTokenKey);
}
