from core import cache, crypto, fs, json, path


workspace = fs.temp_dir("corepy-pass3-")
payload = json.dumps({"name": "corepy", "pass": 3})
target = path.join(workspace, "payload.json")
fs.write_file(target, payload)
print(crypto.sha256(fs.read_file(target)))

store = cache.new(path.join(workspace, "cache"), 60)
cache.set(store, "payload", {"path": target})
print(cache.has(store, "payload"))
