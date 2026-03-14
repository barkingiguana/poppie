"""Tests for the poppie Python SDK client.

These tests validate the SDK types and configuration. Integration tests
against a live server require a running poppie instance and are run separately.
"""

from poppie import PoppieClient, VersionWarning, SDK_VERSION


def test_version_warning_dataclass():
    """VersionWarning stores status and message."""
    w = VersionWarning(status="deprecated", message="please upgrade")
    assert w.status == "deprecated"
    assert w.message == "please upgrade"


def test_sdk_version_is_set():
    """SDK_VERSION matches the root VERSION file."""
    assert SDK_VERSION == "0.1.0"


def test_client_default_socket_path():
    """PoppieClient uses default socket path when none is given."""
    import os
    from pathlib import Path

    expected = str(Path.home() / ".config" / "poppie" / "poppie.sock")
    client = PoppieClient.__new__(PoppieClient)
    # We can't fully init without a server, but we can check defaults.
    from poppie._client import _default_socket_path

    assert _default_socket_path() == expected


def test_client_context_manager():
    """PoppieClient supports context manager protocol."""
    assert hasattr(PoppieClient, "__enter__")
    assert hasattr(PoppieClient, "__exit__")


def test_algorithm_mapping():
    """Algorithm strings map to proto enum values."""
    from poppie._client import _ALGORITHMS
    from poppie._generated.poppie import poppie_pb2

    assert _ALGORITHMS["sha1"] == poppie_pb2.ALGORITHM_SHA1
    assert _ALGORITHMS["sha256"] == poppie_pb2.ALGORITHM_SHA256
    assert _ALGORITHMS["sha512"] == poppie_pb2.ALGORITHM_SHA512
