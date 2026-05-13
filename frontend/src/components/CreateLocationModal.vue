<script setup lang="ts">
import { ref, watch } from "vue";
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { api, postJson } from "../api";
import { toast } from "../composables/useToast";
import FxModal from "./FxModal.vue";
import FxAlert from "./primitives/FxAlert.vue";
import FxButton from "./primitives/FxButton.vue";

const { t } = useI18n();

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
    toast.success(t("toast.locationCreated"));
    close();
    await router.push(r.next);
  } catch (e) {
    err.value = e instanceof Error ? e.message : t("common.saveFailed");
  } finally {
    busy.value = false;
  }
}
</script>

<template>
  <FxModal
    :model-value="modelValue"
    :title="t('createLocation.title')"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <FxAlert v-if="err" class="mb-4">{{ err }}</FxAlert>
    <form id="fx-create-location-form" class="space-y-4" @submit.prevent="save">
      <div>
        <label class="fx-label" for="fx-create-loc-name">{{ $t("labelForm.name") }}</label>
        <input id="fx-create-loc-name" v-model="name" class="fx-input" required />
      </div>
      <div>
        <label class="fx-label" for="fx-create-loc-desc">{{ $t("itemDetail.description") }}</label>
        <textarea id="fx-create-loc-desc" v-model="description" class="fx-input min-h-[100px]" />
      </div>
      <div>
        <label class="fx-label" for="fx-create-loc-parent">{{ $t("common.parent") }}</label>
        <select id="fx-create-loc-parent" v-model="parentId" class="fx-input">
          <option v-for="o in parentOptions" :key="o.ID || 'root'" :value="o.ID">{{ o.Label }}</option>
        </select>
      </div>
    </form>

    <template #footer>
      <div class="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
        <FxButton type="button" variant="secondary" :disabled="busy" @click="close">{{ $t("common.cancel") }}</FxButton>
        <button type="submit" form="fx-create-location-form" class="fx-btn-primary w-full sm:w-auto" :disabled="busy" :aria-busy="busy || undefined">
          {{ $t("common.save") }}
        </button>
      </div>
    </template>
  </FxModal>
</template>
