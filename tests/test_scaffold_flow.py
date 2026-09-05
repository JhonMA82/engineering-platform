"""Scaffold flow test: project idea -> generated tree meets expectations.

For each scenario this writes a blueprint project (no network, no installs)
and verifies: selected starters/surfaces, generated files (GENTLE.md,
AGENTS.md, ARCHITECTURE.md, .engineering/*), their key contents, and a
clean `doctor` (inspect_project without errors).
"""

from __future__ import annotations

import tempfile
import unittest
from json import loads
from pathlib import Path

from scripts.eng import inspect_project, write_project


QUEJAS_INTAKE = {
    "name": "quejas-ciudadanas",
    "project_type": "multi-app",
    "signals": ["shared-api", "public-intake", "kiosk-mode"],
    "database": "postgresql-managed",
    "surfaces": [
        {
            "id": "public-intake",
            "capabilities": ["form-capture", "offline-outbox", "tracking-token", "kiosk-mode"],
        }
    ],
}

ADMIN_INTAKE = {
    "name": "admin-solo",
    "project_type": "institutional-admin",
    "signals": ["workflow", "requests"],
    "database": "postgresql-managed",
}

SIGNALS_ONLY_INTAKE = {
    "name": "portal-senales",
    "project_type": "multi-app",
    "signals": ["shared-api", "kiosk-mode", "tracking-token"],
    "database": "postgresql-managed",
}


def write_blueprint(intake: dict) -> Path:
    temporary = tempfile.TemporaryDirectory()
    output = Path(temporary.name) / intake["name"]
    write_project(intake, output)
    # Keep the directory alive via the test case cleanup.
    return output, temporary


class ScaffoldFlowTests(unittest.TestCase):
    def assert_clean_doctor(self, output: Path) -> dict:
        manifest, errors, warnings = inspect_project(output)
        self.assertEqual(errors, [], f"doctor errors: {errors}")
        return manifest

    def assert_base_files(self, output: Path) -> None:
        for relative in (
            "GENTLE.md",
            "AGENTS.md",
            "ARCHITECTURE.md",
            ".engineering/project.json",
            ".engineering/gentle-handoff.json",
        ):
            self.assertTrue((output / relative).is_file(), f"missing {relative}")

    def test_quejas_scaffold_matches_expectations(self) -> None:
        output, _temporary = write_blueprint(QUEJAS_INTAKE)
        self.addCleanup(_temporary.cleanup)
        self.assert_base_files(output)
        manifest = self.assert_clean_doctor(output)

        self.assertEqual(manifest["recipe"]["id"], "GP-06")
        self.assertEqual(
            [(item["id"], item.get("role"), item.get("destination")) for item in manifest["starters"]],
            [
                ("hono-api", "primary", "services/api"),
                ("tanstack-admin", "primary", "apps/admin"),
                ("tanstack-transactional-pwa", "surface", "apps/intake"),
            ],
        )
        self.assertEqual(
            [(item["id"], item["provider"]) for item in manifest.get("surfaces", [])],
            [("public-intake", "tanstack-transactional-pwa")],
        )

        handoff = loads((output / ".engineering/gentle-handoff.json").read_text(encoding="utf-8"))
        self.assertEqual(handoff["strategy"]["owner"], "gentle-ai")
        self.assertIn("apps/intake/AGENTS.md", handoff["composition"]["starter_docs"]["tanstack-transactional-pwa"])
        self.assertIn("services/api/AGENTS.md", handoff["composition"]["starter_docs"]["hono-api"])
        self.assertEqual(handoff["composition"]["primary_recipe"], "GP-06")
        self.assertEqual(handoff["composition"]["future_surfaces"], ["mobile", "desktop"])

        gentle = (output / "GENTLE.md").read_text(encoding="utf-8")
        self.assertIn("GP-06", gentle)
        self.assertIn("apps/intake", gentle)
        self.assertIn("## Superficies futuras", gentle)
        self.assertIn("`mobile`", gentle)
        self.assertIn("## Documentación por starter", gentle)
        self.assertIn("apps/intake/docs/API-CONTRACT.md", gentle)
        self.assertIn("destino por defecto: `services/api`", gentle)
        self.assertIn("Aceptación por capability", gentle)

    def test_admin_scaffold_stays_minimal(self) -> None:
        output, _temporary = write_blueprint(ADMIN_INTAKE)
        self.addCleanup(_temporary.cleanup)
        self.assert_base_files(output)
        manifest = self.assert_clean_doctor(output)

        self.assertEqual(manifest["recipe"]["id"], "GP-02")
        self.assertEqual(
            [item["id"] for item in manifest["starters"]], ["hono-api", "tanstack-admin"]
        )
        self.assertEqual(manifest.get("surfaces", []), [])

        handoff = loads((output / ".engineering/gentle-handoff.json").read_text(encoding="utf-8"))
        self.assertEqual(handoff["composition"]["future_surfaces"], [])
        gentle = (output / "GENTLE.md").read_text(encoding="utf-8")
        self.assertIn("Ninguna: la Recipe no permite componer más surfaces.", gentle)

    def test_signals_only_scaffold_infers_portal(self) -> None:
        output, _temporary = write_blueprint(SIGNALS_ONLY_INTAKE)
        self.addCleanup(_temporary.cleanup)
        self.assert_base_files(output)
        manifest = self.assert_clean_doctor(output)

        self.assertEqual(manifest["recipe"]["id"], "GP-06")
        self.assertIn(
            "tanstack-transactional-pwa", [item["id"] for item in manifest["starters"]]
        )
        gentle = (output / "GENTLE.md").read_text(encoding="utf-8")
        self.assertIn("apps/intake", gentle)


if __name__ == "__main__":
    unittest.main()
