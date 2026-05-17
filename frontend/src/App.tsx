import { Routes, Route } from 'react-router-dom'
import { AuthProvider } from './hooks/useAuth'
import { AuthGuard } from './components/AuthGuard'
import { TabBar } from './components/TabBar'
import { Login } from './pages/Login'
import { Register } from './pages/Register'
import { History } from './pages/History'
import { Profile } from './pages/Profile'

function App() {
  return (
    <AuthProvider>
      <div className="min-h-screen bg-gray-50">
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route element={<AuthGuard><TabBar /></AuthGuard>}>
            <Route path="/register" element={<Register />} />
            <Route path="/history" element={<History />} />
            <Route path="/profile" element={<Profile />} />
          </Route>
          <Route path="*" element={<Login />} />
        </Routes>
      </div>
    </AuthProvider>
  )
}

export default App
