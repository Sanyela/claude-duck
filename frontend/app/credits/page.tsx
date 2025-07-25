export const dynamic = 'force-dynamic'

import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { CreditsContent } from "@/components/credits/credits-content"

export default function CreditsPage() {
  return (
    <DashboardLayout>
      <CreditsContent />
    </DashboardLayout>
  )
}
