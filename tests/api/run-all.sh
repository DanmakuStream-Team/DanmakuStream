#!/usr/bin/env bash
# 统一 API 测试入口：按顺序执行全部 API 脚本（前置：后端 8080 + MySQL 已启动）。
# 任一脚本失败即以非零退出；可用 API_BASE / MYSQL_CMD / VIDEO_DIR 适配环境。
set -uo pipefail
cd "$(dirname "$0")/../.."

FAIL=0
for script in tests/api/uc13-admin-test.sh tests/api/member-c-content-test.sh tests/api/uc06-library-test.sh; do
  echo "===== $script ====="
  if ! bash "$script"; then
    echo "!!!!! FAILED: $script"
    FAIL=1
  fi
done

if [ "$FAIL" -ne 0 ]; then
  echo "===== run-all: 存在失败 ====="
  exit 1
fi
echo "===== run-all: 全部通过 ====="
