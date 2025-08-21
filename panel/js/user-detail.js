// 用户详情页面逻辑
class UserDetailManager {
    constructor() {
        this.userId = null;
        this.user = null;
        // 用户行为日志分页
        this.userLogsCurrentPage = 1;
        this.userLogsPageSize = 10;
        this.userLogsTotal = 0;
        this.userLogs = [];
        this.userLogsFilterAction = '';
        
        this.init();
    }

    async init() {
        // 检查登录状态
        if (!adminAPI.isLoggedIn()) {
            window.location.href = 'index.html';
            return;
        }

        // 从URL参数获取用户ID
        this.userId = this.getUserIdFromUrl();
        if (!this.userId) {
            this.showNotification('未找到用户ID', 'error');
            setTimeout(() => {
                window.location.href = 'dashboard.html#users';
            }, 2000);
            return;
        }

        // 加载用户详情
        await this.loadUserDetail();
        
        // 绑定事件
        this.bindEvents();
        
        // 更新管理员用户名
        this.updateAdminUsername();
    }

    // 从URL参数获取用户ID
    getUserIdFromUrl() {
        const urlParams = new URLSearchParams(window.location.search);
        return urlParams.get('id');
    }

    // 绑定事件监听器
    bindEvents() {
        // 返回按钮
        document.getElementById('backBtn').addEventListener('click', () => {
            window.location.href = 'dashboard.html#users';
        });

        // 退出登录
        document.getElementById('logoutBtn').addEventListener('click', () => this.logout());

        // 标签页切换
        document.querySelectorAll('.tab-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const tabName = e.target.closest('.tab-btn').dataset.tab;
                this.switchTab(tabName);
            });
        });

        // 日志过滤器
        document.querySelector('.chip-group').addEventListener('click', (e) => {
            const btn = e.target.closest('.chip-filter');
            if (!btn) return;
            
            // 切换激活状态
            document.querySelectorAll('.chip-filter').forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            
            this.userLogsFilterAction = (btn.dataset.action || '').toLowerCase();
            this.renderUserActionTimeline();
        });

        // 用户日志分页
        document.getElementById('userLogsPrevPage').addEventListener('click', () => this.changeUserLogsPage(-1));
        document.getElementById('userLogsNextPage').addEventListener('click', () => this.changeUserLogsPage(1));

        // 操作按钮
        document.getElementById('updateStatusBtn').addEventListener('click', () => this.updateUserStatus());
        document.getElementById('resetPasswordBtn').addEventListener('click', () => this.resetUserPassword());
        document.getElementById('deleteUserBtn').addEventListener('click', () => this.deleteUser());
    }

    // 加载用户详情
    async loadUserDetail() {
        try {
            const result = await adminAPI.getUserDetail(this.userId);
            
            if (result.success) {
                this.user = result.data;
                this.renderUserInfo();
                await this.loadUserActionLogs();
            } else {
                this.showNotification(result.message || '获取用户详情失败', 'error');
                setTimeout(() => {
                    window.location.href = 'dashboard.html#users';
                }, 2000);
            }
        } catch (error) {
            console.error('加载用户详情失败:', error);
            this.showNotification('加载用户详情失败，请稍后重试', 'error');
        }
    }

    // 渲染用户信息
    renderUserInfo() {
        if (!this.user) return;

        // 基本信息
        document.getElementById('userDisplayName').textContent = this.user.username || '未知用户';
        document.getElementById('userEmail').textContent = this.user.email || '';
        document.getElementById('userId').textContent = `ID: ${this.user.id}`;
        
        // 状态
        const statusBadge = document.getElementById('userStatus');
        statusBadge.textContent = this.getStatusText(this.user.status);
        statusBadge.className = `status-badge status-${this.user.status}`;
        
        // 时间信息
        document.getElementById('userCreatedAt').textContent = this.formatDate(this.user.created_at);
        document.getElementById('userLastLogin').textContent = this.user.last_login_at ? this.formatDate(this.user.last_login_at) : '从未登录';
        
        // 操作表单
        document.getElementById('userStatusSelect').value = this.user.status || 'active';
    }

    // 切换标签页
    switchTab(tabName) {
        // 更新按钮状态
        document.querySelectorAll('.tab-btn').forEach(btn => {
            btn.classList.toggle('active', btn.dataset.tab === tabName);
        });
        
        // 更新内容显示
        document.querySelectorAll('.tab-content').forEach(content => {
            content.classList.toggle('active', content.id === `tab-${tabName}`);
        });
        
        // 如果切换到日志页面且还没有加载，则加载日志
        if (tabName === 'logs' && this.userLogs.length === 0) {
            this.loadUserActionLogs();
        }
    }

    // 加载用户行为日志
    async loadUserActionLogs() {
        try {
            const result = await adminAPI.getUserActionLogs(this.userId, this.userLogsCurrentPage, this.userLogsPageSize);
            
            if (result.success) {
                const data = result.data || {};
                this.userLogs = data.logs || [];
                this.userLogsTotal = data.total || 0;
                this.renderUserActionTimeline();
                this.updateUserLogsPagination();
            } else {
                this.showNotification(result.message || '加载用户行为日志失败', 'error');
            }
        } catch (error) {
            console.error('加载用户行为日志失败:', error);
            this.showNotification('加载用户行为日志失败', 'error');
        }
    }

    // 渲染用户行为时间线
    renderUserActionTimeline() {
        const timeline = document.getElementById('userActionLogsTimeline');
        if (!timeline) return;
        
        timeline.innerHTML = '';
        timeline.className = 'timeline';

        // 过滤日志
        let logs = Array.isArray(this.userLogs) ? this.userLogs.slice() : [];
        if (this.userLogsFilterAction) {
            const key = this.userLogsFilterAction;
            logs = logs.filter(l => (l.action || '').toLowerCase() === key);
        }
        
        if (!logs || logs.length === 0) {
            timeline.innerHTML = '<div class="empty">暂无匹配的活动</div>';
            return;
        }

        const frag = document.createDocumentFragment();
        logs.forEach(log => {
            const item = document.createElement('div');
            item.className = 'timeline-item activity-card';

            const actionKey = (log.action || '').toLowerCase().replace(/[^a-z0-9_\-]/g, '_');
            const iconClass = `activity-icon icon-${actionKey || 'login'}`;
            const actionLabelMap = {
                'login': '登录',
                'update_profile': '资料变更',
                'reset_password': '重置密码'
            };
            const actionLabel = actionLabelMap[actionKey] || (log.action || '活动');

            // 详情文本：格式化 JSON
            let detailsText = '';
            if (typeof log.details === 'string') {
                try { 
                    detailsText = JSON.stringify(JSON.parse(log.details), null, 2); 
                } catch (_) { 
                    detailsText = log.details; 
                }
            } else if (log.details) {
                try { 
                    detailsText = JSON.stringify(log.details, null, 2); 
                } catch (_) { 
                    detailsText = String(log.details); 
                }
            }

            const ua = log.user_agent || '';
            const uaShort = ua.slice(0, 36) + (ua.length > 36 ? '…' : '');

            item.innerHTML = `
                <div class="activity-row">
                    <div class="${iconClass}">${this.escapeHTML(actionLabel.substring(0,1))}</div>
                    <div>
                        <div class="activity-title">${this.escapeHTML(actionLabel)}</div>
                        <div class="activity-sub">${this.escapeHTML(log.username || '')}</div>
                    </div>
                    <div class="activity-time">${this.formatDate(log.created_at)}</div>
                </div>
                <div class="activity-meta">
                    <span>IP: ${this.escapeHTML(log.ip_address || '')}</span>
                    <span title="${this.escapeHTML(ua)}">UA: ${this.escapeHTML(uaShort)}</span>
                </div>
                ${detailsText ? `<pre class="code-block">${this.escapeHTML(detailsText)}</pre>` : ''}
            `;
            frag.appendChild(item);
        });
        timeline.appendChild(frag);
    }

    // 用户行为日志分页
    changeUserLogsPage(direction) {
        const totalPages = Math.ceil(this.userLogsTotal / this.userLogsPageSize) || 1;
        const newPage = this.userLogsCurrentPage + direction;
        
        if (newPage >= 1 && newPage <= totalPages) {
            this.userLogsCurrentPage = newPage;
            this.loadUserActionLogs();
        }
    }

    // 更新用户日志分页信息
    updateUserLogsPagination() {
        const totalPages = Math.ceil(this.userLogsTotal / this.userLogsPageSize) || 1;
        const info = document.getElementById('userLogsPageInfo');
        if (info) info.textContent = `第 ${this.userLogsCurrentPage} 页 / 共 ${totalPages} 页`;
        
        const prev = document.getElementById('userLogsPrevPage');
        const next = document.getElementById('userLogsNextPage');
        if (prev) prev.disabled = this.userLogsCurrentPage === 1;
        if (next) next.disabled = this.userLogsCurrentPage === totalPages;
    }

    // 更新用户状态
    async updateUserStatus() {
        const newStatus = document.getElementById('userStatusSelect').value;
        
        if (!this.userId) return;

        try {
            const result = await adminAPI.updateUserStatus(this.userId, newStatus);
            
            if (result.success) {
                this.showNotification('用户状态更新成功', 'success');
                // 重新加载用户信息
                await this.loadUserDetail();
            } else {
                this.showNotification(result.message || '更新失败', 'error');
            }
        } catch (error) {
            this.showNotification('更新失败，请稍后重试', 'error');
        }
    }

    // 重置用户密码
    async resetUserPassword() {
        const newPasswordInput = document.getElementById('newPasswordInput');
        const newPassword = (newPasswordInput?.value || '').trim();

        if (!this.userId) {
            this.showNotification('未找到用户ID', 'error');
            return;
        }

        if (newPassword.length < 6) {
            this.showNotification('新密码至少6位', 'error');
            return;
        }

        try {
            const result = await adminAPI.resetUserPassword(this.userId, newPassword);
            
            if (result.success) {
                this.showNotification('密码重置成功', 'success');
                newPasswordInput.value = '';
            } else {
                this.showNotification(result.message || '重置失败', 'error');
            }
        } catch (error) {
            this.showNotification('重置失败，请稍后重试', 'error');
        }
    }

    // 删除用户
    async deleteUser() {
        if (!this.userId) return;

        if (!confirm('确定要删除该用户吗？此操作不可恢复！')) {
            return;
        }

        try {
            const result = await adminAPI.deleteUser(this.userId);
            
            if (result.success) {
                this.showNotification('用户删除成功', 'success');
                setTimeout(() => {
                    window.location.href = 'dashboard.html#users';
                }, 1500);
            } else {
                this.showNotification(result.message || '删除失败', 'error');
            }
        } catch (error) {
            this.showNotification('删除失败，请稍后重试', 'error');
        }
    }

    // 退出登录
    logout() {
        adminAPI.logout();
        window.location.href = 'index.html';
    }

    // 更新管理员用户名
    updateAdminUsername() {
        const username = localStorage.getItem('admin_username') || '管理员';
        document.getElementById('adminUsername').textContent = username;
    }

    // 工具函数：格式化日期
    formatDate(dateString) {
        if (!dateString) return '未知';
        const date = new Date(dateString);
        return date.toLocaleString('zh-CN');
    }

    // 工具函数：转义HTML
    escapeHTML(str) {
        if (typeof str !== 'string') return '';
        return str
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#39;');
    }

    // 工具函数：获取状态文本
    getStatusText(status) {
        const statusMap = {
            'active': '正常',
            'inactive': '未激活',
            'banned': '禁用'
        };
        return statusMap[status] || status;
    }

    // 显示通知
    showNotification(message, type = 'info') {
        // 创建通知元素
        const notification = document.createElement('div');
        notification.className = `notification notification-${type}`;
        notification.textContent = message;
        
        // 添加样式
        notification.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            padding: 15px 20px;
            border-radius: 5px;
            color: white;
            font-weight: 500;
            z-index: 1001;
            animation: slideIn 0.3s ease;
        `;
        
        if (type === 'success') {
            notification.style.backgroundColor = '#28a745';
        } else if (type === 'error') {
            notification.style.backgroundColor = '#dc3545';
        } else {
            notification.style.backgroundColor = '#17a2b8';
        }
        
        document.body.appendChild(notification);
        
        // 3秒后自动移除
        setTimeout(() => {
            notification.remove();
        }, 3000);
    }
}

// 初始化用户详情页面
let userDetail;
document.addEventListener('DOMContentLoaded', function() {
    userDetail = new UserDetailManager();
});
