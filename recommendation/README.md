# Recommendation Service（离线部分）

DanmakuStream 推荐系统的**离线训练与评估**模块：在公开数据集上验证 ItemCF 等算法是否有效。

## 安装（conda）

```bash
# 新建 conda 环境（一次性）
conda create -y -n danmaku-rec python=3.11 numpy pandas scipy

# 激活
conda activate danmaku-rec
```

> 也可用 `requirements.txt` + pip 安装到任意环境，但本项目约定使用 conda。

## 快速冒烟（无需下载数据）

```bash
conda activate danmaku-rec
python -m offline --dataset synthetic
```

## 跑 MovieLens

```bash
conda activate danmaku-rec

# 下载并解压（约 1MB）
mkdir -p data && cd data
curl -LO https://files.grouplens.org/datasets/movielens/ml-latest-small.zip
unzip ml-latest-small.zip
cd ..

python -m offline --dataset movielens --data-dir data/ml-latest-small --k 10
```

## 跑 KuaiRec

KuaiRec 数据量大，下载后（`data/big_matrix.csv`）：

```bash
conda activate danmaku-rec
python -m offline --dataset kuairec --data-dir data/KuaiRec --min-item-interactions 5
```

## 输出说明

输出一张对比表，指标含义：

| 指标 | 含义 |
|---|---|
| hit_rate@k | 测试集中至少命中一条的用户占比 |
| recall@k | 测试项被召回的比例 |
| ndcg@k | 归一化折损累计增益（越靠前越高） |
| coverage | 推荐结果覆盖目录的比例 |
| diversity | 1 - 列表内 tag 平均相似度（越大越多样） |

**验收标准**：ItemCF 的 recall@k / ndcg@k 应显著优于 popular 基线，明显优于 random。

**实测参考（ml-latest-small，k=10，时间切分 20%）**：

| 模型 | recall@10 | ndcg@10 | hit_rate@10 |
|---|---|---|---|
| random | 0.002 | 0.003 | 0.054 |
| popular | 0.041 | 0.070 | 0.377 |
| itemcf | 0.062 | 0.090 | 0.482 |

> MovieLens 的 genre 只有 19 个粗粒度标签，tag 基线偏弱属正常现象；ItemCF 借助物品共现信号表现最佳。

## 目录结构

```text
offline/
├── adapters/     # movielens / kuairec / synthetic → 统一 Dataset
├── base_model.py # 模型基类（统一 fit/recommend + 候选集管理）
├── baselines.py  # Random / Popular / Tag 基线
├── itemcf.py     # ItemCF（稀疏余弦 + shrink + top-k 截断）
├── metrics.py    # HitRate / Recall / NDCG / Coverage / Diversity
├── evaluate.py   # 时间切分 + 多模型对比流水线
└── run.py        # CLI 入口
```
