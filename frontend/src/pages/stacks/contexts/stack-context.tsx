import { createContext, useState, useContext } from 'react';
import type { Dispatch, ReactNode, SetStateAction } from 'react';
import type { Stack } from '@/pages/stacks/types';

// Context type definition
interface StackContextType {
  stacks: Stack[];
  // React's own setter type so consumers can use functional updates —
  // required for write-through from async callbacks (avoids clobbering the
  // list with a stale render-closure snapshot).
  setStacks: Dispatch<SetStateAction<Stack[]>>;
  addStack: (stack: Omit<Stack, 'id' | 'created_at' | 'lifecycle' | 'converged_release' | 'latest_release'>) => void;
  removeStack: (id: string) => void;
}

// Create the context
const StackContext = createContext<StackContextType | undefined>(undefined);

// Create a provider component
export function StackProvider({ children }: { children: ReactNode }) {
  const [stacks, setStacks] = useState<Stack[]>([]);

  // Add a new stack
  const addStack = (stackData: Omit<Stack, 'id' | 'created_at' | 'lifecycle' | 'converged_release' | 'latest_release'>) => {
    // id, created_at, and lifecycle are always set here; spec comes from stackData
    const newStack: Stack = {
      id: `stack-${Date.now()}`,
      created_at: new Date().toISOString(),
      lifecycle: 'active',
      ...stackData,
    };

    setStacks((prevStacks) => [...prevStacks, newStack]);
  };

  // Remove a stack
  const removeStack = (id: string) => {
    setStacks((prevStacks) => prevStacks.filter((stack) => stack.id !== id));
  };

  return (
    <StackContext.Provider value={{ stacks, setStacks, addStack, removeStack }}>
      {children}
    </StackContext.Provider>
  );
}

// Custom hook to use the stack context
export function useStacks() {
  const context = useContext(StackContext);

  if (context === undefined) {
    throw new Error('useStacks must be used within a StackProvider');
  }

  return context;
}
