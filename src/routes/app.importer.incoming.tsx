import { useMemo, useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import '../routeTree.gen'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Checkbox } from '@/components/ui/checkbox'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Calendar } from '@/components/ui/calendar'
import { useAuth } from '@/lib/auth'

// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore - File-based route types are injected by the generated route tree
export const Route = createFileRoute('/app/importer/incoming')({
    component: ImporterIncomingPage,
})

type PassportStatus = 'Received' | 'Accepted'

type IncomingPassport = {
    id: string
    lot: string
    form: 'Billet' | 'Ingot' | 'Slab' | 'Sheet' | 'Extrusion'
    siteGln: string
    createdAt: string
    status: PassportStatus
    renewablePct: number
    scope1: number
    scope2: number
    coaCid?: string
    cfpCid?: string
    acceptedBy?: string
    acceptTx?: string
    acceptedAt?: string
}

const INITIAL: IncomingPassport[] = [
    {
        id: 'ALP-2025-00001',
        lot: 'HEAT-1001',
        form: 'Billet',
        siteGln: '1234567890123',
        createdAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 5).toISOString(),
        status: 'Received',
        renewablePct: 58,
        scope1: 0.92,
        scope2: 3.05,
        coaCid: 'ipfs://bafycoacoa1',
        cfpCid: 'ipfs://bafycfpcfp1',
    },
    {
        id: 'ALP-2025-00002',
        lot: 'HEAT-1002',
        form: 'Ingot',
        siteGln: '1234567890456',
        createdAt: new Date(Date.now() - 1000 * 60 * 60 * 48).toISOString(),
        status: 'Accepted',
        renewablePct: 61,
        scope1: 0.84,
        scope2: 2.88,
        coaCid: 'ipfs://bafycoacoa2',
        cfpCid: 'ipfs://bafycfpcfp2',
        acceptedBy: '0xA17F0Bf1a3C9bC0A77bE3E0aC2E4F19cB8bB42d1',
        acceptTx: '0x7f1a2b3c4d5e6f',
        acceptedAt: new Date(Date.now() - 1000 * 60 * 60 * 40).toISOString(),
    },
    {
        id: 'ALP-2025-00003',
        lot: 'HEAT-1003',
        form: 'Sheet',
        siteGln: '2234567890123',
        createdAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 8).toISOString(),
        status: 'Received',
        renewablePct: 55,
        scope1: 0.97,
        scope2: 3.22,
        coaCid: 'ipfs://bafycoacoa3',
        cfpCid: 'ipfs://bafycfpcfp3',
    },
    {
        id: 'ALP-2025-00004',
        lot: 'HEAT-1004',
        form: 'Extrusion',
        siteGln: '3234567890123',
        createdAt: new Date(Date.now() - 1000 * 60 * 60 * 96).toISOString(),
        status: 'Received',
        renewablePct: 63,
        scope1: 0.78,
        scope2: 2.71,
        coaCid: 'ipfs://bafycoacoa4',
        cfpCid: 'ipfs://bafycfpcfp4',
    },
    {
        id: 'ALP-2025-00005',
        lot: 'HEAT-1005',
        form: 'Slab',
        siteGln: '4234567890123',
        createdAt: new Date(Date.now() - 1000 * 60 * 60 * 5).toISOString(),
        status: 'Received',
        renewablePct: 59,
        scope1: 0.91,
        scope2: 2.95,
        coaCid: 'ipfs://bafycoacoa5',
        cfpCid: 'ipfs://bafycfpcfp5',
    },
]

