#!/usr/bin/env python3
"""Minimal CLI for selecting, recording and checking project recipes."""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import re
import shutil
import subprocess
import sys
import tempfile
from copy import deepcopy
from datetime import date
from pathlib import Path
from typing import Any
from urllib.parse import urlparse


ROOT = Path(__file__).resolve().parents[1]
PLATFORM_VERSION = json.loads((ROOT / "package.json").read_text(encoding="utf-8"))["version"]
DEFINITION_SCHEMA_VERSION = 1
INSTALL_IGNORES = shutil.ignore_patterns(
    ".git",
    ".atl",
    ".env",
    ".env.*",
    "__pycache__",
    ".pytest_cache",
    "node_modules",
    "*.pyc",
    "*.zip",
)


class PlatformError(ValueError):
    """A user-facing platform decision or validation error."""


def read_json(path: Path) -> dict[str, Any]:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except FileNotFoundError as exc:
        raise PlatformError(f"No existe {path}") from exc
    except json.JSONDecodeError as exc:
        raise PlatformError(f"JSON inválido en {path}: {exc}") from exc


def dump_json(value: Any) -> str:
    return json.dumps(value, ensure_ascii=False, indent=2) + "\n"


def data() -> dict[str, Any]:
    return {
        "boilerplates": read_json(ROOT / "platform/boilerplates.json"),
        "recipes": read_json(ROOT / "platform/golden-paths.json"),
        "features": read_json(ROOT / "platform/feature-packs.json"),
        "databases": read_json(ROOT / "platform/database-profiles.json"),
        "skills": read_json(ROOT / "skills/registry.json"),
    }


def by_id(items: list[dict[str, Any]]) -> dict[str, dict[str, Any]]:
    return {item["id"]: item for item in items}


def unique(items: list[str]) -> list[str]:
    return list(dict.fromkeys(items))


def normalize_repository(value: str) -> str:
    raw = value.strip()
    if not raw:
        raise PlatformError("La URL del repositorio está vacía.")
    if "://" not in raw:
        raw = "https://" + raw
    parsed = urlparse(raw)
    host = parsed.netloc.lower().removeprefix("www.")
    path = re.sub(r"/+", "/", parsed.path).strip("/")
    if path.endswith(".git"):
        path = path[:-4]
    if not host or len(path.split("/")) < 2:
        raise PlatformError(f"URL de repositorio no reconocida: {value}")
    if host == "github.com":
        owner, repo, *_ = path.split("/")
        path = f"{owner.lower()}/{repo.lower()}"
    return f"https://{host}/{path}"


def evaluate_boilerplate(
    repository: str,
    *,
    observed_commit: str | None = None,
    observed_date: str | None = None,
    category: str | None = None,
) -> dict[str, Any]:
    registry = data()["boilerplates"]
    normalized = normalize_repository(repository)
    entries = registry["entries"]
    exact = next(
        (
            item
            for item in entries
            if item.get("repository")
            and normalize_repository(item["repository"]) == normalized
        ),
        None,
    )
    if exact:
        upstream = exact.get("upstream", {})
        stored_commit = upstream.get("commit")
        stored_date = upstream.get("observed_at")
        needs_refresh = bool(
            (observed_commit and stored_commit and observed_commit != stored_commit)
            or (observed_date and stored_date and observed_date > stored_date)
            or (observed_commit and not stored_commit)
        )
        return {
            "decision": "ALREADY_REGISTERED_REFRESH" if needs_refresh else "ALREADY_REGISTERED",
            "repository": normalized,
            "entry_id": exact["id"],
            "decision_status": exact["decision_status"],
            "delivery_status": exact["delivery_status"],
            "reason": (
                "La URL ya corresponde a una entrada y existe evidencia upstream más reciente."
                if needs_refresh
                else "La URL ya corresponde a una entrada canónica; crear otra duplicaría catálogo, mantenimiento y decisiones."
            ),
            "next_action": (
                f"Actualizar {exact['id']} mediante {exact['integration']['update_strategy']} y conservar el mismo id."
                if needs_refresh
                else f"Reutilizar {exact['id']} o reevaluar su estado; no crear otra entrada."
            ),
            "stored_commit": stored_commit,
            "observed_commit": observed_commit,
        }

    alternatives = [
        item["id"]
        for item in entries
        if category and item.get("category") == category
    ]
    return {
        "decision": "ADD_AS_CANDIDATE",
        "repository": normalized,
        "entry_id": None,
        "reason": (
            "No existe una URL canónica equivalente. La entrada debe quedar como candidata hasta completar licencia, mantenimiento, seguridad, prueba y comparación."
        ),
        "compare_with": alternatives,
        "next_action": "Crear ficha con evidencia y promover solo si cubre una brecha o mejora materialmente una opción existente.",
    }


def _recipe_for_intake(intake: dict[str, Any], recipes: list[dict[str, Any]]) -> dict[str, Any]:
    project_type = intake.get("project_type")
    signals = set(intake.get("signals", []))
    ranked: list[tuple[int, dict[str, Any]]] = []
    for recipe in recipes:
        match = recipe["match"]
        score = (100 if project_type in match["project_types"] else 0) + len(
            signals.intersection(match["signals"])
        )
        ranked.append((score, recipe))
    ranked.sort(key=lambda pair: (-pair[0], pair[1]["id"]))
    if not ranked or ranked[0][0] < 100:
        supported = sorted(
            {kind for recipe in recipes for kind in recipe["match"]["project_types"]}
        )
        raise PlatformError(
            f"project_type={project_type!r} no está soportado. Opciones: {', '.join(supported)}"
        )
    return ranked[0][1]


def _expand_features(
    selected: list[str],
    feature_index: dict[str, dict[str, Any]],
    database: str | None,
) -> list[str]:
    result: list[str] = []
    visiting: set[str] = set()

    def add(feature_id: str) -> None:
        if feature_id in result:
            return
        if feature_id not in feature_index:
            raise PlatformError(f"Feature desconocido: {feature_id}")
        if feature_id in visiting:
            raise PlatformError(f"Dependencia cíclica de feature: {feature_id}")
        visiting.add(feature_id)
        for requirement in feature_index[feature_id]["requires"]:
            if requirement == "database":
                if not database:
                    raise PlatformError(f"{feature_id} requiere una base de datos")
            else:
                add(requirement)
        visiting.remove(feature_id)
        result.append(feature_id)

    for feature in selected:
        add(feature)
    return result


