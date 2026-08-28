#!/usr/bin/env python3
"""Minimal CLI for selecting, recording and checking project recipes."""

from __future__ import annotations

import argparse
import json
import re
import sys
from copy import deepcopy
from datetime import date
from pathlib import Path
from typing import Any
from urllib.parse import urlparse


ROOT = Path(__file__).resolve().parents[1]
PLATFORM_VERSION = "0.3.0"


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
        starters.append(
            {
                "id": starter_id,
                "delivery_status": item["delivery_status"],
                "integration_mode": item["integration"]["mode"],
                "update_strategy": item["integration"]["update_strategy"],
                "pin": pin,
            }
        )
        if item["delivery_status"] != "released":
            warnings.append(
                f"{starter_id} está {item['delivery_status']}: el resultado es blueprint, no código productivo liberado."
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
        "scaffold_status": "materialized" if all(item["delivery_status"] == "released" for item in starters) else "blueprint",
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
    return f"""# Arquitectura: {manifest['project']['name']}

Documento generado desde la Recipe `{manifest['recipe']['id']}@{manifest['recipe']['version']}`. Las decisiones de dominio deben registrarse como ADR; no se cambia el stack silenciosamente.

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
- No agregues frameworks, bases de datos o feature packs sin actualizar primero el intake y resolver de nuevo la Recipe.
- Respeta ownership: no sobrescribas archivos `user_owned`; los cambios a `managed_sections` deben limitarse a su sección identificada.
"""


def readme_markdown(manifest: dict[str, Any]) -> str:
    return f"""# {manifest['project']['name']}

Proyecto definido por Engineering Platform `{manifest['platform_version']}` con Recipe `{manifest['recipe']['id']}`.

Estado: **{manifest['scaffold_status']}**. Consulta `ARCHITECTURE.md` y `.engineering/project.json` antes de materializar o desarrollar el producto.
"""


def write_project(intake: dict[str, Any], output: Path) -> dict[str, Any]:
    manifest = resolve_recipe(intake)
    if output.exists():
        if not output.is_dir():
            raise PlatformError(f"La salida existe y no es un directorio: {output}")
        if any(output.iterdir()):
            raise PlatformError(f"El directorio de salida no está vacío: {output}")
    engineering = output / ".engineering"
    engineering.mkdir(parents=True, exist_ok=True)
    (engineering / "project.json").write_text(dump_json(manifest), encoding="utf-8")
    (engineering / "intake.json").write_text(dump_json(intake), encoding="utf-8")
    (output / "ARCHITECTURE.md").write_text(architecture_markdown(manifest), encoding="utf-8")
    (output / "AGENTS.md").write_text(agents_markdown(manifest), encoding="utf-8")
    (output / "README.md").write_text(readme_markdown(manifest), encoding="utf-8")
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
        elif entry["delivery_status"] != "released":
            warnings.append(f"{entry['id']} continúa {entry['delivery_status']}")
            if manifest.get("scaffold_status") == "materialized":
                errors.append(
                    f"{entry['id']} no puede constar materialized mientras está {entry['delivery_status']}"
                )
        if starter.get("pin") == "PINNED":
            errors.append(f"{starter.get('id')} conserva el marcador PINNED")
        if manifest.get("scaffold_status") == "materialized" and not starter.get("pin"):
            errors.append(f"{starter.get('id')} materializado sin pin reproducible")
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
    for required in ["ARCHITECTURE.md", "AGENTS.md"]:
        if not (project / required).exists():
            warnings.append(f"Falta {required}")
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
        intake_path.write_text(dump_json(updated_intake), encoding="utf-8")
        (engineering / "project.json").write_text(dump_json(updated_manifest), encoding="utf-8")
        (project / "ARCHITECTURE.md").write_text(
            architecture_markdown(updated_manifest), encoding="utf-8"
        )
        (project / "AGENTS.md").write_text(
            agents_markdown(updated_manifest), encoding="utf-8"
        )
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


def command_doctor(args: argparse.Namespace) -> int:
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
    actions = [
        {
            "id": starter["id"],
            "pin": starter.get("pin"),
            "strategy": starter.get("update_strategy"),
            "action": "check-upstream" if args.check else "create-reviewed-update-branch",
        }
        for starter in manifest["starters"]
    ]
    print(dump_json({"actions": actions, "warnings": warnings}), end="")
    return 0


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(prog="eng", description="Engineering Platform: recetas mínimas y verificables")
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

    doctor = sub.add_parser("doctor", help="Revisar coherencia de un proyecto")
    doctor.add_argument("--project", default=".")
    doctor.set_defaults(handler=command_doctor)

    plan = sub.add_parser("plan", help="Seleccionar skills y gates para un cambio")
    plan.add_argument("--project", default=".")
    plan.add_argument("--change-type", choices=sorted(CHANGE_PLANS), required=True)
    plan.set_defaults(handler=command_plan)

    check = sub.add_parser("check", help="Seleccionar gates por manifest y archivos")
    check.add_argument("--project", default=".")
    check.add_argument("--changed-files", nargs="*")
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
