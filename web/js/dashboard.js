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

  var _redirecting = false;
  async function api(endpoint, opts) {
    if (_redirecting) return null;
    const token = typeof getAccessToken === 'function' ? getAccessToken() : null;
    var res;
    try {
      res = await fetch(endpoint, {
        ...opts,
        headers: { 'Content-Type': 'application/json', ...(token ? { Authorization: 'Bearer ' + token } : {}), ...(opts && opts.headers || {}) }
      });
    } catch (e) {
      console.warn('API fetch error:', e);
      return null;
    }
    if (res.status === 401) {
      if (_redirecting) return null;
      _redirecting = true;
      if (typeof clearTokens === 'function') clearTokens();
      window.location.href = '/login';
      return null;
    }
    const text = await res.text();
    if (!text) return null;
    try { return JSON.parse(text); } catch (e) { return text; }
  }

  async function loadUser() {
    const data = await api('/api/auth/user');
    if (data && data.user && data.user.email) userEmail.textContent = data.user.email;
    else if (data && data.email) userEmail.textContent = data.email;
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

  function escapeHtml(s) {
    return String(s || '')
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/\"/g, '&quot;')
      .replace(/'/g, '&#039;');
  }

  function providerFromModelId(modelId) {
    const s = String(modelId || '');
    const idx = s.indexOf(':');
    if (idx > 0) return s.slice(0, idx);
    return '';
  }

  function renderModelsV2(state) {
    const providerKeys = state.providerKeys || [];
    const customProviders = (state.customProviders && state.customProviders.providers) || [];
    const tenantModels = (state.tenantModels && state.tenantModels.models) || [];
    const tenantAvailable = (state.tenantAvailable && state.tenantAvailable.models) || [];
    const catalog = (state.catalog && state.catalog.models) || [];
    const prefs = state.preferences || {};

    const legacyProviders = ['openrouter', 'google', 'huggingface'];
    const byLegacy = {};
    providerKeys.forEach(function(k) { byLegacy[k.provider] = k; });

    const byCustom = {};
    customProviders.forEach(function(p) { byCustom[p.provider_key] = p; });

    let html = '';
    html += '<h3>Connect providers</h3>';
    html += '<p class="muted">Pick models first if you want. When you select a model, we’ll tell you which provider key you need to connect.</p>';

    // Legacy/built-in providers (existing UX)
    legacyProviders.forEach(function(prov) {
      const k = byLegacy[prov];
      html += '<div class="card" style="margin-bottom:1rem;"><strong>' + escapeHtml(prov) + '</strong> ';
      if (k && k.key_hint) html += 'Connected (' + escapeHtml(k.key_hint || '') + ') <button class="btn btn-secondary btn-remove-key" data-provider="' + escapeHtml(prov) + '">Remove</button>';
      else html += 'Not connected <button class="btn btn-add-key" data-provider="' + escapeHtml(prov) + '">Add key</button>';
      html += '<div class="form-group form-add-key" id="form-' + escapeHtml(prov) + '" style="display:none; margin-top:0.5rem;">' +
        '<input type="password" placeholder="API key" id="input-' + escapeHtml(prov) + '">' +
        '<button class="btn btn-save-key" data-provider="' + escapeHtml(prov) + '">Save</button>' +
        '</div></div>';
    });

    // Custom providers (new universal path)
    html += '<h3 style="margin-top:1.5rem;">Custom providers (advanced)</h3>';
    html += '<p class="muted">Add any provider endpoint + API key. For OpenAI-compatible APIs we call <code>/v1/chat/completions</code>. For Anthropic we use <code>/v1/messages</code>.</p>';

    const templates = [
      // Foundation model providers
      { key: 'openai', label: 'OpenAI', provider_type: 'openai_compatible', base_url: 'https://api.openai.com', note: 'GPT models via OpenAI-compatible API.' },
      { key: 'anthropic', label: 'Anthropic', provider_type: 'anthropic_messages', base_url: 'https://api.anthropic.com', note: 'Claude models via Anthropic Messages API.' },
      { key: 'deepseek', label: 'DeepSeek', provider_type: 'openai_compatible', base_url: 'https://api.deepseek.com', note: 'DeepSeek models via OpenAI-compatible API.' },
      { key: 'xai', label: 'xAI', provider_type: 'openai_compatible', base_url: 'https://api.x.ai/v1', note: 'Grok models (by xAI) via OpenAI-compatible API.' },
      { key: 'groq', label: 'Groq', provider_type: 'openai_compatible', base_url: 'https://api.groq.com/openai/v1', note: 'Groq inference platform (not xAI Grok). Hosts Llama, Mistral, Gemma, etc. at low latency.' },
      { key: 'together', label: 'Together', provider_type: 'openai_compatible', base_url: 'https://api.together.xyz/v1', note: 'Open weights + hosted models via OpenAI-compatible API.' },
      { key: 'fireworks', label: 'Fireworks', provider_type: 'openai_compatible', base_url: 'https://api.fireworks.ai/inference/v1', note: 'Fireworks models via OpenAI-compatible API.' },
      { key: 'mistral', label: 'Mistral', provider_type: 'openai_compatible', base_url: 'https://api.mistral.ai/v1', note: 'Mistral models via OpenAI-compatible API.' },
      { key: 'perplexity', label: 'Perplexity', provider_type: 'openai_compatible', base_url: 'https://api.perplexity.ai', note: 'Perplexity models via OpenAI-compatible API.' },

      // Generic catch-all
      { key: 'custom', label: 'Any OpenAI-compatible', provider_type: 'openai_compatible', base_url: '', note: 'Any OpenAI-style provider/proxy/local gateway that supports /v1/chat/completions.' },
    ];

    templates.forEach(function(tpl) {
      const existing = tpl.key !== 'custom' ? byCustom[tpl.key] : null;
      const cardTitle = tpl.key === 'custom' ? 'openai_compatible provider' : (tpl.label + ' (' + tpl.provider_type + ')');
      html += '<div class="card" style="margin-bottom:1rem;">';
      html += '<div style="display:flex;justify-content:space-between;gap:1rem;align-items:center;flex-wrap:wrap;">';
      html += '<div><strong>' + escapeHtml(cardTitle) + '</strong><div class="muted" style="font-size:0.85rem;margin-top:0.25rem;">' + escapeHtml(tpl.note) + '</div></div>';
      if (existing && existing.key_hint) {
        html += '<div>Connected (' + escapeHtml(existing.key_hint) + ') <button class="btn btn-secondary btn-remove-custom-provider" data-provider-key="' + escapeHtml(existing.provider_key) + '">Remove</button></div>';
      } else {
        html += '<button class="btn btn-secondary btn-add-custom-provider" data-template="' + escapeHtml(tpl.key) + '">Connect</button>';
      }
      html += '</div>';

      const formId = 'custom-provider-form-' + tpl.key;
      html += '<div id="' + escapeHtml(formId) + '" style="display:none;margin-top:0.75rem;">';
      html += '<div class="form-group"><label>Provider key</label><input id="cp-key-' + escapeHtml(tpl.key) + '" value="' + escapeHtml(tpl.key === 'custom' ? '' : tpl.key) + '" placeholder="e.g. together, groq, my-gateway"></div>';
      html += '<div class="form-group"><label>Provider type</label><input id="cp-type-' + escapeHtml(tpl.key) + '" value="' + escapeHtml(tpl.provider_type) + '" placeholder="openai_compatible or anthropic_messages"></div>';
      html += '<div class="form-group"><label>Base URL</label><input id="cp-url-' + escapeHtml(tpl.key) + '" value="' + escapeHtml(tpl.base_url) + '" placeholder="e.g. https://api.openai.com"></div>';
      html += '<div class="form-group"><label>API key</label><input type="password" id="cp-api-' + escapeHtml(tpl.key) + '" placeholder="secret"></div>';
      html += '<button class="btn btn-save-custom-provider" data-template="' + escapeHtml(tpl.key) + '">Save provider</button>';
      html += '</div></div>';
    });

    // Register models for custom providers
    html += '<h3 style="margin-top:1.5rem;">Register models</h3>';
    html += '<p class="muted">For custom providers, add the model IDs you want to route to. You can then request them via <code>provider_key:model_id</code> (or call <code>/v1/chat</code> with <code>provider_key</code> + <code>model_id</code>).</p>';
    const selectableProviders = customProviders.map(function(p) { return p.provider_key; });
    html += '<div class="card" style="margin-bottom:1rem;max-width:720px;">';
    html += '<div class="form-group"><label>Provider</label><select id="tmProvider" style="width:100%;max-width:400px;padding:0.5rem;background:var(--surface);border:1px solid var(--border);border-radius:6px;color:var(--text);">';
    html += '<option value="">Select provider</option>';
    selectableProviders.forEach(function(pk) { html += '<option value="' + escapeHtml(pk) + '">' + escapeHtml(pk) + '</option>'; });
    html += '</select></div>';
    html += '<div class="form-group"><label>Model ID</label><input id="tmModelId" placeholder="e.g. claude-3-5-sonnet-20241022, gpt-4o-mini, deepseek-chat"></div>';
    html += '<div class="form-group"><label>Display name (optional)</label><input id="tmDisplayName" placeholder="e.g. Claude 3.5 Sonnet"></div>';
    html += '<button class="btn" id="btnSaveTenantModel">Save model</button>';
    html += '</div>';

    if (tenantModels.length > 0) {
      html += '<h3>Your registered models</h3>';
      html += '<table><thead><tr><th>Provider</th><th>Model ID</th><th>Display name</th><th></th></tr></thead><tbody>';
      tenantModels.forEach(function(m) {
        html += '<tr>' +
          '<td><code>' + escapeHtml(m.provider_key || '') + '</code></td>' +
          '<td><code>' + escapeHtml(m.model_id || '') + '</code></td>' +
          '<td>' + escapeHtml(m.display_name || '') + '</td>' +
          '<td><button class="btn btn-secondary btn-delete-tenant-model" data-provider-key="' + escapeHtml(m.provider_key || '') + '" data-model-id="' + escapeHtml(m.model_id || '') + '">Remove</button></td>' +
          '</tr>';
      });
      html += '</tbody></table>';
    }

    // Tenant-available models list (what can actually be used right now)
    html += '<h3 style="margin-top:1.5rem;">Models available (usable now)</h3>';
    if (tenantAvailable.length > 0) {
      html += '<p class="muted">These are the model IDs that are currently available for your tenant.</p>';
      html += '<div class="form-group"><input id="modelSearch" placeholder="Search models (id, provider, name)" style="max-width:480px;"></div>';
      html += '<table><thead><tr><th>ID</th><th>Provider</th><th></th></tr></thead><tbody id="modelCatalogBody"></tbody></table>';
      // Body will be filled by JS for filtering
      html += '<div class="muted" style="margin-top:0.5rem;">Default model: <code>' + escapeHtml(prefs.default_model_id || 'auto') + '</code></div>';
    } else {
      html += '<p class="empty">No models available yet. Connect a provider key or add a custom provider + model.</p>';
    }

    // Global catalog (help users pick a model, mostly OpenRouter/HF/Gemini)
    if (catalog.length > 0) {
      html += '<h3 style="margin-top:1.5rem;">Model catalog</h3>';
      html += '<p class="muted">Browse models we know about (mainly built-in providers). Selecting a model will tell you which provider key to connect.</p>';
      html += '<div class="form-group"><input id="globalModelSearch" placeholder="Search global catalog" style="max-width:480px;"></div>';
      html += '<table><thead><tr><th>ID</th><th>Display name</th><th>Provider</th><th></th></tr></thead><tbody id="globalCatalogBody"></tbody></table>';
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
      const [providerKeys, customProviders, tenantModels, tenantAvailable, catalog, preferences] = await Promise.all([
        api('/api/settings/provider-keys'),
        api('/api/settings/providers'),
        api('/api/settings/models'),
        api('/api/tenant/models'),
        api('/api/models'),
        api('/api/settings/preferences')
      ]);
      content.innerHTML = renderModelsV2({
        providerKeys: providerKeys || [],
        customProviders: customProviders || { providers: [] },
        tenantModels: tenantModels || { models: [] },
        tenantAvailable: tenantAvailable || { models: [] },
        catalog: catalog || { models: [] },
        preferences: preferences || {}
      });

      // Legacy provider key handlers
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

      // Custom providers handlers
      content.querySelectorAll('.btn-add-custom-provider').forEach(function(btn) {
        btn.onclick = function() {
          const tpl = btn.dataset.template;
          const el = document.getElementById('custom-provider-form-' + tpl);
          if (el) el.style.display = 'block';
        };
      });
      content.querySelectorAll('.btn-save-custom-provider').forEach(function(btn) {
        btn.onclick = async function() {
          const tpl = btn.dataset.template;
          const providerKey = (document.getElementById('cp-key-' + tpl) || {}).value || '';
          const providerType = (document.getElementById('cp-type-' + tpl) || {}).value || '';
          const baseUrl = (document.getElementById('cp-url-' + tpl) || {}).value || '';
          const apiKey = (document.getElementById('cp-api-' + tpl) || {}).value || '';
          if (!providerKey || !apiKey) return;
          const res = await fetch('/api/settings/providers', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json', Authorization: 'Bearer ' + getAccessToken() },
            body: JSON.stringify({ provider_key: providerKey, provider_type: providerType, base_url: baseUrl, api_key: apiKey })
          });
          if (res.ok) showPage('models');
          else alert(await res.text());
        };
      });
      content.querySelectorAll('.btn-remove-custom-provider').forEach(function(btn) {
        btn.onclick = async function() {
          if (!confirm('Remove this provider?')) return;
          await fetch('/api/settings/providers?provider_key=' + encodeURIComponent(btn.dataset.providerKey), { method: 'DELETE', headers: { Authorization: 'Bearer ' + getAccessToken() } });
          showPage('models');
        };
      });

      // Tenant model registration handlers
      const btnSaveTenantModel = document.getElementById('btnSaveTenantModel');
      if (btnSaveTenantModel) {
        btnSaveTenantModel.onclick = async function() {
          const providerKey = (document.getElementById('tmProvider') || {}).value || '';
          const modelId = (document.getElementById('tmModelId') || {}).value || '';
          const displayName = (document.getElementById('tmDisplayName') || {}).value || '';
          if (!providerKey || !modelId) return;
          const res = await fetch('/api/settings/models', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json', Authorization: 'Bearer ' + getAccessToken() },
            body: JSON.stringify({ provider_key: providerKey, model_id: modelId, display_name: displayName })
          });
          if (res.ok) showPage('models');
          else alert(await res.text());
        };
      }
      content.querySelectorAll('.btn-delete-tenant-model').forEach(function(btn) {
        btn.onclick = async function() {
          if (!confirm('Remove this model?')) return;
          const pk = btn.dataset.providerKey;
          const mid = btn.dataset.modelId;
          await fetch('/api/settings/models?provider_key=' + encodeURIComponent(pk) + '&model_id=' + encodeURIComponent(mid), { method: 'DELETE', headers: { Authorization: 'Bearer ' + getAccessToken() } });
          showPage('models');
        };
      });

      // Fill "usable now" table and filter.
      const usable = (tenantAvailable && tenantAvailable.models) || [];
      const bodyEl = document.getElementById('modelCatalogBody');
      const searchEl = document.getElementById('modelSearch');
      function renderUsable(filter) {
        if (!bodyEl) return;
        const q = (filter || '').toLowerCase().trim();
        const rows = usable.filter(function(m) {
          const id = String(m.id || '');
          const p = String(m.provider || '');
          const name = String(m.display_name || '');
          if (!q) return true;
          return id.toLowerCase().includes(q) || p.toLowerCase().includes(q) || name.toLowerCase().includes(q);
        }).slice(0, 200);
        bodyEl.innerHTML = rows.map(function(m) {
          const id = String(m.id || '');
          const provider = String(m.provider || providerFromModelId(id) || '');
          return '<tr>' +
            '<td><code>' + escapeHtml(id) + '</code></td>' +
            '<td>' + escapeHtml(provider) + '</td>' +
            '<td><button class="btn btn-secondary btn-set-default" data-model-id="' + escapeHtml(id) + '">Set default</button></td>' +
            '</tr>';
        }).join('');
        bodyEl.querySelectorAll('.btn-set-default').forEach(function(b) {
          b.onclick = async function() {
            const mid = b.dataset.modelId;
            await fetch('/api/settings/preferences', { method: 'PUT', headers: { 'Content-Type': 'application/json', Authorization: 'Bearer ' + getAccessToken() }, body: JSON.stringify({ default_model_id: mid }) });
            showPage('models');
          };
        });
      }
      renderUsable('');
      if (searchEl) searchEl.oninput = function() { renderUsable(searchEl.value); };

      // Fill global catalog (helps user pick a model then connect provider)
      const global = (catalog && catalog.models) || [];
      const gBody = document.getElementById('globalCatalogBody');
      const gSearch = document.getElementById('globalModelSearch');
      function connectHint(provider) {
        if (provider === 'openrouter' || provider === 'huggingface' || provider === 'google' || provider === 'gemini') return provider === 'gemini' ? 'google' : provider;
        if (provider === 'ollama') return 'ollama';
        return provider;
      }
      function renderGlobal(filter) {
        if (!gBody) return;
        const q = (filter || '').toLowerCase().trim();
        const rows = global.filter(function(m) {
          const id = String(m.id || '');
          const p = String(m.provider || '');
          const name = String(m.display_name || '');
          if (!q) return true;
          return id.toLowerCase().includes(q) || p.toLowerCase().includes(q) || name.toLowerCase().includes(q);
        }).slice(0, 200);
        gBody.innerHTML = rows.map(function(m) {
          const id = String(m.id || '');
          const provider = String(m.provider || providerFromModelId(id) || '');
          const display = String(m.display_name || '');
          const hint = connectHint(provider);
          let action = '';
          if (hint === 'ollama') action = '<span class="muted">Local (no key)</span>';
          else action = '<button class="btn btn-secondary btn-connect-for-model" data-provider="' + escapeHtml(hint) + '">Connect provider</button>';
          return '<tr>' +
            '<td><code>' + escapeHtml(id) + '</code></td>' +
            '<td>' + escapeHtml(display) + '</td>' +
            '<td>' + escapeHtml(provider) + '</td>' +
            '<td>' + action + '</td>' +
            '</tr>';
        }).join('');
        gBody.querySelectorAll('.btn-connect-for-model').forEach(function(b) {
          b.onclick = function() {
            const prov = b.dataset.provider;
            // If it's a legacy provider, open its key form.
            const legacyForm = document.getElementById('form-' + prov);
            if (legacyForm) {
              legacyForm.style.display = 'block';
              const input = document.getElementById('input-' + prov);
              if (input) input.focus();
              return;
            }
            // Otherwise show advanced connect template if present.
            const adv = document.getElementById('custom-provider-form-' + prov);
            if (adv) adv.style.display = 'block';
          };
        });
      }
      renderGlobal('');
      if (gSearch) gSearch.oninput = function() { renderGlobal(gSearch.value); };
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
