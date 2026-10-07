import 'dart:async';
import 'dart:convert';
import 'dart:io';

import '../../../core/network/api_client.dart';
import '../../../core/network/token_storage.dart';
import '../domain/game_models.dart';
import '../domain/game_repository.dart';
import '../domain/game_socket.dart';

class ApiGameRepository implements GameRepository {
  ApiGameRepository(this._client, this._tokens);

  final ApiClient _client;
  final TokenStorage _tokens;

  @override
  Future<GameState> getGame(String gameId) async =>
      gameStateFromJson(await _client.get('/games/$gameId'), gameId);

  @override
  GameSocket createSocket() => ApiGameSocket(_client, _tokens);
}

class ApiGameSocket implements GameSocket {
  ApiGameSocket(this._client, this._tokens);

  final ApiClient _client;
  final TokenStorage _tokens;
  final _events = StreamController<GameState>.broadcast();
  WebSocket? _socket;
  GameState? _state;
  String? _gameId;
  DateTime? _lastEventAt;
  String? _lastEventKey;

  @override
  Stream<GameState> get events => _events.stream;

  @override
  Future<void> connect(String gameId) async {
    _gameId = gameId;
    _emit(_state?.copyWith(connection: GameConnectionStatus.connecting));
    final token = await _tokens.readAccessToken();
    if (token == null || token.isEmpty) {
      throw const GameActionException(
        'Authentication is required for online games.',
      );
    }
    final base = Uri.parse(_client.baseUrl);
    final scheme = base.scheme == 'https' ? 'wss' : 'ws';
    final uri = base.replace(
      scheme: scheme,
      path: '${base.path}/games/$gameId/ws',
      query: '',
    );
    _socket = await WebSocket.connect(
      uri.toString(),
      headers: {'Authorization': 'Bearer $token'},
    );
    _socket!.listen(
      (message) => _handleMessage(message),
      onError: (Object error, StackTrace stackTrace) {
        _emit(_state?.copyWith(connection: GameConnectionStatus.failed));
        _events.addError(error, stackTrace);
      },
      onDone: () => _emit(
        _state?.copyWith(connection: GameConnectionStatus.disconnected),
      ),
      cancelOnError: false,
    );
    _emit(_state?.copyWith(connection: GameConnectionStatus.connected));
  }

  @override
  Future<void> reconnect() async {
    final gameId = _gameId;
    if (gameId == null) return;
    await disconnect();
    _emit(_state?.copyWith(connection: GameConnectionStatus.reconnecting));
    await connect(gameId);
  }

  @override
  Future<void> sendMove(String uci) async {
    final gameId = _requireGameId();
    _emitFromResponse(
      await _client.post('/games/$gameId/moves', body: {'uci': uci}),
    );
  }

  @override
  Future<void> offerDraw() async {
    final gameId = _requireGameId();
    _emitFromResponse(await _client.post('/games/$gameId/draw-offers'));
  }

  @override
  Future<void> respondToDraw({required bool accept}) async {
    final gameId = _requireGameId();
    if (accept) {
      _emitFromResponse(
        await _client.post('/games/$gameId/draw-offers/accept'),
      );
    }
  }

  @override
  Future<void> resign() async {
    final gameId = _requireGameId();
    _emitFromResponse(await _client.post('/games/$gameId/resign'));
  }

  @override
  Future<void> disconnect() async {
    await _socket?.close();
    _socket = null;
    _emit(_state?.copyWith(connection: GameConnectionStatus.disconnected));
  }

  @override
  Future<void> dispose() async {
    await disconnect();
    await _events.close();
  }

  String _requireGameId() {
    final gameId = _gameId;
    if (gameId == null || _socket == null) {
      throw const GameActionException('The game is not connected.');
    }
    return gameId;
  }

  void _handleMessage(Object? message) {
    try {
      if (message is! String) return;
      final decoded = jsonDecode(message);
      if (decoded is! Map<String, dynamic>) return;
      final type = decoded['type']?.toString();
      if (type == 'ping' || type == 'pong' || type == 'connected') return;
      final createdAt = DateTime.tryParse(
        decoded['created_at']?.toString() ?? '',
      );
      final key =
          '$type|${decoded['created_at']}|${jsonEncode(decoded['game'])}';
      if (key == _lastEventKey ||
          (createdAt != null &&
              _lastEventAt != null &&
              createdAt.isBefore(_lastEventAt!))) {
        return;
      }
      _lastEventKey = key;
      if (createdAt != null) _lastEventAt = createdAt;
      if (type == 'error' || type == 'server_error') {
        _events.addError(
          GameActionException(
            decoded['message']?.toString() ?? 'Game server error',
          ),
        );
        return;
      }
      final game = decoded['game'] ?? decoded['data'];
      if (game is Map) {
        _emit(gameStateFromJson(Map<String, dynamic>.from(game), _gameId!));
      }
    } on Object catch (error, stackTrace) {
      _events.addError(error, stackTrace);
    }
  }

  void _emitFromResponse(Map<String, dynamic> response) {
    _emit(gameStateFromJson(response, _gameId!));
  }

  void _emit(GameState? state) {
    if (state == null || _events.isClosed) return;
    _state = state;
    _events.add(state);
  }
}

GameState gameStateFromJson(Map<String, dynamic> data, String fallbackId) {
  final moves = data['moves'];
  final whiteId = data['white_player_id']?.toString() ?? '';
  final blackId = data['black_player_id']?.toString() ?? '';
  final fenFields = (data['current_fen']?.toString() ?? '').split(' ');
  final side = fenFields.length > 1 && fenFields[1] == 'b' ? blackId : whiteId;
  return GameState(
    id: data['id']?.toString() ?? fallbackId,
    fen: data['current_fen']?.toString() ?? '',
    white: GamePlayer(id: whiteId, name: whiteId, rating: 0, color: 'white'),
    black: GamePlayer(id: blackId, name: blackId, rating: 0, color: 'black'),
    whiteTime: Duration(milliseconds: _int(data, 'white_clock_ms')),
    blackTime: Duration(milliseconds: _int(data, 'black_clock_ms')),
    turn: side,
    status: GameStatus.values.firstWhere(
      (value) => value.name == _string(data, 'status'),
      orElse: () => GameStatus.waiting,
    ),
    result: _result(_string(data, 'result')),
    moves: moves is List
        ? moves
              .whereType<Map>()
              .map((item) => _parseMove(Map<String, dynamic>.from(item)))
              .toList()
        : const [],
    connection: GameConnectionStatus.connected,
    drawOfferBy: data['draw_offer_by']?.toString(),
    endReason: data['end_reason']?.toString(),
    lastMove: moves is List && moves.isNotEmpty && moves.last is Map
        ? _parseMove(Map<String, dynamic>.from(moves.last as Map))
        : null,
  );
}

GameMove _parseMove(Map<String, dynamic> data) {
  final uci = data['uci']?.toString() ?? '';
  return GameMove(
    number: _int(data, 'number'),
    playerId: data['player_id']?.toString() ?? '',
    from: uci.length >= 2 ? uci.substring(0, 2) : '',
    to: uci.length >= 4 ? uci.substring(2, 4) : '',
    san: data['san']?.toString(),
  );
}

int _int(Map<String, dynamic> data, String key) =>
    (data[key] as num?)?.toInt() ?? 0;

String _string(Map<String, dynamic> data, String key) =>
    data[key]?.toString() ?? '';

GameResult _result(String value) => switch (value) {
  '1-0' => GameResult.whiteWin,
  '0-1' => GameResult.blackWin,
  '1/2-1/2' => GameResult.draw,
  _ => GameResult.ongoing,
};