def resolve_recipe(intake: dict[str, Any]) -> dict[str, Any]:
    platform = data()
    recipe = _recipe_for_intake(intake, platform["recipes"]["paths"])
    feature_index = by_id(platform["features"]["features"])
    database_index = by_id(platform["databases"]["profiles"])
    boilerplate_index = by_id(platform["boilerplates"]["entries"])

    requested_database = intake.get("database", recipe["stack"]["database"]["default"])
    allowed_databases = recipe["stack"]["database"]["allowed"]
    if requested_database is not None:
        if requested_database not in database_index:
            raise PlatformError(f"Perfil de base de datos desconocido: {requested_database}")
        if requested_database not in allowed_databases:
            raise PlatformError(
                f"{requested_database} no está permitido por {recipe['id']}; permitidos: {', '.join(allowed_databases) or 'ninguno'}"
            )

    requested = list(recipe["features"]["default"])
    optional = set(recipe["features"]["optional"])
    for feature in intake.get("features", []):
        if feature not in optional and feature not in requested:
            raise PlatformError(f"{feature} no está contemplado por {recipe['id']}")
        if feature not in requested:
            requested.append(feature)
    excluded = set(intake.get("excluded_features", []))
    conflict = excluded.intersection(intake.get("features", []))
    if conflict:
        raise PlatformError(f"Features incluidos y excluidos a la vez: {', '.join(sorted(conflict))}")
    requested = [feature for feature in requested if feature not in excluded]
    features = _expand_features(requested, feature_index, requested_database)
    excluded_dependencies = excluded.intersection(features)
    if excluded_dependencies:
        raise PlatformError(
            "La Recipe necesita features excluidos como dependencia: "
            + ", ".join(sorted(excluded_dependencies))
        )

    resolved_skills = list(recipe["skills"])
    resolved_gates = list(recipe["gates"])
    for feature_id in features:
        resolved_skills.extend(feature_index[feature_id].get("skills", []))
        resolved_gates.extend(feature_index[feature_id].get("gates", []))
    if requested_database:
        resolved_gates.extend(database_index[requested_database].get("required_gates", []))

    starters: list[dict[str, Any]] = []
    warnings: list[str] = []
    for starter_id in recipe["stack"]["starters"]:
        item = boilerplate_index[starter_id]
        pin = item.get("upstream", {}).get("commit")
        adapter_path = item.get("integration", {}).get("adapter")
        adapter = read_json(ROOT / adapter_path) if adapter_path else None
        materializer = (adapter or {}).get("materializer", {})
        starters.append(
            {
                "id": starter_id,
                "delivery_status": item["delivery_status"],
                "integration_mode": item["integration"]["mode"],
                "update_strategy": item["integration"]["update_strategy"],
                "pin": pin,
                "repository": item.get("repository"),
                "adapter": adapter_path,
                "destination": materializer.get("destination"),
            }
        )
        if not adapter_path or not pin:
            warnings.append(
                f"{starter_id} no tiene adapter y pin materializables: el resultado será blueprint."
            )
    if recipe["status"] != "stable":
        warnings.append(f"{recipe['id']} está en canal {recipe['status']} y requiere aceptación explícita.")
    if requested_database and database_index[requested_database]["status"] != "stable":
        warnings.append(
            f"{requested_database} está en canal {database_index[requested_database]['status']} y requiere piloto."
        )

    project_name = intake.get("name")
    if not project_name or not re.fullmatch(r"[a-z0-9]+(?:-[a-z0-9]+)*", project_name):
        raise PlatformError("name debe usar kebab-case, por ejemplo school-requests")

    return {
        "$schema": "https://raw.githubusercontent.com/JhonMA82/engineering-platform/main/schemas/project.schema.json",
        "schema_version": 2,
        "platform_version": PLATFORM_VERSION,
        "generated_at": date.today().isoformat(),
        "scaffold_status": "blueprint",
        "readiness": "code-ready",
        "project": {"name": project_name, "type": intake["project_type"]},
        "recipe": {"id": recipe["id"], "version": recipe["version"], "status": recipe["status"]},
        "starters": starters,
        "database": requested_database,
        "features": features,
        "skills": unique(resolved_skills),
        "gates": unique(resolved_gates),
        "exclusions": sorted(set(recipe["exclusions"]).union(excluded)),
        "ownership": {
            "managed": [".engineering/project.json", ".engineering/intake.json"],
            "managed_sections": ["AGENTS.md", "ARCHITECTURE.md"],
            "seeded": ["README.md"],
            "user_owned": ["apps/**", "packages/**", "src/**", "tests/**"],
        },
        "warnings": warnings,
    }


def architecture_markdown(manifest: dict[str, Any]) -> str:
    starters = "\n".join(
        f"- `{item['id']}` ({item['delivery_status']}, {item['integration_mode']})"
        for item in manifest["starters"]
    )
    features = ", ".join(f"`{item}`" for item in manifest["features"]) or "ninguno"
    gates = ", ".join(f"`{item}`" for item in manifest["gates"])
    exclusions = "\n".join(f"- `{item}`" for item in manifest["exclusions"])
    warnings = "\n".join(f"- {item}" for item in manifest["warnings"]) or "- Ninguna."
    summary = manifest["project"].get(
        "summary", "Proyecto definido desde intake; no hay resumen de descubrimiento."
    )
    return f"""# Arquitectura: {manifest['project']['name']}

Documento generado desde la Recipe `{manifest['recipe']['id']}@{manifest['recipe']['version']}`. Las decisiones de dominio deben registrarse como ADR; no se cambia el stack silenciosamente.

## Idea

{summary}

## Stack

{starters}

- Base de datos: `{manifest['database'] or 'ninguna'}`
- Features: {features}
- Estado del scaffold: `{manifest['scaffold_status']}`

## Quality Gates

{gates}

## Exclusiones explícitas

{exclusions}

## Advertencias

{warnings}
"""


def agents_markdown(manifest: dict[str, Any]) -> str:
    return f"""# AGENTS.md

Antes de cambiar código, lee `.engineering/project.json` y `ARCHITECTURE.md`.

- Recipe: `{manifest['recipe']['id']}@{manifest['recipe']['version']}`.
- Skills permitidos para esta arquitectura: {', '.join(f'`{item}`' for item in manifest['skills'])}.
- Ejecuta los gates aplicables antes de terminar: {', '.join(f'`{item}`' for item in manifest['gates'])}.
- Lee `GENTLE.md` para la intención del producto y la estrategia de entrega.
- No agregues frameworks, bases de datos o feature packs sin actualizar primero el intake y resolver de nuevo la Recipe.
- Respeta ownership: no sobrescribas archivos `user_owned`; los cambios a `managed_sections` deben limitarse a su sección identificada.
"""


def readme_markdown(manifest: dict[str, Any]) -> str:
    return f"""# {manifest['project']['name']}

Proyecto definido por Engineering Platform `{manifest['platform_version']}` con Recipe `{manifest['recipe']['id']}`.

Estado: **{manifest['scaffold_status']}**. Consulta `ARCHITECTURE.md` y `.engineering/project.json` antes de materializar o desarrollar el producto.
"""


def _string_list(value: Any, field: str, *, allow_empty: bool = True) -> list[str]:
    if not isinstance(value, list) or any(not isinstance(item, str) or not item.strip() for item in value):
        raise PlatformError(f"{field} debe ser una lista de textos no vacíos")
    if not allow_empty and not value:
        raise PlatformError(f"{field} necesita al menos un elemento")
    return value


def validate_project_definition(definition: dict[str, Any]) -> dict[str, Any]:
    if definition.get("schema_version") != DEFINITION_SCHEMA_VERSION:
        raise PlatformError(
            f"schema_version de definición no soportado: {definition.get('schema_version')}"
        )
    idea = definition.get("idea")
    intake = definition.get("intake")
    delivery = definition.get("delivery")
    discovery = definition.get("discovery")
    if not isinstance(idea, dict) or not isinstance(intake, dict):
        raise PlatformError("La definición necesita objetos idea e intake")
    if not isinstance(delivery, dict) or not isinstance(discovery, dict):
        raise PlatformError("La definición necesita objetos delivery y discovery")
    for field in ("summary", "problem"):
        if not isinstance(idea.get(field), str) or not idea[field].strip():
            raise PlatformError(f"idea.{field} es obligatorio")
    for field in ("users", "outcomes", "must_have"):
        _string_list(idea.get(field), f"idea.{field}", allow_empty=False)
    _string_list(idea.get("out_of_scope", []), "idea.out_of_scope")
    for field in ("acceptance_criteria", "risks", "unknowns"):
        _string_list(delivery.get(field, []), f"delivery.{field}")
    if discovery.get("status") != "confirmed":
        raise PlatformError("La definición debe estar confirmada por el usuario antes del bootstrap")
    if discovery.get("confirmed_by") != "user":
        raise PlatformError("discovery.confirmed_by debe ser user")
    if not isinstance(discovery.get("confirmed_at"), str) or not discovery["confirmed_at"].strip():
        raise PlatformError("discovery.confirmed_at es obligatorio")
    resolve_recipe(intake)
    return definition


def _ensure_target_is_safe(
    output: Path,
    allowed_existing: set[Path] | None = None,
    require_managed_metadata: bool = False,
) -> None:
    if output.is_symlink():
        raise PlatformError(f"La salida es un enlace simbólico: {output}")
    if not output.exists():
        return
    if not output.is_dir():
        raise PlatformError(f"La salida existe y no es un directorio: {output}")
    if require_managed_metadata:
        entries = {path.name for path in output.iterdir()}
        managed = {".atl", ".gitignore", ".git"}
        if entries and not entries.issubset(managed):
            raise PlatformError(
                f"El destino ya existe y no contiene únicamente metadatos gestionados: {output}"
            )
    allowed = {path.resolve() for path in (allowed_existing or set())}
    unexpected: list[str] = []
    for path in output.rglob("*"):
        relative = path.relative_to(output)
        if path.is_symlink():
            unexpected.append(str(relative))
            continue
        if ".git" in relative.parts:
            continue
        if relative == Path(".atl"):
            if not path.is_dir():
                unexpected.append(str(relative))
            continue
        if relative == Path(".gitignore"):
            if not path.is_file():
                unexpected.append(str(relative))
            continue
        if relative.parts and relative.parts[0] == ".atl":
            continue
        if path.is_dir():
            continue
        if path.resolve() not in allowed:
            unexpected.append(str(relative))
    if unexpected:
        raise PlatformError(
            "El directorio de salida contiene archivos ajenos al bootstrap: "
            + ", ".join(sorted(unexpected))
        )


