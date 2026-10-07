import 'settings_models.dart';

abstract interface class SettingsRepository {
  Future<SettingsState> load();
  Future<void> save(SettingsState settings);
}
