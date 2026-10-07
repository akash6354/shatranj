import 'package:flutter/material.dart';

import 'chess_models.dart';

class ChessPieceWidget extends StatelessWidget {
  const ChessPieceWidget({
    required this.piece,
    this.theme = PieceTheme.unicode,
    this.size,
    super.key,
  });

  final ChessPiece piece;
  final PieceTheme theme;
  final double? size;

  @override
  Widget build(BuildContext context) => FittedBox(
        fit: BoxFit.contain,
        child: Text(
          _glyph,
          style: TextStyle(
            fontSize: size ?? 42,
            height: 1,
            color: piece.color == ChessColor.white
                ? const Color(0xFFF8F2DC)
                : const Color(0xFF17221F),
            shadows: const [
              Shadow(blurRadius: 2, color: Colors.black54),
            ],
          ),
        ),
      );

  String get _glyph {
    if (theme == PieceTheme.minimal) {
      return switch (piece.type) {
        ChessPieceType.king => '♚',
        ChessPieceType.queen => '♛',
        ChessPieceType.rook => '♜',
        ChessPieceType.bishop => '♝',
        ChessPieceType.knight => '♞',
        ChessPieceType.pawn => '♟',
      };
    }
    return switch (piece.notation) {
      'K' => '♔',
      'Q' => '♕',
      'R' => '♖',
      'B' => '♗',
      'N' => '♘',
      'P' => '♙',
      'k' => '♚',
      'q' => '♛',
      'r' => '♜',
      'b' => '♝',
      'n' => '♞',
      _ => '♟',
    };
  }
}
