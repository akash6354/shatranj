import '../domain/community_models.dart';
import '../domain/community_repositories.dart';

class MockCommunityRepository implements CommunityRepository {
  @override
  Future<List<Club>> getFeaturedClubs() async => MockClubRepository.clubs;

  @override
  Future<List<Coach>> getCoaches() async => MockCoachRepository.coaches;
}

class MockClubRepository implements ClubRepository {
  static final clubs = [
    const Club(
      id: 'tactics-lab',
      name: 'Tactics Lab',
      description: 'Daily puzzles, analysis, and friendly competition.',
      visibility: 'public',
      memberCount: 1284,
      isMember: true,
      activity: 'Mira shared a new puzzle',
    ),
    const Club(
      id: 'rapid-rising',
      name: 'Rapid Rising',
      description: 'Improve your rapid rating with players from around the world.',
      visibility: 'public',
      memberCount: 642,
      isMember: false,
      activity: 'Weekly arena starts Friday',
    ),
  ];

  static final members = [
    ClubMember(id: 'you', username: 'Akash', role: 'member', joinedAt: DateTime(2026, 1, 2)),
    ClubMember(id: 'mira', username: 'Mira', role: 'moderator', joinedAt: DateTime(2025, 8, 4)),
    ClubMember(id: 'rohan', username: 'Rohan', role: 'member', joinedAt: DateTime(2025, 11, 12)),
  ];

  @override
  Future<List<Club>> listClubs({String query = ''}) async => clubs
      .where((club) => club.name.toLowerCase().contains(query.toLowerCase()))
      .toList();

  @override
  Future<Club> getClub(String id) async => clubs.firstWhere((club) => club.id == id);

  @override
  Future<List<ClubMember>> getMembers(String id) async => members;

  @override
  Future<void> join(String id) async {}

  @override
  Future<void> leave(String id) async {}
}

class MockFriendRepository implements FriendRepository {
  static final friends = [
    const Friend(id: 'mira', username: 'Mira', status: 'online', requestPending: false),
    const Friend(id: 'rohan', username: 'Rohan', status: 'offline', requestPending: false),
    const Friend(id: 'neha', username: 'Neha', status: 'pending', requestPending: true),
  ];

  @override
  Future<List<Friend>> listFriends() async => friends;
  @override
  Future<void> request(String userId) async {}
  @override
  Future<void> accept(String userId) async {}
  @override
  Future<void> reject(String userId) async {}
  @override
  Future<void> remove(String userId) async {}
}

class MockCoachRepository implements CommunityRepository {
  @override
  Future<List<Club>> getFeaturedClubs() async => MockClubRepository.clubs;

  @override
  Future<List<Coach>> getCoaches() async => coaches;

  static const coaches = [
    Coach(
      id: 'coach-1',
      displayName: 'Arjun Mehta',
      title: 'International Master',
      bio: 'Opening preparation and practical tournament coaching.',
      ratePaise: 180000,
      currency: 'INR',
      specialties: ['Openings', 'Tournament prep'],
      availability: 'Available this week',
      status: 'active',
    ),
    Coach(
      id: 'coach-2',
      displayName: 'Sara Khan',
      title: 'Chess Coach',
      bio: 'Clear lessons for improving players and juniors.',
      ratePaise: 120000,
      currency: 'INR',
      specialties: ['Endgames', 'Fundamentals'],
      availability: 'Next slot tomorrow',
      status: 'active',
    ),
  ];
}

class MockChatRepository implements ChatRepository {
  static final rooms = [
    ChatRoom(
      id: 'room-mira',
      name: 'Mira',
      kind: 'direct',
      lastMessage: 'Good luck in the arena!',
      updatedAt: DateTime(2026, 10, 7, 19, 42),
      online: true,
    ),
    ChatRoom(
      id: 'room-tactics',
      name: 'Tactics Lab',
      kind: 'club',
      lastMessage: 'New puzzle is live',
      updatedAt: DateTime(2026, 10, 7, 17, 12),
      online: false,
    ),
  ];

  static final messages = <String, List<ChatMessage>>{
    'room-mira': [
      ChatMessage(id: 'm1', senderId: 'mira', username: 'Mira', content: 'Ready for the tournament?', createdAt: DateTime(2026, 10, 7, 19, 40), isMine: false),
      ChatMessage(id: 'm2', senderId: 'you', username: 'Akash', content: 'Absolutely. Good luck!', createdAt: DateTime(2026, 10, 7, 19, 42), isMine: true),
    ],
    'room-tactics': [
      ChatMessage(id: 'm3', senderId: 'mira', username: 'Mira', content: 'New puzzle is live', createdAt: DateTime(2026, 10, 7, 17, 12), isMine: false),
    ],
  };

  @override
  Future<List<ChatRoom>> listRooms() async => rooms;

  @override
  Future<List<ChatMessage>> getMessages(String roomId) async => messages[roomId] ?? [];

  @override
  Future<ChatMessage> send(String roomId, String content) async {
    final message = ChatMessage(
      id: 'local-${DateTime.now().microsecondsSinceEpoch}',
      senderId: 'you',
      username: 'Akash',
      content: content,
      createdAt: DateTime.now(),
      isMine: true,
    );
    messages.putIfAbsent(roomId, () => []).add(message);
    return message;
  }
}
