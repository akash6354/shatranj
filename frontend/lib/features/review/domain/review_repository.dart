import 'review_models.dart';

abstract interface class ReviewRepository {
  Future<GameReview> getReview(String gameId);
}
