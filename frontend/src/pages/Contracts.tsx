import { useState, useMemo } from 'react'
import { useAppContext, Contract, Invoice, CLINItem, ContractMod } from '../store/AppContext'

// --- Types ---

type RiskLevel = 'RED' | 'YELLOW' | 'GREEN'

interface SBIRPhase {
  label: string
  status: 'COMPLETED' | 'IN_PROGRESS' | 'PENDING'
  dateRange: string
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
    case 'RED': return 'bg-red-100 text-red-800 border-red-200'
    case 'YELLOW': return 'bg-yellow-100 text-yellow-800 border-yellow-200'
    case 'GREEN': return 'bg-green-100 text-green-800 border-green-200'
  }
}

function getProgressBarColor(percentage: number): string {
  if (percentage >= 100) return 'bg-red-500'
  if (percentage >= 90) return 'bg-yellow-500'
  return 'bg-blue-500'
}

function getClinRisk(clin: CLINItem): RiskLevel {
  if (clin.ceiling === 0) return 'GREEN'
  const pct = clin.expended / clin.ceiling
  if (pct >= 0.9) return 'RED'
  if (pct >= 0.7) return 'YELLOW'
  return 'GREEN'
}

// --- SBIR Lifecycle Timeline Component ---

const defaultSBIRPhases: SBIRPhase[] = [
  { label: 'Phase I (Feasibility)', status: 'COMPLETED', dateRange: 'Sep 2023 – Mar 2024' },
  { label: 'Phase II (R&D)', status: 'IN_PROGRESS', dateRange: 'Jan 2025 – Jul 2025' },
  { label: 'Phase III (Commercialization)', status: 'PENDING', dateRange: 'Option pending exercise' },
]

