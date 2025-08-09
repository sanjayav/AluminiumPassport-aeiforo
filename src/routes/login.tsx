import { createFileRoute, redirect, useRouter } from '@tanstack/react-router'
import { useAuth } from '../lib/auth'
import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Loader2 } from 'lucide-react'

export const Route = createFileRoute('/login')({
    beforeLoad: ({ context }: { context: { auth: ReturnType<typeof useAuth> } }) => {
        const { auth } = context
        if (auth.isAuthenticated) {
            throw redirect({ to: '/app' })
        }
    },
    component: LoginPage,
})

function LoginPage() {
    const { connectWithMetaMask } = useAuth()
    const router = useRouter()
    const [isConnecting, setIsConnecting] = useState(false)
    const [errorMessage, setErrorMessage] = useState<string | null>(null)


    async function handleConnectMetaMask() {
        setErrorMessage(null)
        setIsConnecting(true)
        try {
            await connectWithMetaMask()
            router.navigate({ to: '/app' })
        } catch (err) {
            const message = err instanceof Error ? err.message : 'Failed to connect MetaMask'
            setErrorMessage(message)
        } finally {
            setIsConnecting(false)
        }
    }

    return (
        <div className="relative min-h-screen flex items-center justify-center p-4">
            {/* global background lines (outside the card) */}
            <div
                className="pointer-events-none absolute inset-0 w-full h-full bg-opacity-95 bg-[#181717]/99"
                style={{
                    backgroundImage:
                        'linear-gradient(to right, rgba(255, 255, 255, 0.08) 1px, transparent 1px), linear-gradient(rgba(255, 255, 255, 0.08) 1px, transparent 1px)',
                    backgroundSize: '4rem 4rem',
                }}
            />
            <Card
                className="relative w-full max-w-5xl grid grid-cols-1 lg:grid-cols-2 overflow-hidden rounded-2xl border border-gray-400/30 shadow-[0_20px_60px_-15px_rgba(0,0,0,0.6),0_8px_24px_-10px_rgba(0,0,0,0.45)] bg-background"
            >
                {/* Left side inside card */}
                <div className="relative flex items-center justify-center p-8 bg-background text-foreground min-h-[420px] lg:min-h-[560px]">
                    <div className="relative w-full max-w-md">
                        {/* Brand */}
                        <div className="mb-10 flex items-center gap-3">
                            <img
                                src="https://static.wixstatic.com/media/1fb5ba_40df6745ff4e44438824953013887c71~mv2.png"
                                alt="Company logo"
                                className="h-8 w-8 object-contain"
                            />
                            <img
                                src="https://static.wixstatic.com/media/1fb5ba_5df0c34a230b456abeb3199df837a17d~mv2.png"
                                alt="Company name"
                                className="h-6 object-contain"
                            />
                        </div>
                        {/* Hero copy */}
                        <div className="space-y-6">
                            <h1 className="font-serif italic text-4xl md:text-5xl">Welcome to Marklytics!</h1>
                            <p className="text-sm text-muted-foreground">Empower your green analytics journey.</p>
                            {errorMessage && (
                                <div className="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">
                                    {errorMessage}
                                </div>
                            )}
                            {/* Keep MetaMask login as-is; show spinner while connecting */}
                            <Button onClick={handleConnectMetaMask} className="w-full" disabled={isConnecting}>
                                {isConnecting ? (
                                    <Loader2 className="h-4 w-4 animate-spin" />
                                ) : (
                                    'Login in with MetaMask'
                                )}
                            </Button>
                        </div>
                    </div>
                </div>

                {/* Right side inside card - banner image */}
                <div className="relative min-h-[300px] lg:min-h-[560px]">
                    <div
                        className="absolute inset-0 bg-center bg-cover"
                        style={{ backgroundImage: "url('/login-banner.jpg')" }}
                        aria-label="Login banner"
                    />
                    <div className="absolute inset-0 bg-gradient-to-br from-black/30 via-black/0 to-black/30" />
                </div>
            </Card>
        </div>
    )
}


