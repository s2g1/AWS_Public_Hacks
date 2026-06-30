import { createContext, useContext, useState, useEffect, useCallback, ReactNode } from 'react'

// --- Types ---

export interface CLINItem {
  clinNumber: string
  description: string
  type: string
  ceiling: number
  obligated: number
  expended: number
}

export interface Solicitation {
  id: string
  solicitationNumber: string
  title: string
  type: string
  agency: string
  description: string
  naicsCode: string
  estimatedValue: string
  closeDate: string
  postedDate: string
  status: 'DRAFT' | 'OPEN' | 'CLOSED' | 'AWARDED' | 'CANCELLED'
  evaluationCriteria: string
  awardedTo?: string
  awardedProposalId?: string
  attachmentName?: string
}

export interface AIEvaluation {
  summary: string
  clinBreakdown: CLINItem[]
  boeAllocation: string
  score: number
  recommendation: 'APPROVE' | 'REVIEW' | 'REJECT'
}

export interface Proposal {
  id: string
  solicitationId: string
  companyName: string
  technicalApproach: string
  priceProposal: number
  pastPerformance: string
  keyPersonnel: string
  submittedAt: string
  status: 'SUBMITTED' | 'UNDER_REVIEW' | 'APPROVED' | 'REJECTED'
  clinStructure?: CLINItem[]
  boeInfo?: string
  attachmentName?: string
  aiEvaluation?: AIEvaluation
}

export interface Contract {
  id: string
  contractNumber: string
  title: string
  contractor: string
  proposalId: string
  solicitationId: string
  status: 'ACTIVE' | 'COMPLETED'
  popStart: string
  popEnd: string
  totalCeiling: number
  totalObligated: number
  totalExpended: number
  clins: CLINItem[]
}

export interface Invoice {
  id: string
  contractId: string
  clinNumber: string
  amount: number
  description: string
  submittedAt: string
  status: 'SUBMITTED' | 'COMPLIANCE_CHECK' | 'FLAGGED' | 'APPROVED' | 'DISBURSED' | 'REJECTED'
  complianceIssues?: string[]
  govJustification?: string
}

export interface Payment {
  id: string
  invoiceId: string
  contractId: string
  amount: number
  status: 'PROCESSING' | 'DISBURSED' | 'FAILED'
  disbursedAt?: string
}

export interface Notification {
  id: string
  targetRole: 'GOV' | 'VENDOR'
  targetCompany?: string
  message: string
  type: 'info' | 'action_required' | 'success' | 'warning'
  relatedId?: string
  createdAt: string
  read: boolean
}

export interface HistoryEntry {
  id: string
  actor: 'GOV' | 'VENDOR'
  actorName: string
  action: string
  details: string
  timestamp: string
  relatedId?: string
}

export interface ContractMod {
  id: string
  contractId: string
  type: 'REA' | 'ECP' | 'GOV_MOD'
  requestedBy: 'GOV' | 'VENDOR'
  title: string
  description: string
  amount: number
  status: 'SUBMITTED' | 'UNDER_REVIEW' | 'APPROVED' | 'REJECTED'
  submittedAt: string
  resolvedAt?: string
}

export interface AppState {
  currentRole: 'GOV' | 'VENDOR'
  vendorCompany: string
  solicitations: Solicitation[]
  proposals: Proposal[]
  contracts: Contract[]
  invoices: Invoice[]
  payments: Payment[]
  notifications: Notification[]
  history: HistoryEntry[]
  contractMods: ContractMod[]
}

// --- Context Interface ---

interface AppContextValue {
  state: AppState
  switchRole: (role: 'GOV' | 'VENDOR') => void
  createSolicitation: (data: Omit<Solicitation, 'id' | 'solicitationNumber' | 'postedDate' | 'status'>) => void
  publishSolicitation: (id: string) => void
  submitProposal: (solicitationId: string, data: Omit<Proposal, 'id' | 'solicitationId' | 'submittedAt' | 'status' | 'aiEvaluation'>) => void
  approveProposal: (proposalId: string) => void
  rejectProposal: (proposalId: string) => void
  submitInvoice: (contractId: string, clinNumber: string, amount: number, description: string) => void
  approveInvoice: (invoiceId: string, justification?: string) => void
  rejectInvoice: (invoiceId: string, reason: string) => void
  generateProposalEvaluation: (proposalId: string) => void
  submitContractMod: (contractId: string, type: 'REA' | 'ECP' | 'GOV_MOD', title: string, description: string, amount: number) => void
  approveContractMod: (modId: string) => void
  rejectContractMod: (modId: string) => void
  markNotificationRead: (id: string) => void
}

