import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Card, CardContent, CardFooter } from "@/components/ui/card";
import { 
  MoreHorizontal, 
  Play, 
  Square, 
  Trash2,
  Database,
  Server,
  Globe,
  Box,
  Layers,
  Loader2
} from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { cn } from "@/lib/utils";
import type { Stack } from "@/pages/stacks/contexts/stack-context";

interface StackCardProps {
  stack: Stack;
  onDelete: (id: string) => void;
}

export function StackCard({ stack, onDelete }: StackCardProps) {
  const navigate = useNavigate();
  const [isRunning, setIsRunning] = useState(stack.status === 'running');
  const [transitionState, setTransitionState] = useState<'idle' | 'stopping' | 'resuming'>('idle');
  
  // Determine icon based on stack template or name
  const getStackIcon = () => {
    const name = stack.name.toLowerCase();
    const template = stack.template?.toLowerCase() || '';
    
    if (template.includes('next') || name.includes('frontend') || name.includes('web')) {
      return <Globe className="h-5 w-5" />;
    } else if (template.includes('express') || name.includes('api') || name.includes('backend')) {
      return <Server className="h-5 w-5" />;
    } else if (template.includes('database') || name.includes('db') || name.includes('postgres') || name.includes('mongo')) {
      return <Database className="h-5 w-5" />;
    } else if (name.includes('nginx') || name.includes('proxy')) {
      return <Box className="h-5 w-5" />;
    } else {
      return <Layers className="h-5 w-5" />;
    }
  };

  const toggleRunning = () => {
    // Show transition state
    setTransitionState(isRunning ? 'stopping' : 'resuming');
    
    // Simulate stack state change with delay
    setTimeout(() => {
      setIsRunning(!isRunning);
      setTransitionState('idle');
    }, 2000);
  };

  // Format the date
  const formattedDate = new Date(stack.created).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric'
  });

  // Get status text and color based on current state
  const getStatusInfo = () => {
    if (transitionState === 'stopping') {
      return {
        text: "Stopping...",
        dotClass: "bg-amber-400"
      };
    } else if (transitionState === 'resuming') {
      return {
        text: "Starting...",
        dotClass: "bg-blue-400"
      };
    } else {
      return {
        text: isRunning ? "Running" : "Stopped",
        dotClass: isRunning ? "bg-green-500" : "bg-gray-400"
      };
    }
  };

  const handleCardClick = (e: React.MouseEvent) => {
    // Don't navigate if clicking on the menu or its children
    if (e.target instanceof Element) {
      const isMenu = e.target.closest('[data-menu="true"]');
      if (isMenu) return;
    }
    navigate(`/stacks/${stack.id}`);
  };

  const statusInfo = getStatusInfo();

  return (
    <Card 
      className="overflow-hidden transition-all duration-200 hover:shadow-md cursor-pointer"
      onClick={handleCardClick}
    >
      <CardContent className="p-0">
        <div className="flex items-center justify-between p-4 border-b">
          <div className="flex items-center">
            <div className={cn(
              "bg-primary/10 p-2 rounded-md mr-3",
              isRunning ? "text-green-600" : "text-gray-500"
            )}>
              {getStackIcon()}
            </div>
            <div>
              <h3 className="font-medium text-base">{stack.name}</h3>
              <p className="text-muted-foreground text-sm">{formattedDate}</p>
            </div>
          </div>
          
          <DropdownMenu>
            <DropdownMenuTrigger asChild disabled={transitionState !== 'idle'}>
              <button 
                className={cn(
                  "w-8 h-8 rounded-md hover:bg-gray-100 flex items-center justify-center",
                  transitionState !== 'idle' && "opacity-50 cursor-not-allowed"
                )}
                data-menu="true"
              >
                <MoreHorizontal className="h-5 w-5" />
                <span className="sr-only">Open menu</span>
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={toggleRunning}>
                {isRunning ? (
                  <>
                    <Square className="mr-2 h-4 w-4" />
                    <span>Stop Stack</span>
                  </>
                ) : (
                  <>
                    <Play className="mr-2 h-4 w-4" />
                    <span>Start Stack</span>
                  </>
                )}
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={() => onDelete(stack.id)} className="text-red-600">
                <Trash2 className="mr-2 h-4 w-4" />
                <span>Delete Stack</span>
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
        
        {stack.description && (
          <div className="px-4 py-2 text-sm text-muted-foreground">
            {stack.description}
          </div>
        )}
      </CardContent>
      
      <CardFooter className="flex justify-between items-center p-3 bg-gray-50">
        <div className="flex items-center">
          {transitionState !== 'idle' ? (
            <Loader2 className="h-3 w-3 mr-2 animate-spin text-blue-500" />
          ) : (
            <div className={cn(
              "h-2.5 w-2.5 rounded-full mr-2 flex-shrink-0",
              statusInfo.dotClass
            )}></div>
          )}
          <span className="text-xs font-medium leading-none">{statusInfo.text}</span>
        </div>
        
        {stack.template && (
          <div className="text-xs bg-white border px-2 py-0.5 rounded-full">
            {stack.template}
          </div>
        )}
      </CardFooter>
    </Card>
  );
}
