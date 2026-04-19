#!/usr/bin/env python3
"""
Patch client-ish binary: replace futex_time64 (syscall 422) with futex (syscall 240).

iSH on iOS does not support Linux syscall 422 (futex_time64), which Go ≥ 1.19
emits for linux/386. This script finds the exact instruction pattern and
rewrites only the syscall number, leaving the rest of the instruction intact.
"""

import sys
import os

PATTERN = bytes([0xb8, 0xa6, 0x01, 0x00, 0x00,  # MOV EAX, 422
                 0x8b, 0x5c, 0x24, 0x04])         # MOV EBX, [ESP+4]
REPLACEMENT_BYTES = bytes([0xb8, 0xf0, 0x00, 0x00, 0x00])  # MOV EAX, 240


def patch(path: str) -> None:
    with open(path, "rb") as f:
        data = bytearray(f.read())

    idx = data.find(PATTERN)
    if idx == -1:
        print(f"ERROR: pattern not found in {path}")
        print("The binary may already be patched or was built with a different Go version.")
        sys.exit(1)

    # Check for duplicate (shouldn't happen, but guard anyway)
    second = data.find(PATTERN, idx + 1)
    if second != -1:
        print(f"WARNING: pattern found twice (0x{idx:x} and 0x{second:x}), patching first occurrence only")

    print(f"Found pattern at offset 0x{idx:x}")
    data[idx:idx + len(REPLACEMENT_BYTES)] = REPLACEMENT_BYTES

    with open(path, "wb") as f:
        f.write(data)

    print(f"Patched OK: syscall 422 (futex_time64) → 240 (futex)")


if __name__ == "__main__":
    binary = sys.argv[1] if len(sys.argv) > 1 else "client-ish"
    if not os.path.exists(binary):
        print(f"ERROR: file not found: {binary}")
        sys.exit(1)
    patch(binary)