const STORAGE_KEY = 'fedpay_app_state'

// --- Seed Data ---

function createSeedData(): AppState {
  return {
    currentRole: 'GOV',
    vendorCompany: 'Nexus AI Solutions LLC',
    solicitations: [
      {
        id: 'sol-1',
        solicitationNumber: 'FA8750-25-SBIR-0042',
        title: 'Agentic AI for Federal Payment Modernization',
        type: 'SBIR Phase II',
        agency: 'AFRL / Air Force Research Laboratory',
        description: 'The Air Force Research Laboratory seeks innovative solutions leveraging agentic AI architectures to modernize federal payment processing systems.',
        naicsCode: '541512',
        estimatedValue: '$750K - $1.5M',
        closeDate: '2025-01-15',
        postedDate: '2024-12-15',
        status: 'AWARDED',
        evaluationCriteria: 'Technical Approach (40%), Past Performance (25%), Price (20%), Key Personnel (15%)',
        awardedTo: 'Quantum Federal Systems LLC',
        awardedProposalId: 'prop-1',
      },
      {
        id: 'sol-2',
        solicitationNumber: 'FA8750-25-RFP-0098',
        title: 'Next-Generation Document Intelligence Platform',
        type: 'Full & Open Competition',
        agency: 'AFRL / Air Force Research Laboratory',
        description: 'AFRL requires a cloud-native document intelligence platform capable of processing unstructured federal financial documents at scale.',
        naicsCode: '541511',
        estimatedValue: '$2.5M - $5M',
        closeDate: '2025-07-15',
        postedDate: '2025-06-01',
        status: 'OPEN',
        evaluationCriteria: 'Technical Capability (35%), Management Approach (20%), Past Performance (25%), Price (20%)',
      },
    ],
    proposals: [
      {
        id: 'prop-1',
        solicitationId: 'sol-1',
        companyName: 'Quantum Federal Systems LLC',
        technicalApproach: 'Multi-agent architecture with autonomous document extraction, intelligent validation pipelines, and adaptive disbursement routing using LLM-based decision engines.',
        priceProposal: 1249800,
        pastPerformance: 'Successfully delivered 3 prior SBIR Phase II contracts for DoD AI/ML systems with average CPARS rating of Exceptional.',
        keyPersonnel: 'Dr. Sarah Chen (PI) - 15 years federal AI systems; Mark Rodriguez (Tech Lead) - Former Treasury systems architect',
        submittedAt: '2025-01-10T14:30:00Z',
        status: 'APPROVED',
        clinStructure: [
          { clinNumber: '0001', description: 'Phase II Research & Development', type: 'CPFF', ceiling: 749880, obligated: 749880, expended: 208300 },
          { clinNumber: '0002', description: 'Phase III Option - Production Pilot', type: 'CPFF', ceiling: 499920, obligated: 0, expended: 0 },
        ],
      },
    ],
    contracts: [
      {
        id: 'contract-1',
        contractNumber: 'FA8750-25-F-0018',
        title: 'Autonomous Payment Processing AI - SBIR Phase II',
        contractor: 'Quantum Federal Systems LLC',
        proposalId: 'prop-1',
        solicitationId: 'sol-1',
        status: 'ACTIVE',
        popStart: '2025-01-02',
        popEnd: '2025-07-01',
        totalCeiling: 1249800,
        totalObligated: 749880,
        totalExpended: 208300,
        clins: [
          { clinNumber: '0001', description: 'Phase II Research & Development', type: 'CPFF', ceiling: 749880, obligated: 749880, expended: 208300 },
          { clinNumber: '0002', description: 'Phase III Option - Production Pilot', type: 'CPFF', ceiling: 499920, obligated: 0, expended: 0 },
        ],
      },
      {
        id: 'contract-2',
        contractNumber: 'FA8750-24-F-0092',
        title: 'Legacy System Integration Connector',
        contractor: 'Atlas Defense Technologies',
        proposalId: 'prop-atlas-1',
        solicitationId: 'sol-atlas-1',
        status: 'ACTIVE',
        popStart: '2024-06-01',
        popEnd: '2025-06-01',
        totalCeiling: 890000,
        totalObligated: 890000,
        totalExpended: 445000,
        clins: [
          { clinNumber: '0001', description: 'Integration Services', type: 'FFP', ceiling: 890000, obligated: 890000, expended: 445000 },
        ],
      },
    ],
    invoices: [],
    payments: [],
    notifications: [
      {
        id: 'notif-seed-1',
        targetRole: 'VENDOR',
        targetCompany: 'Quantum Federal Systems LLC',
        message: 'Your proposal for "Agentic AI for Federal Payment Modernization" has been approved! Contract FA8750-25-F-0018 created.',
        type: 'success',
        relatedId: 'contract-1',
        createdAt: '2025-01-02T10:00:00Z',
        read: true,
      },
      {
        id: 'notif-seed-2',
        targetRole: 'GOV',
        message: 'New proposal received from Quantum Federal Systems LLC for "Agentic AI for Federal Payment Modernization"',
        type: 'action_required',
        relatedId: 'prop-1',
        createdAt: '2025-01-10T14:30:00Z',
        read: true,
      },
      {
        id: 'notif-seed-3',
        targetRole: 'VENDOR',
        message: 'New solicitation published: "Next-Generation Document Intelligence Platform"',
        type: 'info',
        relatedId: 'sol-2',
        createdAt: '2025-06-01T08:00:00Z',
        read: false,
      },
      {
        id: 'notif-seed-4',
        targetRole: 'GOV',
        message: 'Invoice submitted for FA8750-24-F-0092 - $75,000',
        type: 'action_required',
        relatedId: 'contract-2',
        createdAt: '2025-06-10T11:00:00Z',
        read: false,
      },
      {
        id: 'notif-seed-5',
        targetRole: 'GOV',
        message: 'Contract mod request from Quantum Federal Systems LLC: Phase III scope expansion',
        type: 'action_required',
        relatedId: 'contract-1',
        createdAt: '2025-06-12T09:30:00Z',
        read: false,
      },
      {
        id: 'notif-seed-6',
        targetRole: 'VENDOR',
        targetCompany: 'Quantum Federal Systems LLC',
        message: 'Payment of $208,300 disbursed for FA8750-25-F-0018',
        type: 'success',
        relatedId: 'contract-1',
        createdAt: '2025-05-20T14:00:00Z',
        read: true,
      },
    ],
    history: [
      {
        id: 'hist-seed-1',
        actor: 'GOV',
        actorName: 'Contracting Officer',
        action: 'Solicitation Published',
        details: 'Published "Agentic AI for Federal Payment Modernization" (FA8750-25-SBIR-0042)',
        timestamp: '2024-12-15T09:00:00Z',
        relatedId: 'sol-1',
      },
      {
        id: 'hist-seed-2',
        actor: 'VENDOR',
        actorName: 'Quantum Federal Systems LLC',
        action: 'Proposal Submitted',
        details: 'Submitted proposal for "Agentic AI for Federal Payment Modernization" — $1,249,800',
        timestamp: '2025-01-10T14:30:00Z',
        relatedId: 'prop-1',
      },
      {
        id: 'hist-seed-3',
        actor: 'GOV',
        actorName: 'Contracting Officer',
        action: 'Proposal Approved',
        details: 'Approved proposal from Quantum Federal Systems LLC. Contract FA8750-25-F-0018 created.',
        timestamp: '2025-01-02T10:00:00Z',
        relatedId: 'contract-1',
      },
      {
        id: 'hist-seed-4',
        actor: 'GOV',
        actorName: 'Contracting Officer',
        action: 'Solicitation Published',
        details: 'Published "Next-Generation Document Intelligence Platform" (FA8750-25-RFP-0098)',
        timestamp: '2025-06-01T08:00:00Z',
        relatedId: 'sol-2',
      },
      {
        id: 'hist-seed-5',
        actor: 'VENDOR',
        actorName: 'Quantum Federal Systems LLC',
        action: 'Invoice Submitted',
        details: 'Submitted invoice for $208,300 on FA8750-25-F-0018 (CLIN 0001)',
        timestamp: '2025-05-15T10:00:00Z',
        relatedId: 'contract-1',
      },
      {
        id: 'hist-seed-6',
        actor: 'VENDOR',
        actorName: 'Quantum Federal Systems LLC',
        action: 'Payment Received',
        details: 'Payment of $208,300 disbursed for FA8750-25-F-0018',
        timestamp: '2025-05-20T14:00:00Z',
        relatedId: 'contract-1',
      },
      {
        id: 'hist-seed-7',
        actor: 'GOV',
        actorName: 'Contracting Officer',
        action: 'Invoice Approved',
        details: 'Approved invoice for $208,300 on FA8750-25-F-0018 — compliance passed',
        timestamp: '2025-05-19T11:30:00Z',
        relatedId: 'contract-1',
      },
      {
        id: 'hist-seed-8',
        actor: 'VENDOR',
        actorName: 'Quantum Federal Systems LLC',
        action: 'Mod Requested',
        details: 'Submitted ECP for FA8750-25-F-0018: "Phase III scope expansion" ($250,000)',
        timestamp: '2025-06-12T09:30:00Z',
        relatedId: 'contract-1',
      },
    ],
    contractMods: [],
  }
}

