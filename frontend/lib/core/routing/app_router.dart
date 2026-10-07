import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import '../../features/auth/presentation/auth_screens.dart';
import '../../features/home/presentation/home_screen.dart';
import '../../features/game/presentation/game_screen.dart';
import '../../features/review/presentation/review_screen.dart';
import '../../features/history/presentation/game_history_detail_screen.dart';
import '../../features/history/presentation/game_history_screen.dart';
import '../../features/history/presentation/game_result_screen.dart';
import '../../features/puzzles/presentation/puzzle_home_screen.dart';
import '../../features/puzzles/presentation/puzzle_play_screen.dart';
import '../../features/puzzles/presentation/puzzle_result_screen.dart';
import '../../features/learning/presentation/learning_home_screen.dart';
import '../../features/learning/presentation/lesson_detail_screen.dart';
import '../../features/learning/presentation/lesson_list_screen.dart';
import '../../features/play/presentation/play_home_screen.dart';
import '../../features/play/presentation/matchmaking_screen.dart';
import '../../features/play/presentation/computer_game_screen.dart';
import '../../features/play/presentation/friend_game_screen.dart';
import '../../features/tournaments/presentation/tournament_home_screen.dart';
import '../../features/tournaments/presentation/tournament_detail_screen.dart';
import '../../features/community/presentation/community_home_screen.dart';
import '../../features/community/presentation/clubs_screen.dart';
import '../../features/community/presentation/club_detail_screen.dart';
import '../../features/community/presentation/friends_screen.dart';
import '../../features/community/presentation/coaches_screen.dart';
import '../../features/community/presentation/chat_screens.dart';
import '../../features/profile/presentation/profile_screen.dart';
import '../../features/settings/presentation/settings_screen.dart';
import '../../features/premium/presentation/premium_screen.dart';
import '../../features/shell/presentation/app_shell.dart';
import '../../features/shell/presentation/placeholder_screen.dart';

abstract final class AppRoutes {
  static const splash = '/auth/splash';
  static const onboarding = '/auth/onboarding';
  static const login = '/auth/login';
  static const register = '/auth/register';
  static const forgotPassword = '/auth/forgot-password';
  static const home = '/';
  static const learn = '/learn';
  static const play = '/play';
  static const community = '/community';
  static const profile = '/profile';
  static const gameHistory = '/games/history';
  static const puzzles = '/puzzles';
  static const game = '/game';
  static const tournaments = '/tournaments';
  static const premium = '/premium';
  static String gameReview(String gameId) => '/game/$gameId/review';
}

