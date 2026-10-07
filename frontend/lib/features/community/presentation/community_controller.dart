import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/network_providers.dart';
import '../data/api_community_repositories.dart';
import '../domain/community_models.dart';
import '../domain/community_repositories.dart';

final communityRepositoryProvider = Provider<CommunityRepository>(
  (ref) => ApiCommunityRepository(ref.watch(networkApiClientProvider)),
);
final clubRepositoryProvider = Provider<ClubRepository>(
  (ref) => ApiClubRepository(ref.watch(networkApiClientProvider)),
);
final friendRepositoryProvider = Provider<FriendRepository>(
  (ref) => ApiFriendRepository(ref.watch(networkApiClientProvider)),
);
final chatRepositoryProvider = Provider<ChatRepository>(
  (ref) => ApiChatRepository(ref.watch(networkApiClientProvider)),
);
final coachRepositoryProvider = Provider<CommunityRepository>(
  (ref) => ApiCoachRepository(ref.watch(networkApiClientProvider)),
);

final clubsProvider = FutureProvider<List<Club>>((ref) => ref.watch(clubRepositoryProvider).listClubs());
final clubProvider = FutureProvider.family<Club, String>((ref, id) => ref.watch(clubRepositoryProvider).getClub(id));
final clubMembersProvider = FutureProvider.family<List<ClubMember>, String>((ref, id) => ref.watch(clubRepositoryProvider).getMembers(id));
final friendsProvider = FutureProvider<List<Friend>>((ref) => ref.watch(friendRepositoryProvider).listFriends());
final coachesProvider = FutureProvider<List<Coach>>((ref) => ref.watch(coachRepositoryProvider).getCoaches());
final chatRoomsProvider = FutureProvider<List<ChatRoom>>((ref) => ref.watch(chatRepositoryProvider).listRooms());
final chatMessagesProvider = FutureProvider.family<List<ChatMessage>, String>((ref, id) => ref.watch(chatRepositoryProvider).getMessages(id));
