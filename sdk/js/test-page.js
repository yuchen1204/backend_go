import createClient from './src/index.js';

let client = null;
const state = { lastFileDetail: null };

const $ = (id) => document.getElementById(id);
const $out = () => $('output');

function now() {
  const d = new Date();
  return d.toLocaleString();
}

function bindChatSection() {
  let conn = null;
  const setConnected = (on) => {
    const statusEl = $('chatStatus');
    if (statusEl) statusEl.textContent = on ? '已连接' : '未连接';
  };

  $('chatConnect').addEventListener('click', () => {
    try {
      guardClient();
      if (conn) { log('聊天已连接'); return; }
      const tokens = client.getTokens();
      conn = client.chat.connect({
        token: tokens.accessToken,
        onOpen: () => { setConnected(true); log('WS open'); },
        onClose: () => { setConnected(false); log('WS close'); conn = null; },
        onError: (e) => { logError('WS error', e); },
        onMessage: (data) => {
          log('WS message', data);
          // 自动填充 room_id 便于后续复用房间
          if (data && data.room_id && !$('chatRoomId').value) {
            $('chatRoomId').value = data.room_id;
            log('已填充 room_id', { room_id: data.room_id });
          }
        },
      });
    } catch (err) { logError('WS connect', err); }
  });

  $('chatDisconnect').addEventListener('click', () => {
    try {
      if (conn) { conn.close(); conn = null; setConnected(false); log('WS closed by client'); }
      else { log('WS 未连接'); }
    } catch (err) { logError('WS disconnect', err); }
  });

  $('chatSend').addEventListener('click', () => {
    try {
      if (!conn) return log('请先连接 WS');
      const to_user_id = $('chatToUserId').value || undefined;
      const room_id = $('chatRoomId').value || undefined;
      const content = $('chatMessage').value;
      conn.send({ to_user_id, room_id, content });
      log('WS send', { to_user_id, room_id, content });
    } catch (err) { logError('WS send', err); }
  });
}

function log(title, payload) {
  const out = $out();
  const line = `\n[${now()}] ${title}\n` + (payload ? JSON.stringify(payload, null, 2) : '');
  out.textContent = (out.textContent || '') + line + '\n';
  out.scrollTop = out.scrollHeight;
}

function logError(title, err) {
  const out = $out();
  let detail = String(err && err.message ? err.message : err);
  if (err && typeof err === 'object') {
    const obj = {
      name: err.name,
      status: err.status,
      code: err.code,
      message: err.message,
      payload: err.payload,
    };
    detail = JSON.stringify(obj, null, 2);
  }
  const line = `\n[${now()}] ${title} ERROR\n${detail}`;
  out.textContent = (out.textContent || '') + line + '\n';
  out.scrollTop = out.scrollHeight;
}

function setFileId(id) {
  if (!id) return;
  $('fileId').value = id;
  log('已填充 fileId', { id });
}

async function copyToClipboard(text) {
  try {
    await navigator.clipboard.writeText(String(text));
    log('已复制到剪贴板', String(text));
  } catch (e) {
    logError('复制到剪贴板失败', e);
  }
}

function openURL(url) {
  if (!url) return;
  window.open(String(url), '_blank');
}

function renderFiles(listElId, listResponse) {
  const el = $(listElId);
  if (!el) return;
  el.innerHTML = '';
  const files = listResponse && Array.isArray(listResponse.files) ? listResponse.files : [];
  if (!files.length) {
    el.textContent = '无数据';
    return;
  }
  for (const f of files) {
    const item = document.createElement('div');
    item.className = 'item';
    const badge = document.createElement('span');
    badge.className = `badge ${f.is_public ? 'public' : 'private'}`;
    badge.textContent = f.is_public ? '公开' : '私有';
    const idCode = document.createElement('code');
    idCode.textContent = (f.id || '').slice(0, 8) + '…';
    const nameSpan = document.createElement('span');
    nameSpan.textContent = f.original_name || '';

    const btnFill = document.createElement('button');
    btnFill.textContent = '填充ID';
    btnFill.addEventListener('click', () => setFileId(f.id));

    const btnCopy = document.createElement('button');
    btnCopy.className = 'secondary';
    btnCopy.textContent = '复制ID';
    btnCopy.addEventListener('click', () => copyToClipboard(f.id));

    const btnOpen = document.createElement('button');
    btnOpen.className = 'secondary';
    btnOpen.textContent = '打开链接';
    btnOpen.addEventListener('click', () => openURL(f.url));

    item.append(badge, idCode, nameSpan, btnFill, btnCopy, btnOpen);
    el.appendChild(item);
  }
}

