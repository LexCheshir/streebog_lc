import ctypes
import datetime
from pathlib import Path
from typing import Union

import gostcrypto

so_path = Path(__file__).parent / "streebog_go" / "streebog_go.so"
streebog_go = ctypes.cdll.LoadLibrary(so_path)

_c_hash_file = streebog_go.HashFileWrapper
_c_hash_file.argtypes = [ctypes.c_char_p]
_c_hash_file.restype = ctypes.c_void_p


class Streebog256:
    def __init__(self, data: bytes = b"") -> None:
        pass

    def update(self) -> None:
        pass

    def digest(self) -> str:
        return ""


def hash_file(path: Union[str, Path]) -> str:
    path = Path(path).resolve()
    c_res = _c_hash_file(str(path).encode("utf-8"))
    c_bytes = ctypes.string_at(c_res)
    res = c_bytes.decode("utf-8")
    return res


def lib_hash(path: Path) -> str:
    hash_obj = gostcrypto.gosthash.new("streebog256")
    buffer_size = 64
    with path.open(mode="rb") as f:
        buffer = f.read(buffer_size)
        while len(buffer) > 0:
            hash_obj.update(buffer)
            buffer = f.read(buffer_size)
    res = hash_obj.hexdigest()
    return res


def test_file(path: Path):
    print(f"## {path}\n")
    print(f"size: {path.stat().st_size}")

    start = datetime.datetime.now()
    my_res = hash_file(path)
    my_end = datetime.datetime.now() - start
    print(f"{my_end = }")

    start = datetime.datetime.now()
    li_res = lib_hash(path)
    li_end = datetime.datetime.now() - start
    print(f"{li_end = }")
    print()

    print(f"{my_res = }")
    print(f"{li_res = }")
    print(f"-{'[x]' if my_res == li_res else '[ ]'} same result")
    print(f"perf improvement: x{li_end / my_end}")
    print("\n")


def main():
    path_list = [
        Path("/mnt/d/OS/rufus-4.5.exe"),
        Path("/mnt/d/OS/rufus-4.5p.exe"),
        Path("/mnt/d/workinprogress/polytech/batch/lex_diploma_bachelor_fin_v3.docx"),
    ]

    print()
    for path in path_list:
        test_file(path)


if __name__ == "__main__":
    main()
