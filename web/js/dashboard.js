(function() {
  if (typeof isAuthenticated !== 'function' || !isAuthenticated()) {
    window.location.href = '/login';
    return;
  }

  const content = document.getElementById('dashboardContent');
  const pageTitle = document.getElementById('pageTitle');
  const userEmail = document.getElementById('userEmail');
  const sidebar = document.getElementById('sidebar');
  const navToggle = document.getElementById('navToggle');
  const userBtn = document.getElementById('userBtn');
  const userDropdown = document.getElementById('userDropdown');

  function getPage() {
    const path = window.location.pathname.replace(/^\/dashboard\/?/, '') || 'home';
    return path || 'home';
  }

  function setActiveNav() {
    const page = getPage();
    document.querySelectorAll('.nav-link').forEach(function(a) {
      a.classList.toggle('active', a.getAttribute('data-page') === page);
    });
  }

  function setTitle(title) {
    pageTitle.textContent = title;
  }

  async function api(endpoint, opts) {
    const token = typeof getAccessToken === 'function' ? getAccessToken() : null;
    const res = await fetch(endpoint, {
      ...opts,
      headers: { 'Content-Type': 'application/json', ...(token ? { Authorization: 'Bearer ' + token } : {}), ...(opts && opts.headers || {}) }
    });
    if (res.status === 401) { window.location.href = '/login'; return; }
    const text = await res.text();
    if (!text) return null;
    try { return JSON.parse(text); } catch (e) { return text; }
  }

  async function loadUser() {
    const data = await api('/api/auth/user');
    if (data && data.email) userEmail.textContent = data.email;
  }

  function renderHome(data) {
    const usage = data.usage || {};
    const summary = usage.summary || {};
    const prefs = data.preferences || {};
    const keys = data.gaiolKeys || [];
    const providers = data.providerKeys || [];
    const budgetLimit = prefs.budget_limit != null ? Number(prefs.budget_limit) : null;
    const cost = Number(summary.cost || 0);
    let html = '';
    if (budgetLimit != null && budgetLimit > 0 && cost > budgetLimit) {
      html += '<div class="card" style="border-color:#f59e0b;margin-bottom:1rem;"><strong>Budget alert:</strong> Usage ($' + cost.toFixed(4) + ') exceeds your limit ($' + budgetLimit.toFixed(2) + '). <a href="/dashboard/settings">Update budget</a></div>';
    }
    html += '<div class="cards">' +
      '<div class="card"><div class="label">Requests</div><div class="value">' + (summary.requests || 0) + '</div></div>' +
      '<div class="card"><div class="label">Cost</div><div class="value">$' + (Number(summary.cost || 0).toFixed(4)) + '</div></div>' +
      '<div class="card"><div class="label">GAIOL keys</div><div class="value">' + (keys.length || 0) + '</div></div>' +
      '<div class="card"><div class="label">Providers</div><div class="value">' + (providers.length || 0) + '</div></div>' +
      '</div>' +
      '<p><a href="/dashboard/usage" class="btn btn-secondary">Usage</a> <a href="/dashboard/billing" class="btn btn-secondary">Billing</a> <a href="/dashboard/models" class="btn btn-secondary">Models</a> <a href="/dashboard/api-keys" class="btn btn-secondary">API keys</a></p>';
    return html;
  }

  function renderUsage(data) {
    const summary = data.summary || {};
    const byDay = data.by_day || [];
    const byProvider = data.by_provider || [];
    const byKey = data.by_key || [];
    byDay.sort(function(a,b) { return (a.date || '').localeCompare(b.date || ''); });
    let html = '<div class="cards"><div class="card"><div class="label">Requests</div><div class="value">' + (summary.requests || 0) + '</div></div>' +
      '<div class="card"><div class="label">Tokens</div><div class="value">' + (summary.tokens || 0) + '</div></div>' +
      '<div class="card"><div class="label">Cost</div><div class="value">$' + (Number(summary.cost || 0).toFixed(4)) + '</div></div></div>';
    html += '<p><button class="btn btn-secondary" id="btnExportUsage">Export CSV</button></p>';
    if (byDay.length > 0) {
      html += '<h3>Usage over time</h3><div style="max-width:600px;height:220px;"><canvas id="usageChart"></canvas></div>';
    }
    html += '<h3>By day</h3><table><thead><tr><th>Date</th><th>Requests</th><th>Tokens</th><th>Cost</th></tr></thead><tbody>';
    byDay.forEach(function(r) { html += '<tr><td>' + (r.date || '') + '</td><td>' + (r.requests || 0) + '</td><td>' + (r.tokens || 0) + '</td><td>$' + (Number(r.cost || 0).toFixed(4)) + '</td></tr>'; });
    html += '</tbody></table>';
    html += '<h3>By provider</h3><table><thead><tr><th>Provider</th><th>Requests</th><th>Tokens</th><th>Cost</th></tr></thead><tbody>';
    byProvider.forEach(function(r) { html += '<tr><td>' + (r.provider || '') + '</td><td>' + (r.requests || 0) + '</td><td>' + (r.tokens || 0) + '</td><td>$' + (Number(r.cost || 0).toFixed(4)) + '</td></tr>'; });
    html += '</tbody></table>';
    if (byKey.length > 0) {
      html += '<h3>By API key</h3><table><thead><tr><th>Key name</th><th>Requests</th><th>Tokens</th><th>Cost</th></tr></thead><tbody>';
      byKey.forEach(function(r) { html += '<tr><td>' + (r.key_name || r.key_id || '') + '</td><td>' + (r.requests || 0) + '</td><td>' + (r.tokens || 0) + '</td><td>$' + (Number(r.cost || 0).toFixed(4)) + '</td></tr>'; });
      html += '</tbody></table>';
    }
    if (byDay.length === 0 && byProvider.length === 0) html += '<p class="empty">No usage data yet.</p>';
    return html;
  }

  function renderActivity(activity) {
    const list = activity || [];
    let html = '<h3>Recent activity</h3><table><thead><tr><th>Time</th><th>Action</th><th>Details</th></tr></thead><tbody>';
    list.forEach(function(e) {
      const details = e.metadata && Object.keys(e.metadata).length ? JSON.stringify(e.metadata) : '';
      html += '<tr><td>' + (e.created_at ? new Date(e.created_at).toLocaleString() : '') + '</td><td>' + (e.action || '') + '</td><td>' + details + '</td></tr>';
    });
    html += '</tbody></table>';
    if (list.length === 0) html += '<p class="empty">No activity yet.</p>';
    return html;
  }

  function renderBilling(summary, history) {
    const s = summary || {};
    const h = (history && history.history) || [];
    let html = '<h3>This month</h3><div class="cards"><div class="card"><div class="label">Total cost</div><div class="value">$' + (Number(s.total_cost || 0).toFixed(4)) + '</div></div></div>';
    if ((s.by_provider || []).length) {
      html += '<table><thead><tr><th>Provider</th><th>Cost</th></tr></thead><tbody>';
      s.by_provider.forEach(function(p) { html += '<tr><td>' + (p.provider || '') + '</td><td>$' + (Number(p.cost || 0).toFixed(4)) + '</td></tr>'; });
      html += '</tbody></table>';
    }
    html += '<h3>History (last 6 months)</h3><table><thead><tr><th>Month</th><th>Cost</th></tr></thead><tbody>';
    h.forEach(function(r) { html += '<tr><td>' + (r.month || '') + '</td><td>$' + (Number(r.total_cost || 0).toFixed(4)) + '</td></tr>'; });
    html += '</tbody></table>';
    if (h.length === 0) html += '<p class="empty">No billing history yet.</p>';
    return html;
  }

  function renderModels(providerKeys, tenantModels) {
    const list = providerKeys || [];
    const providers = ['openrouter', 'google', 'huggingface'];
    let html = '<p>Connect provider API keys so GAIOL can route requests. Keys are stored encrypted.</p>';
    const byProvider = {};
    list.forEach(function(k) { byProvider[k.provider] = k; });
    providers.forEach(function(prov) {
      const k = byProvider[prov];
      html += '<div class="card" style="margin-bottom:1rem;"><strong>' + prov + '</strong> ';
      if (k && k.key_hint) html += 'Connected (' + (k.key_hint || '') + ') <button class="btn btn-secondary btn-remove-key" data-provider="' + prov + '">Remove</button>';
      else html += 'Not connected <button class="btn btn-add-key" data-provider="' + prov + '">Add key</button>';
      html += '<div class="form-group form-add-key" id="form-' + prov + '" style="display:none; margin-top:0.5rem;"><input type="password" placeholder="API key" id="input-' + prov + '"><button class="btn btn-save-key" data-provider="' + prov + '">Save</button></div></div>';
    });
    const models = (tenantModels && tenantModels.models) || [];
    if (models.length > 0) {
      html += '<h3>Models available</h3><p class="muted">You can use these model IDs with your GAIOL key.</p><table><thead><tr><th>ID</th><th>Display name</th><th>Provider</th></tr></thead><tbody>';
      models.forEach(function(m) { html += '<tr><td><code>' + (m.id || '') + '</code></td><td>' + (m.display_name || '') + '</td><td>' + (m.provider || '') + '</td></tr>'; });
      html += '</tbody></table>';
    } else {
      html += '<p class="empty">Add provider keys above to see models you can use.</p>';
    }
    return html;
  }

  function renderApiKeys(keys, createdKey) {
    const list = keys || [];
    let html = '';
    if (createdKey) {
      html += '<div class="key-reveal">' + createdKey + '</div><p class="key-warning">Copy this key now. We won\'t show it again.</p>';
    }
    html += '<button class="btn" id="btnCreateKey">Create key</button><table style="margin-top:1rem;"><thead><tr><th>Name</th><th>Last used</th><th>Created</th><th></th></tr></thead><tbody>';
    list.forEach(function(k) {
      html += '<tr><td>' + (k.name || 'default') + '</td><td>' + (k.last_used_at ? new Date(k.last_used_at).toLocaleString() : 'Never') + '</td><td>' + (k.created_at ? new Date(k.created_at).toLocaleString() : '') + '</td><td><button class="btn btn-secondary btn-revoke-key" data-id="' + (k.id || '') + '">Revoke</button></td></tr>';
    });
    html += '</tbody></table>';
    if (list.length === 0 && !createdKey) html += '<p class="empty">No API keys yet. Create one to use the inference API.</p>';
    return html;
  }

  function renderSettings(user, prefs, tenantModels) {
    const email = (user && user.email) || '';
    const budget = (prefs && prefs.budget_limit != null) ? prefs.budget_limit : '';
    const strategy = (prefs && prefs.strategy) || 'balanced';
    const defaultModel = (prefs && prefs.default_model_id) || '';
    const models = (tenantModels && tenantModels.models) || [];
    let html = '<div class="card" style="max-width:500px;"><div class="form-group"><label>Email</label><div>' + email + '</div></div>';
    html += '<h3>Preferences</h3><div class="form-group"><label>Monthly budget limit ($)</label><input type="number" id="prefBudget" min="0" step="0.01" placeholder="e.g. 10" value="' + budget + '"></div>';
    html += '<div class="form-group"><label>Strategy (cost vs quality)</label><select id="prefStrategy"><option value="balanced"' + (strategy === 'balanced' ? ' selected' : '') + '>Balanced</option><option value="cost"' + (strategy === 'cost' ? ' selected' : '') + '>Cost</option><option value="quality"' + (strategy === 'quality' ? ' selected' : '') + '>Quality</option></select></div>';
    html += '<div class="form-group"><label>Default model</label><select id="prefDefaultModel"><option value="">Use auto</option>';
    models.forEach(function(m) { html += '<option value="' + (m.id || '') + '"' + (defaultModel === m.id ? ' selected' : '') + '>' + (m.display_name || m.id) + '</option>'; });
    html += '</select></div><button class="btn" id="btnSavePrefs">Save preferences</button>';
    html += '<p style="margin-top:1rem;"><a href="/dashboard/models" class="btn btn-secondary">Manage provider keys</a></p></div>';
    return html;
  }

  async function showPage(page) {
    setActiveNav();
    if (page === 'home') {
      setTitle('Dashboard');
      const [usage, gaiolKeys, providerKeys, preferences] = await Promise.all([api('/api/usage'), api('/api/gaiol-keys'), api('/api/settings/provider-keys'), api('/api/settings/preferences')]);
      content.innerHTML = renderHome({ usage, gaiolKeys, providerKeys, preferences });
    } else if (page === 'usage') {
      setTitle('Usage');
      const data = await api('/api/usage');
      content.innerHTML = renderUsage(data || {});
      var byDay = (data && data.by_day) || [];
      if (byDay.length > 0 && typeof Chart !== 'undefined') {
        byDay.sort(function(a,b) { return (a.date || '').localeCompare(b.date || ''); });
        var ctx = document.getElementById('usageChart');
        if (ctx) new Chart(ctx.getContext('2d'), { type: 'line', data: { labels: byDay.map(function(r) { return r.date; }), datasets: [{ label: 'Cost ($)', data: byDay.map(function(r) { return r.cost || 0; }), borderColor: '#6366f1', fill: false }, { label: 'Requests', data: byDay.map(function(r) { return r.requests || 0; }), borderColor: '#22c55e', fill: false }] }, options: { responsive: true, maintainAspectRatio: false } });
      }
      content.querySelector('#btnExportUsage').onclick = async function() {
        const token = getAccessToken();
        const res = await fetch('/api/usage/export', { headers: token ? { Authorization: 'Bearer ' + token } : {} });
        const blob = await res.blob();
        const a = document.createElement('a'); a.href = URL.createObjectURL(blob); a.download = 'usage.csv'; a.click(); URL.revokeObjectURL(a.href);
      };
    } else if (page === 'activity') {
      setTitle('Activity');
      const data = await api('/api/activity');
      content.innerHTML = renderActivity(data && data.activity ? data.activity : []);
    } else if (page === 'billing') {
      setTitle('Billing');
      const [summary, history] = await Promise.all([api('/api/billing/summary'), api('/api/billing/history')]);
      content.innerHTML = renderBilling(summary || {}, history || {});
    } else if (page === 'models') {
      setTitle('Models');
      const [providerKeys, tenantModels] = await Promise.all([api('/api/settings/provider-keys'), api('/api/tenant/models')]);
      content.innerHTML = renderModels(providerKeys || [], tenantModels || {});
      content.querySelectorAll('.btn-add-key').forEach(function(btn) {
        btn.onclick = function() { document.getElementById('form-' + btn.dataset.provider).style.display = 'block'; };
      });
      content.querySelectorAll('.btn-save-key').forEach(function(btn) {
        btn.onclick = async function() {
          const prov = btn.dataset.provider;
          const key = document.getElementById('input-' + prov).value;
          if (!key) return;
          const res = await fetch('/api/settings/provider-keys', { method: 'POST', headers: { 'Content-Type': 'application/json', Authorization: 'Bearer ' + getAccessToken() }, body: JSON.stringify({ provider: prov, api_key: key }) });
          if (res.ok) showPage('models');
        };
      });
      content.querySelectorAll('.btn-remove-key').forEach(function(btn) {
        btn.onclick = async function() {
          if (!confirm('Remove this provider key?')) return;
          await fetch('/api/settings/provider-keys?provider=' + encodeURIComponent(btn.dataset.provider), { method: 'DELETE', headers: { Authorization: 'Bearer ' + getAccessToken() } });
          showPage('models');
        };
      });
    } else if (page === 'api-keys') {
      setTitle('API keys');
      const keys = await api('/api/gaiol-keys');
      content.innerHTML = renderApiKeys(keys || [], window._createdKey || null);
      window._createdKey = null;
      content.querySelector('#btnCreateKey').onclick = async function() {
        const res = await fetch('/api/gaiol-keys', { method: 'POST', headers: { 'Content-Type': 'application/json', Authorization: 'Bearer ' + getAccessToken() }, body: JSON.stringify({ name: 'default' }) });
        const data = await res.json();
        if (data && data.api_key) { window._createdKey = data.api_key; showPage('api-keys'); }
      };
      content.querySelectorAll('.btn-revoke-key').forEach(function(btn) {
        btn.onclick = async function() {
          if (!confirm('Revoke this key? It will stop working immediately.')) return;
          await fetch('/api/gaiol-keys/' + btn.dataset.id, { method: 'DELETE', headers: { Authorization: 'Bearer ' + getAccessToken() } });
          showPage('api-keys');
        };
      });
    } else if (page === 'settings') {
      setTitle('Settings');
      const [user, prefs, tenantModels] = await Promise.all([api('/api/auth/user'), api('/api/settings/preferences'), api('/api/tenant/models')]);
      content.innerHTML = renderSettings(user, prefs, tenantModels);
      content.querySelector('#btnSavePrefs').onclick = async function() {
        const budgetEl = document.getElementById('prefBudget');
        const budgetVal = budgetEl && budgetEl.value.trim() !== '' ? parseFloat(budgetEl.value) : null;
        const strategyEl = document.getElementById('prefStrategy');
        const modelEl = document.getElementById('prefDefaultModel');
        await fetch('/api/settings/preferences', { method: 'PUT', headers: { 'Content-Type': 'application/json', Authorization: 'Bearer ' + getAccessToken() }, body: JSON.stringify({ budget_limit: budgetVal, strategy: strategyEl ? strategyEl.value : 'balanced', default_model_id: modelEl ? modelEl.value : '' }) });
        showPage('settings');
      };
    } else {
      setTitle('Dashboard');
      content.innerHTML = '<p><a href="/dashboard">Go to Home</a></p>';
    }
  }

  userBtn.onclick = function() { userDropdown.classList.toggle('show'); };
  document.addEventListener('click', function(e) { if (!userBtn.contains(e.target) && !userDropdown.contains(e.target)) userDropdown.classList.remove('show'); });
  document.getElementById('logoutLink').onclick = function(e) { e.preventDefault(); (async function() { try { if (typeof signOut === 'function') await signOut(); } catch(err) {} window.location.href = '/'; })(); };

  navToggle.onclick = function() { sidebar.classList.toggle('open'); };
  document.querySelectorAll('.nav-link').forEach(function(a) {
    a.onclick = function(e) { e.preventDefault(); window.history.pushState({}, '', a.getAttribute('href')); showPage(getPage()); if (window.innerWidth <= 768) sidebar.classList.remove('open'); };
  });
  window.addEventListener('popstate', function() { showPage(getPage()); });

  loadUser().then(function() { showPage(getPage()); });
})();
