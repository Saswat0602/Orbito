import { createSlice, PayloadAction } from '@reduxjs/toolkit';

export type ViewMode = 'list' | 'board' | 'gantt' | 'calendar';

interface UiState {
  sidebarCollapsed: boolean;
  activeView: ViewMode;
  commandMenuOpen: boolean;
  taskDrawerId: string | null;
  theme: 'dark' | 'light';
  isAiChatOpen: boolean;
}

const initialState: UiState = {
  sidebarCollapsed: false,
  activeView: 'board',
  commandMenuOpen: false,
  taskDrawerId: null,
  theme: 'dark',
  isAiChatOpen: false,
};

export const uiSlice = createSlice({
  name: 'ui',
  initialState,
  reducers: {
    toggleSidebar: (state) => {
      state.sidebarCollapsed = !state.sidebarCollapsed;
    },
    setSidebarCollapsed: (state, action: PayloadAction<boolean>) => {
      state.sidebarCollapsed = action.payload;
    },
    setActiveView: (state, action: PayloadAction<ViewMode>) => {
      state.activeView = action.payload;
    },
    setCommandMenuOpen: (state, action: PayloadAction<boolean>) => {
      state.commandMenuOpen = action.payload;
    },
    setTaskDrawerId: (state, action: PayloadAction<string | null>) => {
      state.taskDrawerId = action.payload;
    },
    setTheme: (state, action: PayloadAction<'dark' | 'light'>) => {
      state.theme = action.payload;
    },
    setIsAiChatOpen: (state, action: PayloadAction<boolean>) => {
      state.isAiChatOpen = action.payload;
    },
    toggleAiChat: (state) => {
      state.isAiChatOpen = !state.isAiChatOpen;
    },
  },
});

export const {
  toggleSidebar,
  setSidebarCollapsed,
  setActiveView,
  setCommandMenuOpen,
  setTaskDrawerId,
  setTheme,
  setIsAiChatOpen,
  toggleAiChat,
} = uiSlice.actions;

export default uiSlice.reducer;

