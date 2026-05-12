<script setup lang="ts">
import { ref, watch } from "vue";
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { api, postJson } from "../api";
import { toast } from "../composables/useToast";
import FxModal from "./FxModal.vue";

const { t } = useI18n();

type ItemTemplate = { ID: string; DisplayName: string };

const props = defineProps<{
  modelValue: boolean;
}>();

const emit = defineEmits<{ "update:modelValue": [boolean] }>();

const router = useRouter();

const name = ref("");
const color = ref("#6366f1");
const defaultTemplateType = ref("");
const templates = ref<ItemTemplate[]>([]);
const err = ref("");
const busy = ref(false);

async function load() {
  err.value = "";
  if (!props.modelValue) return;
  const r = await api<{ templates: ItemTemplate[] }>("/api/labels/new");
  templates.value = r.templates;
  defaultTemplateType.value = r.templates[0]?.ID || "";
  name.value = "";
  color.value = "#6366f1";
}

watch(
  () => props.modelValue,
  (open) => {
    if (open) void load();
  },
);

function close() {
  emit("update:modelValue", false);
}

async function save() {
  err.value = "";
  busy.value = true;
  try {
    const r = await postJson<{ next: string }>("/api/labels", {
      name: name.value,
      color: color.value,
      default_template_type: defaultTemplateType.value,
    });
    toast.success(t("toast.labelCreated"));
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
    :title="t('createLabel.title')"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <p v-if="err" class="mb-4 rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">{{ err }}</p>
    <form id="fx-create-label-form" class="space-y-4" @submit.prevent="save">
      <div>
        <label class="fx-label" for="fx-create-label-name">{{ $t("labelForm.name") }}</label>
        <input id="fx-create-label-name" v-model="name" class="fx-input" required />
      </div>
      <div>
        <label class="fx-label" for="fx-create-label-color">{{ $t("labelForm.color") }}</label>
        <input id="fx-create-label-color" v-model="color" type="color" class="h-10 w-20 cursor-pointer rounded border border-zinc-200" />
      </div>
      <div>
        <label class="fx-label" for="fx-create-label-tpl">{{ $t("labelForm.defaultTemplate") }}</label>
        <select id="fx-create-label-tpl" v-model="defaultTemplateType" class="fx-input">
          <option value="">{{ $t("common.noneChoice") }}</option>
          <option v-for="tpl in templates" :key="tpl.ID" :value="tpl.ID">{{ tpl.DisplayName || tpl.ID }}</option>
        </select>
      </div>
    </form>

    <template #footer>
      <div class="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
        <button type="button" class="fx-btn-secondary w-full sm:w-auto" :disabled="busy" @click="close">{{ $t("common.cancel") }}</button>
        <button type="submit" form="fx-create-label-form" class="fx-btn-primary w-full sm:w-auto" :disabled="busy">
          {{ $t("common.save") }}
        </button>
      </div>
    </template>
  </FxModal>
</template>