function readConfig() {
  return {
    baseURL: $('baseURL').value || 'http://localhost:8080',
    autoRefresh: $('autoRefresh').checked,
  };
}

function createOrUpdateClient() {
  const { baseURL, autoRefresh } = readConfig();
  client = createClient({ baseURL, autoRefresh });
  log('Client initialized', { baseURL, basePath: client.basePath, autoRefresh });
}

function guardClient() {
  if (!client) {
    createOrUpdateClient();
  }
}

function nonEmpty(v) { return v !== undefined && v !== null && String(v).trim() !== ''; }

function bindConfigSection() {
  $('initClient').addEventListener('click', () => {
    createOrUpdateClient();
  });

  $('showTokens').addEventListener('click', () => {
    guardClient();
    log('Tokens', client.getTokens());
  });

  $('clearTokens').addEventListener('click', () => {
    guardClient();
    client.clearTokens();
    log('Tokens cleared');
  });
}

function bindActivationSection() {
  // POST /users/send-activation-code
  $('sendActivationCode').addEventListener('click', async () => {
    try {
      guardClient();
      const email = $('activationEmail').value;
      if (!nonEmpty(email)) return log('请输入邮箱');
      const data = await client.users.sendActivationCode({ email });
      log('POST /users/send-activation-code', data);
    } catch (err) { logError('POST /users/send-activation-code', err); }
  });

  // POST /users/activate
  $('activateAccount').addEventListener('click', async () => {
    try {
      guardClient();
      const email = $('activationEmail').value;
      const verification_code = $('activationCode').value;
      if (!nonEmpty(email)) return log('请输入邮箱');
      if (!nonEmpty(verification_code)) return log('请输入验证码');
      const data = await client.users.activateAccount({ email, verification_code });
      log('POST /users/activate', data);
    } catch (err) { logError('POST /users/activate', err); }
  });
}

function bindAuthSection() {
  $('login').addEventListener('click', async () => {
    try {
      guardClient();
      const username = $('username').value;
      const password = $('password').value;
      const data = await client.auth.login({ username, password });
      log('传统登录成功', data);
      log('Stored tokens', client.getTokens());
    } catch (err) {
      logError('传统登录', err);
    }
  });

  $('refresh').addEventListener('click', async () => {
    try {
      guardClient();
      const data = await client.auth.refresh();
      log('Refresh success', data);
      log('Stored tokens', client.getTokens());
    } catch (err) {
      logError('Refresh', err);
    }
  });

  $('logout').addEventListener('click', async () => {
    try {
      guardClient();
      const tokens = client.getTokens();
      await client.auth.logout(tokens.accessToken && tokens.refreshToken ? { access_token: tokens.accessToken, refresh_token: tokens.refreshToken } : undefined);
      log('Logout success');
      log('Stored tokens', client.getTokens());
    } catch (err) {
      logError('Logout', err);
    }
  });
}

