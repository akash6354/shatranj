import 'package:flutter/material.dart';

enum BoardOrientation { white, black }

enum ChessColor { white, black }

enum ChessPieceType { king, queen, rook, bishop, knight, pawn }

enum BoardTheme { classic, forest, midnight }

enum PieceTheme { unicode, minimal }

class ChessPiece {
  const ChessPiece({
    required this.type,
    required this.color,
  });

  final ChessPieceType type;
  final ChessColor color;

  String get notation {
    final symbol = switch (type) {
      ChessPieceType.king => 'k',
      ChessPieceType.queen => 'q',
      ChessPieceType.rook => 'r',
      ChessPieceType.bishop => 'b',
      ChessPieceType.knight => 'n',
      ChessPieceType.pawn => 'p',
    };
    return color == ChessColor.white ? symbol.toUpperCase() : symbol;
  }

  static ChessPiece? fromNotation(String value) {
    final color = value == value.toUpperCase()
        ? ChessColor.white
        : ChessColor.black;
    final type = switch (value.toLowerCase()) {
      'k' => ChessPieceType.king,
      'q' => ChessPieceType.queen,
      'r' => ChessPieceType.rook,
      'b' => ChessPieceType.bishop,
      'n' => ChessPieceType.knight,
      'p' => ChessPieceType.pawn,
      _ => null,
    };
    return type == null ? null : ChessPiece(type: type, color: color);
  }
}

class ChessSquare {
  const ChessSquare({required this.file, required this.rank});

  final int file;
  final int rank;

  String get algebraic =>
      '${String.fromCharCode(97 + file)}${rank + 1}';

  static ChessSquare? parse(String value) {
    if (value.length != 2) return null;
    final file = value.codeUnitAt(0) - 97;
    final rank = int.tryParse(value[1]) ?? 0;
    if (file < 0 || file > 7 || rank < 1 || rank > 8) return null;
    return ChessSquare(file: file, rank: rank - 1);
  }

  @override
  bool operator ==(Object other) =>
      other is ChessSquare && other.file == file && other.rank == rank;

  @override
  int get hashCode => Object.hash(file, rank);
}

class ChessPosition {
  ChessPosition._(this.squares);

  final Map<ChessSquare, ChessPiece> squares;

  factory ChessPosition.fromFen(String fen) {
    final ranks = fen.trim().split(' ').first.split('/');
    final pieces = <ChessSquare, ChessPiece>{};
    if (ranks.length != 8) return ChessPosition._(pieces);

    for (var row = 0; row < 8; row++) {
      var file = 0;
      for (final symbol in ranks[row].split('')) {
        final empty = int.tryParse(symbol);
        if (empty != null) {
          file += empty;
          continue;
        }
        final piece = ChessPiece.fromNotation(symbol);
        if (piece != null && file < 8) {
          pieces[ChessSquare(file: file, rank: 7 - row)] = piece;
        }
        file++;
      }
    }
    return ChessPosition._(pieces);
  }

  ChessPiece? pieceAt(ChessSquare square) => squares[square];
}

class ChessBoardColors {
  const ChessBoardColors({
    required this.light,
    required this.dark,
    required this.selected,
    required this.lastMove,
    required this.check,
    required this.legalMove,
    required this.coordinate,
  });

  final Color light;
  final Color dark;
  final Color selected;
  final Color lastMove;
  final Color check;
  final Color legalMove;
  final Color coordinate;

  factory ChessBoardColors.forTheme(BoardTheme theme) {
    return switch (theme) {
      BoardTheme.classic => const ChessBoardColors(
          light: Color(0xFFE8E3C8),
          dark: Color(0xFF76966A),
          selected: Color(0xB8FFC34D),
          lastMove: Color(0x809FCA68),
          check: Color(0xCCF07867),
          legalMove: Color(0x995C8268),
          coordinate: Color(0xFF416450),
        ),
      BoardTheme.forest => const ChessBoardColors(
          light: Color(0xFFDCE7D2),
          dark: Color(0xFF5F876C),
          selected: Color(0xB8FFC34D),
          lastMove: Color(0x8097C775),
          check: Color(0xCCF07867),
          legalMove: Color(0x99517660),
          coordinate: Color(0xFF315A42),
        ),
      BoardTheme.midnight => const ChessBoardColors(
          light: Color(0xFFB5C5C1),
          dark: Color(0xFF456B68),
          selected: Color(0xB8FFC34D),
          lastMove: Color(0x8095BFA3),
          check: Color(0xCCF07867),
          legalMove: Color(0x996E9B8A),
          coordinate: Color(0xFF31514F),
        ),
    };
  }
}
