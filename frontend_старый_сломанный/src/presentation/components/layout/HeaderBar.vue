<template>
  <header class="bg-gray-800 border-b border-gray-700 px-4 py-3 flex items-center justify-between min-h-[64px]">
    <!-- Project Info Card -->
    <div class="flex items-center space-x-4">
      <div 
        v-tooltip="{
          content: getProjectTooltipContent,
          position: 'bottom',
          maxWidth: 300,
          allowHTML: true
        }"
        class="flex items-center space-x-3 cursor-pointer hover:bg-gray-700/50 rounded-lg px-2 py-1 transition-colors"
        @click="openProjectSettings"
      >
        <!-- Project Icon -->
        <div 
          v-tooltip="'Project root directory'"
          class="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center"
        >
          <FolderIcon class="w-5 h-5 text-white" />
        </div>
        
        <!-- Project Details -->
        <div class="flex flex-col">
          <h1 class="text-lg font-semibold text-gray-100 truncate max-w-xs">
            {{ projectName }}
          </h1>
          <div class="flex items-center space-x-2 text-xs text-gray-400">
            <span 
              v-tooltip="{
                content: `Full path: ${projectPath}`,
                position: 'bottom'
              }" 
              class="truncate max-w-64 hover:text-gray-200 transition-colors"
              :title="projectPath"
            >
              {{ truncatedPath }}
            </span>
            <div 
              v-tooltip="{
                content: getStatusTooltipContent,
                position: 'bottom',
                allowHTML: true
              }"
              class="flex items-center space-x-1"
            >
              <StatusIndicator :status="projectStatus" />
              <span class="capitalize">{{ projectStatus }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Mode Toggle -->
    <div class="flex items-center space-x-6">
      <div class="flex items-center space-x-3">
        <span 
          v-tooltip="'Switch between manual control and autonomous AI mode'"
          class="text-sm text-gray-300 font-medium"
        >
          Mode:
        </span>
        <div class="relative">
          <div 
            v-tooltip="{
              content: getModeToggleTooltip,
              position: 'bottom',
              maxWidth: 250,
              allowHTML: true,
              interactive: true
            }"
            class="flex items-center bg-gray-700 rounded-lg p-1 cursor-pointer transition-all duration-300"
            :class="headerBarService.getModeToggleClasses()"
            @click="handleModeToggle"
          >
            <!-- Manual Mode Button -->
            <button
              v-tooltip="{
                content: getManualModeTooltip,
                position: 'bottom',
                maxWidth: 200,
                allowHTML: true
              }"
              class="px-4 py-2 rounded-md text-sm font-medium transition-all duration-200 min-w-[80px]"
              :class="headerBarService.getModeButtonClasses(true)"
              :disabled="workspaceStore.isTransitioning"
              @click.stop="workspaceStore.setMode('manual')"
            >
              <div class="flex items-center justify-center space-x-1">
                <CodeBracketIcon class="w-4 h-4" />
                <span>Manual</span>
              </div>
            </button>
            
            <!-- Autonomous Mode Button -->
            <button
              v-tooltip="{
                content: getAutonomousModeTooltip,
                position: 'bottom',
                maxWidth: 200,
                allowHTML: true
              }"
              class="px-4 py-2 rounded-md text-sm font-medium transition-all duration-200 min-w-[80px]"
              :class="headerBarService.getModeButtonClasses(false)"
              :disabled="workspaceStore.isTransitioning"
              @click.stop="workspaceStore.setMode('autonomous')"
            >
              <div class="flex items-center justify-center space-x-1">
                <CpuChipIcon class="w-4 h-4" />
                <span>Auto</span>
              </div>
            </button>
          </div>
          
          <!-- Loading indicator during transition -->
          <div 
            v-if="workspaceStore.isTransitioning"
            class="absolute inset-0 flex items-center justify-center bg-gray-700 bg-opacity-50 rounded-lg"
          >
            <LoadingSpinner class="w-4 h-4" />
          </div>
        </div>
      </div>
    </div>

    <!-- Quick Actions -->
    <div class="flex items-center space-x-2">
      <!-- Language Switcher -->
      <div class="relative">
        <IconButton
          v-tooltip="{
            content: 'Switch Language (Переключить язык)',
            position: 'bottom',
            allowHTML: true
          }"
          icon="LanguageIcon"
          @click="toggleLanguageMenu"
        />
        
        <!-- Language Dropdown Menu -->
        <div 
          v-if="isLanguageMenuOpen"
          class="absolute right-0 mt-2 w-48 rounded-md shadow-lg bg-gray-800 border border-gray-700 z-50"
        >
          <div class="py-1">
            <button 
              class="w-full px-4 py-2 text-sm text-left text-gray-300 hover:bg-gray-700 flex items-center space-x-2"
              @click="changeLanguage('en')"
            >
              <span class="opacity-70">🇬🇧</span>
              <span>English</span>
              <span v-if="currentLanguage === 'en'" class="ml-auto text-blue-400">✓</span>
            </button>
            <button 
              class="w-full px-4 py-2 text-sm text-left text-gray-300 hover:bg-gray-700 flex items-center space-x-2"
              @click="changeLanguage('ru')"
            >
              <span class="opacity-70">🇷🇺</span>
              <span>Русский</span>
              <span v-if="currentLanguage === 'ru'" class="ml-auto text-blue-400">✓</span>
            </button>
          </div>
        </div>
      </div>
      
      <!-- Context Build Status -->
      <div 
        v-if="contextStore.buildStatus !== 'idle'" 
        v-tooltip="{
          content: getContextStatusTooltip,
          position: 'bottom',
          maxWidth: 250,
          allowHTML: true
        }"
        class="flex items-center space-x-2 px-3 py-1.5 rounded-lg bg-gray-700 cursor-pointer hover:bg-gray-600 transition-colors"
        @click="openContextDetails"
      >
        <LoadingSpinner 
          v-if="contextStore.buildStatus === 'building'" 
          class="w-4 h-4 text-blue-400" 
        />
        <CheckCircleIcon 
          v-else-if="contextStore.buildStatus === 'complete'" 
          class="w-4 h-4 text-green-400" 
        />
        <ExclamationCircleIcon 
          v-else-if="contextStore.buildStatus === 'error'" 
          class="w-4 h-4 text-red-400" 
        />
        <span class="text-xs text-gray-300">
          {{ getContextStatusText() }}
        </span>
      </div>

      <!-- Action Buttons -->
      <div class="flex items-center space-x-1">
        <!-- Settings -->
        <IconButton
          v-tooltip="{
            content: 'Open Settings<br><kbd>Ctrl+,</kbd>',
            position: 'bottom',
            allowHTML: true
          }"
          icon="Cog6ToothIcon"
          @click="openSettings"
        />
        
        <!-- Export -->
        <IconButton
          v-tooltip="{
            content: contextStore.hasSelectedFiles ? 
              'Export current context<br><kbd>Ctrl+E</kbd>' : 
              'No files selected for export',
            position: 'bottom',
            allowHTML: true
          }"
          icon="ArrowUpTrayIcon"
          :disabled="!contextStore.hasSelectedFiles"
          @click="openExport"
        />
        
        <!-- Help -->
        <IconButton
          v-tooltip="{
            content: 'Help & Keyboard Shortcuts<br><kbd>F1</kbd> or <kbd>?</kbd>',
            position: 'bottom',
            allowHTML: true
          }"
          icon="QuestionMarkCircleIcon"
          @click="showHelp"
        />
      </div>
    </div>
  </header>
</template>