def _safe_relative(value: str, field: str) -> Path:
    relative = Path(value)
    if relative.is_absolute() or ".." in relative.parts:
        raise PlatformError(f"{field} debe ser una ruta relativa segura: {value}")
    return relative


def _run_record(command: list[str], cwd: Path, *, gate: str | None = None) -> dict[str, Any]:
    if not command or any(not isinstance(item, str) or not item for item in command):
        raise PlatformError("El adapter contiene un comando inválido")
    executable = shutil.which(command[0])
    if not executable:
        raise PlatformError(f"Falta el ejecutable requerido por el adapter: {command[0]}")
    completed = subprocess.run(command, cwd=cwd, text=True, capture_output=True, check=False)
    record: dict[str, Any] = {
        "command": command,
        "workdir": ".",
        "returncode": completed.returncode,
    }
    if gate:
        record["gate"] = gate
    if completed.stdout.strip():
        record["stdout"] = completed.stdout.strip()[-4000:]
    if completed.stderr.strip():
        record["stderr"] = completed.stderr.strip()[-4000:]
    if completed.returncode != 0:
        detail = completed.stderr.strip() or completed.stdout.strip() or "sin salida"
        raise PlatformError(f"Falló {' '.join(command)} en {cwd}: {detail[-1000:]}")
    return record


def _copy_materialized_tree(source: Path, destination: Path) -> None:
    if not source.is_dir():
        raise PlatformError(f"La fuente materializable no existe: {source}")
    for path in source.rglob("*"):
        relative = path.relative_to(source)
        if ".git" in relative.parts or "node_modules" in relative.parts:
            continue
        if path.is_symlink():
            raise PlatformError(f"El starter contiene un enlace simbólico no permitido: {relative}")
        target = destination / relative
        if path.is_dir():
            target.mkdir(parents=True, exist_ok=True)
            continue
        if target.exists():
            if relative == Path(".gitignore") and target.is_file():
                current = target.read_text(encoding="utf-8").splitlines()
                incoming = path.read_text(encoding="utf-8").splitlines()
                target.write_text("\n".join(unique(current + incoming)).rstrip() + "\n", encoding="utf-8")
                continue
            raise PlatformError(f"El starter intentó sobrescribir {target}")
        target.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(path, target)


def _content_sha256(root: Path) -> str:
    digest = hashlib.sha256()
    for path in sorted(root.rglob("*"), key=lambda item: item.as_posix()):
        relative = path.relative_to(root)
        if path.is_dir() or {".git", "node_modules", ".engineering"}.intersection(relative.parts):
            continue
        digest.update(relative.as_posix().encode())
        digest.update(b"\0")
        digest.update(path.read_bytes())
        digest.update(b"\0")
    return digest.hexdigest()


def _actual_project_data(project: Path) -> dict[str, Any]:
    tree: list[str] = []
    packages: list[dict[str, Any]] = []
    environment_files: list[str] = []
    for path in sorted(project.rglob("*"), key=lambda item: item.as_posix()):
        relative = path.relative_to(project)
        if {".git", "node_modules", "dist", "target", "__pycache__"}.intersection(relative.parts):
            continue
        if len(relative.parts) <= 3:
            tree.append(relative.as_posix() + ("/" if path.is_dir() else ""))
        if path.is_file() and path.name in {".env.example", ".env.sample", "env.example"}:
            environment_files.append(relative.as_posix())
        if path.is_file() and path.name == "package.json" and len(relative.parts) <= 4:
            package = read_json(path)
            parent = path.parent
            manager = "npm"
            if (parent / "bun.lock").exists() or (parent / "bun.lockb").exists():
                manager = "bun"
            elif (parent / "pnpm-lock.yaml").exists():
                manager = "pnpm"
            elif (parent / "yarn.lock").exists():
                manager = "yarn"
            packages.append({
                "path": parent.relative_to(project).as_posix() or ".",
                "name": package.get("name"),
                "package_manager": manager,
                "scripts": sorted(package.get("scripts", {})),
            })
    return {"actual_tree": tree, "packages": packages, "environment_files": environment_files}


def _preserve_upstream_instructions(project: Path, starter_id: str) -> None:
    docs = project / "docs/boilerplates"
    for name in ("README.md", "AGENTS.md"):
        source = project / name
        if not source.exists():
            continue
        docs.mkdir(parents=True, exist_ok=True)
        target = docs / f"{starter_id}-{name}"
        if not target.exists():
            source.rename(target)


def materialize_project(
    manifest: dict[str, Any],
    project: Path,
    *,
    skip_setup: bool = False,
    skip_checks: bool = False,
) -> dict[str, Any]:
    starter_records: list[dict[str, Any]] = []
    setup_records: list[dict[str, Any]] = []
    check_records: list[dict[str, Any]] = []
    for starter in manifest["starters"]:
        adapter_path = starter.get("adapter")
        if not adapter_path or not starter.get("pin"):
            raise PlatformError(f"{starter['id']} no tiene adapter y pin materializables")
        adapter = read_json(ROOT / adapter_path)
        materializer = adapter.get("materializer", {})
        kind = materializer.get("type")
        destination_relative = _safe_relative(materializer.get("destination", "."), "destination")
        destination = project / destination_relative
        destination.mkdir(parents=True, exist_ok=True)
        source = adapter.get("source", {})
        generator_record: dict[str, Any] | None = None
        if kind == "git-copy":
            repository = source.get("repository")
            commit = source.get("commit")
            if not repository or not commit:
                raise PlatformError(f"{starter['id']} necesita repository y commit exacto")
            with tempfile.TemporaryDirectory(prefix=f"eng-{starter['id']}-") as temporary:
                clone = Path(temporary) / "source"
                clone.mkdir()
                _run_record(["git", "init", "--quiet"], clone)
                _run_record(["git", "remote", "add", "origin", repository], clone)
                _run_record(["git", "fetch", "--quiet", "--depth", "1", "origin", commit], clone)
                _run_record(["git", "checkout", "--quiet", "FETCH_HEAD"], clone)
                _copy_materialized_tree(clone, destination)
        elif kind == "local-copy":
            local_path = _safe_relative(source.get("local_path", ""), "source.local_path")
            _copy_materialized_tree(ROOT / local_path, destination)
        elif kind == "command-generator":
            working = project / _safe_relative(materializer.get("working_directory", "."), "working_directory")
            working.mkdir(parents=True, exist_ok=True)
            generator_record = _run_record(materializer.get("command", []), working)
            generator_record["workdir"] = working.relative_to(project).as_posix() or "."
            if not destination.exists() or not any(destination.iterdir()):
                raise PlatformError(f"El generador de {starter['id']} no creó {destination_relative}")
        else:
            raise PlatformError(f"Materializer desconocido para {starter['id']}: {kind}")
        if destination_relative == Path("."):
            _preserve_upstream_instructions(project, starter["id"])
        for command in materializer.get("setup", []):
            if skip_setup:
                setup_records.append({"starter": starter["id"], "command": command, "workdir": destination_relative.as_posix(), "status": "skipped"})
            else:
                record = _run_record(command, destination)
                record.update({"starter": starter["id"], "workdir": destination_relative.as_posix()})
                setup_records.append(record)
        for check in materializer.get("checks", []):
            if skip_checks:
                check_records.append({"starter": starter["id"], "gate": check.get("gate"), "command": check.get("command", []), "workdir": destination_relative.as_posix(), "status": "skipped"})
            else:
                record = _run_record(check.get("command", []), destination, gate=check.get("gate"))
                record.update({"starter": starter["id"], "workdir": destination_relative.as_posix()})
                check_records.append(record)
        starter_records.append({
            "id": starter["id"], "type": kind, "destination": destination_relative.as_posix(),
            "repository": source.get("repository"), "branch": source.get("branch"),
            "commit": source.get("commit"), "content_sha256": _content_sha256(destination),
            "generator": generator_record,
        })
    actual = _actual_project_data(project)
    readiness = "verified" if not skip_checks and check_records else "code-ready"
    return {
        "schema_version": 1,
        "platform_version": PLATFORM_VERSION,
        "generated_at": date.today().isoformat(),
        "readiness": readiness,
        "starters": starter_records,
        "setup": setup_records,
        "checks": check_records,
        **actual,
    }


