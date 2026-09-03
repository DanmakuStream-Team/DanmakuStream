-- 微服务数据库初始化（D06）：三库三账号，最小权限。
-- 由 docker-compose.microservices.yml 的 mysql 以 initdb 脚本方式执行；
-- K8s 环境用 deploy/k8s/microservices/mysql-init-configmap.yaml 挂载同内容。
-- 注意：每个应用账号只授权自己的 Schema，禁止跨 Schema 访问（对应《微服务划分》§4 规则 1）。

CREATE DATABASE IF NOT EXISTS user_db       CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS content_db    CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS engagement_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 账号口令仅限本地/演示环境；正式环境由 Secret 注入。
CREATE USER IF NOT EXISTS 'user_app'       @'%' IDENTIFIED BY 'user_app_pass';
CREATE USER IF NOT EXISTS 'content_app'    @'%' IDENTIFIED BY 'content_app_pass';
CREATE USER IF NOT EXISTS 'engagement_app' @'%' IDENTIFIED BY 'engagement_app_pass';

-- 先撤销历史授权，保证脚本重复执行时也收敛到同一权限集。
REVOKE ALL PRIVILEGES, GRANT OPTION FROM 'user_app'@'%';
REVOKE ALL PRIVILEGES, GRANT OPTION FROM 'content_app'@'%';
REVOKE ALL PRIVILEGES, GRANT OPTION FROM 'engagement_app'@'%';

-- 运行期 DML + 当前 GORM AutoMigrate 所需 DDL；不授予全局、跨库或授权管理权限。
GRANT SELECT, INSERT, UPDATE, DELETE, CREATE, ALTER, INDEX, DROP, REFERENCES
  ON user_db.* TO 'user_app'@'%';
GRANT SELECT, INSERT, UPDATE, DELETE, CREATE, ALTER, INDEX, DROP, REFERENCES
  ON content_db.* TO 'content_app'@'%';
GRANT SELECT, INSERT, UPDATE, DELETE, CREATE, ALTER, INDEX, DROP, REFERENCES
  ON engagement_db.* TO 'engagement_app'@'%';
