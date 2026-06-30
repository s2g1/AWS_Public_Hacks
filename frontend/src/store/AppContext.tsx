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

export interface AppState {
  currentRole: 'GOV' | 'VENDOR'
  solicitations: Solicitation[]
  proposals: Proposal[]
  contracts: Contract[]
  invoices: Invoice[]
  payments: Payment[]
}

// --- Context Interface ---

interface AppContextValue {
  state: AppState
  switchRole: (role: 'GOV' | 'VENDOR') => void
  createSolicitation: (data: Omit<Solicitation, 'id' | 'solicitationNumber' | 'postedDate' | 'status'>) => void
  publishSolicitation: (id: string) => void
  submitProposal: (solicitationId: string, data: Omit<Proposal, 'id' | 'solicitationId' | 'submittedAt' | 'status'>) => void
  approveProposal: (proposalId: string) => void
  rejectProposal: (proposalId: string) => void
  submitInvoice: (contractId: string, clinNumber: string, amount: number, description: string) => void
  approveInvoice: (invoiceId: string, justification?: string) => void
  rejectInvoice: (invoiceId: string, reason: string) => void
}

const STORAGE_KEY = 'fedpay_app_state'

// --- Seed Data ---

function createSeedData(): AppState {
  return {
    currentRole: 'GOV',
    solicitations: [
      {
        id: 'sol-1',
        solicitationNumber: 'FA8750-25-SBIR-0042',
        title: 'Agentic AI for Federal Payment Modernization',
        type: 'SBIR Phase II',
        agency: 'AFRL / Air Force Research Laboratory',
        description: 'The Air Force Research Laboratory seeks innovative solutions leveraging agentic AI architectures to modernize federal payment processing systems. The solution must demonstrate autonomous document extraction, intelligent validation, and adaptive disbursement routing capabilities while maintaining FedRAMP High compliance.',
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
        description: 'AFRL requires a cloud-native document intelligence platform capable of processing unstructured federal financial documents at scale. The platform must support OCR, NLP-based entity extraction, automated classification, and integration with existing Treasury systems. Must achieve 99.5% accuracy on standard government form types (SF-1034, SF-1035, DD-250).',
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
    ],
    invoices: [],
    payments: [],
  }
}

// --- Persistence ---

function loadState(): AppState {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored) {
      return JSON.parse(stored) as AppState
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
    }))
  }, [])

  const publishSolicitation = useCallback((id: string) => {
    setState(prev => ({
      ...prev,
      solicitations: prev.solicitations.map(s =>
        s.id === id && s.status === 'DRAFT' ? { ...s, status: 'OPEN' as const } : s
      ),
    }))
  }, [])

  const submitProposal = useCallback((solicitationId: string, data: Omit<Proposal, 'id' | 'solicitationId' | 'submittedAt' | 'status'>) => {
    const id = `prop-${Date.now()}`
    const newProposal: Proposal = {
      ...data,
      id,
      solicitationId,
      submittedAt: new Date().toISOString(),
      status: 'SUBMITTED',
    }
    setState(prev => ({
      ...prev,
      proposals: [...prev.proposals, newProposal],
    }))
  }, [])

  const approveProposal = useCallback((proposalId: string) => {
    setState(prev => {
      const proposal = prev.proposals.find(p => p.id === proposalId)
      if (!proposal) return prev

      const solicitation = prev.solicitations.find(s => s.id === proposal.solicitationId)
      if (!solicitation) return prev

      // Create contract from proposal
      const contractId = `contract-${Date.now()}`
      const contractNumber = `FA8750-25-F-${String(Math.floor(Math.random() * 9000) + 1000)}`
      const clins: CLINItem[] = proposal.clinStructure || [
        {
          clinNumber: '0001',
          description: 'Base Performance',
          type: 'CPFF',
          ceiling: proposal.priceProposal,
          obligated: proposal.priceProposal * 0.6,
          expended: 0,
        },
      ]

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
      }
    })
  }, [])

  const rejectProposal = useCallback((proposalId: string) => {
    setState(prev => ({
      ...prev,
      proposals: prev.proposals.map(p =>
        p.id === proposalId ? { ...p, status: 'REJECTED' as const } : p
      ),
    }))
  }, [])

  const submitInvoice = useCallback((contractId: string, clinNumber: string, amount: number, description: string) => {
    setState(prev => {
      const id = `inv-${Date.now()}`
      const issues = runComplianceCheck({ amount, clinNumber, contractId }, prev.contracts)

      let invoiceStatus: Invoice['status']
      let newPayments = prev.payments
      let updatedContracts = prev.contracts

      if (issues.length === 0) {
        // Auto-approve and create payment
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
        // Update CLIN expended amount
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
      } else {
        invoiceStatus = 'FLAGGED'
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
      }
    })
  }, [])

  const approveInvoice = useCallback((invoiceId: string, justification?: string) => {
    setState(prev => {
      const invoice = prev.invoices.find(i => i.id === invoiceId)
      if (!invoice || invoice.status !== 'FLAGGED') return prev

      const paymentId = `pay-${Date.now()}`
      const newPayment: Payment = {
        id: paymentId,
        invoiceId,
        contractId: invoice.contractId,
        amount: invoice.amount,
        status: 'DISBURSED',
        disbursedAt: new Date().toISOString(),
      }

      // Update CLIN expended
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
      }
    })
  }, [])

  const rejectInvoice = useCallback((invoiceId: string, reason: string) => {
    setState(prev => ({
      ...prev,
      invoices: prev.invoices.map(i =>
        i.id === invoiceId ? { ...i, status: 'REJECTED' as const, govJustification: reason } : i
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
    submitInvoice,
    approveInvoice,
    rejectInvoice,
  }

  return (
    <AppContext.Provider value={value}>
      {children}
    </AppContext.Provider>
  )
}
