// Full upstream page moved here from miner.upstream
import React, { useMemo, useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { useForm, useWatch } from 'react-hook-form'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { motion, AnimatePresence } from 'framer-motion'
import { Check, Copy, ExternalLink } from 'lucide-react'
import * as yup from 'yup'
import { yupResolver } from '@hookform/resolvers/yup'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Calendar } from '@/components/ui/calendar'

// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore - types are injected by TanStack Router generated file
export const Route = createFileRoute('/app/miner')({
    component: UpstreamPage,
})

type UpstreamFormValues = {
    batchId: string
    country: string
    mine: string
    operator: string
    siteGln: string
    dateStart: string
    dateEnd: string
    method: string
    oreGrade: string
    lat: string
    lon: string
    renewablePct: string
    scope12: string
    water: string
    hse: string
    evidenceFiles: FileList | null
    transportMode: string
    transportKm: string
    transportCo2e: string
}

const STEPS = ['Batch', 'Extraction', 'ESG', 'Evidence', 'Transport'] as const
const stepFields: (keyof UpstreamFormValues)[][] = [
    ['batchId', 'country', 'mine', 'operator', 'siteGln', 'dateStart', 'dateEnd'],
    ['method', 'oreGrade', 'lat', 'lon'],
    ['renewablePct', 'scope12', 'water', 'hse'],
    ['evidenceFiles'],
    ['transportMode', 'transportKm', 'transportCo2e'],
]

const schema: yup.ObjectSchema<UpstreamFormValues> = yup
    .object({
        batchId: yup.string().required('Batch ID is required'),
        country: yup.string().required('Country is required'),
        mine: yup.string().required('Mine is required'),
        operator: yup.string().required('Operator is required'),
        siteGln: yup.string().required('Site GLN is required'),
        dateStart: yup.string().required('Start date is required'),
        dateEnd: yup.string().required('End date is required'),
        method: yup.string().required('Method is required'),
        oreGrade: yup
            .number()
            .typeError('Ore grade must be a number')
            .min(0, 'Must be at least 0')
            .required('Ore grade is required') as unknown as yup.AnySchema,
        lat: yup.string().required('Latitude is required'),
        lon: yup.string().required('Longitude is required'),
        renewablePct: yup
            .number()
            .typeError('Renewable % must be a number')
            .min(0)
            .max(100)
            .required('Renewable % is required') as unknown as yup.AnySchema,
        scope12: yup
            .number()
            .typeError('Scope1/2 must be a number')
            .min(0)
            .required('Scope1/2 is required') as unknown as yup.AnySchema,
        water: yup
            .number()
            .typeError('Water must be a number')
            .min(0)
            .required('Water is required') as unknown as yup.AnySchema,
        hse: yup
            .number()
            .typeError('HSE incidents must be a number')
            .min(0)
            .required('HSE is required') as unknown as yup.AnySchema,
        evidenceFiles: yup
            .mixed<FileList>()
            .test('required', 'At least one file is required', (value) => !!value && value.length > 0)
            .nullable()
            .required('Evidence is required'),
        transportMode: yup.string().required('Mode is required'),
        transportKm: yup
            .number()
            .typeError('Distance must be a number')
            .min(0)
            .required('Distance is required') as unknown as yup.AnySchema,
        transportCo2e: yup
            .number()
            .typeError('CO₂e must be a number')
            .min(0)
            .required('CO₂e is required') as unknown as yup.AnySchema,
    })
    .required() as yup.ObjectSchema<UpstreamFormValues>

