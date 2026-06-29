import { useState } from 'react'

// --- Sample/Mock Data ---

type RiskLevel = 'RED' | 'YELLOW' | 'GREEN'
type CLINStatus = 'ACTIVE' | 'EXERCISED' | 'COMPLETED' | 'OPTION'

interface CLINFinancials {
  clinNumber: string
  description: string
  type: string
  status: CLINStatus
  ceiling: number
  obligated: number
  expended: number
  eac: number
  burnRate: number // average monthly spend (last 3 months)
  riskLevel: RiskLevel
  monthlySpend: number[] // last 6 months for sparkline
}

interface ContractData {
  contractNumber: string
  title: string
  contractor: string
  popStart: string
  popEnd: string
  totalCeiling: number
  totalObligated: number
  totalExpended: number
  totalEAC: number
  clins: CLINFinancials[]
}

const sampleContract: ContractData = {
  contractNumber: 'FA8750-23-C-0042',
  title: 'Advanced AI Research & Development - Phase II SBIR',
  contractor: 'NovaTech Solutions Inc.',
  popStart: '2023-10-01',
  popEnd: '2025-09-30',
  totalCeiling: 4_750_000,
  totalObligated: 3_850_000,
  totalExpended: 2_410_000,
  totalEAC: 4_200_000,
  clins: [
    {
      clinNumber: '0001',
      description: 'Research & Engineering Labor',
      type: 'CPFF',
      status: 'ACTIVE',
      ceiling: 2_500_000,
      obligated: 2_100_000,
      expended: 1_620_000,
      eac: 2_350_000,
      burnRate: 185_000,
      riskLevel: 'YELLOW',
      monthlySpend: [170_000, 175_000, 180_000, 190_000, 185_000, 185_000],
    },
    {
      clinNumber: '0002',
      description: 'Cloud Infrastructure & Compute',
      type: 'FFP',
      status: 'ACTIVE',
      ceiling: 800_000,
      obligated: 750_000,
      expended: 510_000,
      eac: 780_000,
      burnRate: 72_000,
      riskLevel: 'GREEN',
      monthlySpend: [65_000, 68_000, 70_000, 72_000, 74_000, 72_000],
    },
    {
      clinNumber: '0003',
      description: 'Travel & ODCs',
      type: 'T&M',
      status: 'ACTIVE',
      ceiling: 200_000,
      obligated: 200_000,
      expended: 210_000,
      eac: 230_000,
      burnRate: 28_000,
      riskLevel: 'RED',
      monthlySpend: [22_000, 24_000, 26_000, 28_000, 30_000, 28_000],
    },
    {
      clinNumber: '0004',
      description: 'Phase III Option - Production Integration',
      type: 'CPIF',
      status: 'OPTION',
      ceiling: 1_250_000,
      obligated: 800_000,
      expended: 70_000,
      eac: 840_000,
      burnRate: 23_000,
      riskLevel: 'GREEN',
      monthlySpend: [0, 0, 10_000, 20_000, 25_000, 23_000],
    },
  ],
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

function getRiskBadgeClasses(risk: RiskLevel): string {
  switch (risk) {
    case 'RED':
      return 'bg-red-100 text-red-800 border-red-200'
    case 'YELLOW':
      return 'bg-yellow-100 text-yellow-800 border-yellow-200'
    case 'GREEN':
      return 'bg-green-100 text-green-800 border-green-200'
  }
}

function getRiskLabel(risk: RiskLevel): string {
  switch (risk) {
    case 'RED':
      return 'High Risk'
    case 'YELLOW':
      return 'Moderate Risk'
    case 'GREEN':
      return 'Low Risk'
  }
}

function getOverallRisk(clins: CLINFinancials[]): RiskLevel {
  if (clins.some((c) => c.riskLevel === 'RED')) return 'RED'
  if (clins.some((c) => c.riskLevel === 'YELLOW')) return 'YELLOW'
  return 'GREEN'
}

function getProgressPercentage(expended: number, ceiling: number): number {
  if (ceiling === 0) return 0
  return Math.min((expended / ceiling) * 100, 100)
}

function getProgressBarColor(percentage: number): string {
  if (percentage >= 100) return 'bg-red-500'
  if (percentage >= 90) return 'bg-yellow-500'
  return 'bg-blue-500'
}

// --- Sparkline Component ---

function BurnRateSparkline({ data }: { data: number[] }) {
  const max = Math.max(...data, 1)

  return (
    <div className="flex items-end gap-0.5 h-6" aria-label="Burn rate trend">
      {data.map((value, i) => (
        <div
          key={i}
          className="bg-blue-400 rounded-t-sm min-w-[3px] flex-1"
          style={{ height: `${(value / max) * 100}%` }}
          title={formatCurrency(value)}
        />
      ))}
    </div>
  )
}

// --- Financial Card Component ---

interface FinancialCardProps {
  label: string
  value: number
  subtitle?: string
  accent?: string
}

function FinancialCard({ label, value, subtitle, accent = 'border-blue-500' }: FinancialCardProps) {
  return (
    <div className={`bg-white rounded-lg shadow-sm border-l-4 ${accent} p-4 sm:p-5`}>
      <p className="text-sm font-medium text-gray-500 uppercase tracking-wide">{label}</p>
      <p className="mt-1 text-xl sm:text-2xl font-bold text-gray-900">{formatCurrency(value)}</p>
      {subtitle && <p className="mt-1 text-xs text-gray-500">{subtitle}</p>}
    </div>
  )
}

// --- CLIN Row Component ---

interface CLINRowProps {
  clin: CLINFinancials
  isExpanded: boolean
  onToggle: () => void
}

function CLINRow({ clin, isExpanded, onToggle }: CLINRowProps) {
  const progress = getProgressPercentage(clin.expended, clin.ceiling)
  const overrun = Math.max(0, clin.expended - clin.ceiling)
  const underRun = Math.max(0, clin.obligated - clin.eac)

  return (
    <div className="border border-gray-200 rounded-lg mb-2 overflow-hidden">
      {/* Header row - always visible */}
      <button
        onClick={onToggle}
        className="w-full flex items-center justify-between p-3 sm:p-4 text-left hover:bg-gray-50 transition-colors"
        aria-expanded={isExpanded}
      >
        <div className="flex items-center gap-3 min-w-0 flex-1">
          <span className="text-gray-400 transition-transform duration-200" style={{ transform: isExpanded ? 'rotate(90deg)' : 'rotate(0deg)' }}>
            ▶
          </span>
          <div className="min-w-0">
            <div className="flex items-center gap-2 flex-wrap">
              <span className="font-semibold text-gray-900 text-sm sm:text-base">CLIN {clin.clinNumber}</span>
              <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium border ${getRiskBadgeClasses(clin.riskLevel)}`}>
                {clin.riskLevel}
              </span>
              <span className="text-xs text-gray-500 bg-gray-100 px-2 py-0.5 rounded">{clin.status}</span>
            </div>
            <p className="text-xs sm:text-sm text-gray-600 truncate mt-0.5">{clin.description}</p>
          </div>
        </div>
        <div className="text-right hidden sm:block ml-4">
          <p className="text-sm font-medium text-gray-900">{formatCurrency(clin.expended)}</p>
          <p className="text-xs text-gray-500">of {formatCurrency(clin.ceiling)}</p>
        </div>
      </button>

      {/* Expanded details */}
      {isExpanded && (
        <div className="border-t border-gray-200 bg-gray-50 p-4 sm:p-5">
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-4">
            <div>
              <p className="text-xs font-medium text-gray-500 uppercase">Ceiling</p>
              <p className="text-sm font-semibold text-gray-900">{formatCurrency(clin.ceiling)}</p>
            </div>
            <div>
              <p className="text-xs font-medium text-gray-500 uppercase">Obligated</p>
              <p className="text-sm font-semibold text-gray-900">{formatCurrency(clin.obligated)}</p>
            </div>
            <div>
              <p className="text-xs font-medium text-gray-500 uppercase">Expended</p>
              <p className="text-sm font-semibold text-gray-900">{formatCurrency(clin.expended)}</p>
            </div>
            <div>
              <p className="text-xs font-medium text-gray-500 uppercase">EAC</p>
              <p className="text-sm font-semibold text-gray-900">{formatCurrency(clin.eac)}</p>
            </div>
          </div>

          {/* Progress bar */}
          <div className="mb-4">
            <div className="flex justify-between text-xs text-gray-500 mb-1">
              <span>Expenditure Progress</span>
              <span>{progress.toFixed(1)}% of ceiling</span>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-2">
              <div
                className={`${getProgressBarColor(progress)} rounded-full h-2 transition-all`}
                style={{ width: `${Math.min(progress, 100)}%` }}
              />
            </div>
          </div>

          {/* Variance & Burn Rate */}
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div className="bg-white rounded p-3 border border-gray-200">
              <p className="text-xs font-medium text-gray-500 uppercase mb-1">Burn Rate (Monthly Avg)</p>
              <p className="text-sm font-semibold text-gray-900">{formatCurrency(clin.burnRate)}/mo</p>
              <div className="mt-2">
                <BurnRateSparkline data={clin.monthlySpend} />
              </div>
            </div>
            <div className="bg-white rounded p-3 border border-gray-200">
              <p className="text-xs font-medium text-gray-500 uppercase mb-1">Overrun</p>
              <p className={`text-sm font-semibold ${overrun > 0 ? 'text-red-600' : 'text-green-600'}`}>
                {overrun > 0 ? formatCurrency(overrun) : 'None'}
              </p>
            </div>
            <div className="bg-white rounded p-3 border border-gray-200">
              <p className="text-xs font-medium text-gray-500 uppercase mb-1">Under-run</p>
              <p className={`text-sm font-semibold ${underRun > 0 ? 'text-blue-600' : 'text-gray-600'}`}>
                {underRun > 0 ? formatCurrency(underRun) : 'None'}
              </p>
            </div>
          </div>

          {/* Contract type */}
          <div className="mt-3 text-xs text-gray-500">
            Contract Type: <span className="font-medium text-gray-700">{clin.type}</span>
          </div>
        </div>
      )}
    </div>
  )
}

// --- Main Component ---

function Contracts() {
  const [expandedClins, setExpandedClins] = useState<Set<string>>(new Set())
  const contract = sampleContract
  const overallRisk = getOverallRisk(contract.clins)

  function toggleClin(clinNumber: string) {
    setExpandedClins((prev) => {
      const next = new Set(prev)
      if (next.has(clinNumber)) {
        next.delete(clinNumber)
      } else {
        next.add(clinNumber)
      }
      return next
    })
  }

  function expandAll() {
    setExpandedClins(new Set(contract.clins.map((c) => c.clinNumber)))
  }

  function collapseAll() {
    setExpandedClins(new Set())
  }

  return (
    <div className="p-4 sm:p-6 lg:p-8 max-w-7xl mx-auto">
      {/* Header */}
      <div className="mb-6">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Contract Financial Dashboard</h1>
            <p className="mt-1 text-sm text-gray-600">
              {contract.contractNumber} — {contract.title}
            </p>
          </div>
          <div className={`inline-flex items-center px-3 py-1.5 rounded-full text-sm font-medium border ${getRiskBadgeClasses(overallRisk)}`}>
            {getRiskLabel(overallRisk)}
          </div>
        </div>
        <div className="mt-2 flex flex-wrap gap-4 text-xs text-gray-500">
          <span>Contractor: <span className="font-medium text-gray-700">{contract.contractor}</span></span>
          <span>PoP: <span className="font-medium text-gray-700">{contract.popStart} to {contract.popEnd}</span></span>
        </div>
      </div>

      {/* Contract-level financial cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        <FinancialCard
          label="Total Ceiling"
          value={contract.totalCeiling}
          subtitle="Maximum contract value"
          accent="border-gray-500"
        />
        <FinancialCard
          label="Total Obligated"
          value={contract.totalObligated}
          subtitle={`${((contract.totalObligated / contract.totalCeiling) * 100).toFixed(1)}% of ceiling`}
          accent="border-blue-500"
        />
        <FinancialCard
          label="Total Expended"
          value={contract.totalExpended}
          subtitle={`${((contract.totalExpended / contract.totalObligated) * 100).toFixed(1)}% of obligated`}
          accent="border-green-500"
        />
        <FinancialCard
          label="Estimate at Completion"
          value={contract.totalEAC}
          subtitle={contract.totalEAC > contract.totalCeiling ? '⚠️ Exceeds ceiling' : 'Within ceiling'}
          accent={contract.totalEAC > contract.totalCeiling ? 'border-red-500' : 'border-purple-500'}
        />
      </div>

      {/* CLIN-level financials section */}
      <div className="mb-4">
        <div className="flex items-center justify-between mb-3">
          <h2 className="text-lg font-semibold text-gray-900">CLIN Financials</h2>
          <div className="flex gap-2">
            <button
              onClick={expandAll}
              className="text-xs text-blue-600 hover:text-blue-800 font-medium px-2 py-1 rounded hover:bg-blue-50 transition-colors"
            >
              Expand All
            </button>
            <button
              onClick={collapseAll}
              className="text-xs text-gray-600 hover:text-gray-800 font-medium px-2 py-1 rounded hover:bg-gray-100 transition-colors"
            >
              Collapse All
            </button>
          </div>
        </div>

        {/* CLIN rows */}
        <div>
          {contract.clins.map((clin) => (
            <CLINRow
              key={clin.clinNumber}
              clin={clin}
              isExpanded={expandedClins.has(clin.clinNumber)}
              onToggle={() => toggleClin(clin.clinNumber)}
            />
          ))}
        </div>
      </div>

      {/* Summary risk table */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4 sm:p-5">
        <h3 className="text-sm font-semibold text-gray-900 mb-3">Risk Summary</h3>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="text-left text-xs font-medium text-gray-500 uppercase border-b border-gray-200">
                <th className="pb-2 pr-4">CLIN</th>
                <th className="pb-2 pr-4">Risk</th>
                <th className="pb-2 pr-4">Burn Rate</th>
                <th className="pb-2 pr-4 hidden sm:table-cell">% Expended</th>
                <th className="pb-2">Status</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {contract.clins.map((clin) => (
                <tr key={clin.clinNumber}>
                  <td className="py-2 pr-4 font-medium text-gray-900">{clin.clinNumber}</td>
                  <td className="py-2 pr-4">
                    <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium border ${getRiskBadgeClasses(clin.riskLevel)}`}>
                      {clin.riskLevel}
                    </span>
                  </td>
                  <td className="py-2 pr-4 text-gray-700">{formatCurrency(clin.burnRate)}/mo</td>
                  <td className="py-2 pr-4 text-gray-700 hidden sm:table-cell">
                    {((clin.expended / clin.ceiling) * 100).toFixed(1)}%
                  </td>
                  <td className="py-2 text-gray-600">{clin.status}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}

export default Contracts
