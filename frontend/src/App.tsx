import { Routes, Route } from 'react-router-dom'
import { AppProvider } from './store/AppContext'
import MainLayout from './components/layout/MainLayout'
import RoleSwitcher from './components/RoleSwitcher'
import Dashboard from './pages/Dashboard'
import Contracts from './pages/Contracts'
import Solicitations from './pages/Solicitations'
import History from './pages/History'
import ChatWidget from './components/ChatWidget'

function App() {
  return (
    <AppProvider>
      <RoleSwitcher />
      <MainLayout>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/contracts" element={<Contracts />} />
          <Route path="/solicitations" element={<Solicitations />} />
          <Route path="/history" element={<History />} />
        </Routes>
      </MainLayout>
      <ChatWidget />
    </AppProvider>
  )
}

export default App
