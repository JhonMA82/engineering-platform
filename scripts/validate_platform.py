#!/usr/bin/env python3
"""Cross-file validation for the Engineering Platform."""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path
from typing import Any
from urllib.parse import unquote


ROOT = Path(__file__).resolve().parents[1]
ERRORS: list[str] = []
SKIPPED_MARKDOWN_DIRECTORIES = frozenset({".git", "node_modules"})


def error(message: str) -> None:
    ERRORS.append(message)


def load(relative: str) -> dict[str, Any]:
    path = ROOT / relative
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except FileNotFoundError:
        error(f"Falta {relative}")
    except json.JSONDecodeError as exc:
        error(f"JSON inválido {relative}:{exc.lineno}: {exc.msg}")
    return {}


def index(items: list[dict[str, Any]], label: str) -> dict[str, dict[str, Any]]:
    result: dict[str, dict[str, Any]] = {}
    for item in items:
        item_id = item.get("id")
        if not isinstance(item_id, str) or not item_id:
            error(f"{label}: entrada sin id")
        elif item_id in result:
            error(f"{label}: id duplicado {item_id}")
        else:
            result[item_id] = item
    return result


def validate_required_files() -> None:
    required = [
        "README.md",
        "AGENTS.md",
        "eng",
        "package.json",
        "extensions/engineering-platform.ts",
        "pi-skills/project-discovery/SKILL.md",
        "platform/catalog.json",
        "platform/boilerplates.json",
        "platform/golden-paths.json",
        "platform/feature-packs.json",
        "platform/database-profiles.json",
        "skills/registry.json",
        "schemas/project.schema.json",
        "schemas/project-definition.schema.json",
        "templates/project-intake.json",
        "docs/01-concepts/CONCEPTS_WITH_EXAMPLES.md",
        "docs/13-examples/END_TO_END_SCHOOL_REQUESTS.md",
        "catalog/legacy-v1.2.3/catalog.json",
    ]
    for relative in required:
        if not (ROOT / relative).exists():
            error(f"Falta {relative}")


def validate_all_json() -> None:
    for path in ROOT.rglob("*.json"):
        if ".git" in path.parts:
            continue
        try:
            value = json.loads(path.read_text(encoding="utf-8"))
        except json.JSONDecodeError as exc:
            error(f"JSON inválido {path.relative_to(ROOT)}:{exc.lineno}: {exc.msg}")
            continue
        if isinstance(value, dict) and "$schema" in value:
            schema = value["$schema"]
            if isinstance(schema, str) and not schema.startswith(("http://", "https://")):
                if not (path.parent / schema).resolve().exists():
                    error(f"Schema inexistente: {path.relative_to(ROOT)} -> {schema}")
        if "PINNED" in json.dumps(value):
            error(f"Marcador PINNED sin resolver en {path.relative_to(ROOT)}")


