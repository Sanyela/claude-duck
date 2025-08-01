import { DashboardLayout } from "@/components/layout/dashboard-layout"
import { DeviceManagement } from "@/components/settings/device-management"

export default function DevicesPage() {
  return (
    <DashboardLayout>
      <div className="max-w-4xl mx-auto">
        <DeviceManagement />
      </div>
    </DashboardLayout>
  )
}