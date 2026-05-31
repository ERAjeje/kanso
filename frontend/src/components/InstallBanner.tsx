interface Props {
  canInstall: boolean
  triggerInstall: () => void
  userDismissed: () => void
}

export function InstallBanner({ canInstall, triggerInstall, userDismissed }: Props) {
  if (!canInstall) return null

  return (
    <div className="fixed bottom-20 left-4 right-4 z-50 mx-auto max-w-md">
      <div className="bg-indigo-600 text-white rounded-xl p-4 shadow-lg flex items-center gap-3">
        <div className="flex-1 min-w-0">
          <p className="text-sm font-medium">Instale o Kanso para uma experiência melhor</p>
        </div>
        <button
          onClick={triggerInstall}
          className="shrink-0 bg-white text-indigo-600 px-4 py-1.5 rounded-lg text-sm font-semibold hover:bg-indigo-50 transition-colors"
        >
          Instalar App
        </button>
        <button
          onClick={userDismissed}
          className="shrink-0 text-white/70 hover:text-white text-sm transition-colors"
          aria-label="Fechar"
        >
          Agora não
        </button>
      </div>
    </div>
  )
}
