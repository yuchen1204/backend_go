# Flutter SDK API 参考

目录：`sdk/fultter_package/`

- 基于源码 `lib/src/**` 汇总的完整方法清单与说明。
- 默认 Base URL `http://localhost:8080`，Base Path `/api/v1`。
- 认证：使用 `Authorization: Bearer <access_token>`。
- 错误：抛出 `BackendApiError`（见 `lib/src/errors.dart`）。

## 初始化

```dart
import 'package:fultter_package/fultter_package.dart';

final client = BackendClient(
  baseUrl: 'http://localhost:8080',
  // basePath: '/api/v1',
  // accessToken: '...',
  // refreshToken: '...',
  // autoRefresh: true,
);

client.setTokens(accessToken: '...', refreshToken: '...');
final tokens = client.getTokens(); // { accessToken, refreshToken }
client.clearTokens();
```

底层 HTTP：`client.http` 暴露 `BackendHttp`。

---

## Auth 模块（client.auth, 源码 `lib/src/modules/auth.dart`）

- login(payload)
  - 方法/路径：POST /users/login
  - 鉴权：否
  - 入参：`{ 'username': String, 'password': String }`
  - 返回：`Map?`，包含 `access_token?`、`refresh_token?`、`user?`、`verification_required?`

- loginWithDevice({ username, password, deviceVerificationCode?, deviceId?, deviceName?, deviceType? })
  - 方法/路径：POST /users/login
  - 鉴权：否
  - 说明：携带（或省略）设备信息进行登录；陌生设备可能返回 `verification_required: true`。

- loginWithCustomDevice(payload)
  - 方法/路径：POST /users/login
  - 鉴权：否
  - 入参示例：`{ 'username': 'u', 'password': 'p', 'device_id': '...', 'device_name': '...', 'device_type': '...', 'device_verification_code': '...' }`

- loginWithCurrentDevice({ username, password, deviceVerificationCode? })
  - 方法/路径：POST /users/login
  - 鉴权：否
  - 说明：内部调用 `getCurrentDevicePayload()` 自动采集 `{ device_id, device_name, device_type }`。

- refresh({ refreshToken? })
  - 方法/路径：POST /users/refresh
  - 鉴权：否
  - 入参：不传则使用 `client` 当前的 `refreshToken`。
  - 返回：`{ access_token }`，并自动更新到 `TokenStore`。

- logout({ accessToken?, refreshToken? })
  - 方法/路径：POST /users/logout
  - 鉴权：否
  - 入参：不传则尝试使用 `client` 当前的 token。
  - 返回：任意；同时会清空本地 tokens。

示例：

```dart
// 登录并保存 Token
final login = await client.auth.login({'username': 'u', 'password': 'p'});
client.setTokens(
  accessToken: login?['access_token'],
  refreshToken: login?['refresh_token'],
);

// 刷新 Access Token（可不传，默认取 client 的 refreshToken）
final refreshed = await client.auth.refresh();
client.setTokens(accessToken: refreshed?['access_token']);

// 登出并清理本地
await client.auth.logout();
client.clearTokens();
```

---

## Users 模块（client.users, 源码 `lib/src/modules/users.dart`）

- getById(id)
  - 方法/路径：GET /users/:id
  - 鉴权：否
  - 返回：`Map?` 用户信息

- getByUsername(username)
  - 方法/路径：GET /users/username/:username
  - 鉴权：否

- me()
  - 方法/路径：GET /users/me
  - 鉴权：是

- updateMe(payload)
  - 方法/路径：PUT /users/me
  - 鉴权：是

- register(payload)
  - 方法/路径：POST /users/register
  - 鉴权：否

- sendCode(payload)
  - 方法/路径：POST /users/send-code
  - 鉴权：否

- sendResetCode(payload)
  - 方法/路径：POST /users/send-reset-code
  - 鉴权：否

- resetPassword(payload)
  - 方法/路径：POST /users/reset-password
  - 鉴权：否

- sendActivationCode(payload)
  - 方法/路径：POST /users/send-activation-code
  - 鉴权：否

- activateAccount(payload)
  - 方法/路径：POST /users/activate
  - 鉴权：否

示例：

```dart
// 获取当前用户信息
final me = await client.users.me();

// 更新当前用户信息
await client.users.updateMe({'nickname': 'Alice', 'avatar_url': '...'});

// 注册 + 发送激活验证码 + 激活
await client.users.register({'username': 'alice', 'email': 'a@ex.com', 'password': '***'});
await client.users.sendActivationCode({'email': 'a@ex.com'});
await client.users.activateAccount({'email': 'a@ex.com', 'verification_code': '123456'});
```

---

## Files 模块（client.files, 源码 `lib/src/modules/files.dart`）

