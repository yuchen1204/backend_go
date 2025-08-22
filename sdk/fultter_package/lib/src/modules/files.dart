import 'package:dio/dio.dart';
import '../http.dart';

/// 文件相关 API 封装。
///
/// 后端路由前缀：`/api/v1`
///
/// 包含以下能力：
/// - GET `/files/{id}` 获取文件信息（可选鉴权，用于私有文件访问）
/// - PUT `/files/{id}` 更新文件信息（需鉴权）
/// - DELETE `/files/{id}` 删除文件（需鉴权）
/// - GET `/files/my` 获取我的文件列表（需鉴权）
/// - GET `/files/public` 获取公开文件列表
/// - GET `/files/storages` 获取可用存储
/// - POST `/files/upload` 上传单个文件（表单，需鉴权）
/// - POST `/files/upload-multiple` 批量上传文件（表单，需鉴权）
class FilesApi {
  final BackendHttp _http;
  FilesApi(this._http);

  /// 获取文件信息。
  ///
  /// 路由：GET `/files/{id}`
  /// - 当 `auth=true` 时，会在请求头中携带 Bearer Token，用于访问私有文件。
  ///
  /// 参数：
  /// - [id] 文件 ID
  /// - [auth] 是否携带鉴权（默认 `false`）
  ///
  /// 返回：文件对象（Map），如果响应不是对象则为 `null`。
  Future<Map<String, dynamic>?> getFile(String id, {bool auth = false}) async {
    final data = await _http.request('GET', '/files/${Uri.encodeComponent(id)}', auth: auth);
    return (data is Map<String, dynamic>) ? data : null;
  }

  /// 更新文件信息。
  ///
  /// 路由：PUT `/files/{id}`（需鉴权）
  ///
  /// 参数：
  /// - [id] 文件 ID
  /// - [payload] 待更新字段，例如 `{"description": "...", "category": "..."}`
  ///
  /// 返回：更新后的文件对象（Map），如果响应不是对象则为 `null`。
  Future<Map<String, dynamic>?> updateFile(String id, Map<String, dynamic> payload) async {
    final data = await _http.request('PUT', '/files/${Uri.encodeComponent(id)}', auth: true, body: payload);
    return (data is Map<String, dynamic>) ? data : null;
  }

  /// 删除文件。
  ///
  /// 路由：DELETE `/files/{id}`（需鉴权）
  ///
  /// 参数：
  /// - [id] 文件 ID
  ///
  /// 返回：后端原始响应 `data` 字段内容。
  Future<dynamic> deleteFile(String id) async {
    return _http.request('DELETE', '/files/${Uri.encodeComponent(id)}', auth: true);
  }

  /// 获取当前用户的文件列表。
  ///
  /// 路由：GET `/files/my`（需鉴权）
  ///
  /// 参数：
  /// - [query] 查询参数，例如：`{"page":1, "limit":20, "category":"image"}`
  ///
  /// 返回：后端原始响应 `data` 字段内容（通常为分页列表）。
  Future<dynamic> listMyFiles(Map<String, dynamic> query) async {
    return _http.request('GET', '/files/my', auth: true, query: query);
  }

  /// 获取公开文件列表。
  ///
  /// 路由：GET `/files/public`
  ///
  /// 参数：
  /// - [query] 查询参数，例如：`{"page":1, "limit":20, "category":"image"}`
  ///
  /// 返回：后端原始响应 `data` 字段内容。
  Future<dynamic> listPublicFiles(Map<String, dynamic> query) async {
    return _http.request('GET', '/files/public', query: query);
  }

  /// 获取可用存储列表。
  ///
  /// 路由：GET `/files/storages`
  ///
  /// 返回：后端原始响应 `data` 字段内容。
  Future<dynamic> getStorages() async {
    return _http.request('GET', '/files/storages');
  }

  /// 上传单个文件（表单提交）。
  ///
  /// 路由：POST `/files/upload`（需鉴权）
  ///
  /// 参数：
  /// - [file] 必填，单个文件（`MultipartFile`）
  /// - [storageName] 可选，存储名（表单字段 `storage_name`）
  /// - [category] 可选，分类（表单字段 `category`）
  /// - [description] 可选，描述（表单字段 `description`）
  /// - [isPublic] 可选，是否公开（表单字段 `is_public`）
  ///
  /// 返回：上传成功后的文件对象（Map），如果响应不是对象则为 `null`。
  Future<Map<String, dynamic>?> upload({
    required MultipartFile file,
    String? storageName,
    String? category,
    String? description,
    bool? isPublic,
  }) async {
    final form = FormData();
    form.files.add(MapEntry('file', file));
    if (storageName != null) form.fields.add(MapEntry('storage_name', storageName));
    if (category != null) form.fields.add(MapEntry('category', category));
    if (description != null) form.fields.add(MapEntry('description', description));
    if (isPublic != null) form.fields.add(MapEntry('is_public', isPublic.toString()));

    final data = await _http.request('POST', '/files/upload', auth: true, body: form, isForm: true);
    return (data is Map<String, dynamic>) ? data : null;
  }

  /// 批量上传文件（表单提交）。
  ///
  /// 路由：POST `/files/upload-multiple`（需鉴权）
  ///
  /// 参数：
  /// - [files] 必填，文件列表（每个为 `MultipartFile`）
  /// - [storageName] 可选，存储名（表单字段 `storage_name`）
  /// - [category] 可选，分类（表单字段 `category`）
  /// - [description] 可选，描述（表单字段 `description`）
  /// - [isPublic] 可选，是否公开（表单字段 `is_public`）
  ///
  /// 返回：后端原始响应 `data` 字段内容。
  Future<dynamic> uploadMultiple({
    required List<MultipartFile> files,
    String? storageName,
    String? category,
    String? description,
    bool? isPublic,
  }) async {
    final form = FormData();
    for (final f in files) {
      form.files.add(MapEntry('files', f));
    }
    if (storageName != null) form.fields.add(MapEntry('storage_name', storageName));
    if (category != null) form.fields.add(MapEntry('category', category));
    if (description != null) form.fields.add(MapEntry('description', description));
    if (isPublic != null) form.fields.add(MapEntry('is_public', isPublic.toString()));

    return _http.request('POST', '/files/upload-multiple', auth: true, body: form, isForm: true);
  }
}
