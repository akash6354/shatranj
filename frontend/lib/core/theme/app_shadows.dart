import 'package:flutter/material.dart';

import 'app_colors.dart';

abstract final class AppShadows {
  static const card = <BoxShadow>[
    BoxShadow(
      color: Color(0x26000000),
      blurRadius: 18,
      offset: Offset(0, 8),
    ),
  ];
  static const elevated = <BoxShadow>[
    BoxShadow(
      color: Color(0x40000000),
      blurRadius: 26,
      offset: Offset(0, 12),
    ),
  ];
  static const goldGlow = <BoxShadow>[
    BoxShadow(
      color: Color(0x55FFC34D),
      blurRadius: 18,
      spreadRadius: -4,
    ),
  ];
  static const mintGlow = <BoxShadow>[
    BoxShadow(
      color: Color(0x4080D7B9),
      blurRadius: 18,
      spreadRadius: -5,
    ),
  ];

  static const focusBorder = BorderSide(color: AppColors.gold, width: 1.2);
}
