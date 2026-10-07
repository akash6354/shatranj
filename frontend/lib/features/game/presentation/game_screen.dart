import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/chess/chess_widgets.dart';
import '../../../core/widgets/app_widgets.dart';
import '../domain/game_models.dart';
import 'game_controller.dart';

class GameScreen extends ConsumerStatefulWidget {
  const GameScreen({required this.gameId, super.key});

  final String gameId;

  @override
  ConsumerState<GameScreen> createState() => _GameScreenState();
}

class _GameScreenState extends ConsumerState<GameScreen> {
  String? _selectedSquare;

  @override
  Widget build(BuildContext context) {
    final game = ref.watch(gameControllerProvider(widget.gameId));
    return Scaffold(
      appBar: AppBar(
        title: const Text('Online game'),
        actions: [
          AppIconButton(
            icon: Icons.more_horiz_rounded,
            tooltip: 'Game options',
            onPressed: () => _showOptions(context),
          ),
        ],
      ),
      body: SafeArea(
        child: game.when(
          loading: () => const AppLoading(message: 'Connecting to game'),
          error: (error, _) => _GameError(
            message: error.toString(),
            onReconnect: () => ref
                .read(gameControllerProvider(widget.gameId).notifier)
                .reconnect(),
          ),
          data: (state) => _GameContent(
            state: state,
            selectedSquare: _selectedSquare,
            onSquareTap: _onSquareTap,
            onDraw: () => _offerDraw(state),
            onResign: () => _confirmResign(state),
            onReconnect: () => ref
                .read(gameControllerProvider(widget.gameId).notifier)
                .reconnect(),
          ),
        ),
      ),
    );
  }

  void _onSquareTap(String square) {
    final state = ref.read(gameControllerProvider(widget.gameId)).valueOrNull;
    if (state == null || state.isFinished || state.turn != state.white.id) return;
    if (_selectedSquare == null) {
      setState(() => _selectedSquare = square);
    } else {
      final from = _selectedSquare!;
      setState(() => _selectedSquare = null);
      ref
          .read(gameControllerProvider(widget.gameId).notifier)
          .sendMove(from, square);
    }
  }

  Future<void> _offerDraw(GameState state) async {
    await ref.read(gameControllerProvider(widget.gameId).notifier).offerDraw();
  }

  Future<void> _confirmResign(GameState state) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Resign game?'),
        content: const Text('This action will end the game for you.'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('Cancel')),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            style: FilledButton.styleFrom(backgroundColor: AppColors.danger),
            child: const Text('Resign'),
          ),
        ],
      ),
    );
    if (confirmed == true && mounted) {
      await ref.read(gameControllerProvider(widget.gameId).notifier).resign();
    }
  }

  void _showOptions(BuildContext context) {
    showModalBottomSheet<void>(
      context: context,
      builder: (context) => SafeArea(
        child: ListTile(
          leading: const Icon(Icons.flag_outlined),
          title: const Text('Resign game'),
          onTap: () {
            Navigator.pop(context);
            final state = ref.read(gameControllerProvider(widget.gameId)).valueOrNull;
            if (state != null) _confirmResign(state);
          },
        ),
      ),
    );
  }
}

class _GameContent extends StatelessWidget {
  const _GameContent({
    required this.state,
    required this.selectedSquare,
    required this.onSquareTap,
    required this.onDraw,
    required this.onResign,
    required this.onReconnect,
  });

  final GameState state;
  final String? selectedSquare;
  final ValueChanged<String> onSquareTap;
  final VoidCallback onDraw;
  final VoidCallback onResign;
  final VoidCallback onReconnect;

  @override
  Widget build(BuildContext context) => LayoutBuilder(
        builder: (context, constraints) {
          final wide = constraints.maxWidth >= 760;
          final board = Column(
            children: [
              _PlayerBar(
                player: state.black,
                clock: ChessClock(
                  initialDuration: state.blackTime,
                  isRunning: state.turn == state.black.id && !state.isFinished,
                  label: 'Opponent',
                ),
              ),
              const SizedBox(height: AppSpacing.sm),
              ChessBoard(
                fen: state.fen,
                selectedSquare: selectedSquare,
                lastMove: state.lastMove == null
                    ? null
                    : (
                        from: state.lastMove!.from,
                        to: state.lastMove!.to,
                      ),
                onSquareTap: onSquareTap,
              ),
              const SizedBox(height: AppSpacing.sm),
              _PlayerBar(
                player: state.white,
                clock: ChessClock(
                  initialDuration: state.whiteTime,
                  isRunning: state.turn == state.white.id && !state.isFinished,
                  label: 'You',
                ),
              ),
            ],
          );

          return SingleChildScrollView(
            padding: const EdgeInsets.fromLTRB(AppSpacing.page, AppSpacing.sm, AppSpacing.page, AppSpacing.xl),
            child: Center(
              child: ConstrainedBox(
                constraints: const BoxConstraints(maxWidth: 980),
                child: wide
                    ? Row(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Expanded(flex: 3, child: board),
                          const SizedBox(width: AppSpacing.lg),
                          Expanded(flex: 2, child: _GameSidePanel(state: state, onDraw: onDraw, onResign: onResign, onReconnect: onReconnect)),
                        ],
                      )
                    : Column(
                        children: [
                          board,
                          const SizedBox(height: AppSpacing.lg),
                          _GameSidePanel(state: state, onDraw: onDraw, onResign: onResign, onReconnect: onReconnect),
                        ],
                      ),
              ),
            ),
          );
        },
      );
}

