import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/app_widgets.dart';
import '../domain/tournament_models.dart';
import 'tournament_controller.dart';

class TournamentDetailScreen extends ConsumerWidget {
  const TournamentDetailScreen({required this.tournamentId, super.key});
  final String tournamentId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final tournament = ref.watch(tournamentProvider(tournamentId));
    return Scaffold(
      appBar: AppBar(title: const Text('Tournament')),
      body: SafeArea(
        child: tournament.when(
          loading: () => const AppLoading(),
          error: (error, _) => AppError(title: 'Tournament unavailable', message: error.toString()),
          data: (item) => _Detail(tournament: item),
        ),
      ),
    );
  }
}

class _Detail extends ConsumerWidget {
  const _Detail({required this.tournament});
  final Tournament tournament;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final standings = ref.watch(standingsProvider(tournament.id));
    return SingleChildScrollView(
      padding: const EdgeInsets.fromLTRB(AppSpacing.page, AppSpacing.sm, AppSpacing.page, AppSpacing.xl),
      child: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 860),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              AppGradientCard(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        const Icon(Icons.emoji_events_rounded, color: AppColors.gold, size: 34),
                        const SizedBox(width: AppSpacing.sm),
                        Expanded(child: Text(tournament.name, style: AppTextStyles.headline)),
                      ],
                    ),
                    const SizedBox(height: AppSpacing.sm),
                    Text(tournament.description, style: AppTextStyles.body),
                    const SizedBox(height: AppSpacing.md),
                    Wrap(
                      spacing: AppSpacing.md,
                      runSpacing: AppSpacing.xs,
                      children: [
                        AppPill(label: tournament.format, icon: Icons.account_tree_outlined),
                        AppPill(label: tournament.timeControl, icon: Icons.timer_outlined),
                        AppPill(label: '${tournament.maxPlayers} players', icon: Icons.people_outline),
                      ],
                    ),
                  ],
                ),
              ),
              const SizedBox(height: AppSpacing.md),
              _Action(tournament: tournament),
              const SizedBox(height: AppSpacing.lg),
              const AppSectionHeader(title: 'Standings', subtitle: 'Current tournament leaderboard'),
              const SizedBox(height: AppSpacing.sm),
              standings.when(
                loading: () => const AppLoading(),
                error: (error, _) => AppError(message: error.toString()),
                data: (players) => _Standings(players: players),
              ),
              const SizedBox(height: AppSpacing.lg),
              const AppSectionHeader(title: 'Rounds'),
              const SizedBox(height: AppSpacing.sm),
              if (tournament.rounds.isEmpty)
                const AppCard(child: Text('Rounds will appear when the tournament starts.', style: AppTextStyles.body))
              else
                ...tournament.rounds.map((round) => AppCard(
                      margin: const EdgeInsets.only(bottom: AppSpacing.sm),
                      child: Row(
                        children: [
                          Text('Round ${round.number}', style: AppTextStyles.bodyStrong),
                          const Spacer(),
                          Text(round.status, style: AppTextStyles.caption),
                          if (round.currentGameId != null) ...[
                            const SizedBox(width: AppSpacing.sm),
                            AppButton(
                              label: 'Current game',
                              variant: AppButtonVariant.ghost,
                              onPressed: () => context.push('/game/${round.currentGameId}'),
                            ),
                          ],
                        ],
                      ),
                    )),
            ],
          ),
        ),
      ),
    );
  }
}

class _Action extends ConsumerWidget {
  const _Action({required this.tournament});
  final Tournament tournament;

  @override
  Widget build(BuildContext context, WidgetRef ref) => AppButton(
        label: tournament.isRegistered ? 'Leave tournament' : 'Join tournament',
        icon: tournament.isRegistered ? Icons.logout_rounded : Icons.login_rounded,
        variant: tournament.isRegistered ? AppButtonVariant.outline : AppButtonVariant.primary,
        expand: true,
        onPressed: tournament.status != TournamentStatus.upcoming
            ? null
            : () async {
                final repository = ref.read(tournamentRepositoryProvider);
                if (tournament.isRegistered) {
                  await repository.withdraw(tournament.id);
                } else {
                  await repository.register(tournament.id);
                }
                ref.invalidate(tournamentProvider(tournament.id));
                ref.invalidate(tournamentListProvider);
              },
      );
}

class _Standings extends StatelessWidget {
  const _Standings({required this.players});
  final List<TournamentPlayer> players;

  @override
  Widget build(BuildContext context) => AppCard(
        padding: EdgeInsets.zero,
        child: Column(
          children: players.map((player) => Padding(
                padding: const EdgeInsets.symmetric(horizontal: AppSpacing.md, vertical: AppSpacing.sm),
                child: Row(
                  children: [
                    SizedBox(width: 28, child: Text('${player.rank ?? '-'}', style: AppTextStyles.bodyStrong)),
                    const CircleAvatar(radius: 16, child: Icon(Icons.person, size: 17)),
                    const SizedBox(width: AppSpacing.sm),
                    Expanded(child: Text(player.username, style: AppTextStyles.bodyStrong)),
                    Text('${player.rating}', style: AppTextStyles.caption),
                    const SizedBox(width: AppSpacing.md),
                    SizedBox(width: 32, child: Text('${player.score}', style: AppTextStyles.bodyStrong)),
                    SizedBox(width: 36, child: Text('${player.games}g', style: AppTextStyles.caption)),
                  ],
                ),
              )).toList(),
        ),
      );
}
