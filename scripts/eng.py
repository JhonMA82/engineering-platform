#!/usr/bin/env python3
"""Minimal CLI for selecting, recording and checking project recipes."""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import re
import shlex
import shutil
import subprocess
import sys
import tempfile
from collections.abc import Mapping
from copy import deepcopy
from datetime import date
from pathlib import Path
from typing import Any
from urllib.parse import urlparse


ROOT = Path(__file__).resolve().parents[1]
PLATFORM_VERSION = json.loads((ROOT / "package.json").read_text(encoding="utf-8"))["version"]
DEFINITION_SCHEMA_VERSION = 1
PROJECT_SCHEMA_URL = (
    "https://raw.githubusercontent.com/JhonMA82/engineering-platform/"
    f"v{PLATFORM_VERSION}/schemas/project.schema.json"
)
DEFINITION_SCHEMA_URL = (
    "https://raw.githubusercontent.com/JhonMA82/engineering-platform/"
    f"v{PLATFORM_VERSION}/schemas/project-definition.schema.json"
)
PI_ENGINEERING_PLATFORM_GIT_SOURCE_PREFIX = "git:github.com/JhonMA82/engineering-platform@"
_ENVIRONMENT_NAME_PATTERN = re.compile(r"[A-Za-z_][A-Za-z0-9_]*")


def _install_ignores(_directory: str, names: list[str]) -> set[str]:
    """Exclude runtime/secrets from the Pi package, keeping env templates."""
    ignored: set[str] = set()
    for name in names:
        if name in {
            ".git",
            ".atl",
            "__pycache__",
            ".pytest_cache",
            "node_modules",
            "dist",
            "target",
        }:
            ignored.add(name)
        elif name == ".env" or (
            name.startswith(".env.") and name not in {".env.example", ".env.sample"}
        ):
            ignored.add(name)
        elif name.endswith((".pyc", ".zip")):
            ignored.add(name)
    return ignored


class PlatformError(ValueError):
    """A user-facing platform decision or validation error."""


def _validate_environment(value: Any, field: str) -> dict[str, str]:
    if not isinstance(value, Mapping):
        raise PlatformError(f"{field} debe ser un objeto de variables de entorno")
    result: dict[str, str] = {}
    for name, environment_value in value.items():
        if not isinstance(name, str) or _ENVIRONMENT_NAME_PATTERN.fullmatch(name) is None:
            raise PlatformError(f"{field} contiene un nombre de variable inválido")
        if not isinstance(environment_value, str) or any(
            character in environment_value for character in "\x00\r\n"
        ):
            raise PlatformError(f"{field} contiene un valor inválido")
        result[name] = environment_value
    return result


def _adapter_environment(adapter: dict[str, Any]) -> dict[str, str]:
    materializer = adapter.get("materializer", {})
    if not isinstance(materializer, Mapping):
        raise PlatformError(f"{adapter.get('boilerplate_id', 'unknown')}: materializer inválido")
    if "environment" not in materializer:
        return {}
    return _validate_environment(
        materializer["environment"],
        f"{adapter.get('boilerplate_id', 'unknown')}.materializer.environment",
    )


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


def boilerplate_references(boilerplate_id: str) -> list[str]:
    references: list[str] = []
    for recipe in data()["recipes"]["paths"]:
        stack = recipe.get("stack", {})
        if boilerplate_id in stack.get("starters", []):
            references.append(f"{recipe['id']}:default")
        if boilerplate_id in stack.get("alternatives", []):
            references.append(f"{recipe['id']}:alternative")
        packs = recipe.get("solution_packs", {})
        if boilerplate_id in packs.get("default", []) + packs.get("optional", []):
            references.append(f"{recipe['id']}:solution-pack")
    return references


def verify_boilerplate(boilerplate_id: str) -> dict[str, Any]:
    entries = by_id(data()["boilerplates"]["entries"])
    entry = entries.get(boilerplate_id)
    if not entry:
        raise PlatformError(f"Boilerplate inexistente: {boilerplate_id}")
    integration = entry.get("integration", {})
    adapter_path = integration.get("adapter")
    evidence_path = integration.get("evidence")
    errors: list[str] = []
    if not adapter_path or not (ROOT / adapter_path).is_file():
        errors.append("adapter ausente")
    if not evidence_path or not (ROOT / evidence_path).is_file():
        errors.append("evidencia AI-friendly ausente")
    if not entry.get("upstream", {}).get("commit"):
        errors.append("pin upstream ausente")
    return {
        "id": boilerplate_id,
        "ok": not errors,
        "errors": errors,
        "references": boilerplate_references(boilerplate_id),
        "adapter": adapter_path,
        "evidence": evidence_path,
    }


def _write_json_atomic(path: Path, value: dict[str, Any]) -> None:
    temporary = path.with_suffix(path.suffix + ".tmp")
    temporary.write_text(dump_json(value), encoding="utf-8")
    temporary.replace(path)


def _write_text_atomic(path: Path, value: str) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    temporary = path.with_suffix(path.suffix + ".tmp")
    temporary.write_text(value, encoding="utf-8")
    temporary.replace(path)


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


def _starter_capabilities(
    adapter: dict[str, Any], selected_features: list[str], database: str | None
) -> tuple[list[str], list[str], list[str]]:
    """Translate platform capabilities without teaching the engine starter names."""
    capabilities = adapter.get("capabilities")
    if not isinstance(capabilities, dict):
        return [], [], []

    supported_databases = capabilities.get("supported_databases", [])
    if database and supported_databases and database not in supported_databases:
        raise PlatformError(
            f"{adapter.get('boilerplate_id')} no soporta {database}; "
            f"soporta: {', '.join(supported_databases)}"
        )

    generated = list(capabilities.get("database_features", {}).get(database or "", []))
    provided: list[str] = []
    unresolved: list[str] = []
    selected = set(selected_features)
    mapping = capabilities.get("feature_map", {})
    unsupported = set(capabilities.get("unsupported_features", []))
    for feature_id in selected_features:
        if feature_id in unsupported:
            unresolved.append(feature_id)
            continue
        rule = mapping.get(feature_id)
        if rule is None:
            continue
        if isinstance(rule, str):
            generated.append(rule)
            provided.append(feature_id)
            continue
        if not isinstance(rule, dict):
            raise PlatformError(
                f"{adapter.get('boilerplate_id')}: mapping inválido para {feature_id}"
            )
        requirements = set(rule.get("requires", []))
        if not requirements.issubset(selected):
            unresolved.append(feature_id)
            continue
        targets = rule.get("to", [])
        if isinstance(targets, str):
            targets = [targets]
        generated.extend(targets)
        provided.append(feature_id)
    return unique(generated), unique(provided), unique(unresolved)


