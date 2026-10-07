enum ReviewTab { overview, moveByMove, insights }

enum MoveClassification {
  brilliant,
  best,
  good,
  inaccuracy,
  mistake,
  blunder,
  missedWin,
}

class ReviewPlayer {
  const ReviewPlayer({
    required this.name,
    required this.rating,
    required this.accuracy,
    required this.color,
  });

  final String name;
  final int rating;
  final double accuracy;
  final String color;
}

class ReviewMove {
  const ReviewMove({
    required this.number,
    required this.white,
    required this.black,
    required this.evaluation,
    required this.classification,
    required this.fenAfter,
    required this.explanation,
    this.bestMove,
  });

  final int number;
  final String white;
  final String black;
  final double evaluation;
  final MoveClassification classification;
  final String fenAfter;
  final String explanation;
  final String? bestMove;

  String get playedMove => number.isOdd ? white : black;
}

class ReviewStats {
  const ReviewStats({
    required this.brilliant,
    required this.best,
    required this.good,
    required this.inaccuracies,
    required this.mistakes,
    required this.blunders,
    required this.missedWins,
  });

  final int brilliant;
  final int best;
  final int good;
  final int inaccuracies;
  final int mistakes;
  final int blunders;
  final int missedWins;
}

class PhaseStats {
  const PhaseStats({
    required this.name,
    required this.accuracy,
    required this.moves,
    required this.highlight,
  });

  final String name;
  final double accuracy;
  final int moves;
  final String highlight;
}

class ReviewInsight {
  const ReviewInsight({
    required this.title,
    required this.description,
    required this.icon,
    required this.color,
    required this.phase,
  });

  final String title;
  final String description;
  final String icon;
  final int color;
  final String phase;
}

class GameReview {
  const GameReview({
    required this.id,
    required this.white,
    required this.black,
    required this.result,
    required this.resultLabel,
    required this.opening,
    required this.stats,
    required this.moves,
    required this.phases,
    required this.insights,
    required this.rating,
    required this.ratingChange,
  });

  final String id;
  final ReviewPlayer white;
  final ReviewPlayer black;
  final String result;
  final String resultLabel;
  final String opening;
  final ReviewStats stats;
  final List<ReviewMove> moves;
  final List<PhaseStats> phases;
  final List<ReviewInsight> insights;
  final int rating;
  final int ratingChange;
}
