// 文件详情页面逻辑
class FileDetailManager {
    constructor() {
        this.fileId = null;
        this.fileData = null;
        this.init();
    }

    async init() {
        // 检查登录状态
        if (!adminAPI.isLoggedIn()) {
            window.location.href = 'index.html';
            return;
        }

        // 从URL获取文件ID
        const urlParams = new URLSearchParams(window.location.search);
        this.fileId = urlParams.get('id');
        
        if (!this.fileId) {
            this.showNotification('未指定文件ID', 'error');
            setTimeout(() => {
                window.location.href = 'dashboard.html#files';
            }, 2000);
            return;
        }

        // 绑定事件
        this.bindEvents();
        
        // 加载文件详情
        await this.loadFileDetail();
        // 若带有 edit=1 参数，自动打开编辑模态框
        const editFlag = urlParams.get('edit');
        if (editFlag === '1') {
            this.openEditModal();
        }
        
        // 更新管理员用户名
        this.updateAdminUsername();
    }

    bindEvents() {
        // 编辑按钮
        document.getElementById('editFileBtn').addEventListener('click', () => this.openEditModal());
        
        // 删除按钮
        document.getElementById('deleteFileBtn').addEventListener('click', () => this.deleteFile());
        
        // 退出登录
        document.getElementById('logoutBtn').addEventListener('click', () => this.logout());
        
        // 编辑表单提交
        document.getElementById('editFileForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.updateFile();
        });

