import { ReactNode } from 'react'
import Sidebar from './Sidebar'
import MobileNav from './MobileNav'
import FloatingActionButton from './FloatingActionButton'

interface MainLayoutProps {
  children: ReactNode
}

function MainLayout({ children }: MainLayoutProps) {
  return (
    <div className="min-h-screen bg-gray-50 flex">
      {/* Sidebar: hidden on mobile, collapsible on tablet, persistent on desktop */}
      <Sidebar />

      {/* Main content area */}
      <main className="flex-1 pb-16 sm:pb-0">
        {/* Mobile: single-column stacked layout */}
        {/* Tablet: two-column split view handled by page content */}
        {/* Desktop: full multi-panel grid layout handled by page content */}
        <div className="w-full lg:grid lg:grid-cols-1">
          {children}
        </div>
      </main>

      {/* Mobile bottom tab navigation */}
      <MobileNav />

      {/* Mobile floating action button for quick upload */}
      <FloatingActionButton />
    </div>
  )
}

export default MainLayout
