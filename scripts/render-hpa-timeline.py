#!/usr/bin/env python3
"""Render an HPA replica timeline CSV as a dependency-free SVG artifact."""

from __future__ import annotations

import csv
import html
import sys
from pathlib import Path


def main() -> int:
    if len(sys.argv) != 3:
        print("usage: render-hpa-timeline.py INPUT.csv OUTPUT.svg", file=sys.stderr)
        return 2
    source, destination = map(Path, sys.argv[1:])
    with source.open(newline="", encoding="utf-8") as handle:
        rows = list(csv.DictReader(handle))
    if not rows:
        print("HPA timeline is empty", file=sys.stderr)
        return 1

    timestamps = [int(row["timestamp"]) for row in rows]
    current = [int(row["currentReplicas"]) for row in rows]
    desired = [int(row["desiredReplicas"]) for row in rows]
    width, height, left, top, right, bottom = 960, 520, 72, 48, 32, 72
    plot_width = width - left - right
    plot_height = height - top - bottom
    t0, t1 = min(timestamps), max(timestamps)
    span = max(1, t1 - t0)
    ymax = max(2, max(current + desired) + 1)

    def x(value: int) -> float:
        return left + (value - t0) / span * plot_width

    def y(value: int) -> float:
        return top + (ymax - value) / ymax * plot_height

    current_points = " ".join(f"{x(t):.1f},{y(v):.1f}" for t, v in zip(timestamps, current))
    desired_points = " ".join(f"{x(t):.1f},{y(v):.1f}" for t, v in zip(timestamps, desired))
    service = html.escape(rows[0]["service"])
    start_label = html.escape(rows[0]["isoTime"])
    end_label = html.escape(rows[-1]["isoTime"])

    grid = []
    for value in range(ymax + 1):
        ypos = y(value)
        grid.append(
            f'<line x1="{left}" y1="{ypos:.1f}" x2="{width-right}" y2="{ypos:.1f}" stroke="#d8dee9"/>'
            f'<text x="{left-12}" y="{ypos+5:.1f}" text-anchor="end">{value}</text>'
        )
    svg = f'''<svg xmlns="http://www.w3.org/2000/svg" width="{width}" height="{height}" viewBox="0 0 {width} {height}">
<rect width="100%" height="100%" fill="#ffffff"/>
<style>text{{font:14px sans-serif;fill:#263238}} .title{{font:bold 20px sans-serif}}</style>
<text class="title" x="{left}" y="28">HPA replica timeline — {service}</text>
{''.join(grid)}
<line x1="{left}" y1="{top}" x2="{left}" y2="{height-bottom}" stroke="#455a64"/>
<line x1="{left}" y1="{height-bottom}" x2="{width-right}" y2="{height-bottom}" stroke="#455a64"/>
<polyline points="{desired_points}" fill="none" stroke="#ff8f00" stroke-width="3" stroke-dasharray="7 5"/>
<polyline points="{current_points}" fill="none" stroke="#1565c0" stroke-width="4"/>
<text x="{left}" y="{height-36}">{start_label}</text>
<text x="{width-right}" y="{height-36}" text-anchor="end">{end_label}</text>
<line x1="{left}" y1="{height-14}" x2="{left+34}" y2="{height-14}" stroke="#1565c0" stroke-width="4"/>
<text x="{left+42}" y="{height-9}">current replicas</text>
<line x1="{left+190}" y1="{height-14}" x2="{left+224}" y2="{height-14}" stroke="#ff8f00" stroke-width="3" stroke-dasharray="7 5"/>
<text x="{left+232}" y="{height-9}">desired replicas</text>
</svg>'''
    destination.write_text(svg, encoding="utf-8")
    print(destination)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
