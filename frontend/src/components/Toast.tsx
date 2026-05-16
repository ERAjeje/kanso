import { useEffect } from 'react'

interface Props {
  message: string
  visible: boolean
  onClose: () => void
}

export function Toast({ message, visible, onClose }: Props) {
  useEffect(() => {
    if (!visible) return
    const timer = setTimeout(onClose, 4000)
    return () => clearTimeout(timer)
  }, [visible, onClose])

  if (!visible) return null

  return (
    <div className="fixed top-4 left-1/2 -translate-x-1/2 z-50 bg-white shadow-lg rounded-lg border border-gray-200 p-4 text-sm text-gray-800 max-w-sm w-full">
      {message}
    </div>
  )
}
