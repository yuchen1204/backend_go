import 'package:dio/dio.dart';
import 'http.dart';
import 'modules/auth.dart';
import 'modules/users.dart';
import 'modules/files.dart';
import 'modules/chat.dart';
import 'modules/friends.dart';

/// 后端客户端入口，聚合各模块 API：`auth`、`users`、`files`、`chat`、`friends`。
///
/// 默认基础路径为 `/api/v1`，默认基础地址为 `http://localhost:8080`。
/// 支持在构造时注入初始令牌与是否自动刷新访问令牌等选项。
class BackendClient {
  final String baseUrl;
  final String basePath;
  final TokenStore _tokens = TokenStore();
  late final BackendHttp _http;

  late final AuthApi auth;
  late final UsersApi users;
  late final FilesApi files;
  late final ChatApi chat;
  late final FriendsApi friends;

  /// 构造函数。
  ///
  /// 参数：
  /// - [baseUrl] 服务地址，默认 `http://localhost:8080`
  /// - [basePath] API 基础路径，默认 `/api/v1`
  /// - [accessToken] 初始访问令牌
  /// - [refreshToken] 初始刷新令牌
  /// - [autoRefresh] 是否自动在 401 时尝试刷新令牌（默认 `true`）
  /// - [dioOptions] 传递给底层 Dio 的可选配置
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
    friends = FriendsApi(_http);
  }

  /// 更新本地令牌（局部覆盖）。
  ///
  /// 仅当参数不为 `null` 时才会更新对应值。
  void setTokens({String? accessToken, String? refreshToken}) {
    if (accessToken != null) _tokens.accessToken = accessToken;
    if (refreshToken != null) _tokens.refreshToken = refreshToken;
  }

  /// 读取当前本地令牌。
  Map<String, String?> getTokens() => {
        'accessToken': _tokens.accessToken,
        'refreshToken': _tokens.refreshToken,
      };

  /// 清空本地令牌。
  void clearTokens() {
    _tokens.accessToken = null;
    _tokens.refreshToken = null;
  }

  /// 低层 HTTP 访问封装（必要时可直接使用）。
  BackendHttp get http => _http;
}