def _starter_manifest_entry(
    starter_id: str,
    *,
    selected_features: list[str],
    database: str | None,
    boilerplate_index: dict[str, dict[str, Any]] | None = None,
) -> tuple[dict[str, Any], list[str], list[str]]:
    entries = boilerplate_index or by_id(data()["boilerplates"]["entries"])
    item = entries.get(starter_id)
    if not item:
        raise PlatformError(f"Boilerplate inexistente: {starter_id}")
    if item.get("delivery_status") not in {"curated", "released"}:
        raise PlatformError(
            f"{starter_id} no es materializable: {item.get('delivery_status')}"
        )
    pin = item.get("upstream", {}).get("commit")
    adapter_path = item.get("integration", {}).get("adapter")
    if not pin or not adapter_path:
        raise PlatformError(f"{starter_id} no tiene adapter y pin materializables")
    adapter = read_json(ROOT / adapter_path)
    materializer = adapter.get("materializer", {})
    generated, provided, unresolved = _starter_capabilities(
        adapter, selected_features, database
    )
    starter: dict[str, Any] = {
        "id": starter_id,
        "delivery_status": item["delivery_status"],
        "integration_mode": item["integration"]["mode"],
        "update_strategy": item["integration"]["update_strategy"],
        "pin": pin,
        "repository": item.get("repository"),
        "adapter": adapter_path,
        "destination": materializer.get("destination"),
    }
    if generated:
        starter["generator_features"] = generated
    if provided:
        starter["provided_features"] = provided
    if unresolved:
        starter["unmaterialized_features"] = unresolved
    return starter, provided, unresolved


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
    capability_providers: dict[str, list[str]] = {feature: [] for feature in features}
    warnings: list[str] = []
    for starter_id in recipe["stack"]["starters"]:
        starter, provided_features, unresolved_features = _starter_manifest_entry(
            starter_id,
            selected_features=features,
            database=requested_database,
            boilerplate_index=boilerplate_index,
        )
        for feature_id in provided_features:
            capability_providers[feature_id].append(starter_id)
        if unresolved_features:
            warnings.append(
                f"{starter_id} no materializa directamente: {', '.join(unresolved_features)}; "
                "Gentle debe implementarlas como capacidades del proyecto."
            )
        starters.append(starter)
    capability_status = {
        feature_id: {
            "state": "materialized" if providers else "pending-implementation",
            "provided_by": providers,
            "owner": "boilerplate" if providers else "gentle-ai",
        }
        for feature_id, providers in capability_providers.items()
    }
    pending = [
        feature_id
        for feature_id, status in capability_status.items()
        if status["state"] == "pending-implementation"
    ]
    if pending:
        warnings.append("Capacidades pendientes para Gentle: " + ", ".join(pending))
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
        "$schema": PROJECT_SCHEMA_URL,
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
        "capability_status": capability_status,
        "skills": unique(resolved_skills),
        "gates": unique(resolved_gates),
        "exclusions": sorted(set(recipe["exclusions"]).union(excluded)),
        "ownership": {
            "managed": [".engineering/project.json", ".engineering/intake.json"],
            "managed_sections": ["AGENTS.md", "ARCHITECTURE.md"],
            "seeded": ["README.md"],
            "user_owned": ["apps/**", "services/**", "packages/**", "src/**", "tests/**"],
        },
        "warnings": warnings,
    }