def _initialize_seed_repository(project: Path, materialization: dict[str, Any]) -> None:
    if not (project / ".git").exists():
        _run_record(["git", "init", "--quiet"], project)
    existing = subprocess.run(["git", "remote"], cwd=project, text=True, capture_output=True, check=False).stdout.split()
    for starter in materialization.get("starters", []):
        repository = starter.get("repository")
        remote = f"seed-{starter['id']}"
        if repository and remote not in existing:
            _run_record(["git", "remote", "add", remote, repository], project)


def gentle_handoff_data(
    manifest: dict[str, Any], definition: dict[str, Any] | None
) -> dict[str, Any]:
    idea = (definition or {}).get("idea", {})
    delivery = (definition or {}).get("delivery", {})
    patterns = ["modular-monolith", "explicit-contracts", "least-privilege", "incremental-delivery"]
    patterns.append("tenant-isolation" if "multitenancy" in manifest["features"] else "single-tenant-first")
    materialization = manifest.get("materialization", {})
    structure = dict(manifest["ownership"])
    structure.update({
        "actual_tree": materialization.get("actual_tree", []),
        "packages": materialization.get("packages", []),
        "environment_files": materialization.get("environment_files", []),
    })
    return {
        "schema_version": 1,
        "platform_version": manifest["platform_version"],
        "scaffold_status": manifest["scaffold_status"],
        "readiness": manifest.get("readiness", "code-ready"),
        "project": manifest["project"],
        "idea": idea,
        "stack": {
            "recipe": manifest["recipe"],
            "starters": manifest["starters"],
            "database": manifest["database"],
            "features": manifest["features"],
            "skills": manifest["skills"],
            "sources": materialization.get("starters", []),
        },
        "structure": structure,
        "verification": {
            "setup": materialization.get("setup", []),
            "checks": materialization.get("checks", []),
        },
        "patterns": patterns,
        "delivery": delivery,
        "quality_gates": manifest["gates"],
        "exclusions": manifest["exclusions"],
        "strategy": {
            "owner": "gentle-ai",
            "allowed": ["direct", "sdd"],
            "instruction": "Gentle elige direct o SDD según riesgo, ambigüedad y alcance; debe registrar la elección antes de implementar.",
            "prefer_sdd_when": [
                "hay contratos entre aplicaciones",
                "existen migraciones o permisos sensibles",
                "persisten incógnitas de producto de alto impacto",
            ],
        },
        "read_first": (
            [".engineering/project-definition.json"] if definition else []
        )
        + [".engineering/project.json", "ARCHITECTURE.md", "AGENTS.md"],
    }


def gentle_markdown(handoff: dict[str, Any]) -> str:
    idea = handoff.get("idea", {})
    stack = handoff["stack"]
    starters = ", ".join(f"`{item['id']}`" for item in stack["starters"])
    must_have = "\n".join(f"- {item}" for item in idea.get("must_have", [])) or "- No definido."
    out_of_scope = "\n".join(f"- {item}" for item in idea.get("out_of_scope", [])) or "- Ninguno adicional."
    acceptance = "\n".join(
        f"- {item}" for item in handoff.get("delivery", {}).get("acceptance_criteria", [])
    ) or "- Completar los quality gates aplicables."
    structure = handoff["structure"]
    sources = "\n".join(
        f"- `{item['id']}` → `{item['destination']}` @ `{item['commit']}`"
        + (f" ({item['repository']})" if item.get("repository") else " (starter interno)")
        for item in stack.get("sources", [])
    ) or "- Aún no materializado."
    packages = "\n".join(
        f"- `{item['path']}`: `{item.get('package_manager', 'desconocido')}`; scripts: "
        + (", ".join(f"`{script}`" for script in item.get("scripts", [])) or "ninguno")
        for item in structure.get("packages", [])
    ) or "- No detectados."
    actual_tree = "\n".join(f"- `{item}`" for item in structure.get("actual_tree", [])[:80]) or "- Aún no detectada."
    checks = "\n".join(
        f"- `{item.get('gate', 'setup')}` en `{item.get('workdir', '.')}`: "
        f"`{' '.join(item.get('command', []))}` ({item.get('status', 'passed')})"
        for item in handoff.get("verification", {}).get("checks", [])
    ) or "- Pendientes de registrar."
    read_first = ", ".join(f"`{item}`" for item in handoff["read_first"])
    return f"""# Handoff a Gentle AI

## Idea confirmada

{idea.get('summary', handoff['project'].get('summary', 'Sin resumen.'))}

Problema: {idea.get('problem', 'Consulta la definición del proyecto.')}

## Stack y base

- Recipe: `{stack['recipe']['id']}@{stack['recipe']['version']}`
- Boilerplates: {starters or 'ninguno'}
- Base de datos: `{stack['database'] or 'ninguna'}`
- Features: {', '.join(f'`{item}`' for item in stack['features']) or 'ninguno'}
- Skills: {', '.join(f'`{item}`' for item in stack['skills']) or 'ninguno'}
- Patrones: {', '.join(f'`{item}`' for item in handoff['patterns'])}
- Estado: `{handoff['scaffold_status']}`
- Readiness: `{handoff['readiness']}`

## Fuentes exactas

{sources}

## Estructura y ownership

- Gestionado por la plataforma: {', '.join(f'`{item}`' for item in structure['managed'])}
- Secciones gestionadas: {', '.join(f'`{item}`' for item in structure['managed_sections'])}
- Código propiedad del proyecto: {', '.join(f'`{item}`' for item in structure['user_owned'])}

### Paquetes detectados

{packages}

### Árbol real

{actual_tree}

### Verificación

{checks}

## Alcance mínimo

{must_have}

## Fuera de alcance

{out_of_scope}

## Criterios de aceptación

{acceptance}

## Instrucciones de ejecución

1. Lee, en orden: {read_first}.
2. Decide entre ejecución directa y SDD según riesgo, ambigüedad, contratos, datos y permisos; registra brevemente el motivo.
3. Conserva la Recipe, el stack, los patrones y las exclusiones. Propón una decisión explícita antes de desviarte.
4. Implementa el incremento vertical mínimo y ejecuta los quality gates indicados en `.engineering/project.json`.
5. Si readiness es `code-ready`, ejecuta `eng check --run` antes de tratar el proyecto como verificado.
"""


def write_handoff(
    project: Path,
    manifest: dict[str, Any],
    definition: dict[str, Any] | None,
) -> dict[str, Any]:
    handoff = gentle_handoff_data(manifest, definition)
    (project / ".engineering/gentle-handoff.json").write_text(
        dump_json(handoff), encoding="utf-8"
    )
    (project / "GENTLE.md").write_text(gentle_markdown(handoff), encoding="utf-8")
    return handoff


def write_project(
    intake: dict[str, Any],
    output: Path,
    *,
    definition: dict[str, Any] | None = None,
    allowed_existing: set[Path] | None = None,
    materialize: bool = False,
    skip_setup: bool = False,
    skip_checks: bool = False,
) -> dict[str, Any]:
    manifest = resolve_recipe(intake)
    _ensure_target_is_safe(output, allowed_existing)
    if definition:
        validate_project_definition(definition)
        manifest["project"]["summary"] = definition["idea"]["summary"]
        manifest["definition_status"] = "confirmed"
    if definition:
        manifest["ownership"]["managed"].append(".engineering/project-definition.json")
    manifest["ownership"]["managed"].extend(
        [".engineering/gentle-handoff.json", "GENTLE.md"]
    )
    if materialize:
        manifest["ownership"]["managed"].append(".engineering/materialization.json")
    output.parent.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory(prefix=f".{output.name}-eng-", dir=output.parent) as temporary:
        staged = Path(temporary) / output.name
        if output.exists():
            shutil.copytree(output, staged, symlinks=True)
        else:
            staged.mkdir()
        if materialize:
            materialization = materialize_project(
                manifest, staged, skip_setup=skip_setup, skip_checks=skip_checks
            )
            manifest["scaffold_status"] = "materialized"
            manifest["readiness"] = materialization["readiness"]
            manifest["materialization"] = materialization
            _initialize_seed_repository(staged, materialization)
        engineering = staged / ".engineering"
        engineering.mkdir(parents=True, exist_ok=True)
        if definition:
            (engineering / "project-definition.json").write_text(
                dump_json(definition), encoding="utf-8"
            )
        (engineering / "intake.json").write_text(dump_json(intake), encoding="utf-8")
        (staged / "ARCHITECTURE.md").write_text(architecture_markdown(manifest), encoding="utf-8")
        (staged / "AGENTS.md").write_text(agents_markdown(manifest), encoding="utf-8")
        (staged / "README.md").write_text(readme_markdown(manifest), encoding="utf-8")
        if materialize:
            manifest["materialization"].update(_actual_project_data(staged))
            (engineering / "materialization.json").write_text(
                dump_json(manifest["materialization"]), encoding="utf-8"
            )
        (engineering / "project.json").write_text(dump_json(manifest), encoding="utf-8")
        write_handoff(staged, manifest, definition)
        if output.exists():
            backup = Path(temporary) / f"{output.name}.existing"
            output.rename(backup)
            try:
                staged.rename(output)
            except Exception:
                backup.rename(output)
                raise
        else:
            staged.rename(output)
    return manifest


