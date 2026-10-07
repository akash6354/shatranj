import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/routing/app_router.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/app_widgets.dart';
import 'auth_controller.dart';

class SplashScreen extends StatefulWidget {
  const SplashScreen({super.key});

  @override
  State<SplashScreen> createState() => _SplashScreenState();
}

class _SplashScreenState extends State<SplashScreen> {
  Timer? _timer;

  @override
  void initState() {
    super.initState();
    _timer = Timer(const Duration(milliseconds: 1500), () {
      if (mounted) context.go(AppRoutes.onboarding);
    });
  }

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => Scaffold(
        body: SafeArea(
          child: Center(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Container(
                  width: 104,
                  height: 104,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    color: AppColors.gold.withValues(alpha: .14),
                    border: Border.all(color: AppColors.goldMuted),
                    boxShadow: AppShadows.goldGlow,
                  ),
                  child: const Icon(
                    Icons.workspace_premium_rounded,
                    color: AppColors.gold,
                    size: 52,
                  ),
                ),
                const SizedBox(height: AppSpacing.lg),
                const Text('Shatranj', style: AppTextStyles.display),
                const SizedBox(height: AppSpacing.xs),
                const Text('Think • Train • Play • Grow', style: AppTextStyles.body),
                const SizedBox(height: AppSpacing.xxl),
                const SizedBox(
                  width: 22,
                  height: 22,
                  child: CircularProgressIndicator(strokeWidth: 2),
                ),
              ],
            ),
          ),
        ),
      );
}

class OnboardingScreen extends StatefulWidget {
  const OnboardingScreen({super.key});

  @override
  State<OnboardingScreen> createState() => _OnboardingScreenState();
}

class _OnboardingScreenState extends State<OnboardingScreen> {
  final _pageController = PageController();
  int _page = 0;

  static const _pages = [
    (
      icon: Icons.auto_awesome_rounded,
      title: 'More than a game',
      body: 'Sharpen your mind with puzzles, lessons, and real opponents.',
    ),
    (
      icon: Icons.school_rounded,
      title: 'Learn at your pace',
      body: 'Build stronger habits with guided courses and thoughtful analysis.',
    ),
    (
      icon: Icons.emoji_events_rounded,
      title: 'Play with purpose',
      body: 'Track your progress and grow into the player you want to become.',
    ),
  ];

