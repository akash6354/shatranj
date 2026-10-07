import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/app_widgets.dart';
import 'community_controller.dart';

class ClubDetailScreen extends ConsumerWidget {
  const ClubDetailScreen({required this.clubId, super.key});
  final String clubId;
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final club = ref.watch(clubProvider(clubId));
    final members = ref.watch(clubMembersProvider(clubId));
    return Scaffold(
      appBar: AppBar(title: const Text('Club')),
      body: SafeArea(child: club.when(
        loading: () => const AppLoading(),
        error: (error, _) => AppError(message: error.toString()),
        data: (item) => SingleChildScrollView(
          padding: const EdgeInsets.all(AppSpacing.page),
          child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            AppGradientCard(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
              Text(item.name, style: AppTextStyles.headline),
              const SizedBox(height: AppSpacing.sm),
              Text(item.description, style: AppTextStyles.body),
              const SizedBox(height: AppSpacing.sm),
              Text('${item.memberCount} members • ${item.visibility}', style: AppTextStyles.caption),
              const SizedBox(height: AppSpacing.md),
              AppButton(label: item.isMember ? 'Leave club' : 'Join club', icon: item.isMember ? Icons.logout : Icons.login, expand: true,
                variant: item.isMember ? AppButtonVariant.outline : AppButtonVariant.primary,
                onPressed: () async {
                  final repo = ref.read(clubRepositoryProvider);
                  if (item.isMember) {
                    await repo.leave(item.id);
                  } else {
                    await repo.join(item.id);
                  }
                  ref.invalidate(clubProvider(item.id));
                  ref.invalidate(clubsProvider);
                }),
            ])),
            const SizedBox(height: AppSpacing.lg),
            const AppSectionHeader(title: 'Club activity'),
            const SizedBox(height: AppSpacing.sm),
            AppCard(child: Row(children: [const Icon(Icons.bolt_rounded, color: AppColors.gold), const SizedBox(width: AppSpacing.sm), Expanded(child: Text(item.activity, style: AppTextStyles.body))])),
            const SizedBox(height: AppSpacing.lg),
            const AppSectionHeader(title: 'Members'),
            const SizedBox(height: AppSpacing.sm),
            members.when(
              loading: () => const AppLoading(),
              error: (error, _) => AppError(message: error.toString()),
              data: (items) => AppCard(child: Column(children: items.map((member) => ListTile(
                contentPadding: EdgeInsets.zero,
                leading: const CircleAvatar(child: Icon(Icons.person)),
                title: Text(member.username),
                subtitle: Text(member.role),
                trailing: Text('${member.joinedAt.day}/${member.joinedAt.month}', style: AppTextStyles.caption),
              )).toList())),
            ),
          ]),
        ),
      )),
    );
  }
}
