<template>
  <div class="h-full flex flex-col bg-transparent">
    <!-- Header with tabs -->
    <div class="border-b border-gray-700/30">
      <div class="flex items-center justify-between p-3">
        <div class="section-title">
          <div class="section-icon section-icon-orange">
            <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
            </svg>
          </div>
          <h2 class="section-title-text">{{ t('git.title') }}</h2>
        </div>
        <!-- Recent repos dropdown -->
        <div v-if="recentRepos.length > 0" class="relative">
          <button
            @click="showRecentRepos = !showRecentRepos"
            class="text-xs text-gray-400 hover:text-white flex items-center gap-1 px-2 py-1 rounded hover:bg-gray-700/50"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            {{ t('git.recent') }}
          </button>
          <div
            v-if="showRecentRepos"
            class="absolute right-0 top-full mt-1 w-72 bg-gray-800 border border-gray-700 rounded-lg shadow-xl z-50 overflow-hidden"
          >
            <div class="p-2 border-b border-gray-700 flex items-center justify-between">
              <span class="text-xs text-gray-400">{{ t('git.recentRepos') }}</span>
              <button @click="clearRecentRepos" class="text-xs text-red-400 hover:text-red-300">
                {{ t('git.clearHistory') }}
              </button>
            </div>
            <div class="max-h-48 overflow-y-auto">
              <button
                v-for="repo in recentRepos"
                :key="repo.url"
                @click="loadRecentRepo(repo)"
                class="w-full px-3 py-2 text-left hover:bg-gray-700/50 flex items-center gap-2"
              >
                <svg v-if="repo.isGitHub" class="w-4 h-4 text-gray-400 flex-shrink-0" fill="currentColor" viewBox="0 0 24 24">
                  <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
                </svg>
                <svg v-else class="w-4 h-4 text-gray-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
                </svg>
                <div class="flex-1 min-w-0">
                  <p class="text-sm text-white truncate">{{ repo.name }}</p>
                  <p class="text-xs text-gray-500 truncate">{{ repo.url }}</p>
                </div>
              </button>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Source Type Tabs -->
      <div class="flex gap-1 px-2 pb-2">
        <button
          @click="sourceType = 'local'"
          :class="['tab-btn', sourceType === 'local' ? 'tab-btn-active tab-btn-active-indigo' : 'tab-btn-inactive']"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
          </svg>
          {{ t('git.localGit') }}
        </button>
        <button
          @click="sourceType = 'remote'"
          :class="['tab-btn', sourceType === 'remote' ? 'tab-btn-active tab-btn-active-purple' : 'tab-btn-inactive']"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
          </svg>
          {{ t('git.remoteUrl') }}
        </button>
      </div>
    </div>

    <!-- Local Git Panel -->
    <div v-if="sourceType === 'local'" class="flex-1 overflow-auto">
      <div v-if="!isGitRepo" class="p-4 text-center text-gray-400">
        <svg class="w-12 h-12 mx-auto mb-3 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
        </svg>
        <p>{{ t('git.notGitRepo') }}</p>
        <p class="text-sm mt-2">{{ t('git.openGitProject') }}</p>
      </div>

      <div v-else class="p-4 space-y-4">
        <!-- Current Branch Info -->
        <div class="branch-info">
          <svg class="w-5 h-5 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
          </svg>
          <div class="flex-1">
            <span class="text-xs text-gray-400">{{ t('git.currentBranch') }}:</span>
            <span class="ml-2 text-sm text-white font-medium">{{ currentBranch }}</span>
          </div>
          <button 
            @click="selectRef(currentBranch)" 
            class="btn btn-primary btn-sm"
            :disabled="selectedRef === currentBranch"
          >
            {{ t('git.load') }}
          </button>
        </div>

        <!-- Branch/Commit Selection -->
        <div class="space-y-3">
          <div class="flex border-b border-gray-700">
            <button
              @click="switchToRefType('branches')"
              :class="['tab-btn text-sm', refType === 'branches' ? 'tab-btn-active' : 'tab-btn-inactive']"
            >
              {{ t('git.branches') }} ({{ branches.length }})
            </button>
            <button
              @click="switchToRefType('commits')"
              :class="['tab-btn text-sm', refType === 'commits' ? 'tab-btn-active' : 'tab-btn-inactive']"
            >
              {{ t('git.commits') }} {{ commitsLoaded ? `(${commits.length})` : '' }}
            </button>
            <button
              v-if="branches.length >= 2"
              @click="openDiffModal"
              class="ml-auto px-2 py-1 text-xs text-purple-400 hover:text-purple-300 hover:bg-purple-500/10 rounded transition-colors flex items-center gap-1"
            >
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
              </svg>
              {{ t('git.diff') }}
            </button>
          </div>

          <!-- Selected Ref Info -->
          <div v-if="selectedRef" class="p-3 bg-indigo-900/30 rounded-lg border border-indigo-500/30">
            <div class="flex items-center justify-between">
              <div>
                <span class="text-xs text-indigo-300">{{ t('git.buildingFrom') }}:</span>
                <span class="ml-2 text-sm text-white font-medium">{{ selectedRef }}</span>
                <span v-if="selectedRef === currentBranch" class="text-xs text-emerald-400 ml-2">({{ t('git.current') }})</span>
              </div>
              <button @click="clearSelectedRef" class="text-xs text-gray-400 hover:text-white">
                {{ t('git.clear') }}
              </button>
            </div>
            <p class="text-xs text-gray-400 mt-1">{{ t('git.noCheckout') }}</p>
          </div>

          <!-- Branches List -->
          <div v-if="refType === 'branches'" class="max-h-64 overflow-y-auto space-y-1">
            <button
              v-for="branch in branches"
              :key="branch"
              @click="selectRef(branch)"
              :class="[
                'w-full px-3 py-2 text-left text-sm rounded transition-colors flex items-center gap-2',
                selectedRef === branch
                  ? 'bg-indigo-600/20 text-indigo-300 border border-indigo-500/30'
                  : branch === currentBranch
                    ? 'bg-emerald-900/20 text-emerald-300 border border-emerald-500/20'
                    : 'text-gray-300 hover:bg-gray-700'
              ]"
            >
              <svg class="w-4 h-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
              <span class="truncate">{{ branch }}</span>
              <span v-if="branch === currentBranch" class="text-xs text-emerald-400 ml-auto">({{ t('git.current') }})</span>
              <span v-else-if="selectedRef === branch" class="text-xs text-indigo-400 ml-auto">({{ t('git.selected') }})</span>
            </button>
          </div>

          <!-- Commits List -->
          <div v-if="refType === 'commits'" class="max-h-64 overflow-y-auto space-y-1">
            <button
              v-for="commit in commits"
              :key="commit.hash"
              @click="selectRef(commit.hash)"
              :class="[
                'w-full px-3 py-2 text-left text-sm rounded transition-colors',
                selectedRef === commit.hash
                  ? 'bg-indigo-600/20 text-indigo-300 border border-indigo-500/30'
                  : 'text-gray-300 hover:bg-gray-700'
              ]"
            >
              <div class="flex items-start gap-2">
                <code class="text-xs text-amber-400 font-mono flex-shrink-0">{{ commit.hash.slice(0, 7) }}</code>
                <div class="flex-1 min-w-0">
                  <p class="truncate text-white">{{ commit.subject }}</p>
                  <p class="text-xs text-gray-500">{{ commit.author }} • {{ formatDate(commit.date) }}</p>
                </div>
              </div>
            </button>
          </div>
        </div>

        <!-- File Tree at Selected Ref -->
        <div v-if="selectedRef && filesAtRef && filesAtRef.length > 0" class="space-y-2">
          <div class="flex items-center justify-between">
            <span class="text-sm text-gray-300">{{ t('git.filesAtRef') }} {{ selectedRef?.slice(0, 7) }}</span>
            <span class="text-xs text-gray-500">{{ filteredLocalFiles.length }}/{{ filesAtRef?.length || 0 }}</span>
          </div>
          
          <!-- Search & Filter Bar -->
          <div class="space-y-2">
            <div class="relative">
              <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
              <input
                v-model="localSearchQuery"
                type="text"
                :placeholder="t('git.searchFiles')"
                class="input pl-10 py-1.5 text-sm"
              />
              <button
                v-if="localSearchQuery"
                @click="localSearchQuery = ''"
                class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-white"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            
            <!-- File Type Filters -->
            <div class="flex flex-wrap gap-1">
              <button
                v-for="filter in fileTypeFilters"
                :key="filter.id"
                @click="toggleLocalFilter(filter.id)"
                :class="[
                  'px-2 py-0.5 text-xs rounded-full transition-colors',
                  localActiveFilters.has(filter.id)
                    ? 'bg-indigo-500/30 text-indigo-300 border border-indigo-500/50'
                    : 'bg-gray-800/50 text-gray-400 border border-gray-700/50 hover:bg-gray-700/50'
                ]"
              >
                {{ filter.label }} ({{ getLocalFilterCount(filter.id) }})
              </button>
            </div>
          </div>
          
          <div class="file-list-container">
            <SimpleFileTree
              :files="filteredLocalFiles"
              :selected-paths="selectedFiles"
              @toggle-select="toggleFileSelection"
              @select-folder="selectFolderFiles"
              @preview-file="handlePreviewFile"
            />
          </div>
          <div class="flex gap-2">
            <button @click="selectAllFiles" class="action-btn action-btn-success btn-sm flex-1">
              {{ t('git.selectAll') }}
            </button>
            <button @click="clearFileSelection" class="action-btn action-btn-danger btn-sm flex-1">
              {{ t('git.clearSelection') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Remote URL Panel -->
    <div v-if="sourceType === 'remote'" class="flex-1 overflow-auto p-4 space-y-4">
      <!-- URL Input -->
      <div class="space-y-3">
        <div>
          <label class="block text-sm text-gray-300 mb-2">{{ t('git.repoUrl') }}</label>
          <div class="flex gap-2">
            <input
              v-model="remoteUrl"
              type="text"
              :placeholder="t('git.urlPlaceholder')"
              class="input flex-1"
              @keydown.enter="loadRemoteRepo"
            />
            <button
              @click="loadRemoteRepo"
              :disabled="!remoteUrl || isLoadingRemote"
              class="btn btn-primary"
            >
              <svg v-if="isLoadingRemote" class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
              </svg>
              <span v-else>{{ t('git.load') || 'Load' }}</span>
            </button>
          </div>
        </div>
        <p class="text-xs text-gray-500">
          <span v-if="isGitHubRepo" class="text-emerald-400">✓ GitHub API (fast, no clone)</span>
          <span v-else-if="isGitLabRepo" class="text-orange-400">✓ GitLab API (fast, no clone)</span>
          <span v-else>{{ t('git.urlHint') }}</span>
        </p>
      </div>

      <!-- Remote Repo Loaded - Show branches/files -->
      <div v-if="remoteRepoLoaded" class="space-y-4">
        <!-- Branch selector -->
        <div class="branch-info">
          <svg class="w-5 h-5 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
          </svg>
          <div class="flex-1">
            <select v-model="remoteSelectedBranch" @change="loadRemoteFiles" class="input w-full text-sm">
              <option v-for="branch in remoteBranches" :key="branch.name" :value="branch.name">
                {{ branch.name }}
              </option>
            </select>
          </div>
        </div>

        <!-- Files list -->
        <div v-if="remoteFiles && remoteFiles.length > 0" class="space-y-2">
          <div class="flex items-center justify-between">
            <span class="text-sm text-gray-300">{{ t('git.filesAtRef') }} {{ remoteSelectedBranch }}</span>
            <span class="text-xs text-gray-500">{{ filteredRemoteFiles.length }}/{{ remoteFiles?.length || 0 }}</span>
          </div>
          
          <!-- Search & Filter Bar for Remote -->
          <div class="space-y-2">
            <div class="relative">
              <svg class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
              <input
                v-model="remoteSearchQuery"
                type="text"
                :placeholder="t('git.searchFiles')"
                class="input pl-10 py-1.5 text-sm"
              />
              <button
                v-if="remoteSearchQuery"
                @click="remoteSearchQuery = ''"
                class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-white"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            
            <!-- File Type Filters -->
            <div class="flex flex-wrap gap-1">
              <button
                v-for="filter in fileTypeFilters"
                :key="filter.id"
                @click="toggleRemoteFilter(filter.id)"
                :class="[
                  'px-2 py-0.5 text-xs rounded-full transition-colors',
                  remoteActiveFilters.has(filter.id)
                    ? 'bg-indigo-500/30 text-indigo-300 border border-indigo-500/50'
                    : 'bg-gray-800/50 text-gray-400 border border-gray-700/50 hover:bg-gray-700/50'
                ]"
              >
                {{ filter.label }} ({{ getRemoteFilterCount(filter.id) }})
              </button>
            </div>
          </div>
          
          <div class="file-list-container">
            <SimpleFileTree
              :files="filteredRemoteFiles"
              :selected-paths="remoteSelectedFiles"
              @toggle-select="toggleRemoteFileSelection"
              @select-folder="selectRemoteFolderFiles"
              @preview-file="handlePreviewFile"
            />
          </div>
          <div class="flex gap-2">
            <button @click="selectAllRemoteFiles" class="action-btn action-btn-success btn-sm flex-1">
              {{ t('git.selectAll') }}
            </button>
            <button @click="clearRemoteFileSelection" class="action-btn action-btn-danger btn-sm flex-1">
              {{ t('git.clearSelection') }}
            </button>
          </div>
        </div>
      </div>

      <!-- Fallback: Clone option for non-GitHub/GitLab repos -->
      <div v-if="!isGitHubRepo && !isGitLabRepo && remoteUrl && !remoteRepoLoaded" class="p-3 bg-amber-900/20 rounded-lg border border-amber-500/30">
        <p class="text-xs text-amber-300 mb-2">{{ t('git.cloneRequired') }}</p>
        <button
          @click="cloneRemote"
          :disabled="isCloning"
          class="action-btn action-btn-accent w-full"
        >
          {{ isCloning ? t('git.cloning') : t('git.clone') }}
        </button>
      </div>

      <!-- Cloned Repo Info -->
      <div v-if="clonedPath" class="p-3 bg-emerald-900/30 rounded-lg border border-emerald-500/30">
        <div class="flex items-center gap-2 mb-2">
          <svg class="w-4 h-4 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
          </svg>
          <span class="text-sm text-emerald-300">{{ t('git.cloned') }}</span>
        </div>
        <p class="text-xs text-gray-400 truncate">{{ clonedPath }}</p>
        <div class="flex gap-2 mt-3">
          <button @click="openClonedRepo" class="btn btn-primary btn-sm flex-1">
            {{ t('git.openProject') }}
          </button>
          <button @click="cleanupClonedRepo" class="btn btn-ghost btn-sm">
            {{ t('git.remove') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Bottom Panel for Remote Files -->
    <div v-if="sourceType === 'remote' && remoteSelectedFiles.size > 0" class="border-t border-gray-700 p-4">
      <div class="flex items-center justify-between mb-3">
        <span class="text-sm text-gray-300">{{ remoteSelectedFiles.size }} {{ t('git.filesSelected') }}</span>
      </div>
      <button
        @click="buildContextFromRemote"
        :disabled="isBuilding"
        class="btn btn-primary w-full"
      >
        <svg v-if="isBuilding" class="animate-spin w-4 h-4 mr-2" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
        </svg>
        {{ t('git.buildContext') }} {{ remoteSelectedBranch }}
      </button>
    </div>

    <!-- Loading Overlay -->
    <div v-if="isLoading" class="loading-overlay absolute">
      <div class="text-center">
        <svg class="loading-spinner mx-auto mb-2" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
        </svg>
        <span class="text-sm text-gray-400">{{ loadingMessage }}</span>
      </div>
    </div>

    <!-- Bottom Panel - Build Context -->
    <div v-if="selectedFiles.size > 0" class="border-t border-gray-700 p-4">
      <div class="flex items-center justify-between mb-3">
        <span class="text-sm text-gray-300">{{ selectedFiles.size }} {{ t('git.filesSelected') }}</span>
      </div>
      <button
        @click="buildContextFromRef"
        :disabled="isBuilding"
        class="btn btn-primary w-full"
      >
        <svg v-if="isBuilding" class="animate-spin w-4 h-4 mr-2" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"></path>
        </svg>
        {{ t('git.buildContext') }} {{ selectedRef?.slice(0, 7) || 'ref' }}
      </button>
    </div>

    <!-- File Preview Modal -->
    <FilePreviewModal
      :is-open="previewOpen"
      :file-path="previewPath"
      :content="previewContent"
      :is-loading="previewLoading"
      :error="previewError"
      @close="closePreview"
    />

    <!-- Branch Diff Modal -->
    <BranchDiffModal
      :is-open="diffModalOpen"
      :branches="branches"
      :project-path="projectPath"
      :current-branch="currentBranch"
      @close="closeDiffModal"
    />
  </div>
</template>

<script setup lang="ts">
import FilePreviewModal from '@/components/FilePreviewModal.vue'
import SimpleFileTree from '@/components/SimpleFileTree.vue'
import { useI18n } from '@/composables/useI18n'
import { useContextStore } from '@/features/context'
import { apiService, type CommitInfo } from '@/services/api.service'
import { useProjectStore } from '@/stores/project.store'
import { useUIStore } from '@/stores/ui.store'
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'

const { t } = useI18n()
const projectStore = useProjectStore()
const uiStore = useUIStore()
const contextStore = useContextStore()

const sourceType = ref<'local' | 'remote'>('local')
const refType = ref<'branches' | 'commits'>('branches')

// Local git state
const isGitRepo = ref(false)
const currentBranch = ref('')
const branches = ref<string[]>([])
const commits = ref<CommitInfo[]>([])
const selectedRef = ref<string | null>(null)
const filesAtRef = ref<string[]>([])
const selectedFiles = ref<Set<string>>(new Set())

// Remote state
const remoteUrl = ref('')
const isCloning = ref(false)
const clonedPath = ref<string | null>(null)
const isGitHubRepo = ref(false)
const isGitLabRepo = ref(false)
const isLoadingRemote = ref(false)
const remoteRepoLoaded = ref(false)
const remoteBranches = ref<{ name: string; commit: { sha: string } }[]>([])
const remoteSelectedBranch = ref('')
const remoteFiles = ref<string[]>([])
const remoteSelectedFiles = ref<Set<string>>(new Set())

// Loading state
const isLoading = ref(false)
const loadingMessage = ref('')
const isBuilding = ref(false)

// Search & Filter state
const localSearchQuery = ref('')
const remoteSearchQuery = ref('')
const localActiveFilters = ref<Set<string>>(new Set())
const remoteActiveFilters = ref<Set<string>>(new Set())

// Recent repos
interface RecentRepo {
  url: string
  name: string
  isGitHub: boolean
  lastUsed: number
}
const recentRepos = ref<RecentRepo[]>([])
const showRecentRepos = ref(false)

// File type filters
const fileTypeFilters = [
  { id: 'code', label: 'Code', extensions: ['.ts', '.js', '.tsx', '.jsx', '.vue', '.go', '.py', '.java', '.rs', '.cpp', '.c', '.h', '.cs', '.rb', '.php', '.swift', '.kt'] },
  { id: 'styles', label: 'Styles', extensions: ['.css', '.scss', '.sass', '.less', '.styl'] },
  { id: 'config', label: 'Config', extensions: ['.json', '.yaml', '.yml', '.toml', '.xml', '.ini', '.env', '.config'] },
  { id: 'docs', label: 'Docs', extensions: ['.md', '.txt', '.rst', '.adoc'] },
]

const projectPath = computed(() => projectStore.currentPath || '')

// Filtered files computed
const filteredLocalFiles = computed(() => {
  let files = filesAtRef.value || []
  
  // Apply search
  if (localSearchQuery.value) {
    const query = localSearchQuery.value.toLowerCase()
    files = files.filter(f => f.toLowerCase().includes(query))
  }
  
  // Apply type filters
  if (localActiveFilters.value.size > 0) {
    const activeExtensions = new Set<string>()
    localActiveFilters.value.forEach(filterId => {
      const filter = fileTypeFilters.find(f => f.id === filterId)
      if (filter) filter.extensions.forEach(ext => activeExtensions.add(ext))
    })
    files = files.filter(f => {
      const ext = '.' + f.split('.').pop()?.toLowerCase()
      return activeExtensions.has(ext)
    })
  }
  
  return files
})

const filteredRemoteFiles = computed(() => {
  let files = remoteFiles.value || []
  
  // Apply search
  if (remoteSearchQuery.value) {
    const query = remoteSearchQuery.value.toLowerCase()
    files = files.filter(f => f.toLowerCase().includes(query))
  }
  
  // Apply type filters
  if (remoteActiveFilters.value.size > 0) {
    const activeExtensions = new Set<string>()
    remoteActiveFilters.value.forEach(filterId => {
      const filter = fileTypeFilters.find(f => f.id === filterId)
      if (filter) filter.extensions.forEach(ext => activeExtensions.add(ext))
    })
    files = files.filter(f => {
      const ext = '.' + f.split('.').pop()?.toLowerCase()
      return activeExtensions.has(ext)
    })
  }
  
  return files
})

// Filter counts
function getLocalFilterCount(filterId: string): number {
  const filter = fileTypeFilters.find(f => f.id === filterId)
  if (!filter || !filesAtRef.value) return 0
  return filesAtRef.value.filter(f => {
    const ext = '.' + f.split('.').pop()?.toLowerCase()
    return filter.extensions.includes(ext)
  }).length
}

function getRemoteFilterCount(filterId: string): number {
  const filter = fileTypeFilters.find(f => f.id === filterId)
  if (!filter || !remoteFiles.value) return 0
  return remoteFiles.value.filter(f => {
    const ext = '.' + f.split('.').pop()?.toLowerCase()
    return filter.extensions.includes(ext)
  }).length
}

function toggleLocalFilter(filterId: string) {
  if (localActiveFilters.value.has(filterId)) {
    localActiveFilters.value.delete(filterId)
  } else {
    localActiveFilters.value.add(filterId)
  }
  localActiveFilters.value = new Set(localActiveFilters.value)
}

function toggleRemoteFilter(filterId: string) {
  if (remoteActiveFilters.value.has(filterId)) {
    remoteActiveFilters.value.delete(filterId)
  } else {
    remoteActiveFilters.value.add(filterId)
  }
  remoteActiveFilters.value = new Set(remoteActiveFilters.value)
}

// Recent repos management
const RECENT_REPOS_KEY = 'git-recent-repos'
const MAX_RECENT_REPOS = 10
const SELECTION_KEY = 'git-file-selection'

function loadRecentReposFromStorage() {
  try {
    const stored = localStorage.getItem(RECENT_REPOS_KEY)
    if (stored) {
      recentRepos.value = JSON.parse(stored)
    }
  } catch (e) {
    console.error('Failed to load recent repos:', e)
  }
}

// Selection persistence
function saveSelectionToStorage(key: string, files: Set<string>) {
  try {
    const data = {
      files: Array.from(files),
      timestamp: Date.now()
    }
    localStorage.setItem(`${SELECTION_KEY}-${key}`, JSON.stringify(data))
  } catch (e) {
    console.error('Failed to save selection:', e)
  }
}

function loadSelectionFromStorage(key: string): Set<string> {
  try {
    const stored = localStorage.getItem(`${SELECTION_KEY}-${key}`)
    if (stored) {
      const data = JSON.parse(stored)
      // Only restore if less than 24 hours old
      if (Date.now() - data.timestamp < 24 * 60 * 60 * 1000) {
        return new Set(data.files)
      }
    }
  } catch (e) {
    console.error('Failed to load selection:', e)
  }
  return new Set()
}

function saveRecentRepo(url: string, name: string, isGitHub: boolean) {
  const existing = recentRepos.value.findIndex(r => r.url === url)
  if (existing >= 0) {
    recentRepos.value.splice(existing, 1)
  }
  
  recentRepos.value.unshift({
    url,
    name,
    isGitHub,
    lastUsed: Date.now()
  })
  
  if (recentRepos.value.length > MAX_RECENT_REPOS) {
    recentRepos.value = recentRepos.value.slice(0, MAX_RECENT_REPOS)
  }
  
  localStorage.setItem(RECENT_REPOS_KEY, JSON.stringify(recentRepos.value))
}

function loadRecentRepo(repo: RecentRepo) {
  remoteUrl.value = repo.url
  showRecentRepos.value = false
  sourceType.value = 'remote'
  loadRemoteRepo()
}

function clearRecentRepos() {
  recentRepos.value = []
  localStorage.removeItem(RECENT_REPOS_KEY)
  showRecentRepos.value = false
}

// Close dropdown on outside click
function handleClickOutside(e: MouseEvent) {
  const target = e.target as HTMLElement
  if (!target.closest('.relative')) {
    showRecentRepos.value = false
  }
}

async function checkGitRepo() {
  if (!projectPath.value) return
  
  isLoading.value = true
  loadingMessage.value = 'Checking repository...'
  
  // Reset lazy loading flags
  branchesLoaded.value = false
  commitsLoaded.value = false
  branches.value = []
  commits.value = []
  
  try {
    isGitRepo.value = await apiService.isGitRepository(projectPath.value)
    
    if (isGitRepo.value) {
      currentBranch.value = await apiService.getCurrentBranch(projectPath.value)
      await loadBranches()
      // Don't load commits until user switches to commits tab
    }
  } catch (error) {
    console.error('Failed to check git repo:', error)
  } finally {
    isLoading.value = false
  }
}

const branchesLoaded = ref(false)
const commitsLoaded = ref(false)

async function loadBranches() {
  if (branchesLoaded.value) return
  
  try {
    const result = await apiService.getBranches(projectPath.value)
    branches.value = JSON.parse(result)
    branchesLoaded.value = true
  } catch (error) {
    console.error('Failed to load branches:', error)
  }
}

async function loadCommits() {
  if (commitsLoaded.value) return
  
  try {
    commits.value = await apiService.getCommitHistory(projectPath.value, 50)
    commitsLoaded.value = true
  } catch (error) {
    console.error('Failed to load commits:', error)
  }
}

function switchToRefType(type: 'branches' | 'commits') {
  refType.value = type
  
  // Lazy load commits when switching to commits tab
  if (type === 'commits' && !commitsLoaded.value) {
    loadCommits()
  }
}

async function selectRef(ref: string) {
  selectedRef.value = ref
  selectedFiles.value = new Set()
  filesAtRef.value = []

  isLoading.value = true
  loadingMessage.value = `Loading files at ${ref.slice(0, 7)}...`

  try {
    const files = await apiService.listFilesAtRef(projectPath.value, ref)
    filesAtRef.value = files || []
    console.log(`[Git] Loaded ${filesAtRef.value.length} files at ref ${ref}`)
    
    // Restore previous selection if available
    const key = `local-${projectPath.value}-${ref}`
    const savedSelection = loadSelectionFromStorage(key)
    if (savedSelection.size > 0) {
      // Filter to only include files that still exist
      const validFiles = new Set([...savedSelection].filter(f => filesAtRef.value.includes(f)))
      if (validFiles.size > 0) {
        selectedFiles.value = validFiles
        uiStore.addToast(`${t('git.restoredSelection')}: ${validFiles.size}`, 'info')
      }
    }
  } catch (error) {
    console.error('Failed to load files at ref:', error)
    uiStore.addToast(t('toast.refreshError'), 'error')
    filesAtRef.value = []
  } finally {
    isLoading.value = false
  }
}

function clearSelectedRef() {
  selectedRef.value = null
  filesAtRef.value = []
  selectedFiles.value.clear()
}

function toggleFileSelection(file: string) {
  if (selectedFiles.value.has(file)) {
    selectedFiles.value.delete(file)
  } else {
    selectedFiles.value.add(file)
  }
  selectedFiles.value = new Set(selectedFiles.value)
}

function selectFolderFiles(files: string[]) {
  // Toggle: if all selected, deselect all; otherwise select all
  const allSelected = files.every(f => selectedFiles.value.has(f))
  if (allSelected) {
    files.forEach(f => selectedFiles.value.delete(f))
  } else {
    files.forEach(f => selectedFiles.value.add(f))
  }
  selectedFiles.value = new Set(selectedFiles.value)
}

function selectAllFiles() {
  selectedFiles.value = new Set(filesAtRef.value)
}

function clearFileSelection() {
  selectedFiles.value = new Set()
}

async function buildContextFromRef() {
  if (!selectedRef.value || selectedFiles.value.size === 0) return
  
  isBuilding.value = true
  
  try {
    const files = Array.from(selectedFiles.value)
    const content = await apiService.buildContextAtRef(projectPath.value, files, selectedRef.value)
    
    // Set context in store
    contextStore.setRawContext(content, files.length)
    
    uiStore.addToast(`Context built from ${selectedRef.value.slice(0, 7)}: ${files.length} files`, 'success')
  } catch (error) {
    console.error('Failed to build context:', error)
    uiStore.addToast('Failed to build context from ref', 'error')
  } finally {
    isBuilding.value = false
  }
}

// ============================================
// Remote Repository Functions (GitHub API)
// ============================================

async function loadRemoteRepo() {
  if (!remoteUrl.value) return

  isLoadingRemote.value = true
  remoteRepoLoaded.value = false
  isGitHubRepo.value = false
  isGitLabRepo.value = false

  try {
    // Check if it's a GitHub URL
    isGitHubRepo.value = await apiService.isGitHubURL(remoteUrl.value)
    
    if (isGitHubRepo.value) {
      // Use GitHub API (fast, no clone)
      const branches = await apiService.gitHubGetBranches(remoteUrl.value)
      remoteBranches.value = branches

      // Get default branch
      const defaultBranch = await apiService.gitHubGetDefaultBranch(remoteUrl.value)
      remoteSelectedBranch.value = defaultBranch

      // Load files for default branch
      await loadRemoteFiles()

      remoteRepoLoaded.value = true
      
      // Save to recent repos
      const repoName = remoteUrl.value.split('/').slice(-2).join('/')
      saveRecentRepo(remoteUrl.value, repoName, true)
      
      uiStore.addToast('Repository loaded via GitHub API', 'success')
      return
    }
    
    // Check if it's a GitLab URL
    isGitLabRepo.value = await apiService.isGitLabURL(remoteUrl.value)
    
    if (isGitLabRepo.value) {
      // Use GitLab API (fast, no clone)
      const branches = await apiService.gitLabGetBranches(remoteUrl.value)
      // Convert GitLab branch format to common format
      remoteBranches.value = branches.map(b => ({
        name: b.name,
        commit: { sha: b.commit.id }
      }))

      // Get default branch
      const defaultBranch = await apiService.gitLabGetDefaultBranch(remoteUrl.value)
      remoteSelectedBranch.value = defaultBranch

      // Load files for default branch
      await loadRemoteFiles()

      remoteRepoLoaded.value = true
      
      // Save to recent repos
      const repoName = remoteUrl.value.split('/').slice(-2).join('/')
      saveRecentRepo(remoteUrl.value, repoName, false)
      
      uiStore.addToast('Repository loaded via GitLab API', 'success')
      return
    }
    
    // Neither GitHub nor GitLab
    uiStore.addToast('Non-GitHub/GitLab URL. Use Clone option.', 'warning')
  } catch (error) {
    console.error('Failed to load remote repo:', error)
    uiStore.addToast('Failed to load repository', 'error')
  } finally {
    isLoadingRemote.value = false
  }
}

async function loadRemoteFiles() {
  if (!remoteUrl.value || !remoteSelectedBranch.value) return

  isLoading.value = true
  loadingMessage.value = `Loading files from ${remoteSelectedBranch.value}...`
  remoteFiles.value = []
  remoteSelectedFiles.value = new Set()

  try {
    let files: string[] = []
    
    if (isGitHubRepo.value) {
      files = await apiService.gitHubListFiles(remoteUrl.value, remoteSelectedBranch.value)
    } else if (isGitLabRepo.value) {
      files = await apiService.gitLabListFiles(remoteUrl.value, remoteSelectedBranch.value)
    }
    
    remoteFiles.value = files || []
    console.log(`[Git] Loaded ${remoteFiles.value.length} remote files`)
    
    // Restore previous selection if available
    const key = `remote-${remoteUrl.value}-${remoteSelectedBranch.value}`
    const savedSelection = loadSelectionFromStorage(key)
    if (savedSelection.size > 0) {
      const validFiles = new Set([...savedSelection].filter(f => remoteFiles.value.includes(f)))
      if (validFiles.size > 0) {
        remoteSelectedFiles.value = validFiles
        uiStore.addToast(`${t('git.restoredSelection')}: ${validFiles.size}`, 'info')
      }
    }
  } catch (error) {
    console.error('Failed to load remote files:', error)
    uiStore.addToast(t('toast.refreshError'), 'error')
    remoteFiles.value = []
  } finally {
    isLoading.value = false
  }
}

function toggleRemoteFileSelection(file: string) {
  if (remoteSelectedFiles.value.has(file)) {
    remoteSelectedFiles.value.delete(file)
  } else {
    remoteSelectedFiles.value.add(file)
  }
  remoteSelectedFiles.value = new Set(remoteSelectedFiles.value)
}

function selectRemoteFolderFiles(files: string[]) {
  const allSelected = files.every(f => remoteSelectedFiles.value.has(f))
  if (allSelected) {
    files.forEach(f => remoteSelectedFiles.value.delete(f))
  } else {
    files.forEach(f => remoteSelectedFiles.value.add(f))
  }
  remoteSelectedFiles.value = new Set(remoteSelectedFiles.value)
}

function selectAllRemoteFiles() {
  remoteSelectedFiles.value = new Set(remoteFiles.value)
}

function clearRemoteFileSelection() {
  remoteSelectedFiles.value = new Set()
}

async function buildContextFromRemote() {
  if (!remoteUrl.value || remoteSelectedFiles.value.size === 0) return

  isBuilding.value = true

  try {
    const files = Array.from(remoteSelectedFiles.value)
    let content = ''
    let source = ''
    
    if (isGitHubRepo.value) {
      content = await apiService.gitHubBuildContext(remoteUrl.value, files, remoteSelectedBranch.value)
      source = 'GitHub'
    } else if (isGitLabRepo.value) {
      content = await apiService.gitLabBuildContext(remoteUrl.value, files, remoteSelectedBranch.value)
      source = 'GitLab'
    }

    contextStore.setRawContext(content, files.length)
    uiStore.addToast(`Context built from ${source}: ${files.length} files`, 'success')
  } catch (error) {
    console.error('Failed to build context from remote:', error)
    uiStore.addToast('Failed to build context', 'error')
  } finally {
    isBuilding.value = false
  }
}

async function cloneRemote() {
  if (!remoteUrl.value) return

  isCloning.value = true
  isLoading.value = true
  loadingMessage.value = 'Cloning repository...'

  try {
    clonedPath.value = await apiService.cloneRepository(remoteUrl.value)
    uiStore.addToast('Repository cloned successfully', 'success')
  } catch (error) {
    console.error('Failed to clone:', error)
    uiStore.addToast('Failed to clone repository', 'error')
  } finally {
    isCloning.value = false
    isLoading.value = false
  }
}

async function openClonedRepo() {
  if (!clonedPath.value) return
  await projectStore.openProjectByPath(clonedPath.value)
  sourceType.value = 'local'
}

async function cleanupClonedRepo() {
  if (!clonedPath.value) return

  try {
    await apiService.cleanupTempRepository(clonedPath.value)
    clonedPath.value = null
    uiStore.addToast('Temporary repository removed', 'success')
  } catch (error) {
    console.error('Failed to cleanup:', error)
  }
}

function formatDate(dateStr: string): string {
  try {
    const date = new Date(dateStr)
    return date.toLocaleDateString()
  } catch {
    return dateStr
  }
}

watch(() => projectPath.value, () => {
  checkGitRepo()
}, { immediate: true })

// Auto-save local selection
watch(selectedFiles, (newVal) => {
  if (selectedRef.value && newVal.size > 0) {
    const key = `local-${projectPath.value}-${selectedRef.value}`
    saveSelectionToStorage(key, newVal)
  }
}, { deep: true })

// Auto-save remote selection
watch(remoteSelectedFiles, (newVal) => {
  if (remoteUrl.value && remoteSelectedBranch.value && newVal.size > 0) {
    const key = `remote-${remoteUrl.value}-${remoteSelectedBranch.value}`
    saveSelectionToStorage(key, newVal)
  }
}, { deep: true })

// Diff modal state
const diffModalOpen = ref(false)

function openDiffModal() {
  diffModalOpen.value = true
}

function closeDiffModal() {
  diffModalOpen.value = false
}

// File preview state
const previewOpen = ref(false)
const previewPath = ref('')
const previewContent = ref('')
const previewLoading = ref(false)
const previewError = ref('')

async function handlePreviewFile(filePath: string) {
  previewPath.value = filePath
  previewContent.value = ''
  previewError.value = ''
  previewLoading.value = true
  previewOpen.value = true

  try {
    let content = ''
    
    if (sourceType.value === 'local' && selectedRef.value) {
      content = await apiService.getFileAtRef(projectPath.value, filePath, selectedRef.value)
    } else if (sourceType.value === 'remote') {
      if (isGitHubRepo.value) {
        content = await apiService.gitHubGetFileContent(remoteUrl.value, filePath, remoteSelectedBranch.value)
      } else if (isGitLabRepo.value) {
        content = await apiService.gitLabGetFileContent(remoteUrl.value, filePath, remoteSelectedBranch.value)
      }
    }
    
    previewContent.value = content
  } catch (error) {
    console.error('Failed to load file preview:', error)
    previewError.value = t('error.loadFailed')
  } finally {
    previewLoading.value = false
  }
}

function closePreview() {
  previewOpen.value = false
}

onMounted(() => {
  checkGitRepo()
  loadRecentReposFromStorage()
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>
