import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue';
import { defaultVirtualScrollingService, type VirtualScrollConfig, type VirtualItem } from '@/domain/services/VirtualScrollingService';
import { APP_CONFIG } from '@/config/app-config';

export interface VirtualScrollOptions {
  itemHeight: number;
  containerHeight: number;
  buffer?: number;
  totalItems: number;
}

/**
 * Composable for virtual scrolling functionality
 * Delegates business logic to VirtualScrollingService domain service
 * Following DDD principles by separating UI concerns from domain logic
 */
export function useVirtualScroll(options: VirtualScrollOptions) {
  const { 
    itemHeight, 
    containerHeight, 
    buffer = APP_CONFIG.ui.virtualScroll.DEFAULT_BUFFER, 
    totalItems: initialTotalItems 
  } = options;
  
  const scrollTop = ref(0);
  const containerRef = ref<HTMLElement | null>(null);
  const totalItems = ref(initialTotalItems);
  
  // Domain service instance
  const virtualScrollService = defaultVirtualScrollingService;
  
  // Cache for scroll metrics to reduce reactivity
  let cachedScrollMetrics: ReturnType<typeof virtualScrollService.calculateScrollMetrics> | null = null;
  let cachedScrollTop = -1;
  let cachedConfig: VirtualScrollConfig | null = null;
  
  // Virtual scroll configuration (memoized)
  const getConfig = (): VirtualScrollConfig => {
    const currentConfig = {
      itemHeight,
      containerHeight,
      buffer,
      totalItems: totalItems.value
    };
    
    // Only recreate config if it has changed
    if (!cachedConfig || 
        cachedConfig.itemHeight !== currentConfig.itemHeight ||
        cachedConfig.containerHeight !== currentConfig.containerHeight ||
        cachedConfig.buffer !== currentConfig.buffer ||
        cachedConfig.totalItems !== currentConfig.totalItems) {
      cachedConfig = currentConfig;
    }
    
    return cachedConfig;
  };
  
  // Computed properties using domain service with caching
  const visibleItems = computed((): VirtualItem[] => {
    const config = getConfig();
    return virtualScrollService.calculateVisibleItems(config, scrollTop.value);
  });
  
  const getScrollMetrics = () => {
    const config = getConfig();
    
    // Only recalculate if scrollTop or config has changed
    if (!cachedScrollMetrics || 
        cachedScrollTop !== scrollTop.value || 
        cachedConfig !== config) {
      cachedScrollMetrics = virtualScrollService.calculateScrollMetrics(config, scrollTop.value);
      cachedScrollTop = scrollTop.value;
    }
    
    return cachedScrollMetrics;
  };
  
  // Optimized computed properties with reduced reactivity
  const visibleRange = computed(() => {
    const metrics = getScrollMetrics();
    return {
      start: metrics.visibleStart,
      end: metrics.visibleEnd
    };
  });
  
  const visibleStart = computed(() => visibleRange.value.start);
  const visibleEnd = computed(() => visibleRange.value.end);
  
  const totalHeight = computed(() => {
    return totalItems.value * itemHeight;
  });
  
  const offsetY = computed(() => {
    return visibleStart.value * itemHeight;
  });
  
  const shouldUseVirtualization = computed(() => {
    return virtualScrollService.shouldUseVirtualization(
      totalItems.value,
      itemHeight,
      containerHeight
    );
  });
  
  // Methods delegating to domain service
  const handleScroll = (event: Event) => {
    const target = event.target as HTMLElement;
    scrollTop.value = target.scrollTop;
    // Invalidate cache on scroll
    cachedScrollMetrics = null;
  };
  
  const scrollToItem = async (index: number, alignment: 'start' | 'center' | 'end' | 'auto' = 'auto') => {
    if (!containerRef.value) return;
    
    const targetScrollTop = virtualScrollService.calculateScrollToItem(
      index,
      getConfig(),
      scrollTop.value,
      alignment
    );
    
    containerRef.value.scrollTop = targetScrollTop;
    scrollTop.value = targetScrollTop;
    
    // Invalidate cache after scroll
    cachedScrollMetrics = null;
    
    await nextTick();
  };
  
  const scrollToTop = async () => {
    await scrollToItem(0, 'start');
  };
  
  const scrollToBottom = async () => {
    await scrollToItem(totalItems.value - 1, 'end');
  };
  
  const getItemAtPosition = (y: number): number => {
    return Math.floor(y / itemHeight);
  };
  
  const isItemVisible = (index: number): boolean => {
    const range = visibleRange.value;
    return index >= range.start && index <= range.end;
  };
  
  const updateTotalItems = (count: number) => {
    totalItems.value = count;
    // Invalidate cache when total items change
    cachedConfig = null;
    cachedScrollMetrics = null;
  };
  
  // Validation using domain service
  const validateConfiguration = () => {
    return virtualScrollService.validateConfiguration(getConfig());
  };
  
  // Performance optimization methods
  const getPerformanceRecommendations = (renderTime: number, renderCalls: number) => {
    const metrics = {
      renderTime,
      itemCount: totalItems.value,
      visibleCount: visibleRange.value.end - visibleRange.value.start + 1,
      lastUpdate: Date.now(),
      renderCalls
    };
    
    return virtualScrollService.getPerformanceRecommendations(metrics);
  };
  
  // Lifecycle
  onMounted(() => {
    if (containerRef.value) {
      containerRef.value.addEventListener('scroll', handleScroll, { passive: true });
    }
    
    // Validate configuration on mount
    const validation = validateConfiguration();
    if (!validation.valid) {
      console.warn('Virtual scroll configuration issues:', validation.errors);
    }
  });
  
  onUnmounted(() => {
    if (containerRef.value) {
      containerRef.value.removeEventListener('scroll', handleScroll);
    }
    // Clean up cache
    cachedScrollMetrics = null;
    cachedConfig = null;
  });
  
  return {
    // Refs
    containerRef,
    scrollTop: computed(() => scrollTop.value),
    
    // Computed
    visibleStart,
    visibleEnd,
    visibleItems,
    totalHeight,
    offsetY,
    totalItems: computed(() => totalItems.value),
    scrollMetrics: computed(() => getScrollMetrics()),
    shouldUseVirtualization,
    
    // Methods
    handleScroll,
    scrollToItem,
    scrollToTop,
    scrollToBottom,
    getItemAtPosition,
    isItemVisible,
    updateTotalItems,
    validateConfiguration,
    getPerformanceRecommendations
  };
}