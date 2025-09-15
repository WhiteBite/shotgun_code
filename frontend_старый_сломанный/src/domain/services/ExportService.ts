import { ref, type Ref } from 'vue'
import { APP_CONFIG } from '@/config/app-config'

export class ExportService {
  // Export format
  private _exportFormat: Ref<string> = ref('plain')
  private _stripComments: Ref<boolean> = ref(false)
  private _includeManifest: Ref<boolean> = ref(false)

  // AI Profile
  private _aiProfile: Ref<string> = ref('Claude-3')

  // Auto-split settings
  private _enableAutoSplit: Ref<boolean> = ref(true)
  private _maxTokensPerChunk: Ref<number> = ref(APP_CONFIG.export.ai.DEFAULT_MAX_TOKENS_PER_CHUNK)
  private _splitStrategy: Ref<string> = ref('smart')
  private _overlapTokens: Ref<number> = ref(APP_CONFIG.export.ai.DEFAULT_OVERLAP_TOKENS)

  // Limits
  private _tokenLimit: Ref<number> = ref(APP_CONFIG.export.ai.DEFAULT_TOKEN_LIMIT)
  private _fileSizeLimitKB: Ref<number> = ref(APP_CONFIG.export.file.DEFAULT_SIZE_LIMIT_KB)

  // Human-readable export settings
  private _theme: Ref<string> = ref('Dark')
  private _includeLineNumbers: Ref<boolean> = ref(true)
  private _includePageNumbers: Ref<boolean> = ref(true)

  constructor(
    private getIsOpen: () => boolean,
    private getIsLoading: () => boolean,
    private doExportClipboard: () => Promise<void>,
    private doExportAI: () => Promise<void>,
    private doExportHuman: () => Promise<void>,
    private close: () => void
  ) {}

  // Export format getters and setters
  get exportFormat() {
    return this._exportFormat.value
  }

  set exportFormat(value: string) {
    this._exportFormat.value = value
  }

  get stripComments() {
    return this._stripComments.value
  }

  set stripComments(value: boolean) {
    this._stripComments.value = value
  }

  get includeManifest() {
    return this._includeManifest.value
  }

  set includeManifest(value: boolean) {
    this._includeManifest.value = value
  }

  // AI Profile getters and setters
  get aiProfile() {
    return this._aiProfile.value
  }

  set aiProfile(value: string) {
    this._aiProfile.value = value
  }

  get aiProfileHint() {
    const hints: Record<string, string> = {
      'Claude-3': 'Optimized for Claude 3 models with appropriate token limits',
      'GPT-4o': 'Optimized for GPT-4o with appropriate token limits',
      'Generic': 'Generic profile with standard settings'
    }
    return hints[this._aiProfile.value] || ''
  }

  // Auto-split settings getters and setters
  get enableAutoSplit() {
    return this._enableAutoSplit.value
  }

  set enableAutoSplit(value: boolean) {
    this._enableAutoSplit.value = value
  }

  get maxTokensPerChunk() {
    return this._maxTokensPerChunk.value
  }

  set maxTokensPerChunk(value: number) {
    this._maxTokensPerChunk.value = value
  }

  get splitStrategy() {
    return this._splitStrategy.value
  }

  set splitStrategy(value: string) {
    this._splitStrategy.value = value
  }

  get splitStrategyHint() {
    const hints: Record<string, string> = {
      'smart': 'Intelligently splits at file boundaries when possible',
      'file': 'Splits only at complete file boundaries',
      'token': 'Splits precisely at token boundaries (may split files)'
    }
    return hints[this._splitStrategy.value] || ''
  }

  get overlapTokens() {
    return this._overlapTokens.value
  }

  set overlapTokens(value: number) {
    this._overlapTokens.value = value
  }

  // Limits getters and setters
  get tokenLimit() {
    return this._tokenLimit.value
  }

  set tokenLimit(value: number) {
    this._tokenLimit.value = value
  }

  get fileSizeLimitKB() {
    return this._fileSizeLimitKB.value
  }

  set fileSizeLimitKB(value: number) {
    this._fileSizeLimitKB.value = value
  }

  // Human-readable export settings getters and setters
  get theme() {
    return this._theme.value
  }

  set theme(value: string) {
    this._theme.value = value
  }

  get includeLineNumbers() {
    return this._includeLineNumbers.value
  }

  set includeLineNumbers(value: boolean) {
    this._includeLineNumbers.value = value
  }

  get includePageNumbers() {
    return this._includePageNumbers.value
  }

  set includePageNumbers(value: boolean) {
    this._includePageNumbers.value = value
  }

  // Computed properties
  get shouldAutoSplit() {
    return this._enableAutoSplit.value && this._tokenLimit.value > 0
  }

  get isManifestFormatSelected() {
    return this._exportFormat.value === 'manifest'
  }

  // Segment button classes
  getSegmentButtonClasses(selectedValue: string) {
    return [
      "px-2 py-1 text-xs",
      this._exportFormat.value === selectedValue
        ? "bg-blue-600 text-white"
        : "text-gray-300 hover:bg-gray-700",
    ]
  }

  // Action handlers
  async handleExportClipboard() {
    await this.doExportClipboard()
  }

  async handleExportAI() {
    await this.doExportAI()
  }

  async handleExportHuman() {
    await this.doExportHuman()
  }

  handleClose() {
    this.close()
  }

  // State accessors
  get isOpen() {
    return this.getIsOpen()
  }

  get isLoading() {
    return this.getIsLoading()
  }
}