def inspect_project(project: Path) -> tuple[dict[str, Any], list[str], list[str]]:
    manifest_path = project / ".engineering/project.json"
    manifest = read_json(manifest_path)
    platform = data()
    boilerplates = by_id(platform["boilerplates"]["entries"])
    recipes = by_id(platform["recipes"]["paths"])
    features = by_id(platform["features"]["features"])
    databases = by_id(platform["databases"]["profiles"])
    skills = by_id(platform["skills"]["skills"])
    errors: list[str] = []
    warnings: list[str] = []
    if manifest.get("schema_version") != 2:
        errors.append(f"schema_version no soportado: {manifest.get('schema_version')}")
    if manifest.get("platform_version") != PLATFORM_VERSION:
        warnings.append(
            f"El proyecto usa platform {manifest.get('platform_version')}; actual es {PLATFORM_VERSION}"
        )
    recipe = recipes.get(manifest.get("recipe", {}).get("id"))
    if not recipe:
        errors.append("Recipe inexistente")
    elif manifest.get("recipe", {}).get("version") != recipe.get("version"):
        warnings.append(
            f"Recipe instalada {manifest.get('recipe', {}).get('version')} != actual {recipe.get('version')}"
        )
    for starter in manifest.get("starters", []):
        entry = boilerplates.get(starter.get("id"))
        if not entry:
            errors.append(f"Boilerplate inexistente: {starter.get('id')}")
        elif entry["delivery_status"] not in {"curated", "released"}:
            warnings.append(f"{entry['id']} continúa {entry['delivery_status']}")
            if manifest.get("scaffold_status") == "materialized":
                errors.append(
                    f"{entry['id']} no puede constar materialized mientras está {entry['delivery_status']}"
                )
        if starter.get("pin") == "PINNED":
            errors.append(f"{starter.get('id')} conserva el marcador PINNED")
        if manifest.get("scaffold_status") == "materialized" and not starter.get("pin"):
            errors.append(f"{starter.get('id')} materializado sin pin reproducible")
        if manifest.get("scaffold_status") == "materialized" and not starter.get("adapter"):
            errors.append(f"{starter.get('id')} materializado sin adapter")
    if manifest.get("scaffold_status") == "materialized":
        materialization_path = project / ".engineering/materialization.json"
        if not materialization_path.exists():
            errors.append("scaffold materialized sin materialization.json")
        elif manifest.get("readiness") not in {"code-ready", "verified"}:
            errors.append(f"readiness inválido: {manifest.get('readiness')}")
    if manifest.get("database") is not None and manifest.get("database") not in databases:
        errors.append(f"Perfil de base de datos inexistente: {manifest.get('database')}")
    if recipe and manifest.get("database") not in recipe.get("stack", {}).get("database", {}).get("allowed", []):
        if manifest.get("database") is not None:
            errors.append(f"Base de datos fuera de la Recipe: {manifest.get('database')}")
    for feature in manifest.get("features", []):
        if feature not in features:
            errors.append(f"Feature inexistente: {feature}")
    installed_features = set(manifest.get("features", []))
    if recipe:
        allowed_features: set[str] = set()
        pending_features = list(recipe.get("features", {}).get("default", [])) + list(
            recipe.get("features", {}).get("optional", [])
        )
        while pending_features:
            candidate = pending_features.pop()
            if candidate in allowed_features or candidate not in features:
                continue
            allowed_features.add(candidate)
            pending_features.extend(
                requirement
                for requirement in features[candidate].get("requires", [])
                if requirement != "database"
            )
        for feature_id in installed_features.difference(allowed_features):
            errors.append(f"Feature fuera de la Recipe: {feature_id}")
    for feature_id in installed_features.intersection(features):
        for requirement in features[feature_id].get("requires", []):
            if requirement == "database" and not manifest.get("database"):
                errors.append(f"{feature_id} requiere base de datos")
            elif requirement != "database" and requirement not in installed_features:
                errors.append(f"{feature_id} requiere feature {requirement}")
    for skill in manifest.get("skills", []):
        if skill not in skills:
            errors.append(f"Skill inexistente: {skill}")
    for required in ["ARCHITECTURE.md", "AGENTS.md", "GENTLE.md", ".engineering/gentle-handoff.json"]:
        if not (project / required).exists():
            if manifest.get("platform_version") == PLATFORM_VERSION:
                errors.append(f"Falta {required}")
            else:
                warnings.append(f"Falta {required}")
    definition_path = project / ".engineering/project-definition.json"
    if manifest.get("definition_status") == "confirmed":
        if not definition_path.exists():
            errors.append("definition_status=confirmed pero falta project-definition.json")
        else:
            try:
                definition = validate_project_definition(read_json(definition_path))
                if definition["intake"].get("name") != manifest.get("project", {}).get("name"):
                    errors.append("La definición y el manifest tienen nombres de proyecto distintos")
            except PlatformError as exc:
                errors.append(f"Definición inválida: {exc}")
    handoff_path = project / ".engineering/gentle-handoff.json"
    if handoff_path.exists():
        handoff = read_json(handoff_path)
        if handoff.get("platform_version") != manifest.get("platform_version"):
            errors.append("El handoff de Gentle y el manifest usan versiones distintas")
        if handoff.get("stack", {}).get("recipe", {}).get("id") != manifest.get("recipe", {}).get("id"):
            errors.append("El handoff de Gentle y el manifest usan Recipes distintas")
        if handoff.get("stack", {}).get("skills") != manifest.get("skills"):
            errors.append("El handoff de Gentle y el manifest declaran skills distintos")
        if handoff.get("strategy", {}).get("owner") != "gentle-ai":
            errors.append("El handoff no delega la estrategia de desarrollo a Gentle AI")
    return manifest, errors, warnings


def add_feature_to_project(project: Path, feature_id: str, *, apply: bool = False) -> dict[str, Any]:
    engineering = project / ".engineering"
    intake_path = engineering / "intake.json"
    current_manifest, errors, warnings = inspect_project(project)
    if errors:
        raise PlatformError("El proyecto no pasa doctor: " + "; ".join(errors))
    intake = read_json(intake_path)
    if feature_id in current_manifest.get("features", []):
        return {
            "changed": False,
            "feature": feature_id,
            "reason": "El feature ya está instalado directa o transitivamente.",
            "warnings": warnings,
        }
    if feature_id in intake.get("excluded_features", []):
        raise PlatformError(
            f"{feature_id} está excluido explícitamente; actualiza la decisión y el ADR antes de agregarlo"
        )
    updated_intake = deepcopy(intake)
    requested = list(updated_intake.get("features", []))
    requested.append(feature_id)
    updated_intake["features"] = requested
    updated_manifest = resolve_recipe(updated_intake)
    definition_path = engineering / "project-definition.json"
    definition = read_json(definition_path) if definition_path.exists() else None
    if definition:
        definition = deepcopy(definition)
        definition["intake"] = updated_intake
        validate_project_definition(definition)
        updated_manifest["project"]["summary"] = definition["idea"]["summary"]
        updated_manifest["definition_status"] = "confirmed"
        updated_manifest["ownership"]["managed"].append(
            ".engineering/project-definition.json"
        )
    updated_manifest["ownership"]["managed"].extend(
        [".engineering/gentle-handoff.json", "GENTLE.md"]
    )
    if current_manifest.get("scaffold_status") == "materialized":
        updated_manifest["scaffold_status"] = "materialized"
        updated_manifest["readiness"] = current_manifest.get("readiness", "code-ready")
        updated_manifest["materialization"] = current_manifest.get("materialization", {})
        updated_manifest["ownership"]["managed"].append(".engineering/materialization.json")
    added = [
        item
        for item in updated_manifest["features"]
        if item not in current_manifest.get("features", [])
    ]
    result = {
        "changed": True,
        "applied": apply,
        "feature": feature_id,
        "added_with_dependencies": added,
        "new_gates": sorted(
            set(updated_manifest["gates"]).difference(current_manifest.get("gates", []))
        ),
        "warnings": updated_manifest["warnings"],
    }
    if apply:
        if updated_manifest.get("scaffold_status") == "materialized":
            updated_manifest["materialization"].update(_actual_project_data(project))
            (engineering / "materialization.json").write_text(
                dump_json(updated_manifest["materialization"]), encoding="utf-8"
            )
        intake_path.write_text(dump_json(updated_intake), encoding="utf-8")
        if definition:
            definition_path.write_text(dump_json(definition), encoding="utf-8")
        (engineering / "project.json").write_text(dump_json(updated_manifest), encoding="utf-8")
        (project / "ARCHITECTURE.md").write_text(
            architecture_markdown(updated_manifest), encoding="utf-8"
        )
        (project / "AGENTS.md").write_text(
            agents_markdown(updated_manifest), encoding="utf-8"
        )
        write_handoff(project, updated_manifest, definition)
        result["handoff_status"] = "handoff-updated-code-change-required"
    return result