- getFile(id)
  - 方法/路径：GET /files/:id
  - 鉴权：否（后端支持可选鉴权；若带上 token，可访问私有文件）
  - 返回：`Map?` 文件详情

- updateFile(id, payload)
  - 方法/路径：PUT /files/:id
  - 鉴权：是
  - 返回：`Map?`

- deleteFile(id)
  - 方法/路径：DELETE /files/:id
  - 鉴权：是

- listMyFiles(query)
  - 方法/路径：GET /files/my
  - 鉴权：是
  - 查询：`{ page?, page_size?, category?, storage_type?, storage_name?, is_public? }`

- listPublicFiles(query)
  - 方法/路径：GET /files/public
  - 鉴权：否

- getStorages()
  - 方法/路径：GET /files/storages
  - 鉴权：否

- upload({ required MultipartFile file, String? storageName, String? category, String? description, bool? isPublic })
  - 方法/路径：POST /files/upload
  - 鉴权：是
  - 体：`FormData`，字段名同上；返回：`Map?`

- uploadMultiple({ required List<MultipartFile> files, String? storageName, String? category, String? description, bool? isPublic })
  - 方法/路径：POST /files/upload-multiple
  - 鉴权：是
  - 体：`FormData`，重复字段 `files`；返回：`dynamic`

示例：

```dart
// 公共文件列表
final pub = await client.files.listPublicFiles({'page': 1, 'page_size': 20, 'category': 'docs'});

// 我的文件（需鉴权）
final mine = await client.files.listMyFiles({'page': 1, 'page_size': 20});

// 获取/更新/删除
final file = await client.files.getFile('FILE_ID');
await client.files.updateFile('FILE_ID', {'description': 'new desc'});
await client.files.deleteFile('FILE_ID');

// 单文件上传
import 'package:dio/dio.dart';
final mf = await MultipartFile.fromFile('/path/file.png', filename: 'file.png');
await client.files.upload(file: mf, isPublic: true, category: 'images');

// 多文件上传
final mf1 = await MultipartFile.fromFile('/path/a.txt', filename: 'a.txt');
final mf2 = await MultipartFile.fromFile('/path/b.txt', filename: 'b.txt');
await client.files.uploadMultiple(files: [mf1, mf2], category: 'docs');
```

---

## Chat 模块（client.chat, 源码 `lib/src/modules/chat.dart`）

- connect({ String? token }) -> ChatConnection
  - 端点：GET /ws/chat（WS 协议；Query `?token=` 或通过 Header 由服务端处理）
  - 若不传 `token`，默认使用 `client` 当前 `accessToken`。
  - 返回：`ChatConnection`，包含：
    - `stream`：消息流（尝试 JSON 解析）
    - `send({ String? toUserId, String? roomId, required String content })`：二选一指定目标
    - `close()`：关闭连接

示例：

```dart
final conn = client.chat.connect(); // 默认使用 client 的 accessToken
final sub = conn.stream.listen((msg) => print('ws: $msg'));

// 发送到目标用户
conn.send(toUserId: 'TARGET_USER_UUID', content: 'hi');

// 或发送到指定会话
// conn.send(roomId: 'ROOM_UUID', content: 'hello');

// 关闭
await Future.delayed(Duration(milliseconds: 500));
conn.close();
await sub.cancel();
```

---

## 设备工具（源码 `lib/src/device.dart`）

- DeviceHelper.getDeviceInfo() -> `Future<Map>`
  - 跨平台返回平台与设备关键信息
- DeviceHelper.computeFingerprint(info) -> `String`
  - 对设备信息按键排序拼接后计算 SHA256 Hex
- getCurrentDevicePayload() -> `Future<Map<String, String>>`
  - 标准登录载荷补全：`{ device_id, device_name, device_type }`

---

## 刷新与自动重试（源码 `lib/src/http.dart`）

- 401 且满足条件时，自动调用 `POST /users/refresh` 获取新 `access_token` 并重试一次原请求。
- 相关属性：`BackendHttp.autoRefresh`、`TokenStore.refreshToken`。

错误捕获示例：

```dart
try {
  await client.files.deleteFile('NOT_EXIST');
} on BackendApiError catch (e) {
  print('api error: ${e.statusCode} ${e.message}');
}
```

---

## 通用约定

- 分页参数（若路由支持）：
  - `page`：从 1 开始的页码，默认 1
  - `page_size` 或 `limit`：每页数量，常见默认 10/20，上限由后端限制

- 错误结构（BackendApiError，见 `lib/src/errors.dart`）：
  - `statusCode`：HTTP 状态码（如 400、401、403、404、409 等）
  - `message`：后端返回的提示信息（若有）
  - `data`：原始响应 JSON（若有）

---

## 说明

- 已支持 Friends 模块，与后端路由保持一致。
