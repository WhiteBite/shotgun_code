<template>
  <div class='bg-gray-800/50 p-6 rounded-lg border border-gray-700'>
    <h2 class='text-xl font-semibold mb-4 text-white'>{{ t('projects.title') }}</h2>
    <div class='space-y-3'>
      <button @click='onOpen' class='w-full flex items-center justify-center gap-2 p-3 bg-blue-600/80 hover:bg-blue-600 rounded-md text-white font-semibold'>
        {{ t('projects.open') }}
      </button>
      <div v-if='projectStore.recentProjects.length > 0' class='space-y-2'>
        <div v-for='proj in projectStore.recentProjects' :key='proj.path' class='p-3 bg-gray-700/50 hover:bg-gray-700/80 rounded-md flex items-center justify-between'>
          <div class="min-w-0">
            <p class='font-semibold text-white truncate'>{{ proj.name }}</p>
            <p class='text-xs text-gray-400 truncate'>{{ proj.path }}</p>
          </div>
          <div class="flex items-center gap-2 flex-shrink-0">
            <button @click.stop='onSelect(proj)' class='px-2 py-1 text-xs bg-blue-600/70 hover:bg-blue-600 rounded text-white'>Open</button>
            <button @click.stop='onRemove(proj.path)' class='px-2 py-1 text-xs bg-gray-600/70 hover:bg-gray-600 rounded text-white'>Удалить из списка</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang='ts'>
import { onMounted } from 'vue';
import { useProjectStore } from '@/stores/project.store';
import { useRouter } from 'vue-router';
import { t } from '@/lib/i18n';

const router = useRouter();
const projectStore = useProjectStore();

onMounted(() => {
  projectStore.loadRecentProjects();
});

async function onOpen() {
  const ok = await projectStore.openProject();
  if (ok) router.push({ name: "workspace" });
}

async function onSelect(proj: {name:string; path:string}) {
  const ok = await projectStore.setCurrentProject(proj);
  if (ok) router.push({ name: "workspace" });
}

function onRemove(path: string) {
  projectStore.removeRecent(path);
}
</script>