import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue';

export interface VirtualScrollOptions {
  itemHeight: number;
  containerHeight: number;
  buffer: number; // Number of items to render outside visible area
  totalItems: number;
}

export interface VirtualScrollItem {
  index: number;
  top: number;
  height: number;
  data?: any;
}

/**
 * Composable for virtual scrolling to efficiently render large lists
 * This prevents DOM bloat when displaying large context content
 */
export function useVirtualScroll(options: VirtualScrollOptions) {
  const { itemHeight, containerHeight, buffer = 5 } = options;
  
  const scrollTop = ref(0);
  const containerRef = ref<HTMLElement | null>(null);
  const totalItems = ref(options.totalItems);
  
  // Computed properties
  const visibleStart = computed(() => {
    return Math.max(0, Math.floor(scrollTop.value / itemHeight) - buffer);
  });
  
  const visibleEnd = computed(() => {
    const visible = Math.ceil(containerHeight / itemHeight);
    return Math.min(totalItems.value - 1, visibleStart.value + visible + buffer * 2);
  });
  
  const visibleItems = computed((): VirtualScrollItem[] => {
    const items: VirtualScrollItem[] = [];
    for (let i = visibleStart.value; i <= visibleEnd.value; i++) {
      items.push({
        index: i,
        top: i * itemHeight,
        height: itemHeight
      });
    }
    return items;
  });
  
  const totalHeight = computed(() => {
    return totalItems.value * itemHeight;
  });
  
  const offsetY = computed(() => {
    return visibleStart.value * itemHeight;
  });
  
  // Methods
  const handleScroll = (event: Event) => {
    const target = event.target as HTMLElement;
    scrollTop.value = target.scrollTop;
  };
  
  const scrollToItem = async (index: number) => {
    if (!containerRef.value) return;
    
    const targetScrollTop = Math.max(0, index * itemHeight - containerHeight / 2);
    containerRef.value.scrollTop = targetScrollTop;
    scrollTop.value = targetScrollTop;
    
    await nextTick();
  };
  
  const scrollToTop = async () => {
    await scrollToItem(0);
  };
  
  const scrollToBottom = async () => {
    await scrollToItem(totalItems.value - 1);
  };
  
  const getItemAtPosition = (y: number): number => {
    return Math.floor(y / itemHeight);
  };
  
  const isItemVisible = (index: number): boolean => {
    return index >= visibleStart.value && index <= visibleEnd.value;
  };
  
  const updateTotalItems = (count: number) => {
    totalItems.value = count;
  };
  
  // Lifecycle
  onMounted(() => {
    if (containerRef.value) {
      containerRef.value.addEventListener('scroll', handleScroll, { passive: true });
    }
  });
  
  onUnmounted(() => {
    if (containerRef.value) {
      containerRef.value.removeEventListener('scroll', handleScroll);
    }
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
    
    // Methods
    handleScroll,
    scrollToItem,
    scrollToTop,
    scrollToBottom,
    getItemAtPosition,
    isItemVisible,
    updateTotalItems
  };
}