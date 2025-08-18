# Backend JS SDK

基于 `docs/swagger.yaml` 的 JavaScript SDK，封装了用户与文件相关的 API，支持浏览器与 Node.js (>=18)。

- 基础路径：默认 `basePath = /api/v1`
- 鉴权：在需要鉴权的请求中自动注入 `Authorization: Bearer <access_token>`
- 响应：统一解包 `ResponseData`，直接返回 `data` 字段；错误抛出 `BackendApiError`

## 安装

本 SDK 作为源码使用，可直接引用：

```bash
# 作为子目录使用
# 路径：sdk/js/
```

或将其发布到私有 npm 仓库后安装。

## 快速开始

```js
import createClient from './sdk/js/src/index.js';

const client = createClient({
  baseURL: 'http://localhost:8080',
  accessToken: undefined, // 初次无需，登录后设置
});

// 登录
const loginResp = await client.auth.login({ username: 'testuser', password: 'password123' });
client.setTokens({ accessToken: loginResp.access_token, refreshToken: loginResp.refresh_token });

// 获取当前用户信息（需要鉴权）
const me = await client.users.me();

// 获取公开文件列表（无需鉴权）
const files = await client.files.listPublicFiles({ page: 1, page_size: 20 });

// 上传单个文件（浏览器 File 或 Blob；Node.js 18+ 支持 Blob）
const fdResult = await client.files.upload({ file: someFile, category: 'docs', is_public: true });
```

## API 概览

- auth
  - `login({ username, password })`
  - `logout({ access_token, refresh_token })`（若不传，默认使用 `client` 中存储的 token）
  - `refresh({ refresh_token })`（若不传，默认使用 `client` 中存储的 refresh token）
- users
  - `getById(id)`
  - `getByUsername(username)`
  - `me()`
  - `updateMe(payload)`
  - `register(payload)`
  - `sendCode(payload)`
  - `sendResetCode(payload)`
  - `resetPassword(payload)`
- files
  - `getFile(id)`
  - `updateFile(id, payload)`
  - `deleteFile(id)`
  - `listMyFiles(query)`
  - `listPublicFiles(query)`
  - `getStorages()`
  - `upload({ file, storage_name?, category?, description?, is_public? })`
  - `uploadMultiple({ files, storage_name?, category?, description?, is_public? })`

## Node.js 与浏览器支持

- 需要 Node.js >= 18（内置 fetch、FormData、Blob）。
- 浏览器中直接使用 `<script type="module">` 方式或打包后引入。

## 错误处理

- 所有请求失败会抛出 `BackendApiError`：
  - `error.status`：HTTP 状态码
  - `error.code`：服务端响应中的 `code`（若有）
  - `error.payload`：完整响应 JSON（若有）

## 许可

MIT
