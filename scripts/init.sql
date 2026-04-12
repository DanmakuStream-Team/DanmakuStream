-- 初始化脚本，建库后自动执行
CREATE DATABASE IF NOT EXISTS danmakustream CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE danmakustream;

-- 初始管理员账号（密码: admin123，bcrypt hash 已预生成）
-- 实际部署时通过注册接口或后台工具创建，此处仅做示例
