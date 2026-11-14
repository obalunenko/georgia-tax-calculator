#!/usr/bin/env python3
"""Resolve the next strict NNN-feature-name directory without writing it."""

from __future__ import annotations

import argparse
import json
import re
import sys
from pathlib import Path


FEATURE_DIR_RE = re.compile(
    r"^(?P<number>[0-9]{3})-(?P<slug>[a-z0-9]+(?:-[a-z0-9]+)*)$"
)
SLUG_RE = re.compile(r"^[a-z0-9]+(?:-[a-z0-9]+)*$")


def find_repo_root(start: Path) -> Path:
    current = start.resolve()
    for candidate in (current, *current.parents):
        if (candidate / ".specify").is_dir():
            return candidate
    raise ValueError("could not find a Spec Kit project root")


def next_feature(root: Path, slug: str) -> dict[str, str | bool]:
    if not SLUG_RE.fullmatch(slug):
        raise ValueError(
            "feature name must be lowercase ASCII kebab-case "
            "([a-z0-9]+ groups separated by single hyphens)"
        )

    highest = 0
    specs_dir = root / "specs"
    if specs_dir.is_dir():
        for child in specs_dir.iterdir():
            if not child.is_dir():
                continue
            match = FEATURE_DIR_RE.fullmatch(child.name)
            if match:
                highest = max(highest, int(match.group("number"), 10))

    number = highest + 1
    if number > 999:
        raise ValueError("feature number overflow: NNN naming supports 001 through 999")

    feature_num = f"{number:03d}"
    relative_dir = Path("specs") / f"{feature_num}-{slug}"
    if (root / relative_dir).exists():
        raise ValueError(f"target already exists: {relative_dir.as_posix()}")

    return {
        "SPECIFY_FEATURE_DIRECTORY": relative_dir.as_posix(),
        "FEATURE_NUM": feature_num,
        "FEATURE_NAME": slug,
        "BRANCH_CREATED": False,
        "HOOK_DIRECTIVE": (
            "Use SPECIFY_FEATURE_DIRECTORY as the binding explicit target for "
            "this specify invocation; do not recompute or rename it."
        ),
    }


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--slug", required=True)
    parser.add_argument("--repo-root", type=Path)
    args = parser.parse_args()

    try:
        root = (
            args.repo_root.resolve()
            if args.repo_root is not None
            else find_repo_root(Path.cwd())
        )
        result = next_feature(root, args.slug)
    except (OSError, ValueError) as exc:
        print(f"[feature-numbering] Error: {exc}", file=sys.stderr)
        return 1

    print(json.dumps(result, separators=(",", ":")))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
