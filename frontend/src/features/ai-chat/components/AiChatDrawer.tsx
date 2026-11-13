'use client';

import React, { useState, useEffect, useRef, useMemo, useCallback } from 'react';
import ReactMarkdown from 'react-markdown';
import {
  Sparkles,
  Send,
  X,
  RotateCcw,
  Plus,
  Trash2,
  History,
  Mic,
  MicOff,
  Copy,
  Check,
  Edit2,
  FolderKanban,
  CheckCircle2,
  AlertCircle,
  Loader2,
} from 'lucide-react';
import { toast } from 'sonner';

import { useAppDispatch, useAppSelector } from '@/lib/store/hooks';
import { setIsAiChatOpen } from '@/lib/store/slices/uiSlice';
import {
  useGetConversationsQuery,
  useCreateConversationMutation,
  useRenameConversationMutation,
  useDeleteConversationMutation,
  useSendChatMessageMutation,
  useClearAiContextMutation,
} from '@/lib/store/api/aiChatApi';
import { useGetWorkspacesQuery } from '@/lib/store/api/workspacesApi';
import { useGetProjectsQuery } from '@/lib/store/api/projectsApi';
import { useSpeechRecognition } from '@/lib/hooks/useSpeechRecognition';
import { AiChatMessage, AiConversation } from '@/types/ai';

const PROMPT_SUGGESTIONS = [
  { label: 'Summarize sprint', prompt: "Summarize this sprint's progress and highlight blockers." },
  { label: 'Overdue tasks', prompt: 'Which tasks are overdue or at risk of missing deadlines?' },
  { label: 'High priority bugs', prompt: 'What high-priority bugs need immediate attention?' },
  { label: 'Draft user story', prompt: 'Draft a user story with acceptance criteria for a new feature.' },
];

