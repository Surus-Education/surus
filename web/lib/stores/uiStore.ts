import { create } from "zustand";

interface UIStore {
  authPromptOpen: boolean;
  authPromptAction: string | null;
  openAuthPrompt: (action: string) => void;
  closeAuthPrompt: () => void;
}

export const useUIStore = create<UIStore>((set) => ({
  authPromptOpen: false,
  authPromptAction: null,
  openAuthPrompt: (action: string) => set({ authPromptOpen: true, authPromptAction: action }),
  closeAuthPrompt: () => set({ authPromptOpen: false, authPromptAction: null }),
}));
