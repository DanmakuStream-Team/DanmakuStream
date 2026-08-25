"""基线模型：Random / Popular / Tag。"""
import numpy as np

from .base_model import BaseModel


class RandomModel(BaseModel):
    name = "random"

    def _build(self):
        self._rng = np.random.default_rng(self._params.get("seed", 42))

    def recommend(self, user_id, k=10):
        scores = self._rng.random(len(self._item_ids))
        return self._topk_scores(user_id, scores, k)


class PopularModel(BaseModel):
    name = "popular"

    def _build(self):
        pop = self._train.groupby("item_id").size()
        self._scores = np.zeros(len(self._item_ids), dtype="float64")
        for idx, it in enumerate(self._item_ids):
            self._scores[idx] = float(pop.get(int(it), 0))

    def recommend(self, user_id, k=10):
        return self._topk_scores(user_id, self._scores, k)


class TagModel(BaseModel):
    name = "tag"

    def _build(self):
        # item_id -> tag 集合
        self._item_tags = {
            int(r.item_id): set(r.tags) for r in self._items.itertuples()
        }
        # 热度兜底（用户无 tag 历史时退化为热门）
        pop = self._train.groupby("item_id").size()
        self._pop_scores = np.zeros(len(self._item_ids), dtype="float64")
        for idx, it in enumerate(self._item_ids):
            self._pop_scores[idx] = float(pop.get(int(it), 0))

        # user_id -> tag 计数
        self._user_tags: dict[int, dict[str, int]] = {}
        for uid, grp in self._train.groupby("user_id"):
            cnt: dict[str, int] = {}
            for it in grp["item_id"].astype("int64"):
                for t in self._item_tags.get(int(it), set()):
                    cnt[t] = cnt.get(t, 0) + 1
            self._user_tags[int(uid)] = cnt

    def recommend(self, user_id, k=10):
        utags = self._user_tags.get(int(user_id), {})
        if not utags:
            return self._topk_scores(user_id, self._pop_scores, k)
        scores = np.zeros(len(self._item_ids), dtype="float64")
        for idx, it in enumerate(self._item_ids):
            tags = self._item_tags.get(int(it), set())
            if not tags:
                continue
            scores[idx] = sum(utags.get(t, 0) for t in tags) / len(tags)
        return self._topk_scores(user_id, scores, k)
