// 登录页面逻辑
document.addEventListener('DOMContentLoaded', function() {
    const loginForm = document.getElementById('loginForm');
    const errorMessage = document.getElementById('errorMessage');

    // 检查是否已登录
    if (adminAPI.isLoggedIn()) {
        window.location.href = 'dashboard.html';
        return;
    }

    // 处理登录表单提交
    loginForm.addEventListener('submit', async function(e) {
        e.preventDefault();
        
        const username = document.getElementById('username').value.trim();
        const password = document.getElementById('password').value;

        if (!username || !password) {
            showError('请输入用户名和密码');
            return;
        }

        // 显示加载状态
        const submitBtn = loginForm.querySelector('button[type="submit"]');
        const originalText = submitBtn.textContent;
        submitBtn.textContent = '登录中...';
        submitBtn.disabled = true;

        try {
            const result = await adminAPI.adminLogin(username, password);
            
            if (result.success) {
                // 登录成功，跳转到仪表盘
                window.location.href = 'dashboard.html';
            } else {
                showError(result.message);
                submitBtn.textContent = originalText;
                submitBtn.disabled = false;
            }
        } catch (error) {
            showError('登录失败，请稍后重试');
            submitBtn.textContent = originalText;
            submitBtn.disabled = false;
        }
    });

    // 显示错误信息
    function showError(message) {
        errorMessage.textContent = message;
        errorMessage.style.display = 'block';
        
        // 3秒后自动隐藏错误信息
        setTimeout(() => {
            errorMessage.style.display = 'none';
        }, 3000);
    }

    // 回车键提交表单
    document.addEventListener('keypress', function(e) {
        if (e.key === 'Enter') {
            loginForm.dispatchEvent(new Event('submit'));
        }
    });
});
