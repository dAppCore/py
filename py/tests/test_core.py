from __future__ import annotations

import importlib
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import unittest
from unittest.mock import patch

from core import action, array, cache, config, crypto, data, dns, echo, entitlement, err, fs, i18n, info, json, log, math as core_math, medium, options, path, process, registry, scm, service, strings, task


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

    def test_config_reads_environment_when_setting_missing(self) -> None:
        runtime_config = config.Config()

        with patch.dict(os.environ, {"DATABASE_HOST": "db.internal", "PORT": "8080", "DEBUG": "true"}, clear=False):
            self.assertEqual(runtime_config.get("database.host"), "db.internal")
            self.assertEqual(runtime_config.string("database.host"), "db.internal")
            self.assertEqual(runtime_config.int("port"), 8080)
            self.assertTrue(runtime_config.bool("debug"))

        runtime_config.set("database.host", "override.internal")
        runtime_config.set("port", 9000)
        runtime_config.set("debug", False)
        with patch.dict(os.environ, {"DATABASE_HOST": "db.internal", "PORT": "8080", "DEBUG": "true"}, clear=False):
            self.assertEqual(runtime_config.string("database.host"), "override.internal")
            self.assertEqual(runtime_config.int("port"), 9000)
            self.assertFalse(runtime_config.bool("debug"))

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
        self.assertEqual(core_math.moving_average([1, 3, 6, 10], window=2), [1.0, 2.0, 4.5, 8.0])
        self.assertEqual(core_math.difference([1, 3, 6, 10]), [2.0, 3.0, 4.0])

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

    def test_math_submodules_are_importable(self) -> None:
        kdtree_module = importlib.import_module("core.math.kdtree")
        knn_module = importlib.import_module("core.math.knn")
        signal_module = importlib.import_module("core.math.signal")

        tree = kdtree_module.build([[0.0, 0.0], [1.0, 1.0]], metric="euclidean")
        self.assertEqual([item["index"] for item in tree.nearest([0.8, 0.8], k=2)], [1, 0])

        neighbours = knn_module.search(
            [[1.0, 0.0], [0.0, 1.0], [0.8, 0.2]],
            [1.0, 0.0],
            k=2,
            metric="cosine",
        )
        self.assertEqual([item["index"] for item in neighbours], [0, 2])
        self.assertEqual(signal_module.moving_average([1, 3, 6, 10], window=2), [1.0, 2.0, 4.5, 8.0])
        self.assertEqual(signal_module.difference([1, 3, 6, 10], lag=2), [5.0, 7.0])

    def test_cache_crypto_and_dns_surface(self) -> None:
        with tempfile.TemporaryDirectory() as directory_name:
            store = cache.new(directory_name, 60)
            cache.set(store, "greeting", {"name": "corepy", "debug": True})
            self.assertTrue(cache.has(store, "greeting"))
            self.assertEqual(cache.get(store, "greeting")["name"], "corepy")
            self.assertEqual(cache.get(store, "missing", "fallback"), "fallback")
            cache.set_with_ttl(store, "nested/config", {"enabled": True}, 60)
            self.assertEqual(cache.keys(store), ["greeting", "nested/config"])
            self.assertEqual(cache.keys(store, "nested"), ["nested/config"])
            self.assertEqual(cache.clear(store, "nested"), 1)
            self.assertTrue(cache.delete(store, "greeting"))
            self.assertFalse(cache.has(store, "greeting"))

        self.assertEqual(crypto.sha1("hello"), "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d")
        self.assertEqual(
            crypto.sha256("hello"),
            "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
        )
        self.assertEqual(
            crypto.hmac_sha256("secret", "hello"),
            "88aab3ede8d3adf94d26ab90d3bafd4a2083070c3bcce9c014ee04a443847c0b",
        )
        self.assertTrue(crypto.compare_digest("corepy", "corepy"))
        self.assertFalse(crypto.compare_digest("corepy", "core"))
        encoded = crypto.base64_encode("hello")
        self.assertEqual(encoded, "aGVsbG8=")
        self.assertEqual(crypto.base64_decode(encoded), b"hello")
        self.assertEqual(len(crypto.random_bytes(16)), 16)

        self.assertEqual(dns.lookup_port("tcp", "http"), 80)
        self.assertTrue(any(address in {"127.0.0.1", "::1"} for address in dns.lookup_host("localhost")))
        self.assertTrue(dns.lookup_ip("localhost"))
        self.assertTrue(dns.reverse_lookup("127.0.0.1"))

    def test_scm_surface(self) -> None:
        if shutil.which("git") is None:
            self.skipTest("git is not available")

        with tempfile.TemporaryDirectory() as directory_name:
            repo = Path(directory_name)
            _git(repo, "init")
            _git(repo, "config", "user.email", "corepy@example.com")
            _git(repo, "config", "user.name", "CorePy Tests")
            (repo / "README.md").write_text("hello\n", encoding="utf-8")
            _git(repo, "add", "README.md")
            _git(repo, "commit", "-m", "initial")

            self.assertTrue(scm.exists(repo))
            self.assertEqual(scm.root(repo), str(repo.resolve()))
            self.assertTrue(scm.branch(repo))
            self.assertEqual(len(scm.head(repo)), 40)
            self.assertIn("README.md", scm.tracked_files(repo))

            clean_status = scm.status(repo)
            self.assertTrue(clean_status["clean"])
            self.assertEqual(clean_status["changes"], [])

            (repo / "README.md").write_text("updated\n", encoding="utf-8")
            dirty_status = scm.status(repo)
            self.assertFalse(dirty_status["clean"])
            self.assertTrue(dirty_status["changes"])

    def test_array_registry_info_and_entitlement_surface(self) -> None:
        values = array.new("a", "b")
        array.add(values, "c")
        array.add_unique(values, "c", "d")
        self.assertTrue(array.contains(values, "d"))
        array.remove(values, "b")
        array.deduplicate(values)
        self.assertEqual(array.as_list(values), ["a", "c", "d"])
        self.assertEqual(array.len(values), 3)
        array.clear(values)
        self.assertEqual(array.as_list(values), [])

        items = registry.new()
        registry.set(items, "alpha", 1)
        registry.set(items, "beta", 2)
        self.assertTrue(registry.has(items, "alpha"))
        self.assertEqual(registry.get(items, "alpha"), 1)
        self.assertEqual(registry.get(items, "missing", "fallback"), "fallback")
        self.assertEqual(registry.names(items), ["alpha", "beta"])
        registry.disable(items, "beta")
        self.assertTrue(registry.disabled(items, "beta"))
        self.assertEqual(registry.list(items, "*"), [1])
        registry.enable(items, "beta")
        self.assertEqual(registry.list(items, "*"), [1, 2])
        registry.seal(items)
        self.assertTrue(registry.sealed(items))
        with self.assertRaises(RuntimeError):
            registry.set(items, "gamma", 3)
        registry.open(items)
        registry.set(items, "gamma", 3)
        registry.lock(items)
        self.assertTrue(registry.locked(items))
        with self.assertRaises(RuntimeError):
            registry.delete(items, "alpha")

        snapshot = info.snapshot()
        self.assertEqual(info.env("OS"), snapshot["OS"])
        self.assertIn("DIR_HOME", info.keys())
        self.assertTrue(info.env("DIR_TMP"))

        grant = entitlement.new(True, False, 5, 4, 1, "")
        self.assertTrue(entitlement.near_limit(grant, 0.8))
        self.assertEqual(entitlement.usage_percent(grant), 80.0)
        self.assertTrue(grant.near_limit(0.8))
        self.assertEqual(grant.usage_percent(), 80.0)

    def test_action_task_and_i18n_surface(self) -> None:
        actions = action.new_registry()
        action.register(actions, "produce", lambda _ctx, _values: "payload")
        action.register(actions, "consume", lambda _ctx, values: f"got:{values['_input']}")

        self.assertEqual(action.names(actions), ["produce", "consume"])
        self.assertTrue(action.exists(action.get(actions, "produce")))
        self.assertEqual(action.run(actions, "produce"), "payload")
        action.disable(actions, "produce")
        with self.assertRaises(RuntimeError):
            action.run(actions, "produce")
        action.enable(actions, "produce")

        plan = task.new(
            "pipeline",
            [
                {"action": "produce"},
                {"action": "consume", "input": "previous"},
            ],
        )
        self.assertTrue(task.exists(plan))
        self.assertEqual(task.run(plan, actions), "got:payload")

        tasks = task.new_registry()
        task.register(tasks, "pipeline", [{"action": "produce"}])
        self.assertEqual(task.names(tasks), ["pipeline"])

        messages = i18n.new()
        self.assertEqual(i18n.translate(messages, "hello.world"), "hello.world")
        self.assertEqual(i18n.language(messages), "en")
        self.assertEqual(i18n.available_languages(messages), ["en"])
        i18n.add_locales(messages, "locales/core")
        self.assertEqual(i18n.locales(messages), ["locales/core"])

        class MockTranslator:
            def __init__(self) -> None:
                self._language = "en"

            def translate(self, message_id: str, *args: object) -> str:
                return f"translated:{message_id}"

            def set_language(self, lang: str) -> None:
                self._language = lang

            def language(self) -> str:
                return self._language

            def available_languages(self) -> list[str]:
                return ["en", "de", "fr"]

        translator = MockTranslator()
        i18n.set_translator(messages, translator)
        self.assertEqual(i18n.translate(messages, "hello.world"), "translated:hello.world")
        i18n.set_language(messages, "de")
        self.assertEqual(i18n.language(messages), "de")
        self.assertEqual(i18n.available_languages(messages), ["en", "de", "fr"])


if __name__ == "__main__":
    unittest.main()


def _git(directory: Path, *arguments: str) -> None:
    subprocess.run(
        ["git", "-C", str(directory), *arguments],
        check=True,
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
    )