def architecture_markdown(manifest: dict[str, Any]) -> str:
    boundaries = "\n".join(
        f"- `{item['destination']}` pertenece a `{item['id']}`; respeta sus instrucciones y comandos."
        for item in manifest["starters"]
    )
    tenancy = (
        "Aislamiento por tenant obligatorio."
        if "multitenancy" in manifest["features"]
        else "Single-tenant por defecto."
    )
    return f"""# Arquitectura: {manifest['project']['name']}

La definición vive en `.engineering/project-definition.json`; stack, features, exclusiones y gates viven únicamente en `.engineering/project.json`.

## Patrones

- Monolito modular por servicio, contratos explícitos y mínimo privilegio.
- Clientes sin autoridad de seguridad ni reglas de dominio duplicadas.
- Cambios de datos con migración, recuperación y auditoría proporcional al riesgo.
- {tenancy}

## Límites materiales

{boundaries}

Las desviaciones permanentes requieren actualizar la Recipe o registrar una decisión.
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


def _adapter_requirements(adapter: dict[str, Any]) -> list[dict[str, str]]:
    requirements = adapter.get("requirements", [])
    if not isinstance(requirements, list):
        raise PlatformError(f"{adapter.get('boilerplate_id')}: requirements inválidos")
    result: list[dict[str, str]] = []
    for item in requirements:
        if not isinstance(item, dict) or not isinstance(item.get("executable"), str):
            raise PlatformError(f"{adapter.get('boilerplate_id')}: requirement inválido")
        result.append(item)
    kind = adapter.get("materializer", {}).get("type")
    if kind in {"git-copy", "git-generator"} and not any(
        item["executable"] == "git" for item in result
    ):
        result.insert(0, {"executable": "git"})
    return result


def _requirement_status(requirement: dict[str, str], starter_id: str) -> dict[str, Any]:
    executable = requirement["executable"]
    path = shutil.which(executable)
    status: dict[str, Any] = {
        "starter": starter_id,
        "executable": executable,
        "available": bool(path),
        "expected": requirement.get("version_prefix"),
        "ok": bool(path),
    }
    if not path:
        return status
    version_args = requirement.get("version_args", "--version").split()
    try:
        completed = subprocess.run(
            [path, *version_args],
            text=True,
            capture_output=True,
            check=False,
            timeout=10,
        )
    except subprocess.TimeoutExpired:
        status.update({"ok": False, "error": "timeout al consultar versión"})
        return status
    output = (completed.stdout or completed.stderr).strip().splitlines()
    version = output[0] if output else ""
    status["version"] = version
    prefix = requirement.get("version_prefix")
    if completed.returncode != 0:
        status.update({"ok": False, "error": "no se pudo consultar versión"})
    elif prefix:
        match = re.search(r"\d+(?:\.\d+){0,2}", version)
        normalized = match.group(0) if match else version
        status["ok"] = normalized.startswith(prefix)
    return status


def adapter_preflight(adapter: dict[str, Any]) -> list[dict[str, Any]]:
    starter_id = adapter.get("boilerplate_id", "unknown")
    return [
        _requirement_status(requirement, starter_id)
        for requirement in _adapter_requirements(adapter)
    ]


def _assert_adapter_requirements(adapter: dict[str, Any]) -> list[dict[str, Any]]:
    statuses = adapter_preflight(adapter)
    failures = [item for item in statuses if not item["ok"]]
    if failures:
        details = []
        for item in failures:
            if not item["available"]:
                details.append(f"{item['executable']} no está instalado")
            else:
                details.append(
                    f"{item['executable']}={item.get('version', '?')} no cumple "
                    f"{item.get('expected') or 'el requisito'}"
                )
        raise PlatformError(
            f"Preflight de {adapter.get('boilerplate_id')} falló: " + "; ".join(details)
        )
    return statuses


def _run_record(
    command: list[str],
    cwd: Path,
    *,
    gate: str | None = None,
    environment: Mapping[str, str] | None = None,
) -> dict[str, Any]:
    if not command or any(not isinstance(item, str) or not item for item in command):
        raise PlatformError("El adapter contiene un comando inválido")
    declared_environment = (
        _validate_environment(environment, "environment")
        if environment is not None
        else {}
    )
    executable = shutil.which(command[0])
    if not executable:
        raise PlatformError(f"Falta el ejecutable requerido por el adapter: {command[0]}")
    timeout = int(os.environ.get("ENG_COMMAND_TIMEOUT", "1800"))
    run_kwargs: dict[str, Any] = {
        "cwd": cwd,
        "text": True,
        "capture_output": True,
        "check": False,
        "timeout": timeout,
    }
    if declared_environment:
        command_environment = os.environ.copy()
        command_environment.update(declared_environment)
        run_kwargs["env"] = command_environment
    try:
        completed = subprocess.run(command, **run_kwargs)
    except subprocess.TimeoutExpired as exc:
        raise PlatformError(
            f"Timeout de {timeout}s al ejecutar {' '.join(command)} en {cwd}"
        ) from exc
    record: dict[str, Any] = {
        "command": command,
        "workdir": ".",
        "returncode": completed.returncode,
        "status": "passed" if completed.returncode == 0 else "failed",
    }
    if gate:
        record["gate"] = gate
    if declared_environment:
        record["environment"] = declared_environment
    if completed.stdout.strip():
        record["stdout"] = completed.stdout.strip()[-4000:]
    if completed.stderr.strip():
        record["stderr"] = completed.stderr.strip()[-4000:]
    if completed.returncode != 0:
        detail = "\n".join(
            part for part in (completed.stdout.strip(), completed.stderr.strip()) if part
        ) or "sin salida"
        raise PlatformError(f"Falló {' '.join(command)} en {cwd}: {detail[-1000:]}")
    return record


def _run_adapter_record(
    command: list[str],
    cwd: Path,
    *,
    environment: Mapping[str, str] | None = None,
    gate: str | None = None,
) -> dict[str, Any]:
    declared_environment = (
        _validate_environment(environment, "environment")
        if environment is not None
        else {}
    )
    kwargs: dict[str, Any] = {}
    if gate is not None:
        kwargs["gate"] = gate
    if declared_environment:
        kwargs["environment"] = declared_environment
    return _run_record(command, cwd, **kwargs)


def _add_declared_environment(
    record: dict[str, Any], environment: Mapping[str, str]
) -> None:
    if environment:
        record["environment"] = dict(environment)


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


def _apply_overlay(adapter: dict[str, Any], destination: Path) -> None:
    overlay = adapter.get("overlay")
    if not overlay:
        return
    overlay_path = ROOT / _safe_relative(overlay, "overlay")
    _copy_materialized_tree(overlay_path, destination)


def _prune_materialized_output(adapter: dict[str, Any], destination: Path) -> list[str]:
    removed: list[str] = []
    for pattern in adapter.get("prune", []):
        _safe_relative(pattern, "prune")
        for target in sorted(destination.glob(pattern)):
            try:
                relative = target.relative_to(destination)
            except ValueError as exc:
                raise PlatformError(f"prune salió del destino: {target}") from exc
            if target.is_symlink():
                target.unlink()
            elif target.is_dir():
                shutil.rmtree(target)
            elif target.exists():
                target.unlink()
            removed.append(relative.as_posix())
    return removed


def _apply_patch_files(values: list[str], destination: Path, field: str) -> list[str]:
    applied: list[str] = []
    for value in values:
        relative = _safe_relative(value, field)
        patch_path = ROOT / relative
        if not patch_path.is_file():
            raise PlatformError(f"Patch inexistente: {relative}")
        _run_record(["git", "apply", "--check", str(patch_path)], destination)
        _run_record(["git", "apply", str(patch_path)], destination)
        applied.append(relative.as_posix())
    return applied


def _apply_adapter_patches(adapter: dict[str, Any], destination: Path) -> list[str]:
    return _apply_patch_files(adapter.get("patches", []), destination, "patches")


def _render_adapter_command(
    command: list[str],
    *,
    source: Path,
    output: Path,
    project_name: str,
    generator_features: list[str],
) -> list[str]:
    replacements = {
        "{source}": str(source),
        "{output}": str(output),
        "{project_name}": project_name,
        "{features_csv}": ",".join(generator_features),
    }
    rendered: list[str] = []
    for token in command:
        value = token
        for marker, replacement in replacements.items():
            value = value.replace(marker, replacement)
        rendered.append(value)
    return rendered


def _checkout_git_source(source: dict[str, Any], destination: Path) -> None:
    repository = source.get("repository")
    commit = source.get("commit")
    if not repository or not commit:
        raise PlatformError("El generador Git necesita repository y commit exactos")
    destination.mkdir()
    _run_record(["git", "init", "--quiet"], destination)
    _run_record(["git", "remote", "add", "origin", repository], destination)
    _run_record(["git", "fetch", "--quiet", "--depth", "1", "origin", commit], destination)
    _run_record(["git", "checkout", "--quiet", "FETCH_HEAD"], destination)


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
            declared_manager = package.get("packageManager", "")
            if declared_manager.startswith("bun@"):
                manager = "bun"
            elif declared_manager.startswith("pnpm@"):
                manager = "pnpm"
            elif declared_manager.startswith("yarn@"):
                manager = "yarn"
            elif (parent / "bun.lock").exists() or (parent / "bun.lockb").exists():
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
    requirement_records: list[dict[str, Any]] = []
    for starter in manifest["starters"]:
        adapter_path = starter.get("adapter")
        if not adapter_path or not starter.get("pin"):
            raise PlatformError(f"{starter['id']} no tiene adapter y pin materializables")
        adapter = read_json(ROOT / adapter_path)
        adapter_environment = _adapter_environment(adapter)
        requirement_records.extend(_assert_adapter_requirements(adapter))
        materializer = adapter.get("materializer", {})
        kind = materializer.get("type")
        destination_relative = _safe_relative(
            starter.get("destination_override") or materializer.get("destination", "."),
            "destination",
        )
        destination = project / destination_relative
        source = adapter.get("source", {})
        generator_record: dict[str, Any] | None = None
        source_pruned: list[str] = []
        source_patches: list[str] = []
        if kind == "git-copy":
            destination.mkdir(parents=True, exist_ok=True)
            repository = source.get("repository")
            commit = source.get("commit")
            if not repository or not commit:
                raise PlatformError(f"{starter['id']} necesita repository y commit exacto")
            with tempfile.TemporaryDirectory(prefix=f"eng-{starter['id']}-") as temporary:
                clone = Path(temporary) / "source"
                _checkout_git_source(source, clone)
                source_pruned = _prune_materialized_output(adapter, clone)
                _copy_materialized_tree(clone, destination)
        elif kind == "local-copy":
            destination.mkdir(parents=True, exist_ok=True)
            local_path = _safe_relative(source.get("local_path", ""), "source.local_path")
            _copy_materialized_tree(ROOT / local_path, destination)
        elif kind == "command-generator":
            working = project / _safe_relative(materializer.get("working_directory", "."), "working_directory")
            working.mkdir(parents=True, exist_ok=True)
            generator_record = _run_adapter_record(
                materializer.get("command", []),
                working,
                environment=adapter_environment,
            )
            _add_declared_environment(generator_record, adapter_environment)
            generator_record["workdir"] = working.relative_to(project).as_posix() or "."
            if not destination.exists() or not any(destination.iterdir()):
                raise PlatformError(f"El generador de {starter['id']} no creó {destination_relative}")
        elif kind == "git-generator":
            destination.mkdir(parents=True, exist_ok=True)
            with tempfile.TemporaryDirectory(prefix=f"eng-{starter['id']}-generator-") as temporary:
                temporary_path = Path(temporary)
                clone = temporary_path / "source"
                generated = temporary_path / "generated"
                _checkout_git_source(source, clone)
                source_patches = _apply_patch_files(
                    materializer.get("source_patches", []), clone, "source_patches"
                )
                for command in materializer.get("source_setup", []):
                    _run_adapter_record(command, clone, environment=adapter_environment)
                command = _render_adapter_command(
                    materializer.get("command", []),
                    source=clone,
                    output=generated,
                    project_name=manifest["project"]["name"],
                    generator_features=starter.get("generator_features", []),
                )
                generator_record = _run_adapter_record(
                    command,
                    clone,
                    environment=adapter_environment,
                )
                _add_declared_environment(generator_record, adapter_environment)
                generator_record["workdir"] = "<pinned-source>"
                if not generated.is_dir() or not any(generated.iterdir()):
                    raise PlatformError(f"El generador de {starter['id']} no produjo archivos")
                _copy_materialized_tree(generated, destination)
        else:
            raise PlatformError(f"Materializer desconocido para {starter['id']}: {kind}")
        if destination_relative == Path("."):
            _preserve_upstream_instructions(project, starter["id"])
        pruned = unique(source_pruned + _prune_materialized_output(adapter, destination))
        patches = _apply_adapter_patches(adapter, destination)
        _apply_overlay(adapter, destination)
        for command in materializer.get("setup", []):
            if skip_setup:
                record = {
                    "starter": starter["id"],
                    "command": command,
                    "workdir": destination_relative.as_posix(),
                    "status": "skipped",
                }
                _add_declared_environment(record, adapter_environment)
                setup_records.append(record)
            else:
                record = _run_adapter_record(
                    command,
                    destination,
                    environment=adapter_environment,
                )
                _add_declared_environment(record, adapter_environment)
                record.update({"starter": starter["id"], "workdir": destination_relative.as_posix()})
                setup_records.append(record)
        for check in materializer.get("checks", []):
            if skip_checks:
                record = {
                    "starter": starter["id"],
                    "gate": check.get("gate"),
                    "command": check.get("command", []),
                    "workdir": destination_relative.as_posix(),
                    "status": "skipped",
                }
                _add_declared_environment(record, adapter_environment)
                check_records.append(record)
            else:
                record = _run_adapter_record(
                    check.get("command", []),
                    destination,
                    gate=check.get("gate"),
                    environment=adapter_environment,
                )
                _add_declared_environment(record, adapter_environment)
                record.update({"starter": starter["id"], "workdir": destination_relative.as_posix()})
                check_records.append(record)
        starter_records.append({
            "id": starter["id"], "type": kind, "destination": destination_relative.as_posix(),
            "repository": source.get("repository"), "branch": source.get("branch"),
            "commit": source.get("commit"), "content_sha256": _content_sha256(destination),
            "generator": generator_record, "pruned": pruned, "patches": patches,
            "source_patches": source_patches,
        })
    actual = _actual_project_data(project)
    readiness = "verified" if not skip_checks and check_records else "code-ready"
    return {
        "schema_version": 1,
        "platform_version": PLATFORM_VERSION,
        "generated_at": date.today().isoformat(),
        "readiness": readiness,
        "starters": starter_records,
        "requirements": requirement_records,
        "setup": setup_records,
        "checks": check_records,
        **actual,
    }


def _initialize_seed_repository(project: Path, materialization: dict[str, Any]) -> None:
    if not (project / ".git").exists():
        _run_record(["git", "init", "--quiet", "--initial-branch", "main"], project)
    existing = subprocess.run(["git", "remote"], cwd=project, text=True, capture_output=True, check=False).stdout.split()
    for starter in materialization.get("starters", []):
        repository = starter.get("repository")
        remote = f"seed-{starter['id']}"
        if repository and remote not in existing:
            _run_record(["git", "remote", "add", remote, repository], project)


def _read_materialization(
    project: Path, manifest: dict[str, Any] | None = None
) -> dict[str, Any]:
    """Read the materialization record, with v0.5.x embedded-data compatibility."""
    path = project / ".engineering/materialization.json"
    if path.exists():
        return read_json(path)
    embedded = (manifest or {}).get("materialization")
    return deepcopy(embedded) if isinstance(embedded, dict) else {}


def project_ci_yaml(manifest: dict[str, Any], materialization: dict[str, Any]) -> str:
    lines = [
        "name: engineering",
        "",
        "on:",
        "  push:",
        "  pull_request:",
        "",
        "permissions:",
        "  contents: read",
        "",
        "jobs:",
    ]
    for starter in manifest.get("starters", []):
        starter_id = starter["id"]
        job_id = re.sub(r"[^a-z0-9_-]", "-", starter_id.lower())
        adapter = read_json(ROOT / starter["adapter"])
        requirements = _adapter_requirements(adapter)
        checks = [
            record
            for record in materialization.get("checks", [])
            if record.get("starter") == starter_id
        ]
        ci_setup = adapter.get("materializer", {}).get("ci_setup")
        if ci_setup:
            records = [
                {
                    "starter": starter_id,
                    "command": command,
                    "workdir": starter.get("destination", "."),
                }
                for command in ci_setup
            ] + checks
        else:
            records = [
                record
                for record in materialization.get("setup", [])
                if record.get("starter") == starter_id
            ] + checks
        lines.extend(
            [
                f"  {job_id}:",
                "    runs-on: ubuntu-24.04",
                "    timeout-minutes: 30",
            ]
        )
        environment = _adapter_environment(adapter)
        if environment:
            lines.append("    env:")
            lines.extend(
                f"      {name}: {json.dumps(value, ensure_ascii=False)}"
                for name, value in environment.items()
            )
        lines.extend(
            [
                "    steps:",
                "      - uses: actions/checkout@v4",
            ]
        )
        executables = {item["executable"]: item for item in requirements}
        if "python3" in executables or "python" in executables or "uv" in executables:
            python_requirement = executables.get("python3") or executables.get("python") or {}
            python_version = python_requirement.get("ci_version", "3.13")
            lines.extend(
                [
                    "      - uses: actions/setup-python@v5",
                    "        with:",
                    f'          python-version: "{python_version}"',
                ]
            )
        if "uv" in executables:
            lines.append("      - uses: astral-sh/setup-uv@v6")
        if "node" in executables or "npx" in executables:
            node_requirement = executables.get("node") or executables.get("npx") or {}
            lines.extend(
                [
                    "      - uses: actions/setup-node@v4",
                    "        with:",
                    f'          node-version: "{node_requirement.get("ci_version", "24")}"',
                ]
            )
        if "bun" in executables:
            bun_version = executables["bun"].get("ci_version", "1.4.0")
            lines.extend(
                [
                    "      - uses: oven-sh/setup-bun@v2",
                    "        with:",
                    f'          bun-version: "{bun_version}"',
                ]
            )
        for index, record in enumerate(records, start=1):
            label = record.get("gate") or "setup"
            lines.extend(
                [
                    f"      - name: {label} ({index})",
                    f"        working-directory: {record.get('workdir', '.')}",
                    f"        run: {shlex.join(record['command'])}",
                ]
            )
    return "\n".join(lines).rstrip() + "\n"


def write_project_ci(
    project: Path, manifest: dict[str, Any], materialization: dict[str, Any]
) -> Path:
    workflow = project / ".github/workflows/engineering.yml"
    workflow.parent.mkdir(parents=True, exist_ok=True)
    workflow.write_text(project_ci_yaml(manifest, materialization), encoding="utf-8")
    return workflow


def gentle_handoff_data(
    manifest: dict[str, Any], definition: dict[str, Any] | None
) -> dict[str, Any]:
    source_of_truth = {
        "decisions": ".engineering/project.json",
        "architecture": "ARCHITECTURE.md",
        "rules": "AGENTS.md",
    }
    if definition:
        source_of_truth["idea"] = ".engineering/project-definition.json"
    if manifest.get("scaffold_status") == "materialized":
        source_of_truth["materialization"] = ".engineering/materialization.json"
    read_first = ["GENTLE.md"]
    if "idea" in source_of_truth:
        read_first.append(source_of_truth["idea"])
    read_first.extend(
        source_of_truth[key]
        for key in ("decisions", "architecture", "rules", "materialization")
        if key in source_of_truth
    )
    return {
        "schema_version": 2,
        "platform_version": manifest["platform_version"],
        "scaffold_status": manifest["scaffold_status"],
        "readiness": manifest.get("readiness", "code-ready"),
        "project": {"name": manifest["project"]["name"]},
        "source_of_truth": source_of_truth,
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
        "read_first": read_first,
    }


def gentle_markdown(
    manifest: dict[str, Any],
    definition: dict[str, Any] | None,
    handoff: dict[str, Any],
) -> str:
    starters = ", ".join(
        f"`{item['id']}` → `{item.get('destination') or 'por definir'}`"
        for item in manifest.get("starters", [])
    ) or "ninguno"
    patterns = [
        "modular-monolith",
        "explicit-contracts",
        "least-privilege",
        "incremental-delivery",
        "tenant-isolation"
        if "multitenancy" in manifest.get("features", [])
        else "single-tenant-first",
    ]
    read_first = ", ".join(f"`{item}`" for item in handoff["read_first"])
    sources = handoff["source_of_truth"]
    pending = [
        feature_id
        for feature_id, status in manifest.get("capability_status", {}).items()
        if status.get("state") == "pending-implementation"
    ]
    pending_text = ", ".join(f"`{item}`" for item in pending) or "ninguna"
    return f"""# Handoff a Gentle AI

