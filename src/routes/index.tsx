import { createFileRoute, redirect } from '@tanstack/react-router'

export const Route = createFileRoute('/')({
    beforeLoad: ({ context }: { context: { auth: { isAuthenticated: boolean } } }) => {
        const { auth } = context
        if (auth.isAuthenticated) {
            throw redirect({ to: '/app' })
        }
    },
    component: () => null,
})


