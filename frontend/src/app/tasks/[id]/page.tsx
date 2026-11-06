'use client';

import React, { useState, useEffect } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useTask, useUpdateTask, useDeleteTask } from '@/lib/hooks/useTasks';
import { Header } from '@/components/layout/Header';
import { PriorityBadge } from '@/components/ui/PriorityBadge';
import { CustomSelect } from '@/components/ui/CustomSelect';
import { WEEKS, STATUSES } from '@/lib/utils/constants';
import { ArrowLeft, Send, Trash2, Calendar, Tag as TagIcon, MessageSquare, Plus, X, Edit2 } from 'lucide-react';
import { Task, TaskStatus, TaskPriority, TaskType } from '@/types';

const priorityOptions: { value: TaskPriority; label: string }[] = [
    { value: 'low', label: 'Low Priority' },
    { value: 'medium', label: 'Medium Priority' },
    { value: 'high', label: 'High Priority' },
];

const typeOptions: { value: TaskType; label: string }[] = [
    { value: 'feature', label: 'Feature' },
    { value: 'bug', label: 'Bug' },
    { value: 'improvement', label: 'Improvement' },
    { value: 'research', label: 'Research' },
    { value: 'meeting', label: 'Meeting' },
    { value: 'other', label: 'Other' },
];

