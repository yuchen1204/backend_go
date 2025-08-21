import 'dart:convert';
import 'package:web_socket_channel/web_socket_channel.dart';
import '../http.dart';

class ChatConnection {
  final WebSocketChannel _channel;
  ChatConnection(this._channel);

  Stream<dynamic> get stream => _channel.stream.map((event) {
        try {
          if (event is String) return jsonDecode(event);
          return event;
        } catch (_) {
          return event;
        }
      });

  void send({String? toUserId, String? roomId, required String content}) {
    if ((toUserId == null || toUserId.isEmpty) && (roomId == null || roomId.isEmpty)) {
      throw ArgumentError('either roomId or toUserId is required');
    }
    final payload = <String, dynamic>{'content': content};
    if (roomId != null) payload['room_id'] = roomId;
    if (toUserId != null) payload['to_user_id'] = toUserId;
    _channel.sink.add(jsonEncode(payload));
  }

  void close() {
    _channel.sink.close();
  }
}

class ChatApi {
  final String baseUrl;
  final String basePath;
  final TokenStore tokens;

  ChatApi({required this.baseUrl, required this.basePath, required this.tokens});

  String _buildWsUrl(String token) {
    final normalized = baseUrl.replaceFirst(RegExp(r'/\$'), '');
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

  ChatConnection connect({String? token}) {
    final tk = token ?? tokens.accessToken ?? '';
    final url = _buildWsUrl(tk);
    final channel = WebSocketChannel.connect(Uri.parse(url));
    return ChatConnection(channel);
  }
}
