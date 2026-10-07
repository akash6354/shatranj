import 'package:flutter/material.dart';

import '../../theme/app_theme.dart';
import 'chess_models.dart';
import 'chess_piece.dart';

class PromotionDialog extends StatelessWidget {
  const PromotionDialog({
    required this.color,
    this.pieceTheme = PieceTheme.unicode,
    super.key,
  });

  final ChessColor color;
  final PieceTheme pieceTheme;

  static Future<ChessPieceType?> show(
    BuildContext context, {
    required ChessColor color,
    PieceTheme pieceTheme = PieceTheme.unicode,
  }) {
    return showDialog<ChessPieceType>(
      context: context,
      builder: (_) => PromotionDialog(color: color, pieceTheme: pieceTheme),
    );
  }

  @override
  Widget build(BuildContext context) => AlertDialog(
        title: const Text('Choose promotion'),
        content: Wrap(
          spacing: AppSpacing.sm,
          children: [
            for (final type in [
              ChessPieceType.queen,
              ChessPieceType.rook,
              ChessPieceType.bishop,
              ChessPieceType.knight,
            ])
              IconButton(
                onPressed: () => Navigator.of(context).pop(type),
                tooltip: type.name,
                icon: ChessPieceWidget(
                  piece: ChessPiece(type: type, color: color),
                  theme: pieceTheme,
                  size: 42,
                ),
              ),
          ],
        ),
      );
}
