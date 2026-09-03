import unittest
from patroy import PatroyClient, ScrapeResult, Chunk


class TestPatroyClient(unittest.TestCase):
    def test_scrape_result_parsing(self):
        data = {
            "url": "https://example.com",
            "title": "Example Domain",
            "markdown": "This is example markdown text.",
            "html": "<article><p>Hello</p></article>",
            "tables": [{"headers": ["A", "B"], "rows": [["1", "2"]]}],
            "duration_ms": 120,
            "is_fallback": False,
        }
        chunks = [
            {"index": 0, "content": "This is example", "char_count": 15},
            {"index": 1, "content": "markdown text.", "char_count": 14},
        ]

        res = ScrapeResult.from_dict(data, chunks)
        self.assertEqual(res.url, "https://example.com")
        self.assertEqual(res.title, "Example Domain")
        self.assertEqual(res.html, "<article><p>Hello</p></article>")
        self.assertEqual(res.clean_html, "<article><p>Hello</p></article>")
        self.assertEqual(len(res.tables), 1)
        self.assertEqual(res.tables[0]["headers"], ["A", "B"])
        self.assertEqual(res.duration_ms, 120)
        self.assertFalse(res.is_fallback)
        self.assertIsNotNone(res.chunks)
        self.assertEqual(len(res.chunks), 2)
        self.assertEqual(res.chunks[0].content, "This is example")


if __name__ == "__main__":
    unittest.main()
