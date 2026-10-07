import 'dart:async';

import 'package:flutter/material.dart';

import '../../theme/app_theme.dart';

class ChessClock extends StatefulWidget {
  const ChessClock({
    required this.initialDuration,
    this.isRunning = false,
    this.onExpired,
    this.label,
    this.compact = false,
    super.key,
  });

  final Duration initialDuration;
  final bool isRunning;
  final VoidCallback? onExpired;
  final String? label;
  final bool compact;

  @override
  State<ChessClock> createState() => _ChessClockState();
}

class _ChessClockState extends State<ChessClock> {
  late Duration _remaining = widget.initialDuration;
  Timer? _timer;

  @override
  void initState() {
    super.initState();
    _syncTimer();
  }

  @override
  void didUpdateWidget(covariant ChessClock oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.initialDuration != widget.initialDuration) {
      _remaining = widget.initialDuration;
    }
    _syncTimer();
  }

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }

  void _syncTimer() {
    _timer?.cancel();
    if (widget.isRunning && _remaining > Duration.zero) {
      _timer = Timer.periodic(const Duration(seconds: 1), (_) {
        if (_remaining <= const Duration(seconds: 1)) {
          setState(() => _remaining = Duration.zero);
          _timer?.cancel();
          widget.onExpired?.call();
        } else {
          setState(() => _remaining -= const Duration(seconds: 1));
        }
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final urgent = _remaining <= const Duration(seconds: 10);
    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.end,
      children: [
        if (widget.label != null)
          Text(widget.label!, style: AppTextStyles.caption),
        Container(
          padding: EdgeInsets.symmetric(
            horizontal: widget.compact ? 8 : 12,
            vertical: widget.compact ? 5 : 8,
          ),
          decoration: BoxDecoration(
            color: urgent
                ? AppColors.danger.withValues(alpha: .15)
                : AppColors.surfaceElevated,
            borderRadius: AppRadius.small,
            border: Border.all(
              color: urgent ? AppColors.danger : AppColors.border,
            ),
          ),
          child: Text(
            _format(_remaining),
            style: AppTextStyles.titleSmall.copyWith(
              color: urgent ? AppColors.dangerSoft : AppColors.text,
              fontFeatures: const [FontFeature.tabularFigures()],
            ),
          ),
        ),
      ],
    );
  }

  String _format(Duration duration) {
    final minutes = duration.inMinutes.remainder(60).toString().padLeft(2, '0');
    final seconds = duration.inSeconds.remainder(60).toString().padLeft(2, '0');
    return '${duration.inHours > 0 ? '${duration.inHours}:' : ''}$minutes:$seconds';
  }
}
