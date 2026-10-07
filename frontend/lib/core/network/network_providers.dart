import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../storage/local_storage.dart';
import 'api_client.dart';
import 'token_storage.dart';

final networkLocalStorageProvider =
    Provider<LocalStorage>((_) => InMemoryLocalStorage());

final networkTokenStorageProvider = Provider<TokenStorage>(
  (_) => SecureTokenStorage(),
);

final networkApiClientProvider = Provider<ApiClient>(
  (ref) => ApiClient(tokenStorage: ref.watch(networkTokenStorageProvider)),
);
