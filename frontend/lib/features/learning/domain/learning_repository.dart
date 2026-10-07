import 'learning_models.dart';

abstract interface class LearningRepository {
  Future<List<Course>> listCourses();
  Future<Course> getCourse(String courseId);
  Future<Lesson> updateLessonProgress(
    String lessonId, {
    required int progressPercent,
    required bool completed,
    String? lastPosition,
  });
}
