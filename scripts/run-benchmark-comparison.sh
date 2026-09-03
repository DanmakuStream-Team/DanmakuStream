#!/usr/bin/env bash
# ==============================================================
# Day 9 性能压测对比脚本（单体 vs 微服务，各 3 轮）
# 产物：artifacts/benchmarks/
#           run.log
#           {monolith|microservices}-bm{01,02,03}-r{1,2,3}.json
#           {monolith|microservices}-bm{01,02,03}-r{1,2,3}.txt
#           {monolith|microservices}-stats.csv
#           comparison.csv / comparison.md
#
# 前置：Docker Engine + Docker Compose v2；无需在本机装 k6（脚本内用 docker.io/grafana/k6）
# 示例：
#   bash scripts/run-benchmark-comparison.sh
#   BENCHMARK_ROUNDS=5 BENCHMARK_VUS_BM01=100 BENCHMARK_DURATION=90s \
#       bash scripts/run-benchmark-comparison.sh
# ==============================================================
set -Eeuo pipefail
cd "$(dirname "$0")/.."
REPO_ROOT="$(pwd)"

# ── 可调参数 ──────────────────────────────────────────────────
ROUNDS="${BENCHMARK_ROUNDS:-3}"
DURATION="${BENCHMARK_DURATION:-60s}"
VUS_BM01="${BENCHMARK_VUS_BM01:-50}"    # 公开搜索
VUS_BM02="${BENCHMARK_VUS_BM02:-30}"    # 登录鉴权
VUS_BM03="${BENCHMARK_VUS_BM03:-40}"    # 资料库历史
K6_IMAGE="${BENCHMARK_K6_IMAGE:-docker.io/grafana/k6:0.52.0}"
MONO_GATEWAY_PORT="${BENCHMARK_MONO_PORT:-28888}"
MONO_FRONTEND_PORT="${BENCHMARK_MONO_FRONTEND_PORT:-28080}"
MONO_MYSQL_PORT="${BENCHMARK_MONO_MYSQL_PORT:-23306}"
MONO_RTMP_PORT="${BENCHMARK_MONO_RTMP_PORT:-29350}"
MONO_SRS_API_PORT="${BENCHMARK_MONO_SRS_API_PORT:-21985}"
MONO_HLS_PORT="${BENCHMARK_MONO_HLS_PORT:-28081}"
MICRO_GATEWAY_PORT="${BENCHMARK_MICRO_PORT:-38888}"
MICRO_FRONTEND_PORT="${BENCHMARK_MICRO_FRONTEND_PORT:-38080}"
MICRO_RTMP_PORT="${BENCHMARK_MICRO_RTMP_PORT:-39350}"
MICRO_HLS_PORT="${BENCHMARK_MICRO_HLS_PORT:-38081}"
ARTIFACT_DIR="$REPO_ROOT/artifacts/benchmarks"
LOG_FILE="$ARTIFACT_DIR/run.log"
PULL_K6="${BENCHMARK_PULL_K6:-1}"

mkdir -p "$ARTIFACT_DIR"
touch "$LOG_FILE"

log()  { printf '[%s] %s\n' "$(date '+%H:%M:%S')" "$*" | tee -a "$LOG_FILE"; }
step() { printf '\n===[%s]=== %s ===\n' "$(date '+%H:%M:%S')" "$*" | tee -a "$LOG_FILE"; }

# ── 工具函数 ──────────────────────────────────────────────────
pull_deps() {
  if [ "$PULL_K6" = "1" ]; then
    docker pull "$K6_IMAGE" >/dev/null 2>&1 || log "k6 镜像拉取失败/已存在，跳过"
  fi
}

docker_compose() {
  # $1 = compose 文件名
  # $2 = project name
  # $@ = subcommand + args
  local f="$1"; shift
  local p="$1"; shift
  docker compose --project-name "$p" -f "$REPO_ROOT/$f" "$@"
}

benchmark_video_values() {
  local include_transcode=$1 separator=''
  for i in $(seq 1 100); do
    if [ "$include_transcode" = 1 ]; then
      printf "%s(NOW(),NOW(),'BM-公开视频-%d','压测数据','/data/videos/seed.mp4','approved','ready',@owner_id)" "$separator" "$i"
    else
      printf "%s(NOW(),NOW(),'BM-公开视频-%d','压测数据','/data/videos/seed.mp4','approved',@owner_id)" "$separator" "$i"
    fi
    separator=,
  done
}

