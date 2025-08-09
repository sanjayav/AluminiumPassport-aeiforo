import React from 'react'
import { Outlet, createFileRoute, redirect, Link, useRouter } from '@tanstack/react-router'
import '../routeTree.gen'
// import { useAuth } from '../lib/auth'
import { AppSidebar } from '@/components/app-sidebar'
import { SidebarInset, SidebarProvider, SidebarTrigger } from '@/components/ui/sidebar'
import {
    Breadcrumb,
    BreadcrumbItem,
    BreadcrumbLink,
    BreadcrumbList,
    BreadcrumbPage,
    BreadcrumbSeparator,
} from '@/components/ui/breadcrumb'

// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore - types are injected by TanStack Router generated file
export const Route = createFileRoute('/app')({
    beforeLoad: ({ context }: { context: { auth: { isAuthenticated: boolean } } }) => {
        const { auth } = context
        if (!auth.isAuthenticated) {
            throw redirect({ to: '/login' })
        }
    },
    component: AppLayout,
})

function AppLayout() {
    const router = useRouter()
    const pathname = router.state.location.pathname
    const afterApp = pathname.replace(/^\/app\/?/, '')
    const segments = afterApp ? afterApp.split('/').filter(Boolean) : []
    const titleMap: Record<string, string> = { roles: 'Role' }
    const toTitleCase = (s: string) => s.replace(/[-_]/g, ' ').replace(/\b\w/g, (m) => m.toUpperCase())
    const crumbs = segments.map((seg, idx) => {
        const label = titleMap[seg] ?? toTitleCase(seg)
        const to = `/app/${segments.slice(0, idx + 1).join('/')}` as const
        return { label, to }
    })
    return (
        <SidebarProvider>
            <AppSidebar />
            <SidebarInset>
                <header className="flex items-center justify-between border-b p-6">
                    <div className="flex items-center gap-3">
                        <SidebarTrigger />
                        <Breadcrumb>
                            <BreadcrumbList>
                                <BreadcrumbItem>
                                    <BreadcrumbLink asChild>
                                        <Link to="/app">Home</Link>
                                    </BreadcrumbLink>
                                </BreadcrumbItem>
                                {crumbs.map((c, i) => (
                                    <React.Fragment key={`crumb-${i}`}>
                                        <BreadcrumbSeparator />
                                        <BreadcrumbItem>
                                            {i < crumbs.length - 1 ? (
                                                <BreadcrumbLink asChild>
                                                    <Link to={c.to as any}>{c.label}</Link>
                                                </BreadcrumbLink>
                                            ) : (
                                                <BreadcrumbPage className="capitalize">{c.label}</BreadcrumbPage>
                                            )}
                                        </BreadcrumbItem>
                                    </React.Fragment>
                                ))}
                            </BreadcrumbList>
                        </Breadcrumb>
                    </div>
                    <div />
                </header>
                <main className="p-6">
                    <Outlet />
                </main>
            </SidebarInset>
        </SidebarProvider>
    )
}


