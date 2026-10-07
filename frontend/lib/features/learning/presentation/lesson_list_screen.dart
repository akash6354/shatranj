import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/app_widgets.dart';
import '../domain/learning_models.dart';
import 'learning_controller.dart';

class LessonListScreen extends ConsumerWidget {
  const LessonListScreen({required this.courseId, super.key});
  final String courseId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final course = ref.watch(learningCourseProvider(courseId));
    return Scaffold(
      appBar: AppBar(title: const Text('Lessons')),
      body: SafeArea(
        child: course.when(
          loading: () => const AppLoading(),
          error: (error, _) => AppError(title: 'Course unavailable', message: error.toString()),
          data: (item) => _LessonList(course: item),
        ),
      ),
    );
  }
}

class _LessonList extends StatelessWidget {
  const _LessonList({required this.course});
  final Course course;

  @override
  Widget build(BuildContext context) => SingleChildScrollView(
        padding: const EdgeInsets.all(AppSpacing.page),
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 720),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(course.title, style: AppTextStyles.headline),
                const SizedBox(height: AppSpacing.xs),
                Text(course.description, style: AppTextStyles.body),
                const SizedBox(height: AppSpacing.md),
                LinearProgressIndicator(value: course.progressPercent / 100),
                const SizedBox(height: AppSpacing.lg),
                ...course.lessons.map((lesson) => _LessonCard(lesson: lesson)),
                if (course.lessons.isEmpty)
                  const AppEmptyState(
                    title: 'Lessons are coming soon',
                    message: 'This learning path will be available when its content is published.',
                    icon: Icons.menu_book_outlined,
                  ),
              ],
            ),
          ),
        ),
      );
}

class _LessonCard extends StatelessWidget {
  const _LessonCard({required this.lesson});
  final Lesson lesson;

  @override
  Widget build(BuildContext context) => AppCard(
        margin: const EdgeInsets.only(bottom: AppSpacing.sm),
        color: lesson.locked ? AppColors.surfaceMuted : null,
        onTap: lesson.locked ? null : () => context.push('/learn/${lesson.courseId}/lessons/${lesson.id}'),
        child: Row(
          children: [
            Icon(
              lesson.locked
                  ? Icons.lock_outline_rounded
                  : lesson.progress.completed
                      ? Icons.check_circle_rounded
                      : Icons.play_circle_outline_rounded,
              color: lesson.locked
                  ? AppColors.muted
                  : lesson.progress.completed
                      ? AppColors.success
                      : AppColors.gold,
              size: 30,
            ),
            const SizedBox(width: AppSpacing.md),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(lesson.title, style: AppTextStyles.bodyStrong),
                  const SizedBox(height: AppSpacing.xs),
                  Text(lesson.description, style: AppTextStyles.caption),
                  const SizedBox(height: AppSpacing.sm),
                  LinearProgressIndicator(value: lesson.progress.percent / 100),
                ],
              ),
            ),
            const SizedBox(width: AppSpacing.md),
            Text('${lesson.progress.percent}%', style: AppTextStyles.caption),
          ],
        ),
      );
}
