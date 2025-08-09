import React, { useMemo, useState } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { useForm } from 'react-hook-form'
import * as yup from 'yup'
import { yupResolver } from '@hookform/resolvers/yup'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Check } from 'lucide-react'
import { motion, AnimatePresence } from 'framer-motion'
import '../routeTree.gen'
import { useAuth } from '@/lib/auth'

// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore - File-based route types are injected by the generated route tree
export const Route = createFileRoute('/app/roles')({
    component: RoleManagementPage,
})

type RoleFormValues = {
    role: string
    ethAddress: string
    name: string
    creditDurationMonths: number
}

const ROLE_OPTIONS = [
    'Admin',
    'Miner',
    'Refiner',
    'Alloy Producer',
    'Product Manufacturer',
    'Distributor',
    'Service Provider',
    'Recycler',
    'Auditor',
    'Regulator',
] as const

const roleFormSchema = yup
    .object({
        role: yup
            .string()
            .oneOf([...ROLE_OPTIONS], 'Select a valid role')
            .required('Role is required'),
        ethAddress: yup
            .string()
            .required('Ethereum address is required')
            .matches(/^0x[a-fA-F0-9]{40}$/u, 'Enter a valid Ethereum address'),
        name: yup.string().required('Name is required').min(2, 'Name must be at least 2 characters'),
        creditDurationMonths: yup
            .number()
            .typeError('Credit duration must be a number')
            .integer('Must be an integer')
            .min(1, 'Must be at least 1 month')
            .max(120, 'Must be 120 months or less')
            .required('Credit duration is required'),
    })
    .required() as yup.ObjectSchema<RoleFormValues>

const STEPS = [
    'Select Role',
    'Assign Role',
    'Register DID',
    'Verify DID',
    'Issue Credential',
    'Generate Signature',
    'Sign Credential',
    'Validate Credentials',
] as const

function RoleManagementPage() {
    const { accountAddress } = useAuth()
    const [completedSteps, setCompletedSteps] = useState<Set<number>>(new Set())

    const defaultValues: Partial<RoleFormValues> = useMemo(
        () => ({
            role: '',
            ethAddress: accountAddress ?? '',
            name: '',
        }),
        [accountAddress],
    )

    const {
        handleSubmit,
        register,
        formState: { errors, isSubmitting },
        reset,
    } = useForm<RoleFormValues>({
        resolver: yupResolver(roleFormSchema),
        defaultValues,
        mode: 'onBlur',
        reValidateMode: 'onChange',
    })

    function handleRoleSubmit(values: RoleFormValues) {
        // In a real app, perform API calls here. We simulate success.
        // Mark the first step as completed and reset the form to pristine values (keeping role selection visible).
        setCompletedSteps((prev) => new Set(prev).add(0))
        reset(values)
    }

    return (
        <div className="space-y-6">
            <header className="space-y-1">
                <h1 className="text-2xl font-semibold tracking-tight">Role Management</h1>
                <p className="text-sm text-muted-foreground">Manage roles and verifiable credential lifecycle.</p>
            </header>

            <Stepper steps={STEPS as unknown as string[]} completedSteps={completedSteps} />

            {/* Generic form used across steps. For now, only the first step is enabled */}
            <Card className="border-0 shadow-none">
                <CardHeader>
                    <CardTitle>Select Role</CardTitle>
                    <CardDescription>Choose a role and provide participant details.</CardDescription>
                </CardHeader>
                <CardContent>
                    <form className="grid grid-cols-1 gap-4 sm:grid-cols-2" onSubmit={handleSubmit(handleRoleSubmit)} noValidate>
                        <div className="space-y-2">
                            <Label htmlFor="role">Role</Label>
                            <Select
                                onValueChange={(val) => {
                                    const { onChange } = register('role')
                                    onChange({ target: { value: val, name: 'role' } } as unknown as React.ChangeEvent<HTMLInputElement>)
                                }}
                                disabled={false}
                            >
                                <SelectTrigger aria-invalid={Boolean(errors.role)}>
                                    <SelectValue placeholder="Select a role" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectGroup>
                                        {ROLE_OPTIONS.map((r) => (
                                            <SelectItem key={r} value={r}>
                                                {r}
                                            </SelectItem>
                                        ))}
                                    </SelectGroup>
                                </SelectContent>
                            </Select>
                            {errors.role ? <FieldError message={errors.role.message} /> : null}
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="ethAddress">Ethereum Address</Label>
                            <Input
                                id="ethAddress"
                                placeholder="0x…"
                                aria-invalid={Boolean(errors.ethAddress)}
                                {...register('ethAddress')}
                                disabled={false}
                            />
                            {errors.ethAddress ? (
                                <FieldError message={errors.ethAddress.message} />
                            ) : null}
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="name">Name</Label>
                            <Input id="name" placeholder="Enter name" aria-invalid={Boolean(errors.name)} {...register('name')} disabled={false} />
                            {errors.name ? <FieldError message={errors.name.message} /> : null}
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="creditDurationMonths">Credit Duration (months)</Label>
                            <Input
                                id="creditDurationMonths"
                                type="number"
                                min={1}
                                step={1}
                                placeholder="e.g., 12"
                                aria-invalid={Boolean(errors.creditDurationMonths)}
                                {...register('creditDurationMonths', { valueAsNumber: true })}
                                disabled={false}
                            />
                            {errors.creditDurationMonths ? (
                                <FieldError message={errors.creditDurationMonths.message} />
                            ) : null}
                        </div>

                        <div className="sm:col-span-2">
                            <Button type="submit" disabled={isSubmitting} className="w-full sm:w-auto">
                                Add
                            </Button>
                        </div>
                    </form>
                </CardContent>
            </Card>

        </div>
    )
}

