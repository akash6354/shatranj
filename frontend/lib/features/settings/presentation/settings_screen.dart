import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/app_widgets.dart';
import '../../../core/widgets/chess/chess_models.dart';
import '../../auth/presentation/auth_controller.dart';
import '../domain/settings_models.dart';
import 'settings_controller.dart';

class SettingsScreen extends ConsumerWidget {
  const SettingsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final settings = ref.watch(settingsControllerProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Settings')),
      body: SafeArea(
        child: settings.when(
          loading: () => const AppLoading(message: 'Loading settings'),
          error: (error, _) => AppError(message: error.toString()),
          data: (value) => _SettingsContent(settings: value),
        ),
      ),
    );
  }
}

class _SettingsContent extends ConsumerWidget {
  const _SettingsContent({required this.settings});
  final SettingsState settings;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final controller = ref.read(settingsControllerProvider.notifier);
    return ListView(
      padding: const EdgeInsets.fromLTRB(
        AppSpacing.page,
        AppSpacing.sm,
        AppSpacing.page,
        AppSpacing.xl,
      ),
      children: [
        const _SectionTitle(title: 'Appearance'),
        AppCard(
          child: Column(
            children: [
              _ChoiceRow<AppThemeMode>(
                title: 'Theme',
                value: settings.themeMode,
                values: AppThemeMode.values,
                label: (value) => value == AppThemeMode.dark ? 'Dark' : 'System',
                onChanged: controller.setTheme,
              ),
              _ChoiceRow<BoardTheme>(
                title: 'Board theme',
                value: settings.boardTheme,
                values: BoardTheme.values,
                label: (value) => _title(value.name),
                onChanged: controller.setBoard,
              ),
              _ChoiceRow<PieceTheme>(
                title: 'Piece style',
                value: settings.pieceTheme,
                values: PieceTheme.values,
                label: (value) => _title(value.name),
                onChanged: controller.setPieces,
              ),
            ],
          ),
        ),
        const SizedBox(height: AppSpacing.lg),
        const _SectionTitle(title: 'Game'),
        AppCard(
          child: Column(
            children: [
              _SwitchRow(title: 'Show legal moves', value: settings.showLegalMoves, onChanged: controller.setLegalMoves),
              _SwitchRow(title: 'Show move suggestions', value: settings.showMoveSuggestions, onChanged: controller.setSuggestions),
              _SwitchRow(title: 'Sound', value: settings.soundEnabled, onChanged: controller.setSound),
              _SwitchRow(title: 'Vibration', value: settings.vibrationEnabled, onChanged: controller.setVibration),
            ],
          ),
        ),
        const SizedBox(height: AppSpacing.lg),
        const _SectionTitle(title: 'Account'),
        AppCard(
          child: Column(
            children: [
              _ActionRow(title: 'Edit profile', icon: Icons.edit_outlined, onTap: () => _showUnavailable(context, 'Profile editing will be connected to the account API.')),
              _ActionRow(title: 'Email & security', icon: Icons.lock_outline_rounded, onTap: () => _showUnavailable(context, 'Security settings will be connected to the account API.')),
              _ActionRow(title: 'Language', icon: Icons.language_rounded, onTap: () => _showLanguage(context)),
              _ActionRow(
                title: 'Log out',
                icon: Icons.logout_rounded,
                destructive: true,
                onTap: () async {
                  await ref.read(authControllerProvider.notifier).logout();
                  if (context.mounted) context.go('/auth/login');
                },
              ),
            ],
          ),
        ),
      ],
    );
  }

  void _showLanguage(BuildContext context) => showModalBottomSheet<void>(
        context: context,
        builder: (context) => SafeArea(
          child: ListTile(
            title: const Text('English'),
            trailing: const Icon(Icons.check_rounded, color: AppColors.gold),
            onTap: () => context.pop(),
          ),
        ),
      );

  void _showUnavailable(BuildContext context, String message) {
    ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(message)));
  }
}

class _SectionTitle extends StatelessWidget {
  const _SectionTitle({required this.title});
  final String title;
  @override
  Widget build(BuildContext context) => Padding(
        padding: const EdgeInsets.only(bottom: AppSpacing.sm),
        child: Text(title, style: AppTextStyles.titleSmall),
      );
}

class _SwitchRow extends StatelessWidget {
  const _SwitchRow({required this.title, required this.value, required this.onChanged});
  final String title;
  final bool value;
  final ValueChanged<bool> onChanged;
  @override
  Widget build(BuildContext context) => SwitchListTile.adaptive(
        contentPadding: EdgeInsets.zero,
        title: Text(title, style: AppTextStyles.body),
        value: value,
        onChanged: onChanged,
      );
}

class _ChoiceRow<T> extends StatelessWidget {
  const _ChoiceRow({
    required this.title,
    required this.value,
    required this.values,
    required this.label,
    required this.onChanged,
  });
  final String title;
  final T value;
  final List<T> values;
  final String Function(T value) label;
  final ValueChanged<T> onChanged;
  @override
  Widget build(BuildContext context) => ListTile(
        contentPadding: EdgeInsets.zero,
        title: Text(title, style: AppTextStyles.body),
        trailing: DropdownButton<T>(
          value: value,
          underline: const SizedBox.shrink(),
          items: values.map((item) => DropdownMenuItem<T>(value: item, child: Text(label(item)))).toList(),
          onChanged: (next) {
            if (next != null) onChanged(next);
          },
        ),
      );
}

class _ActionRow extends StatelessWidget {
  const _ActionRow({required this.title, required this.icon, required this.onTap, this.destructive = false});
  final String title;
  final IconData icon;
  final VoidCallback onTap;
  final bool destructive;
  @override
  Widget build(BuildContext context) => ListTile(
        contentPadding: EdgeInsets.zero,
        leading: Icon(icon, color: destructive ? AppColors.danger : AppColors.textSecondary),
        title: Text(title, style: AppTextStyles.body.copyWith(color: destructive ? AppColors.danger : null)),
        trailing: const Icon(Icons.chevron_right_rounded, color: AppColors.muted),
        onTap: onTap,
      );
}

String _title(String value) => '${value[0].toUpperCase()}${value.substring(1)}';