## Contexto

- Proyecto: `{manifest['project']['name']}`
- Recipe: `{manifest['recipe']['id']}@{manifest['recipe']['version']}`
- Boilerplates: {starters}
- Base de datos: `{manifest.get('database') or 'ninguna'}`
- Patrones: {', '.join(f'`{item}`' for item in patterns)}
- Estado: `{handoff['scaffold_status']}` · readiness `{handoff['readiness']}`
- Capacidades pendientes de implementación: {pending_text}

## Fuentes de verdad

- Idea y alcance: `{sources.get('idea', 'no disponible')}`
- Stack, skills, gates y ownership: `{sources['decisions']}`
- Arquitectura: `{sources['architecture']}`
- Reglas: `{sources['rules']}`
- Materialización, pins y verificación: `{sources.get('materialization', 'no materializado')}`

## Instrucciones de ejecución

1. Lee, en orden: {read_first}. La idea no se duplica aquí.
2. Decide entre ejecución directa y SDD según riesgo, ambigüedad, contratos, datos y permisos; registra brevemente el motivo.
3. Conserva la Recipe, el stack, los patrones y las exclusiones; consulta las fuentes antes de desviarte.
4. Implementa únicamente las capacidades con estado `pending-implementation` que pertenezcan al incremento solicitado; no asumas que una feature declarada ya existe.
5. Implementa el incremento vertical mínimo y ejecuta los quality gates indicados en `.engineering/project.json`.
6. Si readiness es `code-ready`, ejecuta `eng check --run` antes de tratar el proyecto como verificado.
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
    (project / "GENTLE.md").write_text(
        gentle_markdown(manifest, definition, handoff), encoding="utf-8"
    )
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
    definition_for_output = deepcopy(definition) if definition else None
    if definition:
        validate_project_definition(definition)
        manifest["project"]["summary"] = definition["idea"]["summary"]
        manifest["definition_status"] = "confirmed"
        definition_for_output["$schema"] = DEFINITION_SCHEMA_URL
    if definition:
        manifest["ownership"]["managed"].append(".engineering/project-definition.json")
    manifest["ownership"]["managed"].extend(
        [".engineering/gentle-handoff.json", "GENTLE.md"]
    )
    if materialize:
        manifest["ownership"]["managed"].extend(
            [".engineering/materialization.json", ".github/workflows/engineering.yml"]
        )
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
            if materialization["readiness"] == "verified":
                for status in manifest.get("capability_status", {}).values():
                    if status.get("state") == "materialized":
                        status["state"] = "verified"
            _initialize_seed_repository(staged, materialization)
        engineering = staged / ".engineering"
        engineering.mkdir(parents=True, exist_ok=True)
        if definition_for_output:
            (engineering / "project-definition.json").write_text(
                dump_json(definition_for_output), encoding="utf-8"
            )
        (engineering / "intake.json").write_text(dump_json(intake), encoding="utf-8")
        (staged / "ARCHITECTURE.md").write_text(architecture_markdown(manifest), encoding="utf-8")
        (staged / "AGENTS.md").write_text(agents_markdown(manifest), encoding="utf-8")
        (staged / "README.md").write_text(readme_markdown(manifest), encoding="utf-8")
        if materialize:
            materialization.update(_actual_project_data(staged))
            write_project_ci(staged, manifest, materialization)
            materialization.update(_actual_project_data(staged))
            (engineering / "materialization.json").write_text(
                dump_json(materialization), encoding="utf-8"
            )
        (engineering / "project.json").write_text(dump_json(manifest), encoding="utf-8")
        write_handoff(staged, manifest, definition_for_output)
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
    required_fields = {
        "schema_version",
        "platform_version",
        "project",
        "recipe",
        "starters",
        "features",
        "capability_status",
        "gates",
        "ownership",
    }
    missing_fields = sorted(required_fields.difference(manifest))
    if (
        "capability_status" in missing_fields
        and manifest.get("platform_version") != PLATFORM_VERSION
    ):
        missing_fields.remove("capability_status")
        warnings.append(
            "Manifest anterior sin capability_status; se completará al aplicar el próximo cambio"
        )
    if missing_fields:
        errors.append("Manifest incompleto: " + ", ".join(missing_fields))
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
        if manifest.get("scaffold_status") == "materialized" and starter.get("adapter"):
            adapter = read_json(ROOT / starter["adapter"])
            for requirement in adapter_preflight(adapter):
                if not requirement["ok"]:
                    expected = requirement.get("expected")
                    detail = "ausente" if not requirement["available"] else requirement.get("version", "incompatible")
                    errors.append(
                        f"Runtime de {starter.get('id')}: {requirement['executable']} "
                        f"{detail}" + (f"; esperado {expected}" if expected else "")
                    )
    if manifest.get("scaffold_status") == "materialized":
        materialization_path = project / ".engineering/materialization.json"
        if not materialization_path.exists():
            errors.append("scaffold materialized sin materialization.json")
        elif manifest.get("readiness") not in {"code-ready", "verified"}:
            errors.append(f"readiness inválido: {manifest.get('readiness')}")
        if not (project / ".github/workflows/engineering.yml").is_file():
            message = "scaffold materialized sin CI raíz de Engineering"
            if manifest.get("platform_version") == PLATFORM_VERSION:
                errors.append(message)
            else:
                warnings.append(message + "; se creará al aplicar el próximo cambio")
    if manifest.get("database") is not None and manifest.get("database") not in databases:
        errors.append(f"Perfil de base de datos inexistente: {manifest.get('database')}")
    if recipe and manifest.get("database") not in recipe.get("stack", {}).get("database", {}).get("allowed", []):
        if manifest.get("database") is not None:
            errors.append(f"Base de datos fuera de la Recipe: {manifest.get('database')}")
    for feature in manifest.get("features", []):
        if feature not in features:
            errors.append(f"Feature inexistente: {feature}")
    capability_status = manifest.get("capability_status", {})
    if not isinstance(capability_status, dict):
        errors.append("capability_status debe ser un objeto")
        capability_status = {}
    allowed_capability_states = {
        "planned",
        "pending-implementation",
        "materialized",
        "verified",
    }
    for feature in manifest.get("features", []):
        status = capability_status.get(feature)
        if not isinstance(status, dict):
            if manifest.get("platform_version") == PLATFORM_VERSION:
                errors.append(f"Falta estado real para la capacidad {feature}")
        elif status.get("state") not in allowed_capability_states:
            errors.append(f"Estado inválido para {feature}: {status.get('state')}")
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
        if handoff.get("schema_version") == 1:
            if handoff.get("stack", {}).get("recipe", {}).get("id") != manifest.get("recipe", {}).get("id"):
                errors.append("El handoff de Gentle y el manifest usan Recipes distintas")
            if handoff.get("stack", {}).get("skills") != manifest.get("skills"):
                errors.append("El handoff de Gentle y el manifest declaran skills distintos")
        elif handoff.get("schema_version") == 2:
            sources = handoff.get("source_of_truth", {})
            if sources.get("decisions") != ".engineering/project.json":
                errors.append("El handoff no apunta al manifest como fuente de decisiones")
            if manifest.get("scaffold_status") == "materialized" and sources.get("materialization") != ".engineering/materialization.json":
                errors.append("El handoff no apunta al registro de materialización")
        else:
            errors.append(f"schema_version de handoff no soportado: {handoff.get('schema_version')}")
        if handoff.get("strategy", {}).get("owner") != "gentle-ai":
            errors.append("El handoff no delega la estrategia de desarrollo a Gentle AI")
    return manifest, errors, warnings


