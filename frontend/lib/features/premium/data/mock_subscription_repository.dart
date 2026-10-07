import '../domain/subscription_models.dart';
import '../domain/subscription_repository.dart';

class MockSubscriptionRepository implements SubscriptionRepository {
  static const _plans = [
    SubscriptionPlan(
      id: 'premium-monthly',
      name: 'Premium Monthly',
      pricePaise: 19900,
      currency: 'INR',
      interval: 'month',
      entitlements: _entitlements,
    ),
    SubscriptionPlan(
      id: 'premium-yearly',
      name: 'Premium Yearly',
      pricePaise: 149900,
      currency: 'INR',
      interval: 'year',
      entitlements: _entitlements,
    ),
  ];

  static const _entitlements = [
    'unlimited_analysis',
    'advanced_puzzles',
    'full_courses',
    'advanced_statistics',
    'no_ads',
    'premium_tournaments',
    'early_access',
  ];

  @override
  Future<List<SubscriptionPlan>> getPlans() async => _plans;

  @override
  Future<SubscriptionState> getSubscription() async =>
      SubscriptionState.initial(_plans);

  @override
  Future<SubscriptionState> restorePurchase() async {
    await Future<void>.delayed(const Duration(milliseconds: 500));
    return SubscriptionState(
      status: SubscriptionStatus.free,
      plans: _plans,
      error: 'No previous purchase was found for this account.',
    );
  }

  @override
  Future<SubscriptionState> selectPlan(String planId) async {
    await Future<void>.delayed(const Duration(milliseconds: 250));
    return SubscriptionState(
      status: SubscriptionStatus.free,
      plans: _plans,
      selectedPlanId: planId,
    );
  }
}
