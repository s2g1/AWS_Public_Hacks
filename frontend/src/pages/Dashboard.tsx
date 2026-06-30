import { useState } from 'react'
import { Link } from 'react-router-dom'
import { useAppContext, Notification, HistoryEntry } from '../store/AppContext'

// --- Types ---

interface AgentHealth {
  name: string
  status: 'healthy' | 'degraded' | 'offline'
  lastHeartbeat: string
  throughput: number
}

const sampleAgents: AgentHealth[] = [
  { name: 'Document Processing', status: 'healthy', lastHeartbeat: '2s ago', throughput: 12 },
  { name: 'Validation', status: 'healthy', lastHeartbeat: '1s ago', throughput: 10 },
  { name: 'Compliance', status: 'healthy', lastHeartbeat: '3s ago', throughput: 9 },
  { name: 'Routing', status: 'degraded', lastHeartbeat: '15s ago', throughput: 4 },
  { name: 'Disbursement', status: 'healthy', lastHeartbeat: '2s ago', throughput: 7 },
]

// --- Utility Functions ---

function formatCurrency(amount: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(amount)
}

function getStatusDotClass(status: AgentHealth['status']): string {
  switch (status) {
    case 'healthy': return 'bg-green-500'
    case 'degraded': return 'bg-yellow-500'
    case 'offline': return 'bg-red-500'
  }
}

function getStatusLabel(status: AgentHealth['status']): string {
  switch (status) {
    case 'healthy': return 'Healthy'
    case 'degraded': return 'Degraded'
    case 'offline': return 'Offline'
  }
}

function getNotifIcon(type: Notification['type']): string {
  switch (type) {
    case 'info': return 'ℹ️'
    case 'action_required': return '⚡'
    case 'success': return '✅'
    case 'warning': return '⚠️'
  }
}

function getActionIcon(action: string): string {
  const lower = action.toLowerCase()
  if (lower.includes('solicitation')) return '📋'
  if (lower.includes('proposal')) return '📝'
  if (lower.includes('invoice')) return '🧾'
  if (lower.includes('payment')) return '💳'
  if (lower.includes('mod')) return '🔧'
  if (lower.includes('approved') || lower.includes('approve')) return '✅'
  if (lower.includes('rejected') || lower.includes('reject')) return '❌'
  return '📄'
}

