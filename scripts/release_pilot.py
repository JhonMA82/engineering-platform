#!/usr/bin/env python3
"""Materialize one stable Golden Path for scheduled release evidence."""

from __future__ import annotations

import argparse
import json
from pathlib import Path

try:
    from scripts.eng import PlatformError, inspect_project, write_project
except ModuleNotFoundError:  # Direct execution outside the repository root.
    from eng import PlatformError, inspect_project, write_project


PILOTS = {
    "GP-01": {
        "name": "pilot-public-web",
        "project_type": "public-web",
        "signals": ["content", "marketing"],
        "features": [],
        "excluded_features": [],
        "database": None,
    },
    "GP-02": {
        "name": "pilot-admin",
        "project_type": "institutional-admin",
        "signals": ["workflow", "requests"],
        "features": [],
        "excluded_features": ["multitenancy", "jobs"],
        "database": "postgresql-managed",
    },
    "GP-03": {
        "name": "pilot-python-data",
        "project_type": "python-data",
        "signals": ["etl", "api"],
        "features": [],
        "excluded_features": [],
        "database": "postgresql-managed",
    },
    "GP-04": {
        "name": "pilot-mobile",
        "project_type": "mobile",
        "signals": ["mobile", "api"],
        "features": [],
        "excluded_features": [],
        "database": "postgresql-managed",
    },
    "GP-05": {
        "name": "pilot-desktop",
        "project_type": "desktop",
        "signals": ["desktop", "local"],
        "features": [],
        "excluded_features": [],
        "database": "sqlite-local",
    },
    "GP-06": {
        "name": "pilot-multi-app",
        "project_type": "multi-app",
        "signals": ["shared-api", "multiple-clients"],
        "features": [],
        "excluded_features": ["multitenancy", "jobs"],
        "database": "postgresql-managed",
    },
}


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("recipe", choices=sorted(PILOTS))
    parser.add_argument("--output", required=True)
    args = parser.parse_args()
    output = Path(args.output).resolve()
    if output.exists():
        raise PlatformError(f"El destino del piloto ya existe: {output}")
    manifest = write_project(PILOTS[args.recipe], output, materialize=True)
    if manifest["recipe"]["id"] != args.recipe:
        raise PlatformError(
            f"El piloto {args.recipe} resolvió {manifest['recipe']['id']}"
        )
    _, errors, warnings = inspect_project(output)
    if errors:
        raise PlatformError("; ".join(errors))
    print(
        json.dumps(
            {
                "recipe": args.recipe,
                "starters": [item["id"] for item in manifest["starters"]],
                "readiness": manifest["readiness"],
                "warnings": warnings,
            },
            indent=2,
        )
    )
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except PlatformError as exc:
        raise SystemExit(f"ERROR: {exc}") from exc
