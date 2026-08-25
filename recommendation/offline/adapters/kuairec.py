"""KuaiRec 适配器（快手真实短视频推荐数据集，尽力兼容其原始格式）。

KuaiRec 原始仓库：https://github.com/chongminggao/KuaiRec
主交互文件为 data/big_matrix.csv，列名形如：
    user_id, video_id, date, ..., watch_ratio, ...

本适配器做了宽松处理：
- 交互文件需包含 user_id 与 video_id（或 item_id）。
- timestamp 优先取 timestamp 列，其次 date 列（形如 20201128），否则按行号。
- rating 优先取 rating 列，其次 watch_ratio（×5 归一化到 1~5），否则 1.0（隐式）。
- 可选 items_file（如 item_categories.csv）：若含 categories 类列则作为 tag。

KuaiRec 数据量大，建议离线评估时配合候选采样（见 evaluate.run_experiment 的 sample 参数）。
"""
import os

import numpy as np
import pandas as pd

from .base import Dataset


def _to_ts(series: pd.Series) -> pd.Series:
    s = pd.to_numeric(series, errors="coerce")
    # date 形如 20201128 视为日期编号，直接当整数时间戳使用即可
    return s.fillna(0).astype("int64")


class KuaiRecAdapter:
    name = "kuairec"

    def load(
        self,
        data_dir: str,
        interactions_file: str = "big_matrix.csv",
        items_file: str | None = None,
    ) -> Dataset:
        if not data_dir:
            raise ValueError("KuaiRec 需要 --data-dir 指向解压后的数据目录")
        path = os.path.join(data_dir, interactions_file)
        if not os.path.isfile(path):
            raise FileNotFoundError(f"找不到交互文件 {path}")

        df = pd.read_csv(path)

        item_col = "video_id" if "video_id" in df.columns else "item_id"
        if item_col not in df.columns:
            raise ValueError("交互文件缺少 video_id / item_id 列")
        if "user_id" not in df.columns:
            raise ValueError("交互文件缺少 user_id 列")

        if "timestamp" in df.columns:
            ts = _to_ts(df["timestamp"])
        elif "date" in df.columns:
            ts = _to_ts(df["date"])
        else:
            ts = pd.Series(np.arange(len(df)), dtype="int64")

        if "rating" in df.columns:
            rating = pd.to_numeric(df["rating"], errors="coerce").fillna(1.0)
        elif "watch_ratio" in df.columns:
            rating = pd.to_numeric(df["watch_ratio"], errors="coerce").fillna(0.0) * 5.0
            rating = rating.clip(1.0, 5.0)
        else:
            rating = pd.Series(np.ones(len(df)), dtype=float)

        events = pd.DataFrame(
            {
                "user_id": df["user_id"].astype("int64"),
                "item_id": df[item_col].astype("int64"),
                "rating": rating.astype(float),
                "timestamp": ts,
            }
        )

        # items：若提供了分类文件则读取 tag，否则空 tag
        items = pd.DataFrame({"item_id": np.sort(events["item_id"].unique())})
        items["title"] = items["item_id"].astype(str)
        items["tags"] = [[] for _ in range(len(items))]
        if items_file:
            cat_path = os.path.join(data_dir, items_file)
            if os.path.isfile(cat_path):
                cat = pd.read_csv(cat_path)
                cat_item = "video_id" if "video_id" in cat.columns else "item_id"
                # 常见分类列名做一次宽松匹配
                cat_col = next(
                    (c for c in cat.columns if "categor" in c.lower() or "type" in c.lower()),
                    None,
                )
                if cat_item in cat.columns:
                    merged = pd.merge(
                        items[["item_id"]],
                        cat[[cat_item, cat_col]] if cat_col else cat[[cat_item]],
                        left_on="item_id",
                        right_on=cat_item,
                        how="left",
                    )
                    if cat_col:
                        merged["tags"] = merged[cat_col].fillna("").astype(str).apply(
                            lambda v: [p for p in v.split(",") if p]
                        )
                        items["tags"] = merged["tags"].tolist()

        return Dataset(name=self.name, events=events, items=items)
