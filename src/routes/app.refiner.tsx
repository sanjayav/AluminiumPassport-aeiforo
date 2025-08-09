import React, { useEffect, useMemo, useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import '../routeTree.gen'
import { useForm, useFieldArray, useWatch } from 'react-hook-form'
import * as yup from 'yup'
import { yupResolver } from '@hookform/resolvers/yup'
import { motion, AnimatePresence } from 'framer-motion'
import { Factory, FileText, Check, Copy, ExternalLink, Clock, BarChart3, QrCode } from 'lucide-react'

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Calendar } from '@/components/ui/calendar'

// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore - File-based route types are injected by the generated route tree
export const Route = createFileRoute('/app/refiner')({
    component: RefinerConsolePage,
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

type CreateFormValues = {
    // Batch/Product basics
    upstreamBatch?: string
    lotNumber: string
    productForm: 'Billet' | 'Ingot' | 'Slab' | 'Sheet' | 'Extrusion'
    productionDate: string
    siteGln: string
    hsCode: string

    // Composition
    compositionType: 'Primary Purity' | 'Alloy'
    purityPct?: number | ''
    alloyGrade?: string
    elements: ElementRow[]

    // ESG
    renewablePct: number | ''
    scope1: number | ''
    scope2: number | ''
    calcMethod: string

    // Documents
    coaCid?: string
    cfpCid?: string
    isoCid?: string
    optionalDocs: OptionalDoc[]

    // Review
    confirm: boolean
}

type UpdateFormValues = CreateFormValues & {
    passportId: string
}

const CREATE_STEPS = ['Batch', 'Product', 'ESG', 'Documents', 'Review', 'Mint'] as const
const UPDATE_STEPS = ['Select', 'Edit', 'Review', 'Update'] as const

const createSchema: yup.ObjectSchema<CreateFormValues> = yup
    .object({
        upstreamBatch: yup.string().optional(),
        lotNumber: yup.string().required('Lot / Heat Number is required'),
        productForm: yup
            .mixed<CreateFormValues['productForm']>()
            .oneOf(['Billet', 'Ingot', 'Slab', 'Sheet', 'Extrusion'])
            .required('Form is required'),
        productionDate: yup.string().required('Production date is required'),
        siteGln: yup.string().required('Site GLN is required').matches(/^\d{13}$/u, 'GLN must be 13 digits'),
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

        coaCid: yup.string().matches(/^ipfs:\/\/.+/u, { message: 'Invalid CID', excludeEmptyString: true }).optional(),
        cfpCid: yup.string().matches(/^ipfs:\/\/.+/u, { message: 'Invalid CID', excludeEmptyString: true }).optional(),
        isoCid: yup.string().matches(/^ipfs:\/\/.+/u, { message: 'Invalid CID', excludeEmptyString: true }).optional(),
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
    .required() as yup.ObjectSchema<CreateFormValues>

function RefinerConsolePage() {
    const [mode, setMode] = useState<'create' | 'update'>('create')

    // Create form state
    const [createStep, setCreateStep] = useState(0)
    const [minted, setMinted] = useState<{ passportId: string; jsonCid: string } | null>(null)
    const [createChecksum, setCreateChecksum] = useState('')

    const createForm = useForm<CreateFormValues>({
        resolver: yupResolver(createSchema),
        defaultValues: {
            upstreamBatch: '',
            lotNumber: '',
            productForm: undefined as unknown as CreateFormValues['productForm'],
            productionDate: '',
            siteGln: '',
            hsCode: '',
            compositionType: 'Primary Purity',
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
    const watchedCreate = useWatch({ control: createForm.control })
    const { fields: elementFields, append: appendElement, remove: removeElement } = useFieldArray({ control: createForm.control, name: 'elements' })
    const { fields: optDocFields, append: appendOptDoc, remove: removeOptDoc, update: updateOptDoc } = useFieldArray({ control: createForm.control, name: 'optionalDocs' })

    const createCompleted = useMemo(() => {
        const s = new Set<number>()
        if (
            isFilled(watchedCreate?.lotNumber) &&
            isFilled(watchedCreate?.productForm) &&
            isFilled(watchedCreate?.productionDate) &&
            /^\d{13}$/u.test(String(watchedCreate?.siteGln ?? '')) &&
            isFilled(watchedCreate?.hsCode)
        ) s.add(0)

        if (watchedCreate?.compositionType === 'Primary Purity') {
            if (isFilled(watchedCreate?.purityPct)) s.add(1)
        } else if (watchedCreate?.compositionType === 'Alloy') {
            const hasElement = Array.isArray(watchedCreate?.elements) && (watchedCreate?.elements ?? []).some((e: any) => isFilled(e?.element) && isFilled(e?.percent))
            if (isFilled(watchedCreate?.alloyGrade) && hasElement) s.add(1)
        }

        if (isFilled(watchedCreate?.renewablePct) && isFilled(watchedCreate?.scope1) && isFilled(watchedCreate?.scope2) && isFilled(watchedCreate?.calcMethod)) s.add(2)

        if (isValidCid(watchedCreate?.coaCid) && isValidCid(watchedCreate?.cfpCid)) s.add(3)

        if (s.has(0) && s.has(1) && s.has(2) && s.has(3)) s.add(4)
        if (minted) s.add(5)
        return s
    }, [watchedCreate, minted])

    useEffect(() => {
        if (createStep >= 4) {
            computeChecksum({ ...watchedCreate, minted: undefined }).then(setCreateChecksum).catch(() => setCreateChecksum(''))
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [createStep, watchedCreate])

    async function handleCreateNext() {
        const fieldsPerStep: (keyof CreateFormValues)[][] = [
            ['upstreamBatch', 'lotNumber', 'productForm', 'productionDate', 'siteGln', 'hsCode'],
            ['compositionType', 'purityPct', 'alloyGrade', 'elements'],
            ['renewablePct', 'scope1', 'scope2', 'calcMethod'],
            ['coaCid', 'cfpCid', 'isoCid', 'optionalDocs'],
            ['confirm'],
            [],
        ]
        const valid = await createForm.trigger(fieldsPerStep[createStep] as any)
        if (!valid) return
        if (createStep === 3) {
            if (!isValidCid(createForm.getValues('coaCid')) || !isValidCid(createForm.getValues('cfpCid'))) return
        }
        setCreateStep((s) => Math.min(s + 1, CREATE_STEPS.length - 1))
    }
    function handleCreateBack() {
        setCreateStep((s) => Math.max(0, s - 1))
    }
    async function handleMint() {
        if (!createForm.getValues('confirm')) return
        const json = JSON.stringify({
            type: 'Aluminium Passport',
            standard: 'Digital Product Passport',
            version: '1.0',
            createdAt: new Date().toISOString(),
            data: createForm.getValues(),
            checksum: createChecksum,
        })
        const jsonCid = await uploadJsonToIpfsMock(json)
        const passportId = generatePassportId()
        setMinted({ passportId, jsonCid })
        pushActivity({ kind: 'created', id: passportId, tx: randomTx(), when: new Date() })
        setCreateStep(5)
    }

    // Update form state
    const [updateStep, setUpdateStep] = useState(0)
    const [updated, setUpdated] = useState<{ passportId: string; jsonCid: string } | null>(null)
    const [baseline, setBaseline] = useState<UpdateFormValues | null>(null)

    const updateForm = useForm<UpdateFormValues>({
        resolver: yupResolver(createSchema.concat(yup.object({ passportId: yup.string().required('Passport ID is required') })) as any),
        defaultValues: {
            passportId: '',
            upstreamBatch: 'BXT-2025-000123',
            lotNumber: 'HEAT-1002',
            productForm: 'Ingot',
            productionDate: new Date().toISOString().slice(0, 10),
            siteGln: '1234567890456',
            hsCode: '7604.10',
            compositionType: 'Alloy',
            alloyGrade: 'AA6063',
            elements: [
                { element: 'Si', percent: 0.7, tolerance: '±0.05' },
                { element: 'Mg', percent: 0.45, tolerance: '±0.03' },
            ],
            renewablePct: 61 as unknown as number,
            scope1: 0.84 as unknown as number,
            scope2: 2.88 as unknown as number,
            calcMethod: 'EN ISO 14067:2018',
            coaCid: 'ipfs://bafycoacoa2',
            cfpCid: 'ipfs://bafycfpcfp2',
            isoCid: '',
            optionalDocs: [],
            confirm: false,
        },
        mode: 'onChange',
    })
    const watchedUpdate = useWatch({ control: updateForm.control })
    const { fields: updElemFields, append: updAppendElement, remove: updRemoveElement } = useFieldArray({ control: updateForm.control, name: 'elements' })
    const { fields: updOptDocFields, append: updAppendOptDoc, remove: updRemoveOptDoc, update: updUpdateOptDoc } = useFieldArray({ control: updateForm.control, name: 'optionalDocs' })

    const [activity, setActivity] = useState<ActivityEvent[]>([
        { kind: 'created', id: 'ALP-2025-00012', when: daysAgo(2), tx: '0x7f1a2b3c4d5e6f' },
        { kind: 'updated', id: 'ALP-2025-00009', when: daysAgo(4), tx: '0x5e6f7a8b9c0d1e' },
    ])

    const stats = useMemo(() => {
        const total = 128
        const pending = 6
        const avgRenewable = average([Number(watchedCreate?.renewablePct) || 0, Number(watchedUpdate?.renewablePct) || 0].filter(Boolean)) || 58
        return { total, pending, avgRenewable }
    }, [watchedCreate?.renewablePct, watchedUpdate?.renewablePct])

    function pushActivity(e: ActivityEvent) {
        setActivity((s) => [e, ...s].slice(0, 5))
    }

    const passportsIndex: Array<{ id: string; lot: string }> = useMemo(
        () => [
            { id: 'ALP-2025-00001', lot: 'HEAT-1001' },
            { id: 'ALP-2025-00002', lot: 'HEAT-1002' },
            { id: 'ALP-2025-00003', lot: 'HEAT-1003' },
            { id: 'ALP-2025-00012', lot: 'HEAT-1014' },
        ],
        [],
    )

    const createCanProceed = (stepIndex: number) => createCompleted.has(stepIndex)

    const updateCompleted = useMemo(() => {
        const s = new Set<number>()
        if (isFilled(watchedUpdate?.passportId)) s.add(0)
        if (s.has(0)) s.add(1)
        if (baseline) s.add(2)
        if (updated) s.add(3)
        return s
    }, [watchedUpdate?.passportId, baseline, updated])

    function handleUpdateBack() {
        setUpdateStep((s) => Math.max(0, s - 1))
    }

    async function handleUpdateNext() {
        if (updateStep === 0) {
            const ok = await updateForm.trigger(['passportId'] as any)
            if (!ok) return
            // Load baseline (mock)
            const base = updateForm.getValues()
            setBaseline({ ...base })
            setUpdateStep(1)
            return
        }
        if (updateStep === 1) {
            // allow proceeding to review
            setUpdateStep(2)
            return
        }
    }

    async function handleDoUpdate() {
        const diff = computeDiff(baseline, updateForm.getValues())
        const json = JSON.stringify({
            type: 'Aluminium Passport Update',
            standard: 'Digital Product Passport',
            version: '1.0',
            updatedAt: new Date().toISOString(),
            passportId: updateForm.getValues('passportId'),
            diff,
        })
        const jsonCid = await uploadJsonToIpfsMock(json)
        const id = updateForm.getValues('passportId') || generatePassportId()
        setUpdated({ passportId: id, jsonCid })
        pushActivity({ kind: 'updated', id, tx: randomTx(), when: new Date() })
        setUpdateStep(3)
    }

    const digitalLinkCreated = minted ? `https://ipfs.io/ipfs/${minted.jsonCid.replace('ipfs://', '')}` : ''
    const digitalLinkUpdated = updated ? `https://ipfs.io/ipfs/${updated.jsonCid.replace('ipfs://', '')}` : ''

    return (
        <div className="space-y-6">
            {/* Header Banner */}
            <div className="relative overflow-hidden rounded-xl border bg-gradient-to-r from-emerald-500/10 to-cyan-500/10">
                <div className="flex items-center gap-4 p-6">
                    <div className="grid size-12 place-items-center rounded-lg bg-gradient-to-br from-emerald-500 to-cyan-500 text-white shadow-md">
                        <Factory className="size-6" />
                    </div>
                    <div>
                        <h1 className="text-2xl font-semibold tracking-tight">Refiner Passport Console</h1>
                        <p className="text-sm text-muted-foreground">Create new or update existing passports</p>
                    </div>
                </div>
            </div>

            {/* Action Cards */}
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                <Card className={`${mode === 'create' ? 'ring-2 ring-emerald-500' : ''} relative overflow-hidden`}>
                    <div className="absolute inset-0 -z-10 bg-gradient-to-br from-emerald-500/10 to-cyan-500/10" />
                    <CardHeader className="flex-row items-center gap-4">
                        <div className="grid size-10 place-items-center rounded-md bg-gradient-to-br from-emerald-500 to-cyan-500 text-white">
                            <Factory className="size-5" />
                        </div>
                        <div>
                            <CardTitle>Create Passport</CardTitle>
                            <CardDescription>Mint a new Digital Product Passport</CardDescription>
                        </div>
                    </CardHeader>
                    <CardContent className="space-y-3">
                        <div className="text-sm text-muted-foreground">Link upstream batch, add composition & ESG, upload CoA/CFP</div>
                        <Button onClick={() => setMode('create')}>Start Create Flow</Button>
                    </CardContent>
                </Card>

                <Card className={`${mode === 'update' ? 'ring-2 ring-cyan-500' : ''}`}>
                    <CardHeader className="flex-row items-center gap-4">
                        <div className="grid size-10 place-items-center rounded-md border">
                            <FileText className="size-5" />
                        </div>
                        <div>
                            <CardTitle>Update Passport</CardTitle>
                            <CardDescription>Edit data or attach new evidence</CardDescription>
                        </div>
                    </CardHeader>
                    <CardContent className="space-y-3">
                        <div className="text-sm text-muted-foreground">Replace CFP, add certificates, fix composition</div>
                        <Button variant="outline" onClick={() => setMode('update')}>Start Update Flow</Button>
                    </CardContent>
                </Card>
            </div>

            {/* Mode Panel + Sidebar */}
            <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
                <div className="lg:col-span-2 space-y-6">
                    <Stepper steps={[...(mode === 'create' ? CREATE_STEPS : UPDATE_STEPS)]} completedSteps={mode === 'create' ? createCompleted : updateCompleted} />

                    <Card className="relative">
                        <CardContent className="space-y-6 pt-6">
                            {mode === 'create' ? (
                                <>
                                    {createStep === 0 && (
                                        <Section title="Step 1 — Link Batch" subtitle="Select upstream or enter manually">
                                            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                                                <div className="space-y-2">
                                                    <Label>Select Upstream Batch (optional)</Label>
                                                    <Select
                                                        defaultValue={createForm.getValues('upstreamBatch') ? String(createForm.getValues('upstreamBatch')) : 'none'}
                                                        onValueChange={(val) => createForm.setValue('upstreamBatch', val === 'none' ? '' : val)}
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
                                                    <Input {...createForm.register('lotNumber')} placeholder="HEAT-XXXX" />
                                                </Field>

                                                <div className="space-y-2">
                                                    <Label>Form</Label>
                                                    <Select
                                                        defaultValue={watchedCreate?.productForm ? String(watchedCreate.productForm) : undefined}
                                                        onValueChange={(val) => createForm.setValue('productForm', val as CreateFormValues['productForm'], { shouldDirty: true, shouldValidate: true })}
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
                                                                {watchedCreate?.productionDate ? new Date(watchedCreate.productionDate).toLocaleDateString() : 'Pick a date'}
                                                            </Button>
                                                        </PopoverTrigger>
                                                        <PopoverContent className="w-auto p-0" align="start">
                                                            <Calendar
                                                                mode="single"
                                                                selected={watchedCreate?.productionDate ? new Date(watchedCreate.productionDate) : undefined}
                                                                onSelect={(d) => createForm.setValue('productionDate', d ? d.toISOString().slice(0, 10) : '')}
                                                            />
                                                        </PopoverContent>
                                                    </Popover>
                                                </Field>

                                                <Field label="Site GLN">
                                                    <Input
                                                        inputMode="numeric"
                                                        maxLength={13}
                                                        placeholder="13-digit GLN"
                                                        value={createForm.getValues('siteGln')}
                                                        onChange={(e) => {
                                                            const only = e.target.value.replace(/\D+/g, '').slice(0, 13)
                                                            createForm.setValue('siteGln', only)
                                                        }}
                                                    />
                                                </Field>

                                                <Field label="HS Code">
                                                    <Input {...createForm.register('hsCode')} placeholder="e.g., 7604.10" />
                                                </Field>
                                            </div>
                                        </Section>
                                    )}

                                    {createStep === 1 && (
                                        <Section title="Step 2 — Composition" subtitle="Primary purity or alloy composition">
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
                                                                        checked={watchedCreate?.compositionType === t}
                                                                        onChange={() => createForm.setValue('compositionType', t)}
                                                                    />
                                                                    {t}
                                                                </label>
                                                            ))}
                                                        </div>
                                                    </div>

                                                    {watchedCreate?.compositionType === 'Primary Purity' ? (
                                                        <Field label="Purity %">
                                                            <Input type="number" step="0.01" {...createForm.register('purityPct', { valueAsNumber: true })} placeholder="e.g., 99.7" />
                                                        </Field>
                                                    ) : (
                                                        <div className="space-y-2">
                                                            <Label>Alloy Grade</Label>
                                                            <Select
                                                                defaultValue={watchedCreate?.alloyGrade || undefined}
                                                                onValueChange={(v) => createForm.setValue('alloyGrade', v, { shouldDirty: true, shouldValidate: true })}
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

                                                {watchedCreate?.compositionType === 'Alloy' && (
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
                                                                                <Input {...createForm.register(`elements.${idx}.element` as const)} placeholder="e.g., Si" />
                                                                            </td>
                                                                            <td className="p-2">
                                                                                <Input type="number" step="0.01" {...createForm.register(`elements.${idx}.percent` as const, { valueAsNumber: true })} placeholder="e.g., 0.7" />
                                                                            </td>
                                                                            <td className="p-2">
                                                                                <Input {...createForm.register(`elements.${idx}.tolerance` as const)} placeholder="e.g., ±0.05" />
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

                                    {createStep === 2 && (
                                        <Section title="Step 3 — ESG" subtitle="Environmental metrics">
                                            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                                                <Field label="Renewable Energy %">
                                                    <Input type="number" step="1" {...createForm.register('renewablePct', { valueAsNumber: true })} placeholder="e.g., 58" />
                                                </Field>
                                                <Field label="Scope 1 Emissions (tCO₂e/t)">
                                                    <Input type="number" step="0.01" {...createForm.register('scope1', { valueAsNumber: true })} placeholder="e.g., 0.92" />
                                                </Field>
                                                <Field label="Scope 2 Emissions (tCO₂e/t)">
                                                    <Input type="number" step="0.01" {...createForm.register('scope2', { valueAsNumber: true })} placeholder="e.g., 3.05" />
                                                </Field>
                                                <div className="space-y-2">
                                                    <Label>Calculation Method</Label>
                                                    <Select defaultValue={watchedCreate?.calcMethod || undefined} onValueChange={(v) => createForm.setValue('calcMethod', v, { shouldDirty: true, shouldValidate: true })}>
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

                                    {createStep === 3 && (
                                        <Section title="Step 4 — Documents" subtitle="Upload certificates to IPFS">
                                            <div className="grid grid-cols-1 gap-6">
                                                <DropRow
                                                    label="Certificate of Analysis (CoA)"
                                                    accept=".pdf,.jpg,.jpeg,.png"
                                                    cid={createForm.getValues('coaCid')}
                                                    onUpload={async (file) => createForm.setValue('coaCid', await uploadToIpfsMock(file, 'QmCoA'))}
                                                />
                                                <DropRow
                                                    label="Carbon Footprint Report (CFP)"
                                                    accept=".pdf,.jpg,.jpeg,.png"
                                                    cid={createForm.getValues('cfpCid')}
                                                    onUpload={async (file) => createForm.setValue('cfpCid', await uploadToIpfsMock(file, 'QmCFP'))}
                                                />
                                                <DropRow
                                                    label="ISO 14001 Certificate"
                                                    accept=".pdf,.jpg,.jpeg,.png"
                                                    cid={createForm.getValues('isoCid')}
                                                    onUpload={async (file) => createForm.setValue('isoCid', await uploadToIpfsMock(file, 'QmISO'))}
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
                                                                                defaultValue={String(createForm.getValues(`optionalDocs.${idx}.type` as const))}
                                                                                onValueChange={(v) => updateOptDoc(idx, { ...(createForm.getValues(`optionalDocs.${idx}` as const) as OptionalDoc), type: v as OptionalDoc['type'] })}
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
                                                                                    const curr = createForm.getValues(`optionalDocs.${idx}` as const) as OptionalDoc
                                                                                    updateOptDoc(idx, { ...curr, name: file.name, cid })
                                                                                }}
                                                                            />
                                                                            {createForm.getValues(`optionalDocs.${idx}.cid` as const) ? (
                                                                                <CidChip cid={createForm.getValues(`optionalDocs.${idx}.cid` as const) ?? ''} />
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

                                                {!isValidCid(createForm.getValues('coaCid')) || !isValidCid(createForm.getValues('cfpCid')) ? (
                                                    <p className="text-sm text-destructive">Upload CoA and CFP to continue.</p>
                                                ) : null}
                                            </div>
                                        </Section>
                                    )}

                                    {createStep === 4 && (
                                        <Section title="Step 5 — Review" subtitle="Confirm and mint">
                                            <div className="space-y-4">
                                                <div className="rounded-md border p-4 text-sm">
                                                    <div className="flex items-center justify-between">
                                                        <div className="font-medium">Checksum</div>
                                                        <code className="text-xs">{createChecksum || '—'}</code>
                                                    </div>
                                                </div>
                                                <label className="flex items-center gap-2 text-sm">
                                                    <input type="checkbox" checked={!!watchedCreate?.confirm} onChange={(e) => createForm.setValue('confirm', e.target.checked)} />
                                                    I confirm all data is correct.
                                                </label>
                                            </div>
                                        </Section>
                                    )}

                                    {createStep === 5 && minted && (
                                        <Section title="Minted" subtitle="Passport successfully minted">
                                            <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
                                                <Card className="md:col-span-2">
                                                    <CardHeader>
                                                        <CardTitle>Passport</CardTitle>
                                                        <CardDescription>Details</CardDescription>
                                                    </CardHeader>
                                                    <CardContent className="space-y-3 text-sm">
                                                        <PreviewRow label="Passport ID" value={minted.passportId} />
                                                        <PreviewRow label="CoA CID" value={createForm.getValues('coaCid') ?? ''} />
                                                        <PreviewRow label="CFP CID" value={createForm.getValues('cfpCid') ?? ''} />
                                                        <PreviewRow label="ISO CID" value={createForm.getValues('isoCid') ?? ''} />
                                                        <div className="flex items-center gap-2 pt-2">
                                                            <Button type="button" variant="outline" size="sm" onClick={() => navigator.clipboard.writeText(minted.passportId)}>
                                                                <Copy className="mr-2 size-4" /> Copy Passport ID
                                                            </Button>
                                                            <Button type="button" variant="outline" size="sm" onClick={() => window.open(digitalLinkCreated, '_blank')}>
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
                                                        {digitalLinkCreated ? (
                                                            <img
                                                                alt="QR Code"
                                                                className="h-40 w-40"
                                                                src={`https://api.qrserver.com/v1/create-qr-code/?size=160x160&data=${encodeURIComponent(digitalLinkCreated)}`}
                                                            />
                                                        ) : null}
                                                        <a className="text-xs text-blue-600 underline" href={digitalLinkCreated} target="_blank" rel="noreferrer">
                                                            Digital Link URL
                                                        </a>
                                                    </CardContent>
                                                </Card>
                                            </div>
                                        </Section>
                                    )}

                                    {/* Sticky Footer */}
                                    <div className="sticky bottom-0 -mx-6 -mb-6 border-t bg-background/80 p-4 backdrop-blur supports-[backdrop-filter]:bg-background/60">
                                        <div className="flex items-center gap-3">
                                            {createStep > 0 ? (
                                                <Button variant="outline" type="button" onClick={handleCreateBack}>Back</Button>
                                            ) : null}
                                            {createStep < 4 ? (
                                                <Button type="button" onClick={handleCreateNext} disabled={!createCanProceed(createStep)}>Next</Button>
                                            ) : createStep === 4 ? (
                                                <Button type="button" onClick={handleMint} disabled={!watchedCreate?.confirm}>Mint</Button>
                                            ) : null}
                                        </div>
                                    </div>
                                </>
                            ) : (
                                <>
                                    {updateStep === 0 && (
                                        <Section title="Step 1 — Select Passport" subtitle="Search by ID or Lot/Heat">
                                            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                                                <Field label="Passport">
                                                    <Select
                                                        defaultValue={updateForm.getValues('passportId') || undefined}
                                                        onValueChange={(v) => updateForm.setValue('passportId', v)}
                                                    >
                                                        <SelectTrigger>
                                                            <SelectValue placeholder="Select passport" />
                                                        </SelectTrigger>
                                                        <SelectContent>
                                                            {passportsIndex.map((p) => (
                                                                <SelectItem key={p.id} value={p.id}>
                                                                    {p.id} · {p.lot}
                                                                </SelectItem>
                                                            ))}
                                                        </SelectContent>
                                                    </Select>
                                                </Field>
                                                <div className="space-y-2">
                                                    <Label>Or enter ID manually</Label>
                                                    <Input value={updateForm.getValues('passportId')} onChange={(e) => updateForm.setValue('passportId', e.target.value)} placeholder="ALP-YYYY-#####" />
                                                </div>
                                            </div>
                                        </Section>
                                    )}

                                    {updateStep === 1 && (
                                        <Section title="Step 2 — Edit Sections" subtitle="Prefilled; changed fields highlighted">
                                            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                                                <Field label="Lot / Heat Number">
                                                    <Input {...updateForm.register('lotNumber')} className={changedCls(baseline?.lotNumber, watchedUpdate?.lotNumber)} />
                                                </Field>
                                                <div className="space-y-2">
                                                    <Label>Form</Label>
                                                    <Select
                                                        defaultValue={watchedUpdate?.productForm ? String(watchedUpdate.productForm) : undefined}
                                                        onValueChange={(val) => updateForm.setValue('productForm', val as UpdateFormValues['productForm'])}
                                                    >
                                                        <SelectTrigger className={changedCls(baseline?.productForm, watchedUpdate?.productForm)}>
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
                                                            <Button variant="outline" className={`justify-start font-normal ${changedCls(baseline?.productionDate, watchedUpdate?.productionDate)}`}>
                                                                {watchedUpdate?.productionDate ? new Date(watchedUpdate.productionDate).toLocaleDateString() : 'Pick a date'}
                                                            </Button>
                                                        </PopoverTrigger>
                                                        <PopoverContent className="w-auto p-0" align="start">
                                                            <Calendar
                                                                mode="single"
                                                                selected={watchedUpdate?.productionDate ? new Date(watchedUpdate.productionDate) : undefined}
                                                                onSelect={(d) => updateForm.setValue('productionDate', d ? d.toISOString().slice(0, 10) : '')}
                                                            />
                                                        </PopoverContent>
                                                    </Popover>
                                                </Field>

                                                <Field label="Site GLN">
                                                    <Input
                                                        inputMode="numeric"
                                                        maxLength={13}
                                                        placeholder="13-digit GLN"
                                                        value={updateForm.getValues('siteGln')}
                                                        onChange={(e) => {
                                                            const only = e.target.value.replace(/\D+/g, '').slice(0, 13)
                                                            updateForm.setValue('siteGln', only)
                                                        }}
                                                        className={changedCls(baseline?.siteGln, watchedUpdate?.siteGln)}
                                                    />
                                                </Field>

                                                <Field label="HS Code">
                                                    <Input {...updateForm.register('hsCode')} className={changedCls(baseline?.hsCode, watchedUpdate?.hsCode)} />
                                                </Field>

                                                <div className="sm:col-span-2 space-y-4">
                                                    <div className="space-y-2">
                                                        <Label>Composition Type</Label>
                                                        <div className="flex gap-3">
                                                            {(['Primary Purity', 'Alloy'] as const).map((t) => (
                                                                <label key={t} className="flex items-center gap-2 text-sm">
                                                                    <input
                                                                        type="radio"
                                                                        name="upd-compositionType"
                                                                        value={t}
                                                                        checked={watchedUpdate?.compositionType === t}
                                                                        onChange={() => updateForm.setValue('compositionType', t)}
                                                                    />
                                                                    {t}
                                                                </label>
                                                            ))}
                                                        </div>
                                                    </div>

                                                    {watchedUpdate?.compositionType === 'Primary Purity' ? (
                                                        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                                                            <Field label="Purity %">
                                                                <Input type="number" step="0.01" {...updateForm.register('purityPct', { valueAsNumber: true })} className={changedCls(baseline?.purityPct, watchedUpdate?.purityPct)} />
                                                            </Field>
                                                        </div>
                                                    ) : (
                                                        <div className="space-y-3">
                                                            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                                                                <div className="space-y-2">
                                                                    <Label>Alloy Grade</Label>
                                                                    <Select defaultValue={watchedUpdate?.alloyGrade || undefined} onValueChange={(v) => updateForm.setValue('alloyGrade', v)}>
                                                                        <SelectTrigger className={changedCls(baseline?.alloyGrade, watchedUpdate?.alloyGrade)}>
                                                                            <SelectValue placeholder="Select alloy" />
                                                                        </SelectTrigger>
                                                                        <SelectContent>
                                                                            {['AA6063', 'AA6061', 'AA6082', 'AA1050', 'Other'].map((g) => (
                                                                                <SelectItem key={g} value={g}>{g}</SelectItem>
                                                                            ))}
                                                                        </SelectContent>
                                                                    </Select>
                                                                </div>
                                                                <div className="flex items-end justify-end">
                                                                    <Button type="button" variant="outline" size="sm" onClick={() => updAppendElement({ element: '', percent: '' })}>Add Element</Button>
                                                                </div>
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
                                                                        {updElemFields.map((row, idx) => (
                                                                            <tr key={row.id} className="border-t">
                                                                                <td className="p-2">
                                                                                    <Input {...updateForm.register(`elements.${idx}.element` as const)} className={changedCls(baseline?.elements?.[idx]?.element, watchedUpdate?.elements?.[idx]?.element)} />
                                                                                </td>
                                                                                <td className="p-2">
                                                                                    <Input type="number" step="0.01" {...updateForm.register(`elements.${idx}.percent` as const, { valueAsNumber: true })} className={changedCls(baseline?.elements?.[idx]?.percent, watchedUpdate?.elements?.[idx]?.percent)} />
                                                                                </td>
                                                                                <td className="p-2">
                                                                                    <Input {...updateForm.register(`elements.${idx}.tolerance` as const)} className={changedCls(baseline?.elements?.[idx]?.tolerance, watchedUpdate?.elements?.[idx]?.tolerance)} />
                                                                                </td>
                                                                                <td className="p-2 text-right">
                                                                                    <Button type="button" variant="ghost" size="sm" onClick={() => updRemoveElement(idx)}>Remove</Button>
                                                                                </td>
                                                                            </tr>
                                                                        ))}
                                                                    </tbody>
                                                                </table>
                                                            </div>
                                                        </div>
                                                    )}
                                                </div>

                                                <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                                                    <Field label="Renewable Energy %">
                                                        <Input type="number" step="1" {...updateForm.register('renewablePct', { valueAsNumber: true })} className={changedCls(baseline?.renewablePct, watchedUpdate?.renewablePct)} />
                                                    </Field>
                                                    <Field label="Scope 1 Emissions (tCO₂e/t)">
                                                        <Input type="number" step="0.01" {...updateForm.register('scope1', { valueAsNumber: true })} className={changedCls(baseline?.scope1, watchedUpdate?.scope1)} />
                                                    </Field>
                                                    <Field label="Scope 2 Emissions (tCO₂e/t)">
                                                        <Input type="number" step="0.01" {...updateForm.register('scope2', { valueAsNumber: true })} className={changedCls(baseline?.scope2, watchedUpdate?.scope2)} />
                                                    </Field>
                                                    <div className="space-y-2">
                                                        <Label>Calculation Method</Label>
                                                        <Select defaultValue={watchedUpdate?.calcMethod || undefined} onValueChange={(v) => updateForm.setValue('calcMethod', v)}>
                                                            <SelectTrigger className={changedCls(baseline?.calcMethod, watchedUpdate?.calcMethod)}>
                                                                <SelectValue placeholder="Select method" />
                                                            </SelectTrigger>
                                                            <SelectContent>
                                                                <SelectItem value="EN ISO 14067:2018">EN ISO 14067:2018</SelectItem>
                                                                <SelectItem value="GHG Protocol">GHG Protocol</SelectItem>
                                                            </SelectContent>
                                                        </Select>
                                                    </div>
                                                </div>

                                                <div className="grid grid-cols-1 gap-6">
                                                    <DropRow
                                                        label="Replace Certificate of Analysis (CoA)"
                                                        accept=".pdf,.jpg,.jpeg,.png"
                                                        cid={updateForm.getValues('coaCid')}
                                                        onUpload={async (file) => updateForm.setValue('coaCid', await uploadToIpfsMock(file, 'QmCoANew'))}
                                                    />
                                                    <DropRow
                                                        label="Replace Carbon Footprint Report (CFP)"
                                                        accept=".pdf,.jpg,.jpeg,.png"
                                                        cid={updateForm.getValues('cfpCid')}
                                                        onUpload={async (file) => updateForm.setValue('cfpCid', await uploadToIpfsMock(file, 'QmCFPNew'))}
                                                    />
                                                    <DropRow
                                                        label="Attach ISO 14001 Certificate"
                                                        accept=".pdf,.jpg,.jpeg,.png"
                                                        cid={updateForm.getValues('isoCid')}
                                                        onUpload={async (file) => updateForm.setValue('isoCid', await uploadToIpfsMock(file, 'QmISONew'))}
                                                    />

                                                    <div className="space-y-3">
                                                        <div className="flex items-center justify-between">
                                                            <div>
                                                                <div className="font-medium">Optional Documents</div>
                                                                <div className="text-xs text-muted-foreground">Add ASI, EPD, LCA, DoC or other</div>
                                                            </div>
                                                            <Button type="button" variant="outline" size="sm" onClick={() => updAppendOptDoc({ type: 'ASI', name: '' })}>
                                                                Add Document
                                                            </Button>
                                                        </div>
                                                        {updOptDocFields.length > 0 && (
                                                            <div className="space-y-3">
                                                                {updOptDocFields.map((f, idx) => (
                                                                    <Card key={f.id}>
                                                                        <CardContent className="grid grid-cols-1 items-end gap-3 pt-6 sm:grid-cols-3">
                                                                            <div className="space-y-2">
                                                                                <Label>Type</Label>
                                                                                <Select
                                                                                    defaultValue={String(updateForm.getValues(`optionalDocs.${idx}.type` as const))}
                                                                                    onValueChange={(v) => updUpdateOptDoc(idx, { ...(updateForm.getValues(`optionalDocs.${idx}` as const) as OptionalDoc), type: v as OptionalDoc['type'] })}
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
                                                                                        const curr = updateForm.getValues(`optionalDocs.${idx}` as const) as OptionalDoc
                                                                                        updUpdateOptDoc(idx, { ...curr, name: file.name, cid })
                                                                                    }}
                                                                                />
                                                                                {updateForm.getValues(`optionalDocs.${idx}.cid` as const) ? (
                                                                                    <CidChip cid={updateForm.getValues(`optionalDocs.${idx}.cid` as const) ?? ''} />
                                                                                ) : null}
                                                                            </div>
                                                                            <div className="sm:text-right">
                                                                                <Button type="button" variant="ghost" onClick={() => updRemoveOptDoc(idx)}>
                                                                                    Remove
                                                                                </Button>
                                                                            </div>
                                                                        </CardContent>
                                                                    </Card>
                                                                ))}
                                                            </div>
                                                        )}
                                                    </div>
                                                </div>
                                            </div>
                                        </Section>
                                    )}

                                    {updateStep === 2 && (
                                        <Section title="Step 3 — Review" subtitle="Old vs New">
                                            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                                                <DiffBlock title="Old" values={baseline ?? undefined} />
                                                <DiffBlock title="New" values={updateForm.getValues()} />
                                            </div>
                                        </Section>
                                    )}

                                    {updateStep === 3 && updated && (
                                        <Section title="Updated" subtitle="Passport successfully updated">
                                            <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
                                                <Card className="md:col-span-2">
                                                    <CardHeader>
                                                        <CardTitle>Passport</CardTitle>
                                                        <CardDescription>Details</CardDescription>
                                                    </CardHeader>
                                                    <CardContent className="space-y-3 text-sm">
                                                        <PreviewRow label="Passport ID" value={updated.passportId} />
                                                        <PreviewRow label="Update JSON CID" value={updated.jsonCid} />
                                                        <div className="flex items-center gap-2 pt-2">
                                                            <Button type="button" variant="outline" size="sm" onClick={() => navigator.clipboard.writeText(updated.passportId)}>
                                                                <Copy className="mr-2 size-4" /> Copy Passport ID
                                                            </Button>
                                                            <Button type="button" variant="outline" size="sm" onClick={() => window.open(digitalLinkUpdated, '_blank')}>
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
                                                        {digitalLinkUpdated ? (
                                                            <img
                                                                alt="QR Code"
                                                                className="h-40 w-40"
                                                                src={`https://api.qrserver.com/v1/create-qr-code/?size=160x160&data=${encodeURIComponent(digitalLinkUpdated)}`}
                                                            />
                                                        ) : null}
                                                        <a className="text-xs text-blue-600 underline" href={digitalLinkUpdated} target="_blank" rel="noreferrer">
                                                            Digital Link URL
                                                        </a>
                                                    </CardContent>
                                                </Card>
                                            </div>
                                        </Section>
                                    )}

                                    {/* Sticky Footer */}
                                    <div className="sticky bottom-0 -mx-6 -mb-6 border-t bg-background/80 p-4 backdrop-blur supports-[backdrop-filter]:bg-background/60">
                                        <div className="flex items-center gap-3">
                                            {updateStep > 0 ? (
                                                <Button variant="outline" type="button" onClick={handleUpdateBack}>Back</Button>
                                            ) : null}
                                            {updateStep < UPDATE_STEPS.length - 1 ? (
                                                updateStep < 2 ? (
                                                    <Button type="button" onClick={handleUpdateNext}>Next</Button>
                                                ) : (
                                                    <Button type="button" onClick={handleDoUpdate}>Update</Button>
                                                )
                                            ) : null}
                                        </div>
                                    </div>
                                </>
                            )}
                        </CardContent>
                    </Card>
                </div>

                {/* Right Sidebar */}
                <div className="space-y-4 lg:sticky lg:top-6 self-start">
                    <Card>
                        <CardHeader>
                            <CardTitle>Preview</CardTitle>
                            <CardDescription>Live summary</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-3 text-sm">
                            {mode === 'create' ? (
                                <>
                                    <PreviewRow label="Upstream Batch" value={watchedCreate?.upstreamBatch || '—'} />
                                    <PreviewRow label="Lot / Heat Number" value={watchedCreate?.lotNumber} />
                                    <PreviewRow label="Form" value={String(watchedCreate?.productForm || '')} />
                                    <PreviewRow label="Production Date" value={watchedCreate?.productionDate} />
                                    <PreviewRow label="Site GLN" value={watchedCreate?.siteGln} />
                                    <PreviewRow label="HS Code" value={watchedCreate?.hsCode} />
                                    <PreviewRow label="Composition" value={watchedCreate?.compositionType} />
                                    {watchedCreate?.compositionType === 'Primary Purity' ? (
                                        <PreviewRow label="Purity %" value={String(watchedCreate?.purityPct ?? '')} />
                                    ) : (
                                        <>
                                            <PreviewRow label="Alloy Grade" value={String(watchedCreate?.alloyGrade ?? '')} />
                                            {(watchedCreate?.elements ?? [])
                                                .filter((e) => e.element)
                                                .map((e, i) => (
                                                    <PreviewRow key={`elc-${i}`} label={`Element ${e.element}`} value={`${e.percent ?? ''}${e.tolerance ? ` (${e.tolerance})` : ''}`} />
                                                ))}
                                        </>
                                    )}
                                    <PreviewRow label="Renewable %" value={String(watchedCreate?.renewablePct ?? '')} />
                                    <PreviewRow label="Scope 1" value={String(watchedCreate?.scope1 ?? '')} />
                                    <PreviewRow label="Scope 2" value={String(watchedCreate?.scope2 ?? '')} />
                                    <PreviewRow label="Method" value={watchedCreate?.calcMethod} />
                                    <PreviewRow label="CoA CID" value={watchedCreate?.coaCid} />
                                    <PreviewRow label="CFP CID" value={watchedCreate?.cfpCid} />
                                    <PreviewRow label="ISO CID" value={watchedCreate?.isoCid} />
                                </>
                            ) : (
                                <>
                                    <PreviewRow label="Passport ID" value={watchedUpdate?.passportId || '—'} />
                                    <PreviewRow label="Lot / Heat Number" value={watchedUpdate?.lotNumber} />
                                    <PreviewRow label="Form" value={String(watchedUpdate?.productForm || '')} />
                                    <PreviewRow label="Production Date" value={watchedUpdate?.productionDate} />
                                    <PreviewRow label="Site GLN" value={watchedUpdate?.siteGln} />
                                    <PreviewRow label="HS Code" value={watchedUpdate?.hsCode} />
                                    <PreviewRow label="Composition" value={watchedUpdate?.compositionType} />
                                    {watchedUpdate?.compositionType === 'Primary Purity' ? (
                                        <PreviewRow label="Purity %" value={String(watchedUpdate?.purityPct ?? '')} />
                                    ) : (
                                        <>
                                            <PreviewRow label="Alloy Grade" value={String(watchedUpdate?.alloyGrade ?? '')} />
                                            {(watchedUpdate?.elements ?? [])
                                                .filter((e) => e.element)
                                                .map((e, i) => (
                                                    <PreviewRow key={`elu-${i}`} label={`Element ${e.element}`} value={`${e.percent ?? ''}${e.tolerance ? ` (${e.tolerance})` : ''}`} />
                                                ))}
                                        </>
                                    )}
                                    <PreviewRow label="Renewable %" value={String(watchedUpdate?.renewablePct ?? '')} />
                                    <PreviewRow label="Scope 1" value={String(watchedUpdate?.scope1 ?? '')} />
                                    <PreviewRow label="Scope 2" value={String(watchedUpdate?.scope2 ?? '')} />
                                    <PreviewRow label="Method" value={watchedUpdate?.calcMethod} />
                                    <PreviewRow label="CoA CID" value={watchedUpdate?.coaCid} />
                                    <PreviewRow label="CFP CID" value={watchedUpdate?.cfpCid} />
                                    <PreviewRow label="ISO CID" value={watchedUpdate?.isoCid} />
                                </>
                            )}
                            {minted ? (
                                <div className="mt-3 grid place-items-center gap-2">
                                    <img
                                        alt="QR Code"
                                        className="h-32 w-32"
                                        src={`https://api.qrserver.com/v1/create-qr-code/?size=128x128&data=${encodeURIComponent(`https://ipfs.io/ipfs/${minted.jsonCid.replace('ipfs://', '')}`)}`}
                                    />
                                    <div className="text-xs text-muted-foreground">Minted: {minted.passportId}</div>
                                </div>
                            ) : updated ? (
                                <div className="mt-3 grid place-items-center gap-2">
                                    <img
                                        alt="QR Code"
                                        className="h-32 w-32"
                                        src={`https://api.qrserver.com/v1/create-qr-code/?size=128x128&data=${encodeURIComponent(`https://ipfs.io/ipfs/${updated.jsonCid.replace('ipfs://', '')}`)}`}
                                    />
                                    <div className="text-xs text-muted-foreground">Updated: {updated.passportId}</div>
                                </div>
                            ) : (
                                <div className="mt-2 grid place-items-center">
                                    <QrCode className="size-6 text-muted-foreground" />
                                    <div className="text-xs text-muted-foreground">QR will appear after {mode === 'create' ? 'mint' : 'update'}</div>
                                </div>
                            )}
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader>
                            <CardTitle>Recent Activity</CardTitle>
                            <CardDescription>Last 5 events</CardDescription>
                        </CardHeader>
                        <CardContent className="space-y-3">
                            {activity.map((a, i) => (
                                <div key={`act-${i}`} className="flex items-center justify-between gap-3 text-sm">
                                    <div className="flex items-center gap-2">
                                        <Clock className="size-4 text-muted-foreground" />
                                        <div>
                                            <div className="font-medium">{a.kind === 'created' ? 'Created' : 'Updated'} {a.id}</div>
                                            <div className="text-xs text-muted-foreground">{a.when.toLocaleString()}</div>
                                        </div>
                                    </div>
                                    <a className="text-xs text-blue-600 underline" href={`https://etherscan.io/tx/${a.tx}`} target="_blank" rel="noreferrer">tx</a>
                                </div>
                            ))}
                        </CardContent>
                    </Card>

                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center gap-2"><BarChart3 className="size-5" /> Quick Stats</CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="grid grid-cols-3 gap-3 text-center">
                                <div className="rounded-md border p-3">
                                    <div className="text-xs text-muted-foreground">Total minted</div>
                                    <div className="text-sm font-semibold">{stats.total}</div>
                                </div>
                                <div className="rounded-md border p-3">
                                    <div className="text-xs text-muted-foreground">Pending updates</div>
                                    <div className="text-sm font-semibold">{stats.pending}</div>
                                </div>
                                <div className="rounded-md border p-3">
                                    <div className="text-xs text-muted-foreground">Avg renewable %</div>
                                    <div className="text-sm font-semibold">{Number(stats.avgRenewable).toFixed(1)}%</div>
                                </div>
                            </div>
                        </CardContent>
                    </Card>
                </div>
            </div>
        </div>
    )
}

// UI helpers
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
            <div className={`relative grid items-center gap-6 ${steps.length === 6 ? 'grid-cols-6' : 'grid-cols-4'}`}>
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
            <div className={`grid gap-6 ${steps.length === 6 ? 'grid-cols-6' : 'grid-cols-4'}`}>
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

function DiffBlock({ title, values }: { title: string; values?: Partial<UpdateFormValues> | null }) {
    const rows: Array<[string, string | undefined]> = [
        ['Passport ID', values?.passportId],
        ['Lot / Heat', values?.lotNumber],
        ['Form', values?.productForm ? String(values.productForm) : undefined],
        ['Production Date', values?.productionDate],
        ['Site GLN', values?.siteGln],
        ['HS Code', values?.hsCode],
        ['Composition', values?.compositionType],
        ['Purity %', values?.compositionType === 'Primary Purity' ? String(values?.purityPct ?? '') : undefined],
        ['Alloy Grade', values?.compositionType === 'Alloy' ? values?.alloyGrade : undefined],
        ['Renewable %', values?.renewablePct != null ? String(values.renewablePct) : undefined],
        ['Scope 1', values?.scope1 != null ? String(values.scope1) : undefined],
        ['Scope 2', values?.scope2 != null ? String(values.scope2) : undefined],
        ['Method', values?.calcMethod],
        ['CoA CID', values?.coaCid],
        ['CFP CID', values?.cfpCid],
        ['ISO CID', values?.isoCid],
    ]
    return (
        <Card>
            <CardHeader>
                <CardTitle className="text-base">{title}</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2 text-sm">
                {rows.map(([k, v]) => (
                    v ? (
                        <div key={k} className="flex items-center justify-between gap-4">
                            <span className="text-muted-foreground">{k}</span>
                            <span className="font-medium">{v}</span>
                        </div>
                    ) : null
                ))}
            </CardContent>
        </Card>
    )
}

// Utils
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
    await new Promise((r) => setTimeout(r, 300))
    return `ipfs://${fixed ?? 'bafy' + Math.random().toString(36).slice(2, 10)}`
}

async function uploadJsonToIpfsMock(_json: string): Promise<string> {
    await new Promise((r) => setTimeout(r, 300))
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

function computeDiff<T extends Record<string, unknown> | null | undefined>(
    prev: T,
    next: Record<string, unknown>,
): Record<string, { old: unknown; new: unknown }> {
    const result: Record<string, { old: unknown; new: unknown }> = {}
    const keys = new Set<string>([...Object.keys(prev ?? {}), ...Object.keys(next ?? {})])
    for (const k of keys) {
        const a = (prev as Record<string, unknown> | null | undefined)?.[k]
        const b = (next as Record<string, unknown>)[k]
        if (!deepEqual(a, b)) {
            result[k] = { old: a, new: b }
        }
    }
    return result
}

function deepEqual(a: unknown, b: unknown): boolean {
    try {
        return JSON.stringify(a) === JSON.stringify(b)
    } catch {
        return a === b
    }
}

type ActivityEvent = { kind: 'created' | 'updated'; id: string; tx: string; when: Date }

function randomTx(): string {
    return '0x' + Math.random().toString(16).slice(2, 10)
}

function daysAgo(n: number): Date {
    const d = new Date()
    d.setDate(d.getDate() - n)
    return d
}

function average(arr: number[]): number {
    if (!arr.length) return 0
    return arr.reduce((a, b) => a + b, 0) / arr.length
}

function generatePassportId(): string {
    const num = Math.floor(Math.random() * 99999).toString().padStart(5, '0')
    return `ALP-2025-${num}`
}

function changedCls(a: unknown, b: unknown): string {
    if (a === undefined || a === null) return ''
    const same = JSON.stringify(a) === JSON.stringify(b)
    return same ? '' : 'ring-2 ring-amber-400'
}


