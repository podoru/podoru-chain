import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter, Routes, Route, Link, useLocation } from 'react-router-dom'
import Dashboard from './pages/Dashboard'
import Blocks from './pages/Blocks'
import Transactions from './pages/Transactions'
import State from './pages/State'
import Network from './pages/Network'
import Wallet from './pages/Wallet'
import './styles.css'

function App() {
  const location = useLocation()

  const isActive = (path: string) => {
    if (path === '/') {
      return location.pathname === '/'
    }
    return location.pathname.startsWith(path)
  }

  return (
    <>
      <div className="header">
        <h1>⛓️ Podoru Chain Explorer</h1>
        <nav className="nav">
          <Link to="/" className={isActive('/') ? 'active' : ''}>Dashboard</Link>
          <Link to="/blocks" className={isActive('/blocks') ? 'active' : ''}>Blocks</Link>
          <Link to="/transactions" className={isActive('/transactions') ? 'active' : ''}>Transactions</Link>
          <Link to="/wallet" className={isActive('/wallet') ? 'active' : ''}>Wallet</Link>
          <Link to="/state" className={isActive('/state') ? 'active' : ''}>State Browser</Link>
          <Link to="/network" className={isActive('/network') ? 'active' : ''}>Network</Link>
        </nav>
      </div>
      <div className="container">
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/blocks" element={<Blocks />} />
          <Route path="/transactions" element={<Transactions />} />
          <Route path="/wallet" element={<Wallet />} />
          <Route path="/state" element={<State />} />
          <Route path="/network" element={<Network />} />
        </Routes>
      </div>
    </>
  )
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <BrowserRouter>
      <App />
    </BrowserRouter>
  </React.StrictMode>,
)
