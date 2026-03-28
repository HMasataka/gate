import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { isAuthenticated } from "../lib/auth";
import { api } from "../lib/api";
import { useToast } from "../components/Toast";
import Sidebar from "../components/Sidebar";

interface Client {
  client_id: string;
  name?: string;
  client_name?: string;
  client_type?: string;
  redirect_uris?: string[];
  scopes?: string[];
  grant_types?: string[];
  created_at?: string;
  updated_at?: string;
}

export default function Clients() {
  const [clients, setClients] = useState<Client[]>([]);
  const [allClients, setAllClients] = useState<Client[]>([]);
  const [searchQuery, setSearchQuery] = useState("");
  const [loading, setLoading] = useState(true);

  const [panelClient, setPanelClient] = useState<Client | null>(null);
  const [panelOpen, setPanelOpen] = useState(false);

  const [deleteId, setDeleteId] = useState<string | null>(null);
  const [deleteLoading, setDeleteLoading] = useState(false);

  const [rotateId, setRotateId] = useState<string | null>(null);
  const [rotateLoading, setRotateLoading] = useState(false);
  const [rotatedSecret, setRotatedSecret] = useState("");

  const [newClientModal, setNewClientModal] = useState(false);
  const [ncName, setNcName] = useState("");
  const [ncType, setNcType] = useState("confidential");
  const [ncUris, setNcUris] = useState("");
  const [ncError, setNcError] = useState("");
  const [ncLoading, setNcLoading] = useState(false);
  const [ncResultId, setNcResultId] = useState("");
  const [ncResultSecret, setNcResultSecret] = useState("");
  const [ncCreated, setNcCreated] = useState(false);

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
        "/api/v1/admin/clients",
      );
      const list = Array.isArray(data)
        ? data
        : (data as { clients?: Client[] }).clients ?? (data as { data?: Client[] }).data ?? [];
      setAllClients(list);
      setClients(list);
    } catch (err) {
      showToast(err instanceof Error ? err.message : "Failed to load clients.", "error");
    } finally {
      setLoading(false);
    }
  }

  function filterClients(query: string) {
    setSearchQuery(query);
    const q = query.toLowerCase();
    const filtered = allClients.filter((c) => {
      const name = (c.name ?? c.client_name ?? "").toLowerCase();
      const id = (c.client_id ?? "").toLowerCase();
      return name.includes(q) || id.includes(q);
    });
    setClients(filtered);
  }

  function openPanel(clientId: string) {
    const client = allClients.find((c) => c.client_id === clientId);
    if (!client) return;
    setPanelClient(client);
    setPanelOpen(true);
  }

  function closePanel() {
    setPanelOpen(false);
    setPanelClient(null);
  }

  async function confirmDelete() {
    if (!deleteId) return;
    setDeleteLoading(true);
    try {
      await api("DELETE", `/api/v1/admin/clients/${deleteId}`);
      showToast("Client deleted.");
      setDeleteId(null);
      await loadClients();
    } catch (err) {
      showToast(err instanceof Error ? err.message : "Delete failed.", "error");
    } finally {
      setDeleteLoading(false);
    }
  }

  async function confirmRotate() {
    if (!rotateId) return;
    setRotateLoading(true);
    try {
      const data = await api<{ client_secret?: string; secret?: string }>(
        "POST",
        `/api/v1/admin/clients/${rotateId}/rotate-secret`,
      );
      setRotatedSecret(data.client_secret ?? data.secret ?? "");
      showToast("Secret rotated successfully.");
    } catch (err) {
      showToast(err instanceof Error ? err.message : "Rotation failed.", "error");
      setRotateId(null);
    } finally {
      setRotateLoading(false);
    }
  }

  function closeRotateModal() {
    setRotateId(null);
    setRotatedSecret("");
  }

  function openNewClientModal() {
    setNcName("");
    setNcType("confidential");
    setNcUris("");
    setNcError("");
    setNcResultId("");
    setNcResultSecret("");
    setNcCreated(false);
    setNewClientModal(true);
  }

  async function createClient() {
    setNcError("");

    if (!ncName.trim()) {
      setNcError("Name is required.");
      return;
    }

    const redirectUris = ncUris
      .split("\n")
      .map((u) => u.trim())
      .filter((u) => u.length > 0);

    if (redirectUris.length === 0) {
      setNcError("At least one redirect URI is required.");
      return;
    }

    setNcLoading(true);
    try {
      const data = await api<{ client_id?: string; client_secret?: string }>(
        "POST",
        "/api/v1/admin/clients",
        {
          name: ncName.trim(),
          client_type: ncType,
          redirect_uris: redirectUris,
        },
      );
      setNcResultId(data.client_id ?? "");
      setNcResultSecret(data.client_secret ?? "");
      setNcCreated(true);
      showToast("Client created successfully.");
      await loadClients();
    } catch (err) {
      setNcError(err instanceof Error ? err.message : "Creation failed.");
    } finally {
      setNcLoading(false);
    }
  }

  function copyText(text: string) {
    navigator.clipboard.writeText(text).then(
      () => showToast("Copied to clipboard."),
      () => showToast("Copy failed.", "error"),
    );
  }

  return (
    <div className="page-layout">
      <Sidebar />
      <div className="main-content">
        <div className="page-header">
          <h1>OAuth Clients</h1>
          <p>Manage registered OAuth applications and credentials.</p>
        </div>

        <div className="content-area">
          <div className="search-bar">
            <div className="search-input-wrap">
              <span className="search-input-icon">/</span>
              <input
                className="search-input"
                type="text"
                placeholder="Search clients..."
                value={searchQuery}
                onChange={(e) => filterClients(e.target.value)}
              />
            </div>
            <button className="btn btn-primary" onClick={openNewClientModal}>
              + New Client
            </button>
          </div>

          {loading && (
            <div className="empty-state" style={{ padding: 32 }}>
              <p>Loading clients...</p>
            </div>
          )}

          {!loading && clients.length === 0 && (
            <div className="empty-state" style={{ padding: 32 }}>
              <h3>No clients found</h3>
              <p>Register a new OAuth client to get started.</p>
            </div>
          )}

          {!loading && clients.length > 0 && (
            <div className="table-container">
              <table className="data-table">
                <thead>
                  <tr>
                    <th>Client ID</th>
                    <th>Name</th>
                    <th>Type</th>
                    <th>Redirect URIs</th>
                    <th>Created</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {clients.map((c) => (
                    <tr key={c.client_id}>
                      <td className="cell-mono">{c.client_id}</td>
                      <td>{c.name ?? c.client_name ?? "Unnamed"}</td>
                      <td>
                        <span className={`badge ${c.client_type === "public" ? "badge-gray" : "badge-amber"}`}>
                          {c.client_type ?? "confidential"}
                        </span>
                      </td>
                      <td>{(c.redirect_uris ?? []).length}</td>
                      <td>{c.created_at ? new Date(c.created_at).toLocaleDateString() : "-"}</td>
                      <td>
                        <div className="table-actions">
                          <button className="btn btn-secondary btn-sm" onClick={() => openPanel(c.client_id)}>Details</button>
                          <button className="btn btn-ghost btn-sm" onClick={() => { setRotateId(c.client_id); setRotatedSecret(""); }}>Rotate</button>
                          <button className="btn btn-danger btn-sm" onClick={() => setDeleteId(c.client_id)}>Delete</button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>

      {/* Detail panel */}
      {panelOpen && panelClient && (
        <>
          <div className="panel-overlay" onClick={closePanel} />
          <div className="side-panel open">
            <div className="side-panel-header">
              <span className="side-panel-title">Client Details</span>
              <button className="side-panel-close" onClick={closePanel}>x</button>
            </div>
            <div className="side-panel-body">
              <div className="detail-group">
                <div className="detail-label">Name</div>
                <div className="detail-value">{panelClient.name ?? panelClient.client_name ?? "-"}</div>
              </div>
              <div className="detail-group">
                <div className="detail-label">Client ID</div>
                <div className="detail-value mono">{panelClient.client_id}</div>
              </div>
              <div className="detail-group">
                <div className="detail-label">Client Type</div>
                <div className="detail-value">
                  <span className={`badge ${panelClient.client_type === "public" ? "badge-gray" : "badge-amber"}`}>
                    {panelClient.client_type ?? "confidential"}
                  </span>
                </div>
              </div>
              <div className="detail-group">
                <div className="detail-label">Redirect URIs</div>
                <div className="detail-value mono" style={{ fontSize: 11, lineHeight: 1.8 }}>
                  {(panelClient.redirect_uris ?? []).length > 0
                    ? panelClient.redirect_uris!.map((u, i) => <div key={i}>{u}</div>)
                    : "-"}
                </div>
              </div>
              <div className="detail-group">
                <div className="detail-label">Scopes</div>
                <div className="detail-value">{(panelClient.scopes ?? []).join(", ") || "-"}</div>
              </div>
              <div className="detail-group">
                <div className="detail-label">Grant Types</div>
                <div className="detail-value">{(panelClient.grant_types ?? []).join(", ") || "-"}</div>
              </div>
              <div className="detail-group">
                <div className="detail-label">Created</div>
                <div className="detail-value">{panelClient.created_at ? new Date(panelClient.created_at).toLocaleString() : "-"}</div>
              </div>
              <div className="detail-group">
                <div className="detail-label">Updated</div>
                <div className="detail-value">{panelClient.updated_at ? new Date(panelClient.updated_at).toLocaleString() : "-"}</div>
              </div>
            </div>
            <div className="side-panel-footer">
              <button className="btn btn-danger btn-sm" onClick={() => { closePanel(); setDeleteId(panelClient.client_id); }}>Delete Client</button>
              <button className="btn btn-secondary btn-sm" onClick={() => { closePanel(); setRotateId(panelClient.client_id); setRotatedSecret(""); }}>Rotate Secret</button>
            </div>
          </div>
        </>
      )}

      {/* Delete modal */}
      {deleteId && (
        <div className="modal-overlay" onClick={() => setDeleteId(null)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <span className="modal-title">Delete Client</span>
              <button className="modal-close" onClick={() => setDeleteId(null)}>x</button>
            </div>
            <p style={{ fontSize: 12, color: "var(--text-secondary)", marginBottom: 8 }}>
              Are you sure you want to delete this client? This action cannot be undone. All tokens issued to this client will be revoked.
            </p>
            <div className="modal-footer">
              <button className="btn btn-secondary" onClick={() => setDeleteId(null)}>Cancel</button>
              <button className="btn btn-danger" disabled={deleteLoading} onClick={confirmDelete}>
                {deleteLoading ? "Deleting..." : "Delete"}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Rotate secret modal */}
      {rotateId && (
        <div className="modal-overlay" onClick={closeRotateModal}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <span className="modal-title">Rotate Client Secret</span>
              <button className="modal-close" onClick={closeRotateModal}>x</button>
            </div>

            {!rotatedSecret && (
              <p style={{ fontSize: 12, color: "var(--text-secondary)", marginBottom: 16 }}>
                A new client secret will be generated. The existing secret will be immediately invalidated.
              </p>
            )}

            {rotatedSecret && (
              <div>
                <div className="alert alert-warning" style={{ marginBottom: 16 }}>
                  Save this secret now. It will not be shown again.
                </div>
                <div className="credential-group">
                  <div className="credential-label">New Client Secret</div>
                  <div className="secret-box">
                    <button className="btn btn-secondary btn-sm copy-btn" onClick={() => copyText(rotatedSecret)}>Copy</button>
                    <span>{rotatedSecret}</span>
                  </div>
                </div>
              </div>
            )}

            <div className="modal-footer">
              <button className="btn btn-secondary" onClick={closeRotateModal}>Close</button>
              {!rotatedSecret && (
                <button className="btn btn-primary" disabled={rotateLoading} onClick={confirmRotate}>
                  {rotateLoading ? "Rotating..." : "Rotate Secret"}
                </button>
              )}
            </div>
          </div>
        </div>
      )}

      {/* New client modal */}
      {newClientModal && (
        <div className="modal-overlay" onClick={() => setNewClientModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <span className="modal-title">New OAuth Client</span>
              <button className="modal-close" onClick={() => setNewClientModal(false)}>x</button>
            </div>

            {!ncCreated && (
              <>
                <div className="form-group">
                  <label className="form-label">Name</label>
                  <input className="form-input" type="text" placeholder="My Application" value={ncName} onChange={(e) => setNcName(e.target.value)} />
                </div>
                <div className="form-group">
                  <label className="form-label">Client Type</label>
                  <select className="form-input" value={ncType} onChange={(e) => setNcType(e.target.value)}>
                    <option value="confidential">confidential</option>
                    <option value="public">public</option>
                  </select>
                </div>
                <div className="form-group">
                  <label className="form-label">Redirect URIs (one per line)</label>
                  <textarea className="form-input" rows={3} placeholder="http://localhost:3000/callback" value={ncUris} onChange={(e) => setNcUris(e.target.value)} />
                </div>
                {ncError && (
                  <div className="alert alert-error" style={{ marginBottom: 8 }}>{ncError}</div>
                )}
              </>
            )}

            {ncCreated && (
              <div>
                <div className="alert alert-warning" style={{ marginBottom: 12 }}>
                  Save credentials below. The secret will not be shown again.
                </div>
                <div className="credential-group">
                  <div className="credential-label">Client ID</div>
                  <div className="secret-box">
                    <button className="btn btn-secondary btn-sm copy-btn" onClick={() => copyText(ncResultId)}>Copy</button>
                    <span>{ncResultId}</span>
                  </div>
                </div>
                {ncResultSecret && (
                  <div className="credential-group">
                    <div className="credential-label">Client Secret</div>
                    <div className="secret-box">
                      <button className="btn btn-secondary btn-sm copy-btn" onClick={() => copyText(ncResultSecret)}>Copy</button>
                      <span>{ncResultSecret}</span>
                    </div>
                  </div>
                )}
              </div>
            )}

            <div className="modal-footer">
              <button className="btn btn-secondary" onClick={() => setNewClientModal(false)}>Close</button>
              {!ncCreated && (
                <button className="btn btn-primary" disabled={ncLoading} onClick={createClient}>
                  {ncLoading ? "Creating..." : "Create Client"}
                </button>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
