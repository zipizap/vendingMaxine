import { useState, useEffect, useRef } from 'react'
import './App.css'
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import Collections from './Collections';
import Home from './Home';
import Profile from './Profile';
import About from './About';

function App() {
  const [menuOpen, setMenuOpen] = useState(false);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [userName, setUserName] = useState('');
  const [userEmail, setUserEmail] = useState('');
  const [userMenuOpen, setUserMenuOpen] = useState(false);
  const userMenuRef = useRef<HTMLDivElement>(null);
  const menuRef = useRef<HTMLElement>(null);

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

  // Close hamburger menu when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        menuOpen &&
        menuRef.current && 
        !menuRef.current.contains(event.target as Node) &&
        !(event.target as Element).closest('.menu-icon')
      ) {
        setMenuOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [menuOpen]);

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

        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/collections" element={<Collections />} />
          <Route path="/profile" element={<Profile />} />
          <Route path="/about" element={<About />} />
        </Routes>
      </div>
    </Router>
  )
}

export default App