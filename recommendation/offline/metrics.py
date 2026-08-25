"""推荐评估指标：HitRate / Recall / NDCG / Coverage / Diversity。"""
import numpy as np


def _rel_keys(relevant):
    return set(int(i) for i in relevant)


def hit_rate_at_k(ranked, relevant, k):
    rel = _rel_keys(relevant)
    top = [int(i) for i in ranked[:k]]
    return 1.0 if any(i in rel for i in top) else 0.0


def recall_at_k(ranked, relevant, k):
    rel = _rel_keys(relevant)
    if not rel:
        return 0.0
    top = [int(i) for i in ranked[:k]]
    return sum(1 for i in top if i in rel) / len(rel)


def precision_at_k(ranked, relevant, k):
    rel = _rel_keys(relevant)
    top = [int(i) for i in ranked[:k]]
    if not top:
        return 0.0
    return sum(1 for i in top if i in rel) / len(top)


def _dcg(scores):
    scores = np.asarray(list(scores), dtype="float64")
    if len(scores) == 0:
        return 0.0
    gains = 2.0 ** scores - 1.0
    discounts = np.log2(np.arange(2, len(scores) + 2))
    return float(np.sum(gains / discounts))


def ndcg_at_k(ranked, relevant, k):
    # relevant: dict[item_id -> relevance score]（评分或二值）
    rel = {int(i): float(v) for i, v in relevant.items()}
    gains = [rel.get(int(i), 0.0) for i in ranked[:k]]
    ideal = sorted(rel.values(), reverse=True)[:k]
    denom = _dcg(ideal)
    return _dcg(gains) / denom if denom > 0 else 0.0


def coverage(ranked_lists, n_items):
    if n_items <= 0:
        return 0.0
    seen = set()
    for rl in ranked_lists:
        seen.update(int(i) for i in rl)
    return len(seen) / n_items


def diversity(ranked_lists, item_tags, k):
    """1 - 平均列表内 tag Jaccard 相似度（越大越多样）。"""
    per_list = []
    for rl in ranked_lists:
        top = [int(i) for i in rl[:k]]
        if len(top) < 2:
            continue
        sims = []
        for a in range(len(top)):
            for b in range(a + 1, len(top)):
                ta = item_tags.get(top[a], set())
                tb = item_tags.get(top[b], set())
                if not ta and not tb:
                    continue
                union = len(ta | tb)
                sims.append(len(ta & tb) / union if union else 0.0)
        per_list.append(1.0 - (float(np.mean(sims)) if sims else 0.0))
    return float(np.mean(per_list)) if per_list else 0.0
