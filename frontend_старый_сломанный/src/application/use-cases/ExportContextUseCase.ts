import type { ExportRepository } from '@/domain/repositories/RepositoryInterfaces';
import type { ExportSettings, ExportResult } from '@/types/api';
import type { SplitSettings, SplitPreview, ClipboardFormat } from '@/types/splitter';
import { createSplitterService } from '@/infrastructure/context/splitter.service';
import { createTokenEstimator } from '@/infrastructure/context/token-estimator.service';

/**
 * Export Context Use Case
 * This use case handles exporting context data using the export repository
 * Encapsulates all business logic for context export including splitting and token estimation
 */
export class ExportContextUseCase {
  private splitter = createSplitterService();
  private estimator = createTokenEstimator();
  
  constructor(
    private exportRepository: ExportRepository
  ) {}

  /**
   * Execute the export context operation with advanced processing
   * @param settings Export settings with optional splitting
   * @returns Export result
   */
  async execute(settings: ExportSettings): Promise<ExportResult> {
    // Validate inputs
    if (!settings) {
      throw new Error('Export settings are required');
    }

    if (!settings.context) {
      throw new Error('Context content is required for export');
    }

    if (!settings.mode) {
      throw new Error('Export mode is required');
    }

    // Perform the export operation
    try {
      const result = await this.exportRepository.exportContext(settings);
      return result;
    } catch (error) {
      throw new Error(`Failed to export context: ${error instanceof Error ? error.message : String(error)}`);
    }
  }

  /**
   * Create a split preview for the given text and settings
   * @param text The text to split
   * @param splitSettings Split configuration
   * @returns Split preview with chunks and metadata
   */
  createSplitPreview(text: string, splitSettings: SplitSettings): SplitPreview | null {
    if (!text) {
      return null;
    }
    
    try {
      const settings = { ...splitSettings, enableAutoSplit: true };
      const preview = this.splitter.split(
        { text },
        settings,
        { estimator: (t) => this.estimator.estimate(t) }
      );
      return preview;
    } catch (error) {
      console.error('Split preview failed:', error);
      return null;
    }
  }

  /**
   * Prepare context payload with optional splitting
   * @param text Original context text
   * @param splitSettings Split configuration
   * @param splitPreview Optional pre-computed split preview
   * @param target Export target ('all' or 'chunk')
   * @param chunkIndex Optional chunk index for single chunk export
   * @returns Processed text payload
   */
  prepareContextPayload(
    text: string, 
    splitSettings: SplitSettings, 
    splitPreview: SplitPreview | null, 
    target: 'all' | 'chunk' = 'all',
    chunkIndex?: number
  ): string {
    if (!text) {
      return '';
    }

    let payload = text;
    
    if (
      splitSettings.enableAutoSplit &&
      splitPreview &&
      splitPreview.chunks.length > 1
    ) {
      if (target === 'chunk' && typeof chunkIndex === 'number') {
        const chunk = splitPreview.chunks[chunkIndex];
        payload = chunk ? chunk.text : text;
      } else {
        const parts = splitPreview.chunks.map(
          (c, i, arr) => `=== Part ${i + 1}/${arr.length} ===\n${c.text}`,
        );
        payload = parts.join('\n\n');
      }
    }
    
    return payload;
  }

  /**
   * Export context to clipboard with advanced processing
   * @param text Original context text
   * @param format Export format
   * @param stripComments Whether to strip comments
   * @param splitSettings Split configuration
   * @param splitPreview Optional pre-computed split preview
   * @param target Export target
   * @param chunkIndex Optional chunk index
   * @returns Export result
   */
  async exportToClipboard(
    text: string,
    format: ClipboardFormat,
    stripComments: boolean,
    splitSettings: SplitSettings,
    splitPreview: SplitPreview | null,
    target: 'all' | 'chunk' = 'all',
    chunkIndex?: number
  ): Promise<ExportResult> {
    const payload = this.prepareContextPayload(text, splitSettings, splitPreview, target, chunkIndex);
    
    const exportSettings: ExportSettings = {
      mode: 'clipboard',
      context: payload,
      exportFormat: format,
      stripComments,
      includeManifest: format === 'manifest',
      aiProfile: 'Generic',
      tokenLimit: 180000,
      fileSizeLimitKB: 2048,
      enableAutoSplit: false,
      maxTokensPerChunk: 50000,
      overlapTokens: 1000,
      splitStrategy: 'token',
      theme: 'Dark',
      includeLineNumbers: true,
      includePageNumbers: true,
    };
    
    return await this.execute(exportSettings);
  }
}