from __future__ import annotations

import re
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
MODELS = ROOT / "docs" / "models"
REQUIRED_DIRS = ("usecase", "system", "component", "object", "deployment")
SEQUENCE_NAME = re.compile(r"^(SYS-SEQ|COMP-SEQ|OBJ-SEQ)\d{2}$")


def main() -> int:
    errors: list[str] = []

    for directory in REQUIRED_DIRS:
        if not (MODELS / directory).is_dir():
            errors.append(f"缺少模型目录: docs/models/{directory}")

    sources = sorted(path for path in MODELS.rglob("*.puml") if path.name != "_theme.puml")
    if not sources:
        errors.append("docs/models 中没有正式 PlantUML 源文件")

    for source in sources:
        content = source.read_text(encoding="utf-8")
        for suffix in (".svg", ".png"):
            exported = source.with_suffix(suffix)
            if not exported.is_file() or exported.stat().st_size == 0:
                errors.append(f"{source.relative_to(ROOT)} 缺少同名 {suffix} 导出图")

        if source.stem.startswith(("SYS-SEQ", "COMP-SEQ", "OBJ-SEQ")) and not SEQUENCE_NAME.fullmatch(source.stem):
            errors.append(f"顺序图编号不规范: {source.relative_to(ROOT)}")

        if "CLASS" in source.stem:
            empty_classes = re.findall(r"^\s*class\s+[^\s{]+\s*$", content, flags=re.MULTILINE)
            if empty_classes:
                errors.append(
                    f"类图存在无属性/方法的空类: {source.relative_to(ROOT)} -> "
                    + ", ".join(item.strip() for item in empty_classes)
                )

    for required in ("DEPLOY-MONO", "DEPLOY-K8S"):
        if not (MODELS / "deployment" / f"{required}.puml").is_file():
            errors.append(f"缺少部署图: docs/models/deployment/{required}.puml")

    for document in (ROOT / "docs").rglob("*.md"):
        content = document.read_text(encoding="utf-8")
        if re.search(r"```\s*mermaid\b", content, flags=re.IGNORECASE):
            errors.append(f"正式文档仍包含 Mermaid 图: {document.relative_to(ROOT)}")

    if errors:
        print("PlantUML 资产检查失败：")
        for error in errors:
            print(f"- {error}")
        return 1

    print(f"PlantUML 资产检查通过：{len(sources)} 张正式图均有 .puml、.svg 和 .png。")
    return 0


if __name__ == "__main__":
    sys.exit(main())
