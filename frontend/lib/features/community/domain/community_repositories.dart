import 'community_models.dart';

abstract interface class CommunityRepository {
  Future<List<Club>> getFeaturedClubs();
  Future<List<Coach>> getCoaches();
}

abstract interface class ClubRepository {
  Future<List<Club>> listClubs({String query = ''});
  Future<Club> getClub(String id);
  Future<List<ClubMember>> getMembers(String id);
  Future<void> join(String id);
  Future<void> leave(String id);
}

abstract interface class FriendRepository {
  Future<List<Friend>> listFriends();
  Future<void> request(String userId);
  Future<void> accept(String userId);
  Future<void> reject(String userId);
  Future<void> remove(String userId);
}

abstract interface class ChatRepository {
  Future<List<ChatRoom>> listRooms();
  Future<List<ChatMessage>> getMessages(String roomId);
  Future<ChatMessage> send(String roomId, String content);
}
