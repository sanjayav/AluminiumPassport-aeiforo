import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/app/')({
    component: () => <div>Welcome to the app</div>,
})