function ImporterIncomingPage() {
    const { accountAddress } = useAuth()
    const [passports, setPassports] = useState<IncomingPassport[]>(INITIAL)
    const [selectedId, setSelectedId] = useState<string | null>(INITIAL[0]?.id ?? null)

    // Access control flags (replace with on-chain checks)
    const hasImporterRole = true
    const supplierApproved = true

    // Filters
    const [search, setSearch] = useState('')
    const [forms, setForms] = useState<Record<IncomingPassport['form'], boolean>>({
        Billet: false,
        Ingot: false,
        Slab: false,
        Sheet: false,
        Extrusion: false,
    })
    const [statusFilter, setStatusFilter] = useState<'All' | PassportStatus>('All')
    const [dateFrom, setDateFrom] = useState<string | null>(null)
    const [dateTo, setDateTo] = useState<string | null>(null)

    const filtered = useMemo(() => {
        const activeForms = Object.entries(forms)
            .filter(([, v]) => v)
            .map(([k]) => k as IncomingPassport['form'])
        return passports.filter((p) => {
            const matchesSearch = [p.id, p.lot].some((v) => v.toLowerCase().includes(search.toLowerCase()))
            const matchesForm = activeForms.length === 0 || activeForms.includes(p.form)
            const created = new Date(p.createdAt)
            const matchesFrom = !dateFrom || created >= new Date(dateFrom)
            const matchesTo = !dateTo || created <= new Date(dateTo)
            const matchesStatus = statusFilter === 'All' || p.status === statusFilter
            return matchesSearch && matchesForm && matchesFrom && matchesTo && matchesStatus
        })
    }, [passports, search, forms, dateFrom, dateTo, statusFilter])

    const selected = useMemo(() => passports.find((p) => p.id === selectedId) ?? null, [passports, selectedId])

    const { toasts, pushToast, removeToast } = useToasts()

    async function onAccept(passportId: string) {
        try {
            if (!hasImporterRole || !supplierApproved) {
                throw new Error('Requires IMPORTER_ROLE and Approved supplier status')
            }
            const { txHash } = await acceptPassport(passportId)
            const nowIso = new Date().toISOString()
            const updated = passports.map((p) =>
                p.id === passportId
                    ? {
                        ...p,
                        status: 'Accepted' as const,
                        acceptedBy: accountAddress ?? undefined,
                        acceptTx: txHash,
                        acceptedAt: nowIso,
                    }
                    : p,
            )
            setPassports(updated)
            pushToast({ type: 'success', title: 'Custody accepted', description: `Tx ${txHash}` })
        } catch (err) {
            const message = err instanceof Error ? err.message : 'Accept failed'
            pushToast({ type: 'error', title: 'Failed to accept', description: message })
        }
    }

    return (
        <div className="space-y-6">
            {/* Header */}
            <header className="space-y-1">
                <h1 className="text-2xl font-semibold tracking-tight">Incoming Passports</h1>
                <p className="text-sm text-muted-foreground">Importer accepts custody before market placement.</p>
            </header>

            {/* Filters */}
            <Card className="border-0 shadow-none">
                <CardContent className="pt-6">
                    <div className="grid grid-cols-1 gap-4 md:grid-cols-4">
                        <div className="space-y-2 md:col-span-1">
                            <Label>Search</Label>
                            <Input placeholder="Passport ID or Lot" value={search} onChange={(e) => setSearch(e.target.value)} />
                        </div>
                        <div className="space-y-2 md:col-span-1">
                            <Label>Form</Label>
                            <div className="flex flex-wrap gap-3">
                                {(['Billet', 'Ingot', 'Slab', 'Sheet', 'Extrusion'] as const).map((f) => (
                                    <label key={f} className="flex items-center gap-2 text-sm">
                                        <Checkbox checked={forms[f]} onCheckedChange={(v) => setForms((s) => ({ ...s, [f]: Boolean(v) }))} />
                                        {f}
                                    </label>
                                ))}
                            </div>
                        </div>
                        <div className="space-y-2 md:col-span-1">
                            <Label>Date Range</Label>
                            <div className="grid grid-cols-2 gap-2">
                                <Popover>
                                    <PopoverTrigger asChild>
                                        <Button variant="outline" className="justify-start font-normal">
                                            {dateFrom ? new Date(dateFrom).toLocaleDateString() : 'From'}
                                        </Button>
                                    </PopoverTrigger>
                                    <PopoverContent className="w-auto p-0" align="start">
                                        <Calendar mode="single" selected={dateFrom ? new Date(dateFrom) : undefined} onSelect={(d) => setDateFrom(d ? d.toISOString().slice(0, 10) : null)} />
                                    </PopoverContent>
                                </Popover>
                                <Popover>
                                    <PopoverTrigger asChild>
                                        <Button variant="outline" className="justify-start font-normal">
                                            {dateTo ? new Date(dateTo).toLocaleDateString() : 'To'}
                                        </Button>
                                    </PopoverTrigger>
                                    <PopoverContent className="w-auto p-0" align="start">
                                        <Calendar mode="single" selected={dateTo ? new Date(dateTo) : undefined} onSelect={(d) => setDateTo(d ? d.toISOString().slice(0, 10) : null)} />
                                    </PopoverContent>
                                </Popover>
                            </div>
                        </div>
                        <div className="space-y-2 md:col-span-1">
                            <Label>Status</Label>
                            <div className="flex items-center gap-3 text-sm">
                                {(['All', 'Received', 'Accepted'] as const).map((s) => (
                                    <label key={s} className="flex items-center gap-2">
                                        <input type="radio" name="status-filter" checked={statusFilter === s} onChange={() => setStatusFilter(s as any)} />
                                        {s}
                                    </label>
                                ))}
                            </div>
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* Grid */}
            <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
                {/* Table */}
                <div className="lg:col-span-2">
                    <Card>
                        <CardHeader>
                            <CardTitle>Incoming</CardTitle>
                            <CardDescription>Filter and select a passport to accept custody.</CardDescription>
                        </CardHeader>
                        <CardContent>
                            <div className="overflow-x-auto rounded-md border">
                                <table className="w-full text-sm">
                                    <thead className="bg-accent/30">
                                        <tr>
                                            <th className="p-2 text-left">Passport ID</th>
                                            <th className="p-2 text-left">Lot/Heat</th>
                                            <th className="p-2 text-left">Form</th>
                                            <th className="p-2 text-left">Origin Site GLN</th>
                                            <th className="p-2 text-left">Created At</th>
                                            <th className="p-2 text-left">Status</th>
                                            <th className="p-2 text-left">Actions</th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {filtered.map((p) => {
                                            const isAccepted = p.status === 'Accepted'
                                            return (
                                                <tr key={p.id} className={`border-t ${selectedId === p.id ? 'bg-accent/20' : ''}`} onClick={() => setSelectedId(p.id)}>
                                                    <td className="p-2 font-mono">{p.id}</td>
                                                    <td className="p-2">{p.lot}</td>
                                                    <td className="p-2">
                                                        <span className="inline-flex items-center rounded-full border px-2 py-0.5 text-xs">{p.form}</span>
                                                    </td>
                                                    <td className="p-2">{p.siteGln}</td>
                                                    <td className="p-2">{new Date(p.createdAt).toLocaleString()}</td>
                                                    <td className="p-2">
                                                        <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs ${isAccepted ? 'border-green-500 text-green-700 border' : 'border-amber-500 text-amber-700 border'}`}>
                                                            {p.status}
                                                        </span>
                                                    </td>
                                                    <td className="p-2">
                                                        <div className="flex flex-wrap gap-2">
                                                            <Button size="sm" variant="outline" disabled={isAccepted || !hasImporterRole || !supplierApproved} onClick={(e) => { e.stopPropagation(); onAccept(p.id) }}>Accept</Button>
                                                            <Button size="sm" variant="ghost" onClick={(e) => { e.stopPropagation(); window.open(`/passport/${p.id}`, '_blank') }}>View</Button>
                                                            <Button size="sm" variant="ghost" disabled={!isAccepted} onClick={(e) => { e.stopPropagation(); window.location.href = `/app/importer/market?pid=${encodeURIComponent(p.id)}` }}>Go to Market</Button>
                                                        </div>
                                                    </td>
                                                </tr>
                                            )
                                        })}
                                        {filtered.length === 0 ? (
                                            <tr>
                                                <td colSpan={7} className="p-6 text-center text-sm text-muted-foreground">No results</td>
                                            </tr>
                                        ) : null}
                                    </tbody>
                                </table>
                            </div>
                        </CardContent>
                    </Card>
                </div>

                {/* Selection panel */}
                <div className="space-y-4">
                    <Card>
                        <CardHeader>
                            <CardTitle>Selection</CardTitle>
                            <CardDescription>Snapshot</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-3 text-sm">
                            {selected ? (
                                <>
                                    <Row label="Passport ID" value={selected.id} mono />
                                    <Row label="Lot / Heat" value={selected.lot} />
                                    <Row label="Form" value={selected.form} />
                                    <Row label="Origin site GLN" value={selected.siteGln} />
                                    <div className="pt-2">
                                        <div className="font-medium">ESG Highlights</div>
                                        <div className="mt-2 grid grid-cols-3 gap-3">
                                            <Metric label="Renewable %" value={`${selected.renewablePct}%`} />
                                            <Metric label="Scope 1" value={String(selected.scope1)} />
                                            <Metric label="Scope 2" value={String(selected.scope2)} />
                                        </div>
                                    </div>
                                    <div className="pt-2">
                                        <div className="font-medium">Linked Docs</div>
                                        <div className="mt-2 flex flex-col gap-2">
                                            <DocLink label="CoA" cid={selected.coaCid} />
                                            <DocLink label="CFP" cid={selected.cfpCid} />
                                        </div>
                                    </div>
                                    <div className="pt-2 flex items-center gap-3">
                                        <Button onClick={() => selected && onAccept(selected.id)} disabled={selected.status === 'Accepted' || !hasImporterRole || !supplierApproved}>Accept</Button>
                                        <Button variant="outline" onClick={() => selected && (window.location.href = `/app/importer/market?pid=${encodeURIComponent(selected.id)}`)} disabled={selected.status !== 'Accepted'}>
                                            Go to Market
                                        </Button>
                                    </div>
                                    {selected.status === 'Accepted' ? (
                                        <div className="rounded-md border p-3 text-xs text-muted-foreground">
                                            {formatAcceptedCopy({ acceptedBy: selected.acceptedBy, acceptTx: selected.acceptTx, acceptedAt: selected.acceptedAt })}
                                        </div>
                                    ) : null}
                                </>
                            ) : (
                                <div className="text-sm text-muted-foreground">No selection</div>
                            )}
                        </CardContent>
                    </Card>
                </div>
            </div>

            {/* Toasts */}
            <ToastViewport toasts={toasts} onClose={removeToast} />
        </div>
    )
}

function Row({ label, value, mono }: { label: string; value?: string | number; mono?: boolean }) {
    if (value == null) return null
    return (
        <div className="flex items-center justify-between gap-4">
            <span className="text-muted-foreground">{label}</span>
            <span className={mono ? 'font-mono font-medium' : 'font-medium'}>{value}</span>
        </div>
    )
}

function Metric({ label, value }: { label: string; value: string }) {
    return (
        <div className="rounded-md border p-3 text-center">
            <div className="text-xs text-muted-foreground">{label}</div>
            <div className="text-sm font-semibold">{value}</div>
        </div>
    )
}

function DocLink({ label, cid }: { label: string; cid?: string }) {
    if (!cid) return (
        <div className="text-xs text-muted-foreground">{label}: —</div>
    )
    const http = `https://ipfs.io/ipfs/${cid.replace('ipfs://', '')}`
    return (
        <a className="text-xs text-blue-600 underline" href={http} target="_blank" rel="noreferrer">
            {label}: {cid}
        </a>
    )
}

function formatAcceptedCopy({ acceptedBy, acceptTx, acceptedAt }: { acceptedBy?: string; acceptTx?: string; acceptedAt?: string }): string {
    const addr = acceptedBy ? shortenHex(acceptedBy) : '—'
    const tx = acceptTx ? shortenHex(acceptTx) : '—'
    const ts = acceptedAt ? new Date(acceptedAt).toLocaleString() : '—'
    return `Accepted by ${addr} · Tx ${tx} · Timestamp ${ts}`
}

function shortenHex(v: string): string {
    if (!v.startsWith('0x')) return v
    return v.slice(0, 6) + '…' + v.slice(-4)
}

async function acceptPassport(_passportId: string): Promise<{ txHash: string }> {
    await new Promise((r) => setTimeout(r, 500))
    const txHash = '0x' + Math.random().toString(16).slice(2, 10)
    return { txHash }
}

type Toast = { id: string; type: 'success' | 'error'; title: string; description?: string }

function useToasts() {
    const [toasts, setToasts] = useState<Toast[]>([])
    function pushToast(t: Omit<Toast, 'id'>) {
        const id = Math.random().toString(36).slice(2, 10)
        const toast = { ...t, id }
        setToasts((s) => [...s, toast])
        setTimeout(() => {
            setToasts((s) => s.filter((x) => x.id !== id))
        }, 4000)
    }
    function removeToast(id: string) {
        setToasts((s) => s.filter((x) => x.id !== id))
    }
    return { toasts, pushToast, removeToast }
}

function ToastViewport({ toasts, onClose }: { toasts: Toast[]; onClose: (id: string) => void }) {
    return (
        <div className="fixed right-4 top-4 z-50 flex w-[360px] max-w-[calc(100vw-2rem)] flex-col gap-2">
            {toasts.map((t) => (
                <div key={t.id} className={`rounded-md border p-3 shadow-md ${t.type === 'success' ? 'border-green-600' : 'border-red-600'}`}>
                    <div className="flex items-start justify-between gap-3">
                        <div>
                            <div className="text-sm font-semibold">{t.title}</div>
                            {t.description ? <div className="text-xs text-muted-foreground">{t.description}</div> : null}
                        </div>
                        <button className="text-xs text-muted-foreground" onClick={() => onClose(t.id)}>Dismiss</button>
                    </div>
                </div>
            ))}
        </div>
    )
}


