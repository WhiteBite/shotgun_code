import { APP_CONFIG } from '@/config/app-config';

export type TooltipPosition = 'top' | 'bottom' | 'left' | 'right' | 'auto';

export interface TooltipPositionResult {
  x: number;
  y: number;
  actualPosition: TooltipPosition;
  adjustments: string[];
}

export interface TooltipBounds {
  width: number;
  height: number;
  maxWidth: number;
  maxHeight: number;
}

export interface ViewportInfo {
  width: number;
  height: number;
  scrollX: number;
  scrollY: number;
}

export interface ElementBounds {
  x: number;
  y: number;
  width: number;
  height: number;
  top: number;
  left: number;
  right: number;
  bottom: number;
}

export interface PositionConstraints {
  minDistance: number;
  maxDistance: number;
  preferredPositions: TooltipPosition[];
  allowOverflow: boolean;
  stickToViewport: boolean;
}

/**
 * Domain service for tooltip positioning and collision detection
 * Encapsulates business rules for optimal tooltip placement
 * Follows DDD principles by containing positioning domain knowledge
 */
export class TooltipPositioningService {
  private readonly defaultOffset: number;
  private readonly arrowSize: number;
  private readonly viewportPadding: number;
  private readonly positionPriority: TooltipPosition[];

  constructor() {
    this.defaultOffset = APP_CONFIG.ui.tooltips.DEFAULT_OFFSET;
    this.arrowSize = APP_CONFIG.ui.tooltips.ARROW_SIZE;
    this.viewportPadding = APP_CONFIG.ui.tooltips.VIEWPORT_PADDING;
    this.positionPriority = ['top', 'bottom', 'right', 'left'];
  }

  /**
   * Calculates optimal tooltip position avoiding viewport boundaries
   * Business rule: Tooltips should be fully visible and not overlap trigger element
   */
  calculateOptimalPosition(
    triggerBounds: ElementBounds,
    tooltipBounds: TooltipBounds,
    preferredPosition: TooltipPosition,
    viewport: ViewportInfo,
    constraints?: Partial<PositionConstraints>
  ): TooltipPositionResult {
    const config = {
      minDistance: this.defaultOffset,
      maxDistance: this.defaultOffset * 3,
      preferredPositions: [preferredPosition, ...this.positionPriority],
      allowOverflow: false,
      stickToViewport: true,
      ...constraints
    };

    // Try preferred position first
    if (preferredPosition !== 'auto') {
      const result = this.tryPosition(
        preferredPosition,
        triggerBounds,
        tooltipBounds,
        viewport,
        config
      );
      
      if (result.fitsInViewport) {
        return {
          x: result.x,
          y: result.y,
          actualPosition: preferredPosition,
          adjustments: result.adjustments
        };
      }
    }

    // Try alternative positions
    for (const position of config.preferredPositions) {
      if (position === preferredPosition) continue;
      
      const result = this.tryPosition(
        position,
        triggerBounds,
        tooltipBounds,
        viewport,
        config
      );
      
      if (result.fitsInViewport) {
        return {
          x: result.x,
          y: result.y,
          actualPosition: position,
          adjustments: [...result.adjustments, `Moved from ${preferredPosition} to ${position}`]
        };
      }
    }

    // If no position fits perfectly, use the best fallback
    return this.calculateFallbackPosition(
      triggerBounds,
      tooltipBounds,
      preferredPosition,
      viewport,
      config
    );
  }

