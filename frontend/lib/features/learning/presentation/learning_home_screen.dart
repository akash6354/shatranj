import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/app_widgets.dart';
import '../domain/learning_models.dart';
import 'learning_controller.dart';

class LearningHomeScreen extends ConsumerWidget {
  const LearningHomeScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final courses = ref.watch(learningCoursesProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Learn')),
      body: SafeArea(
        child: courses.when(
          loading: () => const AppLoading(message: 'Loading lessons'),
          error: (error, _) => AppError(title: 'Learning unavailable', message: error.toString()),
          data: (items) => _LearningContent(courses: items),
        ),
      ),
    );
  }
}

class _LearningContent extends StatelessWidget {
  const _LearningContent({required this.courses});
  final List<Course> courses;

  @override
  Widget build(BuildContext context) {
    final completed = courses.where((course) => course.progressPercent > 0).length;
    return SingleChildScrollView(
      padding: const EdgeInsets.fromLTRB(AppSpacing.page, AppSpacing.sm, AppSpacing.page, AppSpacing.xl),
      child: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 920),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              AppGradientCard(
                child: Row(
                  children: [
                    const Icon(Icons.school_rounded, color: AppColors.gold, size: 48),
                    const SizedBox(width: AppSpacing.md),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          const Text('Grow your chess understanding', style: AppTextStyles.headline),
                          const SizedBox(height: AppSpacing.xs),
                          Text('$completed of ${courses.length} learning paths started', style: AppTextStyles.body),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: AppSpacing.lg),
              const AppSectionHeader(title: 'Learning paths', subtitle: 'Structured lessons for every stage'),
              const SizedBox(height: AppSpacing.sm),
              ...courses.map((course) => _CourseCard(course: course)),
            ],
          ),
        ),
      ),
    );
  }
}

class _CourseCard extends StatelessWidget {
  const _CourseCard({required this.course});
  final Course course;

  @override
  Widget build(BuildContext context) => AppCard(
        margin: const EdgeInsets.only(bottom: AppSpacing.sm),
        onTap: () => context.push('/learn/${course.id}'),
        child: Row(
          children: [
            Container(
              width: 48,
              height: 48,
              decoration: BoxDecoration(
                color: AppColors.mint.withValues(alpha: .13),
                borderRadius: AppRadius.buttonRadius,
              ),
              child: Icon(_iconFor(course.category), color: AppColors.mint),
            ),
            const SizedBox(width: AppSpacing.md),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(course.title, style: AppTextStyles.titleSmall),
                  const SizedBox(height: AppSpacing.xs),
                  Text(course.description, style: AppTextStyles.caption),
                  const SizedBox(height: AppSpacing.sm),
                  LinearProgressIndicator(value: course.progressPercent / 100),
                ],
              ),
            ),
            const SizedBox(width: AppSpacing.md),
            Text('${course.progressPercent}%', style: AppTextStyles.bodyStrong.copyWith(color: AppColors.gold)),
            const Icon(Icons.chevron_right_rounded, color: AppColors.muted),
          ],
        ),
      );
}

IconData _iconFor(String category) => switch (category) {
      'Openings' => Icons.account_tree_outlined,
      'Strategy' => Icons.psychology_outlined,
      'Endgames' => Icons.flag_outlined,
      _ => Icons.auto_awesome_outlined,
    };
