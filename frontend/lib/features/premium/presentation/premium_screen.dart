import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/app_widgets.dart';
import '../domain/subscription_models.dart';
import 'subscription_controller.dart';

class PremiumScreen extends ConsumerWidget {
  const PremiumScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final subscription = ref.watch(subscriptionControllerProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Shatranj Premium')),
      body: SafeArea(
        child: subscription.when(
          loading: () => const AppLoading(message: 'Loading plans'),
          error: (error, _) => AppError(title: 'Premium unavailable', message: error.toString()),
          data: (state) => _PremiumContent(state: state),
        ),
      ),
    );
  }
}

class _PremiumContent extends ConsumerWidget {
  const _PremiumContent({required this.state});
  final SubscriptionState state;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final controller = ref.read(subscriptionControllerProvider.notifier);
    return ListView(
      padding: const EdgeInsets.fromLTRB(AppSpacing.page, AppSpacing.sm, AppSpacing.page, AppSpacing.xl),
      children: [
        AppGradientCard(
          child: Column(
            children: [
              const Icon(Icons.workspace_premium_rounded, color: AppColors.gold, size: 58),
              const SizedBox(height: AppSpacing.sm),
              const Text('Unlock your full chess potential', style: AppTextStyles.headline, textAlign: TextAlign.center),
              const SizedBox(height: AppSpacing.xs),
              const Text('Go deeper, improve faster, and enjoy the complete Shatranj experience.', style: AppTextStyles.body, textAlign: TextAlign.center),
              if (state.isPremium) ...[
                const SizedBox(height: AppSpacing.md),
                const AppPill(label: 'Premium active', icon: Icons.check_circle_rounded, color: AppColors.success),
              ],
            ],
          ),
        ),
        const SizedBox(height: AppSpacing.lg),
        const Text('Choose your plan', style: AppTextStyles.title),
        const SizedBox(height: AppSpacing.sm),
        ...state.plans.map((plan) => _PlanCard(
              plan: plan,
              selected: plan.id == state.selectedPlanId,
              onTap: () => controller.selectPlan(plan.id),
            )),
        const SizedBox(height: AppSpacing.lg),
        const Text('Premium includes', style: AppTextStyles.title),
        const SizedBox(height: AppSpacing.sm),
        const _FeatureList(),
        const SizedBox(height: AppSpacing.lg),
        if (state.error != null) ...[
          AppCard(
            borderColor: AppColors.danger.withValues(alpha: .6),
            child: Row(
              children: [
                const Icon(Icons.info_outline_rounded, color: AppColors.warning),
                const SizedBox(width: AppSpacing.sm),
                Expanded(child: Text(state.error!, style: AppTextStyles.body)),
              ],
            ),
          ),
          const SizedBox(height: AppSpacing.md),
        ],
        AppButton(
          label: state.isPremium ? 'Premium active' : 'Continue with Premium',
          icon: Icons.lock_open_rounded,
          expand: true,
          loading: state.status == SubscriptionStatus.restoring,
          onPressed: state.isPremium ? null : controller.startSubscription,
        ),
        const SizedBox(height: AppSpacing.sm),
        AppButton(
          label: 'Restore purchase',
          icon: Icons.restore_rounded,
          variant: AppButtonVariant.outline,
          expand: true,
          loading: state.status == SubscriptionStatus.restoring,
          onPressed: controller.restorePurchase,
        ),
        const SizedBox(height: AppSpacing.sm),
        Text(
          'Payment processing will be connected later. No payment is taken in this preview.',
          style: AppTextStyles.caption,
          textAlign: TextAlign.center,
        ),
      ],
    );
  }
}

class _PlanCard extends StatelessWidget {
  const _PlanCard({required this.plan, required this.selected, required this.onTap});
  final SubscriptionPlan plan;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) => AppCard(
        margin: const EdgeInsets.only(bottom: AppSpacing.sm),
        color: selected ? AppColors.surfaceElevated : null,
        borderColor: selected ? AppColors.gold : null,
        onTap: onTap,
        child: Row(
          children: [
            Icon(selected ? Icons.radio_button_checked : Icons.radio_button_off, color: selected ? AppColors.gold : AppColors.muted),
            const SizedBox(width: AppSpacing.sm),
            Expanded(
              child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                Text(plan.interval == 'year' ? 'Yearly' : 'Monthly', style: AppTextStyles.titleSmall),
                Text(plan.interval == 'year' ? 'Best value for committed players' : 'Flexible monthly access', style: AppTextStyles.caption),
              ]),
            ),
            Column(crossAxisAlignment: CrossAxisAlignment.end, children: [
              Text(plan.priceLabel, style: AppTextStyles.bodyStrong.copyWith(color: AppColors.gold)),
              Text(plan.intervalLabel, style: AppTextStyles.caption),
            ]),
          ],
        ),
      );
}

class _FeatureList extends StatelessWidget {
  const _FeatureList();
  static const features = [
    'Unlimited game analysis',
    'Advanced puzzles and themes',
    'Full courses and lessons',
    'Advanced statistics',
    'Ad-free experience',
    'Premium tournaments',
    'Early-access features',
  ];

  @override
  Widget build(BuildContext context) => AppCard(
        child: Column(
          children: features
              .map((feature) => Padding(
                    padding: const EdgeInsets.symmetric(vertical: AppSpacing.xs),
                    child: Row(
                      children: [
                        const Icon(Icons.check_circle_rounded, color: AppColors.mint, size: 19),
                        const SizedBox(width: AppSpacing.sm),
                        Expanded(child: Text(feature, style: AppTextStyles.body)),
                      ],
                    ),
                  ))
              .toList(),
        ),
      );
}
