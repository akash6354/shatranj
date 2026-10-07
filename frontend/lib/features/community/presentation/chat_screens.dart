import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/app_widgets.dart';
import '../domain/community_models.dart';
import 'community_controller.dart';

class ChatRoomsScreen extends ConsumerWidget {
  const ChatRoomsScreen({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final rooms = ref.watch(chatRoomsProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Chat')),
      body: SafeArea(child: rooms.when(
        loading: () => const AppLoading(),
        error: (error, _) => AppError(message: error.toString()),
        data: (items) => ListView(padding: const EdgeInsets.all(AppSpacing.page), children: items.map((room) => AppCard(
          margin: const EdgeInsets.only(bottom: AppSpacing.sm),
          onTap: () => Navigator.of(context).push(MaterialPageRoute<void>(builder: (_) => ChatScreen(room: room))),
          child: Row(children: [
            Stack(children: [
              CircleAvatar(child: Icon(room.kind == 'club' ? Icons.groups : Icons.person)),
              if (room.online) Positioned(right: 0, bottom: 0, child: Container(width: 10, height: 10, decoration: const BoxDecoration(color: AppColors.success, shape: BoxShape.circle))),
            ]),
            const SizedBox(width: AppSpacing.md),
            Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
              Text(room.name, style: AppTextStyles.bodyStrong),
              const SizedBox(height: AppSpacing.xs),
              Text(room.lastMessage, style: AppTextStyles.caption, maxLines: 1, overflow: TextOverflow.ellipsis),
            ])),
            Text('${room.updatedAt.hour}:${room.updatedAt.minute.toString().padLeft(2, '0')}', style: AppTextStyles.caption),
          ]),
        )).toList()),
      )),
    );
  }
}

class ChatScreen extends ConsumerStatefulWidget {
  const ChatScreen({required this.room, super.key});
  final ChatRoom room;
  @override
  ConsumerState<ChatScreen> createState() => _ChatScreenState();
}

class _ChatScreenState extends ConsumerState<ChatScreen> {
  final _controller = TextEditingController();
  @override
  void dispose() { _controller.dispose(); super.dispose(); }
  @override
  Widget build(BuildContext context) {
    final messages = ref.watch(chatMessagesProvider(widget.room.id));
    return Scaffold(
      appBar: AppBar(title: Text(widget.room.name), actions: [if (widget.room.online) const Padding(padding: EdgeInsets.only(right: AppSpacing.md), child: Center(child: Text('Online', style: TextStyle(color: AppColors.success))))]),
      body: SafeArea(child: Column(children: [
        Expanded(child: messages.when(
          loading: () => const AppLoading(),
          error: (error, _) => AppError(message: error.toString()),
          data: (items) => ListView.builder(
            padding: const EdgeInsets.all(AppSpacing.md),
            itemCount: items.length,
            itemBuilder: (context, index) => _MessageBubble(message: items[index]),
          ),
        )),
        Padding(
          padding: const EdgeInsets.fromLTRB(AppSpacing.md, AppSpacing.sm, AppSpacing.md, AppSpacing.md),
          child: Row(children: [
            Expanded(child: TextField(controller: _controller, textInputAction: TextInputAction.send, onSubmitted: (_) => _send(), decoration: const InputDecoration(hintText: 'Write a message'))),
            const SizedBox(width: AppSpacing.sm),
            AppIconButton(icon: Icons.send_rounded, tooltip: 'Send message', onPressed: _send),
          ]),
        ),
      ])),
    );
  }

  Future<void> _send() async {
    final content = _controller.text.trim();
    if (content.isEmpty) return;
    await ref.read(chatRepositoryProvider).send(widget.room.id, content);
    _controller.clear();
    ref.invalidate(chatMessagesProvider(widget.room.id));
  }
}

class _MessageBubble extends StatelessWidget {
  const _MessageBubble({required this.message});
  final ChatMessage message;
  @override
  Widget build(BuildContext context) => Align(
        alignment: message.isMine ? Alignment.centerRight : Alignment.centerLeft,
        child: Container(
          constraints: const BoxConstraints(maxWidth: 300),
          margin: const EdgeInsets.only(bottom: AppSpacing.sm),
          padding: const EdgeInsets.symmetric(horizontal: AppSpacing.md, vertical: AppSpacing.sm),
          decoration: BoxDecoration(color: message.isMine ? AppColors.gold.withValues(alpha: .2) : AppColors.surface, borderRadius: AppRadius.cardRadius, border: Border.all(color: message.isMine ? AppColors.gold.withValues(alpha: .4) : AppColors.border)),
          child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
            if (!message.isMine) Text(message.username, style: AppTextStyles.caption.copyWith(color: AppColors.mint)),
            Text(message.content, style: AppTextStyles.body),
            const SizedBox(height: AppSpacing.xs),
            Text('${message.createdAt.hour}:${message.createdAt.minute.toString().padLeft(2, '0')}', style: AppTextStyles.caption),
          ]),
        ),
      );
}
