import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/network_providers.dart';
import '../data/api_tournament_repository.dart';
import '../domain/tournament_models.dart';
import '../domain/tournament_repository.dart';

final tournamentRepositoryProvider = Provider<TournamentRepository>(
  (ref) => ApiTournamentRepository(ref.watch(networkApiClientProvider)),
);

final tournamentListProvider = FutureProvider<List<Tournament>>(
  (ref) => ref.watch(tournamentRepositoryProvider).listTournaments(),
);

final tournamentProvider =
    FutureProvider.family<Tournament, String>((ref, id) {
  return ref.watch(tournamentRepositoryProvider).getTournament(id);
});

final standingsProvider =
    FutureProvider.family<List<TournamentPlayer>, String>((ref, id) {
  return ref.watch(tournamentRepositoryProvider).getStandings(id);
});
