import { Routes, Route } from 'react-router-dom'
import { AuthProvider } from './hooks/useAuth'
import { AuthGuard } from './components/AuthGuard'
import { Login } from './pages/Login'
import { Register } from './pages/Register'

function App() {
  return (
    <AuthProvider>
      <div className="min-h-screen bg-gray-50">
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={
            <AuthGuard>
              <Register />
            </AuthGuard>
          } />
          <Route path="*" element={<Login />} />
        </Routes>
      </div>
    </AuthProvider>
  )
}

export default App
