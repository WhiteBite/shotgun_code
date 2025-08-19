/* Auto-generated: content splitter service (framework-agnostic) */
import type {
  SplitInput, SplitSettings, SplitPreview, TokenEstimateOptions,
  Segment, SplitChunk
} from "@/types/splitter";

export interface SplitterService {
  estimateTokens(text: string, opts?: TokenEstimateOptions): number;
  detectSegmentsFromContextText(text: string): Segment[];
  split(input: SplitInput, settings: SplitSettings, opts?: TokenEstimateOptions): SplitPreview;
}

function defaultEstimateTokens(text: string): number {
  // Fallback heuristic; can be replaced by precise estimator (tiktoken etc.)
  if (!text) return 0;
  return Math.max(1, Math.ceil(text.length / 4));
}

function clamp(n: number, min: number, max: number): number {
  return Math.max(min, Math.min(max, n));
}

function makeChunk(text: string, start: number, end: number, estimator: (t: string)=>number): SplitChunk {
  const slice = text.slice(start, end);
  return {
    index: 0, // will be assigned later
    start, end,
    text: slice,
    tokens: estimator(slice),
    chars: slice.length,
  };
}

function packSegmentsByTokens(
  text: string,
  segments: Segment[],
  limitTokens: number,
  estimator: (t: string)=>number,
  overlapTokens: number
): { chunks: SplitChunk[]; warnings: string[] } {
  const chunks: SplitChunk[] = [];
  const warnings: string[] = [];
  let currentStart = segments.length ? segments[0].start : 0;
  let accTokens = 0;
  let accStart = currentStart;
  let lastEnd = currentStart;

  const avgCharsPerToken = (text.length || 1) / Math.max(1, estimator(text));
  const overlapChars = Math.floor(overlapTokens * avgCharsPerToken);

  const pushChunk = (from: number, to: number) => {
    const s = clamp(from, 0, text.length);
    const e = clamp(to, s, text.length);
    const chunk = makeChunk(text, s, e, estimator);
    chunks.push(chunk);
  };

  for (const seg of segments) {
    const segText = text.slice(seg.start, seg.end);
    const segTokens = estimator(segText);
    if (segTokens <= limitTokens) {
      if (accTokens + segTokens > limitTokens && lastEnd > accStart) {
        pushChunk(accStart, lastEnd);
        accStart = Math.max(accStart, lastEnd - overlapChars);
        accTokens = estimator(text.slice(accStart, accStart)); // 0
      }
      if (accStart === 0 && chunks.length === 0) {
        accStart = seg.start;
      } else if (lastEnd === 0) {
        accStart = seg.start;
      }
      lastEnd = seg.end;
      accTokens += segTokens;
    } else {
      if (lastEnd > accStart) {
        pushChunk(accStart, lastEnd);
        accStart = Math.max(accStart, lastEnd - overlapChars);
      }
      const charsPerToken = avgCharsPerToken || 4;
      const stepChars = Math.max(1, Math.floor(limitTokens * charsPerToken));
      let pos = seg.start;
      while (pos < seg.end) {
        const endPos = Math.min(seg.end, pos + stepChars);
        pushChunk(pos, endPos);
        pos = endPos - overlapChars;
        if (pos <= seg.start) pos = endPos; // guard
      }
      warnings.push(`Segment too large, split by token within: ${seg.title || seg.relPath || seg.id}`);
      accTokens = 0;
      lastEnd = 0;
      accStart = segments.find(s => s.start >= seg.end)?.start ?? seg.end;
    }
  }

  if (lastEnd > accStart) {
    pushChunk(accStart, lastEnd);
  }

  chunks.forEach((c, i) => c.index = i);
  return { chunks, warnings };
}

export function createSplitterService(): SplitterService {
  return {
    estimateTokens(text: string, opts?: TokenEstimateOptions): number {
      if (opts?.estimator) return Math.max(0, opts.estimator(text));
      return defaultEstimateTokens(text);
    },

    detectSegmentsFromContextText(text: string): Segment[] {
      if (!text) return [];
      const lines = text.split(/\r?\n/);
      const segments: Segment[] = [];
      let currentStart = 0;
      let currentTitle = "Context";
      let currentRel = "";

      const push = (start: number, end: number, title: string, rel?: string) => {
        if (end <= start) return;
        segments.push({
          id: `${start}:${end}`,
          title,
          start, end,
          isFile: !!rel,
          relPath: rel,
        });
      };

      for (let i = 0; i < lines.length; i++) {
        const line = lines[i];
        const m1 = /^File:\s+(.+)$/.exec(line);
        const m2 = /^\+\+\+\s.+?\s(.+)$/.exec(line);
        const m3 = /^\s*[-=]{3,}\s*(.+?)\s*[-=]{3,}\s*$/.exec(line);
        if (m1 || m2 || m3) {
          const prevEnd = text.split(/\r?\n/).slice(0, i).join("\n").length;
          push(currentStart, prevEnd, currentTitle, currentRel || undefined);
          currentStart = prevEnd + (i > 0 ? 1 : 0);
          currentTitle = "File";
          currentRel = (m1?.[1] || m2?.[1] || m3?.[1] || "").trim();
        }
      }
      push(currentStart, text.length, currentTitle, currentRel || undefined);

      if (segments.length === 0) {
        return [{ id: "0:len", title: "Context", start: 0, end: text.length }];
      }
      return segments.filter(s => s.end > s.start).sort((a, b) => a.start - b.start);
    },

    split(input: SplitInput, settings: SplitSettings, opts?: TokenEstimateOptions): SplitPreview {
      const text = input.text || "";
      const estimator = (t: string) => (opts?.estimator ? opts.estimator(t) : defaultEstimateTokens(t));
      const totalTokens = estimator(text);
      const totalChars = text.length;

      if (!settings.enableAutoSplit || !text) {
        const chunk: SplitChunk = {
          index: 0, start: 0, end: text.length, text,
          tokens: totalTokens, chars: totalChars,
        };
        return { totalTokens, totalChars, chunkCount: 1, chunks: [chunk], warnings: [] };
      }

      const limit = Math.max(1, settings.maxTokensPerChunk);
      const overlap = Math.max(0, settings.overlapTokens || 0);
      const strategy = settings.splitStrategy || "smart";

      let segments: Segment[] | undefined = input.segments;
      if ((!segments || !segments.length) && (strategy === "file" || strategy === "smart")) {
        segments = this.detectSegmentsFromContextText(text);
      }

      if ((strategy === "file" || strategy === "smart") && segments && segments.length) {
        const { chunks, warnings } = packSegmentsByTokens(text, segments, limit, estimator, overlap);
        return { totalTokens, totalChars, chunkCount: chunks.length, chunks, warnings };
      }

      const avgCharsPerToken = totalTokens > 0 ? totalChars / totalTokens : 4;
      const stepChars = Math.max(1, Math.floor(limit * avgCharsPerToken));
      const overlapChars = Math.floor(overlap * avgCharsPerToken);
      const chunks: SplitChunk[] = [];
      let i = 0;
      while (i < text.length) {
        const end = Math.min(text.length, i + stepChars);
        chunks.push(makeChunk(text, i, end, estimator));
        i = end - overlapChars;
        if (i <= 0) i = end;
        if (i <= 0) break;
      }
      chunks.forEach((c, idx) => (c.index = idx));
      return { totalTokens, totalChars, chunkCount: chunks.length, chunks, warnings: [] };
    },
  };
}
