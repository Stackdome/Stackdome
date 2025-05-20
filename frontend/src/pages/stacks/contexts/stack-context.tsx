import { createContext, useState, useContext } from 'react';
import type { ReactNode } from 'react';

// Stack type definition
export interface Stack {
  id: string;
  name: string;
  description?: string;
  status: 'running' | 'stopped' | 'deploying';
  template?: string;
  created: Date;
  icon?: string;
}

// Context type definition
interface StackContextType {
  stacks: Stack[];
  addStack: (stack: Omit<Stack, 'id' | 'created' | 'status'>) => void;
  removeStack: (id: string) => void;
}

// Create the context
const StackContext = createContext<StackContextType | undefined>(undefined);

// Create a provider component
export function StackProvider({ children }: { children: ReactNode }) {
  const [stacks, setStacks] = useState<Stack[]>([]);

  // Add a new stack
  const addStack = (stackData: Omit<Stack, 'id' | 'created' | 'status'>) => {
    const newStack: Stack = {
      id: `stack-${Date.now()}`,
      created: new Date(),
      status: 'running',
      ...stackData,
    };
    
    setStacks((prevStacks) => [...prevStacks, newStack]);
  };

  // Remove a stack
  const removeStack = (id: string) => {
    setStacks((prevStacks) => prevStacks.filter((stack) => stack.id !== id));
  };

  return (
    <StackContext.Provider value={{ stacks, addStack, removeStack }}>
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
