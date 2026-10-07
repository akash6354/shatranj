import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/app_widgets.dart';
import '../domain/community_models.dart';
import 'community_controller.dart';

class FriendsScreen extends ConsumerWidget {
  const FriendsScreen({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final friends = ref.watch(friendsProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Friends')),
      body: SafeArea(child: friends.when(
        loading: () => const AppLoading(),
        error: (error, _) => AppError(message: error.toString()),
        data: (items) => ListView(
          padding: const EdgeInsets.all(AppSpacing.page),
          children: [
            AppButton(label: 'Find friends', icon: Icons.person_add_alt_1, variant: AppButtonVariant.outline, expand: true, onPressed: () {}),
            const SizedBox(height: AppSpacing.md),
            ...items.map((friend) => _FriendTile(friend: friend)),
          ],
        ),
      )),
    );
  }
}

class _FriendTile extends ConsumerWidget {
  const _FriendTile({required this.friend});
  final Friend friend;
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final repo = ref.read(friendRepositoryProvider);
    return AppCard(
      margin: const EdgeInsets.only(bottom: AppSpacing.sm),
      child: Row(children: [
        Stack(children: [
          const CircleAvatar(child: Icon(Icons.person)),
          Positioned(right: 0, bottom: 0, child: Container(width: 10, height: 10, decoration: BoxDecoration(color: friend.isOnline ? AppColors.success : AppColors.muted, shape: BoxShape.circle, border: Border.all(color: AppColors.surface, width: 2)))),
        ]),
        const SizedBox(width: AppSpacing.md),
        Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
          Text(friend.username, style: AppTextStyles.bodyStrong),
          Text(friend.requestPending ? 'Friend request' : friend.status, style: AppTextStyles.caption),
        ])),
        if (friend.requestPending) ...[
          AppIconButton(icon: Icons.check_rounded, tooltip: 'Accept request', onPressed: () async { await repo.accept(friend.id); ref.invalidate(friendsProvider); }),
          AppIconButton(icon: Icons.close_rounded, tooltip: 'Reject request', onPressed: () async { await repo.reject(friend.id); ref.invalidate(friendsProvider); }),
        ] else
          AppIconButton(icon: Icons.person_remove_outlined, tooltip: 'Remove friend', onPressed: () async { await repo.remove(friend.id); ref.invalidate(friendsProvider); }),
      ]),
    );
  }
}
