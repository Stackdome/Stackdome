import { BreadcrumbContext } from '@/contexts/breadcrumb-context';
import { useContext } from 'react';

export function useBreadcrumb() {
  return useContext(BreadcrumbContext);
}
