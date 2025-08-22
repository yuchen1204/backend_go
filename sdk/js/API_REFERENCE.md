# JavaScript SDK API 参考

基于 `sdk/js/src/index.js` 的完整方法清单与说明。默认 Base URL `http://localhost:8080`，Base Path `/api/v1`。

- 认证：使用 `Authorization: Bearer <access_token>`（大小写不敏感）。
- 错误：抛出 `BackendApiError`，包含 `status`、`code`、`payload`。

## 初始化

```js
import createClient from './sdk/js/src/index.js';
const client = createClient({
  baseURL: 'http://localhost:8080',
  basePath: '/api/v1',
  accessToken: undefined,
  refreshToken: undefined,
  autoRefresh: true,
});
```

- client.setTokens({ accessToken?, refreshToken? })
- client.clearTokens()
- client.getTokens() -> { accessToken, refreshToken }

---

## Auth 模块（client.auth）

- auth.login(payload)
  - 方法/路径：POST /users/login
  - 鉴权：否
  - 入参：{ username: string, password: string }
  - 返回：{ access_token?, refresh_token?, user?, verification_required? }

- auth.loginWithDevice({ username, password, deviceVerifyCode?, customDeviceId?, customDeviceName?, customDeviceType? })
  - 方法/路径：POST /users/login
  - 鉴权：否
  - 说明：自动生成/或自定义设备信息完成登录，首次陌生设备可能返回 verification_required=true。
  - 返回：同上

- auth.loginWithCustomDevice(payload)
  - 方法/路径：POST /users/login
  - 鉴权：否
  - 入参：{ username, password, device_id?, device_name?, device_type?, device_verification_code? }

- auth.logout(payload?)
  - 方法/路径：POST /users/logout
  - 鉴权：否
  - 入参：{ access_token, refresh_token }（若未显式传入，将尝试使用 client 当前 token）
  - 返回：任意

- auth.refresh(payload?)
  - 方法/路径：POST /users/refresh
  - 鉴权：否
  - 入参：{ refresh_token }（若未显式传入，将尝试使用 client 当前 refreshToken）
  - 返回：{ access_token }

示例：

```js
// 登录并设置 Token
const login = await client.auth.login({ username: 'u', password: 'p' });
client.setTokens({ accessToken: login.access_token, refreshToken: login.refresh_token });

// 刷新 Access Token（可省略参数，默认取 client 内的 refreshToken）
const refreshed = await client.auth.refresh();
client.setTokens({ accessToken: refreshed.access_token });

// 登出（可不传，默认读取 client 内的 tokens）
await client.auth.logout();
client.clearTokens();
```

---

## Users 模块（client.users）

- users.getById(id)
  - 方法/路径：GET /users/:id
  - 鉴权：否
  - 返回：用户对象

- users.getByUsername(username)
  - 方法/路径：GET /users/username/:username
  - 鉴权：否

- users.me()
  - 方法/路径：GET /users/me
  - 鉴权：是

- users.updateMe(payload)
  - 方法/路径：PUT /users/me
  - 鉴权：是

- users.register(payload)
  - 方法/路径：POST /users/register
  - 鉴权：否

- users.sendCode(payload)
  - 方法/路径：POST /users/send-code
  - 鉴权：否

- users.sendResetCode(payload)
  - 方法/路径：POST /users/send-reset-code
  - 鉴权：否

- users.resetPassword(payload)
  - 方法/路径：POST /users/reset-password
  - 鉴权：否

- users.sendActivationCode(payload)
  - 方法/路径：POST /users/send-activation-code
  - 鉴权：否

- users.activateAccount(payload)
  - 方法/路径：POST /users/activate
  - 鉴权：否

示例：

```js
// 获取当前用户
const me = await client.users.me();

// 更新当前用户
await client.users.updateMe({ nickname: 'Alice', avatar_url: '...' });

// 注册 + 邮箱验证码 + 激活
await client.users.register({ username: 'alice', email: 'a@ex.com', password: '***' });
await client.users.sendActivationCode({ email: 'a@ex.com' });
await client.users.activateAccount({ email: 'a@ex.com', verification_code: '123456' });
```

---

## Files 模块（client.files）

- files.getFile(id, { auth?: boolean } = {})
  - 方法/路径：GET /files/:id
  - 鉴权：可选（默认否）。当需要访问私有文件时传入 `{ auth: true }`，将自动携带 `Authorization` 头。

- files.updateFile(id, payload)
  - 方法/路径：PUT /files/:id
  - 鉴权：是

- files.deleteFile(id)
  - 方法/路径：DELETE /files/:id
  - 鉴权：是

- files.listMyFiles(query?)
  - 方法/路径：GET /files/my
  - 鉴权：是
  - 查询：{ page?, page_size?, category?, storage_type?, storage_name?, is_public? }

- files.listPublicFiles(query?)
  - 方法/路径：GET /files/public
  - 鉴权：否

- files.getStorages()
  - 方法/路径：GET /files/storages
  - 鉴权：否

- files.upload({ file, storage_name?, category?, description?, is_public? })
  - 方法/路径：POST /files/upload
  - 鉴权：是
  - 体：multipart/form-data，字段如上

- files.uploadMultiple({ files, storage_name?, category?, description?, is_public? })
  - 方法/路径：POST /files/upload-multiple
  - 鉴权：是
  - 体：multipart/form-data，字段如上，`files` 为数组

