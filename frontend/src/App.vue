<script setup lang="ts">
import { RouterView } from "vue-router";
import { Toaster } from "vue-sonner";
import "vue-sonner/style.css";
import FxAlertDialogHost from "./components/FxAlertDialogHost.vue";
</script>

<template>
  <RouterView />
  <FxAlertDialogHost />
  <Toaster position="bottom-right" :duration="4000" :rich-colors="true" theme="light" />
</template>
