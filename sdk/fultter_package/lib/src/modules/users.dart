import '../http.dart';

/// 用户相关 API。
///
/// 包含用户查询、个人信息获取与更新、注册与账号激活、密码重置等。
class UsersApi {
  final BackendHttp _http;
  UsersApi(this._http);

  /// 根据用户 ID 获取用户信息。
  ///
  /// 路由：GET `/users/{id}`
  ///
  /// 返回：用户对象（Map），否则为 `null`。
  Future<Map<String, dynamic>?> getById(String id) async {
    final data = await _http.request('GET', '/users/${Uri.encodeComponent(id)}');
    return (data is Map<String, dynamic>) ? data : null;
  }

  /// 根据用户名获取用户信息。
  ///
  /// 路由：GET `/users/username/{username}`
  ///
  /// 返回：用户对象（Map），否则为 `null`。
  Future<Map<String, dynamic>?> getByUsername(String username) async {
    final data = await _http.request('GET', '/users/username/${Uri.encodeComponent(username)}');
    return (data is Map<String, dynamic>) ? data : null;
  }

  /// 获取当前登录用户信息。
  ///
  /// 路由：GET `/users/me`（需鉴权）
  ///
  /// 返回：用户对象（Map），否则为 `null`。
  Future<Map<String, dynamic>?> me() async {
    final data = await _http.request('GET', '/users/me', auth: true);
    return (data is Map<String, dynamic>) ? data : null;
  }

  /// 更新当前登录用户信息。
  ///
  /// 路由：PUT `/users/me`（需鉴权）
  ///
  /// 参数：
  /// - [payload] 需要更新的字段
  ///
  /// 返回：用户对象（Map），否则为 `null`。
  Future<Map<String, dynamic>?> updateMe(Map<String, dynamic> payload) async {
    final data = await _http.request('PUT', '/users/me', auth: true, body: payload);
    return (data is Map<String, dynamic>) ? data : null;
  }

  /// 注册账号。
  ///
  /// 路由：POST `/users/register`
  Future<dynamic> register(Map<String, dynamic> payload) async {
    return _http.request('POST', '/users/register', body: payload);
  }

  /// 发送验证码（注册/通用）。
  ///
  /// 路由：POST `/users/send-code`
  Future<dynamic> sendCode(Map<String, dynamic> payload) async {
    return _http.request('POST', '/users/send-code', body: payload);
  }

  /// 发送重置密码验证码。
  ///
  /// 路由：POST `/users/send-reset-code`
  Future<dynamic> sendResetCode(Map<String, dynamic> payload) async {
    return _http.request('POST', '/users/send-reset-code', body: payload);
  }

  /// 重置密码。
  ///
  /// 路由：POST `/users/reset-password`
  Future<dynamic> resetPassword(Map<String, dynamic> payload) async {
    return _http.request('POST', '/users/reset-password', body: payload);
    }

  /// 发送账号激活验证码。
  ///
  /// 路由：POST `/users/send-activation-code`
  Future<dynamic> sendActivationCode(Map<String, dynamic> payload) async {
    return _http.request('POST', '/users/send-activation-code', body: payload);
  }

  /// 激活账号。
  ///
  /// 路由：POST `/users/activate`
  Future<dynamic> activateAccount(Map<String, dynamic> payload) async {
    return _http.request('POST', '/users/activate', body: payload);
  }
}