def _infer_capability_status(manifest: dict[str, Any]) -> dict[str, dict[str, Any]]:
    statuses: dict[str, dict[str, Any]] = {}
    verified = manifest.get("readiness") == "verified"
    for feature_id in manifest.get("features", []):
        providers = [
            starter["id"]
            for starter in manifest.get("starters", [])
            if feature_id in starter.get("provided_features", [])
        ]
        statuses[feature_id] = {
            "state": ("verified" if verified else "materialized") if providers else "pending-implementation",
            "provided_by": providers,
            "owner": "boilerplate" if providers else "gentle-ai",
        }
    return statuses


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
    updated_starter_ids = {item["id"] for item in updated_manifest["starters"]}
    updated_manifest["starters"].extend(
        deepcopy(item)
        for item in current_manifest.get("starters", [])
        if item.get("id") not in updated_starter_ids
    )
    if current_manifest.get("extensions"):
        updated_manifest["extensions"] = deepcopy(current_manifest["extensions"])
    updated_manifest["gates"] = unique(
        updated_manifest.get("gates", []) + current_manifest.get("gates", [])
    )
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
        updated_manifest["ownership"]["managed"].extend(
            [".engineering/materialization.json", ".github/workflows/engineering.yml"]
        )
    added = [
        item
        for item in updated_manifest["features"]
        if item not in current_manifest.get("features", [])
    ]
    for existing_feature in current_manifest.get("features", []):
        if existing_feature in current_manifest.get("capability_status", {}):
            updated_manifest["capability_status"][existing_feature] = deepcopy(
                current_manifest["capability_status"][existing_feature]
            )
    for added_feature in added:
        updated_manifest["capability_status"][added_feature] = {
            "state": "pending-implementation",
            "provided_by": [],
            "owner": "gentle-ai",
        }
    implementation_required = [
        feature_id
        for feature_id in added
        if updated_manifest["capability_status"][feature_id]["state"]
        == "pending-implementation"
    ]
    result = {
        "changed": True,
        "applied": apply,
        "feature": feature_id,
        "requested_with_dependencies": added,
        "implementation_required": implementation_required,
        "materialized": [],
        "new_gates": sorted(
            set(updated_manifest["gates"]).difference(current_manifest.get("gates", []))
        ),
        "warnings": updated_manifest["warnings"],
    }
    if apply:
        if updated_manifest.get("scaffold_status") == "materialized":
            materialization = _read_materialization(project, current_manifest)
            materialization.update(_actual_project_data(project))
            (engineering / "materialization.json").write_text(
                dump_json(materialization), encoding="utf-8"
            )
            write_project_ci(project, updated_manifest, materialization)
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
        result["warning"] = (
            "--apply registró la capacidad y el trabajo pendiente; no afirmó que el código exista."
        )
    return result


