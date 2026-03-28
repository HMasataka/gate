import { NavLink, useNavigate } from "react-router-dom";
import { getEmail, clearAuth } from "../lib/auth";

export default function Sidebar() {
  const navigate = useNavigate();
  const email = getEmail();

  function signOut() {
    clearAuth();
    navigate("/login");
  }

  return (
    <aside className="sidebar">
      <div className="sidebar-logo">
        <span>gate</span>
        <small>admin console</small>
      </div>

      <nav>
        <div className="nav-section-label">Management</div>
        <NavLink
          to="/dashboard"
          className={({ isActive }) => `nav-item${isActive ? " active" : ""}`}
        >
          <span className="nav-icon">#</span>
          Dashboard
        </NavLink>
        <NavLink
          to="/clients"
          className={({ isActive }) => `nav-item${isActive ? " active" : ""}`}
        >
          <span className="nav-icon">@</span>
          Clients
        </NavLink>
        <NavLink
          to="/users"
          className={({ isActive }) => `nav-item${isActive ? " active" : ""}`}
        >
          <span className="nav-icon">*</span>
          Users
        </NavLink>
        <NavLink
          to="/roles"
          className={({ isActive }) => `nav-item${isActive ? " active" : ""}`}
        >
          <span className="nav-icon">&amp;</span>
          Roles
        </NavLink>
        <NavLink
          to="/permissions"
          className={({ isActive }) => `nav-item${isActive ? " active" : ""}`}
        >
          <span className="nav-icon">!</span>
          Permissions
        </NavLink>
        <NavLink
          to="/audit-logs"
          className={({ isActive }) => `nav-item${isActive ? " active" : ""}`}
        >
          <span className="nav-icon">~</span>
          Audit Logs
        </NavLink>
      </nav>

      <div className="sidebar-bottom">
        <div className="sidebar-user">{email}</div>
        <button
          className="btn btn-ghost btn-sm"
          style={{ marginTop: 8, width: "100%" }}
          onClick={signOut}
        >
          Sign out
        </button>
      </div>
    </aside>
  );
}
