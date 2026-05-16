import { useState } from 'react'
import { RegistrationForm } from '../components/RegistrationForm'
import { Toast } from '../components/Toast'
import { SyncStatus } from '../components/SyncStatus'

export function Register() {
  const [toastVisible, setToastVisible] = useState(false)
  const [toastMessage, setToastMessage] = useState('')

  const handleShowToast = (message: string) => {
    setToastMessage(message)
    setToastVisible(true)
  }

  return (
    <div className="p-8 max-w-lg mx-auto">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-800">Novo Registro</h1>
        <SyncStatus />
      </div>

      <div className="bg-white rounded-xl p-6 shadow-sm border border-gray-100">
        <RegistrationForm
          onSaved={() => {}}
          onShowToast={handleShowToast}
        />
      </div>

      <Toast
        message={toastMessage}
        visible={toastVisible}
        onClose={() => setToastVisible(false)}
      />
    </div>
  )
}
