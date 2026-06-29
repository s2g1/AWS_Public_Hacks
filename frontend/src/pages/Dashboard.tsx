import { Link } from 'react-router-dom'

// --- Sample Data ---

interface SummaryStats {
  totalContracts: number
  activePayments: number
  pendingREAs: number
  alerts: number
}

interface ActivityItem {
  id: string
  timestamp: string
  description: string
  type: 'payment' | 'contract' | 'alert' | 'rea'
  status: 'success' | 'warning' | 'error' | 'info'
}

interface AgentHealth {
  name: string
  status: 'healthy' | 'degraded' | 'offline'
  lastHeartbeat: string
  throughput: number // messages/min
}

interface ObligationSummary {
  totalCeiling: number
  totalObligated: number
  totalExpended: number
  contractCount: number
}

const sampleStats: SummaryStats = {
  totalContracts: 12,
  activePayments: 34,
  pendingREAs: 5,
  alerts: 3,
}

const sampleActivity: ActivityItem[] = [
  {
    id: '1',
    timestamp: '2024-12-10T14:32:00Z',
    description: 'Payment PAY-2024-0891 disbursed to NovaTech Solutions ($45,200)',
    type: 'payment',
    status: 'success',
  },
  {
    id: '2',
    timestamp: '2024-12-10T14:15:00Z',
    description: 'REA-042 submitted for FA8750-23-C-0042 CLIN 0001 ($125,000)',
    type: 'rea',
    status: 'info',
  },
  {
    id: '3',
    timestamp: '2024-12-10T13:55:00Z',
    description: 'Compliance alert: OFAC screening flagged payee on PAY-2024-0893',
    type: 'alert',
    status: 'error',
  },
  {
    id: '4',
    timestamp: '2024-12-10T13:40:00Z',
    description: 'Option CLIN 0004 exercised on contract FA8750-23-C-0042',
    type: 'contract',
    status: 'success',
  },
  {
    id: '5',
    timestamp: '2024-12-10T13:20:00Z',
    description: 'Validation Agent escalated PAY-2024-0890 (low confidence: 0.62)',
    type: 'payment',
    status: 'warning',
  },
]

const sampleAgents: AgentHealth[] = [
  { name: 'Document Processing', status: 'healthy', lastHeartbeat: '2s ago', throughput: 12 },
  { name: 'Validation', status: 'healthy', lastHeartbeat: '1s ago', throughput: 10 },
  { name: 'Compliance', status: 'healthy', lastHeartbeat: '3s ago', throughput: 9 },
  { name: 'Routing', status: 'degraded', lastHeartbeat: '15s ago', throughput: 4 },
  { name: 'Disbursement', status: 'healthy', lastHeartbeat: '2s ago', throughput: 7 },
]

const sampleObligation: ObligationSummary = {
  totalCeiling: 52_400_000,
  totalObligated: 38_750_000,
  totalExpended: 24_100_000,
  contractCount: 12,
}

// --- Utility Functions ---

function formatCurrency(amount: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(amount)
}

function formatTime(timestamp: string): string {
  const date = new Date(timestamp)
  return date.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })
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

function getActivityIcon(type: ActivityItem['type']): string {
  switch (type) {
    case 'payment':
      return '💳'
    case 'contract':
      return '📄'
    case 'alert':
      return '🚨'
    case 'rea':
      return '📋'
  }
}

function getActivityBorderClass(status: ActivityItem['status']): string {
  switch (status) {
    case 'success':
      return 'border-l-green-500'
    case 'warning':
      return 'border-l-yellow-500'
    case 'error':
      return 'border-l-red-500'
    case 'info':
      return 'border-l-blue-500'
  }
}

// --- Components ---

