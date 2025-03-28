import { useState, useEffect, useRef } from 'react'
import './App.css'
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import Collections from './Collections';
import Home from './Home';

function App() {
  const [menuOpen, setMenuOpen] = useState(false);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [userName, setUserName] = useState('');
  const [userEmail, setUserEmail] = useState('');
  const [userMenuOpen, setUserMenuOpen] = useState(false);
  const userMenuRef = useRef<HTMLDivElement>(null);

  const toggleMenu = () => {
    setMenuOpen(!menuOpen);
  };

  const toggleUserMenu = () => {
    setUserMenuOpen(!userMenuOpen);
  };

  const handleLogin = () => {
    const currentUrl = encodeURIComponent(window.location.href);
    window.location.href = `/login?redirect=${currentUrl}`;
  };

  const handleLogout = () => {
    window.location.href = '/logout';
  };

  // Close user menu when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (userMenuRef.current && !userMenuRef.current.contains(event.target as Node)) {
        setUserMenuOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, []);

  useEffect(() => {
    // Check authentication status when the component mounts
    const checkAuth = async () => {
      try {
        const response = await fetch('/api/public/check_auth');
        const data = await response.json();
        /*
        example data:
              {
                  "authenticated": true,
                  "claims": {
                      "at_hash": "fMB-Zq2MAHm-rv0gn5mDhA",
                      "aud": "my-vending-maxine-clientid",
                      "c_hash": "Wn53WLrDIW2N01qFsZGj6A",
                      "email": "kilgore@kilgore.trout",
                      "email_verified": true,
                      "exp": 1743250847,
                      "groups": [
                          "authors"
                      ],
                      "iat": 1743164447,
                      "iss": "http://127.0.0.1:5556/dex",
                      "name": "Kilgore Trout",
                      "sub": "Cg0wLTM4NS0yODA4OS0wEgRtb2Nr"
                  }
              }
        */
        
        setIsAuthenticated(data.authenticated);
        
        if (data.authenticated && data.claims) {
          if (data.claims.name) {
            setUserName(data.claims.name);
          }
          if (data.claims.email) {
            setUserEmail(data.claims.email);
          }
        }
      } catch (error) {
        console.error('Error checking authentication:', error);
      }
    };

    checkAuth();
  }, []);

  return (
    <Router>
      <div className="app-container">
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
                    <li onClick={handleLogout}>Logout</li>
                    <li><Link to="/profile">My Profile</Link></li>
                    <li><Link to="/settings">Settings</Link></li>
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

        <nav className={`nav-menu ${menuOpen ? 'open' : ''}`}>
          <ul>
            <li><Link to="/">Home</Link></li>
            <li><a href="/login" target="_self">Login</a></li>
            <li><a href="/logout" target="_self">Logout</a></li>
            <li><Link to="/collections">Collections</Link></li>
            <li><a href="#">About</a></li>
          </ul>
        </nav>

        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/collections" element={<Collections />} />
        </Routes>
      </div>
    </Router>
  )
}

export default App