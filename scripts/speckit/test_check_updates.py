#!/usr/bin/env python3
"""Focused tests for the read-only Spec Kit update checker."""

from __future__ import annotations

import importlib.util
import unittest
from pathlib import Path


SCRIPT = Path(__file__).with_name("check-updates.py")
SPEC = importlib.util.spec_from_file_location("check_speckit_updates", SCRIPT)
if SPEC is None or SPEC.loader is None:
    raise RuntimeError(f"could not load {SCRIPT}")
checker = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(checker)


class CheckSpecKitUpdatesTest(unittest.TestCase):
    def test_mixed_companion_releases_are_split_by_product(self) -> None:
        releases = [
            {"tag_name": "v0.31.4"},
            {"tag_name": "speckit-ext-v0.20.2"},
            {"tag_name": "v0.31.3"},
            {"tag_name": "speckit-ext-v0.20.1"},
        ]

        self.assertEqual(
            "0.31.4",
            checker.latest_matching_version(
                releases,
                "tag_name",
                checker.COMPANION_EDITOR_TAG_RE,
                checker.COMPANION_REPOSITORY,
            ),
        )
        self.assertEqual(
            "0.20.2",
            checker.latest_matching_version(
                releases,
                "tag_name",
                checker.COMPANION_SPECKIT_TAG_RE,
                checker.COMPANION_REPOSITORY,
            ),
        )

    def test_editor_version_accepts_current_and_legacy_publishers(self) -> None:
        self.assertEqual(
            "0.31.4",
            checker.editor_extension_version(
                "other.extension@1.0.0\nalfredoperez.speckit-companion@0.31.4\n",
            ),
        )
        self.assertEqual(
            "0.1.3",
            checker.editor_extension_version("alfredo-dev.speckit-companion@0.1.3\n"),
        )

    def test_editor_update_state_is_not_described_as_a_pin(self) -> None:
        self.assertEqual("UPDATE AVAILABLE", checker.installed_update_state("0.31.3", "0.31.4"))
        self.assertEqual("installed is current", checker.installed_update_state("0.31.4", "0.31.4"))


if __name__ == "__main__":
    unittest.main()
