"""Version warning types for poppie SDK."""

from dataclasses import dataclass


@dataclass
class VersionWarning:
    """Contains version negotiation information from the server."""

    status: str
    """The server's assessment: "supported", "deprecated", or "unknown"."""

    message: str
    """An optional human-readable deprecation message."""
