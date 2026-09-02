/**
 * k6 脚本 BM-01：公开视频搜索（GET /api/v1/videos?keyword=...）
 *
 * 环境变量：
 *   K6_BASE_URL   压测目标（默认单体 http://127.0.0.1:8888）
 *   K6_VUS        并发用户数（默认 50）
 *   K6_DURATION   单轮持续时间（默认 60s，含 10s 预热）
 *   K6_KEYWORDS   搜索关键词列表，逗号分隔（默认提供覆盖常见分布）
 *
 * 运行：
 *   k6 run -e K6_BASE_URL=http://127.0.0.1:18888 benchmarks/k6/01-public-video-search.js
 *   k6 run --summary-export=artifacts/benchmarks/micro-01-r1.json benchmarks/k6/01-public-video-search.js
 */
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend } from 'k6/metrics';

const errorRate = new Rate('bm01_errors');
const ttfbTrend = new Trend('bm01_ttfb', true);

const keywords = __ENV.K6_KEYWORDS
  ? __ENV.K6_KEYWORDS.split(',')
  : ['E2E-MC-公开视频', 'E2E-UC05-互动测试视频', 'E2E-MEMBER-B-分享视频', '公开', '视频', 'E2E', ''];

export const options = {
  vus: Number(__ENV.K6_VUS || 50),
  duration: __ENV.K6_DURATION || '60s',
  thresholds: {
    http_req_failed: ['rate<0.02'],
    'bm01_errors': ['rate<0.02'],
    http_req_duration: ['p(95)<800', 'avg<400'],
  },
  noConnectionReuse: false,
  userAgent: 'DanmakuStreamBenchmark/1.0 (k6; BM-01)',
};

const BASE = (__ENV.K6_BASE_URL || 'http://127.0.0.1:8888').replace(/\/$/, '');

export default function () {
  const kw = keywords[Math.floor(Math.random() * keywords.length)];
  const url = `${BASE}/api/v1/videos?page=1&pageSize=20&keyword=${encodeURIComponent(kw)}`;
  const res = http.get(url, { tags: { api: 'bm01_search' } });

  check(res, {
    'status is 200': (r) => r.status === 200,
    'has json data.list': (r) => {
      try {
        const body = r.json();
        return body && body.data && Array.isArray(body.data.list);
      } catch (e) {
        return false;
      }
    },
  }) || errorRate.add(1);

  if (res.timings) ttfbTrend.add(res.timings.waiting);
  sleep(0.3 + Math.random() * 0.5);
}
