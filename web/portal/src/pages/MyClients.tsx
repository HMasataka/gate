import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { isAuthenticated, getEmail, clearAuth } from "../lib/auth";
import { api } from "../lib/api";
import { useToast } from "../components/Toast";

interface Client {
  client_id: string;
  name?: string;
  client_name?: string;
  client_type?: string;
  redirect_uris?: string[];
  created_at?: string;
}

export default function MyClients() {
  const [clients, setClients] = useState<Client[]>([]);
  const [loading, setLoading] = useState(true);
  const [rotateClientId, setRotateClientId] = useState<string | null>(null);
  const [rotateLoading, setRotateLoading] = useState(false);
  const [newSecret, setNewSecret] = useState("");
  const navigate = useNavigate();
  const { showToast } = useToast();

  useEffect(() => {
    if (!isAuthenticated()) {
      navigate("/login");
      return;
    }
    loadClients();
  }, [navigate]);

  async function loadClients() {
    setLoading(true);
    try {
      const data = await api<Client[] | { clients?: Client[]; data?: Client[] }>(
        "GET",
        "/api/v1/clients",
      );
      const list = Array.isArray(data)
        ? data
        : (data as { clients?: Client[] }).clients ?? (data as { data?: Client[] }).data ?? [];
      setClients(list);
    } catch (err) {
      showToast(err instanceof Error ? err.message : "Failed to load clients.", "error");
    } finally {
      setLoading(false);
    }
  }

  function signOut() {
    clearAuth();
    navigate("/login");
  }

  async function confirmRotateSecret() {
    if (!rotateClientId) return;
    setRotateLoading(true);

    try {
      const data = await api<{ client_secret?: string; secret?: string }>(
        "POST",
        `/api/v1/clients/${rotateClientId}/rotate-secret`,
      );
      setNewSecret(data.client_secret ?? data.secret ?? "");
      showToast("Secret rotated successfully.");
    } catch (err) {
      showToast(err instanceof Error ? err.message : "Failed to rotate secret.", "error");
      setRotateClientId(null);
    } finally {
      setRotateLoading(false);
    }
  }

  function closeRotateModal() {
    setRotateClientId(null);
    setNewSecret("");
  }

  function copyText(text: string) {
    navigator.clipboard.writeText(text).then(
      () => showToast("Copied to clipboard."),
      () => showToast("Copy failed.", "error"),
    );
  }

  const email = getEmail();

  return (
    <div className="portal-layout">
      <nav className="portal-nav">
        <Link to="/clients" className="portal-nav-logo">
          <span>gate</span> / clients
        </Link>
        <div className="portal-nav-links">
          <Link to="/clients" className="portal-nav-link active">My Clients</Link>
          <Link to="/register-client" className="portal-nav-link">+ Register New</Link>
        </div>
        <div className="portal-nav-right">
          <span className="portal-user">{email}</span>
          <button className="btn btn-ghost btn-sm" onClick={signOut}>Sign out</button>
        </div>
      </nav>

      <div className="portal-content">
        <div className="portal-header">
          <h1>My Applications</h1>
          <p>OAuth clients registered under your account.</p>
        </div>

        {loading && (
          <div className="empty-state">
            <p>Loading clients...</p>
          </div>
        )}

        {!loading && clients.length === 0 && (
          <div className="empty-state">
            <h3>No applications yet</h3>
            <p>Register your first OAuth client to get started.</p>
            <br />
            <Link to="/register-client" className="btn btn-primary">
              Register a client
            </Link>
          </div>
        )}

        {!loading && clients.length > 0 && (
          <div className="clients-grid">
            {clients.map((client) => {
              const name = client.name ?? client.client_name ?? "Unnamed";
              const typeClass = client.client_type === "public" ? "badge-gray" : "badge-amber";
              const redirectURI =
                client.redirect_uris && client.redirect_uris.length > 0
                  ? client.redirect_uris[0]
                  : "-";
              const createdAt = client.created_at
                ? new Date(client.created_at).toLocaleDateString()
                : "-";

              return (
                <div key={client.client_id} className="client-card">
                  <div className="client-card-header">
                    <div>
                      <div className="client-name">{name}</div>
                      <div className="client-id">{client.client_id}</div>
                    </div>
                    <span className={`badge ${typeClass}`}>
                      {client.client_type ?? "confidential"}
                    </span>
                  </div>
                  <div className="client-meta">
                    <div className="client-meta-row">
                      <span className="client-meta-label">Redirect URI</span>
                      <span className="client-meta-value">{redirectURI}</span>
                    </div>
                    <div className="client-meta-row">
                      <span className="client-meta-label">Created</span>
                      <span className="client-meta-value">{createdAt}</span>
                    </div>
                    <div className="client-meta-row">
                      <span className="client-meta-label">Redirect URIs</span>
                      <span className="client-meta-value">
                        {(client.redirect_uris ?? []).length}
                      </span>
                    </div>
                  </div>
                  <div className="client-actions">
                    <Link
                      to={`/register-client?view=${encodeURIComponent(client.client_id)}`}
                      className="btn btn-secondary btn-sm"
                    >
                      View
                    </Link>
                    <button
                      className="btn btn-ghost btn-sm"
                      onClick={() => setRotateClientId(client.client_id)}
                    >
                      Rotate Secret
                    </button>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>

      {rotateClientId && (
        <div className="modal-overlay" onClick={closeRotateModal}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <span className="modal-title">Rotate Client Secret</span>
              <button className="modal-close" onClick={closeRotateModal}>x</button>
            </div>

            {!newSecret && (
              <p style={{ fontSize: 12, color: "var(--text-secondary)", marginBottom: 16 }}>
                A new client secret will be generated. The previous secret will be
                immediately invalidated.
              </p>
            )}

            {newSecret && (
              <div>
                <div className="alert alert-warning" style={{ marginBottom: 16 }}>
                  Save this secret now. It will not be shown again.
                </div>
                <div className="credential-group">
                  <div className="credential-label">New Client Secret</div>
                  <div className="secret-box">
                    <button
                      className="btn btn-secondary btn-sm copy-btn"
                      onClick={() => copyText(newSecret)}
                    >
                      Copy
                    </button>
                    <span>{newSecret}</span>
                  </div>
                </div>
              </div>
            )}

            <div className="modal-footer">
              <button className="btn btn-secondary" onClick={closeRotateModal}>Close</button>
              {!newSecret && (
                <button
                  className="btn btn-danger"
                  disabled={rotateLoading}
                  onClick={confirmRotateSecret}
                >
                  {rotateLoading ? "Rotating..." : "Rotate Secret"}
                </button>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