function UpstreamPage() {
    const [activeStep, setActiveStep] = useState(0)
    const [txId, setTxId] = useState<string | null>(null)
    const [cid, setCid] = useState<string | null>(null)

    const form = useForm<UpstreamFormValues>({
        resolver: yupResolver(schema),
        defaultValues: {
            batchId: '',
            country: '',
            mine: '',
            operator: '',
            siteGln: '',
            dateStart: '',
            dateEnd: '',
            method: '',
            oreGrade: '',
            lat: '',
            lon: '',
            renewablePct: '',
            scope12: '',
            water: '',
            hse: '',
            evidenceFiles: null,
            transportMode: '',
            transportKm: '',
            transportCo2e: '',
        },
        mode: 'onChange',
    })

    const watched = useWatch({ control: form.control })

    const completedSteps = useMemo(() => {
        const filled = (v?: string | number | FileList | null) => v !== undefined && v !== null && String(v).toString().trim().length > 0
        const s = new Set<number>()
        if (filled(watched?.batchId) && filled(watched?.country) && filled(watched?.mine) && filled(watched?.operator) && filled(watched?.siteGln) && filled(watched?.dateStart) && filled(watched?.dateEnd)) s.add(0)
        if (filled(watched?.method) && filled(watched?.oreGrade) && filled(watched?.lat) && filled(watched?.lon)) s.add(1)
        if (filled(watched?.renewablePct) && filled(watched?.scope12) && filled(watched?.water) && filled(watched?.hse)) s.add(2)
        if (watched?.evidenceFiles && watched?.evidenceFiles.length > 0) s.add(3)
        if (filled(watched?.transportMode) && filled(watched?.transportKm) && filled(watched?.transportCo2e)) s.add(4)
        return s
    }, [watched])

    function onRegister() {
        const fakeTx = '0x' + Math.random().toString(16).slice(2).padEnd(16, '0')
        const fakeCid = 'bafy' + Math.random().toString(36).slice(2, 12)
        setTxId(fakeTx)
        setCid(fakeCid)
    }

    const isFilled = (value: unknown) => {
        if (value == null) return false
        if (typeof value === 'string') return value.trim().length > 0
        if (typeof value === 'number') return !Number.isNaN(value)
        if (typeof FileList !== 'undefined' && value instanceof FileList) return value.length > 0
        return Boolean(value)
    }

    const currentFields = stepFields[activeStep]
    const stepIsFilled = currentFields.every((f) => isFilled((watched as any)?.[f]))

    async function handleNext() {
        const valid = await form.trigger(currentFields as (keyof UpstreamFormValues)[])
        if (!valid) return
        setActiveStep((s) => Math.min(s + 1, STEPS.length - 1))
    }

    function handleBack() {
        setActiveStep((s) => Math.max(0, s - 1))
    }

    return (
        <div className="space-y-6">
            <header className="space-y-1">
                <h1 className="text-2xl font-semibold tracking-tight">Register Upstream</h1>
            </header>

            <Stepper steps={[...STEPS]} completedSteps={completedSteps} />

            <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
                <div className="lg:col-span-2 space-y-6">
                    {activeStep === 0 && (
                        <Section title="Step 1 – Batch" subtitle="Batch details">
                            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                                <Field label="Batch ID"><Input {...form.register('batchId')} placeholder="BATCH-0001" /></Field>
                                <Field label="Country"><Input {...form.register('country')} placeholder="e.g., Australia" /></Field>
                                <Field label="Mine"><Input {...form.register('mine')} placeholder="Mine name" /></Field>
                                <Field label="Operator"><Input {...form.register('operator')} placeholder="Operator name" /></Field>
                                <Field label="Site GLN"><Input {...form.register('siteGln')} placeholder="Global Location Number" /></Field>
                                <div className="grid grid-cols-2 gap-4">
                                    <Field label="Start Date">
                                        <Popover>
                                            <PopoverTrigger asChild>
                                                <Button variant="outline" className="justify-start font-normal">
                                                    {watched?.dateStart ? new Date(watched.dateStart).toLocaleDateString() : 'Pick a date'}
                                                </Button>
                                            </PopoverTrigger>
                                            <PopoverContent className="w-auto p-0" align="start">
                                                <Calendar
                                                    mode="single"
                                                    selected={watched?.dateStart ? new Date(watched.dateStart) : undefined}
                                                    onSelect={(d) => form.setValue('dateStart', d ? d.toISOString().slice(0, 10) : '')}
                                                />
                                            </PopoverContent>
                                        </Popover>
                                    </Field>
                                    <Field label="End Date">
                                        <Popover>
                                            <PopoverTrigger asChild>
                                                <Button variant="outline" className="justify-start font-normal">
                                                    {watched?.dateEnd ? new Date(watched.dateEnd).toLocaleDateString() : 'Pick a date'}
                                                </Button>
                                            </PopoverTrigger>
                                            <PopoverContent className="w-auto p-0" align="start">
                                                <Calendar
                                                    mode="single"
                                                    selected={watched?.dateEnd ? new Date(watched.dateEnd) : undefined}
                                                    onSelect={(d) => form.setValue('dateEnd', d ? d.toISOString().slice(0, 10) : '')}
                                                />
                                            </PopoverContent>
                                        </Popover>
                                    </Field>
                                </div>
                            </div>
                        </Section>
                    )}

                    {activeStep === 1 && (
                        <Section title="Step 2 – Extraction" subtitle="Extraction details">
                            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                                <Field label="Method"><Input {...form.register('method')} placeholder="e.g., Open-pit" /></Field>
                                <Field label="Ore grade (Al₂O₃%)"><Input type="number" step="0.01" {...form.register('oreGrade', { valueAsNumber: true })} placeholder="e.g., 45" /></Field>
                                <Field label="Latitude"><Input {...form.register('lat')} placeholder="e.g., -23.12" /></Field>
                                <Field label="Longitude"><Input {...form.register('lon')} placeholder="e.g., 130.99" /></Field>
                            </div>
                        </Section>
                    )}

                    {activeStep === 2 && (
                        <Section title="Step 3 – ESG" subtitle="Sustainability metrics">
                            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                                <Field label="Renewable %"><Input type="number" step="0.1" {...form.register('renewablePct', { valueAsNumber: true })} placeholder="e.g., 60" /></Field>
                                <Field label="Scope1/2 (tCO₂e/t ore)"><Input type="number" step="0.001" {...form.register('scope12', { valueAsNumber: true })} placeholder="e.g., 0.75" /></Field>
                                <Field label="Water (m³/t)"><Input type="number" step="0.01" {...form.register('water', { valueAsNumber: true })} placeholder="e.g., 1.2" /></Field>
                                <Field label="HSE incidents"><Input type="number" step="1" {...form.register('hse', { valueAsNumber: true })} placeholder="e.g., 0" /></Field>
                            </div>
                        </Section>
                    )}

                    {activeStep === 3 && (
                        <Section title="Step 4 – Evidence" subtitle="Upload supporting documents (ISO, Supplier Declaration)">
                            <div className="space-y-3">
                                <div className="rounded-md border border-dashed p-6 text-center">
                                    <input id="evidenceFiles" type="file" multiple onChange={(e) => form.setValue('evidenceFiles', e.target.files)} />
                                    <p className="mt-2 text-xs text-muted-foreground">Drag & drop or click to upload</p>
                                </div>
                                <ul className="text-sm text-muted-foreground">{Array.from(form.getValues('evidenceFiles') ?? []).map((f) => (<li key={f.name}>{f.name}</li>))}</ul>
                            </div>
                        </Section>
                    )}

                    {activeStep === 4 && (
                        <Section title="Step 5 – Transport" subtitle="Outbound movement">
                            <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
                                <Field label="Mode"><Input {...form.register('transportMode')} placeholder="e.g., Truck" /></Field>
                                <Field label="Distance (km)"><Input type="number" step="0.1" {...form.register('transportKm', { valueAsNumber: true })} placeholder="e.g., 250" /></Field>
                                <Field label="CO₂e (t)"><Input type="number" step="0.001" {...form.register('transportCo2e', { valueAsNumber: true })} placeholder="e.g., 1.23" /></Field>
                            </div>
                        </Section>
                    )}

                    <div className="flex items-center gap-3">
                        {activeStep > 0 ? (
                            <Button variant="outline" type="button" onClick={handleBack}>Back</Button>
                        ) : null}
                        {activeStep < STEPS.length - 1 ? (
                            <Button type="button" onClick={handleNext} disabled={!stepIsFilled}>Next</Button>
                        ) : (
                            <Button onClick={form.handleSubmit(onRegister)}>Register Upstream</Button>
                        )}
                        {txId ? <span className="text-sm text-muted-foreground">Tx: {txId}</span> : null}
                        {cid ? <span className="text-sm text-muted-foreground">CID: {cid}</span> : null}
                        {cid ? (<Button variant="outline" className="ml-auto" onClick={() => alert('Shared to Refiner!')}>Share to Refiner</Button>) : null}
                    </div>
                </div>

                <div>
                    <Card>
                        <CardHeader>
                            <CardTitle>Preview</CardTitle>
                            <CardDescription>Summary of the upstream registration</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-3 text-sm">
                            <PreviewRow label="Batch ID" value={watched?.batchId} />
                            <PreviewRow label="Country" value={watched?.country} />
                            <PreviewRow label="Mine" value={watched?.mine} />
                            <PreviewRow label="Operator" value={watched?.operator} />
                            <PreviewRow label="Site GLN" value={watched?.siteGln} />
                            <PreviewRow label="Dates" value={[watched?.dateStart, watched?.dateEnd].filter(Boolean).join(' → ')} />
                            <PreviewRow label="Method" value={watched?.method} />
                            <PreviewRow label="Ore grade (Al₂O₃%)" value={watched?.oreGrade} />
                            <PreviewRow label="Geo" value={[watched?.lat, watched?.lon].filter(Boolean).join(', ')} />
                            <PreviewRow label="Renewable %" value={watched?.renewablePct} />
                            <PreviewRow label="Scope1/2" value={watched?.scope12} />
                            <PreviewRow label="Water" value={watched?.water} />
                            <PreviewRow label="HSE" value={watched?.hse} />
                            <PreviewRow label="Transport" value={[watched?.transportMode, watched?.transportKm + ' km'].filter(Boolean).join(', ')} />
                            <div className="flex items-center gap-2 pt-2">
                                <Button type="button" variant="outline" size="sm" onClick={() => watched?.batchId && navigator.clipboard.writeText(watched.batchId)} disabled={!watched?.batchId}><Copy className="mr-2 size-4" /> Copy Batch ID</Button>
                                <Button type="button" variant="outline" size="sm" onClick={() => cid && window.open(`https://ipfs.io/ipfs/${cid}`, '_blank')} disabled={!cid}><ExternalLink className="mr-2 size-4" /> Open IPFS</Button>
                            </div>
                        </CardContent>
                    </Card>
                </div>
            </div>
        </div>
    )
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
    return (
        <div className="space-y-2">
            <Label>{label}</Label>
            {children}
        </div>
    )
}

