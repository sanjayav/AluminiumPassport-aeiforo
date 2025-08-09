import { Outlet, createRootRouteWithContext } from '@tanstack/react-router'
import type { AuthContextValue } from '../lib/auth'

export const Route = createRootRouteWithContext<{ auth: AuthContextValue }>()({
    component: () => <Outlet />,
})


