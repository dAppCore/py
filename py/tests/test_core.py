from __future__ import annotations

from pathlib import Path
import sys
import tempfile
import unittest

from core import config, data, echo, err, fs, json, log, math as core_math, medium, options, path, process, service, strings


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

    def test_module_level_surface_matches_tier1_shape(self) -> None:
        values = options.new({"name": "corepy", "port": 8080})
        options.set(values, "debug", True)
        self.assertEqual(options.string(values, "name"), "corepy")
        self.assertEqual(options.int(values, "port"), 8080)
        self.assertTrue(options.bool(values, "debug"))
        self.assertEqual(options.items(values)["name"], "corepy")

        runtime_config = config.new()
        config.set(runtime_config, "debug", True)
        config.enable(runtime_config, "tier1")
        self.assertTrue(config.bool(runtime_config, "debug"))
        self.assertTrue(config.enabled(runtime_config, "tier1"))
        self.assertEqual(config.enabled_features(runtime_config), ["tier1"])

        assets = data.new()
        with tempfile.TemporaryDirectory() as directory_name:
            fixture_directory = Path(directory_name) / "fixtures"
            fixture_directory.mkdir()
            (fixture_directory / "note.txt").write_text("hello from data", encoding="utf-8")
            data.mount(assets, "fixtures", fixture_directory)
            self.assertEqual(data.read_string(assets, "fixtures/note.txt"), "hello from data")
            self.assertEqual(data.list_names(assets, "fixtures"), ["note"])
            self.assertEqual(data.mounts(assets), ["fixtures"])

        registry = service.new("corepy")
        service.register(registry, "brain")
        self.assertEqual(service.names(registry), ["cli", "brain"])

    def test_data_and_service_registry(self) -> None:
        assets = data.Data()
        with tempfile.TemporaryDirectory() as directory_name:
            fixture_directory = Path(directory_name) / "fixtures"
            fixture_directory.mkdir()
            (fixture_directory / "note.txt").write_text("hello from data", encoding="utf-8")
            template_directory = fixture_directory / "templates"
            template_directory.mkdir()
            (template_directory / "greeting-{{.Name}}.txt.tmpl").write_text("hello {{.Name}}", encoding="utf-8")
            assets.mount("fixtures", fixture_directory)
            self.assertEqual(assets.read_string("fixtures/note.txt"), "hello from data")
            self.assertEqual(assets.list_names("fixtures"), ["note", "templates"])
            workspace = Path(directory_name) / "workspace"
            self.assertEqual(assets.extract("fixtures/templates", workspace, {"Name": "corepy"}), str(workspace.resolve()))
            self.assertEqual((workspace / "greeting-corepy.txt").read_text(encoding="utf-8"), "hello corepy")

        registry = service.ServiceRegistry()
        registry.register("brain", service.Service(name="brain"))
        self.assertEqual(registry.names(), ["cli", "brain"])

    def test_medium_process_log_and_errors(self) -> None:
        buffer = medium.memory("hello")
        self.assertEqual(buffer.read_text(), "hello")
        buffer.write_text("updated")
        self.assertEqual(buffer.read_text(), "updated")
        medium.write_text(buffer, "via module")
        self.assertEqual(medium.read_text(buffer), "via module")

        output = process.run(sys.executable, "-c", "print('ok')")
        self.assertEqual(output.strip(), "ok")
        env_output = process.run_with_env(Path.cwd(), ["COREPY_MODE=test"], sys.executable, "-c", "import os; print(os.environ['COREPY_MODE'])")
        self.assertEqual(env_output.strip(), "test")
        self.assertTrue(process.exists())

        issue = err.e("core.test", "boom", None, "BOOM")
        wrapped = err.wrap(issue, "core.outer", "outer boom", "OUTER")
        self.assertIsNotNone(wrapped)
        assert wrapped is not None
        self.assertEqual(err.operation(wrapped), "core.outer")
        self.assertEqual(err.error_code(wrapped), "OUTER")
        self.assertEqual(err.message(wrapped), "outer boom")
        self.assertEqual(str(wrapped), "core.outer: outer boom [OUTER]: core.test: boom [BOOM]")

        log.set_level("debug")
        log.info("corepy test", "module", "core")
        log.set_level("quiet")
        with self.assertRaises(ValueError):
            log.set_level("verbose")

    def test_path_and_strings_helpers(self) -> None:
        self.assertEqual(path.join("deploy", "to", "homelab"), "deploy/to/homelab")
        self.assertEqual(path.base("/tmp/corepy/config.json"), "config.json")
        self.assertEqual(path.dir("/tmp/corepy/config.json"), "/tmp/corepy")
        self.assertEqual(path.ext("config.json"), ".json")
        self.assertFalse(path.is_abs("deploy/to/homelab"))
        self.assertEqual(path.clean("deploy//to/../from"), "deploy/from")

        self.assertTrue(strings.contains("hello world", "world"))
        self.assertEqual(strings.trim("  corepy  "), "corepy")
        self.assertEqual(strings.trim_prefix("--debug", "--"), "debug")
        self.assertEqual(strings.trim_suffix("config.json", ".json"), "config")
        self.assertEqual(strings.split_n("key=value=extra", "=", 2), ["key", "value=extra"])
        self.assertEqual(strings.join("/", "deploy", "to", "homelab"), "deploy/to/homelab")
        self.assertEqual(strings.concat("deploy", "/", "to"), "deploy/to")

    def test_math_surface(self) -> None:
        self.assertEqual(core_math.sort([3, 1, 2]), [1, 2, 3])
        self.assertEqual(core_math.binary_search([1, 2, 3], 2), 1)
        self.assertAlmostEqual(core_math.mean([1, 2, 3]), 2.0)
        self.assertAlmostEqual(core_math.median([1, 2, 3]), 2.0)
        self.assertAlmostEqual(core_math.variance([1, 2, 3]), 2.0 / 3.0)
        self.assertAlmostEqual(core_math.stdev([1, 2, 3]), (2.0 / 3.0) ** 0.5)
        self.assertTrue(core_math.epsilon_equal(0.1 + 0.2, 0.3, 1e-9))
        self.assertEqual(core_math.normalize([10, 20, 30]), [0.0, 0.5, 1.0])
        self.assertEqual(core_math.rescale([10, 20, 30], -1.0, 1.0), [-1.0, 0.0, 1.0])

        tree = core_math.kdtree.build([[0.0, 0.0], [1.0, 1.0], [3.0, 3.0]], metric="euclidean")
        neighbours = tree.nearest([0.8, 0.8], k=2)
        self.assertEqual([item["index"] for item in neighbours], [1, 0])

        cosine_neighbours = core_math.knn.search(
            [[1.0, 0.0], [0.0, 1.0], [0.8, 0.2]],
            [1.0, 0.0],
            k=2,
            metric="cosine",
        )
        self.assertEqual([item["index"] for item in cosine_neighbours], [0, 2])


if __name__ == "__main__":
    unittest.main()
