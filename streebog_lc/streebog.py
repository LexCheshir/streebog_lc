import ctypes
from pathlib import Path

so_path = Path(__file__).parent / "streebog_go" / "streebog_go.so"
streebog_go = ctypes.cdll.LoadLibrary(so_path)


class Streebog256:
    def __init__(self, data: bytes = b"") -> None:
        pass

    def update(self) -> None:
        pass

    def digest(self) -> str:
        return ""


def main():
    sb = Streebog256()
    res = sb.digest()
    print(res)


if __name__ == "__main__":
    main()
