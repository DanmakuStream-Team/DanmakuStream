"""数据适配器：把不同数据源转成统一的 Dataset 格式。"""

from .base import Dataset
from .movielens import MovieLensAdapter
from .kuairec import KuaiRecAdapter
from .synthetic import generate

__all__ = ["Dataset", "MovieLensAdapter", "KuaiRecAdapter", "generate"]
