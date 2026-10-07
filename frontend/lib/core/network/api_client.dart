import 'dart:convert';

import 'package:http/http.dart' as http;

import 'token_storage.dart';

const requestTimeout = Duration(seconds: 15);

String normalizeBaseUrl(String value) =>
    value.endsWith('/') ? value.substring(0, value.length - 1) : value;

class ApiClient {
  ApiClient({http.Client? client, this.tokenStorage, String? baseUrl})
    : _client = client ?? http.Client(),
      baseUrl = normalizeBaseUrl(
        baseUrl ??
            const String.fromEnvironment(
              'SHATRANJ_API_URL',
              defaultValue: 'http://localhost:8080/api/v1',
            ),
      );

  final http.Client _client;
  final TokenStorage? tokenStorage;
  final String baseUrl;

  Future<Map<String, dynamic>> get(String path) async {
    final response = await _client
        .get(Uri.parse('$baseUrl$path'), headers: await _headers())
        .timeout(requestTimeout);
    return _decode(response);
  }

  Future<Map<String, dynamic>> post(
    String path, {
    Map<String, dynamic>? body,
  }) async {
    final response = await _client
        .post(
          Uri.parse('$baseUrl$path'),
          headers: await _headers(),
          body: body == null ? null : jsonEncode(body),
        )
        .timeout(requestTimeout);
    return _decode(response);
  }

  Future<Map<String, dynamic>> put(
    String path, {
    Map<String, dynamic>? body,
  }) async {
    final response = await _client
        .put(
          Uri.parse('$baseUrl$path'),
          headers: await _headers(),
          body: body == null ? null : jsonEncode(body),
        )
        .timeout(requestTimeout);
    return _decode(response);
  }

  Future<Map<String, dynamic>> delete(String path) async {
    final response = await _client
        .delete(Uri.parse('$baseUrl$path'), headers: await _headers())
        .timeout(requestTimeout);
    return _decode(response);
  }

  Future<Map<String, String>> _headers() async {
    final token = await tokenStorage?.readAccessToken();
    return {
      'Accept': 'application/json',
      'Content-Type': 'application/json',
      if (token != null && token.isNotEmpty) 'Authorization': 'Bearer $token',
    };
  }

  Future<Map<String, dynamic>> _decode(http.Response response) async {
    dynamic decoded;
    try {
      decoded = response.body.isEmpty
          ? <String, dynamic>{}
          : jsonDecode(response.body);
    } on FormatException {
      throw ApiException(
        response.statusCode,
        'The server returned an invalid response.',
      );
    }
    if (response.statusCode < 200 || response.statusCode >= 300) {
      if (response.statusCode == 401) {
        await tokenStorage?.clear();
      }
      throw ApiException(
        response.statusCode,
        decoded is Map<String, dynamic> &&
                decoded['error'] is Map<String, dynamic>
            ? (decoded['error'] as Map<String, dynamic>)['message']
                      ?.toString() ??
                  'Request failed'
            : 'Request failed',
      );
    }
    return decoded is Map<String, dynamic> ? decoded : {'data': decoded};
  }
}

class ApiException implements Exception {
  const ApiException(this.statusCode, this.message);

  final int statusCode;
  final String message;

  @override
  String toString() => 'ApiException($statusCode): $message';
}
