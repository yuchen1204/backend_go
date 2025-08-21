import 'package:dio/dio.dart';
import 'http.dart';
import 'modules/auth.dart';
import 'modules/users.dart';
import 'modules/files.dart';
import 'modules/chat.dart';

class BackendClient {
  final String baseUrl;
  final String basePath;
  final TokenStore _tokens = TokenStore();
  late final BackendHttp _http;

  late final AuthApi auth;
  late final UsersApi users;
  late final FilesApi files;
  late final ChatApi chat;

  BackendClient({
    String? baseUrl,
    this.basePath = '/api/v1',
    String? accessToken,
    String? refreshToken,
    bool autoRefresh = true,
    BaseOptions? dioOptions,
  }) : baseUrl = baseUrl ?? 'http://localhost:8080' {
    _tokens.accessToken = accessToken;
    _tokens.refreshToken = refreshToken;
    // Avoid using `this.` qualifiers below by resolving to locals
    final resolvedBaseUrl = this.baseUrl;
    final resolvedBasePath = basePath;
    _http = BackendHttp(
      baseUrl: resolvedBaseUrl,
      basePath: resolvedBasePath,
      tokens: _tokens,
      autoRefresh: autoRefresh,
      options: dioOptions,
    );

    auth = AuthApi(_http, _tokens);
    users = UsersApi(_http);
    files = FilesApi(_http);
    chat = ChatApi(baseUrl: resolvedBaseUrl, basePath: resolvedBasePath, tokens: _tokens);
  }

  void setTokens({String? accessToken, String? refreshToken}) {
    if (accessToken != null) _tokens.accessToken = accessToken;
    if (refreshToken != null) _tokens.refreshToken = refreshToken;
  }

  Map<String, String?> getTokens() => {
        'accessToken': _tokens.accessToken,
        'refreshToken': _tokens.refreshToken,
      };

  void clearTokens() {
    _tokens.accessToken = null;
    _tokens.refreshToken = null;
  }

  // Low-level access if needed
  BackendHttp get http => _http;
}
