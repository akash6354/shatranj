import '../domain/profile_models.dart';
import '../domain/profile_repository.dart';

class MockProfileRepository implements ProfileRepository {
  @override
  Future<Profile> getProfile() async => const Profile(
        username: 'akash',
        displayName: 'Akash',
        avatarUrl: '',
        rating: 1248,
        blitzRating: 1284,
        rapidRating: 1212,
        puzzleRating: 1382,
        games: 186,
        winRate: 58,
        streak: 7,
        achievements: ['First victory', 'Puzzle streak', 'Opening student'],
        friends: 24,
      );
}
