/** Mirrors backend domain.TemplateField / FieldOption JSON keys. */

export type TemplateWidget = "text" | "select";

export type FieldOption = {
  value: string;
  label: string;
};

export type TemplateField = {
  key: string;
  label: string;
  widget: TemplateWidget;
  required: boolean;
  placeholder?: string;
  pattern?: string;
  min_int?: number;
  max_int?: number;
  max_len?: number;
  options?: FieldOption[];
};

export function emptyTemplateField(): TemplateField {
  return {
    key: "",
    label: "",
    widget: "text",
    required: false,
    placeholder: "",
    pattern: "",
    options: [],
  };
}

function isRecord(x: unknown): x is Record<string, unknown> {
  return typeof x === "object" && x !== null && !Array.isArray(x);
}

function asWidget(x: unknown): TemplateWidget {
  return x === "select" ? "select" : "text";
}

function normalizeOption(x: unknown): FieldOption {
  if (!isRecord(x)) return { value: "", label: "" };
  return {
    value: typeof x.value === "string" ? x.value : "",
    label: typeof x.label === "string" ? x.label : "",
  };
}

export type ParseFieldsIssue = { kind: "invalid_json" } | { kind: "not_array" };

export type ValidateFieldsIssue =
  | { kind: "key_required"; index: number }
  | { kind: "duplicate_key"; key: string }
  | { kind: "label_required"; key: string }
  | { kind: "widget_invalid"; key: string }
  | { kind: "select_no_options"; key: string }
  | { kind: "option_value_label"; key: string; optionIndex: number }
  | { kind: "duplicate_option_value"; key: string; value: string }
  | { kind: "invalid_regex"; key: string }
  | { kind: "min_max_int"; key: string };

/** Parse API `fields_json` into editable field rows (lenient import). */
export function parseTemplateFieldsJson(
  raw: string,
): { ok: true; fields: TemplateField[] } | { ok: false; issue: ParseFieldsIssue } {
  const s = raw.trim();
  if (s === "") {
    return { ok: true, fields: [] };
  }
  let data: unknown;
  try {
    data = JSON.parse(s) as unknown;
  } catch {
    return { ok: false, issue: { kind: "invalid_json" } };
  }
  if (!Array.isArray(data)) {
    return { ok: false, issue: { kind: "not_array" } };
  }
  const fields: TemplateField[] = [];
  for (const el of data) {
    if (!isRecord(el)) {
      fields.push(emptyTemplateField());
      continue;
    }
    const widget = asWidget(el.widget);
    const optsRaw = el.options;
    const options = Array.isArray(optsRaw) ? optsRaw.map(normalizeOption) : [];
    let minInt: number | undefined;
    let maxInt: number | undefined;
    if (typeof el.min_int === "number" && Number.isFinite(el.min_int)) minInt = Math.trunc(el.min_int);
    if (typeof el.max_int === "number" && Number.isFinite(el.max_int)) maxInt = Math.trunc(el.max_int);
    let maxLen: number | undefined;
    if (typeof el.max_len === "number" && Number.isFinite(el.max_len) && el.max_len > 0) maxLen = Math.trunc(el.max_len);
    fields.push({
      key: typeof el.key === "string" ? el.key : "",
      label: typeof el.label === "string" ? el.label : "",
      widget,
      required: Boolean(el.required),
      placeholder: typeof el.placeholder === "string" ? el.placeholder : "",
      pattern: typeof el.pattern === "string" ? el.pattern : "",
      min_int: minInt,
      max_int: maxInt,
      max_len: maxLen,
      options: widget === "select" ? (options.length ? options : [{ value: "", label: "" }]) : [],
    });
  }
  return { ok: true, fields };
}

/** Build compact `fields_json` for the API (omits empty optional keys). */
export function serializeTemplateFields(fields: TemplateField[]): string {
  const wire = fields.map((f) => {
    const key = f.key.trim();
    const label = f.label.trim();
    const base: Record<string, unknown> = {
      key,
      label,
      widget: f.widget,
      required: Boolean(f.required),
    };
    const ph = (f.placeholder ?? "").trim();
    if (ph !== "") base.placeholder = ph;
    if (f.widget === "text") {
      const pat = (f.pattern ?? "").trim();
      if (pat !== "") base.pattern = pat;
      if (f.min_int !== undefined && Number.isFinite(f.min_int)) base.min_int = Math.trunc(f.min_int);
      if (f.max_int !== undefined && Number.isFinite(f.max_int)) base.max_int = Math.trunc(f.max_int);
      if (f.max_len !== undefined && f.max_len > 0) base.max_len = Math.trunc(f.max_len);
    } else {
      base.options = (f.options ?? []).map((o) => ({
        value: o.value.trim(),
        label: o.label.trim(),
      }));
    }
    return base;
  });
  return JSON.stringify(wire);
}

/** Returns structured validation issue or null if OK (aligned with backend rules). */
export function validateTemplateFields(fields: TemplateField[]): ValidateFieldsIssue | null {
  const seen = new Set<string>();
  for (let i = 0; i < fields.length; i++) {
    const f = fields[i];
    const key = f.key.trim();
    if (key === "") {
      return { kind: "key_required", index: i };
    }
    if (seen.has(key)) {
      return { kind: "duplicate_key", key };
    }
    seen.add(key);
    if (f.label.trim() === "") {
      return { kind: "label_required", key };
    }
    if (f.widget !== "text" && f.widget !== "select") {
      return { kind: "widget_invalid", key };
    }
    if (f.widget === "select") {
      const opts = f.options ?? [];
      if (opts.length === 0) {
        return { kind: "select_no_options", key };
      }
      const values = new Set<string>();
      for (let j = 0; j < opts.length; j++) {
        const o = opts[j];
        const v = o.value.trim();
        const lb = o.label.trim();
        if (v === "" || lb === "") {
          return { kind: "option_value_label", key, optionIndex: j };
        }
        if (values.has(v)) {
          return { kind: "duplicate_option_value", key, value: v };
        }
        values.add(v);
      }
    }
    const pat = (f.pattern ?? "").trim();
    if (pat !== "") {
      try {
        new RegExp(pat);
      } catch {
        return { kind: "invalid_regex", key };
      }
    }
    if (f.widget === "text" && f.min_int !== undefined && f.max_int !== undefined && f.min_int > f.max_int) {
      return { kind: "min_max_int", key };
    }
  }
  return null;
}
