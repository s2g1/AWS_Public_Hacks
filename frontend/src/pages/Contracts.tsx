import { useState } from 'react'

// --- Types ---

type RiskLevel = 'RED' | 'YELLOW' | 'GREEN'
type CLINStatus = 'ACTIVE' | 'OPTION' | 'COMPLETED' | 'EXERCISED'

interface CLINFinancials {
  clinNumber: string
  description: string
  type: string
  status: CLINStatus
  ceiling: number
  obligated: number
  expended: number
  eac: number
  burnRate: number
  riskLevel: RiskLevel
  monthlySpend: number[]
  optionDeadline?: string
  note?: string
}

interface SBIRPhase {
  label: string
  status: 'COMPLETED' | 'IN_PROGRESS' | 'PENDING'
  dateRange: string
}

interface ContractData {
  contractNumber: string
  title: string
  contractor: string
  awardingAgency: string
  contractType: string
  parentIDIQ: string
  sbirTopic: string
  sbirTopicTitle: string
  popStart: string
  popEnd: string
  popMonths: number
  monthsElapsed: number
  totalCeiling: number
  totalObligated: number
  totalExpended: number
  totalEAC: number
  overallRisk: RiskLevel
  status: string
  clins: CLINFinancials[]
  sbirPhases: SBIRPhase[]
}

// --- Mock Data: Realistic SBIR IDIQ Task Order ---

