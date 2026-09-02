/**
 * k6 脚本 BM-02：用户登录鉴权（POST /api/v1/auth/login）
 *
 * setup 阶段创建一个专用账号，压测阶段只测登录，避免把注册耗时混入登录指标。
 * CPU 密集路径：bcrypt/pbkdf2 校验 + JWT 签发。
 *
 * 环境变量：
 *   K6_BASE_URL   压测目标（默认 http://127.0.0.1:8888）
 *   K6_VUS        并发（默认 30，登录为写操作，低于公开搜索）
 *   K6_DURATION   单轮持续时间（默认 60s）
 *   K6_PASSWORD   注册/登录统一密码（默认 Benchmark123!）
 */
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const errorRate = new Rate('bm02_errors');
const ttfbTrend = new Trend('bm02_ttfb', true);

export const options = {
  vus: Number(__ENV.K6_VUS || 30),
  duration: __ENV.K6_DURATION || '60s',
  thresholds: {
    http_req_failed: ['rate<0.02'],
    'bm02_errors': ['rate<0.02'],
    http_req_duration: ['p(95)<1500', 'avg<700'],
  },
  noConnectionReuse: false,
  userAgent: 'DanmakuStreamBenchmark/1.0 (k6; BM-02)',
};

const BASE = (__ENV.K6_BASE_URL || 'http://127.0.0.1:8888').replace(/\/$/, '');
const PASSWORD = __ENV.K6_PASSWORD || 'Benchmark123!';

export function setup() {
  const nickname = `bm02-login-${Date.now()}`;
  const headers = { 'Content-Type': 'application/json' };
  const reg = http.post(`${BASE}/api/v1/auth/register`,
    JSON.stringify({ nickname, password: PASSWORD }),
    { headers, tags: { phase: 'setup' } });
  if (reg.status !== 200) {
    throw new Error(`setup register failed: ${reg.status} ${reg.body}`);
  }
  return { nickname };
}

export default function (data) {
  const headers = { 'Content-Type': 'application/json' };
  const login = http.post(`${BASE}/api/v1/auth/login`,
    JSON.stringify({ nickname: data.nickname, password: PASSWORD }),
    { headers, tags: { api: 'bm02_login' } });

  const ok = check(login, {
    'login status 200': (r) => r.status === 200,
    'login has token': (r) => {
      try {
        const d = r.json().data;
        return d && typeof d.token === 'string' && d.token.length > 20;
      } catch (e) {
        return false;
      }
    },
  });
  errorRate.add(!ok);
  if (login.timings) ttfbTrend.add(login.timings.waiting);
  sleep(0.6 + Math.random() * 0.4);
}
