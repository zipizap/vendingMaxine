import { useState } from 'react'
import reactLogo from './assets/react.svg'
import viteLogo from '/vite.svg'
import './App.css'
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import Public from './Public';

function App() {
  const [count, setCount] = useState(0)
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
            <li><Link to="/public">Public</Link></li>
            <li><a href="/login" target="_self">Login</a></li>
            <li><a href="/private" target="_self">Private</a></li>
            <li><a href="/date_from_server" target="_self">Date from server</a></li>
            {/* <li><Link to="/date_from_server">Date from Server</Link></li> */}
            <li><a href="/logout" target="_self">Logout</a></li>
            <li><a href="#">About</a></li>
            <li><a href="#">Services</a></li>
            <li><a href="#">Contact</a></li>
          </ul>
        </nav>

        <Routes>
          <Route path="/" element={
            <div>
              <div>
                <a href="https://vite.dev" target="_blank">
                  <img src={viteLogo} className="logo" alt="Vite logo" />
                </a>
                <a href="https://react.dev" target="_blank">
                  <img src={reactLogo} className="logo react" alt="React logo" />
                </a>
              </div>
              <h1>Vite + React</h1>
              <div className="card">
                <button onClick={() => setCount((count) => count + 1)}>
                  count is {count}
                </button>
                <p>
                  Edit <code>src/App.tsx</code> and save to test HMR
                </p>
              </div>
              <p className="read-the-docs">
                Click on the Vite and React logos to learn more
              </p>
            </div>
          } />
          <Route path="/public" element={<Public />} />
        </Routes>
      </div>
    </Router>
  )
}

export default App