GoRouter createAppRouter() => GoRouter(
      initialLocation: AppRoutes.splash,
      routes: [
        GoRoute(
          path: 'splash',
          builder: (_, _) => const SplashScreen(),
        ),
        GoRoute(
          path: 'onboarding',
          builder: (_, _) => const OnboardingScreen(),
        ),
        GoRoute(
          path: '/auth',
          redirect: (_, state) => state.fullPath == '/auth' ? AppRoutes.login : null,
          routes: [
            GoRoute(
              path: 'login',
              builder: (_, _) => const LoginScreen(),
            ),
            GoRoute(
              path: 'register',
              builder: (_, _) => const RegisterScreen(),
            ),
            GoRoute(
              path: 'forgot-password',
              builder: (_, _) => const ForgotPasswordScreen(),
            ),
          ],
        ),
        GoRoute(
          path: '/game/:gameId',
          builder: (context, state) => GameScreen(
            gameId: state.pathParameters['gameId']!,
          ),
        ),
        GoRoute(
          path: '/game/:gameId/review',
          builder: (context, state) => ReviewScreen(
            gameId: state.pathParameters['gameId']!,
          ),
        ),
        GoRoute(
          path: '/game/:gameId/result',
          builder: (context, state) => GameResultScreen(
            gameId: state.pathParameters['gameId']!,
          ),
        ),
        GoRoute(
          path: AppRoutes.gameHistory,
          builder: (_, _) => const GameHistoryScreen(),
          routes: [
            GoRoute(
              path: ':gameId',
              builder: (context, state) => GameHistoryDetailScreen(
                gameId: state.pathParameters['gameId']!,
              ),
            ),
          ],
        ),
        GoRoute(
          path: AppRoutes.puzzles,
          builder: (_, _) => const PuzzleHomeScreen(),
          routes: [
            GoRoute(
              path: ':puzzleId/play',
              builder: (context, state) => PuzzlePlayScreen(
                puzzleId: state.pathParameters['puzzleId']!,
              ),
            ),
            GoRoute(
              path: ':puzzleId/result',
              builder: (context, state) => PuzzleResultScreen(
                puzzleId: state.pathParameters['puzzleId']!,
              ),
            ),
          ],
        ),
        GoRoute(
          path: AppRoutes.tournaments,
          builder: (_, _) => const TournamentHomeScreen(),
          routes: [
            GoRoute(
              path: ':tournamentId',
              builder: (context, state) => TournamentDetailScreen(
                tournamentId: state.pathParameters['tournamentId']!,
              ),
            ),
          ],
        ),
        GoRoute(
          path: AppRoutes.premium,
          builder: (_, _) => const PremiumScreen(),
        ),
        StatefulShellRoute.indexedStack(
          builder: (context, state, navigationShell) =>
              AppShell(navigationShell: navigationShell),
          branches: [
            _branch(
              AppRoutes.home,
              const HomeScreen(),
            ),
            _branch(
              AppRoutes.learn,
              const LearningHomeScreen(),
              routes: [
                GoRoute(
                  path: ':courseId',
                  builder: (context, state) => LessonListScreen(
                    courseId: state.pathParameters['courseId']!,
                  ),
                  routes: [
                    GoRoute(
                      path: 'lessons/:lessonId',
                      builder: (context, state) => LessonDetailScreen(
                        courseId: state.pathParameters['courseId']!,
                        lessonId: state.pathParameters['lessonId']!,
                      ),
                    ),
                  ],
                ),
              ],
            ),
            _branch(
              AppRoutes.play,
              const PlayHomeScreen(),
              routes: [
                GoRoute(
                  path: 'matchmaking',
                  builder: (_, _) => const MatchmakingScreen(),
                  routes: [
                    GoRoute(
                      path: 'found',
                      builder: (_, _) => const MatchFoundScreen(),
                    ),
                  ],
                ),
                GoRoute(
                  path: 'computer',
                  builder: (_, _) => const ComputerGameScreen(),
                ),
                GoRoute(
                  path: 'friend',
                  builder: (_, _) => const FriendGameScreen(),
                ),
              ],
            ),
            _branch(
              AppRoutes.community,
              const CommunityHomeScreen(),
              routes: [
                GoRoute(path: 'clubs', builder: (_, _) => const ClubsScreen(), routes: [
                  GoRoute(path: ':clubId', builder: (context, state) => ClubDetailScreen(clubId: state.pathParameters['clubId']!)),
                ]),
                GoRoute(path: 'friends', builder: (_, _) => const FriendsScreen()),
                GoRoute(path: 'coaches', builder: (_, _) => const CoachesScreen()),
                GoRoute(path: 'chat', builder: (_, _) => const ChatRoomsScreen()),
              ],
            ),
            _branch(
              AppRoutes.profile,
              const ProfileScreen(),
              routes: [
                GoRoute(
                  path: 'settings',
                  builder: (_, _) => const SettingsScreen(),
                ),
              ],
            ),
          ],
        ),
      ],
      errorBuilder: (context, state) => PlaceholderScreen(
        title: 'Page not found',
        icon: Icons.search_off_rounded,
        message: state.error?.toString(),
      ),
    );

StatefulShellBranch _branch(
  String path,
  Widget screen, {
  List<GoRoute> routes = const [],
}) {
  return StatefulShellBranch(
    routes: [
      GoRoute(
        path: path,
        builder: (_, _) => screen,
        routes: routes,
      ),
    ],
  );
}
