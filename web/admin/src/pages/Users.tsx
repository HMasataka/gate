import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { isAuthenticated } from "../lib/auth";
import { api } from "../lib/api";
import { useToast } from "../components/Toast";
import Sidebar from "../components/Sidebar";

interface User {
  id?: string;
  uid?: string;
  user_id?: string;
  email?: string;
  status?: string;
  created_at?: string;
  last_sign_in_at?: string;
  last_login_at?: string;
}

function getUserId(user: User): string {
  return user.id ?? user.uid ?? user.user_id ?? "";
}

function StatusBadge({ status }: { status: string }) {
  const s = (status ?? "").toLowerCase();
  switch (s) {
    case "active":
      return <span className="badge badge-green">active</span>;
    case "locked":
      return <span className="badge badge-amber">locked</span>;
    case "deleted":
    case "pending_purge":
      return <span className="badge badge-red">{status}</span>;
    default:
      return <span className="badge badge-gray">{status || "unknown"}</span>;
  }
}

export default function Users() {
  const [allUsers, setAllUsers] = useState<User[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [searchQuery, setSearchQuery] = useState("");
  const [loading, setLoading] = useState(true);

  const [deleteId, setDeleteId] = useState<string | null>(null);
  const [deleteEmail, setDeleteEmail] = useState("");
  const [deleteLoading, setDeleteLoading] = useState(false);

  const [revokeId, setRevokeId] = useState<string | null>(null);
  const [revokeEmail, setRevokeEmail] = useState("");
  const [revokeLoading, setRevokeLoading] = useState(false);

  const navigate = useNavigate();
  const { showToast } = useToast();

  useEffect(() => {
    if (!isAuthenticated()) {
      navigate("/login");
      return;
    }
    loadUsers();
  }, [navigate]);

  async function loadUsers() {
    setLoading(true);
    try {
      const data = await api<User[] | { users?: User[]; data?: User[] }>(
        "GET",
        "/api/v1/admin/users",
      );
      const list = Array.isArray(data)
        ? data
        : (data as { users?: User[] }).users ?? (data as { data?: User[] }).data ?? [];
      setAllUsers(list);
      setUsers(list);
    } catch (err) {
      showToast(err instanceof Error ? err.message : "Failed to load users.", "error");
    } finally {
      setLoading(false);
    }
  }

  function filterUsers(query: string) {
    setSearchQuery(query);
    const q = query.toLowerCase();
    const filtered = allUsers.filter((u) => {
      const email = (u.email ?? "").toLowerCase();
      const uid = getUserId(u).toLowerCase();
      return email.includes(q) || uid.includes(q);
    });
    setUsers(filtered);
  }

  const totalUsers = allUsers.length;
  const activeUsers = allUsers.filter((u) => (u.status ?? "").toLowerCase() === "active").length;
  const lockedUsers = allUsers.filter((u) => (u.status ?? "").toLowerCase() === "locked").length;

  async function lockUser(uid: string) {
    try {
      await api("POST", `/api/v1/admin/users/${uid}/lock`);
      showToast("User locked.");
      await loadUsers();
    } catch (err) {
      showToast(err instanceof Error ? err.message : "Lock failed.", "error");
    }
  }

  async function unlockUser(uid: string) {
    try {
      await api("POST", `/api/v1/admin/users/${uid}/unlock`);
      showToast("User unlocked.");
      await loadUsers();
    } catch (err) {
      showToast(err instanceof Error ? err.message : "Unlock failed.", "error");
    }
  }

  async function confirmDelete() {
    if (!deleteId) return;
    setDeleteLoading(true);
    try {
      await api("DELETE", `/api/v1/admin/users/${deleteId}`);
      showToast("User deleted.");
      setDeleteId(null);
      await loadUsers();
    } catch (err) {
      showToast(err instanceof Error ? err.message : "Delete failed.", "error");
    } finally {
      setDeleteLoading(false);
    }
  }

  async function confirmRevoke() {
    if (!revokeId) return;
    setRevokeLoading(true);
    try {
      await api("POST", `/api/v1/admin/users/${revokeId}/revoke-tokens`);
      showToast("All tokens revoked.");
      setRevokeId(null);
    } catch (err) {
      showToast(err instanceof Error ? err.message : "Revocation failed.", "error");
    } finally {
      setRevokeLoading(false);
    }
  }

  return (
    <div className="page-layout">
      <Sidebar />
      <div className="main-content">
        <div className="page-header">
          <h1>Users</h1>
          <p>Manage user accounts and access control.</p>
        </div>

        <div className="content-area">
          <div className="stats-bar">
            <div className="stat-item">
              <span className="stat-value-inline">{totalUsers}</span>
              <span>Total Users</span>
            </div>
            <div className="stat-item">
              <span className="badge badge-green">{activeUsers}</span>
              <span>Active</span>
            </div>
            <div className="stat-item">
              <span className="badge badge-amber">{lockedUsers}</span>
              <span>Locked</span>
            </div>
          </div>

          <div className="search-bar">
            <div className="search-input-wrap">
              <span className="search-input-icon">/</span>
              <input
                className="search-input"
                type="text"
                placeholder="Search by email or UID..."
                value={searchQuery}
                onChange={(e) => filterUsers(e.target.value)}
              />
            </div>
          </div>

          {loading && (
            <div className="empty-state" style={{ padding: 32 }}>
              <p>Loading users...</p>
            </div>
          )}

          {!loading && users.length === 0 && (
            <div className="empty-state" style={{ padding: 32 }}>
              <h3>No users found</h3>
              <p>No user accounts match your search.</p>
            </div>
          )}

          {!loading && users.length > 0 && (
            <div className="table-container">
              <table className="data-table">
                <thead>
                  <tr>
                    <th>UID</th>
                    <th>Email</th>
                    <th>Status</th>
                    <th>Created</th>
                    <th>Last Sign-in</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {users.map((u) => {
                    const uid = getUserId(u);
                    const uidShort = uid.length > 16 ? uid.substring(0, 16) + "..." : uid;
                    const created = u.created_at ? new Date(u.created_at).toLocaleDateString() : "-";
                    const lastSignIn = u.last_sign_in_at ?? u.last_login_at
                      ? new Date((u.last_sign_in_at ?? u.last_login_at)!).toLocaleDateString()
                      : "-";
                    const statusLower = (u.status ?? "").toLowerCase();

                    return (
                      <tr key={uid}>
                        <td className="cell-mono" title={uid}>{uidShort}</td>
                        <td>{u.email ?? "-"}</td>
                        <td><StatusBadge status={u.status ?? ""} /></td>
                        <td>{created}</td>
                        <td>{lastSignIn}</td>
                        <td>
                          <div className="table-actions">
                            {statusLower === "active" && (
                              <button className="btn btn-ghost btn-sm" onClick={() => lockUser(uid)}>Lock</button>
                            )}
                            {statusLower === "locked" && (
                              <button className="btn btn-ghost btn-sm" onClick={() => unlockUser(uid)}>Unlock</button>
                            )}
                            <button className="btn btn-ghost btn-sm" onClick={() => { setRevokeId(uid); setRevokeEmail(u.email ?? uid); }}>Revoke Tokens</button>
                            <button className="btn btn-danger btn-sm" onClick={() => { setDeleteId(uid); setDeleteEmail(u.email ?? uid); }}>Delete</button>
                          </div>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>

      {/* Delete modal */}
      {deleteId && (
        <div className="modal-overlay" onClick={() => setDeleteId(null)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <span className="modal-title">Delete User</span>
              <button className="modal-close" onClick={() => setDeleteId(null)}>x</button>
            </div>
            <p style={{ fontSize: 12, color: "var(--text-secondary)", marginBottom: 4 }}>
              Are you sure you want to permanently delete this user account? This action cannot be undone and all user data will be removed.
            </p>
            <p style={{ fontSize: 11, color: "var(--danger)", marginTop: 8 }}>
              {deleteEmail}
            </p>
            <div className="modal-footer">
              <button className="btn btn-secondary" onClick={() => setDeleteId(null)}>Cancel</button>
              <button className="btn btn-danger" disabled={deleteLoading} onClick={confirmDelete}>
                {deleteLoading ? "Deleting..." : "Delete User"}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Revoke tokens modal */}
      {revokeId && (
        <div className="modal-overlay" onClick={() => setRevokeId(null)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <span className="modal-title">Revoke Tokens</span>
              <button className="modal-close" onClick={() => setRevokeId(null)}>x</button>
            </div>
            <p style={{ fontSize: 12, color: "var(--text-secondary)", marginBottom: 4 }}>
              All active access tokens and refresh tokens for this user will be immediately revoked. The user will be signed out of all sessions.
            </p>
            <p style={{ fontSize: 11, color: "var(--text-muted)", marginTop: 8 }}>
              {revokeEmail}
            </p>
            <div className="modal-footer">
              <button className="btn btn-secondary" onClick={() => setRevokeId(null)}>Cancel</button>
              <button className="btn btn-danger" disabled={revokeLoading} onClick={confirmRevoke}>
                {revokeLoading ? "Revoking..." : "Revoke All Tokens"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