# 等待一个 URL 就绪
wait_url() {
  local name=$1 url=$2 timeout=${3:-120}
  for i in $(seq 1 "$timeout"); do
    if curl --fail --silent --show-error "$url" >/dev/null 2>&1; then
      log "✓ $name ready: $url"; return 0
    fi
    [ "$i" = "$timeout" ] && { log "✗ $name 超时未就绪 $url"; return 1; }
    sleep 2
  done
}

# 采集目标容器的 CPU% / MEM usage（压测最后 20 秒采样）
collect_stats() {
  local project=$1 output_file=$2 duration_sec=${3:-20}
  shift 3
  local containers=("$@")
  local tmp
  tmp=$(mktemp)
  log "📊 采样容器资源 (${duration_sec}s): ${containers[*]}"
  ( for i in $(seq 1 "$duration_sec"); do
      docker stats --no-stream --format '{{.Name}},{{.CPUPerc}},{{.MemUsage}}' 2>/dev/null \
        | grep -E "^${project}" >> "$tmp" || true
      sleep 1
    done ) &
  local sampler_pid=$!
  sleep "$duration_sec"
  kill "$sampler_pid" 2>/dev/null || true
  wait "$sampler_pid" 2>/dev/null || true
  if [ ! -s "$output_file" ]; then
    echo "container,cpu_percent,mem_bytes,mem_limit" > "$output_file"
  fi
  python3 - "$tmp" "$output_file" <<'PY'
import csv, re, sys
src, dst = sys.argv[1], sys.argv[2]
rows = []
with open(src) as f:
    for line in f:
        parts = line.strip().split(',')
        if len(parts) < 3: continue
        name, cpu_s, mem_s = parts[0], parts[1], parts[2]
        try:
            cpu = float(cpu_s.replace('%',''))
        except Exception:
            continue
        m = re.match(r'([\d.]+)([a-zA-Z]+) / ([\d.]+)([a-zA-Z]+)', mem_s)
        if not m: continue
        mem_num, mem_unit, lim_num, lim_unit = m.group(1), m.group(2), m.group(3), m.group(4)
        def to_bytes(n, u):
            n = float(n); u = u.upper()
            for p, mul in [('B',1),('KB',1e3),('MB',1e6),('GB',1e9),('KIB',1024),('MIB',1024**2),('GIB',1024**3)]:
                if u == p: return int(n*mul)
            return int(n)
        rows.append((name, cpu, to_bytes(mem_num, mem_unit), to_bytes(lim_num, lim_unit)))
if not rows:
    sys.exit(0)
from collections import defaultdict
agg = defaultdict(lambda: [0, 0, 0, 0])  # cpu_sum, cpu_n, mem_sum, mem_n
for name, cpu, mem, lim in rows:
    agg[name][0] += cpu; agg[name][1] += 1
    agg[name][2] += mem; agg[name][3] += 1
with open(dst, 'a', newline='') as f:
    w = csv.writer(f)
    limits = {name: lim for name, _, _, lim in rows}
    for name, (cs, cn, ms, mn) in agg.items():
        w.writerow([name, round(cs/cn,2), int(ms/mn) if mn else 0, limits[name]])
PY
  rm -f "$tmp"
}

