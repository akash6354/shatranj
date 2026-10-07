import 'package:flutter/material.dart';

import 'chess_models.dart';
import 'chess_piece.dart';
import 'board_coordinates.dart';

class ChessSquareWidget extends StatelessWidget {
  const ChessSquareWidget({
    required this.square,
    required this.piece,
    required this.colors,
    required this.showCoordinate,
    this.isSelected = false,
    this.isLastMove = false,
    this.isCheck = false,
    this.isLegalMove = false,
    this.onTap,
    this.pieceTheme = PieceTheme.unicode,
    super.key,
  });

  final ChessSquare square;
  final ChessPiece? piece;
  final ChessBoardColors colors;
  final bool showCoordinate;
  final bool isSelected;
  final bool isLastMove;
  final bool isCheck;
  final bool isLegalMove;
  final VoidCallback? onTap;
  final PieceTheme pieceTheme;

  @override
  Widget build(BuildContext context) {
    final base = (square.file + square.rank).isEven
        ? colors.light
        : colors.dark;
    final background = isCheck
        ? colors.check
        : isSelected
            ? colors.selected
            : isLastMove
                ? colors.lastMove
                : base;

    return GestureDetector(
      onTap: onTap,
      child: ColoredBox(
        color: background,
        child: Stack(
          fit: StackFit.expand,
          children: [
            if (isLegalMove)
              Center(
                child: Container(
                  width: piece == null ? 18 : double.infinity,
                  height: piece == null ? 18 : double.infinity,
                  decoration: BoxDecoration(
                    color: piece == null
                        ? colors.legalMove
                        : Colors.transparent,
                    shape: piece == null ? BoxShape.circle : BoxShape.rectangle,
                    border: piece == null
                        ? null
                        : Border.all(color: colors.legalMove, width: 5),
                  ),
                ),
              ),
            if (piece != null)
              Padding(
                padding: const EdgeInsets.all(3),
                child: ChessPieceWidget(piece: piece!, theme: pieceTheme),
              ),
            if (showCoordinate)
              Positioned(
                left: 3,
                bottom: 2,
                child: BoardCoordinates(
                  square: square,
                  color: colors.coordinate,
                ),
              ),
          ],
        ),
      ),
    );
  }
}