  /**
   * Tries to position tooltip in a specific direction
   * Business rule: Ensure minimum distance and avoid collisions
   */
  private tryPosition(
    position: TooltipPosition,
    triggerBounds: ElementBounds,
    tooltipBounds: TooltipBounds,
    viewport: ViewportInfo,
    constraints: PositionConstraints
  ): {
    x: number;
    y: number;
    fitsInViewport: boolean;
    adjustments: string[];
  } {
    const adjustments: string[] = [];
    let x: number;
    let y: number;

    switch (position) {
      case 'top':
        x = triggerBounds.x + (triggerBounds.width - tooltipBounds.width) / 2;
        y = triggerBounds.y - tooltipBounds.height - constraints.minDistance;
        break;

      case 'bottom':
        x = triggerBounds.x + (triggerBounds.width - tooltipBounds.width) / 2;
        y = triggerBounds.bottom + constraints.minDistance;
        break;

      case 'left':
        x = triggerBounds.x - tooltipBounds.width - constraints.minDistance;
        y = triggerBounds.y + (triggerBounds.height - tooltipBounds.height) / 2;
        break;

      case 'right':
        x = triggerBounds.right + constraints.minDistance;
        y = triggerBounds.y + (triggerBounds.height - tooltipBounds.height) / 2;
        break;

      default:
        throw new Error(`Invalid position: ${position}`);
    }

    // Apply viewport constraints
    const adjusted = this.constrainToViewport(
      { x, y },
      tooltipBounds,
      viewport,
      constraints.stickToViewport
    );

    if (adjusted.x !== x || adjusted.y !== y) {
      adjustments.push(`Adjusted for viewport boundaries`);
    }

    const fitsInViewport = this.checkViewportFit(
      adjusted,
      tooltipBounds,
      viewport
    );

    return {
      x: adjusted.x,
      y: adjusted.y,
      fitsInViewport,
      adjustments
    };
  }

  /**
   * Constrains tooltip position to viewport boundaries
   * Business rule: Keep tooltips visible within viewport when possible
   */
  private constrainToViewport(
    position: { x: number; y: number },
    tooltipBounds: TooltipBounds,
    viewport: ViewportInfo,
    stickToViewport: boolean
  ): { x: number; y: number } {
    if (!stickToViewport) {
      return position;
    }

    const padding = this.viewportPadding;
    
    let { x, y } = position;

    // Constrain horizontal position
    const leftBound = viewport.scrollX + padding;
    const rightBound = viewport.scrollX + viewport.width - tooltipBounds.width - padding;
    
    x = Math.max(leftBound, Math.min(rightBound, x));

    // Constrain vertical position
    const topBound = viewport.scrollY + padding;
    const bottomBound = viewport.scrollY + viewport.height - tooltipBounds.height - padding;
    
    y = Math.max(topBound, Math.min(bottomBound, y));

    return { x, y };
  }

  /**
   * Checks if tooltip fits within viewport at given position
   * Business rule: Tooltips should not extend beyond viewport boundaries
   */
  private checkViewportFit(
    position: { x: number; y: number },
    tooltipBounds: TooltipBounds,
    viewport: ViewportInfo
  ): boolean {
    const { x, y } = position;
    const { width, height } = tooltipBounds;

    const left = x;
    const right = x + width;
    const top = y;
    const bottom = y + height;

    const viewportLeft = viewport.scrollX;
    const viewportRight = viewport.scrollX + viewport.width;
    const viewportTop = viewport.scrollY;
    const viewportBottom = viewport.scrollY + viewport.height;

    return (
      left >= viewportLeft &&
      right <= viewportRight &&
      top >= viewportTop &&
      bottom <= viewportBottom
    );
  }

  /**
   * Calculates fallback position when no ideal position fits
   * Business rule: Prefer visibility over perfect positioning
   */
  private calculateFallbackPosition(
    triggerBounds: ElementBounds,
    tooltipBounds: TooltipBounds,
    preferredPosition: TooltipPosition,
    viewport: ViewportInfo,
    constraints: PositionConstraints
  ): TooltipPositionResult {
    // Find position with maximum visibility
    let bestPosition = preferredPosition;
    let bestVisibility = 0;
    let bestCoords = { x: 0, y: 0 };
    let adjustments: string[] = ['Using fallback positioning'];

    for (const position of this.positionPriority) {
      const result = this.tryPosition(
        position,
        triggerBounds,
        tooltipBounds,
        viewport,
        constraints
      );

      const visibility = this.calculateVisibilityRatio(
        result,
        tooltipBounds,
        viewport
      );

      if (visibility > bestVisibility) {
        bestVisibility = visibility;
        bestPosition = position;
        bestCoords = { x: result.x, y: result.y };
        adjustments = [...result.adjustments, 'Using fallback positioning'];
      }
    }

    return {
      x: bestCoords.x,
      y: bestCoords.y,
      actualPosition: bestPosition,
      adjustments
    };
  }

