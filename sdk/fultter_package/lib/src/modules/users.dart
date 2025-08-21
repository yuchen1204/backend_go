import '../http.dart';

class UsersApi {
  final BackendHttp _http;
  UsersApi(this._http);

  Future<Map<String, dynamic>?> getById(String id) async {
    final data = await _http.request('GET', '/users/${Uri.encodeComponent(id)}');
    return (data is Map<String, dynamic>) ? data : null;
  }

  Future<Map<String, dynamic>?> getByUsername(String username) async {
    final data = await _http.request('GET', '/users/username/${Uri.encodeComponent(username)}');
    return (data is Map<String, dynamic>) ? data : null;
  }

  Future<Map<String, dynamic>?> me() async {
    final data = await _http.request('GET', '/users/me', auth: true);
    return (data is Map<String, dynamic>) ? data : null;
  }

  Future<Map<String, dynamic>?> updateMe(Map<String, dynamic> payload) async {
    final data = await _http.request('PUT', '/users/me', auth: true, body: payload);
    return (data is Map<String, dynamic>) ? data : null;
  }

  Future<dynamic> register(Map<String, dynamic> payload) async {
    return _http.request('POST', '/users/register', body: payload);
  }

  Future<dynamic> sendCode(Map<String, dynamic> payload) async {
    return _http.request('POST', '/users/send-code', body: payload);
  }

  Future<dynamic> sendResetCode(Map<String, dynamic> payload) async {
    return _http.request('POST', '/users/send-reset-code', body: payload);
  }

  Future<dynamic> resetPassword(Map<String, dynamic> payload) async {
    return _http.request('POST', '/users/reset-password', body: payload);
    }

  Future<dynamic> sendActivationCode(Map<String, dynamic> payload) async {
    return _http.request('POST', '/users/send-activation-code', body: payload);
  }

  Future<dynamic> activateAccount(Map<String, dynamic> payload) async {
    return _http.request('POST', '/users/activate', body: payload);
  }
}
