import React, { useEffect, useMemo, useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import '../routeTree.gen'
import { Globe, Archive, PackageCheck, Copy, Loader2 } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Checkbox } from '@/components/ui/checkbox'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { Calendar } from '@/components/ui/calendar'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { QRCodeSVG } from 'qrcode.react'

// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore - File-based route types are injected by the generated route tree
export const Route = createFileRoute('/app/importer/market')({
    component: ImporterMarketPage,
})

type ProductForm = 'Billet' | 'Ingot' | 'Slab' | 'Sheet' | 'Extrusion'
type Passport = {
    id: string
    lot: string
    form: ProductForm
    siteGln: string
    createdAt: string
    renewablePct: number
    scope1: number
    scope2: number
    coaCid?: string
    cfpCid?: string
}

const SAMPLE_PASSPORTS: Passport[] = [
    {
        id: 'ALP-2025-00001',
        lot: 'HEAT-1001',
        form: 'Billet',
        siteGln: '1234567890123',
        createdAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 5).toISOString(),
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
        renewablePct: 61,
        scope1: 0.84,
        scope2: 2.88,
        coaCid: 'ipfs://bafycoacoa2',
        cfpCid: 'ipfs://bafycfpcfp2',
    },
    {
        id: 'ALP-2025-00003',
        lot: 'HEAT-1003',
        form: 'Sheet',
        siteGln: '2234567890123',
        createdAt: new Date(Date.now() - 1000 * 60 * 60 * 24 * 8).toISOString(),
        renewablePct: 55,
        scope1: 0.97,
        scope2: 3.22,
        coaCid: 'ipfs://bafycoacoa3',
        cfpCid: 'ipfs://bafycfpcfp3',
    },
]

type Country = { code: string; name: string }
// EU + EEA + GB
const COUNTRIES: Country[] = [
    { code: 'AT', name: 'Austria' },
    { code: 'BE', name: 'Belgium' },
    { code: 'BG', name: 'Bulgaria' },
    { code: 'HR', name: 'Croatia' },
    { code: 'CY', name: 'Cyprus' },
    { code: 'CZ', name: 'Czechia' },
    { code: 'DK', name: 'Denmark' },
    { code: 'EE', name: 'Estonia' },
    { code: 'FI', name: 'Finland' },
    { code: 'FR', name: 'France' },
    { code: 'DE', name: 'Germany' },
    { code: 'GR', name: 'Greece' },
    { code: 'HU', name: 'Hungary' },
    { code: 'IE', name: 'Ireland' },
    { code: 'IT', name: 'Italy' },
    { code: 'LV', name: 'Latvia' },
    { code: 'LT', name: 'Lithuania' },
    { code: 'LU', name: 'Luxembourg' },
    { code: 'MT', name: 'Malta' },
    { code: 'NL', name: 'Netherlands' },
    { code: 'PL', name: 'Poland' },
    { code: 'PT', name: 'Portugal' },
    { code: 'RO', name: 'Romania' },
    { code: 'SK', name: 'Slovakia' },
    { code: 'SI', name: 'Slovenia' },
    { code: 'ES', name: 'Spain' },
    { code: 'SE', name: 'Sweden' },
    { code: 'IS', name: 'Iceland' },
    { code: 'LI', name: 'Liechtenstein' },
    { code: 'NO', name: 'Norway' },
    { code: 'GB', name: 'United Kingdom' },
]

function getFlagEmoji(iso2: string): string {
    if (!iso2 || iso2.length !== 2) return ''
    const codePoints = iso2
        .toUpperCase()
        .split('')
        .map((char) => 127397 + char.charCodeAt(0))
    return String.fromCodePoint(...codePoints)
}

function isFutureDate(d: Date): boolean {
    const today = new Date()
    today.setHours(0, 0, 0, 0)
    const cmp = new Date(d)
    cmp.setHours(0, 0, 0, 0)
    return cmp.getTime() > today.getTime()
}

function computeGs1CheckDigit17Digits(digits17: string): number {
    // GS1 mod10 algorithm: weights from right: 3,1,3,1...
    let sum = 0
    for (let i = 0; i < digits17.length; i++) {
        const digit = Number(digits17[digits17.length - 1 - i])
        const weight = i % 2 === 0 ? 3 : 1
        sum += digit * weight
    }
    const mod = sum % 10
    const check = mod === 0 ? 0 : 10 - mod
    return check
}

