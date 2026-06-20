from __future__ import annotations

from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
EXPERTS_DIR = ROOT / "experts"
INDEX_PATH = EXPERTS_DIR / "INDEX.md"


SECTIONS: list[tuple[str, list[tuple[str, str]]]] = [
    (
        "Main-chain Core",
        [
            ("参谋主帅", "staff/参谋主帅.md"),
        ],
    ),
    (
        "Frontline Owners",
        [
            ("内容主帅", "content/内容主帅.md"),
            ("开发主帅", "dev/开发主帅.md"),
            ("情报主帅", "intel/情报主帅.md"),
            ("数据主帅", "data/数据主帅.md"),
            ("视觉主帅", "visual/视觉主帅.md"),
        ],
    ),
    (
        "On-demand Owner Profiles",
        [
            ("执行底座主帅", "execution_base/执行底座主帅.md"),
            ("进化主帅", "evolve/进化主帅.md"),
        ],
    ),
    (
        "Specialized Entrances",
        [
            ("交付主帅", "expedition/交付主帅.md"),
            ("提示词主帅", "prompt/提示词主帅.md"),
            ("ComfyUI主帅", "comfyui/ComfyUI主帅.md"),
            ("安全主帅", "security/安全主帅.md"),
        ],
    ),
    (
        "Independent Oversight",
        [
            ("白帽纠察官", "oversight/白帽纠察官.md"),
            ("根因雷达官", "oversight/根因雷达官.md"),
            ("审计官", "oversight/审计官.md"),
            ("质检官", "oversight/质检官.md"),
            ("性能基准官", "oversight/性能基准官.md"),
            ("保卫科", "security/保卫科.md"),
            ("合规审计官", "security/合规审计官.md"),
        ],
    ),
]


def ensure_exists(rel_path: str) -> Path:
    path = EXPERTS_DIR / rel_path
    if not path.exists():
        raise SystemExit(f"missing expert card: {path}")
    return path


def render_index() -> str:
    lines = [
        "# 专家索引 / Expert Mirror",
        "",
        "Mirror source: `kernel-source.json`",
        "",
        "本索引只保留当前仍有效的 owner 归口，不再承担第二套运行时说明。",
        "",
    ]
    for title, entries in SECTIONS:
        lines.append(f"## {title}")
        lines.append("")
        for name, rel_path in entries:
            ensure_exists(rel_path)
            target = (Path("E:/wuji-projects/wuji-legion-codex/experts") / rel_path).as_posix()
            lines.append(f"- [{name}]({target})")
        lines.append("")

    lines.extend(
        [
            "## Mirror Rule",
            "",
            "如果本索引与 `kernel-source.json`、`fusion-matrix.json`、`residual-entrypoints.json` 冲突，以结构化真相源为准。",
            "",
            "说明：",
            "",
            "- `参谋主帅` 属于主链参谋内核，不是独立官。",
            "- `执行底座主帅`、`进化主帅` 对应真实 on-demand owner profile。",
            "- `交付主帅`、`提示词主帅`、`ComfyUI主帅`、`安全主帅` 是专项入口，不单列为顶层 owner profile，也不拥有独立写权限。",
            "",
        ]
    )
    return "\n".join(lines)


def main() -> None:
    INDEX_PATH.write_text(render_index(), encoding="utf-8")
    print(f"OK rebuilt expert index: {INDEX_PATH}")
    print("NOTE legacy expert-card regeneration is retired; checked-in cards remain the source for role text.")


if __name__ == "__main__":
    main()
