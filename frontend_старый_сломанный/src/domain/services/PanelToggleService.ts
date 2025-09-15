import { ref } from 'vue'
import { APP_CONFIG } from '@/config/app-config'

export class PanelToggleService {
  private activePanels = ref<string[]>(["files", "reports"])

  constructor() {}

  // Get active panels
  getActivePanels() {
    return this.activePanels.value
  }

  // Set active panels
  setActivePanels(panels: string[]) {
    this.activePanels.value = panels
  }

  // Toggle panel visibility
  togglePanel(panelId: string) {
    const index = this.activePanels.value.indexOf(panelId)
    if (index > -1) {
      this.activePanels.value.splice(index, 1)
    } else {
      this.activePanels.value.push(panelId)
    }
  }

  // Check if panel is active
  isPanelActive(panelId: string) {
    return this.activePanels.value.includes(panelId)
  }
}