# GAIOL Poster (Engineering Applications of Artificial Intelligence)

One-page conference/journal poster derived from the submitted manuscript.

## Build

From this directory:

```bash
pdflatex GAIOL_poster.tex
```

Run twice if references or layout change. Output: `GAIOL_poster.pdf`.

## Figures

Place the paper's figures in `figures/` so they appear on the poster:

| File               | Description                    |
|--------------------|--------------------------------|
| `figures/Figure_2.png` | System architecture (used in poster) |

If `Figure_2.png` is missing, the poster still compiles and shows a placeholder. See `figures/README.txt` for optional figures.

## Layout

- **Size:** 36 in (width) x 24 in (height), one page. Set in the preamble via `geometry`: `paperwidth=36in`, `paperheight=24in`. For A0 use `paperwidth=33.1in`, `paperheight=46.8in` (portrait) or swap for landscape.
- **Font:** 14pt body; title uses Huge/LARGE. Change first line of the document to `\documentclass[11pt]` or `[17pt]` if needed.
- Three-column body: Abstract, Problem/Contributions, Architecture, Algorithms, Evaluation, Results, Conclusion, References.
- Title, authors, and affiliations match the manuscript; journal name: *Engineering Applications of Artificial Intelligence*.

## Customization

- Margins and column spacing: `geometry` and `\columnsep` in the preamble.
- Section color: `\definecolor{accent}{RGB}{0,51,102}`.
- To add the reasoning-pipeline or consensus figure, insert `\includegraphics` in the Algorithmic Framework or Results section and ensure the file is in `figures/`.