# 执行一次单接口压测
run_k6_once() {
  local mode=$1       # monolith | microservices
  local bm_id=$2      # 01 / 02 / 03
  local round_n=$3
  local base_url=$4
  local vus=$5
  local duration=$6
  local scripts=("$REPO_ROOT/benchmarks/k6/${bm_id}-"*.js)
  if [ "${#scripts[@]}" -ne 1 ] || [ ! -f "${scripts[0]}" ]; then
    log "[error] bm${bm_id} 必须且只能匹配一个 k6 脚本"
    return 1
  fi
  local script_name
  script_name=$(basename "${scripts[0]}")
  local json_out="$ARTIFACT_DIR/${mode}-bm${bm_id}-r${round_n}.json"
  local text_out="$ARTIFACT_DIR/${mode}-bm${bm_id}-r${round_n}.txt"
  log "▶ k6 $mode bm${bm_id} r${round_n} — VUS=$vus DUR=$duration URL=$base_url"
  docker run --rm --add-host=host.docker.internal:host-gateway \
    --network=host \
    --user "$(id -u):$(id -g)" \
    -v "$REPO_ROOT/benchmarks/k6:/scripts:ro" \
    -v "$ARTIFACT_DIR:/artifacts" \
    -e "K6_BASE_URL=$base_url" \
    -e "K6_VUS=$vus" -e "K6_DURATION=$duration" \
    "$K6_IMAGE" run \
      --summary-export="/artifacts/$(basename "$json_out")" \
      "/scripts/$script_name" > >(tee "$text_out") 2>&1 \
  || log "[warn] k6 bm${bm_id} r${round_n} 返回非零，仍继续后续轮次"
  sleep 3
}

# ── 启动并预热单体栈 ─────────────────────────────────────────
setup_monolith() {
  step "启动单体栈（gateway=$MONO_GATEWAY_PORT frontend=$MONO_FRONTEND_PORT）"
  (
    cd "$REPO_ROOT"
    export GATEWAY_PORT="$MONO_GATEWAY_PORT"
    export FRONTEND_PORT="$MONO_FRONTEND_PORT"
    export MYSQL_PORT="$MONO_MYSQL_PORT"
    export RTMP_PORT="$MONO_RTMP_PORT"
    export SRS_API_PORT="$MONO_SRS_API_PORT"
    export HLS_PORT="$MONO_HLS_PORT"
    docker_compose docker-compose.yml danmaku-bench-mono down --volumes --remove-orphans >/dev/null 2>&1 || true
    docker_compose docker-compose.yml danmaku-bench-mono up --detach --build --remove-orphans
  )
  wait_url "mono gateway"  "http://127.0.0.1:${MONO_GATEWAY_PORT}/gateway/health" 180
  wait_url "mono frontend" "http://127.0.0.1:${MONO_FRONTEND_PORT}/" 120
  log "单体预热数据：运行 seed-test-data.sh（若容器存在则通过 compose exec 触发）"
  (
    cd "$REPO_ROOT"
    PASSWORD_SEED='Test1234!'
    for nick in test_user test_moderator test_admin; do
      curl -s -o /dev/null -X POST "http://127.0.0.1:${MONO_GATEWAY_PORT}/api/v1/auth/register" \
        -H 'Content-Type: application/json' \
        -d "{\"nickname\":\"$nick\",\"password\":\"$PASSWORD_SEED\"}" || true
    done
    # 通过容器 id 更新角色
    mono_mysql_cid=$(docker compose --project-name danmaku-bench-mono ps -q mysql 2>/dev/null || true)
    if [ -n "$mono_mysql_cid" ]; then
      video_values=$(benchmark_video_values 0)
      docker exec "$mono_mysql_cid" mysql -uroot -ppassword danmakustream -e "
        UPDATE users SET role='user'      WHERE nickname='test_user';
        UPDATE users SET role='moderator' WHERE nickname='test_moderator';
        UPDATE users SET role='admin'     WHERE nickname='test_admin';
        SET @owner_id=(SELECT id FROM users WHERE nickname='test_user' LIMIT 1);
        DELETE FROM videos WHERE title LIKE 'BM-公开视频-%';
        INSERT INTO videos (created_at,updated_at,title,description,video_url,status,author_id)
        VALUES $video_values;
      " >/dev/null 2>&1 || true
    fi
  )
  sleep 5
}

