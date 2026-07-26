#!/usr/bin/env python3
"""Dependency-free tests for the offline fitters.

Run either with the stdlib test runner (no pytest needed):
    python3 -m unittest discover tools/tutor
or directly:
    python3 tools/tutor/test_tutor.py
"""
import math
import unittest

from calibrate_irt import calibrate_irt
from fit_bkt import fit_bkt, sequence_log_likelihood


class TestFitBKT(unittest.TestCase):
    def test_one_step_probabilities_sum_to_one(self):
        p = (0.3, 0.1, 0.1, 0.2)
        pt = math.exp(sequence_log_likelihood([True], *p))
        pf = math.exp(sequence_log_likelihood([False], *p))
        self.assertAlmostEqual(pt + pf, 1.0, places=9)

    def test_fit_beats_bad_baseline(self):
        seqs = [[False, True, True, True, True]] * 15 + [[True, True, True]] * 15
        fit = fit_bkt(seqs)
        good = sum(sequence_log_likelihood(s, fit["p_init"], fit["p_learn"],
                                           fit["p_slip"], fit["p_guess"]) for s in seqs)
        bad = sum(sequence_log_likelihood(s, 0.5, 0.01, 0.45, 0.45) for s in seqs)
        self.assertGreater(good, bad)
        self.assertLessEqual(fit["p_slip"], 0.45)
        self.assertLessEqual(fit["p_guess"], 0.45)
        self.assertEqual(fit["sequences_fit"], 30)

    def test_empty_rejected(self):
        with self.assertRaises(ValueError):
            fit_bkt([])


class TestCalibrateIRT(unittest.TestCase):
    def _data(self):
        rows = []
        for _ in range(10):
            rows.append({"easy": True, "hard": True, "disc": True, "noise": True})
        for _ in range(10):
            rows.append({"easy": True, "hard": False, "disc": False, "noise": True})
        return rows

    def test_difficulty_order(self):
        by = {it["question_id"]: it for it in calibrate_irt(self._data())["calibrated"]}
        self.assertLess(by["easy"]["difficulty"], by["hard"]["difficulty"])

    def test_discrimination(self):
        by = {it["question_id"]: it for it in calibrate_irt(self._data())["calibrated"]}
        self.assertGreater(by["disc"]["discrimination"], by["noise"]["discrimination"])
        self.assertEqual(by["noise"]["discrimination"], 0.0)

    def test_empty_rejected(self):
        with self.assertRaises(ValueError):
            calibrate_irt([])


if __name__ == "__main__":
    unittest.main()
