import '../http.dart';

class AuthApi {
  final BackendHttp _http;
  final TokenStore _tokens;
  AuthApi(this._http, this._tokens);

  Future<Map<String, dynamic>?> login(Map<String, dynamic> payload) async {
    final data = await _http.request('POST', '/users/login', body: payload);
    if (data is Map<String, dynamic>) {
      _maybeUpdateTokens(data);
    }
    return (data is Map<String, dynamic>) ? data : null;
  }

  Future<Map<String, dynamic>?> loginWithDevice({
    required String username,
    required String password,
    String? deviceVerificationCode,
    String? deviceId,
    String? deviceName,
    String? deviceType,
  }) async {
    final payload = <String, dynamic>{
      'username': username,
      'password': password,
    };
    if (deviceId != null) payload['device_id'] = deviceId;
    if (deviceName != null) payload['device_name'] = deviceName;
    if (deviceType != null) payload['device_type'] = deviceType;
    if (deviceVerificationCode != null) payload['device_verification_code'] = deviceVerificationCode;

    final data = await _http.request('POST', '/users/login', body: payload);
    if (data is Map<String, dynamic>) {
      _maybeUpdateTokens(data);
    }
    return (data is Map<String, dynamic>) ? data : null;
  }

  Future<Map<String, dynamic>?> loginWithCustomDevice(Map<String, dynamic> payload) async {
    final data = await _http.request('POST', '/users/login', body: payload);
    if (data is Map<String, dynamic>) {
      _maybeUpdateTokens(data);
    }
    return (data is Map<String, dynamic>) ? data : null;
  }

  Future<Map<String, dynamic>?> refresh({String? refreshToken}) async {
    final token = refreshToken ?? _tokens.refreshToken;
    if (token == null) {
      throw ArgumentError('refresh requires refreshToken');
    }
    final data = await _http.request('POST', '/users/refresh', body: {'refresh_token': token});
    if (data is Map<String, dynamic>) {
      final access = data['access_token'] as String?;
      if (access != null) _tokens.accessToken = access;
    }
    return (data is Map<String, dynamic>) ? data : null;
  }

  Future<dynamic> logout({String? accessToken, String? refreshToken}) async {
    final at = accessToken ?? _tokens.accessToken;
    final rt = refreshToken ?? _tokens.refreshToken;
    if (at == null || rt == null) {
      throw ArgumentError('logout requires access_token and refresh_token');
    }
    final res = await _http.request('POST', '/users/logout', body: {
      'access_token': at,
      'refresh_token': rt,
    });
    _tokens.accessToken = null;
    _tokens.refreshToken = null;
    return res;
  }

  void _maybeUpdateTokens(Map<String, dynamic> data) {
    final access = data['access_token'] as String?;
    final refresh = data['refresh_token'] as String?;
    if (access != null) _tokens.accessToken = access;
    if (refresh != null) _tokens.refreshToken = refresh;
  }
}
