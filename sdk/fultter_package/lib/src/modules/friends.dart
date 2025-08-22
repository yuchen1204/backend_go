import '../http.dart';

/// 好友相关 API。
///
/// 涵盖：好友请求、好友列表、备注更新、删除好友、黑名单管理。
///
/// 主要路由（均位于 `/api/v1` 下）：
/// - POST `/friends/requests` 创建请求（需鉴权）
/// - POST `/friends/requests/{id}/accept` 接受（需鉴权）
/// - POST `/friends/requests/{id}/reject` 拒绝（需鉴权）
/// - DELETE `/friends/requests/{id}` 取消我发出的请求（需鉴权）
/// - GET `/friends/list` 好友列表（需鉴权）
/// - GET `/friends/requests/incoming|outgoing` 请求列表（需鉴权）
/// - PATCH `/friends/remarks/{friend_id}` 更新备注（需鉴权）
/// - DELETE `/friends/{friend_id}` 删除好友（需鉴权）
/// - POST/DELETE/GET `/friends/blocks` 黑名单（需鉴权）
class FriendsApi {
  final BackendHttp _http;
  FriendsApi(this._http);

  // Friend requests
  /// 创建好友请求。
  ///
  /// 路由：POST `/friends/requests`（需鉴权）
  ///
  /// 参数：
  /// - [receiverId] 接收方用户 ID
  /// - [note] 可选附言
  ///
  /// 返回：请求对象（Map），否则为 `null`。
  Future<Map<String, dynamic>?> createRequest({
    required String receiverId,
    String? note,
  }) async {
    final data = await _http.request(
      'POST',
      '/friends/requests',
      auth: true,
      body: {
        'receiver_id': receiverId,
        if (note != null) 'note': note,
      },
    );
    return (data is Map<String, dynamic>) ? data : null;
  }

  /// 接受收到的好友请求。
  ///
  /// 路由：POST `/friends/requests/{id}/accept`（需鉴权）
  Future<dynamic> acceptRequest(String id) async {
    return _http.request(
      'POST',
      '/friends/requests/${Uri.encodeComponent(id)}/accept',
      auth: true,
    );
  }

  /// 拒绝收到的好友请求。
  ///
  /// 路由：POST `/friends/requests/{id}/reject`（需鉴权）
  Future<dynamic> rejectRequest(String id) async {
    return _http.request(
      'POST',
      '/friends/requests/${Uri.encodeComponent(id)}/reject',
      auth: true,
    );
  }

  /// 取消我发出的好友请求。
  ///
  /// 路由：DELETE `/friends/requests/{id}`（需鉴权）
  Future<dynamic> cancelRequest(String id) async {
    return _http.request(
      'DELETE',
      '/friends/requests/${Uri.encodeComponent(id)}',
      auth: true,
    );
  }

  // Lists
  /// 获取好友列表，支持分页与搜索。
  ///
  /// 路由：GET `/friends/list`（需鉴权）
  Future<dynamic> listFriends({int? page, int? limit, String? search}) async {
    return _http.request(
      'GET',
      '/friends/list',
      auth: true,
      query: {
        if (page != null) 'page': page,
        if (limit != null) 'limit': limit,
        if (search != null) 'search': search,
      },
    );
  }

  /// 获取收到的好友请求列表，可按状态过滤。
  ///
  /// 路由：GET `/friends/requests/incoming`（需鉴权）
  Future<dynamic> listIncoming({int? page, int? limit, String? status}) async {
    return _http.request(
      'GET',
      '/friends/requests/incoming',
      auth: true,
      query: {
        if (page != null) 'page': page,
        if (limit != null) 'limit': limit,
        if (status != null) 'status': status,
      },
    );
  }

  /// 获取我发出的好友请求列表，可按状态过滤。
  ///
  /// 路由：GET `/friends/requests/outgoing`（需鉴权）
  Future<dynamic> listOutgoing({int? page, int? limit, String? status}) async {
    return _http.request(
      'GET',
      '/friends/requests/outgoing',
      auth: true,
      query: {
        if (page != null) 'page': page,
        if (limit != null) 'limit': limit,
        if (status != null) 'status': status,
      },
    );
  }

  // Friend operations
  /// 更新好友备注。
  ///
  /// 路由：PATCH `/friends/remarks/{friend_id}`（需鉴权）
  Future<dynamic> updateRemark(String friendId, String remark) async {
    return _http.request(
      'PATCH',
      '/friends/remarks/${Uri.encodeComponent(friendId)}',
      auth: true,
      body: {'remark': remark},
    );
  }

  /// 删除好友。
  ///
  /// 路由：DELETE `/friends/{friend_id}`（需鉴权）
  Future<dynamic> deleteFriend(String friendId) async {
    return _http.request(
      'DELETE',
      '/friends/${Uri.encodeComponent(friendId)}',
      auth: true,
    );
  }

  // Block list
  /// 将用户加入黑名单。
  ///
  /// 路由：POST `/friends/blocks/{user_id}`（需鉴权）
  Future<dynamic> block(String userId) async {
    return _http.request(
      'POST',
      '/friends/blocks/${Uri.encodeComponent(userId)}',
      auth: true,
    );
  }

  /// 将用户从黑名单移除。
  ///
  /// 路由：DELETE `/friends/blocks/{user_id}`（需鉴权）
  Future<dynamic> unblock(String userId) async {
    return _http.request(
      'DELETE',
      '/friends/blocks/${Uri.encodeComponent(userId)}',
      auth: true,
    );
  }

  /// 获取黑名单列表，支持分页。
  ///
  /// 路由：GET `/friends/blocks`（需鉴权）
  Future<dynamic> listBlocks({int? page, int? limit}) async {
    return _http.request(
      'GET',
      '/friends/blocks',
      auth: true,
      query: {
        if (page != null) 'page': page,
        if (limit != null) 'limit': limit,
      },
    );
  }
}
