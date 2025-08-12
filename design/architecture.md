# Shotgun App Architecture (Revised)

## 1. Overview

The Shotgun App is a desktop application built with Wails (v2) and Vue.js (v3) using Pinia for state management. It is designed as a developer assistant to streamline the process of preparing project context for Large Language Models (LLMs) and applying the resulting changes.

The architecture is designed following the principles of **Clean Architecture**, ensuring a clear **Separation of Concerns (SoC)**, **high cohesion**, and **low coupling**. This makes the application testable, scalable, and maintainable.

-   **Backend (Go):** Structured into three distinct layers: Domain, Application, and Infrastructure. It handles all business logic, file system operations, and external service integrations (like Git).
-   **Frontend (Vue.js + Pinia):** A reactive SPA providing the user interface. State management is decentralized into focused, single-responsibility stores.
-   **Communication:** Wails provides the bridge for invoking Go methods from JavaScript and for the backend to emit events to the frontend.

## 2. Backend (Go) - Clean Architecture

The backend is strictly divided into three layers to isolate business logic from external dependencies.

### 2.1. Domain Layer (`/backend/domain`)

This is the core of the application. It contains the enterprise logic and is completely independent of any framework or external library.

-   **`models.go`**: Defines core entities like `FileNode` and `Commit`. These are pure data structures.
-   **`interfaces.go`**: Defines the contracts (Go interfaces) for all external dependencies required by the application layer. This is the key to the **Dependency Inversion Principle (DIP)**. Key interfaces include:
    -   `Logger`: For abstract logging.
    -   `EventBus`: For emitting events to the outside world (i.e., the frontend).
    -   `GitRepository`: For all Git-related operations.
    -   `SettingsRepository`: For managing user settings.
    -   `FileTreeBuilder`: For scanning the file system and building a project tree.
    -   `FileSystemWatcher`: For observing real-time changes in the file system.

### 2.2. Application Layer (`/backend/application`)

This layer contains the application-specific business rules. It orchestrates the domain entities and interfaces to perform use cases.

-   **`project_service.go`**: The primary service of this layer.
    -   It depends **only on the interfaces** defined in the domain layer, which are injected via its constructor.
    -   It contains the logic for use cases like:
        -   Generating the "shotgun" context (asynchronously, with progress reporting).
        -   Listing project files by coordinating the `FileTreeBuilder`.
        -   Handling Git operations by calling the `GitRepository`.
        -   Splitting diffs.
    -   This layer is **fully testable** in isolation, without needing the Wails runtime.

### 2.3. Infrastructure Layer (`/backend/infrastructure`)

This layer contains the concrete implementations of the domain interfaces. It's the "dirty" layer that deals with external frameworks, databases, and APIs.

-   **`wailsbridge/`**: Implements the `domain.Logger` and `domain.EventBus` interfaces using the Wails runtime functions (`runtime.Log...`, `runtime.EventsEmit`).
-   **`git/`**: Implements the `domain.GitRepository` interface by executing `git` commands on the command line (`os/exec`).
-   **`settingsfs/`**: Implements `domain.SettingsRepository` by reading/writing a JSON configuration file to the user's disk (using `xdg` library).
-   **`fsscanner/`**: Implements the `domain.FileTreeBuilder` interface by scanning the file system.
-   **`fswatcher/`**: Implements the `domain.FileSystemWatcher` interface using the efficient, event-based `fsnotify` library.

### 2.4. Composition Root (`app.go`, `main.go`)

-   **`main.go`**: The application entry point. It initializes Wails and the main `App` struct.
-   **`app.go`**: Acts as the **Composition Root**.
    -   Its primary responsibility is to instantiate the application during the `startup` lifecycle hook.
    -   It creates concrete instances from the **Infrastructure** layer and injects them into the constructor of the **Application** layer's `ProjectService`.
    -   It exposes the public methods of the `ProjectService` to the Wails runtime, making them callable from the frontend. It acts as a thin adapter, delegating all work to the application service.

## 3. Frontend (Vue.js) - Decoupled State Management

The frontend uses Vue 3 (Composition API), Vite, Tailwind CSS, and Pinia. The key architectural decision is the decomposition of a monolithic state store into smaller, focused stores.

### 3.1. Pinia Stores (`/frontend/src/stores`)

-   **`projectStore.ts` (now `projectStore.ts`):** Manages only the top-level state, such as the `projectRoot` path. It orchestrates the selection of a new project, triggering resets in other stores.
-   **`fileTreeStore.js`**: Manages the state of the file tree (`tree`), loading status, and user interactions (expanding/collapsing nodes, toggling exclusions).
-   **`context.store.ts`**: Manages the generated "shotgun" context (`shotgunPromptContext`), its generation status (`isGeneratingContext`), and progress.
-   **`gitStore.ts`**: Manages all Git-related state and operations (`isAvailable`, `isLoading`, `recentCommits`, `commitHistoryModalVisible`).
-   **`promptStore.js`**: Manages the user's task input, the final composed prompt, and template selection. It uses a debounced ref for user input to optimize performance.
-   **`diffStore.js`**: Manages the logic for the "Review & Apply" step, including the input diff text and the resulting split diffs.
-   **`settingsStore.js`**: Manages user-configurable settings (`useGitignore`, `useCustomIgnore`, rules text).
-   **`notificationsStore.ts`**: A global store for logs and notifications displayed in the console.
-   **`stepsStore.js`**: Manages the state of the multi-step UI.

### 3.2. Composables (`/frontend/src/composables`)

-   **`useClipboard.js`**: A self-contained utility for copying text to the clipboard. It reliably handles cross-platform differences by checking for the presence of the Wails runtime (`window.wails.runtime`) and falling back to the browser's `navigator.clipboard` API if it's not available.
-   **`useDebounce.js`**: A simple composable for debouncing reactive data, used to prevent excessive re-computation in the `promptStore`.

### 3.3. App Coordinator (`/frontend/src/stores/appCoordinator.js`)

This file is the "frontend composition root." It doesn't hold state but sets up `watch` effects to react to changes across different stores, ensuring they work together cohesively. For example, it watches for changes in prompt inputs and triggers the regeneration of the final prompt.

## 4. User Experience (UX) Flow

The application guides the user through a refined, three-step process:

1.  **Step 1: Prepare Context**: The user selects a project folder. A file tree is displayed, and they can exclude files/folders. Git integration allows selecting files from recent or uncommitted changes. The user then triggers the generation of the project context, which is an asynchronous operation with a progress bar.
2.  **Step 2: Compose Prompt**: The user writes a task for the LLM. The application combines this task, custom rules, and the context from Step 1 into a final, optimized prompt ready to be copied.
3.  **Step 3: Review & Apply**: The user pastes the `git diff` output received from the LLM. The application splits the diff into manageable parts, which can be easily copied and applied using standard developer tools.

This revised architecture provides a robust, testable, and maintainable foundation for the Shotgun App, enabling future feature development with confidence.