function Section({ title, subtitle, children }: { title: string; subtitle?: string; children: React.ReactNode }) {
    return (
        <Card className="border-0 shadow-none">
            <CardHeader>
                <CardTitle className="text-base">{title}</CardTitle>
                {subtitle ? <CardDescription>{subtitle}</CardDescription> : null}
            </CardHeader>
            <CardContent>{children}</CardContent>
        </Card>
    )
}

function PreviewRow({ label, value }: { label: string; value?: string }) {
    if (!value) return null
    return (
        <div className="flex items-center justify-between gap-4">
            <span className="text-muted-foreground">{label}</span>
            <span className="font-medium">{value}</span>
        </div>
    )
}

function Stepper({ steps, completedSteps }: { steps: string[]; completedSteps: Set<number> }) {
    return (
        <div className="space-y-3">
            <div className="relative grid grid-cols-5 items-center gap-6">
                <div className="pointer-events-none absolute left-0 right-0 top-1/2 -translate-y-1/2">
                    <div className="h-0.5 w-full rounded-full bg-gray-200" />
                    {(() => {
                        const progressPct = (Math.max(...Array.from(completedSteps.values()).concat([-1])) + 1) / Math.max(1, steps.length - 1) * 100
                        return (<motion.div initial={false} animate={{ width: `${progressPct}%` }} transition={{ duration: 0.35 }} className="h-0.5 rounded-full bg-[color:var(--color-primary)]" style={{ width: '0%' }} />)
                    })()}
                </div>
                {steps.map((label, i) => (
                    <div key={label} className="relative grid place-items-center">
                        <motion.span aria-label={completedSteps.has(i) ? 'completed' : 'not completed'} className="grid size-8 place-items-center rounded-full border text-xs bg-background" initial={false} animate={{ backgroundColor: completedSteps.has(i) ? 'var(--color-primary)' : 'var(--color-background)', color: completedSteps.has(i) ? 'var(--color-primary-foreground)' : 'rgb(17 24 39)', borderColor: completedSteps.has(i) ? 'var(--color-primary)' : 'rgb(229 231 235)' }} transition={{ duration: 0.25 }}>
                            <AnimatePresence initial={false} mode="wait">
                                {completedSteps.has(i) ? (
                                    <motion.span key="check" initial={{ scale: 0, rotate: -45, opacity: 0 }} animate={{ scale: 1, rotate: 0, opacity: 1 }} exit={{ scale: 0, rotate: 45, opacity: 0 }} transition={{ type: 'spring', stiffness: 400, damping: 28 }} className="grid place-items-center"><Check className="size-4" /></motion.span>
                                ) : (
                                    <motion.span key="index" initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}>{i + 1}</motion.span>
                                )}
                            </AnimatePresence>
                        </motion.span>
                    </div>
                ))}
            </div>
            <div className="grid grid-cols-5 gap-6">{steps.map((label) => (<div key={`${label}-title`} className="text-center text-sm font-semibold">{label}</div>))}</div>
        </div>
    )
}


