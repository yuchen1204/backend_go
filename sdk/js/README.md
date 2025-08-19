# Backend Go - JavaScript SDK

[![npm](https://img.shields.io/badge/npm-ready-green.svg)](https://www.npmjs.com/)
[![Node.js](https://img.shields.io/badge/Node.js-18+-green.svg)](https://nodejs.org/)
[![Browser](https://img.shields.io/badge/Browser-ES6+-blue.svg)](https://caniuse.com/es6)

基于 `docs/swagger.yaml` 的 JavaScript SDK，为 Backend Go 项目提供完整的前端集成解决方案。

## ✨ 特性

- 🔐 **完整认证支持**：双Token机制 + 陌生设备验证
- 📁 **文件管理**：上传、下载、分类管理
- 🌐 **跨平台**：支持浏览器与 Node.js (>=18)
- 🛡️ **类型安全**：基于 Swagger 自动生成
- 🔄 **自动重试**：Token刷新和错误处理
- 📱 **设备指纹**：自动生成设备唯一标识

## 📦 安装

### 方式一：直接使用源码（推荐）

```bash
# 克隆项目后直接使用
git clone https://github.com/yuchen1204/backend_go.git
cd backend_go/sdk/js/
```

### 方式二：npm 安装（待发布）

```bash
# 发布到 npm 后可通过以下方式安装
npm install @backend-go/js-sdk
```

### 方式三：CDN 引入

```html
<!-- 通过 CDN 引入（适合快速原型开发） -->
<script type="module">
  import createClient from 'https://cdn.jsdelivr.net/gh/yuchen1204/backend_go@main/sdk/js/src/index.js';
</script>
```

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

## 📚 API 概览

### 🔐 认证模块 (auth)

| 方法 | 参数 | 说明 |
|------|------|------|
| `login()` | `{ username, password }` | 传统登录（无设备验证） |
| `loginWithDevice()` | `{ username, password, deviceVerifyCode?, ... }` | 智能设备登录（自动指纹） |
| `loginWithCustomDevice()` | `payload` | 自定义设备登录 |
| `logout()` | `{ access_token?, refresh_token? }` | 登出（可选参数） |
| `refresh()` | `{ refresh_token? }` | 刷新Token |

### 👤 用户模块 (users)

| 方法 | 参数 | 说明 |
|------|------|------|
| `getById()` | `id` | 根据ID获取用户 |
| `getByUsername()` | `username` | 根据用户名获取用户 |
| `me()` | - | 获取当前用户信息 |
| `updateMe()` | `payload` | 更新当前用户信息 |
| `register()` | `payload` | 用户注册 |
| `sendCode()` | `payload` | 发送注册验证码 |
| `sendResetCode()` | `payload` | 发送重置验证码 |
| `resetPassword()` | `payload` | 重置密码 |
| `sendActivationCode()` | `{ email }` | 发送激活验证码到邮箱 |
| `activateAccount()` | `{ email, verification_code }` | 使用验证码激活账号 |

### 📁 文件模块 (files)

| 方法 | 参数 | 说明 |
|------|------|------|
| `getFile()` | `id` | 获取文件详情 |
| `updateFile()` | `id, payload` | 更新文件信息 |
| `deleteFile()` | `id` | 删除文件 |
| `listMyFiles()` | `query` | 获取我的文件列表 |
| `listPublicFiles()` | `query` | 获取公开文件列表 |
| `getStorages()` | - | 获取存储配置信息 |
| `upload()` | `{ file, storage_name?, ... }` | 上传单个文件 |
| `uploadMultiple()` | `{ files, storage_name?, ... }` | 批量上传文件 |

### 📱 设备工具函数

| 方法 | 返回值 | 说明 |
|------|--------|------|
| `generateDeviceFingerprint()` | `string` | 生成设备指纹 |
| `getDeviceName()` | `string` | 获取设备名称 |
| `getDeviceType()` | `string` | 获取设备类型 |

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

## 用户自助激活

```js
import createClient from './sdk/js/src/index.js';

const client = createClient({ baseURL: 'http://localhost:8080' });

// 1) 发送激活验证码（未激活用户）
await client.users.sendActivationCode({ email: 'test@example.com' });

// 2) 用户收取邮件并输入验证码，调用激活接口
await client.users.activateAccount({
  email: 'test@example.com',
  verification_code: '123456',
});

// 成功后，用户状态变为 active，即可正常登录使用
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
