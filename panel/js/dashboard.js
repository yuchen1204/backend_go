// 仪表盘页面逻辑
class DashboardManager {
    constructor() {
        this.currentPage = 1;
        this.pageSize = 10;
        this.totalUsers = 0;
        this.searchQuery = '';
        this.users = [];
        // 日志列表状态
        this.logsCurrentPage = 1;
        this.logsPageSize = 10;
        this.logsTotal = 0;
        this.logsAdminFilter = '';
        this.logsActionFilter = '';
        this.logs = [];
        // 用户行为日志（详情模态框内）
        this.userLogsCurrentPage = 1;
        this.userLogsPageSize = 10;
        this.userLogsTotal = 0;
        this.userLogs = [];
        this.userLogsFilterAction = '';
        
        // 文件管理状态
        this.filesCurrentPage = 1;
        this.filesPageSize = 10;
        this.filesTotal = 0;
        this.files = [];
        this.storageInfo = null;
        this.fileSelectedType = 'local'; // 'local' | 's3'
        this.fileSelectedBucket = '';
        this.fileCategory = '';
        this.filePublic = '';

        this.init();
    }

    async init() {
        // 检查登录状态
        if (!adminAPI.isLoggedIn()) {
            window.location.href = 'index.html';
            return;
        }

        // 初始化页面
        await this.loadDashboard();
        this.bindEvents();

        // 根据当前hash设置初始菜单与区域
        const initial = (window.location.hash || '#dashboard').substring(1);
        this.updateActiveMenu(initial);
        this.handleNavigation(initial);

        // 监听hash变化
        window.addEventListener('hashchange', () => {
            const target = (window.location.hash || '#dashboard').substring(1);
            this.updateActiveMenu(target);
            this.handleNavigation(target);
        });

        // 文件视图事件（两个下拉：存储 + Bucket）
        const storageSelect = document.getElementById('storageSelect');
        if (storageSelect) {
            storageSelect.addEventListener('change', () => {
                const v = (storageSelect.value || 'local').toLowerCase();
                this.fileSelectedType = v === 's3' ? 's3' : 'local';
                this.fileSelectedBucket = '';
                this.populateBuckets();
                this.filesCurrentPage = 1;
                this.loadFiles();
            });
        }

        const bucketSelect = document.getElementById('bucketSelect');
        if (bucketSelect) {
            bucketSelect.addEventListener('change', () => {
                this.fileSelectedBucket = bucketSelect.value || '';
                this.filesCurrentPage = 1;
                this.loadFiles();
            });
        }
    }