  /**
   * Calculates how much of the tooltip is visible in viewport
   * Business rule: Maximize tooltip visibility for better UX
   */
  private calculateVisibilityRatio(
    position: { x: number; y: number },
    tooltipBounds: TooltipBounds,
    viewport: ViewportInfo
  ): number {
    const { x, y } = position;
    const { width, height } = tooltipBounds;

    const tooltipLeft = x;
    const tooltipRight = x + width;
    const tooltipTop = y;
    const tooltipBottom = y + height;

    const viewportLeft = viewport.scrollX;
    const viewportRight = viewport.scrollX + viewport.width;
    const viewportTop = viewport.scrollY;
    const viewportBottom = viewport.scrollY + viewport.height;

    // Calculate intersection
    const intersectionLeft = Math.max(tooltipLeft, viewportLeft);
    const intersectionRight = Math.min(tooltipRight, viewportRight);
    const intersectionTop = Math.max(tooltipTop, viewportTop);
    const intersectionBottom = Math.min(tooltipBottom, viewportBottom);

    const intersectionWidth = Math.max(0, intersectionRight - intersectionLeft);
    const intersectionHeight = Math.max(0, intersectionBottom - intersectionTop);
    const intersectionArea = intersectionWidth * intersectionHeight;

    const tooltipArea = width * height;
    
    return tooltipArea > 0 ? intersectionArea / tooltipArea : 0;
  }

  /**
   * Detects collision with other elements
   * Business rule: Avoid overlapping with important UI elements
   */
  detectCollisions(
    tooltipPosition: { x: number; y: number },
    tooltipBounds: TooltipBounds,
    excludeElements: HTMLElement[] = []
  ): HTMLElement[] {
    const collisions: HTMLElement[] = [];
    
    // Get elements at tooltip position
    const elementsAtPosition = document.elementsFromPoint(
      tooltipPosition.x + tooltipBounds.width / 2,
      tooltipPosition.y + tooltipBounds.height / 2
    );

    for (const element of elementsAtPosition) {
      if (element instanceof HTMLElement && !excludeElements.includes(element)) {
        // Check if element has important UI role
        if (this.isImportantUIElement(element)) {
          collisions.push(element);
        }
      }
    }

    return collisions;
  }

  /**
   * Determines if an element is important for collision detection
   * Business rule: Avoid covering critical UI elements
   */
  private isImportantUIElement(element: HTMLElement): boolean {
    const importantRoles = ['button', 'link', 'menuitem', 'option', 'tab'];
    const importantTags = ['BUTTON', 'A', 'INPUT', 'SELECT', 'TEXTAREA'];
    const importantClasses = ['menu', 'dropdown', 'modal', 'dialog'];

    // Check ARIA role
    const role = element.getAttribute('role');
    if (role && importantRoles.includes(role)) {
      return true;
    }

    // Check tag name
    if (importantTags.includes(element.tagName)) {
      return true;
    }

    // Check classes
    const classList = element.className.toLowerCase();
    for (const importantClass of importantClasses) {
      if (classList.includes(importantClass)) {
        return true;
      }
    }

    return false;
  }

