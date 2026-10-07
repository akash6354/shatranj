import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import 'package:frontend/core/routing/app_router.dart';
import 'package:frontend/core/theme/app_theme.dart';

void main() {
  testWidgets('authenticated shell exposes every main destination', (tester) async {
    final router = createAppRouter();
    await tester.pumpWidget(_RouterHost(router: router));
    await tester.pumpAndSettle();

    router.go(AppRoutes.home);
    await tester.pumpAndSettle();
    expect(find.text('Home', findRichText: true), findsOneWidget);

    for (final route in [
      AppRoutes.learn,
      AppRoutes.play,
      AppRoutes.community,
      AppRoutes.profile,
    ]) {
      router.go(route);
      await tester.pumpAndSettle();
      expect(find.text(route.substring(1).capitalize()), findsNWidgets(2));
    }
  });

  testWidgets('authentication routes and back navigation work', (tester) async {
    final router = createAppRouter();
    await tester.pumpWidget(_RouterHost(router: router));

    router.go(AppRoutes.login);
    await tester.pumpAndSettle();
    expect(find.text('Welcome back'), findsOneWidget);

    router.go(AppRoutes.register);
    await tester.pumpAndSettle();
    expect(find.text('Create your account'), findsOneWidget);

    router.push(AppRoutes.forgotPassword);
    await tester.pumpAndSettle();
    expect(find.text('Reset password'), findsOneWidget);

    router.pop();
    await tester.pumpAndSettle();
    expect(find.text('Create your account'), findsOneWidget);
  });
}

class _RouterHost extends StatelessWidget {
  const _RouterHost({required this.router});

  final GoRouter router;

  @override
  Widget build(BuildContext context) => ProviderScope(
        child: MaterialApp.router(
          routerConfig: router,
          theme: AppTheme.dark,
        ),
      );
}

extension on String {
  String capitalize() => '${this[0].toUpperCase()}${substring(1)}';
}