CHANGE_PLANS = {
    "api": {"skills": ["contracts", "authorization"], "gates": ["contract", "integration", "security"]},
    "schema": {"skills": ["database"], "gates": ["migration", "test", "backup-restore"]},
    "permission": {"skills": ["authorization", "security-review"], "gates": ["test", "security"]},
    "ui": {"skills": ["gate-runner"], "gates": ["lint", "typecheck", "test", "accessibility"]},
    "upgrade": {"skills": ["react-starter-kit-updater", "gate-runner"], "gates": ["lint", "typecheck", "test", "integration", "build", "security"]},
    "incident": {"skills": ["project-doctor", "knowledge-capture"], "gates": ["test", "integration"]},
}


def change_plan(manifest: dict[str, Any], change_type: str) -> dict[str, Any]:
    if change_type not in CHANGE_PLANS:
        raise PlatformError(f"Tipo de cambio desconocido: {change_type}")
    specific = CHANGE_PLANS[change_type]
    available = set(manifest["skills"])
    selected_skills = [item for item in specific["skills"] if item in available]
    missing_skills = [item for item in specific["skills"] if item not in available]
    return {
        "change_type": change_type,
        "recipe": manifest["recipe"],
        "skills": selected_skills,
        "missing_skills": missing_skills,
        "gates": sorted(set(specific["gates"]).intersection(manifest["gates"]) or set(specific["gates"])),
        "steps": [
            "Confirmar alcance y contrato observable.",
            "Implementar el cambio mínimo respetando ownership.",
            "Ejecutar los gates seleccionados.",
            "Registrar ADR o knowledge entry si cambia una decisión o resuelve un incidente repetible.",
        ],
    }


def command_catalog(args: argparse.Namespace) -> int:
    entries = data()["boilerplates"]["entries"]
    filtered = [
        item
        for item in entries
        if (not args.category or item["category"] == args.category)
        and (not args.decision_status or item["decision_status"] == args.decision_status)
        and (not args.delivery_status or item["delivery_status"] == args.delivery_status)
    ]
    print(dump_json(filtered), end="")
    return 0


def command_evaluate(args: argparse.Namespace) -> int:
    print(
        dump_json(
            evaluate_boilerplate(
                args.repository,
                observed_commit=args.observed_commit,
                observed_date=args.observed_date,
                category=args.category,
            )
        ),
        end="",
    )
    return 0


def command_recommend(args: argparse.Namespace) -> int:
    intake = read_json(Path(args.input))
    print(dump_json(resolve_recipe(intake)), end="")
    return 0


def command_new(args: argparse.Namespace) -> int:
    intake = read_json(Path(args.input))
    manifest = resolve_recipe(intake)
    if args.dry_run:
        print(dump_json(manifest), end="")
        return 0
    output = Path(args.output).resolve()
    result = write_project(intake, output)
    print(f"Proyecto {result['scaffold_status']} creado en {output}")
    for warning in result["warnings"]:
        print(f"ADVERTENCIA: {warning}")
    return 0


def command_bootstrap(args: argparse.Namespace) -> int:
    definition_path = Path(args.input).resolve()
    definition = validate_project_definition(read_json(definition_path))
    output = Path(args.output).resolve()
    try:
        definition_path.relative_to(output)
        allowed = {definition_path}
    except ValueError:
        allowed = set()
    if args.dry_run:
        manifest = resolve_recipe(definition["intake"])
        manifest["project"]["summary"] = definition["idea"]["summary"]
        print(dump_json(manifest), end="")
        return 0
    result = write_project(
        definition["intake"],
        output,
        definition=definition,
        allowed_existing=allowed,
        materialize=True,
        skip_setup=args.skip_setup,
        skip_checks=args.skip_checks,
    )
    print(f"Proyecto {result['scaffold_status']} creado en {output}")
    print(f"Handoff para Gentle generado en {output / 'GENTLE.md'}")
    for warning in result["warnings"]:
        print(f"ADVERTENCIA: {warning}")
    return 0


def command_handoff(args: argparse.Namespace) -> int:
    project = Path(args.project).resolve()
    manifest, errors, warnings = inspect_project(project)
    if errors:
        raise PlatformError("El proyecto no pasa doctor: " + "; ".join(errors))
    definition_path = project / ".engineering/project-definition.json"
    definition = validate_project_definition(read_json(definition_path)) if definition_path.exists() else None
    if manifest.get("scaffold_status") == "materialized":
        manifest.setdefault("materialization", {}).update(_actual_project_data(project))
        (project / ".engineering/materialization.json").write_text(
            dump_json(manifest["materialization"]), encoding="utf-8"
        )
        (project / ".engineering/project.json").write_text(dump_json(manifest), encoding="utf-8")
    write_handoff(project, manifest, definition)
    print(
        dump_json(
            {
                "ok": True,
                "handoff": str(project / "GENTLE.md"),
                "definition": "confirmed" if definition else "missing",
                "warnings": warnings,
            }
        ),
        end="",
    )
    return 0


def _managed_launcher_target(launcher: Path, install_root: Path) -> Path | None:
    if not launcher.is_symlink():
        return None
    managed_root = install_root.parent
    if managed_root.is_symlink():
        return None
    try:
        target = launcher.resolve()
    except (OSError, RuntimeError):
        return None
    managed_root = managed_root.resolve()
    if target.name != "eng" or target.parent.parent != managed_root:
        return None
    return target


def _managed_installations(home: Path) -> list[Path]:
    managed_root = home / ".local/share/engineering-platform"
    if managed_root.is_symlink() or not managed_root.is_dir():
        return []
    try:
        candidates = list(managed_root.iterdir())
    except OSError as exc:
        raise PlatformError(f"No se pudo leer {managed_root}: {exc}") from exc
    installations: list[Path] = []
    for candidate in sorted(candidates, key=lambda item: item.name):
        if candidate.is_symlink() or not candidate.is_dir():
            continue
        if not re.fullmatch(r"\d+\.\d+\.\d+", candidate.name):
            continue
        try:
            package = read_json(candidate / "package.json")
        except PlatformError:
            continue
        if package.get("name") == "engineering-platform" and package.get("version") == candidate.name:
            installations.append(candidate)
    return installations


def _retire_stale_managed_installations(
    pi_executable: str, home: Path, install_root: Path
) -> list[Path]:
    stale_installations = [
        installation
        for installation in _managed_installations(home)
        if installation != install_root
    ]
    for installation in stale_installations:
        completed = subprocess.run(
            [pi_executable, "remove", str(installation)],
            text=True,
            capture_output=True,
            check=False,
        )
        if completed.returncode != 0:
            detail = completed.stderr.strip() or completed.stdout.strip() or "sin salida"
            raise PlatformError(
                f"Pi no pudo retirar {installation}; no se borraron archivos: {detail[-1000:]}"
            )
    for installation in stale_installations:
        shutil.rmtree(installation)
    return stale_installations


def _global_install_status(home: Path) -> dict[str, Any]:
    install_root = home / ".local/share/engineering-platform" / PLATFORM_VERSION
    launcher = home / ".local/bin/eng"
    pi_executable = shutil.which("pi")
    launcher_ok = launcher.is_symlink() and launcher.resolve() == (install_root / "eng").resolve()
    pi_registered = False
    if pi_executable:
        listed = subprocess.run(
            [pi_executable, "list"], text=True, capture_output=True, check=False
        )
        package_list = listed.stdout + listed.stderr
        pi_registered = listed.returncode == 0 and (
            str(install_root) in package_list or "engineering-platform" in package_list
        )
    return {
        "ok": bool(pi_executable and install_root.exists() and launcher_ok and pi_registered),
        "pi": pi_executable,
        "pi_registered": pi_registered,
        "package": str(install_root),
        "package_exists": install_root.exists(),
        "launcher": str(launcher),
        "launcher_ok": launcher_ok,
        "path_hint": None if str(launcher.parent) in os.environ.get("PATH", "").split(os.pathsep) else str(launcher.parent),
    }