def validate_registry() -> None:
    catalog = load("platform/catalog.json")
    boilerplates_doc = load("platform/boilerplates.json")
    recipes_doc = load("platform/golden-paths.json")
    features_doc = load("platform/feature-packs.json")
    databases_doc = load("platform/database-profiles.json")
    skills_doc = load("skills/registry.json")

    boilerplates = index(boilerplates_doc.get("entries", []), "boilerplates")
    recipes = index(recipes_doc.get("paths", []), "recipes")
    features = index(features_doc.get("features", []), "features")
    databases = index(databases_doc.get("profiles", []), "databases")
    skills = index(skills_doc.get("skills", []), "skills")

    decisions = set(boilerplates_doc.get("decision_statuses", {}))
    deliveries = set(boilerplates_doc.get("delivery_statuses", {}))
    tiers = set(boilerplates_doc.get("maintenance_tiers", {}))
    seen_repositories: dict[str, str] = {}
    seen_names = set(boilerplates)
    for item in boilerplates.values():
        if item.get("decision_status") not in decisions:
            error(f"{item['id']}: decision_status inválido")
        if item.get("delivery_status") not in deliveries:
            error(f"{item['id']}: delivery_status inválido")
        if item.get("maintenance_tier") not in tiers:
            error(f"{item['id']}: maintenance_tier inválido")
        profile = item.get("profile")
        if not profile or not (ROOT / profile).exists():
            error(f"{item['id']}: profile inexistente {profile}")
        repository = item.get("repository")
        if repository:
            normalized = repository.lower().removesuffix(".git").rstrip("/")
            if normalized in seen_repositories:
                error(f"Repositorio duplicado: {item['id']} y {seen_repositories[normalized]}")
            seen_repositories[normalized] = item["id"]
        for legacy_id in item.get("legacy_ids", []):
            if legacy_id in seen_names:
                error(f"Alias duplicado o en conflicto: {legacy_id}")
            seen_names.add(legacy_id)
        adapter = item.get("integration", {}).get("adapter")
        if adapter and not (ROOT / adapter).exists():
            error(f"{item['id']}: adapter inexistente {adapter}")
        if adapter and (ROOT / adapter).exists():
            adapter_doc = load(adapter)
            if adapter_doc.get("boilerplate_id") != item["id"]:
                error(f"{item['id']}: adapter declara otro boilerplate_id")
            source = adapter_doc.get("source", {})
            upstream = item.get("upstream", {})
            if source.get("commit") != upstream.get("commit"):
                error(f"{item['id']}: pin distinto entre registro y adapter")
            if source.get("repository") != item.get("repository"):
                error(f"{item['id']}: repository distinto entre registro y adapter")
            materializer = adapter_doc.get("materializer", {})
            kind = materializer.get("type")
            if kind not in {"git-copy", "local-copy", "command-generator", "git-generator"}:
                error(f"{item['id']}: materializer inválido {kind}")
            destination = materializer.get("destination")
            if not isinstance(destination, str) or destination.startswith("/") or ".." in destination.split("/"):
                error(f"{item['id']}: destination inseguro")
            if kind in {"git-copy", "git-generator"} and (
                not source.get("repository") or not source.get("commit")
            ):
                error(f"{item['id']}: {kind} necesita repository y commit")
            if kind == "local-copy":
                local_path = source.get("local_path")
                if not isinstance(local_path, str) or not (ROOT / local_path).is_dir():
                    error(f"{item['id']}: local_path inexistente {local_path}")
            if kind in {"command-generator", "git-generator"} and not materializer.get("command"):
                error(f"{item['id']}: {kind} sin comando")
            if adapter_doc.get("integration", {}).get("mode") != item.get("integration", {}).get("mode"):
                error(f"{item['id']}: mode distinto entre registro y adapter")
            if adapter_doc.get("integration", {}).get("update_strategy") != item.get("integration", {}).get("update_strategy"):
                error(f"{item['id']}: estrategia distinta entre registro y adapter")
            overlay = adapter_doc.get("overlay")
            if overlay and not (ROOT / overlay).is_dir():
                error(f"{item['id']}: overlay inexistente {overlay}")
            evidence = item.get("integration", {}).get("evidence")
            if evidence:
                evidence_doc = load(evidence)
                if evidence_doc.get("boilerplate_id") != item["id"]:
                    error(f"{item['id']}: evidence declara otro boilerplate_id")
            for command in materializer.get("setup", []):
                if not isinstance(command, list) or not command:
                    error(f"{item['id']}: comando setup inválido")
            for check in materializer.get("checks", []):
                if not check.get("gate") or not check.get("command"):
                    error(f"{item['id']}: check inválido")

    for skill in skills.values():
        path = skill.get("path")
        if not path or not (ROOT / path).exists():
            error(f"Skill sin implementación: {skill['id']} -> {path}")

    for feature in features.values():
        for requirement in feature.get("requires", []):
            if requirement != "database" and requirement not in features:
                error(f"{feature['id']}: dependencia desconocida {requirement}")
        for skill in feature.get("skills", []):
            if skill not in skills:
                error(f"{feature['id']}: skill desconocido {skill}")

    for recipe in recipes.values():
        stack = recipe.get("stack", {})
        for starter_id in stack.get("starters", []) + stack.get("alternatives", []):
            if starter_id not in boilerplates:
                error(f"{recipe['id']}: boilerplate desconocido {starter_id}")
        for starter_id in stack.get("starters", []):
            starter = boilerplates.get(starter_id, {})
            if not starter.get("integration", {}).get("adapter") or not starter.get("upstream", {}).get("commit"):
                error(f"{recipe['id']}: starter default no materializable {starter_id}")
        for pack_id in recipe.get("solution_packs", {}).get("default", []) + recipe.get("solution_packs", {}).get("optional", []):
            if pack_id not in boilerplates or boilerplates[pack_id].get("kind") != "solution-pack":
                error(f"{recipe['id']}: solution pack inválido {pack_id}")
        database = stack.get("database", {})
        allowed = database.get("allowed", [])
        default = database.get("default")
        if default is not None and default not in allowed:
            error(f"{recipe['id']}: base default fuera de allowed")
        for database_id in allowed:
            if database_id not in databases:
                error(f"{recipe['id']}: base desconocida {database_id}")
        recipe_features = recipe.get("features", {})
        default_features = recipe_features.get("default", [])
        optional_features = recipe_features.get("optional", [])
        if set(default_features).intersection(optional_features):
            error(f"{recipe['id']}: feature default y opcional a la vez")
        for feature_id in default_features + optional_features:
            if feature_id not in features:
                error(f"{recipe['id']}: feature desconocido {feature_id}")
        for skill_id in recipe.get("skills", []):
            if skill_id not in skills:
                error(f"{recipe['id']}: skill desconocido {skill_id}")

    for label, relative in catalog.get("source_of_truth", {}).items():
        if not (ROOT / relative).exists():
            error(f"catalog.source_of_truth.{label} no existe: {relative}")

