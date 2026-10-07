import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/network_providers.dart';
import '../data/api_profile_repository.dart';
import '../domain/profile_models.dart';
import '../domain/profile_repository.dart';

final profileRepositoryProvider = Provider<ProfileRepository>(
  (ref) => ApiProfileRepository(ref.watch(networkApiClientProvider)),
);

final profileProvider = FutureProvider<Profile>(
  (ref) => ref.watch(profileRepositoryProvider).getProfile(),
);