// --- Persistence ---

function loadState(): AppState {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored) {
      const parsed = JSON.parse(stored) as AppState
      // Ensure new fields exist (migration)
      if (!parsed.notifications) parsed.notifications = []
      if (!parsed.history) parsed.history = []
      if (!parsed.contractMods) parsed.contractMods = []
      if (!parsed.vendorCompany) parsed.vendorCompany = 'Nexus AI Solutions LLC'
      return parsed
    }
  } catch {
    // If corrupted, fall through to seed
  }
  return createSeedData()
}

function persistState(state: AppState) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(state))
}

// --- Compliance Check Logic ---

function runComplianceCheck(invoice: { amount: number; clinNumber: string; contractId: string }, contracts: Contract[]): string[] {
  const issues: string[] = []
  const contract = contracts.find(c => c.id === invoice.contractId)
  if (!contract) return ['Contract not found']

  const clin = contract.clins.find(c => c.clinNumber === invoice.clinNumber)
  if (!clin) return ['CLIN not found on contract']

  const remaining = clin.ceiling - clin.expended
  if (invoice.amount > remaining) {
    issues.push(`Exceeds CLIN obligation: invoice $${invoice.amount.toLocaleString()} > remaining $${remaining.toLocaleString()}`)
  }

  if (invoice.amount > 25000 && clin.type === 'FFP') {
    issues.push('FFP invoice above $25,000 threshold requires milestone verification')
  }

  return issues
}

