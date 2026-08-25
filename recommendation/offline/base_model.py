"""模型基类：定义统一的 fit / recommend 接口与候选集管理。

所有模型共享同一候选集（candidates），保证评估公平：
- candidates 默认 = 训练集中出现过的所有 item；
- 大数据集下可传入 min_item_interactions 过滤后的候选集（见 evaluate.build_candidates）。
"""
import numpy as np


class BaseModel:
    name = "base"

    def __init__(self, **kwargs):
        self._params = kwargs
        self._train = None
        self._items = None
        self._item_ids = None  # 候选 item id（np.int64 数组）
        self._item_index = {}  # item_id -> 候选下标
        self._user_hist = {}  # user_id -> [item_id, ...]

    def fit(self, events, items, candidates=None):
        self._train = events
        self._items = items
        if candidates is None:
            candidates = events["item_id"].unique()
        self._set_candidates(np.sort(np.asarray(candidates, dtype="int64")))
        self._user_hist = (
            events.groupby("user_id")["item_id"].apply(lambda s: [int(x) for x in s]).to_dict()
        )
        self._build()
        return self

    def _set_candidates(self, item_ids):
        self._item_ids = np.asarray(item_ids, dtype="int64")
        self._item_index = {int(i): idx for idx, i in enumerate(self._item_ids)}

    def _build(self):
        """子类在 fit 后构建模型结构。"""

    def recommend(self, user_id, k=10):
        raise NotImplementedError

    def _topk_scores(self, user_id, scores, k):
        """按分数降序取 top-k，并屏蔽训练期已看过的 item。"""
        s = np.asarray(scores, dtype="float64").copy()
        for it in self._user_hist.get(int(user_id), []):
            idx = self._item_index.get(int(it))
            if idx is not None:
                s[idx] = -np.inf
        order = np.argsort(s)[::-1][:k]
        return [int(self._item_ids[i]) for i in order if np.isfinite(s[i])]
