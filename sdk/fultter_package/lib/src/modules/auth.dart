import '../http.dart';
import '../device.dart';

/// 认证与令牌相关 API。
///
/// 涵盖：登录（含设备信息）、刷新访问令牌、登出并清理本地令牌。
class AuthApi {
  final BackendHttp _http;
  final TokenStore _tokens;
  AuthApi(this._http, this._tokens);

  /// 登录（自定义负载）。
  ///
  /// 路由：POST `/users/login`
  ///
  /// 说明：若响应中包含 `access_token`/`refresh_token`，会自动更新到本地 `TokenStore`。
  Future<Map<String, dynamic>?> login(Map<String, dynamic> payload) async {
    final data = await _http.request('POST', '/users/login', body: payload);
    if (data is Map<String, dynamic>) {
      _maybeUpdateTokens(data);
    }
    return (data is Map<String, dynamic>) ? data : null;
  }

  /// 使用显式设备信息进行登录。
  ///
  /// 路由：POST `/users/login`
  ///
  /// 参数：
  /// - [username], [password] 必填
  /// - [deviceVerificationCode] 可选，陌生设备验证代码
  /// - [deviceId], [deviceName], [deviceType] 设备信息
  ///
  /// 返回：登录响应对象（Map），并在必要时更新本地令牌。
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

  /// 使用自定义登录负载（可能包含设备字段）进行登录。
  ///
  /// 路由：POST `/users/login`
  ///
  /// 返回：登录响应对象（Map），并在必要时更新本地令牌。
  Future<Map<String, dynamic>?> loginWithCustomDevice(Map<String, dynamic> payload) async {
    final data = await _http.request('POST', '/users/login', body: payload);
    if (data is Map<String, dynamic>) {
      _maybeUpdateTokens(data);
    }
    return (data is Map<String, dynamic>) ? data : null;
  }

  /// 使用当前设备信息快速登录。
  ///
  /// 路由：POST `/users/login`
  ///
  /// 说明：会自动收集当前设备指纹、名称、类型等并合并到请求体。
  Future<Map<String, dynamic>?> loginWithCurrentDevice({
    required String username,
    required String password,
    String? deviceVerificationCode,
  }) async {
    final device = await getCurrentDevicePayload();
    final payload = <String, dynamic>{
      'username': username,
      'password': password,
      ...device,
      if (deviceVerificationCode != null) 'device_verification_code': deviceVerificationCode,
    };
    final data = await _http.request('POST', '/users/login', body: payload);
    if (data is Map<String, dynamic>) {
      _maybeUpdateTokens(data);
    }
    return (data is Map<String, dynamic>) ? data : null;
  }

  /// 刷新访问令牌。
  ///
  /// 路由：POST `/users/refresh`
  ///
  /// 参数：
  /// - [refreshToken] 可选，不传则读取本地 `TokenStore.refreshToken`
  ///
  /// 返回：刷新响应（Map），若包含 `access_token` 会更新本地。
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

  /// 登出并清理本地令牌。
  ///
  /// 路由：POST `/users/logout`
  ///
  /// 参数：
  /// - [accessToken], [refreshToken] 可选，不传则读取本地 `TokenStore`。
  ///
  /// 返回：后端原始响应 `data` 字段内容。调用成功后会清空本地令牌。
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
