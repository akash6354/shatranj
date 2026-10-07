import 'package:flutter/material.dart';

import '../domain/home_models.dart';

abstract interface class HomeRepository {
  Future<HomeSnapshot> getSnapshot();
}

class DemoHomeRepository implements HomeRepository {
  @override
  Future<HomeSnapshot> getSnapshot() async => const HomeSnapshot(
        displayName: 'Akash',
        username: 'akash_the_great',
        rating: 1248,
        streak: 7,
        gamesPlayed: 38,
        puzzlesSolved: 126,
        learningHours: 12,
        dailyPuzzle: DailyPuzzle(
          title: 'Find the best move',
          rating: 1382,
          difficulty: 'Medium',
          fen: 'r1bqk2r/pppp1ppp/2n2n2/4p3/4P3/2N2N2/PPPP1PPP/R1BQKB1R w KQkq - 2 4',
        ),
        courses: [
          CourseProgress(
            title: 'Opening Fundamentals',
            category: 'Openings',
            progress: 62,
            icon: Icons.menu_book_rounded,
          ),
          CourseProgress(
            title: 'Middlegame Strategy',
            category: 'Strategy',
            progress: 35,
            icon: Icons.hub_rounded,
          ),
          CourseProgress(
            title: 'Endgame Mastery',
            category: 'Endgames',
            progress: 18,
            icon: Icons.flag_rounded,
          ),
        ],
        quickActions: [
          HomeAction(
            title: 'Play',
            subtitle: 'Find a game',
            icon: Icons.sports_esports_rounded,
            color: Color(0xFF80D7B9),
          ),
          HomeAction(
            title: 'Puzzles',
            subtitle: 'Sharpen tactics',
            icon: Icons.extension_rounded,
            color: Color(0xFFFFC34D),
          ),
          HomeAction(
            title: 'Learn',
            subtitle: 'Build skills',
            icon: Icons.menu_book_rounded,
            color: Color(0xFF72B8E8),
          ),
          HomeAction(
            title: 'Tournaments',
            subtitle: 'Compete live',
            icon: Icons.emoji_events_rounded,
            color: Color(0xFFF07867),
          ),
        ],
      );
}
