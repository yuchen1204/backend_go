# Backend JS SDK

基于 `docs/swagger.yaml` 的 JavaScript SDK，封装了用户与文件相关的 API，支持浏览器与 Node.js (>=18)。

- 基础路径：默认 `basePath = /api/v1`
- 鉴权：在需要鉴权的请求中自动注入 `Authorization: Bearer <access_token>`
- 响应：统一解包 `ResponseData`，直接返回 `data` 字段；错误抛出 `BackendApiError`
- 设备验证：支持陌生设备登录邮箱验证码二次验证功能

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

// 传统登录（无设备验证）
const loginResp = await client.auth.login({ username: 'testuser', password: 'password123' });
client.setTokens({ accessToken: loginResp.access_token, refreshToken: loginResp.refresh_token });

// 设备登录验证（自动生成设备指纹）
const deviceLoginResp = await client.auth.loginWithDevice({ 
  username: 'testuser', 
  password: 'password123' 
});

if (deviceLoginResp.verification_required) {
  // 首次登录陌生设备，需要邮箱验证码
  console.log('请查收邮件并输入验证码');
  
  // 输入验证码后完成登录
  const verifiedResp = await client.auth.loginWithDevice({
    username: 'testuser',
    password: 'password123',
    deviceVerifyCode: '123456' // 邮件中的验证码
  });
  
  client.setTokens({ 
    accessToken: verifiedResp.access_token, 
    refreshToken: verifiedResp.refresh_token 
  });
}

// 获取当前用户信息（需要鉴权）
const me = await client.users.me();

// 获取公开文件列表（无需鉴权）
const files = await client.files.listPublicFiles({ page: 1, page_size: 20 });

// 上传单个文件（浏览器 File 或 Blob；Node.js 18+ 支持 Blob）
const fdResult = await client.files.upload({ file: someFile, category: 'docs', is_public: true });
```

## API 概览

- auth
  - `login({ username, password })`：传统登录（无设备验证）
  - `loginWithDevice({ username, password, deviceVerifyCode?, customDeviceId?, customDeviceName?, customDeviceType? })`：设备登录验证（自动生成设备指纹）
  - `loginWithCustomDevice(payload)`：手动设备登录（完全自定义参数）
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
- 设备工具函数
  - `generateDeviceFingerprint()`：生成设备指纹
  - `getDeviceName()`：获取设备名称
  - `getDeviceType()`：获取设备类型

## 设备登录验证

SDK 支持陌生设备登录邮箱验证功能：

### 自动设备验证流程

```js
// 1. 首次登录陌生设备（自动生成设备指纹）
const result = await client.auth.loginWithDevice({
  username: 'testuser',
  password: 'password123'
});

if (result.verification_required) {
  // 2. 系统发送邮件验证码，用户输入验证码
  const verifiedResult = await client.auth.loginWithDevice({
    username: 'testuser',
    password: 'password123',
    deviceVerifyCode: '123456' // 邮件验证码
  });
  
  // 3. 验证成功，设备被标记为信任
  client.setTokens({
    accessToken: verifiedResult.access_token,
    refreshToken: verifiedResult.refresh_token
  });
}

// 4. 同设备再次登录将直接成功，无需验证
```

### 手动设备参数

```js
// 完全自定义设备信息
const customResult = await client.auth.loginWithCustomDevice({
  username: 'testuser',
  password: 'password123',
  device_id: 'my-custom-device-id',
  device_name: '我的设备',
  device_type: 'mobile'
});
```

### 设备工具函数

```js
// 生成设备指纹（基于浏览器特征）
const fingerprint = client.generateDeviceFingerprint();

// 检测设备信息
const deviceName = client.getDeviceName(); // "Windows电脑", "iPhone" 等
const deviceType = client.getDeviceType(); // "desktop", "mobile", "tablet"
```

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
