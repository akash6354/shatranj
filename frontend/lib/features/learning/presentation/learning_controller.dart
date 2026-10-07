import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/network_providers.dart';
import '../data/api_learning_repository.dart';
import '../domain/learning_models.dart';
import '../domain/learning_repository.dart';

final learningRepositoryProvider = Provider<LearningRepository>(
  (ref) => ApiLearningRepository(ref.watch(networkApiClientProvider)),
);

final learningCoursesProvider = FutureProvider<List<Course>>(
  (ref) => ref.watch(learningRepositoryProvider).listCourses(),
);

final learningCourseProvider = FutureProvider.family<Course, String>((
  ref,
  courseId,
) {
  return ref.watch(learningRepositoryProvider).getCourse(courseId);
});
