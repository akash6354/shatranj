class LessonProgress {
  const LessonProgress({
    required this.percent,
    required this.completed,
    this.lastPosition,
  });

  final int percent;
  final bool completed;
  final String? lastPosition;
}

class Lesson {
  const Lesson({
    required this.id,
    required this.courseId,
    required this.title,
    required this.description,
    required this.sortOrder,
    required this.content,
    required this.progress,
    this.fen,
    this.locked = false,
  });

  final String id;
  final String courseId;
  final String title;
  final String description;
  final int sortOrder;
  final String content;
  final LessonProgress progress;
  final String? fen;
  final bool locked;
}

class Course {
  const Course({
    required this.id,
    required this.title,
    required this.description,
    required this.category,
    required this.progressPercent,
    required this.lessons,
  });

  final String id;
  final String title;
  final String description;
  final String category;
  final int progressPercent;
  final List<Lesson> lessons;
}

class LearningState {
  const LearningState({
    required this.courses,
    this.selectedCourse,
    this.error,
  });

  final List<Course> courses;
  final Course? selectedCourse;
  final String? error;
}
