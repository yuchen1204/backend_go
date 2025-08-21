/*
  Backend JS SDK
  - Base URL: defaults to http://localhost:8080
  - Base Path: /api/v1
  - Auth: Authorization: Bearer <access_token>
  - Auto refresh access token on 401 using /users/refresh (optional)
*/

const DEFAULT_BASE_URL = 'http://localhost:8080';
const DEFAULT_BASE_PATH = '/api/v1';

class BackendApiError extends Error {
  constructor(message, { status, code, payload }) {
    super(message || 'Backend API Error');
    this.name = 'BackendApiError';
    this.status = status;
    this.code = code;
    this.payload = payload;
  }
}

function isNullish(v) {
  return v === undefined || v === null;
}

function buildURL(baseURL, basePath, path, query) {
  const url = new URL(String(baseURL).replace(/\/$/, '') + String(basePath) + path);
  if (query && typeof query === 'object') {
    Object.entries(query).forEach(([k, v]) => {
      if (!isNullish(v)) {
        url.searchParams.append(k, String(v));
      }
    });
  }
  return url.toString();
}

function ensureFetch(fetchImpl) {
  if (fetchImpl) return fetchImpl;
  if (typeof fetch !== 'undefined') return fetch;
  if (typeof globalThis !== 'undefined' && globalThis.fetch) return globalThis.fetch;
  throw new Error('No fetch implementation found. Provide options.fetch.');
}

