'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { LayoutDashboard, ListTodo, BarChart3, Menu, X } from 'lucide-react';

const navigation = [
    { name: 'Dashboard', href: '/', icon: LayoutDashboard },
    { name: 'Tasks', href: '/tasks', icon: ListTodo },
    { name: 'Month View', href: '/month', icon: BarChart3 },
];

export const Sidebar: React.FC = () => {
    const pathname = usePathname();
    const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);

    return (
        <>
            {/* Mobile Menu Button */}
            <button
                onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
                className="md:hidden fixed top-4 left-4 z-50 p-2.5 rounded-xl bg-[rgb(var(--color-bg-primary))] border border-[rgb(var(--color-border-primary))] shadow-lg"
            >
                {isMobileMenuOpen ? <X size={20} /> : <Menu size={20} />}
            </button>

            {/* Mobile Overlay */}
            {isMobileMenuOpen && (
                <div
                    className="md:hidden fixed inset-0 bg-black/60 backdrop-blur-sm z-40"
                    onClick={() => setIsMobileMenuOpen(false)}
                />
            )}

            {/* Sidebar */}
            <aside
                className={`
          w-64 bg-[rgb(var(--color-bg-primary))] border-r border-[rgb(var(--color-border-primary))] p-6 flex flex-col shrink-0
          fixed md:static inset-y-0 left-0 z-40 transform transition-transform duration-300
          ${isMobileMenuOpen ? 'translate-x-0' : '-translate-x-full md:translate-x-0'}
        `}
            >
                {/* Logo */}
                <div className="flex items-center gap-3 px-2 mb-10 mt-12 md:mt-0">
                    <div className="w-9 h-9 bg-[rgb(var(--color-primary))] rounded-xl flex items-center justify-center shadow-lg">
                        <BarChart3 size={20} className="text-white" />
                    </div>
                    <span className="text-xl font-bold tracking-tight">Orbito</span>
                </div>

                {/* Navigation */}
                <nav className="flex flex-col gap-1.5 flex-1">
                    {navigation.map((item) => {
                        const isActive = pathname === item.href;
                        const Icon = item.icon;

                        return (
                            <Link
                                key={item.name}
                                href={item.href}
                                onClick={() => setIsMobileMenuOpen(false)}
                                className={`flex items-center gap-3 px-4 py-3 rounded-xl transition-all ${isActive
                                    ? 'bg-[rgb(var(--color-primary))] text-white shadow-md'
                                    : 'text-[rgb(var(--color-text-secondary))] hover:bg-[rgb(var(--color-bg-tertiary))]'
                                    }`}
                            >
                                <Icon size={18} />
                                <span className="text-sm font-semibold">{item.name}</span>
                            </Link>
                        );
                    })}
                </nav>

                {/* Bottom Section */}
                <div className="mt-auto pt-6 border-t border-[rgb(var(--color-border-secondary))]">
                    <div className="p-4 bg-[rgb(var(--color-bg-tertiary))] rounded-2xl">
                        <p className="text-[10px] font-bold text-[rgb(var(--color-text-tertiary))] uppercase tracking-widest mb-2">
                            Month View
                        </p>
                        <p className="text-sm font-bold">January 2026</p>
                    </div>
                </div>
            </aside>
        </>
    );
};
