import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/network_providers.dart';
import '../data/api_review_repository.dart';
import '../domain/review_models.dart';
import '../domain/review_repository.dart';

final reviewRepositoryProvider = Provider<ReviewRepository>(
  (ref) => ApiReviewRepository(ref.watch(networkApiClientProvider)),
);

final reviewProvider = FutureProvider.family<GameReview, String>(
  (ref, gameId) => ref.watch(reviewRepositoryProvider).getReview(gameId),
);

final reviewTabProvider = StateProvider<ReviewTab>((_) => ReviewTab.overview);
final selectedReviewMoveProvider = StateProvider<int>((_) => 0);
