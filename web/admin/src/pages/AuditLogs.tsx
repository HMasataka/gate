import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { isAuthenticated } from "../lib/auth";
import { api } from "../lib/api";
import { useToast } from "../components/Toast";
import Sidebar from "../components/Sidebar";

interface AuditLog {
  ID: string;
  UserID?: string | null;
  Action: string;
  IPAddress: string;
  UserAgent: string;
  Metadata?: Record<string, unknown>;
  CreatedAt?: string;
}

export default function AuditLogs() {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [filterAction, setFilterAction] = useState("");
  const [filterUserId, setFilterUserId] = useState("");

  const navigate = useNavigate();
  const { showToast } = useToast();

  useEffect(() => {
    if (!isAuthenticated()) {
      navigate("/login");
      return;
    }
    loadLogs();
  }, [navigate]);

  async function loadLogs() {
    setLoading(true);
    try {
      const params = new URLSearchParams();
      if (filterAction) params.set("action", filterAction);
      if (filterUserId) params.set("user_id", filterUserId);
      const qs = params.toString();
      const url = `/api/v1/admin/audit-logs${qs ? `?${qs}` : ""}`;
      const data = await api<{ audit_logs?: AuditLog[]; data?: AuditLog[] } | AuditLog[]>(
        "GET",
        url,
      );
      const list = Array.isArray(data)
        ? data
        : (data as { audit_logs?: AuditLog[] }).audit_logs ?? (data as { data?: AuditLog[] }).data ?? [];
      setLogs(list);
    } catch (err) {
      showToast(err instanceof Error ? err.message : "Failed to load audit logs.", "error");
    } finally {
      setLoading(false);
    }
  }

  function applyFilters() {
    loadLogs();
  }

  function clearFilters() {
    setFilterAction("");
    setFilterUserId("");
  }

  function actionBadgeClass(action: string): string {
    switch (action) {
      case "login":
      case "register":
        return "badge-green";
      case "login_failed":
        return "badge-red";
      case "logout":
      case "token_revoke":
        return "badge-amber";
      default:
        return "badge-gray";
    }
  }

  return (
    <div className="page-layout">
      <Sidebar />
      <div className="main-content">
        <div className="page-header">
          <h1>Audit Logs</h1>
          <p>View authentication and authorization activity.</p>
        </div>

        <div className="content-area">
          <div className="search-bar">
            <div className="search-input-wrap">
              <span className="search-input-icon">/</span>
              <input
                className="search-input"
                type="text"
                placeholder="Filter by User ID..."
                value={filterUserId}
                onChange={(e) => setFilterUserId(e.target.value)}
              />
            </div>
            <select
              className="form-input"
              style={{ width: "auto", minWidth: 160 }}
              value={filterAction}
              onChange={(e) => setFilterAction(e.target.value)}
            >
              <option value="">All actions</option>
              <option value="login">login</option>
              <option value="logout">logout</option>
              <option value="login_failed">login_failed</option>
              <option value="register">register</option>
              <option value="password_change">password_change</option>
              <option value="password_reset">password_reset</option>
              <option value="token_issue">token_issue</option>
              <option value="token_revoke">token_revoke</option>
              <option value="role_change">role_change</option>
              <option value="permission_change">permission_change</option>
              <option value="mfa_setup">mfa_setup</option>
              <option value="mfa_disable">mfa_disable</option>
              <option value="admin_action">admin_action</option>
              <option value="social_login">social_login</option>
            </select>
            <button className="btn btn-primary" onClick={applyFilters}>
              Filter
            </button>
            <button className="btn btn-ghost" onClick={clearFilters}>
              Clear
            </button>
          </div>

          {loading && (
            <div className="empty-state" style={{ padding: 32 }}>
              <p>Loading audit logs...</p>
            </div>
          )}

          {!loading && logs.length === 0 && (
            <div className="empty-state" style={{ padding: 32 }}>
              <h3>No audit logs found</h3>
              <p>No activity matches the current filters.</p>
            </div>
          )}

          {!loading && logs.length > 0 && (
            <div className="table-container">
              <table className="data-table">
                <thead>
                  <tr>
                    <th>Timestamp</th>
                    <th>Action</th>
                    <th>User ID</th>
                    <th>IP Address</th>
                    <th>User Agent</th>
                  </tr>
                </thead>
                <tbody>
                  {logs.map((log) => (
                    <tr key={log.ID}>
                      <td>{log.CreatedAt ? new Date(log.CreatedAt).toLocaleString() : "-"}</td>
                      <td>
                        <span className={`badge ${actionBadgeClass(log.Action)}`}>
                          {log.Action}
                        </span>
                      </td>
                      <td className="cell-mono">{log.UserID ?? "-"}</td>
                      <td className="cell-mono">{log.IPAddress || "-"}</td>
                      <td style={{ maxWidth: 200, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }} title={log.UserAgent}>
                        {log.UserAgent || "-"}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
