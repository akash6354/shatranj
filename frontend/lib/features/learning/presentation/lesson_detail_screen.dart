import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/app_widgets.dart';
import '../../../core/widgets/chess/chess_widgets.dart';
import '../domain/learning_models.dart';
import 'learning_controller.dart';

class LessonDetailScreen extends ConsumerStatefulWidget {
  const LessonDetailScreen({required this.courseId, required this.lessonId, super.key});
  final String courseId;
  final String lessonId;

  @override
  ConsumerState<LessonDetailScreen> createState() => _LessonDetailScreenState();
}

class _LessonDetailScreenState extends ConsumerState<LessonDetailScreen> {
  @override
  Widget build(BuildContext context) {
    final course = ref.watch(learningCourseProvider(widget.courseId));
    return Scaffold(
      appBar: AppBar(title: const Text('Lesson')),
      body: SafeArea(
        child: course.when(
          loading: () => const AppLoading(),
          error: (error, _) => AppError(title: 'Lesson unavailable', message: error.toString()),
          data: (item) {
            final lesson = item.lessons.firstWhere((entry) => entry.id == widget.lessonId);
            final index = item.lessons.indexOf(lesson);
            final next = index + 1 < item.lessons.length ? item.lessons[index + 1] : null;
            final previous = index > 0 ? item.lessons[index - 1] : null;
            return _DetailContent(
              lesson: lesson,
              previous: previous,
              next: next,
              onOpen: (target) => context.go('/learn/${widget.courseId}/lessons/${target.id}'),
              onComplete: () => _complete(lesson),
            );
          },
        ),
      ),
    );
  }

  Future<void> _complete(Lesson lesson) async {
    await ref.read(learningRepositoryProvider).updateLessonProgress(
          lesson.id,
          progressPercent: 100,
          completed: true,
        );
    if (mounted) context.pop();
  }
}

class _DetailContent extends StatelessWidget {
  const _DetailContent({
    required this.lesson,
    required this.previous,
    required this.next,
    required this.onOpen,
    required this.onComplete,
  });

  final Lesson lesson;
  final Lesson? previous;
  final Lesson? next;
  final ValueChanged<Lesson> onOpen;
  final VoidCallback onComplete;

  @override
  Widget build(BuildContext context) => SingleChildScrollView(
        padding: const EdgeInsets.fromLTRB(AppSpacing.page, AppSpacing.sm, AppSpacing.page, AppSpacing.xl),
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 760),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(lesson.title, style: AppTextStyles.headline),
                const SizedBox(height: AppSpacing.xs),
                Text(lesson.description, style: AppTextStyles.body),
                const SizedBox(height: AppSpacing.md),
                LinearProgressIndicator(value: lesson.progress.percent / 100),
                if (lesson.fen != null) ...[
                  const SizedBox(height: AppSpacing.lg),
                  ChessBoard(fen: lesson.fen!, showCoordinates: true),
                ],
                const SizedBox(height: AppSpacing.lg),
                AppCard(
                  child: Text(lesson.content, style: AppTextStyles.body.copyWith(height: 1.55)),
                ),
                const SizedBox(height: AppSpacing.lg),
                AppButton(
                  label: lesson.progress.completed ? 'Completed' : 'Mark lesson complete',
                  icon: Icons.check_rounded,
                  expand: true,
                  onPressed: lesson.progress.completed ? null : onComplete,
                ),
                const SizedBox(height: AppSpacing.md),
                Row(
                  children: [
                    Expanded(
                      child: AppButton(
                        label: 'Previous',
                        icon: Icons.arrow_back_rounded,
                        variant: AppButtonVariant.outline,
                        onPressed: previous == null ? null : () => onOpen(previous!),
                      ),
                    ),
                    const SizedBox(width: AppSpacing.sm),
                    Expanded(
                      child: AppButton(
                        label: 'Next',
                        icon: Icons.arrow_forward_rounded,
                        variant: AppButtonVariant.secondary,
                        onPressed: next == null ? null : () => onOpen(next!),
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ),
      );
}
