from __future__ import annotations

from contextlib import redirect_stdout
from io import StringIO
import os
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
    _apply_overlay,
    _adapter_environment,
    _render_adapter_command,
    _run_record,
    _starter_capabilities,
    add_feature_to_project,
    adapter_preflight,
    boilerplate_references,
    change_plan,
    command_install,
    command_check,
    command_boilerplate_verify,
    command_start,
    command_uninstall,
    evaluate_boilerplate,
    extend_project,
    inspect_project,
    materialize_project,
    normalize_repository,
    project_ci_yaml,
    read_json,
    resolve_recipe,
    validate_project_definition,
    verify_boilerplate,
    write_project,
)
from scripts.validate_platform import adapter_environment_errors, repository_markdown_files


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

    def test_primary_boilerplates_have_pin_adapter_and_ai_evidence(self) -> None:
        for boilerplate_id in (
            "stardrive",
            "tanstack-admin",
            "hono-api",
            "ignite",
            "tauri-ui",
            "speedpy",
            "react-starter-kit",
            "ai-assistant-starter",
        ):
            with self.subTest(boilerplate_id=boilerplate_id):
                self.assertTrue(verify_boilerplate(boilerplate_id)["ok"])

    def test_boilerplate_verify_exit_status_matches_result_and_keeps_json(self) -> None:
        for result, expected_status in (
            ({"id": "fixture", "ok": False, "errors": ["missing"]}, 1),
            ({"id": "fixture", "ok": True, "errors": []}, 0),
        ):
            with self.subTest(expected_status=expected_status):
                output = StringIO()
                with patch("scripts.eng.verify_boilerplate", return_value=result), redirect_stdout(output):
                    status = command_boilerplate_verify(Namespace(id="fixture"))
                self.assertEqual(status, expected_status)
                self.assertEqual(loads(output.getvalue()), result)

    def test_tauri_generator_uses_pinned_source_cli_for_vite(self) -> None:
        root = Path(__file__).parents[1]
        adapter = loads(
            (root / "curation/tauri-ui/adapter.json").read_text(encoding="utf-8")
        )
        self.assertEqual(
            adapter["materializer"]["command"],
            [
                "node",
                "packages/create-tauri-ui/index.js",
                "{output}",
                "--template",
                "vite",
                "--yes",
                "--no-workflow",
            ],
        )
        self.assertEqual(
            adapter["materializer"]["source_patches"],
            ["curation/tauri-ui/patches/create-tauri-app-version.patch"],
        )

    def test_tauri_source_patch_pins_nested_generator_before_build(self) -> None:
        root = Path(__file__).parents[1]
        patch_relative = "curation/tauri-ui/patches/create-tauri-app-version.patch"
        patch_text = (root / patch_relative).read_text(encoding="utf-8")
        self.assertIn('-        "create-tauri-app",', patch_text)
        self.assertIn('+        "create-tauri-app@4.6.2",', patch_text)
        self.assertIn('        "vanilla-ts",', patch_text)

        manifest = {
            "project": {"name": "sample-tauri"},
            "starters": [
                {
                    "id": "tauri-ui",
                    "pin": "8eb86d894c19b6df04ff883ab28b412b1e5f23ea",
                    "adapter": "curation/tauri-ui/adapter.json",
                }
            ],
        }
        commands: list[list[str]] = []
        patched_source: list[str] = []

        def fake_checkout(_source: dict, destination: Path) -> None:
            destination.mkdir()
            source_file = destination / "packages/create-tauri-ui/src/scaffold.ts"
            source_file.parent.mkdir(parents=True)
            source_file.write_text(
                "\n" * 243
                + """export async function scaffoldTauri(options: ProjectOptions): Promise<TauriScaffoldResult> {
    await execSafe(
      "bunx",
      [
        "create-tauri-app",
        tempProjectName,
        "--template",
        "vanilla-ts",
        "--manager",
        "bun",
        "--identifier",
        options.identifier,
        "--yes",
      ],
    );
}
""",
                encoding="utf-8",
            )
            subprocess.run(
                ["git", "init", "--quiet"],
                cwd=destination,
                capture_output=True,
                check=True,
                text=True,
            )
            subprocess.run(
                ["git", "add", str(source_file.relative_to(destination))],
                cwd=destination,
                capture_output=True,
                check=True,
                text=True,
            )

        def fake_run(
            command: list[str],
            cwd: Path,
            *,
            gate: str | None = None,
            environment: dict[str, str] | None = None,
        ) -> dict:
            commands.append(command)
            if command[:2] == ["git", "apply"]:
                return _run_record(command, cwd, gate=gate, environment=environment)
            if command[0] == "node":
                patched_source.append(
                    (cwd / "packages/create-tauri-ui/src/scaffold.ts").read_text(
                        encoding="utf-8"
                    )
                )
                generated = Path(command[2])
                generated.mkdir(parents=True)
                (generated / "package.json").write_text(
                    dumps({"name": "sample-tauri"}), encoding="utf-8"
                )
            record = {"command": command, "workdir": ".", "returncode": 0, "status": "passed"}
            if gate:
                record["gate"] = gate
            return record

        with tempfile.TemporaryDirectory() as temporary:
            with patch("scripts.eng._checkout_git_source", side_effect=fake_checkout), patch(
                "scripts.eng._run_record", side_effect=fake_run
            ), patch("scripts.eng._assert_adapter_requirements", return_value=[]):
                result = materialize_project(
                    manifest,
                    Path(temporary),
                    skip_setup=True,
                    skip_checks=True,
                )

        self.assertEqual(commands[0][:3], ["git", "apply", "--check"])
        self.assertEqual(commands[1][:2], ["git", "apply"])
        self.assertEqual(commands[2][:2], ["bun", "install"])
        self.assertEqual(commands[3][:2], ["bun", "run"])
        self.assertEqual(commands[4][:2], ["node", "packages/create-tauri-ui/index.js"])
        self.assertEqual(result["starters"][0]["source_patches"], [patch_relative])
        self.assertEqual(result["starters"][0]["patches"], [])
        self.assertEqual(len(patched_source), 1)
        self.assertIn('"create-tauri-app@4.6.2"', patched_source[0])
        self.assertIn('"--template"', patched_source[0])
        self.assertIn('"vanilla-ts"', patched_source[0])

    def test_referenced_boilerplate_cannot_be_removed_silently(self) -> None:
        references = boilerplate_references("hono-api")
        self.assertIn("GP-02:default", references)
        self.assertIn("GP-04:default", references)
        self.assertIn("GP-06:default", references)


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
        self.assertEqual([item["id"] for item in result["starters"]], ["hono-api", "tanstack-admin"])
        self.assertEqual(result["database"], "postgresql-managed")
        self.assertIn("auth", result["features"])
        self.assertIn("files", result["features"])
        self.assertEqual(result["scaffold_status"], "blueprint")
        self.assertEqual(result["capability_status"]["auth"]["state"], "materialized")
        self.assertEqual(
            result["capability_status"]["files"]["state"],
            "pending-implementation",
        )

    def test_blueprint_template_does_not_claim_verified_capabilities(self) -> None:
        template = loads(
            (Path(__file__).parents[1] / "templates/project-manifest.json").read_text(
                encoding="utf-8"
            )
        )
        states = [status["state"] for status in template["capability_status"].values()]
        self.assertEqual(template["scaffold_status"], "blueprint")
        self.assertEqual(template["readiness"], "code-ready")
        self.assertNotIn("verified", states)
        self.assertTrue(all(state == "materialized" for state in states))

    def test_rejects_turso_when_default_api_only_supports_postgresql(self) -> None:
        intake = self.school_intake()
        intake["database"] = "turso-libsql"
        with self.assertRaises(PlatformError):
            resolve_recipe(intake)

    def test_api_starter_receives_only_semantically_compatible_features(self) -> None:
        result = resolve_recipe(self.school_intake())
        api = result["starters"][0]
        self.assertEqual(api["destination"], "services/api")
        self.assertEqual(
            api["generator_features"],
            ["persistence", "auth", "authorization", "audit", "observability"],
        )
        self.assertEqual(api["unmaterialized_features"], ["files"])
        self.assertTrue(any("files" in warning for warning in result["warnings"]))

    def test_api_starter_maps_tenant_scoped_features_when_tenancy_is_selected(self) -> None:
        adapter = loads(
            (Path(__file__).parents[1] / "curation/hono-api/adapter.json").read_text(
                encoding="utf-8"
            )
        )
        generated, provided, unresolved = _starter_capabilities(
            adapter,
            ["auth", "rbac", "multitenancy", "api-keys", "files", "webhooks"],
            "postgresql-managed",
        )
        self.assertEqual(
            generated,
            ["persistence", "auth", "authorization", "tenancy", "apiKeys", "files", "webhooks"],
        )
        self.assertEqual(
            provided,
            ["auth", "rbac", "multitenancy", "api-keys", "files", "webhooks"],
        )
        self.assertEqual(unresolved, [])

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
            self.assertIn("webhooks", result["requested_with_dependencies"])
            self.assertIn("webhooks", result["implementation_required"])
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
        self.assertTrue((root / "pi-skills/project-evolution/SKILL.md").exists())

    def test_runtime_preflight_rejects_incompatible_version(self) -> None:
        adapter = {
            "boilerplate_id": "sample",
            "requirements": [{"executable": "bun", "version_prefix": "1.4."}],
        }
        completed = Namespace(returncode=0, stdout="1.3.14\n", stderr="")
        with patch("scripts.eng.shutil.which", return_value="/fake/bun"), patch(
            "scripts.eng.subprocess.run", return_value=completed
        ):
            status = adapter_preflight(adapter)
        self.assertFalse(status[0]["ok"])
        self.assertEqual(status[0]["expected"], "1.4.")

    def test_run_record_merges_declared_environment_without_recording_ambient_values(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            with patch.dict(
                os.environ,
                {"PATH": "/ambient", "PRIVATE_TOKEN": "secret"},
                clear=True,
            ), patch("scripts.eng.shutil.which", return_value="/fake/tool"), patch(
                "scripts.eng.subprocess.run",
                return_value=Namespace(returncode=0, stdout="", stderr=""),
            ) as run:
                record = _run_record(
                    ["tool"],
                    Path(temporary),
                    environment={"SHARP_IGNORE_GLOBAL_LIBVIPS": "1"},
                )
                command_environment = run.call_args.kwargs["env"]
                self.assertEqual(command_environment["PATH"], "/ambient")
                self.assertEqual(command_environment["PRIVATE_TOKEN"], "secret")
                self.assertEqual(command_environment["SHARP_IGNORE_GLOBAL_LIBVIPS"], "1")
                self.assertEqual(
                    record["environment"], {"SHARP_IGNORE_GLOBAL_LIBVIPS": "1"}
                )
                self.assertNotIn("PRIVATE_TOKEN", record["environment"])

                run.reset_mock()
                record = _run_record(["tool"], Path(temporary))
                self.assertNotIn("env", run.call_args.kwargs)
                self.assertNotIn("environment", record)

    def test_runtime_rejects_malformed_adapter_environment(self) -> None:
        invalid_environments = (
            {"INVALID-NAME": "1"},
            {"VALID_NAME": "line\nbreak"},
            {"VALID_NAME": "nul\x00value"},
            {1: "1"},
            [],
        )
        for environment in invalid_environments:
            with self.subTest(environment=environment):
                with self.assertRaises(PlatformError):
                    _adapter_environment(
                        {
                            "boilerplate_id": "sample",
                            "materializer": {"environment": environment},
                        }
                    )

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
            gentle = (output / "GENTLE.md").read_text(encoding="utf-8")
            self.assertIn(".engineering/project-definition.json", gentle)
            self.assertNotIn("Sistema interno", gentle)
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
            gentle = (output / "GENTLE.md").read_text(encoding="utf-8")
            self.assertIn(".engineering/project-definition.json", gentle)
            self.assertNotIn("Sistema interno", gentle)

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
            self.assertTrue((output / ".github/workflows/engineering.yml").exists())
            self.assertTrue((output / ".git").is_dir())

    def test_doctor_keeps_previous_materialized_manifest_upgradeable(self) -> None:
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
            manifest_path = output / ".engineering/project.json"
            manifest = loads(manifest_path.read_text(encoding="utf-8"))
            manifest["platform_version"] = "0.7.0"
            manifest.pop("capability_status")
            manifest_path.write_text(dumps(manifest), encoding="utf-8")
            handoff_path = output / ".engineering/gentle-handoff.json"
            handoff = loads(handoff_path.read_text(encoding="utf-8"))
            handoff["platform_version"] = "0.7.0"
            handoff_path.write_text(dumps(handoff), encoding="utf-8")
            (output / ".github/workflows/engineering.yml").unlink()

            _, errors, warnings = inspect_project(output)

            self.assertEqual(errors, [])
            self.assertTrue(any("capability_status" in item for item in warnings))
            self.assertTrue(any("CI raíz" in item for item in warnings))

    def test_project_schema_keeps_capability_status_optional(self) -> None:
        schema = loads(
            (Path(__file__).parents[1] / "schemas/project.schema.json").read_text(
                encoding="utf-8"
            )
        )
        self.assertNotIn("capability_status", schema["required"])
        self.assertIn("capability_status", schema["properties"])

    def test_extend_plans_new_starter_without_mutating_project(self) -> None:
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
            before = (output / ".engineering/project.json").read_bytes()
            with patch("scripts.eng.adapter_preflight", return_value=[]):
                result = extend_project(output, "tauri-ui")
            self.assertEqual(result["destination"], "apps/desktop")
            self.assertFalse(result["applied"])
            self.assertEqual(before, (output / ".engineering/project.json").read_bytes())
            self.assertFalse((output / "apps/desktop").exists())

    def test_adapter_command_rendering_is_data_driven(self) -> None:
        command = _render_adapter_command(
            ["tool", "--out={output}", "--features={features_csv}", "{project_name}"],
            source=Path("/tmp/source"),
            output=Path("/tmp/output"),
            project_name="sample-project",
            generator_features=["auth", "audit"],
        )
        self.assertEqual(
            command,
            ["tool", "--out=/tmp/output", "--features=auth,audit", "sample-project"],
        )

    def test_git_generator_materializes_from_temporary_output(self) -> None:
        manifest = {
            "project": {"name": "sample-api"},
            "starters": [
                {
                    "id": "hono-api",
                    "pin": "360eb274cc5936fee5aab88eb8bd94977e95dfc9",
                    "adapter": "curation/hono-api/adapter.json",
                    "generator_features": ["persistence", "auth"],
                }
            ],
        }

        def fake_checkout(_source: dict, destination: Path) -> None:
            destination.mkdir()

        def fake_run(command: list[str], _cwd: Path, *, gate: str | None = None) -> dict:
            for token in command:
                if token.startswith("--out="):
                    generated = Path(token.removeprefix("--out="))
                    (generated / "apps/api/src").mkdir(parents=True)
                    (generated / "apps/api/src/server.ts").write_text("export {};\n", encoding="utf-8")
                    (generated / "package.json").write_text(
                        dumps({"name": "generated-api", "packageManager": "bun@1.4.0"}),
                        encoding="utf-8",
                    )
            record = {"command": command, "workdir": ".", "returncode": 0, "status": "passed"}
            if gate:
                record["gate"] = gate
            return record

        with tempfile.TemporaryDirectory() as temporary:
            project = Path(temporary)
            with patch("scripts.eng._checkout_git_source", side_effect=fake_checkout), patch(
                "scripts.eng._run_record", side_effect=fake_run
            ), patch("scripts.eng._assert_adapter_requirements", return_value=[]):
                result = materialize_project(
                    manifest, project, skip_setup=True, skip_checks=True
                )
            self.assertTrue((project / "services/api/apps/api/src/server.ts").exists())
            self.assertEqual(result["starters"][0]["type"], "git-generator")
            self.assertEqual(result["packages"][0]["package_manager"], "bun")

    def test_materializer_environment_reaches_source_generator_setup_and_checks(self) -> None:
        adapter = {
            "boilerplate_id": "sample-generator",
            "source": {"repository": "https://example.com/sample", "commit": "sample-pin"},
            "materializer": {
                "type": "git-generator",
                "destination": "generated",
                "source_setup": [["tool", "prepare"]],
                "command": ["tool", "--out={output}"],
                "setup": [["tool", "setup"]],
                "checks": [{"gate": "test", "command": ["tool", "check"]}],
                "environment": {"STATIC_FLAG": "enabled"},
            },
        }
        calls: list[tuple[list[str], dict[str, str] | None]] = []

        def fake_checkout(_source: dict, destination: Path) -> None:
            destination.mkdir()

        def fake_run(
            command: list[str],
            _cwd: Path,
            *,
            gate: str | None = None,
            environment: dict[str, str] | None = None,
        ) -> dict:
            calls.append((command, environment))
            for token in command:
                if token.startswith("--out="):
                    generated = Path(token.removeprefix("--out="))
                    generated.mkdir(parents=True)
                    (generated / "package.json").write_text(
                        dumps({"name": "generated"}), encoding="utf-8"
                    )
            record = {"command": command, "workdir": ".", "returncode": 0, "status": "passed"}
            if gate:
                record["gate"] = gate
            return record

        def fake_read_json(path: Path) -> dict:
            if path.as_posix().endswith("custom/sample-generator.json"):
                return adapter
            return read_json(path)

        manifest = {
            "project": {"name": "sample-project"},
            "starters": [
                {
                    "id": "sample-generator",
                    "pin": "sample-pin",
                    "adapter": "custom/sample-generator.json",
                }
            ],
        }
        with tempfile.TemporaryDirectory() as temporary:
            project = Path(temporary)
            with patch("scripts.eng.read_json", side_effect=fake_read_json), patch(
                "scripts.eng._checkout_git_source", side_effect=fake_checkout
            ), patch("scripts.eng._run_record", side_effect=fake_run), patch(
                "scripts.eng._assert_adapter_requirements", return_value=[]
            ):
                result = materialize_project(project=project, manifest=manifest)

        self.assertEqual(len(calls), 4)
        self.assertTrue(all(environment == {"STATIC_FLAG": "enabled"} for _, environment in calls))
        self.assertEqual(result["starters"][0]["generator"]["environment"], {"STATIC_FLAG": "enabled"})
        self.assertEqual(result["setup"][0]["environment"], {"STATIC_FLAG": "enabled"})
        self.assertEqual(result["checks"][0]["environment"], {"STATIC_FLAG": "enabled"})

    def test_skipped_materializer_records_retain_declared_environment(self) -> None:
        manifest = {
            "project": {"name": "sample-web"},
            "starters": [
                {
                    "id": "stardrive",
                    "pin": "5c449810b763140ac72133ff4ae63d8497cce77a",
                    "adapter": "curation/stardrive/adapter.json",
                }
            ],
        }

        def fake_checkout(_source: dict, destination: Path) -> None:
            destination.mkdir()
            (destination / "package.json").write_text(
                dumps({"name": "stardrive"}), encoding="utf-8"
            )

        with tempfile.TemporaryDirectory() as temporary:
            with patch("scripts.eng._checkout_git_source", side_effect=fake_checkout), patch(
                "scripts.eng._assert_adapter_requirements", return_value=[]
            ):
                result = materialize_project(
                    manifest,
                    Path(temporary),
                    skip_setup=True,
                    skip_checks=True,
                )

        expected = {"SHARP_IGNORE_GLOBAL_LIBVIPS": "1"}
        self.assertEqual(result["setup"][0]["status"], "skipped")
        self.assertEqual(result["checks"][0]["status"], "skipped")
        self.assertEqual(result["setup"][0]["environment"], expected)
        self.assertTrue(all(record["environment"] == expected for record in result["checks"]))

    def test_check_reuses_environment_from_materialization_records(self) -> None:
        intake = {
            "name": "assistant-app",
            "project_type": "ai-assistant",
            "signals": ["chat", "tools"],
            "features": [],
            "excluded_features": [],
            "database": "postgresql-managed",
        }
        expected = {"STATIC_FLAG": "enabled"}
        with tempfile.TemporaryDirectory() as temporary:
            output = Path(temporary) / "assistant-app"
            write_project(
                intake,
                output,
                materialize=True,
                skip_setup=True,
                skip_checks=True,
            )
            materialization_path = output / ".engineering/materialization.json"
            materialization = loads(materialization_path.read_text(encoding="utf-8"))
            for record in materialization["setup"] + materialization["checks"]:
                record["environment"] = expected
            materialization_path.write_text(dumps(materialization), encoding="utf-8")
            calls: list[dict[str, str] | None] = []

            def fake_run(
                command: list[str],
                _cwd: Path,
                *,
                gate: str | None = None,
                environment: dict[str, str] | None = None,
            ) -> dict:
                calls.append(environment)
                record = {"command": command, "workdir": ".", "returncode": 0, "status": "passed"}
                if gate:
                    record["gate"] = gate
                return record

            args = Namespace(project=str(output), changed_files=[], run=True)
            with patch("scripts.eng.adapter_preflight", return_value=[]), patch(
                "scripts.eng._run_record", side_effect=fake_run
            ), patch("builtins.print"):
                self.assertEqual(command_check(args), 0)

            after = loads(materialization_path.read_text(encoding="utf-8"))

        self.assertEqual(calls, [expected, *([expected] * 2)])
        self.assertTrue(all(record["environment"] == expected for record in after["setup"]))
        self.assertTrue(all(record["environment"] == expected for record in after["checks"]))

    def test_generated_ci_emits_only_declared_adapter_environment(self) -> None:
        manifest = {
            "starters": [
                {
                    "id": "stardrive",
                    "adapter": "curation/stardrive/adapter.json",
                    "destination": ".",
                }
            ]
        }
        materialization = {
            "setup": [
                {"starter": "stardrive", "command": ["npm", "ci"], "workdir": "."}
            ],
            "checks": [
                {
                    "starter": "stardrive",
                    "gate": "typecheck",
                    "command": ["npm", "run", "check"],
                    "workdir": ".",
                }
            ],
        }
        with patch.dict(os.environ, {"AMBIENT_SECRET": "do-not-export"}):
            workflow = project_ci_yaml(manifest, materialization)

        self.assertIn(
            '    env:\n      SHARP_IGNORE_GLOBAL_LIBVIPS: "1"\n    steps:',
            workflow,
        )
        self.assertNotIn("AMBIENT_SECRET", workflow)
        self.assertNotIn("do-not-export", workflow)

    def test_command_generator_owns_creation_of_its_destination(self) -> None:
        manifest = {
            "project": {"name": "sample-mobile"},
            "starters": [
                {
                    "id": "ignite",
                    "pin": "e829d2f922c5568a59a77bfb6232aeb500be3f13",
                    "adapter": "curation/ignite/adapter.json",
                }
            ],
        }

        def fake_run(command: list[str], cwd: Path, *, gate: str | None = None) -> dict:
            destination = cwd / "mobile"
            self.assertFalse(destination.exists())
            destination.mkdir()
            (destination / "package.json").write_text(
                dumps({"name": "mobile", "scripts": {}}), encoding="utf-8"
            )
            return {"command": command, "workdir": ".", "returncode": 0, "status": "passed"}

        with tempfile.TemporaryDirectory() as temporary:
            project = Path(temporary)
            with patch("scripts.eng._run_record", side_effect=fake_run), patch(
                "scripts.eng._assert_adapter_requirements", return_value=[]
            ), patch(
                "scripts.eng._apply_adapter_patches", return_value=[]
            ):
                result = materialize_project(
                    manifest, project, skip_setup=True, skip_checks=True
                )
            self.assertTrue((project / "apps/mobile/package.json").exists())
            self.assertEqual(result["starters"][0]["destination"], "apps/mobile")

    def test_overlay_adds_agent_context_without_engine_special_cases(self) -> None:
        root = Path(__file__).parents[1]
        adapter = loads((root / "curation/ignite/adapter.json").read_text(encoding="utf-8"))
        with tempfile.TemporaryDirectory() as temporary:
            destination = Path(temporary)
            _apply_overlay(adapter, destination)
            self.assertIn("Ignite mobile app", (destination / "AGENTS.md").read_text(encoding="utf-8"))

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
    def test_rejects_unsafe_adapter_environment_values(self) -> None:
        self.assertTrue(
            adapter_environment_errors(
                {"INVALID-NAME": "1"}, "stardrive: materializer.environment"
            )
        )
        self.assertTrue(
            adapter_environment_errors(
                {"VALID_NAME": "line\nbreak"}, "stardrive: materializer.environment"
            )
        )
        self.assertTrue(
            adapter_environment_errors(
                {"VALID_NAME": "nul\x00value"}, "stardrive: materializer.environment"
            )
        )
        self.assertTrue(
            adapter_environment_errors([], "stardrive: materializer.environment")
        )

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
