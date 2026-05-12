<script setup lang="ts">
import { ref, watch } from "vue";
import { useRouter } from "vue-router";
import { api, postJson } from "../api";
import { toast } from "../composables/useToast";
import FxModal from "./FxModal.vue";

type LocOpt = { ID: string; Label: string };

const props = defineProps<{
  modelValue: boolean;
  initialParentId?: string;
}>();

const emit = defineEmits<{ "update:modelValue": [boolean] }>();

const router = useRouter();

const name = ref("");
const description = ref("");
const parentId = ref("");
const parentOptions = ref<LocOpt[]>([]);
const err = ref("");
const busy = ref(false);

async function load() {
  err.value = "";
  if (!props.modelValue) return;
  const pid = props.initialParentId?.trim() || "";
  const r = await api<{ parent_options: LocOpt[]; selected_parent: string }>(
    "/api/locations/new" + (pid ? "?parent_id=" + encodeURIComponent(pid) : ""),
  );
  parentOptions.value = r.parent_options;
  parentId.value = r.selected_parent || pid || "";
  name.value = "";
  description.value = "";
}

watch(
  () => props.modelValue,
  (open) => {
    if (open) void load();
  },
);

watch(
  () => props.initialParentId,
  () => {
    if (props.modelValue) void load();
  },
);

function close() {
  emit("update:modelValue", false);
}

async function save() {
  err.value = "";
  busy.value = true;
  try {
    const r = await postJson<{ next: string }>("/api/locations", {
      name: name.value,
      description: description.value,
      parent_id: parentId.value,
    });
    toast.success("Location created.");
    close();
    await router.push(r.next);
  } catch (e) {
    err.value = e instanceof Error ? e.message : "Save failed";
  } finally {
    busy.value = false;
  }
}
</script>

<template>
  <FxModal
    :model-value="modelValue"
    title="New location"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <p v-if="err" class="mb-4 rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">{{ err }}</p>
    <form id="fx-create-location-form" class="space-y-4" @submit.prevent="save">
      <div>
        <label class="fx-label" for="fx-create-loc-name">Name</label>
        <input id="fx-create-loc-name" v-model="name" class="fx-input" required />
      </div>
      <div>
        <label class="fx-label" for="fx-create-loc-desc">Description</label>
        <textarea id="fx-create-loc-desc" v-model="description" class="fx-input min-h-[100px]" />
      </div>
      <div>
        <label class="fx-label" for="fx-create-loc-parent">Parent</label>
        <select id="fx-create-loc-parent" v-model="parentId" class="fx-input">
          <option v-for="o in parentOptions" :key="o.ID || 'root'" :value="o.ID">{{ o.Label }}</option>
        </select>
      </div>
    </form>

    <template #footer>
      <div class="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
        <button type="button" class="fx-btn-secondary w-full sm:w-auto" :disabled="busy" @click="close">Cancel</button>
        <button type="submit" form="fx-create-location-form" class="fx-btn-primary w-full sm:w-auto" :disabled="busy">
          Save
        </button>
      </div>
    </template>
  </FxModal>
</template>
