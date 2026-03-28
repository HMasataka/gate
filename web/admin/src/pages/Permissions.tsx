import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { isAuthenticated } from "../lib/auth";
import { api } from "../lib/api";
import { useToast } from "../components/Toast";
import Sidebar from "../components/Sidebar";

interface Permission {
  ID: string;
  Name: string;
  Description: string;
  CreatedAt?: string;
  UpdatedAt?: string;
}

export default function Permissions() {
  const [allPermissions, setAllPermissions] = useState<Permission[]>([]);
  const [permissions, setPermissions] = useState<Permission[]>([]);
  const [searchQuery, setSearchQuery] = useState("");
  const [loading, setLoading] = useState(true);

  const [deleteId, setDeleteId] = useState<string | null>(null);
  const [deleteName, setDeleteName] = useState("");
  const [deleteLoading, setDeleteLoading] = useState(false);

  const [editPerm, setEditPerm] = useState<Permission | null>(null);
  const [editName, setEditName] = useState("");
  const [editDesc, setEditDesc] = useState("");
  const [editLoading, setEditLoading] = useState(false);
  const [editError, setEditError] = useState("");

  const [newModal, setNewModal] = useState(false);
  const [newName, setNewName] = useState("");
  const [newDesc, setNewDesc] = useState("");
  const [newLoading, setNewLoading] = useState(false);
  const [newError, setNewError] = useState("");

  const navigate = useNavigate();
  const { showToast } = useToast();

  useEffect(() => {
    if (!isAuthenticated()) {
      navigate("/login");
      return;
    }
    loadPermissions();
  }, [navigate]);

  async function loadPermissions() {
    setLoading(true);
    try {
      const data = await api<{ permissions?: Permission[]; data?: Permission[] } | Permission[]>(
        "GET",
        "/api/v1/admin/permissions",
      );
      const list = Array.isArray(data)
        ? data
        : (data as { permissions?: Permission[] }).permissions ?? (data as { data?: Permission[] }).data ?? [];
      setAllPermissions(list);
      setPermissions(list);
    } catch (err) {
      showToast(err instanceof Error ? err.message : "Failed to load permissions.", "error");
    } finally {
      setLoading(false);
    }
  }

  function filterPermissions(query: string) {
    setSearchQuery(query);
    const q = query.toLowerCase();
    const filtered = allPermissions.filter((p) => {
      return p.Name.toLowerCase().includes(q) || p.ID.toLowerCase().includes(q);
    });
    setPermissions(filtered);
  }

  function openNew() {
    setNewName("");
    setNewDesc("");
    setNewError("");
    setNewModal(true);
  }

  async function createPermission() {
    setNewError("");
    if (!newName.trim()) {
      setNewError("Name is required.");
      return;
    }
    setNewLoading(true);
    try {
      await api("POST", "/api/v1/admin/permissions", {
        name: newName.trim(),
        description: newDesc.trim(),
      });
      showToast("Permission created.");
      setNewModal(false);
      await loadPermissions();
    } catch (err) {
      setNewError(err instanceof Error ? err.message : "Creation failed.");
    } finally {
      setNewLoading(false);
    }
  }

  function openEdit(perm: Permission) {
    setEditPerm(perm);
    setEditName(perm.Name);
    setEditDesc(perm.Description);
    setEditError("");
  }

  async function updatePermission() {
    if (!editPerm) return;
    setEditError("");
    if (!editName.trim()) {
      setEditError("Name is required.");
      return;
    }
    setEditLoading(true);
    try {
      await api("PUT", `/api/v1/admin/permissions/${editPerm.ID}`, {
        name: editName.trim(),
        description: editDesc.trim(),
      });
      showToast("Permission updated.");
      setEditPerm(null);
      await loadPermissions();
    } catch (err) {
      setEditError(err instanceof Error ? err.message : "Update failed.");
    } finally {
      setEditLoading(false);
    }
  }

  async function confirmDelete() {
    if (!deleteId) return;
    setDeleteLoading(true);
    try {
      await api("DELETE", `/api/v1/admin/permissions/${deleteId}`);
      showToast("Permission deleted.");
      setDeleteId(null);
      await loadPermissions();
    } catch (err) {
      showToast(err instanceof Error ? err.message : "Delete failed.", "error");
    } finally {
      setDeleteLoading(false);
    }
  }

  return (
    <div className="page-layout">
      <Sidebar />
      <div className="main-content">
        <div className="page-header">
          <h1>Permissions</h1>
          <p>Manage permissions for role-based access control.</p>
        </div>

        <div className="content-area">
          <div className="search-bar">
            <div className="search-input-wrap">
              <span className="search-input-icon">/</span>
              <input
                className="search-input"
                type="text"
                placeholder="Search permissions..."
                value={searchQuery}
                onChange={(e) => filterPermissions(e.target.value)}
              />
            </div>
            <button className="btn btn-primary" onClick={openNew}>
              + New Permission
            </button>
          </div>

          {loading && (
            <div className="empty-state" style={{ padding: 32 }}>
              <p>Loading permissions...</p>
            </div>
          )}

          {!loading && permissions.length === 0 && (
            <div className="empty-state" style={{ padding: 32 }}>
              <h3>No permissions found</h3>
              <p>Create a new permission to get started.</p>
            </div>
          )}

          {!loading && permissions.length > 0 && (
            <div className="table-container">
              <table className="data-table">
                <thead>
                  <tr>
                    <th>ID</th>
                    <th>Name</th>
                    <th>Description</th>
                    <th>Created</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {permissions.map((p) => (
                    <tr key={p.ID}>
                      <td className="cell-mono">{p.ID}</td>
                      <td>{p.Name}</td>
                      <td>{p.Description || "-"}</td>
                      <td>{p.CreatedAt ? new Date(p.CreatedAt).toLocaleDateString() : "-"}</td>
                      <td>
                        <div className="table-actions">
                          <button className="btn btn-secondary btn-sm" onClick={() => openEdit(p)}>Edit</button>
                          <button className="btn btn-danger btn-sm" onClick={() => { setDeleteId(p.ID); setDeleteName(p.Name); }}>Delete</button>
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

      {/* New permission modal */}
      {newModal && (
        <div className="modal-overlay" onClick={() => setNewModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <span className="modal-title">New Permission</span>
              <button className="modal-close" onClick={() => setNewModal(false)}>x</button>
            </div>
            <div className="form-group">
              <label className="form-label">Name</label>
              <input className="form-input" type="text" placeholder="read:users" value={newName} onChange={(e) => setNewName(e.target.value)} />
            </div>
            <div className="form-group">
              <label className="form-label">Description</label>
              <input className="form-input" type="text" placeholder="Can read user data" value={newDesc} onChange={(e) => setNewDesc(e.target.value)} />
            </div>
            {newError && <div className="alert alert-error" style={{ marginBottom: 8 }}>{newError}</div>}
            <div className="modal-footer">
              <button className="btn btn-secondary" onClick={() => setNewModal(false)}>Cancel</button>
              <button className="btn btn-primary" disabled={newLoading} onClick={createPermission}>
                {newLoading ? "Creating..." : "Create Permission"}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Edit permission modal */}
      {editPerm && (
        <div className="modal-overlay" onClick={() => setEditPerm(null)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <span className="modal-title">Edit Permission</span>
              <button className="modal-close" onClick={() => setEditPerm(null)}>x</button>
            </div>
            <div className="form-group">
              <label className="form-label">Name</label>
              <input className="form-input" type="text" value={editName} onChange={(e) => setEditName(e.target.value)} />
            </div>
            <div className="form-group">
              <label className="form-label">Description</label>
              <input className="form-input" type="text" value={editDesc} onChange={(e) => setEditDesc(e.target.value)} />
            </div>
            {editError && <div className="alert alert-error" style={{ marginBottom: 8 }}>{editError}</div>}
            <div className="modal-footer">
              <button className="btn btn-secondary" onClick={() => setEditPerm(null)}>Cancel</button>
              <button className="btn btn-primary" disabled={editLoading} onClick={updatePermission}>
                {editLoading ? "Saving..." : "Save Changes"}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Delete modal */}
      {deleteId && (
        <div className="modal-overlay" onClick={() => setDeleteId(null)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <span className="modal-title">Delete Permission</span>
              <button className="modal-close" onClick={() => setDeleteId(null)}>x</button>
            </div>
            <p style={{ fontSize: 12, color: "var(--text-secondary)", marginBottom: 4 }}>
              Are you sure you want to delete this permission? This action cannot be undone.
            </p>
            <p style={{ fontSize: 11, color: "var(--danger)", marginTop: 8 }}>
              {deleteName}
            </p>
            <div className="modal-footer">
              <button className="btn btn-secondary" onClick={() => setDeleteId(null)}>Cancel</button>
              <button className="btn btn-danger" disabled={deleteLoading} onClick={confirmDelete}>
                {deleteLoading ? "Deleting..." : "Delete Permission"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