function bindDeviceAuthSection() {
  // 生成设备指纹
  $('generateFingerprint').addEventListener('click', () => {
    try {
      guardClient();
      const fingerprint = client.generateDeviceFingerprint();
      $('deviceId').value = fingerprint;
      log('生成设备指纹', { fingerprint });
    } catch (err) {
      logError('生成设备指纹', err);
    }
  });

  // 检测设备信息
  $('detectDevice').addEventListener('click', () => {
    try {
      guardClient();
      const deviceName = client.getDeviceName();
      const deviceType = client.getDeviceType();
      $('deviceName').value = deviceName;
      $('deviceType').value = deviceType;
      log('检测设备信息', { deviceName, deviceType });
    } catch (err) {
      logError('检测设备信息', err);
    }
  });

  // 设备登录（第一步）
  $('loginWithDevice').addEventListener('click', async () => {
    try {
      guardClient();
      const username = $('username').value;
      const password = $('password').value;
      
      // 如果没有设备指纹，自动生成
      if (!$('deviceId').value) {
        $('generateFingerprint').click();
      }
      if (!$('deviceName').value || !$('deviceType').value) {
        $('detectDevice').click();
      }

      const data = await client.auth.loginWithDevice({ username, password });
      
      if (data.verification_required) {
        log('设备登录 - 需要验证', data);
        log('请查收邮件并输入验证码');
      } else {
        log('设备登录成功', data);
        log('Stored tokens', client.getTokens());
      }
    } catch (err) {
      logError('设备登录', err);
    }
  });

  // 验证码登录（第二步）
  $('loginWithVerifyCode').addEventListener('click', async () => {
    try {
      guardClient();
      const username = $('username').value;
      const password = $('password').value;
      const deviceVerifyCode = $('deviceVerifyCode').value;
      
      if (!deviceVerifyCode) {
        log('请输入验证码');
        return;
      }

      const data = await client.auth.loginWithDevice({ 
        username, 
        password, 
        deviceVerifyCode 
      });
      
      log('验证码登录成功', data);
      log('Stored tokens', client.getTokens());
      $('deviceVerifyCode').value = ''; // 清空验证码
    } catch (err) {
      logError('验证码登录', err);
    }
  });

  // 自定义设备登录
  $('loginWithCustomDevice').addEventListener('click', async () => {
    try {
      guardClient();
      const username = $('username').value;
      const password = $('password').value;
      const device_id = $('customDeviceId').value;
      const device_name = $('customDeviceName').value;
      const device_type = $('customDeviceType').value;
      
      if (!device_id) {
        log('请输入自定义设备ID');
        return;
      }

      const payload = { username, password, device_id };
      if (device_name) payload.device_name = device_name;
      if (device_type) payload.device_type = device_type;

      const data = await client.auth.loginWithCustomDevice(payload);
      
      if (data.verification_required) {
        log('自定义设备登录 - 需要验证', data);
      } else {
        log('自定义设备登录成功', data);
        log('Stored tokens', client.getTokens());
      }
    } catch (err) {
      logError('自定义设备登录', err);
    }
  });

  // 完整测试流程
  $('testDeviceFlow').addEventListener('click', async () => {
    try {
      guardClient();
      log('开始完整设备验证流程测试...');
      
      // 1. 生成设备指纹和信息
      $('generateFingerprint').click();
      $('detectDevice').click();
      
      // 2. 执行设备登录
      await new Promise(resolve => setTimeout(resolve, 500)); // 等待UI更新
      $('loginWithDevice').click();
      
      log('流程说明：');
      log('1. 已生成设备指纹和检测设备信息');
      log('2. 已发起设备登录请求');
      log('3. 请查收邮件获取验证码');
      log('4. 在"邮件验证码"框中输入验证码');
      log('5. 点击"提交验证码登录"完成验证');
      
    } catch (err) {
      logError('完整测试流程', err);
    }
  });

  // 清空设备信息
  $('clearDeviceInputs').addEventListener('click', () => {
    $('deviceId').value = '';
    $('deviceName').value = '';
    $('deviceType').value = '';
    $('deviceVerifyCode').value = '';
    $('customDeviceId').value = '';
    $('customDeviceName').value = '';
    $('customDeviceType').value = '';
    log('已清空所有设备信息');
  });
}

