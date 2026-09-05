"""Intent flow test: user idea in different shapes -> expected platform result.

Each scenario mirrors a real discovery outcome (complete, signals-only,
wrong project_type, invented vocabulary, future client, admin-only) and
asserts the end-to-end resolution: recipe, starters, surfaces, warnings,
or the exact actionable correction. Run with `make test`.
"""

from __future__ import annotations

import unittest

from scripts.eng import (
    PlatformError,
    evolution_hints,
    resolve_recipe,
    suggest_intake_correction,
)


def starters_of(manifest: dict) -> list[str]:
    return [item["id"] for item in manifest["starters"]]


def warnings_of(manifest: dict) -> list[str]:
    return list(manifest.get("warnings", []))


class IntentFlowTests(unittest.TestCase):
    def test_complete_idea_resolves_full_stack(self) -> None:
        manifest = resolve_recipe(
            {
                "name": "quejas-ciudadanas",
                "project_type": "multi-app",
                "signals": ["shared-api", "public-intake", "kiosk-mode"],
                "database": "postgresql-managed",
                "surfaces": [
                    {
                        "id": "public-intake",
                        "capabilities": [
                            "form-capture",
                            "offline-outbox",
                            "tracking-token",
                            "kiosk-mode",
                        ],
                    }
                ],
            }
        )
        self.assertEqual(manifest["recipe"]["id"], "GP-06")
        self.assertEqual(
            starters_of(manifest),
            ["hono-api", "tanstack-admin", "tanstack-transactional-pwa"],
        )
        self.assertEqual(
            [(item["id"], item["provider"], item["destination"]) for item in manifest["surfaces"]],
            [("public-intake", "tanstack-transactional-pwa", "apps/intake")],
        )

    def test_signals_only_idea_still_brings_portal(self) -> None:
        manifest = resolve_recipe(
            {
                "name": "quejas-senales",
                "project_type": "multi-app",
                "signals": ["shared-api", "kiosk-mode", "tracking-token"],
                "database": "postgresql-managed",
            }
        )
        self.assertEqual(manifest["recipe"]["id"], "GP-06")
        self.assertIn("tanstack-transactional-pwa", starters_of(manifest))
        self.assertTrue(
            any("inferidas desde señales" in warning for warning in warnings_of(manifest)),
            f"expected inference warning, got: {warnings_of(manifest)}",
        )

    def test_admin_requesting_portal_gets_actionable_correction(self) -> None:
        intake = {
            "name": "admin-portal",
            "project_type": "institutional-admin",
            "signals": ["workflow"],
            "database": "postgresql-managed",
            "surfaces": [{"id": "public-intake", "capabilities": ["form-capture"]}],
        }
        with self.assertRaises(PlatformError) as context:
            resolve_recipe(intake)
        self.assertIn("multi-app", str(context.exception))
        corrected = suggest_intake_correction(intake, str(context.exception))
        self.assertIsNotNone(corrected)
        assert corrected is not None
        manifest = resolve_recipe(corrected)
        self.assertEqual(manifest["recipe"]["id"], "GP-06")
        self.assertIn("tanstack-transactional-pwa", starters_of(manifest))

    def test_admin_with_portal_signals_only_suggests_multi_app(self) -> None:
        with self.assertRaises(PlatformError) as context:
            resolve_recipe(
                {
                    "name": "admin-senales",
                    "project_type": "institutional-admin",
                    "signals": ["workflow", "kiosk-mode"],
                    "database": "postgresql-managed",
                }
            )
        self.assertIn("multi-app", str(context.exception))

    def test_invented_vocabulary_teaches_canonical_terms(self) -> None:
        with self.assertRaises(PlatformError) as context:
            resolve_recipe(
                {
                    "name": "vocabulario",
                    "project_type": "multi-app",
                    "signals": ["shared-api"],
                    "database": "postgresql-managed",
                    "surfaces": [{"id": "public-intake", "capabilities": ["qr-capture"]}],
                }
            )
        message = str(context.exception)
        self.assertIn("Capabilities no reconocidas", message)
        self.assertIn("qr-capture -> form-capture, tracking-token", message)

    def test_future_mobile_keeps_composable_recipe_without_surfaces(self) -> None:
        manifest = resolve_recipe(
            {
                "name": "movil-futuro",
                "project_type": "multi-app",
                "signals": ["shared-api", "multiple-clients"],
                "database": "postgresql-managed",
            }
        )
        self.assertEqual(manifest["recipe"]["id"], "GP-06")
        self.assertEqual(starters_of(manifest), ["hono-api", "tanstack-admin"])
        hints = evolution_hints(manifest)
        self.assertTrue(any("mobile" in hint for hint in hints), f"got: {hints}")

    def test_admin_only_stays_minimal(self) -> None:
        manifest = resolve_recipe(
            {
                "name": "admin-solo",
                "project_type": "institutional-admin",
                "signals": ["workflow", "requests"],
                "database": "postgresql-managed",
            }
        )
        self.assertEqual(manifest["recipe"]["id"], "GP-02")
        self.assertEqual(starters_of(manifest), ["hono-api", "tanstack-admin"])
        self.assertEqual(manifest["surfaces"], [])
        self.assertEqual(evolution_hints(manifest), [])


if __name__ == "__main__":
    unittest.main()