// --- Helper: create notification ---

function addNotification(
  notifications: Notification[],
  targetRole: 'GOV' | 'VENDOR',
  message: string,
  type: Notification['type'],
  relatedId?: string,
  targetCompany?: string,
): Notification[] {
  const notif: Notification = {
    id: `notif-${Date.now()}-${Math.random().toString(36).slice(2, 6)}`,
    targetRole,
    targetCompany,
    message,
    type,
    relatedId,
    createdAt: new Date().toISOString(),
    read: false,
  }
  return [notif, ...notifications]
}

// --- Helper: create history entry ---

function addHistory(
  history: HistoryEntry[],
  actor: 'GOV' | 'VENDOR',
  actorName: string,
  action: string,
  details: string,
  relatedId?: string,
): HistoryEntry[] {
  const entry: HistoryEntry = {
    id: `hist-${Date.now()}-${Math.random().toString(36).slice(2, 6)}`,
    actor,
    actorName,
    action,
    details,
    timestamp: new Date().toISOString(),
    relatedId,
  }
  return [entry, ...history]
}

// --- Context ---

const AppContext = createContext<AppContextValue | null>(null)

export function useAppContext(): AppContextValue {
  const ctx = useContext(AppContext)
  if (!ctx) {
    throw new Error('useAppContext must be used within an AppProvider')
  }
  return ctx
}

// --- Provider ---

