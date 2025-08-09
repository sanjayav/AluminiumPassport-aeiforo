import React, { useEffect, useMemo, useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import '../routeTree.gen'
import { useForm, useFieldArray, useWatch } from 'react-hook-form'
import * as yup from 'yup'
import { yupResolver } from '@hookform/resolvers/yup'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Calendar } from '@/components/ui/calendar'
import { AnimatePresence, motion } from 'framer-motion'
import { Check, Copy, ExternalLink } from 'lucide-react'

// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore - File-based route types are injected by the generated route tree
export const Route = createFileRoute('/app/refiner/create')({
    component: RefinerCreatePage,
})

type ElementRow = {
    element: string
    percent: number | ''
    tolerance?: string
}

type OptionalDoc = {
    type: 'ASI' | 'EPD' | 'LCA' | 'DoC' | 'Other'
    name?: string
    cid?: string
}

type RefinerFormValues = {
    upstreamBatch?: string
    lotNumber: string
    productForm: 'Billet' | 'Ingot' | 'Slab' | 'Sheet' | 'Extrusion'
    productionDate: string
    siteGln: string
    hsCode: string

    compositionType: 'Primary Purity' | 'Alloy'
    purityPct?: number | ''
    alloyGrade?: string
    elements: ElementRow[]

    renewablePct: number | ''
    scope1: number | ''
    scope2: number | ''
    calcMethod: string

    coaCid?: string
    cfpCid?: string
    isoCid?: string
    optionalDocs: OptionalDoc[]

    confirm: boolean
}

const STEP_LABELS = ['Batch', 'Product', 'ESG', 'Documents', 'Review', 'Mint'] as const

const schema: yup.ObjectSchema<RefinerFormValues> = yup
    .object({
        upstreamBatch: yup.string().optional(),
        lotNumber: yup.string().required('Lot / Heat Number is required'),
        productForm: yup.mixed<RefinerFormValues['productForm']>().oneOf(['Billet', 'Ingot', 'Slab', 'Sheet', 'Extrusion']).required('Form is required'),
        productionDate: yup.string().required('Production date is required'),
        siteGln: yup
            .string()
            .required('Site GLN is required')
            .matches(/^\d{13}$/u, 'GLN must be 13 digits'),
        hsCode: yup.string().required('HS Code is required'),

        compositionType: yup.mixed<'Primary Purity' | 'Alloy'>().oneOf(['Primary Purity', 'Alloy']).required(),
        purityPct: (yup
            .number()
            .typeError('Purity must be a number')
            .min(0)
            .max(100)
            .when('compositionType', {
                is: (v: unknown) => v === 'Primary Purity',
                then: (s) => s.required('Purity is required'),
                otherwise: (s) => s.optional(),
            })) as unknown as yup.AnySchema,
        alloyGrade: yup
            .string()
            .when('compositionType', {
                is: (v: unknown) => v === 'Alloy',
                then: (s) => s.required('Alloy Grade is required'),
                otherwise: (s) => s.optional(),
            }),
        elements: yup
            .array()
            .of(
                yup.object({
                    element: yup.string().required('Element is required'),
                    percent: (yup
                        .number()
                        .typeError('% must be a number')
                        .min(0)
                        .max(100)
                        .required('% is required')) as unknown as yup.AnySchema,
                    tolerance: yup.string().optional(),
                }),
            )
            .default([]),

        renewablePct: yup.number().typeError('Renewable % must be a number').min(0).max(100).required('Renewable % is required') as unknown as yup.AnySchema,
        scope1: yup.number().typeError('Scope 1 must be a number').min(0).required('Scope 1 is required') as unknown as yup.AnySchema,
        scope2: yup.number().typeError('Scope 2 must be a number').min(0).required('Scope 2 is required') as unknown as yup.AnySchema,
        calcMethod: yup.string().required('Calculation Method is required'),

        coaCid: yup
            .string()
            .matches(/^ipfs:\/\/.+/u, { message: 'Invalid CID', excludeEmptyString: true })
            .optional(),
        cfpCid: yup
            .string()
            .matches(/^ipfs:\/\/.+/u, { message: 'Invalid CID', excludeEmptyString: true })
            .optional(),
        isoCid: yup
            .string()
            .matches(/^ipfs:\/\/.+/u, { message: 'Invalid CID', excludeEmptyString: true })
            .optional(),
        optionalDocs: yup
            .array()
            .of(
                yup.object({
                    type: yup.mixed<OptionalDoc['type']>().oneOf(['ASI', 'EPD', 'LCA', 'DoC', 'Other']).required(),
                    name: yup.string().optional(),
                    cid: yup.string().matches(/^ipfs:\/\/.+/u, { message: 'Invalid CID', excludeEmptyString: true }).optional(),
                }),
            )
            .default([]),
        confirm: yup.boolean().default(false),
    })
    .required() as yup.ObjectSchema<RefinerFormValues>

