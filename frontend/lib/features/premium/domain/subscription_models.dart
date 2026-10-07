enum SubscriptionStatus { free, active, expired, cancelled, restoring, error }

class SubscriptionPlan {
  const SubscriptionPlan({
    required this.id,
    required this.name,
    required this.pricePaise,
    required this.currency,
    required this.interval,
    required this.entitlements,
  });

  final String id;
  final String name;
  final int pricePaise;
  final String currency;
  final String interval;
  final List<String> entitlements;

  String get priceLabel => '$currency ${(pricePaise / 100).toStringAsFixed(0)}';
  String get intervalLabel => interval == 'year' ? 'per year' : 'per month';
}

class SubscriptionState {
  const SubscriptionState({
    required this.status,
    required this.plans,
    this.selectedPlanId = 'premium-yearly',
    this.activePlanId,
    this.endsAt,
    this.error,
  });

  factory SubscriptionState.initial(List<SubscriptionPlan> plans) =>
      SubscriptionState(status: SubscriptionStatus.free, plans: plans);

  final SubscriptionStatus status;
  final List<SubscriptionPlan> plans;
  final String selectedPlanId;
  final String? activePlanId;
  final DateTime? endsAt;
  final String? error;

  SubscriptionPlan? get selectedPlan =>
      plans.where((plan) => plan.id == selectedPlanId).firstOrNull;

  bool get isPremium => status == SubscriptionStatus.active;

  SubscriptionState copyWith({
    SubscriptionStatus? status,
    String? selectedPlanId,
    String? activePlanId,
    DateTime? endsAt,
    String? error,
  }) =>
      SubscriptionState(
        status: status ?? this.status,
        plans: plans,
        selectedPlanId: selectedPlanId ?? this.selectedPlanId,
        activePlanId: activePlanId ?? this.activePlanId,
        endsAt: endsAt ?? this.endsAt,
        error: error,
      );
}
