export const dynamic = 'force-dynamic'

import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { OAuthLinkGenerator } from "@/components/tools/oauth-link-generator"
import { Card } from "@/components/ui/card"

export default function AdminOAuthPage() {
  return (
    <DashboardLayout>
      <Card className="shadow-lg bg-card text-card-foreground border-border">
        <OAuthLinkGenerator />
      </Card>
    </DashboardLayout>
  )
} 