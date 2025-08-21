import 'dart:async';
import 'package:dio/dio.dart';
import 'errors.dart';

class TokenStore {
  String? accessToken;
  String? refreshToken;
}

class BackendHttp {
  final Dio _dio;
  final String baseUrl; // e.g. http://localhost:8080
  final String basePath; // e.g. /api/v1
  final TokenStore tokens;
  final bool autoRefresh;
  bool _isRefreshing = false;

  BackendHttp({
    required this.baseUrl,
    this.basePath = '/api/v1',
    required this.tokens,
    this.autoRefresh = true,
    BaseOptions? options,
  }) : _dio = Dio(options ?? BaseOptions(connectTimeout: const Duration(seconds: 15), receiveTimeout: const Duration(seconds: 30)));

  String _buildUrl(String path, [Map<String, dynamic>? query]) {
    final base = baseUrl.replaceFirst(RegExp(r'/\$'), '');
    final uri = Uri.parse('$base$basePath$path').replace(queryParameters: _cleanupQuery(query));
    return uri.toString();
  }

  Map<String, dynamic>? _cleanupQuery(Map<String, dynamic>? q) {
    if (q == null) return null;
    final m = <String, dynamic>{};
    q.forEach((k, v) {
      if (v != null) m[k] = v;
    });
    return m;
  }

  Future<dynamic> request(
    String method,
    String path, {
    Map<String, dynamic>? query,
    dynamic body,
    bool auth = false,
    Map<String, String>? headers,
    bool isForm = false,
  }) async {
    final url = _buildUrl(path, query);
    final reqHeaders = <String, dynamic>{...?(headers)};
    if (auth && tokens.accessToken != null) {
      reqHeaders['Authorization'] = 'Bearer ${tokens.accessToken}';
    }

    final options = Options(method: method, headers: reqHeaders, responseType: ResponseType.json);

    Future<Response<dynamic>> send() {
      if (isForm) {
        return _dio.request(url, options: options, data: body);
      }
      return _dio.request(url, options: options, data: body);
    }

    try {
      final resp = await send();
      return _unwrap(resp);
    } on DioException catch (e) {
      final resp = e.response;
      if (_shouldTryRefresh(resp, auth)) {
        try {
          await _refreshAccessToken();
        } catch (_) {
          // fallthrough to original error unwrap
          return _unwrapResponseMaybe(resp);
        }
        // retry once without recursive refresh
        try {
          final retryResp = await _dio.request(url, options: options.copyWith(headers: {
            ...?options.headers,
            if (auth && tokens.accessToken != null) 'Authorization': 'Bearer ${tokens.accessToken}',
          }), data: body);
          return _unwrap(retryResp);
        } on DioException catch (e2) {
          return _unwrapResponseMaybe(e2.response);
        }
      }
      return _unwrapResponseMaybe(resp);
    }
  }

  bool _shouldTryRefresh(Response? resp, bool auth) {
    return resp != null && resp.statusCode == 401 && auth && autoRefresh && tokens.refreshToken != null;
  }

  dynamic _unwrap(Response resp) {
    final status = resp.statusCode ?? 0;
    final data = resp.data;
    if (status < 200 || status >= 300) {
      final message = data is Map && data['message'] is String ? data['message'] as String : 'HTTP $status';
      final code = data is Map ? data['code']?.toString() : null;
      throw BackendApiError(message, status: status, code: code, payload: data);
    }
    if (data is Map && data.containsKey('data')) {
      return data['data'];
    }
    return data;
  }

  dynamic _unwrapResponseMaybe(Response? resp) {
    if (resp == null) {
      throw BackendApiError('Network error', status: null);
    }
    return _unwrap(resp);
  }

  Future<String> _refreshAccessToken() async {
    if (_isRefreshing) {
      // simple guard: avoid concurrent refresh; let the first throw if it fails
      throw BackendApiError('Refresh in progress');
    }
    if (tokens.refreshToken == null) {
      throw BackendApiError('Missing refresh token');
    }
    _isRefreshing = true;
    try {
      final resp = await _dio.post(_buildUrl('/users/refresh'), data: {
        'refresh_token': tokens.refreshToken,
      });
      final data = _unwrap(resp);
      final access = data is Map ? data['access_token'] as String? : null;
      if (access == null) {
        throw BackendApiError('Invalid refresh response');
      }
      tokens.accessToken = access;
      return access;
    } finally {
      _isRefreshing = false;
    }
  }
}
