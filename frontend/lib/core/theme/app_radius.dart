import 'package:flutter/material.dart';

abstract final class AppRadius {
  static const sm = 10.0;
  static const button = 14.0;
  static const card = 18.0;
  static const dialog = 24.0;
  static const pill = 100.0;

  static final small = BorderRadius.circular(sm);
  static final buttonRadius = BorderRadius.circular(button);
  static final cardRadius = BorderRadius.circular(card);
  static final dialogRadius = BorderRadius.circular(dialog);
  static final pillRadius = BorderRadius.circular(pill);
}
