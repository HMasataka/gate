import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { isAuthenticated } from "../lib/auth";
import { api } from "../lib/api";
import Sidebar from "../components/Sidebar";

export default function Dashboard() {
  const [clientCount, setClientCount] = useState<string>("-");
  const navigate = useNavigate();

  useEffect(() => {
    if (!isAuthenticated()) {
      navigate("/login");
      return;
    }
    loadDashboard();
  }, [navigate]);

  async function loadDashboard() {
    try {
      const data = await api<{ total?: number; clients?: unknown[] }>(
        "GET",
        "/api/v1/admin/clients?limit=1&offset=0",
      );
      setClientCount(String(data.total ?? "-"));
    } catch {
      // dashboard stats are non-critical
    }
  }

  return (
    <div className="page-layout">
      <Sidebar />
      <div className="main-content">
        <div className="page-header">
          <h1>Dashboard</h1>
          <p>gate / admin</p>
        </div>
        <div className="content-area">
          <div className="stats-grid">
            <div className="stat-card">
              <div className="stat-lbl">Clients</div>
              <div className="stat-val">{clientCount}</div>
              <div className="stat-sub">registered OAuth clients</div>
            </div>
            <div className="stat-card">
              <div className="stat-lbl">Status</div>
              <div className="stat-val" style={{ fontSize: 16, color: "var(--success)" }}>
                ● online
              </div>
              <div className="stat-sub">API server</div>
            </div>
          </div>

          <div style={{ fontSize: 10, color: "var(--text-muted)", letterSpacing: 1, textTransform: "uppercase", marginBottom: 12 }}>
            Quick Access
          </div>
          <div className="nav-cards">
            <div className="nav-card" onClick={() => navigate("/clients")}>
              <div className="nc-icon">@</div>
              <div className="nc-title">OAuth Clients</div>
              <div className="nc-desc">
                Register and manage OAuth 2.0 clients. Configure redirect URIs and allowed scopes.
              </div>
            </div>
            <div className="nav-card" onClick={() => navigate("/users")}>
              <div className="nc-icon">*</div>
              <div className="nc-title">Users</div>
              <div className="nc-desc">
                Manage accounts, roles, and permissions. Lock, unlock, and review activity.
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