function validateSSCC(sscc: string): boolean {
    if (!/^\d{18}$/.test(sscc)) return false
    const body = sscc.slice(0, 17)
    const check = Number(sscc.slice(17))
    return computeGs1CheckDigit17Digits(body) === check
}

async function uploadToIpfs(_file: File): Promise<{ cid: string }>
// Simulated IPFS upload
{
    await new Promise((r) => setTimeout(r, 800))
    const rand = Math.random().toString(36).slice(2, 10)
    return { cid: `bafy${rand}` }
}

function shortenHex(v?: string): string {
    if (!v) return '—'
    if (!v.startsWith('0x')) return v
    return v.slice(0, 6) + '…' + v.slice(-4)
}

type Toast = { id: string; type: 'success' | 'error' | 'info'; title: string; description?: string }

function useToasts() {
    const [toasts, setToasts] = useState<Toast[]>([])
    function pushToast(t: Omit<Toast, 'id'>) {
        const id = Math.random().toString(36).slice(2, 10)
        const toast = { ...t, id }
        setToasts((s) => [...s, toast])
        setTimeout(() => {
            setToasts((s) => s.filter((x) => x.id !== id))
        }, 4500)
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
                <div
                    key={t.id}
                    className={`rounded-md border p-3 shadow-md ${t.type === 'success'
                            ? 'border-emerald-600 bg-emerald-500/10'
                            : t.type === 'error'
                                ? 'border-red-600 bg-red-500/10'
                                : 'border-amber-600 bg-amber-500/10'
                        }`}
                >
                    <div className="flex items-start justify-between gap-3">
                        <div>
                            <div className="text-sm font-semibold">{t.title}</div>
                            {t.description ? <div className="text-xs text-muted-foreground">{t.description}</div> : null}
                        </div>
                        <button className="text-xs text-muted-foreground" onClick={() => onClose(t.id)}>
                            Dismiss
                        </button>
                    </div>
                </div>
            ))}
        </div>
    )
}

