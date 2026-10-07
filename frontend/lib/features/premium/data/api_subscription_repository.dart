import '../../../core/network/api_client.dart';
import '../domain/subscription_models.dart';
import '../domain/subscription_repository.dart';

class ApiSubscriptionRepository implements SubscriptionRepository {
  ApiSubscriptionRepository(this._client);
  final ApiClient _client;

  @override
  Future<List<SubscriptionPlan>> getPlans() async {
    final response = await _client.get('/subscriptions/plans');
    final plans = response['data'];
    if (plans is! List) return const [];
    return plans
        .whereType<Map>()
        .map((item) {
          final data = Map<String, dynamic>.from(item);
          return SubscriptionPlan(
            id: data['id']?.toString() ?? '',
            name: data['name']?.toString() ?? '',
            pricePaise: (data['price_paise'] as num?)?.toInt() ?? 0,
            currency: data['currency']?.toString() ?? '',
            interval: data['interval']?.toString() ?? '',
            entitlements: data['entitlements'] is List
                ? (data['entitlements'] as List)
                    .map((value) => value.toString())
                    .toList()
                : const [],
          );
        })
        .toList();
  }

  @override
  Future<SubscriptionState> getSubscription() async {
    final plans = await getPlans();
    try {
      final data = await _client.get('/subscriptions/me');
      return _state(plans, data);
    } on ApiException catch (error) {
      if (error.statusCode == 404) {
        return SubscriptionState.initial(plans);
      }
      rethrow;
    }
  }

  @override
  Future<SubscriptionState> restorePurchase() => getSubscription();

  @override
  Future<SubscriptionState> selectPlan(String planId) async {
    final state = await getSubscription();
    return state.copyWith(selectedPlanId: planId);
  }

  SubscriptionState _state(
    List<SubscriptionPlan> plans,
    Map<String, dynamic> data,
  ) {
    final status = data['status']?.toString() ?? 'free';
    return SubscriptionState(
      status: switch (status) {
        'active' => SubscriptionStatus.active,
        'expired' => SubscriptionStatus.expired,
        'cancelled' => SubscriptionStatus.cancelled,
        _ => SubscriptionStatus.free,
      },
      plans: plans,
      activePlanId: data['plan_id']?.toString(),
      selectedPlanId: data['plan_id']?.toString() ?? 'premium-yearly',
      endsAt: DateTime.tryParse(data['ends_at']?.toString() ?? ''),
    );
  }
}