        // 点击模态框外部关闭
        window.addEventListener('click', (e) => {
            const modal = document.getElementById('editFileModal');
            if (e.target === modal) {
                this.closeEditModal();
            }
        });
    }

    // 加载文件详情
    async loadFileDetail() {
        const result = await adminAPI.getFileDetail(this.fileId);
        
        if (result.success) {
            this.fileData = result.data;
            this.renderFileDetail();
        } else {
            this.showNotification(result.message || '获取文件详情失败', 'error');
            setTimeout(() => {
                window.location.href = 'dashboard.html#files';
            }, 2000);
        }
    }

    // 渲染文件详情
    renderFileDetail() {
        const file = this.fileData;

        // 顶部标题与徽章
        const titleEl = document.getElementById('pageTitle');
        const nameBadge = document.getElementById('fileNameBadge');
        const publicBadge = document.getElementById('filePublicBadge');
        if (titleEl) titleEl.textContent = this.escapeHTML(file.original_name || '文件详情');
        if (nameBadge) nameBadge.textContent = `${this.escapeHTML(file.mime_type || 'unknown')} · ${this.formatFileSize(file.size || 0)}`;
        if (publicBadge) {
            publicBadge.textContent = file.is_public ? '公开' : '私有';
            publicBadge.classList.remove('status-active', 'status-inactive');
            publicBadge.classList.add(file.is_public ? 'status-active' : 'status-inactive');
        }

        // 基本信息
        document.getElementById('fileBasicInfo').innerHTML = `
            <div class="info-item">
                <label>文件ID</label>
                <span>${file.id}</span>
            </div>
            <div class="info-item">
                <label>原始名称</label>
                <span>${this.escapeHTML(file.original_name)}</span>
            </div>
            <div class="info-item">
                <label>MIME类型</label>
                <span>${this.escapeHTML(file.mime_type)}</span>
            </div>
            <div class="info-item">
                <label>文件大小</label>
                <span>${this.formatFileSize(file.size)}</span>
            </div>
            <div class="info-item">
                <label>分类</label>
                <span>${this.escapeHTML(file.category || '未分类')}</span>
            </div>
            <div class="info-item">
                <label>描述</label>
                <span>${this.escapeHTML(file.description || '无描述')}</span>
            </div>
        `;

        // 存储信息
        document.getElementById('fileStorageInfo').innerHTML = `
            <div class="info-item">
                <label>存储类型</label>
                <span>${this.escapeHTML(file.storage_type)}</span>
            </div>
            <div class="info-item">
                <label>存储名称</label>
                <span>${this.escapeHTML(file.storage_name)}</span>
            </div>
            <div class="info-item">
                <label>存储路径</label>
                <span>${this.escapeHTML(file.storage_path)}</span>
            </div>
            <div class="info-item">
                <label>上传时间</label>
                <span>${this.formatDate(file.created_at)}</span>
            </div>
            <div class="info-item">
                <label>更新时间</label>
                <span>${this.formatDate(file.updated_at)}</span>
            </div>
        `;

        // 访问信息
        document.getElementById('fileAccessInfo').innerHTML = `
            <div class="info-item">
                <label>公开访问</label>
                <span class="status-badge ${file.is_public ? 'status-active' : 'status-inactive'}">
                    ${file.is_public ? '是' : '否'}
                </span>
            </div>
            <div class="info-item">
                <label>上传用户</label>
                <span>${file.user_id ? (file.user_id.substring(0, 8) + '...') : '系统'}</span>
            </div>
            <div class="info-item">
                <label>访问URL</label>
                <div class="inline-actions">
                    <a href="${file.url}" target="_blank" class="link">${file.url}</a>
                    <button class="btn btn-secondary btn-xs" id="copyUrlBtn">复制</button>
                </div>
            </div>
        `;

        // 复制链接按钮
        const copyBtn = document.getElementById('copyUrlBtn');
        if (copyBtn) {
            copyBtn.addEventListener('click', () => this.copyText(file.url));
        }

        // 文件预览
        this.renderFilePreview(file);
    }

    // 渲染文件预览
    renderFilePreview(file) {
        const previewSection = document.getElementById('filePreviewSection');
        const previewContainer = document.getElementById('filePreview');
        
        // 检查是否可以预览
        if (this.canPreview(file.mime_type)) {
            previewSection.style.display = 'block';
            
            if (file.mime_type.startsWith('image/')) {
                previewContainer.innerHTML = `
                    <div class="image-preview card">
                        <img src="${file.url}" alt="${this.escapeHTML(file.original_name)}" class="image-preview-img">
                    </div>
                `;
            } else if (file.mime_type === 'text/plain') {
                // 对于文本文件，可以考虑异步加载内容
                previewContainer.innerHTML = `
                    <div class="text-preview card">
                        <p>文本文件预览功能暂未实现</p>
                        <a href="${file.url}" target="_blank" class="btn btn-primary">查看文件</a>
                    </div>
                `;
            }
        } else {
            previewSection.style.display = 'none';
        }
    }

    // 复制文本到剪贴板
    async copyText(text) {
        try {
            if (navigator.clipboard && window.isSecureContext) {
                await navigator.clipboard.writeText(text);
            } else {
                const ta = document.createElement('textarea');
                ta.value = text; ta.style.position = 'fixed'; ta.style.opacity = '0';
                document.body.appendChild(ta); ta.focus(); ta.select();
                document.execCommand('copy'); document.body.removeChild(ta);
            }
            this.showNotification('已复制到剪贴板', 'success');
        } catch (e) {
            this.showNotification('复制失败，请手动复制', 'error');
        }
    }

    // 检查是否可以预览
    canPreview(mimeType) {
        const previewableTypes = [
            'image/jpeg', 'image/png', 'image/gif', 'image/webp', 'image/svg+xml',
            'text/plain', 'text/html', 'text/css', 'text/javascript',
            'application/json'
        ];
        return previewableTypes.includes(mimeType);
    }

    // 打开编辑模态框
    openEditModal() {
        if (!this.fileData) return;
        
        const file = this.fileData;
        document.getElementById('fileCategory').value = file.category || '';
        document.getElementById('fileDescription').value = file.description || '';
        document.getElementById('fileIsPublic').checked = file.is_public;
        
        document.getElementById('editFileModal').style.display = 'block';
    }

    // 关闭编辑模态框
    closeEditModal() {
        document.getElementById('editFileModal').style.display = 'none';
    }

    // 更新文件信息
    async updateFile() {
        const formData = new FormData(document.getElementById('editFileForm'));
        const updateData = {
            category: formData.get('category').trim(),
            description: formData.get('description').trim(),
            is_public: document.getElementById('fileIsPublic').checked
        };

        const result = await adminAPI.updateFile(this.fileId, updateData);
        
        if (result.success) {
            this.showNotification('文件信息更新成功', 'success');
            this.closeEditModal();
            await this.loadFileDetail(); // 重新加载详情
        } else {
            this.showNotification(result.message || '更新失败', 'error');
        }
    }

    // 删除文件
    async deleteFile() {
        if (!confirm('确定要删除该文件吗？此操作不可恢复！')) {
            return;
        }

        const result = await adminAPI.deleteFile(this.fileId);
        
        if (result.success) {
            this.showNotification('文件删除成功', 'success');
            setTimeout(() => {
                window.location.href = 'dashboard.html#files';
            }, 1500);
        } else {
            this.showNotification(result.message || '删除失败', 'error');
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

    // 工具函数：格式化文件大小
    formatFileSize(bytes) {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
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

    // 显示通知
    showNotification(message, type = 'info') {
        const notification = document.createElement('div');
        notification.className = `notification notification-${type}`;
        notification.textContent = message;
        
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
        
        setTimeout(() => {
            notification.remove();
        }, 3000);
    }
}

// 初始化文件详情页面
let fileDetail;
document.addEventListener('DOMContentLoaded', function() {
    fileDetail = new FileDetailManager();
});
