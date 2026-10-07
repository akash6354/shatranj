import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/app_widgets.dart';
import '../domain/community_models.dart';
import 'community_controller.dart';

class CoachesScreen extends ConsumerWidget {
  const CoachesScreen({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final coaches = ref.watch(coachesProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Coaches')),
      body: SafeArea(child: coaches.when(
        loading: () => const AppLoading(),
        error: (error, _) => AppError(message: error.toString()),
        data: (items) => ListView(padding: const EdgeInsets.all(AppSpacing.page), children: items.map((coach) => _CoachCard(coach: coach)).toList()),
      )),
    );
  }
}

class _CoachCard extends StatelessWidget {
  const _CoachCard({required this.coach});
  final Coach coach;
  @override
  Widget build(BuildContext context) => AppCard(
        margin: const EdgeInsets.only(bottom: AppSpacing.sm),
        child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
          Row(children: [
            const CircleAvatar(radius: 24, child: Icon(Icons.school_rounded)),
            const SizedBox(width: AppSpacing.md),
            Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
              Text(coach.displayName, style: AppTextStyles.titleSmall),
              Text(coach.title, style: AppTextStyles.caption),
            ])),
            AppPill(label: coach.status == 'active' ? 'Available' : coach.status, color: AppColors.success),
          ]),
          const SizedBox(height: AppSpacing.md),
          Text(coach.bio, style: AppTextStyles.body),
          const SizedBox(height: AppSpacing.sm),
          Wrap(spacing: AppSpacing.xs, children: coach.specialties.map((item) => AppPill(label: item)).toList()),
          const SizedBox(height: AppSpacing.sm),
          Row(children: [
            Text('${coach.currency} ${(coach.ratePaise / 100).toStringAsFixed(0)} / session', style: AppTextStyles.bodyStrong),
            const Spacer(),
            Text(coach.availability, style: AppTextStyles.caption.copyWith(color: AppColors.mint)),
          ]),
        ]),
      );
}
