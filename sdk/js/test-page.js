import createClient from './src/index.js';

let client = null;
const state = { lastFileDetail: null };

const $ = (id) => document.getElementById(id);
const $out = () => $('output');

function now() {
  const d = new Date();
  return d.toLocaleString();
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

function bindAuthSection() {
  $('login').addEventListener('click', async () => {
    try {
      guardClient();
      const username = $('username').value;
      const password = $('password').value;
      const data = await client.auth.login({ username, password });
      log('Login success', data);
      log('Stored tokens', client.getTokens());
    } catch (err) {
      logError('Login', err);
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
  bindRegisterSection();
  bindUsersSection();
  bindFilesSection();
  // 初始化客户端（可选）
  createOrUpdateClient();
}

window.addEventListener('DOMContentLoaded', main);
