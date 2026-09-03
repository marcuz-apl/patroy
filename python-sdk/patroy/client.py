"""Patroy Python client implementation with zero external runtime dependencies."""

import json
import subprocess
import urllib.error
import urllib.request
from dataclasses import dataclass
from typing import Any, Dict, List, Optional


@dataclass
class Chunk:
    index: int
    content: str
    heading: Optional[str] = None
    char_count: int = 0


@dataclass
class ScrapeResult:
    url: str
    title: str
    markdown: str
    html: Optional[str] = None
    clean_html: Optional[str] = None
    raw_html: Optional[str] = None
    author: Optional[str] = None
    description: Optional[str] = None
    date: Optional[str] = None
    site_name: Optional[str] = None
    tables: Optional[List[Dict[str, Any]]] = None
    next_data: Optional[Dict[str, Any]] = None
    json_ld: Optional[List[Any]] = None
    screenshot: Optional[str] = None
    pdf: Optional[str] = None
    is_fallback: bool = False
    duration_ms: int = 0
    chunks: Optional[List[Chunk]] = None

    @classmethod
    def from_dict(cls, data: Dict[str, Any], chunks_data: Optional[List[Dict[str, Any]]] = None) -> "ScrapeResult":
        chunks = None
        if chunks_data:
            chunks = [
                Chunk(
                    index=c.get("index", 0),
                    content=c.get("content", ""),
                    heading=c.get("heading"),
                    char_count=c.get("char_count", 0),
                )
                for c in chunks_data
            ]

        clean_html = data.get("clean_html") or data.get("html")
        html = data.get("html") or data.get("clean_html")

        return cls(
            url=data.get("url", ""),
            title=data.get("title", ""),
            markdown=data.get("markdown", ""),
            html=html,
            clean_html=clean_html,
            raw_html=data.get("raw_html"),
            author=data.get("author"),
            description=data.get("description"),
            date=data.get("date"),
            site_name=data.get("site_name"),
            tables=data.get("tables"),
            next_data=data.get("next_data"),
            json_ld=data.get("json_ld"),
            screenshot=data.get("screenshot"),
            pdf=data.get("pdf"),
            is_fallback=data.get("is_fallback", False),
            duration_ms=data.get("duration_ms", 0),
            chunks=chunks,
        )


class PatroyClient:
    """Client for communicating with the Patroy REST API microservice or local binary."""

    def __init__(self, endpoint: str = "http://localhost:4023", binary_path: Optional[str] = None):
        self.endpoint = endpoint.rstrip("/")
        self.binary_path = binary_path

    def health(self) -> Dict[str, Any]:
        """Check the health and telemetry status of the Patroy microservice."""
        url = f"{self.endpoint}/health"
        req = urllib.request.Request(url, headers={"User-Agent": "patroy-python/1.1.0"})
        with urllib.request.urlopen(req, timeout=5) as resp:
            return json.loads(resp.read().decode("utf-8"))

    def scrape(
        self,
        url: str,
        wait_for: Optional[str] = None,
        screenshot: bool = False,
        pdf: bool = False,
        timeout: int = 30,
        chunk: bool = False,
        chunk_size: int = 4000,
        chunk_overlap: int = 400,
    ) -> ScrapeResult:
        """Scrape a target URL into clean LLM Markdown and structured data."""
        payload = {
            "url": url,
            "wait_for": wait_for or "",
            "screenshot": screenshot,
            "pdf": pdf,
            "timeout_sec": timeout,
            "chunk": chunk,
            "chunk_size": chunk_size,
            "chunk_overlap": chunk_overlap,
        }

        try:
            req = urllib.request.Request(
                f"{self.endpoint}/scrape",
                data=json.dumps(payload).encode("utf-8"),
                headers={
                    "Content-Type": "application/json",
                    "User-Agent": "patroy-python/1.1.0",
                },
                method="POST",
            )
            with urllib.request.urlopen(req, timeout=timeout + 5) as resp:
                data = json.loads(resp.read().decode("utf-8"))

                if chunk and "chunks" in data:
                    return ScrapeResult.from_dict(data.get("result", {}), data.get("chunks", []))
                return ScrapeResult.from_dict(data)

        except urllib.error.URLError:
            # Fall back to local binary execution if configured
            if self.binary_path:
                return self._scrape_via_binary(url, wait_for=wait_for, timeout=timeout)
            raise

    def scrape_batch(
        self,
        urls: List[str],
        concurrency: int = 4,
        timeout: int = 60,
    ) -> List[ScrapeResult]:
        """Concurrently scrape multiple URLs through the microservice."""
        payload = {
            "urls": urls,
            "concurrency": concurrency,
            "timeout_sec": timeout,
        }

        req = urllib.request.Request(
            f"{self.endpoint}/scrape/batch",
            data=json.dumps(payload).encode("utf-8"),
            headers={
                "Content-Type": "application/json",
                "User-Agent": "patroy-python/1.1.0",
            },
            method="POST",
        )
        with urllib.request.urlopen(req, timeout=timeout + 10) as resp:
            data = json.loads(resp.read().decode("utf-8"))
            return [ScrapeResult.from_dict(item) for item in data]

    def _scrape_via_binary(self, url: str, wait_for: Optional[str] = None, timeout: int = 30) -> ScrapeResult:
        cmd = [self.binary_path, url, "-f", "json", "--silent", "--timeout", f"{timeout}s"]
        if wait_for:
            cmd.extend(["--wait-for", wait_for])

        result = subprocess.run(cmd, capture_output=True, text=True, check=True)
        data = json.loads(result.stdout)
        return ScrapeResult.from_dict(data)