export default function TaskDetailPage() {
    const params = useParams();
    const router = useRouter();
    const taskId = params.id as string;

    const { data: task, isLoading } = useTask(taskId);
    const updateTask = useUpdateTask();
    const deleteTask = useDeleteTask();

    const [commentText, setCommentText] = useState('');

    // Edit states
    const [isEditingTitle, setIsEditingTitle] = useState(false);
    const [isEditingDescription, setIsEditingDescription] = useState(false);
    const [editedTitle, setEditedTitle] = useState('');
    const [editedDescription, setEditedDescription] = useState('');
    const [tagInput, setTagInput] = useState('');
    const [isSaving, setIsSaving] = useState(false);

    // Update local state when task changes
    useEffect(() => {
        if (task) {
            setEditedTitle(task.text);
            setEditedDescription(task.description || '');
        }
    }, [task]);

    const handleDelete = () => {
        if (task && confirm('Are you sure you want to delete this task?')) {
            deleteTask.mutate(task.id, {
                onSuccess: () => router.push('/tasks'),
            });
        }
    };

    const handleSaveTitle = () => {
        if (!task || editedTitle.trim() === task.text) {
            setIsEditingTitle(false);
            return;
        }

        setIsSaving(true);
        updateTask.mutate({
            id: task.id,
            updates: { text: editedTitle.trim() },
        }, {
            onSettled: () => {
                setIsSaving(false);
                setIsEditingTitle(false);
            },
        });
    };

    const handleSaveDescription = () => {
        if (!task || editedDescription === (task.description || '')) {
            setIsEditingDescription(false);
            return;
        }

        setIsSaving(true);
        updateTask.mutate({
            id: task.id,
            updates: { description: editedDescription },
        }, {
            onSettled: () => {
                setIsSaving(false);
                setIsEditingDescription(false);
            },
        });
    };

    const handleAddTag = () => {
        if (!task || !tagInput.trim() || task.tags?.includes(tagInput.trim())) {
            setTagInput('');
            return;
        }

        setIsSaving(true);
        updateTask.mutate({
            id: task.id,
            updates: { tags: [...(task.tags || []), tagInput.trim()] },
        }, {
            onSettled: () => {
                setIsSaving(false);
                setTagInput('');
            },
        });
    };

    const handleRemoveTag = (tag: string) => {
        if (!task) return;

        setIsSaving(true);
        updateTask.mutate({
            id: task.id,
            updates: { tags: task.tags?.filter((t) => t !== tag) || [] },
        }, {
            onSettled: () => setIsSaving(false),
        });
    };

    const handleUpdateField = (field: keyof Task, value: any) => {
        if (!task) return;

        setIsSaving(true);
        updateTask.mutate({
            id: task.id,
            updates: { [field]: value },
        }, {
            onSettled: () => setIsSaving(false),
        });
    };

    const handleAddComment = (e: React.FormEvent) => {
        e.preventDefault();
        if (!commentText.trim() || !task) return;

        const newComment = {
            id: Date.now().toString(),
            taskId: task.id,
            user: 'You',
            text: commentText,
            time: 'Just now',
            createdAt: new Date().toISOString(),
        };

        updateTask.mutate({
            id: task.id,
            updates: {
                comments: [...(task.comments || []), newComment],
            },
        });

        setCommentText('');
    };

    if (isLoading) {
        return (
            <div className="flex items-center justify-center h-full">
                <div className="text-center">
                    <div className="w-16 h-16 border-4 border-[rgb(var(--color-primary))] border-t-transparent rounded-full animate-spin mx-auto mb-4" />
                    <p className="text-[rgb(var(--color-text-secondary))] font-semibold">Loading task...</p>
                </div>
            </div>
        );
    }

    if (!task) {
        return (
            <div className="flex flex-col items-center justify-center h-full p-8">
                <p className="text-xl font-bold text-[rgb(var(--color-text-secondary))] mb-4">Task not found</p>
                <button
                    onClick={() => router.push('/tasks')}
                    className="flex items-center gap-2 px-4 py-2 bg-[rgb(var(--color-primary))] text-white rounded-xl hover:bg-[rgb(var(--color-primary-hover))] transition-all"
                >
                    <ArrowLeft size={18} />
                    Back to Tasks
                </button>
            </div>
        );
    }

    const currentWeek = WEEKS.find((w) => w.id === task.week);
    const currentStatus = STATUSES.find((s) => s.id === task.status);

    return (
        <div className="flex flex-col h-full overflow-hidden bg-[rgb(var(--color-bg-secondary))]">
            <Header
                title="Task Details"
                subtitle={currentWeek?.range}
                showAddButton={false}
            />

            <div className="flex-1 overflow-y-auto scrollbar-thin">
                <div className="max-w-7xl mx-auto px-3 py-4 md:px-6 md:py-6">
                    {/* Back Button */}
                    <button
                        onClick={() => router.push('/tasks')}
                        className="flex items-center gap-2 text-[rgb(var(--color-text-secondary))] hover:text-[rgb(var(--color-primary))] transition-all text-sm font-semibold mb-4"
                    >
                        <ArrowLeft size={18} />
                        Back to Tasks
                    </button>

                    <div className="grid grid-cols-1 lg:grid-cols-[1fr_320px] gap-4">
                        {/* Main Content - Left Side */}
                        <div className="space-y-4">
                            {/* Task Header Card */}
                            <div className="bg-[rgb(var(--color-bg-primary))] rounded-lg border border-[rgb(var(--color-border-primary))] p-5">
                                <div className="flex items-start justify-between mb-4">
                                    {/* Editable Title */}
                                    {isEditingTitle ? (
                                        <input
                                            type="text"
                                            value={editedTitle}
                                            onChange={(e) => setEditedTitle(e.target.value)}
                                            onBlur={handleSaveTitle}
                                            onKeyDown={(e) => {
                                                if (e.key === 'Enter') handleSaveTitle();
                                                if (e.key === 'Escape') {
                                                    setEditedTitle(task.text);
                                                    setIsEditingTitle(false);
                                                }
                                            }}
                                            autoFocus
                                            className="text-2xl md:text-3xl font-bold flex-1 leading-tight bg-transparent border-b-2 border-[rgb(var(--color-primary))] focus:outline-none"
                                        />
                                    ) : (
                                        <h1
                                            onClick={() => setIsEditingTitle(true)}
                                            className="text-2xl md:text-3xl font-bold flex-1 leading-tight cursor-pointer hover:text-[rgb(var(--color-primary))] transition-colors group"
                                        >
                                            {task.text}
                                            <Edit2 size={18} className="inline-block ml-2 opacity-0 group-hover:opacity-50 transition-opacity" />
                                        </h1>
                                    )}
                                    <button
                                        onClick={handleDelete}
                                        className="p-2 hover:bg-red-50 dark:hover:bg-red-950 rounded-lg transition-all ml-3"
                                    >
                                        <Trash2 size={18} className="text-red-500" />
                                    </button>
                                </div>

                                {/* Priority Badge - Now Editable */}
                                <div className="mb-4">
                                    <CustomSelect
                                        value={task.priority}
                                        onValueChange={(value) => handleUpdateField('priority', value as TaskPriority)}
                                        options={priorityOptions}
                                        placeholder="Select priority"
                                    />
                                </div>

                                {/* Editable Description */}
                                <div className="mb-4">
                                    <h3 className="text-xs font-bold text-[rgb(var(--color-text-tertiary))] uppercase tracking-wide mb-2">
                                        Description
                                    </h3>
                                    {isEditingDescription ? (
                                        <textarea
                                            value={editedDescription}
                                            onChange={(e) => setEditedDescription(e.target.value)}
                                            onBlur={handleSaveDescription}
                                            onKeyDown={(e) => {
                                                if (e.key === 'Escape') {
                                                    setEditedDescription(task.description || '');
                                                    setIsEditingDescription(false);
                                                }
                                            }}
                                            autoFocus
                                            rows={3}
                                            className="w-full text-[rgb(var(--color-text-secondary))] leading-relaxed text-sm bg-[rgb(var(--color-bg-tertiary))] border border-[rgb(var(--color-primary))] rounded-lg p-3 focus:outline-none focus:ring-2 focus:ring-[rgb(var(--color-primary))]"
                                        />
                                    ) : (
                                        <p
                                            onClick={() => setIsEditingDescription(true)}
                                            className="text-[rgb(var(--color-text-secondary))] leading-relaxed text-sm cursor-pointer hover:bg-[rgb(var(--color-bg-tertiary))] p-2 rounded-lg transition-colors min-h-[60px] group"
                                        >
                                            {task.description || 'Click to add description...'}
                                            <Edit2 size={14} className="inline-block ml-2 opacity-0 group-hover:opacity-50 transition-opacity" />
                                        </p>
                                    )}
                                </div>

                                {/* Editable Tags */}
                                <div>
                                    <h3 className="text-xs font-bold text-[rgb(var(--color-text-tertiary))] uppercase tracking-wide mb-2 flex items-center gap-2">
                                        <TagIcon size={12} />
                                        Tags
                                    </h3>
                                    <div className="flex flex-wrap gap-2 mb-2">
                                        {task.tags && task.tags.map((tag, index) => (
                                            <span
                                                key={index}
                                                className="inline-flex items-center gap-1.5 px-3 py-1 bg-[rgb(var(--color-primary-light))] text-[rgb(var(--color-primary))] rounded-md text-xs font-semibold group"
                                            >
                                                {tag}
                                                <button
                                                    onClick={() => handleRemoveTag(tag)}
                                                    className="hover:text-red-500 transition-colors"
                                                >
                                                    <X size={12} />
                                                </button>
                                            </span>
                                        ))}
                                    </div>
                                    <div className="flex gap-2">
                                        <input
                                            type="text"
                                            placeholder="Add a tag..."
                                            value={tagInput}
                                            onChange={(e) => setTagInput(e.target.value)}
                                            onKeyDown={(e) => {
                                                if (e.key === 'Enter') {
                                                    e.preventDefault();
                                                    handleAddTag();
                                                }
                                            }}
                                            className="flex-1 px-3 py-1.5 rounded-lg bg-[rgb(var(--color-bg-tertiary))] border border-[rgb(var(--color-border-primary))] focus:outline-none focus:ring-2 focus:ring-[rgb(var(--color-primary))] text-sm"
                                        />
                                        <button
                                            onClick={handleAddTag}
                                            className="px-3 py-1.5 bg-[rgb(var(--color-primary))] text-white rounded-lg hover:bg-[rgb(var(--color-primary-hover))] transition-all"
                                        >
                                            <Plus size={16} />
                                        </button>
                                    </div>
                                </div>
                            </div>

                            {/* Comments Section */}
                            <div className="bg-[rgb(var(--color-bg-primary))] rounded-lg border border-[rgb(var(--color-border-primary))] p-5">
                                <h2 className="text-base font-bold text-[rgb(var(--color-text-primary))] mb-4 flex items-center gap-2">
                                    <MessageSquare size={16} />
                                    Comments ({task.comments?.length || 0})
                                </h2>

                                {/* Comments List */}
                                <div className="space-y-4 mb-4">
                                    {task.comments && task.comments.length > 0 ? (
                                        task.comments.map((comment) => (
                                            <div key={comment.id} className="flex gap-3">
                                                <div className="w-8 h-8 rounded-full bg-gradient-to-br from-[rgb(var(--color-primary))] to-purple-600 flex items-center justify-center text-white shrink-0 font-semibold text-xs">
                                                    {comment.user[0]}
                                                </div>
                                                <div className="flex-1">
                                                    <div className="flex items-center gap-2 mb-1">
                                                        <span className="font-semibold text-sm">{comment.user}</span>
                                                        <span className="text-xs text-[rgb(var(--color-text-tertiary))]">{comment.time}</span>
                                                    </div>
                                                    <div className="bg-[rgb(var(--color-bg-tertiary))] p-3 rounded-lg rounded-tl-none border border-[rgb(var(--color-border-secondary))]">
                                                        <p className="text-sm text-[rgb(var(--color-text-secondary))] leading-relaxed">
                                                            {comment.text}
                                                        </p>
                                                    </div>
                                                </div>
                                            </div>
                                        ))
                                    ) : (
                                        <div className="text-center py-8 bg-[rgb(var(--color-bg-tertiary))] rounded-lg border border-dashed border-[rgb(var(--color-border-primary))]">
                                            <MessageSquare size={32} className="mx-auto mb-2 text-[rgb(var(--color-text-tertiary))]" />
                                            <p className="text-[rgb(var(--color-text-tertiary))] text-sm">
                                                No comments yet. Be the first to comment!
                                            </p>
                                        </div>
                                    )}
                                </div>

                                {/* Add Comment Form */}
                                <form onSubmit={handleAddComment} className="relative">
                                    <textarea
                                        value={commentText}
                                        onChange={(e) => setCommentText(e.target.value)}
                                        placeholder="Add a comment..."
                                        className="w-full bg-[rgb(var(--color-bg-tertiary))] border border-[rgb(var(--color-border-primary))] rounded-lg p-3 pr-12 text-sm focus:outline-none focus:ring-2 focus:ring-[rgb(var(--color-primary))] focus:border-transparent transition-all resize-none"
                                        rows={3}
                                    />
                                    <button
                                        type="submit"
                                        disabled={!commentText.trim()}
                                        className="absolute bottom-3 right-3 p-2 bg-[rgb(var(--color-primary))] text-white rounded-lg hover:bg-[rgb(var(--color-primary-hover))] transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                                    >
                                        <Send size={16} />
                                    </button>
                                </form>
                            </div>
                        </div>

                        {/* Sidebar - Right Side */}
                        <div className="space-y-4">
                            {/* Editable Status Card */}
                            <div className="bg-[rgb(var(--color-bg-primary))] rounded-lg border border-[rgb(var(--color-border-primary))] p-4">
                                <h3 className="text-xs font-bold text-[rgb(var(--color-text-tertiary))] uppercase tracking-wide mb-3">
                                    Status
                                </h3>
                                <CustomSelect
                                    value={task.status}
                                    onValueChange={(value) => handleUpdateField('status', value as TaskStatus)}
                                    options={STATUSES.map((s) => ({ value: s.id, label: s.label }))}
                                    placeholder="Select status"
                                />
                            </div>

                            {/* Editable Week Card */}
                            <div className="bg-[rgb(var(--color-bg-primary))] rounded-lg border border-[rgb(var(--color-border-primary))] p-4">
                                <h3 className="text-xs font-bold text-[rgb(var(--color-text-tertiary))] uppercase tracking-wide mb-3 flex items-center gap-2">
                                    <Calendar size={12} />
                                    Week
                                </h3>
                                <CustomSelect
                                    value={task.week.toString()}
                                    onValueChange={(value) => handleUpdateField('week', parseInt(value))}
                                    options={WEEKS.map((w) => ({ value: w.id.toString(), label: w.label }))}
                                    placeholder="Select week"
                                />
                                <p className="text-xs text-[rgb(var(--color-text-secondary))] mt-2">{currentWeek?.range}</p>
                            </div>

                            {/* Editable Type Card */}
                            {task.type && (
                                <div className="bg-[rgb(var(--color-bg-primary))] rounded-lg border border-[rgb(var(--color-border-primary))] p-4">
                                    <h3 className="text-xs font-bold text-[rgb(var(--color-text-tertiary))] uppercase tracking-wide mb-3">
                                        Type
                                    </h3>
                                    <CustomSelect
                                        value={task.type}
                                        onValueChange={(value) => handleUpdateField('type', value as TaskType)}
                                        options={typeOptions}
                                        placeholder="Select type"
                                    />
                                </div>
                            )}

                            {/* Saving Indicator */}
                            {isSaving && (
                                <div className="bg-[rgb(var(--color-primary-light))] text-[rgb(var(--color-primary))] px-4 py-2 rounded-lg text-center text-sm font-semibold">
                                    Saving...
                                </div>
                            )}

                            {/* Back Button */}
                            <button
                                onClick={() => router.push(`/tasks`)}
                                className="w-full bg-[rgb(var(--color-primary))] text-white px-4 py-3 rounded-lg hover:bg-[rgb(var(--color-primary-hover))] transition-all font-semibold text-sm"
                            >
                                Back to Board
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div >
    );
}
