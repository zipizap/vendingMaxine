import { useRef, useState } from 'react'
import './App.css'
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import Collections from './Collections';
import Home from './Home';
import Profile from './Profile';
import About from './About';
import { AuthProvider, useAuth } from './context/AuthContext';
import PrivateRoute from './components/PrivateRoute';

// Separate menu component that uses auth context
const AppMenu = () => {
  const { isAuthenticated, userName, userEmail, handleLogin, handleLogout } = useAuth();
  const [menuOpen, setMenuOpen] = useState(false);
  const [userMenuOpen, setUserMenuOpen] = useState(false);
  const userMenuRef = useRef<HTMLDivElement>(null);
  const menuRef = useRef<HTMLElement>(null);

  const toggleMenu = () => {
    setMenuOpen(!menuOpen);
  };

  const toggleUserMenu = () => {
    setUserMenuOpen(!userMenuOpen);
  };

  // Close menus code remains the same...
  
  return (
    <>
      <div className="menu-icon" onClick={toggleMenu}>
        ☰
      </div>

      <div className="welcome-message" ref={userMenuRef}>
        {isAuthenticated ? (
          <>
            <div className="user-info" onClick={toggleUserMenu}>
              <p className="user-name">{userName} ▼</p>
              <p className="user-email">{userEmail}</p>
            </div>
            {userMenuOpen && (
              <div className="user-dropdown-menu">
                <ul>
                  <li><Link to="/profile">My Profile</Link></li>
                  <li onClick={handleLogout}>Logout</li>
                </ul>
              </div>
            )}
          </>
        ) : (
          <button className="login-button" onClick={handleLogin}>
            Login
          </button>
        )}
      </div>

      <nav className={`nav-menu ${menuOpen ? 'open' : ''}`} ref={menuRef}>
        <ul>
          <li><Link to="/">Home</Link></li>
          <li><Link to="/collections">Collections</Link></li>
          <li><Link to="/about">About</Link></li>
        </ul>
      </nav>
    </>
  );
};

function App() {
  return (
    <AuthProvider>
      <Router>
        <div className="app-container">
          <AppMenu />
          
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/collections" element={
              <PrivateRoute>
                <Collections />
              </PrivateRoute>
            } />
            <Route path="/profile" element={
              <PrivateRoute>
                <Profile />
              </PrivateRoute>
            } />
            <Route path="/about" element={<About />} />
          </Routes>
        </div>
      </Router>
    </AuthProvider>
  )
}

export default App