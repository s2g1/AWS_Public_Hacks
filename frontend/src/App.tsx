import { Routes, Route } from 'react-router-dom'
import MainLayout from './components/layout/MainLayout'
import Dashboard from './pages/Dashboard'
import Payments from './pages/Payments'
import Upload from './pages/Upload'
import Contracts from './pages/Contracts'
import Alerts from './pages/Alerts'
import ChatWidget from './components/ChatWidget'

function App() {
  return (
    <>
      <MainLayout>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/payments" element={<Payments />} />
          <Route path="/upload" element={<Upload />} />
          <Route path="/contracts" element={<Contracts />} />
          <Route path="/alerts" element={<Alerts />} />
        </Routes>
      </MainLayout>
      <ChatWidget />
    </>
  )
}

export default App
