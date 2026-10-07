import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/app_widgets.dart';
import '../domain/community_models.dart';
import 'community_controller.dart';

class ClubsScreen extends ConsumerStatefulWidget {
  const ClubsScreen({super.key});
  @override
  ConsumerState<ClubsScreen> createState() => _ClubsScreenState();
}

class _ClubsScreenState extends ConsumerState<ClubsScreen> {
  String _query = '';

  @override
  Widget build(BuildContext context) {
    final clubs = ref.watch(clubsProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Clubs')),
      body: SafeArea(child: clubs.when(
        loading: () => const AppLoading(),
        error: (error, _) => AppError(message: error.toString()),
        data: (items) => ListView(
          padding: const EdgeInsets.all(AppSpacing.page),
          children: [
            TextField(
              decoration: const InputDecoration(prefixIcon: Icon(Icons.search_rounded), hintText: 'Search clubs'),
              onChanged: (value) => setState(() => _query = value),
            ),
            const SizedBox(height: AppSpacing.md),
            ...items.where((club) => club.name.toLowerCase().contains(_query.toLowerCase())).map((club) => _ClubCard(club: club)),
          ],
        ),
      )),
    );
  }
}

class _ClubCard extends StatelessWidget {
  const _ClubCard({required this.club});
  final Club club;
  @override
  Widget build(BuildContext context) => AppCard(
        margin: const EdgeInsets.only(bottom: AppSpacing.sm),
        onTap: () => context.push('/community/clubs/${club.id}'),
        child: Row(children: [
          const CircleAvatar(child: Icon(Icons.groups_rounded)),
          const SizedBox(width: AppSpacing.md),
          Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            Text(club.name, style: AppTextStyles.titleSmall),
            const SizedBox(height: AppSpacing.xs),
            Text(club.description, style: AppTextStyles.caption, maxLines: 2, overflow: TextOverflow.ellipsis),
            const SizedBox(height: AppSpacing.sm),
            Text('${club.memberCount} members • ${club.visibility}', style: AppTextStyles.caption),
          ])),
          const Icon(Icons.chevron_right_rounded, color: AppColors.muted),
        ]),
      );
}
