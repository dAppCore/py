from __future__ import annotations

from pathlib import Path
import sys
import tempfile
import unittest

from core import config, data, echo, err, fs, json, log, medium, options, process, service


class CorePyTests(unittest.TestCase):
    def test_echo_and_json_round_trip(self) -> None:
        self.assertEqual(echo("hello"), "hello")

        with tempfile.TemporaryDirectory() as directory_name:
            filename = Path(directory_name) / "config.json"
            fs.write_file(filename, json.dumps({"name": "corepy"}))
            payload = fs.read_file(filename)
            self.assertEqual(json.loads(payload)["name"], "corepy")

    def test_options_and_config(self) -> None:
        values = options.Options({"name": "corepy", "port": 8080})
        values.set("debug", True)
        self.assertEqual(values.string("name"), "corepy")
        self.assertEqual(values.int("port"), 8080)
        self.assertTrue(values.bool("debug"))

        runtime_config = config.Config()
        runtime_config.set("debug", True)
        runtime_config.enable("tier1")
        self.assertTrue(runtime_config.bool("debug"))
        self.assertTrue(runtime_config.enabled("tier1"))
        self.assertEqual(runtime_config.enabled_features(), ["tier1"])

    def test_data_and_service_registry(self) -> None:
        assets = data.Data()
        with tempfile.TemporaryDirectory() as directory_name:
            fixture_directory = Path(directory_name) / "fixtures"
            fixture_directory.mkdir()
            (fixture_directory / "note.txt").write_text("hello from data", encoding="utf-8")
            assets.mount("fixtures", fixture_directory)
            self.assertEqual(assets.read_string("fixtures/note.txt"), "hello from data")
            self.assertEqual(assets.list_names("fixtures"), ["note"])

        registry = service.ServiceRegistry()
        registry.register("brain", service.Service(name="brain"))
        self.assertEqual(registry.names(), ["brain"])

    def test_medium_process_log_and_errors(self) -> None:
        buffer = medium.memory("hello")
        self.assertEqual(buffer.read_text(), "hello")
        buffer.write_text("updated")
        self.assertEqual(buffer.read_text(), "updated")

        output = process.run(sys.executable, "-c", "print('ok')")
        self.assertEqual(output.strip(), "ok")

        issue = err.e("core.test", "boom")
        wrapped = err.wrap(issue, "core.outer", "outer boom")
        self.assertIsNotNone(wrapped)
        assert wrapped is not None
        self.assertEqual(err.operation(wrapped), "core.outer")
        self.assertEqual(err.message(wrapped), "outer boom")
        self.assertEqual(str(wrapped), "core.outer: outer boom: core.test: boom")

        log.set_level("debug")
        log.info("corepy test", "module", "core")


if __name__ == "__main__":
    unittest.main()