function StatCard({ label, value, icon, accent }: { label: string; value: number; icon: string; accent: string }) {
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
  const stats = sampleStats
  const activity = sampleActivity
  const agents = sampleAgents
  const obligation = sampleObligation
  const totalThroughput = agents.reduce((sum, a) => sum + a.throughput, 0)

  return (
    <div className="p-4 sm:p-6 lg:p-8 max-w-7xl mx-auto">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
        <p className="mt-1 text-sm text-gray-600">Federal Payment Processing Overview</p>
      </div>

      {/* Summary Statistics Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        <StatCard label="Total Contracts" value={stats.totalContracts} icon="📑" accent="border-blue-500" />
        <StatCard label="Active Payments" value={stats.activePayments} icon="💳" accent="border-green-500" />
        <StatCard label="Pending REAs" value={stats.pendingREAs} icon="📋" accent="border-yellow-500" />
        <StatCard label="Alerts" value={stats.alerts} icon="🔔" accent="border-red-500" />
      </div>

      {/* Main content grid: Activity Feed + Quick Actions / System Status */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
        {/* Recent Activity Feed */}
        <div className="lg:col-span-2 bg-white rounded-lg shadow-sm border border-gray-200 p-4 sm:p-5">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Recent Activity</h2>
          <div className="space-y-3">
            {activity.map((item) => (
              <div
                key={item.id}
                className={`border-l-4 ${getActivityBorderClass(item.status)} bg-gray-50 rounded-r-lg p-3`}
              >
                <div className="flex items-start gap-2">
                  <span className="text-base flex-shrink-0" aria-hidden="true">{getActivityIcon(item.type)}</span>
                  <div className="min-w-0 flex-1">
                    <p className="text-sm text-gray-800">{item.description}</p>
                    <p className="text-xs text-gray-500 mt-1">{formatTime(item.timestamp)}</p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Quick Actions + System Status */}
        <div className="space-y-6">
          {/* Quick Actions */}
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4 sm:p-5">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">Quick Actions</h2>
            <div className="space-y-3">
              <Link
                to="/upload"
                className="flex items-center gap-3 w-full p-3 rounded-lg border border-gray-200 hover:bg-blue-50 hover:border-blue-300 transition-colors text-left"
              >
                <span className="flex items-center justify-center w-9 h-9 rounded-lg bg-blue-100 text-blue-600 text-lg" aria-hidden="true">📤</span>
                <div>
                  <p className="text-sm font-medium text-gray-900">Upload Document</p>
                  <p className="text-xs text-gray-500">Submit a new document for processing</p>
                </div>
              </Link>
              <Link
                to="/contracts"
                className="flex items-center gap-3 w-full p-3 rounded-lg border border-gray-200 hover:bg-green-50 hover:border-green-300 transition-colors text-left"
              >
                <span className="flex items-center justify-center w-9 h-9 rounded-lg bg-green-100 text-green-600 text-lg" aria-hidden="true">📄</span>
                <div>
                  <p className="text-sm font-medium text-gray-900">View Contracts</p>
                  <p className="text-xs text-gray-500">Contract financials and CLIN details</p>
                </div>
              </Link>
              <Link
                to="/payments"
                className="flex items-center gap-3 w-full p-3 rounded-lg border border-gray-200 hover:bg-purple-50 hover:border-purple-300 transition-colors text-left"
              >
                <span className="flex items-center justify-center w-9 h-9 rounded-lg bg-purple-100 text-purple-600 text-lg" aria-hidden="true">💰</span>
                <div>
                  <p className="text-sm font-medium text-gray-900">View Payments</p>
                  <p className="text-xs text-gray-500">Real-time pipeline status</p>
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
                    <span className="hidden sm:inline">{agent.lastHeartbeat}</span>
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
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4 sm:p-6">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 mb-5">
          <div>
            <h2 className="text-lg font-semibold text-gray-900">Obligation Tracking Summary</h2>
            <p className="text-sm text-gray-500">Across {obligation.contractCount} active contracts</p>
          </div>
          <Link
            to="/contracts"
            className="text-sm text-blue-600 hover:text-blue-800 font-medium"
          >
            View Details →
          </Link>
        </div>

        {/* Summary cards */}
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6">
          <div className="text-center p-4 bg-gray-50 rounded-lg border border-gray-200">
            <p className="text-xs font-medium text-gray-500 uppercase tracking-wide">Total Ceiling</p>
            <p className="mt-1 text-xl font-bold text-gray-900">{formatCurrency(obligation.totalCeiling)}</p>
          </div>
          <div className="text-center p-4 bg-blue-50 rounded-lg border border-blue-200">
            <p className="text-xs font-medium text-blue-600 uppercase tracking-wide">Total Obligated</p>
            <p className="mt-1 text-xl font-bold text-blue-900">{formatCurrency(obligation.totalObligated)}</p>
          </div>
          <div className="text-center p-4 bg-green-50 rounded-lg border border-green-200">
            <p className="text-xs font-medium text-green-600 uppercase tracking-wide">Total Expended</p>
            <p className="mt-1 text-xl font-bold text-green-900">{formatCurrency(obligation.totalExpended)}</p>
          </div>
        </div>

        {/* Progress bars */}
        <div className="space-y-4">
          <ObligationBar
            label="Obligated vs Ceiling"
            value={obligation.totalObligated}
            total={obligation.totalCeiling}
            color="bg-blue-500"
          />
          <ObligationBar
            label="Expended vs Obligated"
            value={obligation.totalExpended}
            total={obligation.totalObligated}
            color="bg-green-500"
          />
          <ObligationBar
            label="Expended vs Ceiling"
            value={obligation.totalExpended}
            total={obligation.totalCeiling}
            color="bg-purple-500"
          />
        </div>

        {/* Anti-deficiency warning */}
        {obligation.totalObligated > obligation.totalCeiling && (
          <div className="mt-4 p-3 bg-red-50 border border-red-200 rounded-lg flex items-center gap-2" role="alert">
            <span className="text-red-600 font-bold text-sm">⚠️</span>
            <p className="text-sm text-red-700">Anti-Deficiency Warning: Total obligations exceed contract ceiling.</p>
          </div>
        )}

        {/* Remaining capacity */}
        <div className="mt-4 pt-4 border-t border-gray-200 grid grid-cols-1 sm:grid-cols-2 gap-4 text-sm">
          <div>
            <span className="text-gray-500">Remaining Ceiling Capacity:</span>
            <span className="ml-2 font-semibold text-gray-900">
              {formatCurrency(obligation.totalCeiling - obligation.totalObligated)}
            </span>
          </div>
          <div>
            <span className="text-gray-500">Unspent Obligations:</span>
            <span className="ml-2 font-semibold text-gray-900">
              {formatCurrency(obligation.totalObligated - obligation.totalExpended)}
            </span>
          </div>
        </div>
      </div>
    </div>
  )
}

export default Dashboard
