import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/app_widgets.dart';
import '../domain/play_models.dart';
import 'play_controller.dart';

class FriendGameScreen extends ConsumerStatefulWidget {
  const FriendGameScreen({super.key});

  @override
  ConsumerState<FriendGameScreen> createState() => _FriendGameScreenState();
}

class _FriendGameScreenState extends ConsumerState<FriendGameScreen> {
  final _codeController = TextEditingController();
  TimeControl _timeControl = const TimeControl(10, 5);
  String? _roomCode;

  @override
  void dispose() {
    _codeController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => Scaffold(
        appBar: AppBar(title: const Text('Play Friend')),
        body: SafeArea(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(AppSpacing.page),
            child: Center(
              child: ConstrainedBox(
                constraints: const BoxConstraints(maxWidth: 560),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text('Private game', style: AppTextStyles.headline),
                    const SizedBox(height: AppSpacing.xs),
                    const Text('Share a room code with a friend or join their room.', style: AppTextStyles.body),
                    const SizedBox(height: AppSpacing.lg),
                    const Text('Time control', style: AppTextStyles.titleSmall),
                    const SizedBox(height: AppSpacing.sm),
                    Wrap(
                      spacing: AppSpacing.sm,
                      children: _friendControls
                          .map((value) => AppChoicePill(
                                label: value.label,
                                selected: _timeControl.label == value.label,
                                onSelected: (_) => setState(() => _timeControl = value),
                              ))
                          .toList(),
                    ),
                    const SizedBox(height: AppSpacing.lg),
                    if (_roomCode != null)
                      AppGradientCard(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            const Text('Room created', style: AppTextStyles.titleSmall),
                            const SizedBox(height: AppSpacing.sm),
                            Row(
                              children: [
                                Expanded(child: Text(_roomCode!, style: AppTextStyles.numeric)),
                                AppIconButton(
                                  icon: Icons.copy_rounded,
                                  tooltip: 'Copy room code',
                                  onPressed: () => _copyRoomCode(context),
                                ),
                              ],
                            ),
                            const SizedBox(height: AppSpacing.xs),
                            const Text('Share this code with your friend.', style: AppTextStyles.caption),
                          ],
                        ),
                      )
                    else
                      AppButton(
                        label: 'Create room',
                        icon: Icons.add_link_rounded,
                        expand: true,
                        onPressed: () async {
                          final code = await ref.read(friendGameRepositoryProvider).createRoom(_timeControl);
                          if (mounted) setState(() => _roomCode = code);
                        },
                      ),
                    const SizedBox(height: AppSpacing.lg),
                    const Text('Join a room', style: AppTextStyles.titleSmall),
                    const SizedBox(height: AppSpacing.sm),
                    TextField(
                      controller: _codeController,
                      textCapitalization: TextCapitalization.characters,
                      decoration: const InputDecoration(hintText: 'Enter room code'),
                    ),
                    const SizedBox(height: AppSpacing.sm),
                    AppButton(
                      label: 'Join room',
                      variant: AppButtonVariant.outline,
                      expand: true,
                      onPressed: () async {
                        if (_codeController.text.trim().isEmpty) return;
                        final id = await ref.read(friendGameRepositoryProvider).joinRoom(_codeController.text);
                        if (context.mounted) context.go('/game/friend-$id');
                      },
                    ),
                  ],
                ),
              ),
            ),
          ),
        ),
      );

  void _copyRoomCode(BuildContext context) {
    // The room-code UI is intentionally platform-neutral until a clipboard
    // integration is introduced.
    ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Room code ready to share')));
  }
}

const _friendControls = [TimeControl(3, 2), TimeControl(5, 0), TimeControl(10, 5)];
