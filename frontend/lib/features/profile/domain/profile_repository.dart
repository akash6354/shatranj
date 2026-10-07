import 'profile_models.dart';

abstract interface class ProfileRepository {
  Future<Profile> getProfile();
}
