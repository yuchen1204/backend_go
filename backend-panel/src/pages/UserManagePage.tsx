import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../services/api';
import LoadingSpinner from '../components/LoadingSpinner';
import '../styles/UserManage.css';

interface User {
  id: string;
  username: string;
  email: string;
  nickname: string;
  bio: string;
  avatar: string;
  status: 'active' | 'inactive' | 'banned';
  created_at: string;
  updated_at: string;
}

interface UserStats {
  total_users: number;
  active_users: number;
  inactive_users: number;
  banned_users: number;
}

const UserManagePage: React.FC = () => {
  const [users, setUsers] = useState<User[]>([]);
  const [stats, setStats] = useState<UserStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedUser, setSelectedUser] = useState<User | null>(null);
  const [showUserDetail, setShowUserDetail] = useState(false);
  const navigate = useNavigate();

  const limit = 10;

  useEffect(() => {
    fetchUsers();
    fetchStats();
  }, [currentPage, searchTerm]);

  const fetchUsers = async () => {
    try {
      setLoading(true);
      const token = localStorage.getItem('admin_token');
      const response = await api.get('/admin/users', {
        headers: { Authorization: `Bearer ${token}` },
        params: {
          page: currentPage,
          limit: limit,
          search: searchTerm || undefined,
        },
      });

      if (response.data && response.data.data) {
        setUsers(response.data.data.users || []);
        const total = response.data.data.total || 0;
        setTotalPages(Math.ceil(total / limit));
      }
    } catch (err: any) {
      if (err.response?.status === 401) {
        localStorage.removeItem('admin_token');
        navigate('/login');
      } else {
        setError('获取用户列表失败');
      }
    } finally {
      setLoading(false);
    }
  };

  const fetchStats = async () => {
    try {
      const token = localStorage.getItem('admin_token');
      const response = await api.get('/admin/stats/users', {
        headers: { Authorization: `Bearer ${token}` },
      });

      if (response.data && response.data.data) {
        setStats(response.data.data);
      }
    } catch (err) {
      console.error('获取用户统计失败:', err);
    }
  };

  const handleStatusChange = async (userId: string, newStatus: string) => {
    try {
      const token = localStorage.getItem('admin_token');
      await api.put(`/admin/users/${userId}/status`, 
        { status: newStatus },
        { headers: { Authorization: `Bearer ${token}` } }
      );
      
      // 更新本地状态
      setUsers(users.map(user => 
        user.id === userId ? { ...user, status: newStatus as any } : user
      ));
      
      // 刷新统计数据
      fetchStats();
    } catch (err: any) {
      setError('更新用户状态失败');
    }
  };

  const handleDeleteUser = async (userId: string) => {
    if (!window.confirm('确定要删除此用户吗？此操作不可撤销。')) {
      return;
    }

    try {
      const token = localStorage.getItem('admin_token');
      await api.delete(`/admin/users/${userId}`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      
      // 刷新用户列表
      fetchUsers();
      fetchStats();
    } catch (err: any) {
      setError('删除用户失败');
    }
  };

  const handleViewDetail = async (userId: string) => {
    try {
      const token = localStorage.getItem('admin_token');
      const response = await api.get(`/admin/users/${userId}`, {
        headers: { Authorization: `Bearer ${token}` },
      });

      if (response.data && response.data.data) {
        setSelectedUser(response.data.data);
        setShowUserDetail(true);
      }
    } catch (err: any) {
      setError('获取用户详情失败');
    }
  };

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    setCurrentPage(1);
    fetchUsers();
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'status-active';
      case 'inactive': return 'status-inactive';
      case 'banned': return 'status-banned';
      default: return '';
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case 'active': return '活跃';
      case 'inactive': return '非活跃';
      case 'banned': return '已封禁';
      default: return status;
    }
  };

  return (
    <div className="user-manage-container">
      <div className="user-manage-header">
        <h1>用户管理</h1>
        <button 
          className="back-button"
          onClick={() => navigate('/dashboard')}
        >
          返回仪表盘
        </button>
      </div>

      {/* 统计卡片 */}
      {stats && (
        <div className="stats-grid">
          <div className="stat-card">
            <div className="stat-number">{stats.total_users}</div>
            <div className="stat-label">总用户数</div>
          </div>
          <div className="stat-card">
            <div className="stat-number">{stats.active_users}</div>
            <div className="stat-label">活跃用户</div>
          </div>
          <div className="stat-card">
            <div className="stat-number">{stats.inactive_users}</div>
            <div className="stat-label">非活跃用户</div>
          </div>
          <div className="stat-card">
            <div className="stat-number">{stats.banned_users}</div>
            <div className="stat-label">已封禁用户</div>
          </div>
        </div>
      )}

      {/* 搜索栏 */}
      <div className="search-section">
        <form onSubmit={handleSearch} className="search-form">
          <input
            type="text"
            placeholder="搜索用户名、邮箱或昵称..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="search-input"
          />
          <button type="submit" className="search-button">搜索</button>
        </form>
      </div>

      {error && <div className="error-message">{error}</div>}

      {/* 用户列表 */}
      <div className="users-section">
        {loading ? (
          <div className="loading-container">
            <LoadingSpinner size="large" />
          </div>
        ) : (
          <>
            <div className="users-table">
              <div className="table-header">
                <div className="header-cell">用户名</div>
                <div className="header-cell">邮箱</div>
                <div className="header-cell">昵称</div>
                <div className="header-cell">状态</div>
                <div className="header-cell">注册时间</div>
                <div className="header-cell">操作</div>
              </div>
              
              {users.map((user) => (
                <div key={user.id} className="table-row">
                  <div className="table-cell">
                    <div className="user-info">
                      <img 
                        src={user.avatar || '/default-avatar.png'} 
                        alt={user.username}
                        className="user-avatar"
                        onError={(e) => {
                          (e.target as HTMLImageElement).src = '/default-avatar.png';
                        }}
                      />
                      <span>{user.username}</span>
                    </div>
                  </div>
                  <div className="table-cell">{user.email}</div>
                  <div className="table-cell">{user.nickname || '-'}</div>
                  <div className="table-cell">
                    <span className={`status-badge ${getStatusColor(user.status)}`}>
                      {getStatusText(user.status)}
                    </span>
                  </div>
                  <div className="table-cell">
                    {new Date(user.created_at).toLocaleDateString('zh-CN')}
                  </div>
                  <div className="table-cell">
                    <div className="action-buttons">
                      <button 
                        className="action-btn view-btn"
                        onClick={() => handleViewDetail(user.id)}
                      >
                        查看
                      </button>
                      <select
                        value={user.status}
                        onChange={(e) => handleStatusChange(user.id, e.target.value)}
                        className="status-select"
                      >
                        <option value="active">活跃</option>
                        <option value="inactive">非活跃</option>
                        <option value="banned">已封禁</option>
                      </select>
                      <button 
                        className="action-btn delete-btn"
                        onClick={() => handleDeleteUser(user.id)}
                      >
                        删除
                      </button>
                    </div>
                  </div>
                </div>
              ))}
            </div>

            {/* 分页 */}
            {totalPages > 1 && (
              <div className="pagination">
                <button 
                  className="page-btn"
                  disabled={currentPage === 1}
                  onClick={() => setCurrentPage(currentPage - 1)}
                >
                  上一页
                </button>
                <span className="page-info">
                  第 {currentPage} 页，共 {totalPages} 页
                </span>
                <button 
                  className="page-btn"
                  disabled={currentPage === totalPages}
                  onClick={() => setCurrentPage(currentPage + 1)}
                >
                  下一页
                </button>
              </div>
            )}
          </>
        )}
      </div>

      {/* 用户详情模态框 */}
      {showUserDetail && selectedUser && (
        <div className="modal-overlay" onClick={() => setShowUserDetail(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h2>用户详情</h2>
              <button 
                className="close-btn"
                onClick={() => setShowUserDetail(false)}
              >
                ×
              </button>
            </div>
            <div className="modal-body">
              <div className="user-detail">
                <div className="detail-row">
                  <label>用户ID:</label>
                  <span>{selectedUser.id}</span>
                </div>
                <div className="detail-row">
                  <label>用户名:</label>
                  <span>{selectedUser.username}</span>
                </div>
                <div className="detail-row">
                  <label>邮箱:</label>
                  <span>{selectedUser.email}</span>
                </div>
                <div className="detail-row">
                  <label>昵称:</label>
                  <span>{selectedUser.nickname || '-'}</span>
                </div>
                <div className="detail-row">
                  <label>个人简介:</label>
                  <span>{selectedUser.bio || '-'}</span>
                </div>
                <div className="detail-row">
                  <label>状态:</label>
                  <span className={`status-badge ${getStatusColor(selectedUser.status)}`}>
                    {getStatusText(selectedUser.status)}
                  </span>
                </div>
                <div className="detail-row">
                  <label>注册时间:</label>
                  <span>{new Date(selectedUser.created_at).toLocaleString('zh-CN')}</span>
                </div>
                <div className="detail-row">
                  <label>更新时间:</label>
                  <span>{new Date(selectedUser.updated_at).toLocaleString('zh-CN')}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default UserManagePage;
