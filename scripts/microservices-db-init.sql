-- 微服务数据库初始化（D06）：三库三账号，最小权限。
-- 由 docker-compose.microservices.yml 的 mysql 以 initdb 脚本方式执行；
-- K8s 环境用 deploy/k8s/microservices/mysql-init-configmap.yaml 挂载同内容。
-- 注意：每个应用账号只授权自己的 Schema，禁止跨 Schema 访问（对应《微服务划分》§4 规则 1）。

CREATE DATABASE IF NOT EXISTS user_db       CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS content_db    CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS engagement_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 账号口令仅限本地/演示环境；正式环境经 Secret 注入后执行 ALTER USER 轮换。
CREATE USER IF NOT EXISTS 'user_app'       @'%' IDENTIFIED BY 'user_app_pass';
CREATE USER IF NOT EXISTS 'content_app'    @'%' IDENTIFIED BY 'content_app_pass';
CREATE USER IF NOT EXISTS 'engagement_app' @'%' IDENTIFIED BY 'engagement_app_pass';

GRANT ALL PRIVILEGES ON user_db.*       TO 'user_app'@'%';
GRANT ALL PRIVILEGES ON content_db.*    TO 'content_app'@'%';
GRANT ALL PRIVILEGES ON engagement_db.* TO 'engagement_app'@'%';

-- 跨库访问：默认不授予（MySQL 无授权即拒绝）。上面的 GRANT 逐账号只指向自己的 Schema，
-- 即满足"每个账号只能访问自己的 Schema"；如未来误授权，用 REVOKE 收回。