# ── 启动并预热微服务栈 ────────────────────────────────────────
setup_microservices() {
  step "启动微服务栈（gateway=$MICRO_GATEWAY_PORT frontend=$MICRO_FRONTEND_PORT）"
  (
    cd "$REPO_ROOT"
    export MICRO_GATEWAY_PORT="$MICRO_GATEWAY_PORT"
    export MICRO_FRONTEND_PORT="$MICRO_FRONTEND_PORT"
    export MICRO_RTMP_PORT="$MICRO_RTMP_PORT"
    export MICRO_HLS_PORT="$MICRO_HLS_PORT"
    docker_compose docker-compose.microservices.yml danmaku-bench-micro down --volumes --remove-orphans >/dev/null 2>&1 || true
    docker_compose docker-compose.microservices.yml danmaku-bench-micro up --detach --build --remove-orphans
  )
  wait_url "micro gateway"  "http://127.0.0.1:${MICRO_GATEWAY_PORT}/gateway/health" 240
  wait_url "micro frontend" "http://127.0.0.1:${MICRO_FRONTEND_PORT}/" 180
  log "微服务数据播种：创建基准账号和公开视频"
  (
    benchmark_user=benchmark-content-owner
    curl -sS -o /dev/null -X POST "http://127.0.0.1:${MICRO_GATEWAY_PORT}/api/v1/auth/register" \
      -H 'Content-Type: application/json' \
      -d "{\"nickname\":\"$benchmark_user\",\"password\":\"Benchmark123!\"}" || true
    micro_mysql_cid=$(docker compose --project-name danmaku-bench-micro ps -q mysql 2>/dev/null || true)
    if [ -n "$micro_mysql_cid" ]; then
      video_values=$(benchmark_video_values 1)
      docker exec "$micro_mysql_cid" mysql -uroot \
        -p"${MYSQL_ROOT_PASSWORD:-local-dev-root-password}" -e "
          SET @owner_id=(SELECT id FROM user_db.users WHERE nickname='$benchmark_user' LIMIT 1);
          DELETE FROM content_db.videos WHERE title LIKE 'BM-公开视频-%';
          INSERT INTO content_db.videos
            (created_at,updated_at,title,description,video_url,status,transcode_status,author_id)
          VALUES $video_values;
        " >/dev/null
    fi
  )
  sleep 5
}

# ── 跑 3 接口 × ROUNDS 轮次 ──────────────────────────────────
run_suite() {
  local mode=$1 base_url=$2
  local stats_project
  case "$mode" in
    monolith) stats_project=danmaku-bench-mono ;;
    microservices) stats_project=danmaku-bench-micro ;;
    *) log "[error] 未知架构模式: $mode"; return 1 ;;
  esac
  step "开始 $mode 压测 × 3 接口 × $ROUNDS 轮"
  local stats_file="$ARTIFACT_DIR/${mode}-stats.csv"
  : > "$stats_file"
  for r in $(seq 1 "$ROUNDS"); do
    step "$mode —— 轮次 $r / $ROUNDS"
    run_k6_once "$mode" "01" "$r" "$base_url" "$VUS_BM01" "$DURATION"
    run_k6_once "$mode" "02" "$r" "$base_url" "$VUS_BM02" "$DURATION"
    # 资源采集放在接口 3（资料库）期间
    local bm03_bg_pid
    ( sleep 10; collect_stats "$stats_project" "$stats_file" 20 ) &
    bm03_bg_pid=$!
    run_k6_once "$mode" "03" "$r" "$base_url" "$VUS_BM03" "$DURATION"
    wait "$bm03_bg_pid" 2>/dev/null || true
  done
}

