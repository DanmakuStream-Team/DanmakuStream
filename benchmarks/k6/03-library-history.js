/**
 * k6 脚本 BM-03：个人资料库-观看历史（GET /api/v1/users/me/history）
 *
 * UC06 核心接口：需要鉴权 + 查 watch_histories 分页 + JOIN videos。
 * Setup 阶段先创建 1 个账号，并为最多 100 个不同视频写入历史。
 *
 * 环境变量：
 *   K6_BASE_URL        压测目标（默认 http://127.0.0.1:8888）
 *   K6_VUS             并发（默认 40）
 *   K6_DURATION        单轮持续时间（默认 60s）
 *   K6_HISTORY_SEEDER  1=在 setup 中写入不同视频的历史；0=跳过（已有数据时）
 *   K6_PASSWORD        账号密码（默认 Benchmark123!）
 */
import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const errorRate = new Rate('bm03_errors');
const ttfbTrend = new Trend('bm03_ttfb', true);

export const options = {
  vus: Number(__ENV.K6_VUS || 40),
  duration: __ENV.K6_DURATION || '60s',
  thresholds: {
    http_req_failed: ['rate<0.02'],
    'bm03_errors': ['rate<0.02'],
    http_req_duration: ['p(95)<1200', 'avg<500'],
  },
  noConnectionReuse: false,
  userAgent: 'DanmakuStreamBenchmark/1.0 (k6; BM-03)',
};

const BASE = (__ENV.K6_BASE_URL || 'http://127.0.0.1:8888').replace(/\/$/, '');
const PASSWORD = __ENV.K6_PASSWORD || 'Benchmark123!';

export function setup() {
  const headers = { 'Content-Type': 'application/json' };
  const nickname = `bm03-benchmark-${Date.now()}`;
  const reg = http.post(`${BASE}/api/v1/auth/register`,
    JSON.stringify({ nickname, password: PASSWORD }), { headers });
  if (reg.status !== 200 && reg.status !== 409) {
    throw new Error(`setup register failed: ${reg.status} ${reg.body}`);
  }
  const login = http.post(`${BASE}/api/v1/auth/login`,
    JSON.stringify({ nickname, password: PASSWORD }), { headers });
  const token = login.json().data.token;
  if (!token) throw new Error('setup register/login did not return token');

  // 查最多 100 个公开视频作为历史写入目标。
  const videosResp = http.get(`${BASE}/api/v1/videos?page=1&pageSize=100`);
  let videoIds = [];
  try {
    const videos = videosResp.json().data.list || [];
    if (videos.length > 0) videoIds = videos.map((v) => v.id).filter(Boolean);
  } catch (e) {}

  if ((__ENV.K6_HISTORY_SEEDER || '1') !== '0') {
    const authHeaders = { ...headers, Authorization: `Bearer ${token}` };
    for (let i = 0; i < videoIds.length; i += 1) {
      const vid = videoIds[i];
      const progress = Math.floor(Math.random() * 1200) + 1;
      http.post(`${BASE}/api/v1/users/me/history/${vid}`,
        JSON.stringify({ progress }), { headers: authHeaders });
      if (i % 20 === 0) sleep(0.02);
    }
  }

  return { token };
}

export default function (data) {
  group('UC06 - history paginated query', () => {
    const page = 1 + Math.floor(Math.random() * 3);
    const url = `${BASE}/api/v1/users/me/history?page=${page}&pageSize=20`;
    const res = http.get(url, {
      headers: { Authorization: `Bearer ${data.token}` },
      tags: { api: 'bm03_history' },
    });
    const ok = check(res, {
      'history status 200': (r) => r.status === 200,
      'history has total field': (r) => {
        try {
          const d = r.json().data;
          return d && typeof d.total === 'number' && Array.isArray(d.list);
        } catch (e) {
          return false;
        }
      },
    });
    errorRate.add(!ok);
    if (res.timings) ttfbTrend.add(res.timings.waiting);
  });
  sleep(0.4 + Math.random() * 0.4);
}