function bindUsersSection() {
  $('getMe').addEventListener('click', async () => {
    try {
      guardClient();
      const data = await client.users.me();
      log('GET /users/me', data);
    } catch (err) { logError('GET /users/me', err); }
  });

  $('getById').addEventListener('click', async () => {
    try {
      guardClient();
      const id = $('userId').value;
      const data = await client.users.getById(id);
      log('GET /users/{id}', data);
    } catch (err) { logError('GET /users/{id}', err); }
  });

  $('getByUsername').addEventListener('click', async () => {
    try {
      guardClient();
      const username = $('userNameQuery').value;
      const data = await client.users.getByUsername(username);
      log('GET /users/username/{username}', data);
    } catch (err) { logError('GET /users/username/{username}', err); }
  });

  $('updateMe').addEventListener('click', async () => {
    try {
      guardClient();
      const payload = {};
      const nickname = $('nickname').value;
      const avatar = $('avatar').value;
      const bio = $('bio').value;
      if (nonEmpty(nickname)) payload.nickname = nickname;
      if (nonEmpty(avatar)) payload.avatar = avatar;
      if (nonEmpty(bio)) payload.bio = bio;
      const data = await client.users.updateMe(payload);
      log('PUT /users/me', data);
    } catch (err) { logError('PUT /users/me', err); }
  });
}

function bindFilesSection() {
  $('listPublic').addEventListener('click', async () => {
    try {
      guardClient();
      const data = await client.files.listPublicFiles({ page: 1, page_size: 20 });
      log('GET /files/public', data);
      renderFiles('publicFilesList', data);
    } catch (err) { logError('GET /files/public', err); }
  });

  $('listMy').addEventListener('click', async () => {
    try {
      guardClient();
      const data = await client.files.listMyFiles({ page: 1, page_size: 20 });
      log('GET /files/my', data);
      renderFiles('myFilesList', data);
    } catch (err) { logError('GET /files/my', err); }
  });

  $('getFile').addEventListener('click', async () => {
    try {
      guardClient();
      const id = $('fileId').value;
      const data = await client.files.getFile(id);
      log('GET /files/{id}', data);
      state.lastFileDetail = data;
      if (!id && data && data.id) setFileId(data.id);
    } catch (err) { logError('GET /files/{id}', err); }
  });

  $('deleteFile').addEventListener('click', async () => {
    try {
      guardClient();
      const id = $('fileId').value;
      const data = await client.files.deleteFile(id);
      log('DELETE /files/{id}', data);
    } catch (err) { logError('DELETE /files/{id}', err); }
  });

  $('updateFile').addEventListener('click', async () => {
    try {
      guardClient();
      const id = $('fileId').value;
      const category = $('fileCategory').value;
      const description = $('fileDesc').value;
      const is_public = $('filePublic').checked;
      const payload = {};
      if (nonEmpty(category)) payload.category = category;
      if (nonEmpty(description)) payload.description = description;
      if ($('filePublic').indeterminate === false) payload.is_public = is_public;
      const data = await client.files.updateFile(id, payload);
      log('PUT /files/{id}', data);
    } catch (err) { logError('PUT /files/{id}', err); }
  });

  $('uploadSingle').addEventListener('click', async () => {
    try {
      guardClient();
      const fileInput = $('singleFile');
      const file = fileInput.files && fileInput.files[0];
      if (!file) return log('Upload single: 请选择文件');
      const category = $('singleCategory').value;
      const description = $('singleDesc').value;
      const storage_name = $('singleStorage').value;
      const is_public = $('singlePublic').checked;
      const data = await client.files.upload({ file, category, description, storage_name: nonEmpty(storage_name) ? storage_name : undefined, is_public });
      log('POST /files/upload', data);
      if (data && data.id) setFileId(data.id);
      if (data && data.url) openURL(data.url);
    } catch (err) { logError('POST /files/upload', err); }
  });

  $('uploadMulti').addEventListener('click', async () => {
    try {
      guardClient();
      const fileInput = $('multiFiles');
      const files = fileInput.files ? Array.from(fileInput.files) : [];
      if (!files.length) return log('Upload multi: 请选择文件');
      const category = $('multiCategory').value;
      const description = $('multiDesc').value;
      const storage_name = $('multiStorage').value;
      const is_public = $('multiPublic').checked;
      const data = await client.files.uploadMultiple({ files, category, description, storage_name: nonEmpty(storage_name) ? storage_name : undefined, is_public });
      log('POST /files/upload-multiple', data);
      if (Array.isArray(data) && data.length) {
        if (data[0].id) setFileId(data[0].id);
        if (data[0].url) openURL(data[0].url);
      }
    } catch (err) { logError('POST /files/upload-multiple', err); }
  });

  $('getStorages').addEventListener('click', async () => {
    try {
      guardClient();
      const data = await client.files.getStorages();
      log('GET /files/storages', data);
    } catch (err) { logError('GET /files/storages', err); }
  });
}

