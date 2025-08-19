import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../services/api';
import '../styles/Dashboard.css';

interface UserStats {
  total_users: number;
  active_users: number;
  inactive_users: number;
  banned_users: number;
}

const DashboardPage: React.FC = () => {
  const navigate = useNavigate();
  const [userStats, setUserStats] = useState<UserStats | null>(null);

  useEffect(() => {
    fetchUserStats();
  }, []);

  const fetchUserStats = async () => {
    try {
      const token = localStorage.getItem('admin_token');
      const response = await api.get('/admin/stats/users', {
        headers: { Authorization: `Bearer ${token}` },
      });

      if (response.data && response.data.data) {
        setUserStats(response.data.data);
      }
    } catch (err) {
      console.error('获取用户统计失败:', err);
    }
  };

  const handleLogout = () => {
    localStorage.removeItem('admin_token');
    navigate('/login');
  };

  return (
    <div className="dashboard-container">
      <header className="dashboard-header">
        <h1 className="dashboard-title">管理员仪表盘</h1>
        <div className="dashboard-user-info">
          <div className="user-avatar">管</div>
          <button onClick={handleLogout} className="logout-button">
            退出登录
          </button>
        </div>
      </header>

      <main className="dashboard-content">
        <section className="welcome-section">
          <h2 className="welcome-title">欢迎回来！</h2>
          <p className="welcome-subtitle">系统运行正常，一切数据都在您的掌控之中</p>
          <div className="welcome-stats">
            <div className="welcome-stat">
              <div className="welcome-stat-number">{userStats?.total_users || 0}</div>
              <div className="welcome-stat-label">总用户数</div>
            </div>
            <div className="welcome-stat">
              <div className="welcome-stat-number">5,678</div>
              <div className="welcome-stat-label">文件总数</div>
            </div>
            <div className="welcome-stat">
              <div className="welcome-stat-number">99.9%</div>
              <div className="welcome-stat-label">系统可用性</div>
            </div>
          </div>
        </section>

        <div className="dashboard-grid">
          <div className="dashboard-card">
            <div className="card-header">
              <div className="card-icon users">👥</div>
              <h3 className="card-title">用户管理</h3>
            </div>
            <p className="card-description">
              管理系统用户，查看用户活动和统计信息
            </p>
            <div className="card-stats">
              <div>
                <div className="stat-number">{userStats?.active_users || 0}</div>
                <div className="stat-label">活跃用户</div>
              </div>
              <button className="card-action" onClick={() => navigate('/users')}>查看详情</button>
            </div>
          </div>

          <div className="dashboard-card">
            <div className="card-header">
              <div className="card-icon files">📁</div>
              <h3 className="card-title">文件管理</h3>
            </div>
            <p className="card-description">
              监控文件上传、存储使用情况和访问统计
            </p>
            <div className="card-stats">
              <div>
                <div className="stat-number">5,678</div>
                <div className="stat-label">总文件数</div>
              </div>
              <button className="card-action">管理文件</button>
            </div>
          </div>

          <div className="dashboard-card">
            <div className="card-header">
              <div className="card-icon analytics">📊</div>
              <h3 className="card-title">数据分析</h3>
            </div>
            <p className="card-description">
              查看系统使用情况、性能指标和趋势分析
            </p>
            <div className="card-stats">
              <div>
                <div className="stat-number">99.9%</div>
                <div className="stat-label">系统可用性</div>
              </div>
              <button className="card-action">查看报告</button>
            </div>
          </div>

          <div className="dashboard-card">
            <div className="card-header">
              <div className="card-icon settings">⚙️</div>
              <h3 className="card-title">系统设置</h3>
            </div>
            <p className="card-description">
              配置系统参数、安全设置和维护选项
            </p>
            <div className="card-stats">
              <div>
                <div className="stat-number">12</div>
                <div className="stat-label">配置项</div>
              </div>
              <button className="card-action">系统设置</button>
            </div>
          </div>
        </div>

        <div className="quick-actions">
          <div className="quick-action-btn">
            <div className="quick-action-icon">🔍</div>
            <div className="quick-action-title">搜索用户</div>
            <div className="quick-action-desc">快速查找特定用户</div>
          </div>
          <div className="quick-action-btn">
            <div className="quick-action-icon">📈</div>
            <div className="quick-action-title">生成报告</div>
            <div className="quick-action-desc">创建系统使用报告</div>
          </div>
          <div className="quick-action-btn">
            <div className="quick-action-icon">🛠️</div>
            <div className="quick-action-title">系统维护</div>
            <div className="quick-action-desc">执行系统维护任务</div>
          </div>
          <div className="quick-action-btn">
            <div className="quick-action-icon">📋</div>
            <div className="quick-action-title">查看日志</div>
            <div className="quick-action-desc">检查系统运行日志</div>
          </div>
        </div>
      </main>
    </div>
  );
};

export default DashboardPage;
