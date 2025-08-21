// 管理员API服务层
class AdminAPIService {
    constructor() {
        this.baseURL = this.getBaseURL();
        this.token = localStorage.getItem('admin_token');
    }

    // 获取用户行为日志（按用户）
    async getUserActionLogs(userId, page = 1, limit = 10) {
        try {
            const params = new URLSearchParams({
                page: page.toString(),
                limit: limit.toString()
            });
            const response = await fetch(`${this.baseURL}/admin/users/${userId}/action-logs?${params.toString()}`, {
                method: 'GET',
                headers: this.getAuthHeaders()
            });
            const data = await response.json();
            if (response.ok) {
                return { success: true, data: data.data };
            } else {
                return { success: false, message: data.message || '获取用户行为日志失败' };
            }
        } catch (error) {
            console.error('获取用户行为日志错误:', error);
            return { success: false, message: '网络连接失败' };
        }
    }

    // 刷新管理员Token
    async refreshAdminToken() {
        try {
            const response = await fetch(`${this.baseURL}/admin/refresh-token`, {
                method: 'POST',
                headers: this.getAuthHeaders()
            });

            const data = await response.json();

            if (response.ok) {
                // 后端统一响应格式：{ code, message, data: { token } }
                const newToken = data?.data?.token;
                if (newToken) {
                    this.token = newToken;
                    localStorage.setItem('admin_token', newToken);
                    return { success: true, data: { token: newToken } };
                }
                return { success: false, message: '响应格式不正确' };
            } else if (response.status === 401) {
                this.handleUnauthorized();
                return { success: false, message: '未授权' };
            } else {
                return { success: false, message: data.message || '刷新Token失败' };
            }
        } catch (error) {
            console.error('刷新管理员Token错误:', error);
            return { success: false, message: '网络连接失败' };
        }
    }

    // 获取存储信息（本地与S3 buckets）
    async getStorageInfo() {
        try {
            const response = await fetch(`${this.baseURL}/admin/storage/info`, {
                method: 'GET',
                headers: this.getAuthHeaders()
            });
            const data = await response.json();
            if (response.ok) {
                return { success: true, data: data.data };
            }
            return { success: false, message: data.message || '获取存储信息失败' };
        } catch (e) {
            console.error('获取存储信息错误:', e);
            return { success: false, message: '网络连接失败' };
        }
    }

    // 获取文件列表（管理员）
    async getFiles({ page = 1, page_size = 10, storage_name = '', category = '', is_public = '' } = {}) {
        try {
            const params = new URLSearchParams({ page: String(page), page_size: String(page_size) });
            if (storage_name) params.append('storage_name', storage_name);
            if (category) params.append('category', category);
            if (is_public !== '' && is_public !== null && is_public !== undefined) params.append('is_public', String(is_public));

            const response = await fetch(`${this.baseURL}/admin/files?${params.toString()}`, {
                method: 'GET',
                headers: this.getAuthHeaders()
            });
            const data = await response.json();
            if (response.ok) {
                return { success: true, data: data.data };
            }
            return { success: false, message: data.message || '获取文件列表失败' };
        } catch (e) {
            console.error('获取文件列表错误:', e);
            return { success: false, message: '网络连接失败' };
        }
    }

    // 获取文件详情（管理员）
    async getFileDetail(fileId) {
        try {
            const response = await fetch(`${this.baseURL}/admin/files/${fileId}`, {
                method: 'GET',
                headers: this.getAuthHeaders()
            });
            const data = await response.json();
            if (response.ok) {
                return { success: true, data: data.data };
            }
            return { success: false, message: data.message || '获取文件详情失败' };
        } catch (e) {
            console.error('获取文件详情错误:', e);
            return { success: false, message: '网络连接失败' };
        }
    }

