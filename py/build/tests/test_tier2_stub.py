from __future__ import annotations

import importlib
import sys
from pathlib import Path
import unittest


class Tier2StubTests(unittest.TestCase):
    def test_import_core_raises_typed_error_good(self) -> None:
        build_root = Path(__file__).resolve().parents[1]
        sys.path.insert(0, str(build_root))
        try:
            from _build_stub import Tier2BuildUnavailableError

            with self.assertRaisesRegex(Tier2BuildUnavailableError, "gopy and CPython 3.13"):
                importlib.import_module("core")
        finally:
            sys.path.remove(str(build_root))
            sys.modules.pop("core", None)
            sys.modules.pop("_build_stub", None)


if __name__ == "__main__":
    unittest.main()
