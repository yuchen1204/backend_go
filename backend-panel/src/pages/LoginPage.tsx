import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../services/api';
import LoadingSpinner from '../components/LoadingSpinner';
import '../styles/Login.css';

const LoginPage: React.FC = () => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const navigate = useNavigate();

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);
    
    try {
      const response = await api.post('/admin/login', {
        username,
        password,
      });

      if (response.data && response.data.data.token) {
        localStorage.setItem('admin_token', response.data.data.token);
        navigate('/dashboard');
      } else {
        setError('登录失败，无效的响应');
      }
    } catch (err: any) {
      if (err.response && err.response.data && err.response.data.message) {
        setError(err.response.data.message);
      } else {
        setError('登录时发生未知错误');
      }
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="login-container">
      <div className="login-card">
        <div className="login-header">
          <h1 className="login-title">管理员登录</h1>
          <p className="login-subtitle">请输入您的管理员凭据</p>
        </div>
        <form className="login-form" onSubmit={handleLogin}>
          <div className="form-group">
            <label htmlFor="username" className="form-label">用户名</label>
            <input
              type="text"
              id="username"
              className="form-input"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
              placeholder="请输入用户名"
            />
          </div>
          <div className="form-group">
            <label htmlFor="password" className="form-label">密码</label>
            <input
              type="password"
              id="password"
              className="form-input"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              placeholder="请输入密码"
            />
          </div>
          {error && <div className="error-message">{error}</div>}
          <button type="submit" className="login-button" disabled={isLoading}>
            {isLoading ? (
              <>
                <LoadingSpinner size="small" color="white" />
                <span style={{ marginLeft: '8px' }}>登录中...</span>
              </>
            ) : (
              '登录'
            )}
          </button>
        </form>
      </div>
    </div>
  );
};

export default LoginPage;