def extend_project(
    project: Path,
    starter_id: str,
    *,
    apply: bool = False,
    skip_setup: bool = False,
    skip_checks: bool = False,
) -> dict[str, Any]:
    manifest, errors, warnings = inspect_project(project)
    if errors:
        raise PlatformError("El proyecto no pasa doctor: " + "; ".join(errors))
    if manifest.get("scaffold_status") != "materialized":
        raise PlatformError("eng extend requiere un proyecto materializado")
    if starter_id in {item.get("id") for item in manifest.get("starters", [])}:
        return {
            "changed": False,
            "starter": starter_id,
            "reason": "El starter ya forma parte del proyecto.",
        }

    starter, provided_features, unresolved = _starter_manifest_entry(
        starter_id,
        selected_features=manifest.get("features", []),
        database=manifest.get("database"),
    )
    adapter = read_json(ROOT / starter["adapter"])
    dependencies = adapter.get("project_dependencies", {})
    current_starters = {item["id"] for item in manifest.get("starters", [])}
    required_all = set(dependencies.get("all_of", []))
    if not required_all.issubset(current_starters):
        raise PlatformError(
            f"{starter_id} requiere starters: {', '.join(sorted(required_all - current_starters))}"
        )
    required_one = set(dependencies.get("one_of", []))
    if required_one and not current_starters.intersection(required_one):
        raise PlatformError(
            f"{starter_id} requiere uno de: {', '.join(sorted(required_one))}"
        )

    extension_destination = adapter.get("extension_destination")
    destination_value = extension_destination or starter.get("destination")
    if destination_value in {None, "."}:
        raise PlatformError(
            f"{starter_id} ocupa la raíz y no declara extension_destination; "
            "no se agregará sobre un proyecto existente"
        )
    destination = _safe_relative(destination_value, "extension_destination")
    target = project / destination
    if target.is_symlink() or (target.exists() and not target.is_dir()):
        raise PlatformError(f"El destino de extensión no es un directorio: {destination}")
    if target.exists() and any(target.iterdir()):
        raise PlatformError(f"El destino de extensión no está vacío: {destination}")
    starter["destination"] = destination.as_posix()
    starter["destination_override"] = destination.as_posix()
    requirement_status = adapter_preflight(adapter)
    requirement_errors = [item for item in requirement_status if not item["ok"]]

    result: dict[str, Any] = {
        "changed": True,
        "applied": apply,
        "starter": starter_id,
        "destination": destination.as_posix(),
        "provided_features": provided_features,
        "pending_features": unresolved,
        "requirements": requirement_status,
        "added": [destination.as_posix()],
        "updated": [
            ".engineering/project.json",
            ".engineering/materialization.json",
            "ARCHITECTURE.md",
            "AGENTS.md",
            "GENTLE.md",
            ".github/workflows/engineering.yml",
        ],
        "preserved": [
            item.get("destination") for item in manifest.get("starters", [])
        ],
        "warnings": warnings,
    }
    if requirement_errors:
        result["blocked"] = True
        if apply:
            _assert_adapter_requirements(adapter)
        return result
    if not apply:
        return result

    _assert_adapter_requirements(adapter)
    old_materialization = _read_materialization(project, manifest)
    with tempfile.TemporaryDirectory(
        prefix=f".{project.name}-{starter_id}-extend-", dir=project.parent
    ) as temporary:
        staged_root = Path(temporary) / "staged"
        staged_root.mkdir()
        partial_manifest = {
            "project": manifest["project"],
            "starters": [starter],
        }
        added_materialization = materialize_project(
            partial_manifest,
            staged_root,
            skip_setup=skip_setup,
            skip_checks=skip_checks,
        )
        staged_destination = staged_root / destination
        if not staged_destination.is_dir():
            raise PlatformError(f"La extensión no produjo {destination}")
        target.parent.mkdir(parents=True, exist_ok=True)
        if target.exists():
            target.rmdir()
        staged_destination.rename(target)

        updated_manifest = deepcopy(manifest)
        updated_manifest["$schema"] = PROJECT_SCHEMA_URL
        updated_manifest["platform_version"] = PLATFORM_VERSION
        if not isinstance(updated_manifest.get("capability_status"), dict):
            updated_manifest["capability_status"] = _infer_capability_status(updated_manifest)
        updated_manifest["starters"].append(starter)
        updated_manifest.setdefault("extensions", []).append(
            {
                "starter": starter_id,
                "destination": destination.as_posix(),
                "applied_at": date.today().isoformat(),
            }
        )
        updated_manifest.setdefault("ownership", {}).setdefault("managed", [])
        updated_manifest["ownership"]["managed"] = unique(
            updated_manifest["ownership"]["managed"]
            + [".engineering/materialization.json", ".github/workflows/engineering.yml"]
        )
        updated_manifest["gates"] = unique(
            updated_manifest.get("gates", [])
            + [
                check["gate"]
                for check in adapter.get("materializer", {}).get("checks", [])
                if check.get("gate")
            ]
        )
        extension_readiness = added_materialization.get("readiness", "code-ready")
        updated_manifest["readiness"] = (
            "verified"
            if manifest.get("readiness") == "verified" and extension_readiness == "verified"
            else "code-ready"
        )
        for feature_id in provided_features:
            if feature_id not in updated_manifest.get("capability_status", {}):
                continue
            status = updated_manifest["capability_status"][feature_id]
            providers = unique(status.get("provided_by", []) + [starter_id])
            status.update(
                {
                    "state": "verified" if updated_manifest["readiness"] == "verified" else "materialized",
                    "provided_by": providers,
                    "owner": "boilerplate",
                }
            )

        updated_materialization = deepcopy(old_materialization)
        for key in ("starters", "setup", "checks", "requirements"):
            updated_materialization[key] = updated_materialization.get(key, []) + added_materialization.get(key, [])
        updated_materialization["platform_version"] = PLATFORM_VERSION
        updated_materialization["readiness"] = updated_manifest["readiness"]
        updated_materialization.update(_actual_project_data(project))

        managed_paths = [
            project / ".engineering/project.json",
            project / ".engineering/materialization.json",
            project / "ARCHITECTURE.md",
            project / "AGENTS.md",
            project / "GENTLE.md",
            project / ".engineering/gentle-handoff.json",
            project / ".github/workflows/engineering.yml",
        ]
        previous = {
            path: path.read_bytes() if path.exists() else None for path in managed_paths
        }
        try:
            _write_json_atomic(project / ".engineering/project.json", updated_manifest)
            _write_json_atomic(
                project / ".engineering/materialization.json", updated_materialization
            )
            _write_text_atomic(project / "ARCHITECTURE.md", architecture_markdown(updated_manifest))
            _write_text_atomic(project / "AGENTS.md", agents_markdown(updated_manifest))
            definition_path = project / ".engineering/project-definition.json"
            definition = (
                validate_project_definition(read_json(definition_path))
                if definition_path.exists()
                else None
            )
            write_handoff(project, updated_manifest, definition)
            write_project_ci(project, updated_manifest, updated_materialization)
        except Exception:
            if target.exists():
                shutil.rmtree(target)
            for path, content in previous.items():
                if content is None:
                    if path.exists():
                        path.unlink()
                else:
                    path.parent.mkdir(parents=True, exist_ok=True)
                    path.write_bytes(content)
            raise
    result["readiness"] = updated_manifest["readiness"]
    result["checks"] = added_materialization.get("checks", [])
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


