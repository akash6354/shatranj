import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'package:frontend/features/home/presentation/home_screen.dart';

void main() {
  testWidgets('renders the home dashboard content', (tester) async {
    await tester.pumpWidget(
      const ProviderScope(
        child: MaterialApp(
          home: HomeScreen(),
        ),
      ),
    );

    await tester.pumpAndSettle();

    expect(find.text('Good morning, Akash'), findsOneWidget);
    expect(find.text('Daily puzzle'), findsOneWidget);
    expect(find.text('Continue learning'), findsOneWidget);
    expect(find.text('Solve puzzle'), findsOneWidget);
  });
}
