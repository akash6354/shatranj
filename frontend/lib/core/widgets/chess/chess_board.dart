import 'package:flutter/material.dart';

import 'chess_models.dart';
import 'chess_square.dart';

class ChessBoard extends StatelessWidget {
  const ChessBoard({
    required this.fen,
    this.orientation = BoardOrientation.white,
    this.boardTheme = BoardTheme.classic,
    this.pieceTheme = PieceTheme.unicode,
    this.selectedSquare,
    this.lastMove,
    this.legalMoves = const {},
    this.checkSquare,
    this.showCoordinates = true,
    this.onSquareTap,
    this.borderRadius = 12,
    super.key,
  });

  final String fen;
  final BoardOrientation orientation;
  final BoardTheme boardTheme;
  final PieceTheme pieceTheme;
  final String? selectedSquare;
  final ({String from, String to})? lastMove;
  final Set<String> legalMoves;
  final String? checkSquare;
  final bool showCoordinates;
  final ValueChanged<String>? onSquareTap;
  final double borderRadius;

  @override
  Widget build(BuildContext context) {
    final position = ChessPosition.fromFen(fen);
    final colors = ChessBoardColors.forTheme(boardTheme);
    final files = orientation == BoardOrientation.white
        ? List.generate(8, (index) => index)
        : List.generate(8, (index) => 7 - index);
    final ranks = orientation == BoardOrientation.white
        ? List.generate(8, (index) => 7 - index)
        : List.generate(8, (index) => index);

    return AspectRatio(
      aspectRatio: 1,
      child: ClipRRect(
        borderRadius: BorderRadius.circular(borderRadius),
        child: GridView.builder(
          physics: const NeverScrollableScrollPhysics(),
          gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
            crossAxisCount: 8,
          ),
          itemCount: 64,
          itemBuilder: (context, index) {
            final square = ChessSquare(
              file: files[index % 8],
              rank: ranks[index ~/ 8],
            );
            final algebraic = square.algebraic;
            final isLastMove = lastMove?.from == algebraic || lastMove?.to == algebraic;
            return ChessSquareWidget(
              square: square,
              piece: position.pieceAt(square),
              colors: colors,
              pieceTheme: pieceTheme,
              showCoordinate: showCoordinates &&
                  (index >= 56 || index % 8 == 0),
              isSelected: selectedSquare == algebraic,
              isLastMove: isLastMove,
              isCheck: checkSquare == algebraic,
              isLegalMove: legalMoves.contains(algebraic),
              onTap: onSquareTap == null ? null : () => onSquareTap!(algebraic),
            );
          },
        ),
      ),
    );
  }
}
