<script setup lang="ts">
import { ref, computed, onMounted, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { api, postJson } from "../api";
import { toast } from "../composables/useToast";
import FxAlert from "../components/primitives/FxAlert.vue";
import FxButton from "../components/primitives/FxButton.vue";
import FxPageHeader from "../components/primitives/FxPageHeader.vue";

const { t } = useI18n();

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
    toast.success(t("toast.locationSaved"));
    await router.push(r.next);
  } catch (e) {
    err.value = e instanceof Error ? e.message : t("common.saveFailed");
  }
}
</script>

<template>
  <div class="max-w-lg space-y-5">
    <FxPageHeader :title="$t('locationForm.titleEdit')" />
    <FxAlert v-if="err">{{ err }}</FxAlert>
    <form class="space-y-4" @submit.prevent="save">
      <div>
        <label class="fx-label" for="ln">{{ $t("labelForm.name") }}</label>
        <input id="ln" v-model="name" class="fx-input" required />
      </div>
      <div>
        <label class="fx-label" for="ld">{{ $t("itemDetail.description") }}</label>
        <textarea id="ld" v-model="description" class="fx-input min-h-[100px]" />
      </div>
      <div>
        <label class="fx-label" for="lp">{{ $t("common.parent") }}</label>
        <select id="lp" v-model="parentId" class="fx-input">
          <option v-for="o in parentOptions" :key="o.ID || 'root'" :value="o.ID">{{ o.Label }}</option>
        </select>
      </div>
      <FxButton type="submit" icon-left="check">{{ $t("common.save") }}</FxButton>
    </form>
  </div>
</template>