# ── 生成 comparison.csv / comparison.md ──────────────────────
summarize() {
  step "生成聚合报告"
  python3 - "$ARTIFACT_DIR" "$VUS_BM01" "$VUS_BM02" "$VUS_BM03" "$DURATION" "$ROUNDS" <<'PY'
import json, os, re, sys, csv
from datetime import datetime, timezone
from pathlib import Path
from collections import defaultdict

artifact = Path(sys.argv[1])
vus = sys.argv[2:5]
duration, rounds = sys.argv[5:7]

rows = []  # mode, bm, round, vus, duration, http_reqs, req_rate, avg_ms, p95_ms, err_rate
for mode in ['monolith','microservices']:
    for bm in ['01','02','03']:
        for r in range(1, 99):
            j = artifact / f'{mode}-bm{bm}-r{r}.json'
            if not j.exists(): break
            try:
                d = json.loads(j.read_text(encoding='utf-8'))
            except Exception: continue
            metrics = d.get('metrics', {})
            # 每次迭代恰好执行一个被测业务请求。使用 iterations 和自定义
            # Trend/Rate，避免 setup 中的注册、登录及历史播种污染主请求指标。
            reqs   = metrics.get('iterations',{}).get('count',0)
            rate   = metrics.get('iterations',{}).get('rate',0)
            trend  = metrics.get(f'bm{bm}_duration',{})
            avg    = trend.get('avg',0)
            p95    = trend.get('p(95)',0)
            fail   = metrics.get(f'bm{bm}_errors',{}).get('value',0)
            rows.append([mode,bm,r,int(reqs),round(rate,2),round(avg,2),round(p95,2),round(fail*100,2)])

csv_path = artifact / 'comparison.csv'
with open(csv_path, 'w', newline='') as f:
    w = csv.writer(f)
    w.writerow(['mode','benchmark','round','business_reqs','req_rate_per_sec','avg_ms','p95_ms','error_pct'])
    w.writerows(rows)

# 聚合 3 轮平均
agg = defaultdict(lambda: {'reqs':0,'rate':[],'avg':[],'p95':[],'err':[]})
for mode,bm,r,n,rate,avg,p95,err in rows:
    k = (mode,bm)
    agg[k]['reqs'] += n
    agg[k]['rate'].append(rate)
    agg[k]['avg'].append(avg)
    agg[k]['p95'].append(p95)
    agg[k]['err'].append(err)

def mean(xs): return round(sum(xs)/len(xs),2) if xs else 0
bm_name = {'01':'BM01 公开搜索 GET /videos','02':'BM02 登录 POST /auth/login','03':'BM03 资料库 GET /me/history'}

md = []
md.append('# Day 9 性能压测对比报告（单体 vs 微服务）')
md.append('')
md.append(f'> 生成时间：{datetime.now(timezone.utc).isoformat(timespec="seconds")}；运行 `bash scripts/run-benchmark-comparison.sh` 生成；参数 VUS={"/".join(vus)}，时长 {duration}，轮次 {rounds}')
md.append('')
md.append('## 1. 核心指标对比（单体 vs 微服务，各 N 轮平均值）')
md.append('')
md.append('| 用例 | 架构 | 吞吐(req/s) | 平均耗时(ms) | P95(ms) | 错误率(%) | 总请求数 |')
md.append('|---|---|---:|---:|---:|---:|---:|')
bm_order = ['01','02','03']
for bm in bm_order:
    for mode in ['monolith','microservices']:
        k = (mode,bm)
        a = agg[k]
        if not a['rate']: continue
        md.append(f'| {bm_name[bm]} | {mode} | {mean(a["rate"])} | {mean(a["avg"])} | {mean(a["p95"])} | {mean(a["err"])} | {a["reqs"]} |')
md.append('')
md.append('## 2. 微服务 vs 单体 差异百分比')
md.append('')
md.append('| 用例 | 吞吐差异(%) | 平均耗时差异(%) | P95 差异(%) |')
md.append('|---|---:|---:|---:|')
for bm in bm_order:
    mk = ('monolith',bm); ik = ('microservices',bm)
    ma, ia = agg[mk], agg[ik]
    if not ma['rate'] or not ia['rate']: continue
    def diff_pct(a,b):
        if not b: return 0
        return round((a-b)/b*100,2)
    md.append(f'| {bm_name[bm]} | {diff_pct(mean(ia["rate"]),mean(ma["rate"]))}% | {diff_pct(mean(ia["avg"]),mean(ma["avg"]))}% | {diff_pct(mean(ia["p95"]),mean(ma["p95"]))}% |')
md.append('')
md.append('## 3. 资源占用（压测期间 CPU% / RSS）')
md.append('')
md.append('| 架构 | 容器 | CPU% 平均 | 内存 RSS 平均 | 内存 Limit |')
md.append('|---|---|---:|---:|---:|')
for mode in ['monolith','microservices']:
    csvf = artifact / f'{mode}-stats.csv'
    if csvf.exists():
        with open(csvf) as f:
            rr = list(csv.DictReader(f))
            by_container = defaultdict(lambda: {'cpu': [], 'mem': [], 'limit': 0})
            for r in rr:
                item = by_container[r['container']]
                item['cpu'].append(float(r['cpu_percent'] or 0))
                item['mem'].append(int(r['mem_bytes'] or 0))
                item['limit'] = int(r['mem_limit'] or 0)
            for container, item in sorted(by_container.items()):
                mem = int(mean(item['mem']))
                lim = item['limit']
                def fmt(n):
                    if n>=1e9: return f'{n/1e9:.2f}GB'
                    if n>=1e6: return f'{n/1e6:.2f}MB'
                    if n>=1e3: return f'{n/1e3:.2f}KB'
                    return str(n)
                md.append(f'| {mode} | {container} | {mean(item["cpu"])}% | {fmt(mem)} | {fmt(lim)} |')
md.append('')
md.append('## 4. 单轮明细索引')
md.append('')
for mode in ['monolith','microservices']:
    for bm in ['01','02','03']:
        for r in range(1,99):
            j = artifact / f'{mode}-bm{bm}-r{r}.json'
            t = artifact / f'{mode}-bm{bm}-r{r}.txt'
            if not j.exists(): break
            md.append(f'- `{mode}-bm{bm}-r{r}`: [json]({j.name}) / [stdout]({t.name})')
md.append('')
md.append('## 5. 注意事项 / 待分析')
md.append('')
md.append('- 单体：所有流量通过 nginx-gateway 单 Go 进程处理；微服务：先 nginx-gateway，再 by-path 分发到 user/content/engagement 三服务。')
md.append('- BM02（登录）受密码哈希 + JWT 签发 CPU 开销影响；两套架构实现细节不同，测得差异不能仅归因于服务拆分。若要验证扩容收益，应另做固定单副本与多副本/HPA 的阶梯负载对照。')
md.append('- BM03（资料库历史）会为压测用户写入最多 100 条不同视频的历史；真实生产数据量应扩大到 10k 级别后复测。')
md.append('- CPU% 采样使用 docker stats 1s 间隔 20 次取平均，粗粒度但足够做趋势对比；若要精确数据，请接入 cAdvisor + Prometheus。')
md.append('')

md_path = artifact / 'comparison.md'
md_path.write_text('\n'.join(md) + '\n', encoding='utf-8')
print(f'✅ 生成: {csv_path}  {md_path}')
PY
}

