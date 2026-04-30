from __future__ import annotations

import importlib
from pathlib import Path
import sys
import unittest

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))

from core import agent, api, container, mcp, store, ws


class RFCStubModuleTests(unittest.TestCase):
    def test_rfc_stub_modules_available_good(self) -> None:
        modules = [agent, api, container, mcp, store, ws]
        for module in modules:
            imported = importlib.import_module(module.__name__)
            self.assertIs(imported, module)
            self.assertFalse(module.available())


if __name__ == "__main__":
    unittest.main()
