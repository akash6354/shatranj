import 'subscription_models.dart';

abstract interface class SubscriptionRepository {
  Future<List<SubscriptionPlan>> getPlans();
  Future<SubscriptionState> getSubscription();
  Future<SubscriptionState> restorePurchase();
  Future<SubscriptionState> selectPlan(String planId);
}
