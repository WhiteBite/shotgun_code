/* Auto-generated: Token Estimator service with DI */
export type EstimatorFn = (text: string) => number;

let globalEstimator: EstimatorFn | null = null;

/**
 * Позволяет подменить глобальный оценщик токенов (например, tiktoken для OpenAI).
 * Передайте null, чтобы вернуть эвристику по умолчанию.
 */
export function setGlobalEstimator(fn: EstimatorFn | null) {
  globalEstimator = fn;
}

export interface TokenEstimator {
  estimate(text: string): number;
}

/**
 * Простая фабрика. В будущем можно учитывать профиль/модель для выбора стратегии.
 */
export function createTokenEstimator(_profile: string = "Generic", _model?: string): TokenEstimator {
  const heuristic: EstimatorFn = (text: string) => Math.max(1, Math.ceil((text || "").length / 4));
  const fn: EstimatorFn = (t) => (globalEstimator ? globalEstimator(t) : heuristic(t));
  return { estimate: fn };
}