def command_boilerplate_verify(args: argparse.Namespace) -> int:
    result = verify_boilerplate(args.id)
    print(dump_json(result), end="")
    return 0 if result["ok"] else 1


def command_boilerplate_add(args: argparse.Namespace) -> int:
    entry = read_json(Path(args.input).resolve())
    required = {
        "id",
        "legacy_ids",
        "name",
        "kind",
        "decision_status",
        "delivery_status",
        "maintenance_tier",
        "category",
        "repository",
        "profile",
        "use_when",
        "avoid_when",
        "integration",
    }
    missing = sorted(required.difference(entry))
    if missing:
        raise PlatformError(f"Entrada incompleta: {', '.join(missing)}")
    registry_path = ROOT / "platform/boilerplates.json"
    registry = read_json(registry_path)
    entries = registry["entries"]
    if entry["id"] in by_id(entries):
        raise PlatformError(f"Ya existe el boilerplate {entry['id']}")
    if entry.get("repository"):
        normalized = normalize_repository(entry["repository"])
        if any(
            item.get("repository")
            and normalize_repository(item["repository"]) == normalized
            for item in entries
        ):
            raise PlatformError("El repositorio ya está registrado")
    result = {"action": "add", "id": entry["id"], "apply": bool(args.apply)}
    if args.apply:
        entries.append(entry)
        registry["reviewed_at"] = date.today().isoformat()
        _write_json_atomic(registry_path, registry)
    print(dump_json(result), end="")
    return 0


def command_boilerplate_remove(args: argparse.Namespace) -> int:
    registry_path = ROOT / "platform/boilerplates.json"
    registry = read_json(registry_path)
    entries = by_id(registry["entries"])
    if args.id not in entries:
        raise PlatformError(f"Boilerplate inexistente: {args.id}")
    references = boilerplate_references(args.id)
    if references:
        raise PlatformError(
            f"{args.id} sigue referenciado por {', '.join(references)}; "
            "actualiza primero las Recipes"
        )
    result = {"action": "remove", "id": args.id, "apply": bool(args.apply)}
    if args.apply:
        registry["entries"] = [item for item in registry["entries"] if item["id"] != args.id]
        registry["reviewed_at"] = date.today().isoformat()
        _write_json_atomic(registry_path, registry)
    print(dump_json(result), end="")
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
        materialization = _read_materialization(project, manifest)
        materialization.update(_actual_project_data(project))
        (project / ".engineering/materialization.json").write_text(
            dump_json(materialization), encoding="utf-8"
        )
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