示例：

```js
// 公共文件列表（分页与筛选）
const pub = await client.files.listPublicFiles({ page: 1, page_size: 20, category: 'docs' });

// 我的文件（需鉴权）
const mine = await client.files.listMyFiles({ page: 1, page_size: 20 });

// 获取公开文件
const file = await client.files.getFile('PUBLIC_FILE_ID');

// 获取私有文件（需鉴权）
const privateFile = await client.files.getFile('PRIVATE_FILE_ID', { auth: true });
await client.files.updateFile('FILE_ID', { description: 'new desc' });
await client.files.deleteFile('FILE_ID');

// 浏览器单文件上传
await client.files.upload({ file: someFile /* File|Blob */, is_public: true, category: 'images' });

// 浏览器多文件上传
await client.files.uploadMultiple({ files: [file1, file2], category: 'docs' });

// Node.js 18+（Blob）示例
import { Blob } from 'buffer';
const blob = new Blob([Buffer.from('hello')], { type: 'text/plain' });
await client.files.upload({ file: blob, category: 'text' });
```

---

## Friends 模块（client.friends）

- friends.createRequest({ receiver_id, note? })
  - 方法/路径：POST /friends/requests
  - 鉴权：是

- friends.acceptRequest(id)
  - 方法/路径：POST /friends/requests/:id/accept
  - 鉴权：是

- friends.rejectRequest(id)
  - 方法/路径：POST /friends/requests/:id/reject
  - 鉴权：是

- friends.cancelRequest(id)
  - 方法/路径：DELETE /friends/requests/:id
  - 鉴权：是

- friends.listFriends({ page?, limit?, search? })
  - 方法/路径：GET /friends/list
  - 鉴权：是

- friends.listIncoming({ page?, limit?, status? })
  - 方法/路径：GET /friends/requests/incoming
  - 鉴权：是

- friends.listOutgoing({ page?, limit?, status? })
  - 方法/路径：GET /friends/requests/outgoing
  - 鉴权：是

- friends.updateRemark(friend_id, remark)
  - 方法/路径：PATCH /friends/remarks/:friend_id
  - 鉴权：是

- friends.deleteFriend(friend_id)
  - 方法/路径：DELETE /friends/:friend_id
  - 鉴权：是

- friends.block(user_id)
  - 方法/路径：POST /friends/blocks/:user_id
  - 鉴权：是

- friends.unblock(user_id)
  - 方法/路径：DELETE /friends/blocks/:user_id
  - 鉴权：是

- friends.listBlocks({ page?, limit? })
  - 方法/路径：GET /friends/blocks
  - 鉴权：是

示例：

```js
// 发送/管理好友请求
const req = await client.friends.createRequest({ receiver_id: 'USER_ID', note: 'hi' });
await client.friends.acceptRequest(req.id);
await client.friends.rejectRequest(req.id);
await client.friends.cancelRequest(req.id);

// 列表
const friends = await client.friends.listFriends({ page: 1, limit: 20, search: 'al' });
const incoming = await client.friends.listIncoming({ page: 1, limit: 20 });
const outgoing = await client.friends.listOutgoing({ page: 1, limit: 20 });

// 备注与删除
await client.friends.updateRemark('FRIEND_ID', '同事');
await client.friends.deleteFriend('FRIEND_ID');

// 拉黑管理
await client.friends.block('USER_ID');
await client.friends.unblock('USER_ID');
const blocks = await client.friends.listBlocks({ page: 1, limit: 50 });
```

---

## Chat 模块（client.chat）

- chat.connect({ token?, onOpen?, onClose?, onError?, onMessage? }) -> { socket, send, close }
  - 端点：GET /ws/chat（WS 协议，`?token=` 或 Header 方式）
  - send({ to_user_id?, room_id?, content })：二选一指定会话或目标用户

示例：

```js
const { socket, send, close } = client.chat.connect({
  // token 可省略，默认使用 client.getTokens().accessToken
  onOpen: () => console.log('ws open'),
  onClose: () => console.log('ws close'),
  onError: (e) => console.error('ws error', e),
  onMessage: (msg) => console.log('ws message', msg),
});

// 发送到目标用户（服务端将基于好友关系创建/查找会话）
send({ to_user_id: 'TARGET_USER_UUID', content: 'hi' });

// 或指定已有会话
// send({ room_id: 'ROOM_UUID', content: 'hello' });

// 关闭
close();
```

---

## 低级能力

- client._request(method, path, { query?, body?, auth?, headers?, isForm? })
- client._refreshAccessToken()：尝试使用 refresh_token 获取新的 access_token

---

## 通用约定

- 分页参数（若路由支持）：
  - `page`：从 1 开始的页码，默认 1
  - `page_size` 或 `limit`：每页数量，常见默认 10/20，上限由后端限制

- 错误结构（BackendApiError）：
  - `status`：HTTP 状态码（如 400、401、403、404、409 等）
  - `code`：后端业务码（若返回体包含）
  - `payload`：原始响应 JSON（若有）
  - 捕获示例：

```js
try {
  await client.files.deleteFile('not-exist');
} catch (e) {
  if (e.name === 'BackendApiError') {
    console.error(e.status, e.code, e.payload);
  } else {
    console.error(e);
  }
}
```