    // 更新文件信息（管理员）
    async updateFile(fileId, updateData) {
        try {
            const response = await fetch(`${this.baseURL}/admin/files/${fileId}`, {
                method: 'PUT',
                headers: this.getAuthHeaders(),
                body: JSON.stringify(updateData)
            });
            const data = await response.json();
            if (response.ok) {
                return { success: true, data: data.data };
            }
            return { success: false, message: data.message || '更新文件失败' };
        } catch (e) {
            console.error('更新文件错误:', e);
            return { success: false, message: '网络连接失败' };
        }
    }

    // 删除文件（管理员）
    async deleteFile(fileId) {
        try {
            const response = await fetch(`${this.baseURL}/admin/files/${fileId}`, {
                method: 'DELETE',
                headers: this.getAuthHeaders()
            });
            const data = await response.json();
            if (response.ok) {
                return { success: true, data: data.data };
            }
            return { success: false, message: data.message || '删除文件失败' };
        } catch (e) {
            console.error('删除文件错误:', e);
            return { success: false, message: '网络连接失败' };
        }
    }

    // 重置用户密码（管理员）
    async resetUserPassword(userId, newPassword) {
        try {
            const response = await fetch(`${this.baseURL}/admin/users/${userId}/password`, {
                method: 'PUT',
                headers: this.getAuthHeaders(),
                body: JSON.stringify({ new_password: newPassword })
            });

            const data = await response.json();
            if (response.ok) {
                return { success: true, data: data.data };
            } else {
                return { success: false, message: data.message || '重置密码失败' };
            }
        } catch (error) {
            console.error('重置密码错误:', error);
            return { success: false, message: '网络连接失败' };
        }
    }

    // 获取管理员操作日志列表
    async getAdminLogs(page = 1, limit = 10, adminUsername = '', action = '') {
        try {
            const params = new URLSearchParams({
                page: page.toString(),
                limit: limit.toString()
            });
            if (adminUsername) params.append('admin_username', adminUsername);
            if (action) params.append('action', action);

            const response = await fetch(`${this.baseURL}/admin/logs?${params.toString()}`, {
                method: 'GET',
                headers: this.getAuthHeaders()
            });

            const data = await response.json();
            if (response.ok) {
                return { success: true, data: data.data };
            } else {
                return { success: false, message: data.message || '获取操作日志失败' };
            }
        } catch (error) {
            console.error('获取操作日志错误:', error);
            return { success: false, message: '网络连接失败' };
        }
    }

    // 创建管理员操作日志（预留：若需前端手动记录）
    async createAdminLog(payload) {
        try {
            const response = await fetch(`${this.baseURL}/admin/logs`, {
                method: 'POST',
                headers: this.getAuthHeaders(),
                body: JSON.stringify(payload)
            });
            const data = await response.json();
            if (response.ok) {
                return { success: true, data: data.data };
            } else {
                return { success: false, message: data.message || '创建操作日志失败' };
            }
        } catch (error) {
            console.error('创建操作日志错误:', error);
            return { success: false, message: '网络连接失败' };
        }
    }

    getBaseURL() {
        // 根据当前环境自动选择API地址
        const protocol = window.location.protocol;
        const hostname = window.location.hostname;
        const port = window.location.port;
        
        // 开发环境
        if (hostname === 'localhost' || hostname === '127.0.0.1') {
            return `${protocol}//${hostname}:8080/api/v1`;
        }
        
        // 生产环境
        return `${protocol}//${hostname}${port ? ':' + port : ''}/api/v1`;
    }

    // 设置认证头
    getAuthHeaders() {
        const headers = {
            'Content-Type': 'application/json'
        };
        
        if (this.token) {
            headers['Authorization'] = `Bearer ${this.token}`;
        }
        
        return headers;
    }