function bindFriendsSection() {
  // Create friend request
  $('createFriendRequest').addEventListener('click', async () => {
    try {
      guardClient();
      const receiver_id = $('createFriendRequestReceiverId').value;
      const note = $('createFriendRequestNote').value;
      const data = await client.friends.createRequest({ receiver_id, note: note || undefined });
      log('POST /friends/requests', data);
      if (data && data.id) $('friendRequestId').value = data.id;
    } catch (err) { logError('POST /friends/requests', err); }
  });

  // Accept
  $('acceptFriendRequest').addEventListener('click', async () => {
    try {
      guardClient();
      const id = $('friendRequestId').value;
      const data = await client.friends.acceptRequest(id);
      log('POST /friends/requests/{id}/accept', data);
    } catch (err) { logError('POST /friends/requests/{id}/accept', err); }
  });

  // Reject
  $('rejectFriendRequest').addEventListener('click', async () => {
    try {
      guardClient();
      const id = $('friendRequestId').value;
      const data = await client.friends.rejectRequest(id);
      log('POST /friends/requests/{id}/reject', data);
    } catch (err) { logError('POST /friends/requests/{id}/reject', err); }
  });

  // Cancel
  $('cancelFriendRequest').addEventListener('click', async () => {
    try {
      guardClient();
      const id = $('friendRequestId').value;
      const data = await client.friends.cancelRequest(id);
      log('DELETE /friends/requests/{id}', data);
    } catch (err) { logError('DELETE /friends/requests/{id}', err); }
  });

  // Lists
  $('listFriends').addEventListener('click', async () => {
    try {
      guardClient();
      const page = Number($('friendPage').value || '1');
      const limit = Number($('friendLimit').value || '20');
      const search = $('friendSearch').value || undefined;
      const data = await client.friends.listFriends({ page, limit, search });
      log('GET /friends/list', data);
    } catch (err) { logError('GET /friends/list', err); }
  });

  $('listIncoming').addEventListener('click', async () => {
    try {
      guardClient();
      const page = Number($('friendPage').value || '1');
      const limit = Number($('friendLimit').value || '20');
      const status = $('incomingStatus').value || undefined;
      const data = await client.friends.listIncoming({ page, limit, status });
      log('GET /friends/requests/incoming', data);
    } catch (err) { logError('GET /friends/requests/incoming', err); }
  });

  $('listOutgoing').addEventListener('click', async () => {
    try {
      guardClient();
      const page = Number($('friendPage').value || '1');
      const limit = Number($('friendLimit').value || '20');
      const status = $('outgoingStatus').value || undefined;
      const data = await client.friends.listOutgoing({ page, limit, status });
      log('GET /friends/requests/outgoing', data);
    } catch (err) { logError('GET /friends/requests/outgoing', err); }
  });

  // Friend ops
  $('updateRemarkFriend').addEventListener('click', async () => {
    try {
      guardClient();
      const friend_id = $('friendFriendId').value;
      const remark = $('friendRemark').value;
      const data = await client.friends.updateRemark(friend_id, remark);
      log('PATCH /friends/remarks/{friend_id}', data);
    } catch (err) { logError('PATCH /friends/remarks/{friend_id}', err); }
  });

  $('deleteFriend').addEventListener('click', async () => {
    try {
      guardClient();
      const friend_id = $('friendFriendId').value;
      const data = await client.friends.deleteFriend(friend_id);
      log('DELETE /friends/{friend_id}', data);
    } catch (err) { logError('DELETE /friends/{friend_id}', err); }
  });

  // Blocks
  $('blockUser').addEventListener('click', async () => {
    try {
      guardClient();
      const user_id = $('blockUserId').value;
      const data = await client.friends.block(user_id);
      log('POST /friends/blocks/{user_id}', data);
    } catch (err) { logError('POST /friends/blocks/{user_id}', err); }
  });

  $('unblockUser').addEventListener('click', async () => {
    try {
      guardClient();
      const user_id = $('blockUserId').value;
      const data = await client.friends.unblock(user_id);
      log('DELETE /friends/blocks/{user_id}', data);
    } catch (err) { logError('DELETE /friends/blocks/{user_id}', err); }
  });

  $('listBlocks').addEventListener('click', async () => {
    try {
      guardClient();
      const page = Number($('friendPage').value || '1');
      const limit = Number($('friendLimit').value || '20');
      const data = await client.friends.listBlocks({ page, limit });
      log('GET /friends/blocks', data);
    } catch (err) { logError('GET /friends/blocks', err); }
  });
}

