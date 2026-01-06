/**
 * MonitoringDashboard
 * Displays system health, model performance, and cost stats
 */
class MonitoringDashboard {
    constructor(containerId) {
        this.container = document.getElementById(containerId);
        this.stats = null;
        this.refreshInterval = null;

        if (this.container) {
            this.render();
            this.fetchStats();
            this.startAutoRefresh();
        }
    }

    render() {
        this.container.innerHTML = `
            <div class="monitoring-dashboard glass-panel">
                <div class="dashboard-header">
                    <h2><i class="fas fa-chart-line"></i> System Observability</h2>
                    <button id="refresh-stats-btn" class="btn-refresh"><i class="fas fa-sync-alt"></i></button>
                </div>
                
                <div class="stats-grid">
                    <div class="stat-card">
                        <div class="stat-label">Total Requests</div>
                        <div id="stat-requests" class="stat-value">--</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-label">Total Cost</div>
                        <div id="stat-cost" class="stat-value">--</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-label">Avg Latency</div>
                        <div id="stat-latency" class="stat-value">--</div>
                    </div>
                    <div class="stat-card">
                        <div class="stat-label">Success Rate</div>
                        <div id="stat-success" class="stat-value">--</div>
                    </div>
                </div>

                <div class="dashboard-sections">
                    <div class="dashboard-section">
                        <h3>Model Performance Leaderboard</h3>
                        <div id="model-performance-list" class="performance-list">
                            <div class="loading-placeholder">Loading performance data...</div>
                        </div>
                    </div>
                    
                    <div class="dashboard-section">
                        <h3>Provider Health Status</h3>
                        <div id="provider-health-grid" class="health-grid">
                            <div class="health-item">
                                <span class="provider-name">OpenRouter</span>
                                <span class="status-indicator online"></span>
                            </div>
                            <div class="health-item">
                                <span class="provider-name">Google Gemini</span>
                                <span class="status-indicator online"></span>
                            </div>
                        </div>
                    </div>
                </div>
                
                <div class="dashboard-footer">
                    <span id="last-updated">Last updated: --</span>
                </div>
            </div>
        `;

        document.getElementById('refresh-stats-btn').addEventListener('click', () => this.fetchStats());
    }

    async fetchStats() {
        try {
            const response = await fetch('/api/monitoring/stats');
            this.stats = await response.json();
            this.updateUI();
        } catch (error) {
            console.error('Failed to fetch monitoring stats:', error);
        }
    }

    updateUI() {
        if (!this.stats) return;

        document.getElementById('stat-requests').textContent = this.stats.total_requests || '0';
        document.getElementById('stat-cost').textContent = `$${(this.stats.total_cost || 0).toFixed(4)}`;
        document.getElementById('stat-latency').textContent = `${this.stats.avg_latency_ms || 0}ms`;
        document.getElementById('stat-success').textContent = `${(this.stats.success_rate * 100 || 100).toFixed(1)}%`;

        // Update performance list
        const perfList = document.getElementById('model-performance-list');
        perfList.innerHTML = '';

        if (this.stats.model_performance && Object.keys(this.stats.model_performance).length > 0) {
            Object.entries(this.stats.model_performance)
                .sort(([, a], [, b]) => b - a)
                .forEach(([model, quality]) => {
                    const item = document.createElement('div');
                    item.className = 'performance-item';
                    item.innerHTML = `
                        <span class="model-id">${model.split('/').pop()}</span>
                        <div class="quality-bar-container">
                            <div class="quality-bar" style="width: ${quality * 100}%"></div>
                        </div>
                        <span class="quality-value">${(quality * 100).toFixed(0)}%</span>
                    `;
                    perfList.appendChild(item);
                });
        } else {
            perfList.innerHTML = '<div class="empty-state">No performance data yet.</div>';
        }

        document.getElementById('last-updated').textContent = `Last updated: ${new Date(this.stats.updated_at).toLocaleTimeString()}`;
    }

    startAutoRefresh() {
        this.refreshInterval = setInterval(() => this.fetchStats(), 30000); // Every 30s
    }

    stopAutoRefresh() {
        if (this.refreshInterval) clearInterval(this.refreshInterval);
    }
}