const contractData: ContractData = {
  contractNumber: 'FA8750-25-F-0018',
  title: 'Autonomous Payment Processing AI - SBIR Phase II',
  contractor: 'Quantum Federal Systems LLC',
  awardingAgency: 'AFRL / Air Force Research Laboratory',
  contractType: 'Task Order under IDIQ',
  parentIDIQ: 'GS-00F-0001A',
  sbirTopic: 'AF241-0042',
  sbirTopicTitle: 'Agentic AI for Federal Payment Modernization',
  popStart: '2025-01-02',
  popEnd: '2025-07-01',
  popMonths: 6,
  monthsElapsed: 2,
  totalCeiling: 1_249_800,
  totalObligated: 749_880,
  totalExpended: 208_300,
  totalEAC: 1_180_000,
  overallRisk: 'GREEN',
  status: 'ACTIVE - On Track',
  clins: [
    {
      clinNumber: '0001',
      description: 'Phase II Research & Development',
      type: 'CPFF',
      status: 'ACTIVE',
      ceiling: 749_880,
      obligated: 749_880,
      expended: 208_300,
      eac: 720_000,
      burnRate: 104_150,
      riskLevel: 'GREEN',
      monthlySpend: [98_000, 104_150],
    },
    {
      clinNumber: '0002',
      description: 'Phase III Option - Production Pilot',
      type: 'CPFF',
      status: 'OPTION',
      ceiling: 499_920,
      obligated: 0,
      expended: 0,
      eac: 460_000,
      burnRate: 0,
      riskLevel: 'GREEN',
      monthlySpend: [0, 0],
      optionDeadline: '2025-06-01',
      note: 'Option not yet exercised. Exercise deadline: Jun 1, 2025.',
    },
  ],
  sbirPhases: [
    { label: 'Phase I (Feasibility)', status: 'COMPLETED', dateRange: 'Sep 2023 – Mar 2024' },
    { label: 'Phase II (R&D)', status: 'IN_PROGRESS', dateRange: 'Jan 2025 – Jul 2025' },
    { label: 'Phase III (Commercialization)', status: 'PENDING', dateRange: 'Option pending exercise' },
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

function getProgressBarColor(percentage: number): string {
  if (percentage >= 100) return 'bg-red-500'
  if (percentage >= 90) return 'bg-yellow-500'
  return 'bg-blue-500'
}

// --- Sparkline Component ---

function BurnRateSparkline({ data }: { data: number[] }) {
  const max = Math.max(...data, 1)

  return (
    <div className="flex items-end gap-0.5 h-8" aria-label="Burn rate trend">
      {data.map((value, i) => (
        <div
          key={i}
          className="bg-blue-400 rounded-t-sm min-w-[6px] flex-1"
          style={{ height: `${(value / max) * 100}%` }}
          title={formatCurrency(value)}
        />
      ))}
    </div>
  )
}

// --- SBIR Lifecycle Timeline Component ---

function SBIRTimeline({ phases }: { phases: SBIRPhase[] }) {
  function getStatusIcon(status: SBIRPhase['status']) {
    switch (status) {
      case 'COMPLETED':
        return <span className="text-green-600 font-bold">✓</span>
      case 'IN_PROGRESS':
        return <span className="text-blue-600 font-bold">●</span>
      case 'PENDING':
        return <span className="text-gray-400 font-bold">○</span>
    }
  }

  function getStatusLabel(status: SBIRPhase['status']) {
    switch (status) {
      case 'COMPLETED':
        return 'COMPLETED'
      case 'IN_PROGRESS':
        return 'IN PROGRESS'
      case 'PENDING':
        return 'PENDING'
    }
  }

  function getConnectorColor(status: SBIRPhase['status']) {
    switch (status) {
      case 'COMPLETED':
        return 'bg-green-400'
      case 'IN_PROGRESS':
        return 'bg-blue-400'
      case 'PENDING':
        return 'bg-gray-200'
    }
  }

  function getNodeBg(status: SBIRPhase['status']) {
    switch (status) {
      case 'COMPLETED':
        return 'bg-green-100 border-green-400'
      case 'IN_PROGRESS':
        return 'bg-blue-100 border-blue-400 ring-2 ring-blue-200'
      case 'PENDING':
        return 'bg-gray-100 border-gray-300'
    }
  }

  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4 sm:p-6">
      <h3 className="text-sm font-semibold text-gray-900 uppercase tracking-wide mb-4">SBIR Lifecycle Timeline</h3>
      <div className="flex flex-col sm:flex-row items-start sm:items-center gap-2 sm:gap-0">
        {phases.map((phase, idx) => (
          <div key={idx} className="flex items-center flex-1 w-full sm:w-auto">
            {/* Node */}
            <div className={`flex-shrink-0 flex items-center justify-center w-8 h-8 rounded-full border-2 ${getNodeBg(phase.status)}`}>
              {getStatusIcon(phase.status)}
            </div>
            {/* Phase Info */}
            <div className="ml-3 sm:ml-2 min-w-0 flex-shrink-0">
              <p className="text-xs sm:text-sm font-medium text-gray-900 whitespace-nowrap">{phase.label}</p>
              <p className="text-xs text-gray-500">{getStatusLabel(phase.status)}</p>
              <p className="text-xs text-gray-400">{phase.dateRange}</p>
            </div>
            {/* Connector line (not after last) */}
            {idx < phases.length - 1 && (
              <div className={`hidden sm:block flex-1 h-0.5 mx-3 ${getConnectorColor(phases[idx].status)}`} />
            )}
          </div>
        ))}
      </div>
    </div>
  )
}

// --- CLIN Detail Row Component ---

interface CLINRowProps {
  clin: CLINFinancials
  isExpanded: boolean
  onToggle: () => void
}

function CLINRow({ clin, isExpanded, onToggle }: CLINRowProps) {
  const progress = clin.ceiling > 0 ? Math.min((clin.expended / clin.ceiling) * 100, 100) : 0
  const overrun = Math.max(0, clin.expended - clin.ceiling)
  const underRun = Math.max(0, clin.obligated - clin.eac)

  return (
    <div className="border border-gray-200 rounded-lg mb-3 overflow-hidden">
      {/* Header row */}
      <button
        onClick={onToggle}
        className="w-full flex items-center justify-between p-3 sm:p-4 text-left hover:bg-gray-50 transition-colors"
        aria-expanded={isExpanded}
      >
        <div className="flex items-center gap-3 min-w-0 flex-1">
          <span
            className="text-gray-400 transition-transform duration-200 text-xs"
            style={{ transform: isExpanded ? 'rotate(90deg)' : 'rotate(0deg)' }}
          >
            ▶
          </span>
          <div className="min-w-0">
            <div className="flex items-center gap-2 flex-wrap">
              <span className="font-semibold text-gray-900 text-sm sm:text-base">CLIN {clin.clinNumber}</span>
              <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium border ${getRiskBadgeClasses(clin.riskLevel)}`}>
                {clin.riskLevel}
              </span>
              <span className="text-xs text-gray-500 bg-gray-100 px-2 py-0.5 rounded">{clin.status}</span>
              <span className="text-xs text-gray-400 bg-gray-50 px-2 py-0.5 rounded">{clin.type}</span>
            </div>
            <p className="text-xs sm:text-sm text-gray-600 truncate mt-0.5">{clin.description}</p>
          </div>
        </div>
        <div className="text-right hidden sm:block ml-4">
          <p className="text-sm font-medium text-gray-900">{formatCurrency(clin.expended)}</p>
          <p className="text-xs text-gray-500">of {formatCurrency(clin.ceiling)} ceiling</p>
        </div>
      </button>

      {/* Expanded details */}
      {isExpanded && (
        <div className="border-t border-gray-200 bg-gray-50 p-4 sm:p-5">
          {/* Financial grid */}
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 mb-4">
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

          {/* Burn Rate & Variance */}
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div className="bg-white rounded p-3 border border-gray-200">
              <p className="text-xs font-medium text-gray-500 uppercase mb-1">Burn Rate</p>
              <p className="text-sm font-semibold text-gray-900">{formatCurrency(clin.burnRate)}/mo</p>
              {clin.burnRate > 0 && (
                <div className="mt-2">
                  <BurnRateSparkline data={clin.monthlySpend} />
                </div>
              )}
            </div>
            <div className="bg-white rounded p-3 border border-gray-200">
              <p className="text-xs font-medium text-gray-500 uppercase mb-1">Overrun</p>
              <p className={`text-sm font-semibold ${overrun > 0 ? 'text-red-600' : 'text-green-600'}`}>
                {overrun > 0 ? formatCurrency(overrun) : 'None'}
              </p>
            </div>
            <div className="bg-white rounded p-3 border border-gray-200">
              <p className="text-xs font-medium text-gray-500 uppercase mb-1">Under-run (Projected)</p>
              <p className={`text-sm font-semibold ${underRun > 0 ? 'text-blue-600' : 'text-gray-600'}`}>
                {underRun > 0 ? formatCurrency(underRun) : 'N/A'}
              </p>
            </div>
          </div>

          {/* Note if present */}
          {clin.note && (
            <div className="mt-3 text-xs text-amber-700 bg-amber-50 border border-amber-200 rounded px-3 py-2">
              ℹ️ {clin.note}
            </div>
          )}
        </div>
      )}
    </div>
  )
}

// --- Contract Summary Card Component ---

function ContractSummaryCard({ contract, onViewClins }: { contract: ContractData; onViewClins: () => void }) {
  const popProgress = (contract.monthsElapsed / contract.popMonths) * 100
  const expendedPct = ((contract.totalExpended / contract.totalCeiling) * 100).toFixed(1)
  const obligatedPct = ((contract.totalObligated / contract.totalCeiling) * 100).toFixed(1)

  return (
    <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
      {/* Top status bar */}
      <div className="bg-slate-800 px-4 sm:px-6 py-3 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <span className="text-white font-mono text-sm sm:text-base font-bold">{contract.contractNumber}</span>
          <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold border ${getRiskBadgeClasses(contract.overallRisk)}`}>
            {contract.overallRisk}
          </span>
        </div>
        <span className="text-emerald-400 text-xs sm:text-sm font-medium">{contract.status}</span>
      </div>

      <div className="p-4 sm:p-6">
        {/* Title and key info */}
        <h2 className="text-lg sm:text-xl font-bold text-gray-900 mb-1">{contract.title}</h2>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-x-8 gap-y-2 mt-3 text-sm">
          <div className="flex gap-2">
            <span className="text-gray-500 min-w-[100px]">Contractor:</span>
            <span className="font-medium text-gray-900">{contract.contractor}</span>
          </div>
          <div className="flex gap-2">
            <span className="text-gray-500 min-w-[100px]">Agency:</span>
            <span className="font-medium text-gray-900">{contract.awardingAgency}</span>
          </div>
          <div className="flex gap-2">
            <span className="text-gray-500 min-w-[100px]">Contract Type:</span>
            <span className="font-medium text-gray-900">{contract.contractType}</span>
          </div>
          <div className="flex gap-2">
            <span className="text-gray-500 min-w-[100px]">Parent IDIQ:</span>
            <span className="font-mono text-sm text-blue-700">{contract.parentIDIQ}</span>
          </div>
          <div className="flex gap-2">
            <span className="text-gray-500 min-w-[100px]">SBIR Topic:</span>
            <span className="font-medium text-gray-900">{contract.sbirTopic} <span className="text-gray-500 text-xs">("{contract.sbirTopicTitle}")</span></span>
          </div>
        </div>

        {/* Period of Performance */}
        <div className="mt-5 p-3 bg-slate-50 rounded-lg border border-slate-200">
          <div className="flex items-center justify-between mb-2">
            <span className="text-xs font-semibold text-gray-700 uppercase tracking-wide">Period of Performance</span>
            <span className="text-xs text-gray-500">
              {contract.monthsElapsed} of {contract.popMonths} months ({popProgress.toFixed(0)}%)
            </span>
          </div>
          <div className="w-full bg-gray-200 rounded-full h-2.5 mb-2">
            <div
              className="bg-blue-600 rounded-full h-2.5 transition-all"
              style={{ width: `${popProgress}%` }}
            />
          </div>
          <div className="flex justify-between text-xs text-gray-500">
            <span>{contract.popStart}</span>
            <span>{contract.popEnd}</span>
          </div>
        </div>

        {/* Financial Summary */}
        <div className="mt-5">
          <h3 className="text-xs font-semibold text-gray-700 uppercase tracking-wide mb-3">Financial Summary</h3>
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
            <div className="bg-gray-50 rounded-lg p-3 border border-gray-200">
              <p className="text-xs text-gray-500 font-medium">Ceiling</p>
              <p className="text-base sm:text-lg font-bold text-gray-900">{formatCurrency(contract.totalCeiling)}</p>
            </div>
            <div className="bg-gray-50 rounded-lg p-3 border border-gray-200">
              <p className="text-xs text-gray-500 font-medium">Obligated</p>
              <p className="text-base sm:text-lg font-bold text-gray-900">{formatCurrency(contract.totalObligated)}</p>
              <p className="text-xs text-gray-400">{obligatedPct}% of ceiling</p>
            </div>
            <div className="bg-gray-50 rounded-lg p-3 border border-gray-200">
              <p className="text-xs text-gray-500 font-medium">Expended</p>
              <p className="text-base sm:text-lg font-bold text-gray-900">{formatCurrency(contract.totalExpended)}</p>
              <p className="text-xs text-gray-400">{expendedPct}% of ceiling</p>
            </div>
            <div className="bg-gray-50 rounded-lg p-3 border border-gray-200">
              <p className="text-xs text-gray-500 font-medium">EAC</p>
              <p className="text-base sm:text-lg font-bold text-gray-900">{formatCurrency(contract.totalEAC)}</p>
              <p className="text-xs text-green-600">Under ceiling</p>
            </div>
          </div>
        </div>

        {/* View CLINs CTA */}
        <div className="mt-6 flex justify-center">
          <button
            onClick={onViewClins}
            className="inline-flex items-center gap-2 px-5 py-2.5 bg-slate-800 text-white text-sm font-medium rounded-lg hover:bg-slate-700 transition-colors focus:outline-none focus:ring-2 focus:ring-slate-500 focus:ring-offset-2"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
            </svg>
            View CLINs
          </button>
        </div>
      </div>
    </div>
  )
}

