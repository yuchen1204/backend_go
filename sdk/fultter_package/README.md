# Backend Flutter SDK

面向 Flutter 的后端访问 SDK，提供用户认证、文件管理、用户信息、聊天 WebSocket 等能力。与同项目的 JS SDK API 形态尽量保持一致，便于多端统一调用。

目录路径：`sdk/fultter_package/`

## 安装

pubspec.yaml 中添加依赖（本地开发直接 path 依赖）：

```yaml
dependencies:
  fultter_package:
    path: ../sdk/fultter_package
```

然后执行：

```bash
flutter pub get
```

## 快速开始

```dart
import 'package:fultter_package/fultter_package.dart';

void main() async {
  final client = BackendClient(baseUrl: 'http://localhost:8080');

  // 登录
  final login = await client.auth.login({
    'username': 'testuser',
    'password': 'password123',
  });
  client.setTokens(
    accessToken: login?['access_token'],
    refreshToken: login?['refresh_token'],
  );

  // 当前用户
  final me = await client.users.me();
  print('me: $me');

  // 查询公开文件
  final list = await client.files.listPublicFiles({'page': 1, 'page_size': 20});
  print('files: $list');

  // 连接聊天（基于 access_token）
  final conn = client.chat.connect();
  final sub = conn.stream.listen((msg) => print('ws: $msg'));
  conn.send(toUserId: 'TARGET-USER-UUID', content: 'hello');

  // 结束时关闭
  await Future.delayed(Duration(seconds: 1));
  conn.close();
  await sub.cancel();
}
```

## 主要 API

- `BackendClient({ String? baseUrl, String basePath = '/api/v1', ... })`
  - 模块：`auth`、`users`、`files`、`chat`
  - Token：`setTokens() / getTokens() / clearTokens()`

### Auth 模块（`client.auth`）
- `login(payload)` 普通登录（返回 `access_token`/`refresh_token`）
- `loginWithDevice(...)` 陌生设备/带设备信息登录
- `refresh({ refreshToken? })` 刷新 Access Token
- `logout({ accessToken?, refreshToken? })`

示例：
```dart
final res = await client.auth.login({'username': 'u', 'password': 'p'});
client.setTokens(
  accessToken: res?['access_token'],
  refreshToken: res?['refresh_token'],
);
```

### Users 模块（`client.users`）
- `me()` 获取当前用户
- `updateMe(payload)` 更新当前用户
- `getById(id)` / `getByUsername(username)`
- `register(payload)` / `sendCode(payload)`
- `sendResetCode(payload)` / `resetPassword(payload)`
- `sendActivationCode(payload)` / `activateAccount(payload)`

### Files 模块（`client.files`）
- `listPublicFiles(query)` / `listMyFiles(query)`
- `getFile(id)` / `updateFile(id, payload)` / `deleteFile(id)`
- `upload({ MultipartFile file, ... })`
- `uploadMultiple({ List<MultipartFile> files, ... })`

示例（单文件上传）：
```dart
import 'package:dio/dio.dart';

final mf = await MultipartFile.fromFile('/path/to/file.png', filename: 'file.png');
final uploaded = await client.files.upload(file: mf, isPublic: true);
```

### Chat 模块（`client.chat`）
- `connect({ String? token }) -> ChatConnection`
  - `stream`：接收消息（自动尝试 JSON 解析）
  - `send({ toUserId?, roomId?, required content })`
  - `close()`

## 错误处理

- HTTP 层会在出现错误时抛出 `BackendApiError`（位于 `src/errors.dart`）。
- 401 时会尝试使用 `refresh_token` 自动刷新一次 `access_token` 并重试原请求（需在 `BackendClient` 中设置了 Tokens 且 `autoRefresh=true`）。

```dart
try {
  final me = await client.users.me();
} on BackendApiError catch (e) {
  print('api error: ${e.statusCode} ${e.message}');
}
```

## 注意事项

- 确保后端与 `basePath`、API 路由匹配（默认 `/api/v1`）。
- WebSocket 地址会基于 `baseUrl` 自动推导为 `ws(s)://.../api/v1/ws/chat?token=...`。
- 上传文件需要使用 Dio 的 `MultipartFile` 与 `FormData`。
- 若你在生产中使用，请将 `baseUrl` 配置为你的服务域名，并在 HTTPS 下使用。

## 本地开发与测试

```bash
flutter pub get
flutter test
```

如需示例 App 或更详细的演示，请在根仓库提交 issue 反馈你的需求。欢迎 PR！
