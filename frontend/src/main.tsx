import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.tsx'
import { getAppConfig } from '@/api/config'

// Warm the public app-config (GitHub OAuth gate) before React mounts so the
// auth screens can render the "Continue with GitHub" button without waiting on
// a per-page request — matters on slow networks.
void getAppConfig()

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
