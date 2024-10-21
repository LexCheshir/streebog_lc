import ctypes
from pathlib import Path

so_path = Path(__file__).parent / "streebog_go" / "streebog_go.so"
streebog_go = ctypes.cdll.LoadLibrary(so_path)
hello_world = streebog_go.helloWorld


def streebog_256(data) -> str:
    return ""


def streebog_512(data) -> str:
    return ""


def main():
    hello_world()


if __name__ == "__main__":
    main()
