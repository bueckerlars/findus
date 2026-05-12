(function () {
  'use strict';

  var dialog = document.getElementById('fx-command-dialog');
  var input = document.getElementById('fx-command-input');
  if (!dialog || !input) return;

  var bodyEl = document.getElementById('fx-command-body');
  var resultsWrap = document.getElementById('fx-command-search-wrap');
  var resultsEl = document.getElementById('fx-command-search-results');
  var openSearchBtn = document.getElementById('fx-command-open-search');
  var openSearchQ = document.getElementById('fx-command-open-search-q');
  var triggerBtn = document.getElementById('fx-command-trigger');

  var selectedIdx = 0;
  var debounceTimer;
  var fetchAbort;
  var lastFocus;

  function isMac() {
    return /Mac|iPhone|iPad|iPod/i.test(navigator.platform || '') || (navigator.userAgentData && navigator.userAgentData.platform === 'macOS');
  }

  document.querySelectorAll('[data-fx-cmd-mod]').forEach(function (el) {
    el.textContent = isMac() ? '⌘' : 'Ctrl+';
  });

  function norm(s) {
    return (s || '').toLowerCase().trim();
  }

  function staticItems() {
    return Array.prototype.slice.call(dialog.querySelectorAll('[data-cmd-item][data-keywords]'));
  }

  function searchResultButtons() {
    return Array.prototype.slice.call(resultsEl.querySelectorAll('[data-cmd-item]'));
  }

  function visibleItems() {
    var a = staticItems().filter(function (el) {
      return !el.classList.contains('hidden');
    });
    var b = searchResultButtons();
    var out = a.concat(b);
    if (openSearchBtn && !openSearchBtn.classList.contains('hidden')) out.push(openSearchBtn);
    return out;
  }

  function syncGroupVisibility() {
    dialog.querySelectorAll('[data-fx-cmd-group]').forEach(function (g) {
      if (g.id === 'fx-command-search-wrap') return;
      var any = g.querySelector('[data-cmd-item]:not(.hidden)');
      g.classList.toggle('hidden', !any);
    });
  }

  function filterStatic(q) {
    var nq = norm(q);
    staticItems().forEach(function (el) {
      var kw = (el.getAttribute('data-keywords') || '') + ' ' + (el.textContent || '');
      if (!nq) {
        el.classList.remove('hidden');
        return;
      }
      el.classList.toggle('hidden', norm(kw).indexOf(nq) === -1);
    });
    syncGroupVisibility();
  }

  function setActive(items) {
    items.forEach(function (el, i) {
      var on = i === selectedIdx;
      el.setAttribute('aria-selected', on ? 'true' : 'false');
      el.classList.toggle('fx-command-item-active', on);
      if (on) {
        try {
          el.scrollIntoView({ block: 'nearest', behavior: 'smooth' });
        } catch (_) {
          el.scrollIntoView(false);
        }
      }
    });
  }

  function clampSelection() {
    var items = visibleItems();
    if (items.length === 0) {
      selectedIdx = 0;
      return;
    }
    if (selectedIdx >= items.length) selectedIdx = items.length - 1;
    if (selectedIdx < 0) selectedIdx = 0;
    setActive(items);
  }

  function renderSearchItems(items, q) {
    resultsEl.innerHTML = '';
    items.forEach(function (it) {
      var btn = document.createElement('button');
      btn.type = 'button';
      btn.setAttribute('data-cmd-item', '');
      btn.setAttribute('data-href', '/items/' + it.id);
      btn.className =
        'fx-command-item group flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-left text-[13px] leading-snug outline-none transition hover:bg-zinc-200/35 focus-visible:bg-zinc-200/35 focus-visible:ring-2 focus-visible:ring-zinc-400/25';
      btn.setAttribute('aria-selected', 'false');
      var type = it.type || '';
      btn.innerHTML =
        '<span class="fx-command-item-icon flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-zinc-100/90 text-zinc-500 transition group-hover:bg-white/95 group-hover:text-zinc-700 group-hover:shadow-sm group-hover:ring-1 group-hover:ring-zinc-200/70">' +
        '<svg class="fx-icon shrink-0" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" d="m21 7.5-9-5.25L3 7.5m18 0v9l-9 5.25m9-14.25v9m0-9-9 5.25M3 7.5v9m9 5.25M3 7.5l9 5.25m0 0 9-5.25"/></svg></span>' +
        '<span class="min-w-0 flex-1 font-medium text-zinc-900">' +
        escapeHtml(it.name) +
        '</span>' +
        '<span class="shrink-0 rounded px-1.5 py-0.5 text-[10px] font-medium capitalize leading-none text-zinc-600 ring-1 ring-zinc-200/60 bg-zinc-100/50">' +
        escapeHtml(type) +
        '</span>';
      resultsEl.appendChild(btn);
    });

    var showWrap = items.length > 0 || norm(q).length > 0;
    resultsWrap.classList.toggle('hidden', !showWrap);
    if (openSearchBtn && openSearchQ) {
      var hasQ = norm(q).length > 0;
      openSearchBtn.classList.toggle('hidden', !hasQ);
      openSearchBtn.setAttribute('data-href', '/search?q=' + encodeURIComponent(q));
      openSearchQ.textContent = hasQ ? ' for “' + q.trim() + '”' : '';
    }
    selectedIdx = 0;
    clampSelection();
  }

  function escapeHtml(s) {
    var d = document.createElement('div');
    d.textContent = s;
    return d.innerHTML;
  }

  function runSearch(q) {
    if (fetchAbort) fetchAbort.abort();
    var nq = norm(q);
    if (!nq) {
      renderSearchItems([], '');
      resultsWrap.classList.add('hidden');
      filterStatic('');
      clampSelection();
      return;
    }

    fetchAbort = new AbortController();
    var signal = fetchAbort.signal;
    fetch('/command-search?q=' + encodeURIComponent(q.trim()), {
      credentials: 'same-origin',
      signal: signal,
      headers: { Accept: 'application/json' },
    })
      .then(function (r) {
        if (!r.ok) throw new Error('search');
        return r.json();
      })
      .then(function (data) {
        var items = data && data.items ? data.items : [];
        filterStatic(q);
        renderSearchItems(items, q);
        clampSelection();
      })
      .catch(function (err) {
        if (err.name === 'AbortError') return;
        filterStatic(q);
        renderSearchItems([], q);
        clampSelection();
      });
  }

  function scheduleSearch() {
    var q = input.value;
    clearTimeout(debounceTimer);
    filterStatic(q);
    debounceTimer = setTimeout(function () {
      runSearch(input.value);
    }, 220);
  }

  function openPalette() {
    if (dialog.open) return;
    lastFocus = document.activeElement;
    dialog.showModal();
    if (triggerBtn) triggerBtn.setAttribute('aria-expanded', 'true');
    input.value = '';
    filterStatic('');
    renderSearchItems([], '');
    resultsWrap.classList.add('hidden');
    selectedIdx = 0;
    clampSelection();
    setTimeout(function () {
      input.focus();
      input.select();
    }, 0);
  }

  function closePalette() {
    if (!dialog.open) return;
    if (triggerBtn) triggerBtn.setAttribute('aria-expanded', 'false');
    dialog.close();
  }

  function activateSelected() {
    var items = visibleItems();
    var el = items[selectedIdx];
    if (!el) return;
    var href = el.getAttribute('data-href');
    if (href) {
      window.location.href = href;
    }
  }

  document.addEventListener(
    'keydown',
    function (e) {
      var cmdk = (e.metaKey || e.ctrlKey) && (e.key === 'k' || e.key === 'K');
      if (!cmdk) return;
      e.preventDefault();
      if (dialog.open) closePalette();
      else openPalette();
    },
    true
  );

  dialog.addEventListener('cancel', function (e) {
    e.preventDefault();
    closePalette();
  });

  dialog.addEventListener('close', function () {
    clearTimeout(debounceTimer);
    if (fetchAbort) fetchAbort.abort();
    if (triggerBtn) triggerBtn.setAttribute('aria-expanded', 'false');
    if (lastFocus && typeof lastFocus.focus === 'function') {
      try {
        lastFocus.focus();
      } catch (_) {}
    }
  });

  input.addEventListener('input', function () {
    scheduleSearch();
  });

  input.addEventListener('keydown', function (e) {
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      var n = visibleItems().length;
      if (n) selectedIdx = (selectedIdx + 1) % n;
      clampSelection();
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      var m = visibleItems().length;
      if (m) selectedIdx = (selectedIdx - 1 + m) % m;
      clampSelection();
    } else if (e.key === 'Enter') {
      e.preventDefault();
      activateSelected();
    }
  });

  bodyEl.addEventListener('click', function (e) {
    var btn = e.target && e.target.closest && e.target.closest('[data-cmd-item][data-href]');
    if (!btn || !dialog.contains(btn)) return;
    e.preventDefault();
    var href = btn.getAttribute('data-href');
    if (href) window.location.href = href;
  });

  bodyEl.addEventListener('mousemove', function (e) {
    var btn = e.target && e.target.closest && e.target.closest('[data-cmd-item][data-href]');
    if (!btn || !dialog.contains(btn)) return;
    var items = visibleItems();
    var idx = items.indexOf(btn);
    if (idx >= 0) {
      selectedIdx = idx;
      setActive(items);
    }
  });

  if (triggerBtn) {
    triggerBtn.setAttribute('aria-expanded', 'false');
    triggerBtn.addEventListener('click', function () {
      if (!dialog.open) openPalette();
    });
  }
})();