export function AiChatDrawer() {
  const dispatch = useAppDispatch();
  const isOpen = useAppSelector((state) => state.ui.isAiChatOpen);
  const authState = useAppSelector((state) => state.auth);

  // Active Context
  const activeOrgId = authState.activeOrganizationId;
  const activeWorkspaceId = authState.activeWorkspaceId;
  const activeWorkspaceSlug = authState.activeWorkspaceSlug;
  const activeProjectId = authState.activeProjectId;
  const activeProjectSlug = authState.activeProjectSlug;

  const { data: workspaces = [] } = useGetWorkspacesQuery(
    activeOrgId ? { organizationId: activeOrgId } : undefined,
    { skip: !activeOrgId }
  );

  const activeWorkspace = useMemo(() => {
    return (
      workspaces.find((w) => w.id === activeWorkspaceId) ||
      workspaces.find((w) => w.slug === activeWorkspaceSlug) ||
      workspaces[0] ||
      null
    );
  }, [workspaces, activeWorkspaceId, activeWorkspaceSlug]);

  const { data: projects = [] } = useGetProjectsQuery(
    activeWorkspace?.id ? { workspaceId: activeWorkspace.id } : undefined,
    { skip: !activeWorkspace?.id }
  );

  const activeProject = useMemo(() => {
    if (!activeProjectId && !activeProjectSlug) return null;
    return (
      projects.find((p) => p.id === activeProjectId) ||
      projects.find((p) => p.slug === activeProjectSlug) ||
      null
    );
  }, [projects, activeProjectId, activeProjectSlug]);

  // Session & Conversations
  const [currentSessionId, setCurrentSessionId] = useState<string>(() => `session_${Date.now()}`);
  const [activeConversationId, setActiveConversationId] = useState<string | null>(null);
  const [showHistory, setShowHistory] = useState<boolean>(false);
  const [editingConvId, setEditingConvId] = useState<string | null>(null);
  const [editTitle, setEditTitle] = useState<string>('');

  // Messages state
  const [messages, setMessages] = useState<AiChatMessage[]>([]);
  const [inputValue, setInputValue] = useState<string>('');
  const [copiedIndex, setCopiedIndex] = useState<number | null>(null);

  const messagesEndRef = useRef<HTMLDivElement | null>(null);
  const textareaRef = useRef<HTMLTextAreaElement | null>(null);

  // RTK Query Hooks
  const { data: conversations = [], refetch: refetchConversations } = useGetConversationsQuery(undefined, {
    skip: !isOpen,
  });
  const [createConversation, { isLoading: isCreatingConv }] = useCreateConversationMutation();
  const [renameConversation] = useRenameConversationMutation();
  const [deleteConversation] = useDeleteConversationMutation();
  const [sendChatMessage, { isLoading: isSending }] = useSendChatMessageMutation();
  const [clearAiContext, { isLoading: isClearing }] = useClearAiContextMutation();

  // Speech Recognition Hook
  const {
    isListening,
    transcript,
    isSupported: isSpeechSupported,
    startListening,
    stopListening,
    resetTranscript,
  } = useSpeechRecognition();

  // Sync transcript to input
  useEffect(() => {
    if (transcript) {
      setInputValue((prev) => {
        const trimmed = prev.trim();
        return trimmed ? `${trimmed} ${transcript}` : transcript;
      });
      resetTranscript();
    }
  }, [transcript, resetTranscript]);

  // Auto-scroll messages
  useEffect(() => {
    if (isOpen) {
      messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    }
  }, [messages, isSending, isOpen]);

  // Focus textarea when opened
  useEffect(() => {
    if (isOpen) {
      setTimeout(() => {
        textareaRef.current?.focus();
      }, 150);
    }
  }, [isOpen]);

  // Sync active conversation when selecting from history
  const handleSelectConversation = useCallback((conv: AiConversation) => {
    setActiveConversationId(conv.id);
    setCurrentSessionId(conv.sessionId);
    setMessages(
      (conv.messages || []).map((m) => ({
        role: m.role,
        content: m.content,
      }))
    );
    setShowHistory(false);
  }, []);

  // Start a new conversation
  const handleNewChat = useCallback(async () => {
    const newSession = `session_${Date.now()}`;
    setCurrentSessionId(newSession);
    setActiveConversationId(null);
    setMessages([]);
    setShowHistory(false);
    try {
      const res = await createConversation({ title: 'New Conversation', sessionId: newSession }).unwrap();
      setActiveConversationId(res.id);
    } catch {
      // Local fallback continues gracefully
    }
  }, [createConversation]);

  // Clear Context
  const handleClearContext = useCallback(async () => {
    try {
      await clearAiContext(currentSessionId).unwrap();
      setMessages([]);
      toast.success('Conversation context cleared');
    } catch {
      setMessages([]);
      toast.success('Context reset');
    }
  }, [clearAiContext, currentSessionId]);

  // Rename Conversation
  const handleSaveRename = useCallback(
    async (id: string) => {
      if (!editTitle.trim()) {
        setEditingConvId(null);
        return;
      }
      try {
        await renameConversation({ id, title: editTitle.trim() }).unwrap();
        setEditingConvId(null);
        refetchConversations();
        toast.success('Conversation renamed');
      } catch {
        toast.error('Failed to rename conversation');
      }
    },
    [editTitle, renameConversation, refetchConversations]
  );

  // Delete Conversation
  const handleDeleteConversation = useCallback(
    async (id: string, e: React.MouseEvent) => {
      e.stopPropagation();
      try {
        await deleteConversation(id).unwrap();
        if (activeConversationId === id) {
          handleNewChat();
        }
        refetchConversations();
        toast.success('Conversation deleted');
      } catch {
        toast.error('Failed to delete conversation');
      }
    },
    [activeConversationId, deleteConversation, handleNewChat, refetchConversations]
  );

  // Copy message to clipboard
  const handleCopy = useCallback((text: string, index: number) => {
    navigator.clipboard.writeText(text);
    setCopiedIndex(index);
    setTimeout(() => setCopiedIndex(null), 2000);
  }, []);

  // Send message
  const handleSendMessage = useCallback(
    async (promptToSend?: string) => {
      const text = promptToSend || inputValue;
      if (!text.trim() || isSending) return;

      const userMsg: AiChatMessage = { role: 'user', content: text.trim() };
      const nextMessages = [...messages, userMsg];
      setMessages(nextMessages);
      setInputValue('');

      try {
        const response = await sendChatMessage({
          message: text.trim(),
          history: messages,
          workspaceId: activeWorkspace?.id,
          projectId: activeProject?.id,
          sessionId: currentSessionId,
          currentOrganizationId: activeOrgId || undefined,
        }).unwrap();

        if (response.success && response.message) {
          setMessages([...nextMessages, { role: 'assistant', content: response.message }]);
        } else {
          setMessages([
            ...nextMessages,
            {
              role: 'assistant',
              content: response.message || 'I could not generate a response. Please check AI settings.',
            },
          ]);
        }
      } catch (err: unknown) {
        const errorMsg =
          (err as { data?: { message?: string } })?.data?.message ||
          'Failed to communicate with AI Copilot. Ensure AI is enabled in Administration Settings.';
        setMessages([
          ...nextMessages,
          {
            role: 'assistant',
            content: `⚠️ **Error:** ${errorMsg}`,
          },
        ]);
      }
    },
    [
      inputValue,
      isSending,
      messages,
      sendChatMessage,
      activeWorkspace,
      activeProject,
      currentSessionId,
      activeOrgId,
    ]
  );

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSendMessage();
    }
  };

  if (!isOpen) return null;

  return (
    <>
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-black/50 backdrop-blur-xs z-40 transition-opacity"
        onClick={() => dispatch(setIsAiChatOpen(false))}
        aria-hidden="true"
      />

      {/* Slide-out Drawer */}
      <aside
        className="fixed inset-y-0 right-0 w-full sm:w-[480px] z-50 flex flex-col bg-[#09090b] border-l border-zinc-800 shadow-2xl transition-transform duration-300 ease-in-out"
        role="dialog"
        aria-label="Orbito AI Copilot"
      >
        {/* Top Header */}
        <div className="h-14 px-4 border-b border-zinc-800/80 flex items-center justify-between shrink-0 bg-[#09090b]/80 backdrop-blur-md">
          <div className="flex items-center gap-2.5">
            <div className="p-1.5 rounded-md bg-zinc-800 border border-zinc-700 text-zinc-200">
              <Sparkles className="w-4 h-4" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="text-sm font-semibold text-zinc-100 tracking-tight">Orbito Copilot</span>
                <span className="text-[10px] uppercase font-mono px-1.5 py-0.5 rounded-sm bg-zinc-850 text-zinc-300 border border-zinc-750">
                  AI
                </span>
              </div>
              <div className="flex items-center gap-1.5 text-[11px] text-zinc-400 truncate max-w-[240px]">
                <FolderKanban className="w-3 h-3 text-zinc-500 shrink-0" />
                <span className="truncate">
                  {activeProject
                    ? `${activeWorkspace?.name || 'Workspace'} / ${activeProject.name}`
                    : activeWorkspace?.name || 'Global Context'}
                </span>
              </div>
            </div>
          </div>

          {/* Action Icons */}
          <div className="flex items-center gap-1">
            <button
              type="button"
              onClick={() => setShowHistory((prev) => !prev)}
              title="Chat History"
              className={`p-1.5 rounded-md transition-colors ${
                showHistory
                  ? 'bg-zinc-800 text-zinc-100'
                  : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800/60'
              }`}
            >
              <History className="w-4 h-4" />
            </button>
            <button
              type="button"
              onClick={handleClearContext}
              disabled={isClearing || messages.length === 0}
              title="Clear Context & Chat"
              className="p-1.5 rounded-md text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800/60 disabled:opacity-40 transition-colors"
            >
              <RotateCcw className="w-4 h-4" />
            </button>
            <button
              type="button"
              onClick={() => dispatch(setIsAiChatOpen(false))}
              title="Close (Esc)"
              className="p-1.5 rounded-md text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800/60 transition-colors"
            >
              <X className="w-4 h-4" />
            </button>
          </div>
        </div>

        {/* Conversation History Drawer (Sub-pane) */}
        {showHistory && (
          <div className="border-b border-zinc-800/80 bg-zinc-950/80 p-3 max-h-56 overflow-y-auto space-y-2">
            <div className="flex items-center justify-between pb-1.5 border-b border-zinc-800/60">
              <span className="text-xs font-medium text-zinc-300">Previous Sessions</span>
              <button
                type="button"
                onClick={handleNewChat}
                disabled={isCreatingConv}
                className="flex items-center gap-1 text-[11px] font-medium text-zinc-200 hover:text-white px-2 py-0.5 rounded-sm bg-zinc-850 border border-zinc-750 hover:bg-zinc-800 transition-colors cursor-pointer"
              >
                <Plus className="w-3 h-3" />
                New Chat
              </button>
            </div>

            {conversations.length === 0 ? (
              <p className="text-[11px] text-zinc-500 py-2 text-center">No previous conversations found.</p>
            ) : (
              <div className="space-y-1">
                {conversations.map((conv) => {
                  const isActive = conv.id === activeConversationId || conv.sessionId === currentSessionId;
                  const isEditing = editingConvId === conv.id;

                  return (
                    <div
                      key={conv.id}
                      onClick={() => !isEditing && handleSelectConversation(conv)}
                      className={`group flex items-center justify-between p-1.5 rounded-md text-xs cursor-pointer border transition-colors ${
                        isActive
                          ? 'bg-zinc-800/90 border-violet-500/30 text-zinc-100'
                          : 'bg-zinc-900/40 border-transparent text-zinc-400 hover:bg-zinc-900 hover:text-zinc-200'
                      }`}
                    >
                      {isEditing ? (
                        <div
                          className="flex items-center gap-1.5 w-full"
                          onClick={(e) => e.stopPropagation()}
                        >
                          <input
                            type="text"
                            value={editTitle}
                            onChange={(e) => setEditTitle(e.target.value)}
                            onKeyDown={(e) => {
                              if (e.key === 'Enter') handleSaveRename(conv.id);
                              if (e.key === 'Escape') setEditingConvId(null);
                            }}
                            className="flex-1 bg-zinc-800 border border-zinc-700 rounded-sm px-2 py-0.5 text-xs text-zinc-100 focus:outline-hidden"
                            autoFocus
                          />
                          <button
                            type="button"
                            onClick={() => handleSaveRename(conv.id)}
                            className="p-1 hover:text-emerald-400"
                          >
                            <Check className="w-3 h-3" />
                          </button>
                          <button
                            type="button"
                            onClick={() => setEditingConvId(null)}
                            className="p-1 hover:text-rose-400"
                          >
                            <X className="w-3 h-3" />
                          </button>
                        </div>
                      ) : (
                        <>
                          <span className="truncate flex-1 pr-2 font-medium">
                            {conv.title || 'Untitled Conversation'}
                          </span>
                          <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                            <button
                              type="button"
                              onClick={(e) => {
                                e.stopPropagation();
                                setEditingConvId(conv.id);
                                setEditTitle(conv.title || '');
                              }}
                              className="p-1 text-zinc-400 hover:text-zinc-200"
                              title="Rename"
                            >
                              <Edit2 className="w-3 h-3" />
                            </button>
                            <button
                              type="button"
                              onClick={(e) => handleDeleteConversation(conv.id, e)}
                              className="p-1 text-zinc-400 hover:text-rose-400"
                              title="Delete"
                            >
                              <Trash2 className="w-3 h-3" />
                            </button>
                          </div>
                        </>
                      )}
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        )}

        {/* Message Feed */}
        <div className="flex-1 overflow-y-auto p-4 space-y-4">
          {messages.length === 0 ? (
            <div className="h-full flex flex-col justify-center items-center text-center p-4 space-y-5">
              <div className="w-12 h-12 rounded-xl bg-zinc-850 border border-zinc-750 flex items-center justify-center text-zinc-200 shadow-inner">
                <Sparkles className="w-6 h-6" />
              </div>
              <div className="space-y-1.5">
                <h3 className="text-sm font-semibold text-zinc-200">How can I help you today?</h3>
                <p className="text-xs text-zinc-400 max-w-xs leading-relaxed">
                  I can summarize sprints, analyze workloads, identify risk bottlenecks, or draft task
                  descriptions.
                </p>
              </div>

              {/* Prompt Suggestions */}
              <div className="w-full space-y-2 pt-2">
                <span className="text-[11px] font-medium text-zinc-500 uppercase tracking-wider block text-left">
                  Suggested Prompts
                </span>
                <div className="grid grid-cols-1 gap-1.5">
                  {PROMPT_SUGGESTIONS.map((item, idx) => (
                    <button
                      key={idx}
                      type="button"
                      onClick={() => handleSendMessage(item.prompt)}
                      className="text-left px-3 py-2 rounded-md bg-zinc-900/60 border border-zinc-800/80 hover:border-zinc-600 hover:bg-zinc-850/60 text-xs text-zinc-300 transition-all flex items-center justify-between group cursor-pointer"
                    >
                      <span className="font-medium text-zinc-200">{item.label}</span>
                      <span className="text-[11px] text-zinc-500 group-hover:text-zinc-300 truncate ml-2">
                        {item.prompt}
                      </span>
                    </button>
                  ))}
                </div>
              </div>
            </div>
          ) : (
            messages.map((msg, idx) => {
              const isUser = msg.role === 'user';

              return (
                <div
                  key={idx}
                  className={`flex flex-col ${isUser ? 'items-end' : 'items-start'} space-y-1`}
                >
                  <div className="flex items-center gap-1.5 px-1">
                    <span className="text-[11px] font-medium text-zinc-500">
                      {isUser ? 'You' : 'Orbito Copilot'}
                    </span>
                  </div>

                  <div
                    className={`relative group max-w-[85%] rounded-md px-3.5 py-2.5 text-xs leading-relaxed ${
                      isUser
                        ? 'bg-zinc-800 text-zinc-100 border border-zinc-700/50'
                        : 'bg-zinc-900/90 text-zinc-200 border border-zinc-800/80 shadow-xs'
                    }`}
                  >
                    {!isUser && (
                      <button
                        type="button"
                        onClick={() => handleCopy(msg.content, idx)}
                        className="absolute top-2 right-2 p-1 rounded-sm bg-zinc-800/80 text-zinc-400 hover:text-zinc-200 opacity-0 group-hover:opacity-100 transition-opacity"
                        title="Copy Response"
                      >
                        {copiedIndex === idx ? (
                          <Check className="w-3 h-3 text-emerald-400" />
                        ) : (
                          <Copy className="w-3 h-3" />
                        )}
                      </button>
                    )}

                    {isUser ? (
                      <p className="whitespace-pre-wrap">{msg.content}</p>
                    ) : (
                      <div className="prose prose-invert prose-xs max-w-none space-y-2 [&_p]:leading-relaxed [&_ul]:list-disc [&_ul]:pl-4 [&_ol]:list-decimal [&_ol]:pl-4 [&_code]:px-1 [&_code]:py-0.5 [&_code]:bg-zinc-800 [&_code]:rounded-xs [&_code]:font-mono [&_pre]:bg-zinc-950 [&_pre]:p-2 [&_pre]:rounded-md [&_pre]:border [&_pre]:border-zinc-800">
                        <ReactMarkdown>{msg.content}</ReactMarkdown>
                      </div>
                    )}
                  </div>
                </div>
              );
            })
          )}

          {isSending && (
            <div className="flex items-center gap-2 text-xs text-zinc-400 py-2">
              <Loader2 className="w-3.5 h-3.5 animate-spin text-violet-400" />
              <span>Thinking and analyzing project data...</span>
            </div>
          )}

          <div ref={messagesEndRef} />
        </div>

        {/* Input & Voice Controls */}
        <div className="p-3 border-t border-zinc-800/80 bg-[#09090b]/90 backdrop-blur-md space-y-2">
          {/* Active Voice Listening Banner */}
          {isListening && (
            <div className="flex items-center justify-between px-2.5 py-1.5 rounded-md bg-rose-950/40 border border-rose-800/50 text-[11px] text-rose-300 animate-pulse">
              <div className="flex items-center gap-1.5">
                <span className="w-2 h-2 rounded-full bg-rose-500 animate-ping" />
                <span>Listening to voice query... Speak now</span>
              </div>
              <button
                type="button"
                onClick={stopListening}
                className="text-xs font-semibold text-rose-400 hover:underline"
              >
                Done
              </button>
            </div>
          )}

          <div className="relative flex items-end gap-1.5 bg-zinc-900 border border-zinc-800 rounded-md p-1.5 focus-within:border-zinc-600 transition-colors">
            <textarea
              ref={textareaRef}
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="Ask Orbito Copilot... (Press Enter to send)"
              rows={2}
              className="flex-1 bg-transparent border-0 resize-none text-xs text-zinc-100 placeholder:text-zinc-500 focus:outline-hidden p-1.5 max-h-32"
            />

            <div className="flex items-center gap-1 shrink-0 pb-0.5">
              {isSpeechSupported && (
                <button
                  type="button"
                  onClick={isListening ? stopListening : startListening}
                  className={`p-1.5 rounded-md transition-colors ${
                    isListening
                      ? 'bg-rose-500/20 text-rose-400 border border-rose-500/40'
                      : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800'
                  }`}
                  title={isListening ? 'Stop listening' : 'Voice input (Speech to text)'}
                >
                  {isListening ? <MicOff className="w-4 h-4 text-rose-400" /> : <Mic className="w-4 h-4" />}
                </button>
              )}

              <button
                type="button"
                onClick={() => handleSendMessage()}
                disabled={!inputValue.trim() || isSending}
                className="p-1.5 rounded-md bg-zinc-100 text-zinc-950 hover:bg-white disabled:opacity-40 disabled:hover:bg-zinc-100 transition-colors cursor-pointer shadow-xs"
                title="Send Message"
              >
                <Send className="w-4 h-4" />
              </button>
            </div>
          </div>

          <div className="flex items-center justify-between text-[10px] text-zinc-500 px-1">
            <span>
              Toggle anytime with <kbd className="px-1 py-0.5 rounded-xs bg-zinc-800 text-zinc-300 font-mono">⌘J</kbd>
            </span>
            <span>Shift+Enter for newline</span>
          </div>
        </div>
      </aside>
    </>
  );
}
