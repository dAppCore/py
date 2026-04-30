import base64
import hashlib
import json
import os

from core import fs, path


workspace = fs.temp_dir("corepy-stdlib-")
target = os.path.join(workspace, "payload.json")
fs.write_file(target, json.dumps({"name": "corepy", "tier": 1}))

digest = hashlib.sha256(fs.read_file(target))
print(os.path.basename(target))
print(base64.b64encode(path.base(target)))
print(digest.hexdigest())
