from __future__ import annotations

import shutil
import subprocess
import tempfile
import unittest
from argparse import Namespace
from json import dumps, loads
from pathlib import Path
from unittest.mock import patch

from scripts.eng import (
    PLATFORM_VERSION,
    PlatformError,
    add_feature_to_project,
    change_plan,
    command_install,
    command_check,
    command_start,
    command_uninstall,
    evaluate_boilerplate,
    inspect_project,
    normalize_repository,
    resolve_recipe,
    validate_project_definition,
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
            self.assertTrue((output / "GENTLE.md").exists())
            manifest, errors, warnings = inspect_project(output)
            self.assertEqual(manifest["project"]["name"], "school-requests")
            self.assertEqual(errors, [])
            self.assertEqual(warnings, [])
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


class PiWorkflowTests(unittest.TestCase):
    def definition(self) -> dict:
        return loads(
            (Path(__file__).parents[1] / "examples/project-definitions/school-requests.json").read_text(
                encoding="utf-8"
            )
        )

    def test_source_definition_matches_canonical_example(self) -> None:
        root = Path(__file__).resolve().parents[1]
        paths = (
            root / "examples/project-definitions/school-requests.json",
            root / "examples/school-requests/.engineering/project-definition.json",
        )

        def normalized(path: Path) -> dict:
            definition = loads(path.read_text(encoding="utf-8"))
            definition["$schema"] = (
                (path.parent / definition["$schema"]).resolve().relative_to(root).as_posix()
            )
            return definition

        self.assertEqual(
            normalized(paths[0]),
            normalized(paths[1]),
            "canonical source and generated project definitions drifted",
        )

    def test_pi_package_declares_native_resources(self) -> None:
        root = Path(__file__).parents[1]
        package = loads((root / "package.json").read_text(encoding="utf-8"))
        self.assertEqual(package["version"], PLATFORM_VERSION)
        self.assertEqual(package["pi"]["extensions"], ["./extensions/engineering-platform.ts"])
        self.assertIn("./.opencode/skills", package["pi"]["skills"])
        self.assertTrue((root / "pi-skills/project-discovery/SKILL.md").exists())

    def test_bootstrap_accepts_gentle_metadata_in_target(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            output = Path(temporary) / "school-requests"
            (output / ".atl/state/nested.json").parent.mkdir(parents=True)
            (output / ".atl/state/nested.json").write_text("{}", encoding="utf-8")
            (output / ".gitignore").write_text(".env\n", encoding="utf-8")
            result = write_project(self.definition()["intake"], output)
            self.assertEqual(result["scaffold_status"], "blueprint")

    def test_bootstrap_accepts_only_definition_in_target(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            output = Path(temporary) / "school-requests"
            definition_path = output / ".engineering/project-definition.json"
            definition_path.parent.mkdir(parents=True)
            definition_path.write_text(
                __import__("json").dumps(self.definition(), ensure_ascii=False), encoding="utf-8"
            )
            result = write_project(
                self.definition()["intake"],
                output,
                definition=self.definition(),
                allowed_existing={definition_path},
            )
            self.assertEqual(result["definition_status"], "confirmed")
            self.assertIn("Sistema interno", (output / "GENTLE.md").read_text(encoding="utf-8"))
            handoff = loads(
                (output / ".engineering/gentle-handoff.json").read_text(encoding="utf-8")
            )
            self.assertEqual(handoff["strategy"]["owner"], "gentle-ai")
            self.assertEqual(handoff["strategy"]["allowed"], ["direct", "sdd"])

    def test_rejects_unconfirmed_definition(self) -> None:
        definition = self.definition()
        definition["discovery"]["status"] = "draft"
        with self.assertRaises(PlatformError):
            validate_project_definition(definition)

    def test_feature_update_preserves_definition_and_handoff(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            output = Path(temporary) / "school-requests"
            definition = self.definition()
            write_project(
                definition["intake"], output, definition=definition
            )
            add_feature_to_project(output, "api-keys", apply=True)
            manifest = loads(
                (output / ".engineering/project.json").read_text(encoding="utf-8")
            )
            updated_definition = loads(
                (output / ".engineering/project-definition.json").read_text(encoding="utf-8")
            )
            self.assertEqual(manifest["definition_status"], "confirmed")
            self.assertIn("api-keys", updated_definition["intake"]["features"])
            self.assertIn("Sistema interno", (output / "GENTLE.md").read_text(encoding="utf-8"))

    def test_start_dry_run_accepts_gentle_metadata_in_target(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            workspace = Path(temporary)
            target = workspace / "new-product"
            (target / ".atl/state/nested.json").parent.mkdir(parents=True)
            (target / ".atl/state/nested.json").write_text("{}", encoding="utf-8")
            (target / ".gitignore").write_text(".env\n", encoding="utf-8")
            with patch("builtins.print") as output:
                command_start(
                    Namespace(name="new-product", workspace=temporary, dry_run=True)
                )
            payload = loads(output.call_args.args[0])
            self.assertEqual(Path(payload["target"]), target)

    def test_start_rejects_user_content_in_target(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            target = Path(temporary) / "new-product"
            target.mkdir()
            (target / "notes.txt").write_text("user data", encoding="utf-8")
            with self.assertRaises(PlatformError):
                command_start(
                    Namespace(name="new-product", workspace=temporary, dry_run=True)
                )

    def test_start_accepts_existing_empty_target(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            target = Path(temporary) / "new-product"
            target.mkdir()
            with patch("builtins.print") as output:
                command_start(
                    Namespace(name="new-product", workspace=temporary, dry_run=True)
                )
            payload = loads(output.call_args.args[0])
            self.assertEqual(Path(payload["target"]), target)

    def test_start_dry_run_uses_workspace_child(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            with patch("builtins.print") as output:
                command_start(
                    Namespace(name="new-product", workspace=temporary, dry_run=True)
                )
            payload = loads(output.call_args.args[0])
            self.assertEqual(Path(payload["target"]).parent, Path(temporary))
            self.assertEqual(payload["command"][0], "pi")

    def test_start_rejects_symlink_target(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            workspace = Path(temporary) / "dev"
            elsewhere = Path(temporary) / "elsewhere"
            workspace.mkdir()
            elsewhere.mkdir()
            (workspace / "new-product").symlink_to(elsewhere, target_is_directory=True)
            with self.assertRaises(PlatformError):
                command_start(
                    Namespace(name="new-product", workspace=str(workspace), dry_run=True)
                )

    def test_canonical_example_has_complete_gentle_handoff(self) -> None:
        project = Path(__file__).parents[1] / "examples/school-requests"
        manifest, errors, warnings = inspect_project(project)
        self.assertEqual(errors, [])
        self.assertEqual(manifest["definition_status"], "confirmed")
        self.assertTrue((project / "GENTLE.md").exists())
        self.assertEqual(warnings, [])

    def test_internal_starter_materializes_real_code(self) -> None:
        intake = {
            "name": "assistant-app",
            "project_type": "ai-assistant",
            "signals": ["chat", "tools"],
            "features": [],
            "excluded_features": [],
            "database": "postgresql-managed",
        }
        with tempfile.TemporaryDirectory() as temporary:
            output = Path(temporary) / "assistant-app"
            manifest = write_project(
                intake, output, materialize=True, skip_setup=True, skip_checks=True
            )
            self.assertEqual(manifest["scaffold_status"], "materialized")
            self.assertEqual(manifest["readiness"], "code-ready")
            self.assertTrue((output / "src/app.ts").exists())
            self.assertTrue((output / ".engineering/materialization.json").exists())
            self.assertTrue((output / ".git").is_dir())

    def test_materialization_is_single_source_and_keeps_safe_templates(self) -> None:
        intake = {
            "name": "assistant-app",
            "project_type": "ai-assistant",
            "signals": ["chat", "tools"],
            "features": [],
            "excluded_features": [],
            "database": "postgresql-managed",
        }
        definition = self.definition()
        definition["intake"] = intake
        definition["idea"]["summary"] = "Asistente interno con controles y auditoría."
        with tempfile.TemporaryDirectory() as temporary:
            output = Path(temporary) / "assistant-app"
            write_project(
                intake,
                output,
                definition=definition,
                materialize=True,
                skip_setup=True,
                skip_checks=True,
            )
            manifest = loads((output / ".engineering/project.json").read_text(encoding="utf-8"))
            handoff = loads(
                (output / ".engineering/gentle-handoff.json").read_text(encoding="utf-8")
            )
            self.assertNotIn("materialization", manifest)
            self.assertEqual(
                manifest["$schema"],
                f"https://raw.githubusercontent.com/JhonMA82/engineering-platform/v{PLATFORM_VERSION}/schemas/project.schema.json",
            )
            self.assertEqual(handoff["schema_version"], 2)
            self.assertNotIn("stack", handoff)
            self.assertTrue((output / ".env.example").exists())
            self.assertEqual((output / ".git/HEAD").read_text(encoding="utf-8"), "ref: refs/heads/main\n")
            self.assertEqual(
                loads((output / ".engineering/project-definition.json").read_text(encoding="utf-8"))["$schema"],
                f"https://raw.githubusercontent.com/JhonMA82/engineering-platform/v{PLATFORM_VERSION}/schemas/project-definition.schema.json",
            )

    def test_partial_check_preserves_unselected_checks_and_reuses_setup(self) -> None:
        intake = {
            "name": "assistant-app",
            "project_type": "ai-assistant",
            "signals": ["chat", "tools"],
            "features": [],
            "excluded_features": [],
            "database": "postgresql-managed",
        }
        with tempfile.TemporaryDirectory() as temporary:
            output = Path(temporary) / "assistant-app"
            write_project(
                intake, output, materialize=True, skip_setup=True, skip_checks=True
            )
            materialization_path = output / ".engineering/materialization.json"
            materialization = loads(materialization_path.read_text(encoding="utf-8"))
            materialization["checks"].append(
                {
                    "starter": "assistant-app",
                    "gate": "build",
                    "command": ["npm", "run", "build"],
                    "workdir": ".",
                    "status": "skipped",
                }
            )
            materialization["readiness"] = "verified"
            materialization_path.write_text(dumps(materialization), encoding="utf-8")
            manifest = loads((output / ".engineering/project.json").read_text(encoding="utf-8"))
            manifest["readiness"] = "verified"
            (output / ".engineering/project.json").write_text(dumps(manifest), encoding="utf-8")
            calls: list[tuple[list[str], str | None]] = []

            def fake_run(command: list[str], _cwd: Path, *, gate: str | None = None) -> dict:
                calls.append((command, gate))
                record = {"command": command, "workdir": ".", "returncode": 0, "status": "passed"}
                if gate:
                    record["gate"] = gate
                return record

            args = Namespace(project=str(output), changed_files=["src/app.ts"], run=True)
            with patch("scripts.eng._run_record", side_effect=fake_run), patch("builtins.print"):
                self.assertEqual(command_check(args), 0)
            self.assertEqual([gate for _, gate in calls], [None, "typecheck", "test"])
            after = loads(materialization_path.read_text(encoding="utf-8"))
            self.assertEqual(len(after["checks"]), 3)
            self.assertEqual(after["checks"][-1]["gate"], "build")
            self.assertEqual(after["readiness"], "code-ready")

            calls.clear()
            with patch("scripts.eng._run_record", side_effect=fake_run), patch("builtins.print"):
                self.assertEqual(command_check(args), 0)
            self.assertEqual([gate for _, gate in calls], ["typecheck", "test"])

    def test_materialization_failure_leaves_target_unchanged(self) -> None:
        intake = self.definition()["intake"]
        with tempfile.TemporaryDirectory() as temporary:
            output = Path(temporary) / "school-requests"
            (output / ".atl").mkdir(parents=True)
            marker = output / ".atl/state.json"
            marker.write_text("{}", encoding="utf-8")
            with patch("scripts.eng.materialize_project", side_effect=PlatformError("boom")):
                with self.assertRaises(PlatformError):
                    write_project(intake, output, materialize=True)
            self.assertEqual(marker.read_text(encoding="utf-8"), "{}")
            self.assertFalse((output / ".engineering").exists())

    def test_pi_command_accepts_managed_metadata(self) -> None:
        extension = (Path(__file__).parents[1] / "extensions/engineering-platform.ts").read_text(
            encoding="utf-8"
        )
        self.assertIn('[".git", ".atl", ".gitignore"]', extension)

    def test_global_install_can_be_verified_without_touching_user_home(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            completed = Namespace(returncode=0, stdout="engineering-platform", stderr="")
            with patch("scripts.eng.shutil.which", return_value="/fake/pi"):
                with patch("scripts.eng.subprocess.run", return_value=completed):
                    with patch("builtins.print") as output:
                        command_install(
                            Namespace(
                                target="pi",
                                home=temporary,
                                force=False,
                                dry_run=False,
                                global_install=True,
                            )
                        )
            status = loads(output.call_args.args[0])
            self.assertTrue(status["ok"])
            home = Path(temporary)
            self.assertTrue((home / ".local/bin/eng").is_symlink())
            self.assertTrue(
                (home / ".local/share/engineering-platform" / PLATFORM_VERSION / "package.json").exists()
            )
            self.assertTrue(
                (
                    home
                    / ".local/share/engineering-platform"
                    / PLATFORM_VERSION
                    / "starters/ai-assistant/.env.example"
                ).exists()
            )
            with patch("scripts.eng.shutil.which", return_value="/fake/pi"):
                with patch("scripts.eng.subprocess.run", return_value=completed):
                    with patch("builtins.print"):
                        command_uninstall(
                            Namespace(
                                target="pi",
                                home=temporary,
                                dry_run=False,
                                global_install=True,
                            )
                        )
            self.assertFalse((home / ".local/bin/eng").exists())
            self.assertFalse((home / ".local/share/engineering-platform" / PLATFORM_VERSION).exists())

    def test_global_install_migrates_managed_launcher(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            home = Path(temporary)
            previous_install = home / ".local/share/engineering-platform/previous"
            previous_install.mkdir(parents=True)
            (previous_install / "eng").write_text("#!/bin/sh\n", encoding="utf-8")
            launcher = home / ".local/bin/eng"
            launcher.parent.mkdir(parents=True)
            launcher.symlink_to(previous_install / "eng")
            completed = Namespace(returncode=0, stdout="engineering-platform", stderr="")
            with patch("scripts.eng.shutil.which", return_value="/fake/pi"):
                with patch("scripts.eng.subprocess.run", return_value=completed):
                    command_install(
                        Namespace(
                            target="pi",
                            home=temporary,
                            force=False,
                            dry_run=False,
                            global_install=True,
                        )
                    )
            expected = home / ".local/share/engineering-platform" / PLATFORM_VERSION / "eng"
            self.assertEqual(launcher.resolve(), expected.resolve())
            self.assertTrue(previous_install.exists())

    def test_global_install_retires_conflicting_git_and_stale_managed_packages(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            home = Path(temporary)
            stale_install = home / ".local/share/engineering-platform/0.5.0"
            stale_install.mkdir(parents=True)
            (stale_install / "package.json").write_text(
                dumps({"name": "engineering-platform", "version": "0.5.0"}),
                encoding="utf-8",
            )
            direct_package = "git:github.com/JhonMA82/engineering-platform@v0.5.1"
            unrelated_package = "npm:unrelated-package"
            project_package = "git:github.com/JhonMA82/engineering-platform@v0.4.0"
            registered = [str(stale_install), direct_package, unrelated_package]
            commands: list[list[str]] = []

            def package_list() -> str:
                entries = [
                    f"  {source}\n    /fake/path/{index}"
                    for index, source in enumerate(registered)
                ]
                return (
                    "User packages:\n"
                    + "\n".join(entries)
                    + f"\nProject packages:\n  {project_package}\n    /project/package"
                )

            def fake_pi(command: list[str], **_: object) -> Namespace:
                commands.append(command)
                if command[1] == "install":
                    registered.append(command[2])
                    return Namespace(returncode=0, stdout="installed", stderr="")
                if command[1] == "remove":
                    registered.remove(command[2])
                    return Namespace(returncode=0, stdout="removed", stderr="")
                return Namespace(returncode=0, stdout=package_list(), stderr="")

            with patch("scripts.eng.shutil.which", return_value="/fake/pi"):
                with patch("scripts.eng.subprocess.run", side_effect=fake_pi):
                    with patch("builtins.print"):
                        command_install(
                            Namespace(
                                target="pi",
                                home=temporary,
                                force=False,
                                dry_run=False,
                                global_install=True,
                            )
                        )

            current_install = home / ".local/share/engineering-platform" / PLATFORM_VERSION
            remove_commands = [command for command in commands if command[1] == "remove"]
            self.assertEqual(
                remove_commands,
                [
                    ["/fake/pi", "remove", direct_package],
                    ["/fake/pi", "remove", str(stale_install)],
                ],
            )
            self.assertLess(
                commands.index(["/fake/pi", "install", str(current_install)]),
                commands.index(remove_commands[0]),
            )
            self.assertIn(["/fake/pi", "list", "--no-approve"], commands)
            self.assertIn(str(current_install), registered)
            self.assertIn(unrelated_package, registered)
            self.assertNotIn(project_package, [command[2] for command in remove_commands])
            self.assertNotIn(direct_package, registered)
            self.assertNotIn(str(stale_install), registered)
            self.assertFalse(stale_install.exists())

    def test_global_install_removes_unregistered_stale_managed_directory(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            home = Path(temporary)
            stale_install = home / ".local/share/engineering-platform/0.4.2"
            stale_install.mkdir(parents=True)
            (stale_install / "package.json").write_text(
                dumps({"name": "engineering-platform", "version": "0.4.2"}),
                encoding="utf-8",
            )
            registered: list[str] = []
            commands: list[list[str]] = []

            def fake_pi(command: list[str], **_: object) -> Namespace:
                commands.append(command)
                if command[1] == "install":
                    registered.append(command[2])
                    return Namespace(returncode=0, stdout="installed", stderr="")
                if command[1] == "remove":
                    self.assertTrue(stale_install.exists())
                    return Namespace(
                        returncode=1,
                        stdout="",
                        stderr=f"No matching package found for {command[2]}",
                    )
                return Namespace(returncode=0, stdout="\n".join(registered), stderr="")

            with patch("scripts.eng.shutil.which", return_value="/fake/pi"):
                with patch("scripts.eng.subprocess.run", side_effect=fake_pi):
                    with patch("builtins.print") as output:
                        result = command_install(
                            Namespace(
                                target="pi",
                                home=temporary,
                                force=False,
                                dry_run=False,
                                global_install=True,
                            )
                        )

            status = loads(output.call_args.args[0])
            self.assertEqual(result, 0)
            self.assertTrue(status["ok"])
            self.assertEqual(
                [command[1] for command in commands], ["install", "list", "remove", "list"]
            )
            self.assertFalse(stale_install.exists())

    def test_global_install_preserves_stale_installation_when_pi_removal_fails(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            home = Path(temporary)
            stale_install = home / ".local/share/engineering-platform/0.5.0"
            stale_install.mkdir(parents=True)
            (stale_install / "package.json").write_text(
                dumps({"name": "engineering-platform", "version": "0.5.0"}),
                encoding="utf-8",
            )
            commands: list[list[str]] = []

            def failing_pi(command: list[str], **_: object) -> Namespace:
                commands.append(command)
                if command[1] == "remove":
                    return Namespace(returncode=1, stdout="", stderr="x" * 5000)
                return Namespace(returncode=0, stdout="installed", stderr="")

            with patch("scripts.eng.shutil.which", return_value="/fake/pi"):
                with patch("scripts.eng.subprocess.run", side_effect=failing_pi):
                    with self.assertRaises(PlatformError) as raised:
                        command_install(
                            Namespace(
                                target="pi",
                                home=temporary,
                                force=False,
                                dry_run=False,
                                global_install=True,
                            )
                        )

            current_install = home / ".local/share/engineering-platform" / PLATFORM_VERSION
            launcher = home / ".local/bin/eng"
            self.assertLessEqual(len(str(raised.exception)), 1200)
            self.assertTrue(stale_install.exists())
            self.assertTrue(current_install.exists())
            self.assertEqual(launcher.resolve(), (current_install / "eng").resolve())
            self.assertEqual([command[1] for command in commands], ["install", "list", "remove"])

    def test_global_install_preserves_stale_installation_when_git_removal_fails(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            home = Path(temporary)
            stale_install = home / ".local/share/engineering-platform/0.5.0"
            stale_install.mkdir(parents=True)
            (stale_install / "package.json").write_text(
                dumps({"name": "engineering-platform", "version": "0.5.0"}),
                encoding="utf-8",
            )
            direct_package = "git:github.com/JhonMA82/engineering-platform@v0.5.1"
            commands: list[list[str]] = []

            def failing_pi(command: list[str], **_: object) -> Namespace:
                commands.append(command)
                if command[1] == "install":
                    return Namespace(returncode=0, stdout="installed", stderr="")
                if command[1] == "list":
                    return Namespace(
                        returncode=0,
                        stdout=(
                            "User packages:\n"
                            f"  {direct_package}\n"
                            "    /fake/git/github.com/JhonMA82/engineering-platform"
                        ),
                        stderr="",
                    )
                return Namespace(returncode=1, stdout="", stderr="permission denied")

            with patch("scripts.eng.shutil.which", return_value="/fake/pi"):
                with patch("scripts.eng.subprocess.run", side_effect=failing_pi):
                    with self.assertRaises(PlatformError) as raised:
                        command_install(
                            Namespace(
                                target="pi",
                                home=temporary,
                                force=False,
                                dry_run=False,
                                global_install=True,
                            )
                        )

            current_install = home / ".local/share/engineering-platform" / PLATFORM_VERSION
            self.assertIn(direct_package, str(raised.exception))
            self.assertTrue(stale_install.exists())
            self.assertTrue(current_install.exists())
            self.assertEqual([command[1] for command in commands], ["install", "list", "remove"])

    def test_global_install_preserves_stale_installation_when_pi_listing_fails(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            home = Path(temporary)
            stale_install = home / ".local/share/engineering-platform/0.5.0"
            stale_install.mkdir(parents=True)
            (stale_install / "package.json").write_text(
                dumps({"name": "engineering-platform", "version": "0.5.0"}),
                encoding="utf-8",
            )
            commands: list[list[str]] = []

            def failing_pi(command: list[str], **_: object) -> Namespace:
                commands.append(command)
                if command[1] == "install":
                    return Namespace(returncode=0, stdout="installed", stderr="")
                return Namespace(returncode=1, stdout="", stderr="x" * 5000)

            with patch("scripts.eng.shutil.which", return_value="/fake/pi"):
                with patch("scripts.eng.subprocess.run", side_effect=failing_pi):
                    with self.assertRaises(PlatformError) as raised:
                        command_install(
                            Namespace(
                                target="pi",
                                home=temporary,
                                force=False,
                                dry_run=False,
                                global_install=True,
                            )
                        )

            current_install = home / ".local/share/engineering-platform" / PLATFORM_VERSION
            self.assertLessEqual(len(str(raised.exception)), 1200)
            self.assertTrue(stale_install.exists())
            self.assertTrue(current_install.exists())
            self.assertEqual([command[1] for command in commands], ["install", "list"])

    def test_global_install_preserves_unmanaged_launcher(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            home = Path(temporary)
            external_launcher = home / "custom/eng"
            external_launcher.parent.mkdir(parents=True)
            external_launcher.write_text("#!/bin/sh\n", encoding="utf-8")
            launcher = home / ".local/bin/eng"
            launcher.parent.mkdir(parents=True)
            launcher.symlink_to(external_launcher)
            with patch("scripts.eng.shutil.which", return_value="/fake/pi"):
                with self.assertRaises(PlatformError):
                    command_install(
                        Namespace(
                            target="pi",
                            home=temporary,
                            force=False,
                            dry_run=False,
                            global_install=True,
                        )
                    )
            self.assertEqual(launcher.resolve(), external_launcher.resolve())

    def test_global_uninstall_removes_all_managed_installations(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            home = Path(temporary)
            managed_root = home / ".local/share/engineering-platform"
            installations = {
                "0.4.2": managed_root / "0.4.2",
                PLATFORM_VERSION: managed_root / PLATFORM_VERSION,
            }
            for version, installation in installations.items():
                installation.mkdir(parents=True)
                (installation / "eng").write_text("#!/bin/sh\n", encoding="utf-8")
                (installation / "package.json").write_text(
                    dumps({"name": "engineering-platform", "version": version}),
                    encoding="utf-8",
                )
            launcher = home / ".local/bin/eng"
            launcher.parent.mkdir(parents=True)
            launcher.symlink_to(installations["0.4.2"] / "eng")
            completed = Namespace(returncode=0, stdout="removed", stderr="")
            with patch("scripts.eng.shutil.which", return_value="/fake/pi"):
                with patch("scripts.eng.subprocess.run", return_value=completed) as run:
                    with patch("builtins.print") as output:
                        command_uninstall(
                            Namespace(
                                target="pi",
                                home=temporary,
                                dry_run=False,
                                global_install=True,
                            )
                        )
            status = loads(output.call_args.args[0])
            self.assertTrue(status["removed"])
            self.assertTrue(status["launcher_removed"])
            self.assertEqual(run.call_count, 2)
            self.assertFalse(launcher.exists())
            self.assertFalse(managed_root.exists())
            self.assertTrue(all(not installation.exists() for installation in installations.values()))

    def test_global_uninstall_preserves_unmanaged_launcher(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            home = Path(temporary)
            installation = home / ".local/share/engineering-platform" / PLATFORM_VERSION
            installation.mkdir(parents=True)
            (installation / "eng").write_text("#!/bin/sh\n", encoding="utf-8")
            (installation / "package.json").write_text(
                dumps({"name": "engineering-platform", "version": PLATFORM_VERSION}),
                encoding="utf-8",
            )
            external_launcher = home / "custom/eng"
            external_launcher.parent.mkdir(parents=True)
            external_launcher.write_text("#!/bin/sh\n", encoding="utf-8")
            launcher = home / ".local/bin/eng"
            launcher.parent.mkdir(parents=True)
            launcher.symlink_to(external_launcher)
            completed = Namespace(returncode=0, stdout="removed", stderr="")
            with patch("scripts.eng.shutil.which", return_value="/fake/pi"):
                with patch("scripts.eng.subprocess.run", return_value=completed):
                    with patch("builtins.print"):
                        command_uninstall(
                            Namespace(
                                target="pi",
                                home=temporary,
                                dry_run=False,
                                global_install=True,
                            )
                        )
            self.assertFalse(installation.exists())
            self.assertTrue(launcher.is_symlink())
            self.assertEqual(launcher.resolve(), external_launcher.resolve())

    def test_global_uninstall_removes_stale_managed_launcher_without_current_install(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            home = Path(temporary)
            previous_install = home / ".local/share/engineering-platform/0.4.2"
            previous_install.mkdir(parents=True)
            (previous_install / "eng").write_text("#!/bin/sh\n", encoding="utf-8")
            (previous_install / "package.json").write_text(
                dumps({"name": "engineering-platform", "version": "0.4.2"}),
                encoding="utf-8",
            )
            launcher = home / ".local/bin/eng"
            launcher.parent.mkdir(parents=True)
            launcher.symlink_to(previous_install / "eng")
            completed = Namespace(returncode=0, stdout="removed", stderr="")
            with patch("scripts.eng.shutil.which", return_value="/fake/pi"):
                with patch("scripts.eng.subprocess.run", return_value=completed):
                    with patch("builtins.print") as output:
                        command_uninstall(
                            Namespace(
                                target="pi",
                                home=temporary,
                                dry_run=False,
                                global_install=True,
                            )
                        )
            status = loads(output.call_args.args[0])
            self.assertTrue(status["removed"])
            self.assertFalse(launcher.exists())
            self.assertFalse(previous_install.exists())


class LauncherTests(unittest.TestCase):
    def test_symlinked_launcher_locates_scripts_eng(self) -> None:
        repository = Path(__file__).resolve().parents[1]
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            package = root / "package"
            launcher = package / "eng"
            launcher.parent.mkdir(parents=True)
            shutil.copy2(repository / "eng", launcher)
            shutil.copy2(repository / "package.json", package / "package.json")
            shutil.copytree(repository / "scripts", package / "scripts")

            link = root / "bin/eng"
            link.parent.mkdir()
            link.symlink_to(launcher)
            completed = subprocess.run(
                [str(link), "--version"],
                capture_output=True,
                text=True,
                check=False,
            )

            self.assertEqual(completed.returncode, 0, completed.stderr)
            self.assertEqual(completed.stdout.strip(), f"eng {PLATFORM_VERSION}")


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