    // 管理员登录
    async adminLogin(username, password) {
        try {
            const response = await fetch(`${this.baseURL}/admin/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ username, password })
            });

            const data = await response.json();
            
            if (response.ok) {
                this.token = data.data.token;
                localStorage.setItem('admin_token', this.token);
                return { success: true, data: data.data };
            } else {
                return { success: false, message: data.message || '登录失败' };
            }
        } catch (error) {
            console.error('登录错误:', error);
            return { success: false, message: '网络连接失败' };
        }
    }

    // 获取用户列表
    async getUsers(page = 1, limit = 10, search = '') {
        try {
            const params = new URLSearchParams({
                page: page.toString(),
                limit: limit.toString()
            });
            
            if (search) {
                params.append('search', search);
            }

            const response = await fetch(`${this.baseURL}/admin/users?${params}`, {
                method: 'GET',
                headers: this.getAuthHeaders()
            });

            const data = await response.json();
            
            if (response.ok) {
                return { success: true, data: data.data };
            } else {
                return { success: false, message: data.message || '获取用户列表失败' };
            }
        } catch (error) {
            console.error('获取用户列表错误:', error);
            return { success: false, message: '网络连接失败' };
        }
    }

    // 获取用户详情
    async getUserDetail(userId) {
        try {
            const response = await fetch(`${this.baseURL}/admin/users/${userId}`, {
                method: 'GET',
                headers: this.getAuthHeaders()
            });

            const data = await response.json();
            
            if (response.ok) {
                return { success: true, data: data.data };
            } else {
                return { success: false, message: data.message || '获取用户详情失败' };
            }
        } catch (error) {
            console.error('获取用户详情错误:', error);
            return { success: false, message: '网络连接失败' };
        }
    }

    // 更新用户状态
    async updateUserStatus(userId, status) {
        try {
            const response = await fetch(`${this.baseURL}/admin/users/${userId}/status`, {
                method: 'PUT',
                headers: this.getAuthHeaders(),
                body: JSON.stringify({ status })
            });

            const data = await response.json();
            
            if (response.ok) {
                return { success: true, data: data.data };
            } else {
                return { success: false, message: data.message || '更新用户状态失败' };
            }
        } catch (error) {
            console.error('更新用户状态错误:', error);
            return { success: false, message: '网络连接失败' };
        }
    }

    // 删除用户
    async deleteUser(userId) {
        try {
            const response = await fetch(`${this.baseURL}/admin/users/${userId}`, {
                method: 'DELETE',
                headers: this.getAuthHeaders()
            });

            const data = await response.json();
            
            if (response.ok) {
                return { success: true, data: data.data };
            } else {
                return { success: false, message: data.message || '删除用户失败' };
            }
        } catch (error) {
            console.error('删除用户错误:', error);
            return { success: false, message: '网络连接失败' };
        }
    }

    // 获取用户统计信息
    async getUserStats() {
        try {
            const response = await fetch(`${this.baseURL}/admin/stats/users`, {
                method: 'GET',
                headers: this.getAuthHeaders()
            });

            const data = await response.json();
            
            if (response.ok) {
                return { success: true, data: data.data };
            } else {
                return { success: false, message: data.message || '获取统计信息失败' };
            }
        } catch (error) {
            console.error('获取统计信息错误:', error);
            return { success: false, message: '网络连接失败' };
        }
    }

    // 获取网络流量统计（管理员）
    async getTrafficStats() {
        try {
            const response = await fetch(`${this.baseURL}/admin/stats/traffic`, {
                method: 'GET',
                headers: this.getAuthHeaders()
            });
            const data = await response.json();
            if (response.ok) {
                return { success: true, data: data.data };
            }
            return { success: false, message: data.message || '获取流量统计失败' };
        } catch (e) {
            console.error('获取流量统计错误:', e);
            return { success: false, message: '网络连接失败' };
        }
    }

    // 检查是否已登录
    isLoggedIn() {
        return !!this.token;
    }

    // 登出
    logout() {
        this.token = null;
        localStorage.removeItem('admin_token');
    }

    // 处理401错误（未授权）
    handleUnauthorized() {
        this.logout();
        window.location.href = 'index.html';
    }
}

// 创建全局API服务实例
const adminAPI = new AdminAPIService();
