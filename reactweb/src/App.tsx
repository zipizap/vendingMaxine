import { useState } from 'react'
import './App.css'
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import Public from './Public';
import Collections from './Collections';
import Home from './Home';

function App() {
  const [menuOpen, setMenuOpen] = useState(false);

  const toggleMenu = () => {
    setMenuOpen(!menuOpen);
  };

  return (
    <Router>
      <div className="app-container">
        <div className="menu-icon" onClick={toggleMenu}>
          ☰
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
          <Route path="/public" element={<Public />} />
          <Route path="/collections" element={<Collections />} />
        </Routes>
      </div>
    </Router>
  )
}

export default App