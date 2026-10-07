enum GameConnectionStatus { disconnected, connecting, connected, reconnecting, failed }

enum GameStatus { waiting, active, finished }

enum GameResult { ongoing, whiteWin, blackWin, draw }

class GamePlayer {
  const GamePlayer({
    required this.id,
    required this.name,
    required this.rating,
    required this.color,
  });

  final String id;
  final String name;
  final int rating;
  final String color;
}

class GameMove {
  const GameMove({
    required this.number,
    required this.playerId,
    required this.from,
    required this.to,
    this.san,
  });

  final int number;
  final String playerId;
  final String from;
  final String to;
  final String? san;
}

class GameState {
  const GameState({
    required this.id,
    required this.fen,
    required this.white,
    required this.black,
    required this.whiteTime,
    required this.blackTime,
    required this.turn,
    required this.status,
    required this.result,
    required this.moves,
    required this.connection,
    this.drawOfferBy,
    this.endReason,
    this.selectedSquare,
    this.lastMove,
  });

  final String id;
  final String fen;
  final GamePlayer white;
  final GamePlayer black;
  final Duration whiteTime;
  final Duration blackTime;
  final String turn;
  final GameStatus status;
  final GameResult result;
  final List<GameMove> moves;
  final GameConnectionStatus connection;
  final String? drawOfferBy;
  final String? endReason;
  final String? selectedSquare;
  final GameMove? lastMove;

  bool get isFinished => status == GameStatus.finished;

  GameState copyWith({
    String? fen,
    Duration? whiteTime,
    Duration? blackTime,
    String? turn,
    GameStatus? status,
    GameResult? result,
    List<GameMove>? moves,
    GameConnectionStatus? connection,
    String? drawOfferBy,
    String? endReason,
    String? selectedSquare,
    GameMove? lastMove,
  }) =>
      GameState(
        id: id,
        fen: fen ?? this.fen,
        white: white,
        black: black,
        whiteTime: whiteTime ?? this.whiteTime,
        blackTime: blackTime ?? this.blackTime,
        turn: turn ?? this.turn,
        status: status ?? this.status,
        result: result ?? this.result,
        moves: moves ?? this.moves,
        connection: connection ?? this.connection,
        drawOfferBy: drawOfferBy ?? this.drawOfferBy,
        endReason: endReason ?? this.endReason,
        selectedSquare: selectedSquare ?? this.selectedSquare,
        lastMove: lastMove ?? this.lastMove,
      );
}

class GameActionException implements Exception {
  const GameActionException(this.message);

  final String message;

  @override
  String toString() => message;
}