function SBIRTimeline({ phases }: { phases: SBIRPhase[] }) {
  function getStatusIcon(status: SBIRPhase['status']) {
    switch (status) {
      case 'COMPLETED': return <span className="text-green-600 font-bold">✓</span>
      case 'IN_PROGRESS': return <span className="text-blue-600 font-bold">●</span>
      case 'PENDING': return <span className="text-gray-400 font-bold">○</span>
    }
  }
  function getStatusLabel(status: SBIRPhase['status']) {
    switch (status) {
      case 'COMPLETED': return 'COMPLETED'
      case 'IN_PROGRESS': return 'IN PROGRESS'
      case 'PENDING': return 'PENDING'
    }
  }
  function getConnectorColor(status: SBIRPhase['status']) {
    switch (status) {
      case 'COMPLETED': return 'bg-green-400'
      case 'IN_PROGRESS': return 'bg-blue-400'
      case 'PENDING': return 'bg-gray-200'
    }
  }
  function getNodeBg(status: SBIRPhase['status']) {
    switch (status) {
      case 'COMPLETED': return 'bg-green-100 border-green-400'
      case 'IN_PROGRESS': return 'bg-blue-100 border-blue-400 ring-2 ring-blue-200'
      case 'PENDING': return 'bg-gray-100 border-gray-300'
    }
  }

  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4 sm:p-6">
      <h3 className="text-sm font-semibold text-gray-900 uppercase tracking-wide mb-4">SBIR Lifecycle Timeline</h3>
      <div className="flex flex-col sm:flex-row items-start sm:items-center gap-2 sm:gap-0">
        {phases.map((phase, idx) => (
          <div key={idx} className="flex items-center flex-1 w-full sm:w-auto">
            <div className={`flex-shrink-0 flex items-center justify-center w-8 h-8 rounded-full border-2 ${getNodeBg(phase.status)}`}>
              {getStatusIcon(phase.status)}
            </div>
            <div className="ml-3 sm:ml-2 min-w-0 flex-shrink-0">
              <p className="text-xs sm:text-sm font-medium text-gray-900 whitespace-nowrap">{phase.label}</p>
              <p className="text-xs text-gray-500">{getStatusLabel(phase.status)}</p>
              <p className="text-xs text-gray-400">{phase.dateRange}</p>
            </div>
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

function CLINRow({ clin, isExpanded, onToggle }: { clin: CLINItem; isExpanded: boolean; onToggle: () => void }) {
  const progress = clin.ceiling > 0 ? Math.min((clin.expended / clin.ceiling) * 100, 100) : 0
  const risk = getClinRisk(clin)

  return (
    <div className="border border-gray-200 rounded-lg mb-3 overflow-hidden">
      <button
        onClick={onToggle}
        className="w-full flex items-center justify-between p-3 sm:p-4 text-left hover:bg-gray-50 transition-colors"
        aria-expanded={isExpanded}
      >
        <div className="flex items-center gap-3 min-w-0 flex-1">
          <span className="text-gray-400 transition-transform duration-200 text-xs" style={{ transform: isExpanded ? 'rotate(90deg)' : 'rotate(0deg)' }}>▶</span>
          <div className="min-w-0">
            <div className="flex items-center gap-2 flex-wrap">
              <span className="font-semibold text-gray-900 text-sm sm:text-base">CLIN {clin.clinNumber}</span>
              <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium border ${getRiskBadgeClasses(risk)}`}>{risk}</span>
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

      {isExpanded && (
        <div className="border-t border-gray-200 bg-gray-50 p-4 sm:p-5">
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 mb-4">
            <div><p className="text-xs font-medium text-gray-500 uppercase">Ceiling</p><p className="text-sm font-semibold text-gray-900">{formatCurrency(clin.ceiling)}</p></div>
            <div><p className="text-xs font-medium text-gray-500 uppercase">Obligated</p><p className="text-sm font-semibold text-gray-900">{formatCurrency(clin.obligated)}</p></div>
            <div><p className="text-xs font-medium text-gray-500 uppercase">Expended</p><p className="text-sm font-semibold text-gray-900">{formatCurrency(clin.expended)}</p></div>
            <div><p className="text-xs font-medium text-gray-500 uppercase">Remaining</p><p className="text-sm font-semibold text-gray-900">{formatCurrency(clin.ceiling - clin.expended)}</p></div>
          </div>
          <div className="mb-4">
            <div className="flex justify-between text-xs text-gray-500 mb-1">
              <span>Expenditure Progress</span>
              <span>{progress.toFixed(1)}% of ceiling</span>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-2">
              <div className={`${getProgressBarColor(progress)} rounded-full h-2 transition-all`} style={{ width: `${Math.min(progress, 100)}%` }} />
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

// --- Invoice Submission Form ---

function InvoiceForm({ contract, onSubmit, onClose }: { contract: Contract; onSubmit: (clinNumber: string, amount: number, description: string) => void; onClose: () => void }) {
  const [activeTab, setActiveTab] = useState<'manual' | 'upload'>('manual')
  const [clinNumber, setClinNumber] = useState(contract.clins[0]?.clinNumber || '')
  const [amount, setAmount] = useState('')
  const [description, setDescription] = useState('')
  const [error, setError] = useState('')
  const [uploadedFile, setUploadedFile] = useState<string>('')

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const parsedAmount = parseFloat(amount.replace(/,/g, ''))
    if (!parsedAmount || parsedAmount <= 0) { setError('Please enter a valid amount'); return }
    if (!description.trim()) { setError('Please enter a description'); return }
    onSubmit(clinNumber, parsedAmount, description)
  }

  function handleFileUpload(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    if (file) {
      setUploadedFile(file.name)
      // Simulate OCR extraction
      const simulatedAmount = Math.floor(Math.random() * 50000) + 25000
      setAmount(String(simulatedAmount))
      setDescription(`Extracted from ${file.name}`)
      setError('')
    }
  }

  function handleFileSubmit() {
    if (!uploadedFile) { setError('Please upload a file'); return }
    const parsedAmount = parseFloat(amount.replace(/,/g, ''))
    if (!parsedAmount || parsedAmount <= 0) { setError('Invalid extracted amount'); return }
    onSubmit(clinNumber, parsedAmount, description || `Extracted from ${uploadedFile}`)
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div className="bg-white rounded-xl shadow-xl w-full max-w-md">
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">Submit Invoice</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600">
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>
          </button>
        </div>

        {/* Tab Switcher */}
        <div className="flex border-b border-gray-200">
          <button
            onClick={() => setActiveTab('manual')}
            className={`flex-1 py-3 text-sm font-medium text-center transition-colors ${activeTab === 'manual' ? 'text-blue-600 border-b-2 border-blue-600' : 'text-gray-500 hover:text-gray-700'}`}
          >
            Manual Input
          </button>
          <button
            onClick={() => setActiveTab('upload')}
            className={`flex-1 py-3 text-sm font-medium text-center transition-colors ${activeTab === 'upload' ? 'text-blue-600 border-b-2 border-blue-600' : 'text-gray-500 hover:text-gray-700'}`}
          >
            File Upload
          </button>
        </div>

        {activeTab === 'manual' ? (
          <form onSubmit={handleSubmit} className="p-6 space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">CLIN</label>
              <select value={clinNumber} onChange={(e) => setClinNumber(e.target.value)} className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm">
                {contract.clins.map(c => (
                  <option key={c.clinNumber} value={c.clinNumber}>CLIN {c.clinNumber} - {c.description} (Remaining: {formatCurrency(c.ceiling - c.expended)})</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Amount ($)</label>
              <input type="text" value={amount} onChange={(e) => setAmount(e.target.value)} className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm" placeholder="e.g., 50000" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
              <textarea value={description} onChange={(e) => setDescription(e.target.value)} rows={3} className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm" placeholder="Describe the work performed..." />
            </div>
            {error && <p className="text-xs text-red-600">{error}</p>}
            <div className="flex justify-end gap-3 pt-2">
              <button type="button" onClick={onClose} className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50">Cancel</button>
              <button type="submit" className="px-4 py-2 text-sm font-medium text-white bg-green-600 rounded-lg hover:bg-green-700">Submit Invoice</button>
            </div>
          </form>
        ) : (
          <div className="p-6 space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">CLIN</label>
              <select value={clinNumber} onChange={(e) => setClinNumber(e.target.value)} className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm">
                {contract.clins.map(c => (
                  <option key={c.clinNumber} value={c.clinNumber}>CLIN {c.clinNumber} - {c.description} (Remaining: {formatCurrency(c.ceiling - c.expended)})</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Upload Invoice</label>
              <input
                type="file"
                accept=".pdf,.doc,.docx,.txt,.zip"
                onChange={handleFileUpload}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm file:mr-4 file:py-1 file:px-3 file:rounded file:border-0 file:text-sm file:font-medium file:bg-green-50 file:text-green-700 hover:file:bg-green-100"
              />
              {uploadedFile && (
                <p className="mt-1 text-xs text-green-600">📎 {uploadedFile}</p>
              )}
            </div>
            {uploadedFile && (
              <>
                <div className="bg-blue-50 border border-blue-200 rounded-lg p-3">
                  <p className="text-xs font-medium text-blue-800 mb-1">🤖 Simulated OCR Extraction:</p>
                  <p className="text-sm text-blue-900">Amount: <span className="font-semibold">${parseFloat(amount).toLocaleString()}</span></p>
                  <p className="text-xs text-blue-700 mt-1">Description: {description}</p>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Extracted Amount ($)</label>
                  <input type="text" value={amount} onChange={(e) => setAmount(e.target.value)} className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm" />
                </div>
              </>
            )}
            {error && <p className="text-xs text-red-600">{error}</p>}
            <div className="flex justify-end gap-3 pt-2">
              <button type="button" onClick={onClose} className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50">Cancel</button>
              <button
                type="button"
                onClick={handleFileSubmit}
                disabled={!uploadedFile}
                className="px-4 py-2 text-sm font-medium text-white bg-green-600 rounded-lg hover:bg-green-700 disabled:bg-gray-300 disabled:cursor-not-allowed"
              >
                Submit Invoice
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

// --- Contract Mod Form ---

function ContractModForm({ isGov, onSubmit, onClose }: { contract: Contract; isGov: boolean; onSubmit: (type: 'REA' | 'ECP' | 'GOV_MOD', title: string, description: string, amount: number) => void; onClose: () => void }) {
  const [modType, setModType] = useState<'REA' | 'ECP' | 'GOV_MOD'>(isGov ? 'GOV_MOD' : 'REA')
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [amount, setAmount] = useState('')
  const [error, setError] = useState('')

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!title.trim()) { setError('Title is required'); return }
    if (!description.trim()) { setError('Description is required'); return }
    const parsedAmount = parseFloat(amount.replace(/,/g, ''))
    if (!parsedAmount || parsedAmount <= 0) { setError('Please enter a valid amount'); return }
    onSubmit(modType, title, description, parsedAmount)
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div className="bg-white rounded-xl shadow-xl w-full max-w-md">
        <div className="flex items-center justify-between px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">{isGov ? 'Issue Contract Mod' : 'Request Contract Mod'}</h2>
          <button onClick={onClose} className="text-gray-400 hover:text-gray-600">
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>
          </button>
        </div>
        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Type</label>
            <select value={modType} onChange={(e) => setModType(e.target.value as 'REA' | 'ECP' | 'GOV_MOD')} className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm">
              {isGov ? (
                <option value="GOV_MOD">Government Modification</option>
              ) : (
                <>
                  <option value="REA">REA (Request for Equitable Adjustment)</option>
                  <option value="ECP">ECP (Engineering Change Proposal)</option>
                </>
              )}
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Title</label>
            <input type="text" value={title} onChange={(e) => setTitle(e.target.value)} className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm" placeholder="Mod title..." />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
            <textarea value={description} onChange={(e) => setDescription(e.target.value)} rows={3} className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm" placeholder="Describe the modification..." />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Amount ($)</label>
            <input type="text" value={amount} onChange={(e) => setAmount(e.target.value)} className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm" placeholder="e.g., 50000" />
          </div>
          {error && <p className="text-xs text-red-600">{error}</p>}
          <div className="flex justify-end gap-3 pt-2">
            <button type="button" onClick={onClose} className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50">Cancel</button>
            <button type="submit" className="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700">Submit</button>
          </div>
        </form>
      </div>
    </div>
  )
}

// --- Pending Mods Panel ---

function PendingModsPanel({ mods, isGov, onApprove, onReject }: { mods: ContractMod[]; contractNumber: string; isGov: boolean; onApprove: (id: string) => void; onReject: (id: string) => void }) {
  const pending = mods.filter(m => m.status === 'SUBMITTED' || m.status === 'UNDER_REVIEW')
  if (pending.length === 0) return null

  function getActionOwner(mod: ContractMod): string {
    if (mod.requestedBy === 'VENDOR') return 'Pending: GOV Review'
    return 'Pending: Vendor Response'
  }

  return (
    <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-4">
      <h4 className="text-sm font-semibold text-blue-900 mb-2">📋 Pending Modifications ({pending.length})</h4>
      <div className="space-y-2">
        {pending.map(mod => (
          <div key={mod.id} className="bg-white border border-blue-200 rounded-lg p-3">
            <div className="flex items-center justify-between mb-1">
              <span className="text-sm font-medium text-gray-900">{mod.title}</span>
              <div className="flex items-center gap-2">
                <span className="text-xs px-2 py-0.5 rounded-full bg-blue-100 text-blue-800 font-medium">{mod.type}</span>
                <span className="text-xs px-2 py-0.5 rounded-full bg-amber-100 text-amber-800 font-medium">{getActionOwner(mod)}</span>
              </div>
            </div>
            <p className="text-xs text-gray-600 mb-1">{mod.description}</p>
            <div className="flex items-center justify-between">
              <span className="text-xs text-gray-500">Amount: {formatCurrency(mod.amount)} • {new Date(mod.submittedAt).toLocaleDateString()}</span>
              {isGov && mod.requestedBy === 'VENDOR' && (
                <div className="flex gap-2">
                  <button onClick={() => onApprove(mod.id)} className="px-2 py-1 text-xs font-medium text-white bg-green-600 rounded hover:bg-green-700">Approve</button>
                  <button onClick={() => onReject(mod.id)} className="px-2 py-1 text-xs font-medium text-red-700 bg-red-50 border border-red-200 rounded hover:bg-red-100">Reject</button>
                </div>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

// --- Flagged Invoices Panel (GOV) ---

function FlaggedInvoicesPanel({ invoices, onApprove, onReject }: { invoices: Invoice[]; onApprove: (id: string, justification: string) => void; onReject: (id: string, reason: string) => void }) {
  const [justification, setJustification] = useState<Record<string, string>>({})
  const flagged = invoices.filter(i => i.status === 'FLAGGED')
  if (flagged.length === 0) return null

  return (
    <div className="bg-amber-50 border border-amber-200 rounded-lg p-4 sm:p-5 mb-6">
      <h3 className="text-sm font-semibold text-amber-900 uppercase tracking-wide mb-3">⚠️ Flagged Invoices Requiring Review ({flagged.length})</h3>
      <div className="space-y-3">
        {flagged.map(invoice => (
          <div key={invoice.id} className="bg-white border border-amber-200 rounded-lg p-4">
            <div className="flex items-center justify-between mb-2">
              <span className="font-medium text-gray-900">Invoice {invoice.id}</span>
              <span className="text-sm font-semibold text-amber-700">{formatCurrency(invoice.amount)}</span>
            </div>
            <p className="text-sm text-gray-600 mb-2">{invoice.description}</p>
            <p className="text-xs text-gray-500 mb-2">CLIN: {invoice.clinNumber} | Submitted: {new Date(invoice.submittedAt).toLocaleDateString()}</p>
            {invoice.complianceIssues && (
              <div className="mb-3">
                {invoice.complianceIssues.map((issue, idx) => (
                  <p key={idx} className="text-xs text-red-700 bg-red-50 border border-red-100 rounded px-2 py-1 mb-1">⚠️ {issue}</p>
                ))}
              </div>
            )}
            <div className="space-y-2">
              <textarea value={justification[invoice.id] || ''} onChange={(e) => setJustification(prev => ({ ...prev, [invoice.id]: e.target.value }))} placeholder="Enter justification or rejection reason..." rows={2} className="w-full px-3 py-2 border border-gray-300 rounded-lg text-xs" />
              <div className="flex gap-2">
                <button onClick={() => onApprove(invoice.id, justification[invoice.id] || '')} className="px-3 py-1.5 text-xs font-medium text-white bg-green-600 rounded-lg hover:bg-green-700">Approve & Disburse</button>
                <button onClick={() => onReject(invoice.id, justification[invoice.id] || 'Rejected by reviewer')} className="px-3 py-1.5 text-xs font-medium text-red-700 bg-red-50 border border-red-200 rounded-lg hover:bg-red-100">Reject</button>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

// --- Contract Card ---

function ContractCard({ contract, onViewClins, onSubmitInvoice, onRequestMod, isGov, mods, onApproveMod, onRejectMod }: {
  contract: Contract
  onViewClins: () => void
  onSubmitInvoice: () => void
  onRequestMod: () => void
  isGov: boolean
  mods: ContractMod[]
  onApproveMod: (id: string) => void
  onRejectMod: (id: string) => void
}) {
  const expendedPct = ((contract.totalExpended / contract.totalCeiling) * 100).toFixed(1)
  const obligatedPct = ((contract.totalObligated / contract.totalCeiling) * 100).toFixed(1)
  const popStart = new Date(contract.popStart).getTime()
  const popEnd = new Date(contract.popEnd).getTime()
  const now = Date.now()
  const popProgress = Math.min(100, Math.max(0, ((now - popStart) / (popEnd - popStart)) * 100))

  return (
    <div className="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden mb-6">
      <div className="bg-slate-800 px-4 sm:px-6 py-3 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <span className="text-white font-mono text-sm sm:text-base font-bold">{contract.contractNumber}</span>
          <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold border ${
            contract.status === 'ACTIVE' ? 'bg-green-100 text-green-800 border-green-200' : 'bg-gray-100 text-gray-800 border-gray-200'
          }`}>{contract.status}</span>
        </div>
        <span className="text-emerald-400 text-xs sm:text-sm font-medium">{contract.contractor}</span>
      </div>

      <div className="p-4 sm:p-6">
        <h2 className="text-lg sm:text-xl font-bold text-gray-900 mb-1">{contract.title}</h2>

        {/* Pending Mods */}
        <PendingModsPanel mods={mods} contractNumber={contract.contractNumber} isGov={isGov} onApprove={onApproveMod} onReject={onRejectMod} />

        {/* Period of Performance */}
        <div className="mt-4 p-3 bg-slate-50 rounded-lg border border-slate-200">
          <div className="flex items-center justify-between mb-2">
            <span className="text-xs font-semibold text-gray-700 uppercase tracking-wide">Period of Performance</span>
            <span className="text-xs text-gray-500">{popProgress.toFixed(0)}% elapsed</span>
          </div>
          <div className="w-full bg-gray-200 rounded-full h-2.5 mb-2">
            <div className="bg-blue-600 rounded-full h-2.5 transition-all" style={{ width: `${popProgress}%` }} />
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
              <p className="text-xs text-gray-500 font-medium">Remaining</p>
              <p className="text-base sm:text-lg font-bold text-gray-900">{formatCurrency(contract.totalCeiling - contract.totalExpended)}</p>
            </div>
          </div>
        </div>

        {/* Actions */}
        <div className="mt-6 flex gap-3 flex-wrap">
          <button onClick={onViewClins} className="inline-flex items-center gap-2 px-5 py-2.5 bg-slate-800 text-white text-sm font-medium rounded-lg hover:bg-slate-700 transition-colors">
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" /></svg>
            View CLINs
          </button>
          {!isGov && contract.status === 'ACTIVE' && (
            <button onClick={onSubmitInvoice} className="inline-flex items-center gap-2 px-5 py-2.5 bg-green-600 text-white text-sm font-medium rounded-lg hover:bg-green-700 transition-colors">
              Submit Invoice
            </button>
          )}
          {contract.status === 'ACTIVE' && (
            <button onClick={onRequestMod} className="inline-flex items-center gap-2 px-5 py-2.5 bg-purple-600 text-white text-sm font-medium rounded-lg hover:bg-purple-700 transition-colors">
              {isGov ? 'Issue Mod' : 'Request Mod'}
            </button>
          )}
        </div>
      </div>
    </div>
  )
}

// --- Main Component ---

function Contracts() {
  const { state, submitInvoice, approveInvoice, rejectInvoice, submitContractMod, approveContractMod, rejectContractMod } = useAppContext()
  const isGov = state.currentRole === 'GOV'

  // Vendor scoping: vendor sees only their contracts
  // Sort by upcoming POP end date (soonest first)
  const contracts = useMemo(() => {
    const filtered = isGov
      ? state.contracts
      : state.contracts.filter(c => c.contractor === state.vendorCompany)
    return [...filtered].sort((a, b) => new Date(a.popEnd).getTime() - new Date(b.popEnd).getTime())
  }, [state.contracts, isGov, state.vendorCompany])

  const invoices = state.invoices

  const [expandedContract, setExpandedContract] = useState<string | null>(null)
  const [expandedClins, setExpandedClins] = useState<Set<string>>(new Set())
  const [invoiceContract, setInvoiceContract] = useState<Contract | null>(null)
  const [modContract, setModContract] = useState<Contract | null>(null)
  const [toast, setToast] = useState<string | null>(null)

  function toggleClin(clinNumber: string) {
    setExpandedClins((prev) => {
      const next = new Set(prev)
      if (next.has(clinNumber)) { next.delete(clinNumber) } else { next.add(clinNumber) }
      return next
    })
  }

  function showToast(message: string) {
    setToast(message)
    setTimeout(() => setToast(null), 4000)
  }

  function handleSubmitInvoice(clinNumber: string, amount: number, description: string) {
    if (invoiceContract) {
      submitInvoice(invoiceContract.id, clinNumber, amount, description)
      setInvoiceContract(null)
      showToast('Invoice submitted! Compliance check complete.')
    }
  }

  function handleApproveInvoice(id: string, justification: string) {
    approveInvoice(id, justification)
    showToast('Invoice approved and payment disbursed.')
  }

  function handleRejectInvoice(id: string, reason: string) {
    rejectInvoice(id, reason)
    showToast('Invoice rejected.')
  }

  function handleSubmitMod(type: 'REA' | 'ECP' | 'GOV_MOD', title: string, description: string, amount: number) {
    if (modContract) {
      submitContractMod(modContract.id, type, title, description, amount)
      setModContract(null)
      showToast('Contract modification submitted.')
    }
  }

  return (
    <div className="p-4 sm:p-6 lg:p-8 max-w-7xl mx-auto">
      {/* Page Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Contract Management</h1>
        <p className="mt-1 text-sm text-gray-500">
          {isGov
            ? 'Federal contract oversight, CLIN-level financial tracking, and invoice review'
            : 'View your contracts and submit invoices for payment'}
        </p>
      </div>

      {/* GOV: Flagged invoices for review */}
      {isGov && (
        <FlaggedInvoicesPanel invoices={invoices} onApprove={handleApproveInvoice} onReject={handleRejectInvoice} />
      )}

      {/* Empty State */}
      {contracts.length === 0 ? (
        <div className="text-center py-16 bg-white rounded-lg border border-gray-200">
          <svg className="mx-auto h-16 w-16 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          <p className="mt-4 text-lg font-medium text-gray-700">No Contracts Yet</p>
          <p className="mt-2 text-sm text-gray-500 max-w-md mx-auto">
            {isGov
              ? 'Contracts are created when a proposal is approved. Go to Solicitations to review proposals and award contracts.'
              : 'Submit proposals on open solicitations to receive contract awards.'}
          </p>
        </div>
      ) : (
        <>
          {/* Contract Cards */}
          {contracts.map(contract => (
            <div key={contract.id}>
              <ContractCard
                contract={contract}
                isGov={isGov}
                mods={state.contractMods.filter(m => m.contractId === contract.id)}
                onViewClins={() => setExpandedContract(expandedContract === contract.id ? null : contract.id)}
                onSubmitInvoice={() => setInvoiceContract(contract)}
                onRequestMod={() => setModContract(contract)}
                onApproveMod={approveContractMod}
                onRejectMod={rejectContractMod}
              />

              {/* CLIN details */}
              {expandedContract === contract.id && (
                <div className="mb-6 animate-in fade-in">
                  <div className="flex items-center justify-between mb-3">
                    <h2 className="text-lg font-semibold text-gray-900">CLIN Details</h2>
                    <button onClick={() => setExpandedContract(null)} className="text-xs text-gray-500 hover:text-gray-700 font-medium px-2 py-1 rounded hover:bg-gray-100">Hide CLINs</button>
                  </div>
                  {contract.clins.map((clin) => (
                    <CLINRow key={clin.clinNumber} clin={clin} isExpanded={expandedClins.has(clin.clinNumber)} onToggle={() => toggleClin(clin.clinNumber)} />
                  ))}
                </div>
              )}
            </div>
          ))}

          {/* SBIR Timeline */}
          {contracts.some(c => c.contractNumber === 'FA8750-25-F-0018') && (
            <div className="mb-6">
              <SBIRTimeline phases={defaultSBIRPhases} />
            </div>
          )}

          {/* Invoice History */}
          {invoices.length > 0 && (
            <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-4 sm:p-5">
              <h3 className="text-sm font-semibold text-gray-900 uppercase tracking-wide mb-3">Invoice History</h3>
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="text-left text-xs font-medium text-gray-500 uppercase border-b border-gray-200">
                      <th className="pb-2 pr-4">ID</th>
                      <th className="pb-2 pr-4">CLIN</th>
                      <th className="pb-2 pr-4">Amount</th>
                      <th className="pb-2 pr-4">Status</th>
                      <th className="pb-2 pr-4">Submitted</th>
                      <th className="pb-2">Description</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-100">
                    {invoices.map(inv => (
                      <tr key={inv.id}>
                        <td className="py-2 pr-4 font-mono text-xs text-gray-700">{inv.id}</td>
                        <td className="py-2 pr-4 text-gray-700">{inv.clinNumber}</td>
                        <td className="py-2 pr-4 font-medium text-gray-900">{formatCurrency(inv.amount)}</td>
                        <td className="py-2 pr-4">
                          <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${
                            inv.status === 'DISBURSED' ? 'bg-green-100 text-green-800' :
                            inv.status === 'FLAGGED' ? 'bg-amber-100 text-amber-800' :
                            inv.status === 'REJECTED' ? 'bg-red-100 text-red-800' :
                            'bg-blue-100 text-blue-800'
                          }`}>{inv.status}</span>
                        </td>
                        <td className="py-2 pr-4 text-xs text-gray-500">{new Date(inv.submittedAt).toLocaleDateString()}</td>
                        <td className="py-2 text-xs text-gray-600 max-w-[200px] truncate">{inv.description}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </>
      )}

      {/* Invoice Form Modal */}
      {invoiceContract && (
        <InvoiceForm contract={invoiceContract} onSubmit={handleSubmitInvoice} onClose={() => setInvoiceContract(null)} />
      )}

      {/* Contract Mod Form Modal */}
      {modContract && (
        <ContractModForm contract={modContract} isGov={isGov} onSubmit={handleSubmitMod} onClose={() => setModContract(null)} />
      )}

      {/* Toast */}
      {toast && (
        <div className="fixed top-4 right-4 z-[60]">
          <div className="bg-green-600 text-white px-4 py-3 rounded-lg shadow-lg flex items-center gap-3">
            <svg className="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
            <span className="text-sm font-medium">{toast}</span>
            <button onClick={() => setToast(null)} className="ml-2 text-green-200 hover:text-white">
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>
      )}
    </div>
  )
}

export default Contracts
