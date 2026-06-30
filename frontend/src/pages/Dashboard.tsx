import { Link } from 'react-router-dom'
import { useAppContext } from '../store/AppContext'

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
    case 'healthy':
      return 'bg-green-500'
    case 'degraded':
      return 'bg-yellow-500'
    case 'offline':
      return 'bg-red-500'
  }
}

function getStatusLabel(status: AgentHealth['status']): string {
  switch (status) {
    case 'healthy':
      return 'Healthy'
    case 'degraded':
      return 'Degraded'
    case 'offline':
      return 'Offline'
  }
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
        <div
          className={`${color} rounded-full h-3 transition-all`}
          style={{ width: `${Math.min(pct, 100)}%` }}
        />
      </div>
    </div>
  )
}

// --- Main Component ---

function Dashboard() {
  const { state } = useAppContext()

  // Compute real stats from state
  const totalContracts = state.contracts.length
  const pendingProposals = state.proposals.filter(p => p.status === 'SUBMITTED').length
  const openSolicitations = state.solicitations.filter(s => s.status === 'OPEN').length
  const disbursedPayments = state.payments.filter(p => p.status === 'DISBURSED').length
  const flaggedInvoices = state.invoices.filter(i => i.status === 'FLAGGED').length

  // Obligation totals from contracts
  const totalCeiling = state.contracts.reduce((sum, c) => sum + c.totalCeiling, 0)
  const totalObligated = state.contracts.reduce((sum, c) => sum + c.totalObligated, 0)
  const totalExpended = state.contracts.reduce((sum, c) => sum + c.totalExpended, 0)

  // Recent activity derived from state
  const recentActivity: { id: string; description: string; type: 'payment' | 'contract' | 'alert' | 'info'; status: 'success' | 'warning' | 'error' | 'info' }[] = []

  // Add recent payments
  for (const p of state.payments.slice(-3).reverse()) {
    recentActivity.push({
      id: p.id,
      description: `Payment ${p.id} ${p.status === 'DISBURSED' ? 'disbursed' : 'processing'} (${formatCurrency(p.amount)})`,
      type: 'payment',
      status: 'success',
    })
  }

  // Add flagged invoices
  for (const inv of state.invoices.filter(i => i.status === 'FLAGGED').slice(-2)) {
    recentActivity.push({
      id: inv.id,
      description: `Invoice ${inv.id} flagged: ${inv.complianceIssues?.[0] || 'Compliance issue'}`,
      type: 'alert',
      status: 'warning',
    })
  }

  // Add recent proposals
  for (const prop of state.proposals.filter(p => p.status === 'SUBMITTED').slice(-2)) {
    recentActivity.push({
      id: prop.id,
      description: `Proposal from ${prop.companyName} pending review`,
      type: 'info',
      status: 'info',
    })
  }

  // Add recent contract creations
  for (const c of state.contracts.slice(-2).reverse()) {
    recentActivity.push({
      id: c.id,
      description: `Contract ${c.contractNumber} awarded to ${c.contractor}`,
      type: 'contract',
      status: 'success',
    })
  }

  const agents = sampleAgents
  const totalThroughput = agents.reduce((sum, a) => sum + a.throughput, 0)

  function getActivityIcon(type: string): string {
    switch (type) {
      case 'payment': return '💳'
      case 'contract': return '📄'
      case 'alert': return '🚨'
      default: return '📋'
    }
  }

  function getActivityBorderClass(status: string): string {
    switch (status) {
      case 'success': return 'border-l-green-500'
      case 'warning': return 'border-l-yellow-500'
      case 'error': return 'border-l-red-500'
      default: return 'border-l-blue-500'
    }
  }

  return (
    <div className="p-4 sm:p-6 lg:p-8 max-w-7xl mx-auto">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
        <p className="mt-1 text-sm text-gray-600">
          Federal Payment Processing Overview — {state.currentRole === 'GOV' ? '🏛️ Government View' : '🏢 Vendor View'}
        </p>
      </div>

      {/* Summary Statistics Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        <StatCard label="Total Contracts" value={totalContracts} icon="📑" accent="border-blue-500" />
        <StatCard label="Open Solicitations" value={openSolicitations} icon="📋" accent="border-green-500" />
        <StatCard label="Pending Proposals" value={pendingProposals} icon="📝" accent="border-yellow-500" />
        <StatCard label={flaggedInvoices > 0 ? 'Flagged Invoices' : 'Disbursed Payments'} value={flaggedInvoices > 0 ? flaggedInvoices : disbursedPayments} icon={flaggedInvoices > 0 ? '⚠️' : '💳'} accent={flaggedInvoices > 0 ? 'border-red-500' : 'border-purple-500'} />
      </div>

      {/* Main content grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
        {/* Recent Activity Feed */}
        <div className="lg:col-span-2 bg-white rounded-lg shadow-sm border border-gray-200 p-4 sm:p-5">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Recent Activity</h2>
          <div className="space-y-3">
            {recentActivity.length === 0 ? (
              <p className="text-sm text-gray-500 py-4 text-center">No recent activity. Start by creating a solicitation or submitting a proposal.</p>
            ) : (
              recentActivity.slice(0, 6).map((item) => (
                <div
                  key={item.id}
                  className={`border-l-4 ${getActivityBorderClass(item.status)} bg-gray-50 rounded-r-lg p-3`}
                >
                  <div className="flex items-start gap-2">
                    <span className="text-base flex-shrink-0" aria-hidden="true">{getActivityIcon(item.type)}</span>
                    <p className="text-sm text-gray-800">{item.description}</p>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>

        {/* Quick Actions + System Status */}
        <div className="space-y-6">
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4 sm:p-5">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Quick Actions</h2>
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
                to="/payments"
                className="flex items-center gap-3 w-full p-3 rounded-lg border border-gray-200 hover:bg-purple-50 hover:border-purple-300 transition-colors text-left"
              >
                <span className="flex items-center justify-center w-9 h-9 rounded-lg bg-purple-100 text-purple-600 text-lg" aria-hidden="true">💰</span>
                <div>
                  <p className="text-sm font-medium text-gray-900">Payments</p>
                  <p className="text-xs text-gray-500">{disbursedPayments} disbursed</p>
                </div>
              </Link>
            </div>
          </div>

          {/* System Status */}
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
      </div>

      {/* Obligation Tracking Summary */}
      {totalContracts > 0 && (
        <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4 sm:p-6">
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 mb-5">
            <div>
              <h2 className="text-lg font-semibold text-gray-900">Obligation Tracking Summary</h2>
              <p className="text-sm text-gray-500">Across {totalContracts} active contract{totalContracts !== 1 ? 's' : ''}</p>
            </div>
            <Link to="/contracts" className="text-sm text-blue-600 hover:text-blue-800 font-medium">
              View Details →
            </Link>
          </div>

          {/* Summary cards */}
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

          {/* Progress bars */}
          <div className="space-y-4">
            <ObligationBar label="Obligated vs Ceiling" value={totalObligated} total={totalCeiling} color="bg-blue-500" />
            <ObligationBar label="Expended vs Obligated" value={totalExpended} total={totalObligated} color="bg-green-500" />
            <ObligationBar label="Expended vs Ceiling" value={totalExpended} total={totalCeiling} color="bg-purple-500" />
          </div>

          {/* Anti-deficiency warning */}
          {totalObligated > totalCeiling && (
            <div className="mt-4 p-3 bg-red-50 border border-red-200 rounded-lg flex items-center gap-2" role="alert">
              <span className="text-red-600 font-bold text-sm">⚠️</span>
              <p className="text-sm text-red-700">Anti-Deficiency Warning: Total obligations exceed contract ceiling.</p>
            </div>
          )}

          {/* Remaining capacity */}
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
    </div>
  )
}

export default Dashboard
