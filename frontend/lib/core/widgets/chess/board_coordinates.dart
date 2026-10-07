import 'package:flutter/material.dart';

import 'chess_models.dart';

class BoardCoordinates extends StatelessWidget {
  const BoardCoordinates({
    required this.square,
    required this.color,
    super.key,
  });

  final ChessSquare square;
  final Color color;

  @override
  Widget build(BuildContext context) => Text(
        square.algebraic,
        style: TextStyle(
          fontSize: 8,
          fontWeight: FontWeight.w700,
          color: color,
        ),
      );
}
