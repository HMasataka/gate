import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { isAuthenticated } from "../lib/auth";
import { api } from "../lib/api";
import { useToast } from "../components/Toast";
import Sidebar from "../components/Sidebar";

interface Role {
  ID: string;
  Name: string;
  Description: string;
  ParentID?: string | null;
  CreatedAt?: string;
  UpdatedAt?: string;
}

export default function Roles() {
  const [allRoles, setAllRoles] = useState<Role[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [searchQuery, setSearchQuery] = useState("");
  const [loading, setLoading] = useState(true);

  const [deleteId, setDeleteId] = useState<string | null>(null);
  const [deleteName, setDeleteName] = useState("");
  const [deleteLoading, setDeleteLoading] = useState(false);

  const [editRole, setEditRole] = useState<Role | null>(null);
  const [editName, setEditName] = useState("");
  const [editDesc, setEditDesc] = useState("");
  const [editParent, setEditParent] = useState("");
  const [editLoading, setEditLoading] = useState(false);
  const [editError, setEditError] = useState("");

  const [newModal, setNewModal] = useState(false);
  const [newName, setNewName] = useState("");
  const [newDesc, setNewDesc] = useState("");
  const [newParent, setNewParent] = useState("");
  const [newLoading, setNewLoading] = useState(false);
  const [newError, setNewError] = useState("");

  const [assignModal, setAssignModal] = useState<{ type: "user" | "permission"; roleId: string; roleName: string } | null>(null);
  const [assignId, setAssignId] = useState("");
  const [assignLoading, setAssignLoading] = useState(false);
  const [assignError, setAssignError] = useState("");

  const navigate = useNavigate();
  const { showToast } = useToast();

  useEffect(() => {
    if (!isAuthenticated()) {
      navigate("/login");
      return;
    }
    loadRoles();
  }, [navigate]);

  async function loadRoles() {
    setLoading(true);
    try {
      const data = await api<{ roles?: Role[]; data?: Role[] } | Role[]>(
        "GET",
        "/api/v1/admin/roles",
      );
      const list = Array.isArray(data)
        ? data
        : (data as { roles?: Role[] }).roles ?? (data as { data?: Role[] }).data ?? [];
      setAllRoles(list);
      setRoles(list);
    } catch (err) {
      showToast(err instanceof Error ? err.message : "Failed to load roles.", "error");
    } finally {
      setLoading(false);
    }
  }

  function filterRoles(query: string) {
    setSearchQuery(query);
    const q = query.toLowerCase();
    const filtered = allRoles.filter((r) => {
      return r.Name.toLowerCase().includes(q) || r.ID.toLowerCase().includes(q);
    });
    setRoles(filtered);
  }

  function openNew() {
    setNewName("");
    setNewDesc("");
    setNewParent("");
    setNewError("");
    setNewModal(true);
  }

  async function createRole() {
    setNewError("");
    if (!newName.trim()) {
      setNewError("Name is required.");
      return;
    }
    setNewLoading(true);
    try {
      await api("POST", "/api/v1/admin/roles", {
        name: newName.trim(),
        description: newDesc.trim(),
        parent_id: newParent.trim() || "",
      });
      showToast("Role created.");
      setNewModal(false);
      await loadRoles();
    } catch (err) {
      setNewError(err instanceof Error ? err.message : "Creation failed.");
    } finally {
      setNewLoading(false);
    }
  }

  function openEdit(role: Role) {
    setEditRole(role);
    setEditName(role.Name);
    setEditDesc(role.Description);
    setEditParent(role.ParentID ?? "");
    setEditError("");
  }

  async function updateRole() {
    if (!editRole) return;
    setEditError("");
    if (!editName.trim()) {
      setEditError("Name is required.");
      return;
    }
    setEditLoading(true);
    try {
      await api("PUT", `/api/v1/admin/roles/${editRole.ID}`, {
        name: editName.trim(),
        description: editDesc.trim(),
        parent_id: editParent.trim() || "",
      });
      showToast("Role updated.");
      setEditRole(null);
      await loadRoles();
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
      await api("DELETE", `/api/v1/admin/roles/${deleteId}`);
      showToast("Role deleted.");
      setDeleteId(null);
      await loadRoles();
    } catch (err) {
      showToast(err instanceof Error ? err.message : "Delete failed.", "error");
    } finally {
      setDeleteLoading(false);
    }
  }

  function openAssign(type: "user" | "permission", roleId: string, roleName: string) {
    setAssignModal({ type, roleId, roleName });
    setAssignId("");
    setAssignError("");
  }

  async function submitAssign() {
    if (!assignModal) return;
    setAssignError("");
    if (!assignId.trim()) {
      setAssignError(`${assignModal.type === "user" ? "User" : "Permission"} ID is required.`);
      return;
    }
    setAssignLoading(true);
    try {
      const suffix = assignModal.type === "user"
        ? `/users/${assignId.trim()}`
        : `/permissions/${assignId.trim()}`;
      await api("POST", `/api/v1/admin/roles/${assignModal.roleId}${suffix}`);
      showToast(`${assignModal.type === "user" ? "User" : "Permission"} assigned to role.`);
      setAssignModal(null);
    } catch (err) {
      setAssignError(err instanceof Error ? err.message : "Assignment failed.");
    } finally {
      setAssignLoading(false);
    }
  }

  return (
    <div className="page-layout">
      <Sidebar />
      <div className="main-content">
        <div className="page-header">
          <h1>Roles</h1>
          <p>Manage roles and their assignments.</p>
        </div>

        <div className="content-area">
          <div className="search-bar">
            <div className="search-input-wrap">
              <span className="search-input-icon">/</span>
              <input
                className="search-input"
                type="text"
                placeholder="Search roles..."
                value={searchQuery}
                onChange={(e) => filterRoles(e.target.value)}
              />
            </div>
            <button className="btn btn-primary" onClick={openNew}>
              + New Role
            </button>
          </div>

          {loading && (
            <div className="empty-state" style={{ padding: 32 }}>
              <p>Loading roles...</p>
            </div>
          )}

          {!loading && roles.length === 0 && (
            <div className="empty-state" style={{ padding: 32 }}>
              <h3>No roles found</h3>
              <p>Create a new role to get started.</p>
            </div>
          )}

          {!loading && roles.length > 0 && (
            <div className="table-container">
              <table className="data-table">
                <thead>
                  <tr>
                    <th>ID</th>
                    <th>Name</th>
                    <th>Description</th>
                    <th>Parent ID</th>
                    <th>Created</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {roles.map((r) => (
                    <tr key={r.ID}>
                      <td className="cell-mono">{r.ID}</td>
                      <td>{r.Name}</td>
                      <td>{r.Description || "-"}</td>
                      <td className="cell-mono">{r.ParentID ?? "-"}</td>
                      <td>{r.CreatedAt ? new Date(r.CreatedAt).toLocaleDateString() : "-"}</td>
                      <td>
                        <div className="table-actions">
                          <button className="btn btn-secondary btn-sm" onClick={() => openEdit(r)}>Edit</button>
                          <button className="btn btn-ghost btn-sm" onClick={() => openAssign("user", r.ID, r.Name)}>Assign User</button>
                          <button className="btn btn-ghost btn-sm" onClick={() => openAssign("permission", r.ID, r.Name)}>Assign Perm</button>
                          <button className="btn btn-danger btn-sm" onClick={() => { setDeleteId(r.ID); setDeleteName(r.Name); }}>Delete</button>
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

      {/* New role modal */}
      {newModal && (
        <div className="modal-overlay" onClick={() => setNewModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <span className="modal-title">New Role</span>
              <button className="modal-close" onClick={() => setNewModal(false)}>x</button>
            </div>
            <div className="form-group">
              <label className="form-label">Name</label>
              <input className="form-input" type="text" placeholder="admin" value={newName} onChange={(e) => setNewName(e.target.value)} />
            </div>
            <div className="form-group">
              <label className="form-label">Description</label>
              <input className="form-input" type="text" placeholder="Administrator role" value={newDesc} onChange={(e) => setNewDesc(e.target.value)} />
            </div>
            <div className="form-group">
              <label className="form-label">Parent Role ID (optional)</label>
              <input className="form-input" type="text" placeholder="" value={newParent} onChange={(e) => setNewParent(e.target.value)} />
            </div>
            {newError && <div className="alert alert-error" style={{ marginBottom: 8 }}>{newError}</div>}
            <div className="modal-footer">
              <button className="btn btn-secondary" onClick={() => setNewModal(false)}>Cancel</button>
              <button className="btn btn-primary" disabled={newLoading} onClick={createRole}>
                {newLoading ? "Creating..." : "Create Role"}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Edit role modal */}
      {editRole && (
        <div className="modal-overlay" onClick={() => setEditRole(null)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <span className="modal-title">Edit Role</span>
              <button className="modal-close" onClick={() => setEditRole(null)}>x</button>
            </div>
            <div className="form-group">
              <label className="form-label">Name</label>
              <input className="form-input" type="text" value={editName} onChange={(e) => setEditName(e.target.value)} />
            </div>
            <div className="form-group">
              <label className="form-label">Description</label>
              <input className="form-input" type="text" value={editDesc} onChange={(e) => setEditDesc(e.target.value)} />
            </div>
            <div className="form-group">
              <label className="form-label">Parent Role ID (optional)</label>
              <input className="form-input" type="text" value={editParent} onChange={(e) => setEditParent(e.target.value)} />
            </div>
            {editError && <div className="alert alert-error" style={{ marginBottom: 8 }}>{editError}</div>}
            <div className="modal-footer">
              <button className="btn btn-secondary" onClick={() => setEditRole(null)}>Cancel</button>
              <button className="btn btn-primary" disabled={editLoading} onClick={updateRole}>
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
              <span className="modal-title">Delete Role</span>
              <button className="modal-close" onClick={() => setDeleteId(null)}>x</button>
            </div>
            <p style={{ fontSize: 12, color: "var(--text-secondary)", marginBottom: 4 }}>
              Are you sure you want to delete this role? This action cannot be undone.
            </p>
            <p style={{ fontSize: 11, color: "var(--danger)", marginTop: 8 }}>
              {deleteName}
            </p>
            <div className="modal-footer">
              <button className="btn btn-secondary" onClick={() => setDeleteId(null)}>Cancel</button>
              <button className="btn btn-danger" disabled={deleteLoading} onClick={confirmDelete}>
                {deleteLoading ? "Deleting..." : "Delete Role"}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Assign modal */}
      {assignModal && (
        <div className="modal-overlay" onClick={() => setAssignModal(null)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <span className="modal-title">
                Assign {assignModal.type === "user" ? "User" : "Permission"} to {assignModal.roleName}
              </span>
              <button className="modal-close" onClick={() => setAssignModal(null)}>x</button>
            </div>
            <div className="form-group">
              <label className="form-label">
                {assignModal.type === "user" ? "User ID" : "Permission ID"}
              </label>
              <input className="form-input" type="text" value={assignId} onChange={(e) => setAssignId(e.target.value)} />
            </div>
            {assignError && <div className="alert alert-error" style={{ marginBottom: 8 }}>{assignError}</div>}
            <div className="modal-footer">
              <button className="btn btn-secondary" onClick={() => setAssignModal(null)}>Cancel</button>
              <button className="btn btn-primary" disabled={assignLoading} onClick={submitAssign}>
                {assignLoading ? "Assigning..." : "Assign"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
