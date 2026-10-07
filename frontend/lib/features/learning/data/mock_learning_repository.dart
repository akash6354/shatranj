import '../domain/learning_models.dart';
import '../domain/learning_repository.dart';

class MockLearningRepository implements LearningRepository {
  static const _fen =
      'r1bqk2r/pppp1ppp/2n2n2/4p3/4P3/2N2N2/PPPP1PPP/R1BQKB1R w KQkq - 2 4';

  static final _courses = [
    Course(
      id: 'opening-fundamentals',
      title: 'Opening Fundamentals',
      description: 'Build strong habits from the very first move.',
      category: 'Openings',
      progressPercent: 42,
      lessons: [
        Lesson(
          id: 'opening-1',
          courseId: 'opening-fundamentals',
          title: 'The principles of development',
          description: 'Learn how to control the centre and develop with purpose.',
          sortOrder: 1,
          content: 'Develop your minor pieces, control the centre, and castle early.',
          progress: LessonProgress(percent: 100, completed: true),
        ),
        Lesson(
          id: 'opening-2',
          courseId: 'opening-fundamentals',
          title: 'King safety and castling',
          description: 'Keep your king safe while connecting your rooks.',
          sortOrder: 2,
          content: 'Castling is both a king-safety move and a development move.',
          progress: LessonProgress(percent: 55, completed: false),
          fen: _fen,
        ),
        Lesson(
          id: 'opening-3',
          courseId: 'opening-fundamentals',
          title: 'Reading opening positions',
          description: 'Turn principles into a reliable decision process.',
          sortOrder: 3,
          content: 'Before every move, check threats, development, and the centre.',
          progress: LessonProgress(percent: 0, completed: false),
          locked: true,
        ),
      ],
    ),
    Course(
      id: 'opening-repertoire',
      title: 'Opening Repertoire',
      description: 'Create a practical set of openings for your games.',
      category: 'Openings',
      progressPercent: 18,
      lessons: const [],
    ),
    Course(
      id: 'middlegame-strategy',
      title: 'Middlegame Strategy',
      description: 'Improve your plans, pawn structures, and piece activity.',
      category: 'Strategy',
      progressPercent: 0,
      lessons: const [],
    ),
    Course(
      id: 'endgame-mastery',
      title: 'Endgame Mastery',
      description: 'Convert advantages and defend difficult positions.',
      category: 'Endgames',
      progressPercent: 0,
      lessons: const [],
    ),
    Course(
      id: 'grandmaster-games',
      title: 'Grandmaster Games',
      description: 'Study the ideas behind memorable games.',
      category: 'Masterclass',
      progressPercent: 0,
      lessons: const [],
    ),
  ];

  @override
  Future<List<Course>> listCourses() async => _courses;

  @override
  Future<Course> getCourse(String courseId) async =>
      _courses.firstWhere((course) => course.id == courseId);

  @override
  Future<Lesson> updateLessonProgress(
    String lessonId, {
    required int progressPercent,
    required bool completed,
    String? lastPosition,
  }) async {
    for (final course in _courses) {
      for (final lesson in course.lessons) {
        if (lesson.id == lessonId) {
          return Lesson(
            id: lesson.id,
            courseId: lesson.courseId,
            title: lesson.title,
            description: lesson.description,
            sortOrder: lesson.sortOrder,
            content: lesson.content,
            progress: LessonProgress(
              percent: progressPercent,
              completed: completed,
              lastPosition: lastPosition,
            ),
            fen: lesson.fen,
            locked: lesson.locked,
          );
        }
      }
    }
    throw StateError('Lesson not found');
  }
}
