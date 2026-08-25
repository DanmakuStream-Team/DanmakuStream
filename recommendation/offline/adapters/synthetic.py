"""合成数据生成器：无网络时用于冒烟测试 / CI / 快速验证算法闭环。

生成规则（保证 ItemCF 能学到结构）：
- 每个 item 有 1~3 个 tag（genres）。
- 每个用户有隐式的 tag 偏好分布（集中式 Dirichlet，偏好少数几个 tag）。
- 交互按"先按偏好选 tag，再选带该 tag 的 item"生成，形成 tag 聚集的共现结构。
- timestamp 单调递增，支持时间切分。
"""
import numpy as np
import pandas as pd

from .base import Dataset


def generate(
    n_users: int = 400,
    n_items: int = 400,
    n_genres: int = 8,
    min_events: int = 15,
    max_events: int = 50,
    seed: int = 42,
) -> Dataset:
    rng = np.random.default_rng(seed)
    genres = [f"g{i}" for i in range(n_genres)]
    genre_index = {g: i for i, g in enumerate(genres)}

    # item -> tags（1~3 个）
    item_tags: list[list[str]] = []
    for _ in range(n_items):
        k = int(rng.integers(1, 4))
        item_tags.append([str(g) for g in rng.choice(genres, size=k, replace=False)])

    # tag -> 含该 tag 的 item 索引（便于采样）
    genre_items = {g: [i for i, tags in enumerate(item_tags) if g in tags] for g in genres}

    # user -> tag 偏好（集中式 Dirichlet，偏好少数 tag）
    user_pref = rng.dirichlet(np.ones(n_genres) * 0.5, size=n_users)

    rows = []
    for u in range(n_users):
        n_events = int(rng.integers(min_events, max_events + 1))
        for _ in range(n_events):
            g = str(rng.choice(genres, p=user_pref[u]))
            i = int(rng.choice(genre_items[g]))
            p = float(np.mean([user_pref[u][genre_index[t]] for t in item_tags[i]]))
            rating = float(np.clip(2.5 + p * 2.5 + rng.normal(0.0, 0.4), 1.0, 5.0))
            rows.append((u, i, rating))

    events = pd.DataFrame(rows, columns=["user_id", "item_id", "rating"])
    events = events.drop_duplicates(subset=["user_id", "item_id"]).reset_index(drop=True)
    # 单调递增时间戳 + 轻微抖动
    events["timestamp"] = (
        np.arange(len(events), dtype="int64") * 10 + rng.integers(0, 10, size=len(events))
    )
    events["user_id"] = events["user_id"].astype("int64")
    events["item_id"] = events["item_id"].astype("int64")

    items = pd.DataFrame(
        {
            "item_id": np.arange(n_items, dtype="int64"),
            "title": [f"video-{i}" for i in range(n_items)],
            "tags": item_tags,
        }
    )

    return Dataset(name="synthetic", events=events, items=items)
