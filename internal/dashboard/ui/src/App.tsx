import { useState, useEffect } from 'react'
import { BrowserRouter, Routes, Route, Link, useLocation } from 'react-router-dom'
import { useWebSocket } from './hooks/useWebSocket'
import { RequestList } from './pages/RequestList'
import { RequestDetail } from './pages/RequestDetail'
import { CostDashboard } from './pages/CostDashboard'
import { Timeline } from './pages/Timeline'

export default function App() {
  const { requests, connected } = useWebSocket()
  const [search, setSearch] = useState('')
  const [dark, setDark] = useState(() =>
    typeof window !== 'undefined' ? window.matchMedia('(prefers-color-scheme: dark)').matches : true
  )

  const totalCost = requests.reduce((s, r) => s + (r.total_cost || 0), 0)

  return (
    <div className={dark ? 'dark' : ''}>
      <div className="min-h-screen bg-gray-950 text-gray-100 font-mono text-sm">
        <BrowserRouter>
          <AppContent
            connected={connected}
            search={search}
            setSearch={setSearch}
            dark={dark}
            setDark={setDark}
            requestCount={requests.length}
            totalCost={totalCost}
            requests={requests}
          />
        </BrowserRouter>
      </div>
    </div>
  )
}

function AppContent({
  connected, search, setSearch, dark, setDark, requestCount, totalCost, requests,
}: {
  connected: boolean
  search: string
  setSearch: (s: string) => void
  dark: boolean
  setDark: (d: boolean) => void
  requestCount: number
  totalCost: number
  requests: ReturnType<typeof useWebSocket>['requests']
}) {
  // Keyboard shortcut: / or Cmd+K focuses search
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.key === '/' || (e.key === 'k' && (e.metaKey || e.ctrlKey))) && !(e.target instanceof HTMLInputElement)) {
        e.preventDefault()
        document.getElementById('search-input')?.focus()
      }
    }
    window.addEventListener('keydown', handler)
    return () => window.removeEventListener('keydown', handler)
  }, [])

  return (
    <>
      <Header
        connected={connected}
        search={search}
        setSearch={setSearch}
        dark={dark}
        setDark={setDark}
        requestCount={requestCount}
        totalCost={totalCost}
      />
      <Routes>
        <Route path="/" element={<RequestList requests={requests} search={search} />} />
        <Route path="/request/:id" element={<RequestDetail requests={requests} />} />
        <Route path="/cost" element={<CostDashboard requests={requests} />} />
        <Route path="/timeline" element={<Timeline requests={requests} />} />
      </Routes>
    </>
  )
}

function Header({
  connected, search, setSearch, dark, setDark, requestCount, totalCost,
}: {
  connected: boolean
  search: string
  setSearch: (s: string) => void
  dark: boolean
  setDark: (d: boolean) => void
  requestCount: number
  totalCost: number
}) {
  return (
    <header className="border-b border-gray-800 px-4 py-2 flex items-center gap-4 sticky top-0 z-50 bg-gray-950/95 backdrop-blur-sm">
      <div className="flex items-center gap-2">
        <span className="text-purple-400 font-bold text-base tracking-tight">probe</span>
        <span
          className={`w-2 h-2 rounded-full flex-shrink-0 ${connected ? 'bg-green-400' : 'bg-gray-600'}`}
          title={connected ? 'Connected' : 'Disconnected'}
        />
      </div>
      <Nav />
      <div className="flex-1" />
      <div className="text-xs text-gray-500 flex gap-3 items-center">
        <span>{requestCount} req</span>
        <span className="text-yellow-400">${totalCost.toFixed(6)}</span>
      </div>
      <input
        id="search-input"
        className="bg-gray-800 border border-gray-700 rounded px-2 py-1 text-xs w-44 focus:outline-none focus:border-purple-500 placeholder-gray-600"
        placeholder="Search (⌘K)"
        value={search}
        onChange={e => setSearch(e.target.value)}
        onKeyDown={e => e.key === 'Escape' && setSearch('')}
      />
      <button
        onClick={() => setDark(!dark)}
        className="text-gray-400 hover:text-gray-200 text-sm px-1.5 border-0 bg-transparent cursor-pointer"
        title={dark ? 'Switch to light mode' : 'Switch to dark mode'}
      >
        {dark ? '☀' : '🌙'}
      </button>
    </header>
  )
}

function Nav() {
  const location = useLocation()
  const links = [
    { to: '/', label: 'Requests' },
    { to: '/cost', label: 'Cost' },
    { to: '/timeline', label: 'Timeline' },
  ]
  return (
    <nav className="flex gap-0.5">
      {links.map(l => {
        const active = l.to === '/' ? location.pathname === '/' : location.pathname.startsWith(l.to)
        return (
          <Link key={l.to} to={l.to}
            className={`px-2.5 py-1 rounded text-xs transition-colors ${active ? 'bg-gray-800 text-white' : 'text-gray-400 hover:text-gray-200 hover:bg-gray-800/60'}`}>
            {l.label}
          </Link>
        )
      })}
    </nav>
  )
}
