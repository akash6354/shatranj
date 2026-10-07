import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/app_widgets.dart';
import '../domain/tournament_models.dart';
import 'tournament_controller.dart';

class TournamentHomeScreen extends ConsumerStatefulWidget {
  const TournamentHomeScreen({super.key});

  @override
  ConsumerState<TournamentHomeScreen> createState() => _TournamentHomeScreenState();
}

class _TournamentHomeScreenState extends ConsumerState<TournamentHomeScreen> {
  var _tab = 0;

  @override
  Widget build(BuildContext context) {
    final tournaments = ref.watch(tournamentListProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Tournaments')),
      body: SafeArea(
        child: tournaments.when(
          loading: () => const AppLoading(message: 'Loading tournaments'),
          error: (error, _) => AppError(title: 'Tournaments unavailable', message: error.toString()),
          data: (items) => Column(
            children: [
              TabBar(
                onTap: (value) => setState(() => _tab = value),
                tabs: const [
                  Tab(text: 'Live'),
                  Tab(text: 'Upcoming'),
                  Tab(text: 'My Tournaments'),
                ],
              ),
              Expanded(child: _TournamentList(items: _filter(items), tab: _tab)),
            ],
          ),
        ),
      ),
    );
  }

  List<Tournament> _filter(List<Tournament> items) => switch (_tab) {
        0 => items.where((item) => item.status == TournamentStatus.live).toList(),
        1 => items.where((item) => item.status == TournamentStatus.upcoming).toList(),
        _ => items.where((item) => item.isRegistered).toList(),
      };
}

class _TournamentList extends StatelessWidget {
  const _TournamentList({required this.items, required this.tab});
  final List<Tournament> items;
  final int tab;

  @override
  Widget build(BuildContext context) => items.isEmpty
      ? const AppEmptyState(
          title: 'No tournaments here',
          message: 'Check back soon for more events.',
          icon: Icons.emoji_events_outlined,
        )
      : ListView(
          padding: const EdgeInsets.fromLTRB(AppSpacing.page, AppSpacing.md, AppSpacing.page, AppSpacing.xl),
          children: items.map((item) => _TournamentCard(tournament: item)).toList(),
        );
}

class _TournamentCard extends StatelessWidget {
  const _TournamentCard({required this.tournament});
  final Tournament tournament;

  @override
  Widget build(BuildContext context) => AppCard(
        margin: const EdgeInsets.only(bottom: AppSpacing.sm),
        onTap: () => context.push('/tournaments/${tournament.id}'),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(
                  tournament.status == TournamentStatus.live
                      ? Icons.circle
                      : Icons.schedule_rounded,
                  color: tournament.status == TournamentStatus.live ? AppColors.success : AppColors.gold,
                  size: 14,
                ),
                const SizedBox(width: AppSpacing.xs),
                Text(
                  tournament.status == TournamentStatus.live ? 'LIVE NOW' : 'UPCOMING',
                  style: AppTextStyles.caption.copyWith(
                    color: tournament.status == TournamentStatus.live ? AppColors.success : AppColors.gold,
                    fontWeight: FontWeight.w700,
                  ),
                ),
                const Spacer(),
                AppPill(label: tournament.format),
              ],
            ),
            const SizedBox(height: AppSpacing.sm),
            Text(tournament.name, style: AppTextStyles.titleSmall),
            const SizedBox(height: AppSpacing.xs),
            Text(tournament.description, style: AppTextStyles.caption, maxLines: 2, overflow: TextOverflow.ellipsis),
            const SizedBox(height: AppSpacing.md),
            Wrap(
              spacing: AppSpacing.md,
              runSpacing: AppSpacing.xs,
              children: [
                _Info(icon: Icons.timer_outlined, label: tournament.timeControl),
                _Info(icon: Icons.people_outline, label: '${tournament.participants.length}/${tournament.maxPlayers}'),
                _Info(
                  icon: Icons.schedule_outlined,
                  label: tournament.status == TournamentStatus.live
                      ? 'In progress'
                      : _startLabel(tournament.startsAt),
                ),
              ],
            ),
          ],
        ),
      );
}

class _Info extends StatelessWidget {
  const _Info({required this.icon, required this.label});
  final IconData icon;
  final String label;

  @override
  Widget build(BuildContext context) => Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 16, color: AppColors.muted),
          const SizedBox(width: AppSpacing.xs),
          Text(label, style: AppTextStyles.caption),
        ],
      );
}

String _startLabel(DateTime date) =>
    '${date.day}/${date.month} at ${date.hour.toString().padLeft(2, '0')}:${date.minute.toString().padLeft(2, '0')}';