function RefinerCreatePage() {
    const [activeStep, setActiveStep] = useState(0)
    const [minted, setMinted] = useState<{ passportId: string; jsonCid: string } | null>(null)
    const [checksum, setChecksum] = useState<string>('')

    const form = useForm<RefinerFormValues>({
        resolver: yupResolver(schema),
        defaultValues: {
            upstreamBatch: '',
            lotNumber: '',
            productionDate: '',
            siteGln: '',
            hsCode: '',
            purityPct: '' as unknown as number,
            alloyGrade: '',
            elements: [],
            renewablePct: '' as unknown as number,
            scope1: '' as unknown as number,
            scope2: '' as unknown as number,
            calcMethod: '',
            optionalDocs: [],
            confirm: false,
        },
        mode: 'onChange',
    })

    const watched = useWatch({ control: form.control })

    const { fields: elementFields, append: appendElement, remove: removeElement } = useFieldArray({ control: form.control, name: 'elements' })
    const { fields: optDocFields, append: appendOptDoc, remove: removeOptDoc, update: updateOptDoc } = useFieldArray({ control: form.control, name: 'optionalDocs' })

    const completedSteps = useMemo(() => {
        const s = new Set<number>()
        // Step 0: Batch
        if (
            isFilled(watched?.lotNumber) &&
            isFilled(watched?.productForm) &&
            isFilled(watched?.productionDate) &&
            /^\d{13}$/u.test(String(watched?.siteGln ?? '')) &&
            isFilled(watched?.hsCode)
        ) s.add(0)

        // Step 1: Product (Composition)
        if (watched?.compositionType === 'Primary Purity') {
            if (isFilled(watched?.purityPct)) s.add(1)
        } else if (watched?.compositionType === 'Alloy') {
            const hasElement = Array.isArray(watched?.elements) && (watched?.elements ?? []).some((e: any) => isFilled(e?.element) && isFilled(e?.percent))
            if (isFilled(watched?.alloyGrade) && hasElement) s.add(1)
        }

        // Step 2: ESG
        if (isFilled(watched?.renewablePct) && isFilled(watched?.scope1) && isFilled(watched?.scope2) && isFilled(watched?.calcMethod)) s.add(2)

        // Step 3: Documents (require CoA & CFP CIDs)
        if (isValidCid(watched?.coaCid) && isValidCid(watched?.cfpCid)) s.add(3)

        // Step 4: Review
        if (s.has(0) && s.has(1) && s.has(2) && s.has(3)) s.add(4)

        // Step 5 completed only after mint
        if (minted) s.add(5)

        return s
    }, [watched, minted])

    const canProceed = (stepIndex: number) => completedSteps.has(stepIndex)

    useEffect(() => {
        // Recompute checksum whenever entering Review or Mint step or when data changes
        if (activeStep >= 4) {
            computeChecksum({ ...watched, minted: undefined }).then(setChecksum).catch(() => setChecksum(''))
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [activeStep, watched])

    async function handleNext() {
        // Validate subset of fields for the current step
        const fieldsPerStep: (keyof RefinerFormValues)[][] = [
            ['upstreamBatch', 'lotNumber', 'productForm', 'productionDate', 'siteGln', 'hsCode'],
            ['compositionType', 'purityPct', 'alloyGrade', 'elements'],
            ['renewablePct', 'scope1', 'scope2', 'calcMethod'],
            ['coaCid', 'cfpCid', 'isoCid', 'optionalDocs'],
            ['confirm'],
            [],
        ]

        const valid = await form.trigger(fieldsPerStep[activeStep] as any)
        if (!valid) return

        // Documents step hard rule
        if (activeStep === 3) {
            if (!isValidCid(form.getValues('coaCid')) || !isValidCid(form.getValues('cfpCid'))) return
        }

        setActiveStep((s) => Math.min(s + 1, STEP_LABELS.length - 1))
    }

    function handleBack() {
        setActiveStep((s) => Math.max(0, s - 1))
    }

    async function handleMint() {
        // Ensure confirmed
        const confirmed = form.getValues('confirm')
        if (!confirmed) return

        // Prepare JSON and mock IPFS upload
        const json = JSON.stringify({
            type: 'Aluminium Passport',
            standard: 'Digital Product Passport',
            version: '1.0',
            createdAt: new Date().toISOString(),
            data: form.getValues(),
            checksum,
        })
        const jsonCid = await uploadJsonToIpfsMock(json)
        const passportId = 'ALP‑2025‑00001'
        setMinted({ passportId, jsonCid })
        setActiveStep(5)
    }

    const digitalLink = minted ? `https://ipfs.io/ipfs/${minted.jsonCid.replace('ipfs://', '')}` : ''

    return (
        <div className="space-y-6">
            <header className="space-y-1">
                <h1 className="text-2xl font-semibold tracking-tight">Create Refiner Passport</h1>
                <p className="text-sm text-muted-foreground">Link batch, define composition, add ESG and documents, then review and mint.</p>
            </header>

            <Stepper steps={[...STEP_LABELS]} completedSteps={completedSteps} />

            <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
                <div className="space-y-6 lg:col-span-2">
                    {activeStep === 0 && (
                        <Section title="Step 1 – Link Batch" subtitle="Batch and product basics">
                            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                                {/* Upstream batch (optional) */}
                                <div className="space-y-2">
                                    <Label>Select Upstream Batch (optional)</Label>
                                    <Select
                                        defaultValue={form.getValues('upstreamBatch') ? String(form.getValues('upstreamBatch')) : 'none'}
                                        onValueChange={(val) => form.setValue('upstreamBatch', val === 'none' ? '' : val)}
                                    >
                                        <SelectTrigger>
                                            <SelectValue placeholder="Select batch" />
                                        </SelectTrigger>
                                        <SelectContent>
                                            <SelectItem value="none">None</SelectItem>
                                            <SelectItem value="BXT-2025-000123">BXT-2025-000123</SelectItem>
                                        </SelectContent>
                                    </Select>
                                </div>

                                <Field label="Lot / Heat Number">
                                    <Input {...form.register('lotNumber')} placeholder="HEAT-XXXX" />
                                </Field>

                                <div className="space-y-2">
                                    <Label>Form</Label>
                                    <Select
                                        defaultValue={watched?.productForm ? String(watched.productForm) : undefined}
                                        onValueChange={(val) => form.setValue('productForm', val as RefinerFormValues['productForm'], { shouldDirty: true, shouldValidate: true })}
                                    >
                                        <SelectTrigger>
                                            <SelectValue placeholder="Select form" />
                                        </SelectTrigger>
                                        <SelectContent>
                                            {(['Billet', 'Ingot', 'Slab', 'Sheet', 'Extrusion'] as const).map((f) => (
                                                <SelectItem key={f} value={f}>{f}</SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                </div>

                                <Field label="Production Date">
                                    <Popover>
                                        <PopoverTrigger asChild>
                                            <Button variant="outline" className="justify-start font-normal">
                                                {watched?.productionDate ? new Date(watched.productionDate).toLocaleDateString() : 'Pick a date'}
                                            </Button>
                                        </PopoverTrigger>
                                        <PopoverContent className="w-auto p-0" align="start">
                                            <Calendar
                                                mode="single"
                                                selected={watched?.productionDate ? new Date(watched.productionDate) : undefined}
                                                onSelect={(d) => form.setValue('productionDate', d ? d.toISOString().slice(0, 10) : '')}
                                            />
                                        </PopoverContent>
                                    </Popover>
                                </Field>

                                <Field label="Site GLN">
                                    <Input
                                        inputMode="numeric"
                                        maxLength={13}
                                        placeholder="13-digit GLN"
                                        value={form.getValues('siteGln')}
                                        onChange={(e) => {
                                            const onlyDigits = e.target.value.replace(/\D+/g, '').slice(0, 13)
                                            form.setValue('siteGln', onlyDigits)
                                        }}
                                    />
                                </Field>

                                <Field label="HS Code">
                                    <Input {...form.register('hsCode')} placeholder="e.g., 7604.10" />
                                </Field>
                            </div>
                        </Section>
                    )}

                    {activeStep === 1 && (
                        <Section title="Step 2 – Composition" subtitle="Primary purity or alloy composition">
                            <div className="space-y-4">
                                <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                                    <div className="space-y-2">
                                        <Label>Composition Type</Label>
                                        <div className="flex gap-3">
                                            {(['Primary Purity', 'Alloy'] as const).map((t) => (
                                                <label key={t} className="flex items-center gap-2 text-sm">
                                                    <input
                                                        type="radio"
                                                        name="compositionType"
                                                        value={t}
                                                        checked={watched?.compositionType === t}
                                                        onChange={() => form.setValue('compositionType', t)}
                                                    />
                                                    {t}
                                                </label>
                                            ))}
                                        </div>
                                    </div>

                                    {watched?.compositionType === 'Primary Purity' ? (
                                        <Field label="Purity %">
                                            <Input type="number" step="0.01" {...form.register('purityPct', { valueAsNumber: true })} placeholder="e.g., 99.7" />
                                        </Field>
                                    ) : (
                                        <div className="space-y-2">
                                            <Label>Alloy Grade</Label>
                                            <Select
                                                defaultValue={watched?.alloyGrade || undefined}
                                                onValueChange={(v) => form.setValue('alloyGrade', v, { shouldDirty: true, shouldValidate: true })}
                                            >
                                                <SelectTrigger>
                                                    <SelectValue placeholder="Select alloy" />
                                                </SelectTrigger>
                                                <SelectContent>
                                                    {['AA6063', 'AA6061', 'AA6082', 'AA1050', 'Other'].map((g) => (
                                                        <SelectItem key={g} value={g}>
                                                            {g}
                                                        </SelectItem>
                                                    ))}
                                                </SelectContent>
                                            </Select>
                                        </div>
                                    )}
                                </div>

                                {watched?.compositionType === 'Alloy' && (
                                    <div className="space-y-3">
                                        <div className="flex items-center justify-between">
                                            <Label>Element Breakdown</Label>
                                            <Button type="button" variant="outline" size="sm" onClick={() => appendElement({ element: '', percent: '', tolerance: '' })}>
                                                Add Row
                                            </Button>
                                        </div>
                                        <div className="overflow-x-auto rounded-md border">
                                            <table className="w-full text-sm">
                                                <thead className="bg-accent/30">
                                                    <tr>
                                                        <th className="p-2 text-left">Element</th>
                                                        <th className="p-2 text-left">%</th>
                                                        <th className="p-2 text-left">Tolerance</th>
                                                        <th className="p-2" />
                                                    </tr>
                                                </thead>
                                                <tbody>
                                                    {elementFields.map((row, idx) => (
                                                        <tr key={row.id} className="border-t">
                                                            <td className="p-2">
                                                                <Input {...form.register(`elements.${idx}.element` as const)} placeholder="e.g., Si" />
                                                            </td>
                                                            <td className="p-2">
                                                                <Input type="number" step="0.01" {...form.register(`elements.${idx}.percent` as const, { valueAsNumber: true })} placeholder="e.g., 0.7" />
                                                            </td>
                                                            <td className="p-2">
                                                                <Input {...form.register(`elements.${idx}.tolerance` as const)} placeholder="e.g., ±0.05" />
                                                            </td>
                                                            <td className="p-2 text-right">
                                                                <Button type="button" variant="ghost" size="sm" onClick={() => removeElement(idx)}>
                                                                    Remove
                                                                </Button>
                                                            </td>
                                                        </tr>
                                                    ))}
                                                </tbody>
                                            </table>
                                        </div>
                                    </div>
                                )}
                            </div>
                        </Section>
                    )}

                    {activeStep === 2 && (
                        <Section title="Step 3 – ESG" subtitle="Environmental metrics">
                            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                                <Field label="Renewable Energy %">
                                    <Input type="number" step="1" {...form.register('renewablePct', { valueAsNumber: true })} placeholder="e.g., 58" />
                                </Field>
                                <Field label="Scope 1 Emissions (tCO₂e/t)">
                                    <Input type="number" step="0.01" {...form.register('scope1', { valueAsNumber: true })} placeholder="e.g., 0.92" />
                                </Field>
                                <Field label="Scope 2 Emissions (tCO₂e/t)">
                                    <Input type="number" step="0.01" {...form.register('scope2', { valueAsNumber: true })} placeholder="e.g., 3.05" />
                                </Field>
                                <div className="space-y-2">
                                    <Label>Calculation Method</Label>
                                    <Select defaultValue={watched?.calcMethod || undefined} onValueChange={(v) => form.setValue('calcMethod', v, { shouldDirty: true, shouldValidate: true })}>
                                        <SelectTrigger>
                                            <SelectValue placeholder="Select method" />
                                        </SelectTrigger>
                                        <SelectContent>
                                            <SelectItem value="EN ISO 14067:2018">EN ISO 14067:2018</SelectItem>
                                            <SelectItem value="GHG Protocol">GHG Protocol</SelectItem>
                                        </SelectContent>
                                    </Select>
                                </div>
                            </div>
                        </Section>
                    )}

                    {activeStep === 3 && (
                        <Section title="Step 4 – Documents" subtitle="Upload certificates to IPFS">
                            <div className="grid grid-cols-1 gap-6">
                                <DropRow
                                    label="Certificate of Analysis (CoA)"
                                    accept=".pdf,.jpg,.jpeg,.png"
                                    cid={form.getValues('coaCid')}
                                    onUpload={async (file) => form.setValue('coaCid', await uploadToIpfsMock(file, 'QmCoA'))}
                                />
                                <DropRow
                                    label="Carbon Footprint Report (CFP)"
                                    accept=".pdf,.jpg,.jpeg,.png"
                                    cid={form.getValues('cfpCid')}
                                    onUpload={async (file) => form.setValue('cfpCid', await uploadToIpfsMock(file, 'QmCFP'))}
                                />
                                <DropRow
                                    label="ISO 14001 Certificate"
                                    accept=".pdf,.jpg,.jpeg,.png"
                                    cid={form.getValues('isoCid')}
                                    onUpload={async (file) => form.setValue('isoCid', await uploadToIpfsMock(file, 'QmISO'))}
                                />

                                <div className="space-y-3">
                                    <div className="flex items-center justify-between">
                                        <div>
                                            <div className="font-medium">Optional Documents</div>
                                            <div className="text-xs text-muted-foreground">Add ASI, EPD, LCA, DoC or other</div>
                                        </div>
                                        <Button type="button" variant="outline" size="sm" onClick={() => appendOptDoc({ type: 'ASI', name: '' })}>
                                            Add Document
                                        </Button>
                                    </div>
                                    {optDocFields.length > 0 && (
                                        <div className="space-y-3">
                                            {optDocFields.map((f, idx) => (
                                                <Card key={f.id}>
                                                    <CardContent className="grid grid-cols-1 items-end gap-3 pt-6 sm:grid-cols-3">
                                                        <div className="space-y-2">
                                                            <Label>Type</Label>
                                                            <Select
                                                                defaultValue={String(form.getValues(`optionalDocs.${idx}.type` as const))}
                                                                onValueChange={(v) => updateOptDoc(idx, { ...(form.getValues(`optionalDocs.${idx}` as const) as OptionalDoc), type: v as OptionalDoc['type'] })}
                                                            >
                                                                <SelectTrigger>
                                                                    <SelectValue placeholder="Select type" />
                                                                </SelectTrigger>
                                                                <SelectContent>
                                                                    {(['ASI', 'EPD', 'LCA', 'DoC', 'Other'] as const).map((t) => (
                                                                        <SelectItem key={t} value={t}>{t}</SelectItem>
                                                                    ))}
                                                                </SelectContent>
                                                            </Select>
                                                        </div>
                                                        <div className="space-y-2">
                                                            <Label>File</Label>
                                                            <input
                                                                type="file"
                                                                accept=".pdf,.jpg,.jpeg,.png"
                                                                onChange={async (e) => {
                                                                    const file = e.target.files?.[0]
                                                                    if (!file) return
                                                                    const cid = await uploadToIpfsMock(file)
                                                                    const curr = form.getValues(`optionalDocs.${idx}` as const) as OptionalDoc
                                                                    updateOptDoc(idx, { ...curr, name: file.name, cid })
                                                                }}
                                                            />
                                                            {form.getValues(`optionalDocs.${idx}.cid` as const) ? (
                                                                <CidChip cid={form.getValues(`optionalDocs.${idx}.cid` as const) ?? ''} />
                                                            ) : null}
                                                        </div>
                                                        <div className="sm:text-right">
                                                            <Button type="button" variant="ghost" onClick={() => removeOptDoc(idx)}>
                                                                Remove
                                                            </Button>
                                                        </div>
                                                    </CardContent>
                                                </Card>
                                            ))}
                                        </div>
                                    )}
                                </div>

                                {!isValidCid(form.getValues('coaCid')) || !isValidCid(form.getValues('cfpCid')) ? (
                                    <p className="text-sm text-destructive">Upload CoA and CFP to continue.</p>
                                ) : null}
                            </div>
                        </Section>
                    )}

                    {activeStep === 4 && (
                        <Section title="Step 5 – Review & Mint" subtitle="Verify all details before minting">
                            <div className="space-y-4">
                                <div className="rounded-md border p-4 text-sm">
                                    <div className="flex items-center justify-between">
                                        <div className="font-medium">Checksum</div>
                                        <code className="text-xs">{checksum || '—'}</code>
                                    </div>
                                </div>
                                <label className="flex items-center gap-2 text-sm">
                                    <input type="checkbox" checked={!!watched?.confirm} onChange={(e) => form.setValue('confirm', e.target.checked)} />
                                    I confirm all data is correct.
                                </label>
                                <div>
                                    <Button type="button" onClick={handleMint} disabled={!watched?.confirm}>
                                        Mint Passport
                                    </Button>
                                </div>
                            </div>
                        </Section>
                    )}

                    {activeStep === 5 && minted && (
                        <Section title="Minted" subtitle="Passport successfully minted">
                            <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
                                <Card className="md:col-span-2">
                                    <CardHeader>
                                        <CardTitle>Passport</CardTitle>
                                        <CardDescription>Details</CardDescription>
                                    </CardHeader>
                                    <CardContent className="space-y-3 text-sm">
                                        <PreviewRow label="Passport ID" value={minted.passportId} />
                                        <PreviewRow label="CoA CID" value={form.getValues('coaCid') ?? ''} />
                                        <PreviewRow label="CFP CID" value={form.getValues('cfpCid') ?? ''} />
                                        <PreviewRow label="ISO CID" value={form.getValues('isoCid') ?? ''} />
                                        <div className="flex items-center gap-2 pt-2">
                                            <Button type="button" variant="outline" size="sm" onClick={() => navigator.clipboard.writeText(minted.passportId)}>
                                                <Copy className="mr-2 size-4" /> Copy Passport ID
                                            </Button>
                                            <Button type="button" variant="outline" size="sm" onClick={() => window.open(digitalLink, '_blank')}>
                                                <ExternalLink className="mr-2 size-4" /> Open IPFS JSON
                                            </Button>
                                        </div>
                                    </CardContent>
                                </Card>

                                <Card className="grid place-items-center">
                                    <CardHeader>
                                        <CardTitle>QR Code</CardTitle>
                                        <CardDescription>Scan to open</CardDescription>
                                    </CardHeader>
                                    <CardContent className="flex flex-col items-center gap-3">
                                        {digitalLink ? (
                                            <img
                                                alt="QR Code"
                                                className="h-40 w-40"
                                                src={`https://api.qrserver.com/v1/create-qr-code/?size=160x160&data=${encodeURIComponent(digitalLink)}`}
                                            />
                                        ) : null}
                                        <a className="text-xs text-blue-600 underline" href={digitalLink} target="_blank" rel="noreferrer">
                                            Digital Link URL
                                        </a>
                                    </CardContent>
                                </Card>
                            </div>
                        </Section>
                    )}

                    {/* Navigation */}
                    <div className="flex items-center gap-3">
                        {activeStep > 0 ? (
                            <Button variant="outline" type="button" onClick={handleBack}>
                                Back
                            </Button>
                        ) : null}
                        {activeStep < STEP_LABELS.length - 1 ? (
                            <Button type="button" onClick={handleNext} disabled={!canProceed(activeStep)}>
                                Next
                            </Button>
                        ) : null}
                    </div>
                </div>

                {/* Preview */}
                <div>
                    <Card>
                        <CardHeader>
                            <CardTitle>Preview</CardTitle>
                            <CardDescription>Summary across steps</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-3 text-sm">
                            <PreviewRow label="Upstream Batch" value={watched?.upstreamBatch || '—'} />
                            <PreviewRow label="Lot / Heat Number" value={watched?.lotNumber} />
                            <PreviewRow label="Form" value={String(watched?.productForm)} />
                            <PreviewRow label="Production Date" value={watched?.productionDate} />
                            <PreviewRow label="Site GLN" value={watched?.siteGln} />
                            <PreviewRow label="HS Code" value={watched?.hsCode} />
                            <PreviewRow label="Composition" value={watched?.compositionType} />
                            {watched?.compositionType === 'Primary Purity' ? (
                                <PreviewRow label="Purity %" value={String(watched?.purityPct ?? '')} />
                            ) : (
                                <>
                                    <PreviewRow label="Alloy Grade" value={String(watched?.alloyGrade ?? '')} />
                                    {(watched?.elements ?? [])
                                        .filter((e) => e.element)
                                        .map((e, i) => (
                                            <PreviewRow key={`el-${i}`} label={`Element ${e.element}`} value={`${e.percent ?? ''}${e.tolerance ? ` (${e.tolerance})` : ''}`} />
                                        ))}
                                </>
                            )}
                            <PreviewRow label="Renewable %" value={String(watched?.renewablePct ?? '')} />
                            <PreviewRow label="Scope 1" value={String(watched?.scope1 ?? '')} />
                            <PreviewRow label="Scope 2" value={String(watched?.scope2 ?? '')} />
                            <PreviewRow label="Method" value={watched?.calcMethod} />
                            <PreviewRow label="CoA CID" value={watched?.coaCid} />
                            <PreviewRow label="CFP CID" value={watched?.cfpCid} />
                            <PreviewRow label="ISO CID" value={watched?.isoCid} />
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
            <div className="relative grid grid-cols-6 items-center gap-6">
                <div className="pointer-events-none absolute left-0 right-0 top-1/2 -translate-y-1/2">
                    <div className="h-0.5 w-full rounded-full bg-gray-200" />
                    {(() => {
                        const progressPct = ((Math.max(...Array.from(completedSteps.values()).concat([-1])) + 1) / Math.max(1, steps.length - 1)) * 100
                        return (
                            <motion.div
                                initial={false}
                                animate={{ width: `${progressPct}%` }}
                                transition={{ duration: 0.35 }}
                                className="h-0.5 rounded-full bg-[color:var(--color-primary)]"
                                style={{ width: '0%' }}
                            />
                        )
                    })()}
                </div>
                {steps.map((label, i) => (
                    <div key={label} className="relative grid place-items-center">
                        <motion.span
                            aria-label={completedSteps.has(i) ? 'completed' : 'not completed'}
                            className="grid size-8 place-items-center rounded-full border text-xs bg-background"
                            initial={false}
                            animate={{
                                backgroundColor: completedSteps.has(i) ? 'var(--color-primary)' : 'var(--color-background)',
                                color: completedSteps.has(i) ? 'var(--color-primary-foreground)' : 'rgb(17 24 39)',
                                borderColor: completedSteps.has(i) ? 'var(--color-primary)' : 'rgb(229 231 235)',
                            }}
                            transition={{ duration: 0.25 }}
                        >
                            <AnimatePresence initial={false} mode="wait">
                                {completedSteps.has(i) ? (
                                    <motion.span
                                        key="check"
                                        initial={{ scale: 0, rotate: -45, opacity: 0 }}
                                        animate={{ scale: 1, rotate: 0, opacity: 1 }}
                                        exit={{ scale: 0, rotate: 45, opacity: 0 }}
                                        transition={{ type: 'spring', stiffness: 400, damping: 28 }}
                                        className="grid place-items-center"
                                    >
                                        <Check className="size-4" />
                                    </motion.span>
                                ) : (
                                    <motion.span key="index" initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }}>
                                        {i + 1}
                                    </motion.span>
                                )}
                            </AnimatePresence>
                        </motion.span>
                    </div>
                ))}
            </div>
            <div className="grid grid-cols-6 gap-6">
                {steps.map((label) => (
                    <div key={`${label}-title`} className="text-center text-sm font-semibold">
                        {label}
                    </div>
                ))}
            </div>
        </div>
    )
}

function CidChip({ cid }: { cid: string }) {
    if (!cid) return null
    const http = `https://ipfs.io/ipfs/${cid.replace('ipfs://', '')}`
    return (
        <div className="inline-flex items-center gap-2 rounded-full border px-2 py-1 text-xs">
            <span className="font-mono">{cid}</span>
            <Button type="button" size="sm" variant="ghost" onClick={() => navigator.clipboard.writeText(cid)}>
                <Copy className="size-3" />
            </Button>
            <Button type="button" size="sm" variant="ghost" onClick={() => window.open(http, '_blank')}>
                <ExternalLink className="size-3" />
            </Button>
        </div>
    )
}

function DropRow({ label, accept, cid, onUpload }: { label: string; accept: string; cid?: string; onUpload: (file: File) => Promise<void | string> }) {
    return (
        <div className="space-y-2">
            <Label>{label}</Label>
            <div className="rounded-md border border-dashed p-6">
                <input
                    type="file"
                    accept={accept}
                    onChange={async (e) => {
                        const file = e.target.files?.[0]
                        if (!file) return
                        await onUpload(file)
                    }}
                />
                <p className="mt-2 text-xs text-muted-foreground">Accepts: PDF/JPG/PNG</p>
                {cid ? (
                    <div className="mt-3">
                        <CidChip cid={cid} />
                    </div>
                ) : null}
            </div>
        </div>
    )
}

function isFilled(value: unknown): boolean {
    if (value == null) return false
    if (typeof value === 'string') return value.trim().length > 0
    if (typeof value === 'number') return !Number.isNaN(value)
    if (Array.isArray(value)) return value.length > 0
    return Boolean(value)
}

function isValidCid(cid?: string): boolean {
    return !!cid && /^ipfs:\/\/.+/u.test(cid)
}

async function uploadToIpfsMock(_file: File, fixed?: string): Promise<string> {
    // Simulate IPFS upload latency
    await new Promise((r) => setTimeout(r, 500))
    return `ipfs://${fixed ?? 'bafy' + Math.random().toString(36).slice(2, 10)}`
}

async function uploadJsonToIpfsMock(_json: string): Promise<string> {
    await new Promise((r) => setTimeout(r, 500))
    return `ipfs://${'bafy' + Math.random().toString(36).slice(2, 12)}`
}

async function computeChecksum(obj: unknown): Promise<string> {
    try {
        const enc = new TextEncoder()
        const data = enc.encode(JSON.stringify(obj))
        const digest = await crypto.subtle.digest('SHA-256', data)
        const bytes = Array.from(new Uint8Array(digest))
        const hex = bytes.map((b) => b.toString(16).padStart(2, '0')).join('')
        return '0x' + hex
    } catch {
        return ''
    }
}


