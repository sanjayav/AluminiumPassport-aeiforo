import React, { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react'

export interface AuthContextValue {
    isAuthenticated: boolean
    accountAddress: string | null
    connectWithMetaMask: () => Promise<void>
    disconnect: () => void
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined)

export function AuthProvider({ children }: { children: React.ReactNode }) {
    const [accountAddress, setAccountAddress] = useState<string | null>(null)

    useEffect(() => {
        const stored = window.localStorage.getItem('walletAddress')
        if (stored) {
            setAccountAddress(stored)
        }
    }, [])

    const connectWithMetaMask = useCallback(async () => {
        type EthereumProvider = {
            request?: <T = unknown>(args: { method: string; params?: unknown[] }) => Promise<T>
        }
        const { ethereum } = window as unknown as { ethereum?: EthereumProvider }
        if (!ethereum?.request) {
            throw new Error('MetaMask not found. Please install the extension.')
        }
        const accounts = await ethereum.request<string[]>({ method: 'eth_requestAccounts' })
        const address = accounts?.[0] ?? null
        if (!address) throw new Error('No account returned from MetaMask')
        setAccountAddress(address)
        window.localStorage.setItem('walletAddress', address)
    }, [])

    const disconnect = useCallback(() => {
        setAccountAddress(null)
        window.localStorage.removeItem('walletAddress')
    }, [])

    const value = useMemo<AuthContextValue>(() => ({
        isAuthenticated: Boolean(accountAddress),
        accountAddress,
        connectWithMetaMask,
        disconnect,
    }), [accountAddress, connectWithMetaMask, disconnect])

    return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

// eslint-disable-next-line react-refresh/only-export-components
export function useAuth(): AuthContextValue {
    const ctx = useContext(AuthContext)
    if (!ctx) throw new Error('useAuth must be used within AuthProvider')
    return ctx
}


