'use client';

import React from 'react';
import * as Select from '@radix-ui/react-select';
import { Check, ChevronDown } from 'lucide-react';

interface SelectOption {
    value: string;
    label: string;
}

interface CustomSelectProps {
    value: string;
    onValueChange: (value: string) => void;
    options: SelectOption[];
    placeholder?: string;
    label?: string;
}

export const CustomSelect: React.FC<CustomSelectProps> = ({
    value,
    onValueChange,
    options,
    placeholder = 'Select...',
    label,
}) => {
    return (
        <div className="w-full">
            {label && (
                <label className="text-[10px] font-black text-[rgb(var(--color-text-tertiary))] uppercase tracking-[0.2em] mb-3 block">
                    {label}
                </label>
            )}
            <Select.Root value={value} onValueChange={onValueChange}>
                <Select.Trigger className="w-full flex items-center justify-between px-4 py-3 bg-[rgb(var(--color-bg-tertiary))] border border-[rgb(var(--color-border-primary))] rounded-xl text-sm font-semibold hover:bg-[rgb(var(--color-bg-primary))] focus:outline-none focus:ring-2 focus:ring-[rgb(var(--color-primary))] transition-all">
                    <Select.Value placeholder={placeholder} />
                    <Select.Icon>
                        <ChevronDown size={16} className="text-[rgb(var(--color-text-tertiary))]" />
                    </Select.Icon>
                </Select.Trigger>

                <Select.Portal>
                    <Select.Content className="overflow-hidden bg-[rgb(var(--color-bg-primary))] border border-[rgb(var(--color-border-primary))] rounded-xl shadow-2xl z-50">
                        <Select.Viewport className="p-1">
                            {options.map((option) => (
                                <Select.Item
                                    key={option.value}
                                    value={option.value}
                                    className="relative flex items-center px-8 py-2.5 text-sm font-semibold rounded-lg cursor-pointer select-none outline-none hover:bg-[rgb(var(--color-bg-tertiary))] focus:bg-[rgb(var(--color-bg-tertiary))] data-[state=checked]:bg-[rgb(var(--color-primary))] data-[state=checked]:text-white transition-colors"
                                >
                                    <Select.ItemText>{option.label}</Select.ItemText>
                                    <Select.ItemIndicator className="absolute left-2 inline-flex items-center">
                                        <Check size={14} />
                                    </Select.ItemIndicator>
                                </Select.Item>
                            ))}
                        </Select.Viewport>
                    </Select.Content>
                </Select.Portal>
            </Select.Root>
        </div>
    );
};
