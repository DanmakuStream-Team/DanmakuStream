"""ItemCF：基于物品的协同过滤。

- 用户-物品交互矩阵按隐式反馈二值化（有交互=1）。
- 物品相似度用余弦相似度 + shrink 平滑，只保留每件物品 top_k 个邻居（稀疏，省内存）。
- 打分 = 用户历史向量 × 物品相似度矩阵。
"""
import numpy as np
from scipy.sparse import csr_matrix, lil_matrix

from .base_model import BaseModel


def build_item_similarity(user_item: csr_matrix, top_k: int, shrink: float) -> csr_matrix:
    """user_item: (n_users x n_items) 稀疏矩阵。返回 (n_items x n_items) 稀疏相似度。"""
    item_user = user_item.T.tocsr()  # n_items x n_users
    n_items = item_user.shape[0]

    pop = np.asarray(user_item.sum(axis=0)).ravel().astype("float64")

    # 共现矩阵（稀疏），余弦 = cooc(i,j) / (sqrt(pop_i * pop_j) + shrink)
    cooc = (item_user @ user_item).tocoo()
    denom = np.sqrt(pop[cooc.row] * pop[cooc.col]) + shrink
    vals = cooc.data / denom
    sim_full = csr_matrix((vals, (cooc.row, cooc.col)), shape=(n_items, n_items))

    sim = _truncate_topk(sim_full, top_k)
    sim.setdiag(0.0)
    sim.eliminate_zeros()
    return sim.tocsr()


def _truncate_topk(mat: csr_matrix, top_k: int) -> csr_matrix:
    """每行只保留 top_k 个最大值，避免稠密矩阵内存爆炸。"""
    if top_k is None or top_k <= 0:
        return mat.tocsr()
    lil = mat.tolil()
    for i in range(lil.shape[0]):
        data = lil.data[i]
        if len(data) > top_k:
            idx = np.argsort(data)[::-1][:top_k]
            lil.rows[i] = [lil.rows[i][j] for j in idx]
            lil.data[i] = [data[j] for j in idx]
    return lil.tocsr()


class ItemCF(BaseModel):
    name = "itemcf"

    def __init__(self, top_k=100, shrink=10.0, **kwargs):
        super().__init__(**kwargs)
        self.top_k = top_k
        self.shrink = shrink
        self._item_sim = None

    def _build(self):
        # 训练交互（只保留候选集内的 item）
        train = self._train
        mask = train["item_id"].astype("int64").isin(self._item_ids)
        sub = train[mask]
        users = sub["user_id"].astype("int64").values
        items = sub["item_id"].astype("int64").values

        user_ids = np.unique(users)
        user_index = {int(u): i for i, u in enumerate(user_ids)}
        rows = np.array([user_index[int(u)] for u in users], dtype="int64")
        cols = np.array([self._item_index[int(i)] for i in items], dtype="int64")
        data = np.ones(len(rows), dtype="float64")

        user_item = csr_matrix((data, (rows, cols)), shape=(len(user_ids), len(self._item_ids)))
        self._item_sim = build_item_similarity(user_item, self.top_k, self.shrink)

    def recommend(self, user_id, k=10):
        hist_vec = np.zeros(len(self._item_ids), dtype="float64")
        for it in self._user_hist.get(int(user_id), []):
            idx = self._item_index.get(int(it))
            if idx is not None:
                hist_vec[idx] += 1.0
        if self._item_sim is None:
            return []
        scores = self._item_sim.dot(hist_vec)
        return self._topk_scores(user_id, scores, k)
