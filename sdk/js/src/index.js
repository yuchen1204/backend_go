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
  };

  const auth = {
    login: async (payload) => {
      const data = await doRequest('POST', '/users/login', { body: payload });
      // data: { access_token, refresh_token, user }
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

    // error class
    BackendApiError,

    // domain APIs
    files,
    users,
    auth,
  };
}

export default createClient;
export { createClient, BackendApiError };