export function AppProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AppState>(loadState)

  // Persist on every change
  useEffect(() => {
    persistState(state)
  }, [state])

  const switchRole = useCallback((role: 'GOV' | 'VENDOR') => {
    setState(prev => ({ ...prev, currentRole: role }))
  }, [])

  const createSolicitation = useCallback((data: Omit<Solicitation, 'id' | 'solicitationNumber' | 'postedDate' | 'status'>) => {
    const id = `sol-${Date.now()}`
    const solNum = `FA8750-25-SBIR-${String(Math.floor(Math.random() * 9000) + 1000)}`
    const newSol: Solicitation = {
      ...data,
      id,
      solicitationNumber: solNum,
      postedDate: new Date().toISOString().split('T')[0],
      status: 'DRAFT',
    }
    setState(prev => ({
      ...prev,
      solicitations: [newSol, ...prev.solicitations],
      history: addHistory(prev.history, 'GOV', 'Contracting Officer', 'Solicitation Created', `Created draft solicitation: "${data.title}" (${solNum})`, id),
    }))
  }, [])

  const publishSolicitation = useCallback((id: string) => {
    setState(prev => {
      const sol = prev.solicitations.find(s => s.id === id)
      if (!sol || sol.status !== 'DRAFT') return prev
      return {
        ...prev,
        solicitations: prev.solicitations.map(s =>
          s.id === id ? { ...s, status: 'OPEN' as const } : s
        ),
        notifications: addNotification(prev.notifications, 'VENDOR', `New solicitation published: "${sol.title}"`, 'info', id),
        history: addHistory(prev.history, 'GOV', 'Contracting Officer', 'Solicitation Published', `Published "${sol.title}" (${sol.solicitationNumber})`, id),
      }
    })
  }, [])

  const generateProposalEvaluation = useCallback((proposalId: string) => {
    setState(prev => {
      const proposal = prev.proposals.find(p => p.id === proposalId)
      if (!proposal) return prev

      const price = proposal.priceProposal
      const rdAmount = Math.round(price * 0.55)
      const integrationAmount = Math.round(price * 0.30)
      const pmAmount = Math.round(price * 0.15)

      const score = Math.floor(Math.random() * 21) + 75 // 75-95

      let recommendation: 'APPROVE' | 'REVIEW' | 'REJECT'
      if (score > 85) recommendation = 'APPROVE'
      else if (score >= 70) recommendation = 'REVIEW'
      else recommendation = 'REJECT'

      const summary = `Proposal demonstrates ${score > 85 ? 'strong' : 'adequate'} technical capability in ${proposal.technicalApproach.split(' ').slice(0, 8).join(' ')}... Price proposal of $${price.toLocaleString()} is ${score > 80 ? 'competitive' : 'within acceptable range'} for the scope of work. Past performance indicators suggest ${score > 85 ? 'high' : 'moderate'} probability of successful execution.`

      const aiEvaluation: AIEvaluation = {
        summary,
        clinBreakdown: [
          { clinNumber: '0001', description: 'Research & Development', type: 'CPFF', ceiling: rdAmount, obligated: 0, expended: 0 },
          { clinNumber: '0002', description: 'System Integration & Testing', type: 'CPFF', ceiling: integrationAmount, obligated: 0, expended: 0 },
          { clinNumber: '0003', description: 'Program Management', type: 'FFP', ceiling: pmAmount, obligated: 0, expended: 0 },
        ],
        boeAllocation: `R&D: ${((rdAmount / price) * 100).toFixed(0)}% ($${rdAmount.toLocaleString()}) | Integration: ${((integrationAmount / price) * 100).toFixed(0)}% ($${integrationAmount.toLocaleString()}) | PM: ${((pmAmount / price) * 100).toFixed(0)}% ($${pmAmount.toLocaleString()})`,
        score,
        recommendation,
      }

      return {
        ...prev,
        proposals: prev.proposals.map(p =>
          p.id === proposalId ? { ...p, aiEvaluation } : p
        ),
        notifications: addNotification(prev.notifications, 'GOV', `AI evaluation complete for proposal from ${proposal.companyName} — Score: ${score}/100, Recommendation: ${recommendation}`, 'info', proposalId),
        history: addHistory(prev.history, 'GOV', 'AI Evaluation Engine', 'Evaluation Generated', `Generated AI evaluation for ${proposal.companyName} proposal — Score: ${score}/100`, proposalId),
      }
    })
  }, [])

  const submitProposal = useCallback((solicitationId: string, data: Omit<Proposal, 'id' | 'solicitationId' | 'submittedAt' | 'status' | 'aiEvaluation'>) => {
    const id = `prop-${Date.now()}`
    const newProposal: Proposal = {
      ...data,
      id,
      solicitationId,
      submittedAt: new Date().toISOString(),
      status: 'SUBMITTED',
    }
    setState(prev => {
      const sol = prev.solicitations.find(s => s.id === solicitationId)
      return {
        ...prev,
        proposals: [...prev.proposals, newProposal],
        notifications: addNotification(prev.notifications, 'GOV', `New proposal received from ${data.companyName} for "${sol?.title || 'Unknown'}"`, 'action_required', id),
        history: addHistory(prev.history, 'VENDOR', data.companyName, 'Proposal Submitted', `Submitted proposal for "${sol?.title || 'Unknown'}" — $${data.priceProposal.toLocaleString()}`, id),
      }
    })
    // Trigger AI evaluation after brief delay
    setTimeout(() => {
      generateProposalEvaluation(id)
    }, 1500)
  }, [generateProposalEvaluation])

  const approveProposal = useCallback((proposalId: string) => {
    setState(prev => {
      const proposal = prev.proposals.find(p => p.id === proposalId)
      if (!proposal) return prev

      const solicitation = prev.solicitations.find(s => s.id === proposal.solicitationId)
      if (!solicitation) return prev

      const contractId = `contract-${Date.now()}`
      const contractNumber = `FA8750-25-F-${String(Math.floor(Math.random() * 9000) + 1000)}`
      // Priority: AI evaluation CLINs > manual clinStructure > default fallback
      let clins: CLINItem[]
      if (proposal.aiEvaluation?.clinBreakdown && proposal.aiEvaluation.clinBreakdown.length > 0) {
        // Use AI-generated CLIN breakdown, set obligated for first funded period
        clins = proposal.aiEvaluation.clinBreakdown.map((clin, idx) => ({
          ...clin,
          obligated: idx === 0 ? Math.round(clin.ceiling * 0.6) : 0,
          expended: 0,
        }))
      } else if (proposal.clinStructure && proposal.clinStructure.length > 0) {
        clins = proposal.clinStructure
      } else {
        clins = [
          {
            clinNumber: '0001',
            description: 'Base Performance',
            type: 'CPFF',
            ceiling: proposal.priceProposal,
            obligated: Math.round(proposal.priceProposal * 0.6),
            expended: 0,
          },
        ]
      }

      const totalCeiling = clins.reduce((sum, c) => sum + c.ceiling, 0)
      const totalObligated = clins.reduce((sum, c) => sum + c.obligated, 0)

      const newContract: Contract = {
        id: contractId,
        contractNumber,
        title: solicitation.title,
        contractor: proposal.companyName,
        proposalId,
        solicitationId: proposal.solicitationId,
        status: 'ACTIVE',
        popStart: new Date().toISOString().split('T')[0],
        popEnd: new Date(Date.now() + 180 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
        totalCeiling,
        totalObligated,
        totalExpended: 0,
        clins,
      }

      return {
        ...prev,
        proposals: prev.proposals.map(p =>
          p.id === proposalId ? { ...p, status: 'APPROVED' as const } : p
        ),
        solicitations: prev.solicitations.map(s =>
          s.id === proposal.solicitationId
            ? { ...s, status: 'AWARDED' as const, awardedTo: proposal.companyName, awardedProposalId: proposalId }
            : s
        ),
        contracts: [...prev.contracts, newContract],
        notifications: addNotification(prev.notifications, 'VENDOR', `Your proposal for "${solicitation.title}" has been approved! Contract ${contractNumber} created.`, 'success', contractId, proposal.companyName),
        history: addHistory(prev.history, 'GOV', 'Contracting Officer', 'Proposal Approved', `Approved proposal from ${proposal.companyName}. Contract ${contractNumber} created.`, contractId),
      }
    })
  }, [])

  const rejectProposal = useCallback((proposalId: string) => {
    setState(prev => ({
      ...prev,
      proposals: prev.proposals.map(p =>
        p.id === proposalId ? { ...p, status: 'REJECTED' as const } : p
      ),
      history: addHistory(prev.history, 'GOV', 'Contracting Officer', 'Proposal Rejected', `Rejected proposal ${proposalId}`, proposalId),
    }))
  }, [])

  const submitInvoice = useCallback((contractId: string, clinNumber: string, amount: number, description: string) => {
    setState(prev => {
      const id = `inv-${Date.now()}`
      const contract = prev.contracts.find(c => c.id === contractId)
      const issues = runComplianceCheck({ amount, clinNumber, contractId }, prev.contracts)

      let invoiceStatus: Invoice['status']
      let newPayments = prev.payments
      let updatedContracts = prev.contracts
      let newNotifications = prev.notifications
      let newHistory = prev.history

      // Notify GOV about invoice submission
      newNotifications = addNotification(newNotifications, 'GOV', `Invoice submitted for ${contract?.contractNumber || contractId} - $${amount.toLocaleString()}`, 'action_required', id)
      newHistory = addHistory(newHistory, 'VENDOR', contract?.contractor || 'Vendor', 'Invoice Submitted', `Submitted invoice for $${amount.toLocaleString()} on ${contract?.contractNumber || contractId} (CLIN ${clinNumber})`, id)

      if (issues.length === 0) {
        invoiceStatus = 'DISBURSED'
        const paymentId = `pay-${Date.now()}`
        newPayments = [...prev.payments, {
          id: paymentId,
          invoiceId: id,
          contractId,
          amount,
          status: 'DISBURSED' as const,
          disbursedAt: new Date().toISOString(),
        }]
        updatedContracts = prev.contracts.map(c => {
          if (c.id !== contractId) return c
          const updatedClins = c.clins.map(cl =>
            cl.clinNumber === clinNumber ? { ...cl, expended: cl.expended + amount } : cl
          )
          return {
            ...c,
            clins: updatedClins,
            totalExpended: updatedClins.reduce((sum, cl) => sum + cl.expended, 0),
          }
        })
        // Notify vendor of disbursement
        newNotifications = addNotification(newNotifications, 'VENDOR', `Payment of $${amount.toLocaleString()} disbursed for ${contract?.contractNumber || contractId}`, 'success', paymentId, contract?.contractor)
        newHistory = addHistory(newHistory, 'VENDOR', contract?.contractor || 'Vendor', 'Payment Received', `Payment of $${amount.toLocaleString()} disbursed for ${contract?.contractNumber || contractId}`, paymentId)
      } else {
        invoiceStatus = 'FLAGGED'
        newNotifications = addNotification(newNotifications, 'GOV', `Invoice flagged for review: ${issues[0]}`, 'warning', id)
      }

      const newInvoice: Invoice = {
        id,
        contractId,
        clinNumber,
        amount,
        description,
        submittedAt: new Date().toISOString(),
        status: invoiceStatus,
        complianceIssues: issues.length > 0 ? issues : undefined,
      }

      return {
        ...prev,
        invoices: [...prev.invoices, newInvoice],
        payments: newPayments,
        contracts: updatedContracts,
        notifications: newNotifications,
        history: newHistory,
      }
    })
  }, [])

  const approveInvoice = useCallback((invoiceId: string, justification?: string) => {
    setState(prev => {
      const invoice = prev.invoices.find(i => i.id === invoiceId)
      if (!invoice || invoice.status !== 'FLAGGED') return prev

      const contract = prev.contracts.find(c => c.id === invoice.contractId)
      const paymentId = `pay-${Date.now()}`
      const newPayment: Payment = {
        id: paymentId,
        invoiceId,
        contractId: invoice.contractId,
        amount: invoice.amount,
        status: 'DISBURSED',
        disbursedAt: new Date().toISOString(),
      }

      const updatedContracts = prev.contracts.map(c => {
        if (c.id !== invoice.contractId) return c
        const updatedClins = c.clins.map(cl =>
          cl.clinNumber === invoice.clinNumber ? { ...cl, expended: cl.expended + invoice.amount } : cl
        )
        return {
          ...c,
          clins: updatedClins,
          totalExpended: updatedClins.reduce((sum, cl) => sum + cl.expended, 0),
        }
      })

      return {
        ...prev,
        invoices: prev.invoices.map(i =>
          i.id === invoiceId ? { ...i, status: 'DISBURSED' as const, govJustification: justification } : i
        ),
        payments: [...prev.payments, newPayment],
        contracts: updatedContracts,
        notifications: addNotification(prev.notifications, 'VENDOR', `Payment of $${invoice.amount.toLocaleString()} approved and disbursed for ${contract?.contractNumber || invoice.contractId}`, 'success', paymentId, contract?.contractor),
        history: addHistory(prev.history, 'GOV', 'Contracting Officer', 'Invoice Approved', `Approved flagged invoice for $${invoice.amount.toLocaleString()} on ${contract?.contractNumber || invoice.contractId}`, invoiceId),
      }
    })
  }, [])

  const rejectInvoice = useCallback((invoiceId: string, reason: string) => {
    setState(prev => ({
      ...prev,
      invoices: prev.invoices.map(i =>
        i.id === invoiceId ? { ...i, status: 'REJECTED' as const, govJustification: reason } : i
      ),
      history: addHistory(prev.history, 'GOV', 'Contracting Officer', 'Invoice Rejected', `Rejected invoice ${invoiceId}: ${reason}`, invoiceId),
    }))
  }, [])

  const submitContractMod = useCallback((contractId: string, type: 'REA' | 'ECP' | 'GOV_MOD', title: string, description: string, amount: number) => {
    setState(prev => {
      const contract = prev.contracts.find(c => c.id === contractId)
      const modId = `mod-${Date.now()}`
      const requestedBy = type === 'GOV_MOD' ? 'GOV' : 'VENDOR'

      const newMod: ContractMod = {
        id: modId,
        contractId,
        type,
        requestedBy,
        title,
        description,
        amount,
        status: 'SUBMITTED',
        submittedAt: new Date().toISOString(),
      }

      let newNotifications = prev.notifications
      let newHistory = prev.history

      if (requestedBy === 'VENDOR') {
        newNotifications = addNotification(newNotifications, 'GOV', `Contract mod request from ${contract?.contractor || 'vendor'}: ${title}`, 'action_required', modId)
        newHistory = addHistory(newHistory, 'VENDOR', contract?.contractor || 'Vendor', 'Mod Requested', `Submitted ${type} for ${contract?.contractNumber || contractId}: "${title}" ($${amount.toLocaleString()})`, modId)
      } else {
        newNotifications = addNotification(newNotifications, 'VENDOR', `Government-initiated contract mod: ${title}`, 'info', modId, contract?.contractor)
        newHistory = addHistory(newHistory, 'GOV', 'Contracting Officer', 'Mod Issued', `Issued ${type} on ${contract?.contractNumber || contractId}: "${title}" ($${amount.toLocaleString()})`, modId)
      }

      return {
        ...prev,
        contractMods: [...prev.contractMods, newMod],
        notifications: newNotifications,
        history: newHistory,
      }
    })
  }, [])

  const approveContractMod = useCallback((modId: string) => {
    setState(prev => {
      const mod = prev.contractMods.find(m => m.id === modId)
      if (!mod) return prev
      const contract = prev.contracts.find(c => c.id === mod.contractId)

      return {
        ...prev,
        contractMods: prev.contractMods.map(m =>
          m.id === modId ? { ...m, status: 'APPROVED' as const, resolvedAt: new Date().toISOString() } : m
        ),
        notifications: addNotification(prev.notifications, 'VENDOR', `Contract mod "${mod.title}" approved`, 'success', modId, contract?.contractor),
        history: addHistory(prev.history, 'GOV', 'Contracting Officer', 'Mod Approved', `Approved mod "${mod.title}" ($${mod.amount.toLocaleString()}) on ${contract?.contractNumber || mod.contractId}`, modId),
      }
    })
  }, [])

  const rejectContractMod = useCallback((modId: string) => {
    setState(prev => {
      const mod = prev.contractMods.find(m => m.id === modId)
      if (!mod) return prev
      const contract = prev.contracts.find(c => c.id === mod.contractId)

      return {
        ...prev,
        contractMods: prev.contractMods.map(m =>
          m.id === modId ? { ...m, status: 'REJECTED' as const, resolvedAt: new Date().toISOString() } : m
        ),
        notifications: addNotification(prev.notifications, 'VENDOR', `Contract mod "${mod.title}" rejected`, 'warning', modId, contract?.contractor),
        history: addHistory(prev.history, 'GOV', 'Contracting Officer', 'Mod Rejected', `Rejected mod "${mod.title}" on ${contract?.contractNumber || mod.contractId}`, modId),
      }
    })
  }, [])

  const markNotificationRead = useCallback((id: string) => {
    setState(prev => ({
      ...prev,
      notifications: prev.notifications.map(n =>
        n.id === id ? { ...n, read: true } : n
      ),
    }))
  }, [])

  const value: AppContextValue = {
    state,
    switchRole,
    createSolicitation,
    publishSolicitation,
    submitProposal,
    approveProposal,
    rejectProposal,
    generateProposalEvaluation,
    submitInvoice,
    approveInvoice,
    rejectInvoice,
    submitContractMod,
    approveContractMod,
    rejectContractMod,
    markNotificationRead,
  }

  return (
    <AppContext.Provider value={value}>
      {children}
    </AppContext.Provider>
  )
}