def command_install(args: argparse.Namespace) -> int:
    if args.target != "pi":
        raise PlatformError("Engineering Platform instala únicamente el target global pi")
    home = Path(args.home).expanduser().resolve() if args.home else Path.home()
    install_root = home / ".local/share/engineering-platform" / PLATFORM_VERSION
    launcher = home / ".local/bin/eng"
    pi_executable = shutil.which("pi")
    result = {
        "target": "pi",
        "source": str(ROOT),
        "package": str(install_root),
        "launcher": str(launcher),
        "pi_command": [pi_executable or "pi", "install", str(install_root)],
    }
    if args.dry_run:
        result["dry_run"] = True
        print(dump_json(result), end="")
        return 0
    if not pi_executable:
        raise PlatformError("No se encontró `pi` en PATH; instala Pi antes de integrar la plataforma")
    if install_root.exists() and not args.force:
        current_version = read_json(install_root / "package.json").get("version")
        if current_version != PLATFORM_VERSION:
            raise PlatformError(f"Existe una instalación incompatible en {install_root}")
    elif install_root.exists():
        shutil.rmtree(install_root)
    if not install_root.exists():
        install_root.parent.mkdir(parents=True, exist_ok=True)
        with tempfile.TemporaryDirectory(prefix="engineering-platform-", dir=install_root.parent) as temporary:
            staged = Path(temporary) / PLATFORM_VERSION
            shutil.copytree(ROOT, staged, ignore=INSTALL_IGNORES)
            staged.rename(install_root)
    launcher.parent.mkdir(parents=True, exist_ok=True)
    expected_launcher = (install_root / "eng").resolve()
    if launcher.exists() or launcher.is_symlink():
        current_launcher = _managed_launcher_target(launcher, install_root)
        if current_launcher is None:
            raise PlatformError(f"No se sobrescribirá el launcher existente: {launcher}")
        if current_launcher != expected_launcher:
            launcher.unlink()
            launcher.symlink_to(install_root / "eng")
    else:
        launcher.symlink_to(install_root / "eng")
    completed = subprocess.run(
        [pi_executable, "install", str(install_root)],
        text=True,
        capture_output=True,
        check=False,
    )
    if completed.returncode != 0:
        raise PlatformError(
            "Pi no pudo registrar el paquete: " + (completed.stderr.strip() or completed.stdout.strip())
        )
    retired_installations = _retire_stale_managed_installations(pi_executable, home, install_root)
    status = _global_install_status(home)
    status["pi_output"] = completed.stdout.strip()
    if retired_installations:
        status["retired_packages"] = [str(path) for path in retired_installations]
    print(dump_json(status), end="")
    return 0


def command_uninstall(args: argparse.Namespace) -> int:
    home = Path(args.home).expanduser().resolve() if args.home else Path.home()
    install_root = home / ".local/share/engineering-platform" / PLATFORM_VERSION
    launcher = home / ".local/bin/eng"
    managed_root = install_root.parent
    installations = _managed_installations(home)
    managed_launcher = _managed_launcher_target(launcher, install_root)
    result = {
        "target": "pi",
        "package": str(install_root),
        "packages": [str(path) for path in installations],
        "launcher": str(launcher),
        "removed": False,
        "launcher_removed": False,
    }
    if args.dry_run:
        result["dry_run"] = True
        result["launcher_managed"] = managed_launcher is not None
        print(dump_json(result), end="")
        return 0
    if not installations:
        if managed_launcher is not None:
            launcher.unlink()
            result["launcher_removed"] = True
            result["removed"] = True
        if managed_root.is_dir() and not managed_root.is_symlink():
            try:
                managed_root.rmdir()
            except OSError:
                pass
        print(dump_json(result), end="")
        return 0
    pi_executable = shutil.which("pi")
    if not pi_executable:
        raise PlatformError("No se encontró `pi` en PATH; no se eliminará una instalación parcialmente registrada")
    pi_outputs: list[str] = []
    for installation in installations:
        completed = subprocess.run(
            [pi_executable, "remove", str(installation)],
            text=True,
            capture_output=True,
            check=False,
        )
        if completed.returncode != 0:
            raise PlatformError(
                f"Pi no pudo retirar {installation}; no se borraron archivos: "
                + (completed.stderr.strip() or completed.stdout.strip())
            )
        if completed.stdout.strip():
            pi_outputs.append(completed.stdout.strip())
    for installation in installations:
        shutil.rmtree(installation)
    if managed_launcher is not None:
        launcher.unlink()
        result["launcher_removed"] = True
    if managed_root.is_dir() and not managed_root.is_symlink():
        try:
            managed_root.rmdir()
        except OSError:
            pass
    result["packages_removed"] = [str(path) for path in installations]
    if pi_outputs:
        result["pi_output"] = "\n".join(pi_outputs)
    result["removed"] = True
    print(dump_json(result), end="")
    return 0


def command_start(args: argparse.Namespace) -> int:
    name = args.name.strip()
    if not re.fullmatch(r"[a-z0-9]+(?:-[a-z0-9]+)*", name):
        raise PlatformError("name debe usar kebab-case, por ejemplo school-requests")
    workspace_value = args.workspace or os.environ.get("ENG_WORKSPACE") or str(Path.home() / "dev")
    workspace = Path(workspace_value).expanduser().resolve()
    target = workspace / name
    if target.parent != workspace:
        raise PlatformError("El nombre del proyecto no puede cambiar el workspace")
    _ensure_target_is_safe(target, require_managed_metadata=True)
    initial_prompt = (
        "Inicia un proyecto nuevo con Engineering Platform. Usa la skill project-discovery, "
        "haz preguntas progresivas hasta confirmar la idea y luego ejecuta el bootstrap en este directorio."
    )
    result = {
        "workspace": str(workspace),
        "target": str(target),
        "command": ["pi", "--name", f"new:{name}", initial_prompt],
    }
    if args.dry_run:
        result["dry_run"] = True
        print(dump_json(result), end="")
        return 0
    pi_executable = shutil.which("pi")
    if not pi_executable:
        raise PlatformError("No se encontró `pi` en PATH")
    workspace.mkdir(parents=True, exist_ok=True)
    target.mkdir(exist_ok=True)
    os.chdir(target)
    os.execv(pi_executable, [pi_executable, "--name", f"new:{name}", initial_prompt])
    return 0


def command_doctor(args: argparse.Namespace) -> int:
    if args.global_install:
        home = Path(args.home).expanduser().resolve() if args.home else Path.home()
        result = _global_install_status(home)
        print(dump_json(result), end="")
        return 0 if result["ok"] else 1
    manifest, errors, warnings = inspect_project(Path(args.project).resolve())
    result = {
        "ok": not errors,
        "project": manifest.get("project", {}).get("name"),
        "errors": errors,
        "warnings": warnings,
    }
    print(dump_json(result), end="")
    return 1 if errors else 0


def command_plan(args: argparse.Namespace) -> int:
    manifest, errors, _ = inspect_project(Path(args.project).resolve())
    if errors:
        raise PlatformError("El proyecto no pasa doctor: " + "; ".join(errors))
    print(dump_json(change_plan(manifest, args.change_type)), end="")
    return 0


