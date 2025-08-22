import 'dart:convert';
import 'package:web_socket_channel/web_socket_channel.dart';
import '../http.dart';

/// 聊天连接封装。
///
/// 提供消息流 `stream`、发送 `send` 与关闭 `close` 能力。
/// WebSocket 服务端路径为 `/ws/chat`（位于后端基础路径下）。
class ChatConnection {
  final WebSocketChannel _channel;
  ChatConnection(this._channel);

  /// 将底层 WebSocket 的事件流映射为 `dynamic`：
  /// - 文本消息尝试 `jsonDecode`
  /// - 失败或非文本则原样返回
  Stream<dynamic> get stream => _channel.stream.map((event) {
        try {
          if (event is String) return jsonDecode(event);
          return event;
        } catch (_) {
          return event;
        }
      });

  /// 发送消息到用户或房间。
  ///
  /// 要求二者其一：`toUserId` 或 `roomId`。
  ///
  /// 参数：
  /// - [toUserId] 可选，私聊目标用户
  /// - [roomId] 可选，群聊房间 ID
  /// - [content] 必填，消息内容
  void send({String? toUserId, String? roomId, required String content}) {
    if ((toUserId == null || toUserId.isEmpty) && (roomId == null || roomId.isEmpty)) {
      throw ArgumentError('either roomId or toUserId is required');
    }
    final payload = <String, dynamic>{'content': content};
    if (roomId != null) payload['room_id'] = roomId;
    if (toUserId != null) payload['to_user_id'] = toUserId;
    _channel.sink.add(jsonEncode(payload));
  }

  /// 关闭连接。
  void close() {
    _channel.sink.close();
  }
}

/// 聊天 API（WebSocket）。
///
/// 连接地址基于 `baseUrl + basePath + '/ws/chat'` 构造，并根据 HTTP/HTTPS 使用 `ws`/`wss`。
/// 支持通过访问令牌（Query `?token=...`）进行鉴权。
class ChatApi {
  final String baseUrl;
  final String basePath;
  final TokenStore tokens;

  ChatApi({required this.baseUrl, required this.basePath, required this.tokens});

  /// 根据当前配置与令牌构造 WebSocket URL。
  String _buildWsUrl(String token) {
    final normalized = baseUrl.replaceFirst(RegExp(r'/$'), '');
    final http = '$normalized$basePath/ws/chat';
    final u = Uri.parse(http);
    final scheme = (u.scheme == 'https') ? 'wss' : 'ws';
    final wsUri = Uri(
      scheme: scheme,
      host: u.host,
      port: u.hasPort ? u.port : null,
      path: u.path,
      queryParameters: token.isNotEmpty ? {'token': token} : null,
    );
    return wsUri.toString();
  }

  /// 建立聊天连接。
  ///
  /// 参数：
  /// - [token] 可选，显式指定访问令牌。不传则从 `TokenStore.accessToken` 读取，缺失则为空字符串（匿名）。
  ///
  /// 返回：已连接的 [ChatConnection]。
  ChatConnection connect({String? token}) {
    final tk = token ?? tokens.accessToken ?? '';
    final url = _buildWsUrl(tk);
    final channel = WebSocketChannel.connect(Uri.parse(url));
    return ChatConnection(channel);
  }
}