  @override
  void dispose() {
    _pageController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final page = _pages[_page];
    return Scaffold(
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.fromLTRB(
            AppSpacing.page,
            AppSpacing.lg,
            AppSpacing.page,
            AppSpacing.md,
          ),
          child: Column(
            children: [
              Align(
                alignment: Alignment.centerRight,
                child: TextButton(
                  onPressed: () => context.go(AppRoutes.login),
                  child: const Text('Skip'),
                ),
              ),
              Expanded(
                child: PageView.builder(
                  controller: _pageController,
                  itemCount: _pages.length,
                  onPageChanged: (value) => setState(() => _page = value),
                  itemBuilder: (context, index) => _OnboardingPage(data: _pages[index]),
                ),
              ),
              Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: List.generate(
                  _pages.length,
                  (index) => AnimatedContainer(
                    duration: const Duration(milliseconds: 180),
                    margin: const EdgeInsets.symmetric(horizontal: 3),
                    width: index == _page ? 24 : 7,
                    height: 7,
                    decoration: BoxDecoration(
                      color: index == _page ? AppColors.gold : AppColors.borderStrong,
                      borderRadius: AppRadius.pillRadius,
                    ),
                  ),
                ),
              ),
              const SizedBox(height: AppSpacing.lg),
              AppButton(
                label: _page == _pages.length - 1 ? 'Get started' : 'Continue',
                icon: _page == _pages.length - 1 ? Icons.arrow_forward_rounded : null,
                expand: true,
                onPressed: () {
                  if (_page == _pages.length - 1) {
                    context.go(AppRoutes.login);
                  } else {
                    _pageController.nextPage(
                      duration: const Duration(milliseconds: 240),
                      curve: Curves.easeOut,
                    );
                  }
                },
              ),
              const SizedBox(height: AppSpacing.sm),
              Text(
                '${page.title} • ${_page + 1}/${_pages.length}',
                style: AppTextStyles.caption,
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _OnboardingPage extends StatelessWidget {
  const _OnboardingPage({required this.data});

  final ({IconData icon, String title, String body}) data;

  @override
  Widget build(BuildContext context) => Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Container(
            width: 220,
            height: 280,
            decoration: BoxDecoration(
              gradient: const LinearGradient(
                colors: [Color(0xFF1A4B42), AppColors.surface],
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
              ),
              borderRadius: AppRadius.dialogRadius,
              border: Border.all(color: AppColors.borderStrong),
              boxShadow: AppShadows.elevated,
            ),
            child: Icon(data.icon, color: AppColors.gold, size: 82),
          ),
          const SizedBox(height: AppSpacing.xl),
          Text(data.title, style: AppTextStyles.headline, textAlign: TextAlign.center),
          const SizedBox(height: AppSpacing.sm),
          Text(data.body, style: AppTextStyles.body, textAlign: TextAlign.center),
        ],
      );
}

abstract class _AuthFormScreen extends ConsumerStatefulWidget {
  const _AuthFormScreen({required this.register, super.key});

  final bool register;
}

class LoginScreen extends _AuthFormScreen {
  const LoginScreen({super.key}) : super(register: false);

  @override
  ConsumerState<LoginScreen> createState() => _AuthFormState<LoginScreen>();
}

class RegisterScreen extends _AuthFormScreen {
  const RegisterScreen({super.key}) : super(register: true);

  @override
  ConsumerState<RegisterScreen> createState() => _AuthFormState<RegisterScreen>();
}

class _AuthFormState<T extends _AuthFormScreen> extends ConsumerState<T> {
  final _nameController = TextEditingController();
  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();

  @override
  void dispose() {
    _nameController.dispose();
    _emailController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    ref.listen(authControllerProvider, (_, next) {
      next.whenOrNull(
        data: (session) {
          if (session != null && context.mounted) context.go(AppRoutes.home);
        },
      );
    });
    final state = ref.watch(authControllerProvider);
    final loading = state.isLoading;
    return Scaffold(
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(AppSpacing.page),
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 430),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  const _AuthBrand(),
                  const SizedBox(height: AppSpacing.xl),
                  Text(
                    widget.register ? 'Create your account' : 'Welcome back',
                    style: AppTextStyles.headline,
                  ),
                  const SizedBox(height: AppSpacing.xs),
                  Text(
                    widget.register
                        ? 'Start learning, playing, and improving.'
                        : 'Sign in to continue your chess journey.',
                    style: AppTextStyles.body,
                  ),
                  const SizedBox(height: AppSpacing.md),
                  const AppPill(
                    label: 'Demo authentication • REST integration pending',
                    icon: Icons.construction_rounded,
                    color: AppColors.goldSoft,
                  ),
                  const SizedBox(height: AppSpacing.lg),
                  if (widget.register) ...[
                    TextField(
                      controller: _nameController,
                      textInputAction: TextInputAction.next,
                      decoration: const InputDecoration(
                        labelText: 'Display name',
                        prefixIcon: Icon(Icons.person_outline_rounded),
                      ),
                    ),
                    const SizedBox(height: AppSpacing.md),
                  ],
                  TextField(
                    controller: _emailController,
                    keyboardType: TextInputType.emailAddress,
                    textInputAction: TextInputAction.next,
                    decoration: const InputDecoration(
                      labelText: 'Email',
                      prefixIcon: Icon(Icons.mail_outline_rounded),
                    ),
                  ),
                  const SizedBox(height: AppSpacing.md),
                  TextField(
                    controller: _passwordController,
                    obscureText: true,
                    onSubmitted: loading ? null : (_) => _submit(),
                    decoration: const InputDecoration(
                      labelText: 'Password',
                      prefixIcon: Icon(Icons.lock_outline_rounded),
                    ),
                  ),
                  if (!widget.register)
                    Align(
                      alignment: Alignment.centerRight,
                      child: TextButton(
                        onPressed: () => context.push(AppRoutes.forgotPassword),
                        child: const Text('Forgot password?'),
                      ),
                    ),
                  const SizedBox(height: AppSpacing.sm),
                  if (state.hasError) ...[
                    _AuthError(message: state.error.toString()),
                    const SizedBox(height: AppSpacing.sm),
                  ],
                  AppButton(
                    label: widget.register ? 'Create account' : 'Sign in',
                    expand: true,
                    loading: loading,
                    onPressed: loading ? null : _submit,
                  ),
                  const SizedBox(height: AppSpacing.lg),
                  const _OrDivider(),
                  const SizedBox(height: AppSpacing.lg),
                  AppButton(
                    label: 'Continue with Google',
                    icon: Icons.g_mobiledata_rounded,
                    variant: AppButtonVariant.outline,
                    expand: true,
                    loading: loading,
                    onPressed: loading ? null : () => ref.read(authControllerProvider.notifier).continueWithGoogle(),
                  ),
                  const SizedBox(height: AppSpacing.sm),
                  AppButton(
                    label: 'Continue with phone',
                    icon: Icons.phone_outlined,
                    variant: AppButtonVariant.outline,
                    expand: true,
                    loading: loading,
                    onPressed: loading ? null : _showPhoneDialog,
                  ),
                  const SizedBox(height: AppSpacing.xl),
                  Center(
                    child: Wrap(
                      alignment: WrapAlignment.center,
                      children: [
                        Text(
                          widget.register ? 'Already have an account?' : 'New to Shatranj?',
                          style: AppTextStyles.caption,
                        ),
                        TextButton(
                          onPressed: () => context.go(
                            widget.register ? AppRoutes.login : AppRoutes.register,
                          ),
                          child: Text(widget.register ? 'Sign in' : 'Create account'),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }

  Future<void> _submit() {
    final controller = ref.read(authControllerProvider.notifier);
    if (widget.register) {
      return controller.signUp(
        displayName: _nameController.text,
        email: _emailController.text,
        password: _passwordController.text,
      );
    }
    return controller.signIn(
      email: _emailController.text,
      password: _passwordController.text,
    );
  }

  Future<void> _showPhoneDialog() async {
    final phoneController = TextEditingController();
    final phone = await showDialog<String>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Continue with phone'),
        content: TextField(
          controller: phoneController,
          keyboardType: TextInputType.phone,
          autofocus: true,
          decoration: const InputDecoration(labelText: 'Phone number'),
        ),
        actions: [
          TextButton(onPressed: () => context.pop(), child: const Text('Cancel')),
          FilledButton(
            onPressed: () => context.pop(phoneController.text),
            child: const Text('Continue'),
          ),
        ],
      ),
    );
    phoneController.dispose();
    if (phone != null && mounted) {
      await ref.read(authControllerProvider.notifier).continueWithPhone(phone);
    }
  }
}

class ForgotPasswordScreen extends ConsumerStatefulWidget {
  const ForgotPasswordScreen({super.key});

  @override
  ConsumerState<ForgotPasswordScreen> createState() => _ForgotPasswordState();
}

class _ForgotPasswordState extends ConsumerState<ForgotPasswordScreen> {
  final _emailController = TextEditingController();

  @override
  void dispose() {
    _emailController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(authControllerProvider);
    return Scaffold(
      appBar: AppBar(
        leading: BackButton(onPressed: () => context.pop()),
        title: const Text('Reset password'),
      ),
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(AppSpacing.page),
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 430),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  const Icon(Icons.lock_reset_rounded, color: AppColors.gold, size: 56),
                  const SizedBox(height: AppSpacing.lg),
                  const Text('Forgot your password?', style: AppTextStyles.headline),
                  const SizedBox(height: AppSpacing.xs),
                  const Text(
                    'Enter your email and we will help you get back into your account.',
                    style: AppTextStyles.body,
                  ),
                  const SizedBox(height: AppSpacing.xl),
                  TextField(
                    controller: _emailController,
                    keyboardType: TextInputType.emailAddress,
                    decoration: const InputDecoration(
                      labelText: 'Email',
                      prefixIcon: Icon(Icons.mail_outline_rounded),
                    ),
                  ),
                  if (state.hasError) ...[
                    const SizedBox(height: AppSpacing.md),
                    _AuthError(message: state.error.toString()),
                  ],
                  const SizedBox(height: AppSpacing.lg),
                  AppButton(
                    label: 'Send reset link',
                    expand: true,
                    loading: state.isLoading,
                    onPressed: state.isLoading
                        ? null
                        : () => ref
                            .read(authControllerProvider.notifier)
                            .sendPasswordReset(_emailController.text),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _AuthError extends StatelessWidget {
  const _AuthError({required this.message});

  final String message;

  @override
  Widget build(BuildContext context) => Container(
        padding: const EdgeInsets.all(AppSpacing.sm),
        decoration: BoxDecoration(
          color: AppColors.danger.withValues(alpha: .12),
          borderRadius: AppRadius.small,
          border: Border.all(color: AppColors.danger.withValues(alpha: .4)),
        ),
        child: Row(
          children: [
            const Icon(Icons.error_outline_rounded, color: AppColors.danger, size: 18),
            const SizedBox(width: AppSpacing.xs),
            Expanded(child: Text(message, style: AppTextStyles.caption.copyWith(color: AppColors.dangerSoft))),
          ],
        ),
      );
}

class _OrDivider extends StatelessWidget {
  const _OrDivider();

  @override
  Widget build(BuildContext context) => Row(
        children: [
          const Expanded(child: Divider()),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: AppSpacing.sm),
            child: Text('or continue with', style: AppTextStyles.caption),
          ),
          const Expanded(child: Divider()),
        ],
      );
}

class _AuthBrand extends StatelessWidget {
  const _AuthBrand();

  @override
  Widget build(BuildContext context) => Column(
        children: [
          Container(
            width: 68,
            height: 68,
            decoration: BoxDecoration(
              color: AppColors.gold.withValues(alpha: .14),
              shape: BoxShape.circle,
              border: Border.all(color: AppColors.goldMuted),
            ),
            child: const Icon(Icons.workspace_premium_rounded, color: AppColors.gold, size: 34),
          ),
          const SizedBox(height: AppSpacing.sm),
          const Text('Shatranj', style: AppTextStyles.title),
          const SizedBox(height: AppSpacing.xxs),
          const Text('Think • Train • Play • Grow', style: AppTextStyles.caption),
        ],
      );
}
