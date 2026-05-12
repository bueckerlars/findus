<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { api, postJson } from "../api";
import { toast } from "../composables/useToast";

type LocOpt = { ID: string; Label: string };
type Location = { ID: string; Name: string; Description: string };

const route = useRoute();
const router = useRouter();
const id = computed(() => route.params.id as string);

const name = ref("");
const description = ref("");
const parentId = ref("");
const parentOptions = ref<LocOpt[]>([]);
const err = ref("");

async function load() {
  err.value = "";
  const r = await api<{ location: Location; parent_options: LocOpt[]; selected_parent: string }>("/api/locations/" + id.value + "/edit");
  name.value = r.location.Name;
  description.value = r.location.Description;
  parentOptions.value = r.parent_options;
  parentId.value = r.selected_parent;
}

onMounted(load);
watch(id, () => {
  void load();
});

async function save() {
  err.value = "";
  try {
    const r = await postJson<{ next: string }>("/api/locations/" + id.value, {
      name: name.value,
      description: description.value,
      parent_id: parentId.value,
    });
    toast.success("Location saved.");
    await router.push(r.next);
  } catch (e) {
    err.value = e instanceof Error ? e.message : "Save failed";
  }
}
</script>

<template>
  <div class="max-w-lg space-y-6">
    <h1 class="text-2xl font-semibold text-zinc-900">Edit location</h1>
    <p v-if="err" class="rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">{{ err }}</p>
    <form class="space-y-4" @submit.prevent="save">
      <div>
        <label class="fx-label" for="ln">Name</label>
        <input id="ln" v-model="name" class="fx-input" required />
      </div>
      <div>
        <label class="fx-label" for="ld">Description</label>
        <textarea id="ld" v-model="description" class="fx-input min-h-[100px]" />
      </div>
      <div>
        <label class="fx-label" for="lp">Parent</label>
        <select id="lp" v-model="parentId" class="fx-input">
          <option v-for="o in parentOptions" :key="o.ID || 'root'" :value="o.ID">{{ o.Label }}</option>
        </select>
      </div>
      <button type="submit" class="fx-btn-primary">Save</button>
    </form>
  </div>
</template>
