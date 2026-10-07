import 'dart:async';

import '../../../core/network/api_client.dart';
import '../domain/review_models.dart';
import '../domain/review_repository.dart';

class ApiReviewRepository implements ReviewRepository {
  ApiReviewRepository(this._client);

  final ApiClient _client;

  @override
  Future<GameReview> getReview(String gameId) async {
    await _requestReview(gameId);
    for (var attempt = 0; attempt < 20; attempt++) {
      final response = await _client.get('/games/$gameId/review');
      final status = response['status']?.toString() ?? '';
      if (status == 'completed') return _parse(response, gameId);
      if (status == 'failed' || status == 'unavailable') {
        throw ApiException(
          status == 'unavailable' ? 503 : 500,
          response['message']?.toString() ?? 'Game analysis failed',
        );
      }
      await Future<void>.delayed(const Duration(seconds: 1));
    }
    throw const ApiException(408, 'Game analysis is still pending.');
  }

  Future<void> _requestReview(String gameId) async {
    try {
      await _client.post('/games/$gameId/review');
    } on ApiException catch (error) {
      if (error.statusCode != 202 && error.statusCode != 200) rethrow;
    }
  }

  GameReview _parse(Map<String, dynamic> data, String gameId) {
    final result = data['result'] is Map
        ? Map<String, dynamic>.from(data['result'] as Map)
        : <String, dynamic>{};
    final moves = result['moves'] is List
        ? (result['moves'] as List)
              .whereType<Map>()
              .map((item) => Map<String, dynamic>.from(item))
              .toList()
        : const <Map<String, dynamic>>[];
    final whiteAccuracy = _double(result, 'white_accuracy');
    final blackAccuracy = _double(result, 'black_accuracy');
    return GameReview(
      id: data['id']?.toString() ?? gameId,
      white: ReviewPlayer(
        name: _string(result, 'white_player_id', fallback: 'White'),
        rating: 0,
        accuracy: whiteAccuracy,
        color: 'White',
      ),
      black: ReviewPlayer(
        name: _string(result, 'black_player_id', fallback: 'Black'),
        rating: 0,
        accuracy: blackAccuracy,
        color: 'Black',
      ),
      result: '*',
      resultLabel: 'Analysis complete',
      opening: 'Backend analysis',
      stats: _stats(moves),
      moves: _reviewMoves(moves),
      phases: const [],
      insights: const [],
      rating: 0,
      ratingChange: 0,
    );
  }

  List<ReviewMove> _reviewMoves(List<Map<String, dynamic>> moves) =>
      moves.map((move) {
        final color = _string(move, 'color');
        final san = _string(move, 'san');
        return ReviewMove(
          number: _int(move, 'move_number'),
          white: color == 'white' ? san : '',
          black: color == 'black' ? san : '',
          evaluation: _int(move, 'score_cp') / 100,
          classification: _classification(_string(move, 'classification')),
          fenAfter: '',
          explanation: '',
          bestMove: _string(move, 'best_move').isEmpty
              ? null
              : _string(move, 'best_move'),
        );
      }).toList();

  ReviewStats _stats(List<Map<String, dynamic>> moves) {
    var brilliant = 0;
    var best = 0;
    var good = 0;
    var inaccuracies = 0;
    var mistakes = 0;
    var blunders = 0;
    for (final move in moves) {
      switch (_string(move, 'classification')) {
        case 'brilliant':
          brilliant++;
        case 'best':
          best++;
        case 'good':
        case 'great':
          good++;
        case 'inaccuracy':
          inaccuracies++;
        case 'mistake':
          mistakes++;
        case 'blunder':
          blunders++;
      }
    }
    return ReviewStats(
      brilliant: brilliant,
      best: best,
      good: good,
      inaccuracies: inaccuracies,
      mistakes: mistakes,
      blunders: blunders,
      missedWins: 0,
    );
  }

  MoveClassification _classification(String value) => switch (value) {
    'brilliant' => MoveClassification.brilliant,
    'best' || 'great' => MoveClassification.best,
    'inaccuracy' => MoveClassification.inaccuracy,
    'mistake' => MoveClassification.mistake,
    'blunder' => MoveClassification.blunder,
    _ => MoveClassification.good,
  };

  String _string(
    Map<String, dynamic> data,
    String key, {
    String fallback = '',
  }) => data[key]?.toString() ?? fallback;

  int _int(Map<String, dynamic> data, String key) =>
      (data[key] as num?)?.toInt() ?? 0;

  double _double(Map<String, dynamic> data, String key) =>
      (data[key] as num?)?.toDouble() ?? 0;
}