function FieldError({ message }: { message?: string }) {
    if (!message) return null
    return <p className="text-sm text-destructive">{message}</p>
}

function Stepper({
    steps,
    completedSteps,
}: {
    steps: string[]
    completedSteps: Set<number>
}) {
    const colsPerRow = steps.length
    const rows: string[][] = [steps]

    // Step descriptions removed as requested

    return (
        <div className="space-y-8">
            {rows.map((row, rowIndex) => (
                <div key={`row-${rowIndex}`} className="space-y-3">
                    <div className="relative grid grid-cols-8 items-center gap-6">
                        {/* Single base line across the whole row */}
                        <div className="pointer-events-none absolute left-0 right-0 top-1/2 -translate-y-1/2">
                            <div className="h-0.5 w-full rounded-full bg-gray-200" />
                            {/* Progress overlay that grows with completed steps */}
                            {
                                (() => {
                                    const lastCompletedIndex = row.reduce((acc, _label, i) => {
                                        const gi = rowIndex * colsPerRow + i
                                        return completedSteps.has(gi) ? i : acc
                                    }, -1)
                                    const denominator = Math.max(1, row.length - 1)
                                    const progressPct = Math.max(0, lastCompletedIndex) / denominator * 100
                                    return (
                                        <motion.div
                                            initial={false}
                                            animate={{ width: `${progressPct}%` }}
                                            transition={{ duration: 0.35 }}
                                            className="h-0.5 rounded-full bg-[color:var(--color-primary)]"
                                            style={{ width: '0%' }}
                                        />
                                    )
                                })()
                            }
                        </div>
                        {row.map((label, i) => {
                            const globalIndex = rowIndex * colsPerRow + i
                            return (
                                <div key={label} className="relative grid place-items-center">
                                    <motion.span
                                        aria-label={completedSteps.has(globalIndex) ? 'completed' : 'not completed'}
                                        className="grid size-8 place-items-center rounded-full border text-xs bg-background"
                                        initial={false}
                                        animate={{
                                            backgroundColor: completedSteps.has(globalIndex) ? 'var(--color-primary)' : 'var(--color-background)',
                                            color: completedSteps.has(globalIndex) ? 'var(--color-primary-foreground)' : 'rgb(17 24 39)',
                                            borderColor: completedSteps.has(globalIndex) ? 'var(--color-primary)' : 'rgb(229 231 235)',
                                            boxShadow: completedSteps.has(globalIndex)
                                                ? '0 0 0 4px color-mix(in oklab, var(--color-primary) 20%, transparent)'
                                                : '0 0 0 0px rgba(0,0,0,0)',
                                        }}
                                        transition={{ duration: 0.25 }}
                                    >
                                        <AnimatePresence initial={false} mode="wait">
                                            {completedSteps.has(globalIndex) ? (
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
                                                    {globalIndex + 1}
                                                </motion.span>
                                            )}
                                        </AnimatePresence>
                                    </motion.span>
                                </div>
                            )
                        })}
                    </div>

                    <div className="grid grid-cols-8 gap-6">
                        {row.map((label) => (
                            <div key={`${label}-meta`} className="text-center">
                                <div className="text-sm font-semibold">{label}</div>
                            </div>
                        ))}
                    </div>
                </div>
            ))}
        </div>
    )
}


