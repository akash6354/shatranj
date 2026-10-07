import 'package:flutter/material.dart';

import '../../theme/app_theme.dart';
import 'chess_models.dart';
import 'chess_piece.dart';

class CapturedPieces extends StatelessWidget {
  const CapturedPieces({
    required this.pieces,
    this.label,
    this.pieceTheme = PieceTheme.unicode,
    super.key,
  });

  final List<ChessPiece> pieces;
  final String? label;
  final PieceTheme pieceTheme;

  @override
  Widget build(BuildContext context) => Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          if (label != null) ...[
            Text(label!, style: AppTextStyles.caption),
            const SizedBox(width: AppSpacing.xs),
          ],
          if (pieces.isEmpty)
            const SizedBox(width: 2)
          else
            ...pieces.map(
              (piece) => SizedBox(
                width: 20,
                height: 20,
                child: ChessPieceWidget(piece: piece, theme: pieceTheme, size: 22),
              ),
            ),
        ],
      );
}
