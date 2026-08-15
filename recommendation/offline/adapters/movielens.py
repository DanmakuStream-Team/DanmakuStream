"""MovieLens 适配器（ml-latest-small / ml-100k 通用）。

下载：https://files.grouplens.org/datasets/movielens/ml-latest-small.zip
解压后目录应包含 ratings.csv / movies.csv（tags.csv 可选）。

genres 字段被当作 tag（管道符分隔）。
"""
import os

import pandas as pd

from .base import Dataset


def _split_genres(value: str) -> list[str]:
    parts = [p.strip() for p in str(value).split("|")]
    return [p for p in parts if p and p != "(no genres listed)"]


class MovieLensAdapter:
    name = "movielens"

    def load(self, data_dir: str) -> Dataset:
        ratings_path = os.path.join(data_dir, "ratings.csv")
        movies_path = os.path.join(data_dir, "movies.csv")
        if not (os.path.isfile(ratings_path) and os.path.isfile(movies_path)):
            raise FileNotFoundError(
                f"{data_dir} 下找不到 ratings.csv / movies.csv，"
                "请先下载并解压 MovieLens 数据集"
            )

        ratings = pd.read_csv(ratings_path)
        movies = pd.read_csv(movies_path)

        events = pd.DataFrame(
            {
                "user_id": ratings["userId"].astype("int64"),
                "item_id": ratings["movieId"].astype("int64"),
                "rating": ratings["rating"].astype(float),
                "timestamp": ratings["timestamp"].astype("int64"),
            }
        )

        items = pd.DataFrame(
            {
                "item_id": movies["movieId"].astype("int64"),
                "title": movies["title"],
            }
        )
        items["tags"] = movies["genres"].apply(_split_genres)

        return Dataset(name=self.name, events=events, items=items)