  /**
   * Calculates arrow position for tooltip
   * Business rule: Arrow should point to the center of trigger element
   */
  calculateArrowPosition(
    tooltipPosition: { x: number; y: number },
    tooltipBounds: TooltipBounds,
    triggerBounds: ElementBounds,
    actualPosition: TooltipPosition
  ): { x: number; y: number; rotation: number } | null {
    const triggerCenterX = triggerBounds.x + triggerBounds.width / 2;
    const triggerCenterY = triggerBounds.y + triggerBounds.height / 2;

    const tooltipLeft = tooltipPosition.x;
    const tooltipRight = tooltipPosition.x + tooltipBounds.width;
    const tooltipTop = tooltipPosition.y;
    const tooltipBottom = tooltipPosition.y + tooltipBounds.height;

    switch (actualPosition) {
      case 'top':
        return {
          x: Math.max(
            this.arrowSize,
            Math.min(
              tooltipBounds.width - this.arrowSize,
              triggerCenterX - tooltipLeft
            )
          ),
          y: tooltipBounds.height,
          rotation: 180
        };

      case 'bottom':
        return {
          x: Math.max(
            this.arrowSize,
            Math.min(
              tooltipBounds.width - this.arrowSize,
              triggerCenterX - tooltipLeft
            )
          ),
          y: 0,
          rotation: 0
        };

      case 'left':
        return {
          x: tooltipBounds.width,
          y: Math.max(
            this.arrowSize,
            Math.min(
              tooltipBounds.height - this.arrowSize,
              triggerCenterY - tooltipTop
            )
          ),
          rotation: 90
        };

      case 'right':
        return {
          x: 0,
          y: Math.max(
            this.arrowSize,
            Math.min(
              tooltipBounds.height - this.arrowSize,
              triggerCenterY - tooltipTop
            )
          ),
          rotation: -90
        };

      default:
        return null;
    }
  }

  /**
   * Gets positioning recommendations based on element context
   * Business rule: Suggest optimal positioning based on UI context
   */
  getPositioningRecommendations(
    triggerBounds: ElementBounds,
    viewport: ViewportInfo,
    context?: 'menu' | 'form' | 'content' | 'navigation'
  ): {
    recommendedPosition: TooltipPosition;
    reasons: string[];
    alternatives: TooltipPosition[];
  } {
    const reasons: string[] = [];
    const alternatives: TooltipPosition[] = [];

    // Analyze trigger position relative to viewport
    const isNearTop = triggerBounds.y < viewport.height * 0.3;
    const isNearBottom = triggerBounds.bottom > viewport.height * 0.7;
    const isNearLeft = triggerBounds.x < viewport.width * 0.3;
    const isNearRight = triggerBounds.right > viewport.width * 0.7;

    let recommendedPosition: TooltipPosition = 'auto';

    // Context-based recommendations
    if (context === 'navigation' || context === 'menu') {
      if (isNearTop) {
        recommendedPosition = 'bottom';
        reasons.push('Navigation elements work better with bottom tooltips');
      } else {
        recommendedPosition = 'top';
        reasons.push('Top position provides better visibility for navigation');
      }
    } else if (context === 'form') {
      recommendedPosition = 'right';
      reasons.push('Form tooltips should not interfere with vertical flow');
      alternatives.push('left', 'top');
    } else {
      // General content - position based on space availability
      if (isNearBottom) {
        recommendedPosition = 'top';
        reasons.push('Positioned above to avoid viewport bottom');
        alternatives.push('left', 'right');
      } else if (isNearTop) {
        recommendedPosition = 'bottom';
        reasons.push('Positioned below to avoid viewport top');
        alternatives.push('left', 'right');
      } else if (isNearRight) {
        recommendedPosition = 'left';
        reasons.push('Positioned left to avoid viewport right edge');
        alternatives.push('top', 'bottom');
      } else if (isNearLeft) {
        recommendedPosition = 'right';
        reasons.push('Positioned right to avoid viewport left edge');
        alternatives.push('top', 'bottom');
      } else {
        recommendedPosition = 'top';
        reasons.push('Top position provides optimal visibility');
        alternatives.push('bottom', 'right', 'left');
      }
    }

    return {
      recommendedPosition,
      reasons,
      alternatives
    };
  }
}

// Default instance for dependency injection
export const defaultTooltipPositioningService = new TooltipPositioningService();