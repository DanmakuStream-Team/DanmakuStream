"""命令行入口：加载数据集并运行多模型对比实验。

用法（在 recommendation/ 目录下）：

    # 合成数据冒烟（无需下载）
    python -m offline --dataset synthetic

    # MovieLens（先下载解压，目录含 ratings.csv / movies.csv）
    python -m offline --dataset movielens --data-dir data/ml-latest-small

    # KuaiRec（目录含 big_matrix.csv）
    python -m offline --dataset kuairec --data-dir data/KuaiRec
"""
import argparse

from .adapters import KuaiRecAdapter, MovieLensAdapter, generate
from .evaluate import run_experiment


def _build_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(description="DanmakuStream 推荐算法离线评估")
    p.add_argument(
        "--dataset",
        choices=["movielens", "kuairec", "synthetic"],
        default="synthetic",
        help="数据集（默认 synthetic 用于冒烟测试）",
    )
    p.add_argument("--data-dir", default=None, help="movielens/kuairec 数据目录")
    p.add_argument("--k", type=int, default=10, help="Top-K（默认 10）")
    p.add_argument("--test-ratio", type=float, default=0.2, help="时间切分测试比例")
    p.add_argument("--min-user-interactions", type=int, default=3)
    p.add_argument("--min-item-interactions", type=int, default=2)
    p.add_argument("--top-k-neighbors", type=int, default=100, help="ItemCF 每个 item 保留的邻居数")
    p.add_argument("--shrink", type=float, default=10.0, help="余弦相似度 shrink 平滑项")
    p.add_argument("--n-users", type=int, default=400, help="synthetic：用户数")
    p.add_argument("--n-items", type=int, default=400, help="synthetic：物品数")
    p.add_argument("--n-genres", type=int, default=8, help="synthetic：tag 数")
    p.add_argument("--min-events", type=int, default=15, help="synthetic：每用户最少交互")
    p.add_argument("--max-events", type=int, default=50, help="synthetic：每用户最多交互")
    p.add_argument("--seed", type=int, default=42)
    return p


def main(argv=None):
    args = _build_parser().parse_args(argv)

    if args.dataset == "movielens":
        if not args.data_dir:
            raise SystemExit("movielens 需要 --data-dir 指向解压后的目录（含 ratings.csv / movies.csv）")
        dataset = MovieLensAdapter().load(args.data_dir)
    elif args.dataset == "kuairec":
        if not args.data_dir:
            raise SystemExit("kuairec 需要 --data-dir 指向解压后的目录（含 big_matrix.csv）")
        dataset = KuaiRecAdapter().load(args.data_dir)
    else:
        dataset = generate(
            n_users=args.n_users,
            n_items=args.n_items,
            n_genres=args.n_genres,
            min_events=args.min_events,
            max_events=args.max_events,
            seed=args.seed,
        )

    print(f"数据集: {dataset.name}  交互数: {len(dataset.events)}  "
          f"用户数: {dataset.events['user_id'].nunique()}  "
          f"物品数: {dataset.events['item_id'].nunique()}")
    print(f"参数: k={args.k} test_ratio={args.test_ratio} "
          f"top_k_neighbors={args.top_k_neighbors} shrink={args.shrink}")
    print()

    table, train, test = run_experiment(
        dataset,
        k=args.k,
        test_ratio=args.test_ratio,
        min_user_interactions=args.min_user_interactions,
        min_item_interactions=args.min_item_interactions,
        top_k_neighbors=args.top_k_neighbors,
        shrink=args.shrink,
        seed=args.seed,
    )

    print(f"train={len(train)}  test={len(test)}")
    print()
    print(table.round(4).to_string())
    return table


if __name__ == "__main__":
    main()