function bindRegisterSection() {
  // POST /users/send-code
  $('sendCode').addEventListener('click', async () => {
    try {
      guardClient();
      const username = $('regUsername').value;
      const email = $('regEmail').value;
      const data = await client.users.sendCode({ username, email });
      log('POST /users/send-code', data);
    } catch (err) { logError('POST /users/send-code', err); }
  });

  // POST /users/register
  $('register').addEventListener('click', async () => {
    try {
      guardClient();
      const payload = {
        username: $('regUsername').value,
        email: $('regEmail').value,
        password: $('regPassword').value,
        verification_code: $('regCode').value,
      };
      const nickname = $('regNickname').value;
      const avatar = $('regAvatar').value;
      const bio = $('regBio').value;
      if (nonEmpty(nickname)) payload.nickname = nickname;
      if (nonEmpty(avatar)) payload.avatar = avatar;
      if (nonEmpty(bio)) payload.bio = bio;
      const data = await client.users.register(payload);
      log('POST /users/register', data);
    } catch (err) { logError('POST /users/register', err); }
  });

  // POST /users/send-reset-code
  $('sendResetCode').addEventListener('click', async () => {
    try {
      guardClient();
      const email = $('resetEmail').value;
      const data = await client.users.sendResetCode({ email });
      log('POST /users/send-reset-code', data);
    } catch (err) { logError('POST /users/send-reset-code', err); }
  });

  // POST /users/reset-password
  $('resetPassword').addEventListener('click', async () => {
    try {
      guardClient();
      const email = $('resetEmail').value;
      const new_password = $('resetNewPassword').value;
      const verification_code = $('resetCode').value;
      const data = await client.users.resetPassword({ email, new_password, verification_code });
      log('POST /users/reset-password', data);
    } catch (err) { logError('POST /users/reset-password', err); }
  });
}

function main() {
  bindConfigSection();
  bindAuthSection();
  bindDeviceAuthSection();
  bindRegisterSection();
  bindActivationSection();
  bindUsersSection();
  bindFilesSection();
  bindFriendsSection();
  bindChatSection();
  // 初始化客户端（可选）
  createOrUpdateClient();
}

window.addEventListener('DOMContentLoaded', main);
