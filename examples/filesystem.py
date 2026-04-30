from core import fs, json


target = fs.write_file("/tmp/corepy-example.json", json.dumps({"name": "corepy"}))
print(fs.read_file(target))
