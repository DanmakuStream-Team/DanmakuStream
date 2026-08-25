"""统一数据格式。

所有 Adapter 输出 Dataset：

- events: DataFrame，列 user_id(int64) / item_id(int64) / rating(float) / timestamp(int64)
  rating 可缺省（隐式反馈视为 1.0）；timestamp 用于时间切分。
- items:  DataFrame，列 item_id(int64) / title(str) / tags(list[str])

算法层只依赖这个格式，不关心原始数据源。
"""
from dataclasses import dataclass

import pandas as pd


@dataclass
class Dataset:
    name: str
    events: pd.DataFrame
    items: pd.DataFrame

    def __post_init__(self) -> None:
        for col in ("user_id", "item_id"):
            if col not in self.events.columns:
                raise ValueError(f"events 缺少列 {col}")
            self.events[col] = self.events[col].astype("int64")

        if "rating" not in self.events.columns:
            self.events["rating"] = 1.0
        self.events["rating"] = self.events["rating"].astype(float)

        if "timestamp" not in self.events.columns:
            self.events["timestamp"] = 0
        self.events["timestamp"] = self.events["timestamp"].astype("int64")

        if "item_id" not in self.items.columns:
            raise ValueError("items 缺少列 item_id")
        self.items["item_id"] = self.items["item_id"].astype("int64")

        if "title" not in self.items.columns:
            self.items["title"] = self.items["item_id"].astype(str)

        if "tags" not in self.items.columns:
            self.items["tags"] = [[] for _ in range(len(self.items))]
