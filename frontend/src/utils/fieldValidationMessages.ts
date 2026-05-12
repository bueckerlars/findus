import type { ComposerTranslation } from "vue-i18n";
import type { ParseFieldsIssue, ValidateFieldsIssue } from "../types/templateFields";

export function formatParseFieldsIssue(issue: ParseFieldsIssue, t: ComposerTranslation): string {
  if (issue.kind === "invalid_json") return t("validation.fields.invalidJson");
  return t("validation.fields.notArray");
}

export function formatValidateFieldsIssue(issue: ValidateFieldsIssue, t: ComposerTranslation): string {
  switch (issue.kind) {
    case "key_required":
      return t("validation.fields.keyRequired", { n: issue.index + 1 });
    case "duplicate_key":
      return t("validation.fields.duplicateKey", { key: issue.key });
    case "label_required":
      return t("validation.fields.labelRequired", { key: issue.key });
    case "widget_invalid":
      return t("validation.fields.widgetInvalid", { key: issue.key });
    case "select_no_options":
      return t("validation.fields.selectNoOptions", { key: issue.key });
    case "option_value_label":
      return t("validation.fields.optionValueLabel", { key: issue.key, j: issue.optionIndex + 1 });
    case "duplicate_option_value":
      return t("validation.fields.duplicateOptionValue", { key: issue.key, value: issue.value });
    case "invalid_regex":
      return t("validation.fields.invalidRegex", { key: issue.key });
    case "min_max_int":
      return t("validation.fields.minMaxInt", { key: issue.key });
  }
}
