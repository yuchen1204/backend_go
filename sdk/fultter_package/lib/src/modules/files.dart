import 'package:dio/dio.dart';
import '../http.dart';

class FilesApi {
  final BackendHttp _http;
  FilesApi(this._http);

  Future<Map<String, dynamic>?> getFile(String id) async {
    final data = await _http.request('GET', '/files/${Uri.encodeComponent(id)}');
    return (data is Map<String, dynamic>) ? data : null;
  }

  Future<Map<String, dynamic>?> updateFile(String id, Map<String, dynamic> payload) async {
    final data = await _http.request('PUT', '/files/${Uri.encodeComponent(id)}', auth: true, body: payload);
    return (data is Map<String, dynamic>) ? data : null;
  }

  Future<dynamic> deleteFile(String id) async {
    return _http.request('DELETE', '/files/${Uri.encodeComponent(id)}', auth: true);
  }

  Future<dynamic> listMyFiles(Map<String, dynamic> query) async {
    return _http.request('GET', '/files/my', auth: true, query: query);
  }

  Future<dynamic> listPublicFiles(Map<String, dynamic> query) async {
    return _http.request('GET', '/files/public', query: query);
  }

  Future<dynamic> getStorages() async {
    return _http.request('GET', '/files/storages');
  }

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
