class Profile {
  const Profile({
    required this.username,
    required this.displayName,
    required this.avatarUrl,
    required this.rating,
    required this.blitzRating,
    required this.rapidRating,
    required this.puzzleRating,
    required this.games,
    required this.winRate,
    required this.streak,
    required this.achievements,
    required this.friends,
  });

  final String username;
  final String displayName;
  final String avatarUrl;
  final int rating;
  final int blitzRating;
  final int rapidRating;
  final int puzzleRating;
  final int games;
  final double winRate;
  final int streak;
  final List<String> achievements;
  final int friends;
}
