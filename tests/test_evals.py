from __future__ import annotations

import json
import unittest
from pathlib import Path

from scripts.eng import evaluate_boilerplate, resolve_recipe


ROOT = Path(__file__).resolve().parents[1]


class DecisionEvalTests(unittest.TestCase):
    def test_recipe_resolution_cases(self) -> None:
        cases = json.loads((ROOT / "evals/recipe-resolution.json").read_text(encoding="utf-8"))["cases"]
        for case in cases:
            with self.subTest(case=case["id"]):
                result = resolve_recipe(case["intake"])
                self.assertEqual(result["recipe"]["id"], case["expected_recipe"])
                self.assertEqual(result["database"], case["expected_database"])
                self.assertEqual(result["features"], case["expected_features"])

    def test_discovery_cases_resolve_to_expected_recipe(self) -> None:
        cases = json.loads((ROOT / "evals/discovery-cases.json").read_text(encoding="utf-8"))["cases"]
        for case in cases:
            with self.subTest(case=case["id"]):
                result = resolve_recipe(case["expected_intake"])
                self.assertEqual(result["recipe"]["id"], case["expected_recipe"])
                self.assertEqual(result["database"], case["expected_database"])
                self.assertEqual(result["features"], case["expected_features"])
                if "expected_starters" in case:
                    self.assertEqual(
                        [item["id"] for item in result["starters"]],
                        case["expected_starters"],
                    )

    def test_boilerplate_curator_cases(self) -> None:
        cases = json.loads((ROOT / "evals/boilerplate-curation.json").read_text(encoding="utf-8"))["cases"]
        for case in cases:
            with self.subTest(case=case["id"]):
                result = evaluate_boilerplate(
                    case["repository"],
                    observed_commit=case.get("observed_commit"),
                    category=case.get("category"),
                )
                self.assertEqual(result["decision"], case["expected_decision"])
                if "expected_entry" in case:
                    self.assertEqual(result["entry_id"], case["expected_entry"])
                if "expected_alternative" in case:
                    self.assertIn(case["expected_alternative"], result["compare_with"])


if __name__ == "__main__":
    unittest.main()