function ImporterMarketPage() {
    const [passports] = useState<Passport[]>(SAMPLE_PASSPORTS)
    const [selectorOpen, setSelectorOpen] = useState(false)
    const [passportSearch, setPassportSearch] = useState('')
    const [selectedId, setSelectedId] = useState<string | null>(null)

    const [country, setCountry] = useState<string>('')
    const [placementDate, setPlacementDate] = useState<Date | undefined>(undefined)
    const [sscc, setSscc] = useState('')
    const [billOfLading, setBillOfLading] = useState('')
    const [evidenceCid, setEvidenceCid] = useState<string | null>(null)
    const [uploading, setUploading] = useState(false)
    const [confirm, setConfirm] = useState(false)
    const [submitting, setSubmitting] = useState(false)
    const [placementTx, setPlacementTx] = useState<string | null>(null)

    const { toasts, pushToast, removeToast } = useToasts()

    // Prefill from query param pid
    useEffect(() => {
        const params = new URLSearchParams(window.location.search)
        const pid = params.get('pid')
        if (pid) setSelectedId(pid)
    }, [])

    const selected = useMemo(() => passports.find((p) => p.id === selectedId) ?? null, [passports, selectedId])

    const filteredPassports = useMemo(() => {
        const term = passportSearch.toLowerCase()
        return passports.filter((p) => [p.id, p.lot, p.form].some((v) => String(v).toLowerCase().includes(term)))
    }, [passports, passportSearch])

    const ssccState: 'empty' | 'valid' | 'invalid' = useMemo(() => {
        if (!sscc) return 'empty'
        return validateSSCC(sscc) ? 'valid' : 'invalid'
    }, [sscc])

    const isValid = Boolean(
        selected && country && placementDate && !isFutureDate(placementDate) && confirm && (sscc ? ssccState === 'valid' : true),
    )

    async function onUpload(file: File) {
        try {
            setUploading(true)
            const { cid } = await uploadToIpfs(file)
            setEvidenceCid(cid)
            pushToast({ type: 'success', title: 'Evidence uploaded', description: `ipfs://${cid}` })
        } catch (err) {
            const message = err instanceof Error ? err.message : 'Upload failed'
            pushToast({ type: 'error', title: 'Upload failed', description: message })
        } finally {
            setUploading(false)
        }
    }

    function onDrop(e: React.DragEvent<HTMLDivElement>) {
        e.preventDefault()
        if (e.dataTransfer.files && e.dataTransfer.files[0]) {
            void onUpload(e.dataTransfer.files[0])
        }
    }

    async function onRecordPlacement() {
        try {
            if (!isValid || !selected || !placementDate) return
            setSubmitting(true)
            await new Promise((r) => setTimeout(r, 1000))
            const txHash = '0x' + Math.random().toString(16).slice(2, 10)
            setPlacementTx(txHash)
            pushToast({ type: 'success', title: 'Placement recorded', description: `Tx ${txHash}` })
        } catch (err) {
            const message = err instanceof Error ? err.message : 'Transaction failed'
            pushToast({ type: 'error', title: 'Failed to record', description: message })
        } finally {
            setSubmitting(false)
        }
    }

    function copy(text: string) {
        void navigator.clipboard.writeText(text)
    }

    return (
        <div className="space-y-6">
            {/* Header */}
            <div className="rounded-2xl border bg-gradient-to-r from-slate-900 to-slate-800 p-6 text-slate-100 shadow-lg">
                <div className="flex items-center justify-between gap-4">
                    <div>
                        <div className="flex items-center gap-3">
                            <Globe className="size-6 text-emerald-400" />
                            <h1 className="text-2xl font-bold tracking-tight">Placement on Market</h1>
                        </div>
                        <p className="mt-1 text-sm text-slate-300">Record official market entry for aluminium passports</p>
                    </div>
                    <div className="flex flex-wrap items-center gap-2">
                        <Button variant="ghost" onClick={() => (window.location.href = '/app/importer/incoming')}>Back to Incoming Shipments</Button>
                        <Button variant="outline" onClick={() => selected && window.open(`/passport/${selected.id}`, '_blank')} disabled={!selected}>
                            View Passport
                        </Button>
                    </div>
                </div>
            </div>

            {/* Grid */}
            <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
                {/* Left column */}
                <div className="space-y-6 lg:col-span-2">
                    {/* Select Passport */}
                    <Card className="rounded-2xl shadow-md">
                        <CardHeader className="bg-gradient-to-r from-slate-700/30 to-slate-600/20 rounded-t-2xl">
                            <CardTitle className="text-xl">Select Passport</CardTitle>
                            <CardDescription>Search and choose a passport to place on market.</CardDescription>
                        </CardHeader>
                        <CardContent className="pt-6">
                            <div className="flex flex-col gap-3">
                                <Popover open={selectorOpen} onOpenChange={setSelectorOpen}>
                                    <PopoverTrigger asChild>
                                        <Button variant="outline" className="justify-between">
                                            <div className="flex items-center gap-3">
                                                {selected ? (
                                                    <div className="flex items-center gap-3">
                                                        <QRCodeSVG value={selected.id} size={24} />
                                                        <div className="text-left">
                                                            <div className="font-mono text-sm font-semibold">{selected.id}</div>
                                                            <div className="text-xs text-muted-foreground">Lot {selected.lot}</div>
                                                        </div>
                                                        <span className="ml-2 inline-flex items-center rounded-full border px-2 py-0.5 text-xs">{selected.form}</span>
                                                    </div>
                                                ) : (
                                                    <span className="text-muted-foreground">Choose passport…</span>
                                                )}
                                            </div>
                                        </Button>
                                    </PopoverTrigger>
                                    <PopoverContent className="w-[420px] p-3">
                                        <div className="space-y-3">
                                            <Input placeholder="Search by ID, Lot, Form" value={passportSearch} onChange={(e) => setPassportSearch(e.target.value)} />
                                            <div className="max-h-64 overflow-auto rounded-md border">
                                                {filteredPassports.map((p) => (
                                                    <button
                                                        key={p.id}
                                                        className="flex w-full items-center gap-3 border-b p-2 text-left hover:bg-accent"
                                                        onClick={() => {
                                                            setSelectedId(p.id)
                                                            setSelectorOpen(false)
                                                        }}
                                                    >
                                                        <QRCodeSVG value={p.id} size={24} />
                                                        <div className="flex-1">
                                                            <div className="font-mono text-sm font-semibold">{p.id}</div>
                                                            <div className="text-xs text-muted-foreground">Lot {p.lot} · GLN {p.siteGln}</div>
                                                        </div>
                                                        <span className="inline-flex items-center rounded-full border px-2 py-0.5 text-xs">{p.form}</span>
                                                    </button>
                                                ))}
                                                {filteredPassports.length === 0 ? (
                                                    <div className="p-3 text-center text-sm text-muted-foreground">No results</div>
                                                ) : null}
                                            </div>
                                        </div>
                                    </PopoverContent>
                                </Popover>
                            </div>
                        </CardContent>
                    </Card>

                    {/* Placement Details */}
                    <Card className="rounded-2xl shadow-md">
                        <CardHeader className="bg-gradient-to-r from-slate-700/30 to-slate-600/20 rounded-t-2xl">
                            <CardTitle className="text-xl">Placement Details</CardTitle>
                            <CardDescription>Destination, date, identifiers and evidence.</CardDescription>
                        </CardHeader>
                        <CardContent className="pt-6 space-y-5">
                            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                                <div className="space-y-2">
                                    <Label>Destination Country <span className="text-red-500">*</span></Label>
                                    <Select value={country} onValueChange={setCountry}>
                                        <SelectTrigger>
                                            <SelectValue placeholder="Select country" />
                                        </SelectTrigger>
                                        <SelectContent>
                                            {COUNTRIES.map((c) => (
                                                <SelectItem key={c.code} value={c.code}>
                                                    <span className="mr-2">{getFlagEmoji(c.code)}</span>
                                                    {c.name}
                                                </SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                </div>
                                <div className="space-y-2">
                                    <Label>Placement Date <span className="text-red-500">*</span></Label>
                                    <Popover>
                                        <PopoverTrigger asChild>
                                            <Button variant="outline" className="justify-start font-normal">
                                                {placementDate ? placementDate.toLocaleDateString() : 'Pick a date'}
                                            </Button>
                                        </PopoverTrigger>
                                        <PopoverContent className="w-auto p-0" align="start">
                                            <div className="p-2">
                                                <Calendar
                                                    mode="single"
                                                    selected={placementDate}
                                                    onSelect={(d) => setPlacementDate(d)}
                                                />
                                                <div className="flex items-center justify-between gap-3 p-2 pt-0">
                                                    <Button size="sm" variant="ghost" onClick={() => setPlacementDate(new Date())}>Today</Button>
                                                    {placementDate && isFutureDate(placementDate) ? (
                                                        <span className="text-xs text-red-600">Cannot be in the future</span>
                                                    ) : null}
                                                </div>
                                            </div>
                                        </PopoverContent>
                                    </Popover>
                                </div>
                                <div className="space-y-2">
                                    <Label>SSCC (18 digits)</Label>
                                    <div className="flex items-center gap-2">
                                        <Input
                                            inputMode="numeric"
                                            pattern="\\d*"
                                            placeholder="e.g. 003123456789012345"
                                            value={sscc}
                                            onChange={(e) => setSscc(e.target.value.replace(/\D/g, '').slice(0, 18))}
                                        />
                                        <span
                                            className={
                                                ssccState === 'valid'
                                                    ? 'inline-flex items-center rounded-full border border-emerald-500/40 bg-emerald-500/10 px-2 py-0.5 text-xs text-emerald-500'
                                                    : ssccState === 'invalid'
                                                        ? 'inline-flex items-center rounded-full border border-red-500/40 bg-red-500/10 px-2 py-0.5 text-xs text-red-500'
                                                        : 'inline-flex items-center rounded-full border border-amber-500/40 bg-amber-500/10 px-2 py-0.5 text-xs text-amber-600'
                                            }
                                        >
                                            {ssccState === 'valid' ? 'Valid' : ssccState === 'invalid' ? 'Invalid' : '—'}
                                        </span>
                                    </div>
                                </div>
                                <div className="space-y-2">
                                    <Label>Bill of Lading</Label>
                                    <Input placeholder="Short reference" value={billOfLading} onChange={(e) => setBillOfLading(e.target.value)} />
                                </div>
                            </div>

                            <div className="space-y-2">
                                <Label>Evidence Upload</Label>
                                <div
                                    className="flex flex-col items-center justify-center gap-3 rounded-lg border border-dashed p-6 text-center hover:bg-accent/30"
                                    onDragOver={(e) => e.preventDefault()}
                                    onDrop={onDrop}
                                >
                                    <Archive className="size-8 text-muted-foreground" />
                                    <div className="text-sm text-muted-foreground">Drag & drop or select a file</div>
                                    <input
                                        type="file"
                                        className="hidden"
                                        id="evidence-file"
                                        onChange={(e) => {
                                            const f = e.target.files?.[0]
                                            if (f) void onUpload(f)
                                        }}
                                    />
                                    <div className="flex items-center gap-2">
                                        <Button variant="outline" size="sm" onClick={() => document.getElementById('evidence-file')?.click()} disabled={uploading}>
                                            {uploading ? (
                                                <span className="inline-flex items-center gap-2"><Loader2 className="size-4 animate-spin" />Uploading…</span>
                                            ) : (
                                                'Choose File'
                                            )}
                                        </Button>
                                        {evidenceCid ? (
                                            <CidChip cid={evidenceCid} onCopy={() => copy(`ipfs://${evidenceCid}`)} />
                                        ) : null}
                                    </div>
                                </div>
                            </div>
                        </CardContent>
                    </Card>

                    {/* Review & Confirmation */}
                    <Card className="rounded-2xl shadow-md">
                        <CardHeader className="bg-gradient-to-r from-slate-700/30 to-slate-600/20 rounded-t-2xl">
                            <CardTitle className="text-xl">Review & Confirmation</CardTitle>
                            <CardDescription>Verify details before recording the placement.</CardDescription>
                        </CardHeader>
                        <CardContent className="pt-6 space-y-5">
                            <div className="overflow-hidden rounded-md border">
                                <div className="grid grid-cols-3 gap-0 text-sm">
                                    <SummaryRow label="Passport" value={selected ? selected.id : '—'} mono />
                                    <SummaryRow label="Destination" value={country ? `${getFlagEmoji(country)} ${COUNTRIES.find((c) => c.code === country)?.name ?? country}` : '—'} />
                                    <SummaryRow label="Date" value={placementDate ? placementDate.toLocaleDateString() : '—'} />
                                    <SummaryRow label="SSCC" value={sscc || '—'} mono />
                                    <SummaryRow label="Bill of Lading" value={billOfLading || '—'} />
                                    <SummaryRow label="Evidence" value={evidenceCid ? `ipfs://${evidenceCid}` : '—'} mono />
                                </div>
                            </div>

                            <label className="flex items-center gap-3 text-sm">
                                <Checkbox checked={confirm} onCheckedChange={(v) => setConfirm(Boolean(v))} />
                                I confirm these details are accurate
                            </label>

                            <div className="pt-2">
                                <Button
                                    className="rounded-full bg-gradient-to-r from-emerald-500 to-cyan-500 px-8 py-6 text-base font-semibold text-white shadow-md hover:from-emerald-600 hover:to-cyan-600 disabled:opacity-50"
                                    onClick={() => void onRecordPlacement()}
                                    disabled={!isValid || submitting}
                                >
                                    {submitting ? (
                                        <span className="inline-flex items-center gap-2"><Loader2 className="size-5 animate-spin" /> Recording…</span>
                                    ) : placementTx ? (
                                        <span className="inline-flex items-center gap-2"><PackageCheck className="size-5 text-white" /> Recorded</span>
                                    ) : (
                                        'Record Placement'
                                    )}
                                </Button>
                            </div>
                        </CardContent>
                    </Card>
                </div>

                {/* Right column */}
                <div className="space-y-6">
                    {/* Timeline */}
                    <Card className="rounded-2xl shadow-md">
                        <CardHeader className="bg-gradient-to-r from-slate-700/30 to-slate-600/20 rounded-t-2xl">
                            <CardTitle className="text-xl">Timeline</CardTitle>
                            <CardDescription>Lifecycle milestones</CardDescription>
                        </CardHeader>
                        <CardContent className="pt-6">
                            <div className="relative ml-3 space-y-6">
                                <TimelineItem
                                    icon={<Archive className="size-4" />}
                                    title="Created (Refiner)"
                                    subtitle={selected ? new Date(selected.createdAt).toLocaleString() : '—'}
                                />
                                <TimelineItem
                                    icon={<Globe className="size-4 text-emerald-500" />}
                                    title={country ? `Placed on Market (${COUNTRIES.find((c) => c.code === country)?.name ?? country})` : 'Placed on Market'}
                                    subtitle={placementTx ? (
                                        <span className="inline-flex items-center gap-2 text-emerald-600">
                                            Tx <a className="font-mono underline" href={`#${placementTx}`} onClick={(e) => e.preventDefault()}>{shortenHex(placementTx)}</a>
                                            <Tooltip>
                                                <TooltipTrigger asChild>
                                                    <button className="text-xs" onClick={() => copy(placementTx)}>
                                                        <Copy className="size-3.5" />
                                                    </button>
                                                </TooltipTrigger>
                                                <TooltipContent>Copy TX</TooltipContent>
                                            </Tooltip>
                                        </span>
                                    ) : '—'}
                                />
                            </div>
                        </CardContent>
                    </Card>

                    {/* Passport Snapshot */}
                    <Card className="rounded-2xl shadow-md">
                        <CardHeader className="bg-gradient-to-r from-slate-700/30 to-slate-600/20 rounded-t-2xl">
                            <CardTitle className="text-xl">Passport Snapshot</CardTitle>
                            <CardDescription>Key attributes</CardDescription>
                        </CardHeader>
                        <CardContent className="pt-6 space-y-3 text-sm">
                            {selected ? (
                                <>
                                    <Row label="Passport ID" value={selected.id} mono />
                                    <Row label="Lot / Heat" value={selected.lot} />
                                    <Row
                                        label="Form"
                                        value={<span className="inline-flex items-center rounded-full border px-2 py-0.5 text-xs">{selected.form}</span>}
                                    />
                                    <Row label="Origin site GLN" value={selected.siteGln} />
                                    <div className="pt-2">
                                        <div className="font-medium">ESG Highlights</div>
                                        <div className="mt-2 grid grid-cols-3 gap-3">
                                            <Metric label="Renewable %" value={`${selected.renewablePct}%`} accent="emerald" />
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
                                    <div className="pt-3 flex flex-wrap gap-2">
                                        <Button size="sm" variant="outline" onClick={() => copy(selected.id)}>Copy Passport ID</Button>
                                        <Button size="sm" variant="outline" onClick={() => sscc && copy(sscc)} disabled={!sscc}>Copy SSCC</Button>
                                        <Button size="sm" variant="outline" onClick={() => evidenceCid && copy(`ipfs://${evidenceCid}`)} disabled={!evidenceCid}>Open Evidence</Button>
                                    </div>
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

function SummaryRow({ label, value, mono }: { label: string; value: React.ReactNode; mono?: boolean }) {
    return (
        <div className="col-span-3 grid grid-cols-[160px_1fr] border-b last:border-b-0">
            <div className="bg-accent/30 p-2 text-xs text-muted-foreground">{label}</div>
            <div className={`p-2 ${mono ? 'font-mono' : ''}`}>{value}</div>
        </div>
    )
}

function TimelineItem({ icon, title, subtitle }: { icon: React.ReactNode; title: string; subtitle?: React.ReactNode }) {
    return (
        <div className="relative">
            <div className="absolute -left-3 top-1.5 flex size-6 items-center justify-center rounded-full border bg-background">
                {icon}
            </div>
            <div className="ml-6">
                <div className="text-sm font-medium">{title}</div>
                <div className="text-xs text-muted-foreground">{subtitle || '—'}</div>
            </div>
        </div>
    )
}

function Row({ label, value, mono }: { label: string; value?: React.ReactNode; mono?: boolean }) {
    if (value == null) return null
    return (
        <div className="flex items-center justify-between gap-4">
            <span className="text-muted-foreground">{label}</span>
            <span className={mono ? 'font-mono font-medium' : 'font-medium'}>{value}</span>
        </div>
    )
}

function Metric({ label, value, accent }: { label: string; value: string; accent?: 'emerald' | 'amber' | 'red' }) {
    const accentClass =
        accent === 'emerald'
            ? 'border-emerald-500/40 bg-emerald-500/10'
            : accent === 'red'
                ? 'border-red-500/40 bg-red-500/10'
                : 'border-amber-500/40 bg-amber-500/10'
    return (
        <div className={`rounded-md border p-3 text-center ${accentClass}`}>
            <div className="text-xs text-muted-foreground">{label}</div>
            <div className="text-sm font-semibold">{value}</div>
        </div>
    )
}

function DocLink({ label, cid }: { label: string; cid?: string }) {
    if (!cid) return <div className="text-xs text-muted-foreground">{label}: —</div>
    const http = `https://ipfs.io/ipfs/${cid.replace('ipfs://', '')}`
    return (
        <a className="text-xs text-blue-600 underline" href={http} target="_blank" rel="noreferrer">
            {label}: {cid}
        </a>
    )
}

function CidChip({ cid, onCopy }: { cid: string; onCopy: () => void }) {
    return (
        <div className="inline-flex items-center gap-2 rounded-full border border-emerald-500/40 bg-emerald-500/10 px-2 py-0.5 text-xs font-mono text-emerald-600">
            ipfs://{cid}
            <Tooltip>
                <TooltipTrigger asChild>
                    <button onClick={onCopy} className="ml-1">
                        <Copy className="size-3.5" />
                    </button>
                </TooltipTrigger>
                <TooltipContent>Copy CID</TooltipContent>
            </Tooltip>
        </div>
    )
}


