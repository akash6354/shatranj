abstract interface class LocalStorage {
  String? readString(String key);
  bool? readBool(String key);
  Future<void> writeString(String key, String value);
  Future<void> writeBool(String key, bool value);
  Future<void> remove(String key);
}

/// Development storage boundary. Replace this implementation with a durable
/// platform adapter when local persistence is added to the app.
class InMemoryLocalStorage implements LocalStorage {
  final _values = <String, Object>{};

  @override
  String? readString(String key) => _values[key] as String?;

  @override
  bool? readBool(String key) => _values[key] as bool?;

  @override
  Future<void> writeString(String key, String value) async {
    _values[key] = value;
  }

  @override
  Future<void> writeBool(String key, bool value) async {
    _values[key] = value;
  }

  @override
  Future<void> remove(String key) async {
    _values.remove(key);
  }
}
