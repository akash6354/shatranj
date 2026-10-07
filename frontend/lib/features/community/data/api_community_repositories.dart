import '../../../core/network/api_client.dart';
import '../domain/community_models.dart';
import '../domain/community_repositories.dart';

class ApiCommunityRepository implements CommunityRepository {
  ApiCommunityRepository(this._client);
  final ApiClient _client;

  @override
  Future<List<Club>> getFeaturedClubs() async =>
      ApiClubRepository(_client).listClubs();

  @override
  Future<List<Coach>> getCoaches() async =>
      ApiCoachRepository(_client).getCoaches();
}

class ApiClubRepository implements ClubRepository {
  ApiClubRepository(this._client);
  final ApiClient _client;

  @override
  Future<List<Club>> listClubs({String query = ''}) async {
    final response = await _client.get('/clubs');
    final clubs = _list(response).map(_parseClub);
    final normalized = query.trim().toLowerCase();
    return normalized.isEmpty
        ? clubs.toList()
        : clubs
            .where((club) =>
                club.name.toLowerCase().contains(normalized) ||
                club.description.toLowerCase().contains(normalized))
            .toList();
  }

  @override
  Future<Club> getClub(String id) async =>
      _parseClub(await _client.get('/clubs/$id'));

  @override
  Future<List<ClubMember>> getMembers(String id) async {
    final response = await _client.get('/clubs/$id/members');
    return _list(response).map(_parseMember).toList();
  }

  @override
  Future<void> join(String id) async {
    await _client.post('/clubs/$id/join');
  }

  @override
  Future<void> leave(String id) async {
    await _client.delete('/clubs/$id/membership');
  }

  Club _parseClub(Map<String, dynamic> data) => Club(
        id: _string(data, 'id'),
        name: _string(data, 'name'),
        description: _string(data, 'description'),
        visibility: _string(data, 'visibility'),
        memberCount: 0,
        isMember: _string(data, 'role').isNotEmpty,
      );

  ClubMember _parseMember(Map<String, dynamic> data) => ClubMember(
        id: _string(data, 'user_id'),
        username: _string(data, 'username'),
        role: _string(data, 'role'),
        joinedAt: _date(data['joined_at']),
      );
}

class ApiFriendRepository implements FriendRepository {
  ApiFriendRepository(this._client);
  final ApiClient _client;

  @override
  Future<List<Friend>> listFriends() async {
    final response = await _client.get('/friends');
    return _list(response)
        .map((data) => Friend(
              id: _string(data, 'user_id'),
              username: _string(data, 'username'),
              status: _string(data, 'status'),
              requestPending: _string(data, 'status') != 'accepted',
            ))
        .toList();
  }

  @override
  Future<void> request(String userId) async {
    await _client.post('/friends/requests', body: {'user_id': userId});
  }

  @override
  Future<void> accept(String userId) async {
    await _client.post('/friends/requests/$userId/accept');
  }

  @override
  Future<void> reject(String userId) async {
    await _client.post('/friends/requests/$userId/reject');
  }

  @override
  Future<void> remove(String userId) async {
    throw const ApiException(
      501,
      'The backend does not expose friend removal; block is the supported action.',
    );
  }
}

class ApiCoachRepository implements CommunityRepository {
  ApiCoachRepository(this._client);
  final ApiClient _client;

  @override
  Future<List<Club>> getFeaturedClubs() async => const [];

  @override
  Future<List<Coach>> getCoaches() async {
    final response = await _client.get('/coaches');
    return _list(response)
        .map(
          (data) => Coach(
            id: _string(data, 'user_id'),
            displayName: _string(data, 'display_name'),
            title: _string(data, 'title'),
            bio: _string(data, 'bio'),
            ratePaise: _int(data, 'rate_paise'),
            currency: _string(data, 'currency'),
            specialties: _strings(data['specialties']),
            availability: data['availability']?.toString() ?? '',
            status: _string(data, 'status'),
          ),
        )
        .toList();
  }
}

class ApiChatRepository implements ChatRepository {
  ApiChatRepository(this._client);
  final ApiClient _client;

  @override
  Future<List<ChatRoom>> listRooms() async {
    final response = await _client.get('/chat/rooms');
    return _list(response)
        .map(
          (data) => ChatRoom(
            id: _string(data, 'id'),
            name: _string(data, 'reference_id').isEmpty
                ? _string(data, 'kind')
                : _string(data, 'reference_id'),
            kind: _string(data, 'kind'),
            lastMessage: '',
            updatedAt: _date(data['created_at']),
            online: false,
          ),
        )
        .toList();
  }

  @override
  Future<List<ChatMessage>> getMessages(String roomId) async {
    final response = await _client.get('/chat/rooms/$roomId/messages');
    return _list(response)
        .map(
          (data) => ChatMessage(
            id: _string(data, 'id'),
            senderId: _string(data, 'sender_id'),
            username: _string(data, 'username'),
            content: _string(data, 'content'),
            createdAt: _date(data['created_at']),
            isMine: false,
          ),
        )
        .toList();
  }

  @override
  Future<ChatMessage> send(String roomId, String content) async {
    final response = await _client.post(
      '/chat/rooms/$roomId/messages',
      body: {'content': content},
    );
    return ChatMessage(
      id: _string(response, 'id'),
      senderId: _string(response, 'sender_id'),
      username: _string(response, 'username'),
      content: _string(response, 'content'),
      createdAt: _date(response['created_at']),
      isMine: true,
    );
  }
}

List<Map<String, dynamic>> _list(Map<String, dynamic> response) {
  final value = response['data'];
  return value is List
      ? value.whereType<Map>().map(Map<String, dynamic>.from).toList()
      : const [];
}

String _string(Map<String, dynamic> data, String key) =>
    data[key]?.toString() ?? '';

int _int(Map<String, dynamic> data, String key) =>
    (data[key] as num?)?.toInt() ?? 0;

List<String> _strings(Object? value) =>
    value is List ? value.map((item) => item.toString()).toList() : const [];

DateTime _date(Object? value) =>
    DateTime.tryParse(value?.toString() ?? '') ?? DateTime.now();
