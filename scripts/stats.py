#!/usr/bin/env python3
"""Fetch download counts from GitHub releases API and write stats.json."""

import json
import sys
import os
from urllib.request import Request, urlopen

REPO = "lfaoro/ssm"
API_URL = f"https://api.github.com/repos/{REPO}/releases?per_page=100"
OUTPUT = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "data", "stats.json")


def main():
    req = Request(API_URL, headers={"Accept": "application/vnd.github+json",
                                     "User-Agent": "ssm-stats/1.0"})
    with urlopen(req) as resp:
        releases = json.load(resp)

    known_os = {"linux", "darwin", "freebsd", "netbsd", "openbsd", "solaris", "windows"}
    platforms = {}
    total = 0
    for rel in releases:
        for asset in rel.get("assets", []):
            count = asset["download_count"]
            name = asset["name"]
            if name.endswith((".asc", "checksums.txt")):
                continue
            stem = name
            for ext in (".tar.gz", ".tgz", ".zip", ".deb", ".rpm"):
                if stem.endswith(ext):
                    stem = stem[: -len(ext)]
                    break
            key = stem
            for os_name in sorted(known_os, key=len, reverse=True):
                idx = stem.find(os_name + "_")
                if idx >= 0:
                    key = f"{os_name}/{stem[idx + len(os_name) + 1:]}"
                    break
            platforms[key] = platforms.get(key, 0) + count
            total += count

    result = {"total": total, "platforms": dict(sorted(platforms.items(), key=lambda x: -x[1]))}
    os.makedirs(os.path.dirname(OUTPUT), exist_ok=True)
    with open(OUTPUT, "w") as f:
        json.dump(result, f, indent=2)
        f.write("\n")
    print(f"Updated {OUTPUT}: {total} total downloads")


if __name__ == "__main__":
    main()
