# UC01 / UC06 自动化测试报告（2026-08-29）

## 结论

| 测试层 | 范围 | 结果 |
|---|---|---|
| 后端单元测试 | `go test -count=1 ./...` | 通过 |
| UC06 API/集成测试 | 真实 MySQL 8.0.46 + 后端；鉴权、异常、历史、进度、稍后再看 | 30/30 通过 |
| UC01/UC06 E2E | Chromium + Playwright；UC01 5 条、UC06 4 条 | 9/9 通过 |
| 前端构建 | `vue-tsc && vite build` | 通过 |

## 执行环境

- Linux x86_64
- Go（版本由 `backend/go.mod` / CI `setup-go` 固定）
- Node.js 20（CI）
- MySQL 8.0.46
- Playwright Chromium

测试数据使用唯一运行标识创建，API 脚本退出时清理用户、视频、观看历史和稍后再看记录。

## 可重复执行命令

```bash
cd backend
go test -count=1 ./...

cd ../frontend
npm run build
MYSQL_CMD='mysql -h127.0.0.1 -P3306 -uroot -ppassword danmakustream' \
  npm run test:e2e:uc01-uc06

cd ..
MYSQL_CMD='mysql -h127.0.0.1 -P3306 -uroot -ppassword danmakustream' \
  tests/api/uc06-library-test.sh
```

CI 会保存原始控制台输出、Playwright HTML 报告以及失败时的 trace、截图和视频。该本地报告记录本次绿灯结果；合并前以后续 GitHub Actions artifact 作为远程复验证据。
