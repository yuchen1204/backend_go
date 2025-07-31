-- 创建数据库
CREATE DATABASE IF NOT EXISTS backend;

-- 创建用户（如果不存在）
DO
$do$
BEGIN
   IF NOT EXISTS (
      SELECT FROM pg_catalog.pg_roles
      WHERE  rolname = 'backend_user') THEN

      CREATE ROLE backend_user LOGIN PASSWORD 'backend_password';
   END IF;
END
$do$;

-- 授予权限
GRANT ALL PRIVILEGES ON DATABASE backend TO backend_user;

-- 连接到backend数据库
\c backend;

-- 授予schema权限
GRANT ALL ON SCHEMA public TO backend_user; 