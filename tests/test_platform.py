from __future__ import annotations

import tempfile
import unittest
from pathlib import Path

from scripts.eng import (
    PlatformError,
    add_feature_to_project,
    change_plan,
    evaluate_boilerplate,
    inspect_project,
    normalize_repository,
    resolve_recipe,
    write_project,
)
from scripts.validate_platform import repository_markdown_files


class BoilerplateCuratorTests(unittest.TestCase):
    def test_normalizes_github_url(self) -> None:
        self.assertEqual(
            normalize_repository("github.com/KRIASOFT/react-starter-kit.git/"),
            "https://github.com/kriasoft/react-starter-kit",
        )

    def test_rejects_exact_duplicate(self) -> None:
        result = evaluate_boilerplate("https://github.com/kriasoft/react-starter-kit")
        self.assertEqual(result["decision"], "ALREADY_REGISTERED")
        self.assertEqual(result["entry_id"], "react-starter-kit")

    def test_marks_new_upstream_commit_as_refresh(self) -> None:
        result = evaluate_boilerplate(
            "https://github.com/kriasoft/react-starter-kit",
            observed_commit="future-commit",
        )
        self.assertEqual(result["decision"], "ALREADY_REGISTERED_REFRESH")
        self.assertIn("merge-seed", result["next_action"])

    def test_unknown_repository_is_candidate_not_automatic_default(self) -> None:
        result = evaluate_boilerplate("https://github.com/example/new-admin", category="admin-web")
        self.assertEqual(result["decision"], "ADD_AS_CANDIDATE")
        self.assertIn("tanstack-admin", result["compare_with"])


class RecipeResolverTests(unittest.TestCase):
    def school_intake(self) -> dict:
        return {
            "name": "school-requests",
            "project_type": "institutional-admin",
            "signals": ["workflow", "requests"],
            "features": ["files"],
            "excluded_features": ["multitenancy", "jobs"],
            "database": "postgresql-managed",
        }

    def test_resolves_school_recipe_and_dependencies(self) -> None:
        result = resolve_recipe(self.school_intake())
        self.assertEqual(result["recipe"]["id"], "GP-02")
        self.assertEqual([item["id"] for item in result["starters"]], ["tanstack-admin", "hono-api"])
        self.assertEqual(result["database"], "postgresql-managed")
        self.assertIn("auth", result["features"])
        self.assertIn("files", result["features"])
        self.assertEqual(result["scaffold_status"], "blueprint")

    def test_turso_is_explicit_and_warned(self) -> None:
        intake = self.school_intake()
        intake["database"] = "turso-libsql"
        result = resolve_recipe(intake)
        self.assertTrue(any("trial" in item for item in result["warnings"]))

    def test_rejects_database_outside_recipe(self) -> None:
        intake = self.school_intake()
        intake["database"] = "sqlite-local"
        with self.assertRaises(PlatformError):
            resolve_recipe(intake)

    def test_rejects_feature_conflict(self) -> None:
        intake = self.school_intake()
        intake["features"] = ["jobs"]
        with self.assertRaises(PlatformError):
            resolve_recipe(intake)

    def test_writes_and_doctors_blueprint(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            output = Path(temporary) / "project"
            result = write_project(self.school_intake(), output)
            self.assertTrue((output / ".engineering/project.json").exists())
            self.assertTrue((output / "ARCHITECTURE.md").exists())
            manifest, errors, warnings = inspect_project(output)
            self.assertEqual(manifest["project"]["name"], "school-requests")
            self.assertEqual(errors, [])
            self.assertTrue(warnings)
            self.assertEqual(result["scaffold_status"], "blueprint")

    def test_does_not_overwrite_nonempty_directory(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            output = Path(temporary)
            (output / "keep.txt").write_text("user data", encoding="utf-8")
            with self.assertRaises(PlatformError):
                write_project(self.school_intake(), output)
            self.assertEqual((output / "keep.txt").read_text(encoding="utf-8"), "user data")

    def test_change_plan_selects_risk_gates(self) -> None:
        manifest = resolve_recipe(self.school_intake())
        plan = change_plan(manifest, "permission")
        self.assertIn("authorization", plan["skills"])
        self.assertIn("security", plan["gates"])

    def test_add_feature_plans_dependencies_without_mutating_by_default(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            output = Path(temporary) / "project"
            intake = self.school_intake()
            intake["excluded_features"] = ["multitenancy"]
            write_project(intake, output)
            before = (output / ".engineering/project.json").read_text(encoding="utf-8")
            result = add_feature_to_project(output, "webhooks")
            after = (output / ".engineering/project.json").read_text(encoding="utf-8")
            self.assertTrue(result["changed"])
            self.assertIn("webhooks", result["added_with_dependencies"])
            self.assertEqual(before, after)

    def test_add_feature_respects_explicit_exclusion(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            output = Path(temporary) / "project"
            write_project(self.school_intake(), output)
            with self.assertRaises(PlatformError):
                add_feature_to_project(output, "jobs", apply=True)


class PlatformValidatorTests(unittest.TestCase):
    def test_includes_repository_markdown_and_skips_nested_dependency_markdown(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            repository_doc = root / "docs/repository.md"
            skill_doc = root / ".opencode/skills/example/SKILL.md"
            dependency_doc = root / ".opencode/node_modules/example/README.md"
            for path in (repository_doc, skill_doc, dependency_doc):
                path.parent.mkdir(parents=True, exist_ok=True)
                path.write_text("# Documentation\n", encoding="utf-8")

            files = {path.relative_to(root) for path in repository_markdown_files(root)}

            self.assertIn(repository_doc.relative_to(root), files)
            self.assertIn(skill_doc.relative_to(root), files)
            self.assertNotIn(dependency_doc.relative_to(root), files)


if __name__ == "__main__":
    unittest.main()