// --- Main Component ---

function Contracts() {
  const [showClins, setShowClins] = useState(false)
  const [expandedClins, setExpandedClins] = useState<Set<string>>(new Set())
  const contract = contractData

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
      {/* Page Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Contract Management</h1>
        <p className="mt-1 text-sm text-gray-500">Federal contract oversight and CLIN-level financial tracking</p>
      </div>

      {/* Contract Summary Card */}
      <div className="mb-6">
        <ContractSummaryCard contract={contract} onViewClins={() => setShowClins(!showClins)} />
      </div>

      {/* SBIR Lifecycle Timeline */}
      <div className="mb-6">
        <SBIRTimeline phases={contract.sbirPhases} />
      </div>

      {/* CLIN Details (shown after clicking View CLINs) */}
      {showClins && (
        <div className="animate-in fade-in">
          <div className="flex items-center justify-between mb-3">
            <h2 className="text-lg font-semibold text-gray-900">CLIN Details</h2>
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
              <button
                onClick={() => setShowClins(false)}
                className="text-xs text-gray-500 hover:text-gray-700 font-medium px-2 py-1 rounded hover:bg-gray-100 transition-colors"
              >
                Hide CLINs
              </button>
            </div>
          </div>

          {/* CLIN Rows */}
          <div className="mb-6">
            {contract.clins.map((clin) => (
              <CLINRow
                key={clin.clinNumber}
                clin={clin}
                isExpanded={expandedClins.has(clin.clinNumber)}
                onToggle={() => toggleClin(clin.clinNumber)}
              />
            ))}
          </div>

          {/* Risk Summary Table */}
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4 sm:p-5">
            <h3 className="text-sm font-semibold text-gray-900 mb-3 uppercase tracking-wide">CLIN Risk Summary</h3>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="text-left text-xs font-medium text-gray-500 uppercase border-b border-gray-200">
                    <th className="pb-2 pr-4">CLIN</th>
                    <th className="pb-2 pr-4">Description</th>
                    <th className="pb-2 pr-4">Type</th>
                    <th className="pb-2 pr-4">Risk</th>
                    <th className="pb-2 pr-4">Burn Rate</th>
                    <th className="pb-2 pr-4 hidden sm:table-cell">% Expended</th>
                    <th className="pb-2">Status</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100">
                  {contract.clins.map((clin) => (
                    <tr key={clin.clinNumber}>
                      <td className="py-2 pr-4 font-mono font-medium text-gray-900">{clin.clinNumber}</td>
                      <td className="py-2 pr-4 text-gray-700 text-xs max-w-[200px] truncate">{clin.description}</td>
                      <td className="py-2 pr-4 text-gray-600 text-xs">{clin.type}</td>
                      <td className="py-2 pr-4">
                        <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium border ${getRiskBadgeClasses(clin.riskLevel)}`}>
                          {clin.riskLevel}
                        </span>
                      </td>
                      <td className="py-2 pr-4 text-gray-700">{clin.burnRate > 0 ? `${formatCurrency(clin.burnRate)}/mo` : '—'}</td>
                      <td className="py-2 pr-4 text-gray-700 hidden sm:table-cell">
                        {clin.ceiling > 0 ? `${((clin.expended / clin.ceiling) * 100).toFixed(1)}%` : '—'}
                      </td>
                      <td className="py-2 text-gray-600 text-xs">{clin.status}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default Contracts