# ── 主流程 ────────────────────────────────────────────────────
main() {
  step "Day 9 压测对比开始"
  log "参数: ROUNDS=$ROUNDS DURATION=$DURATION VUS=$VUS_BM01/$VUS_BM02/$VUS_BM03"
  log "端口: 单体(GW=$MONO_GATEWAY_PORT FE=$MONO_FRONTEND_PORT) 微服务(GW=$MICRO_GATEWAY_PORT FE=$MICRO_FRONTEND_PORT)"
  pull_deps

  setup_monolith
  run_suite monolith "http://127.0.0.1:${MONO_GATEWAY_PORT}"
  step "单体栈完成，释放资源"
  ( cd "$REPO_ROOT" && docker_compose docker-compose.yml danmaku-bench-mono down --volumes --remove-orphans ) >/dev/null 2>&1 || true
  sleep 10

  setup_microservices
  run_suite microservices "http://127.0.0.1:${MICRO_GATEWAY_PORT}"
  step "微服务栈完成，释放资源"
  ( cd "$REPO_ROOT" && docker_compose docker-compose.microservices.yml danmaku-bench-micro down --volumes --remove-orphans ) >/dev/null 2>&1 || true

  summarize
  step "Day 9 压测对比完成"
  log "产物目录: $ARTIFACT_DIR"
  log "报告入口: $ARTIFACT_DIR/comparison.md"
}

cleanup_anyway() {
  local ec=$?
  [ "$ec" != "0" ] && step "异常退出($ec)，尝试清理容器"
  ( cd "$REPO_ROOT" && docker_compose docker-compose.yml danmaku-bench-mono down --volumes --remove-orphans ) >/dev/null 2>&1 || true
  ( cd "$REPO_ROOT" && docker_compose docker-compose.microservices.yml danmaku-bench-micro down --volumes --remove-orphans ) >/dev/null 2>&1 || true
  exit "$ec"
}
trap cleanup_anyway EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

main