function formatTimestamp(ts: string): string {
  return new Date(ts).toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

// --- Components ---

function StatCard({ label, value, icon, accent }: { label: string; value: number | string; icon: string; accent: string }) {
  return (
    <div className={`bg-white rounded-lg shadow-sm border-l-4 ${accent} p-4 sm:p-5`}>
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium text-gray-500 uppercase tracking-wide">{label}</p>
          <p className="mt-1 text-2xl sm:text-3xl font-bold text-gray-900">{value}</p>
        </div>
        <span className="text-2xl" aria-hidden="true">{icon}</span>
      </div>
    </div>
  )
}

function ObligationBar({ label, value, total, color }: { label: string; value: number; total: number; color: string }) {
  const pct = total > 0 ? (value / total) * 100 : 0
  return (
    <div>
      <div className="flex items-center justify-between text-sm mb-1">
        <span className="font-medium text-gray-700">{label}</span>
        <span className="text-gray-600">{formatCurrency(value)} ({pct.toFixed(1)}%)</span>
      </div>
      <div className="w-full bg-gray-200 rounded-full h-3">
        <div className={`${color} rounded-full h-3 transition-all`} style={{ width: `${Math.min(pct, 100)}%` }} />
      </div>
    </div>
  )
}

// --- Notifications Section (Enhanced with action buttons) ---

function NotificationsSection({ notifications, onMarkRead, isGov }: { notifications: Notification[]; onMarkRead: (id: string) => void; isGov: boolean }) {
  const [expanded, setExpanded] = useState(true)
  const unreadCount = notifications.filter(n => !n.read).length

  if (notifications.length === 0) return null

  function getActionButtons(notif: Notification) {
    if (notif.type !== 'action_required') return null

    if (isGov) {
      // GOV sees action buttons for proposals, invoices, mods from vendors
      const relatedId = notif.relatedId || ''
      let viewLink = '/contracts'
      if (relatedId.startsWith('prop')) viewLink = '/solicitations'
      else if (relatedId.startsWith('inv')) viewLink = '/payments'
      else if (relatedId.startsWith('mod')) viewLink = '/contracts'
      else if (relatedId.startsWith('contract')) viewLink = '/contracts'

      return (
        <div className="flex items-center gap-2 mt-2">
          <Link to={viewLink} className="px-2.5 py-1 text-xs font-medium rounded bg-blue-100 text-blue-700 hover:bg-blue-200 transition-colors">
            View
          </Link>
          <Link to={viewLink} className="px-2.5 py-1 text-xs font-medium rounded bg-green-100 text-green-700 hover:bg-green-200 transition-colors">
            Approve
          </Link>
          <Link to={viewLink} className="px-2.5 py-1 text-xs font-medium rounded bg-red-100 text-red-700 hover:bg-red-200 transition-colors">
            Reject
          </Link>
        </div>
      )
    } else {
      // VENDOR sees View button linking to relevant page
      const relatedId = notif.relatedId || ''
      let viewLink = '/contracts'
      if (relatedId.startsWith('sol')) viewLink = '/solicitations'
      else if (relatedId.startsWith('contract')) viewLink = '/contracts'

      return (
        <div className="flex items-center gap-2 mt-2">
          <Link to={viewLink} className="px-2.5 py-1 text-xs font-medium rounded bg-blue-100 text-blue-700 hover:bg-blue-200 transition-colors">
            View
          </Link>
        </div>
      )
    }
  }

  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200 mb-6">
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full flex items-center justify-between px-4 sm:px-5 py-3 hover:bg-gray-50 transition-colors"
      >
        <div className="flex items-center gap-2">
          <span className="text-lg" aria-hidden="true">🔔</span>
          <span className="text-sm font-semibold text-gray-900">Notifications</span>
          {unreadCount > 0 && (
            <span className="inline-flex items-center justify-center px-2 py-0.5 rounded-full text-xs font-bold bg-red-500 text-white min-w-[20px]">
              {unreadCount}
            </span>
          )}
        </div>
        <span className="text-gray-400 text-xs">{expanded ? '▼' : '▶'}</span>
      </button>

      {expanded && (
        <div className="border-t border-gray-200 divide-y divide-gray-100 max-h-72 overflow-y-auto">
          {notifications.slice(0, 10).map((notif) => (
            <div
              key={notif.id}
              className={`px-4 sm:px-5 py-3 flex items-start gap-3 ${!notif.read ? 'bg-blue-50/50' : ''}`}
            >
              <span className="text-base flex-shrink-0 mt-0.5" aria-hidden="true">{getNotifIcon(notif.type)}</span>
              <div className="min-w-0 flex-1">
                <p className={`text-sm ${!notif.read ? 'font-medium text-gray-900' : 'text-gray-700'}`}>
                  {notif.message}
                </p>
                <p className="text-xs text-gray-400 mt-0.5">
                  {new Date(notif.createdAt).toLocaleString('en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })}
                </p>
                {getActionButtons(notif)}
              </div>
              {!notif.read && (
                <button
                  onClick={(e) => { e.stopPropagation(); onMarkRead(notif.id) }}
                  className="text-xs text-blue-600 hover:text-blue-800 font-medium flex-shrink-0"
                >
                  Mark read
                </button>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

// --- Quick Actions Section ---

function QuickActionsSection({ actionNotifications, isGov, openSolicitations, totalContracts }: { actionNotifications: Notification[]; isGov: boolean; openSolicitations: number; totalContracts: number }) {
  if (actionNotifications.length === 0) {
    // No pending actions — show fallback links
    return (
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4 sm:p-5">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Quick Actions</h2>
        <p className="text-sm text-gray-500 mb-4">N/A — No pending actions</p>
        <div className="space-y-3">
          <Link
            to="/solicitations"
            className="flex items-center gap-3 w-full p-3 rounded-lg border border-gray-200 hover:bg-blue-50 hover:border-blue-300 transition-colors text-left"
          >
            <span className="flex items-center justify-center w-9 h-9 rounded-lg bg-blue-100 text-blue-600 text-lg" aria-hidden="true">📋</span>
            <div>
              <p className="text-sm font-medium text-gray-900">Solicitations</p>
              <p className="text-xs text-gray-500">{openSolicitations} open opportunities</p>
            </div>
          </Link>
          <Link
            to="/contracts"
            className="flex items-center gap-3 w-full p-3 rounded-lg border border-gray-200 hover:bg-green-50 hover:border-green-300 transition-colors text-left"
          >
            <span className="flex items-center justify-center w-9 h-9 rounded-lg bg-green-100 text-green-600 text-lg" aria-hidden="true">📄</span>
            <div>
              <p className="text-sm font-medium text-gray-900">Contracts</p>
              <p className="text-xs text-gray-500">{totalContracts} active contracts</p>
            </div>
          </Link>
          <Link
            to="/history"
            className="flex items-center gap-3 w-full p-3 rounded-lg border border-gray-200 hover:bg-purple-50 hover:border-purple-300 transition-colors text-left"
          >
            <span className="flex items-center justify-center w-9 h-9 rounded-lg bg-purple-100 text-purple-600 text-lg" aria-hidden="true">📜</span>
            <div>
              <p className="text-sm font-medium text-gray-900">History</p>
              <p className="text-xs text-gray-500">View your action history</p>
            </div>
          </Link>
        </div>
      </div>
    )
  }

  // Show actionable notifications with quick-action buttons
  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4 sm:p-5">
      <h2 className="text-lg font-semibold text-gray-900 mb-4">Quick Actions</h2>
      <div className="space-y-3">
        {actionNotifications.slice(0, 5).map((notif) => {
          const relatedId = notif.relatedId || ''
          let viewLink = '/contracts'
          if (isGov) {
            if (relatedId.startsWith('prop')) viewLink = '/solicitations'
            else if (relatedId.startsWith('inv')) viewLink = '/payments'
            else if (relatedId.startsWith('mod')) viewLink = '/contracts'
            else if (relatedId.startsWith('contract')) viewLink = '/contracts'
          } else {
            if (relatedId.startsWith('sol')) viewLink = '/solicitations'
            else if (relatedId.startsWith('contract')) viewLink = '/contracts'
          }

          return (
            <div key={notif.id} className="p-3 rounded-lg border border-orange-200 bg-orange-50/50">
              <div className="flex items-start gap-2">
                <span className="text-base flex-shrink-0 mt-0.5" aria-hidden="true">⚡</span>
                <div className="min-w-0 flex-1">
                  <p className="text-sm font-medium text-gray-900">{notif.message}</p>
                  <div className="flex items-center gap-2 mt-2">
                    {isGov ? (
                      <>
                        <Link to={viewLink} className="px-2.5 py-1 text-xs font-medium rounded bg-blue-100 text-blue-700 hover:bg-blue-200 transition-colors">
                          View
                        </Link>
                        <Link to={viewLink} className="px-2.5 py-1 text-xs font-medium rounded bg-green-100 text-green-700 hover:bg-green-200 transition-colors">
                          Approve
                        </Link>
                        <Link to={viewLink} className="px-2.5 py-1 text-xs font-medium rounded bg-red-100 text-red-700 hover:bg-red-200 transition-colors">
                          Reject
                        </Link>
                      </>
                    ) : (
                      <Link to={viewLink} className="px-2.5 py-1 text-xs font-medium rounded bg-blue-100 text-blue-700 hover:bg-blue-200 transition-colors">
                        View
                      </Link>
                    )}
                  </div>
                </div>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

// --- Recent Activity Section (from state.history) ---

function RecentActivitySection({ entries }: { entries: HistoryEntry[] }) {
  return (
    <div className="lg:col-span-2 bg-white rounded-lg shadow-sm border border-gray-200 p-4 sm:p-5">
      <h2 className="text-lg font-semibold text-gray-900 mb-4">Recent Activity</h2>
      <div className="space-y-3">
        {entries.length === 0 ? (
          <p className="text-sm text-gray-500 py-4 text-center">No recent activity. Start by creating a solicitation or submitting a proposal.</p>
        ) : (
          entries.map((entry) => (
            <div key={entry.id} className="flex items-start gap-3 p-3 bg-gray-50 rounded-lg border-l-4 border-l-blue-400">
              <span className="text-base flex-shrink-0 mt-0.5" aria-hidden="true">
                {getActionIcon(entry.action)}
              </span>
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2 flex-wrap">
                  <span className="text-sm font-semibold text-gray-900">{entry.action}</span>
                  <span className="text-xs text-gray-400">•</span>
                  <span className="text-xs text-gray-500">{entry.actorName}</span>
                </div>
                <p className="mt-0.5 text-sm text-gray-600">{entry.details}</p>
                <p className="mt-1 text-xs text-gray-400">{formatTimestamp(entry.timestamp)}</p>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  )
}

// --- Main Component ---

function Dashboard() {
  const { state, markNotificationRead } = useAppContext()
  const isGov = state.currentRole === 'GOV'

  // Filter notifications by role and vendor company
  const myNotifications = state.notifications.filter(n => {
    if (n.targetRole !== state.currentRole) return false
    if (state.currentRole === 'VENDOR' && n.targetCompany && n.targetCompany !== state.vendorCompany) return false
    return true
  })

  // Action-required notifications for Quick Actions
  const actionNotifications = myNotifications.filter(n => n.type === 'action_required' && !n.read)

  // Compute real stats from state
  const visibleContracts = isGov
    ? state.contracts
    : state.contracts.filter(c => c.contractor === state.vendorCompany)
  const totalContracts = visibleContracts.length
  const pendingProposals = state.proposals.filter(p => p.status === 'SUBMITTED').length
  const openSolicitations = state.solicitations.filter(s => s.status === 'OPEN').length
  const disbursedPayments = state.payments.filter(p => p.status === 'DISBURSED').length
  const flaggedInvoices = state.invoices.filter(i => i.status === 'FLAGGED').length

  // Obligation totals from visible contracts
  const totalCeiling = visibleContracts.reduce((sum, c) => sum + c.totalCeiling, 0)
  const totalObligated = visibleContracts.reduce((sum, c) => sum + c.totalObligated, 0)
  const totalExpended = visibleContracts.reduce((sum, c) => sum + c.totalExpended, 0)

  // Recent activity from state.history (filtered by current role), last 8, most recent first
  const recentHistoryEntries = state.history
    .filter(h => h.actor === state.currentRole)
    .sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
    .slice(0, 8)

  // System status
  const agents = sampleAgents
  const totalThroughput = agents.reduce((sum, a) => sum + a.throughput, 0)

  return (
    <div className="p-4 sm:p-6 lg:p-8 max-w-7xl mx-auto">
      {/* 1. Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
        <p className="mt-1 text-sm text-gray-600">
          Federal Payment Processing Overview — {isGov ? '🏛️ Government View' : '🏢 Vendor View'}
          {' • '}
          <span className="font-medium text-gray-700">
            {isGov ? 'Gov1 (AFRL)' : `Vendor1 (${state.vendorCompany})`}
          </span>
        </p>
      </div>

      {/* 2. Obligation Tracking Summary (moved up) */}
      {totalContracts > 0 && (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4 sm:p-6 mb-6">
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 mb-5">
            <div>
              <h2 className="text-lg font-semibold text-gray-900">Obligation Tracking Summary</h2>
              <p className="text-sm text-gray-500">Across {totalContracts} active contract{totalContracts !== 1 ? 's' : ''}</p>
            </div>
            <Link to="/contracts" className="text-sm text-blue-600 hover:text-blue-800 font-medium">
              View Details →
            </Link>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6">
            <div className="text-center p-4 bg-gray-50 rounded-lg border border-gray-200">
              <p className="text-xs font-medium text-gray-500 uppercase tracking-wide">Total Ceiling</p>
              <p className="mt-1 text-xl font-bold text-gray-900">{formatCurrency(totalCeiling)}</p>
            </div>
            <div className="text-center p-4 bg-blue-50 rounded-lg border border-blue-200">
              <p className="text-xs font-medium text-blue-600 uppercase tracking-wide">Total Obligated</p>
              <p className="mt-1 text-xl font-bold text-blue-900">{formatCurrency(totalObligated)}</p>
            </div>
            <div className="text-center p-4 bg-green-50 rounded-lg border border-green-200">
              <p className="text-xs font-medium text-green-600 uppercase tracking-wide">Total Expended</p>
              <p className="mt-1 text-xl font-bold text-green-900">{formatCurrency(totalExpended)}</p>
            </div>
          </div>

          <div className="space-y-4">
            <ObligationBar label="Obligated vs Ceiling" value={totalObligated} total={totalCeiling} color="bg-blue-500" />
            <ObligationBar label="Expended vs Obligated" value={totalExpended} total={totalObligated} color="bg-green-500" />
            <ObligationBar label="Expended vs Ceiling" value={totalExpended} total={totalCeiling} color="bg-purple-500" />
          </div>

          {totalObligated > totalCeiling && (
            <div className="mt-4 p-3 bg-red-50 border border-red-200 rounded-lg flex items-center gap-2" role="alert">
              <span className="text-red-600 font-bold text-sm">⚠️</span>
              <p className="text-sm text-red-700">Anti-Deficiency Warning: Total obligations exceed contract ceiling.</p>
            </div>
          )}

          <div className="mt-4 pt-4 border-t border-gray-200 grid grid-cols-1 sm:grid-cols-2 gap-4 text-sm">
            <div>
              <span className="text-gray-500">Remaining Ceiling Capacity:</span>
              <span className="ml-2 font-semibold text-gray-900">{formatCurrency(totalCeiling - totalObligated)}</span>
            </div>
            <div>
              <span className="text-gray-500">Unspent Obligations:</span>
              <span className="ml-2 font-semibold text-gray-900">{formatCurrency(totalObligated - totalExpended)}</span>
            </div>
          </div>
        </div>
      )}

      {/* 3. Notifications (with action buttons) */}
      <NotificationsSection notifications={myNotifications} onMarkRead={markNotificationRead} isGov={isGov} />

      {/* 4. Stats Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        <StatCard label="Total Contracts" value={totalContracts} icon="📑" accent="border-blue-500" />
        <StatCard label="Open Solicitations" value={openSolicitations} icon="📋" accent="border-green-500" />
        <StatCard label="Pending Proposals" value={pendingProposals} icon="📝" accent="border-yellow-500" />
        <StatCard label={flaggedInvoices > 0 ? 'Flagged Invoices' : 'Disbursed Payments'} value={flaggedInvoices > 0 ? flaggedInvoices : disbursedPayments} icon={flaggedInvoices > 0 ? '⚠️' : '💳'} accent={flaggedInvoices > 0 ? 'border-red-500' : 'border-purple-500'} />
      </div>

      {/* 5. Main grid: Recent Activity (left) + Quick Actions (right) */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
        <RecentActivitySection entries={recentHistoryEntries} />
        <QuickActionsSection
          actionNotifications={actionNotifications}
          isGov={isGov}
          openSolicitations={openSolicitations}
          totalContracts={totalContracts}
        />
      </div>

      {/* 6. System Status (very bottom) */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4 sm:p-5">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-gray-900">System Status</h2>
          <span className="text-xs text-gray-500">{totalThroughput} msg/min</span>
        </div>
        <div className="space-y-3">
          {agents.map((agent) => (
            <div key={agent.name} className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <span className={`w-2 h-2 rounded-full ${getStatusDotClass(agent.status)}`} aria-hidden="true" />
                <span className="text-sm text-gray-700">{agent.name}</span>
              </div>
              <div className="flex items-center gap-3 text-xs text-gray-500">
                <span>{agent.throughput} msg/min</span>
                <span className={`px-1.5 py-0.5 rounded text-xs font-medium ${
                  agent.status === 'healthy' ? 'bg-green-100 text-green-700' :
                  agent.status === 'degraded' ? 'bg-yellow-100 text-yellow-700' :
                  'bg-red-100 text-red-700'
                }`}>
                  {getStatusLabel(agent.status)}
                </span>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

export default Dashboard
