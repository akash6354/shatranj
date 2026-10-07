import 'package:flutter/material.dart';

class HomeSnapshot {
  const HomeSnapshot({
    required this.displayName,
    required this.username,
    required this.rating,
    required this.streak,
    required this.gamesPlayed,
    required this.puzzlesSolved,
    required this.learningHours,
    required this.dailyPuzzle,
    required this.courses,
    required this.quickActions,
  });

  final String displayName;
  final String username;
  final int rating;
  final int streak;
  final int gamesPlayed;
  final int puzzlesSolved;
  final int learningHours;
  final DailyPuzzle dailyPuzzle;
  final List<CourseProgress> courses;
  final List<HomeAction> quickActions;
}

class DailyPuzzle {
  const DailyPuzzle({
    required this.title,
    required this.rating,
    required this.difficulty,
    required this.fen,
  });

  final String title;
  final int rating;
  final String difficulty;
  final String fen;
}

class CourseProgress {
  const CourseProgress({
    required this.title,
    required this.category,
    required this.progress,
    required this.icon,
  });

  final String title;
  final String category;
  final int progress;
  final IconData icon;
}

class HomeAction {
  const HomeAction({
    required this.title,
    required this.subtitle,
    required this.icon,
    required this.color,
  });

  final String title;
  final String subtitle;
  final IconData icon;
  final Color color;
}
