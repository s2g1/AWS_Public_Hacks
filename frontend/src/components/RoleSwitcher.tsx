import { useAppContext } from '../store/AppContext'

function RoleSwitcher() {
  const { state, switchRole } = useAppContext()
  const isGov = state.currentRole === 'GOV'

  return (
    <div className="fixed top-3 right-3 z-50 flex items-center gap-2 bg-white border border-gray-200 shadow-lg rounded-full px-3 py-1.5">
      <span className={`text-xs font-semibold uppercase tracking-wide ${isGov ? 'text-blue-700' : 'text-gray-400'}`}>
        GOV
      </span>
      <button
        onClick={() => switchRole(isGov ? 'VENDOR' : 'GOV')}
        className={`relative w-11 h-6 rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-offset-1 focus:ring-blue-500 ${
          isGov ? 'bg-blue-600' : 'bg-emerald-600'
        }`}
        aria-label={`Switch to ${isGov ? 'Vendor' : 'Government'} role`}
      >
        <span
          className={`absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow transition-transform ${
            isGov ? 'translate-x-0' : 'translate-x-5'
          }`}
        />
      </button>
      <span className={`text-xs font-semibold uppercase tracking-wide ${!isGov ? 'text-emerald-700' : 'text-gray-400'}`}>
        VENDOR
      </span>
      <span className={`ml-1 px-2 py-0.5 rounded text-xs font-bold text-white ${isGov ? 'bg-blue-600' : 'bg-emerald-600'}`}>
        {isGov ? '🏛️ Gov' : '🏢 Vendor'}
      </span>
    </div>
  )
}

export default RoleSwitcher
