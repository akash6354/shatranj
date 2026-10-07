import 'dart:convert';

import '../../../core/network/api_client.dart';
import '../domain/learning_models.dart';
import '../domain/learning_repository.dart';

class ApiLearningRepository implements LearningRepository {
  ApiLearningRepository(this._client);

  final ApiClient _client;

  @override
  Future<List<Course>> listCourses() async {
    final response = await _client.get('/lessons');
    final items = response['data'];
    if (items is! List) return const [];
    return items
        .whereType<Map>()
        .map((item) => _parseCourse(Map<String, dynamic>.from(item)))
        .toList();
  }

  @override
  Future<Course> getCourse(String courseId) async {
    final response = await _client.get('/lessons/$courseId');
    return _parseCourse(response);
  }

  @override
  Future<Lesson> updateLessonProgress(
    String lessonId, {
    required int progressPercent,
    required bool completed,
    String? lastPosition,
  }) async {
    final body = <String, dynamic>{
      'progress_percent': progressPercent,
      'completed': completed,
    };
    if (lastPosition != null && lastPosition.trim().isNotEmpty) {
      final decoded = jsonDecode(lastPosition);
      if (decoded is Map<String, dynamic>) {
        body['last_position'] = decoded;
      }
    }
    final response = await _client.put('/lessons/chapters/$lessonId/progress',
        body: body);
    return _parseLesson(response);
  }

  Course _parseCourse(Map<String, dynamic> data) {
    final chapters = data['chapters'];
    final lessons = chapters is List
        ? chapters
            .whereType<Map>()
            .map((item) => _parseLesson(Map<String, dynamic>.from(item)))
            .toList()
        : <Lesson>[];
    return Course(
      id: _string(data, 'id'),
      title: _string(data, 'title'),
      description: _string(data, 'description'),
      category: _string(data, 'category'),
      progressPercent: _int(data, 'progress_percent'),
      lessons: lessons,
    );
  }

  Lesson _parseLesson(Map<String, dynamic> data) {
    final rawContent = data['content'];
    return Lesson(
      id: _string(data, 'id'),
      courseId: _string(data, 'course_id'),
      title: _string(data, 'title'),
      description: _string(data, 'description'),
      sortOrder: _int(data, 'sort_order'),
      content: rawContent is String ? rawContent : jsonEncode(rawContent ?? {}),
      progress: LessonProgress(
        percent: _int(data, 'progress_percent'),
        completed: data['completed'] == true,
        lastPosition: _optionalJsonString(data['last_position']),
      ),
    );
  }

  String _string(Map<String, dynamic> value, String key) =>
      value[key]?.toString() ?? '';

  int _int(Map<String, dynamic> value, String key) =>
      (value[key] as num?)?.toInt() ?? 0;

  String? _optionalJsonString(Object? value) {
    if (value == null) return null;
    return value is String ? value : jsonEncode(value);
  }
}
