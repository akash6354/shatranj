import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/network_providers.dart';
import '../data/api_subscription_repository.dart';
import '../domain/subscription_models.dart';
import '../domain/subscription_repository.dart';

final subscriptionRepositoryProvider = Provider<SubscriptionRepository>(
  (ref) => ApiSubscriptionRepository(ref.watch(networkApiClientProvider)),
);

final subscriptionControllerProvider =
    AsyncNotifierProvider<SubscriptionController, SubscriptionState>(
      SubscriptionController.new,
    );

class SubscriptionController extends AsyncNotifier<SubscriptionState> {
  late SubscriptionRepository _repository;

  @override
  Future<SubscriptionState> build() async {
    _repository = ref.watch(subscriptionRepositoryProvider);
    return _repository.getSubscription();
  }

  Future<void> selectPlan(String planId) async {
    final current = state.valueOrNull;
    if (current == null) return;
    state = AsyncData(current.copyWith(selectedPlanId: planId, error: null));
  }

  Future<void> restorePurchase() async {
    final current = state.valueOrNull;
    if (current == null) return;
    state = AsyncData(
      current.copyWith(status: SubscriptionStatus.restoring, error: null),
    );
    final restored = await AsyncValue.guard(_repository.restorePurchase);
    state = restored;
  }

  Future<void> startSubscription() async {
    final current = state.valueOrNull;
    if (current == null) return;
    state = AsyncData(
      current.copyWith(
        status: SubscriptionStatus.error,
        error:
            'Payment processing is not available yet. Your plan was not charged.',
      ),
    );
  }
}
