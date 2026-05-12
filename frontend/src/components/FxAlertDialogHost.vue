<script setup lang="ts">
import FxAlertDialog from "./FxAlertDialog.vue";
import { useAlertDialogHost } from "../composables/useAlertDialog";

const { visible, dialogState, onConfirm, onCancel } = useAlertDialogHost();
</script>

<template>
  <FxAlertDialog
    v-if="visible && dialogState"
    :title="dialogState.title"
    :message="dialogState.message"
    :confirm-label="dialogState.confirmLabel"
    :cancel-label="dialogState.cancelLabel"
    :variant="dialogState.variant"
    @confirm="onConfirm"
    @cancel="onCancel"
  />
</template>
