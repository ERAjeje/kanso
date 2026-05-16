import { useAuth } from '../hooks/useAuth'

export function Register() {
  const { user } = useAuth()

  return (
    <div className="p-8 max-w-lg mx-auto">
      <h1 className="text-2xl font-bold text-gray-800 mb-2">Novo Registro</h1>
      <p className="text-gray-500 mb-6">
        Bem-vindo, {user?.name || 'usuário'}
      </p>
      <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-100">
        <p className="text-gray-400 text-center py-12">
          Formulário de registro emocional — será implementado na Fase 2
        </p>
      </div>
    </div>
  )
}
