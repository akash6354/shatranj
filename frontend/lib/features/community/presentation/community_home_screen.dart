import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/app_widgets.dart';

class CommunityHomeScreen extends StatelessWidget {
  const CommunityHomeScreen({super.key});

  @override
  Widget build(BuildContext context) => Scaffold(
        appBar: AppBar(title: const Text('Community')),
        body: SafeArea(
          child: SingleChildScrollView(
            padding: const EdgeInsets.fromLTRB(AppSpacing.page, AppSpacing.sm, AppSpacing.page, AppSpacing.xl),
            child: Center(
              child: ConstrainedBox(
                constraints: const BoxConstraints(maxWidth: 760),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text('Connect through chess', style: AppTextStyles.headline),
                    const SizedBox(height: AppSpacing.xs),
                    const Text('Find players, learn together, and share the game.', style: AppTextStyles.body),
                    const SizedBox(height: AppSpacing.lg),
                    _CommunityOption(icon: Icons.groups_rounded, title: 'Clubs', description: 'Join communities built around your chess interests', path: '/community/clubs'),
                    _CommunityOption(icon: Icons.people_alt_outlined, title: 'Friends', description: 'Manage friends and connection requests', path: '/community/friends'),
                    _CommunityOption(icon: Icons.school_outlined, title: 'Coaches', description: 'Find a coach for your next level', path: '/community/coaches'),
                    _CommunityOption(icon: Icons.chat_bubble_outline_rounded, title: 'Chat', description: 'Continue your conversations', path: '/community/chat'),
                  ],
                ),
              ),
            ),
          ),
        ),
      );
}

class _CommunityOption extends StatelessWidget {
  const _CommunityOption({required this.icon, required this.title, required this.description, required this.path});
  final IconData icon;
  final String title;
  final String description;
  final String path;

  @override
  Widget build(BuildContext context) => AppCard(
        margin: const EdgeInsets.only(bottom: AppSpacing.sm),
        onTap: () => context.push(path),
        child: Row(
          children: [
            Icon(icon, color: AppColors.mint, size: 32),
            const SizedBox(width: AppSpacing.md),
            Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
              Text(title, style: AppTextStyles.titleSmall),
              const SizedBox(height: AppSpacing.xs),
              Text(description, style: AppTextStyles.caption),
            ])),
            const Icon(Icons.chevron_right_rounded, color: AppColors.muted),
          ],
        ),
      );
}