def _user_engineering_platform_git_sources(output: str) -> list[str]:
    sources: list[str] = []
    in_user_packages = False
    for line in output.splitlines():
        if line.strip() == "User packages:":
            in_user_packages = True
            continue
        if not in_user_packages:
            continue
        if line and not line[0].isspace():
            break
        indentation = len(line) - len(line.lstrip())
        tokens = line.strip().split()
        if indentation != 2 or not tokens:
            continue
        source = tokens[0]
        if source.startswith(PI_ENGINEERING_PLATFORM_GIT_SOURCE_PREFIX):
            sources.append(source)
    return unique(sources)


def _remove_pi_package(pi_executable: str, source: str) -> None:
    completed = subprocess.run(
        [pi_executable, "remove", source],
        text=True,
        capture_output=True,
        check=False,
    )
    if completed.returncode == 0:
        return
    detail = completed.stderr.strip() or completed.stdout.strip() or "sin salida"
    if detail == f"No matching package found for {source}":
        return
    raise PlatformError(
        f"Pi no pudo retirar {source}; no se borraron archivos: {detail[-1000:]}"
    )


def _retire_conflicting_engineering_platform_sources(pi_executable: str) -> None:
    listed = subprocess.run(
        [pi_executable, "list", "--no-approve"],
        text=True,
        capture_output=True,
        check=False,
    )
    if listed.returncode != 0:
        detail = listed.stderr.strip() or listed.stdout.strip() or "sin salida"
        raise PlatformError(
            f"Pi no pudo listar los paquetes de usuario; no se retiraron paquetes: {detail[-1000:]}"
        )
    for source in _user_engineering_platform_git_sources(listed.stdout):
        _remove_pi_package(pi_executable, source)


def _retire_stale_managed_installations(
    pi_executable: str, home: Path, install_root: Path
) -> list[Path]:
    stale_installations = [
        installation
        for installation in _managed_installations(home)
        if installation != install_root
    ]
    for installation in stale_installations:
        _remove_pi_package(pi_executable, str(installation))
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
            shutil.copytree(ROOT, staged, ignore=_install_ignores)
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
    _retire_conflicting_engineering_platform_sources(pi_executable)
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


def _execution_record_key(record: dict[str, Any]) -> tuple[Any, ...]:
    return (
        record.get("starter"),
        record.get("gate"),
        tuple(record.get("command", [])),
        record.get("workdir", "."),
    )


def _execution_record_passed(record: dict[str, Any]) -> bool:
    # returncode keeps compatibility with records written before status existed.
    return record.get("status") == "passed" or record.get("returncode") == 0


def _execution_record_environment(record: dict[str, Any]) -> dict[str, str]:
    if "environment" not in record:
        return {}
    return _validate_environment(record["environment"], "materialization.environment")


def command_check(args: argparse.Namespace) -> int:
    manifest, errors, warnings = inspect_project(Path(args.project).resolve())
    if args.run and errors:
        raise PlatformError("El proyecto no pasa doctor: " + "; ".join(errors))
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
        materialization = _read_materialization(Path(args.project).resolve(), manifest)
        executed_setup: list[dict[str, Any]] = []
        executed_checks: list[dict[str, Any]] = []
        project = Path(args.project).resolve()
        setup_updates: dict[tuple[Any, ...], dict[str, Any]] = {}
        for registered in materialization.get("setup", []):
            environment = _execution_record_environment(registered)
            if not getattr(args, "force_setup", False) and _execution_record_passed(registered):
                executed_setup.append(registered)
                continue
            record = _run_adapter_record(
                registered["command"],
                project / _safe_relative(registered.get("workdir", "."), "workdir"),
                environment=environment,
            )
            _add_declared_environment(record, environment)
            record.update({"starter": registered.get("starter"), "workdir": registered.get("workdir", ".")})
            executed_setup.append(record)
            setup_updates[_execution_record_key(record)] = record
        materialization["setup"] = [
            setup_updates.get(_execution_record_key(registered), registered)
            for registered in materialization.get("setup", [])
        ]
        check_updates: dict[tuple[Any, ...], dict[str, Any]] = {}
        for registered in materialization.get("checks", []):
            gate = registered.get("gate")
            if files and gate not in gates:
                continue
            environment = _execution_record_environment(registered)
            record = _run_adapter_record(
                registered["command"],
                project / _safe_relative(registered.get("workdir", "."), "workdir"),
                gate=gate,
                environment=environment,
            )
            _add_declared_environment(record, environment)
            record.update({"starter": registered.get("starter"), "workdir": registered.get("workdir", ".")})
            executed_checks.append(record)
            check_updates[_execution_record_key(record)] = record
        materialization["checks"] = [
            check_updates.get(_execution_record_key(registered), registered)
            for registered in materialization.get("checks", [])
        ]
        materialization.update(_actual_project_data(project))
        if not files and executed_checks:
            materialization["readiness"] = "verified"
            manifest["readiness"] = "verified"
            for status in manifest.get("capability_status", {}).values():
                if status.get("state") == "materialized":
                    status["state"] = "verified"
        elif files and executed_checks:
            # A filtered run provides evidence for selected gates only.
            materialization["readiness"] = "code-ready"
            manifest["readiness"] = "code-ready"
        materialization_path.write_text(dump_json(materialization), encoding="utf-8")
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


def command_extend(args: argparse.Namespace) -> int:
    result = extend_project(
        Path(args.project).resolve(),
        args.starter,
        apply=args.apply,
        skip_setup=args.skip_setup,
        skip_checks=args.skip_checks,
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
            "action": "review-required",
            "applied": False,
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

    verify = boilerplate_sub.add_parser("verify", help="Verificar contrato, evidencia y referencias")
    verify.add_argument("id")
    verify.set_defaults(handler=command_boilerplate_verify)

    add_boilerplate = boilerplate_sub.add_parser("add", help="Registrar una entrada declarativa")
    add_boilerplate.add_argument("--from", dest="input", required=True)
    add_boilerplate.add_argument("--apply", action="store_true")
    add_boilerplate.set_defaults(handler=command_boilerplate_add)

    remove_boilerplate = boilerplate_sub.add_parser("remove", help="Retirar una entrada no referenciada")
    remove_boilerplate.add_argument("id")
    remove_boilerplate.add_argument("--apply", action="store_true")
    remove_boilerplate.set_defaults(handler=command_boilerplate_remove)

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
    check.add_argument("--force-setup", action="store_true", help="Repetir la instalación aunque ya haya pasado")
    check.set_defaults(handler=command_check)

    add = sub.add_parser("add", help="Planear o aplicar un feature pack")
    add.add_argument("feature")
    add.add_argument("--project", default=".")
    add.add_argument("--apply", action="store_true", help="Actualizar intake, manifest y secciones gestionadas")
    add.set_defaults(handler=command_add)

    extend = sub.add_parser(
        "extend", help="Planear o agregar un starter a un proyecto existente"
    )
    extend.add_argument("starter")
    extend.add_argument("--project", default=".")
    extend.add_argument("--apply", action="store_true")
    extend.add_argument("--skip-setup", action="store_true")
    extend.add_argument("--skip-checks", action="store_true")
    extend.set_defaults(handler=command_extend)

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
