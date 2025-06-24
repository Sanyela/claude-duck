"use client"

import type React from "react"
import { Sidebar } from "./sidebar"
import { MainHeader } from "./main-header"
import { PageTitle } from "@/components/page-title"

export function DashboardLayout({
  children,
}: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen w-full bg-background">
      <PageTitle />
      <Sidebar />
      <div className="flex flex-1 flex-col">
        <MainHeader />
        <main className="flex-1 p-4 pt-3 sm:p-6 sm:pt-4 lg:p-8 lg:pt-5 overflow-y-auto">{children}</main>
      </div>
    </div>
  )
}
