#!/usr/bin/env python3
"""契约文档一致性校验。

用法（tu-xun 仓库根目录）：
    python3 check_contract.py        # 或 uv run python check_contract.py

校验内容：
1. apifox-import.json 是合法 JSON，且包含 openapi / info / paths / components.schemas；
2. 每个 operation 都有唯一的 operationId 和非空 responses；
3. 所有 $ref 都能在 components.schemas 中解析；
4. api.md 中的接口清单（`METHOD /api/...` 行）与 apifox-import.json 的 paths 双向一致。

全部通过退出码为 0；任何一项失败退出码为 1。改动契约后、提交前必须运行本脚本。
仅依赖标准库。
"""

import json
import re
import sys
from pathlib import Path

# 权威契约文件在仓库根目录（docs/ 的上一级）；docs/ 下仅存放本校验脚本与历史快照。
ROOT = Path(__file__).resolve().parent.parent
JSON_FILE = ROOT / "apifox-import.json"
MD_FILE = ROOT / "api.md"
METHODS = ("get", "post", "put", "delete", "patch")

errors = []
warnings = []


def fail(msg: str) -> None:
    errors.append(msg)


def main() -> int:
    # 1. JSON 合法性与顶层结构
    try:
        spec = json.loads(JSON_FILE.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as e:
        fail(f"apifox-import.json 无法读取或解析: {e}")
        return report()

    for key in ("openapi", "info", "paths"):
        if key not in spec:
            fail(f"apifox-import.json 缺少顶层字段 {key}")
    schemas = spec.get("components", {}).get("schemas")
    if not isinstance(schemas, dict) or not schemas:
        fail("apifox-import.json 缺少 components.schemas")
        schemas = {}
    if errors:
        return report()

    # 2. operationId 与 responses
    json_ops = {}
    op_ids = {}
    for path, item in spec["paths"].items():
        for method in METHODS:
            op = item.get(method)
            if op is None:
                continue
            label = f"{method.upper()} {path}"
            json_ops[label] = op
            op_id = op.get("operationId")
            if not op_id:
                fail(f"{label} 缺少 operationId")
            elif op_id in op_ids:
                fail(f"operationId 重复: {op_id}（{op_ids[op_id]} 与 {label}）")
            else:
                op_ids[op_id] = label
            if not op.get("responses"):
                fail(f"{label} 缺少 responses")

    # 3. $ref 可解析
    def walk(node, where):
        if isinstance(node, dict):
            ref = node.get("$ref")
            if isinstance(ref, str):
                if not ref.startswith("#/components/schemas/"):
                    fail(f"{where} 存在非 components.schemas 的 $ref: {ref}")
                elif ref.rsplit("/", 1)[-1] not in schemas:
                    fail(f"{where} 的 $ref 无法解析: {ref}")
            for k, v in node.items():
                walk(v, where if k != "$ref" else where)
        elif isinstance(node, list):
            for v in node:
                walk(v, where)

    walk(spec["paths"], "paths")
    walk(schemas, "components.schemas")

    # 4. api.md 与 JSON 的接口清单双向一致
    try:
        md_text = MD_FILE.read_text(encoding="utf-8")
    except OSError as e:
        fail(f"api.md 无法读取: {e}")
        return report()

    md_ops = set()
    for m in re.finditer(r"^(GET|POST|PUT|DELETE|PATCH) (/\S+)\s*$", md_text, re.M):
        method, path = m.group(1), m.group(2)
        if not path.startswith("/api/"):
            fail(f"api.md 接口行未以 /api/ 开头: {method} {path}")
            continue
        if re.search(r"/:\w+", path):
            fail(f"api.md 路径参数须用与 OpenAPI 一致的 {{param}} 风格，不用 :param: {method} {path}")
            path = re.sub(r"/:(\w+)", r"/{\1}", path)
        label = f"{method} {path[len('/api'):]}"
        if label in md_ops:
            warnings.append(f"api.md 中接口重复出现: {method} {path}")
        md_ops.add(label)

    for label in sorted(set(json_ops) - md_ops):
        fail(f"apifox-import.json 有而 api.md 没有: {label}")
    for label in sorted(md_ops - set(json_ops)):
        fail(f"api.md 有而 apifox-import.json 没有: {label}")

    # 未被引用的公共 schema（不算失败，仅提示）
    spec_text = json.dumps(spec)
    for name in schemas:
        if f"#/components/schemas/{name}" not in spec_text:
            warnings.append(f"components.schemas.{name} 未被任何地方引用")

    return report(len(json_ops), len(md_ops))


def report(json_count: int = 0, md_count: int = 0) -> int:
    for w in warnings:
        print(f"[提示] {w}")
    if errors:
        for e in errors:
            print(f"[失败] {e}")
        print(f"\n校验失败：{len(errors)} 个问题。")
        return 1
    print(f"校验通过：apifox-import.json {json_count} 个接口，与 api.md {md_count} 个接口一致。")
    return 0


if __name__ == "__main__":
    sys.exit(main())