def command_check(args: argparse.Namespace) -> int:
    manifest, errors, warnings = inspect_project(Path(args.project).resolve())
    files = args.changed_files or []
    gates = set(manifest.get("gates", []))
    if files:
        selected = {"lint", "typecheck", "test"}
        if any("migrations/" in item or "schema" in item for item in files):
            selected.update({"migration", "integration"})
        if any("auth" in item or "permission" in item for item in files):
            selected.add("security")
        if any(item.endswith((".tsx", ".jsx", ".html", ".css")) for item in files):
            selected.add("accessibility")
        gates.intersection_update(selected)
    if args.run:
        if manifest.get("scaffold_status") != "materialized":
            raise PlatformError("El proyecto es blueprint; ejecuta eng bootstrap para materializar código")
        materialization_path = Path(args.project).resolve() / ".engineering/materialization.json"
        materialization = read_json(materialization_path)
        executed_setup: list[dict[str, Any]] = []
        executed_checks: list[dict[str, Any]] = []
        project = Path(args.project).resolve()
        for registered in materialization.get("setup", []):
            record = _run_record(registered["command"], project / _safe_relative(registered.get("workdir", "."), "workdir"))
            record.update({"starter": registered.get("starter"), "workdir": registered.get("workdir", ".")})
            executed_setup.append(record)
        for registered in materialization.get("checks", []):
            gate = registered.get("gate")
            if files and gate not in gates:
                continue
            record = _run_record(
                registered["command"],
                project / _safe_relative(registered.get("workdir", "."), "workdir"),
                gate=gate,
            )
            record.update({"starter": registered.get("starter"), "workdir": registered.get("workdir", ".")})
            executed_checks.append(record)
        materialization["setup"] = executed_setup
        materialization["checks"] = executed_checks
        materialization.update(_actual_project_data(project))
        if not files and executed_checks:
            materialization["readiness"] = "verified"
            manifest["readiness"] = "verified"
        manifest["materialization"] = materialization
        materialization_path.write_text(dump_json(materialization), encoding="utf-8")
        (project / ".engineering/project.json").write_text(dump_json(manifest), encoding="utf-8")
        definition_path = project / ".engineering/project-definition.json"
        definition = validate_project_definition(read_json(definition_path)) if definition_path.exists() else None
        write_handoff(project, manifest, definition)
        result = {
            "ok": True,
            "mode": "executed",
            "readiness": manifest.get("readiness"),
            "selected_gates": sorted(gates),
            "setup": executed_setup,
            "checks": executed_checks,
            "errors": [],
            "warnings": warnings,
        }
        print(dump_json(result), end="")
        return 0
    result = {
        "ok": not errors,
        "mode": "selection-only",
        "selected_gates": sorted(gates),
        "errors": errors,
        "warnings": warnings + ["Este comando selecciona gates; el starter materializado debe mapearlos a comandos reproducibles."],
    }
    print(dump_json(result), end="")
    return 1 if errors else 0


def command_add(args: argparse.Namespace) -> int:
    result = add_feature_to_project(
        Path(args.project).resolve(), args.feature, apply=args.apply
    )
    print(dump_json(result), end="")
    return 0


def command_update(args: argparse.Namespace) -> int:
    manifest, errors, warnings = inspect_project(Path(args.project).resolve())
    if errors:
        raise PlatformError("El proyecto no pasa doctor: " + "; ".join(errors))
    entries = by_id(data()["boilerplates"]["entries"])
    actions = []
    for starter in manifest["starters"]:
        entry = entries[starter["id"]]
        upstream = entry.get("upstream", {})
        repository = entry.get("repository")
        action = {
            "id": starter["id"], "pin": starter.get("pin"),
            "strategy": starter.get("update_strategy"),
            "action": "create-reviewed-update-branch",
        }
        if args.check:
            if repository and upstream.get("branch"):
                completed = subprocess.run(
                    ["git", "ls-remote", repository, f"refs/heads/{upstream['branch']}"],
                    text=True, capture_output=True, check=False,
                )
                if completed.returncode != 0 or not completed.stdout.strip():
                    raise PlatformError(f"No se pudo consultar upstream de {starter['id']}")
                observed = completed.stdout.split()[0]
                action.update({
                    "observed_commit": observed,
                    "action": "up-to-date" if observed == starter.get("pin") else "update-available",
                })
            else:
                action["action"] = "internal-release-current"
        actions.append(action)
    print(dump_json({"actions": actions, "warnings": warnings}), end="")
    return 0


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(prog="eng", description="Engineering Platform: recetas mínimas y verificables")
    parser.add_argument("--version", action="version", version=f"%(prog)s {PLATFORM_VERSION}")
    sub = parser.add_subparsers(dest="command", required=True)

    catalog = sub.add_parser("catalog", help="Listar boilerplates y estados")
    catalog.add_argument("--category")
    catalog.add_argument("--decision-status")
    catalog.add_argument("--delivery-status")
    catalog.set_defaults(handler=command_catalog)

    boilerplate = sub.add_parser("boilerplate", help="Curar un boilerplate")
    boilerplate_sub = boilerplate.add_subparsers(dest="boilerplate_command", required=True)
    evaluate = boilerplate_sub.add_parser("evaluate", help="Detectar duplicado o alta candidata")
    evaluate.add_argument("repository")
    evaluate.add_argument("--observed-commit")
    evaluate.add_argument("--observed-date")
    evaluate.add_argument("--category")
    evaluate.set_defaults(handler=command_evaluate)

    recommend = sub.add_parser("recommend", help="Resolver una Recipe desde un intake")
    recommend.add_argument("--input", required=True)
    recommend.set_defaults(handler=command_recommend)

    new = sub.add_parser("new", help="Crear blueprint reproducible desde un intake")
    new.add_argument("--from", dest="input", required=True)
    new.add_argument("--output", required=True)
    new.add_argument("--dry-run", action="store_true")
    new.set_defaults(handler=command_new)

    bootstrap = sub.add_parser(
        "bootstrap", help="Crear proyecto y handoff desde una definición confirmada"
    )
    bootstrap.add_argument("--from", dest="input", required=True)
    bootstrap.add_argument("--output", default=".")
    bootstrap.add_argument("--dry-run", action="store_true")
    bootstrap.add_argument("--skip-setup", action="store_true", help="Materializar sin instalar dependencias")
    bootstrap.add_argument("--skip-checks", action="store_true", help="Materializar sin ejecutar gates")
    bootstrap.set_defaults(handler=command_bootstrap)

    start = sub.add_parser(
        "start", help="Crear una carpeta de proyecto en el workspace y abrir Pi"
    )
    start.add_argument("name")
    start.add_argument("--workspace")
    start.add_argument("--dry-run", action="store_true")
    start.set_defaults(handler=command_start)

    install = sub.add_parser("install", help="Instalar globalmente la integración Pi")
    install.add_argument("--global", dest="global_install", action="store_true", required=True)
    install.add_argument("--target", choices=["pi"], default="pi")
    install.add_argument("--home", help=argparse.SUPPRESS)
    install.add_argument("--force", action="store_true")
    install.add_argument("--dry-run", action="store_true")
    install.set_defaults(handler=command_install)

    uninstall = sub.add_parser("uninstall", help="Retirar la integración Pi global")
    uninstall.add_argument("--global", dest="global_install", action="store_true", required=True)
    uninstall.add_argument("--target", choices=["pi"], default="pi")
    uninstall.add_argument("--home", help=argparse.SUPPRESS)
    uninstall.add_argument("--dry-run", action="store_true")
    uninstall.set_defaults(handler=command_uninstall)

    doctor = sub.add_parser("doctor", help="Revisar coherencia de un proyecto")
    doctor.add_argument("--project", default=".")
    doctor.add_argument("--global", dest="global_install", action="store_true")
    doctor.add_argument("--home", help=argparse.SUPPRESS)
    doctor.set_defaults(handler=command_doctor)

    handoff = sub.add_parser("handoff", help="Regenerar instrucciones para Gentle AI")
    handoff.add_argument("--project", default=".")
    handoff.set_defaults(handler=command_handoff)

    plan = sub.add_parser("plan", help="Seleccionar skills y gates para un cambio")
    plan.add_argument("--project", default=".")
    plan.add_argument("--change-type", choices=sorted(CHANGE_PLANS), required=True)
    plan.set_defaults(handler=command_plan)

    check = sub.add_parser("check", help="Seleccionar gates por manifest y archivos")
    check.add_argument("--project", default=".")
    check.add_argument("--changed-files", nargs="*")
    check.add_argument("--run", action="store_true", help="Ejecutar setup y checks registrados por los adapters")
    check.set_defaults(handler=command_check)

    add = sub.add_parser("add", help="Planear o aplicar un feature pack")
    add.add_argument("feature")
    add.add_argument("--project", default=".")
    add.add_argument("--apply", action="store_true", help="Actualizar intake, manifest y secciones gestionadas")
    add.set_defaults(handler=command_add)

    update = sub.add_parser("update", help="Planear actualización según cada upstream")
    update.add_argument("--project", default=".")
    update.add_argument("--check", action="store_true")
    update.set_defaults(handler=command_update)
    return parser


def main(argv: list[str] | None = None) -> int:
    parser = build_parser()
    args = parser.parse_args(argv)
    try:
        return int(args.handler(args))
    except PlatformError as exc:
        print(f"ERROR: {exc}", file=sys.stderr)
        return 2


if __name__ == "__main__":
    raise SystemExit(main())