class _PlayerBar extends StatelessWidget {
  const _PlayerBar({required this.player, required this.clock});

  final GamePlayer player;
  final Widget clock;

  @override
  Widget build(BuildContext context) => Row(
        children: [
          const AppAvatar(initials: 'P', size: 40),
          const SizedBox(width: AppSpacing.sm),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(player.name, style: AppTextStyles.bodyStrong),
                Text('${player.rating} • ${player.color}', style: AppTextStyles.caption),
              ],
            ),
          ),
          clock,
        ],
      );
}

class _GameSidePanel extends StatelessWidget {
  const _GameSidePanel({
    required this.state,
    required this.onDraw,
    required this.onResign,
    required this.onReconnect,
  });

  final GameState state;
  final VoidCallback onDraw;
  final VoidCallback onResign;
  final VoidCallback onReconnect;

  @override
  Widget build(BuildContext context) => Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          _ConnectionBanner(state: state, onReconnect: onReconnect),
          if (state.isFinished)
            AppCard(
              margin: const EdgeInsets.only(top: AppSpacing.md),
              child: Column(
                children: [
                  const Icon(Icons.emoji_events_rounded, color: AppColors.gold, size: 36),
                  const SizedBox(height: AppSpacing.sm),
                  const Text('Game complete', style: AppTextStyles.title),
                  const SizedBox(height: AppSpacing.xs),
                  Text(state.endReason ?? 'The game has ended.', style: AppTextStyles.caption),
                ],
              ),
            ),
          AppCard(
            margin: const EdgeInsets.only(top: AppSpacing.md),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const AppSectionHeader(title: 'Moves'),
                const SizedBox(height: AppSpacing.sm),
                SizedBox(
                  height: 180,
                  child: state.moves.isEmpty
                      ? const AppEmptyState(title: 'No moves yet', message: 'Select a source and destination square.')
                      : ListView.builder(
                          itemCount: state.moves.length,
                          itemBuilder: (context, index) {
                            final move = state.moves[index];
                            return ListTile(
                              dense: true,
                              leading: Text('${move.number}.', style: AppTextStyles.caption),
                              title: Text('${move.from} → ${move.to}', style: AppTextStyles.bodyStrong),
                              trailing: Text(move.playerId == state.white.id ? 'White' : 'Black', style: AppTextStyles.caption),
                            );
                          },
                        ),
                ),
              ],
            ),
          ),
          if (state.drawOfferBy != null && state.drawOfferBy!.isNotEmpty)
            AppCard(
              margin: const EdgeInsets.only(top: AppSpacing.md),
              child: Row(
                children: [
                  const Icon(Icons.handshake_outlined, color: AppColors.gold),
                  const SizedBox(width: AppSpacing.sm),
                  const Expanded(child: Text('Draw offer pending', style: AppTextStyles.bodyStrong)),
                  TextButton(onPressed: () {}, child: const Text('Review')),
                ],
              ),
            ),
          if (!state.isFinished)
            Padding(
              padding: const EdgeInsets.only(top: AppSpacing.md),
              child: Row(
                children: [
                  Expanded(child: AppButton(label: 'Offer draw', variant: AppButtonVariant.outline, onPressed: onDraw)),
                  const SizedBox(width: AppSpacing.sm),
                  Expanded(child: AppButton(label: 'Resign', variant: AppButtonVariant.danger, onPressed: onResign)),
                ],
              ),
            ),
        ],
      );
}

class _ConnectionBanner extends StatelessWidget {
  const _ConnectionBanner({required this.state, required this.onReconnect});

  final GameState state;
  final VoidCallback onReconnect;

  @override
  Widget build(BuildContext context) {
    final connected = state.connection == GameConnectionStatus.connected;
    final connecting = state.connection == GameConnectionStatus.connecting ||
        state.connection == GameConnectionStatus.reconnecting;
    return AppPill(
      label: connected
          ? 'Connected'
          : connecting
              ? 'Reconnecting…'
              : 'Offline',
      icon: connected ? Icons.wifi_rounded : Icons.wifi_off_rounded,
      color: connected ? AppColors.success : AppColors.warning,
      onTap: connected ? null : onReconnect,
    );
  }
}

class _GameError extends StatelessWidget {
  const _GameError({required this.message, required this.onReconnect});

  final String message;
  final VoidCallback onReconnect;

  @override
  Widget build(BuildContext context) => AppError(
        title: 'Game connection failed',
        message: message,
        onRetry: onReconnect,
      );
}