function createClient(options = {}) {
  const {
    baseURL = DEFAULT_BASE_URL,
    basePath = DEFAULT_BASE_PATH,
    accessToken: initialAccessToken,
    refreshToken: initialRefreshToken,
    autoRefresh = true,
    fetch: fetchImpl,
    defaultHeaders = {},
  } = options;

  const $fetch = ensureFetch(fetchImpl);

  const state = {
    accessToken: initialAccessToken || undefined,
    refreshToken: initialRefreshToken || undefined,
    isRefreshing: false,
  };

  // Friends APIs (user-side)
  const friends = {
    // requests
    createRequest: ({ receiver_id, note } = {}) =>
      doRequest('POST', '/friends/requests', { auth: true, body: { receiver_id, note } }),
    acceptRequest: (id) =>
      doRequest('POST', `/friends/requests/${encodeURIComponent(id)}/accept`, { auth: true }),
    rejectRequest: (id) =>
      doRequest('POST', `/friends/requests/${encodeURIComponent(id)}/reject`, { auth: true }),
    cancelRequest: (id) =>
      doRequest('DELETE', `/friends/requests/${encodeURIComponent(id)}`, { auth: true }),

    // lists
    listFriends: ({ page, limit, search } = {}) =>
      doRequest('GET', '/friends/list', { auth: true, query: { page, limit, search } }),
    listIncoming: ({ page, limit, status } = {}) =>
      doRequest('GET', '/friends/requests/incoming', { auth: true, query: { page, limit, status } }),
    listOutgoing: ({ page, limit, status } = {}) =>
      doRequest('GET', '/friends/requests/outgoing', { auth: true, query: { page, limit, status } }),

    // friend operations
    updateRemark: (friend_id, remark) =>
      doRequest('PATCH', `/friends/remarks/${encodeURIComponent(friend_id)}`, { auth: true, body: { remark } }),
    deleteFriend: (friend_id) =>
      doRequest('DELETE', `/friends/${encodeURIComponent(friend_id)}`, { auth: true }),

    // block list
    block: (user_id) =>
      doRequest('POST', `/friends/blocks/${encodeURIComponent(user_id)}`, { auth: true }),
    unblock: (user_id) =>
      doRequest('DELETE', `/friends/blocks/${encodeURIComponent(user_id)}`, { auth: true }),
    listBlocks: ({ page, limit } = {}) =>
      doRequest('GET', '/friends/blocks', { auth: true, query: { page, limit } }),
  };

  async function unwrapResponse(resp) {
    let data;
    const text = await resp.text();
    try {
      data = text ? JSON.parse(text) : null;
    } catch (e) {
      // not JSON
      data = text;
    }

    if (!resp.ok) {
      const message = data && data.message ? data.message : `HTTP ${resp.status}`;
      const code = data && typeof data.code !== 'undefined' ? data.code : undefined;
      throw new BackendApiError(message, { status: resp.status, code, payload: data });
    }

    if (data && typeof data === 'object' && Object.prototype.hasOwnProperty.call(data, 'data')) {
      return data.data;
    }
    return data;
  }

  async function doRequest(method, path, { query, body, auth = false, headers, isForm = false } = {}, retryOn401 = true) {
    const url = buildURL(baseURL, basePath, path, query);

    const reqHeaders = { ...defaultHeaders, ...(headers || {}) };

    if (auth && state.accessToken) {
      reqHeaders['Authorization'] = `Bearer ${state.accessToken}`; // middleware expects Bearer (case-insensitive)
    }

    const init = { method, headers: reqHeaders };

    if (!isForm && body && typeof body === 'object') {
      reqHeaders['Content-Type'] = 'application/json';
      init.body = JSON.stringify(body);
    } else if (isForm && body) {
      // do not set Content-Type, let browser/node set boundary automatically
      init.body = body;
    }

    const resp = await $fetch(url, init);

    if (resp.status === 401 && auth && autoRefresh && state.refreshToken && retryOn401) {
      // try refresh once
      try {
        await refreshAccessToken();
      } catch (_) {
        // refresh failed; fall through to throw original 401
        // but prefer unwrap of original response for consistent error
        return unwrapResponse(resp);
      }
      return doRequest(method, path, { query, body, auth, headers, isForm }, false);
    }

    return unwrapResponse(resp);
  }

  async function refreshAccessToken() {
    if (!state.refreshToken) throw new Error('Missing refreshToken');
    if (state.isRefreshing) throw new Error('Refresh already in progress');
    state.isRefreshing = true;
    try {
      const data = await doRequest('POST', '/users/refresh', {
        body: { refresh_token: state.refreshToken },
        auth: false,
      }, false);
      if (!data || !data.access_token) throw new Error('Invalid refresh response');
      state.accessToken = data.access_token;
      return state.accessToken;
    } finally {
      state.isRefreshing = false;
    }
  }

  function setTokens({ accessToken, refreshToken } = {}) {
    if (!isNullish(accessToken)) state.accessToken = accessToken || undefined;
    if (!isNullish(refreshToken)) state.refreshToken = refreshToken || undefined;
  }

  function clearTokens() {
    state.accessToken = undefined;
    state.refreshToken = undefined;
  }

  function getTokens() {
    return { accessToken: state.accessToken, refreshToken: state.refreshToken };
  }

  // Files APIs
  const files = {
    getFile: (id) => doRequest('GET', `/files/${encodeURIComponent(id)}`),
    updateFile: (id, payload) => doRequest('PUT', `/files/${encodeURIComponent(id)}`, { auth: true, body: payload }),
    deleteFile: (id) => doRequest('DELETE', `/files/${encodeURIComponent(id)}`, { auth: true }),
    listMyFiles: (query) => doRequest('GET', '/files/my', { auth: true, query }),
    listPublicFiles: (query) => doRequest('GET', '/files/public', { query }),
    getStorages: () => doRequest('GET', '/files/storages'),
    upload: ({ file, storage_name, category, description, is_public } = {}) => {
      if (!file) throw new Error('file is required');
      const fd = new FormData();
      fd.append('file', file);
      if (!isNullish(storage_name)) fd.append('storage_name', String(storage_name));
      if (!isNullish(category)) fd.append('category', String(category));
      if (!isNullish(description)) fd.append('description', String(description));
      if (!isNullish(is_public)) fd.append('is_public', String(is_public));
      return doRequest('POST', '/files/upload', { auth: true, isForm: true, body: fd });
    },
    uploadMultiple: ({ files, storage_name, category, description, is_public } = {}) => {
      if (!files || !Array.isArray(files) || files.length === 0) throw new Error('files is required (non-empty array)');
      const fd = new FormData();
      for (const f of files) fd.append('files', f);
      if (!isNullish(storage_name)) fd.append('storage_name', String(storage_name));
      if (!isNullish(category)) fd.append('category', String(category));
      if (!isNullish(description)) fd.append('description', String(description));
      if (!isNullish(is_public)) fd.append('is_public', String(is_public));
      return doRequest('POST', '/files/upload-multiple', { auth: true, isForm: true, body: fd });
    },
  };

  // Device fingerprint generation utility
  function generateDeviceFingerprint() {
    const canvas = document.createElement('canvas');
    const ctx = canvas.getContext('2d');
    ctx.textBaseline = 'top';
    ctx.font = '14px Arial';
    ctx.fillText('Device fingerprint', 2, 2);
    
    const fingerprint = [
      navigator.userAgent,
      navigator.language,
      screen.width + 'x' + screen.height,
      screen.colorDepth,
      new Date().getTimezoneOffset(),
      navigator.platform,
      navigator.cookieEnabled,
      canvas.toDataURL(),
    ].join('|');
    
    // Simple hash function (for demo purposes)
    let hash = 0;
    for (let i = 0; i < fingerprint.length; i++) {
      const char = fingerprint.charCodeAt(i);
      hash = ((hash << 5) - hash) + char;
      hash = hash & hash; // Convert to 32-bit integer
    }
    return Math.abs(hash).toString(16).padStart(8, '0');
  }

  function getDeviceName() {
    const ua = navigator.userAgent;
    if (/iPhone/i.test(ua)) return 'iPhone';
    if (/iPad/i.test(ua)) return 'iPad';
    if (/Android/i.test(ua)) return 'Android设备';
    if (/Windows/i.test(ua)) return 'Windows电脑';
    if (/Mac/i.test(ua)) return 'Mac电脑';
    if (/Linux/i.test(ua)) return 'Linux设备';
    return '未知设备';
  }

  function getDeviceType() {
    const ua = navigator.userAgent;
    if (/Mobile|Android|iPhone/i.test(ua)) return 'mobile';
    if (/iPad|Tablet/i.test(ua)) return 'tablet';
    return 'desktop';
  }

  // Users & Auth APIs
  const users = {
    getById: (id) => doRequest('GET', `/users/${encodeURIComponent(id)}`),
    getByUsername: (username) => doRequest('GET', `/users/username/${encodeURIComponent(username)}`),
    me: () => doRequest('GET', '/users/me', { auth: true }),
    updateMe: (payload) => doRequest('PUT', '/users/me', { auth: true, body: payload }),
    register: (payload) => doRequest('POST', '/users/register', { body: payload }),
    sendCode: (payload) => doRequest('POST', '/users/send-code', { body: payload }),
    sendResetCode: (payload) => doRequest('POST', '/users/send-reset-code', { body: payload }),
    resetPassword: (payload) => doRequest('POST', '/users/reset-password', { body: payload }),
    sendActivationCode: (payload) => doRequest('POST', '/users/send-activation-code', { body: payload }),
    activateAccount: (payload) => doRequest('POST', '/users/activate', { body: payload }),
  };

  const auth = {
    // 传统登录（不带设备验证）
    login: async (payload) => {
      const data = await doRequest('POST', '/users/login', { body: payload });
      // data: { access_token, refresh_token, user } or { user, verification_required: true }
      if (data && data.access_token) state.accessToken = data.access_token;
      if (data && data.refresh_token) state.refreshToken = data.refresh_token;
      return data;
    },
    
    // 设备登录验证（自动生成设备指纹）
    loginWithDevice: async ({ username, password, deviceVerifyCode, customDeviceId, customDeviceName, customDeviceType } = {}) => {
      const payload = {
        username,
        password,
        device_id: customDeviceId || generateDeviceFingerprint(),
        device_name: customDeviceName || getDeviceName(),
        device_type: customDeviceType || getDeviceType(),
      };
      
      if (deviceVerifyCode) {
        payload.device_verification_code = deviceVerifyCode;
      }
      
      const data = await doRequest('POST', '/users/login', { body: payload });
      
      // 如果返回了 token，说明登录成功
      if (data && data.access_token) {
        state.accessToken = data.access_token;
        state.refreshToken = data.refresh_token;
      }
      
      return data;
    },
    
    // 手动设备登录（完全自定义参数）
    loginWithCustomDevice: async (payload) => {
      const data = await doRequest('POST', '/users/login', { body: payload });
      if (data && data.access_token) state.accessToken = data.access_token;
      if (data && data.refresh_token) state.refreshToken = data.refresh_token;
      return data;
    },
    
    logout: async (payload) => {
      const actualPayload = payload || (state.accessToken && state.refreshToken ? { access_token: state.accessToken, refresh_token: state.refreshToken } : undefined);
      if (!actualPayload) throw new Error('logout requires { access_token, refresh_token }');
      const res = await doRequest('POST', '/users/logout', { body: actualPayload });
      // clear local tokens regardless of server response success
      clearTokens();
      return res;
    },
    refresh: async (payload) => {
      const actualPayload = payload || (state.refreshToken ? { refresh_token: state.refreshToken } : undefined);
      if (!actualPayload) throw new Error('refresh requires { refresh_token }');
      const data = await doRequest('POST', '/users/refresh', { body: actualPayload });
      if (data && data.access_token) state.accessToken = data.access_token;
      return data;
    },
  };

  return {
    // config
    baseURL,
    basePath,

    // token controls
    setTokens,
    clearTokens,
    getTokens,

    // raw helpers
    _request: doRequest,
    _refreshAccessToken: refreshAccessToken,

    // device utilities
    generateDeviceFingerprint,
    getDeviceName,
    getDeviceType,

    // error class
    BackendApiError,

    // domain APIs
    files,
    users,
    auth,
    friends,
    // Chat WS module
    chat: (() => {
      function buildWSURL(baseURL, basePath, token) {
        const httpUrl = String(baseURL).replace(/\/$/, '') + String(basePath) + '/ws/chat';
        const u = new URL(httpUrl);
        const wsProto = u.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = wsProto + '//' + u.host + u.pathname + (token ? `?token=${encodeURIComponent(token)}` : '');
        return wsUrl;
      }

      function connect({ token, onOpen, onClose, onError, onMessage } = {}) {
        const tk = token || state.accessToken;
        const url = buildWSURL(baseURL, basePath, tk);
        const ws = new WebSocket(url);
        if (onOpen) ws.addEventListener('open', onOpen);
        if (onClose) ws.addEventListener('close', onClose);
        if (onError) ws.addEventListener('error', onError);
        if (onMessage) ws.addEventListener('message', (ev) => {
          try {
            const data = ev.data ? JSON.parse(ev.data) : null;
            onMessage(data);
          } catch (_) {
            onMessage(ev.data);
          }
        });
        function send({ to_user_id, room_id, content }) {
          if (!content) throw new Error('content is required');
          const payload = { content };
          if (room_id) payload.room_id = room_id;
          else if (to_user_id) payload.to_user_id = to_user_id;
          else throw new Error('either room_id or to_user_id is required');
          ws.send(JSON.stringify(payload));
        }
        function close() { try { ws.close(); } catch (_) {} }
        return { socket: ws, send, close };
      }

      return { connect };
    })(),
  };
}

export default createClient;
export { createClient, BackendApiError };
