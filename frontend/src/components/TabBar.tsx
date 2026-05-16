import { NavLink, Outlet } from 'react-router-dom'
import { Pencil, Clock, User } from 'lucide-react'

const tabs = [
  { path: '/register', icon: Pencil, label: 'Registrar' },
  { path: '/history', icon: Clock, label: 'Histórico' },
  { path: '/profile', icon: User, label: 'Perfil' },
]

export function TabBar() {
  return (
    <div className="flex flex-col min-h-screen">
      <main className="flex-1 overflow-y-auto pb-16">
        <Outlet />
      </main>

      <nav className="fixed bottom-0 left-0 right-0 bg-white border-t border-gray-200 z-50">
        <div className="flex justify-around items-center h-16 max-w-lg mx-auto">
          {tabs.map(({ path, icon: Icon, label }) => (
            <NavLink
              key={path}
              to={path}
              className={({ isActive }) =>
                `flex flex-col items-center gap-0.5 px-4 py-2 text-xs font-medium transition-colors
                ${isActive ? 'text-indigo-600' : 'text-gray-400 hover:text-gray-600'}`
              }
            >
              <Icon className="h-5 w-5" />
              <span>{label}</span>
            </NavLink>
          ))}
        </div>
      </nav>
    </div>
  )
}
