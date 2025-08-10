import os
import json
import errno

def get_json_file(path):
    if os.path.isfile(path):
        with open(path, 'r') as file:
            return json.loads(file.read())
    else:
        raise FileNotFoundError(errno.ENOENT, os.strerror(errno.ENOENT), path)


def write_json_file(path, data):
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, 'w') as f:
        json.dump(data, f, indent=2)


def dict_to_json(data):
    return json.dumps(data, indent=2)