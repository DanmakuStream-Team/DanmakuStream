"""评估流水线：时间切分 + 多模型对比。"""
import numpy as np
import pandas as pd

from .baselines import PopularModel, RandomModel, TagModel
from .itemcf import ItemCF
from .metrics import coverage, diversity, hit_rate_at_k, ndcg_at_k, recall_at_k


def temporal_split(events, test_ratio=0.2, min_user_interactions=3):
    """按时间切分：每个用户按时间排序，最近 test_ratio 比例的交互进测试集。

    仅保留交互数 >= min_user_interactions 的用户（否则无法同时训练与测试）。
    """
    df = events.sort_values("timestamp", kind="mergesort")
    counts = df.groupby("user_id").size()
    valid_users = counts[counts >= min_user_interactions].index
    df = df[df["user_id"].isin(valid_users)]

    train_parts, test_parts = [], []
    for _, grp in df.groupby("user_id", sort=False):
        grp = grp.sort_values("timestamp", kind="mergesort")
        n_test = max(1, int(len(grp) * test_ratio))
        test_parts.append(grp.iloc[-n_test:])
        train_parts.append(grp.iloc[:-n_test])

    train = pd.concat(train_parts) if train_parts else df.iloc[0:0]
    test = pd.concat(test_parts) if test_parts else df.iloc[0:0]
    return train.reset_index(drop=True), test.reset_index(drop=True)


def build_candidates(events, min_item_interactions=2):
    """候选集 = 训练集中交互数 >= min_item_interactions 的 item。"""
    counts = events.groupby("item_id").size()
    return counts[counts >= min_item_interactions].index.values


def evaluate_model(model, train, test, items, k=10, candidates=None, sample=None, seed=42):
    model.fit(train, items, candidates=candidates)

    item_tags = {int(r.item_id): set(r.tags) for r in items.itertuples()}
    candidate_set = set(int(i) for i in model._item_ids)

    # 组装每个用户的测试相关集（item -> relevance）
    test_rel: dict[int, dict[int, float]] = {}
    test_users = []
    for uid, grp in test.groupby("user_id", sort=False):
        rel = {}
        for item_id, rating in zip(grp["item_id"].astype("int64"), grp["rating"].astype(float)):
            if int(item_id) in candidate_set:
                rel[int(item_id)] = float(rating)
        if rel:
            test_rel[int(uid)] = rel
            test_users.append(int(uid))

    rng = np.random.default_rng(seed)
    hit, rec, ndcg = [], [], []
    ranked_lists = []
    for u in test_users:
        rl = model.recommend(u, k=k)
        ranked_lists.append(rl)
        rel = test_rel[u]
        hit.append(hit_rate_at_k(rl, rel, k))
        rec.append(recall_at_k(rl, rel, k))
        ndcg.append(ndcg_at_k(rl, rel, k))

    n_items = len(model._item_ids)
    return {
        "hit_rate": float(np.mean(hit)) if hit else 0.0,
        "recall": float(np.mean(rec)) if rec else 0.0,
        "ndcg": float(np.mean(ndcg)) if ndcg else 0.0,
        "coverage": coverage(ranked_lists, n_items),
        "diversity": diversity(ranked_lists, item_tags, k),
        "n_users": len(test_users),
    }


def run_experiment(
    dataset,
    k=10,
    test_ratio=0.2,
    min_user_interactions=3,
    min_item_interactions=2,
    top_k_neighbors=100,
    shrink=10.0,
    seed=42,
):
    train, test = temporal_split(
        dataset.events, test_ratio=test_ratio, min_user_interactions=min_user_interactions
    )
    candidates = build_candidates(train, min_item_interactions=min_item_interactions)

    models = [
        RandomModel(seed=seed),
        PopularModel(),
        TagModel(),
        ItemCF(top_k=top_k_neighbors, shrink=shrink),
    ]

    rows = []
    for m in models:
        res = evaluate_model(m, train, test, dataset.items, k=k, candidates=candidates, seed=seed)
        res["model"] = m.name
        rows.append(res)

    table = pd.DataFrame(rows).set_index("model")
    table = table[["hit_rate", "recall", "ndcg", "coverage", "diversity", "n_users"]]
    return table, train, test
