import 'dart:async';

import '../../../core/widgets/chess/chess_models.dart';
import '../domain/game_models.dart';
import '../domain/game_repository.dart';
import '../domain/game_socket.dart';

class MockGameRepository implements GameRepository {
  @override
  Future<GameState> getGame(String gameId) async => _initialState(gameId);

  @override
  GameSocket createSocket() => MockGameSocket(_initialState('demo-game'));

  GameState _initialState(String gameId) => GameState(
    id: gameId,
    fen: 'r1bqk2r/pppp1ppp/2n2n2/4p3/4P3/2N2N2/PPPP1PPP/R1BQKB1R w KQkq - 2 4',
    white: const GamePlayer(
      id: 'you',
      name: 'Akash',
      rating: 1248,
      color: 'white',
    ),
    black: const GamePlayer(
      id: 'opponent',
      name: 'Rohan',
      rating: 1223,
      color: 'black',
    ),
    whiteTime: const Duration(minutes: 3),
    blackTime: const Duration(minutes: 3),
    turn: 'you',
    status: GameStatus.active,
    result: GameResult.ongoing,
    moves: const [],
    connection: GameConnectionStatus.disconnected,
  );
}

/// Development-only socket. It mirrors state updates but never validates chess.
class MockGameSocket implements GameSocket {
  MockGameSocket(this._state);

  GameState _state;
  final _events = StreamController<GameState>.broadcast();

  @override
  Stream<GameState> get events => _events.stream;

  @override
  Future<void> connect(String gameId) async {
    _state = _state.copyWith(connection: GameConnectionStatus.connecting);
    _events.add(_state);
    await Future<void>.delayed(const Duration(milliseconds: 300));
    _state = _state.copyWith(connection: GameConnectionStatus.connected);
    _events.add(_state);
  }

  @override
  Future<void> reconnect() => connect(_state.id);

  @override
  Future<void> sendMove(String uci) async {
    if (_state.connection != GameConnectionStatus.connected) {
      throw const GameActionException(
        'You are offline. Reconnect to send a move.',
      );
    }
    final move = _parseMove(uci);
    if (move == null) {
      throw const GameActionException(
        'Select a source and destination square.',
      );
    }
    await Future<void>.delayed(const Duration(milliseconds: 250));
    final moves = [..._state.moves, move];
    _state = _state.copyWith(
      fen: _movePiece(_state.fen, move.from, move.to),
      moves: moves,
      lastMove: move,
      turn: _state.turn == _state.white.id ? _state.black.id : _state.white.id,
    );
    _events.add(_state);
  }

  @override
  Future<void> offerDraw() async {
    _state = _state.copyWith(drawOfferBy: _state.white.id);
    _events.add(_state);
  }

  @override
  Future<void> respondToDraw({required bool accept}) async {
    if (!accept) {
      _state = _state.copyWith(drawOfferBy: '');
    } else {
      _state = _state.copyWith(
        drawOfferBy: '',
        status: GameStatus.finished,
        result: GameResult.draw,
        endReason: 'Draw agreed',
      );
    }
    _events.add(_state);
  }

  @override
  Future<void> resign() async {
    _state = _state.copyWith(
      status: GameStatus.finished,
      result: GameResult.blackWin,
      endReason: 'You resigned',
    );
    _events.add(_state);
  }

  @override
  Future<void> disconnect() async {
    _state = _state.copyWith(connection: GameConnectionStatus.disconnected);
    _events.add(_state);
    await _events.close();
  }

  @override
  Future<void> dispose() => disconnect();

  GameMove? _parseMove(String uci) {
    if (uci.length < 4) return null;
    return GameMove(
      number: _state.moves.length + 1,
      playerId: _state.turn,
      from: uci.substring(0, 2),
      to: uci.substring(2, 4),
    );
  }

  String _movePiece(String fen, String from, String to) {
    final position = ChessPosition.fromFen(fen);
    final fromSquare = ChessSquare.parse(from);
    final toSquare = ChessSquare.parse(to);
    if (fromSquare == null || toSquare == null) return fen;
    final piece = position.squares[fromSquare];
    if (piece == null) return fen;
    final squares = Map<ChessSquare, ChessPiece>.from(position.squares)
      ..remove(fromSquare)
      ..[toSquare] = piece;
    final rows = <String>[];
    for (var rank = 7; rank >= 0; rank--) {
      var row = '';
      var empty = 0;
      for (var file = 0; file < 8; file++) {
        final current = squares[ChessSquare(file: file, rank: rank)];
        if (current == null) {
          empty++;
        } else {
          if (empty > 0) row += empty.toString();
          empty = 0;
          row += current.notation;
        }
      }
      if (empty > 0) row += empty.toString();
      rows.add(row);
    }
    return '${rows.join('/')} w - - 0 ${_state.moves.length + 5}';
  }
}
