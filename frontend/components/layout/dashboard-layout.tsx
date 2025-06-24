import type React from "react"
import { Sidebar } from "./sidebar"
import { MainHeader } from "./main-header"

export function DashboardLayout({
  children,
  currentPageTitle,
}: { children: React.ReactNode; currentPageTitle: string }) {
  return (
    <div className="flex min-h-screen w-full bg-background">
      <Sidebar />
      <div className="flex flex-1 flex-col">
        <MainHeader pageTitle={currentPageTitle} />
        <main className="flex-1 p-4 sm:p-6 lg:p-8 overflow-y-auto">{children}</main>
      </div>
    </div>
  )
}
