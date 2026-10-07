class Club {
  const Club({
    required this.id,
    required this.name,
    required this.description,
    required this.visibility,
    required this.memberCount,
    required this.isMember,
    this.activity = 'No recent activity',
  });

  final String id;
  final String name;
  final String description;
  final String visibility;
  final int memberCount;
  final bool isMember;
  final String activity;
}

class ClubMember {
  const ClubMember({
    required this.id,
    required this.username,
    required this.role,
    required this.joinedAt,
  });

  final String id;
  final String username;
  final String role;
  final DateTime joinedAt;
}

class Friend {
  const Friend({
    required this.id,
    required this.username,
    required this.status,
    required this.requestPending,
  });

  final String id;
  final String username;
  final String status;
  final bool requestPending;

  bool get isOnline => status == 'online';
}

class Coach {
  const Coach({
    required this.id,
    required this.displayName,
    required this.title,
    required this.bio,
    required this.ratePaise,
    required this.currency,
    required this.specialties,
    required this.availability,
    required this.status,
  });

  final String id;
  final String displayName;
  final String title;
  final String bio;
  final int ratePaise;
  final String currency;
  final List<String> specialties;
  final String availability;
  final String status;
}

class ChatRoom {
  const ChatRoom({
    required this.id,
    required this.name,
    required this.kind,
    required this.lastMessage,
    required this.updatedAt,
    required this.online,
  });

  final String id;
  final String name;
  final String kind;
  final String lastMessage;
  final DateTime updatedAt;
  final bool online;
}

class ChatMessage {
  const ChatMessage({
    required this.id,
    required this.senderId,
    required this.username,
    required this.content,
    required this.createdAt,
    required this.isMine,
  });

  final String id;
  final String senderId;
  final String username;
  final String content;
  final DateTime createdAt;
  final bool isMine;
}
