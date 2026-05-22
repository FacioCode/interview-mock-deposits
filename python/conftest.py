"""Top-level pytest config.

Adds src/ to sys.path so each lambda + shared module is importable by its
folder name (e.g. `from deposit import ...`). Matches the layout used by
the build script when packaging lambda zips.
"""
import sys
from pathlib import Path

_SRC = Path(__file__).parent / "src"
if str(_SRC) not in sys.path:
    sys.path.insert(0, str(_SRC))
