import ctypes
from pathlib import Path
from typing import Union

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


def main():
    path = Path("/mnt/d/OS/balenaEtcher-Portable-1.7.9.exe")
    res = hash_file(path)
    print(f"{res = }")


if __name__ == "__main__":
    main()