    // 绑定事件监听器
    bindEvents() {
        // 搜索功能
        const searchInput = document.getElementById('userSearch');
        const searchBtn = searchInput.nextElementSibling;
        
        searchBtn.addEventListener('click', () => this.searchUsers());
        searchInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                this.searchUsers();
            }
        });

        // 分页按钮
        document.getElementById('prevPage').addEventListener('click', () => this.changePage(-1));
        document.getElementById('nextPage').addEventListener('click', () => this.changePage(1));

        // 退出登录
        document.getElementById('logoutBtn').addEventListener('click', () => this.logout());

        // 侧边栏导航（兼容点击图标/文字）
        document.querySelectorAll('.sidebar-menu a').forEach(link => {
            link.addEventListener('click', (e) => {
                e.preventDefault();
                const a = e.target.closest('a');
                const href = a && a.getAttribute('href');
                if (!href) return;
                const target = href.substring(1);
                this.handleNavigation(target);
            });
        });

        // 模态框事件
        document.getElementById('updateStatusBtn').addEventListener('click', () => this.updateUserStatus());
        document.getElementById('deleteUserBtn').addEventListener('click', () => this.deleteUser());
        const openResetBtn = document.getElementById('openResetPwdBtn');
        if (openResetBtn) openResetBtn.addEventListener('click', () => this.openResetPasswordModal());
        const confirmResetBtn = document.getElementById('confirmResetPwdBtn');
        if (confirmResetBtn) confirmResetBtn.addEventListener('click', () => this.resetUserPassword());

        // 点击模态框外部关闭
        window.addEventListener('click', (e) => {
            const modal = document.getElementById('userModal');
            if (e.target === modal) {
                this.closeUserModal();
            }
        });

        // 行为日志筛选 chips 事件
        const filters = document.getElementById('userLogsFilters');
        if (filters) {
            filters.addEventListener('click', (e) => {
                const btn = e.target.closest('.chip-filter');
                if (!btn) return;
                // 激活态切换
                filters.querySelectorAll('.chip-filter').forEach(b => b.classList.remove('active'));
                btn.classList.add('active');
                this.userLogsFilterAction = (btn.dataset.action || '').toLowerCase();
                this.renderUserActionTimeline();
            });
        }
    }

    // 加载仪表盘数据
    async loadDashboard() {
        try {
            // 加载用户统计
            await this.loadUserStats();
            // 加载文件统计（本地/S3）
            await this.loadFileStats();
            // 加载网络流量（占位：若后端无接口则显示 '-')
            await this.loadTrafficStats();
            
            // 加载用户列表
            await this.loadUsers();
            
            // 更新管理员用户名
            this.updateAdminUsername();
        } catch (error) {
            console.error('加载仪表盘失败:', error);
            this.showNotification('加载数据失败，请刷新页面重试', 'error');
        }
    }

    // 加载用户统计信息
    async loadUserStats() {
        const result = await adminAPI.getUserStats();
        
        if (result.success) {
            const stats = result.data;
            // 仅渲染活跃用户，其他卡片已移除
            const activeEl = document.getElementById('activeUsers');
            if (activeEl) activeEl.textContent = stats.active_users || 0;
        } else {
            console.error('获取用户统计失败:', result.message);
        }
    }

    // 加载文件统计（按存储类型汇总）
    async loadFileStats() {
        try {
            // 获取存储信息
            const infoRes = await adminAPI.getStorageInfo();
            if (!infoRes.success) {
                console.warn('获取存储信息失败，文件统计不可用:', infoRes.message);
                this.updateFileStatsUI('-', '-');
                return;
            }

            const data = infoRes.data || {};
            const localList = Array.isArray(data.local_storages) ? data.local_storages : [];
            const s3List = Array.isArray(data.s3_storages) ? data.s3_storages : [];

            // 并发请求每个 bucket 的 total（只需 total，不取数据，page_size 可设 1）
            const localPromises = localList.map(name => adminAPI.getFiles({ page: 1, page_size: 1, storage_name: name }));
            const s3Promises = s3List.map(name => adminAPI.getFiles({ page: 1, page_size: 1, storage_name: name }));

            const [localResults, s3Results] = await Promise.all([
                Promise.all(localPromises),
                Promise.all(s3Promises)
            ]);

            const sumTotals = (arr) => arr.reduce((acc, res) => acc + (res && res.success ? (res.data?.total || 0) : 0), 0);

            const localTotal = sumTotals(localResults);
            const s3Total = sumTotals(s3Results);

            this.updateFileStatsUI(localTotal, s3Total);
        } catch (e) {
            console.error('加载文件统计失败:', e);
            this.updateFileStatsUI('-', '-');
        }
    }

    updateFileStatsUI(localVal, s3Val) {
        const elLocal = document.getElementById('totalFilesLocal');
        const elS3 = document.getElementById('totalFilesS3');
        if (elLocal) elLocal.textContent = localVal;
        if (elS3) elS3.textContent = s3Val;
    }

    // 加载网络流量（若无后端接口则占位为 '-'）
    async loadTrafficStats() {
        const inEl = document.getElementById('trafficIn');
        const outEl = document.getElementById('trafficOut');
        try {
            const res = await adminAPI.getTrafficStats();
            if (res.success) {
                const data = res.data || {};
                const inVal = typeof data.in_bytes === 'number' ? data.in_bytes : 0;
                const outVal = typeof data.out_bytes === 'number' ? data.out_bytes : 0;
                if (inEl) inEl.textContent = this.formatBytes(inVal);
                if (outEl) outEl.textContent = this.formatBytes(outVal);
                return;
            }
        } catch (e) {
            console.warn('获取流量统计失败:', e);
        }
        if (inEl) inEl.textContent = '-';
        if (outEl) outEl.textContent = '-';
    }

    // 加载用户列表
    async loadUsers() {
        const result = await adminAPI.getUsers(this.currentPage, this.pageSize, this.searchQuery);
        
        if (result.success) {
            this.users = result.data.users || [];
            this.totalUsers = result.data.total || 0;
            this.renderUsersTable();
            this.updatePagination();
        } else {
            this.showNotification(result.message || '加载用户列表失败', 'error');
        }
    }

    // 渲染用户表格
    renderUsersTable() {
        const tbody = document.getElementById('usersTableBody');
        tbody.innerHTML = '';

        if (this.users.length === 0) {
            tbody.innerHTML = '<tr><td colspan="6" style="text-align: center; padding: 20px;">暂无用户数据</td></tr>';
            return;
        }

        this.users.forEach(user => {
            const row = document.createElement('tr');
            row.innerHTML = `
                <td>${user.id.substring(0, 8)}...</td>
                <td>${user.username}</td>
                <td>${user.email}</td>
                <td><span class="status-badge status-${user.status}">${this.getStatusText(user.status)}</span></td>
                <td>${this.formatDate(user.created_at)}</td>
                <td>
                    <button class="btn btn-primary action-btn" onclick="window.location.href='user-detail.html?id=${user.id}'">详情</button>
                    <button class="btn btn-danger action-btn" onclick="dashboard.deleteUser('${user.id}')">删除</button>
                </td>
            `;
            tbody.appendChild(row);
        });
    }

    // 搜索用户
    searchUsers() {
        this.searchQuery = document.getElementById('userSearch').value.trim();
        this.currentPage = 1;
        this.loadUsers();
    }

    // 切换页面
    changePage(direction) {
        const totalPages = Math.ceil(this.totalUsers / this.pageSize);
        const newPage = this.currentPage + direction;
        
        if (newPage >= 1 && newPage <= totalPages) {
            this.currentPage = newPage;
            this.loadUsers();
        }
    }

    // 更新分页信息
    updatePagination() {
        const totalPages = Math.ceil(this.totalUsers / this.pageSize);
        document.getElementById('pageInfo').textContent = `第 ${this.currentPage} 页 / 共 ${totalPages} 页`;
        
        document.getElementById('prevPage').disabled = this.currentPage === 1;
        document.getElementById('nextPage').disabled = this.currentPage === totalPages;
    }

    // 加载操作日志
    async loadLogs() {
        const result = await adminAPI.getAdminLogs(
            this.logsCurrentPage,
            this.logsPageSize,
            this.logsAdminFilter,
            this.logsActionFilter
        );
        if (result.success) {
            const data = result.data || {};
            this.logs = data.logs || [];
            this.logsTotal = data.total || 0;
            this.renderLogsTable();
            this.updateLogsPagination();
        } else {
            this.showNotification(result.message || '加载操作日志失败', 'error');
        }
    }

    // 渲染日志表格
    renderLogsTable() {
        const tbody = document.getElementById('logsTableBody');
        if (!tbody) return;
        tbody.innerHTML = '';
        if (!this.logs || this.logs.length === 0) {
            tbody.innerHTML = '<tr><td colspan="7" style="text-align:center;padding:20px;">暂无日志记录</td></tr>';
            return;
        }
        this.logs.forEach(log => {
            const row = document.createElement('tr');
            const targetUser = log.target_user_id ? (log.target_user_id.substring(0, 8) + '...') : '-';
            const detailsText = typeof log.details === 'string' ? log.details : JSON.stringify(log.details || '');
            const details = this.escapeHTML(detailsText).slice(0, 120);
            row.innerHTML = `
                <td>${this.formatDate(log.created_at)}</td>
                <td>${this.escapeHTML(log.admin_username || '')}</td>
                <td>${this.escapeHTML(log.action || '')}</td>
                <td>${this.escapeHTML(targetUser)}</td>
                <td>${this.escapeHTML(log.ip_address || '')}</td>
                <td title="${this.escapeHTML(log.user_agent || '')}">${this.escapeHTML((log.user_agent || '').slice(0, 16))}${(log.user_agent && log.user_agent.length>16)?'...':''}</td>
                <td title="${this.escapeHTML(detailsText || '')}">${details}</td>
            `;
            tbody.appendChild(row);
        });
    }

    // 日志分页
    changeLogsPage(direction) {
        const totalPages = Math.ceil(this.logsTotal / this.logsPageSize) || 1;
        const newPage = this.logsCurrentPage + direction;
        if (newPage >= 1 && newPage <= totalPages) {
            this.logsCurrentPage = newPage;
            this.loadLogs();
        }
    }

    updateLogsPagination() {
        const totalPages = Math.ceil(this.logsTotal / this.logsPageSize) || 1;
        const info = document.getElementById('logsPageInfo');
        if (info) info.textContent = `第 ${this.logsCurrentPage} 页 / 共 ${totalPages} 页`;
        const prev = document.getElementById('logsPrevPage');
        const next = document.getElementById('logsNextPage');
        if (prev) prev.disabled = this.logsCurrentPage === 1;
        if (next) next.disabled = this.logsCurrentPage === totalPages;
    }

    // 日志筛选
    searchLogs() {
        const adminInput = document.getElementById('logAdminSearch');
        const actionInput = document.getElementById('logActionSearch');
        this.logsAdminFilter = (adminInput?.value || '').trim();
        this.logsActionFilter = (actionInput?.value || '').trim();
        this.logsCurrentPage = 1;
        this.loadLogs();
    }

    resetLogFilters() {
        const adminInput = document.getElementById('logAdminSearch');
        const actionInput = document.getElementById('logActionSearch');
        if (adminInput) adminInput.value = '';
        if (actionInput) actionInput.value = '';
        this.logsAdminFilter = '';
        this.logsActionFilter = '';
        this.logsCurrentPage = 1;
        this.loadLogs();
    }

    // ====== 文件管理：加载与渲染 ======
    async ensureStorageLoadedThenLoadFiles() {
        try {
            if (!this.storageInfo) {
                const res = await adminAPI.getStorageInfo();
                if (res.success) {
                    this.storageInfo = res.data || {};
                    this.populateBuckets();
                } else {
                    this.showNotification(res.message || '获取存储信息失败', 'error');
                    return;
                }
            }
            await this.loadFiles();
        } catch (e) {
            this.showNotification('加载存储信息失败', 'error');
        }
    }

    populateBuckets() {
        const select = document.getElementById('bucketSelect');
        if (!select) return;
        const storageSelect = document.getElementById('storageSelect');
        if (storageSelect) storageSelect.value = this.fileSelectedType === 's3' ? 's3' : 'local';
        const local = Array.isArray(this.storageInfo?.local_storages) ? this.storageInfo.local_storages : [];
        const s3 = Array.isArray(this.storageInfo?.s3_storages) ? this.storageInfo.s3_storages : [];
        const list = this.fileSelectedType === 's3' ? s3 : local;
        const def = this.storageInfo?.default_storage || '';
        if (!this.fileSelectedBucket) {
            this.fileSelectedBucket = list.includes(def) ? def : (list[0] || '');
        } else if (!list.includes(this.fileSelectedBucket)) {
            this.fileSelectedBucket = list[0] || '';
        }
        select.innerHTML = '';
        list.forEach(name => {
            const opt = document.createElement('option');
            opt.value = name; opt.textContent = name; select.appendChild(opt);
        });
        select.value = this.fileSelectedBucket;
    }

    async loadFiles() {
        const categoryInput = document.getElementById('fileCategory');
        this.fileCategory = (categoryInput?.value || '').trim();
        const res = await adminAPI.getFiles({
            page: this.filesCurrentPage,
            page_size: this.filesPageSize,
            storage_name: this.fileSelectedBucket,
            category: this.fileCategory,
            is_public: this.filePublic
        });
        if (res.success) {
            const data = res.data || {};
            this.files = data.files || data.items || [];
            this.filesTotal = data.total || 0;
            this.renderFilesTable();
            this.updateFilesPagination();
        } else {
            this.showNotification(res.message || '加载文件列表失败', 'error');
        }
    }

    renderFilesTable() {
        const tbody = document.getElementById('filesTableBody');
        if (!tbody) return;
        tbody.innerHTML = '';
        if (!this.files || this.files.length === 0) {
            tbody.innerHTML = '<tr><td colspan="9" style="text-align:center;padding:20px;">暂无文件</td></tr>';
            return;
        }
        const frag = document.createDocumentFragment();
        this.files.forEach(f => {
            const tr = document.createElement('tr');
            const sizeKB = f.size ? (Math.round((f.size / 1024) * 10) / 10) : 0;
            tr.innerHTML = `
                <td>${(f.id || '').toString().substring(0,8)}...</td>
                <td title="${this.escapeHTML(f.original_name || '')}">${this.escapeHTML((f.original_name || '').slice(0,24))}${(f.original_name||'').length>24?'…':''}</td>
                <td>${this.escapeHTML(f.storage_name || '')}</td>
                <td>${this.escapeHTML(f.category || '')}</td>
                <td>${sizeKB} KB</td>
                <td>${f.is_public ? '是' : '否'}</td>
                <td>${f.user_id ? (this.escapeHTML(f.user_id).substring(0,8)+'...') : '-'}</td>
                <td>${this.formatDate(f.created_at)}</td>
                <td>
                    <button class="btn btn-primary action-btn" onclick="window.location.href='file-detail.html?id=${f.id}'">详情</button>
                    <button class="btn btn-secondary action-btn" onclick="window.location.href='file-detail.html?id=${f.id}&edit=1'">编辑</button>
                    <button class="btn btn-danger action-btn" onclick="dashboard.deleteFileById('${f.id}')">删除</button>
                </td>
            `;
            frag.appendChild(tr);
        });
        tbody.appendChild(frag);
    }

    // 从列表删除文件
    async deleteFileById(fileId) {
        if (!fileId) return;
        if (!confirm('确定要删除该文件吗？此操作不可恢复！')) return;
        const res = await adminAPI.deleteFile(fileId);
        if (res.success) {
            this.showNotification('文件删除成功', 'success');
            // 重新加载当前页
            this.loadFiles();
        } else {
            this.showNotification(res.message || '删除失败', 'error');
        }
    }

    updateFilesPagination() {
        const totalPages = Math.ceil(this.filesTotal / this.filesPageSize) || 1;
        const info = document.getElementById('filesPageInfo');
        if (info) info.textContent = `第 ${this.filesCurrentPage} 页 / 共 ${totalPages} 页`;
        const prev = document.getElementById('filesPrevPage');
        const next = document.getElementById('filesNextPage');
        if (prev) prev.disabled = this.filesCurrentPage === 1;
        if (next) next.disabled = this.filesCurrentPage === totalPages;
    }

    changeFilesPage(direction) {
        const totalPages = Math.ceil(this.filesTotal / this.filesPageSize) || 1;
        const newPage = this.filesCurrentPage + direction;
        if (newPage >= 1 && newPage <= totalPages) {
            this.filesCurrentPage = newPage;
            this.loadFiles();
        }
    }

    // 文件筛选
    searchFiles() {
        this.filesCurrentPage = 1;
        this.loadFiles();
    }

    resetFileFilters() {
        const categoryInput = document.getElementById('fileCategory');
        if (categoryInput) categoryInput.value = '';
        this.fileCategory = '';
        // 存储重置为 local
        const storageSelect = document.getElementById('storageSelect');
        if (storageSelect) storageSelect.value = 'local';
        this.fileSelectedType = 'local';
        // bucket 回默认/首个
        this.fileSelectedBucket = '';
        this.populateBuckets();
        this.filesCurrentPage = 1;
        this.loadFiles();
    }

    // 显示用户详情
    async showUserDetail(userId) {
        const result = await adminAPI.getUserDetail(userId);
        
        if (result.success) {
            const user = result.data;
            this.renderUserDetails(user);
            // 初始化用户行为日志分页并加载
            this.userLogsCurrentPage = 1;
            this.userLogsPageSize = 10;
            this.userLogsTotal = 0;
            await this.loadUserActionLogs();
            this.openUserModal();
        } else {
            this.showNotification(result.message || '获取用户详情失败', 'error');
        }
    }

    // 渲染用户详情
    renderUserDetails(user) {
        const detailsDiv = document.getElementById('userDetails');
        detailsDiv.innerHTML = `
            <div class="user-detail-item">
                <strong>用户ID:</strong> ${user.id}
            </div>
            <div class="user-detail-item">
                <strong>用户名:</strong> ${user.username}
            </div>
            <div class="user-detail-item">
                <strong>邮箱:</strong> ${user.email}
            </div>
            <div class="user-detail-item">
                <strong>昵称:</strong> ${user.nickname || '未设置'}
            </div>
            <div class="user-detail-item">
                <strong>状态:</strong> 
                <select id="userStatusSelect" class="form-control">
                    <option value="active" ${user.status === 'active' ? 'selected' : ''}>正常</option>
                    <option value="inactive" ${user.status === 'inactive' ? 'selected' : ''}>未激活</option>
                    <option value="banned" ${user.status === 'banned' ? 'selected' : ''}>禁用</option>
                </select>
            </div>
            <div class="user-detail-item">
                <strong>注册时间:</strong> ${this.formatDate(user.created_at)}
            </div>
            <div class="user-detail-item">
                <strong>最后登录:</strong> ${user.last_login_at ? this.formatDate(user.last_login_at) : '从未登录'}
            </div>
        `;
        
        // 存储当前用户ID
        detailsDiv.dataset.userId = user.id;
    }

    // 加载某用户的行为日志
    async loadUserActionLogs() {
        const userId = document.getElementById('userDetails').dataset.userId;
        if (!userId) return;
        const result = await adminAPI.getUserActionLogs(userId, this.userLogsCurrentPage, this.userLogsPageSize);
        if (result.success) {
            const data = result.data || {};
            this.userLogs = data.logs || [];
            this.userLogsTotal = data.total || 0;
            this.renderUserActionTimeline();
            this.updateUserLogsPagination();
        } else {
            this.showNotification(result.message || '加载用户行为日志失败', 'error');
        }
    }

    // 渲染用户行为时间线
    renderUserActionTimeline() {
        const timeline = document.getElementById('userActionLogsTimeline');
        if (!timeline) return;
        timeline.innerHTML = '';
        timeline.className = 'timeline';

        // 过滤
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

            // 详情文本：尽量美化 JSON
            let detailsText = '';
            if (typeof log.details === 'string') {
                try { detailsText = JSON.stringify(JSON.parse(log.details), null, 2); } catch (_) { detailsText = log.details; }
            } else if (log.details) {
                try { detailsText = JSON.stringify(log.details, null, 2); } catch (_) { detailsText = String(log.details); }
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
        const userId = document.getElementById('userDetails').dataset.userId;
        const newStatus = document.getElementById('userStatusSelect').value;
        
        if (!userId) return;

        try {
            const result = await adminAPI.updateUserStatus(userId, newStatus);
            
            if (result.success) {
                this.showNotification('用户状态更新成功', 'success');
                this.closeUserModal();
                this.loadUsers();
                this.loadUserStats();
            } else {
                this.showNotification(result.message || '更新失败', 'error');
            }
        } catch (error) {
            this.showNotification('更新失败，请稍后重试', 'error');
        }
    }

    // 删除用户
    async deleteUser(userId = null) {
        if (!userId) {
            userId = document.getElementById('userDetails').dataset.userId;
        }
        
        if (!userId) return;

        if (!confirm('确定要删除该用户吗？此操作不可恢复！')) {
            return;
        }

        try {
            const result = await adminAPI.deleteUser(userId);
            
            if (result.success) {
                this.showNotification('用户删除成功', 'success');
                this.closeUserModal();
                this.loadUsers();
                this.loadUserStats();
            } else {
                this.showNotification(result.message || '删除失败', 'error');
            }
        } catch (error) {
            this.showNotification('删除失败，请稍后重试', 'error');
        }
    }

    // 打开用户详情模态框
    openUserModal() {
        document.getElementById('userModal').style.display = 'block';
    }

    // 关闭用户详情模态框
    closeUserModal() {
        document.getElementById('userModal').style.display = 'none';
    }

    // 打开重置密码模态框
    openResetPasswordModal() {
        const resetModal = document.getElementById('resetPwdModal');
        if (resetModal) {
            document.getElementById('newPasswordInput').value = '';
            resetModal.style.display = 'block';
        }
    }

    // 关闭重置密码模态框
    closeResetPasswordModal() {
        const resetModal = document.getElementById('resetPwdModal');
        if (resetModal) resetModal.style.display = 'none';
    }

    // 提交重置密码
    async resetUserPassword() {
        const userId = document.getElementById('userDetails').dataset.userId;
        const newPwdInput = document.getElementById('newPasswordInput');
        const newPassword = (newPwdInput?.value || '').trim();

        if (!userId) {
            this.showNotification('未找到用户ID', 'error');
            return;
        }

        if (newPassword.length < 6) {
            this.showNotification('新密码至少6位', 'error');
            return;
        }

        try {
            const result = await adminAPI.resetUserPassword(userId, newPassword);
            if (result.success) {
                this.showNotification('密码重置成功', 'success');
                this.closeResetPasswordModal();
                this.closeUserModal();
            } else {
                this.showNotification(result.message || '重置失败', 'error');
            }
        } catch (e) {
            this.showNotification('重置失败，请稍后重试', 'error');
        }
    }

    // 处理导航
    handleNavigation(target) {
        switch (target) {
            case 'dashboard':
                window.location.hash = '#dashboard';
                this.updateActiveMenu('dashboard');
                // 应显示仪表盘，而不是用户管理
                this.showSection('dashboard');
                this.loadDashboard();
                break;
            case 'users':
                window.location.hash = '#users';
                this.updateActiveMenu('users');
                this.showSection('users');
                this.loadUsers();
                break;
            case 'logs':
                window.location.hash = '#logs';
                this.updateActiveMenu('logs');
                this.showSection('logs');
                this.loadLogs();
                break;
            case 'files':
                window.location.hash = '#files';
                this.updateActiveMenu('files');
                this.showSection('files');
                this.ensureStorageLoadedThenLoadFiles();
                break;
            case 'logout':
                this.logout();
                break;
            default:
                console.log('导航到:', target);
        }
    }

    // 更新侧边栏菜单激活态
    updateActiveMenu(target) {
        const links = document.querySelectorAll('.sidebar-menu a');
        links.forEach(a => {
            const href = a.getAttribute('href') || '';
            if (href === `#${target}`) {
                a.classList.add('active');
            } else {
                a.classList.remove('active');
            }
        });
    }

    // 显示指定主区域
    showSection(section) {
        const usersSec = document.getElementById('section-users');
        const logsSec = document.getElementById('section-logs');
        const filesSec = document.getElementById('section-files');
        const statsCards = document.querySelector('.stats-cards');
        if (usersSec) usersSec.style.display = (section === 'users') ? 'block' : 'none';
        if (logsSec) logsSec.style.display = (section === 'logs') ? 'block' : 'none';
        if (filesSec) filesSec.style.display = (section === 'files') ? 'block' : 'none';
        // 仅在仪表盘视图显示统计卡片
        if (statsCards) {
            statsCards.style.display = (section === 'dashboard') ? 'grid' : 'none';
        }
    }

    // 退出登录
    logout() {
        adminAPI.logout();
        window.location.href = 'index.html';
    }

    // 更新管理员用户名
    updateAdminUsername() {
        // 从token中解析用户名（简化处理）
        const username = localStorage.getItem('admin_username') || '管理员';
        document.getElementById('adminUsername').textContent = username;
    }

    // 工具函数：格式化日期
    formatDate(dateString) {
        if (!dateString) return '未知';
        const date = new Date(dateString);
        return date.toLocaleString('zh-CN');
    }

    // 工具函数：字节转可读单位
    formatBytes(bytes) {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        const val = bytes / Math.pow(k, i);
        return `${val.toFixed(val >= 100 ? 0 : val >= 10 ? 1 : 2)} ${sizes[i]}`;
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

// 初始化仪表盘
let dashboard;
document.addEventListener('DOMContentLoaded', function() {
    dashboard = new DashboardManager();
});

// 全局函数
function closeUserModal() {
    dashboard.closeUserModal();
}

function changePage(direction) {
    dashboard.changePage(direction);
}

function searchUsers() {
    dashboard.searchUsers();
}
