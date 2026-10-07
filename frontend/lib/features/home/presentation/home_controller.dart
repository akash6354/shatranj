import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/home_repository.dart';
import '../domain/home_models.dart';

final homeRepositoryProvider = Provider<HomeRepository>(
  (ref) => DemoHomeRepository(),
);

final homeSnapshotProvider = FutureProvider<HomeSnapshot>(
  (ref) => ref.watch(homeRepositoryProvider).getSnapshot(),
);

final homeTabProvider = StateProvider<int>((ref) => 0);