def repository_markdown_files(root: Path = ROOT) -> list[Path]:
    files = []
    for path in root.rglob("*.md"):
        if SKIPPED_MARKDOWN_DIRECTORIES.intersection(path.relative_to(root).parts):
            continue
        files.append(path)
    return files


def validate_markdown_links() -> None:
    link_pattern = re.compile(r"(?<!!)\[[^\]]+\]\(([^)]+)\)")
    for path in repository_markdown_files():
        content = path.read_text(encoding="utf-8")
        for target in link_pattern.findall(content):
            target = target.strip().split(" ", 1)[0]
            if target.startswith(("http://", "https://", "mailto:", "#")):
                continue
            target = unquote(target.split("#", 1)[0])
            if not target:
                continue
            resolved = (path.parent / target).resolve()
            if not resolved.exists():
                error(f"Enlace roto: {path.relative_to(ROOT)} -> {target}")


def validate_concepts() -> None:
    path = ROOT / "docs/01-concepts/CONCEPTS_WITH_EXAMPLES.md"
    if not path.exists():
        return
    content = path.read_text(encoding="utf-8")
    terms = [
        "Golden Path",
        "Feature Pack",
        "Project Manifest",
        "Harness",
        "Skill",
        "Guard",
        "Quality Gate",
        "Canonical Example",
        "Eval",
        "ADR",
        "Knowledge Entry",
        "Upgrade Recipe",
        "Design System",
    ]
    for term in terms:
        if term not in content:
            error(f"Concepto sin documentar: {term}")


def validate_pi_package() -> None:
    package = load("package.json")
    if not isinstance(package.get("version"), str) or not package["version"]:
        error("package.json: falta la versión Pi")
    if "pi-package" not in package.get("keywords", []):
        error("package.json: falta keyword pi-package")
    manifest = package.get("pi", {})
    expected = {
        "extensions": "./extensions/engineering-platform.ts",
        "skills": "./pi-skills",
        "prompts": "./prompts",
    }
    for resource, relative in expected.items():
        if relative not in manifest.get(resource, []):
            error(f"package.json: pi.{resource} no declara {relative}")
        if not (ROOT / relative).exists():
            error(f"package.json: recurso Pi inexistente {relative}")
    if "./.opencode/skills" not in manifest.get("skills", []):
        error("package.json: Pi no declara los skills operativos existentes")
    extension_path = ROOT / "extensions/engineering-platform.ts"
    if extension_path.exists():
        extension = extension_path.read_text(encoding="utf-8")
        for command in ('registerCommand("new-project"', 'registerCommand("engineering-status"'):
            if command not in extension:
                error(f"Extensión Pi sin comando requerido: {command}")
    skill_path = ROOT / "pi-skills/project-discovery/SKILL.md"
    if skill_path.exists():
        skill = skill_path.read_text(encoding="utf-8")
        if not skill.startswith("---\nname: project-discovery\n") or "description:" not in skill:
            error("Skill Pi project-discovery sin frontmatter válido")


def main() -> int:
    validate_required_files()
    validate_all_json()
    validate_registry()
    validate_markdown_links()
    validate_concepts()
    validate_pi_package()
    if ERRORS:
        print("VALIDATION FAILED", file=sys.stderr)
        for item in ERRORS:
            print(f"- {item}", file=sys.stderr)
        return 1
    markdown_count = len(repository_markdown_files())
    json_count = sum(1 for _ in ROOT.rglob("*.json"))
    print(f"OK: {markdown_count} Markdown, {json_count} JSON y referencias cruzadas válidas.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
