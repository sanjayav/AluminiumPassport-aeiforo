import React, { useEffect, useMemo, useRef, useState } from 'react'
import { createFileRoute, Link } from '@tanstack/react-router'
import '../routeTree.gen'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { motion, AnimatePresence } from 'framer-motion'
import { BadgeCheck, ExternalLink, QrCode, ShieldCheck, Copy, Lock, Search, X, ChevronLeft, ChevronRight } from 'lucide-react'
import { useAuth } from '@/lib/auth'

// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore - file-based route types injected by TanStack Router
export const Route = createFileRoute('/app/auditor/review')({
    component: AuditAttestationPage,
})

type AttestationType = 'ESG Verification' | 'Regulatory' | 'Safety' | 'Other'

type PreviousAttestation = {
    id: string
    nameOrOrg: string
    type: AttestationType
    dateIso: string
    cid: string
    txHash: string
}

function AuditAttestationPage() {
    const { accountAddress, isAuthenticated } = useAuth()

    // Demo role check: treat presence of 'Auditor' in localStorage walletRoles as true; default true if not set to avoid hard-blocking
    const hasAuditorRole = useMemo(() => {
        try {
            const raw = window.localStorage.getItem('walletRoles')
            if (!raw) return true
            const roles = JSON.parse(raw) as string[]
            return roles.includes('Auditor')
        } catch {
            return true
        }
    }, [])

    const [viewTxHash, setViewTxHash] = useState<string | null>(null)
    const [toast, setToast] = useState<{ tone: 'success' | 'error'; message: string } | null>(null)
    const liveRegionRef = useRef<HTMLDivElement | null>(null)

    // Optional: search and filters for previous attestations
    const [searchTerm, setSearchTerm] = useState('')
    const [selectedTypes, setSelectedTypes] = useState<AttestationType[]>([])
    const [dateFrom, setDateFrom] = useState('')
    const [dateTo, setDateTo] = useState('')
    const [rowsPerPage, setRowsPerPage] = useState(10)
    const [page, setPage] = useState(0)

    // Demo dataset
    const allPrev: PreviousAttestation[] = useMemo(
        () => [
            {
                id: '1',
                nameOrOrg: 'GreenCert Labs',
                type: 'ESG Verification',
                dateIso: new Date(Date.now() - 1000 * 60 * 60 * 24 * 7).toISOString(),
                cid: 'ipfs://bafybeiabc123examplecid0001',
                txHash: '0x9a8b7c6d5e4f3a2b1c0d9e8f7a6b5c4d3e2f1a0b',
            },
            {
                id: '2',
                nameOrOrg: 'Regulus Compliance',
                type: 'Regulatory',
                dateIso: new Date(Date.now() - 1000 * 60 * 60 * 24 * 18).toISOString(),
                cid: 'ipfs://bafybeidemo0002cid',
                txHash: '0x1111111111111111111111111111111111111111',
            },
            {
                id: '3',
                nameOrOrg: 'SafeMet Labs',
                type: 'Safety',
                dateIso: new Date(Date.now() - 1000 * 60 * 60 * 24 * 30).toISOString(),
                cid: 'ipfs://bafybeidemo0003cid',
                txHash: '0x2222222222222222222222222222222222222222',
            },
        ],
        [],
    )

    // Debounced search/filtering
    const [filteredPrev, setFilteredPrev] = useState<PreviousAttestation[]>(allPrev)
    useEffect(() => {
        const handle = setTimeout(() => {
            const from = dateFrom ? new Date(dateFrom).getTime() : -Infinity
            const to = dateTo ? new Date(dateTo).getTime() : Infinity
            const next = allPrev.filter((a) => {
                const hay = `${a.nameOrOrg} ${a.type} ${a.cid} ${a.txHash}`.toLowerCase()
                const match = !searchTerm || hay.includes(searchTerm.toLowerCase())
                const inType = selectedTypes.length === 0 || selectedTypes.includes(a.type)
                const t = new Date(a.dateIso).getTime()
                const inDate = t >= from && t <= to
                return match && inType && inDate
            })
            setFilteredPrev(next)
            setPage(0)
        }, 250)
        return () => clearTimeout(handle)
    }, [searchTerm, selectedTypes, dateFrom, dateTo, allPrev])

    // Pagination calculations
    const total = filteredPrev.length
    const startIdx = page * rowsPerPage
    const endIdx = Math.min(total, startIdx + rowsPerPage)
    const pageItems = filteredPrev.slice(startIdx, endIdx)

    // Attestation form
    const [attType, setAttType] = useState<AttestationType | ''>('')
    const [notes, setNotes] = useState('')
    const [cid, setCid] = useState('')
    const [cidError, setCidError] = useState<string | null>(null)
    const [isSubmitting, setIsSubmitting] = useState(false)

    function validateCid(value: string) {
        if (!value) return true
        return value.startsWith('ipfs://')
    }

    async function handleUploadFile(_file: File) {
        // Placeholder: in a real app, upload to IPFS and return the CID
        // Simulate network delay
        await new Promise((r) => setTimeout(r, 600))
        const fakeCid = `ipfs://bafybeidemo${Math.random().toString(36).slice(2, 8)}`
        setCid(fakeCid)
        setCidError(null)
        setToast({ tone: 'success', message: 'File uploaded. CID attached.' })
    }

    async function handleSubmit(e: React.FormEvent) {
        e.preventDefault()
        if (!attType) {
            setToast({ tone: 'error', message: 'Select an attestation type.' })
            return
        }
        if (cid && !validateCid(cid)) {
            setCidError('CID must start with ipfs://')
            setToast({ tone: 'error', message: 'Invalid CID format.' })
            return
        }
        setIsSubmitting(true)
        try {
            // Simulate signing + onchain submission
            await new Promise((r) => setTimeout(r, 900))
            const tx = `0x${Math.random().toString(16).slice(2).padEnd(64, 'a')}`
            setViewTxHash(tx)
            setToast({ tone: 'success', message: 'Attestation submitted successfully.' })
            // Add to the top of previous attestations
            const newItem: PreviousAttestation = {
                id: `${Date.now()}`,
                nameOrOrg: accountAddress ?? 'You',
                type: (attType as AttestationType) || 'Other',
                dateIso: new Date().toISOString(),
                cid: cid || 'ipfs://',
                txHash: tx,
            }
            // Note: updating allPrev (memo) is not ideal; append into filtered for demo purposes
            setFilteredPrev((prev) => [newItem, ...prev])
            setPage(0)
        } catch {
            setToast({ tone: 'error', message: 'Submission failed. Try again.' })
        } finally {
            setIsSubmitting(false)
        }
    }

    useEffect(() => {
        if (!toast) return
        // Announce via aria-live
        liveRegionRef.current?.focus()
        const t = setTimeout(() => setToast(null), 2400)
        return () => clearTimeout(t)
    }, [toast])

    const evidenceCids = ['ipfs://bafybeiabc123examplecid0001', 'ipfs://bafybeidemo0002cid']

    return (
        <div className="space-y-6">
            {/* Header */}
            <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
                <div className="space-y-1">
                    <div className="flex items-center gap-2 text-slate-800">
                        <ShieldCheck className="size-6 text-emerald-600" aria-hidden />
                        <h1 className="text-2xl font-semibold tracking-tight">Audit & Attestation</h1>
                    </div>
                    <p className="text-sm text-slate-500">Review aluminium passport data and attach verification.</p>
                </div>
                <div className="flex gap-2">
                    <Button variant="ghost" className="rounded-lg" asChild>
                        <Link to=".." from="/app/auditor/review">
                            Back to List
                        </Link>
                    </Button>
                    <Button
                        variant="outline"
                        className="rounded-lg"
                        disabled={!viewTxHash}
                        asChild={Boolean(viewTxHash)}
                    >
                        {viewTxHash ? (
                            <a
                                href={`https://etherscan.io/tx/${viewTxHash}`}
                                target="_blank"
                                rel="noreferrer"
                                aria-disabled={!viewTxHash}
                            >
                                <ExternalLink className="size-4" /> View on Blockchain
                            </a>
                        ) : (
                            <span className="inline-flex items-center gap-2 opacity-60">
                                <ExternalLink className="size-4" /> View on Blockchain
                            </span>
                        )}
                    </Button>
                </div>
            </div>

            {/* Optional search & filters */}
            <Card className="rounded-2xl border-slate-200 bg-white/90 shadow-sm">
                <CardContent className="pt-6">
                    <div className="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
                        <div className="flex flex-1 flex-wrap items-center gap-3">
                            <div className="relative w-full max-w-xs">
                                <Input
                                    placeholder="Search (ID/Lot/CID/tx)"
                                    value={searchTerm}
                                    onChange={(e) => setSearchTerm(e.target.value)}
                                    className="pl-9"
                                    aria-label="Search attestations"
                                />
                                <Search className="pointer-events-none absolute left-2 top-1/2 size-4 -translate-y-1/2 text-slate-400" />
                            </div>
                            <div className="flex flex-wrap gap-2">
                                {(['ESG Verification', 'Regulatory', 'Safety', 'Other'] as AttestationType[]).map((t) => {
                                    const active = selectedTypes.includes(t)
                                    return (
                                        <button
                                            key={t}
                                            type="button"
                                            onClick={() =>
                                                setSelectedTypes((prev) =>
                                                    prev.includes(t) ? prev.filter((x) => x !== t) : [...prev, t],
                                                )
                                            }
                                            className={
                                                'rounded-full border px-3 py-1 text-xs transition-colors focus-visible:ring-[3px] focus-visible:ring-ring/50 ' +
                                                (active
                                                    ? 'border-emerald-500 bg-emerald-50 text-emerald-700'
                                                    : 'border-slate-200 bg-white text-slate-600 hover:bg-slate-50')
                                            }
                                            aria-pressed={active}
                                        >
                                            {t}
                                        </button>
                                    )
                                })}
                            </div>
                            <div className="flex items-center gap-2">
                                <Input
                                    type="date"
                                    value={dateFrom}
                                    onChange={(e) => setDateFrom(e.target.value)}
                                    aria-label="From date"
                                    className="w-[11rem]"
                                />
                                <span className="text-slate-400">to</span>
                                <Input
                                    type="date"
                                    value={dateTo}
                                    onChange={(e) => setDateTo(e.target.value)}
                                    aria-label="To date"
                                    className="w-[11rem]"
                                />
                            </div>
                        </div>
                        <div className="flex items-center gap-2">
                            <Button
                                variant="ghost"
                                className="h-9 rounded-lg"
                                onClick={() => {
                                    setSearchTerm('')
                                    setSelectedTypes([])
                                    setDateFrom('')
                                    setDateTo('')
                                }}
                            >
                                <X className="size-4" /> Clear all
                            </Button>
                        </div>
                    </div>
                </CardContent>
            </Card>

            <div className="grid gap-6 lg:grid-cols-3">
                {/* Left: Passport viewer (2 cols) */}
                <div className="lg:col-span-2 space-y-6">
                    <HoverCard>
                        <Card className="rounded-2xl border-slate-200 bg-white shadow-md">
                            <CardHeader className="flex flex-row items-center justify-between">
                                <CardTitle className="text-lg">Batch & Origin</CardTitle>
                                <StatusBadge status="Verified" />
                            </CardHeader>
                            <CardContent>
                                <KeyValueGrid
                                    items={[
                                        ['Passport ID', 'AL-2024-08-00123'],
                                        ['Batch', 'LOT-4F9A'],
                                        ['Origin', 'Western Australia'],
                                        ['Producer', 'Aurora Metals Pty'],
                                        ['Created', new Date().toLocaleString()],
                                        ['Market Placement', 'EU'],
                                    ]}
                                />
                            </CardContent>
                        </Card>
                    </HoverCard>

                    <HoverCard>
                        <Card className="rounded-2xl border-slate-200 bg-white shadow-md">
                            <CardHeader className="flex flex-row items-center justify-between">
                                <CardTitle className="text-lg">Composition & Form</CardTitle>
                                <StatusBadge status="Pending" />
                            </CardHeader>
                            <CardContent>
                                <KeyValueGrid
                                    items={[
                                        ['Purity', '99.7% Al'],
                                        ['Alloy', 'AA 6063'],
                                        ['Form', 'Billet'],
                                        ['Mass', '12.5 t'],
                                    ]}
                                />
                            </CardContent>
                        </Card>
                    </HoverCard>

                    <HoverCard>
                        <Card className="rounded-2xl border-slate-200 bg-white shadow-md">
                            <CardHeader className="flex flex-row items-center justify-between">
                                <CardTitle className="text-lg">ESG</CardTitle>
                                <StatusBadge status="Verified" />
                            </CardHeader>
                            <CardContent>
                                <KeyValueGrid
                                    items={[
                                        ['Scope 1/2', '4.2 tCO₂e/t'],
                                        ['Energy Source', 'Hydro (87%)'],
                                        ['Water Use', '0.8 m³/t'],
                                        ['Certification', 'ISO 14001'],
                                    ]}
                                />
                            </CardContent>
                        </Card>
                    </HoverCard>

                    <HoverCard>
                        <Card className="rounded-2xl border-slate-200 bg-white shadow-md">
                            <CardHeader className="flex flex-row items-center justify-between">
                                <CardTitle className="text-lg">Transport & Market Placement</CardTitle>
                                <StatusBadge status="Flagged" />
                            </CardHeader>
                            <CardContent>
                                <KeyValueGrid
                                    items={[
                                        ['Route', 'Karratha → Hamburg'],
                                        ['Incoterm', 'DAP'],
                                        ['Carrier', 'Oceanic Freight'],
                                        ['Docs', 'CMR #90213'],
                                    ]}
                                />
                            </CardContent>
                        </Card>
                    </HoverCard>

                    <HoverCard>
                        <Card className="rounded-2xl border-slate-200 bg-white shadow-md">
                            <CardHeader className="flex flex-row items-center justify-between">
                                <CardTitle className="text-lg">Linked Evidence</CardTitle>
                                <div className="flex items-center gap-2 text-xs text-slate-500">
                                    <QrCode className="size-4" /> CID attachments
                                </div>
                            </CardHeader>
                            <CardContent>
                                <div className="flex flex-wrap gap-2">
                                    {evidenceCids.map((value) => (
                                        <EvidenceChip key={value} value={value} />
                                    ))}
                                </div>
                            </CardContent>
                        </Card>
                    </HoverCard>
                </div>

                {/* Right: Attestation panel + Previous attestations */}
                <div className="space-y-6">
                    <motion.div initial={{ opacity: 0, y: 8 }} animate={{ opacity: 1, y: 0 }}>
                        <Card
                            className={
                                'relative overflow-hidden rounded-2xl border-slate-200 bg-white shadow-md ' +
                                (viewTxHash ? 'ring-1 ring-emerald-500/50' : '')
                            }
                        >
                            {viewTxHash ? (
                                <div className="absolute inset-x-0 top-0 h-1 bg-emerald-500" aria-hidden />
                            ) : null}
                            <CardHeader>
                                <div className="flex items-center justify-between">
                                    <CardTitle className="text-lg">Attestation</CardTitle>
                                    <div className="flex items-center gap-2 text-xs text-slate-500">
                                        <BadgeCheck className="size-4 text-emerald-600" />
                                        Your signature links this attestation immutably to the passport.
                                    </div>
                                </div>
                            </CardHeader>
                            <CardContent>
                                <form className="space-y-4" onSubmit={handleSubmit} noValidate>
                                    <div className="grid grid-cols-1 gap-4">
                                        <div className="space-y-2">
                                            <label htmlFor="attType" className="text-sm text-slate-600">
                                                Attestation Type
                                            </label>
                                            <Select
                                                onValueChange={(val) => setAttType(val as AttestationType)}
                                                disabled={!isAuthenticated || !hasAuditorRole}
                                            >
                                                <SelectTrigger id="attType" aria-label="Attestation type" aria-invalid={attType === ''}>
                                                    <SelectValue placeholder="Select type" />
                                                </SelectTrigger>
                                                <SelectContent>
                                                    {(['ESG Verification', 'Regulatory', 'Safety', 'Other'] as AttestationType[]).map((t) => (
                                                        <SelectItem key={t} value={t}>
                                                            {t}
                                                        </SelectItem>
                                                    ))}
                                                </SelectContent>
                                            </Select>
                                        </div>

                                        <div className="space-y-2">
                                            <label htmlFor="notes" className="text-sm text-slate-600">
                                                Notes
                                            </label>
                                            <Textarea
                                                id="notes"
                                                maxLength={300}
                                                placeholder="Add context (max 300 chars)"
                                                value={notes}
                                                onChange={(e) => setNotes(e.target.value)}
                                                disabled={!isAuthenticated || !hasAuditorRole}
                                            />
                                            <div className="text-xs text-slate-400">{notes.length}/300</div>
                                        </div>

                                        <div className="space-y-2">
                                            <label htmlFor="cid" className="text-sm text-slate-600">
                                                Supporting Document CID
                                            </label>
                                            <div className="flex items-center gap-2">
                                                <Input
                                                    id="cid"
                                                    placeholder="ipfs://..."
                                                    value={cid}
                                                    onChange={(e) => setCid(e.target.value)}
                                                    aria-invalid={Boolean(cid) && !validateCid(cid)}
                                                    disabled={!isAuthenticated || !hasAuditorRole}
                                                />
                                                <UploadButton onUploaded={handleUploadFile} disabled={!isAuthenticated || !hasAuditorRole} />
                                            </div>
                                            {cidError ? <p className="text-sm text-red-600">{cidError}</p> : null}
                                        </div>
                                    </div>

                                    <Button
                                        type="submit"
                                        disabled={isSubmitting || !isAuthenticated || !hasAuditorRole}
                                        className="w-full rounded-lg bg-emerald-600 text-white hover:bg-emerald-700"
                                    >
                                        {isSubmitting ? 'Signing…' : 'Sign & Submit Attestation'}
                                    </Button>

                                    {viewTxHash ? (
                                        <div className="mt-2 flex flex-wrap items-center gap-2 text-xs">
                                            <span className="rounded-full bg-emerald-50 px-2.5 py-1 text-emerald-700">Success</span>
                                            <span className="rounded-full border border-slate-200 bg-white px-2.5 py-1 text-slate-700">
                                                tx: <a className="underline" href={`https://etherscan.io/tx/${viewTxHash}`} target="_blank" rel="noreferrer">{shorten(viewTxHash)}</a>
                                            </span>
                                        </div>
                                    ) : null}
                                </form>

                                {!hasAuditorRole ? (
                                    <div
                                        className="pointer-events-none absolute inset-0 rounded-2xl bg-white/70 backdrop-blur-[2px]"
                                        aria-hidden
                                    />
                                ) : null}
                                {!hasAuditorRole ? (
                                    <div className="absolute inset-x-6 bottom-6 rounded-lg border border-amber-200 bg-amber-50 p-3 text-amber-800">
                                        <div className="flex items-center gap-2">
                                            <Lock className="size-4" />
                                            <div className="text-sm">
                                                Wallet lacks AUDITOR_ROLE. Manage roles in your profile.
                                                <Button asChild variant="link" className="px-1 text-amber-900 underline">
                                                    <Link to="/app/roles">Go to Roles</Link>
                                                </Button>
                                            </div>
                                        </div>
                                    </div>
                                ) : null}
                            </CardContent>
                        </Card>
                    </motion.div>

                    {/* Previous Attestations */}
                    <Card className="rounded-2xl border-slate-200 bg-white shadow-md">
                        <CardHeader>
                            <CardTitle className="text-lg">Previous Attestations</CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="space-y-3">
                                <AnimatePresence initial={false}>
                                    {pageItems.length === 0 ? (
                                        <motion.div
                                            key="empty"
                                            initial={{ opacity: 0 }}
                                            animate={{ opacity: 1 }}
                                            exit={{ opacity: 0 }}
                                            className="grid place-items-center rounded-lg border border-dashed border-slate-200 p-10 text-center"
                                        >
                                            <div className="flex flex-col items-center gap-2 text-slate-500">
                                                <QrCode className="size-10 opacity-50" />
                                                <div className="text-sm">No attestations yet. Be the first to verify.</div>
                                            </div>
                                        </motion.div>
                                    ) : (
                                        pageItems.map((a) => (
                                            <motion.div
                                                key={a.id}
                                                initial={{ opacity: 0, y: 6 }}
                                                animate={{ opacity: 1, y: 0 }}
                                                exit={{ opacity: 0, y: -6 }}
                                                transition={{ duration: 0.18 }}
                                                className="flex items-center justify-between rounded-xl border border-slate-200 bg-white p-4 transition-shadow hover:shadow-sm"
                                            >
                                                <div className="min-w-0">
                                                    <div className="flex flex-wrap items-center gap-x-2 gap-y-1 text-sm font-medium text-slate-800">
                                                        <span className="truncate">{a.nameOrOrg}</span>
                                                        <span className="text-slate-400">·</span>
                                                        <span className="text-slate-600">{a.type}</span>
                                                        <span className="text-slate-400">·</span>
                                                        <span className="text-slate-600">{formatDate(a.dateIso)}</span>
                                                    </div>
                                                    <div className="mt-1 flex flex-wrap items-center gap-2 text-xs text-slate-600">
                                                        <span className="rounded-full border border-slate-200 bg-slate-50 px-2 py-0.5">
                                                            CID: <a className="underline" href={a.cid} target="_blank" rel="noreferrer">{shorten(a.cid)}</a>
                                                        </span>
                                                        <span className="rounded-full border border-slate-200 bg-slate-50 px-2 py-0.5">
                                                            tx: <a className="underline" href={`https://etherscan.io/tx/${a.txHash}`} target="_blank" rel="noreferrer">{shorten(a.txHash)}</a>
                                                        </span>
                                                    </div>
                                                </div>
                                                <div className="ml-3 shrink-0">
                                                    <a
                                                        className="inline-flex items-center gap-1 rounded-lg border px-3 py-1.5 text-xs hover:bg-slate-50"
                                                        href={`https://etherscan.io/tx/${a.txHash}`}
                                                        target="_blank"
                                                        rel="noreferrer"
                                                    >
                                                        <ExternalLink className="size-3.5" /> Open
                                                    </a>
                                                </div>
                                            </motion.div>
                                        ))
                                    )}
                                </AnimatePresence>
                            </div>

                            {/* Pager */}
                            <div className="mt-4 flex items-center justify-end gap-3">
                                <div className="flex items-center gap-2 text-sm text-slate-600">
                                    <span>Rows per page</span>
                                    <Select onValueChange={(v) => setRowsPerPage(parseInt(v))}>
                                        <SelectTrigger className="h-8 w-[80px]">
                                            <SelectValue placeholder={String(rowsPerPage)} />
                                        </SelectTrigger>
                                        <SelectContent>
                                            {[10, 25, 50].map((n) => (
                                                <SelectItem key={n} value={String(n)}>
                                                    {n}
                                                </SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                </div>
                                <div className="text-sm text-slate-600">
                                    {total === 0 ? '0–0 of 0' : `${startIdx + 1}–${endIdx} of ${total}`}
                                </div>
                                <div className="flex items-center gap-1">
                                    <Button
                                        type="button"
                                        variant="outline"
                                        size="sm"
                                        className="h-8 w-8 p-0"
                                        onClick={() => setPage((p) => Math.max(0, p - 1))}
                                        disabled={page === 0}
                                        aria-label="Previous page"
                                    >
                                        <ChevronLeft className="size-4" />
                                    </Button>
                                    <Button
                                        type="button"
                                        variant="outline"
                                        size="sm"
                                        className="h-8 w-8 p-0"
                                        onClick={() => setPage((p) => (endIdx >= total ? p : p + 1))}
                                        disabled={endIdx >= total}
                                        aria-label="Next page"
                                    >
                                        <ChevronRight className="size-4" />
                                    </Button>
                                </div>
                            </div>
                        </CardContent>
                    </Card>

                    {/* Optional timeline (collapsible simple style) */}
                    <Card className="rounded-2xl border-slate-200 bg-white shadow-md">
                        <CardHeader>
                            <CardTitle className="text-lg">Timeline</CardTitle>
                        </CardHeader>
                        <CardContent>
                            <ol className="relative ml-1 border-l border-slate-200 pl-4">
                                {[
                                    { label: 'Created', ago: '2 mo ago', tx: '0xabc', icon: <QrCode className="size-4" /> },
                                    { label: 'Placed on Market', ago: '6 wk ago', tx: '0xdef', icon: <ExternalLink className="size-4" /> },
                                    { label: 'Attested (you)', ago: viewTxHash ? 'moments ago' : '—', tx: viewTxHash ?? '', icon: <BadgeCheck className="size-4 text-emerald-600" /> },
                                ].map((s, i) => (
                                    <motion.li key={i} initial={{ opacity: 0, x: -6 }} animate={{ opacity: 1, x: 0 }} className="mb-4 ml-2">
                                        <div className="mb-1 flex items-center gap-2 text-sm font-medium text-slate-800">
                                            <span className="grid size-6 place-items-center rounded-full border border-slate-200 bg-white">{s.icon}</span>
                                            {s.label}
                                        </div>
                                        <div className="text-xs text-slate-500">
                                            {s.ago}
                                            {s.tx ? (
                                                <>
                                                    <span className="px-1">·</span>
                                                    <a className="underline" href={`https://etherscan.io/tx/${s.tx}`} target="_blank" rel="noreferrer">
                                                        {shorten(s.tx)}
                                                    </a>
                                                </>
                                            ) : null}
                                        </div>
                                    </motion.li>
                                ))}
                            </ol>
                        </CardContent>
                    </Card>
                </div>
            </div>

            {/* Toast (aria-live) */}
            <div
                ref={liveRegionRef}
                tabIndex={-1}
                aria-live="polite"
                className="pointer-events-none fixed bottom-4 right-4 z-50"
            >
                <AnimatePresence>
                    {toast ? (
                        <motion.div
                            initial={{ opacity: 0, y: 8 }}
                            animate={{ opacity: 1, y: 0 }}
                            exit={{ opacity: 0, y: 8 }}
                            className={
                                'pointer-events-auto grid max-w-sm grid-cols-[auto,1fr] items-center gap-3 rounded-xl border p-3 pr-4 shadow-md ' +
                                (toast.tone === 'success'
                                    ? 'border-emerald-200 bg-emerald-50 text-emerald-900'
                                    : 'border-red-200 bg-red-50 text-red-900')
                            }
                        >
                            {toast.tone === 'success' ? (
                                <BadgeCheck className="size-5" />
                            ) : (
                                <X className="size-5" />
                            )}
                            <div className="text-sm">{toast.message}</div>
                        </motion.div>
                    ) : null}
                </AnimatePresence>
            </div>
        </div>
    )
}

function HoverCard({ children }: { children: React.ReactNode }) {
    return (
        <motion.div initial={{ opacity: 0, y: 8 }} animate={{ opacity: 1, y: 0 }} whileHover={{ y: -2 }} transition={{ duration: 0.2 }}>
            {children}
        </motion.div>
    )
}

function StatusBadge({ status }: { status: 'Verified' | 'Pending' | 'Flagged' }) {
    const styles =
        status === 'Verified'
            ? 'bg-emerald-50 text-emerald-700 border-emerald-200'
            : status === 'Pending'
                ? 'bg-amber-50 text-amber-700 border-amber-200'
                : 'bg-red-50 text-red-700 border-red-200'
    return (
        <span className={`rounded-full border px-2.5 py-1 text-xs ${styles}`}>{status}</span>
    )
}

function KeyValueGrid({ items }: { items: [string, React.ReactNode][] }) {
    return (
        <div className="grid grid-cols-1 gap-x-6 gap-y-3 sm:grid-cols-2">
            {items.map(([k, v]) => (
                <div key={k} className="flex items-baseline justify-between gap-4 rounded-lg bg-slate-50/50 px-3 py-2">
                    <div className="text-xs text-slate-500">{k}</div>
                    <div className="text-sm font-semibold text-slate-800">{v}</div>
                </div>
            ))}
        </div>
    )
}

function EvidenceChip({ value }: { value: string }) {
    const [copied, setCopied] = useState(false)
    async function copy() {
        try {
            await navigator.clipboard.writeText(value)
            setCopied(true)
            setTimeout(() => setCopied(false), 1200)
        } catch {
            // noop
        }
    }
    return (
        <div className="flex items-center gap-1 rounded-full border border-slate-200 bg-white pl-2 pr-1 text-xs text-slate-700">
            <span className="truncate max-w-[12rem]">{value}</span>
            <Tooltip>
                <TooltipTrigger asChild>
                    <button
                        className="rounded-md p-1 hover:bg-slate-100 focus-visible:ring-[3px] focus-visible:ring-ring/50"
                        aria-label="Copy CID"
                        onClick={copy}
                        type="button"
                    >
                        <Copy className="size-3.5" />
                    </button>
                </TooltipTrigger>
                <TooltipContent>{copied ? 'Copied' : 'Copy'}</TooltipContent>
            </Tooltip>
            <Tooltip>
                <TooltipTrigger asChild>
                    <a
                        className="rounded-md p-1 hover:bg-slate-100 focus-visible:ring-[3px] focus-visible:ring-ring/50"
                        aria-label="Open CID"
                        href={value}
                        target="_blank"
                        rel="noreferrer"
                    >
                        <ExternalLink className="size-3.5" />
                    </a>
                </TooltipTrigger>
                <TooltipContent>Open</TooltipContent>
            </Tooltip>
        </div>
    )
}

function UploadButton({ onUploaded, disabled }: { onUploaded: (file: File) => Promise<void> | void; disabled?: boolean }) {
    const inputRef = useRef<HTMLInputElement | null>(null)
    return (
        <>
            <Button
                type="button"
                variant="outline"
                className="rounded-lg"
                onClick={() => inputRef.current?.click()}
                disabled={disabled}
            >
                Upload
            </Button>
            <input
                ref={inputRef}
                type="file"
                hidden
                onChange={(e) => {
                    const file = e.target.files?.[0]
                    if (file) onUploaded(file)
                    e.currentTarget.value = ''
                }}
            />
        </>
    )
}

function formatDate(iso: string) {
    try {
        return new Date(iso).toLocaleString()
    } catch {
        return iso
    }
}

function shorten(v: string) {
    if (!v) return ''
    if (v.startsWith('ipfs://') && v.length > 20) return `${v.slice(0, 14)}…${v.slice(-6)}`
    if (v.startsWith('0x') && v.length > 16) return `${v.slice(0, 8)}…${v.slice(-6)}`
    return v
}


