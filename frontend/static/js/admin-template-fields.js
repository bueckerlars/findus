/**
 * Visual item-template field editor: syncs to a hidden input as JSON (fields_json).
 */
(function () {
  function el(tag, className, text) {
    var e = document.createElement(tag);
    if (className) e.className = className;
    if (text != null) e.textContent = text;
    return e;
  }

  function fieldCard(field, index, onChange) {
    var card = el('div', 'rounded-xl border border-zinc-200 bg-white p-4 space-y-3');
    card.dataset.index = String(index);

    var head = el('div', 'flex flex-wrap items-center justify-between gap-2');
    head.appendChild(el('span', 'text-sm font-semibold text-zinc-800', 'Field ' + (index + 1)));
    var rm = el('button', 'text-sm font-medium text-red-600 hover:text-red-700', 'Remove');
    rm.type = 'button';
    rm.addEventListener('click', function () {
      card.remove();
      onChange();
    });
    head.appendChild(rm);
    card.appendChild(head);

    var grid = el('div', 'grid gap-3 sm:grid-cols-2');
    function row(label, input) {
      var w = el('div', '');
      var lb = el('label', 'fx-label', label);
      w.appendChild(lb);
      w.appendChild(input);
      return w;
    }

    var keyIn = el('input', 'fx-input font-mono text-sm');
    keyIn.value = field.key || '';
    keyIn.placeholder = 'e.g. serial_no';
    grid.appendChild(row('Key', keyIn));

    var labIn = el('input', 'fx-input');
    labIn.value = field.label || '';
    labIn.placeholder = 'Label shown in forms';
    grid.appendChild(row('Label', labIn));

    var wid = el('select', 'fx-input');
    [['text', 'Text'], ['select', 'Select']].forEach(function (p) {
      var o = document.createElement('option');
      o.value = p[0];
      o.textContent = p[1];
      wid.appendChild(o);
    });
    wid.value = field.widget === 'select' ? 'select' : 'text';
    grid.appendChild(row('Widget', wid));

    var req = el('label', 'flex cursor-pointer items-center gap-2 text-sm text-zinc-700');
    var reqCb = document.createElement('input');
    reqCb.type = 'checkbox';
    reqCb.className = 'rounded border-zinc-300 text-sky-600';
    reqCb.checked = !!field.required;
    req.appendChild(reqCb);
    req.appendChild(document.createTextNode(' Required'));
    grid.appendChild(row('Validation', req));

    var phWrap = el('div', 'sm:col-span-2');
    var phLb = el('label', 'fx-label', 'Placeholder (text only)');
    var phIn = el('input', 'fx-input');
    phIn.value = field.placeholder || '';
    phWrap.appendChild(phLb);
    phWrap.appendChild(phIn);
    grid.appendChild(phWrap);

    var adv = el('details', 'sm:col-span-2 rounded-lg border border-zinc-100 bg-zinc-50/80 p-3');
    var sum = el('summary', 'cursor-pointer text-sm font-medium text-zinc-700', 'Advanced (text fields)');
    adv.appendChild(sum);
    var advGrid = el('div', 'mt-3 grid gap-3 sm:grid-cols-2');

    var patIn = el('input', 'fx-input font-mono text-xs');
    patIn.value = field.pattern || '';
    patIn.placeholder = 'Regex, e.g. ^\\\\d{4}$';
    advGrid.appendChild(row('Pattern', patIn));

    var maxLen = el('input', 'fx-input');
    maxLen.type = 'number';
    maxLen.min = '0';
    maxLen.value = field.max_len > 0 ? String(field.max_len) : '';
    maxLen.placeholder = '0 = default max';
    advGrid.appendChild(row('Max length (chars)', maxLen));

    var minI = el('input', 'fx-input');
    minI.type = 'number';
    minI.value = field.min_int != null ? String(field.min_int) : '';
    advGrid.appendChild(row('Min int', minI));

    var maxI = el('input', 'fx-input');
    maxI.type = 'number';
    maxI.value = field.max_int != null ? String(field.max_int) : '';
    advGrid.appendChild(row('Max int', maxI));

    adv.appendChild(advGrid);
    grid.appendChild(adv);

    var optSection = el('div', 'sm:col-span-2 space-y-2 rounded-lg border border-zinc-100 bg-sky-50/40 p-3');
    var optTitle = el('div', 'text-sm font-medium text-zinc-800', 'Select options');
    optSection.appendChild(optTitle);
    var optList = el('div', 'space-y-2');
    optSection.appendChild(optList);

    function addOptionRow(val, label) {
      var row = el('div', 'flex flex-wrap gap-2 items-end');
      var v = el('input', 'fx-input flex-1 min-w-[6rem] font-mono text-xs');
      v.placeholder = 'value';
      v.value = val || '';
      var l = el('input', 'fx-input flex-1 min-w-[6rem]');
      l.placeholder = 'label';
      l.value = label || '';
      var brm = el('button', 'fx-btn-secondary text-xs shrink-0', 'Remove');
      brm.type = 'button';
      brm.addEventListener('click', function () {
        row.remove();
        onChange();
      });
      row.appendChild(v);
      row.appendChild(l);
      row.appendChild(brm);
      optList.appendChild(row);
      [v, l].forEach(function (x) {
        x.addEventListener('input', onChange);
      });
    }

    function syncOptVisibility() {
      optSection.style.display = wid.value === 'select' ? '' : 'none';
    }

    (field.options || []).forEach(function (o) {
      addOptionRow(o.value, o.label);
    });
    if (wid.value === 'select' && (!field.options || !field.options.length)) {
      addOptionRow('', '');
    }

    var addOpt = el('button', 'fx-btn-secondary text-xs', '+ Option');
    addOpt.type = 'button';
    addOpt.addEventListener('click', function () {
      addOptionRow('', '');
      onChange();
    });
    optSection.appendChild(addOpt);

    grid.appendChild(optSection);
    card.appendChild(grid);

    wid.addEventListener('change', function () {
      syncOptVisibility();
      onChange();
    });
    syncOptVisibility();

    [keyIn, labIn, wid, reqCb, phIn, patIn, maxLen, minI, maxI].forEach(function (x) {
      x.addEventListener('input', onChange);
      x.addEventListener('change', onChange);
    });

    card._collect = function () {
      var wv = wid.value === 'select' ? 'select' : 'text';
      var out = {
        key: keyIn.value.trim(),
        label: labIn.value.trim(),
        widget: wv,
        required: !!reqCb.checked,
      };
      var ph = phIn.value.trim();
      if (ph) out.placeholder = ph;
      if (wv === 'text') {
        var pat = patIn.value.trim();
        if (pat) out.pattern = pat;
        var ml = parseInt(maxLen.value, 10);
        if (!isNaN(ml) && ml > 0) out.max_len = ml;
        var mn = minI.value.trim();
        if (mn !== '') {
          var n = parseInt(mn, 10);
          if (!isNaN(n)) out.min_int = n;
        }
        var mx = maxI.value.trim();
        if (mx !== '') {
          var n2 = parseInt(mx, 10);
          if (!isNaN(n2)) out.max_int = n2;
        }
      }
      if (wv === 'select') {
        var opts = [];
        optList.querySelectorAll(':scope > div').forEach(function (r) {
          var ins = r.querySelectorAll('input');
          if (ins.length < 2) return;
          var ov = ins[0].value.trim();
          var ol = ins[1].value.trim();
          if (ov || ol) opts.push({ value: ov, label: ol || ov });
        });
        out.options = opts;
      }
      return out;
    };

    return card;
  }

  function collect(root, hidden) {
    var cards = root.querySelectorAll('[data-index]');
    var arr = [];
    cards.forEach(function (c) {
      if (typeof c._collect === 'function') arr.push(c._collect());
    });
    hidden.value = JSON.stringify(arr);
  }

  function render(root, hidden, fields, onChange) {
    root.innerHTML = '';
    fields.forEach(function (f, i) {
      root.appendChild(fieldCard(f, i, onChange));
    });
    onChange();
  }

  function init(rootId, hiddenId, initScriptId) {
    var root = document.getElementById(rootId);
    var hidden = document.getElementById(hiddenId);
    if (!root || !hidden) return;

    var raw = '[]';
    var seed = initScriptId ? document.getElementById(initScriptId) : null;
    if (seed && seed.textContent) raw = seed.textContent.trim();

    var fields;
    try {
      fields = JSON.parse(raw);
      if (!Array.isArray(fields)) fields = [];
    } catch (e) {
      fields = [];
    }

    function onChange() {
      collect(root, hidden);
    }

    render(root, hidden, fields, onChange);

    var addBtn = document.getElementById(rootId + '-add');
    if (addBtn) {
      addBtn.addEventListener('click', function () {
        root.appendChild(
          fieldCard({ key: '', label: '', widget: 'text', required: false }, root.children.length, onChange)
        );
        onChange();
      });
    }

    var form = root.closest('form');
    if (form) {
      form.addEventListener('submit', function () {
        collect(root, hidden);
      });
    }
  }

  window.TemplateFieldEditor = { init: init };
})();
