"""Poppie gRPC client with version negotiation."""

from __future__ import annotations

import logging
import os
from pathlib import Path
from typing import Callable

import grpc

from ._generated.poppie import poppie_pb2, poppie_pb2_grpc
from ._version import SDK_NAME, SDK_VERSION
from ._warnings import VersionWarning

logger = logging.getLogger("poppie")

# Header constants matching the Go server.
_HEADER_SDK_VERSION = "x-poppie-sdk-version"
_HEADER_SDK_NAME = "x-poppie-sdk-name"
_HEADER_VERSION_STATUS = "x-poppie-version-status"
_HEADER_DEPRECATION_MESSAGE = "x-poppie-deprecation-message"


def _default_socket_path() -> str:
    home = Path.home()
    return str(home / ".config" / "poppie" / "poppie.sock")


def _default_warning_handler(warning: VersionWarning) -> None:
    logger.warning(
        "poppie server version warning: status=%s message=%s",
        warning.status,
        warning.message,
    )


# Map string algorithm names to proto enum values.
_ALGORITHMS = {
    "sha1": poppie_pb2.ALGORITHM_SHA1,
    "sha256": poppie_pb2.ALGORITHM_SHA256,
    "sha512": poppie_pb2.ALGORITHM_SHA512,
}


class PoppieClient:
    """Client for the poppie TOTP manager gRPC server.

    Supports context manager protocol for automatic cleanup.
    """

    def __init__(
        self,
        socket_path: str | None = None,
        warning_handler: Callable[[VersionWarning], None] | None = _default_warning_handler,
    ) -> None:
        self._socket_path = socket_path or _default_socket_path()
        self._warning_handler = warning_handler
        self._channel = grpc.insecure_channel(f"unix:///{self._socket_path}")
        self._stub = poppie_pb2_grpc.PoppieServiceStub(self._channel)

    def __enter__(self) -> PoppieClient:
        return self

    def __exit__(self, *args: object) -> None:
        self.close()

    def close(self) -> None:
        """Close the underlying gRPC channel."""
        self._channel.close()

    def store_secret(
        self,
        label: str,
        secret: str,
        *,
        algorithm: str = "sha1",
        digits: int = 0,
        period: int = 0,
    ) -> tuple[str, str]:
        """Store a TOTP secret.

        Returns (label, verification_code).
        """
        algo = _ALGORITHMS.get(algorithm.lower(), poppie_pb2.ALGORITHM_SHA1)
        resp, call = self._stub.StoreSecret.with_call(
            poppie_pb2.StoreSecretRequest(
                label=label,
                secret=secret,
                algorithm=algo,
                digits=digits,
                period=period,
            ),
            metadata=self._version_metadata(),
        )
        self._handle_warnings(call)
        return resp.label, resp.verification_code

    def get_code(self, label: str) -> tuple[str, int]:
        """Get a current TOTP code.

        Returns (code, valid_for_seconds).
        """
        resp, call = self._stub.GetCode.with_call(
            poppie_pb2.GetCodeRequest(label=label),
            metadata=self._version_metadata(),
        )
        self._handle_warnings(call)
        return resp.code, resp.valid_for_seconds

    def list_secrets(self) -> list[str]:
        """Return labels of all stored secrets."""
        resp, call = self._stub.ListSecrets.with_call(
            poppie_pb2.ListSecretsRequest(),
            metadata=self._version_metadata(),
        )
        self._handle_warnings(call)
        return list(resp.labels)

    def delete_secret(self, label: str) -> bool:
        """Delete a secret. Returns True if it was actually deleted."""
        resp, call = self._stub.DeleteSecret.with_call(
            poppie_pb2.DeleteSecretRequest(label=label),
            metadata=self._version_metadata(),
        )
        self._handle_warnings(call)
        return resp.deleted

    def _version_metadata(self) -> list[tuple[str, str]]:
        return [
            (_HEADER_SDK_VERSION, SDK_VERSION),
            (_HEADER_SDK_NAME, SDK_NAME),
        ]

    def _handle_warnings(self, call: grpc.Call) -> None:
        if self._warning_handler is None:
            return

        headers = dict(call.initial_metadata())
        status = headers.get(_HEADER_VERSION_STATUS, "")

        if not status or status in ("supported", "unknown"):
            return

        message = headers.get(_HEADER_DEPRECATION_MESSAGE, "")
        self._warning_handler(VersionWarning(status=status, message=message))
